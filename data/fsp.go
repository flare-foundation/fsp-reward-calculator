package data

import (
	"bytes"
	"encoding/hex"
	voters "fsp-rewards-calculator/lib"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type TxInfo struct {
	TimestampSec uint64
	Reverted     bool
	From         common.Address
}

type SignatureType0 struct {
	bytes      []byte
	merkleRoot ProtocolMerkleRoot
	message    []byte // TODO: remove once we stop accepting signature type 0 for FDC submissions
}

type SignatureType1 struct {
	bytes   []byte
	message []byte
}

type Finalization struct {
	Policy     voters.SigningPolicy
	MerkleRoot ProtocolMerkleRoot
	Signatures []ECDSASignature
	Info       TxInfo
}

func GetSubmit1(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[uint8][]payload.Message, error) {
	logger.Info("Fetching submit1 for rounds %d-%d", fromRound, toRound)

	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound)
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound)

	msgs, err := querySubmissions(db, fromSec, toSec, utils.FunctionSignatures.Submit1, params.Net.Contracts.Submission)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	logger.Info("Done fetching submit1 for rounds %d-%d", fromRound, toRound)
	return msgs, nil
}

func GetSubmit2(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[uint8][]payload.Message, error) {
	logger.Info("Fetching submit2 for rounds %d-%d", fromRound, toRound)

	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound.Add(1))
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound.Add(1))

	msgs, err := querySubmissions(db, fromSec, toSec, utils.FunctionSignatures.Submit2, params.Net.Contracts.Submission)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	logger.Info("Done fetching submit2 for rounds %d-%d", fromRound, toRound)
	return msgs, nil
}

func GetSubmitSignatures(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[uint8][]payload.Message, error) {
	logger.Info("Fetching submitSignatures for rounds %d-%d", fromRound, toRound)

	fromSec := params.Net.Epoch.RevealDeadlineSec(fromRound+1) + 1
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows))

	msgs, err := querySubmissions(db, fromSec, toSec, utils.FunctionSignatures.SubmitSignatures, params.Net.Contracts.Submission)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	logger.Info("Done fetching submitSignatures for rounds %d-%d", fromRound, toRound)
	return msgs, nil
}

func GetFinalizations(db *gorm.DB, re *RewardEpoch, fromRound ty.RoundId, toRound ty.RoundId) (map[uint8][]*Finalization, error) {
	logger.Info("Fetching finalizations for rounds %d-%d", fromRound, toRound)

	fromSec := params.Net.Epoch.RevealDeadlineSec(fromRound+1) + 1
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows))

	txns, err := fetchTransactions(db, params.Net.Contracts.Relay, utils.FunctionSignatures.Relay, int64(fromSec), int64(toSec))
	if err != nil {
		return nil, errors.Errorf("error fetching txns From DB: %s", err)
	}

	finalizationsByProtocol := map[uint8][]*Finalization{}
	for _, txn := range txns {
		finalization, err := DecodeFinalization(txn.Input[8:])
		if err != nil {
			logger.Debug("error parsing finalization, skipping: %+v", err)
			continue
		}

		if ty.EpochId(finalization.Policy.RewardEpochId) != re.Epoch {
			logger.Debug("Finalization reward epoch %d does not match expected epoch %d, skipping", finalization.Policy.RewardEpochId, re.Epoch)
			continue
		}

		if !bytes.Equal(finalization.Policy.RawBytes, re.Policy.RawBytes) {
			logger.Debug("Finalization signing policy does not match expected, skipping")
			continue
		}

		expectedRound := params.Net.Epoch.VotingRoundForTimeSec(txn.Timestamp) - 1
		round := finalization.MerkleRoot.round
		if round != expectedRound {
			logger.Debug("finalization round %d does not match expected round %d, skipping", round, expectedRound)
			continue
		}
		if round < fromRound || round > toRound {
			logger.Debug("finalization round %d is not in range [%d, %d], skipping", round, fromRound, toRound)
			continue
		}

		// TODO: Clean up filling in tx info: should be done on creation
		finalization.Info = TxInfo{
			From:         common.HexToAddress(txn.FromAddress),
			TimestampSec: txn.Timestamp,
			Reverted:     txn.Status != 1,
		}

		protocolId := uint8(finalization.MerkleRoot.protocolId)
		if _, ok := finalizationsByProtocol[protocolId]; !ok {
			finalizationsByProtocol[protocolId] = []*Finalization{}
		}
		finalizationsByProtocol[protocolId] = append(finalizationsByProtocol[protocolId], finalization)
	}

	logger.Info("Done fetching finalizations for rounds %d-%d", fromRound, toRound)
	return finalizationsByProtocol, nil
}

func querySubmissions(db *gorm.DB, fromSec uint64, toSec uint64, signature [4]byte, contractAddress common.Address) (map[uint8][]payload.Message, error) {
	txns, err := fetchTransactions(db, contractAddress, signature, int64(fromSec), int64(toSec))
	if err != nil {
		return nil, errors.Errorf("error fetching txns From DB: %s", err)
	}

	messagesByProtocol := map[uint8][]payload.Message{}
	for _, tx := range txns {
		payloadsByProtocol, err := payload.ExtractPayloads(&tx)
		if err != nil {
			logger.Info("error extracting payloads, skipping submission: %s", err)
			continue
		}

		for protocolId, message := range payloadsByProtocol {
			if _, ok := messagesByProtocol[protocolId]; !ok {
				messagesByProtocol[protocolId] = []payload.Message{}
			}
			messagesByProtocol[protocolId] = append(messagesByProtocol[protocolId], message)
		}
	}

	return messagesByProtocol, nil
}

// fetchTransactions retrieves transactions from the database that match the given criteria.
// This is an optimised version that selects only the necessary columns.
func fetchTransactions(
	db *gorm.DB, toAddress common.Address, functionSel [4]byte, from int64, to int64,
) ([]database.Transaction, error) {
	var transactions []database.Transaction

	err := db.Model(database.Transaction{}).
		Where(
			"to_address = ? AND function_sig = ? AND timestamp >= ? AND timestamp <= ?",
			hex.EncodeToString(toAddress[:]), // encodes without 0x prefix and without checksum
			hex.EncodeToString(functionSel[:]),
			from, to,
		).
		Order("timestamp ASC").
		Order("block_number ASC").
		Order("transaction_index ASC").
		// Optimisation: select only the necessary columns
		Select("function_sig", "input", "block_number", "from_address", "status", "timestamp").
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

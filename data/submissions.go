package data

import (
	"encoding/hex"
	voters "fsp-rewards-calculator/lib"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fupdater"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
)

const (
	FtsoScalingProtocolId = 100
)

type Commit struct {
	Hash common.Hash
}

type Reveal struct {
	Random        common.Hash
	EncodedValues []byte
}

type SignatureSubmission struct {
	Signature *Signature
	Info      TxInfo
}

type TxInfo struct {
	TimestampSec uint64
	Reverted     bool
	From         common.Address
}

type Signature struct {
	bytes      []byte
	merkleRoot ProtocolMerkleRoot
}

type Finalization struct {
	Policy     voters.SigningPolicy
	MerkleRoot ProtocolMerkleRoot
	Signatures []ECDSASignature
	Info       TxInfo
}

type FUpdateFeed struct {
	Values   []*big.Int
	Decimals []int8
}

// GetCommits retrieves the last commit submission for each voter for each round in the given range
func GetCommits(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[ty.RoundId]map[ty.VoterSubmit]*Commit, error) {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound)
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound)

	msgs, err := querySubmissions(db, fromSec, toSec, utils.FunctionSignatures.Submit1, params.Net.Contracts.Submission)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var commitsByRound = map[ty.RoundId]map[ty.VoterSubmit]*Commit{}
	for _, msg := range msgs {
		round := ty.RoundId(msg.VotingRound)
		submitRound := params.Net.Epoch.VotingRoundForTimeSec(msg.Timestamp)
		if round != submitRound {
			logger.Debug("commit round %d does not match expected round %d, skipping", round, submitRound)
			continue
		}

		commit, err := DecodeCommit(msg.Payload)
		if err != nil {
			logger.Debug("error parsing commit, skipping: %s", err)
			continue
		}

		if _, ok := commitsByRound[round]; !ok {
			commitsByRound[round] = map[ty.VoterSubmit]*Commit{}
		}

		from := ty.VoterSubmit(msg.From)
		commitsByRound[round][from] = commit
	}

	return commitsByRound, nil
}

// GetReveals retrieves the last reveal submission for voter for each round in the given range
func GetReveals(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId, feeds []Feed) (map[ty.RoundId]map[ty.VoterSubmit]*Reveal, error) {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound.Add(1))
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound.Add(1))

	msgs, err := querySubmissions(db, fromSec, toSec, utils.FunctionSignatures.Submit2, params.Net.Contracts.Submission)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var revealsByRound = map[ty.RoundId]map[ty.VoterSubmit]*Reveal{}
	for _, msg := range msgs {
		round := ty.RoundId(msg.VotingRound)
		submitRound := params.Net.Epoch.VotingRoundForTimeSec(msg.Timestamp)
		if round != submitRound.Add(-1) {
			logger.Debug("reveal round %d does not match expected round %d, skipping", round, submitRound.Add(-1))
			continue
		}

		if msg.Timestamp > params.Net.Epoch.RevealDeadlineSec(submitRound) {
			logger.Debug("reveal from %s too late", msg.From)
			continue
		}

		reveal, err := DecodeReveal(msg.Payload, len(feeds))
		if err != nil {
			logger.Debug("error parsing reveal, skipping: %s", err)
			continue
		}

		if _, ok := revealsByRound[round]; !ok {
			revealsByRound[round] = map[ty.VoterSubmit]*Reveal{}
		}

		from := ty.VoterSubmit(msg.From)
		revealsByRound[round][from] = reveal
	}

	return revealsByRound, nil
}

func getSignatures(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[ty.RoundId][]*SignatureSubmission, error) {
	fromSec := params.Net.Epoch.RevealDeadlineSec(fromRound+1) + 1
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows))

	msgs, err := querySubmissions(db, fromSec, toSec, utils.FunctionSignatures.SubmitSignatures, params.Net.Contracts.Submission)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var signaturesByRound = map[ty.RoundId][]*SignatureSubmission{}
	for _, msg := range msgs {
		round := ty.RoundId(msg.VotingRound)
		expectedRound := params.Net.Epoch.VotingRoundForTimeSec(msg.Timestamp) - 1
		if round != expectedRound {
			logger.Debug("Signature round %d does not match expected round %d, skipping", round, expectedRound)
			continue
		}

		signature, err := DecodeSignature(msg.Payload)
		if err != nil {
			logger.Debug("error parsing Signature, skipping: %s", err)
			continue
		}

		if _, ok := signaturesByRound[round]; !ok {
			signaturesByRound[round] = []*SignatureSubmission{}
		}

		signaturesByRound[round] = append(signaturesByRound[round], &SignatureSubmission{
			Signature: signature,
			Info: TxInfo{
				From:         msg.From,
				TimestampSec: msg.Timestamp,
				Reverted:     false,
			},
		})

	}

	return signaturesByRound, nil
}

func getFinalizations(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[ty.RoundId][]*Finalization, error) {
	fromSec := params.Net.Epoch.RevealDeadlineSec(fromRound+1) + 1
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows))

	txns, err := fetchTransactions(db, params.Net.Contracts.Relay, utils.FunctionSignatures.Relay, int64(fromSec), int64(toSec))
	if err != nil {
		return nil, errors.Errorf("error fetching txns From DB: %s", err)
	}

	var byRound = map[ty.RoundId][]*Finalization{}

	for _, txn := range txns {
		expectedRound := params.Net.Epoch.VotingRoundForTimeSec(txn.Timestamp) - 1

		finalization, err := DecodeFinalization(txn.Input[8:])
		if err != nil {
			logger.Info("error parsing finalization, skipping: %+v", err)
			continue
		}
		if finalization.MerkleRoot.protocolId != FtsoScalingProtocolId {
			logger.Debug("finalization protocol %d does not match expected protocol %d, skipping", finalization.MerkleRoot.protocolId, FtsoScalingProtocolId)
			continue
		}
		if finalization.MerkleRoot.round != expectedRound {
			logger.Debug("finalization round %d does not match expected round %d, skipping", finalization.MerkleRoot.round, expectedRound)
			continue
		}

		if _, ok := byRound[expectedRound]; !ok {
			byRound[expectedRound] = []*Finalization{}
		}

		// TODO: Clean up filling in tx info: should be done on creation
		finalization.Info = TxInfo{
			From:         common.HexToAddress(txn.FromAddress),
			TimestampSec: txn.Timestamp,
			Reverted:     txn.Status != 1,
		}

		byRound[expectedRound] = append(byRound[expectedRound], finalization)
	}

	return byRound, nil
}

func getFUpdateFeeds(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[ty.RoundId]*FUpdateFeed, error) {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound + 1)
	toSec := params.Net.Epoch.VotingRoundStartSec(toRound.Add(2)) // extra round for buffer

	instance, _ := fupdater.NewFUpdater(common.Address{}, nil)
	parse := func(log types.Log) (*fupdater.FUpdaterFastUpdateFeeds, error) {
		return instance.FUpdaterFilterer.ParseFastUpdateFeeds(log)
	}

	events, err := queryEvents(
		db,
		fromSec,
		toSec,
		params.Net.Contracts.FastUpdater,
		utils.EventTopic0.FastUpdateFeeds,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	var byRound = map[ty.RoundId]*FUpdateFeed{}

	for _, event := range events {
		round := ty.RoundId(event.VotingEpochId.Uint64())
		if round < fromRound || round > toRound {
			continue
		}

		byRound[round] = &FUpdateFeed{
			Values:   event.Feeds,
			Decimals: event.Decimals,
		}
	}

	return byRound, nil
}

func getFUpdateSubmits(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[ty.RoundId][]ty.VoterSigning, error) {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound)
	toSec := params.Net.Epoch.VotingRoundStartSec(toRound.Add(2)) // Add extra round as buffer

	instance, _ := fupdater.NewFUpdater(common.Address{}, nil)
	parse := func(log types.Log) (*fupdater.FUpdaterFastUpdateFeedsSubmitted, error) {
		return instance.FUpdaterFilterer.ParseFastUpdateFeedsSubmitted(log)
	}

	events, err := queryEvents(
		db,
		fromSec,
		toSec,
		params.Net.Contracts.FastUpdater,
		utils.EventTopic0.FastUpdateFeedsSubmitted,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	var byRound = map[ty.RoundId][]ty.VoterSigning{}

	for _, event := range events {
		round := ty.RoundId(event.VotingRoundId)
		if round < fromRound || round > toRound {
			continue
		}

		if _, ok := byRound[round]; !ok {
			byRound[round] = []ty.VoterSigning{}
		}

		byRound[round] = append(byRound[round], ty.VoterSigning(event.SigningPolicyAddress))
	}

	return byRound, nil
}

func querySubmissions(db *gorm.DB, fromSec uint64, toSec uint64, signature [4]byte, contractAddress common.Address) ([]payload.Message, error) {
	txns, err := fetchTransactions(db, contractAddress, signature, int64(fromSec), int64(toSec))
	if err != nil {
		return nil, errors.Errorf("error fetching txns From DB: %s", err)
	}

	var payloads []payload.Message
	for _, tx := range txns {
		payloadsByProtocol, err := payload.ExtractPayloads(&tx)
		if err != nil {
			logger.Info("error extracting payloads, skipping submission: %s", err)
			continue
		}

		scalingPayload, ok := payloadsByProtocol[FtsoScalingProtocolId]
		if !ok {
			continue
		}

		payloads = append(payloads, scalingPayload)
	}

	return payloads, nil
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

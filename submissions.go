package main

import (
	"flare-common/database"
	"flare-common/payload"
	"flare-common/policy"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/types"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	FtsoScalingProtocolId = 100
)

var (
	submissionContractAddress = params.Net.Contracts.Submission
	relayContractAddress      = params.Net.Contracts.Relay
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
	Policy     policy.SigningPolicy
	merkleRoot ProtocolMerkleRoot
	Signatures []ECDSASignature
	Info       TxInfo
}

// TODO: Make sure DB query sorts both by timestamp and tx index

// getCommits retrieves the last commit submission for each voter for each round in the given range
func getCommits(db *gorm.DB, fromRound types.RoundId, toRound types.RoundId) (map[types.RoundId]map[VoterSubmit]*Commit, error) {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound)
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound)

	msgs, err := querySubmissions(db, fromSec, toSec, utils.FunctionSignatures.Submit1, submissionContractAddress)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var commitsByRound = map[types.RoundId]map[VoterSubmit]*Commit{}
	for _, msg := range msgs {
		round := types.RoundId(msg.VotingRound)
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
			commitsByRound[round] = map[VoterSubmit]*Commit{}
		}

		from := VoterSubmit(common.HexToAddress(msg.From))
		commitsByRound[round][from] = commit
	}

	return commitsByRound, nil
}

// getReveals retrieves the last reveal submission for voter for each round in the given range
func getReveals(db *gorm.DB, fromRound types.RoundId, toRound types.RoundId) (map[types.RoundId]map[VoterSubmit]*Reveal, error) {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound.Add(1))
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound.Add(1))

	msgs, err := querySubmissions(db, fromSec, toSec, utils.FunctionSignatures.Submit2, submissionContractAddress)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var revealsByRound = map[types.RoundId]map[VoterSubmit]*Reveal{}
	for _, msg := range msgs {
		round := types.RoundId(msg.VotingRound)
		submitRound := params.Net.Epoch.VotingRoundForTimeSec(msg.Timestamp)
		if round != submitRound.Add(-1) {
			logger.Debug("reveal round %d does not match expected round %d, skipping", round, submitRound.Add(-1))
			continue
		}

		if msg.Timestamp > params.Net.Epoch.RevealDeadlineSec(submitRound) {
			// TODO: all seem to fail here??
			logger.Debug("reveal from %s too late", common.HexToAddress(msg.From))
			continue
		}

		reveal, err := DecodeReveal(msg.Payload)
		if err != nil {
			logger.Debug("error parsing reveal, skipping: %s", err)
			continue
		}

		if _, ok := revealsByRound[round]; !ok {
			revealsByRound[round] = map[VoterSubmit]*Reveal{}
		}

		from := VoterSubmit(common.HexToAddress(msg.From))
		revealsByRound[round][from] = reveal
	}

	return revealsByRound, nil
}

func getSignatures(db *gorm.DB, fromRound types.RoundId, toRound types.RoundId) (map[types.RoundId][]*SignatureSubmission, error) {
	fromSec := params.Net.Epoch.RevealDeadlineSec(fromRound+1) + 1
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows))

	msgs, err := querySubmissions(db, fromSec, toSec, utils.FunctionSignatures.SubmitSignatures, submissionContractAddress)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var signaturesByRound = map[types.RoundId][]*SignatureSubmission{}
	for _, msg := range msgs {
		round := types.RoundId(msg.VotingRound)
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
				From:         common.HexToAddress(msg.From),
				TimestampSec: msg.Timestamp,
				Reverted:     false,
			},
		})

	}

	return signaturesByRound, nil
}

func getFinalizations(db *gorm.DB, fromRound types.RoundId, toRound types.RoundId) (map[types.RoundId][]*Finalization, error) {
	fromSec := params.Net.Epoch.RevealDeadlineSec(fromRound+1) + 1
	toSec := params.Net.Epoch.VotingRoundEndSec(toRound.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows))

	txns, err := database.FetchTransactionsByAddressAndSelectorTimestamp(db, relayContractAddress, utils.FunctionSignatures.Relay, int64(fromSec), int64(toSec))
	if err != nil {
		return nil, errors.Errorf("error fetching txns From DB: %s", err)
	}

	var finalizationsByRound = map[types.RoundId][]*Finalization{}

	for _, txn := range txns {
		expectedRound := params.Net.Epoch.VotingRoundForTimeSec(txn.Timestamp) - 1

		finalization, err := DecodeFinalization(txn.Input[8:])
		if err != nil {
			logger.Info("error parsing finalization, skipping: %+v", err)
			continue
		}
		if finalization.merkleRoot.protocolId != FtsoScalingProtocolId {
			logger.Debug("finalization protocol %d does not match expected protocol %d, skipping", finalization.merkleRoot.protocolId, FtsoScalingProtocolId)
			continue
		}
		if finalization.merkleRoot.round != expectedRound {
			logger.Debug("finalization round %d does not match expected round %d, skipping", finalization.merkleRoot.round, expectedRound)
			continue
		}

		if _, ok := finalizationsByRound[expectedRound]; !ok {
			finalizationsByRound[expectedRound] = []*Finalization{}
		}

		// TODO: Clean up filling in tx info: should be done on creation
		finalization.Info = TxInfo{
			From:         common.HexToAddress(txn.FromAddress),
			TimestampSec: txn.Timestamp,
			Reverted:     txn.Status != 1,
		}

		finalizationsByRound[expectedRound] = append(finalizationsByRound[expectedRound], finalization)
	}

	return finalizationsByRound, nil
}

func querySubmissions(db *gorm.DB, fromSec uint64, toSec uint64, signature [4]byte, contractAddress common.Address) ([]payload.Message, error) {
	txns, err := database.FetchTransactionsByAddressAndSelectorTimestamp(db, contractAddress, signature, int64(fromSec), int64(toSec))
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

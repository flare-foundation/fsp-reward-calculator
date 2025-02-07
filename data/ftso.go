package data

import (
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fupdater"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
)

const (
	FtsoProtocolId = 100
)

type Commit struct {
	Hash common.Hash
}

type Reveal struct {
	Random        common.Hash
	EncodedValues []byte
}

type SignatureSubmission struct {
	Signature *SignatureType0
	Info      TxInfo
}

func extractCommits(messages []payload.Message) (map[ty.RoundId]map[ty.VoterSubmit]*Commit, error) {
	var commitsByRound = map[ty.RoundId]map[ty.VoterSubmit]*Commit{}
	for _, msg := range messages {
		if msg.ProtocolID != FtsoProtocolId {
			logger.Fatal("message protocol %d does not match expected protocol %d - likely an issue with submission retrieval")
		}

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

func extractReveals(messages []payload.Message, epochs RewardEpochs) (map[ty.RoundId]map[ty.VoterSubmit]*Reveal, error) {
	var revealsByRound = map[ty.RoundId]map[ty.VoterSubmit]*Reveal{}
	for _, msg := range messages {
		if msg.ProtocolID != FtsoProtocolId {
			logger.Fatal("message protocol %d does not match expected protocol %d - likely an issue with submission retrieval")
		}

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

		feeds := epochs.EpochForRound(round).OrderedFeeds
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

func extractSignatures(messages []payload.Message) (map[ty.RoundId][]*SignatureSubmission, error) {
	var signaturesByRound = map[ty.RoundId][]*SignatureSubmission{}
	for _, msg := range messages {
		if msg.ProtocolID != FtsoProtocolId {
			logger.Fatal("message protocol %d does not match expected protocol %d - likely an issue with submission retrieval")
		}

		round := ty.RoundId(msg.VotingRound)
		expectedRound := params.Net.Epoch.VotingRoundForTimeSec(msg.Timestamp) - 1
		if round != expectedRound {
			logger.Debug("Signature round %d does not match expected round %d, skipping", round, expectedRound)
			continue
		}

		signature, err := DecodeSignatureType0(msg.Payload)
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

type FUpdateFeed struct {
	Values   []*big.Int
	Decimals []int8
}

func getFUpdateFeeds(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[ty.RoundId]*FUpdateFeed, error) {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound + 1)
	toSec := params.Net.Epoch.VotingRoundStartSec(toRound.Add(2)) // extra round for buffer

	instance, _ := fupdater.NewFUpdater(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*fupdater.FUpdaterFastUpdateFeeds, error) {
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
	parse := func(log types.Log, _ uint64) (*fupdater.FUpdaterFastUpdateFeedsSubmitted, error) {
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

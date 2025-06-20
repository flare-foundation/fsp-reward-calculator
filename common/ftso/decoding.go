package ftso

import (
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"github.com/pkg/errors"
)

func DecodeFeedValues(bytes []byte, feeds []fsp.Feed) ([]FeedValue, error) {
	if (len(bytes) % feedValueBytes) != 0 {
		return nil, errors.New("invalid message length for feed values")
	}

	var feedValues []FeedValue
	for i := 0; i < len(bytes); i += feedValueBytes {
		rawValue := fsp.DecodeUint32(bytes[i : i+feedValueBytes])

		var feedValue FeedValue
		if rawValue == noValue {
			feedValue = EmptyFeedValue
		} else {
			feedValue = FeedValue{
				IsEmpty: false,
				Value:   int32(rawValue - 1<<31), // Values encoded in Excess-2^31
			}
		}
		feedValues = append(feedValues, feedValue)
	}

	// Fill in values for truncated empty feeds
	for i := len(feedValues); i < len(feeds); i++ {
		feedValues = append(feedValues, EmptyFeedValue)
	}

	return feedValues, nil
}

func DecodeCommit(bytes []byte) (*Commit, error) {
	if len(bytes) != common.HashLength {
		return nil, errors.New("invalid message length")
	}
	hash := common.BytesToHash(bytes)
	return &Commit{
		Hash: hash,
	}, nil
}

func DecodeReveal(bytes []byte, expectedFeeds int) (*Reveal, error) {
	// The message should be long enough to contain the random and at least one feed value
	if len(bytes) < (common.HashLength + feedValueBytes) {
		return nil, errors.New("message too short")
	}

	random := common.BytesToHash(bytes[:common.HashLength])
	encodedFeeds := bytes[common.HashLength:]

	if (len(encodedFeeds) % feedValueBytes) != 0 {
		return nil, errors.Errorf("invalid message length %d for feed values", len(encodedFeeds))
	}
	if (len(encodedFeeds) / feedValueBytes) > expectedFeeds {
		return nil, errors.Errorf("encoded feed values payload %d exceeds expected number of feeds %d", len(encodedFeeds)/feedValueBytes, expectedFeeds)
	}

	return &Reveal{
		Random:        random,
		EncodedValues: encodedFeeds,
	}, nil
}

func ExtractCommits(messages []payload.Message) (map[ty.RoundId]map[ty.VoterSubmit]*Commit, error) {
	var commitsByRound = map[ty.RoundId]map[ty.VoterSubmit]*Commit{}
	for _, msg := range messages {
		if msg.ProtocolID != ProtocolId {
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

func ExtractReveals(messages []payload.Message, getRoundEpoch func(ty.RoundId) *fsp.RewardEpoch) (map[ty.RoundId]map[ty.VoterSubmit]*Reveal, error) {
	var revealsByRound = map[ty.RoundId]map[ty.VoterSubmit]*Reveal{}
	for _, msg := range messages {
		if msg.ProtocolID != ProtocolId {
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

		feeds := getRoundEpoch(round).OrderedFeeds
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

func ExtractSignatures(messages []payload.Message) (map[ty.RoundId][]*SignatureSubmission, error) {
	var signaturesByRound = map[ty.RoundId][]*SignatureSubmission{}
	for _, msg := range messages {
		if msg.ProtocolID != ProtocolId {
			logger.Fatal("message protocol %d does not match expected protocol %d - likely an issue with submission retrieval")
		}

		round := ty.RoundId(msg.VotingRound)
		expectedRound := params.Net.Epoch.VotingRoundForTimeSec(msg.Timestamp) - 1
		if round != expectedRound {
			logger.Debug("Signature round %d does not match expected round %d, skipping", round, expectedRound)
			continue
		}

		var signature *fsp.RawSignature
		var err error

		if msg.Payload[0] == 1 {
			signature, err = fsp.DecodeSignatureType1(msg.Payload)
			if err != nil {
				logger.Debug("error parsing Signature, skipping: %signature", err)
				continue
			}
		} else {
			signature, err = fsp.DecodeSignatureType0(msg.Payload)
			if err != nil {
				logger.Debug("error parsing Signature, skipping: %signature", err)
				continue
			}
		}

		if _, ok := signaturesByRound[round]; !ok {
			signaturesByRound[round] = []*SignatureSubmission{}
		}

		signaturesByRound[round] = append(signaturesByRound[round], &SignatureSubmission{
			Signature: signature,
			Info: fsp.TxInfo{
				From:         msg.From,
				TimestampSec: msg.Timestamp,
				Reverted:     false,
			},
		})
	}

	return signaturesByRound, nil
}

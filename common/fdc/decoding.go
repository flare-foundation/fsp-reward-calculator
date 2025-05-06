package fdc

import (
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"github.com/pkg/errors"
	"math/big"
)

func ExtractBitVotes(messages []payload.Message) map[ty.RoundId]map[ty.VoterSubmit]*big.Int {
	var bitVotesByRound = map[ty.RoundId]map[ty.VoterSubmit]*big.Int{}
	for _, msg := range messages {
		if msg.ProtocolID != ProtocolId {
			logger.Fatal("message protocol %d does not match expected protocol %d - likely an issue with submission retrieval")
		}

		round := ty.RoundId(msg.VotingRound)
		submitRound := params.Net.Epoch.VotingRoundForTimeSec(msg.Timestamp)
		if round != submitRound.Add(-1) {
			logger.Debug("bitvote round %d does not match expected round %d, skipping", round, submitRound.Add(-1))
			continue
		}

		if msg.Timestamp > params.Net.Epoch.RevealDeadlineSec(submitRound) {
			logger.Debug("bitvote from %s too late", msg.From)
			continue
		}

		if _, ok := bitVotesByRound[round]; !ok {
			bitVotesByRound[round] = map[ty.VoterSubmit]*big.Int{}
		}

		from := ty.VoterSubmit(msg.From)
		bitVote, err := ParseBitVote(msg.Payload)
		if err != nil {
			logger.Debug("error parsing bitVote from %s, skipping: %s", msg.From.String(), err)
			continue
		}
		bitVotesByRound[round][from] = bitVote
	}

	return bitVotesByRound
}

func ParseBitVote(bytes []byte) (*big.Int, error) {
	if len(bytes) < 2 {
		return nil, errors.New("bitVote too short")
	}

	bitVector := new(big.Int).SetBytes(bytes[2:])

	// TODO: should we check the length of the bitVector? TS impl doesn't
	//lengthBytes := bytes[0:2]
	//length := binary.BigEndian.Uint16(lengthBytes)
	//if bitVector.BitLen() > int(length) {
	//	return nil, errors.New("bitvote length does not match bitvector")
	//}
	return bitVector, nil
}

func ExtractFdcSignatures(messages []payload.Message) map[ty.RoundId][]*SignatureSubmission {
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

		sigType := msg.Payload[0]
		if sigType == 1 {
			signature, err = fsp.DecodeSignatureType1(msg.Payload)
			if err != nil {
				logger.Debug("error parsing Signature, skipping: %s", err)
				continue
			}
		} else {
			signature, err = fsp.DecodeSignatureType0(msg.Payload)
			if err != nil {
				logger.Debug("error parsing Signature, skipping: %s", err)
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

	return signaturesByRound
}

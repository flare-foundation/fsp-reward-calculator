package data

import (
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
)

const (
	FdcProtocolId = 200
)

type FdcSignatureSubmission struct {
	Signature *SignatureType1
	Info      TxInfo
}

type BitVote struct {
	bytes []byte
}

func ExtractBitVotes(messages []payload.Message) map[ty.RoundId]map[ty.VoterSubmit]BitVote {
	var bitVotesByRound = map[ty.RoundId]map[ty.VoterSubmit]BitVote{}
	for _, msg := range messages {
		if msg.ProtocolID != FdcProtocolId {
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
			bitVotesByRound[round] = map[ty.VoterSubmit]BitVote{}
		}

		from := ty.VoterSubmit(msg.From)
		bitVotesByRound[round][from] = BitVote{msg.Payload}
	}

	return bitVotesByRound
}

var WrongSignatureIndicatorMessageHash = common.Hash{}

func GetFdcSignersByRound(msgs []payload.Message, roundHash map[ty.RoundId]common.Hash, re *RewardEpoch) SignerMap {
	allSignatures := extractFdcSignatures(msgs)

	signers := SignerMap{}
	for round, signatures := range allSignatures {
		sigsByHash := map[common.Hash]map[ty.VoterSigning]SigInfo{}
		for _, signatureSubmission := range signatures {

			sender := ty.VoterSubmitSignatures(signatureSubmission.Info.From)
			expectedSigner := re.VoterIndex.BySubmitSignatures[sender]

			if expectedSigner == nil {
				logger.Debug("sender %s not registered, skipping signature", sender)
				continue
			}

			signature := signatureSubmission.Signature
			signedHash := roundHash[round]
			signerKey, err := crypto.SigToPub(
				signedHash.Bytes(),
				append(signature.bytes[1:65], signature.bytes[0]-27),
			)
			if err != nil {
				logger.Debug("error recovering signerKey, skipping signature: %s", err)
				continue
			}

			signer := ty.VoterSigning(crypto.PubkeyToAddress(*signerKey))
			if signer != expectedSigner.Signing {
				signedHash = WrongSignatureIndicatorMessageHash
			}

			if _, ok := re.VoterIndex.BySigning[signer]; ok {
				if _, ok := sigsByHash[signedHash]; !ok {
					sigsByHash[signedHash] = map[ty.VoterSigning]SigInfo{}
				}
				sigsByHash[signedHash][signer] = SigInfo{
					Signer:    signer,
					Timestamp: signatureSubmission.Info.TimestampSec,
				}
			}
		}

		signers[round] = sigsByHash
	}
	return signers
}

func extractFdcSignatures(messages []payload.Message) map[ty.RoundId][]*FdcSignatureSubmission {
	var signaturesByRound = map[ty.RoundId][]*FdcSignatureSubmission{}
	for _, msg := range messages {
		if msg.ProtocolID != FdcProtocolId {
			logger.Fatal("message protocol %d does not match expected protocol %d - likely an issue with submission retrieval")
		}

		round := ty.RoundId(msg.VotingRound)
		expectedRound := params.Net.Epoch.VotingRoundForTimeSec(msg.Timestamp) - 1
		if round != expectedRound {
			logger.Debug("Signature round %d does not match expected round %d, skipping", round, expectedRound)
			continue
		}

		signature, err := DecodeSignatureType1(msg.Payload)
		if err != nil {
			logger.Debug("error parsing Signature, skipping: %s", err)
			continue
		}

		if _, ok := signaturesByRound[round]; !ok {
			signaturesByRound[round] = []*FdcSignatureSubmission{}
		}

		signaturesByRound[round] = append(signaturesByRound[round], &FdcSignatureSubmission{
			Signature: signature,
			Info: TxInfo{
				From:         msg.From,
				TimestampSec: msg.Timestamp,
				Reverted:     false,
			},
		})
	}

	return signaturesByRound
}

package data

import (
	"bytes"
	"encoding/binary"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fdchub"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
)

const (
	FdcProtocolId = 200
)

type FdcSignatureSubmission struct {
	Signature *SignatureType1
	Info      TxInfo
}

func ExtractBitVotes(messages []payload.Message) map[ty.RoundId]map[ty.VoterSubmit]*big.Int {
	var bitVotesByRound = map[ty.RoundId]map[ty.VoterSubmit]*big.Int{}
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
			bitVotesByRound[round] = map[ty.VoterSubmit]*big.Int{}
		}

		from := ty.VoterSubmit(msg.From)
		bitVote, err := ParseBitVote(msg.Payload)
		if err != nil {
			logger.Debug("error parsing bitVote from %s, skipping: %s", msg.From, err)
			continue
		}
		bitVotesByRound[round][from] = bitVote
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
				if _, ok := sigsByHash[signedHash][signer]; ok {
					logger.Debug("earlier signature from %s already added, skipping", signer)
					continue
				}

				sigsByHash[signedHash][signer] = SigInfo{
					Signer:          signer,
					Timestamp:       signatureSubmission.Info.TimestampSec,
					UnsignedMessage: signature.message,
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

		var signature *SignatureType1
		var err error

		sigType := msg.Payload[0]
		if sigType == 1 {
			signature, err = DecodeSignatureType1(msg.Payload)
			if err != nil {
				logger.Debug("error parsing Signature, skipping: %s", err)
				continue
			}
		} else {
			// FDC should not be using type 0 signatures, we should punish the submitter in the future
			// but for now we allow it so adding a temporary workaround
			signature0, err := DecodeSignatureType0(msg.Payload)
			if err != nil {
				logger.Debug("error parsing Signature, skipping: %s", err)
				continue
			}
			signature = &SignatureType1{
				bytes:   signature0.bytes,
				message: signature0.message,
			}
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

type AttestationRequest struct {
	Data      []byte
	MergedFee *big.Int // Combined fee for all duplicates
}

func GetAttestationRequestsByRound(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) map[ty.RoundId][]AttestationRequest {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound)
	toSec := params.Net.Epoch.VotingRoundStartSec(toRound.Add(1))

	instance, _ := fdchub.NewFdcHub(common.Address{}, nil)

	type eventWithTs struct {
		event     *fdchub.FdcHubAttestationRequest
		timestamp uint64
	}

	parse := func(log types.Log, ts uint64) (eventWithTs, error) {
		event, err := instance.FdcHubFilterer.ParseAttestationRequest(log)
		return eventWithTs{event, ts}, err
	}

	events, err := queryEvents(
		db,
		fromSec,
		toSec,
		params.Net.Contracts.FdcHub,
		utils.EventTopic0.FdcAttestationRequest,
		parse,
	)
	if err != nil {
		logger.Fatal("error fetching events: %s", err)
	}

	var eventsByRound = map[ty.RoundId][]*fdchub.FdcHubAttestationRequest{}

	for _, event := range events {
		round := params.Net.Epoch.VotingRoundForTime(event.timestamp * 1000)
		if round < fromRound || round > toRound {
			continue
		}
		if _, ok := eventsByRound[round]; !ok {
			eventsByRound[round] = []*fdchub.FdcHubAttestationRequest{}
		}
		eventsByRound[round] = append(eventsByRound[round], event.event)
	}

	return mergeDuplicates(fromRound, toRound, eventsByRound)
}

// TODO: This deduplication logic is simple to read but n^2, optimise when we have more requests
func mergeDuplicates(fromRound ty.RoundId, toRound ty.RoundId, eventsByRound map[ty.RoundId][]*fdchub.FdcHubAttestationRequest) map[ty.RoundId][]AttestationRequest {
	var requestsByRound = map[ty.RoundId][]AttestationRequest{}
	for round := fromRound; round <= toRound; round++ {
		events, ok := eventsByRound[round]
		if !ok {
			continue
		}

		var uniqueRequests []AttestationRequest
		for i := range events {
			merged := false
			for j := range uniqueRequests {
				if bytes.Equal(uniqueRequests[j].Data, events[i].Data) {
					uniqueRequests[j].MergedFee.Add(uniqueRequests[j].MergedFee, events[i].Fee)
					merged = true
					break
				}
			}
			if !merged {
				uniqueRequests = append(uniqueRequests, AttestationRequest{
					Data:      events[i].Data,
					MergedFee: new(big.Int).Set(events[i].Fee),
				})
			}
		}

		requestsByRound[round] = uniqueRequests
	}
	return requestsByRound
}

func ParseBitVote(bytes []byte) (*big.Int, error) {
	if len(bytes) < 2 {
		return nil, errors.New("bitVote too short")
	}

	lengthBytes := bytes[0:2]
	length := binary.BigEndian.Uint16(lengthBytes)
	bitVector := new(big.Int).SetBytes(bytes[2:])

	if bitVector.BitLen() > int(length) {
		return nil, errors.New("bitvote length does not match bitvector")
	}
	return bitVector, nil
}

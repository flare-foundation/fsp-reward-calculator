package fdc

import (
	common2 "fsp-rewards-calculator/common"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fdchub"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"gorm.io/gorm"
	"math/big"
)

var WrongSignatureIndicatorMessageHash = common.Hash{}

func GetSignersByRound(msgs []payload.Message, roundHash map[ty.RoundId]common.Hash, re *fsp.RewardEpoch) fsp.SignerMap {
	allSignatures := ExtractFdcSignatures(msgs)

	signers := fsp.SignerMap{}
	for round, signatures := range allSignatures {
		sigsByHash := map[common.Hash]map[ty.VoterSigning]fsp.SigInfo{}
		for _, signatureSubmission := range signatures {
			sender := ty.VoterSubmitSignatures(signatureSubmission.Info.From)
			voter := re.VoterIndex.BySubmitSignatures[sender]

			if voter == nil {
				logger.Debug("sender %s not registered, skipping signature", sender)
				continue
			}

			signature := signatureSubmission.Signature
			signedHash := roundHash[round]
			signerKey, err := crypto.SigToPub(
				signedHash.Bytes(),
				append(signature.Bytes[1:65], signature.Bytes[0]-27),
			)
			if err != nil {
				logger.Debug("error recovering signerKey, skipping signature: %s", err)
				continue
			}

			recoveredSigner := ty.VoterSigning(crypto.PubkeyToAddress(*signerKey))
			if recoveredSigner != voter.Signing {
				signedHash = WrongSignatureIndicatorMessageHash
			}

			if _, ok := sigsByHash[signedHash]; !ok {
				sigsByHash[signedHash] = map[ty.VoterSigning]fsp.SigInfo{}
			}
			if _, ok := sigsByHash[signedHash][voter.Signing]; ok {
				logger.Debug("earlier signature from %s already added, skipping", voter.Signing)
				continue
			}

			sigsByHash[signedHash][voter.Signing] = fsp.SigInfo{
				Signer:          voter.Signing,
				Timestamp:       signatureSubmission.Info.TimestampSec,
				UnsignedMessage: signature.Message,
			}
		}

		signers[round] = sigsByHash
	}
	return signers
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
		event, err := instance.ParseAttestationRequest(log)
		return eventWithTs{event, ts}, err
	}

	events, err := fsp.QueryEvents(
		db,
		fromSec,
		toSec,
		params.Net.Contracts.FdcHub,
		common2.EventTopic0.FdcAttestationRequest,
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

func mergeDuplicates(fromRound ty.RoundId, toRound ty.RoundId, eventsByRound map[ty.RoundId][]*fdchub.FdcHubAttestationRequest) map[ty.RoundId][]AttestationRequest {
	var requestsByRound = map[ty.RoundId][]AttestationRequest{}
	for round := fromRound; round <= toRound; round++ {
		events, ok := eventsByRound[round]
		if !ok {
			continue
		}

		var requestsByData = map[string]*AttestationRequest{}
		for _, event := range events {
			dataKey := string(event.Data)
			if req, exists := requestsByData[dataKey]; exists {
				req.MergedFee.Add(req.MergedFee, event.Fee)
			} else {
				requestsByData[dataKey] = &AttestationRequest{
					Data:      event.Data,
					MergedFee: new(big.Int).Set(event.Fee),
				}
			}
		}
		var uniqueRequests []AttestationRequest
		// Need to maintain original order
		for _, event := range events {
			dataKey := string(event.Data)
			if req, exists := requestsByData[dataKey]; exists {
				uniqueRequests = append(uniqueRequests, *req)
				delete(requestsByData, dataKey)
			}
		}

		requestsByRound[round] = uniqueRequests
	}
	return requestsByRound
}

package main

import (
	"bytes"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type SignerMap map[types.RoundId]map[common.Hash]map[VoterSigning]SigInfo

type SigInfo struct {
	Signer    VoterSigning
	Timestamp uint64
}

// getSigners fetches all signatures for all rounds in the reward epoch, and for each round
// computes the list of valid signatures by signed hash.
// For each signer, only the last signature for a specific round and hash is retained.
func getSignersByRound(db *gorm.DB, re RewardEpoch) (SignerMap, error) {
	logger.Info("Fetching signers for rounds %d-%d", re.StartRound, re.EndRound)
	allSignatures, err := getSignatures(db, re.StartRound, re.EndRound)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching signatures")
	}

	signers := SignerMap{}
	for round, signatures := range allSignatures {
		sigsByHash := map[common.Hash]map[VoterSigning]SigInfo{}
		for _, signatureSubmission := range signatures {
			signature := signatureSubmission.Signature
			signedHash := signature.merkleRoot.EncodedHash()
			signerKey, err := crypto.SigToPub(
				signedHash.Bytes(),
				append(signature.bytes[1:65], signature.bytes[0]-27),
			)
			if err != nil {
				logger.Debug("error recovering signerKey, skipping signature: %s", err)
				continue
			}

			signer := VoterSigning(crypto.PubkeyToAddress(*signerKey))
			if _, ok := re.VoterIndex.bySigning[signer]; ok {
				if _, ok := sigsByHash[signedHash]; !ok {
					sigsByHash[signedHash] = map[VoterSigning]SigInfo{}
				}
				sigsByHash[signedHash][signer] = SigInfo{
					Signer:    signer,
					Timestamp: signatureSubmission.Info.TimestampSec,
				}
			} else {
				logger.Debug("signer %s not registered, skipping signature", signer)
			}
		}

		signers[round] = sigsByHash
	}
	return signers, nil
}

func getFinalizationsByRound(db *gorm.DB, re RewardEpoch) (map[types.RoundId][]*Finalization, error) {
	logger.Info("Fetching finalizations for rounds %d-%d", re.StartRound, re.EndRound)
	allFinalizationsByRound, err := getFinalizations(db, re.StartRound, re.EndRound)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching finalizations")
	}

	logger.Info("Finalizations: %d", len(allFinalizationsByRound))

	finalizationsByRound := make(map[types.RoundId][]*Finalization)

	//var firstSuccessful *Finalization
	for round, finalizations := range allFinalizationsByRound {
		seenSender := map[common.Address]bool{}
		for _, finalization := range finalizations {
			if types.EpochId(finalization.Policy.RewardEpochId) != re.Epoch {
				logger.Info("finalization reward epoch %d does not match expected epoch %d, skipping", finalization.Policy.RewardEpochId, re.Epoch)
				continue
			}

			if !bytes.Equal(finalization.Policy.RawBytes, re.Policy.RawBytes) {
				logger.Info("finalization signing policy does not match expected, skipping")
				continue
			}

			if _, ok := seenSender[finalization.Info.From]; ok {
				logger.Info("finalization from %s already seen, skipping", finalization.Info.From)
				continue
			} else {
				seenSender[finalization.Info.From] = true
			}

			//if firstSuccessful == nil && !finalization.Info.Reverted {
			//	firstSuccessful = finalization
			//}
			// TODO: Store first successful

			finalizationsByRound[round] = append(finalizationsByRound[round], finalization)
		}
	}
	return finalizationsByRound, nil
}

type FUpdate struct {
	feeds      *FUpdateFeed
	submitters []VoterSigning
}

func getFUpdatesByRound(db *gorm.DB, re RewardEpoch) (map[types.RoundId]*FUpdate, error) {
	logger.Info("Fetching FastUpdates data for rounds %d-%d", re.StartRound, re.EndRound)

	feeds, err := getFUpdateFeeds(db, re.StartRound, re.EndRound)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching FUpdate feeds")
	}

	submitters, err := getFUpdateSubmits(db, re.StartRound, re.EndRound)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching FUpdate submitters")
	}

	byRound := make(map[types.RoundId]*FUpdate)

	for round := re.StartRound; round <= re.EndRound; round++ {
		byRound[round] = &FUpdate{
			feeds:      feeds[round],
			submitters: submitters[round],
		}

	}
	return byRound, nil
}

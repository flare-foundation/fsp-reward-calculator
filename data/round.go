package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
)

type SignerMap map[ty.RoundId]map[common.Hash]map[ty.VoterSigning]SigInfo

type SigInfo struct {
	Signer    ty.VoterSigning
	Timestamp uint64
}

// GetSignersByRound fetches signatures for the specified round range, and for each round
// computes the list of valid signatures by signed hash.
// For each signer, only the last signature for a specific round and hash is retained.
func GetSignersByRound(db *gorm.DB, from ty.RoundId, to ty.RoundId, re *RewardEpoch) (SignerMap, error) {
	logger.Info("Fetching signers for rounds %d-%d", from, to)
	allSignatures, err := getSignatures(db, from, to)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching signatures")
	}

	signers := SignerMap{}
	for round, signatures := range allSignatures {
		sigsByHash := map[common.Hash]map[ty.VoterSigning]SigInfo{}
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

			signer := ty.VoterSigning(crypto.PubkeyToAddress(*signerKey))
			if _, ok := re.VoterIndex.BySigning[signer]; ok {
				if _, ok := sigsByHash[signedHash]; !ok {
					sigsByHash[signedHash] = map[ty.VoterSigning]SigInfo{}
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

func GetFinalizationsByRound(db *gorm.DB, from ty.RoundId, to ty.RoundId, re *RewardEpoch) (map[ty.RoundId][]*Finalization, error) {
	logger.Info("Fetching finalizations for rounds %d-%d", from, to)
	allFinalizationsByRound, err := getFinalizations(db, from, to)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching finalizations")
	}

	logger.Info("Finalizations: %d", len(allFinalizationsByRound))

	finalizationsByRound := make(map[ty.RoundId][]*Finalization)

	for round, finalizations := range allFinalizationsByRound {
		seenSender := map[common.Address]bool{}
		for _, finalization := range finalizations {
			if ty.EpochId(finalization.Policy.RewardEpochId) != re.Epoch {
				logger.Info("Finalization reward epoch %d does not match expected epoch %d, skipping", finalization.Policy.RewardEpochId, re.Epoch)
				continue
			}

			if !bytes.Equal(finalization.Policy.RawBytes, re.Policy.RawBytes) {
				logger.Info("Finalization signing policy does not match expected, skipping")
				continue
			}

			if _, ok := seenSender[finalization.Info.From]; ok {
				logger.Info("Finalization from %s already seen, skipping", finalization.Info.From)
				continue
			} else {
				seenSender[finalization.Info.From] = true
			}

			finalizationsByRound[round] = append(finalizationsByRound[round], finalization)
		}
	}
	return finalizationsByRound, nil
}

type FUpdate struct {
	Feeds      *FUpdateFeed
	Submitters []ty.VoterSigning
}

func GetFUpdatesByRound(db *gorm.DB, from ty.RoundId, to ty.RoundId) (map[ty.RoundId]*FUpdate, error) {
	logger.Info("Fetching FastUpdates data for rounds %d-%d", from, to)

	feeds, err := getFUpdateFeeds(db, from, to)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching FUpdate feeds")
	}

	submitters, err := getFUpdateSubmits(db, from, to)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching FUpdate submitters")
	}

	byRound := make(map[ty.RoundId]*FUpdate)

	for round := from; round <= to; round++ {
		byRound[round] = &FUpdate{
			Feeds:      feeds[round],
			Submitters: submitters[round],
		}

	}
	return byRound, nil
}

type RoundReveals struct {
	Reveals             map[ty.VoterSubmit]*Reveal
	RegisteredOffenders []ty.VoterSubmit
	AllOffenders        []ty.VoterSubmit
}

type PrintReveals struct {
	Submitted string
	Random    string
}

type RoundPrintData struct {
	Reveals      []PrintReveals
	Offenders    []string
	AllOffenders []string

	Medians      []Result
	Random       string
	SecureRandom bool

	NextRandom string

	MedianFeed  string
	FeedDecoded string
}

func PrintRoundData(results RoundResult, reveals RoundReveals, feed *Feed, selection *big.Int, epoch ty.EpochId, round ty.RoundId) {
	var roundData RoundPrintData

	for voter, reveal := range reveals.Reveals {
		roundData.Reveals = append(roundData.Reveals, PrintReveals{
			Submitted: common.Address(voter).String(),
			Random:    reveal.Random.String(),
		})
	}

	for _, offender := range reveals.RegisteredOffenders {
		roundData.Offenders = append(roundData.Offenders, common.Address(offender).String())
	}
	for _, offender := range reveals.AllOffenders {
		roundData.AllOffenders = append(roundData.AllOffenders, common.Address(offender).String())
	}

	roundData.Random = results.Random.Value.String()
	roundData.SecureRandom = results.Random.IsSecure
	roundData.NextRandom = selection.String()

	if feed != nil {
		roundData.MedianFeed = feed.Id.Hex()
		roundData.FeedDecoded = feed.String()
	}

	for _, median := range results.Median {
		medianCpy := *median
		medianCpy.InputValues = nil
		roundData.Medians = append(roundData.Medians, medianCpy)
	}

	jsonData, err := json.MarshalIndent(roundData, "", "    ")
	if err != nil {
		logger.Error("Error serializing to JSON:", err)
		return
	}
	filePath := fmt.Sprintf("results/%s/%d/%d/data.json", params.Net.Name, epoch, round)
	utils.WriteToFile(jsonData, filePath)
}

func GetRoundReveals(db *gorm.DB, from ty.RoundId, to ty.RoundId, epochs RewardEpochs) map[ty.RoundId]RoundReveals {
	var (
		allCommitsByRound map[ty.RoundId]map[ty.VoterSubmit]*Commit
		allRevealsByRound map[ty.RoundId]map[ty.VoterSubmit]*Reveal
	)

	logger.Info("Fetching commits for rounds %d-%d", from, to)
	allCommitsByRound, err := GetCommits(db, from, to)
	logger.Info("All commits fetched")
	if err != nil {
		logger.Fatal("error fetching commitsByRound: %s", err)
	}

	logger.Info("Fetching reveals for rounds %d-%d", from, to)
	allRevealsByRound, err = GetReveals(db, from, to, epochs)
	logger.Info("All reveals fetched")
	if err != nil {
		logger.Fatal("error fetching revealsByRound: %s", err)
	}

	logger.Info("All commits and reveals fetched, processing.")

	return getRoundReveals(from, to, epochs, allCommitsByRound, allRevealsByRound)
}

func getRoundReveals(
	from ty.RoundId,
	to ty.RoundId,
	epochs RewardEpochs,
	allCommitsByRound map[ty.RoundId]map[ty.VoterSubmit]*Commit,
	allRevealsByRound map[ty.RoundId]map[ty.VoterSubmit]*Reveal,
) map[ty.RoundId]RoundReveals {
	roundData := map[ty.RoundId]RoundReveals{}

	for round := from; round < to; round++ {
		voterIndex := epochs.EpochForRound(round).VoterIndex

		validCommits := map[ty.VoterSubmit]*Commit{}
		for voter, commit := range allCommitsByRound[round] {
			if voterIndex.BySubmit[voter] != nil {
				validCommits[voter] = commit
			}
		}

		validReveals := map[ty.VoterSubmit]*Reveal{}
		for voter, reveal := range allRevealsByRound[round] {
			if voterIndex.BySubmit[voter] != nil {
				validReveals[voter] = reveal
			} else {
				logger.Info("Voter %s not found in voterIndex, skipping reveal", common.Address(voter))
			}
		}

		registeredOffenders, matchingReveals := getRevealsAndOffenders(validCommits, validReveals, round)
		allOffenders, _ := getRevealsAndOffenders(allCommitsByRound[round], allRevealsByRound[round], round)

		roundData[round] = RoundReveals{
			Reveals:             matchingReveals,
			RegisteredOffenders: registeredOffenders,
			AllOffenders:        allOffenders,
		}
	}

	return roundData
}

func getRevealsAndOffenders(commits map[ty.VoterSubmit]*Commit, reveals map[ty.VoterSubmit]*Reveal, round ty.RoundId) ([]ty.VoterSubmit, map[ty.VoterSubmit]*Reveal) {
	var offenders []ty.VoterSubmit
	matchingReveals := map[ty.VoterSubmit]*Reveal{}

	for voter, commit := range commits {
		reveal, ok := reveals[voter]
		if !ok {
			logger.Debug("voter %s committed but did not reveal", common.Address(voter))
			offenders = append(offenders, voter)
			continue
		}

		expected := utils.CommitHash(common.Address(voter), uint32(round), reveal.Random, reveal.EncodedValues)

		if expected.Cmp(commit.Hash) != 0 {
			logger.Debug("voter %s reveal hash did not match commit: %s != %s", common.Address(voter), expected.String(), commit.Hash.String())
			offenders = append(offenders, voter)
			continue
		}

		matchingReveals[voter] = reveal
	}
	return offenders, matchingReveals
}

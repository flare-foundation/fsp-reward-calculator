package data

import (
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type SignerMap map[ty.RoundId]map[common.Hash]map[ty.VoterSigning]SigInfo

type SigInfo struct {
	Signer          ty.VoterSigning
	Timestamp       uint64
	UnsignedMessage []byte // TOOD: Only use by FDC
}

// GetSignersByRound fetches signatures for the specified round range, and for each round
// computes the list of valid signatures by signed hash. TODO: update doc
// For each signer, only the last signature for a specific round and hash is retained.
func GetSignersByRound(msgs []payload.Message, roundHash map[ty.RoundId]common.Hash, re *RewardEpoch) SignerMap {
	allSignatures, err := extractSignatures(msgs)
	if err != nil {
		logger.Fatal("error extracting signatures: %s", err)
	}

	signers := SignerMap{}
	for round, signatures := range allSignatures {
		sigsByHash := map[common.Hash]map[ty.VoterSigning]SigInfo{}
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
				append(signature.bytes[1:65], signature.bytes[0]-27),
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
				sigsByHash[signedHash] = map[ty.VoterSigning]SigInfo{}
			}
			if _, ok := sigsByHash[signedHash][voter.Signing]; ok {
				logger.Debug("earlier signature from %s already added, skipping", voter.Signing)
				continue
			}

			sigsByHash[signedHash][voter.Signing] = SigInfo{
				Signer:          voter.Signing,
				Timestamp:       signatureSubmission.Info.TimestampSec,
				UnsignedMessage: signature.message,
			}
		}

		signers[round] = sigsByHash
	}
	return signers
}

func GetFinalizationsByRound(fnz []*Finalization) map[ty.RoundId][]*Finalization {
	finalizationsByRound := make(map[ty.RoundId][]*Finalization)

loop:
	for _, f := range fnz {
		round := f.MerkleRoot.round

		if _, ok := finalizationsByRound[round]; !ok {
			finalizationsByRound[round] = []*Finalization{}
		} else {
			for _, rf := range finalizationsByRound[round] {
				if rf.Info.From == f.Info.From {
					logger.Info("Finalization for round %d from %s already seen, skipping", round, f.Info.From)
					continue loop
				}
			}
		}
		finalizationsByRound[round] = append(finalizationsByRound[round], f)
	}

	return finalizationsByRound
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

func GetRoundReveals(commitsMsgs []payload.Message, revealMsgs []payload.Message, epochs RewardEpochs) map[ty.RoundId]RoundReveals {
	var (
		commitsByRound map[ty.RoundId]map[ty.VoterSubmit]*Commit
		revealsByRound map[ty.RoundId]map[ty.VoterSubmit]*Reveal
	)

	commitsByRound, err := extractCommits(commitsMsgs)
	if err != nil {
		logger.Fatal("error extracting commitsByRound: %s", err)
	}

	revealsByRound, err = extractReveals(revealMsgs, epochs)
	if err != nil {
		logger.Fatal("error extracting revealsByRound: %s", err)
	}

	if len(commitsByRound) != len(revealsByRound) {
		logger.Fatal("commitsByRound and revealsByRound have different lengths")
	}

	roundData := map[ty.RoundId]RoundReveals{}
	for round := range commitsByRound {
		voterIndex := epochs.EpochForRound(round).VoterIndex

		validCommits := map[ty.VoterSubmit]*Commit{}
		for voter, commit := range commitsByRound[round] {
			if voterIndex.BySubmit[voter] != nil {
				validCommits[voter] = commit
			}
		}

		validReveals := map[ty.VoterSubmit]*Reveal{}
		for voter, reveal := range revealsByRound[round] {
			if voterIndex.BySubmit[voter] != nil {
				validReveals[voter] = reveal
			} else {
				logger.Debug("Voter %s not found in voterIndex, skipping reveal", common.Address(voter))
			}
		}

		registeredOffenders, matchingReveals := getRevealsAndOffenders(validCommits, validReveals, round)
		allOffenders, _ := getRevealsAndOffenders(commitsByRound[round], revealsByRound[round], round)

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

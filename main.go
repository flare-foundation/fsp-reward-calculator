package main

import (
	"flare-common/database"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"gorm.io/gorm"
)

func main() {
	config := database.DBConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "flare_ftso_indexer",
		Username: "root",
		Password: "root",
	}

	db, err := database.Connect(&config)

	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}

	_, err = calculateRewards(db, 2690)
	if err != nil {
		logger.Fatal("Error calculating rewards: %s", err)
		return
	}
}

type RewardClaim struct {
}

func calculateRewards(db *gorm.DB, epoch uint64) (RewardClaim, error) {
	re, err := getRewardEpoch(epoch, db)
	if err != nil {
		return RewardClaim{}, errors.Errorf("err fetching reward epoch: %s", err)
	}

	commitsByRound, err := getCommits(db, re.StartRound, re.EndRound, re.Voters)
	if err != nil {
		return RewardClaim{}, errors.Errorf("error fetching commitsByRound: %s", err)
	}
	revealsByRound, err := getReveals(db, re.StartRound, re.EndRound, re.Voters, re.OrderedFeeds)
	if err != nil {
		return RewardClaim{}, errors.Errorf("error fetching revealsByRound: %s", err)
	}

	for round := re.StartRound; round < re.EndRound; round++ {
		commits := commitsByRound[round]
		reveals := revealsByRound[round]

		logger.Info("Round: %d, Commits: %d, Reveals: %d", round, len(commits), len(reveals))

		var offenders []VoterSubmit
		validReveals := map[VoterSubmit]Reveal{}

		for voter, commit := range commits {
			reveal, ok := reveals[voter]
			if !ok {
				logger.Debug("Voter %s committed but did not reveal", voter)
				offenders = append(offenders, voter)
				continue
			}

			expected, err := utils.CommitHash(common.Address(voter), uint32(round), reveal.Random, reveal.EncodedValues)
			if err != nil {
				return RewardClaim{}, errors.Errorf("error computing reveal hash: %s", err)
			}

			if expected.Cmp(commit.Hash) != 0 {
				logger.Debug("Voter %s reveal hash did not match commit: %s != %s", voter, expected.String(), commit.Hash.String())
				offenders = append(offenders, voter)
				continue
			}

			validReveals[voter] = reveal
		}

		logger.Info("Round: %d, Offenders: %d, Valid Reveals: %d", round, len(offenders), len(validReveals))

		_, err := calculateMedians(re, validReveals)
		if err != nil {
			return RewardClaim{}, err

		}
	}

	return RewardClaim{}, nil
}

func calculateMedians(re RewardEpoch, validReveals map[VoterSubmit]Reveal) ([]MedianResult, error) {
	var medianResults []MedianResult
	for feedIndex, feed := range re.OrderedFeeds {
		var weightedValues []WeightedValue

		for voterSubmit, reveal := range validReveals {
			feedValue := reveal.Values[feedIndex]
			voterId := re.Voters.submitToIdentity[voterSubmit]
			weight := re.Voters.cappedWeight[voterId]
			if feedValue.isEmpty || weight == nil {
				continue
			}
			weightedValues = append(weightedValues, WeightedValue{
				value:  feedValue.Value,
				weight: weight,
			})
		}

		median, err := CalculateFeedMedian(weightedValues)
		if err != nil {
			logger.Error("error calculating median for feed %s: %s", feed.String(), err)
			continue
		}

		medianResults = append(medianResults, median)

		logger.Info("Feed: %s, Median: %+v", feed.String(), median)
	}

	return medianResults, nil
}

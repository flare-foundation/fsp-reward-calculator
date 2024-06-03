package main

import (
	"flare-common/database"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
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

	_, err = calculateRewards(db, 2620)
	if err != nil {
		logger.Fatal("Error calculating rewards: %s", err)
		return
	}
}

var randomMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)

type ClaimType int

// TODO: Check correct
const (
	Direct ClaimType = iota
	Fee
	WNat
	Mirror
	CChain
)

type RewardClaim struct {
	Epoch       uint64
	Beneficiary common.Address
	Amount      *big.Int
	Type        ClaimType
}

type RandomResult struct {
	Round    uint64
	Random   *big.Int
	IsSecure bool
}

func calculateRewards(db *gorm.DB, epoch uint64) (RewardClaim, error) {
	re, err := getRewardEpoch(epoch, db)
	if err != nil {
		return RewardClaim{}, errors.Errorf("err fetching reward epoch: %s", err)
	}

	windowStart := re.StartRound - params.Coston.Ftso.RandomGenerationBenchingWindow
	windowEnd := re.EndRound

	commitsByRound, err := getCommits(db, windowStart, re.EndRound)
	if err != nil {
		return RewardClaim{}, errors.Errorf("error fetching commitsByRound: %s", err)
	}
	revealsByRound, err := getReveals(db, re.StartRound, re.EndRound)
	if err != nil {
		return RewardClaim{}, errors.Errorf("error fetching revealsByRound: %s", err)
	}

	offendersByRound := map[uint64][]VoterSubmit{}
	validRevealsByRound := map[uint64]map[VoterSubmit]*Reveal{}

	for round := windowStart; round < windowEnd; round++ {
		var offenders []VoterSubmit
		validReveals := map[VoterSubmit]*Reveal{}

		commits := commitsByRound[round]
		reveals := revealsByRound[round]

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

		offendersByRound[round] = offenders
		validRevealsByRound[round] = validReveals
	}

	for round := re.StartRound; round < re.EndRound; round++ {
		validReveals := validRevealsByRound[round]

		eligibleReveals := map[VoterSubmit]*Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := re.Voters.submitToId[voter]; ok {
				eligibleReveals[voter] = reveal
			}
		}

		// Median
		feedValues := map[VoterSubmit][]FeedValue{}
		for voter, reveal := range eligibleReveals {
			values, err := DecodeFeedValues(reveal.EncodedValues, re.OrderedFeeds)
			if err != nil {
				logger.Error("error decoding feed values for voter %s: %s", voter, err)
				continue
			}
			feedValues[voter] = values
		}
		median, err := calculateMedians(re, feedValues)
		if err != nil {
			return RewardClaim{}, err
		}
		logger.Info("Median: %+v", median)

		// Random calc
		benchingWindowOffenders := map[VoterSubmit]bool{}
		for i := round - params.Coston.Ftso.RandomGenerationBenchingWindow; i < round; i++ {
			for j := range offendersByRound[i] {
				benchingWindowOffenders[offendersByRound[i][j]] = true
			}
		}
		nonBenchedOffenders := 0
		for k := range offendersByRound[round] {
			currentOffender := offendersByRound[round][k]
			if _, ok := benchingWindowOffenders[currentOffender]; !ok {
				nonBenchedOffenders++
			}
		}
		validCount := 0
		random := big.NewInt(0)
		for voter, reveal := range eligibleReveals {
			if _, ok := benchingWindowOffenders[voter]; !ok {
				random.Add(random, new(big.Int).SetBytes(reveal.Random[:]))
				validCount++
			}
		}
		random.Mod(random, randomMod)

		res := RandomResult{
			Round:    round,
			Random:   random,
			IsSecure: nonBenchedOffenders == 0 && validCount >= params.Coston.Ftso.NonBenchedRandomVotersMinCount,
		}
		logger.Info("Random result: %+v", res)

	}

	return RewardClaim{}, nil
}

func calculateMedians(re RewardEpoch, validReveals map[VoterSubmit][]FeedValue) ([]MedianResult, error) {
	var medianResults []MedianResult
	for feedIndex, feed := range re.OrderedFeeds {
		var weightedValues []WeightedValue

		for voterSubmit, values := range validReveals {
			feedValue := values[feedIndex]
			voterId := re.Voters.submitToId[voterSubmit]
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

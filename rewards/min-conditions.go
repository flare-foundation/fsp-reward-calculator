package rewards

import (
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"math/big"
)

var FtsoScalingClosenessThresholdPpm = big.NewInt(5000)      // 0.5%
var FtsoScalingAvailabilityThresholdPpm = big.NewInt(800000) // 80%

func metFtsoCondition(voterIndex *data.VoterIndex, totalFeeds int, results map[ty.RoundId]data.RoundResult) map[ty.VoterId]bool {
	voterHits := map[ty.VoterId]int{}

	for _, result := range results {
		for _, feedResult := range result.Median {
			for _, voterValue := range feedResult.InputValues {
				median := big.NewInt(int64(feedResult.Median))
				delta := new(big.Int).Mul(median, FtsoScalingClosenessThresholdPpm)
				delta.Div(delta, bigTotalPPM)

				lowerBound := new(big.Int).Sub(median, delta)
				upperBound := new(big.Int).Add(median, delta)

				value := big.NewInt(int64(voterValue.Value))

				if value.Cmp(lowerBound) >= 0 && value.Cmp(upperBound) <= 0 {
					voterId := voterIndex.BySubmit[voterValue.Voter].Identity
					voterHits[voterId]++
				}
			}
		}
	}

	rounds := len(results)
	availableHits := rounds * totalFeeds

	threshold := new(big.Int).Div(
		new(big.Int).Mul(
			FtsoScalingAvailabilityThresholdPpm,
			big.NewInt(int64(availableHits))),
		bigTotalPPM,
	)

	metCondition := map[ty.VoterId]bool{}
	for voter, hits := range voterHits {
		bigHits := big.NewInt(int64(hits))
		if bigHits.Cmp(threshold) >= 0 {
			metCondition[voter] = true
		}
		logger.Info("Voter %s hits: %d, met condition: %b", voter.String(), hits, metCondition[voter])
	}

	return metCondition
}

var FuThresholdPpm = big.NewInt(800000)            // 80%
var FuConsiderationThresholdPpm = big.NewInt(2000) // 0.2%

func metFUCondition(index *data.VoterIndex, updates map[ty.RoundId]*data.FUpdate) map[ty.VoterId]bool {
	voterUpdates := map[ty.VoterSigning]int{}
	totalUpdates := 0
	for _, update := range updates {
		for _, submitter := range update.Submitters {
			voterUpdates[submitter]++
		}
		totalUpdates += len(update.Submitters)
	}

	for _, voter := range index.PolicyOrder {
		logger.Info("Voter %s submits: %d", voter.Identity.String(), voterUpdates[voter.Signing])
	}

	metCondition := map[ty.VoterId]bool{}
	totalSigningWeight := big.NewInt(int64(index.TotalSigningPolicyWeight))
	for _, voter := range index.PolicyOrder {
		signingWeight := big.NewInt(int64(voter.SigningPolicyWeight))
		relativeWeightPpm := new(big.Int).Div(
			new(big.Int).Mul(signingWeight, bigTotalPPM),
			totalSigningWeight,
		)
		expectedUpdatesPpm := new(big.Int).Div(
			new(big.Int).Mul(signingWeight, FuThresholdPpm),
			totalSigningWeight,
		)
		expectedUpdates := new(big.Int).Div(
			new(big.Int).Mul(expectedUpdatesPpm, big.NewInt(int64(totalUpdates))),
			bigTotalPPM,
		)

		if relativeWeightPpm.Cmp(FuConsiderationThresholdPpm) < 0 {
			metCondition[voter.Identity] = true
			continue
		}

		if big.NewInt(int64(voterUpdates[voter.Signing])).Cmp(expectedUpdates) >= 0 {
			metCondition[voter.Identity] = true
		}

		logger.Info("Voter %s expected updates: %d, actual updates: %d", voter.Identity.String(), expectedUpdates, voterUpdates[voter.Signing])
	}

	logger.Info("Total submitters: %d", totalUpdates)
	return metCondition
}

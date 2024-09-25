package data

import (
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
)

type RoundResult struct {
	Round  ty.RoundId
	Median map[FeedId]*Result
	Random RandomResult
}

func CalculateResults(
	re RewardEpoch,
	reveals map[ty.RoundId]RoundReveals,
) (map[ty.RoundId]RoundResult, error) {
	var results = map[ty.RoundId]RoundResult{}

	for round := re.StartRound; round <= re.EndRound; round++ {
		validReveals := reveals[round].Reveals

		logger.Info("Reveals for round %d: %d", round, len(validReveals))

		eligibleReveals := map[ty.VoterSubmit]*Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := re.VoterIndex.BySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
			}
		}

		logger.Info("Eligible reveals for round %d: %d", round, len(eligibleReveals))

		// Median
		feedValues := map[ty.VoterSubmit][]FeedValue{}
		for voter, reveal := range eligibleReveals {
			values, err := DecodeFeedValues(reveal.EncodedValues, re.OrderedFeeds)
			if err != nil {
				logger.Error("error decoding feed values for voter %s: %s", voter, err)
				continue
			}
			feedValues[voter] = values
		}

		logger.Info("Calculating median for round %d", round)

		median, err := calculateMedians(re, feedValues)
		if err != nil {
			return nil, err
		}

		random := CalculateRandom(round, reveals, eligibleReveals)
		logger.Info("Round %d, random result: %d", round, random.Value)

		results[round] = RoundResult{
			Round:  round,
			Median: median,
			Random: random,
		}
	}
	return results, nil
}

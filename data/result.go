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
	from ty.RoundId,
	to ty.RoundId,
	re *RewardEpoch,
	reveals map[ty.RoundId]RoundReveals,
) (map[ty.RoundId]RoundResult, error) {
	var results = map[ty.RoundId]RoundResult{}

	for round := from; round <= to; round++ {
		validReveals := reveals[round].Reveals

		logger.Debug("Reveals for round %d: %d", round, len(validReveals))

		eligibleReveals := map[ty.VoterSubmit]*Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := re.VoterIndex.BySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
			}
		}

		logger.Debug("Eligible reveals for round %d: %d", round, len(eligibleReveals))

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

		logger.Debug("Calculating median for round %d", round)

		median, err := calculateMedians(re, feedValues)
		if err != nil {
			return nil, err
		}

		random := CalculateRandom(round, reveals, eligibleReveals)
		logger.Debug("Round %d, random result: %d", round, random.Value)

		results[round] = RoundResult{
			Round:  round,
			Median: median,
			Random: random,
		}
	}
	return results, nil
}

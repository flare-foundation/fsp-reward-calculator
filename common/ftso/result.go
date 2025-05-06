package ftso

import (
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
)

type RoundResult struct {
	Round  ty.RoundId
	Median map[fsp.FeedId]*Result
	Random RandomResult
}

func CalculateResults(
	from ty.RoundId,
	to ty.RoundId,
	feeds []fsp.Feed,
	voterIndex *fsp.VoterIndex,
	reveals map[ty.RoundId]RoundReveals,
) (map[ty.RoundId]RoundResult, error) {
	var results = map[ty.RoundId]RoundResult{}

	for round := from; round <= to; round++ {
		validReveals := reveals[round].Reveals

		logger.Debug("Reveals for round %d: %d", round, len(validReveals))

		eligibleReveals := map[ty.VoterSubmit]*Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := voterIndex.BySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
			} else {
				logger.Fatal("Voter %s not found in voterIndex, skipping reveal - should have been prefiletered before", voter)
			}
		}

		logger.Debug("Eligible reveals for round %d: %d", round, len(eligibleReveals))

		// Median
		feedValues := map[ty.VoterSubmit][]FeedValue{}
		for voter, reveal := range eligibleReveals {
			values, err := DecodeFeedValues(reveal.EncodedValues, feeds)
			if err != nil {
				logger.Error("error decoding feed values for voter %s: %s", voter, err)
				continue
			}
			feedValues[voter] = values
		}

		logger.Debug("Calculating median for round %d", round)

		median, err := calculateMedians(feeds, voterIndex, feedValues)
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

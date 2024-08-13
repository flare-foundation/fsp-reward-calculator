package main

import (
	"encoding/hex"
	"fmt"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/types"
	"github.com/pkg/errors"
	"math/big"
	"sort"
)

type MedianResult struct {
	Q1                int32
	Median            int32
	Q3                int32
	ParticipantWeight *big.Int
	inputValues       []VoterValue
}

type VoterValue struct {
	voter  VoterSubmit
	value  int32
	weight *big.Int
}

func (v VoterValue) String() string {
	return fmt.Sprintf("VoterValue{voter: %s, value: %d, weight: %s}", hex.EncodeToString(v.voter[:]), v.value, v.weight.String())
}

type nullInt32 struct {
	value int32
}

func CalculateMedians(round types.RoundId, re RewardEpoch, validReveals map[VoterSubmit][]FeedValue) (map[FeedId]*MedianResult, error) {
	medianResults := map[FeedId]*MedianResult{}
	for feedIndex, feed := range re.OrderedFeeds {
		var weightedValues []VoterValue

		for voterSubmit, values := range validReveals {
			feedValue := values[feedIndex]
			weight := re.VoterIndex.bySubmit[voterSubmit].CappedWeight
			if feedValue.isEmpty || weight == nil {
				continue
			}
			weightedValues = append(weightedValues, VoterValue{
				voter:  voterSubmit,
				value:  feedValue.Value,
				weight: weight,
			})
		}

		//logger.Info("Calculating median for round %d feed %s, valid values: %d", round, feed.Id.Hex(), len(weightedValues))

		median, err := CalculateFeedMedian(weightedValues)
		if err != nil {
			logger.Error("error calculating median for feed %s: %s", feed.String(), err)
			continue
		}

		//logger.Info("Calculated median for round %s feed %s, %s: result %+v", round, feed.String(), hex.EncodeToString(feed.Id[:]), median)

		medianResults[feed.Id] = median

		//logger.Info("Feed: %s, Median: %+v", feed.String(), median)
	}

	return medianResults, nil
}

func CalculateFeedMedian(voterValues []VoterValue) (*MedianResult, error) {
	if len(voterValues) < 1 {
		return nil, nil
	}

	//logger.Info("Calculating median for %d values: %+v", len(voterValues), voterValues)

	sort.Slice(voterValues, func(i, j int) bool {
		return voterValues[i].value < voterValues[j].value
	})

	totalWeight := big.NewInt(0)
	for _, vw := range voterValues {
		totalWeight.Add(totalWeight, vw.weight)
	}

	q1Weight := new(big.Int).Div(totalWeight, big.NewInt(4))
	medianWeight, medianMod := new(big.Int).DivMod(totalWeight, big.NewInt(2), new(big.Int))
	q3Weight := new(big.Int).Sub(totalWeight, q1Weight)

	var q1, median, q3 *nullInt32
	accumulatedWeight := big.NewInt(0)

	i := 0
	for ; i < len(voterValues); i++ {
		wv := voterValues[i]
		accumulatedWeight.Add(accumulatedWeight, wv.weight)

		if q1 == nil && accumulatedWeight.Cmp(q1Weight) > 0 {
			q1 = &nullInt32{wv.value}
		}
		if median == nil && accumulatedWeight.Cmp(medianWeight) >= 0 {
			if accumulatedWeight.Cmp(medianWeight) == 0 && medianMod.Cmp(big.NewInt(0)) == 0 {
				median = &nullInt32{(wv.value + voterValues[i+1].value) / 2}
			} else {
				median = &nullInt32{wv.value}
			}
		}
		if accumulatedWeight.Cmp(q3Weight) > 0 {
			break
		}
	}

	q3 = &nullInt32{voterValues[i].value}

	if q1 == nil || median == nil {
		return nil, errors.New("could not calculate quartiles")
	}

	return &MedianResult{
		Q1:                q1.value,
		Median:            median.value,
		Q3:                q3.value,
		ParticipantWeight: totalWeight,
		inputValues:       voterValues,
	}, nil
}

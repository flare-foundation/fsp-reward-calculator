package data

import (
	"encoding/hex"
	"fmt"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"github.com/pkg/errors"
	"math/big"
	"sort"
)

type Result struct {
	Q1                int32
	Median            int32
	Q3                int32
	ParticipantWeight *big.Int
	InputValues       []VoterValue
}

type VoterValue struct {
	Voter  ty.VoterSubmit
	Value  int32
	Weight *big.Int
}

func (v VoterValue) String() string {
	return fmt.Sprintf("VoterValue{Voter: %s, Value: %d, Weight: %s}", hex.EncodeToString(v.Voter[:]), v.Value, v.Weight.String())
}

type nullInt32 struct {
	value int32
}

func calculateMedians(re RewardEpoch, validReveals map[ty.VoterSubmit][]FeedValue) (map[FeedId]*Result, error) {
	medianResults := map[FeedId]*Result{}
	for feedIndex, feed := range re.OrderedFeeds {
		var weightedValues []VoterValue

		for voterSubmit, values := range validReveals {
			feedValue := values[feedIndex]
			weight := re.VoterIndex.BySubmit[voterSubmit].CappedWeight
			if feedValue.IsEmpty || weight == nil {
				continue
			}
			weightedValues = append(weightedValues, VoterValue{
				Voter:  voterSubmit,
				Value:  feedValue.Value,
				Weight: weight,
			})
		}

		median, err := calculateFeedMedian(weightedValues)
		if err != nil {
			logger.Error("error calculating median for feed %s: %s", feed.String(), err)
			continue
		}

		medianResults[feed.Id] = median
	}

	return medianResults, nil
}

func calculateFeedMedian(voterValues []VoterValue) (*Result, error) {
	if len(voterValues) < 1 {
		return nil, nil
	}

	sort.Slice(voterValues, func(i, j int) bool {
		return voterValues[i].Value < voterValues[j].Value
	})

	totalWeight := big.NewInt(0)
	for _, vw := range voterValues {
		totalWeight.Add(totalWeight, vw.Weight)
	}

	q1Weight := new(big.Int).Div(totalWeight, big.NewInt(4))
	medianWeight, medianMod := new(big.Int).DivMod(totalWeight, big.NewInt(2), new(big.Int))
	q3Weight := new(big.Int).Sub(totalWeight, q1Weight)

	var q1, median, q3 *nullInt32
	accumulatedWeight := big.NewInt(0)

	i := 0
	for ; i < len(voterValues); i++ {
		wv := voterValues[i]
		accumulatedWeight.Add(accumulatedWeight, wv.Weight)

		if q1 == nil && accumulatedWeight.Cmp(q1Weight) > 0 {
			q1 = &nullInt32{wv.Value}
		}
		if median == nil && accumulatedWeight.Cmp(medianWeight) >= 0 {
			if accumulatedWeight.Cmp(medianWeight) == 0 && medianMod.Cmp(big.NewInt(0)) == 0 {
				median = &nullInt32{(wv.Value + voterValues[i+1].Value) / 2}
			} else {
				median = &nullInt32{wv.Value}
			}
		}
		if accumulatedWeight.Cmp(q3Weight) > 0 {
			break
		}
	}

	q3 = &nullInt32{voterValues[i].Value}

	if q1 == nil || median == nil {
		return nil, errors.New("could not calculate quartiles")
	}

	return &Result{
		Q1:                q1.value,
		Median:            median.value,
		Q3:                q3.value,
		ParticipantWeight: totalWeight,
		InputValues:       voterValues,
	}, nil
}

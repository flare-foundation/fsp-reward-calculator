package main

import (
	"github.com/pkg/errors"
	"math/big"
	"sort"
)

type MedianResult struct {
	Q1          int32
	Median      int32
	Q3          int32
	TotalWeight *big.Int
}

type weightedValue struct {
	value  int32
	weight *big.Int
}

type nullInt32 struct {
	value int32
}

func CalculateFeedMedian(values []FeedValue, voterWeights []*big.Int) (MedianResult, error) {
	if len(values) != len(voterWeights) || len(values) == 0 {
		return MedianResult{}, errors.New("values and weights must have the same length and be non-empty")
	}

	var weightedValues []weightedValue
	for i := 0; i < len(values); i++ {
		if values[i].isEmpty {
			continue
		}
		weightedValues = append(weightedValues, weightedValue{
			value:  values[i].Value,
			weight: voterWeights[i],
		})
	}

	return calculateMedian(weightedValues)
}

func calculateMedian(weightedValues []weightedValue) (MedianResult, error) {
	sort.Slice(weightedValues, func(i, j int) bool {
		return weightedValues[i].value < weightedValues[j].value
	})

	totalWeight := big.NewInt(0)
	for _, vw := range weightedValues {
		totalWeight.Add(totalWeight, vw.weight)
	}

	q1Weight := new(big.Int).Div(totalWeight, big.NewInt(4))
	medianWeight, medianMod := new(big.Int).DivMod(totalWeight, big.NewInt(2), new(big.Int))
	q3Weight := new(big.Int).Sub(totalWeight, q1Weight)

	var q1, median, q3 *nullInt32
	accumulatedWeight := big.NewInt(0)

	i := 0
	for ; i < len(weightedValues); i++ {
		wv := weightedValues[i]
		accumulatedWeight.Add(accumulatedWeight, wv.weight)

		if q1 == nil && accumulatedWeight.Cmp(q1Weight) > 0 {
			q1 = &nullInt32{wv.value}
		}
		if median == nil && accumulatedWeight.Cmp(medianWeight) >= 0 {
			if accumulatedWeight.Cmp(medianWeight) == 0 && medianMod.Cmp(big.NewInt(0)) == 0 {
				median = &nullInt32{(wv.value + weightedValues[i+1].value) / 2}
			} else {
				median = &nullInt32{wv.value}
			}
		}
		if accumulatedWeight.Cmp(q3Weight) > 0 {
			break
		}
	}
	q3 = &nullInt32{weightedValues[i-1].value}

	if q1 == nil || median == nil {
		return MedianResult{}, errors.New("could not calculate quartiles")
	}

	return MedianResult{
		Q1:          q1.value,
		Median:      median.value,
		Q3:          q3.value,
		TotalWeight: totalWeight,
	}, nil
}

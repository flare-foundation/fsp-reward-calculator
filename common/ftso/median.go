package ftso

import (
	"encoding/hex"
	"fmt"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"math/big"
	"sort"

	"github.com/pkg/errors"
)

type Quartiles struct {
	Q1                int32
	Median            int32
	Q3                int32
	ParticipantWeight *big.Int
}

type Result struct {
	Quartiles   *Quartiles
	InputValues []VoterValue
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

func calculateMedians(feeds []fsp.Feed, voterIndex *fsp.VoterIndex, validReveals map[ty.VoterSubmit][]FeedValue) (map[fsp.FeedId]*Result, error) {
	medianResults := map[fsp.FeedId]*Result{}
	for feedIndex, feed := range feeds {
		var weightedValues []VoterValue

		for voterSubmit, values := range validReveals {
			feedValue := values[feedIndex]
			weight := voterIndex.BySubmit[voterSubmit].CappedWeight
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

		medianResults[feed.Id] = &Result{
			Quartiles:   median,
			InputValues: weightedValues,
		}
	}

	return medianResults, nil
}

func calculateFeedMedian(voterValues []VoterValue) (*Quartiles, error) {
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

	// Quartile boundary weight is floor(T/4); the median threshold is ceil(T/2). This matches the TS reference
	// (ftso-scaling: libs/ftso-core/src/ftso-calculation/ftso-median.ts), which is what the on-chain/consensus
	// median follows. Using floor(T/2) or a forward Q3 scan diverges on exact-boundary cumulative weights.
	quartileWeight := new(big.Int).Div(totalWeight, big.NewInt(4))
	medianQuotient, medianMod := new(big.Int).DivMod(totalWeight, big.NewInt(2), new(big.Int))
	medianWeight := new(big.Int).Add(medianQuotient, medianMod) // ceil(T/2)

	var q1, median, q3 *nullInt32
	accumulatedWeight := big.NewInt(0)

	for i := 0; i < len(voterValues); i++ {
		wv := voterValues[i]
		accumulatedWeight.Add(accumulatedWeight, wv.Weight)

		if q1 == nil && accumulatedWeight.Cmp(quartileWeight) > 0 {
			q1 = &nullInt32{wv.Value}
		}
		if median == nil && accumulatedWeight.Cmp(medianWeight) >= 0 {
			if accumulatedWeight.Cmp(medianWeight) == 0 && medianMod.Sign() == 0 {
				// Even total weight: average the two middle values, rounding toward negative infinity
				// (JS Math.floor) and widening to int64 to avoid int32 overflow.
				sum := int64(wv.Value) + int64(voterValues[i+1].Value)
				avg := sum / 2
				if sum%2 != 0 && sum < 0 {
					avg--
				}
				median = &nullInt32{int32(avg)}
			} else {
				median = &nullInt32{wv.Value}
			}
		}
		if q1 != nil && median != nil {
			break
		}
	}

	// Third quartile: highest value whose cumulative weight from the top exceeds floor(T/4), matching the TS
	// reference's backward scan (a forward scan diverges when a prefix lands exactly on T - floor(T/4)).
	suffixWeight := big.NewInt(0)
	for j := len(voterValues) - 1; j >= 0; j-- {
		suffixWeight.Add(suffixWeight, voterValues[j].Weight)
		if suffixWeight.Cmp(quartileWeight) > 0 {
			q3 = &nullInt32{voterValues[j].Value}
			break
		}
	}

	if q1 == nil || median == nil || q3 == nil {
		return nil, errors.New("could not calculate quartiles")
	}

	return &Quartiles{
		Q1:                q1.value,
		Median:            median.value,
		Q3:                q3.value,
		ParticipantWeight: totalWeight,
	}, nil
}

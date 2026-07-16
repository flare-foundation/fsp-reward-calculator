package ftso

import (
	"math/big"
	"testing"
)

func voterValues(values []int32, weights []int64) []VoterValue {
	out := make([]VoterValue, len(values))
	for i := range values {
		out[i] = VoterValue{Value: values[i], Weight: big.NewInt(weights[i])}
	}
	return out
}

// Odd total weight: the median threshold is ceil(T/2), matching the TS reference. With weights [1,1,1] (T=3)
// the middle value must be selected, not the lower one (which floor(T/2) would pick).
func TestCalculateFeedMedian_OddBoundaryUsesCeil(t *testing.T) {
	q, err := calculateFeedMedian(voterValues([]int32{10, 20, 30}, []int64{1, 1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if q.Median != 20 {
		t.Errorf("median = %d, want 20 (ceil(T/2) threshold)", q.Median)
	}
	if q.Q1 != 10 {
		t.Errorf("q1 = %d, want 10", q.Q1)
	}
	if q.Q3 != 30 {
		t.Errorf("q3 = %d, want 30", q.Q3)
	}
}

// Q3 is the highest value whose cumulative weight from the top exceeds floor(T/4) (backward scan). With
// weights [1,1,1,1] (T=4, floor(T/4)=1) the third quartile is 30; a forward scan would wrongly pick 40.
func TestCalculateFeedMedian_Q3BackwardBoundary(t *testing.T) {
	q, err := calculateFeedMedian(voterValues([]int32{10, 20, 30, 40}, []int64{1, 1, 1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if q.Q3 != 30 {
		t.Errorf("q3 = %d, want 30 (backward scan)", q.Q3)
	}
	if q.Median != 25 {
		t.Errorf("median = %d, want 25 (even-count average of 20 and 30)", q.Median)
	}
	if q.Q1 != 20 {
		t.Errorf("q1 = %d, want 20", q.Q1)
	}
}

// Even-count average rounds toward negative infinity (JS Math.floor), not toward zero.
func TestCalculateFeedMedian_EvenAverageFloorsNegative(t *testing.T) {
	q, err := calculateFeedMedian(voterValues([]int32{-3, -2}, []int64{1, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if q.Median != -3 {
		t.Errorf("median = %d, want -3 (floor of -2.5)", q.Median)
	}
}

// Weighted majority: a heavy middle voter sets the median regardless of the exact threshold.
func TestCalculateFeedMedian_WeightedMajority(t *testing.T) {
	q, err := calculateFeedMedian(voterValues([]int32{100, 200, 300}, []int64{1, 5, 1}))
	if err != nil {
		t.Fatal(err)
	}
	if q.Median != 200 {
		t.Errorf("median = %d, want 200", q.Median)
	}
}

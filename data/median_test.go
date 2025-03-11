package data

import (
	"math/big"
	"reflect"
	"testing"
)

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name    string
		arg     []VoterValue
		want    Result
		wantErr bool
	}{
		{
			name: "Same Weight",
			arg: []VoterValue{
				{Value: 1, Weight: big.NewInt(1)},
				{Value: 2, Weight: big.NewInt(1)},
				{Value: 3, Weight: big.NewInt(1)},
				{Value: 4, Weight: big.NewInt(1)},
				{Value: 5, Weight: big.NewInt(1)},
				{Value: 6, Weight: big.NewInt(1)},
			},
			want: Result{
				Q1:                2,
				Median:            3,
				Q3:                5,
				ParticipantWeight: big.NewInt(6),
			},
			wantErr: false,
		},
		{
			name: "Two values",
			arg: []VoterValue{
				{Value: 1, Weight: big.NewInt(1)},
				{Value: 2, Weight: big.NewInt(1)},
			},

			want: Result{
				Q1:                1,
				Median:            1,
				Q3:                2,
				ParticipantWeight: big.NewInt(2),
			},
			wantErr: false,
		},
		{
			name: "Different weights",
			arg: []VoterValue{
				{Value: 1, Weight: big.NewInt(10)},
				{Value: 2, Weight: big.NewInt(1)},
				{Value: 3, Weight: big.NewInt(1)},
				{Value: 4, Weight: big.NewInt(1)},
				{Value: 5, Weight: big.NewInt(1)},
			},
			want: Result{
				Q1:                1,
				Median:            1,
				Q3:                2,
				ParticipantWeight: big.NewInt(14),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateFeedMedian(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateFeedMedian() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateFeedMedian() got = %v, want %v", got, tt.want)
			}
		})
	}
}

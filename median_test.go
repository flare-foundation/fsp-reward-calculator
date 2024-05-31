package main

import (
	"math/big"
	"reflect"
	"testing"
)

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name    string
		arg     []weightedValue
		want    MedianResult
		wantErr bool
	}{
		{
			name: "Same weight",
			arg: []weightedValue{
				{value: 1, weight: big.NewInt(1)},
				{value: 2, weight: big.NewInt(1)},
				{value: 3, weight: big.NewInt(1)},
				{value: 4, weight: big.NewInt(1)},
				{value: 5, weight: big.NewInt(1)},
				{value: 6, weight: big.NewInt(1)},
			},
			want: MedianResult{
				Q1:          2,
				Median:      3,
				Q3:          5,
				TotalWeight: big.NewInt(6),
			},
			wantErr: false,
		},
		{
			name: "Single value",
			arg: []weightedValue{
				{value: 1, weight: big.NewInt(1)},
			},

			want: MedianResult{
				Q1:          1,
				Median:      1,
				Q3:          1,
				TotalWeight: big.NewInt(1),
			},
			wantErr: false,
		},
		{
			name: "Two values",
			arg: []weightedValue{
				{value: 1, weight: big.NewInt(1)},
				{value: 2, weight: big.NewInt(1)},
			},

			want: MedianResult{
				Q1:          1,
				Median:      1,
				Q3:          2,
				TotalWeight: big.NewInt(2),
			},
			wantErr: false,
		},
		{
			name: "Different weights",
			arg: []weightedValue{
				{value: 1, weight: big.NewInt(10)},
				{value: 2, weight: big.NewInt(1)},
				{value: 3, weight: big.NewInt(1)},
				{value: 4, weight: big.NewInt(1)},
				{value: 5, weight: big.NewInt(1)},
			},
			want: MedianResult{
				Q1:          1,
				Median:      1,
				Q3:          2,
				TotalWeight: big.NewInt(14),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateMedian(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateFeedMedian() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateFeedMedian() got = %v, want %v", got, tt.want)
			}
		})
	}
}

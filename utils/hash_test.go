package utils

import (
	"ftsov2-rewarding/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"reflect"
	"testing"
)

func TestCommitHash(t *testing.T) {
	type args struct {
		voter      common.Address
		round      uint32
		random     common.Hash
		feedValues []byte
	}
	tests := []struct {
		name string
		args args
		want common.Hash
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CommitHash(tt.args.voter, tt.args.round, tt.args.random, tt.args.feedValues); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CommitHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeedSelectionRandom(t *testing.T) {
	type args struct {
		random *big.Int
		round  types.RoundId
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FeedSelectionRandom(tt.args.random, tt.args.round); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FeedSelectionRandom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinalizerSelectionSeed(t *testing.T) {
	t.Run("encoding matches reference Typescript implementation", func(t *testing.T) {
		seed := common.HexToHash("0x561c1876599d7b1583693711da55508f038908e7b91b6dc893b099a2eeb024bd").Big()
		expected := "0x9a22e13bd742533ab33080e2b3f96d1571bec59267ef1b721cb23561263eecae"
		res := FinalizerSelectionSeed(seed, 1, types.RoundId(2))
		if res.Hex() != expected {
			t.Errorf("expected %s, got %s", expected, res.Hex())
		}
	})

}

package utils

import (
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/ty"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestFinalizerSelectionSeed(t *testing.T) {
	t.Run("encoding matches reference Typescript implementation", func(t *testing.T) {
		seed := common.HexToHash("0x561c1876599d7b1583693711da55508f038908e7b91b6dc893b099a2eeb024bd").Big()
		expected := "0x9a22e13bd742533ab33080e2b3f96d1571bec59267ef1b721cb23561263eecae"
		res := FinalizerSelectionSeed(seed, 1, ty2.RoundId(2))
		if res.Hex() != expected {
			t.Errorf("expected %s, got %s", expected, res.Hex())
		}
	})

}

func TestRewardClaimHash(t *testing.T) {
	t.Run("encodes correctly", func(t *testing.T) {
		amount, _ := new(big.Int).SetString("48398380199697751340269", 10)

		claim := ty.RewardClaim{
			Beneficiary: common.HexToAddress("0xa174d46ef49d7d4a0328f9910222689e9eab2f45"),
			Amount:      amount,
			Type:        2,
		}
		hash := RewardClaimHash(213, claim)
		expected := "0xd6ebe34021a480411c18676c77aba5eb104ee3caf21537879548ab05833bfedf"

		if hash.Hex() != expected {
			t.Errorf("expected %s, got %s", expected, hash.Hex())
		}
	})
}

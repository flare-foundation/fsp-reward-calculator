package utils

import (
	common2 "fsp-rewards-calculator/common"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

var rndArgs = abi.Arguments{
	{
		Type: common2.Uint256Type,
	},
	{
		Type: common2.Uint256Type,
	},
}

func FeedSelectionRandom(random *big.Int, round ty2.RoundId) *big.Int {
	encoded, err := rndArgs.Pack(random, big.NewInt(int64(round)))
	if err != nil {
		logger.Fatal("error packing arguments %d, %v: %s", round, random, err)
	}
	hash := crypto.Keccak256(encoded)
	return new(big.Int).SetBytes(hash[:])
}

var finalizerArgs = abi.Arguments{
	{Type: common2.Uint256Type},
	{Type: common2.Uint8Type},
	{Type: common2.Uint64Type},
}

func FinalizerSelectionSeed(seed *big.Int, protocolId byte, round ty2.RoundId) common.Hash {
	encoded, err := finalizerArgs.Pack(seed, protocolId, uint64(round))
	if err != nil {
		logger.Fatal("error packing arguments %d, %v: %s", round, seed, err)
	}
	return crypto.Keccak256Hash(encoded)
}

var (
	bytes20Type, _  = abi.NewType("bytes20", "", nil)
	uint120Type, _  = abi.NewType("uint120", "", nil)
	rewardClaimArgs = abi.Arguments{
		{Type: common2.Uint64Type},
		{Type: bytes20Type},
		{Type: uint120Type},
		{Type: common2.Uint8Type},
	}
)

func RewardClaimHash(epoch ty2.EpochId, claim ty.RewardClaim) common.Hash {
	encoded, err := rewardClaimArgs.Pack(
		epoch,
		claim.Beneficiary,
		claim.Amount,
		claim.Type,
	)
	if err != nil {
		logger.Fatal("error packing arguments %v: %s", claim, err)
	}
	return crypto.Keccak256Hash(encoded)
}

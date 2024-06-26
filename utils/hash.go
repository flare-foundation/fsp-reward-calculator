package utils

import (
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

var (
	AddressType, _ = abi.NewType("address", "", nil)
	Uint32Type, _  = abi.NewType("uint32", "", nil)
	Uint256Type, _ = abi.NewType("uint256", "", nil)
	BytesType, _   = abi.NewType("bytes", "", nil)
)

var commitArgs = abi.Arguments{
	{
		Type: AddressType,
	},
	{
		Type: Uint32Type,
	},
	{
		Type: Uint256Type,
	},
	{
		Type: BytesType,
	},
}

func CommitHash(voter common.Address, round uint32, random common.Hash, feedValues []byte) common.Hash {
	encoded, err := commitArgs.Pack(voter, round, random.Big(), feedValues)
	if err != nil {
		logger.Fatal("error packing arguments: %s", err)
	}
	return crypto.Keccak256Hash(encoded)
}

var rndArgs = abi.Arguments{
	{
		Type: Uint256Type,
	},
	{
		Type: Uint256Type,
	},
}

func FeedSelectionRandom(random *big.Int, round types.RoundId) *big.Int {
	encoded, err := rndArgs.Pack(random, round)
	if err != nil {
		logger.Fatal("error packing arguments %d, %v: %s", round, random, err)
	}
	hash := crypto.Keccak256(encoded)
	return new(big.Int).SetBytes(hash[:])
}

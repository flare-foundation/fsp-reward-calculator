package common

import (
	"fsp-rewards-calculator/logger"

	"github.com/ethereum/go-ethereum/accounts/abi"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	AddressType, _ = abi.NewType("address", "", nil)
	Uint8Type, _   = abi.NewType("uint8", "", nil)
	Uint32Type, _  = abi.NewType("uint32", "", nil)
	Uint64Type, _  = abi.NewType("uint64", "", nil)
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

func CommitHash(voter common2.Address, round uint32, random common2.Hash, feedValues []byte) common2.Hash {
	encoded, err := commitArgs.Pack(voter, round, random.Big(), feedValues)
	if err != nil {
		logger.Fatal("error packing arguments: %s", err)
	}
	return crypto.Keccak256Hash(encoded)
}

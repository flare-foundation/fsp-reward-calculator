package utils

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

var (
	addressType, _ = abi.NewType("address", "", nil)
	uint32Type, _  = abi.NewType("uint32", "", nil)
	uint256Type, _ = abi.NewType("uint256", "", nil)
	bytesType, _   = abi.NewType("bytes", "", nil)

	args = abi.Arguments{
		{
			Type: addressType,
		},
		{
			Type: uint32Type,
		},
		{
			Type: uint256Type,
		},
		{
			Type: bytesType,
		},
	}
)

func CommitHash(voter common.Address, round uint32, random common.Hash, feedValues []byte) (common.Hash, error) {
	encoded, err := args.Pack(voter, round, random.Big(), feedValues)
	if err != nil {
		return common.Hash{}, errors.Wrap(err, "error packing arguments")
	}
	return crypto.Keccak256Hash(encoded), nil
}

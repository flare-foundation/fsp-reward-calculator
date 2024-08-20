package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type EpochId uint64

func (r *EpochId) Add(n uint64) EpochId {
	return EpochId(uint64(*r) + n)
}

type RoundId uint64

func (r *RoundId) Add(n int) RoundId {
	return RoundId(int(*r) + n)
}

type ClaimType uint8

// TODO: Check correct
const (
	Direct ClaimType = iota
	Fee
	WNat
	Mirror
	CChain
)

type RewardClaim struct {
	//Round 	 types.RoundId
	Beneficiary common.Address
	Amount      *big.Int
	Type        ClaimType
}

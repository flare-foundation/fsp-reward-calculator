package ty

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ClaimType uint8

const (
	Direct ClaimType = iota
	Fee
	WNat
	Mirror
	CChain
)

type RewardClaim struct {
	Beneficiary common.Address
	Amount      *big.Int
	Type        ClaimType
}

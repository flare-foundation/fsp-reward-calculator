package ty

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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

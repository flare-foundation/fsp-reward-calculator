package ty

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

type VoterId common.Address
type VoterSubmit common.Address
type VoterSubmitSignatures common.Address
type VoterSigning common.Address
type VoterDelegation common.Address

func (v *VoterId) String() string {
	return common.Address(*v).String()
}
func (v *VoterSubmit) String() string { return common.Address(*v).String() }

func (v *VoterSubmitSignatures) String() string {
	return common.Address(*v).String()
}

func (v *VoterSigning) String() string {
	return common.Address(*v).String()
}

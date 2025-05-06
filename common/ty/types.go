package ty

import (
	"github.com/ethereum/go-ethereum/common"
)

type EpochId uint64

func (r *EpochId) Add(n uint64) EpochId {
	return EpochId(uint64(*r) + n)
}

type RoundId uint64

func (r *RoundId) Add(n int) RoundId {
	return RoundId(int(*r) + n)
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
func (v *VoterDelegation) String() string {
	return common.Address(*v).String()
}

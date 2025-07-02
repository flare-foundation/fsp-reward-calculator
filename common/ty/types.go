package ty

import (
	"github.com/ethereum/go-ethereum/common"
)

type RewardEpochId uint64

func (r *RewardEpochId) Add(n uint64) RewardEpochId {
	return RewardEpochId(uint64(*r) + n)
}

type VotingEpochId uint32

func (v *VotingEpochId) Add(n uint32) VotingEpochId {
	return VotingEpochId(uint32(*v) + n)
}
func (v *VotingEpochId) Value() uint32 { return uint32(*v) }

type RoundId uint32

func (r *RoundId) Add(n int) RoundId {
	return RoundId(int(*r) + n)
}
func (r *RoundId) Value() uint32 { return uint32(*r) }

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

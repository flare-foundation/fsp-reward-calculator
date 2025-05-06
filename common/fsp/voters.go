package fsp

import (
	"fsp-rewards-calculator/common/ty"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/calculator"
	"github.com/flare-foundation/go-flare-common/pkg/voters"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
	"sort"
)

type VoterInfo struct {
	Identity            ty.VoterId
	Submit              ty.VoterSubmit
	SubmitSignatures    ty.VoterSubmitSignatures
	Signing             ty.VoterSigning
	Delegation          ty.VoterDelegation
	CappedWeight        *big.Int
	DelegationFeeBips   uint16
	NodeIds             [][20]byte
	NodeWeights         []*big.Int
	SigningPolicyWeight uint16
}

type VoterIndex struct {
	PolicyOrder              []*VoterInfo
	ById                     map[ty.VoterId]*VoterInfo
	BySubmit                 map[ty.VoterSubmit]*VoterInfo
	BySubmitSignatures       map[ty.VoterSubmitSignatures]*VoterInfo
	BySigning                map[ty.VoterSigning]*VoterInfo
	ByDelegation             map[ty.VoterDelegation]*VoterInfo
	TotalCappedWeight        *big.Int
	TotalSigningPolicyWeight uint16
}

func GetVoterIndex(db *gorm.DB, epoch ty.EpochId, fromSec, toSec uint64, policyVoters map[common.Address]voters.VoterData) (*VoterIndex, error) {
	regs, err := getVoterRegisteredEvents(db, fromSec, toSec)
	if err != nil {
		return nil, errors.Errorf("error fetching voter registered regs: %s", err)
	}

	infos, err := getVoterInfoEvents(db, fromSec, toSec)
	if err != nil {
		return nil, errors.Errorf("error fetching voter info events: %s", err)
	}

	infoByIdentity := make(map[common.Address]*calculator.CalculatorVoterRegistrationInfo)
	for _, info := range infos {
		infoByIdentity[info.Voter] = info
	}

	if len(regs) != len(infos) {
		return nil, errors.Errorf("voter registered and voter info event count mismatch: %d != %d", len(regs), len(infos))
	}

	var voters []*VoterInfo
	for _, reg := range regs {
		if reg.RewardEpochId.Uint64() != uint64(epoch) {
			continue
		}

		info := infoByIdentity[reg.Voter]

		voters = append(voters, &VoterInfo{
			Identity:            ty.VoterId(reg.Voter),
			Submit:              ty.VoterSubmit(reg.SubmitAddress),
			SubmitSignatures:    ty.VoterSubmitSignatures(reg.SubmitSignaturesAddress),
			Signing:             ty.VoterSigning(reg.SigningPolicyAddress),
			Delegation:          ty.VoterDelegation(info.DelegationAddress),
			CappedWeight:        info.WNatCappedWeight,
			DelegationFeeBips:   info.DelegationFeeBIPS,
			NodeIds:             info.NodeIds,
			NodeWeights:         info.NodeWeights,
			SigningPolicyWeight: policyVoters[reg.SigningPolicyAddress].Weight,
		})
	}

	// Sort according to signing policy order
	sort.Slice(voters, func(i, j int) bool {
		indexI := policyVoters[common.Address(voters[i].Signing)].Index
		indexJ := policyVoters[common.Address(voters[j].Signing)].Index
		return indexI < indexJ
	})

	return newVoterIndex(voters), nil
}

func newVoterIndex(voters []*VoterInfo) *VoterIndex {
	byId := make(map[ty.VoterId]*VoterInfo)
	bySubmit := make(map[ty.VoterSubmit]*VoterInfo)
	bySubmitSignatures := make(map[ty.VoterSubmitSignatures]*VoterInfo)
	bySigning := make(map[ty.VoterSigning]*VoterInfo)
	byDelegation := make(map[ty.VoterDelegation]*VoterInfo)
	for _, v := range voters {
		byId[v.Identity] = v
		bySubmit[v.Submit] = v
		bySubmitSignatures[v.SubmitSignatures] = v
		bySigning[v.Signing] = v
		byDelegation[v.Delegation] = v
	}
	totalCappedWeight := big.NewInt(0)
	totalSigningPolicyWeight := uint16(0)
	for _, v := range voters {
		totalCappedWeight.Add(totalCappedWeight, v.CappedWeight)
		totalSigningPolicyWeight += v.SigningPolicyWeight
	}
	return &VoterIndex{
		PolicyOrder:              voters,
		ById:                     byId,
		BySubmit:                 bySubmit,
		BySubmitSignatures:       bySubmitSignatures,
		BySigning:                bySigning,
		ByDelegation:             byDelegation,
		TotalCappedWeight:        totalCappedWeight,
		TotalSigningPolicyWeight: totalSigningPolicyWeight,
	}
}

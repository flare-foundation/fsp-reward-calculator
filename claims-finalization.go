package main

import (
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"slices"
)

func calcFinalizationRewardClaims(
	round types.RoundId,
	reward *big.Int,
	finalizations []*Finalization,
	eligibleVoters []*VoterInfo,
	eligibleFinalizers map[common.Address]bool,
) []RewardClaim {

	// TODO: Pre-compute
	successIndex := slices.IndexFunc(finalizations, func(f *Finalization) bool {
		return f.Info.Reverted == false
	})

	if successIndex < 0 {
		return []RewardClaim{burnClaim(reward)}
	}

	firstSuccessfulFinalization := finalizations[successIndex]
	gracePeriodDeadline := params.Net.Epoch.RevealDeadlineSec(round+1) + params.Net.Ftso.GracePeriodForFinalizationDurationSec
	gracePeriodDeadline += 1 // TODO: This is to match TS implementation logic, need to check if it's correct

	if firstSuccessfulFinalization.Info.TimestampSec > gracePeriodDeadline {
		// No voter provided finalization in grace period. The first successful finalizer gets the full reward.
		return []RewardClaim{
			{
				Beneficiary: firstSuccessfulFinalization.Info.From,
				Amount:      reward,
				Type:        Direct,
			},
		}
	}

	// TODO: Handle case when finalization is late and sent in the following round

	var graceFinalizations []*Finalization
	for _, finalization := range finalizations {
		if eligibleFinalizers[finalization.Info.From] && finalization.Info.TimestampSec <= gracePeriodDeadline {
			graceFinalizations = append(graceFinalizations, finalization)
		}
	}
	// We have at least one successful finalization in the grace period, but from non-eligible voters -> burn the reward.
	if len(graceFinalizations) == 0 || len(eligibleVoters) == 0 {
		return []RewardClaim{burnClaim(reward)}
	}

	var claims []RewardClaim

	// The reward should be distributed equally among all the eligible finalizers.
	// Note that each finalizer was chosen by probability corresponding to its relative weight.
	// Consequently, the real weight should not be taken into account here.
	undistributedAmount := big.NewInt(0).Set(reward)
	undistributedWeight := big.NewInt(int64(len(eligibleFinalizers)))

	eligibleVoterBySigning := map[VoterSigning]*VoterInfo{}
	for _, voter := range eligibleVoters {
		eligibleVoterBySigning[voter.Signing] = voter
	}
	for _, finalization := range graceFinalizations {
		voter := eligibleVoterBySigning[VoterSigning(finalization.Info.From)]
		if voter == nil {
			continue
		}

		claimAmount := big.NewInt(0).Div(undistributedAmount, undistributedWeight)
		undistributedAmount.Sub(undistributedAmount, claimAmount)
		undistributedWeight.Sub(undistributedWeight, big.NewInt(1))

		claims = append(claims, signingWeightClaimsForVoter(voter, claimAmount)...)
	}

	if undistributedAmount.Cmp(bigZero) != 0 {
		logger.Info("Burning undistributed finalization reward amount: %s", undistributedAmount.String())
		claims = append(claims, burnClaim(undistributedAmount))
	}

	return claims
}

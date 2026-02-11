package rewards

import (
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/policy"
	"github.com/pkg/errors"
)

func burnClaim(amount *big.Int) ty.RewardClaim {
	return ty.RewardClaim{
		Beneficiary: params.Net.Ftso.BurnAddress,
		Amount:      amount,
		Type:        ty.Direct,
	}
}

func getFinalizationClaims(
	round ty2.RoundId,
	reward *big.Int,
	finalizations []*fsp.Finalization,
	eligibleVoters []*fsp.VoterInfo,
	eligibleFinalizers map[common.Address]bool,
) []ty.RewardClaim {
	firstSuccessfulFinalization := firstSuccessful(finalizations)

	if firstSuccessfulFinalization == nil {
		return []ty.RewardClaim{burnClaim(reward)}
	}

	gracePeriodDeadline := params.Net.Epoch.RevealDeadlineSec(ty2.VotingEpochId(round)+1) + params.Net.Ftso.GracePeriodForFinalizationDurationSec + 1

	if firstSuccessfulFinalization.Info.TimestampSec > gracePeriodDeadline {
		// No voter provided finalization in the grace period. The first successful finalizer gets the full reward.
		return []ty.RewardClaim{
			{
				Beneficiary: firstSuccessfulFinalization.Info.From,
				Amount:      reward,
				Type:        ty.Direct,
			},
		}
	}

	var graceFinalizations []*fsp.Finalization
	for _, finalization := range finalizations {
		if eligibleFinalizers[finalization.Info.From] && finalization.Info.TimestampSec <= gracePeriodDeadline {
			graceFinalizations = append(graceFinalizations, finalization)
		}
	}
	// We have at least one successful finalization in the grace period, but from non-eligible voters -> burn the reward.
	if len(graceFinalizations) == 0 || len(eligibleVoters) == 0 {
		return []ty.RewardClaim{burnClaim(reward)}
	}

	var claims []ty.RewardClaim

	// The reward should be distributed equally among all the eligible finalizers.
	// Note that each finalizer was chosen by probability corresponding to its relative weight.
	// Consequently, the real weight should not be taken into account here.
	undistributedAmount := new(big.Int).Set(reward)
	undistributedWeight := big.NewInt(int64(len(eligibleFinalizers)))

	eligibleVoterBySigning := map[ty2.VoterSigning]*fsp.VoterInfo{}
	for _, voter := range eligibleVoters {
		eligibleVoterBySigning[voter.Signing] = voter
	}
	for _, finalization := range graceFinalizations {
		voter := eligibleVoterBySigning[ty2.VoterSigning(finalization.Info.From)]
		if voter == nil {
			continue
		}

		claimAmount := new(big.Int).Div(undistributedAmount, undistributedWeight)
		undistributedAmount.Sub(undistributedAmount, claimAmount)
		undistributedWeight.Sub(undistributedWeight, big.NewInt(1))

		claims = append(claims, SigningWeightClaimsForVoter(voter, claimAmount)...)
	}

	if undistributedAmount.Cmp(BigZero) != 0 {
		logger.Debug("Burning undistributed finalization reward amount: %s", undistributedAmount.String())
		claims = append(claims, burnClaim(undistributedAmount))
	}

	return claims
}

func selectFinalizers(
	round ty2.RoundId,
	policy *policy.SigningPolicy,
	protocol byte,
	threshold uint16,
) (map[common.Address]bool, error) {
	res, err := policy.Voters.SelectVoters(policy.Seed, protocol, uint32(round), threshold)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting finalizers")
	}
	return res, nil
}

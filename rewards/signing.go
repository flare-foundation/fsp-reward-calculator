package rewards

import (
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common"
)

func getSigningClaims(
	round ty2.RoundId,
	re *fsp.RewardEpoch,
	reward *big.Int,
	eligibleVoters []*fsp.VoterInfo,
	signers map[common.Hash]map[ty2.VoterSigning]fsp.SigInfo,
	finalizations []*fsp.Finalization,
) []ty.RewardClaim {
	doubleSigners := getDoubleSigners(signers)

	revealDeadline := params.Net.Epoch.RevealDeadlineSec(ty2.VotingEpochId(round) + 1)
	roundEnd := params.Net.Epoch.VotingRoundRewardEndSec(
		round.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows),
	)

	acceptedSigs := map[common.Hash]map[ty2.VoterSigning]fsp.SigInfo{}
	for hash, sigs := range signers {
		acceptedSigs[hash] = map[ty2.VoterSigning]fsp.SigInfo{}
		for signer, sig := range sigs {
			if sig.Timestamp < revealDeadline || sig.Timestamp > roundEnd {
				continue
			}
			acceptedSigs[hash][signer] = sig
		}
	}

	var rewardEligibleSigs []fsp.SigInfo

	// TODO: Pre-compute
	successIndex := slices.IndexFunc(finalizations, func(f *fsp.Finalization) bool {
		return !f.Info.Reverted
	})

	if successIndex < 0 {
		return []ty.RewardClaim{burnClaim(reward)}
	} else {
		successfulFinalization := finalizations[successIndex]

		deadline := min(
			successfulFinalization.Info.TimestampSec,
			roundEnd,
		)
		gracePeriod := revealDeadline + params.Net.Ftso.GracePeriodForSignaturesDurationSec

		finalizedHash := successfulFinalization.MerkleRoot.EncodedHash()

		for _, s := range acceptedSigs[finalizedHash] {
			if _, ok := doubleSigners[s.Signer]; ok {
				continue
			}

			if s.Timestamp <= gracePeriod || s.Timestamp <= deadline {
				rewardEligibleSigs = append(rewardEligibleSigs, s)
			}
		}
	}

	// Distribute rewards
	remainingWeight := uint16(0)
	for _, sig := range rewardEligibleSigs {
		remainingWeight += re.Policy.Voters.VoterDataMap[common.Address(sig.Signer)].Weight
	}

	if remainingWeight == 0 {
		return []ty.RewardClaim{burnClaim(reward)}
	}
	remainingAmount := new(big.Int).Set(reward)

	var claims []ty.RewardClaim
	// Sort signatures according to voter order in signing policy
	slices.SortFunc(rewardEligibleSigs, func(i, j fsp.SigInfo) int {
		indexI := re.Policy.Voters.VoterDataMap[common.Address(i.Signer)].Index
		indexJ := re.Policy.Voters.VoterDataMap[common.Address(j.Signer)].Index
		return indexI - indexJ
	})

	eligibleSigners := map[ty2.VoterSigning]*fsp.VoterInfo{}
	for _, voter := range eligibleVoters {
		eligibleSigners[voter.Signing] = voter
	}

	for _, sig := range rewardEligibleSigs {
		weight := re.Policy.Voters.VoterDataMap[common.Address(sig.Signer)].Weight
		if weight == 0 {
			continue
		}

		// TODO: clean up big.Int calculations
		claimAmount := new(big.Int).Div(
			bigTmp.Mul(remainingAmount, big.NewInt(int64(weight))),
			big.NewInt(int64(remainingWeight)),
		)

		remainingAmount.Sub(remainingAmount, claimAmount)
		remainingWeight -= weight

		if voter, ok := eligibleSigners[sig.Signer]; ok {
			claims = append(claims, SigningWeightClaimsForVoter(voter, claimAmount)...)
		} else {
			claims = append(claims, burnClaim(claimAmount))
		}
	}

	return claims
}

func SigningWeightClaimsForVoter(voter *fsp.VoterInfo, amount *big.Int) []ty.RewardClaim {
	var claims []ty.RewardClaim

	stakedWeight := big.NewInt(0)
	for _, w := range voter.NodeWeights {
		stakedWeight.Add(stakedWeight, w)
	}

	totalWeight := new(big.Int).Add(voter.CappedWeight, stakedWeight)
	if totalWeight.Cmp(BigZero) == 0 {
		logger.Fatal("voter totalWeight is zero, this should never happen")
	}

	stakingAmount := new(big.Int).Div(
		bigTmp.Mul(amount, stakedWeight),
		totalWeight,
	)
	delegationAmount := new(big.Int).Sub(amount, stakingAmount)
	delegationFee := new(big.Int).Div(
		bigTmp.Mul(delegationAmount, big.NewInt(int64(voter.DelegationFeeBips))),
		bigTotalBips,
	)
	cappedStakingFeeBips := big.NewInt(min(int64(voter.DelegationFeeBips), params.Net.Ftso.CappedStakingFeeBips))
	stakingFee := new(big.Int).Div(
		bigTmp.Mul(stakingAmount, cappedStakingFeeBips),
		bigTotalBips,
	)
	feeBeneficiary := common.Address(voter.Identity)
	delegationBeneficiary := common.Address(voter.Delegation)

	fee := new(big.Int).Add(delegationFee, stakingFee)
	if fee.Cmp(BigZero) != 0 {
		claims = append(claims, ty.RewardClaim{
			Beneficiary: feeBeneficiary,
			Amount:      fee,
			Type:        ty.Fee,
		})
	}

	delegationCommunityReward := new(big.Int).Sub(delegationAmount, delegationFee)
	claims = append(claims, ty.RewardClaim{
		Beneficiary: delegationBeneficiary,
		Amount:      delegationCommunityReward,
		Type:        ty.WNat,
	})

	remainingStakeWeight := new(big.Int).Set(stakedWeight)
	remainingStakeAmount := new(big.Int).Sub(stakingAmount, stakingFee)

	for i := range voter.NodeIds {
		nodeId := voter.NodeIds[i]
		nodeWeight := voter.NodeWeights[i]

		nodeAmount := big.NewInt(0)

		if nodeWeight.Cmp(BigZero) > 0 {
			nodeAmount = new(big.Int).Div(
				bigTmp.Mul(remainingStakeAmount, nodeWeight),
				remainingStakeWeight,
			)
		}

		remainingStakeAmount.Sub(remainingStakeAmount, nodeAmount)
		remainingStakeWeight.Sub(remainingStakeWeight, nodeWeight)

		claims = append(claims, ty.RewardClaim{
			Beneficiary: nodeId,
			Amount:      nodeAmount,
			Type:        ty.Mirror,
		})
	}

	if remainingStakeAmount.Cmp(BigZero) != 0 {
		logger.Fatal("remainingStakeAmount is not zero, this should never happen")
	}

	return claims
}

func getDoubleSigners(roundSigners map[common.Hash]map[ty2.VoterSigning]fsp.SigInfo) map[ty2.VoterSigning]bool {
	signed := map[ty2.VoterSigning]bool{}
	doubleSigners := map[ty2.VoterSigning]bool{}

	for _, signers := range roundSigners {
		for signer := range signers {
			if _, ok := signed[signer]; ok {
				doubleSigners[signer] = true
			}
			signed[signer] = true
		}
	}

	return doubleSigners
}

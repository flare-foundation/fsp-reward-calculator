package rewards

import (
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"slices"
)

func calcSigningRewardClaims(
	round types.RoundId,
	re RewardEpoch,
	reward *big.Int,
	eligibleVoters []*VoterInfo,
	signers map[common.Hash]map[VoterSigning]SigInfo,
	finalizations []*Finalization,
) []types.RewardClaim {
	doubleSigners := getDoubleSigners(signers)

	revealDeadline := params.Net.Epoch.RevealDeadlineSec(round + 1)
	roundEnd := params.Net.Epoch.VotingRoundEndSec(
		round.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows),
	)

	acceptedSigs := map[common.Hash]map[VoterSigning]SigInfo{}
	for hash, sigs := range signers {
		acceptedSigs[hash] = map[VoterSigning]SigInfo{}
		for signer, sig := range sigs {
			if sig.Timestamp < revealDeadline || sig.Timestamp > roundEnd {
				continue
			}
			acceptedSigs[hash][signer] = sig
		}
	}

	var rewardEligibleSigs []SigInfo

	// TODO: Pre-compute
	successIndex := slices.IndexFunc(finalizations, func(f *Finalization) bool {
		return f.Info.Reverted == false
	})

	if successIndex < 0 {
		signatures := acceptedHashSignatures(re, acceptedSigs)
		if signatures == nil {
			return []types.RewardClaim{burnClaim(reward)}
		} else {
			for _, s := range signatures {
				if _, ok := doubleSigners[s.Signer]; !ok {
					rewardEligibleSigs = append(rewardEligibleSigs, s)
				}
			}
		}
	} else {
		successfulFinalization := finalizations[successIndex]

		deadline := min(
			successfulFinalization.Info.TimestampSec,
			roundEnd,
		)
		gracePeriod := revealDeadline + params.Net.Ftso.GracePeriodForSignaturesDurationSec

		finalizedHash := successfulFinalization.merkleRoot.EncodedHash()

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
		return []types.RewardClaim{burnClaim(reward)}
	}
	remainingAmount := new(big.Int).Set(reward)

	var claims []types.RewardClaim
	// Sort signatures according to voter order in signing policy
	slices.SortFunc(rewardEligibleSigs, func(i, j SigInfo) int {
		indexI := re.Policy.Voters.VoterDataMap[common.Address(i.Signer)].Index
		indexJ := re.Policy.Voters.VoterDataMap[common.Address(j.Signer)].Index
		return indexI - indexJ
	})

	eligibleSigners := map[VoterSigning]*VoterInfo{}
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
			claims = append(claims, signingWeightClaimsForVoter(voter, claimAmount)...)
		} else {
			claims = append(claims, burnClaim(claimAmount))
		}
	}

	return claims
}

func signingWeightClaimsForVoter(voter *VoterInfo, amount *big.Int) []types.RewardClaim {
	var claims []types.RewardClaim

	stakedWeight := big.NewInt(0)
	for _, w := range voter.NodeWeights {
		stakedWeight.Add(stakedWeight, w)
	}

	totalWeight := new(big.Int).Add(voter.CappedWeight, stakedWeight)
	if totalWeight.Cmp(bigZero) == 0 {
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
	if fee.Cmp(bigZero) != 0 {
		claims = append(claims, types.RewardClaim{
			Beneficiary: feeBeneficiary,
			Amount:      fee,
			Type:        types.Fee,
		})
	}

	delegationCommunityReward := new(big.Int).Sub(delegationAmount, delegationFee)
	claims = append(claims, types.RewardClaim{
		Beneficiary: delegationBeneficiary,
		Amount:      delegationCommunityReward,
		Type:        types.WNat,
	})

	remainingStakeWeight := new(big.Int).Set(stakedWeight)
	remainingStakeAmount := new(big.Int).Sub(stakingAmount, stakingFee)

	for i := range voter.NodeIds {
		nodeId := voter.NodeIds[i]
		nodeWeight := voter.NodeWeights[i]

		nodeAmount := big.NewInt(0)

		if nodeWeight.Cmp(bigZero) > 0 {
			nodeAmount = new(big.Int).Div(
				bigTmp.Mul(remainingStakeAmount, nodeWeight),
				remainingStakeWeight,
			)
		}

		remainingStakeAmount.Sub(remainingStakeAmount, nodeAmount)
		remainingStakeWeight.Sub(remainingStakeWeight, nodeWeight)

		claims = append(claims, types.RewardClaim{
			Beneficiary: nodeId,
			Amount:      nodeAmount,
			Type:        types.Mirror,
		})
	}

	if remainingStakeAmount.Cmp(bigZero) != 0 {
		logger.Fatal("remainingStakeAmount is not zero, this should never happen")
	}

	return claims
}

func acceptedHashSignatures(
	re RewardEpoch,
	signaturesByHash map[common.Hash]map[VoterSigning]SigInfo,
) map[VoterSigning]SigInfo {
	threshold := re.Policy.Voters.TotalWeight * params.Net.Ftso.MinimalRewardedNonConsensusDepositedSignaturesPerHashBips / totalBips

	maxWeight := uint16(0)
	var result map[VoterSigning]SigInfo

	for _, signatures := range signaturesByHash {
		hashWeight := uint16(0)
		for _, info := range signatures {
			signerWeight := re.Policy.Voters.VoterDataMap[common.Address(info.Signer)].Weight
			hashWeight += signerWeight
		}
		if hashWeight > threshold && hashWeight > maxWeight {
			maxWeight = hashWeight
			result = signatures
		}
	}

	return result
}

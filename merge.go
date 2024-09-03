package main

import (
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// mergeClaims merges claims with the same beneficiary, type, and amount sign (penalty or reward).
func mergeClaims(claims []types.RewardClaim) []types.RewardClaim {
	byBeneficiaryTypeAndSign := make(map[common.Address]map[types.ClaimType]map[bool]*big.Int)

	for _, claim := range claims {
		if _, ok := byBeneficiaryTypeAndSign[claim.Beneficiary]; !ok {
			byBeneficiaryTypeAndSign[claim.Beneficiary] = make(map[types.ClaimType]map[bool]*big.Int)
		}
		if _, ok := byBeneficiaryTypeAndSign[claim.Beneficiary][claim.Type]; !ok {
			byBeneficiaryTypeAndSign[claim.Beneficiary][claim.Type] = make(map[bool]*big.Int)
		}

		isPositive := claim.Amount.Cmp(bigZero) > 0
		sum, ok := byBeneficiaryTypeAndSign[claim.Beneficiary][claim.Type][isPositive]
		if !ok {
			byBeneficiaryTypeAndSign[claim.Beneficiary][claim.Type][isPositive] = new(big.Int).Set(claim.Amount)
		} else {
			sum.Add(sum, claim.Amount)
		}
	}

	var merged []types.RewardClaim
	for beneficiary, byTypeAndSign := range byBeneficiaryTypeAndSign {
		for claimType, bySign := range byTypeAndSign {
			for _, amount := range bySign {
				merged = append(merged, types.RewardClaim{
					Beneficiary: beneficiary,
					Amount:      amount,
					Type:        claimType,
				})
			}
		}
	}

	return merged
}

func applyPenalties(claims []types.RewardClaim) []types.RewardClaim {
	var result []types.RewardClaim

	rewardByBeneficiaryAndType := make(map[common.Address]map[types.ClaimType]*types.RewardClaim)
	penaltyByBeneficiaryAndType := make(map[common.Address]map[types.ClaimType]*types.RewardClaim)

	burnAmount := big.NewInt(0)

	for i, claim := range claims {
		if claim.Beneficiary == BurnAddress {
			burnAmount.Add(burnAmount, claim.Amount)
			continue
		}
		if claim.Amount.Cmp(bigZero) > 0 {
			if _, ok := rewardByBeneficiaryAndType[claim.Beneficiary]; !ok {
				rewardByBeneficiaryAndType[claim.Beneficiary] = make(map[types.ClaimType]*types.RewardClaim)
			}
			rewardByBeneficiaryAndType[claim.Beneficiary][claim.Type] = &claims[i]
		} else {
			if _, ok := penaltyByBeneficiaryAndType[claim.Beneficiary]; !ok {
				penaltyByBeneficiaryAndType[claim.Beneficiary] = make(map[types.ClaimType]*types.RewardClaim)
			}
			penaltyByBeneficiaryAndType[claim.Beneficiary][claim.Type] = &claims[i]
		}
	}

	for beneficiary, rewardByType := range rewardByBeneficiaryAndType {
		for claimType, rewardClaim := range rewardByType {
			penaltyClaim, ok := penaltyByBeneficiaryAndType[beneficiary][claimType]
			if !ok {
				result = append(result, *rewardClaim)
				continue
			}

			// Penalty claim amount should be negative
			remainder := new(big.Int).Add(rewardClaim.Amount, penaltyClaim.Amount)

			if remainder.Cmp(bigZero) <= 0 {
				burnAmount.Add(burnAmount, rewardClaim.Amount)
			} else {
				burnAmount.Add(burnAmount, new(big.Int).Abs(penaltyClaim.Amount))

				remainderClaim := types.RewardClaim{
					Beneficiary: rewardClaim.Beneficiary,
					Amount:      remainder,
					Type:        rewardClaim.Type,
				}
				result = append(result, remainderClaim)
			}
		}
	}

	if burnAmount.Cmp(bigZero) > 0 {
		result = append(result, burnClaim(burnAmount))
	}

	claimSum := big.NewInt(0)
	for _, claim := range claims {
		if claim.Amount.Cmp(bigZero) > 0 {
			logger.Debug("Original Claim: %s, %s, %d", claim.Beneficiary.Hex(), claim.Type, claim.Amount)

			claimSum.Add(claimSum, claim.Amount)
		}
	}
	resultSum := big.NewInt(0)
	for _, claim := range result {
		logger.Debug("Result Claim: %s, %s, %d", claim.Beneficiary.Hex(), claim.Type, claim.Amount)
		resultSum.Add(resultSum, claim.Amount)
	}

	if claimSum.Cmp(resultSum) != 0 {
		panic("Claim sum does not match result sum")
	}

	return result
}

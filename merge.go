package main

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// mergeClaims merges claims with the same beneficiary, type, and amount sign (penalty or reward).
func mergeClaims(claims []RewardClaim) []RewardClaim {
	byBeneficiaryTypeAndSign := make(map[common.Address]map[ClaimType]map[bool]*big.Int)

	for _, claim := range claims {
		if _, ok := byBeneficiaryTypeAndSign[claim.Beneficiary]; !ok {
			byBeneficiaryTypeAndSign[claim.Beneficiary] = make(map[ClaimType]map[bool]*big.Int)
		}
		if _, ok := byBeneficiaryTypeAndSign[claim.Beneficiary][claim.Type]; !ok {
			byBeneficiaryTypeAndSign[claim.Beneficiary][claim.Type] = make(map[bool]*big.Int)
		}

		isPositive := claim.Amount.Cmp(bigZero) > 0
		sum, ok := byBeneficiaryTypeAndSign[claim.Beneficiary][claim.Type][isPositive]
		if !ok {
			byBeneficiaryTypeAndSign[claim.Beneficiary][claim.Type][isPositive] = big.NewInt(0).Set(claim.Amount)
		} else {
			sum.Add(sum, claim.Amount)
		}
	}

	var merged []RewardClaim
	for beneficiary, byTypeAndSign := range byBeneficiaryTypeAndSign {
		for claimType, bySign := range byTypeAndSign {
			for _, amount := range bySign {
				merged = append(merged, RewardClaim{
					Beneficiary: beneficiary,
					Amount:      amount,
					Type:        claimType,
				})
			}
		}
	}

	return merged
}

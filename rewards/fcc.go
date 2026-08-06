package rewards

import (
	"fsp-rewards-calculator/common/fcc"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"maps"
	"slices"
)

// GetFccRewards builds the reward claims for the FCC fees of a reward epoch.
//
// Both FCC fee sources are credited to the RewardManager at the moment they are paid, so every wei
// observed in these events is part of the funds available for the reward epoch and must appear in
// exactly one claim. Until the TEE rewarding logic exists, all of it is redirected to FccFeesAddress as
// a direct claim, so that all claims keep summing to the funds available on the RewardManager.
func GetFccRewards(fees *fcc.Fees) []ty.RewardClaim {
	epochClaims := make([]ty.RewardClaim, 0, len(fees.ByRound))

	for _, round := range slices.Sorted(maps.Keys(fees.ByRound)) {
		roundClaims := []ty.RewardClaim{{
			Beneficiary: params.Net.FccFeesAddress,
			Amount:      fees.ByRound[round],
			Type:        ty.Direct,
		}}
		utils.PrintRoundResults(roundClaims, fees.Epoch, round, "fcc-fees")
		epochClaims = append(epochClaims, roundClaims...)
	}

	logger.Info(
		"FCC fees for epoch %d: %s wei in %d voting round(s), claimed to %s",
		fees.Epoch, fees.Total, len(fees.ByRound), params.Net.FccFeesAddress,
	)

	return epochClaims
}

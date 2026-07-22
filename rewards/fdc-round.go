package rewards

import (
	"encoding/hex"
	"fsp-rewards-calculator/common/fdc"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"math/big"
)

func calculateFdcRoundRewards(
	re *fsp.RewardEpoch,
	countByType map[string]int,
	attestationRequestsByRound map[ty.RoundId][]fdc.AttestationRequest,
	consensusBitVoteByRound map[ty.RoundId]*big.Int,
) map[ty.RoundId]roundReward {
	rewardPerRound := map[ty.RoundId]roundReward{}

	totalBurnAmount := big.NewInt(0)
	totalRewardAmount := big.NewInt(0)

	if len(re.Offers.FdcInflation) == 0 {
		logger.Warn("no inflation offer for reward epoch %d", re.Epoch)
	} else {
		totalWeight := big.NewInt(0)
		burnWeight := big.NewInt(0)

		inflationOffer := re.Offers.FdcInflation[0]
		for _, conf := range inflationOffer.FdcConfigurations {
			t := hex.EncodeToString(append(conf.AttestationType[:], conf.Source[:]...))
			count := countByType[t]
			totalWeight.Add(totalWeight, conf.InflationShare)
			if count < int(conf.MinRequestsThreshold) {
				burnWeight.Add(burnWeight, conf.InflationShare)
			}
		}

		totalBurnAmount = big.NewInt(0).Div(
			big.NewInt(0).Mul(burnWeight, inflationOffer.Amount),
			totalWeight,
		)
		totalRewardAmount = big.NewInt(0).Sub(inflationOffer.Amount, totalBurnAmount)
	}

	logger.Info("Total FDC reward amount: %s, total amount to be burned: %s", totalRewardAmount, totalBurnAmount)

	perRound, rem := totalRewardAmount.DivMod(totalRewardAmount, big.NewInt(int64(re.EndRound-re.StartRound+1)), big.NewInt(0))
	burnPerRound, remB := totalBurnAmount.DivMod(totalBurnAmount, big.NewInt(int64(re.EndRound-re.StartRound+1)), big.NewInt(0))

	for round := re.StartRound; round <= re.EndRound; round++ {
		amount := new(big.Int).Set(perRound)
		if big.NewInt(int64(round-re.StartRound)).Cmp(rem) < 0 {
			amount.Add(amount, big.NewInt(1))
		}

		burnAmount := new(big.Int).Set(burnPerRound)
		if big.NewInt(int64(round-re.StartRound)).Cmp(remB) < 0 {
			burnAmount.Add(burnAmount, big.NewInt(1))
		}

		feeAmount := big.NewInt(0)
		feeBurnAmount := big.NewInt(0)

		for i, r := range attestationRequestsByRound[round] {
			if isConfirmed(i, consensusBitVoteByRound[round]) {
				feeAmount.Add(feeAmount, r.MergedFee)
			} else {
				feeBurnAmount.Add(feeBurnAmount, r.MergedFee)
			}
		}

		// FIP.16: a share of the confirmed request fees goes to the FIRE pool instead of the FDC
		// reward pool; the rounding remainder stays distributable.
		fireFeeAmount := big.NewInt(0)
		if params.Fip16Active(re.Epoch) {
			fireFeeAmount.Div(
				new(big.Int).Mul(feeAmount, params.Net.Fdc.FireFeeSplitBips),
				bigTotalBips,
			)
		}

		rewardPerRound[round] = roundReward{
			amount: amount.Add(amount, new(big.Int).Sub(feeAmount, fireFeeAmount)),
			burn:   burnAmount.Add(burnAmount, feeBurnAmount),
			fire:   fireFeeAmount,
		}
	}

	return rewardPerRound
}

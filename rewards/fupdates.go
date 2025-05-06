package rewards

import (
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/ftso"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func gatFUpdateClaims(re *fsp.RewardEpoch, roundUpdates *ftso.FUpdate, rewardOffer FUFeedReward, medianResult *ftso.Result, medianDecimals int) []ty.RewardClaim {
	var claims []ty.RewardClaim

	burnClaim := checkBurnReward(rewardOffer, roundUpdates, medianResult, medianDecimals)
	if burnClaim != nil {
		claims = append(claims, *burnClaim)
		return claims
	}

	subs := big.NewInt(int64(len(roundUpdates.Submitters)))
	perRound, rem := new(big.Int).DivMod(rewardOffer.Amount, subs, big.NewInt(0))

	logger.Info("Reward offer amount %s, per round %s, remainder %s", rewardOffer.Amount, perRound, rem)

	for i := range roundUpdates.Submitters {
		signing := roundUpdates.Submitters[i]
		voter := re.VoterIndex.BySigning[signing]
		if voter == nil {
			logger.Fatal("Voter not found for FU submitter signing address %s", signing)
		}

		feeBips := big.NewInt(int64(voter.DelegationFeeBips))

		amount := new(big.Int).Set(perRound)
		if big.NewInt(int64(i)).Cmp(rem) < 0 {
			amount.Add(amount, big.NewInt(1))
		}
		feeAmount := new(big.Int).Div(bigTmp.Mul(amount, feeBips), bigTotalBips)

		claims = append(claims, ty.RewardClaim{
			Beneficiary: common.Address(voter.Identity),
			Amount:      feeAmount,
			Type:        ty.Fee,
		})
		claims = append(claims, ty.RewardClaim{
			Beneficiary: common.Address(voter.Delegation),
			Amount:      new(big.Int).Sub(amount, feeAmount),
			Type:        ty.WNat,
		})
	}

	return claims
}

func checkBurnReward(rewardOffer FUFeedReward, roundUpdates *ftso.FUpdate, medianResult *ftso.Result, medianDecimals int) *ty.RewardClaim {
	if rewardOffer.ShouldBurn {
		return &ty.RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      new(big.Int).Set(rewardOffer.Amount),
			Type:        ty.Direct,
		}

	}
	if len(roundUpdates.Submitters) == 0 || medianResult == nil {
		return &ty.RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      new(big.Int).Set(rewardOffer.Amount),
			Type:        ty.Direct,
		}
	}

	median := big.NewInt(int64(medianResult.Median))
	value := roundUpdates.Feeds.Values[rewardOffer.FeedIndex]
	decimals := int(roundUpdates.Feeds.Decimals[rewardOffer.FeedIndex])

	if decimals > medianDecimals {
		median.Mul(median, PowerOfTen(int64(decimals-medianDecimals)))
	} else if decimals < medianDecimals {
		value.Mul(value, PowerOfTen(int64(medianDecimals-decimals)))
	}

	delta := new(big.Int).Div(
		bigTmp.Mul(median, bigTmp.SetInt64(int64(rewardOffer.FeedConfig.RewardBandValue))),
		bigTotalPPM,
	)

	low := new(big.Int).Sub(median, delta)
	high := new(big.Int).Add(median, delta)

	if value.Cmp(low) < 0 || value.Cmp(high) > 0 {
		return &ty.RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      new(big.Int).Set(rewardOffer.Amount),
			Type:        ty.Direct,
		}
	}

	return nil
}

// PowerOfTen calculates 10^n using big.Int
func PowerOfTen(n int64) *big.Int {
	base := big.NewInt(10)
	exponent := big.NewInt(n)
	return new(big.Int).Exp(base, exponent, nil)
}

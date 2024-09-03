package rewards

import (
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func calculateFUpdateClaims(re RewardEpoch, roundUpdates *FUpdate, rewardOffer FUFeedReward, medianResult *MedianResult, medianDecimals int) []types.RewardClaim {
	var claims []types.RewardClaim

	burnClaim := checkBurnReward(rewardOffer, roundUpdates, medianResult, medianDecimals)
	if burnClaim != nil {
		claims = append(claims, *burnClaim)
		return claims
	}

	subs := big.NewInt(int64(len(roundUpdates.submitters)))
	perRound, rem := new(big.Int).DivMod(rewardOffer.Amount, subs, big.NewInt(0))

	for i := range roundUpdates.submitters {
		signing := roundUpdates.submitters[i]
		voter := re.VoterIndex.bySigning[signing]
		if voter == nil {
			logger.Fatal("Voter not found for FU submitter signing address %s", signing.String())
		}

		feeBips := big.NewInt(int64(voter.DelegationFeeBips))

		amount := new(big.Int).Set(perRound)
		if big.NewInt(int64(i)).Cmp(rem) < 0 {
			amount.Add(amount, big.NewInt(1))
		}
		feeAmount := new(big.Int).Div(bigTmp.Mul(amount, feeBips), bigTotalBips)

		claims = append(claims, types.RewardClaim{
			Beneficiary: common.Address(voter.Identity),
			Amount:      feeAmount,
			Type:        types.Fee,
		})
		claims = append(claims, types.RewardClaim{
			Beneficiary: common.Address(voter.Delegation),
			Amount:      new(big.Int).Sub(amount, feeAmount),
			Type:        types.WNat,
		})
	}

	return claims
}

func checkBurnReward(rewardOffer FUFeedReward, roundUpdates *FUpdate, medianResult *MedianResult, medianDecimals int) *types.RewardClaim {
	if rewardOffer.ShouldBurn {
		return &types.RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      new(big.Int).Set(rewardOffer.Amount),
			Type:        types.Direct,
		}

	}
	if len(roundUpdates.submitters) == 0 || medianResult == nil {
		return &types.RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      new(big.Int).Set(rewardOffer.Amount),
			Type:        types.Direct,
		}
	}

	median := big.NewInt(int64(medianResult.Median))
	value := roundUpdates.feeds.Values[rewardOffer.FeedIndex]
	decimals := int(roundUpdates.feeds.Decimals[rewardOffer.FeedIndex])

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
		return &types.RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      new(big.Int).Set(rewardOffer.Amount),
			Type:        types.Direct,
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

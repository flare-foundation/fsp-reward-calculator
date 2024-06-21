package main

import (
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/types"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

func abs(v int32) uint32 {
	if v < 0 {
		return uint32(-v)
	}
	return uint32(v)
}

type voterRecord struct {
	voter        VoterSubmit
	weight       *big.Int
	isPct, isIqr bool
}

var randomArgs = abi.Arguments{{Type: utils.BytesType}, {Type: utils.Uint256Type}, {Type: utils.AddressType}}

func calcMedianRewardClaims(round types.RoundId, re RewardEpoch, rewardShare *big.Int, rewardOffer FeedReward, medianResult *MedianResult) []RewardClaim {
	var epochClaims []RewardClaim

	// Burn rewardOffer if turnout condition not reached
	if medianResult == nil || !isEnoughParticipation(medianResult.ParticipantWeight, re.Voters.totalCappedWeight, rewardOffer.Feed.MinRewardedTurnoutBIPS) {
		epochClaims = append(epochClaims, RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      big.NewInt(0).Set(rewardShare),
			Type:        Direct,
		})
		return epochClaims
	}

	secondaryBandDiff := abs(medianResult.Median) * rewardOffer.Feed.SecondaryBandWidthPPMs / totalPpm
	lowPct := medianResult.Median - int32(secondaryBandDiff)
	highPct := medianResult.Median + int32(secondaryBandDiff)

	lowIQR := medianResult.Q1
	highIQR := medianResult.Q3

	iqrSum := big.NewInt(0) // eligible Weight for IQR rewardOffer
	pctSum := big.NewInt(0) // eligible Weight for PCT rewardOffer

	var voterRecords []voterRecord
	for _, submission := range medianResult.inputValues {
		value := submission.value

		isPct := value > lowPct && value < highPct
		isIqr := (value > lowIQR && value < highIQR) || (value == lowIQR || value == highIQR) && randomSelect(rewardOffer.Feed.Id, round, submission.voter)

		if isPct {
			pctSum.Add(pctSum, submission.weight)
		}
		if isIqr {
			iqrSum.Add(iqrSum, submission.weight)
		}

		voterRecords = append(voterRecords, voterRecord{
			voter:  submission.voter,
			weight: submission.weight,
			isPct:  isPct,
			isIqr:  isIqr,
		})
	}

	totalNormWeight := big.NewInt(0)
	for i, record := range voterRecords {
		newWeight := big.NewInt(0)
		if pctSum.Cmp(bigZero) == 0 {
			if record.isIqr {
				newWeight.Set(record.weight)
			}
		} else {
			if record.isIqr {
				newWeight.Mul(
					big.NewInt(int64(rewardOffer.Feed.PrimaryBandRewardSharePPM)),
					bigTmp.Mul(
						record.weight,
						pctSum,
					),
				)
			}
			if record.isPct {
				newWeight.Add(
					newWeight,
					bigTmp.Mul(
						big.NewInt(int64(totalPpm-rewardOffer.Feed.PrimaryBandRewardSharePPM)),
						bigTmp.Mul(
							record.weight,
							iqrSum,
						),
					),
				)
			}
		}
		voterRecords[i].weight = newWeight
		totalNormWeight.Add(totalNormWeight, newWeight)
	}

	if totalNormWeight.Cmp(bigZero) == 0 {
		// Burn rewardOffer if no eligible submissions
		epochClaims = append(epochClaims, RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      big.NewInt(0).Set(rewardShare),
			Type:        Direct,
		})
		return epochClaims
	}

	totalReward := big.NewInt(0)
	availableReward := big.NewInt(0).Set(rewardShare)
	availableWeight := big.NewInt(0).Set(totalNormWeight)

	var claims []RewardClaim
	for _, record := range voterRecords {
		if record.weight.Cmp(bigZero) == 0 {
			continue
		}
		reward := big.NewInt(0)
		if record.weight.Cmp(bigZero) > 0 {
			if availableWeight.Cmp(bigZero) == 0 {
				logger.Fatal("availableWeight is zero, this should never happen")
			}
			reward.Div(
				bigTmp.Mul(
					record.weight,
					availableReward,
				),
				availableWeight,
			)
		}
		totalReward.Add(totalReward, reward)

		claims = append(claims, generateClaimsForVoter(re.Voters.bySubmit[record.voter], reward, rewardOffer)...)
	}

	return claims
}

func randomSelect(feedId FeedId, round types.RoundId, voter VoterSubmit) bool {
	pack, err := randomArgs.Pack(feedId[:], big.NewInt(int64(round)), common.Address(voter))
	if err != nil {
		logger.Fatal("error packing arguments, this should never happen: %s", err)
	}
	hash := crypto.Keccak256Hash(pack)
	return hash[len(hash)-1]%2 == 1
}

func isEnoughParticipation(participatingWeight, totalWeight *big.Int, minBips uint16) bool {
	return big.NewInt(0).Mul(
		participatingWeight,
		bigTotalBips,
	).Cmp(
		big.NewInt(0).Mul(
			totalWeight,
			big.NewInt(int64(minBips)),
		),
	) >= 0
}

func generateClaimsForVoter(voter *VoterInfo, reward *big.Int, offer FeedReward) []RewardClaim {
	var claims []RewardClaim

	voterFee := voter.delegationFeeBips
	fee := big.NewInt(0).Div(
		bigTmp.Mul(
			reward,
			big.NewInt(int64(voterFee)),
		),
		bigTotalBips,
	)

	if fee.Cmp(bigZero) > 0 {
		claims = append(claims, RewardClaim{
			Beneficiary: common.Address(voter.Identity),
			Amount:      fee,
			Type:        Fee,
		})
	}

	participationReward := big.NewInt(0).Sub(reward, fee)
	if participationReward.Cmp(bigZero) > 0 {
		claims = append(claims, RewardClaim{
			Beneficiary: common.Address(voter.Delegation),
			Amount:      participationReward,
			Type:        WNat,
		})
	}

	return claims
}

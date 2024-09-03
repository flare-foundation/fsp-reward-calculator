package main

import (
	"encoding/hex"
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

func calcMedianRewardClaims(round types.RoundId, re RewardEpoch, rewardShare *big.Int, rewardOffer FeedReward, medianResult *MedianResult) []types.RewardClaim {
	var epochClaims []types.RewardClaim

	// Burn rewardOffer if turnout condition not reached
	if medianResult == nil || !isEnoughParticipation(medianResult.ParticipantWeight, re.VoterIndex.totalCappedWeight, rewardOffer.Feed.MinRewardedTurnoutBIPS) {
		epochClaims = append(epochClaims, types.RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      new(big.Int).Set(rewardShare),
			Type:        types.Direct,
		})
		return epochClaims
	}

	sortedRecords, pctSum, iqrSum := getRecords(round, re, medianResult, rewardOffer)

	totalNormWeight := big.NewInt(0)
	for i, record := range sortedRecords {
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
		sortedRecords[i].weight = newWeight
		totalNormWeight.Add(totalNormWeight, newWeight)
	}

	if totalNormWeight.Cmp(bigZero) == 0 {
		// Burn rewardOffer if no eligible submissions
		epochClaims = append(epochClaims, types.RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      new(big.Int).Set(rewardShare),
			Type:        types.Direct,
		})
		return epochClaims
	}

	totalReward := big.NewInt(0)
	availableReward := new(big.Int).Set(rewardShare)
	availableWeight := new(big.Int).Set(totalNormWeight)

	for _, record := range sortedRecords {
		logger.Debug("Voter %s, weight %d, isPct %t, isIqr %t", hex.EncodeToString(record.voter[:]), record.weight, record.isPct, record.isIqr)
	}

	var claims []types.RewardClaim
	for _, record := range sortedRecords {
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
			logger.Debug("Dividing for %s: %d * %d / %d, res %d",
				hex.EncodeToString(record.voter[:]),
				record.weight,
				availableReward, availableWeight, reward)

		}

		availableReward.Sub(availableReward, reward)
		availableWeight.Sub(availableWeight, record.weight)
		totalReward.Add(totalReward, reward)

		claims = append(claims, generateClaimsForVoter(re.VoterIndex.bySubmit[record.voter], reward)...)
	}

	if totalReward.Cmp(rewardShare) != 0 {
		logger.Fatal("totalReward %d is not equal to rewardShare %d, this should never happen", totalReward, rewardShare)
	}

	return claims
}

func getRecords(round types.RoundId, re RewardEpoch, medianResult *MedianResult, rewardOffer FeedReward) ([]voterRecord, *big.Int, *big.Int) {
	secondaryBandDiff := abs(medianResult.Median) * rewardOffer.Feed.SecondaryBandWidthPPMs / totalPpm
	lowPct := medianResult.Median - int32(secondaryBandDiff)
	highPct := medianResult.Median + int32(secondaryBandDiff)

	lowIQR := medianResult.Q1
	highIQR := medianResult.Q3

	pctSum := big.NewInt(0)
	iqrSum := big.NewInt(0)
	voterRecords := map[VoterSubmit]voterRecord{}
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

		voterRecords[submission.voter] = voterRecord{
			voter:  submission.voter,
			weight: submission.weight,
			isPct:  isPct,
			isIqr:  isIqr,
		}
	}

	sortedRecords := make([]voterRecord, 0, len(voterRecords))
	for _, signingAddr := range re.OrderedVoters {
		submit := re.VoterIndex.bySigning[signingAddr].Submit
		if record, ok := voterRecords[submit]; ok {
			sortedRecords = append(sortedRecords, record)
		}
	}
	return sortedRecords, pctSum, iqrSum
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
	return new(big.Int).Mul(
		participatingWeight,
		bigTotalBips,
	).Cmp(
		new(big.Int).Mul(
			totalWeight,
			big.NewInt(int64(minBips)),
		),
	) >= 0
}

func generateClaimsForVoter(voter *VoterInfo, reward *big.Int) []types.RewardClaim {
	logger.Debug("Generating claims for voter %s, amount %d", hex.EncodeToString(voter.Identity[:]), reward)

	var claims []types.RewardClaim

	voterFee := voter.DelegationFeeBips
	fee := new(big.Int).Div(
		bigTmp.Mul(
			reward,
			big.NewInt(int64(voterFee)),
		),
		bigTotalBips,
	)

	if fee.Cmp(bigZero) > 0 {
		claims = append(claims, types.RewardClaim{
			Beneficiary: common.Address(voter.Identity),
			Amount:      fee,
			Type:        types.Fee,
		})
	}

	participationReward := new(big.Int).Sub(reward, fee)
	if participationReward.Cmp(bigZero) > 0 {
		claims = append(claims, types.RewardClaim{
			Beneficiary: common.Address(voter.Delegation),
			Amount:      participationReward,
			Type:        types.WNat,
		})
	}

	return claims
}

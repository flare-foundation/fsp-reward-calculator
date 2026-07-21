package rewards

import (
	"encoding/hex"
	common2 "fsp-rewards-calculator/common"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/ftso"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type voterRecord struct {
	voter        ty2.VoterSubmit
	weight       *big.Int
	isPct, isIqr bool
}

// getMedianClaims calculates the median (accuracy) reward claims for the round. It also returns
// the voters that earned a non-zero median reward — once FIP.16 is active this precise set drives
// signing/finalization eligibility (a stake-only voter can earn an accuracy reward without
// producing a WNAT claim).
func getMedianClaims(round ty2.RoundId, re *fsp.RewardEpoch, rewardShare *big.Int, rewardOffer FeedReward, medianResult *ftso.Result) ([]ty.RewardClaim, []*fsp.VoterInfo) {
	var epochClaims []ty.RewardClaim

	// FIP.16: turnout is measured against the total normalized signing weight instead of the
	// total capped delegation weight (matching the unified median voting weight).
	totalVotingWeight := re.VoterIndex.TotalCappedWeight
	if params.Fip16Active(re.Epoch) {
		totalVotingWeight = big.NewInt(int64(re.VoterIndex.TotalSigningPolicyWeight))
	}

	// Burn rewardOffer if turnout condition not reached
	if medianResult.Quartiles == nil || !isEnoughParticipation(medianResult.Quartiles.ParticipantWeight, totalVotingWeight, rewardOffer.Feed.MinRewardedTurnoutBIPS) {
		epochClaims = append(epochClaims, ty.RewardClaim{
			Beneficiary: params.Net.Ftso.BurnAddress,
			Amount:      new(big.Int).Set(rewardShare),
			Type:        ty.Direct,
		})
		return epochClaims, nil
	}

	sortedRecords, pctSum, iqrSum := getRecords(round, re, medianResult, rewardOffer)

	totalNormWeight := big.NewInt(0)
	for i, record := range sortedRecords {
		newWeight := big.NewInt(0)
		if pctSum.Cmp(BigZero) == 0 {
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

	if totalNormWeight.Cmp(BigZero) == 0 {
		// Burn rewardOffer if no eligible submissions
		epochClaims = append(epochClaims, ty.RewardClaim{
			Beneficiary: params.Net.Ftso.BurnAddress,
			Amount:      new(big.Int).Set(rewardShare),
			Type:        ty.Direct,
		})
		return epochClaims, nil
	}

	totalReward := big.NewInt(0)
	availableReward := new(big.Int).Set(rewardShare)
	availableWeight := new(big.Int).Set(totalNormWeight)

	for _, record := range sortedRecords {
		logger.Debug("Voter %s, weight %d, isPct %t, isIqr %t", hex.EncodeToString(record.voter[:]), record.weight, record.isPct, record.isIqr)
	}

	var claims []ty.RewardClaim
	var rewardedVoters []*fsp.VoterInfo
	fip16Active := params.Fip16Active(re.Epoch)
	for _, record := range sortedRecords {
		if record.weight.Cmp(BigZero) == 0 {
			continue
		}
		reward := big.NewInt(0)
		if record.weight.Cmp(BigZero) > 0 {
			if availableWeight.Cmp(BigZero) == 0 {
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

		availableReward.Sub(availableReward, reward)
		availableWeight.Sub(availableWeight, record.weight)
		totalReward.Add(totalReward, reward)

		voter := re.VoterIndex.BySubmit[record.voter]
		// FIP.16: the accuracy reward is split between delegators (WNAT) and stakers (MIRROR)
		// exactly like signing/finalization rewards. Before activation it is split into
		// fee + WNAT (delegation) only.
		if fip16Active {
			claims = append(claims, SigningWeightClaimsForVoter(voter, reward, re.Epoch)...)
		} else {
			claims = append(claims, generateClaimsForVoter(voter, reward)...)
		}
		if reward.Cmp(BigZero) > 0 {
			rewardedVoters = append(rewardedVoters, voter)
		}
	}

	if totalReward.Cmp(rewardShare) != 0 {
		logger.Fatal("totalReward %d is not equal to rewardShare %d, this should never happen", totalReward, rewardShare)
	}

	return claims, rewardedVoters
}

func getRecords(round ty2.RoundId, re *fsp.RewardEpoch, medianResult *ftso.Result, rewardOffer FeedReward) ([]voterRecord, *big.Int, *big.Int) {
	secondaryBandDiff := new(big.Int).Div(
		new(big.Int).Mul(
			big.NewInt(int64(abs(medianResult.Quartiles.Median))),
			big.NewInt(int64(rewardOffer.Feed.SecondaryBandWidthPPMs)),
		),
		bigTotalPPM,
	).Uint64()
	lowPct := medianResult.Quartiles.Median - int32(secondaryBandDiff)
	highPct := medianResult.Quartiles.Median + int32(secondaryBandDiff)

	lowIQR := medianResult.Quartiles.Q1
	highIQR := medianResult.Quartiles.Q3

	pctSum := big.NewInt(0)
	iqrSum := big.NewInt(0)
	voterRecords := map[ty2.VoterSubmit]voterRecord{}
	for _, submission := range medianResult.InputValues {
		value := submission.Value

		isPct := value > lowPct && value < highPct
		isIqr := (value > lowIQR && value < highIQR) || (value == lowIQR || value == highIQR) && randomSelect(rewardOffer.Feed.Id, round, submission.Voter)

		if isPct {
			pctSum.Add(pctSum, submission.Weight)
		}
		if isIqr {
			iqrSum.Add(iqrSum, submission.Weight)
		}

		voterRecords[submission.Voter] = voterRecord{
			voter:  submission.Voter,
			weight: submission.Weight,
			isPct:  isPct,
			isIqr:  isIqr,
		}
	}

	sortedRecords := make([]voterRecord, 0, len(voterRecords))
	for _, signingAddr := range re.OrderedVoters {
		submit := re.VoterIndex.BySigning[signingAddr].Submit
		if record, ok := voterRecords[submit]; ok {
			sortedRecords = append(sortedRecords, record)
		}
	}
	return sortedRecords, pctSum, iqrSum
}

var randomArgs = abi.Arguments{{Type: common2.BytesType}, {Type: common2.Uint256Type}, {Type: common2.AddressType}}

func randomSelect(feedId fsp.FeedId, round ty2.RoundId, voter ty2.VoterSubmit) bool {
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

func generateClaimsForVoter(voter *fsp.VoterInfo, reward *big.Int) []ty.RewardClaim {
	logger.Debug("Generating claims for voter %s, amount %d", hex.EncodeToString(voter.Identity[:]), reward)

	var claims []ty.RewardClaim

	voterFee := voter.DelegationFeeBips
	fee := new(big.Int).Div(
		bigTmp.Mul(
			reward,
			big.NewInt(int64(voterFee)),
		),
		bigTotalBips,
	)

	if fee.Cmp(BigZero) > 0 {
		claims = append(claims, ty.RewardClaim{
			Beneficiary: common.Address(voter.Identity),
			Amount:      fee,
			Type:        ty.Fee,
		})
	}

	participationReward := new(big.Int).Sub(reward, fee)
	if participationReward.Cmp(BigZero) > 0 {
		claims = append(claims, ty.RewardClaim{
			Beneficiary: common.Address(voter.Delegation),
			Amount:      participationReward,
			Type:        ty.WNat,
		})
	}

	return claims
}

func abs(v int32) uint32 {
	if v < 0 {
		return uint32(-v)
	}
	return uint32(v)
}

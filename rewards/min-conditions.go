package rewards

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/ftso"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"io"
	"math/big"
	"net/http"

	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
)

var FtsoScalingClosenessThresholdPpm = big.NewInt(5000)      // 0.5%
var FtsoScalingAvailabilityThresholdPpm = big.NewInt(800000) // 80%

func metFtsoCondition(voterIndex *fsp.VoterIndex, totalFeeds int, results map[ty.RoundId]ftso.RoundResult) map[ty.VoterId]bool {
	voterHits := map[ty.VoterId]int{}

	for _, result := range results {
		for _, feedResult := range result.Median {
			if feedResult == nil {
				continue
			}
			for _, voterValue := range feedResult.InputValues {
				median := big.NewInt(int64(feedResult.Quartiles.Median))
				delta := new(big.Int).Mul(median, FtsoScalingClosenessThresholdPpm)
				delta.Div(delta, bigTotalPPM)

				lowerBound := new(big.Int).Sub(median, delta)
				upperBound := new(big.Int).Add(median, delta)

				value := big.NewInt(int64(voterValue.Value))

				if value.Cmp(lowerBound) >= 0 && value.Cmp(upperBound) <= 0 {
					voterId := voterIndex.BySubmit[voterValue.Voter].Identity
					voterHits[voterId]++
				}
			}
		}
	}

	rounds := len(results)
	availableHits := rounds * totalFeeds

	metCondition := map[ty.VoterId]bool{}
	for voter, hits := range voterHits {
		bigHits := big.NewInt(int64(hits))

		// hits >= availableHits * FtsoScalingAvailabilityThresholdPpm/TotalPPM
		if new(big.Int).Mul(bigTotalPPM, bigHits).Cmp(new(big.Int).Mul(FtsoScalingAvailabilityThresholdPpm, big.NewInt(int64(availableHits)))) >= 0 {
			metCondition[voter] = true
		}
		logger.Debug("Voter %s hits: %d, met condition: %b", voter.String(), hits, metCondition[voter])
	}

	return metCondition
}

var FuThresholdPpm = big.NewInt(800000)            // 80%
var FuConsiderationThresholdPpm = big.NewInt(2000) // 0.2%

func metFUCondition(index *fsp.VoterIndex, updates map[ty.RoundId]*ftso.FUpdate) map[ty.VoterId]bool {
	voterUpdates := map[ty.VoterSigning]int{}
	totalUpdates := 0
	for _, update := range updates {
		for _, submitter := range update.Submitters {
			voterUpdates[submitter]++
		}
		totalUpdates += len(update.Submitters)
	}

	for _, voter := range index.PolicyOrder {
		logger.Debug("Voter %s submits: %d", voter.Identity.String(), voterUpdates[voter.Signing])
	}

	metCondition := map[ty.VoterId]bool{}
	totalSigningWeight := big.NewInt(int64(index.TotalSigningPolicyWeight))
	for _, voter := range index.PolicyOrder {
		signingWeight := big.NewInt(int64(voter.SigningPolicyWeight))
		relativeWeightPpm := new(big.Int).Div(
			new(big.Int).Mul(signingWeight, bigTotalPPM),
			totalSigningWeight,
		)
		expectedUpdatesPpm := new(big.Int).Div(
			new(big.Int).Mul(signingWeight, FuThresholdPpm),
			totalSigningWeight,
		)
		expectedUpdates := new(big.Int).Div(
			new(big.Int).Mul(expectedUpdatesPpm, big.NewInt(int64(totalUpdates))),
			bigTotalPPM,
		)

		if relativeWeightPpm.Cmp(FuConsiderationThresholdPpm) < 0 {
			metCondition[voter.Identity] = true
			continue
		}

		bigUpdates := big.NewInt(int64(voterUpdates[voter.Signing]))
		bigTotalUpdates := big.NewInt(int64(totalUpdates))
		// updates >= totalUpdates * expectedUpdatesPpm/TotalPPM
		if new(big.Int).Mul(bigUpdates, bigTotalPPM).Cmp(new(big.Int).Mul(expectedUpdatesPpm, bigTotalUpdates)) >= 0 {
			metCondition[voter.Identity] = true
		}

		logger.Debug("Voter %s expected updates: %d, actual updates: %d", voter.Identity.String(), expectedUpdates, voterUpdates[voter.Signing])
	}

	logger.Debug("Total submitters: %d", totalUpdates)
	return metCondition
}

var FDCThresholdPpm = big.NewInt(600000) // 60%

func metFDCCondition(totalRewardedRounds int, rewardedRounds map[ty.VoterId]int) map[ty.VoterId]bool {
	metCondition := map[ty.VoterId]bool{}

	bigTotalRounds := big.NewInt(int64(totalRewardedRounds))
	for voter, rewardedCount := range rewardedRounds {
		bigRewardedCount := big.NewInt(int64(rewardedCount))
		logger.Debug("Voter %s rewarded count: %d", voter.String(), rewardedCount)
		// rewardedCount >= totalRewardedRounds * FDCThresholdPpm/TotalPPM
		if new(big.Int).Mul(bigRewardedCount, bigTotalPPM).Cmp(new(big.Int).Mul(FDCThresholdPpm, bigTotalRounds)) >= 0 {
			metCondition[voter] = true
		}
	}

	return metCondition
}

var StakingUptimeThresholdPpm = big.NewInt(800000)               // 80%
var StakingMinSelfBondGwei = big.NewInt(1000000000000000)        // 1M FLR
var StakingMinDesiredSelfBondGwei = big.NewInt(3000000000000000) // 3M FLR
var StakingMinDesiredStakeGwei = big.NewInt(15000000000000000)   // 15M FLR

type StakingCondition uint8

const (
	NotMet StakingCondition = iota
	Met
	MetNoPass
)

func MetStakingCondition(epoch ty.RewardEpochId, voters *fsp.VoterIndex) map[ty.VoterId]StakingCondition {
	metCondition := map[ty.VoterId]StakingCondition{}

	validatorInfoByNode, err := FetchValidatorInfo(epoch)
	if err != nil {
		logger.Fatal("Failed to fetch validator info: %s", err)
	}

	for _, voter := range voters.PolicyOrder {
		stakeWithUptime := big.NewInt(0)
		totalSelfBond := big.NewInt(0)
		totalStake := big.NewInt(0)

		for _, node := range voter.NodeIds {
			nodeHex := hex.EncodeToString(node[:])
			validatorInfo, ok := validatorInfoByNode[nodeHex]
			if !ok {
				logger.Warn("Validator info not found for voter %s node %s", voter.Identity.String(), nodeHex)
				continue
			}

			if validatorInfo.UptimeEligible {
				stakeWithUptime.Add(stakeWithUptime, validatorInfo.TotalStakeAmount)
			}

			totalSelfBond.Add(totalSelfBond, validatorInfo.SelfBond)
			totalStake.Add(totalStake, validatorInfo.TotalStakeAmount)
		}

		// stakeWithUptime >= totalStake * StakingUptimeThresholdPpm/TotalPPM
		uptimeOk := new(big.Int).Mul(bigTotalPPM, stakeWithUptime).Cmp(new(big.Int).Mul(StakingUptimeThresholdPpm, totalStake)) >= 0

		if uptimeOk && totalSelfBond.Cmp(StakingMinSelfBondGwei) >= 0 {
			if totalSelfBond.Cmp(StakingMinDesiredSelfBondGwei) < 0 || totalStake.Cmp(StakingMinDesiredStakeGwei) < 0 {
				metCondition[voter.Identity] = MetNoPass
			} else {
				metCondition[voter.Identity] = Met
			}
		}
	}

	return metCondition
}

type ValidatorInfo struct {
	NodeId           string
	SelfBond         *big.Int
	TotalStakeAmount *big.Int
	UptimeEligible   bool
}

func FetchValidatorInfo(epoch ty.RewardEpochId) (map[string]ValidatorInfo, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/flare-foundation/reward-scripts/refs/heads/main/generated-files/reward-epoch-%d/initial-nodes-data.json", epoch)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch raw: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type jsonData struct {
		NodeId           string `json:"nodeId"`
		SelfBond         string `json:"selfBond"`
		TotalStakeAmount string `json:"totalStakeAmount"`
		UptimeEligible   bool   `json:"uptimeEligible"`
	}

	var raw []jsonData
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, err
	}

	res := map[string]ValidatorInfo{}
	for _, d := range raw {
		selfBond, ok := new(big.Int).SetString(d.SelfBond[:len(d.SelfBond)-1], 10)
		if !ok {
			return nil, errors.Errorf("failed to parse selfBond: %s", d.SelfBond)
		}

		totalStakeAmount, ok := new(big.Int).SetString(d.TotalStakeAmount[:len(d.TotalStakeAmount)-1], 10)
		if !ok {
			return nil, errors.Errorf("failed to parse totalStakeAmount: %s", d.TotalStakeAmount)
		}

		nodeIdHex, err := parseNodeIdHex(d.NodeId)
		if err != nil {
			return nil, err
		}

		res[nodeIdHex] = ValidatorInfo{
			NodeId:           nodeIdHex,
			SelfBond:         selfBond,
			TotalStakeAmount: totalStakeAmount,
			UptimeEligible:   d.UptimeEligible,
		}
	}

	return res, nil
}

func parseNodeIdHex(node string) (string, error) {
	// Strip prefix and decode
	decoded := base58.Decode(node[7:])
	if len(decoded) < 4 {
		return "", errors.Errorf("Decoded length is less than 4")
	}
	return hex.EncodeToString(decoded[:len(decoded)-4]), nil
}

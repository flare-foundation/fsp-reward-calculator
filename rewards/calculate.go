package rewards

import (
	"encoding/json"
	"fmt"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	ty2 "fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
	"io"
	"math/big"
	"net/http"
)

func CalculateResults(db *gorm.DB, epoch ty.EpochId) EpochResult {
	logger.Info("Calculating reward claims for epoch %d", epoch)

	allClaims, minCond := GetEpochClaims(db, epoch)
	PrintConditions(minCond, epoch)

	merged := MergeClaims(allClaims)
	logger.Info("Merged claims: %d, all claims %d", len(merged), len(allClaims))

	utils.PrintEpochClaims(merged, epoch, "all")
	finalClaims := ApplyPenalties(merged)
	utils.PrintEpochClaims(finalClaims, epoch, "merged")

	var resultClaims []ty2.RewardClaim
	if params.Net.Name == "flare" || params.Net.Name == "songbird" {
		nonConditions := buildResults(epoch, finalClaims)
		PrintResults(nonConditions, "-raw")

		resultClaims = applyMinConditions(epoch, finalClaims, minCond)
	} else {
		resultClaims = finalClaims
	}

	return buildResults(epoch, resultClaims)
}

func applyMinConditions(epoch ty.EpochId, merged []ty2.RewardClaim, cond map[*fsp.VoterInfo]MinConditions) []ty2.RewardClaim {
	currentPasses, err := fetCurrentPasses(epoch - 1)
	if err != nil {
		logger.Error("Error fetching current passes: %s, defaulting to 0 for all providers", err)
	}

	burnClaims := map[common.Address]bool{}
	burnClaims[params.Net.Ftso.BurnAddress] = true

	for voter, c := range cond {
		passes := currentPasses[voter.Identity] + c.PassDelta

		if passes < 0 {
			logger.Warn("Voter id %s del %s has negative passes: %d, burning claims", voter.Identity.String(), voter.Delegation.String(), passes)

			burnClaims[common.Address(voter.Identity)] = true
			burnClaims[common.Address(voter.Delegation)] = true
			for _, node := range voter.NodeIds {
				burnClaims[node] = true
			}
		}

		// TODO: Calculate resulting passes
		//if passes < 0 {
		//	passes = 0
		//}
		//if passes > 3 {
		//	passes = 3
		//}
	}

	var resRewards []ty2.RewardClaim
	finalBurnClaim := ty2.RewardClaim{
		Beneficiary: params.Net.Ftso.BurnAddress,
		Amount:      big.NewInt(0),
		Type:        ty2.Direct,
	}

	for _, claim := range merged {
		if burnClaims[claim.Beneficiary] {
			finalBurnClaim.Amount.Add(finalBurnClaim.Amount, claim.Amount)
			logger.Info("Burn claim: %s %d %d", claim.Beneficiary.String(), claim.Amount, claim.Type)
		} else {
			resRewards = append(resRewards, claim)
		}
	}

	if finalBurnClaim.Amount.Cmp(BigZero) > 0 {
		resRewards = append([]ty2.RewardClaim{finalBurnClaim}, resRewards...)
	}

	logger.Info("Original claims: %d, final claims: %d", len(merged), len(resRewards))

	// TODO: Print out current passes locally and support reading from file?
	return resRewards
}

func fetCurrentPasses(epoch ty.EpochId) (map[ty.VoterId]int, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/flare-foundation/fsp-rewards/refs/heads/main/%s/%d/passes.json", params.Net.Name, epoch)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch passes: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type jsonData struct {
		RewardEpochId int    `json:"rewardEpochId"`
		VoterAddress  string `json:"voterAddress"`
		Passes        int    `json:"passes"`
	}

	var raw []jsonData
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, err
	}

	passes := map[ty.VoterId]int{}
	for _, d := range raw {
		parsedEpoch := ty.EpochId(d.RewardEpochId)
		if parsedEpoch != epoch {
			return nil, fmt.Errorf("epoch mismatch: %d != %d", parsedEpoch, epoch)
		}
		voter := ty.VoterId(common.HexToAddress(d.VoterAddress))
		passes[voter] = d.Passes
	}
	return passes, nil
}

func PrintConditions(conditions map[*fsp.VoterInfo]MinConditions, epoch ty.EpochId) {
	byIdAddress := make(map[string]MinConditions)
	for voterInfo, condition := range conditions {
		byIdAddress[voterInfo.Identity.String()] = condition
	}
	jsonData, err := json.MarshalIndent(byIdAddress, "", "    ")
	if err != nil {
		logger.Error("Error serializing to JSON:", err)
		return
	}
	filePath := fmt.Sprintf("results/%s/%d/min-conditions.json", params.Net.Name, epoch)
	utils.WriteToFile(jsonData, filePath)
}

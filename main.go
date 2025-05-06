package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/rewards"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"gorm.io/gorm"
	"io"
	"math/big"
	"net/http"
	"time"
)

type ClientFlags struct {
	Network *string
	Epoch   *uint64

	DbHost *string
	DbPort *int
	DbName *string
	DbUser *string
	DbPass *string
}

func parseFlags() *ClientFlags {
	f := &ClientFlags{
		Network: flag.String("n", "coston", "Network"),
		Epoch:   flag.Uint64("e", 0, "Epoch number"),
		DbHost:  flag.String("h", "localhost", "Database host"),
		DbPort:  flag.Int("p", 3306, "Database port"),
		DbName:  flag.String("d", "flare_ftso_indexer", "Database name"),
		DbUser:  flag.String("u", "root", "Database user"),
		DbPass:  flag.String("w", "root", "Database password"),
	}
	flag.Parse()
	return f
}

func main() {
	flags := parseFlags()

	if *flags.Network == "" {
		flag.PrintDefaults()
		logger.Fatal("Network is required: coston, songbird, flare")
	}
	params.InitNetwork(*flags.Network)

	var epoch ty2.EpochId
	if *flags.Epoch == 0 {
		logger.Warn("Epoch number not provided, defaulting to last epoch")
		current, err := params.Net.Epoch.RewardEpochForTimeSec(uint64(time.Now().UnixMilli() / 1000))
		if err != nil {
			logger.Fatal("Error calculating epoch: %s", err)
		}
		epoch = ty2.EpochId(current - 1)
	} else {
		epoch = ty2.EpochId(*flags.Epoch)
	}

	db := getDb(flags)

	start := time.Now()
	res := calculateResults(db, epoch)
	printResults(res, "")

	elapsed := time.Since(start)
	logger.Info("Merkle root for epoch %d: %s, no weight based %d, duration: %s", epoch, res.MerkleRoot, res.NoOfWeightBasedClaims, elapsed)
}

func getDb(flags *ClientFlags) *gorm.DB {
	var config = database.Config{
		Host:     *flags.DbHost,
		Port:     *flags.DbPort,
		Database: *flags.DbName,
		Username: *flags.DbUser,
		Password: *flags.DbPass,
	}

	logger.Info("Connecting to database: +%v", config)
	db, err := database.Connect(&config)
	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}
	return db
}

func calculateResults(db *gorm.DB, epoch ty2.EpochId) epochResult {
	logger.Info("Calculating reward claims for epoch %d", epoch)

	allClaims, minCond := rewards.GetEpochClaims(db, epoch)
	PrintConditions(minCond, epoch)

	merged := rewards.MergeClaims(allClaims)
	logger.Info("Merged claims: %d, all claims %d", len(merged), len(allClaims))

	utils.PrintEpochClaims(merged, epoch, "all")
	finalClaims := rewards.ApplyPenalties(merged)
	utils.PrintEpochClaims(finalClaims, epoch, "merged")

	var resultClaims []ty.RewardClaim
	if params.Net.Name == "flare" || params.Net.Name == "songbird" {
		nonConditions := buildResults(epoch, finalClaims)
		printResults(nonConditions, "-raw")

		resultClaims = applyMinConditions(epoch, finalClaims, minCond)
	} else {
		resultClaims = finalClaims
	}

	return buildResults(epoch, resultClaims)
}

func applyMinConditions(epoch ty2.EpochId, merged []ty.RewardClaim, cond map[*fsp.VoterInfo]rewards.MinConditions) []ty.RewardClaim {
	currentPasses, err := fetCurrentPasses(epoch - 1)
	if err != nil {
		logger.Error("Error fetching current passes: %s, defaulting to 0 for all providers", err)
	}

	burnClaims := map[common.Address]bool{}
	burnClaims[rewards.BurnAddress] = true

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

		if passes < 0 {
			passes = 0
		}
		if passes > 3 {
			passes = 3
		}
	}

	var resRewards []ty.RewardClaim
	finalBurnClaim := ty.RewardClaim{
		Beneficiary: rewards.BurnAddress,
		Amount:      big.NewInt(0),
		Type:        ty.Direct,
	}

	for _, claim := range merged {
		if burnClaims[claim.Beneficiary] {
			finalBurnClaim.Amount.Add(finalBurnClaim.Amount, claim.Amount)
			logger.Info("Burn claim: %s %d %d", claim.Beneficiary.String(), claim.Amount, claim.Type)
		} else {
			resRewards = append(resRewards, claim)
		}
	}

	if finalBurnClaim.Amount.Cmp(rewards.BigZero) > 0 {
		resRewards = append([]ty.RewardClaim{finalBurnClaim}, resRewards...)
	}

	logger.Info("Original claims: %d, final claims: %d", len(merged), len(resRewards))

	// TODO: Print out current passes locally and support reading from file?
	return resRewards
}

func fetCurrentPasses(epoch ty2.EpochId) (map[ty2.VoterId]int, error) {
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

	passes := map[ty2.VoterId]int{}
	for _, d := range raw {
		parsedEpoch := ty2.EpochId(d.RewardEpochId)
		if parsedEpoch != epoch {
			return nil, fmt.Errorf("epoch mismatch: %d != %d", parsedEpoch, epoch)
		}
		voter := ty2.VoterId(common.HexToAddress(d.VoterAddress))
		passes[voter] = d.Passes
	}
	return passes, nil
}

func PrintConditions(conditions map[*fsp.VoterInfo]rewards.MinConditions, epoch ty2.EpochId) {
	byIdAddress := make(map[string]rewards.MinConditions)
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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/rewards"
	"fsp-rewards-calculator/ty"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"gorm.io/gorm"
	"io"
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

	var epoch ty.EpochId
	if *flags.Epoch == 0 {
		logger.Warn("Epoch number not provided, defaulting to last epoch")
		current, err := params.Net.Epoch.RewardEpochForTimeSec(uint64(time.Now().UnixMilli() / 1000))
		if err != nil {
			logger.Fatal("Error calculating epoch: %s", err)
		}
		epoch = ty.EpochId(current - 1)
	} else {
		epoch = ty.EpochId(*flags.Epoch)
	}

	db := getDb(flags)

	start := time.Now()
	res := calculateResults(db, epoch)
	printResults(res)

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

func calculateResults(db *gorm.DB, epoch ty.EpochId) epochResult {
	logger.Info("Calculating reward claims for epoch %d", epoch)

	allClaims, minCond := rewards.GetEpochClaims(db, epoch)

	merged := rewards.MergeClaims(allClaims)
	logger.Info("Merged claims: %d, all claims %d", len(merged), len(allClaims))

	resultClaims := applyMinConditions(epoch, merged, minCond)

	return buildResults(epoch, resultClaims)
}

func applyMinConditions(epoch ty.EpochId, merged []ty.RewardClaim, cond map[ty.VoterId]rewards.MinConditions) []ty.RewardClaim {
	currentPasses, err := fetCurrentPasses(epoch)
	if err != nil {
		logger.Error("Error fetching current passes: %s, defaulting to 0 for all providers", err)
	}

	claimByVoter := map[common.Address]*ty.RewardClaim{}
	for _, c := range merged {
		claimByVoter[c.Beneficiary] = &c
	}

	for voterId, c := range cond {
		passes := currentPasses[voterId] + c.PassDelta

		if passes < 0 {
			// burn claims for voter
			burnClaim := claimByVoter[rewards.BurnAddress]
			voterClaim := claimByVoter[common.Address(voterId)]

			if voterClaim != nil {
				burnClaim.Amount.Add(burnClaim.Amount, voterClaim.Amount)
				claimByVoter[common.Address(voterId)] = nil
			}
		}
	}

	var resRewards []ty.RewardClaim
	for _, c := range claimByVoter {
		if c != nil {
			resRewards = append(resRewards, *c)
		}
	}

	// TODO: Print out current passes locally and support reading from file?
	return resRewards
}

func fetCurrentPasses(epoch ty.EpochId) (map[ty.VoterId]int, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/flare-foundation/fsp-rewards/refs/heads/main/flare/%d/passes.json", epoch)

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

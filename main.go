package main

import (
	"encoding/json"
	"flare-common/database"
	"fmt"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/types"
	"os"
	"strconv"
)

func main() {
	config := database.DBConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "flare_ftso_indexer",
		Username: "root",
		Password: "root",
	}

	db, err := database.Connect(&config)

	os.Setenv("NETWORK", "coston")

	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}

	epoch := types.EpochId(2960)

	allClaims, err := calculateRewardClaims(db, epoch)
	if err != nil {
		logger.Fatal("Error calculating reward claims for epoch %d: %s", epoch, err)
		return
	}

	merged := mergeClaims(allClaims)
	logger.Info("Merged claims: %d, all claims %d", len(merged), len(allClaims))

	printResults(merged, strconv.Itoa(int(epoch)))
}

func printResults(records []RewardClaim, suffix string) {
	jsonData, err := json.MarshalIndent(records, "", "    ")
	if err != nil {
		fmt.Println("Error serializing to JSON:", err)
		return
	}

	file, err := os.Create(fmt.Sprintf("results/claims-%s.json", suffix))
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

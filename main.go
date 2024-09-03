package main

import (
	"encoding/json"
	"flare-common/database"
	"flare-common/merkle"
	"fmt"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/rewards"
	"ftsov2-rewarding/types"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"time"
)

var db *gorm.DB

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbPortInt, err := strconv.Atoi(dbPort)

	var config = database.DBConfig{
		Host:     dbHost,
		Port:     dbPortInt,
		Database: "flare_ftso_indexer",
		Username: "root",
		Password: "root",
	}

	logger.Info("Connecting to database: +%v", config)

	db, err = database.Connect(&config)
	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/rewards/{epoch}/{ignored...}", calculateRewardClaimsHandler)
	logger.Info("Starting server on port 8089")
	err = http.ListenAndServe(":8089", mux)
	if err != nil {
		logger.Fatal("Error starting server: %s", err)
	}
}

func calculateRewardClaimsHandler(w http.ResponseWriter, r *http.Request) {
	epochStr := r.PathValue("epoch")
	if epochStr == "" {
		http.Error(w, "Missing epoch parameter", http.StatusBadRequest)
		return
	}

	epochNo, err := strconv.Atoi(epochStr)
	if err != nil {
		http.Error(w, "Invalid epoch parameter", http.StatusBadRequest)
		return
	}

	epoch := types.EpochId(epochNo)

	start := time.Now()
	res := getResult(epoch, err)
	printEpochResult(res)
	elapsed := time.Since(start)
	logger.Info("Merkle root for epoch %d: %s, no weight based %d, duration: %s", epoch, res.MerkleRoot, res.NoOfWeightBasedClaims, elapsed)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		return
	}
}

func getResult(epoch types.EpochId, err error) roundResult {
	logger.Info("Calculating reward claims for epoch %d", epoch)

	allClaims, err := rewards.CalculateRewardClaims(db, epoch)
	if err != nil {
		logger.Fatal("Error calculating reward claims for epoch %d: %s", epoch, err)
	}

	merged := rewards.MergeClaims(allClaims)
	logger.Info("Merged claims: %d, all claims %d", len(merged), len(allClaims))

	utils.PrintResults(merged, strconv.Itoa(int(epoch)))

	finalClaims := rewards.ApplyPenalties(merged)
	utils.PrintResults(finalClaims, strconv.Itoa(int(epoch))+"-merged")

	var hashes []common.Hash
	var weightBasedClaims = 0
	for _, claim := range finalClaims {
		if claim.Type == types.WNat || claim.Type == types.Mirror || claim.Type == types.CChain {
			weightBasedClaims++
		}
		hash := utils.RewardClaimHash(epoch, claim)
		hashes = append(hashes, hash)
		logger.Info("Claim: %s, claim +v", hash.Hex(), claim)
	}

	merkleTree := merkle.Build(hashes, false)

	root, err := merkleTree.Root()
	if err != nil {
		logger.Fatal("Error calculating merkle root: %s", err)
	}

	res := roundResult{
		RewardEpochId:         int(epoch),
		NoOfWeightBasedClaims: weightBasedClaims,
		MerkleRoot:            root.Hex(),
	}
	return res
}

type roundResult struct {
	RewardEpochId         int    `json:"rewardEpochId"`
	NoOfWeightBasedClaims int    `json:"noOfWeightBasedClaims"`
	MerkleRoot            string `json:"merkleRoot"`
}

func printEpochResult(result roundResult) {
	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Println("Error serializing to JSON:", err)
		return
	}

	file, err := os.Create(fmt.Sprintf("results/result-%d.json", result.RewardEpochId))
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

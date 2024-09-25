package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/rewards"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/flarenetwork/libs/go-flare-common/pkg/database"
	"gitlab.com/flarenetwork/libs/go-flare-common/pkg/merkle"
	"gorm.io/gorm"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var db *gorm.DB

func main() {
	epochFlag := flag.Uint64("e", 0, "Epoch number")
	flag.Parse()

	if *epochFlag == 0 {
		flag.PrintDefaults()
		logger.Fatal("Epoch number is required")
	}

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

	epoch := ty.EpochId(*epochFlag)

	start := time.Now()
	res := getEpochResult(epoch, err)
	printEpochResult(res)
	elapsed := time.Since(start)
	logger.Info("Merkle root for epoch %d: %s, no weight based %d, duration: %s", epoch, res.MerkleRoot, res.NoOfWeightBasedClaims, elapsed)
}

func getEpochResult(epoch ty.EpochId, err error) epochResult {
	logger.Info("Calculating reward claims for epoch %d", epoch)

	allClaims, err := rewards.GetEpochClaims(db, epoch)
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
		if claim.Type == ty.WNat || claim.Type == ty.Mirror || claim.Type == ty.CChain {
			weightBasedClaims++
		}
		hash := utils.RewardClaimHash(epoch, claim)
		hashes = append(hashes, hash)
		logger.Info("Claim: %s, claim +v", hash.Hex(), claim)
	}

	merkleTree := merkle.Build(hashes, false)

	root, err := merkleTree.Root()

	var cl []ClaimWithProof
	for i := range finalClaims {
		claim := finalClaims[i]

		proof, err := merkleTree.GetProofFromHash(hashes[i])
		if err != nil {
			logger.Fatal("Error calculating proof: %s", err)
		}

		proofHex := make([]string, len(proof))
		for i, p := range proof {
			proofHex[i] = p.Hex()
		}

		cl = append(cl, ClaimWithProof{
			Proof: proof,
			Claim: claim2{
				Beneficiary: claim.Beneficiary,
				Amount:      claim.Amount,
				Type:        int(claim.Type),
				Epoch:       epoch,
			},
		})
	}

	if err != nil {
		logger.Fatal("Error calculating merkle root: %s", err)
	}

	res := epochResult{
		RewardEpochId:         int(epoch),
		NoOfWeightBasedClaims: weightBasedClaims,
		MerkleRoot:            root.Hex(),
		RewardClaims:          cl,
	}
	return res
}

type claim2 struct {
	Beneficiary common.Address `json:"beneficiary"`
	Amount      *big.Int       `json:"amount"`
	Type        int            `json:"claimType"`
	Epoch       ty.EpochId     `json:"rewardEpochId"`
}

type ClaimWithProof struct {
	Proof []common.Hash `json:"merkleProof"`
	Claim claim2        `json:"body"`
}

type epochResult struct {
	RewardEpochId         int              `json:"rewardEpochId"`
	NoOfWeightBasedClaims int              `json:"noOfWeightBasedClaims"`
	MerkleRoot            string           `json:"merkleRoot"`
	RewardClaims          []ClaimWithProof `json:"rewardClaims"`
}

func printEpochResult(result epochResult) {
	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Println("Error serializing to JSON:", err)
		return
	}

	filePath := fmt.Sprintf("results/%s/result-%d.json", os.Getenv("NETWORK"), result.RewardEpochId)

	dir := filepath.Dir(filePath)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating folders:", err)
	}

	file, err := os.Create(filePath)
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

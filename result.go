package main

import (
	"encoding/json"
	"fmt"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/rewards"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"gitlab.com/flarenetwork/libs/go-flare-common/pkg/merkle"
	"math/big"
)

type epochResult struct {
	RewardEpochId         int              `json:"rewardEpochId"`
	NoOfWeightBasedClaims int              `json:"noOfWeightBasedClaims"`
	MerkleRoot            string           `json:"merkleRoot"`
	RewardClaims          []claimWithProof `json:"rewardClaims"`
}

type claimWithProof struct {
	Proof []common.Hash `json:"merkleProof"`
	Claim outputClaim   `json:"body"`
}

type outputClaim struct {
	Beneficiary common.Address `json:"beneficiary"`
	Amount      *big.Int       `json:"amount"`
	Type        int            `json:"claimType"`
	Epoch       ty.EpochId     `json:"rewardEpochId"`
}

func buildResults(epoch ty.EpochId, claims []ty.RewardClaim) epochResult {
	utils.PrintEpochClaims(claims, epoch, "all")
	finalClaims := rewards.ApplyPenalties(claims)
	utils.PrintEpochClaims(finalClaims, epoch, "merged")

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

	var cl []claimWithProof
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

		cl = append(cl, claimWithProof{
			Proof: proof,
			Claim: outputClaim{
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

func printResults(result epochResult) {
	filePath := fmt.Sprintf("results/%s/%d/result.json", params.Net.Name, result.RewardEpochId)

	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Println("Error serializing to JSON:", err)
		return
	}
	utils.WriteToFile(jsonData, filePath)
}

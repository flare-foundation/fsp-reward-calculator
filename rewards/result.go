package rewards

import (
	"encoding/json"
	"fmt"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/merkle"
)

type EpochResult struct {
	RewardEpochId         int              `json:"rewardEpochId"`
	NoOfWeightBasedClaims int              `json:"noOfWeightBasedClaims"`
	MerkleRoot            string           `json:"merkleRoot"`
	RewardClaims          []ClaimWithProof `json:"rewardClaims"`
}

type ClaimWithProof struct {
	Proof []common.Hash `json:"merkleProof"`
	Claim outputClaim   `json:"body"`
}

type outputClaim struct {
	Beneficiary common.Address    `json:"beneficiary"`
	Amount      *big.Int          `json:"amount"`
	Type        int               `json:"claimType"`
	Epoch       ty2.RewardEpochId `json:"rewardEpochId"`
}

func buildResults(epoch ty2.RewardEpochId, finalClaims []ty.RewardClaim) EpochResult {
	var hashes []common.Hash
	var weightBasedClaims = 0
	for _, claim := range finalClaims {
		if claim.Type == ty.WNat || claim.Type == ty.Mirror || claim.Type == ty.CChain {
			weightBasedClaims++
		}
		hash := utils.RewardClaimHash(epoch, claim)
		hashes = append(hashes, hash)
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

	res := EpochResult{
		RewardEpochId:         int(epoch),
		NoOfWeightBasedClaims: weightBasedClaims,
		MerkleRoot:            root.Hex(),
		RewardClaims:          cl,
	}
	return res
}

func PrintResults(result EpochResult, suffix string) {
	filePath := fmt.Sprintf("results/%s/%d/result%s.json", params.Net.Name, result.RewardEpochId, suffix)

	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Println("Error serializing to JSON:", err)
		return
	}
	utils.WriteToFile(jsonData, filePath)
}

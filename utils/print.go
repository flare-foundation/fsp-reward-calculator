package utils

import (
	"encoding/json"
	"fmt"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"os"
	"path/filepath"
)

func PrintEpochClaims(records []ty.RewardClaim, epoch ty2.RewardEpochId, suffix string) {
	jsonData, err := json.MarshalIndent(records, "", "    ")
	if err != nil {
		logger.Error("Error serializing to JSON:", err)
		return
	}
	filePath := fmt.Sprintf("results/%s/%d/claims-%s.json", params.Net.Name, epoch, suffix)
	WriteToFile(jsonData, filePath)
}

func PrintRoundResults(records []ty.RewardClaim, epoch ty2.RewardEpochId, round ty2.RoundId, suffix string) {
	jsonData, err := json.MarshalIndent(records, "", "    ")
	if err != nil {
		logger.Error("Error serializing to JSON:", err)
		return
	}
	filePath := fmt.Sprintf("results/%s/%d/rounds/%d/claims-%s.json", params.Net.Name, epoch, round, suffix)
	WriteToFile(jsonData, filePath)
}

func WriteToFile(jsonData []byte, filePath string) {
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		logger.Error("Error creating folders:", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		logger.Error("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		logger.Error("Error writing to file:", err)
		return
	}
}

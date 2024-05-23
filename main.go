package main

import (
	"flare-common/contracts/registry"
	"flare-common/database"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/utils"
	"time"
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

	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}

	topic0, err := utils.EventIDFromMetadata(registry.RegistryMetaData, "VoterRegistered")
	if err != nil {
		logger.Fatal("Error getting VoterRegistered event: %s", err)
	}

	currentTimestamp := time.Now().Unix()

	res, err := database.FetchLogsByAddressAndTopic0Timestamp(db, "0x051E9Cb16A8676C011faa10efA1ABE95372e7825", topic0, currentTimestamp-3600*12, currentTimestamp)
	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}

	logger.Info("Logs: %v", res)
}

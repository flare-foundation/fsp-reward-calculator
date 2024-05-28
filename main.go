package main

import (
	"flare-common/database"
	"ftsov2-rewarding/logger"
	"github.com/pkg/errors"

	"gorm.io/gorm"
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

	calculateRewards(db, 2690)
}

type RewardClaim struct {
}

func calculateRewards(db *gorm.DB, epoch int) (RewardClaim, error) {
	_, err := getRewardEpoch(epoch, db)
	if err != nil {
		return RewardClaim{}, errors.Errorf("err fetching reward epoch: %s", err)
	}

	return RewardClaim{}, nil
}

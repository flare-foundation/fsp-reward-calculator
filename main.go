package main

import (
	"flag"
	"fmt"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/rewards"
	"fsp-rewards-calculator/utils"
	"time"

	"github.com/flare-foundation/go-flare-common/pkg/database"
	"gorm.io/gorm"
)

type ClientFlags struct {
	Network *string
	Epoch   *uint64

	DbHost *string
	DbPort *int
	DbName *string
	DbUser *string
	DbPass *string

	Verbose *bool
}

func main() {
	start := time.Now()
	logger.Info("Starting FSP rewards calculation")

	flags, err := parseFlags()
	if err != nil {
		logger.Fatal("Configuration error: %v", err)
	}

	utils.SetVerbose(*flags.Verbose)

	logger.Info("Configuration: network=%s, epoch=%d, indexer_db=%s:%d",
		*flags.Network, *flags.Epoch, *flags.DbHost, *flags.DbPort)

	params.InitNetwork(*flags.Network)

	var epoch ty.RewardEpochId
	if *flags.Epoch == 0 {
		current, err := params.Net.Epoch.RewardEpochForTimeSec(uint64(time.Now().UnixMilli() / 1000))
		if err != nil {
			logger.Fatal("Error calculating epoch: %s", err)
		}
		epoch = current - 1
		logger.Info("Epoch number not provided, defaulting to last epoch: %d", epoch)
	} else {
		epoch = ty.RewardEpochId(*flags.Epoch)
	}

	db := getDb(flags)
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	res := rewards.CalculateResults(db, epoch)
	rewards.PrintResults(res, "")

	logger.Info(
		"Calculation completed successfully - Epoch: %d, Hash: %s, Weight-based claims: %d, Total Time: %s",
		epoch, res.MerkleRoot, res.NoOfWeightBasedClaims, time.Since(start),
	)
}

func (f *ClientFlags) validate() error {
	if *f.Network == "" {
		return fmt.Errorf("network is required: must be one of coston, songbird, flare")
	}

	validNetworks := map[string]bool{"coston": true, "songbird": true, "flare": true}
	if !validNetworks[*f.Network] {
		return fmt.Errorf("invalid network '%s': must be one of coston, songbird, flare", *f.Network)
	}

	if *f.DbPort <= 0 || *f.DbPort > 65535 {
		return fmt.Errorf("invalid database port %d: must be between 1 and 65535", *f.DbPort)
	}

	return nil
}

func parseFlags() (*ClientFlags, error) {
	f := &ClientFlags{
		Network: flag.String("n", "", "Network (coston, songbird, flare)"),
		Epoch:   flag.Uint64("e", 0, "Epoch number"),
		DbHost:  flag.String("h", "localhost", "Database host"),
		DbPort:  flag.Int("p", 3306, "Database port"),
		DbName:  flag.String("d", "flare_ftso_indexer", "Database name"),
		DbUser:  flag.String("u", "root", "Database user"),
		DbPass:  flag.String("w", "root", "Database password"),
		Verbose: flag.Bool("v", false, "Verbose output - enable writing per-round claim data"),
	}
	flag.Parse()

	if err := f.validate(); err != nil {
		flag.PrintDefaults()
		return nil, err
	}

	return f, nil
}

func getDb(flags *ClientFlags) *gorm.DB {
	var config = database.Config{
		Host:     *flags.DbHost,
		Port:     *flags.DbPort,
		Database: *flags.DbName,
		Username: *flags.DbUser,
		Password: *flags.DbPass,
	}

	db, err := database.Connect(&config)
	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}
	return db
}

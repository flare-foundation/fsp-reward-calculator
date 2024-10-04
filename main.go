package main

import (
	"flag"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/rewards"
	"fsp-rewards-calculator/ty"
	"gitlab.com/flarenetwork/libs/go-flare-common/pkg/database"
	"gorm.io/gorm"
	"time"
)

type ClientFlags struct {
	Network string
	Epoch   uint64

	DbHost string
	DbPort int
	DbName string
	DbUser string
	DbPass string
}

func parseFlags() *ClientFlags {
	flag.Parse()

	return &ClientFlags{
		Network: *flag.String("n", "coston", "Network"),
		Epoch:   *flag.Uint64("e", 0, "Epoch number"),
		DbHost:  *flag.String("h", "localhost", "Database host"),
		DbPort:  *flag.Int("p", 3306, "Database port"),
		DbName:  *flag.String("d", "flare_ftso_indexer", "Database name"),
		DbUser:  *flag.String("u", "root", "Database user"),
		DbPass:  *flag.String("p", "root", "Database password"),
	}
}

func main() {
	flags := parseFlags()

	if flags.Network == "" {
		flag.PrintDefaults()
		logger.Fatal("Network is required: coston, songbird, flare")
	}
	params.InitNetwork(flags.Network)

	if flags.Epoch == 0 {
		flag.PrintDefaults()
		logger.Fatal("Epoch number is required")
	}

	db := getDb(flags)

	epoch := ty.EpochId(flags.Epoch)

	start := time.Now()
	res := calculateResults(db, epoch)
	printResults(res)

	elapsed := time.Since(start)
	logger.Info("Merkle root for epoch %d: %s, no weight based %d, duration: %s", epoch, res.MerkleRoot, res.NoOfWeightBasedClaims, elapsed)
}

func getDb(flags *ClientFlags) *gorm.DB {
	var config = database.DBConfig{
		Host:     flags.DbHost,
		Port:     flags.DbPort,
		Database: flags.DbName,
		Username: flags.DbUser,
		Password: flags.DbPass,
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

	allClaims, err := rewards.GetEpochClaims(db, epoch)
	if err != nil {
		logger.Fatal("Error calculating reward claims for epoch %d: %s", epoch, err)
	}

	merged := rewards.MergeClaims(allClaims)
	logger.Info("Merged claims: %d, all claims %d", len(merged), len(allClaims))

	return buildResults(epoch, merged)
}

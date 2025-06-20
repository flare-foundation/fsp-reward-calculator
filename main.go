package main

import (
	"flag"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/rewards"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"gorm.io/gorm"
	"time"
)

type ClientFlags struct {
	Network *string
	Epoch   *uint64

	DbHost *string
	DbPort *int
	DbName *string
	DbUser *string
	DbPass *string
}

func parseFlags() *ClientFlags {
	f := &ClientFlags{
		Network: flag.String("n", "coston", "Network"),
		Epoch:   flag.Uint64("e", 0, "Epoch number"),
		DbHost:  flag.String("h", "localhost", "Database host"),
		DbPort:  flag.Int("p", 3306, "Database port"),
		DbName:  flag.String("d", "flare_ftso_indexer", "Database name"),
		DbUser:  flag.String("u", "root", "Database user"),
		DbPass:  flag.String("w", "root", "Database password"),
	}
	flag.Parse()
	return f
}

func main() {
	// Start CPU profiling
	//f, err := os.Create("cpu.prof")
	//if err != nil {
	//	logger.Fatal("could not create CPU profile: %s", err)
	//}
	//defer f.Close()
	//if err := pprof.StartCPUProfile(f); err != nil {
	//	logger.Fatal("could not start CPU profile: %s", err)
	//}
	//defer pprof.StopCPUProfile()

	//analytics.RunAnalytics()

	flags := parseFlags()

	if *flags.Network == "" {
		flag.PrintDefaults()
		logger.Fatal("Network is required: coston, songbird, flare")
	}
	params.InitNetwork(*flags.Network)

	var epoch ty2.EpochId
	if *flags.Epoch == 0 {
		logger.Warn("Epoch number not provided, defaulting to last epoch")
		current, err := params.Net.Epoch.RewardEpochForTimeSec(uint64(time.Now().UnixMilli() / 1000))
		if err != nil {
			logger.Fatal("Error calculating epoch: %s", err)
		}
		epoch = ty2.EpochId(current - 1)
	} else {
		epoch = ty2.EpochId(*flags.Epoch)
	}

	db := getDb(flags)

	start := time.Now()
	res := rewards.CalculateResults(db, epoch)
	rewards.PrintResults(res, "")

	elapsed := time.Since(start)
	logger.Info("Merkle root for epoch %d: %s, no weight based %d, duration: %s", epoch, res.MerkleRoot, res.NoOfWeightBasedClaims, elapsed)
}

func getDb(flags *ClientFlags) *gorm.DB {
	var config = database.Config{
		Host:     *flags.DbHost,
		Port:     *flags.DbPort,
		Database: *flags.DbName,
		Username: *flags.DbUser,
		Password: *flags.DbPass,
	}

	logger.Info("Connecting to database: +%v", config)
	db, err := database.Connect(&config)
	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}
	return db
}

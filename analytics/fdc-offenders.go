package analytics

import (
	"fmt"
	"fsp-rewards-calculator/common/fdc"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/rewards"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"os"
	"time"
)

func RunAnalytics() {
	cfg := database.Config{
		Host:     "localhost",
		Port:     3336,
		Database: "flare_ftso_indexer",
		Username: "root",
		Password: "root",
	}
	// Initialize database connection (replace with actual DB initialization)
	db, err := database.Connect(&cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database: %s", err)
	}

	epoch := 288

	params.InitNetwork("flare")

	epochs, err := fsp.GetRewardEpoch(ty.EpochId(epoch), db)
	if err != nil {
		logger.Fatal("Error fetching reward epoch: %s", err)
	}

	re := &epochs

	currentRound := params.Net.Epoch.VotingRoundForTimeSec(uint64(time.Now().Unix()))
	re.EndRound = currentRound - 2

	// Fetch required data for FDC claims
	submit2, err := fsp.GetSubmit2(db, re.StartRound, re.EndRound)
	if err != nil {
		logger.Fatal("Error fetching submit2: %s", err)
	}

	submitSignatures, err := fsp.GetSubmitSignatures(db, re.StartRound, re.EndRound)
	if err != nil {
		logger.Fatal("Error fetching submitSignatures: %s", err)
	}

	finalizations, err := fsp.GetFinalizations(db, re, re.StartRound, re.EndRound)
	if err != nil {
		logger.Fatal("Error fetching finalizations: %s", err)
	}

	// Calculate FDC claims for the current epoch
	fdcClaims, _ := rewards.GetFdcRewards(db, re, submit2[fdc.ProtocolId], submitSignatures[fdc.ProtocolId], finalizations[fdc.ProtocolId])

	cl := rewards.MergeClaims(fdcClaims)

	// Output the FDC claims
	fmt.Println("FDC Claims for Current Epoch:", cl)

	os.Exit(1)
}

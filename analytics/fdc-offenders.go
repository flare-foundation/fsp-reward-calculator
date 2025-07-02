package analytics

import (
	"fmt"
	"fsp-rewards-calculator/common/fdc"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/ftso"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/rewards"
	"fsp-rewards-calculator/utils"
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

	epoch := 305

	params.InitNetwork("flare")

	epochs, err := fsp.GetRewardEpoch(ty.RewardEpochId(epoch), db)
	if err != nil {
		logger.Fatal("Error fetching reward epoch: %s", err)
	}

	re := &epochs

	currentRound := ty.RoundId(params.Net.Epoch.VotingEpochForTimeSec(uint64(time.Now().Unix())))
	re.EndRound = currentRound - 2

	submit1, err := fsp.GetSubmit1(db, re.StartRound, re.EndRound)
	if err != nil {
		logger.Fatal("Error fetching submit2: %s", err)
	}
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

	echs := rewards.RewardEpochs{
		Current: re,
	}
	// Calculate FDC claims for the current epoch
	ftsoClaims, _ := rewards.GetFtsoRewards(db, echs, re.EndRound, submit1[ftso.ProtocolId], submit2[ftso.ProtocolId], submitSignatures[ftso.ProtocolId], finalizations[ftso.ProtocolId])
	fdcClaims, _ := rewards.GetFdcRewards(db, re, submit2[fdc.ProtocolId], submitSignatures[fdc.ProtocolId], finalizations[fdc.ProtocolId])

	cl := rewards.MergeClaims(append(fdcClaims, ftsoClaims...))
	cl = rewards.ApplyPenalties(cl)

	utils.PrintEpochClaims(cl, re.Epoch, "final-claims")

	// Output the FDC claims
	fmt.Println("FDC Claims for Current Epoch:", cl)

	os.Exit(1)
}

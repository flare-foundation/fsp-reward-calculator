package rewards

import (
	"context"
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"gorm.io/gorm"
)

// GetEpochClaims calculates the reward claims for a reward epoch
func GetEpochClaims(db *gorm.DB, epoch ty.EpochId) ([]ty.RewardClaim, map[*data.VoterInfo]MinConditions) {
	epochs, err := data.LoadRewardEpochs(epoch, db)
	if err != nil {
		logger.Fatal("Error fetching reward epochs, required data may not be indexed: %s", err)
	}

	re := epochs.Current
	windowStart := ty.RoundId(uint64(re.StartRound) - params.Net.Ftso.RandomGenerationBenchingWindow)
	windowEnd := re.EndRound.Add(params.Net.Ftso.FutureSecureRandomWindow)

	ensureDataRange(db, windowStart, windowEnd)

	submit1, err := data.GetSubmit1(db, windowStart, windowEnd)
	if err != nil {
		logger.Fatal("error fetching submit1")
	}

	submit2, err := data.GetSubmit2(db, windowStart, windowEnd)
	if err != nil {
		logger.Fatal("error fetching submit2")
	}
	submitSignatures, err := data.GetSubmitSignatures(db, re.StartRound, re.EndRound)
	if err != nil {
		logger.Fatal("error fetching submitSignatures")
	}

	finalizations, err := data.GetFinalizations(db, re, re.StartRound, re.EndRound)
	if err != nil {
		logger.Fatal("error fetching finalizations")
	}
	logger.Info("Finalizations fetched")

	epochClaims := make([]ty.RewardClaim, 0)

	ftsoClaims, ftsoCond := getFtsoRewards(db, epochs, windowEnd, submit1[data.FtsoProtocolId], submit2[data.FtsoProtocolId], submitSignatures[data.FtsoProtocolId], finalizations[data.FtsoProtocolId])
	fdcClaims, fdcCond := GetFdcRewards(db, epochs.Current, submit2[data.FdcProtocolId], submitSignatures[data.FdcProtocolId], finalizations[data.FdcProtocolId])

	epochClaims = append(epochClaims, ftsoClaims...)
	epochClaims = append(epochClaims, fdcClaims...)

	cond := calcConditions(epoch, re.VoterIndex, ftsoCond, fdcCond)

	return epochClaims, cond
}

func ensureDataRange(db *gorm.DB, start, end ty.RoundId) {
	startSec := params.Net.Epoch.VotingRoundStartSec(start)
	endSec := params.Net.Epoch.VotingRoundEndSec(end)

	firstSec, err := database.FetchFirstDBBlockTs(context.Background(), db)
	if err != nil {
		logger.Fatal("Error fetching first block timestamp: %s", err)
	}
	if firstSec > startSec {
		logger.Fatal("First block timestamp %d is greater than required start %d", firstSec, startSec)
	}

	lastSec, err := database.FetchLastDBBlockTs(context.Background(), db)
	if err != nil {
		logger.Fatal("Error fetching last block timestamp: %s", err)
	}
	if lastSec < endSec {
		logger.Fatal("Last required round %d not indexed in DB. Last block timestamp from DB: %d, required at least %d.", end, lastSec, endSec)
	}
}

func calcConditions(epoch ty.EpochId, voters *data.VoterIndex, conditions FtsoMinConditions, fdcCond map[ty.VoterId]bool) map[*data.VoterInfo]MinConditions {
	stakingCond := map[ty.VoterId]StakingCondition{}
	if params.Net.Name == "flare" {
		stakingCond = MetStakingCondition(epoch, voters)
	}

	cond := map[*data.VoterInfo]MinConditions{}

	for _, voter := range voters.PolicyOrder {
		c := MinConditions{}

		if params.Net.Name != "flare" {
			stakingCond[voter.Identity] = Met
		}

		if !conditions.FastUpdates[voter.Identity] {
			c.PassDelta--
		} else {
			c.MetFU = true
		}
		if !conditions.Scaling[voter.Identity] {
			c.PassDelta--
		} else {
			c.MetFtso = true
		}
		if !fdcCond[voter.Identity] {
			c.PassDelta--
		} else {
			c.MetFdc = true
		}
		if stakingCond[voter.Identity] == NotMet {
			c.PassDelta--
		} else {
			c.MetStaking = true
		}

		if c.PassDelta == 0 { // No penalties
			if stakingCond[voter.Identity] == Met { // No extra pass if MetNoPass
				c.PassDelta = 1
			}
		}

		cond[voter] = c
	}

	return cond
}

type MinConditions struct {
	MetFtso    bool
	MetFU      bool
	MetFdc     bool
	MetStaking bool
	PassDelta  int
}

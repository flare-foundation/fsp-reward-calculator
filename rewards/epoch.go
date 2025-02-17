package rewards

import (
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"gorm.io/gorm"
)

// GetEpochClaims calculates the reward claims for a reward epoch
func GetEpochClaims(db *gorm.DB, epoch ty.EpochId) ([]ty.RewardClaim, map[ty.VoterId]MinConditions) {
	epochs, err := data.LoadRewardEpochs(epoch, db)
	if err != nil {
		logger.Fatal("error fetching reward epochs", err)
	}

	re := epochs.Current
	windowStart := ty.RoundId(uint64(re.StartRound) - params.Net.Ftso.RandomGenerationBenchingWindow)
	windowEnd := re.EndRound.Add(params.Net.Ftso.FutureSecureRandomWindow)

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
	fdcClaims := getFdcRewards(db, epochs, submit2[data.FdcProtocolId], submitSignatures[data.FdcProtocolId], finalizations[data.FdcProtocolId])

	epochClaims = append(epochClaims, ftsoClaims...)
	epochClaims = append(epochClaims, fdcClaims...)

	cond := calcConditions(epoch, re.VoterIndex, ftsoCond)

	return epochClaims, cond
}

func calcConditions(epoch ty.EpochId, voters *data.VoterIndex, conditions FtsoMinConditions) map[ty.VoterId]MinConditions {
	stakingCond := MetStakingContiion(epoch, voters)

	cond := map[ty.VoterId]MinConditions{}

	for _, voter := range voters.PolicyOrder {
		c := MinConditions{}

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

		cond[voter.Identity] = c
	}

	return cond
}

type MinConditions struct {
	MetFtso    bool
	MetFU      bool
	MetStaking bool
	PassDelta  int
}

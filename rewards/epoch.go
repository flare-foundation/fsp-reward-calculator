package rewards

import (
	"context"
	"fsp-rewards-calculator/common/fdc"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/ftso"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// GetEpochClaims calculates the reward claims for a reward epoch
func GetEpochClaims(db *gorm.DB, epoch ty2.RewardEpochId) ([]ty.RewardClaim, map[*fsp.VoterInfo]MinConditions) {
	epochs, err := LoadRewardEpochs(epoch, db)
	if err != nil {
		logger.Fatal("Error fetching reward epochs, required data may not be indexed: %s", err)
	}

	re := epochs.Current
	logger.Info("Epoch round range: %d-%d", re.StartRound, re.EndRound)

	windowStart := ty2.RoundId(uint64(re.StartRound) - params.Net.Ftso.RandomGenerationBenchingWindow)
	windowEnd := re.EndRound.Add(params.Net.Ftso.FutureSecureRandomWindow)

	ensureDataRange(db, windowStart, windowEnd)

	logger.Info("Fetching submission data, this will take a while...")
	submit1, err := fsp.GetSubmit1(db, windowStart, windowEnd)
	if err != nil {
		logger.Fatal("error fetching submit1")
	}

	submit2, err := fsp.GetSubmit2(db, windowStart, windowEnd)
	if err != nil {
		logger.Fatal("error fetching submit2")
	}
	submitSignatures, err := fsp.GetSubmitSignatures(db, re.StartRound, re.EndRound)
	if err != nil {
		logger.Fatal("error fetching submitSignatures")
	}

	finalizations, err := fsp.GetFinalizations(db, re, re.StartRound, re.EndRound)
	if err != nil {
		logger.Fatal("error fetching finalizations")
	}
	logger.Info("Finalizations fetched")

	logger.Info("Done fetching submission data, calculating rewards...")

	epochClaims := make([]ty.RewardClaim, 0)

	logger.Info("Calculating FTSO rewards")
	ftsoClaims, ftsoCond := GetFtsoRewards(db, epochs, windowEnd, submit1[ftso.ProtocolId], submit2[ftso.ProtocolId], submitSignatures[ftso.ProtocolId], finalizations[ftso.ProtocolId])
	logger.Info("Calculating FDC rewards")
	fdcClaims, fdcCond := GetFdcRewards(db, epochs.Current, submit2[fdc.ProtocolId], submitSignatures[fdc.ProtocolId], finalizations[fdc.ProtocolId])

	epochClaims = append(epochClaims, ftsoClaims...)
	epochClaims = append(epochClaims, fdcClaims...)

	cond := calcConditions(epoch, re.VoterIndex, ftsoCond, fdcCond)

	return epochClaims, cond
}

func ensureDataRange(db *gorm.DB, start, end ty2.RoundId) {
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

func calcConditions(epoch ty2.RewardEpochId, voters *fsp.VoterIndex, conditions FtsoMinConditions, fdcCond map[ty2.VoterId]bool) map[*fsp.VoterInfo]MinConditions {
	stakingCond := map[ty2.VoterId]StakingCondition{}
	if params.Net.Name == "flare" {
		stakingCond = MetStakingCondition(epoch, voters)
	}

	cond := map[*fsp.VoterInfo]MinConditions{}

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

type RewardEpochs struct {
	prev    *fsp.RewardEpoch
	Current *fsp.RewardEpoch
	next    *fsp.RewardEpoch
}

func LoadRewardEpochs(epoch ty2.RewardEpochId, db *gorm.DB) (RewardEpochs, error) {
	prev, err := fsp.GetRewardEpoch(epoch-1, db)
	if err != nil {
		return RewardEpochs{}, errors.Wrap(err, "error fetching previous epoch")
	}
	current, err := fsp.GetRewardEpoch(epoch, db)
	if err != nil {
		return RewardEpochs{}, errors.Wrap(err, "error fetching current epoch")
	}
	next, err := fsp.GetRewardEpoch(epoch+1, db)
	if err != nil {
		return RewardEpochs{}, errors.Wrap(err, "error fetching next epoch")
	}
	return RewardEpochs{
		prev:    &prev,
		Current: &current,
		next:    &next,
	}, nil
}

func (re *RewardEpochs) EpochForRound(round ty2.RoundId) *fsp.RewardEpoch {
	switch {
	case round < re.Current.StartRound:
		return re.prev
	case round > re.Current.EndRound:
		return re.next
	default:
		return re.Current
	}
}

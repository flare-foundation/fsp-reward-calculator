package rewards

import (
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// GetEpochClaims calculates the reward claims for a reward epoch
func GetEpochClaims(db *gorm.DB, epoch ty.EpochId) ([]ty.RewardClaim, error) {
	epochs, err := data.LoadRewardEpochs(epoch, db)
	if err != nil {
		return nil, errors.Wrap(err, "err fetching reward epochs")
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

	ftsoClaims, _ := getFtsoRewards(db, epochs, windowEnd, submit1[data.FtsoProtocolId], submit2[data.FtsoProtocolId], submitSignatures[data.FtsoProtocolId], finalizations[data.FtsoProtocolId])
	fdcClaims, _ := getFdcRewards(db, epochs, submit2[data.FdcProtocolId], submitSignatures[data.FdcProtocolId], finalizations[data.FdcProtocolId])

	epochClaims = append(epochClaims, ftsoClaims...)
	epochClaims = append(epochClaims, fdcClaims...)
	return epochClaims, nil
}

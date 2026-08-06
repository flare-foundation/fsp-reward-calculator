package fsp

import (
	common2 "fsp-rewards-calculator/common"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// FundingWindow is the time interval during which fees are credited to a reward epoch on chain.
// Both edges are inclusive.
type FundingWindow struct {
	StartSec uint64
	EndSec   uint64
}

// RewardEpochFundingWindow returns the window during which fees are credited to a reward epoch.
//
// RewardManager.receiveRewards is called with getCurrentRewardEpochId(), which flips exactly when
// RewardEpochStarted is emitted. That moment lags the voting round schedule by an unbounded amount, so
// the schedule cannot be used to decide which epoch a fee funded. The window is bounded by the two
// RewardEpochStarted events instead, and both edges are inclusive: timestamps have one second
// granularity and are shared by every event in a block, so an inclusive window is a superset that
// cannot miss an event. Callers narrow it exactly by the reward epoch id the events carry themselves.
//
// Fails if the next epoch has not started, since its start is what closes the window; the funding of a
// reward epoch is not final until then.
func RewardEpochFundingWindow(db *gorm.DB, epoch ty.RewardEpochId) (FundingWindow, error) {
	starts, err := rewardEpochStartTimestamps(db, epoch)
	if err != nil {
		return FundingWindow{}, err
	}

	start, ok := starts[epoch]
	if !ok {
		return FundingWindow{}, errors.Errorf(
			"no RewardEpochStarted event for reward epoch %d, cannot attribute fees to it", epoch,
		)
	}
	next, ok := starts[epoch+1]
	if !ok {
		return FundingWindow{}, errors.Errorf(
			"no RewardEpochStarted event for reward epoch %d: reward epoch %d is not closed yet, so the fees credited to it are not final",
			epoch+1, epoch,
		)
	}

	return FundingWindow{StartSec: start, EndSec: next}, nil
}

// rewardEpochStartTimestamps returns the timestamps of the RewardEpochStarted events of the given epoch
// and the one after it, keyed by reward epoch id. Events are searched in a generous interval around the
// scheduled epoch start, because the actual start can lag the schedule.
func rewardEpochStartTimestamps(db *gorm.DB, epoch ty.RewardEpochId) (map[ty.RewardEpochId]uint64, error) {
	epochDuration := params.Net.Epoch.RewardEpochDurationInVotingEpochs * params.Net.Epoch.VotingEpochDurationSeconds
	searchIntervalStartSec := params.Net.Epoch.ExpectedRewardEpochStartTimeSec(epoch) - 2*epochDuration
	searchIntervalEndSec := params.Net.Epoch.ExpectedRewardEpochStartTimeSec(epoch+1) + 2*epochDuration

	instance, _ := system.NewFlareSystemsManager(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*system.FlareSystemsManagerRewardEpochStarted, error) {
		return instance.ParseRewardEpochStarted(log)
	}

	events, err := QueryEvents(
		db,
		searchIntervalStartSec,
		searchIntervalEndSec,
		params.Net.Contracts.FlareSystemsManager,
		common2.EventTopic0.RewardEpochStarted,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("error fetching reward epoch started events: %s", err)
	}

	starts := map[ty.RewardEpochId]uint64{}
	for _, event := range events {
		// The event's own timestamp parameter is block.timestamp of the emitting block, the same value
		// the indexer records for the log, so either is the moment the epoch id flipped.
		starts[ty.RewardEpochId(event.RewardEpochId.Uint64())] = event.Timestamp
	}

	return starts, nil
}

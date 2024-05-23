package parameters

import (
	"math"

	"github.com/pkg/errors"
)

// TODO: This was auto-converted from TS, needs reviewing

type EpochParameters struct {
	FirstVotingRoundStartTs            int64
	VotingEpochDurationSeconds         int64
	FirstRewardEpochStartVotingRoundId int64
	RewardEpochDurationInVotingEpochs  int64
	RevealDeadlineSeconds              int64
}

func (e *EpochParameters) VotingEpochForTimeSec(unixSeconds int64) int64 {
	return int64(math.Floor(float64(unixSeconds-e.FirstVotingRoundStartTs) / float64(e.VotingEpochDurationSeconds)))
}

func (e *EpochParameters) VotingEpochForTime(unixMilli int64) int64 {
	unixSeconds := unixMilli / 1000
	return e.VotingEpochForTimeSec(unixSeconds)
}

func (e *EpochParameters) NextVotingEpochStartMs(unixMilli int64) int64 {
	currentEpoch := e.VotingEpochForTime(unixMilli)
	return e.VotingEpochStartMs(currentEpoch + 1)
}

func (e *EpochParameters) VotingEpochStartSec(votingEpochId int64) int64 {
	return e.FirstVotingRoundStartTs + votingEpochId*e.VotingEpochDurationSeconds
}

func (e *EpochParameters) VotingEpochStartMs(votingEpochId int64) int64 {
	return e.VotingEpochStartSec(votingEpochId) * 1000
}

func (e *EpochParameters) VotingEpochEndSec(votingEpochId int64) int64 {
	return e.VotingEpochStartSec(votingEpochId+1) - 1
}

func (e *EpochParameters) RevealDeadlineSec(votingEpochId int64) int64 {
	return e.VotingEpochStartSec(votingEpochId) + e.RevealDeadlineSeconds - 1
}

func (e *EpochParameters) ExpectedFirstVotingRoundForRewardEpoch(rewardEpochId int64) int64 {
	return e.FirstRewardEpochStartVotingRoundId + rewardEpochId*e.RewardEpochDurationInVotingEpochs
}

func (e *EpochParameters) ExpectedRewardEpochStartTimeSec(rewardEpochId int64) int64 {
	return e.VotingEpochStartSec(e.ExpectedFirstVotingRoundForRewardEpoch(rewardEpochId))
}

func (e *EpochParameters) RewardEpochForTimeSec(timeSec int64) (int64, error) {
	votingEpochId := e.VotingEpochForTimeSec(timeSec)
	return e.ExpectedRewardEpochForVotingEpoch(votingEpochId)
}

func (e *EpochParameters) ExpectedRewardEpochForVotingEpoch(votingEpochId int64) (int64, error) {
	if votingEpochId < e.FirstRewardEpochStartVotingRoundId {
		return 0, errors.Errorf(
			"votingEpochId %d is before firstRewardEpochStartVotingRoundId %d", votingEpochId, e.FirstRewardEpochStartVotingRoundId,
		)
	}
	return int64(math.Floor(float64(votingEpochId-e.FirstRewardEpochStartVotingRoundId) / float64(e.RewardEpochDurationInVotingEpochs))), nil
}

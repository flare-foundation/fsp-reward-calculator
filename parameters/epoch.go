package parameters

import (
	"math"

	"github.com/pkg/errors"
)

// TODO: This was auto-converted from TS, needs reviewing

type EpochParameters struct {
	FirstVotingRoundStartTs            int64
	VotingRoundDurationSeconds         int64
	FirstRewardEpochStartVotingRoundId int64
	RewardEpochDurationInVotingEpochs  int64
	RevealDeadlineSeconds              int64
}

func (e *EpochParameters) VotingRoundForTimeSec(unixSeconds int64) int64 {
	return int64(math.Floor(float64(unixSeconds-e.FirstVotingRoundStartTs) / float64(e.VotingRoundDurationSeconds)))
}

func (e *EpochParameters) VotingRoundForTime(unixMilli int64) int64 {
	unixSeconds := unixMilli / 1000
	return e.VotingRoundForTimeSec(unixSeconds)
}

func (e *EpochParameters) NextVotingRoundStartMs(unixMilli int64) int64 {
	currentEpoch := e.VotingRoundForTime(unixMilli)
	return e.VotingRoundStartMs(currentEpoch + 1)
}

func (e *EpochParameters) VotingRoundStartSec(round int64) int64 {
	return e.FirstVotingRoundStartTs + round*e.VotingRoundDurationSeconds
}

func (e *EpochParameters) VotingRoundStartMs(round int64) int64 {
	return e.VotingRoundStartSec(round) * 1000
}

func (e *EpochParameters) VotingRoundEndSec(round int64) int64 {
	return e.VotingRoundStartSec(round+1) - 1
}

func (e *EpochParameters) RevealDeadlineSec(round int64) int64 {
	return e.VotingRoundStartSec(round) + e.RevealDeadlineSeconds - 1
}

func (e *EpochParameters) ExpectedFirstVotingRoundForRewardEpoch(epoch int64) int64 {
	return e.FirstRewardEpochStartVotingRoundId + epoch*e.RewardEpochDurationInVotingEpochs
}

func (e *EpochParameters) ExpectedRewardEpochStartTimeSec(epoch int64) int64 {
	return e.VotingRoundStartSec(e.ExpectedFirstVotingRoundForRewardEpoch(epoch))
}

func (e *EpochParameters) RewardEpochForTimeSec(timeSec int64) (int64, error) {
	votingEpochId := e.VotingRoundForTimeSec(timeSec)
	return e.ExpectedRewardEpochForVotingRound(votingEpochId)
}

func (e *EpochParameters) ExpectedRewardEpochForVotingRound(round int64) (int64, error) {
	if round < e.FirstRewardEpochStartVotingRoundId {
		return 0, errors.Errorf(
			"votingEpochId %d is before firstRewardEpochStartVotingRoundId %d", round, e.FirstRewardEpochStartVotingRoundId,
		)
	}
	return int64(math.Floor(float64(round-e.FirstRewardEpochStartVotingRoundId) / float64(e.RewardEpochDurationInVotingEpochs))), nil
}

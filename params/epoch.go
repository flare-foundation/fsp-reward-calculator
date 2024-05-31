package params

import (
	"math"

	"github.com/pkg/errors"
)

// TODO: This was auto-converted from TS, needs reviewing

type Epoch struct {
	FirstVotingRoundStartTs            uint64
	VotingRoundDurationSeconds         uint64
	FirstRewardEpochStartVotingRoundId uint64
	RewardEpochDurationInVotingEpochs  uint64
	RevealDeadlineSeconds              uint64
}

func (e *Epoch) VotingRoundForTimeSec(unixSeconds uint64) uint64 {
	return uint64(math.Floor(float64(unixSeconds-e.FirstVotingRoundStartTs) / float64(e.VotingRoundDurationSeconds)))
}

func (e *Epoch) VotingRoundForTime(unixMilli uint64) uint64 {
	unixSeconds := unixMilli / 1000
	return e.VotingRoundForTimeSec(unixSeconds)
}

func (e *Epoch) NextVotingRoundStartMs(unixMilli uint64) uint64 {
	currentEpoch := e.VotingRoundForTime(unixMilli)
	return e.VotingRoundStartMs(currentEpoch + 1)
}

func (e *Epoch) VotingRoundStartSec(round uint64) uint64 {
	return e.FirstVotingRoundStartTs + round*e.VotingRoundDurationSeconds
}

func (e *Epoch) VotingRoundStartMs(round uint64) uint64 {
	return e.VotingRoundStartSec(round) * 1000
}

func (e *Epoch) VotingRoundEndSec(round uint64) uint64 {
	return e.VotingRoundStartSec(round+1) - 1
}

func (e *Epoch) RevealDeadlineSec(round uint64) uint64 {
	return e.VotingRoundStartSec(round) + e.RevealDeadlineSeconds - 1
}

func (e *Epoch) ExpectedFirstVotingRoundForRewardEpoch(epoch uint64) uint64 {
	return e.FirstRewardEpochStartVotingRoundId + epoch*e.RewardEpochDurationInVotingEpochs
}

func (e *Epoch) ExpectedRewardEpochStartTimeSec(epoch uint64) uint64 {
	return e.VotingRoundStartSec(e.ExpectedFirstVotingRoundForRewardEpoch(epoch))
}

func (e *Epoch) RewardEpochForTimeSec(timeSec uint64) (uint64, error) {
	votingEpochId := e.VotingRoundForTimeSec(timeSec)
	return e.ExpectedRewardEpochForVotingRound(votingEpochId)
}

func (e *Epoch) ExpectedRewardEpochForVotingRound(round uint64) (uint64, error) {
	if round < e.FirstRewardEpochStartVotingRoundId {
		return 0, errors.Errorf(
			"votingEpochId %d is before firstRewardEpochStartVotingRoundId %d", round, e.FirstRewardEpochStartVotingRoundId,
		)
	}
	return uint64(math.Floor(float64(round-e.FirstRewardEpochStartVotingRoundId) / float64(e.RewardEpochDurationInVotingEpochs))), nil
}

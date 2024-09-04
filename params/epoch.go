package params

import (
	"ftsov2-rewarding/ty"
	"math"

	"github.com/pkg/errors"
)

// TODO: This was auto-converted from TS, needs reviewing

type Epoch struct {
	FirstVotingRoundStartTs                    uint64
	VotingRoundDurationSeconds                 uint64
	FirstRewardEpochStartVotingRoundId         ty.RoundId
	RewardEpochDurationInVotingEpochs          uint64
	RevealDeadlineSeconds                      uint64
	NewSigningPolicyInitializationStartSeconds uint64
}

func (e *Epoch) VotingRoundForTimeSec(unixSeconds uint64) ty.RoundId {
	return ty.RoundId(uint64(math.Floor(float64(unixSeconds-e.FirstVotingRoundStartTs) / float64(e.VotingRoundDurationSeconds))))
}

func (e *Epoch) VotingRoundForTime(unixMilli uint64) ty.RoundId {
	unixSeconds := unixMilli / 1000
	return e.VotingRoundForTimeSec(unixSeconds)
}

func (e *Epoch) NextVotingRoundStartMs(unixMilli uint64) uint64 {
	currentEpoch := e.VotingRoundForTime(unixMilli)
	return e.VotingRoundStartMs(currentEpoch + 1)
}

func (e *Epoch) VotingRoundStartSec(round ty.RoundId) uint64 {
	return e.FirstVotingRoundStartTs + uint64(round)*e.VotingRoundDurationSeconds
}

func (e *Epoch) VotingRoundStartMs(round ty.RoundId) uint64 {
	return e.VotingRoundStartSec(round) * 1000
}

func (e *Epoch) VotingRoundEndSec(round ty.RoundId) uint64 {
	return e.VotingRoundStartSec(round+1) - 1
}

func (e *Epoch) RevealDeadlineSec(round ty.RoundId) uint64 {
	return e.VotingRoundStartSec(round) + e.RevealDeadlineSeconds - 1
}

func (e *Epoch) ExpectedFirstVotingRoundForRewardEpoch(epoch ty.EpochId) ty.RoundId {
	return ty.RoundId(uint64(e.FirstRewardEpochStartVotingRoundId) + uint64(epoch)*e.RewardEpochDurationInVotingEpochs)
}

func (e *Epoch) ExpectedRewardEpochStartTimeSec(epoch ty.EpochId) uint64 {
	return e.VotingRoundStartSec(e.ExpectedFirstVotingRoundForRewardEpoch(epoch))
}

func (e *Epoch) RewardEpochForTimeSec(timeSec uint64) (uint64, error) {
	votingEpochId := e.VotingRoundForTimeSec(timeSec)
	return e.ExpectedRewardEpochForVotingRound(votingEpochId)
}

func (e *Epoch) ExpectedRewardEpochForVotingRound(round ty.RoundId) (uint64, error) {
	if round < e.FirstRewardEpochStartVotingRoundId {
		return 0, errors.Errorf(
			"votingEpochId %d is before firstRewardEpochStartVotingRoundId %d", round, e.FirstRewardEpochStartVotingRoundId,
		)
	}
	return uint64(math.Floor(float64(round-e.FirstRewardEpochStartVotingRoundId) / float64(e.RewardEpochDurationInVotingEpochs))), nil
}

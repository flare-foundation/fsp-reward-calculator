package params

import (
	"fsp-rewards-calculator/common/ty"
	"math"

	"github.com/pkg/errors"
)

type Epoch struct {
	FirstVotingRoundStartTs                    uint64
	VotingEpochDurationSeconds                 uint64
	FirstRewardEpochStartVotingRoundId         ty.RoundId
	RewardEpochDurationInVotingEpochs          uint64
	RevealDeadlineSeconds                      uint64
	NewSigningPolicyInitializationStartSeconds uint64
}

func (e *Epoch) VotingEpochForTimeSec(unixSeconds uint64) ty.VotingEpochId {
	return ty.VotingEpochId(math.Floor(float64(unixSeconds-e.FirstVotingRoundStartTs) / float64(e.VotingEpochDurationSeconds)))
}

// VotingRoundStartSec returns Unix seconds for the start of the voting round.
// A voting round begins at the start of the voting epoch with the same id.
func (e *Epoch) VotingRoundStartSec(round ty.RoundId) uint64 {
	return e.VotingEpochStartSec(ty.VotingEpochId(round))
}

func (e *Epoch) VotingEpochStartSec(votingEpoch ty.VotingEpochId) uint64 {
	return e.FirstVotingRoundStartTs + uint64(votingEpoch)*e.VotingEpochDurationSeconds
}

// VotingRoundEndSec returns Unix seconds for the end of the voting round.
// A voting round begins at the start of the voting epoch with the same id and ends at the end of the next voting epoch.
func (e *Epoch) VotingRoundEndSec(round ty.RoundId) uint64 {
	startEpoch := ty.VotingEpochId(round)
	endEpoch := startEpoch + 1
	return e.VotingEpochStartSec(endEpoch+1) - 1
}

func (e *Epoch) RevealDeadlineSec(votingEpoch ty.VotingEpochId) uint64 {
	return e.VotingEpochStartSec(votingEpoch) + e.RevealDeadlineSeconds - 1
}

func (e *Epoch) ExpectedFirstVotingRoundForRewardEpoch(rewardEpoch ty.RewardEpochId) ty.RoundId {
	return ty.RoundId(uint64(e.FirstRewardEpochStartVotingRoundId) + uint64(rewardEpoch)*e.RewardEpochDurationInVotingEpochs)
}

func (e *Epoch) ExpectedRewardEpochStartTimeSec(rewardEpoch ty.RewardEpochId) uint64 {
	return e.VotingRoundStartSec(e.ExpectedFirstVotingRoundForRewardEpoch(rewardEpoch))
}

func (e *Epoch) RewardEpochForTimeSec(timeSec uint64) (ty.RewardEpochId, error) {
	votingEpoch := e.VotingEpochForTimeSec(timeSec)
	return e.ExpectedRewardEpochForVotingEpoch(votingEpoch)
}

func (e *Epoch) ExpectedRewardEpochForVotingEpoch(votingEpoch ty.VotingEpochId) (ty.RewardEpochId, error) {
	if votingEpoch.Value() < e.FirstRewardEpochStartVotingRoundId.Value() {
		return 0, errors.Errorf(
			"votingEpoch %d is before firstRewardEpochStartVotingRoundId %d", votingEpoch, e.FirstRewardEpochStartVotingRoundId,
		)
	}
	return ty.RewardEpochId(
		math.Floor(
			float64(votingEpoch.Value()-e.FirstRewardEpochStartVotingRoundId.Value()) / float64(e.RewardEpochDurationInVotingEpochs),
		),
	), nil
}

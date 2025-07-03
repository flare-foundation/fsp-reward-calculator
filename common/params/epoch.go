package params

import (
	"fsp-rewards-calculator/common/ty"
	"math"

	"github.com/pkg/errors"
)

/*
    FSP Voting Epochs and Rounds

	Voting epochs start at timestamp specified by the FirstVotingRoundStartTs system parameter, and run continuously
	every VotingEpochDurationSeconds.

	Voting Epochs (90 seconds each):
    ┌─────────────────────┬─────────────────────┬─────────────────────┐
    │      Epoch 0        │      Epoch 1        │      Epoch 2        │
    │       90s           │       90s           │       90s           │
    └─────────────────────┴─────────────────────┴─────────────────────┘
    0                    90                   180                   270

	Voting rounds are aligned with epochs: round X starts at the beginning of epoch X.
	However, voting rounds contain multiple phases and overlap with each other: round X finishes once the finalization
	phase completes during epoch X+1.



    Voting Rounds, using FTSO as an example:
    ┌──────────────────────────────────────┐
    │                  Round 0             │  ← Epochs 0-1
    ├─────────────────────┬──────────┬──┬──┤
    │       Commit        │  Reveal  │Sg│F │
    │        90s          │   45s    │15│10│
    └─────────────────────┴──────────┴──┴──┘

                          ┌──────────────────────────────────────┐
                          │                  Round 1             │  ← Epochs 1-2
                          ├─────────────────────┬──────────┬──┬──┤
                          │       Commit        │  Reveal  │Sg│F │
                          │        90s          │   45s    │15│10│
                          └─────────────────────┴──────────┴──┴──┘

    Phase details:
	- Commit (90s):   Providers submit hash of feed values for the round.
	- Reveal (45s):   Providers reveal feed values.
	- Sign (15s):     Providers sign the round result with median values.
	- Finalize (10s): Providers collect signatures and finalize the round.
*/

type Epoch struct {
	FirstVotingRoundStartTs                    uint64
	VotingEpochDurationSeconds                 uint64 // 90 seconds
	FirstRewardEpochStartVotingRoundId         ty.RoundId
	RewardEpochDurationInVotingEpochs          uint64
	RevealDeadlineSeconds                      uint64 // 45 seconds
	NewSigningPolicyInitializationStartSeconds uint64
}

func (e *Epoch) VotingEpochForTimeSec(unixSeconds uint64) ty.VotingEpochId {
	return ty.VotingEpochId(math.Floor(float64(unixSeconds-e.FirstVotingRoundStartTs) / float64(e.VotingEpochDurationSeconds)))
}

// VotingRoundStartSec returns Unix time for the start of the voting round.
// A voting round begins at the start of the voting epoch with the same id.
func (e *Epoch) VotingRoundStartSec(round ty.RoundId) uint64 {
	return e.VotingEpochStartSec(ty.VotingEpochId(round))
}

func (e *Epoch) VotingEpochStartSec(votingEpoch ty.VotingEpochId) uint64 {
	return e.FirstVotingRoundStartTs + uint64(votingEpoch)*e.VotingEpochDurationSeconds
}

// VotingRoundEndSec returns Unix time for the end of the voting round.
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

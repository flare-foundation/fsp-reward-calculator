// Package fcc collects the fees of the Flare Confidential Compute (FCC) contracts from the indexer.
//
// FCC funds reach the RewardManager through exactly two receiveRewards call sites, each paired one to
// one with an event emitted in the same function:
//
//   - FlareTeeManager.TeeInstructionsSent.fee - the full msg.value of a TEE instruction dispatch
//   - Fdc2Hub.AttestationRequested.fee        - the configured FDC2 attestation type/source fee
//
// The two are disjoint: an FDC2 request paying P credits the configured fee F through Fdc2Hub and
// forwards P - F into FlareTeeManager, so the two events sum to exactly P with no overlap and no gap.
// Summing both over a reward epoch therefore yields exactly the funds FCC added to the RewardManager.
//
// Nothing here is shared with the legacy FDC (common/fdc): different contracts, different events
// (AttestationRequested vs AttestationRequest) and a different reward path.
package fcc

import (
	common2 "fsp-rewards-calculator/common"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/contracts/fdc2"
	"fsp-rewards-calculator/contracts/tee"
	"fsp-rewards-calculator/logger"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Fees are the FCC fees credited to a reward epoch, per voting round.
type Fees struct {
	Epoch ty.RewardEpochId
	// ByRound holds the fee total of each voting round that carries FCC activity.
	ByRound map[ty.RoundId]*big.Int
	// Total is exactly the funds FCC added to the RewardManager during the epoch.
	Total *big.Int
}

// teeInstruction is a TEE instruction dispatch fee credited to the RewardManager.
//
// rewardEpochId is the very value passed to RewardManager.receiveRewards, which makes it authoritative
// for fund attribution: it says which epoch the funds actually landed in, independently of any voting
// round bucketing.
type teeInstruction struct {
	instructionId common.Hash
	rewardEpochId uint32
	fee           *big.Int
	timestampSec  uint64
}

// fdc2Request is an FDC2 attestation request fee credited to the RewardManager.
//
// The event carries no reward epoch id of its own. It is always emitted in the same transaction as the
// TeeInstructionsSent that shares its instruction id, so that event's reward epoch id is its own.
type fdc2Request struct {
	instructionId common.Hash
	fee           *big.Int
	timestampSec  uint64
}

// GetEpochFees collects the FCC fees credited to a reward epoch, bucketed per voting round.
func GetEpochFees(db *gorm.DB, re *fsp.RewardEpoch) (*Fees, error) {
	window, err := fsp.RewardEpochFundingWindow(db, re.Epoch)
	if err != nil {
		return nil, err
	}

	teeInstructions, err := getTeeInstructionsSentEvents(db, window)
	if err != nil {
		return nil, err
	}
	fdc2Requests, err := getFdc2AttestationRequestedEvents(db, window)
	if err != nil {
		return nil, err
	}

	return attributeFees(re.Epoch, re.StartRound, re.EndRound, teeInstructions, fdc2Requests)
}

// attributeFees keeps the fees of the funding window that the reward epoch was actually credited with,
// and buckets them per voting round.
//
// The TEE event carries the reward epoch id passed to receiveRewards, and an FDC2 request inherits it
// from the TeeInstructionsSent sharing its instruction id. Everything else in the window belongs to a
// neighbouring epoch, which is expected: the window overshoots both boundaries on purpose so that
// nothing is missed.
func attributeFees(
	epoch ty.RewardEpochId,
	startRound ty.RoundId,
	endRound ty.RoundId,
	teeInstructions []teeInstruction,
	fdc2Requests []fdc2Request,
) (*Fees, error) {
	fees := &Fees{Epoch: epoch, ByRound: map[ty.RoundId]*big.Int{}, Total: big.NewInt(0)}

	// Built over the whole window, before any filtering, so that an FDC2 request whose pair falls in a
	// different voting round still resolves.
	rewardEpochIdByInstructionId := map[common.Hash]uint32{}
	for _, instruction := range teeInstructions {
		rewardEpochIdByInstructionId[instruction.instructionId] = instruction.rewardEpochId
	}

	excluded := 0
	for _, instruction := range teeInstructions {
		if instruction.rewardEpochId != uint32(epoch) {
			excluded++
			continue
		}
		fees.add(votingRound(instruction.timestampSec, startRound, endRound), instruction.fee)
	}

	for _, request := range fdc2Requests {
		pairedRewardEpochId, paired := rewardEpochIdByInstructionId[request.instructionId]
		switch {
		case !paired:
			// Every FDC2 request forwards its remainder into FlareTeeManager in the same transaction, so the
			// pair is always emitted together. Without it the fee cannot be attributed to any epoch at all,
			// so this is an indexer gap rather than something to attribute around.
			return nil, errors.Errorf(
				"FDC2 attestation request %s has no TeeInstructionsSent with the same instruction id in the funding window of reward epoch %d, so its fee cannot be attributed: events are missing from the indexer",
				request.instructionId, epoch,
			)
		case pairedRewardEpochId != uint32(epoch):
			excluded++
		default:
			fees.add(votingRound(request.timestampSec, startRound, endRound), request.fee)
		}
	}

	// Expected at the epoch boundaries, where the window deliberately overshoots.
	logger.Debug("FCC events in the funding window credited to a neighbouring reward epoch: %d", excluded)

	return fees, nil
}

func (f *Fees) add(round ty.RoundId, fee *big.Int) {
	if f.ByRound[round] == nil {
		f.ByRound[round] = big.NewInt(0)
	}
	f.ByRound[round].Add(f.ByRound[round], fee)
	f.Total.Add(f.Total, fee)
}

// votingRound decides which voting round an FCC fee event is recorded against: the round of its
// timestamp, clamped into the reward epoch's range.
//
// The funding window is bounded by the RewardEpochStarted events and overshoots the epoch's voting round
// schedule at both ends, so a fee paid after the last scheduled round but still credited to this epoch
// belongs to its final round, and one paid before the first scheduled round belongs to the first.
func votingRound(timestampSec uint64, startRound, endRound ty.RoundId) ty.RoundId {
	round := ty.RoundId(params.Net.Epoch.VotingEpochForTimeSec(timestampSec))
	return min(max(round, startRound), endRound)
}

// parseTeeInstructionsSent decodes a FlareTeeManager TeeInstructionsSent log. The instruction payload
// fields of the event (teeMachines, message, cosigners) are payload rather than accounting data and are
// not retained.
func parseTeeInstructionsSent(log types.Log, timestampSec uint64) (teeInstruction, error) {
	instance, _ := tee.NewTeeManager(common.Address{}, nil)
	event, err := instance.ParseTeeInstructionsSent(log)
	if err != nil {
		return teeInstruction{}, err
	}
	return teeInstruction{
		instructionId: common.Hash(event.InstructionId),
		rewardEpochId: event.RewardEpochId,
		fee:           event.Fee,
		timestampSec:  timestampSec,
	}, nil
}

// parseFdc2AttestationRequested decodes an Fdc2Hub AttestationRequested log.
func parseFdc2AttestationRequested(log types.Log, timestampSec uint64) (fdc2Request, error) {
	instance, _ := fdc2.NewFdc2(common.Address{}, nil)
	event, err := instance.ParseAttestationRequested(log)
	if err != nil {
		return fdc2Request{}, err
	}
	return fdc2Request{
		instructionId: common.Hash(event.InstructionId),
		fee:           event.Fee,
		timestampSec:  timestampSec,
	}, nil
}

func getTeeInstructionsSentEvents(db *gorm.DB, window fsp.FundingWindow) ([]teeInstruction, error) {
	events, err := queryFundingWindow(db, window, params.Net.Contracts.FlareTeeManager, common2.EventTopic0.TeeInstructionsSent, parseTeeInstructionsSent)
	if err != nil {
		return nil, errors.Errorf("error fetching TEE instruction events: %s", err)
	}
	return events, nil
}

func getFdc2AttestationRequestedEvents(db *gorm.DB, window fsp.FundingWindow) ([]fdc2Request, error) {
	events, err := queryFundingWindow(db, window, params.Net.Contracts.Fdc2Hub, common2.EventTopic0.Fdc2AttestationRequested, parseFdc2AttestationRequested)
	if err != nil {
		return nil, errors.Errorf("error fetching FDC2 attestation request events: %s", err)
	}
	return events, nil
}

// queryFundingWindow queries events over a funding window. Both edges of the window are inclusive,
// while QueryEvents takes an exclusive end, hence the +1.
func queryFundingWindow[T any](
	db *gorm.DB,
	window fsp.FundingWindow,
	contractAddress common.Address,
	topic0 string,
	parseEvent func(types.Log, uint64) (T, error),
) ([]T, error) {
	return fsp.QueryEventsStrict(db, window.StartSec, window.EndSec+1, contractAddress, topic0, parseEvent)
}

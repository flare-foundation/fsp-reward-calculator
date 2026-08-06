package fcc

import (
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

const (
	testEpoch      = ty.RewardEpochId(419)
	testStartRound = ty.RoundId(1407840)
	testEndRound   = ty.RoundId(1411199)
)

func init() {
	params.InitNetwork("songbird")
}

func instructionId(id byte) common.Hash {
	return common.Hash{id}
}

func roundStartSec(round ty.RoundId) uint64 {
	return params.Net.Epoch.VotingRoundStartSec(round)
}

func attribute(t *testing.T, tee []teeInstruction, fdc2 []fdc2Request) *Fees {
	t.Helper()
	fees, err := attributeFees(testEpoch, testStartRound, testEndRound, tee, fdc2)
	if err != nil {
		t.Fatalf("attributing fees: %s", err)
	}
	return fees
}

func wantRound(t *testing.T, fees *Fees, round ty.RoundId, want int64) {
	t.Helper()
	got := fees.ByRound[round]
	if got == nil || got.Cmp(big.NewInt(want)) != 0 {
		t.Errorf("fees of round %d: got %s, want %d", round, got, want)
	}
}

// A TEE instruction and the FDC2 request it was dispatched for are the two disjoint halves of one
// payment, and both belong to the round of their timestamp.
func TestAttributeFeesPairedRequest(t *testing.T) {
	round := testStartRound + 7
	fees := attribute(t,
		[]teeInstruction{{
			instructionId: instructionId(1),
			rewardEpochId: uint32(testEpoch),
			fee:           big.NewInt(4),
			timestampSec:  roundStartSec(round),
		}},
		[]fdc2Request{{
			instructionId: instructionId(1),
			fee:           big.NewInt(6),
			timestampSec:  roundStartSec(round),
		}},
	)

	if got := fees.Total; got.Cmp(big.NewInt(10)) != 0 {
		t.Errorf("total fees: got %s, want 10", got)
	}
	if got := len(fees.ByRound); got != 1 {
		t.Errorf("rounds with fees: got %d, want 1", got)
	}
	wantRound(t, fees, round, 10)
}

// Events inside the window but credited to a neighbouring epoch fund that epoch, not this one. Expected
// at the boundaries, where the window deliberately overshoots.
func TestAttributeFeesExcludesOtherRewardEpochs(t *testing.T) {
	fees := attribute(t,
		[]teeInstruction{
			{
				instructionId: instructionId(1),
				rewardEpochId: uint32(testEpoch) - 1,
				fee:           big.NewInt(5),
				timestampSec:  roundStartSec(testStartRound),
			},
			{
				instructionId: instructionId(2),
				rewardEpochId: uint32(testEpoch) + 1,
				fee:           big.NewInt(7),
				timestampSec:  roundStartSec(testEndRound),
			},
		},
		// Both requests inherit their pair's epoch, so neither belongs to this one either.
		[]fdc2Request{
			{instructionId: instructionId(1), fee: big.NewInt(1), timestampSec: roundStartSec(testStartRound)},
			{instructionId: instructionId(2), fee: big.NewInt(2), timestampSec: roundStartSec(testEndRound)},
		},
	)

	if got := fees.Total; got.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("total fees: got %s, want 0", got)
	}
	if got := len(fees.ByRound); got != 0 {
		t.Errorf("rounds with fees: got %d, want 0", got)
	}
}

// An FDC2 request carries no reward epoch id of its own, so without its TEE pair its fee cannot be
// attributed to any epoch at all. Attributing it anywhere would silently move funds, so it fails.
func TestAttributeFeesUnpairedRequestFails(t *testing.T) {
	_, err := attributeFees(testEpoch, testStartRound, testEndRound, nil,
		[]fdc2Request{{instructionId: instructionId(9), fee: big.NewInt(3), timestampSec: roundStartSec(testStartRound)}},
	)

	if err == nil {
		t.Fatal("expected an unpaired FDC2 request to fail")
	}
	if !strings.Contains(err.Error(), "missing from the indexer") {
		t.Errorf("error should name the cause, got: %s", err)
	}
}

// The window overshoots the epoch's voting rounds at both ends, so a fee credited to this epoch but paid
// outside its schedule belongs to the nearest round of the epoch, never to a round it does not own.
func TestAttributeFeesClampsToEpochRounds(t *testing.T) {
	fees := attribute(t,
		[]teeInstruction{
			{
				instructionId: instructionId(1),
				rewardEpochId: uint32(testEpoch),
				fee:           big.NewInt(11),
				timestampSec:  roundStartSec(testStartRound) - 1,
			},
			{
				instructionId: instructionId(2),
				rewardEpochId: uint32(testEpoch),
				fee:           big.NewInt(13),
				timestampSec:  roundStartSec(testEndRound + 2),
			},
		},
		nil,
	)

	wantRound(t, fees, testStartRound, 11)
	wantRound(t, fees, testEndRound, 13)
	if got := fees.Total; got.Cmp(big.NewInt(24)) != 0 {
		t.Errorf("total fees: got %s, want 24", got)
	}
}

// A request and its pair can fall in different voting rounds. The pairing map is built over the whole
// window before any bucketing, so the request still resolves to the epoch its pair funded.
func TestAttributeFeesPairsAcrossRounds(t *testing.T) {
	fees := attribute(t,
		[]teeInstruction{{
			instructionId: instructionId(1),
			rewardEpochId: uint32(testEpoch),
			fee:           big.NewInt(1),
			timestampSec:  roundStartSec(testStartRound + 100),
		}},
		[]fdc2Request{{
			instructionId: instructionId(1),
			fee:           big.NewInt(2),
			timestampSec:  roundStartSec(testStartRound + 101),
		}},
	)

	wantRound(t, fees, testStartRound+100, 1)
	wantRound(t, fees, testStartRound+101, 2)
}

// An epoch without FCC activity carries no fees and so produces no claim, leaving its distribution
// unchanged.
func TestAttributeFeesWithoutActivity(t *testing.T) {
	fees := attribute(t, nil, nil)

	if got := fees.Total; got.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("total fees: got %s, want 0", got)
	}
	if got := len(fees.ByRound); got != 0 {
		t.Errorf("rounds with fees: got %d, want 0", got)
	}
}

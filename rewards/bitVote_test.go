package rewards

import (
	"fsp-rewards-calculator/common/fsp"
	ty2 "fsp-rewards-calculator/common/ty"
	"math/big"
	"testing"
)

// bitVoteMessage encodes a bitVote as an unsigned message: two length bytes
// followed by the big-endian bitVector, as expected by fdc.DecodeBitVote.
func bitVoteMessage(bitVote *big.Int) []byte {
	return append([]byte{0x00, 0x10}, bitVote.Bytes()...)
}

func testConsensusSigs(voters []*fsp.VoterInfo, bitVotes []*big.Int) map[ty2.VoterSigning]fsp.SigInfo {
	sigs := map[ty2.VoterSigning]fsp.SigInfo{}
	for i, voter := range voters {
		sigs[voter.Signing] = fsp.SigInfo{
			Signer:          voter.Signing,
			UnsignedMessage: bitVoteMessage(bitVotes[i]),
		}
	}
	return sigs
}

func TestGetConsensusBitVoteMaxWeightWinsRegardlessOfOrder(t *testing.T) {
	voters := []*fsp.VoterInfo{
		testFdcVoter(1, 10),
		testFdcVoter(2, 10),
		testFdcVoter(3, 0),
	}
	voterIndex := testFdcVoterIndex(voters...)

	// Both weighted voters vote 7; a zero-weight voter votes a numerically
	// smaller bitVote. The zero-weight vote must never win, no matter the
	// map iteration order.
	bitVotes := []*big.Int{big.NewInt(7), big.NewInt(7), big.NewInt(1)}
	sigs := testConsensusSigs(voters, bitVotes)

	for i := 0; i < 50; i++ {
		consensus := getConsensusBitVote(sigs, ty2.RoundId(1), voterIndex)
		if consensus == nil || consensus.Cmp(big.NewInt(7)) != 0 {
			t.Fatalf("run %d: got consensus %s, want 7", i, consensus)
		}
	}
}

func TestGetConsensusBitVoteLowerWeightSmallerValueDoesNotWin(t *testing.T) {
	voters := []*fsp.VoterInfo{
		testFdcVoter(1, 5),
		testFdcVoter(2, 10),
	}
	voterIndex := testFdcVoterIndex(voters...)

	// The lighter voter has a numerically smaller bitVote; the heavier vote
	// must win outright, without tie-breaking against the lighter one.
	bitVotes := []*big.Int{big.NewInt(3), big.NewInt(12)}
	sigs := testConsensusSigs(voters, bitVotes)

	for i := 0; i < 50; i++ {
		consensus := getConsensusBitVote(sigs, ty2.RoundId(1), voterIndex)
		if consensus == nil || consensus.Cmp(big.NewInt(12)) != 0 {
			t.Fatalf("run %d: got consensus %s, want 12", i, consensus)
		}
	}
}

func TestGetConsensusBitVoteTieBreaksLexicographically(t *testing.T) {
	voters := []*fsp.VoterInfo{
		testFdcVoter(1, 10),
		testFdcVoter(2, 10),
	}
	voterIndex := testFdcVoterIndex(voters...)

	// Equal weight on bitVotes 9 and 10. The TS implementation sorts BigInts
	// as decimal strings, so "10" < "9" and 10 wins the tie.
	bitVotes := []*big.Int{big.NewInt(9), big.NewInt(10)}
	sigs := testConsensusSigs(voters, bitVotes)

	for i := 0; i < 50; i++ {
		consensus := getConsensusBitVote(sigs, ty2.RoundId(1), voterIndex)
		if consensus == nil || consensus.Cmp(big.NewInt(10)) != 0 {
			t.Fatalf("run %d: got consensus %s, want 10", i, consensus)
		}
	}
}

func TestGetConsensusBitVoteNoValidVotes(t *testing.T) {
	voters := []*fsp.VoterInfo{testFdcVoter(1, 10)}
	voterIndex := testFdcVoterIndex(voters...)

	sigs := map[ty2.VoterSigning]fsp.SigInfo{
		voters[0].Signing: {Signer: voters[0].Signing, UnsignedMessage: []byte{0x00}},
	}

	if consensus := getConsensusBitVote(sigs, ty2.RoundId(1), voterIndex); consensus != nil {
		t.Fatalf("got consensus %s, want nil", consensus)
	}
}

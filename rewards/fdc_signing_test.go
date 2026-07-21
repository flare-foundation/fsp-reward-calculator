package rewards

import (
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestGenerateFdcSigningClaimsSkipsZeroWeightVotersAfterDistribution(t *testing.T) {
	params.InitNetwork("songbird")

	voters := []*fsp.VoterInfo{
		testFdcVoter(1, 1),
		testFdcVoter(2, 1),
		testFdcVoter(3, 0),
	}
	voterIndex := testFdcVoterIndex(voters...)

	round := ty2.RoundId(1)
	consensusBitVote := big.NewInt(1)
	finalizations := []*fsp.Finalization{
		{
			MerkleRoot: fsp.ProtocolMerkleRoot{Round: round},
			Info: fsp.TxInfo{
				TimestampSec: params.Net.Epoch.RevealDeadlineSec(ty2.VotingEpochId(round) + 1),
			},
		},
	}

	bitVotes := map[ty2.VoterSubmit]*big.Int{}
	consensusSigs := map[ty2.VoterSigning]fsp.SigInfo{}
	for _, voter := range voters {
		bitVotes[voter.Submit] = new(big.Int).Set(consensusBitVote)
		consensusSigs[voter.Signing] = fsp.SigInfo{
			Signer:    voter.Signing,
			Timestamp: 0,
		}
	}

	claims := generateFdcSigningClaims(
		416,
		finalizations,
		round,
		big.NewInt(1000),
		bitVotes,
		consensusBitVote,
		consensusSigs,
		voterIndex,
	)

	totalClaimed := big.NewInt(0)
	for _, claim := range claims {
		if claim.Beneficiary == params.Net.Ftso.BurnAddress {
			t.Fatalf("unexpected burn claim: %s", claim.Amount)
		}
		totalClaimed.Add(totalClaimed, claim.Amount)
	}
	if totalClaimed.Cmp(big.NewInt(1000)) != 0 {
		t.Fatalf("claimed amount mismatch: got %s, want 1000", totalClaimed)
	}
}

func testFdcVoter(id int64, signingWeight uint16) *fsp.VoterInfo {
	return &fsp.VoterInfo{
		Identity:            ty2.VoterId(common.BigToAddress(big.NewInt(id))),
		Submit:              ty2.VoterSubmit(common.BigToAddress(big.NewInt(100 + id))),
		SubmitSignatures:    ty2.VoterSubmitSignatures(common.BigToAddress(big.NewInt(200 + id))),
		Signing:             ty2.VoterSigning(common.BigToAddress(big.NewInt(300 + id))),
		Delegation:          ty2.VoterDelegation(common.BigToAddress(big.NewInt(400 + id))),
		CappedWeight:        big.NewInt(1),
		DelegationFeeBips:   0,
		SigningPolicyWeight: signingWeight,
	}
}

func testFdcVoterIndex(voters ...*fsp.VoterInfo) *fsp.VoterIndex {
	index := &fsp.VoterIndex{
		PolicyOrder:        voters,
		ById:               map[ty2.VoterId]*fsp.VoterInfo{},
		BySubmit:           map[ty2.VoterSubmit]*fsp.VoterInfo{},
		BySubmitSignatures: map[ty2.VoterSubmitSignatures]*fsp.VoterInfo{},
		BySigning:          map[ty2.VoterSigning]*fsp.VoterInfo{},
		ByDelegation:       map[ty2.VoterDelegation]*fsp.VoterInfo{},
		TotalCappedWeight:  big.NewInt(0),
	}

	for _, voter := range voters {
		index.ById[voter.Identity] = voter
		index.BySubmit[voter.Submit] = voter
		index.BySubmitSignatures[voter.SubmitSignatures] = voter
		index.BySigning[voter.Signing] = voter
		index.ByDelegation[voter.Delegation] = voter
		index.TotalCappedWeight.Add(index.TotalCappedWeight, voter.CappedWeight)
		index.TotalSigningPolicyWeight += voter.SigningPolicyWeight
	}

	return index
}

func TestDominatesConsensusBitVoteDoesNotMutateInput(t *testing.T) {
	bitVote := big.NewInt(3)
	consensusBitVote := big.NewInt(1)

	if !dominatesConsensusBitVote(bitVote, consensusBitVote) {
		t.Fatal("expected bitVote to dominate consensusBitVote")
	}
	if bitVote.Cmp(big.NewInt(3)) != 0 {
		t.Fatalf("bitVote mutated: got %s, want 3", bitVote)
	}
}

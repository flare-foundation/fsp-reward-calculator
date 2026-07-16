package rewards

import (
	"fsp-rewards-calculator/common/fdc"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"math/big"
)

func dominatesConsensusBitVote(bitVote *big.Int, consensusBitVote *big.Int) bool {
	if consensusBitVote == nil || bitVote == nil {
		return false
	}
	return new(big.Int).And(bitVote, consensusBitVote).Cmp(consensusBitVote) == 0
}

func isConfirmed(attestationIndex int, consensusBitVote *big.Int) bool {
	if consensusBitVote == nil {
		return false
	}
	return consensusBitVote.Bit(attestationIndex) == 1
}

func getConsensusBitVote(sigs map[ty.VoterSigning]fsp.SigInfo, round ty.RoundId, voters *fsp.VoterIndex) *big.Int {
	bitVoteWeight := map[string]uint64{}
	for signer, sig := range sigs {
		if len(sig.UnsignedMessage) < 3 { // first two bytes are length
			logger.Debug("bitVote message too short for signer %s in round %d", signer, round)
			continue
		}
		bitVote, err := fdc.DecodeBitVote(sig.UnsignedMessage)
		if err != nil {
			logger.Debug("error parsing bitVote for signer %s in round %d: %s", signer.String(), round, err)
			continue
		}
		bitVoteWeight[bitVote.String()] += uint64(voters.BySigning[signer].SigningPolicyWeight)
	}

	if len(bitVoteWeight) == 0 {
		return nil
	}

	maxWeight := uint64(0)
	for _, weight := range bitVoteWeight {
		maxWeight = max(maxWeight, weight)
	}

	// If there is more than one candidate with max weight, the TS implementation sorts them
	// with the default Array.sort, which compares BigInts as decimal strings, and takes the
	// first one. Compare lexicographically to match.
	consensusBitVote := ""
	for bitVote, weight := range bitVoteWeight {
		if weight != maxWeight {
			continue
		}
		if consensusBitVote == "" || bitVote < consensusBitVote {
			consensusBitVote = bitVote
		}
	}

	bitVector, _ := new(big.Int).SetString(consensusBitVote, 10)
	return bitVector
}

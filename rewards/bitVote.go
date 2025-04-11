package rewards

import (
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"math/big"
)

func dominatesConsensusBitVote(bitVote *big.Int, consensusBitVote *big.Int) bool {
	if consensusBitVote == nil || bitVote == nil {
		return false
	}
	return bitVote.And(bitVote, consensusBitVote).Cmp(consensusBitVote) == 0
}

func isConfirmed(attestationIndex int, consensusBitVote *big.Int) bool {
	if consensusBitVote == nil {
		return false
	}
	return consensusBitVote.Bit(attestationIndex) == 1
}

// getConsensusBitVote returns the
func getConsensusBitVote(sigs map[ty.VoterSigning]data.SigInfo, round ty.RoundId, voters *data.VoterIndex) *big.Int {
	bitVoteWeight := map[string]uint64{}
	for signer, sig := range sigs {
		if len(sig.UnsignedMessage) < 3 { // first two bytes are length
			logger.Warn("bitVote message too short for signer %s in round %d", signer, round)
			continue
		}
		bitVote, err := data.ParseBitVote(sig.UnsignedMessage)
		if err != nil {
			logger.Warn("error parsing bitVote for signer %s in round %d: %s", signer.String(), round, err)
			continue
		}
		bitVoteWeight[bitVote.String()] += uint64(voters.BySigning[signer].SigningPolicyWeight)
	}

	var consensusBitVote *string
	if len(bitVoteWeight) > 0 {
		maxWeight := uint64(0)
		for bitVote, weight := range bitVoteWeight {
			if weight >= maxWeight {
				if consensusBitVote == nil {
					consensusBitVote = &bitVote
				} else {
					// if we have more than one candidate with max weight, choose the one with the smaller bitVote
					minBitVote := utils.MinDec(bitVote, *consensusBitVote)
					consensusBitVote = &minBitVote
				}

				maxWeight = weight
			}
		}
	}

	if consensusBitVote != nil {
		bitVector, _ := new(big.Int).SetString(*consensusBitVote, 10)
		return bitVector
	} else {
		return nil
	}
}

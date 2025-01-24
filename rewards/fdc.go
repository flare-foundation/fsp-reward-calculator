package rewards

import (
	"encoding/hex"
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"gorm.io/gorm"
	"math/big"
	"slices"
)

func getFdcRewards(db *gorm.DB, epochs data.RewardEpochs, submit2 []payload.Message, submitSignatures []payload.Message, finalizations []*data.Finalization) ([]ty.RewardClaim, error) {
	re := epochs.Current
	bitVotesByRound := data.ExtractBitVotes(submit2)
	finalizationsByRound := data.GetFinalizationsByRound(finalizations)

	consensusHashByRound := map[ty.RoundId]common.Hash{}
	for round, fs := range finalizationsByRound {
		firstSuccessfulIndex := slices.IndexFunc(fs, func(f *data.Finalization) bool {
			return f.Info.Reverted == false
		})
		if firstSuccessfulIndex < 0 {
			continue
		}
		consensusHashByRound[round] = fs[firstSuccessfulIndex].MerkleRoot.EncodedHash()
	}

	signersByRound := data.GetFdcSignersByRound(submitSignatures, consensusHashByRound, re)

	attestationRequestsByRound := data.GetAttestationRequestsByRound(db, re.StartRound, re.EndRound)

	// Get reward offers
	// calculate

	epochClaims := make([]ty.RewardClaim, 0)
	for round := re.StartRound; round <= re.EndRound; round++ {
		roundSigs, ok := signersByRound[round]
		if !ok {
			logger.Warn("no signatures for round %d", round)
			continue
		}

		hash := consensusHashByRound[round]
		consensusSigs, ok := roundSigs[hash]
		if !ok {
			logger.Warn("no signatures for finalized hash %s in round %d", hash, round)
			continue
		}

		consensusBitVote := getConsensusBitVote(consensusSigs, round, epochs.Current.VoterIndex)
		logger.Info("Consensus bitVote for round %d: %v", round, *consensusBitVote)

		offenders := getOffenders(bitVotesByRound[round], consensusSigs, roundSigs[data.WrongSignatureIndicatorMessageHash], epochs.Current.VoterIndex, consensusBitVote)
		logger.Info("Offenders for round %d: %v", round, offenders)

		//penalties := getFdcPenalties(epochs.Current.Reward, epochs.Current.PenaltyFactor, offenders, epochs.Current.VoterIndex)

	}

	logger.Info("Signers fetched %d", len(signersByRound), len(consensusHashByRound), len(finalizationsByRound), len(bitVotesByRound), len(attestationRequestsByRound))

	return epochClaims, nil
}

func getOffenders(
	bitVotes map[ty.VoterSubmit]data.BitVote,
	consensusSigs map[ty.VoterSigning]data.SigInfo,
	wrongSigs map[ty.VoterSigning]data.SigInfo,
	voterIndex *data.VoterIndex,
	consensusBitVote *string,
) map[ty.VoterId]bool {
	offenders := map[ty.VoterId]bool{}

	var revealOffenders []ty.VoterId
	for voterSubmit := range bitVotes {
		voter := voterIndex.BySubmit[voterSubmit]
		_, ok := consensusSigs[voter.Signing]
		if !ok {
			revealOffenders = append(revealOffenders, voter.Identity)
			offenders[voter.Identity] = true
		}
	}

	var wrongSignatureOffenders []ty.VoterId
	for voterSigning := range wrongSigs {
		voter, ok := voterIndex.BySigning[voterSigning]
		if !ok {
			logger.Debug("voter not found for wrong signature %s", voterSigning)
			continue
		}
		wrongSignatureOffenders = append(wrongSignatureOffenders, voter.Identity)
		offenders[voter.Identity] = true
	}

	var bitVoteOffenders []ty.VoterId
	for voterSigning, sig := range consensusSigs {
		voter := voterIndex.BySigning[voterSigning]
		offender := false

		if len(sig.UnsignedMessage) < 3 {
			offender = true
		} else {
			bitVote := hex.EncodeToString(sig.UnsignedMessage[2:])
			if bitVote != *consensusBitVote {
				offender = true
			}
		}
		if offender {
			bitVoteOffenders = append(bitVoteOffenders, voter.Identity)
			offenders[voter.Identity] = true
		}
	}

	// TODO: log different types of offenders for debugging
	logger.Warn("Offenders: reveal %d, wrong signature %d, bitVote %d", len(revealOffenders), len(wrongSignatureOffenders), len(bitVoteOffenders))

	return offenders
}

// getConsensusBitVote returns the
func getConsensusBitVote(sigs map[ty.VoterSigning]data.SigInfo, round ty.RoundId, voters *data.VoterIndex) *string {
	bitVoteWeight := map[string]uint64{}
	for signer, sig := range sigs {
		if len(sig.UnsignedMessage) < 3 { // first two bytes are length
			logger.Warn("bitVote message too short for signer %s in round %d", signer, round)
			continue
		}
		bitVote := hex.EncodeToString(sig.UnsignedMessage[2:])
		bitVoteWeight[bitVote] += uint64(voters.BySigning[signer].SigningPolicyWeight)
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
					minBitVote := utils.MinHex(bitVote, *consensusBitVote)
					consensusBitVote = &minBitVote
				}

				maxWeight = weight
			}
		}
	}
	return consensusBitVote
}

func getFdcPenalties(
	reward *big.Int,
	penaltyFactor *big.Int,
	offenders map[ty.VoterId]bool,
	voters *data.VoterIndex,
) []ty.RewardClaim {
	var penalties []ty.RewardClaim

	// TODO: precompute?
	totalSigningWeight := uint64(0)
	for _, info := range voters.ById {
		totalSigningWeight += uint64(info.SigningPolicyWeight)
	}
	bigTotalSigningWeight := big.NewInt(int64(totalSigningWeight))

	for id := range offenders {
		offender := voters.ById[id]
		if offender.SigningPolicyWeight > 0 {
			bigWeight := big.NewInt(int64(offender.SigningPolicyWeight))
			amount := new(big.Int).Div(
				bigTmp.Mul(bigWeight, bigTmp.Mul(reward, penaltyFactor)),
				bigTotalSigningWeight,
			)
			claims := SigningWeightClaimsForVoter(offender, amount)
			for i := range claims {
				claims[i].Amount.Neg(claims[i].Amount)
			}
			penalties = append(penalties, claims...)
		}
	}
	return penalties
}

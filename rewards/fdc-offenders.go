package rewards

import (
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"math/big"
)

func getOffenders(
	bitVotes map[ty.VoterSubmit]*big.Int,
	consensusSigs map[ty.VoterSigning]data.SigInfo,
	wrongSigs map[ty.VoterSigning]data.SigInfo,
	voterIndex *data.VoterIndex,
	consensusBitVote *big.Int,
) map[ty.VoterSigning]bool {
	offenders := map[ty.VoterSigning]bool{}

	var revealOffenders []ty.VoterId
	var bitVoteOffenders []ty.VoterId

	if consensusBitVote != nil {
		for voterSubmit, bitVote := range bitVotes {
			voter := voterIndex.BySubmit[voterSubmit]
			if voter == nil {
				logger.Info("voter not found for bitVote submit address %s", voterSubmit.String())
				continue
			}
			if !dominatesConsensusBitVote(bitVote, consensusBitVote) {
				continue
			}
			if consensusSigs == nil {
				logger.Warn("consensusSigs is nil, skipping reveal offenders check")
				continue
			}
			_, ok := consensusSigs[voter.Signing]
			if !ok {
				revealOffenders = append(revealOffenders, voter.Identity)
				offenders[voter.Signing] = true
			}
		}

		for voterSigning, sig := range consensusSigs {
			voter := voterIndex.BySigning[voterSigning]
			offender := false

			if len(sig.UnsignedMessage) < 3 {
				offender = true
			} else {
				bitVote, _ := data.ParseBitVote(sig.UnsignedMessage)
				if consensusBitVote.Cmp(bitVote) != 0 {
					offender = true
				}
			}
			if offender {
				bitVoteOffenders = append(bitVoteOffenders, voter.Identity)
				offenders[voter.Signing] = true
			}
		}
	} else {
		logger.Debug("consensusBitVote is nil, skipping reveal & bitVote offenders checks")
	}

	var wrongSignatureOffenders []ty.VoterId
	for voterSigning := range wrongSigs {
		voter, ok := voterIndex.BySigning[voterSigning]
		if !ok {
			logger.Info("voter not found for wrong signature %s", voterSigning.String())
			continue
		}
		wrongSignatureOffenders = append(wrongSignatureOffenders, voter.Identity)
		offenders[voter.Signing] = true
	}

	// TODO: log different types of offenders for debugging
	logger.Warn("Offenders: reveal %d, wrong signature %d, bitVote %d", len(revealOffenders), len(wrongSignatureOffenders), len(bitVoteOffenders))

	return offenders
}

func getFdcPenalties(reward *big.Int, penaltyFactor *big.Int, offenders map[ty.VoterSigning]bool, voters *data.VoterIndex) []ty.RewardClaim {
	var penalties []ty.RewardClaim

	// TODO: precompute?
	totalSigningWeight := uint64(0)
	for _, info := range voters.ById {
		totalSigningWeight += uint64(info.SigningPolicyWeight)
	}
	bigTotalSigningWeight := big.NewInt(int64(totalSigningWeight))

	for signing := range offenders {
		offender := voters.BySigning[signing]
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

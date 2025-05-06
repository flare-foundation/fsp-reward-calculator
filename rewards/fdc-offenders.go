package rewards

import (
	"encoding/hex"
	"fsp-rewards-calculator/common/fdc"
	"fsp-rewards-calculator/common/fsp"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"math/big"
)

func getOffenders(
	bitVotes map[ty2.VoterSubmit]*big.Int,
	consensusSigs map[ty2.VoterSigning]fsp.SigInfo,
	wrongSigs map[ty2.VoterSigning]fsp.SigInfo,
	voterIndex *fsp.VoterIndex,
	consensusBitVote *big.Int,
) map[ty2.VoterSigning]bool {
	offenders := map[ty2.VoterSigning]bool{}

	var revealOffenders []ty2.VoterId
	var bitVoteOffenders []ty2.VoterId

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
				bitVote, err := fdc.DecodeBitVote(sig.UnsignedMessage)
				if err != nil {
					logger.Warn("error parsing bitVote for signer %s: %s", voterSigning.String(), hex.EncodeToString(sig.UnsignedMessage), err)
					// TODO: should this be counted as an offence?
					//offender = true
				} else if consensusBitVote.Cmp(bitVote) != 0 {
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

	var wrongSignatureOffenders []ty2.VoterId
	for voterSigning := range wrongSigs {
		voter, ok := voterIndex.BySigning[voterSigning]
		if !ok {
			logger.Info("voter not found for wrong signature %s", voterSigning.String())
			continue
		}
		wrongSignatureOffenders = append(wrongSignatureOffenders, voter.Identity)
		offenders[voter.Signing] = true
	}

	logger.Warn("Offenders: reveal %d, wrong signature %d, bitVote %d", len(revealOffenders), len(wrongSignatureOffenders), len(bitVoteOffenders))

	return offenders
}

func getFdcPenalties(reward *big.Int, penaltyFactor *big.Int, offenders map[ty2.VoterSigning]bool, voters *fsp.VoterIndex) []ty.RewardClaim {
	var penalties []ty.RewardClaim

	bigTotalSigningWeight := big.NewInt(int64(voters.TotalSigningPolicyWeight))

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

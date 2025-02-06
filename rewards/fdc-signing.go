package rewards

import (
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"math/big"
)

func generateFdcSigningClaims(finalizations []*data.Finalization, round ty.RoundId, reward *big.Int, bitVotes map[ty.VoterSubmit]*big.Int, consensusBitVote *big.Int, consensusSigs map[ty.VoterSigning]data.SigInfo, voterIndex *data.VoterIndex) []ty.RewardClaim {
	var signingClaims []ty.RewardClaim

	if consensusBitVote == nil || consensusBitVote.Cmp(big.NewInt(0)) == 0 {
		logger.Warn("no consensus bitVote for round %d, burning rewards", round)
		return []ty.RewardClaim{burnClaim(reward)}
	}

	successfulFinalization := firstSuccessful(finalizations)

	revealDeadline := params.Net.Epoch.RevealDeadlineSec(round + 1)
	roundEnd := params.Net.Epoch.VotingRoundEndSec(
		round.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows),
	)

	deadline := min(
		successfulFinalization.Info.TimestampSec,
		roundEnd,
	)
	gracePeriod := revealDeadline + params.Net.Ftso.GracePeriodForSignaturesDurationSec

	signersToReward := map[ty.VoterSigning]data.SigInfo{}
	for voter, sig := range consensusSigs {
		if sig.Timestamp <= gracePeriod || sig.Timestamp <= deadline {
			signersToReward[voter] = sig
		} else {
			logger.Warn("signer %s is late for round %d", voter, round)
		}
	}

	undistributedWeight := big.NewInt(int64(voterIndex.TotalSigningPolicyWeight))
	undistributedAmount := big.NewInt(0).Set(reward)

	for index, voter := range voterIndex.PolicyOrder {
		weight := big.NewInt(int64(voter.SigningPolicyWeight))

		if undistributedWeight.Cmp(big.NewInt(0)) == 0 {
			logger.Fatal("no weight for signer %s, index %d", voter.Signing, index)
		}

		voterAmount := big.NewInt(0).Div(
			big.NewInt(0).Mul(undistributedAmount, weight),
			undistributedWeight,
		)

		_, ok := signersToReward[voter.Signing]
		if ok {
			undistributedWeight.Sub(undistributedWeight, weight)
			undistributedAmount.Sub(undistributedAmount, voterAmount)

			bitVote, _ := bitVotes[voter.Submit]

			if !dominatesConsensusBitVote(bitVote, consensusBitVote) {
				burnAmount := big.NewInt(0).Div(big.NewInt(0).Mul(voterAmount, big.NewInt(200000)),
					bigTotalPPM)
				if burnAmount.Cmp(bigZero) >= 0 {
					signingClaims = append(signingClaims, burnClaim(burnAmount))
					voterAmount.Sub(voterAmount, burnAmount)
				}
			}
			signingClaims = append(signingClaims, SigningWeightClaimsForVoter(voter, voterAmount)...)
		}
	}

	if undistributedAmount.Cmp(big.NewInt(0)) > 0 {
		logger.Warn("undistributed amount %s for round %d", undistributedAmount, round)
		signingClaims = append(signingClaims, burnClaim(undistributedAmount))
	}

	return signingClaims
}

package rewards

import (
	"encoding/hex"
	"fsp-rewards-calculator/common/fdc"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	ty2 "fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"gorm.io/gorm"
	"math/big"
	"slices"
)

type roundReward struct {
	amount *big.Int
	burn   *big.Int
}

func GetFdcRewards(db *gorm.DB, re *fsp.RewardEpoch, submit2 []payload.Message, submitSignatures []payload.Message, finalizations []*fsp.Finalization) ([]ty.RewardClaim, map[ty2.VoterId]bool) {
	bitVotesByRound := fdc.ExtractBitVotes(submit2)
	finalizationsByRound := fsp.GetFinalizationsByRound(finalizations)

	consensusHashByRound := map[ty2.RoundId]common.Hash{}
	for round, fs := range finalizationsByRound {
		first := firstSuccessful(fs)
		if first == nil {
			continue
		}
		consensusHashByRound[round] = first.MerkleRoot.EncodedHash()
	}

	signersByRound := fdc.GetSignersByRound(submitSignatures, consensusHashByRound, re)
	attestationRequestsByRound := fdc.GetAttestationRequestsByRound(db, re.StartRound, re.EndRound)

	consensusBitVoteByRound := map[ty2.RoundId]*big.Int{}
	countByType := map[string]int{}
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

		consensusBitVote := getConsensusBitVote(consensusSigs, round, re.VoterIndex)
		logger.Debug("Consensus bitVote for round %d: %d", round, consensusBitVote)
		consensusBitVoteByRound[round] = consensusBitVote

		// Get appearances by type
		for i, request := range attestationRequestsByRound[round] {
			if len(request.Data) < 64 {
				logger.Debug("attestation request malformed: data less than 64 bytes")
				continue
			}
			if !isConfirmed(i, consensusBitVote) {
				continue
			}
			t := hex.EncodeToString(request.Data[0:64])
			countByType[t]++
		}
	}

	rewardByRound := calculateFdcRoundRewards(re, countByType, attestationRequestsByRound, consensusBitVoteByRound)

	rewardedRounds := map[ty2.VoterId]int{}
	totalRewardedRounds := 0

	epochClaims := make([]ty.RewardClaim, 0)
	for round := re.StartRound; round <= re.EndRound; round++ {
		var roundClaims []ty.RewardClaim
		roundClaims = append(roundClaims, burnClaim(rewardByRound[round].burn))
		utils.PrintRoundResults(roundClaims, re.Epoch, round, "fdc-claimback")

		if firstSuccessful(finalizationsByRound[round]) == nil {
			logger.Warn("no successful finalization for round %d, burning round rewards", round)
			roundClaims = append(roundClaims, burnClaim(rewardByRound[round].amount))
		} else {
			reward := rewardByRound[round].amount
			finalizationReward := big.NewInt(0).Div(big.NewInt(0).Mul(reward, params.Net.Fdc.FinalizationBips), bigTotalBips)
			signingReward := big.NewInt(0).Sub(reward, finalizationReward)

			finalizers, err := selectFinalizers(round, re.Policy, params.Net.Fdc.ProtocolId, params.Net.Ftso.FinalizationVoterSelectionThresholdWeightBips)
			if err != nil {
				logger.Fatal("error selecting finalizers for round %d: %s", round, err)
			}

			var eligibleVoters []*fsp.VoterInfo
			for addr := range finalizers {
				voter, ok := re.VoterIndex.BySigning[ty2.VoterSigning(addr)]
				if ok {
					eligibleVoters = append(eligibleVoters, voter)
				}
			}
			finalizationClaims := getFinalizationClaims(round, finalizationReward, finalizationsByRound[round], eligibleVoters, finalizers)
			logger.Debug("Finalization rewards calculated for round %d: %d", round, len(finalizationClaims))
			roundClaims = append(roundClaims, finalizationClaims...)
			utils.PrintRoundResults(finalizationClaims, re.Epoch, round, "fdc-finalz-claims")

			consensusBitVote := consensusBitVoteByRound[round]
			roundSigs := signersByRound[round]
			hash := consensusHashByRound[round]
			consensusSigs := roundSigs[hash]

			signingClaims := generateFdcSigningClaims(finalizationsByRound[round], round, signingReward, bitVotesByRound[round], consensusBitVote, consensusSigs, re.VoterIndex)
			roundClaims = append(roundClaims, signingClaims...)
			utils.PrintRoundResults(signingClaims, re.Epoch, round, "fdc-signing-claims")

			offenders := getOffenders(bitVotesByRound[round], consensusSigs, roundSigs[fdc.WrongSignatureIndicatorMessageHash], re.VoterIndex, consensusBitVoteByRound[round])

			penalties := getFdcPenalties(reward, params.Net.Fdc.PenaltyFactor, offenders, re.VoterIndex)
			utils.PrintRoundResults(penalties, re.Epoch, round, "fdc-penalties")

			roundClaims = append(roundClaims, penalties...)

			if updateCond(rewardedRounds, re.VoterIndex, signingClaims, finalizationClaims, penalties) {
				totalRewardedRounds++
			}
		}

		utils.PrintRoundResults(roundClaims, re.Epoch, round, "fdc-round-claims")
		epochClaims = append(epochClaims, roundClaims...)
	}

	logger.Debug("FDC rewards calculated for re %d: %d", re.Epoch, len(epochClaims))
	return epochClaims, metFDCCondition(totalRewardedRounds, rewardedRounds)
}

func updateCond(rewardedRounds map[ty2.VoterId]int, index *fsp.VoterIndex, signingClaims []ty.RewardClaim, finalizationClaims []ty.RewardClaim, penalties []ty.RewardClaim) bool {
	roundRewarded := false
	eligibleForRound := map[ty2.VoterId]bool{}
	for _, claim := range append(append([]ty.RewardClaim{}, signingClaims...), finalizationClaims...) {
		beneficiary := ty2.VoterId(claim.Beneficiary)
		if index.ById[beneficiary] == nil {
			continue
		}
		eligibleForRound[beneficiary] = true
		roundRewarded = true
	}
	for _, penalty := range penalties {
		eligibleForRound[ty2.VoterId(penalty.Beneficiary)] = false
	}
	for voter, eligible := range eligibleForRound {
		if eligible {
			rewardedRounds[voter]++
		}
	}
	return roundRewarded
}

func firstSuccessful(finalizations []*fsp.Finalization) *fsp.Finalization {
	successIndex := slices.IndexFunc(finalizations, func(f *fsp.Finalization) bool {
		return !f.Info.Reverted
	})
	if successIndex < 0 {
		return nil
	}
	return finalizations[successIndex]
}

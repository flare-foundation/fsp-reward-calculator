package rewards

import (
	"encoding/hex"
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"gorm.io/gorm"
	"math/big"
	"slices"
)

type FdcMinConditions struct {
	rewardedRounds      map[ty.VoterId]int
	totalRewardedRounds int
}

type roundReward struct {
	amount *big.Int
	burn   *big.Int
}

func GetFdcRewards(db *gorm.DB, re *data.RewardEpoch, submit2 []payload.Message, submitSignatures []payload.Message, finalizations []*data.Finalization) ([]ty.RewardClaim, map[ty.VoterId]bool) {
	bitVotesByRound := data.ExtractBitVotes(submit2)
	finalizationsByRound := data.GetFinalizationsByRound(finalizations)

	consensusHashByRound := map[ty.RoundId]common.Hash{}
	for round, fs := range finalizationsByRound {
		first := firstSuccessful(fs)
		if first == nil {
			continue
		}
		consensusHashByRound[round] = first.MerkleRoot.EncodedHash()
	}

	signersByRound := data.GetFdcSignersByRound(submitSignatures, consensusHashByRound, re)
	attestationRequestsByRound := data.GetAttestationRequestsByRound(db, re.StartRound, re.EndRound)

	consensusBitVoteByRound := map[ty.RoundId]*big.Int{}
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
		logger.Info("Consensus bitVote for round %d: %d", round, consensusBitVote)
		consensusBitVoteByRound[round] = consensusBitVote

		// Get appearances by type
		for i, request := range attestationRequestsByRound[round] {
			if len(request.Data) < 64 {
				logger.Warn("attestation request malformed: data less than 64 bytes")
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

	rewardedRounds := map[ty.VoterId]int{}
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

			var eligibleVoters []*data.VoterInfo
			for addr := range finalizers {
				voter, ok := re.VoterIndex.BySigning[ty.VoterSigning(addr)]
				if ok {
					eligibleVoters = append(eligibleVoters, voter)
				}

			}
			finalizationClaims := getFinalizationClaims(round, finalizationReward, finalizationsByRound[round], eligibleVoters, finalizers)
			logger.Info("Finalization rewards calculated for round %d: %d", round, len(finalizationClaims))
			roundClaims = append(roundClaims, finalizationClaims...)
			utils.PrintRoundResults(finalizationClaims, re.Epoch, round, "fdc-finalz-claims")

			consensusBitVote := consensusBitVoteByRound[round]
			roundSigs, _ := signersByRound[round]
			hash := consensusHashByRound[round]
			consensusSigs, _ := roundSigs[hash]

			signingClaims := generateFdcSigningClaims(finalizationsByRound[round], round, signingReward, bitVotesByRound[round], consensusBitVote, consensusSigs, re.VoterIndex)
			roundClaims = append(roundClaims, signingClaims...)
			utils.PrintRoundResults(signingClaims, re.Epoch, round, "fdc-signing-claims")

			offenders := getOffenders(bitVotesByRound[round], consensusSigs, roundSigs[data.WrongSignatureIndicatorMessageHash], re.VoterIndex, consensusBitVoteByRound[round])

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

	logger.Info("FDC rewards calculated for re %d: %d", re.Epoch, len(epochClaims))
	return epochClaims, metFDCCondition(totalRewardedRounds, rewardedRounds)
}

func updateCond(rewardedRounds map[ty.VoterId]int, index *data.VoterIndex, signingClaims []ty.RewardClaim, finalizationClaims []ty.RewardClaim, penalties []ty.RewardClaim) bool {
	roundRewarded := false
	eligibleForRound := map[ty.VoterId]bool{}
	for _, claim := range append(append([]ty.RewardClaim{}, signingClaims...), finalizationClaims...) {
		beneficiary := ty.VoterId(claim.Beneficiary)
		if index.ById[beneficiary] == nil {
			continue
		}
		eligibleForRound[beneficiary] = true
		roundRewarded = true
	}
	for _, penalty := range penalties {
		eligibleForRound[ty.VoterId(penalty.Beneficiary)] = false
	}
	for voter, eligible := range eligibleForRound {
		if eligible {
			rewardedRounds[voter]++
		}
	}
	return roundRewarded
}

func firstSuccessful(finalizations []*data.Finalization) *data.Finalization {
	successIndex := slices.IndexFunc(finalizations, func(f *data.Finalization) bool {
		return f.Info.Reverted == false
	})
	if successIndex < 0 {
		return nil
	}
	return finalizations[successIndex]
}

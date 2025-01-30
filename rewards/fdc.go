package rewards

import (
	"encoding/binary"
	"encoding/hex"
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
	"slices"
)

type roundReward struct {
	amount *big.Int
	burn   *big.Int
}

func getFdcRewards(db *gorm.DB, epochs data.RewardEpochs, submit2 []payload.Message, submitSignatures []payload.Message, finalizations []*data.Finalization) ([]ty.RewardClaim, error) {
	re := epochs.Current
	_ = data.ExtractBitVotes(submit2)

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

	// get confirmed/duplicate reuqests

	// calculate

	consensusBitVoteByRound := map[ty.RoundId]*big.Int{}
	countByType := map[string]int{} // Used for calculating reward offers

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
		logger.Info("Consensus bitVote for round %d: %d", round, consensusBitVote)
		consensusBitVoteByRound[round] = consensusBitVote

		// Get appearances by type
		for i, r := range attestationRequestsByRound[round] {
			if len(r.Data) < 64 {
				logger.Warn("attestation request malformed: data less than 64 bytes")
			}
			if !isConfirmed(i, consensusBitVote) {
				continue
			}
			t := hex.EncodeToString(r.Data[0:64])
			countByType[t]++
		}
	}

	rewardPerRound := calculateFdcRoundRewards(re, countByType, attestationRequestsByRound, consensusBitVoteByRound)

	for round := re.StartRound; round <= re.EndRound; round++ {
		var roundClaims []ty.RewardClaim
		roundClaims = append(roundClaims, burnClaim(rewardPerRound[round].burn))

		// calc reward offers

		//offenders := getOffenders(bitVotesByRound[round], consensusSigs, roundSigs[data.WrongSignatureIndicatorMessageHash], epochs.Current.VoterIndex, consensusBitVote)
		//logger.Info("Offenders for round %d: %v", round, offenders)

		//penalties := getFdcPenalties(epochs.Current.Reward, epochs.Current.PenaltyFactor, offenders, epochs.Current.VoterIndex)

		epochClaims = append(epochClaims, roundClaims...)
	}

	logger.Info("FDC rewards calculated for epoch %d: %d", re.Epoch, len(epochClaims))

	return epochClaims, nil
}

func calculateFdcRoundRewards(
	re *data.RewardEpoch,
	countByType map[string]int,
	attestationRequestsByRound map[ty.RoundId][]data.AttestationRequest,
	consensusBitVoteByRound map[ty.RoundId]*big.Int,
) map[ty.RoundId]roundReward {
	rewardPerRound := map[ty.RoundId]roundReward{}

	totalBurnAmount := big.NewInt(0)
	totalRewardAmount := big.NewInt(0)

	if len(re.Offers.FdcInflation) == 0 {
		logger.Warn("no inflation offer for reward epoch %d", re.Epoch)
	} else {
		totalWeight := big.NewInt(0)
		burnWeight := big.NewInt(0)

		inflationOffer := re.Offers.FdcInflation[0]
		for _, conf := range inflationOffer.FdcConfigurations {
			t := hex.EncodeToString(append(conf.AttestationType[:], conf.Source[:]...))
			count := countByType[t]
			totalWeight.Add(totalWeight, conf.InflationShare)
			if count < int(conf.MinRequestsThreshold) {
				burnWeight.Add(burnWeight, conf.InflationShare)
			}
		}

		totalBurnAmount = big.NewInt(0).Div(
			big.NewInt(0).Mul(burnWeight, inflationOffer.Amount),
			totalWeight,
		)
		totalRewardAmount = big.NewInt(0).Sub(inflationOffer.Amount, totalBurnAmount)
	}

	logger.Info("Total reward amount %s, total burn amount %s", totalRewardAmount, totalBurnAmount)

	perRound, rem := totalRewardAmount.DivMod(totalRewardAmount, big.NewInt(int64(re.EndRound-re.StartRound+1)), big.NewInt(0))
	burnPerRound, remB := totalBurnAmount.DivMod(totalBurnAmount, big.NewInt(int64(re.EndRound-re.StartRound+1)), big.NewInt(0))

	for round := re.StartRound; round <= re.EndRound; round++ {

		amount := new(big.Int).Set(perRound)
		if big.NewInt(int64(round-re.StartRound)).Cmp(rem) < 0 {
			amount.Add(amount, big.NewInt(1))
		}

		burnAmount := new(big.Int).Set(burnPerRound)
		if big.NewInt(int64(round-re.StartRound)).Cmp(remB) < 0 {
			burnAmount.Add(burnAmount, big.NewInt(1))
		}

		feeAmount := big.NewInt(0)
		feeBurnAmount := big.NewInt(0)

		for i, r := range attestationRequestsByRound[round] {
			if isConfirmed(i, consensusBitVoteByRound[round]) {
				feeAmount.Add(feeBurnAmount, r.MergedFee)
			} else {
				feeBurnAmount.Add(feeBurnAmount, r.MergedFee)
			}
		}

		rewardPerRound[round] = roundReward{
			amount: amount.Add(amount, feeAmount),
			burn:   burnAmount.Add(burnAmount, feeBurnAmount),
		}
	}

	return rewardPerRound
}

func isConfirmed(attestationIndex int, consensusBitVote *big.Int) bool {
	if consensusBitVote == nil {
		return false
	}
	return consensusBitVote.Bit(attestationIndex) == 1
}

func getOffenders(
	bitVotes map[ty.VoterSubmit]data.BitVote,
	consensusSigs map[ty.VoterSigning]data.SigInfo,
	wrongSigs map[ty.VoterSigning]data.SigInfo,
	voterIndex *data.VoterIndex,
	consensusBitVote *big.Int,
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
			bitVote, _ := parseBitVote(sig.UnsignedMessage)
			if consensusBitVote.Cmp(bitVote) != 0 {
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

func parseBitVote(bytes []byte) (*big.Int, error) {
	if len(bytes) < 2 {
		return nil, errors.New("bitVote too short")
	}

	lengthBytes := bytes[0:2]
	length := binary.BigEndian.Uint16(lengthBytes)
	bitVector := new(big.Int).SetBytes(bytes[2:])

	if bitVector.BitLen() > int(length) {
		return nil, errors.New("bitvote length does not match bitvector")
	}
	return bitVector, nil
}

// getConsensusBitVote returns the
func getConsensusBitVote(sigs map[ty.VoterSigning]data.SigInfo, round ty.RoundId, voters *data.VoterIndex) *big.Int {
	bitVoteWeight := map[string]uint64{}
	for signer, sig := range sigs {
		if len(sig.UnsignedMessage) < 3 { // first two bytes are length
			logger.Warn("bitVote message too short for signer %s in round %d", signer, round)
			continue
		}
		bitVote, err := parseBitVote(sig.UnsignedMessage)
		if err != nil {
			logger.Warn("error parsing bitVote for signer %s in round %d: %s", signer, round, err)
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

package rewards

import (
	"encoding/hex"
	"fmt"
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
	"slices"
	"sync"
)

// GetEpochClaims calculates the reward claims for a reward epoch
func GetEpochClaims(db *gorm.DB, epoch ty.EpochId) ([]ty.RewardClaim, error) {
	re, err := data.GetRewardEpoch(epoch, db)
	if err != nil {
		return nil, errors.Wrap(err, "err fetching reward epoch")
	}

	windowStart := ty.RoundId(uint64(re.StartRound) - params.Net.Ftso.RandomGenerationBenchingWindow)
	windowEnd := re.EndRound.Add(params.Net.Ftso.FutureSecureRandomWindow)

	var (
		revealsByRound       map[ty.RoundId]data.RoundReveals
		signersByRound       data.SignerMap
		finalizationsByRound map[ty.RoundId][]*data.Finalization
		fUpdatesByRound      map[ty.RoundId]*data.FUpdate
	)

	revealsByRoundChan := data.GetRoundRevealsAsync(db, windowStart, windowEnd, re)

	var wg sync.WaitGroup
	wg.Add(3)

	// TODO: Better error handling
	go func() {
		defer wg.Done()

		var err error
		signersByRound, err = data.GetSignersByRound(db, re)
		logger.Info("Signers fetched")
		if err != nil {
			logger.Fatal("error calculating signers: %s", err)
		}
	}()
	go func() {
		defer wg.Done()

		var err error
		finalizationsByRound, err = data.GetFinalizationsByRound(db, re)
		logger.Info("Finalizations fetched")
		if err != nil {
			logger.Fatal("err fetching finalizations: %s", err)
		}
	}()
	go func() {
		defer wg.Done()

		var err error
		fUpdatesByRound, err = data.GetFUpdatesByRound(db, re)
		logger.Info("Fast update data fetched")
		if err != nil {
			logger.Fatal("err fetching fast updates: %s", err)
		}
	}()

	revealsByRound = <-revealsByRoundChan
	results, err := data.CalculateResults(re, revealsByRound)
	if err != nil {
		return nil, errors.Wrap(err, "error calculating results")
	}

	feedSelectionRandoms := getFeedSelectionRandoms(re, windowEnd, revealsByRound, results)
	roundRewards := calculateRoundRewards(re, feedSelectionRandoms)
	fuRoundRewards := calculateFURoundRewards(re, feedSelectionRandoms)

	wg.Wait()
	logger.Info("All data fetched, calculating rewards.")

	epochClaims := make([]ty.RewardClaim, 0)

	// Calculate reward claims
	for round := re.StartRound; round <= re.EndRound; round++ {

		totalRoundReward := roundRewards[round]

		logger.Info("Round: %d, total reward: %s, feed: %s", round, totalRoundReward.Amount.String(), hex.EncodeToString(totalRoundReward.Feed.Id[:]))
		logger.Debug("Median: %+v", results[round].Median[totalRoundReward.Feed.Id])

		if totalRoundReward.ShouldBurn {
			epochClaims = append(epochClaims, ty.RewardClaim{
				Beneficiary: burnAddress,
				Amount:      new(big.Int).Set(totalRoundReward.Amount),
				Type:        ty.Direct,
			})
			continue
		}

		signingReward := new(big.Int).Div(
			bigTmp.Mul(totalRoundReward.Amount, params.Net.Ftso.SigningBips),
			bigTotalBips,
		)
		finalizationReward := new(big.Int).Div(
			bigTmp.Mul(totalRoundReward.Amount, params.Net.Ftso.FinalizationBips),
			bigTotalBips,
		)
		medianReward := new(big.Int).Sub(
			totalRoundReward.Amount,
			bigTmp.Add(signingReward, finalizationReward),
		)

		logger.Info("Reward shares for round %d: signing %s, finalization %s, median %s", round, signingReward.String(), finalizationReward.String(), medianReward.String())

		logger.Info("Calculating median claims for round %d", round)
		medianClaims := getMedianClaims(round, re, medianReward, totalRoundReward, results[round].Median[totalRoundReward.Feed.Id])

		utils.PrintResults(medianClaims, fmt.Sprintf("%d-median-claims", round))

		// Only voters receiving median rewards are eligible for signing and finalization rewards
		var eligibleVoters []*data.VoterInfo
		for _, claim := range medianClaims {
			if claim.Type != ty.WNat || claim.Amount.Cmp(bigZero) <= 0 {
				continue
			}
			voter, ok := re.VoterIndex.ByDelegation[ty.VoterDelegation(claim.Beneficiary)]
			if ok {
				eligibleVoters = append(eligibleVoters, voter)
			}
		}
		logger.Info("Calculating signing claims for round %d", round)
		signingClaims := getSigningClaims(round, re, signingReward, eligibleVoters, signersByRound[round], finalizationsByRound[round])

		utils.PrintResults(signingClaims, fmt.Sprintf("%d-signing-claims", round))

		logger.Info("Calculating finalization claims for round %d", round)
		finalizers, err := selectFinalizers(round, re.Policy, params.Net.Ftso.FinalizationVoterSelectionThresholdWeightBips)
		if err != nil {
			return nil, errors.Wrap(err, "error selecting finalizers")
		}
		finalizationClaims := getFinalizationClaims(round, finalizationReward, finalizationsByRound[round], eligibleVoters, finalizers)

		utils.PrintResults(finalizationClaims, fmt.Sprintf("%d-finalz-claims", round))

		dSigners := getDoubleSigners(signersByRound[round])
		var dSignerInfos []*data.VoterInfo
		for dSigner := range dSigners {
			dSignerInfos = append(dSignerInfos, re.VoterIndex.BySigning[dSigner])
		}

		doubleSigningPenalties := getPenalties(totalRoundReward.Amount, params.Net.Ftso.PenaltyFactor, dSignerInfos, re.VoterIndex)

		var offenderInfos []*data.VoterInfo
		for _, offender := range revealsByRound[round].Offenders {
			info := re.VoterIndex.BySubmit[offender]
			if info != nil {
				offenderInfos = append(offenderInfos, re.VoterIndex.BySubmit[offender])
			}
		}
		revealPenalties := getPenalties(totalRoundReward.Amount, params.Net.Ftso.PenaltyFactor, offenderInfos, re.VoterIndex)

		logger.Info("Round: %d, computed median claims: %d, signing claims: %d, finalz claims: %d", round, len(medianClaims), len(signingClaims), len(finalizationClaims))

		var roundClaims []ty.RewardClaim

		roundClaims = append(roundClaims, medianClaims...)
		roundClaims = append(roundClaims, signingClaims...)
		roundClaims = append(roundClaims, finalizationClaims...)
		roundClaims = append(roundClaims, doubleSigningPenalties...)
		roundClaims = append(roundClaims, revealPenalties...)

		utils.PrintResults(doubleSigningPenalties, fmt.Sprintf("%d-doublesig-claims", round))
		utils.PrintResults(revealPenalties, fmt.Sprintf("%d-reveal-claims", round))

		// Fast updates
		reward := fuRoundRewards[round]
		feedId := data.FeedId(reward.FeedConfig.FeedId)
		feedIndex := slices.IndexFunc(re.OrderedFeeds, func(f data.Feed) bool {
			return f.Id == feedId
		})
		if feedIndex == -1 {
			logger.Fatal("FastUpdate feed not found for round %d, feedId %s", round, feedId)
		}
		medianDecimals := int(re.OrderedFeeds[feedIndex].Decimals)
		logger.Info("Calculating FastUpdate claims for round %d, feed %s", round, feedId.Hex())
		fuClaims := gatFUpdateClaims(re, fUpdatesByRound[round], fuRoundRewards[round], results[round].Median[feedId], medianDecimals)
		roundClaims = append(roundClaims, fuClaims...)
		utils.PrintResults(fuClaims, fmt.Sprintf("%d-fu-claims", round))

		logger.Info("Round %d, computed FU claims: %d", round, len(fuClaims))

		utils.PrintResults(roundClaims, fmt.Sprintf("%d-round-claims", round))
		epochClaims = append(epochClaims, roundClaims...)
	}

	return epochClaims, nil
}

func getFeedSelectionRandoms(
	re data.RewardEpoch,
	windowEnd ty.RoundId,
	reveals map[ty.RoundId]data.RoundReveals,
	results map[ty.RoundId]data.RoundResult,
) []*big.Int {
	totalRounds := int64(re.EndRound - re.StartRound + 1)

	feedSelectionRandoms := make([]*big.Int, 0, totalRounds)

	for round := re.StartRound + 1; round <= re.EndRound; round++ {
		if results[round].Random.IsSecure {
			feedRandom := utils.FeedSelectionRandom(results[round].Random.Value, round)
			for len(feedSelectionRandoms) < int(round-re.StartRound) {
				feedSelectionRandoms = append(feedSelectionRandoms, feedRandom)
			}
		}
	}

	var lastRandom *data.RandomResult
	var lastRandomRound ty.RoundId

	for round := re.EndRound + 1; round < windowEnd; round++ {
		validReveals := reveals[round].Reveals

		eligibleReveals := map[ty.VoterSubmit]*data.Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := re.NextVoters.BySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
			}
		}
		random := data.CalculateRandom(round, reveals, eligibleReveals)
		if random.IsSecure {
			lastRandom = &random
			lastRandomRound = round
			break
		}
		logger.Info("Extra random: %d %+v", lastRandom)
	}

	// Random for last round is the first secure random from next reward epoch,
	// or nil if none found within a certain window.
	if lastRandom != nil {
		rnd := utils.FeedSelectionRandom(lastRandom.Value, lastRandomRound)
		for len(feedSelectionRandoms) < int(totalRounds) {
			feedSelectionRandoms = append(feedSelectionRandoms, rnd)
		}
	}
	return feedSelectionRandoms
}

func getPenalties(
	reward *big.Int,
	penaltyFactor *big.Int,
	offenders []*data.VoterInfo,
	voters *data.VoterIndex,
) []ty.RewardClaim {
	var penalties []ty.RewardClaim
	for _, offender := range offenders {
		amount := new(big.Int).Div(
			bigTmp.Mul(offender.CappedWeight, bigTmp.Mul(reward, penaltyFactor)),
			voters.TotalCappedWeight,
		)

		claims := SigningWeightClaimsForVoter(offender, amount)
		// big.Int uses Euclidean division behaves differently when dividing negative numbers compared
		// to BigInt in JS. So we calculate an absolute penalty amount first and then negate it.
		for i := range claims {
			claims[i].Amount.Neg(claims[i].Amount)
		}

		penalties = append(penalties, claims...)
	}
	return penalties
}

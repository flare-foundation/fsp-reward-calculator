package rewards

import (
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"gorm.io/gorm"
	"math/big"
	"slices"
)

type FtsoMinConditions struct {
	Scaling     map[ty.VoterId]bool
	FastUpdates map[ty.VoterId]bool
}

func getFtsoRewards(db *gorm.DB, epochs data.RewardEpochs, windowEnd ty.RoundId, submit1 []payload.Message, submit2 []payload.Message, submitSignatures []payload.Message, finalizations []*data.Finalization) ([]ty.RewardClaim, FtsoMinConditions) {
	var (
		revealsByRound       map[ty.RoundId]data.RoundReveals
		signersByRound       data.SignerMap
		finalizationsByRound map[ty.RoundId][]*data.Finalization
		fUpdatesByRound      map[ty.RoundId]*data.FUpdate
	)

	re := epochs.Current
	revealsByRound = data.GetRoundReveals(submit1, submit2, epochs)

	signersByRound, err := data.GetSignersByRound(submitSignatures, re)
	logger.Info("Signers fetched")
	if err != nil {
		logger.Fatal("error calculating signers: %s", err)
	}

	finalizationsByRound = data.GetFinalizationsByRound(finalizations)

	fUpdatesByRound, err = data.GetFUpdatesByRound(db, re.StartRound, re.EndRound)
	logger.Info("Fast update data fetched")

	results, err := data.CalculateResults(re.StartRound, re.EndRound, re, revealsByRound)
	if err != nil {
		logger.Fatal("error calculating results: %s", err)
	}

	feedSelectionRandoms := getFeedSelectionRandoms(epochs, windowEnd, revealsByRound, results)
	roundRewards := calculateRoundRewards(re, feedSelectionRandoms)
	fuRoundRewards := calculateFURoundRewards(re, feedSelectionRandoms)

	for round := re.StartRound; round <= re.EndRound; round++ {
		data.PrintRoundData(results[round], revealsByRound[round], roundRewards[round].Feed, feedSelectionRandoms[round-re.StartRound], re.Epoch, round)
	}

	logger.Info("All data fetched, calculating rewards.")

	// Calculate reward claims
	epochClaims := make([]ty.RewardClaim, 0)
	for round := re.StartRound; round <= re.EndRound; round++ {
		totalRoundReward := roundRewards[round]

		if totalRoundReward.Feed != nil {
			logger.Info("Round: %d, total reward: %s, feed: %s", round, totalRoundReward.Amount.String(), totalRoundReward.Feed.Id[:])
			logger.Debug("Median: %+v", results[round].Median[totalRoundReward.Feed.Id])
		} else {
			logger.Info("Round: %d, total reward: %s, burned", round, totalRoundReward.Amount.String())
		}

		if totalRoundReward.ShouldBurn {
			epochClaims = append(epochClaims, ty.RewardClaim{
				Beneficiary: BurnAddress,
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

		utils.PrintRoundResults(medianClaims, re.Epoch, round, "median-claims")

		// Only voters receiving median rewards are eligible for signing and finalization rewards
		var eligibleVoters []*data.VoterInfo
		for _, claim := range medianClaims {
			if claim.Type != ty.WNat || claim.Amount.Cmp(BigZero) <= 0 {
				continue
			}
			voter, ok := re.VoterIndex.ByDelegation[ty.VoterDelegation(claim.Beneficiary)]
			if ok {
				eligibleVoters = append(eligibleVoters, voter)
			}
		}
		logger.Info("Calculating signing claims for round %d", round)
		signingClaims := getSigningClaims(round, re, signingReward, eligibleVoters, signersByRound[round], finalizationsByRound[round])

		utils.PrintRoundResults(signingClaims, re.Epoch, round, "signing-claims")

		logger.Info("Calculating finalization claims for round %d", round)
		finalizers, err := selectFinalizers(round, re.Policy, params.Net.Ftso.ProtocolId, params.Net.Ftso.FinalizationVoterSelectionThresholdWeightBips)
		if err != nil {
			logger.Fatal("error selecting finalizers: %s", err)
		}
		finalizationClaims := getFinalizationClaims(round, finalizationReward, finalizationsByRound[round], eligibleVoters, finalizers)

		utils.PrintRoundResults(finalizationClaims, re.Epoch, round, "finalz-claims")

		dSigners := getDoubleSigners(signersByRound[round])
		var dSignerInfos []*data.VoterInfo
		for dSigner := range dSigners {
			dSignerInfos = append(dSignerInfos, re.VoterIndex.BySigning[dSigner])
		}

		doubleSigningPenalties := getPenalties(totalRoundReward.Amount, params.Net.Ftso.PenaltyFactor, dSignerInfos, re.VoterIndex)

		var offenderInfos []*data.VoterInfo
		for _, offender := range revealsByRound[round].RegisteredOffenders {
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

		utils.PrintRoundResults(doubleSigningPenalties, re.Epoch, round, "doublesig-pen")
		utils.PrintRoundResults(revealPenalties, re.Epoch, round, "reveal-pen")

		// Fast updates
		reward := fuRoundRewards[round]
		feedId := ty.FeedId(reward.FeedConfig.FeedId)
		feedIndex := slices.IndexFunc(re.OrderedFeeds, func(f ty.Feed) bool {
			return f.Id == feedId
		})
		if feedIndex == -1 { // Check if feed was renamed
			logger.Warn("FastUpdate feed not found for round %d, feedId %s, checking if renamed", round, feedId)

			feedId = params.OldToNewFeed[feedId]
			feedIndex = slices.IndexFunc(re.OrderedFeeds, func(f ty.Feed) bool {
				return f.Id == feedId
			})
			if feedIndex == -1 {
				logger.Fatal("FastUpdate feed not found for round %d, feedId %s", round, feedId)
			} else {
				logger.Warn("Using renamed feed %s", feedId)
			}
		}
		medianDecimals := int(re.OrderedFeeds[feedIndex].Decimals)
		logger.Info("Calculating FastUpdate claims for round %d, feed %s", round, feedId.Hex())

		median := results[round].Median[feedId]
		if median == nil {
			logger.Fatal("Median not found for round %d, fast updates feed %s", round, feedId.Hex())
		}

		fuClaims := gatFUpdateClaims(re, fUpdatesByRound[round], fuRoundRewards[round], results[round].Median[feedId], medianDecimals)
		roundClaims = append(roundClaims, fuClaims...)
		utils.PrintRoundResults(fuClaims, re.Epoch, round, "fu-claims")

		logger.Info("Round %d, computed FU claims: %d", round, len(fuClaims))
		utils.PrintRoundResults(roundClaims, re.Epoch, round, "round-claims")

		epochClaims = append(epochClaims, roundClaims...)
	}

	var cond = metFtsoCondition(re.VoterIndex, len(re.OrderedFeeds), results)
	var fuCond = metFUCondition(re.VoterIndex, fUpdatesByRound)

	return epochClaims, FtsoMinConditions{cond, fuCond}
}

func getFeedSelectionRandoms(
	epochs data.RewardEpochs,
	windowEnd ty.RoundId,
	reveals map[ty.RoundId]data.RoundReveals,
	results map[ty.RoundId]data.RoundResult,
) []*big.Int {
	re := epochs.Current
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
		voterIndex := epochs.EpochForRound(round).VoterIndex
		for voter, reveal := range validReveals {
			if _, ok := voterIndex.BySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
			}
		}
		random := data.CalculateRandom(round, reveals, eligibleReveals)
		if random.IsSecure {
			lastRandom = &random
			lastRandomRound = round
			break
		}
	}

	logger.Info("Extra random: %+v", lastRandom)

	var rnd *big.Int
	// Random for last round is the first secure random from next reward epoch,
	// or nil if none found within a certain window.
	if lastRandom != nil {
		rnd = utils.FeedSelectionRandom(lastRandom.Value, lastRandomRound)
	}
	for len(feedSelectionRandoms) < int(totalRounds) {
		feedSelectionRandoms = append(feedSelectionRandoms, rnd)
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
		for i := range claims {
			claims[i].Amount.Neg(claims[i].Amount)
		}

		penalties = append(penalties, claims...)
	}
	return penalties
}

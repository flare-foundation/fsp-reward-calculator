package main

import (
	"encoding/hex"
	"flare-common/contracts/fumanager"
	"flare-common/policy"
	"fmt"
	voters "ftsov2-rewarding/lib"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/types"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
	"slices"
	"sync"
)

const totalBips = 10000
const totalPpm = 1000000

var (
	randomMod   = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
	BurnAddress = common.HexToAddress("0x000000000000000000000000000000000000dEaD")

	bigTotalBips = big.NewInt(int64(totalBips))
	bigTotalPPM  = big.NewInt(int64(totalPpm))
	bigZero      = big.NewInt(0)
	// Used for temporary big.Int calculation results
	bigTmp = new(big.Int)
)

type RandomResult struct {
	Round    types.RoundId
	Value    *big.Int
	IsSecure bool
}

type RoundResult struct {
	Round  types.RoundId
	Median map[FeedId]*MedianResult
	Random RandomResult
}

func calculateRewardClaims(db *gorm.DB, epoch types.EpochId) ([]types.RewardClaim, error) {
	re, err := getRewardEpoch(epoch, db)
	if err != nil {
		return nil, errors.Wrap(err, "err fetching reward epoch")
	}

	windowStart := types.RoundId(uint64(re.StartRound) - params.Net.Ftso.RandomGenerationBenchingWindow)
	windowEnd := re.EndRound.Add(params.Net.Ftso.FutureSecureRandomWindow)

	var wgCR sync.WaitGroup // Commits and reveals
	var wgSF sync.WaitGroup // Signatures and finalizations
	var wgFU sync.WaitGroup // Fast updates
	errCR := make(chan error, 2)
	errSF := make(chan error, 2)
	errFU := make(chan error, 1)

	var (
		allCommitsByRound    map[types.RoundId]map[VoterSubmit]*Commit
		allRevealsByRound    map[types.RoundId]map[VoterSubmit]*Reveal
		signersByRound       SignerMap
		finalizationsByRound map[types.RoundId][]*Finalization
		fUpdatesByRound      map[types.RoundId]*FUpdate
	)

	wgCR.Add(2)
	go func() {
		defer wgCR.Done()

		var err error
		logger.Info("Fetching commits for rounds %d-%d", windowStart, windowEnd)
		allCommitsByRound, err = getCommits(db, windowStart, windowEnd)
		logger.Info("All commits fetched")
		if err != nil {
			errCR <- errors.Errorf("error fetching commitsByRound: %s", err)
		}
	}()
	go func() {
		defer wgCR.Done()

		var err error
		logger.Info("Fetching reveals for rounds %d-%d", windowStart, windowEnd)
		allRevealsByRound, err = getReveals(db, windowStart, windowEnd)
		logger.Info("All reveals fetched")
		if err != nil {
			errCR <- errors.Errorf("error fetching revealsByRound: %s", err)
		}
	}()

	wgSF.Add(2)
	go func() {
		defer wgSF.Done()

		var err error
		signersByRound, err = getSignersByRound(db, re)
		logger.Info("All signers fetched")
		if err != nil {
			errSF <- errors.Errorf("error calculating signers: %s", err)
		}
	}()
	go func() {
		defer wgSF.Done()

		var err error
		finalizationsByRound, err = getFinalizationsByRound(db, re)
		logger.Info("All finalizations fetched")
		if err != nil {
			errSF <- errors.Errorf("err fetching finalizations: %s", err)
		}
	}()
	wgFU.Add(1)
	go func() {
		defer wgFU.Done()

		var err error
		fUpdatesByRound, err = getFUpdatesByRound(db, re)
		if err != nil {
			errFU <- errors.Errorf("err fetching fast updates: %s", err)
		}
	}()

	wgCR.Wait()
	close(errCR)
	for err := range errCR {
		if err != nil {
			return nil, err
		}
	}
	logger.Info("All commits and reveals fetched, processing.")

	revealsByRound := getRoundReveals(windowStart, windowEnd, re, allCommitsByRound, allRevealsByRound)

	results, err := calculateResults(re, revealsByRound)
	if err != nil {
		return nil, errors.Wrap(err, "error calculating results")
	}

	feedSelectionRandoms := calculateFeedSelectionRandoms(re, windowEnd, revealsByRound, results)
	roundRewards := calculateRoundRewards(re, feedSelectionRandoms)

	fuRoundRewards := calculateFURoundRewards(re, feedSelectionRandoms)

	epochClaims := make([]types.RewardClaim, 0)

	wgSF.Wait()
	close(errSF)
	for err := range errSF {
		if err != nil {
			return nil, err
		}
	}
	wgFU.Wait()
	close(errFU)
	for err := range errFU {
		if err != nil {
			return nil, err
		}
	}
	logger.Info("s %v", fUpdatesByRound)
	logger.Info("All signers and finalizations fetched, calculating rewards.")

	// Calculate reward claims
	for round := re.StartRound; round <= re.EndRound; round++ {
		totalRoundReward := roundRewards[round]

		logger.Info("Round: %d, total reward: %s, feed: %s", round, totalRoundReward.Amount.String(), hex.EncodeToString(totalRoundReward.Feed.Id[:]))
		logger.Debug("Median: %+v", results[round].Median[totalRoundReward.Feed.Id])

		if totalRoundReward.ShouldBurn {
			epochClaims = append(epochClaims, types.RewardClaim{
				Beneficiary: BurnAddress,
				Amount:      new(big.Int).Set(totalRoundReward.Amount),
				Type:        types.Direct,
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
		medianClaims := calcMedianRewardClaims(round, re, medianReward, totalRoundReward, results[round].Median[totalRoundReward.Feed.Id])

		printResults(medianClaims, fmt.Sprintf("%d-median-claims", round))

		// Only voters receiving median rewards are eligible for signing and finalization rewards
		var eligibleVoters []*VoterInfo
		for _, claim := range medianClaims {
			if claim.Type != types.WNat || claim.Amount.Cmp(bigZero) <= 0 {
				continue
			}
			voter, ok := re.VoterIndex.byDelegation[VoterDelegation(claim.Beneficiary)]
			if ok {
				eligibleVoters = append(eligibleVoters, voter)
			}
		}
		logger.Info("Calculating signing claims for round %d", round)
		signingClaims := calcSigningRewardClaims(round, re, signingReward, eligibleVoters, signersByRound[round], finalizationsByRound[round])

		printResults(signingClaims, fmt.Sprintf("%d-signing-claims", round))

		logger.Info("Calculating finalization claims for round %d", round)
		finalizers, err := selectFinalizers(round, re.Policy, params.Net.Ftso.FinalizationVoterSelectionThresholdWeightBips)
		if err != nil {
			return nil, errors.Wrap(err, "error selecting finalizers")
		}
		finalizationClaims := calcFinalizationRewardClaims(round, finalizationReward, finalizationsByRound[round], eligibleVoters, finalizers)

		printResults(finalizationClaims, fmt.Sprintf("%d-finalz-claims", round))

		dSigners := getDoubleSigners(signersByRound[round])
		var dSignerInfos []*VoterInfo
		for dSigner := range dSigners {
			dSignerInfos = append(dSignerInfos, re.VoterIndex.bySigning[dSigner])
		}

		doubleSigningPenalties := calcPenalties(totalRoundReward.Amount, params.Net.Ftso.PenaltyFactor, dSignerInfos, re.VoterIndex)

		var offenderInfos []*VoterInfo
		for _, offender := range revealsByRound[round].offenders {
			info := re.VoterIndex.bySubmit[offender]
			if info != nil {
				offenderInfos = append(offenderInfos, re.VoterIndex.bySubmit[offender])
			}
		}
		revealPenalties := calcPenalties(totalRoundReward.Amount, params.Net.Ftso.PenaltyFactor, offenderInfos, re.VoterIndex)

		logger.Info("Round: %d, computed median claims: %d, signing claims: %d, finalz claims: %d", round, len(medianClaims), len(signingClaims), len(finalizationClaims))

		var roundClaims []types.RewardClaim

		roundClaims = append(roundClaims, medianClaims...)
		roundClaims = append(roundClaims, signingClaims...)
		roundClaims = append(roundClaims, finalizationClaims...)
		roundClaims = append(roundClaims, doubleSigningPenalties...)
		roundClaims = append(roundClaims, revealPenalties...)

		printResults(doubleSigningPenalties, fmt.Sprintf("%d-doublesig-claims", round))
		printResults(revealPenalties, fmt.Sprintf("%d-reveal-claims", round))

		// Fast updates
		reward := fuRoundRewards[round]
		feedId := FeedId(reward.FeedConfig.FeedId)
		feedIndex := slices.IndexFunc(re.OrderedFeeds, func(f Feed) bool {
			return f.Id == feedId
		})
		if feedIndex == -1 {
			logger.Fatal("FastUpdate feed not found for round %d, feedId %s", round, feedId)
		}
		medianDecimals := int(re.OrderedFeeds[feedIndex].Decimals)
		logger.Info("Calculating FastUpdate claims for round %d, feed %s", round, feedId.Hex())
		fuClaims := calculateFUpdateClaims(re, fUpdatesByRound[round], fuRoundRewards[round], results[round].Median[feedId], medianDecimals)
		roundClaims = append(roundClaims, fuClaims...)
		printResults(fuClaims, fmt.Sprintf("%d-fu-claims", round))

		logger.Info("Round %d, computed FU claims: %d", round, len(fuClaims))

		printResults(roundClaims, fmt.Sprintf("%d-round-claims", round))
		checkRoundClaims(round, roundClaims)
		epochClaims = append(epochClaims, roundClaims...)
	}

	return epochClaims, nil
}

type RoundReveals struct {
	reveals   map[VoterSubmit]*Reveal
	offenders []VoterSubmit
}

func getRoundReveals(
	windowStart types.RoundId,
	windowEnd types.RoundId,
	re RewardEpoch,
	allCommitsByRound map[types.RoundId]map[VoterSubmit]*Commit,
	allRevealsByRound map[types.RoundId]map[VoterSubmit]*Reveal,
) map[types.RoundId]RoundReveals {
	roundData := map[types.RoundId]RoundReveals{}

	for round := windowStart; round < windowEnd; round++ {
		var voterIndex *VoterIndex
		switch {
		case round < re.StartRound:
			voterIndex = re.PrevVoters
		case round > re.EndRound:
			voterIndex = re.NextVoters
		default:
			voterIndex = re.VoterIndex
		}

		validCommits := map[VoterSubmit]*Commit{}
		for voter, commit := range allCommitsByRound[round] {
			if voterIndex.bySubmit[voter] != nil {
				validCommits[voter] = commit
			}
		}

		validReveals := map[VoterSubmit]*Reveal{}
		for voter, reveal := range allRevealsByRound[round] {
			if voterIndex.bySubmit[voter] != nil {
				validReveals[voter] = reveal
			}
		}

		var offenders []VoterSubmit
		matchingReveals := map[VoterSubmit]*Reveal{}

		for voter, commit := range validCommits {
			reveal, ok := validReveals[voter]
			if !ok {
				logger.Debug("voter %s committed but did not reveal", common.Address(voter))
				offenders = append(offenders, voter)
				continue
			}

			expected := utils.CommitHash(common.Address(voter), uint32(round), reveal.Random, reveal.EncodedValues)

			if expected.Cmp(commit.Hash) != 0 {
				logger.Debug("voter %s reveal hash did not match commit: %s != %s", common.Address(voter), expected.String(), commit.Hash.String())
				offenders = append(offenders, voter)
				continue
			}

			matchingReveals[voter] = reveal
		}

		roundData[round] = RoundReveals{
			reveals:   matchingReveals,
			offenders: offenders,
		}
	}

	return roundData
}

// calculateFURoundRewards total FastUpdates reward offer share per round
func calculateFURoundRewards(re RewardEpoch, feedSelectionRandoms []*big.Int) map[types.RoundId]FUFeedReward {
	totalReward := big.NewInt(0)
	for i := range re.Offers.fastUpdates {
		totalReward.Add(totalReward, re.Offers.fastUpdates[i].Amount)
	}
	for i := range re.Offers.fastUpdatesI {
		totalReward.Add(totalReward, re.Offers.fastUpdatesI[i].OfferAmount)
	}

	roundRewards := make(map[types.RoundId]FUFeedReward)

	perRound, rem := totalReward.DivMod(totalReward, big.NewInt(int64(re.EndRound-re.StartRound+1)), big.NewInt(0))

	feedConfigs := re.Offers.fastUpdates[0].FeedConfigurations
	numFeeds := big.NewInt(int64(len(feedConfigs)))

	for round := re.StartRound; round <= re.EndRound; round++ {
		random := feedSelectionRandoms[round-re.StartRound]

		logger.Info("[FU] Selected random for round %d: %d", round, random)

		amount := new(big.Int).Set(perRound)
		if big.NewInt(int64(round-re.StartRound)).Cmp(rem) < 0 {
			amount.Add(amount, big.NewInt(1))
		}

		if random == nil {
			roundRewards[round] = FUFeedReward{
				Amount:     amount,
				ShouldBurn: true,
			}
			logger.Info("[FU] No secure random found for round %d, burning reward", round)
			continue
		}

		feedIndex := new(big.Int).Mod(random, numFeeds).Uint64()

		roundRewards[round] = FUFeedReward{
			FeedIndex:  feedIndex,
			FeedConfig: &feedConfigs[feedIndex],
			Amount:     amount,
		}
	}
	return roundRewards
}

func checkRoundClaims(round types.RoundId, claims []types.RewardClaim) {
	total := big.NewInt(0)
	for _, claim := range claims {
		total.Add(total, claim.Amount)
	}
	logger.Info("Round %d total claims: %s", round, total.String())

}

// calculateRoundRewards total reward offer share per round
func calculateRoundRewards(re RewardEpoch, feedSelectionRandoms []*big.Int) map[types.RoundId]FeedReward {
	totalReward := big.NewInt(0)
	for i := range re.Offers.inflation {
		offer := re.Offers.inflation[i]
		totalReward.Add(totalReward, offer.Amount)
	}
	for i := range re.Offers.community {
		totalReward.Add(totalReward, re.Offers.community[i].Amount)
	}

	roundRewards := make(map[types.RoundId]FeedReward)

	perRound, rem := totalReward.DivMod(totalReward, big.NewInt(int64(re.EndRound-re.StartRound+1)), big.NewInt(0))
	numFeeds := big.NewInt(int64(len(re.OrderedFeeds)))
	// TODO: Can reduce allocations in loop by re-using big.Int vars OR use uint64 if safe
	for round := re.StartRound; round <= re.EndRound; round++ {
		random := feedSelectionRandoms[round-re.StartRound]

		logger.Info("Selected random for round %d: %d", round, random)

		amount := new(big.Int).Set(perRound)
		if big.NewInt(int64(round-re.StartRound)).Cmp(rem) < 0 {
			amount.Add(amount, big.NewInt(1))
		}

		if random == nil {
			roundRewards[round] = FeedReward{
				Amount:     amount,
				ShouldBurn: true,
			}
			logger.Info("No secure random found for round %d, burning reward", round)
			continue
		}

		feedIndex := new(big.Int).Mod(random, numFeeds).Uint64()

		randomFeed := &re.OrderedFeeds[feedIndex]

		roundRewards[round] = FeedReward{
			Feed:   randomFeed,
			Amount: amount,
		}
	}
	return roundRewards
}

func calculateFeedSelectionRandoms(
	re RewardEpoch,
	windowEnd types.RoundId,
	reveals map[types.RoundId]RoundReveals,
	results map[types.RoundId]RoundResult,
) []*big.Int {
	totalRounds := int64(re.EndRound - re.StartRound + 1)

	feedSelectionRandoms := make([]*big.Int, 0, totalRounds)

	for round := re.StartRound + 1; round <= re.EndRound; round++ {
		logger.Info("Calculating feed selection random for round %d", round)

		if results[round].Random.IsSecure {
			feedRandom := utils.FeedSelectionRandom(results[round].Random.Value, round)
			for len(feedSelectionRandoms) < int(round-re.StartRound) {
				feedSelectionRandoms = append(feedSelectionRandoms, feedRandom)
			}
		}
		//feedSelectionRandoms[round-re.StartRound] = utils.FeedSelectionRandom(results[round+1].Random.Value, round+1)
	}

	var lastRandom *RandomResult
	var lastRandomRound types.RoundId

	for round := re.EndRound + 1; round < windowEnd; round++ {
		validReveals := reveals[round].reveals

		eligibleReveals := map[VoterSubmit]*Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := re.NextVoters.bySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
			}
		}
		random := calculateRandom(round, reveals, eligibleReveals)
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

func calculateResults(
	re RewardEpoch,
	reveals map[types.RoundId]RoundReveals,
) (map[types.RoundId]RoundResult, error) {
	var results = map[types.RoundId]RoundResult{}

	for round := re.StartRound; round <= re.EndRound; round++ {
		validReveals := reveals[round].reveals

		logger.Info("Reveals for round %d: %d", round, len(validReveals))

		eligibleReveals := map[VoterSubmit]*Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := re.VoterIndex.bySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
			}
		}

		logger.Info("Eligible reveals for round %d: %d", round, len(eligibleReveals))

		// Median
		feedValues := map[VoterSubmit][]FeedValue{}
		for voter, reveal := range eligibleReveals {
			values, err := DecodeFeedValues(reveal.EncodedValues, re.OrderedFeeds)
			if err != nil {
				logger.Error("error decoding feed values for voter %s: %s", voter, err)
				continue
			}
			feedValues[voter] = values
		}

		logger.Info("Calculating median for round %d", round)

		median, err := CalculateMedians(round, re, feedValues)
		if err != nil {
			return nil, err
		}

		//for _, feed := range re.OrderedFeeds {
		//	jsonData, _ := json.MarshalIndent(median[feed.Id], "", "    ")
		//
		//	logger.Info("Feed: %s, median: %s", feed.Id.Hex(), jsonData)
		//}

		//logger.Info("Median: %+v", median)

		random := calculateRandom(round, reveals, eligibleReveals)
		logger.Info("Round %d, random result: %d", round, random.Value)

		results[round] = RoundResult{
			Round:  round,
			Median: median,
			Random: random,
		}
	}
	return results, nil
}

func calcPenalties(
	reward *big.Int,
	penaltyFactor *big.Int,
	offenders []*VoterInfo,
	voters *VoterIndex,
) []types.RewardClaim {
	var penalties []types.RewardClaim
	for _, offender := range offenders {
		amount := new(big.Int).Div(
			bigTmp.Mul(offender.CappedWeight, bigTmp.Mul(reward, penaltyFactor)),
			voters.totalCappedWeight,
		)

		claims := signingWeightClaimsForVoter(offender, amount)
		// big.Int uses Euclidian division behaves differently when dividing negative numbers compared
		// to BigInt in JS. So we calculate an absolute penalty amount first and then negate it.
		for i := range claims {
			claims[i].Amount.Neg(claims[i].Amount)
		}

		penalties = append(penalties, claims...)
	}
	return penalties
}

func selectFinalizers(
	round types.RoundId,
	policy *policy.SigningPolicy,
	threshold uint16,
) (map[common.Address]bool, error) {
	// TODO: We have duplicate VoterSet definitions
	seed := voters.InitialHashSeed(policy.Seed, params.Net.Ftso.ProtocolId, uint32(round))
	vs := voters.NewVoterSet(policy.Voters.Voters, policy.Voters.Weights)
	res, err := vs.RandomSelectThresholdWeightVoters(seed, threshold)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting finalizers")
	}

	selected := map[common.Address]bool{}
	for voter := range res.Iter() {
		selected[voter] = true
	}

	return selected, nil
}

func burnClaim(amount *big.Int) types.RewardClaim {
	return types.RewardClaim{
		Beneficiary: BurnAddress,
		Amount:      amount,
		Type:        types.Direct,
	}
}

func getDoubleSigners(roundSigners map[common.Hash]map[VoterSigning]SigInfo) map[VoterSigning]bool {
	signed := map[VoterSigning]bool{}
	doubleSigners := map[VoterSigning]bool{}

	for _, signers := range roundSigners {
		for signer := range signers {
			if _, ok := signed[signer]; ok {
				doubleSigners[signer] = true
			}
			signed[signer] = true
		}
	}

	return doubleSigners
}

type FeedReward struct {
	Feed       *Feed
	Amount     *big.Int
	ShouldBurn bool
}

type FUFeedReward struct {
	FeedIndex  uint64
	FeedConfig *fumanager.IFastUpdatesConfigurationFeedConfiguration
	Amount     *big.Int
	ShouldBurn bool
}

func calculateRandom(round types.RoundId, reveals map[types.RoundId]RoundReveals, eligibleReveals map[VoterSubmit]*Reveal) RandomResult {
	benchingWindowOffenders := map[VoterSubmit]bool{}
	for i := types.RoundId(uint64(round) - params.Net.Ftso.RandomGenerationBenchingWindow); i < round; i++ {
		for j := range reveals[i].offenders {
			benchingWindowOffenders[reveals[i].offenders[j]] = true
		}
	}

	nonBenchedOffenders := 0
	for k := range reveals[round].offenders {
		currentOffender := reveals[round].offenders[k]
		if _, ok := benchingWindowOffenders[currentOffender]; !ok {
			nonBenchedOffenders++
		}
	}
	validCount := 0
	random := big.NewInt(0)
	for voter, reveal := range eligibleReveals {
		if _, ok := benchingWindowOffenders[voter]; !ok {
			random.Add(random, new(big.Int).SetBytes(reveal.Random[:]))
			validCount++
		}
	}
	random.Mod(random, randomMod)

	res := RandomResult{
		Round:    round,
		Value:    random,
		IsSecure: nonBenchedOffenders == 0 && validCount >= params.Net.Ftso.NonBenchedRandomVotersMinCount,
	}
	return res
}

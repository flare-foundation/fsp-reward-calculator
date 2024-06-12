package main

import (
	"flare-common/database"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/types"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
)

func main() {
	config := database.DBConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "flare_ftso_indexer",
		Username: "root",
		Password: "root",
	}

	db, err := database.Connect(&config)

	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}

	_, err = calculateRewards(db, 2745)
	if err != nil {
		logger.Fatal("Error calculating rewards: %s", err)
		return
	}
}

var randomMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
var BurnAddress = common.HexToAddress("0x000000000000000000000000000000000000dEaD")

type ClaimType int

// TODO: Check correct
const (
	Direct ClaimType = iota
	Fee
	WNat
	Mirror
	CChain
)

var (
	totalBips    uint16 = 10000
	bigTotalBips        = big.NewInt(int64(totalBips))
	bigZero             = big.NewInt(0)
	// Used for temporary big.Int calculations
	bigTmp = new(big.Int)
)

const totalPpm = 1000000

type RewardClaim struct {
	//Epoch       types.EpochId
	//Round 	 types.RoundId
	Beneficiary common.Address
	Amount      *big.Int
	Type        ClaimType
}

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

func calculateRewards(db *gorm.DB, epoch types.EpochId) ([]RewardClaim, error) {
	re, err := getRewardEpoch(epoch, db)
	if err != nil {
		return nil, errors.Wrap(err, "err fetching reward epoch")
	}

	windowStart := types.RoundId(uint64(re.StartRound) - params.Coston.Ftso.RandomGenerationBenchingWindow)
	windowEnd := re.EndRound.Add(params.Coston.Ftso.FutureSecureRandomWindow)

	commitsByRound, err := getCommits(db, windowStart, windowEnd)
	if err != nil {
		return nil, errors.Errorf("error fetching commitsByRound: %s", err)
	}
	revealsByRound, err := getReveals(db, re.StartRound, windowEnd)
	if err != nil {
		return nil, errors.Errorf("error fetching revealsByRound: %s", err)
	}

	offendersByRound := map[types.RoundId][]VoterSubmit{}
	matchingRevealsByRound := map[types.RoundId]map[VoterSubmit]*Reveal{}

	for round := windowStart; round < windowEnd; round++ {
		var offenders []VoterSubmit
		matchingReveals := map[VoterSubmit]*Reveal{}

		commits := commitsByRound[round]
		reveals := revealsByRound[round]

		for voter, commit := range commits {
			reveal, ok := reveals[voter]
			if !ok {
				logger.Debug("voter %s committed but did not reveal", common.Address(voter))
				offenders = append(offenders, voter)
				continue
			}

			expected, err := utils.CommitHash(common.Address(voter), uint32(round), reveal.Random, reveal.EncodedValues)
			if err != nil {
				return nil, errors.Errorf("error computing reveal hash: %s", err)
			}

			if expected.Cmp(commit.Hash) != 0 {
				logger.Debug("voter %s reveal hash did not match commit: %s != %s", voter, expected.String(), commit.Hash.String())
				offenders = append(offenders, voter)
				continue
			}

			matchingReveals[voter] = reveal
		}

		offendersByRound[round] = offenders
		matchingRevealsByRound[round] = matchingReveals
	}

	results := map[types.RoundId]RoundResult{}

	for round := re.StartRound; round < re.EndRound; round++ {
		validReveals := matchingRevealsByRound[round]

		eligibleReveals := map[VoterSubmit]*Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := re.Voters.bySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
			}
		}

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
		median, err := calculateMedians(re, feedValues)
		if err != nil {
			return nil, err
		}
		logger.Info("Median: %+v", median)

		// Value calc
		res := calculateRandom(round, offendersByRound, eligibleReveals)
		logger.Info("Value result: %+v", res)

		results[round] = RoundResult{
			Round:  round,
			Median: median,
		}
	}

	var lastRandom *RandomResult

	for round := re.EndRound + 1; round < windowEnd; round++ {
		validReveals := matchingRevealsByRound[round]

		eligibleReveals := map[VoterSubmit]*Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := re.NextVoters.bySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
			}
		}
		random := calculateRandom(round, offendersByRound, eligibleReveals)
		if random.IsSecure {
			lastRandom = &random
			break
		}
		logger.Info("Extra random: %d %+v", lastRandom)
	}

	totalRounds := int64(re.EndRound - re.StartRound + 1)

	feedSelectionRandoms := make([]*big.Int, totalRounds)
	for i := re.StartRound; i < re.EndRound; i++ {
		feedSelectionRandoms[i-re.StartRound] = results[i].Random.Value
	}
	// Random for last round is the first secure random from next reward epoch,
	// or nil if none found within a certain window.
	if lastRandom != nil {
		feedSelectionRandoms = append(feedSelectionRandoms, lastRandom.Value)
	}

	// Calculate reward offer share for round
	totalReward := big.NewInt(0)
	for i := range re.Offers.inflation {
		offer := re.Offers.inflation[i]
		totalReward.Add(totalReward, offer.Amount)
	}
	for i := range re.Offers.community {
		totalReward.Add(totalReward, re.Offers.community[i].Amount)
	}

	roundRewards := make(map[types.RoundId]FeedReward)

	perRound, rem := totalReward.DivMod(totalReward, big.NewInt(totalRounds), nil)
	numFeeds := big.NewInt(int64(len(re.OrderedFeeds)))
	// TODO: Can reduce allocations in loop by re-using big.Int vars OR use uint64 if safe
	for round := re.StartRound; round <= re.EndRound; round++ {
		random := feedSelectionRandoms[round-re.StartRound]

		if random == nil {
			roundRewards[round] = FeedReward{
				ShouldBurn: true,
			}
			logger.Info("No secure random found for round %d, burning reward", round)
			continue
		}

		feedIndex := new(big.Int).Mod(random, numFeeds).Uint64()

		randomFeed := &re.OrderedFeeds[feedIndex]

		amount := big.NewInt(0).Set(perRound)
		if big.NewInt(int64(round-re.StartRound)).Cmp(rem) < 0 {
			amount.Add(amount, big.NewInt(1))
		}
		roundRewards[round] = FeedReward{
			Feed:   randomFeed,
			Amount: amount,
		}
	}

	epochClaims := make([]RewardClaim, 0)

	// Calculate reward claims
	for round := re.StartRound; round <= re.EndRound; round++ {
		totalReward := roundRewards[round]
		if totalReward.ShouldBurn {
			epochClaims = append(epochClaims, RewardClaim{
				Beneficiary: BurnAddress,
				Amount:      big.NewInt(0).Set(totalReward.Amount),
				Type:        Direct,
			})
			continue
		}

		signingReward := big.NewInt(0).Div(
			bigTmp.Mul(totalReward.Amount, params.Coston.Ftso.SigningBips),
			bigTotalBips,
		)
		finalizationReward := big.NewInt(0).Div(
			bigTmp.Mul(totalReward.Amount, params.Coston.Ftso.FinalizationBips),
			bigTotalBips,
		)
		medianReward := big.NewInt(0).Sub(
			totalReward.Amount,
			bigTmp.Add(signingReward, finalizationReward),
		)

		medianClaims := calcMedianRewardClaims(round, re, medianReward, totalReward, results[round].Median[totalReward.Feed.Id])
		epochClaims = append(epochClaims, medianClaims...)

		// Only voters receiving median rewards are eligible for signing and finalization rewards
		var eligibleVoters []*VoterInfo
		for _, claim := range medianClaims {
			if claim.Type != WNat && claim.Amount.Cmp(bigZero) == 0 {
				continue
			}
			voter := re.Voters.byDelegation[VoterDelegation(claim.Beneficiary)]
			eligibleVoters = append(eligibleVoters, voter)
		}

	}

	return epochClaims, nil
}

type FeedReward struct {
	Feed       *Feed
	Amount     *big.Int
	ShouldBurn bool
}

func abs(v int32) uint32 {
	if v < 0 {
		return uint32(-v)
	}
	return uint32(v)
}

func calcMedianRewardClaims(round types.RoundId, re RewardEpoch, rewardShare *big.Int, rewardOffer FeedReward, medianResult *MedianResult) []RewardClaim {
	var epochClaims []RewardClaim

	// Burn rewardOffer if turnout condition not reached
	if medianResult == nil || !isEnoughParticipation(medianResult.ParticipantWeight, re.Voters.totalCappedWeight, rewardOffer.Feed.MinRewardedTurnoutBIPS) {
		epochClaims = append(epochClaims, RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      big.NewInt(0).Set(rewardShare),
			Type:        Direct,
		})
		return epochClaims
	}

	secondaryBandDiff := abs(medianResult.Median) * rewardOffer.Feed.SecondaryBandWidthPPMs / totalPpm
	lowPct := medianResult.Median - int32(secondaryBandDiff)
	highPct := medianResult.Median + int32(secondaryBandDiff)

	lowIQR := medianResult.Q1
	highIQR := medianResult.Q3

	iqrSum := big.NewInt(0) // eligible Weight for IQR rewardOffer
	pctSum := big.NewInt(0) // eligible Weight for PCT rewardOffer

	var voterRecords []voterRecord
	for _, submission := range medianResult.inputValues {
		value := submission.value

		isPct := value > lowPct && value < highPct
		isIqr := (value > lowIQR && value < highIQR) || (value == lowIQR || value == highIQR) && randomSelect(rewardOffer.Feed.Id, round, submission.voter)

		if isPct {
			pctSum.Add(pctSum, submission.weight)
		}
		if isIqr {
			iqrSum.Add(iqrSum, submission.weight)
		}

		voterRecords = append(voterRecords, voterRecord{
			voter:  submission.voter,
			weight: submission.weight,
			isPct:  isPct,
			isIqr:  isIqr,
		})
	}

	totalNormWeight := big.NewInt(0)
	for i, record := range voterRecords {
		newWeight := big.NewInt(0)
		if pctSum.Cmp(bigZero) == 0 {
			if record.isIqr {
				newWeight.Set(record.weight)
			}
		} else {
			if record.isIqr {
				newWeight.Mul(
					big.NewInt(int64(rewardOffer.Feed.PrimaryBandRewardSharePPM)),
					bigTmp.Mul(
						record.weight,
						pctSum,
					),
				)
			}
			if record.isPct {
				newWeight.Add(
					newWeight,
					bigTmp.Mul(
						big.NewInt(int64(totalPpm-rewardOffer.Feed.PrimaryBandRewardSharePPM)),
						bigTmp.Mul(
							record.weight,
							iqrSum,
						),
					),
				)
			}
		}
		voterRecords[i].weight = newWeight
		totalNormWeight.Add(totalNormWeight, newWeight)
	}

	if totalNormWeight.Cmp(bigZero) == 0 {
		// Burn rewardOffer if no eligible submissions
		epochClaims = append(epochClaims, RewardClaim{
			Beneficiary: BurnAddress,
			Amount:      big.NewInt(0).Set(rewardShare),
			Type:        Direct,
		})
		return epochClaims
	}

	totalReward := big.NewInt(0)
	availableReward := big.NewInt(0).Set(rewardShare)
	availableWeight := big.NewInt(0).Set(totalNormWeight)

	var claims []RewardClaim
	for _, record := range voterRecords {
		if record.weight.Cmp(bigZero) == 0 {
			continue
		}
		reward := big.NewInt(0)
		if record.weight.Cmp(bigZero) > 0 {
			if availableWeight.Cmp(bigZero) == 0 {
				logger.Fatal("availableWeight is zero, this should never happen")
			}
			reward.Div(
				bigTmp.Mul(
					record.weight,
					availableReward,
				),
				availableWeight,
			)
		}
		totalReward.Add(totalReward, reward)

		claims = append(claims, generateClaimsForVoter(re.Voters.bySubmit[record.voter], reward, rewardOffer)...)
	}

	return claims
}

func generateClaimsForVoter(voter *VoterInfo, reward *big.Int, offer FeedReward) []RewardClaim {
	var claims []RewardClaim

	voterFee := voter.delegationFeeBips
	fee := big.NewInt(0).Div(
		bigTmp.Mul(
			reward,
			big.NewInt(int64(voterFee)),
		),
		bigTotalBips,
	)

	if fee.Cmp(bigZero) > 0 {
		claims = append(claims, RewardClaim{
			Beneficiary: common.Address(voter.Identity),
			Amount:      fee,
			Type:        Fee,
		})
	}

	participationReward := big.NewInt(0).Sub(reward, fee)
	if participationReward.Cmp(bigZero) > 0 {
		claims = append(claims, RewardClaim{
			Beneficiary: common.Address(voter.Delegation),
			Amount:      participationReward,
			Type:        WNat,
		})
	}

	return claims
}

type voterRecord struct {
	voter        VoterSubmit
	weight       *big.Int
	isPct, isIqr bool
}

var randomArgs = abi.Arguments{{Type: utils.BytesType}, {Type: utils.Uint256Type}, {Type: utils.AddressType}}

func randomSelect(feedId FeedId, round types.RoundId, voter VoterSubmit) bool {
	pack, err := randomArgs.Pack(feedId, round, voter)
	if err != nil {
		logger.Fatal("error packing arguments, this should never happen: %s", err)
	}
	hash := crypto.Keccak256Hash(pack)
	return hash[len(hash)-1]%2 == 1
}

func isEnoughParticipation(participatingWeight, totalWeight *big.Int, minBips uint16) bool {
	return big.NewInt(0).Mul(
		participatingWeight,
		bigTotalBips,
	).Cmp(
		big.NewInt(0).Mul(
			totalWeight,
			big.NewInt(int64(minBips)),
		),
	) >= 0
}

type RewardShare struct {
	Original *FeedReward
	Amount   *big.Int
}

func calculateRandom(round types.RoundId, offendersByRound map[types.RoundId][]VoterSubmit, eligibleReveals map[VoterSubmit]*Reveal) RandomResult {
	benchingWindowOffenders := map[VoterSubmit]bool{}
	for i := types.RoundId(uint64(round) - params.Coston.Ftso.RandomGenerationBenchingWindow); i < round; i++ {
		for j := range offendersByRound[i] {
			benchingWindowOffenders[offendersByRound[i][j]] = true
		}
	}

	nonBenchedOffenders := 0
	for k := range offendersByRound[round] {
		currentOffender := offendersByRound[round][k]
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
		IsSecure: nonBenchedOffenders == 0 && validCount >= params.Coston.Ftso.NonBenchedRandomVotersMinCount,
	}
	return res
}

func calculateMedians(re RewardEpoch, validReveals map[VoterSubmit][]FeedValue) (map[FeedId]*MedianResult, error) {
	var medianResults map[FeedId]*MedianResult
	for feedIndex, feed := range re.OrderedFeeds {
		var weightedValues []VoterValue

		for voterSubmit, values := range validReveals {
			feedValue := values[feedIndex]
			weight := re.Voters.bySubmit[voterSubmit].cappedWeight
			if feedValue.isEmpty || weight == nil {
				continue
			}
			weightedValues = append(weightedValues, VoterValue{
				value:  feedValue.Value,
				weight: weight,
			})
		}

		median, err := CalculateFeedMedian(weightedValues)
		if err != nil {
			logger.Error("error calculating median for feed %s: %s", feed.String(), err)
			continue
		}

		medianResults[feed.Id] = median

		logger.Info("Feed: %s, Median: %+v", feed.String(), median)
	}

	return medianResults, nil
}

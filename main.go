package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flare-common/database"
	"flare-common/policy"
	"fmt"
	voters "ftsov2-rewarding/lib"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/types"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
	"os"
	"slices"
	"strconv"
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

	os.Setenv("NETWORK", "coston")

	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}

	epoch := types.EpochId(2960)

	allClaims, err := calculateRewardClaims(db, epoch)
	if err != nil {
		logger.Fatal("Error calculating reward claims for epoch %d: %s", epoch, err)
		return
	}

	merged := mergeClaims(allClaims)
	logger.Info("Merged claims: %d, all claims %d", len(merged), len(allClaims))

	printResults(allClaims, strconv.Itoa(int(epoch)))
}

func printResults(records []RewardClaim, suffix string) {
	jsonData, err := json.MarshalIndent(records, "", "    ")
	if err != nil {
		fmt.Println("Error serializing to JSON:", err)
		return
	}

	file, err := os.Create(fmt.Sprintf("results/claims-%s.json", suffix))
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
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

func calculateRewardClaims(db *gorm.DB, epoch types.EpochId) ([]RewardClaim, error) {
	re, err := getRewardEpoch(epoch, db)
	if err != nil {
		return nil, errors.Wrap(err, "err fetching reward epoch")
	}

	windowStart := types.RoundId(uint64(re.StartRound) - params.Net.Ftso.RandomGenerationBenchingWindow)
	windowEnd := re.EndRound.Add(params.Net.Ftso.FutureSecureRandomWindow)

	commitsByRound, err := getCommits(db, windowStart, windowEnd)
	if err != nil {
		return nil, errors.Errorf("error fetching commitsByRound: %s", err)
	}
	revealsByRound, err := getReveals(db, windowStart, windowEnd)
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

			expected := utils.CommitHash(common.Address(voter), uint32(round), reveal.Random, reveal.EncodedValues)

			if expected.Cmp(commit.Hash) != 0 {
				logger.Debug("voter %s reveal hash did not match commit: %s != %s", common.Address(voter), expected.String(), commit.Hash.String())
				offenders = append(offenders, voter)
				continue
			}

			matchingReveals[voter] = reveal
		}

		offendersByRound[round] = offenders
		matchingRevealsByRound[round] = matchingReveals
	}

	results := map[types.RoundId]RoundResult{}

	for round := re.StartRound; round <= re.EndRound; round++ {
		validReveals := matchingRevealsByRound[round]

		logger.Info("Reveals for round %d: %d", round, len(validReveals))

		eligibleReveals := map[VoterSubmit]*Reveal{}
		for voter, reveal := range validReveals {
			if _, ok := re.Voters.bySubmit[voter]; ok {
				eligibleReveals[voter] = reveal
				logger.Info("By voter %s", hex.EncodeToString(voter[:]))

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

		median, err := calculateMedians(round, re, feedValues)
		if err != nil {
			return nil, err
		}

		for _, feed := range re.OrderedFeeds {
			jsonData, _ := json.MarshalIndent(median[feed.Id], "", "    ")

			logger.Info("Feed: %s, median: %s", feed.Id.Hex(), jsonData)
		}

		//logger.Info("Median: %+v", median)

		random := calculateRandom(round, offendersByRound, eligibleReveals)
		logger.Info("Round %d, random result: %d", round, random.Value)

		results[round] = RoundResult{
			Round:  round,
			Median: median,
			Random: random,
		}
	}

	var lastRandom *RandomResult
	var lastRandomRound types.RoundId

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
			lastRandomRound = round
			break
		}
		logger.Info("Extra random: %d %+v", lastRandom)
	}

	totalRounds := int64(re.EndRound - re.StartRound + 1)

	feedSelectionRandoms := make([]*big.Int, totalRounds)
	for i := re.StartRound; i < re.EndRound; i++ {
		logger.Info("Calculating feed selection random for round %d", i)
		feedSelectionRandoms[i-re.StartRound] = utils.FeedSelectionRandom(results[i+1].Random.Value, i+1)
	}
	// Random for last round is the first secure random from next reward epoch,
	// or nil if none found within a certain window.
	if lastRandom != nil {
		lastRound := re.EndRound - re.StartRound
		feedSelectionRandoms[lastRound] = utils.FeedSelectionRandom(lastRandom.Value, lastRandomRound)
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

	perRound, rem := totalReward.DivMod(totalReward, big.NewInt(totalRounds), big.NewInt(0))
	numFeeds := big.NewInt(int64(len(re.OrderedFeeds)))
	// TODO: Can reduce allocations in loop by re-using big.Int vars OR use uint64 if safe
	for round := re.StartRound; round <= re.EndRound; round++ {
		random := feedSelectionRandoms[round-re.StartRound]

		logger.Info("Selected random for round %d: %d", round, random)

		amount := big.NewInt(0).Set(perRound)
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

	epochClaims := make([]RewardClaim, 0)

	_, err = getSigners(db, re)
	if err != nil {
		return nil, errors.Wrap(err, "error calculating signers")
	}

	signers, err := getSigners(db, re)
	if err != nil {
		return nil, errors.Wrap(err, "error calculating signers")
	}

	finalz, err := getFinalz(db, re)
	if err != nil {
		return nil, errors.Wrap(err, "err fetching finalizations")
	}
	logger.Info("Got finalizations: %d", len(finalz))

	// Calculate reward claims
	for round := re.StartRound; round <= re.EndRound; round++ {
		totalRoundReward := roundRewards[round]

		logger.Info("Round: %d, total reward: %s, feed: %s", round, totalRoundReward.Amount.String(), hex.EncodeToString(totalRoundReward.Feed.Id[:]))
		logger.Info("Median: %+v", results[round].Median[totalRoundReward.Feed.Id])

		if totalRoundReward.ShouldBurn {
			epochClaims = append(epochClaims, RewardClaim{
				Beneficiary: BurnAddress,
				Amount:      big.NewInt(0).Set(totalRoundReward.Amount),
				Type:        Direct,
			})
			continue
		}

		signingReward := big.NewInt(0).Div(
			bigTmp.Mul(totalRoundReward.Amount, params.Net.Ftso.SigningBips),
			bigTotalBips,
		)
		finalizationReward := big.NewInt(0).Div(
			bigTmp.Mul(totalRoundReward.Amount, params.Net.Ftso.FinalizationBips),
			bigTotalBips,
		)
		medianReward := big.NewInt(0).Sub(
			totalRoundReward.Amount,
			bigTmp.Add(signingReward, finalizationReward),
		)

		logger.Info("Reward shares for round %d: signing %s, finalization %s, median %s", round, signingReward.String(), finalizationReward.String(), medianReward.String())

		medianClaims := calcMedianRewardClaims(round, re, medianReward, totalRoundReward, results[round].Median[totalRoundReward.Feed.Id])
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

		signingClaims := calcSigningRewardClaims(round, re, signingReward, eligibleVoters, signers[round], finalz[round])

		finalizers, err := selectFinalizers(round, re.Policy, params.Net.Ftso.FinalizationVoterSelectionThresholdWeightBips)
		if err != nil {
			return nil, errors.Wrap(err, "error selecting finalizers")
		}
		finalizationClaims := calcFinalizationRewardClaims(round, finalizationReward, finalz[round], eligibleVoters, finalizers)

		dSigners := getDoubleSigners(signers[round])
		var dSignerInfos []*VoterInfo
		for dSigner := range dSigners {
			dSignerInfos = append(dSignerInfos, re.Voters.bySigning[dSigner])
		}

		doubleSigningPenalties := calcPenalties(totalRoundReward.Amount, params.Net.Ftso.PenaltyFactor, dSignerInfos, re.Voters)

		var offenderInfos []*VoterInfo
		for _, offender := range offendersByRound[round] {
			info := re.Voters.bySubmit[offender]
			if info != nil {
				offenderInfos = append(offenderInfos, re.Voters.bySubmit[offender])
			}
		}
		revealPenalties := calcPenalties(totalRoundReward.Amount, params.Net.Ftso.PenaltyFactor, offenderInfos, re.Voters)

		logger.Info("Round: %d, computed median claims: %d, signing claims: %d, finalz claims: %d", round, len(medianClaims), len(signingClaims), len(finalizationClaims))

		epochClaims = append(epochClaims, signingClaims...)
		epochClaims = append(epochClaims, finalizationClaims...)
		epochClaims = append(epochClaims, doubleSigningPenalties...)
		epochClaims = append(epochClaims, revealPenalties...)

		printResults(medianClaims, fmt.Sprintf("%d-median-claims", round))
		printResults(signingClaims, fmt.Sprintf("%d-signing-claims", round))
		printResults(finalizationClaims, fmt.Sprintf("%d-finalz-claims", round))
		printResults(doubleSigningPenalties, fmt.Sprintf("%d-doublesig-claims", round))
		printResults(revealPenalties, fmt.Sprintf("%d-reveal-claims", round))
	}

	return epochClaims, nil
}

func calcPenalties(
	reward *big.Int,
	penaltyFactor *big.Int,
	offenders []*VoterInfo,
	voters *VoterIndex,
) []RewardClaim {
	var penalties []RewardClaim
	for _, offender := range offenders {
		penalty := bigTmp.Div(
			bigTmp.Mul(offender.CappedWeight, bigTmp.Mul(reward, penaltyFactor)),
			voters.totalCappedWeight,
		)
		penalties = append(penalties, signingClaimsForVoter(offender, penalty)...)
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

func calcFinalizationRewardClaims(
	round types.RoundId,
	reward *big.Int,
	finalizations []*Finalization,
	eligibleVoters []*VoterInfo,
	eligibleFinalizers map[common.Address]bool,
) []RewardClaim {

	// TODO: Pre-compute
	successIndex := slices.IndexFunc(finalizations, func(f *Finalization) bool {
		return f.Info.Reverted == false
	})

	if successIndex < 0 {
		return []RewardClaim{burnClaim(reward)}
	}

	firstSuccessfulFinalization := finalizations[successIndex]
	gracePeriodDeadline := params.Net.Epoch.RevealDeadlineSec(round+1) + params.Net.Ftso.GracePeriodForFinalizationDurationSec

	if firstSuccessfulFinalization.Info.TimestampSec > gracePeriodDeadline {
		// No voter provided finalization in grace period. The first successful finalizer gets the full reward.
		return []RewardClaim{
			{
				Beneficiary: firstSuccessfulFinalization.Info.From,
				Amount:      reward,
				Type:        Direct,
			},
		}
	}

	// TODO: Handle case when finalization is late and sent in the following round

	var graceFinalizations []*Finalization
	for _, finalization := range finalizations {
		if eligibleFinalizers[finalization.Info.From] && finalization.Info.TimestampSec <= gracePeriodDeadline {
			graceFinalizations = append(graceFinalizations, finalization)
		}
	}
	// We have at least one successful finalization in the grace period, but from non-eligible voters -> burn the reward.
	if len(graceFinalizations) == 0 {
		return []RewardClaim{burnClaim(reward)}
	}

	var claims []RewardClaim

	// The reward should be distributed equally among all the eligible finalizers.
	// Note that each finalizer was chosen by probability corresponding to its relative weight.
	// Consequently, the real weight should not be taken into account here.
	undistributedAmount := big.NewInt(0).Set(reward)
	undistributedWeight := big.NewInt(int64(len(eligibleFinalizers)))

	eligibleVoterBySigning := map[VoterSigning]*VoterInfo{}
	for _, voter := range eligibleVoters {
		eligibleVoterBySigning[voter.Signing] = voter
	}
	for _, finalization := range graceFinalizations {
		voter := eligibleVoterBySigning[VoterSigning(finalization.Info.From)]
		if voter == nil {
			continue
		}

		claimAmount := big.NewInt(0).Div(undistributedAmount, undistributedWeight)
		undistributedAmount.Sub(undistributedAmount, claimAmount)
		undistributedWeight.Sub(undistributedWeight, big.NewInt(1))

		claims = append(claims, signingClaimsForVoter(voter, claimAmount)...)
	}

	if undistributedAmount.Cmp(bigZero) != 0 {
		logger.Info("Burning undistributed finalization reward amount: %s", undistributedAmount.String())
		return []RewardClaim{burnClaim(undistributedAmount)}
	}

	return claims
}

func calcSigningRewardClaims(
	round types.RoundId,
	re RewardEpoch,
	reward *big.Int,
	eligibleVoters []*VoterInfo,
	signers map[common.Hash]map[VoterSigning]SigInfo,
	finalizations []*Finalization,
) []RewardClaim {
	doubleSigners := getDoubleSigners(signers)

	revealDeadline := params.Net.Epoch.RevealDeadlineSec(round + 1)
	roundEnd := params.Net.Epoch.VotingRoundEndSec(
		round.Add(1 + params.Net.Ftso.AdditionalRewardFinalizationWindows),
	)

	acceptedSigs := map[common.Hash]map[VoterSigning]SigInfo{}
	for hash, sigs := range signers {
		acceptedSigs[hash] = map[VoterSigning]SigInfo{}
		for signer, sig := range sigs {
			if sig.Timestamp < revealDeadline || sig.Timestamp > roundEnd {
				continue
			}
			acceptedSigs[hash][signer] = sig
		}
	}

	var rewardEligibleSigs []SigInfo

	// TODO: Pre-compute
	successIndex := slices.IndexFunc(finalizations, func(f *Finalization) bool {
		return f.Info.Reverted == false
	})

	if successIndex < 0 {
		signatures := acceptedHashSignatures(re, acceptedSigs)
		if signatures == nil {
			return []RewardClaim{burnClaim(reward)}
		} else {
			for _, s := range signatures {
				if _, ok := doubleSigners[s.Signer]; !ok {
					rewardEligibleSigs = append(rewardEligibleSigs, s)
				}
			}
		}
	} else {
		successfulFinalization := finalizations[successIndex]

		deadline := min(
			successfulFinalization.Info.TimestampSec,
			roundEnd,
		)
		gracePeriod := revealDeadline + params.Net.Ftso.GracePeriodForSignaturesDurationSec

		for _, s := range acceptedSigs[successfulFinalization.merkleRoot.hash] {
			if _, ok := doubleSigners[s.Signer]; ok {
				continue
			}

			if s.Timestamp <= gracePeriod || s.Timestamp <= deadline {
				rewardEligibleSigs = append(rewardEligibleSigs, s)
			}
		}
	}

	// Distribute rewards
	remainingWeight := uint16(0)
	for _, sig := range rewardEligibleSigs {
		remainingWeight += re.Policy.Voters.VoterDataMap[common.Address(sig.Signer)].Weight
	}

	if remainingWeight == 0 {
		return []RewardClaim{burnClaim(reward)}
	}
	remainingAmount := big.NewInt(0).Set(reward)

	var claims []RewardClaim
	// Sort signatures according to voter order in signing policy
	slices.SortFunc(rewardEligibleSigs, func(i, j SigInfo) int {
		indexI := re.Policy.Voters.VoterDataMap[common.Address(i.Signer)].Index
		indexJ := re.Policy.Voters.VoterDataMap[common.Address(j.Signer)].Index
		return indexI - indexJ
	})

	eligibleSigners := map[VoterSigning]*VoterInfo{}
	for _, voter := range eligibleVoters {
		eligibleSigners[voter.Signing] = voter
	}

	for _, sig := range rewardEligibleSigs {
		weight := re.Policy.Voters.VoterDataMap[common.Address(sig.Signer)].Weight
		if weight == 0 {
			continue
		}

		// TODO: clean up big.Int calculations
		claimAmount := big.NewInt(0).Div(
			bigTmp.Mul(remainingAmount, big.NewInt(int64(weight))),
			big.NewInt(int64(remainingWeight)),
		)

		remainingAmount.Sub(remainingAmount, claimAmount)
		remainingWeight -= weight

		if voter, ok := eligibleSigners[sig.Signer]; ok {
			claims = append(claims, signingClaimsForVoter(voter, claimAmount)...)
		} else {
			claims = append(claims, burnClaim(claimAmount))
		}
	}

	return claims
}

func signingClaimsForVoter(voter *VoterInfo, amount *big.Int) []RewardClaim {
	var claims []RewardClaim

	stakedWeight := big.NewInt(0)
	for _, w := range voter.NodeWeights {
		stakedWeight.Add(stakedWeight, w)
	}

	totalWeight := big.NewInt(0).Add(voter.CappedWeight, stakedWeight)
	if totalWeight.Cmp(bigZero) == 0 {
		logger.Fatal("voter totalWeight is zero, this should never happen")
	}

	stakingAmount := big.NewInt(0).Div(
		bigTmp.Mul(amount, stakedWeight),
		totalWeight,
	)
	delegationAmount := big.NewInt(0).Sub(amount, stakingAmount)
	delegationFee := big.NewInt(0).Div(
		bigTmp.Mul(delegationAmount, big.NewInt(int64(voter.DelegationFeeBips))),
		bigTotalBips,
	)
	cappedStakingFeeBips := big.NewInt(min(int64(voter.DelegationFeeBips), params.Net.Ftso.CappedStakingFeeBips))
	stakingFee := big.NewInt(0).Div(
		bigTmp.Mul(stakingAmount, cappedStakingFeeBips),
		bigTotalBips,
	)
	feeBeneficiary := common.Address(voter.Identity)
	delegationBeneficiary := common.Address(voter.Delegation)

	fee := big.NewInt(0).Add(delegationFee, stakingFee)
	if fee.Cmp(bigZero) != 0 {
		claims = append(claims, RewardClaim{
			Beneficiary: feeBeneficiary,
			Amount:      fee,
			Type:        Fee,
		})
	}

	delegationCommunityReward := big.NewInt(0).Sub(delegationAmount, delegationFee)
	claims = append(claims, RewardClaim{
		Beneficiary: delegationBeneficiary,
		Amount:      delegationCommunityReward,
		Type:        WNat,
	})

	remainingStakeWeight := big.NewInt(0).Set(stakedWeight)
	remainingStakeAmount := big.NewInt(0).Sub(stakingAmount, stakingFee)

	for i := range voter.NodeIds {
		nodeId := voter.NodeIds[i]
		nodeWeight := voter.NodeWeights[i]

		nodeAmount := big.NewInt(0)

		if nodeWeight.Cmp(bigZero) > 0 {
			nodeAmount = big.NewInt(0).Div(
				bigTmp.Mul(remainingStakeAmount, nodeWeight),
				remainingStakeWeight,
			)
		}

		remainingStakeAmount.Sub(remainingStakeAmount, nodeAmount)
		remainingStakeWeight.Sub(remainingStakeWeight, nodeWeight)

		claims = append(claims, RewardClaim{
			Beneficiary: nodeId,
			Amount:      nodeAmount,
			Type:        Mirror,
		})
	}

	if remainingStakeAmount.Cmp(bigZero) != 0 {
		logger.Fatal("remainingStakeAmount is not zero, this should never happen")
	}

	return claims
}

func burnClaim(amount *big.Int) RewardClaim {
	return RewardClaim{
		Beneficiary: BurnAddress,
		Amount:      amount,
		Type:        Direct,
	}
}

func acceptedHashSignatures(
	re RewardEpoch,
	signaturesByHash map[common.Hash]map[VoterSigning]SigInfo,
) map[VoterSigning]SigInfo {
	threshold := re.Policy.Voters.TotalWeight * params.Net.Ftso.MinimalRewardedNonConsensusDepositedSignaturesPerHashBips / totalBips

	maxWeight := uint16(0)
	var result map[VoterSigning]SigInfo

	for _, signatures := range signaturesByHash {
		hashWeight := uint16(0)
		for _, info := range signatures {
			signerWeight := re.Policy.Voters.VoterDataMap[common.Address(info.Signer)].Weight
			hashWeight += signerWeight
		}
		if hashWeight > threshold && hashWeight > maxWeight {
			maxWeight = hashWeight
			result = signatures
		}
	}

	return result
}

type SignerMap map[types.RoundId]map[common.Hash]map[VoterSigning]SigInfo

type SigInfo struct {
	Signer    VoterSigning
	Timestamp uint64
}

// getSigners fetches all signatures for all rounds in the reward epoch, and for each round
// computes the list of valid signatures by signed hash.
// For each signer, only the last signature for a specific round and hash is retained.
func getSigners(db *gorm.DB, re RewardEpoch) (SignerMap, error) {
	allSignatures, err := getSignatures(db, re.StartRound, re.EndRound)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching signatures")
	}

	signers := SignerMap{}
	for round, signatures := range allSignatures {
		sigsByHash := map[common.Hash]map[VoterSigning]SigInfo{}
		for _, signatureSubmission := range signatures {
			signature := signatureSubmission.Signature
			signedHash := signature.merkleRoot.EncodedHash()
			signerKey, err := crypto.SigToPub(
				signedHash.Bytes(),
				append(signature.bytes[1:65], signature.bytes[0]-27),
			)
			if err != nil {
				logger.Debug("error recovering signerKey, skipping signature: %s", err)
				continue
			}

			signer := VoterSigning(crypto.PubkeyToAddress(*signerKey))
			if _, ok := re.Voters.bySigning[signer]; ok {
				if _, ok := sigsByHash[signedHash]; !ok {
					sigsByHash[signedHash] = map[VoterSigning]SigInfo{}
				}
				sigsByHash[signedHash][signer] = SigInfo{
					Signer:    signer,
					Timestamp: signatureSubmission.Info.TimestampSec,
				}
			} else {
				logger.Debug("signer %s not registered, skipping signature", signer)
			}
		}

		signers[round] = sigsByHash
	}
	return signers, nil
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

func getFinalz(db *gorm.DB, re RewardEpoch) (map[types.RoundId][]*Finalization, error) {
	allFinalizationsByRound, err := getFinalizations(db, re.StartRound, re.EndRound)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching finalizations")
	}

	logger.Info("Finalizations: %d", len(allFinalizationsByRound))

	finalizationsByRound := make(map[types.RoundId][]*Finalization)

	//var firstSuccessful *Finalization
	for round, finalizations := range allFinalizationsByRound {
		seenSender := map[common.Address]bool{}
		for _, finalization := range finalizations {
			if types.EpochId(finalization.Policy.RewardEpochId) != re.Epoch {
				logger.Info("finalization reward epoch %d does not match expected epoch %d, skipping", finalization.Policy.RewardEpochId, re.Epoch)
				continue
			}

			if !bytes.Equal(finalization.Policy.RawBytes, re.Policy.RawBytes) {
				logger.Info("finalization signing policy does not match expected, skipping")
				continue
			}

			if _, ok := seenSender[finalization.Info.From]; ok {
				logger.Info("finalization from %s already seen, skipping", finalization.Info.From)
				continue
			} else {
				seenSender[finalization.Info.From] = true
			}

			//if firstSuccessful == nil && !finalization.Info.Reverted {
			//	firstSuccessful = finalization
			//}
			// TODO: Store first successful

			finalizationsByRound[round] = append(finalizationsByRound[round], finalization)
		}
	}
	return finalizationsByRound, nil
}

type FeedReward struct {
	Feed       *Feed
	Amount     *big.Int
	ShouldBurn bool
}

type RewardShare struct {
	Original *FeedReward
	Amount   *big.Int
}

func calculateRandom(round types.RoundId, offendersByRound map[types.RoundId][]VoterSubmit, eligibleReveals map[VoterSubmit]*Reveal) RandomResult {
	benchingWindowOffenders := map[VoterSubmit]bool{}
	for i := types.RoundId(uint64(round) - params.Net.Ftso.RandomGenerationBenchingWindow); i < round; i++ {
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
		IsSecure: nonBenchedOffenders == 0 && validCount >= params.Net.Ftso.NonBenchedRandomVotersMinCount,
	}
	return res
}

func calculateMedians(round types.RoundId, re RewardEpoch, validReveals map[VoterSubmit][]FeedValue) (map[FeedId]*MedianResult, error) {
	medianResults := map[FeedId]*MedianResult{}
	for feedIndex, feed := range re.OrderedFeeds {
		var weightedValues []VoterValue

		for voterSubmit, values := range validReveals {
			feedValue := values[feedIndex]
			weight := re.Voters.bySubmit[voterSubmit].CappedWeight
			if feedValue.isEmpty || weight == nil {
				continue
			}
			weightedValues = append(weightedValues, VoterValue{
				voter:  voterSubmit,
				value:  feedValue.Value,
				weight: weight,
			})
		}

		logger.Info("Calculating median for round %d feed %s, valid values: %d", round, feed.Id.Hex(), len(weightedValues))

		median, err := CalculateFeedMedian(weightedValues)
		if err != nil {
			logger.Error("error calculating median for feed %s: %s", feed.String(), err)
			continue
		}

		logger.Info("Calculated median for round %s feed %s, %s: result %+v", round, feed.String(), hex.EncodeToString(feed.Id[:]), median)

		medianResults[feed.Id] = median

		//logger.Info("Feed: %s, Median: %+v", feed.String(), median)
	}

	return medianResults, nil
}

package main

import (
	"bytes"
	"encoding/json"
	"flare-common/database"
	"fmt"
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

	epoch := types.EpochId(2745)

	allClaims, err := calculateRewardClaims(db, epoch)
	if err != nil {
		logger.Fatal("Error calculating reward claims for epoch %d: %s", epoch, err)
		return
	}

	merged := mergeClaims(allClaims)
	logger.Info("Merged claims: %d, all claims %d", len(merged), len(allClaims))

	printResults(merged, epoch)
}

func printResults(records []RewardClaim, id types.EpochId) {
	jsonData, err := json.MarshalIndent(records, "", "    ")
	if err != nil {
		fmt.Println("Error serializing to JSON:", err)
		return
	}

	file, err := os.Create(fmt.Sprintf("claims-%d.json", id))
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

	getFinalz(db, re)

	os.Exit(0)

	windowStart := types.RoundId(uint64(re.StartRound) - params.Net.Ftso.RandomGenerationBenchingWindow)
	windowEnd := re.EndRound.Add(params.Net.Ftso.FutureSecureRandomWindow)

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

			expected := utils.CommitHash(common.Address(voter), uint32(round), reveal.Random, reveal.EncodedValues)

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
		//logger.Info("Median: %+v", median)

		random := calculateRandom(round, offendersByRound, eligibleReveals)
		//logger.Info("Random result: %+v", random)

		results[round] = RoundResult{
			Round:  round,
			Median: median,
			Random: random,
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
		feedSelectionRandoms[i-re.StartRound] = utils.FeedSelectionRandom(results[i].Random.Value, i)
	}
	// Random for last round is the first secure random from next reward epoch,
	// or nil if none found within a certain window.
	if lastRandom != nil {
		lastRound := re.EndRound - re.StartRound
		feedSelectionRandoms[lastRound] = utils.FeedSelectionRandom(lastRandom.Value, lastRound)
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
			bigTmp.Mul(totalReward.Amount, params.Net.Ftso.SigningBips),
			bigTotalBips,
		)
		finalizationReward := big.NewInt(0).Div(
			bigTmp.Mul(totalReward.Amount, params.Net.Ftso.FinalizationBips),
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

		logger.Info("Round: %d, computed median claims: %d", round, len(medianClaims))
	}

	return epochClaims, nil
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
			signedHash := signature.merkleRoot.hash
			signerKey, err := crypto.SigToPub(
				signedHash.Bytes(),
				append(signature.bytes[:64], signature.bytes[64]),
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

func getDoubleSigners(roundSigners map[common.Hash]map[VoterSigning]SigInfo) []VoterSigning {
	signed := map[VoterSigning]bool{}
	var doubleSigners []VoterSigning

	for _, signers := range roundSigners {
		for signer := range signers {
			if _, ok := signed[signer]; ok {
				doubleSigners = append(doubleSigners, signer)
			}
			signed[signer] = true
		}
	}

	return doubleSigners
}

func getFinalz(db *gorm.DB, re RewardEpoch) (*Finalization, error) {
	allFinalizations, err := getFinalizations(db, re.StartRound, re.EndRound)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching finalizations")
	}

	logger.Info("Finalizations: %d", len(allFinalizations))

	for _, finalizations := range allFinalizations {
		for _, finalization := range finalizations {

			if types.EpochId(finalization.Policy.RewardEpochId) != re.Epoch {
				logger.Debug("finalization reward epoch %d does not match expected epoch %d, skipping", finalization.Policy.RewardEpochId, re.Epoch)
				continue
			}

			if !bytes.Equal(finalization.Policy.RawBytes, re.Policy.SigningPolicyBytes) {
				logger.Debug("finalization signing policy does not match expected, skipping")
				continue
			}

		}

	}
	return nil, nil
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

func calculateMedians(re RewardEpoch, validReveals map[VoterSubmit][]FeedValue) (map[FeedId]*MedianResult, error) {
	medianResults := map[FeedId]*MedianResult{}
	for feedIndex, feed := range re.OrderedFeeds {
		var weightedValues []VoterValue

		for voterSubmit, values := range validReveals {
			feedValue := values[feedIndex]
			weight := re.Voters.bySubmit[voterSubmit].cappedWeight
			if feedValue.isEmpty || weight == nil {
				continue
			}
			weightedValues = append(weightedValues, VoterValue{
				voter:  voterSubmit,
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

		//logger.Info("Feed: %s, Median: %+v", feed.String(), median)
	}

	return medianResults, nil
}

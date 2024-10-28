package rewards

import (
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fumanager"
	"math/big"
)

type FeedReward struct {
	Feed       *data.Feed
	Amount     *big.Int
	ShouldBurn bool
}

type FUFeedReward struct {
	FeedIndex  uint64
	FeedConfig *fumanager.IFastUpdatesConfigurationFeedConfiguration
	Amount     *big.Int
	ShouldBurn bool
}

// calculateFURoundRewards total FastUpdates reward offer share per round
func calculateFURoundRewards(re data.RewardEpoch, feedSelectionRandoms []*big.Int) map[ty.RoundId]FUFeedReward {
	totalReward := big.NewInt(0)
	for i := range re.Offers.FastUpdates {
		totalReward.Add(totalReward, re.Offers.FastUpdates[i].Amount)
	}
	for i := range re.Offers.FastUpdatesI {
		totalReward.Add(totalReward, re.Offers.FastUpdatesI[i].OfferAmount)
	}

	roundRewards := make(map[ty.RoundId]FUFeedReward)

	perRound, rem := totalReward.DivMod(totalReward, big.NewInt(int64(re.EndRound-re.StartRound+1)), big.NewInt(0))

	feedConfigs := re.Offers.FastUpdates[0].FeedConfigurations
	numFeeds := big.NewInt(int64(len(feedConfigs)))

	for round := re.StartRound; round <= re.EndRound; round++ {
		random := feedSelectionRandoms[round-re.StartRound]

		logger.Debug("[FU] Selected random for round %d: %d", round, random)

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

// calculateRoundRewards total reward offer share per round
func calculateRoundRewards(re data.RewardEpoch, feedSelectionRandoms []*big.Int) map[ty.RoundId]FeedReward {
	totalReward := big.NewInt(0)
	for i := range re.Offers.Inflation {
		offer := re.Offers.Inflation[i]
		totalReward.Add(totalReward, offer.Amount)
	}
	for i := range re.Offers.Community {
		totalReward.Add(totalReward, re.Offers.Community[i].Amount)
	}

	roundRewards := make(map[ty.RoundId]FeedReward)

	perRound, rem := totalReward.DivMod(totalReward, big.NewInt(int64(re.EndRound-re.StartRound+1)), big.NewInt(0))
	numFeeds := big.NewInt(int64(len(re.OrderedFeeds)))
	for round := re.StartRound; round <= re.EndRound; round++ {
		random := feedSelectionRandoms[round-re.StartRound]

		logger.Debug("Selected random for round %d: %d", round, random)

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

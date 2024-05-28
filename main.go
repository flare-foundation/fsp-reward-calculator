package main

import (
	"flare-common/contracts/offers"
	"flare-common/contracts/relay"
	"flare-common/database"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/parameters"
	"ftsov2-rewarding/utils"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
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

	calculateRewards(db, 2690)
}

type RewardClaim struct {
}

type RewardEpoch struct {
	Epoch      int
	StartRound int
	EndRound   int
	Policy     *relay.RelaySigningPolicyInitialized
	Offers     RewardOffers
}

type RewardOffers struct {
	community []*offers.OffersRewardsOffered
	inflation []*offers.OffersInflationRewardsOffered
}

func calculateRewards(db *gorm.DB, epoch int) (RewardClaim, error) {
	_, err := getRewardEpoch(epoch, db)
	if err != nil {
		return RewardClaim{}, errors.Errorf("err fetching reward epoch: %s", err)
	}

	return RewardClaim{}, nil
}

func getRewardEpoch(epoch int, db *gorm.DB) (RewardEpoch, error) {
	currentTimestamp := time.Now().Unix()

	// TODO: Use lowest index in indexer db as start
	expectedStartSec := parameters.Coston.Epoch.ExpectedRewardEpochStartTimeSec(int64(epoch))
	epochDuration := parameters.Coston.Epoch.RewardEpochDurationInVotingEpochs * parameters.Coston.Epoch.VotingRoundDurationSeconds

	searchIntervalStartSec := expectedStartSec - (epochDuration * 2)
	searchIntervalEndSec := min(expectedStartSec+(epochDuration*2), currentTimestamp)

	relayInst, _ := relay.NewRelay(common.Address{}, nil)
	parsePolicyInitialized := func(log types.Log) (*relay.RelaySigningPolicyInitialized, error) {
		return relayInst.RelayFilterer.ParseSigningPolicyInitialized(log)
	}
	policies, err := QueryEvents(db, searchIntervalStartSec, searchIntervalEndSec, parameters.Coston.Contracts.Relay, utils.EventTopic0.SigningPolicyInitialized, parsePolicyInitialized)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("Error fetching signing policy events: %s", err)
	}

	var rewardEpoch = RewardEpoch{}
	for _, event := range policies {
		if event.RewardEpochId.Int64() == int64(epoch) {
			rewardEpoch.Policy = event
			rewardEpoch.StartRound = int(event.StartVotingRoundId)
		}
		if event.RewardEpochId.Int64() == int64(epoch+1) {
			rewardEpoch.EndRound = int(event.StartVotingRoundId) - 1
		}
	}

	if rewardEpoch.Policy == nil {
		return RewardEpoch{}, errors.Errorf("No signing policy found for epoch %d", epoch)
	}
	if rewardEpoch.EndRound == 0 {
		return RewardEpoch{}, errors.Errorf("Unable to determine last voting round for epoch %d: no signing policy found for next epoch %d. It may not have been indexed yet or the current epoch is not finished", epoch, epoch+1)
	}

	actualStartSec := parameters.Coston.Epoch.VotingRoundStartSec(int64(rewardEpoch.StartRound))
	actualEndSec := parameters.Coston.Epoch.VotingRoundEndSec(int64(rewardEpoch.EndRound))

	rewardEpoch.Offers, err = getRewardOffers(db, actualStartSec, actualEndSec)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("Error fetching reward offers: %s", err)
	}

	feeds := GetOrderedFeeds(rewardEpoch.Offers)
	logger.Info("Feeds: %v", len(feeds))
	for _, f := range feeds {
		logger.Info("Feed: %s, Decimals: %d", f.String(), f.Decimals)
	}

	err = getReveals(db, feeds, actualStartSec, actualEndSec)
	if err != nil {
		return RewardEpoch{}, err
	}

	return rewardEpoch, nil
}

func getRewardOffers(db *gorm.DB, epochStartSec int64, epochEndSec int64) (RewardOffers, error) {
	community, err := GetRewardOfferEvents(db, epochStartSec, epochEndSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching reward offer events: %s", err)
	}
	inflation, err := GetInflationRewardOfferEvents(db, epochStartSec, epochEndSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching inflation reward offer events: %s", err)
	}

	return RewardOffers{
		community,
		inflation,
	}, nil
}

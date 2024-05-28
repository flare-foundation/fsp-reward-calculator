package main

import (
	"flare-common/contracts/offers"
	"flare-common/contracts/relay"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/parameters"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type RewardEpoch struct {
	Epoch        int
	StartRound   int
	EndRound     int
	Policy       *relay.RelaySigningPolicyInitialized
	Offers       RewardOffers
	OrderedFeeds []Feed
}

type RewardOffers struct {
	community []*offers.OffersRewardsOffered
	inflation []*offers.OffersInflationRewardsOffered
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
		return RewardEpoch{}, errors.Errorf("error fetching signing policy events: %s", err)
	}

	var policy *relay.RelaySigningPolicyInitialized
	var startRound int
	var endRound int

	for _, event := range policies {
		if event.RewardEpochId.Int64() == int64(epoch) {
			policy = event
			startRound = int(event.StartVotingRoundId)
		}
		if event.RewardEpochId.Int64() == int64(epoch+1) {
			endRound = int(event.StartVotingRoundId) - 1
		}
	}

	if policy == nil {
		return RewardEpoch{}, errors.Errorf("no signing policy found for epoch %d", epoch)
	}
	if endRound == 0 {
		return RewardEpoch{}, errors.Errorf("unable to determine last voting round for epoch %d: no signing policy found for next epoch %d. It may not have been indexed yet or the current epoch is not finished", epoch, epoch+1)
	}

	actualStartSec := parameters.Coston.Epoch.VotingRoundStartSec(int64(startRound))
	actualEndSec := parameters.Coston.Epoch.VotingRoundEndSec(int64(endRound))

	rewardOffers, err := getRewardOffers(db, actualStartSec, actualEndSec)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching reward rewardOffers: %s", err)
	}

	feeds := GetOrderedFeeds(rewardOffers)
	logger.Info("Feeds: %v", len(feeds))
	for _, f := range feeds {
		logger.Info("Feed: %s, Decimals: %d", f.String(), f.Decimals)
	}

	err = getReveals(db, feeds, actualStartSec, actualEndSec)
	if err != nil {
		return RewardEpoch{}, err
	}

	return RewardEpoch{
		Epoch:        epoch,
		StartRound:   startRound,
		EndRound:     endRound,
		Policy:       policy,
		Offers:       rewardOffers,
		OrderedFeeds: feeds,
	}, nil
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

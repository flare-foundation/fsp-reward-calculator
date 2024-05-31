package main

import (
	"flare-common/contracts/offers"
	"flare-common/contracts/relay"
	"fmt"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"slices"
	"sort"
	"time"
)

type RewardEpoch struct {
	Epoch        uint64
	StartRound   uint64
	EndRound     uint64
	Policy       *relay.RelaySigningPolicyInitialized
	Offers       RewardOffers
	OrderedFeeds []Feed
	Voters       VoterIndex
}

type RewardOffers struct {
	community []*offers.OffersRewardsOffered
	inflation []*offers.OffersInflationRewardsOffered
}

type VoterAddresses struct {
	Identity         common.Address
	Submit           common.Address
	SubmitSignatures common.Address
	SigningPolicy    common.Address
}

func getRewardEpoch(epoch uint64, db *gorm.DB) (RewardEpoch, error) {
	currentTimestamp := time.Now().Unix()

	// TODO: Use lowest index in indexer db as start
	expectedStartSec := params.Coston.Epoch.ExpectedRewardEpochStartTimeSec(epoch)
	epochDuration := params.Coston.Epoch.RewardEpochDurationInVotingEpochs * params.Coston.Epoch.VotingRoundDurationSeconds

	searchIntervalStartSec := expectedStartSec - (epochDuration * 2)
	searchIntervalEndSec := min(expectedStartSec+(epochDuration*2), uint64(currentTimestamp))

	relayInst, _ := relay.NewRelay(common.Address{}, nil)
	parsePolicyInitialized := func(log types.Log) (*relay.RelaySigningPolicyInitialized, error) {
		return relayInst.RelayFilterer.ParseSigningPolicyInitialized(log)
	}
	policies, err := QueryEvents(db, searchIntervalStartSec, searchIntervalEndSec, params.Coston.Contracts.Relay, utils.EventTopic0.SigningPolicyInitialized, parsePolicyInitialized)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching signing policy events: %s", err)
	}

	var policy *relay.RelaySigningPolicyInitialized
	var startRound uint64
	var endRound uint64

	for _, event := range policies {
		if event.RewardEpochId.Uint64() == epoch {
			policy = event
			startRound = uint64(event.StartVotingRoundId)
		}
		if event.RewardEpochId.Uint64() == epoch+1 {
			endRound = uint64(event.StartVotingRoundId) - 1
		}
	}

	if policy == nil {
		return RewardEpoch{}, errors.Errorf("no signing policy found for epoch %d", epoch)
	}
	if endRound == 0 {
		return RewardEpoch{}, errors.Errorf("unable to determine last voting round for epoch %d: no signing policy found for next epoch %d. It may not have been indexed yet or the current epoch is not finished", epoch, epoch+1)
	}

	actualStartSec := params.Coston.Epoch.VotingRoundStartSec(startRound)
	actualEndSec := params.Coston.Epoch.VotingRoundEndSec(endRound)

	// TODO: voters and offers should be queried from prev epoch

	rewardOffers, err := getRewardOffers(db, epoch, actualStartSec-epochDuration-10000, actualEndSec-epochDuration)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching reward rewardOffers: %s", err)
	}

	feeds := GetOrderedFeeds(rewardOffers)
	logger.Info("Feeds: %v", len(feeds))
	for _, f := range feeds {
		logger.Info("Feed: %s, Decimals: %d", f.String(), f.Decimals)
	}

	voters, err := getVoterAddresses(db, epoch, actualStartSec-epochDuration, actualEndSec-epochDuration)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching voter addresses: %s", err)
	}

	return RewardEpoch{
		Epoch:        epoch,
		StartRound:   startRound,
		EndRound:     endRound,
		Policy:       policy,
		Offers:       rewardOffers,
		OrderedFeeds: feeds,
		Voters:       NewVoterIndex(voters),
	}, nil
}

func analyseReveals(revealMap map[uint64][]Reveal, feeds []Feed) {
	for round, reveal := range revealMap {
		feedValues := make(map[int][]int32)
		invalidCount := make([]int, len(feeds))
		validCount := make([]int, len(feeds))
		for _, r := range reveal {
			for feedIndex := range feeds {
				if !r.Values[feedIndex].isEmpty {
					if isPowerOfTen(int(r.Values[feedIndex].Value)) {
						invalidCount[feedIndex]++
					} else {
						validCount[feedIndex]++
					}
				}
				feedValues[feedIndex] = append(feedValues[feedIndex], r.Values[feedIndex].Value)
			}
		}

		totalInvalid := 0

		invalidFeeds := make([]string, 0)
		for i, v := range feedValues {
			invalidp := float64(invalidCount[i]) / float64(invalidCount[i]+validCount[i]) * 100
			feedS := feeds[i].String()
			feeds2 := feedS
			if invalidp >= 50 {
				totalInvalid++
				invalidFeeds = append(invalidFeeds, feedS)
			}
			fmt.Printf("Round %d, feed %10s, total %2d, valid %2d, invalid %2d, invalid%% %.2f: %v\n", round, feeds2, invalidCount[i]+validCount[i], validCount[i], invalidCount[i], invalidp, v)
		}

		sort.Slice(invalidFeeds, func(i, j int) bool {
			return invalidFeeds[i] < invalidFeeds[j]
		})

		fmt.Printf("Round %d, total invalid > 50%%: %d\n, feeds: %v", round, totalInvalid, invalidFeeds)
		break
	}
}

func isPowerOfTen(n int) bool {
	if n < 1 {
		return false
	}
	for n > 1 {
		if n%10 != 0 {
			return false
		}
		n /= 10
	}
	return true
}

func getRewardOffers(db *gorm.DB, epoch, epochStartSec, epochEndSec uint64) (RewardOffers, error) {
	community, err := GetRewardOfferEvents(db, epochStartSec, epochEndSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching reward offer events: %s", err)
	}
	inflation, err := GetInflationRewardOfferEvents(db, epochStartSec, epochEndSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching inflation reward offer events: %s", err)
	}

	community = slices.DeleteFunc(community, func(offer *offers.OffersRewardsOffered) bool {
		return offer.RewardEpochId.Uint64() != epoch
	})
	inflation = slices.DeleteFunc(inflation, func(offer *offers.OffersInflationRewardsOffered) bool {
		return offer.RewardEpochId.Uint64() != epoch
	})

	return RewardOffers{
		community,
		inflation,
	}, nil
}

func getVoterAddresses(db *gorm.DB, epoch, epochStartSec, epochEndSec uint64) ([]VoterAddresses, error) {
	events, err := GetVoterRegisteredEvents(db, epochStartSec, epochEndSec)
	if err != nil {
		return nil, errors.Errorf("error fetching voter registered events: %s", err)
	}

	var addresses []VoterAddresses
	for i := range events {
		if events[i].RewardEpochId.Uint64() != epoch {
			continue
		}

		addresses = append(addresses, VoterAddresses{
			Identity:         events[i].Voter,
			Submit:           events[i].SubmitAddress,
			SubmitSignatures: events[i].SubmitSignaturesAddress,
			SigningPolicy:    events[i].SigningPolicyAddress,
		})
	}

	return addresses, nil
}

type VoterIndex struct {
	voters   []VoterAddresses
	bySubmit map[common.Address]common.Address
}

func NewVoterIndex(voters []VoterAddresses) VoterIndex {
	bySubmit := make(map[common.Address]common.Address)
	for _, v := range voters {
		bySubmit[v.Submit] = v.Identity
	}
	return VoterIndex{
		voters:   voters,
		bySubmit: bySubmit,
	}
}

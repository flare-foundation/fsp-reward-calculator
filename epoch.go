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
	"math/big"
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
	Identity         VoterId
	Submit           VoterSubmit
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
	//actualEndSec := params.Coston.Epoch.VotingRoundEndSec(endRound)

	rewardOffers, err := getRewardOffers(db, epoch, actualStartSec-(epochDuration+epochDuration/2), actualStartSec)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching reward rewardOffers: %s", err)
	}

	feeds := GetOrderedFeeds(rewardOffers)
	logger.Info("Feeds: %v", len(feeds))
	for _, f := range feeds {
		logger.Info("Feed: %s, Decimals: %d", f.String(), f.Decimals)
	}

	voters, err := getVoters(db, epoch, actualStartSec-(epochDuration+epochDuration/2), actualStartSec)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching voter info: %s", err)
	}

	return RewardEpoch{
		Epoch:        epoch,
		StartRound:   startRound,
		EndRound:     endRound,
		Policy:       policy,
		Offers:       rewardOffers,
		OrderedFeeds: feeds,
		Voters:       voters,
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

func getVoters(db *gorm.DB, epoch, fromSec, toSec uint64) (VoterIndex, error) {
	regs, err := GetVoterRegisteredEvents(db, fromSec, toSec)
	if err != nil {
		return VoterIndex{}, errors.Errorf("error fetching voter registered regs: %s", err)
	}

	var addresses []VoterAddresses
	for i := range regs {
		if regs[i].RewardEpochId.Uint64() != epoch {
			continue
		}

		addresses = append(addresses, VoterAddresses{
			Identity:         VoterId(regs[i].Voter),
			Submit:           VoterSubmit(regs[i].SubmitAddress),
			SubmitSignatures: regs[i].SubmitSignaturesAddress,
			SigningPolicy:    regs[i].SigningPolicyAddress,
		})

		logger.Info("Voter %s, submit %s, submit signatures %s, signing policy %s", regs[i].Voter.String(), regs[i].SubmitAddress.String(), regs[i].SubmitSignaturesAddress.String(), regs[i].SigningPolicyAddress.String())
	}

	infos, err := GetVoterInfoEvents(db, fromSec, toSec)
	if err != nil {
		return VoterIndex{}, errors.Errorf("error fetching voter info events: %s", err)
	}

	if len(regs) != len(infos) {
		return VoterIndex{}, errors.Errorf("mismatched voter registered and voter info events: %d != %d", len(regs), len(infos))

	}

	cappedWeight := map[VoterId]*big.Int{}
	for i := range infos {
		if infos[i].RewardEpochId.Uint64() != epoch {
			continue
		}
		cappedWeight[VoterId(infos[i].Voter)] = infos[i].WNatCappedWeight
		logger.Info("Voter %s, capped weight %s", infos[i].Voter.String(), infos[i].WNatCappedWeight.String())
	}

	return NewVoterIndex(addresses, cappedWeight), nil
}

type VoterId common.Address
type VoterSubmit common.Address

type VoterIndex struct {
	identityToAddrs  map[VoterId]VoterAddresses
	submitToIdentity map[VoterSubmit]VoterId
	cappedWeight     map[VoterId]*big.Int
}

func NewVoterIndex(voters []VoterAddresses, cappedWeight map[VoterId]*big.Int) VoterIndex {
	addrMap := make(map[VoterId]VoterAddresses)
	submitToIdentity := make(map[VoterSubmit]VoterId)
	for _, v := range voters {
		submitToIdentity[v.Submit] = v.Identity
		addrMap[v.Identity] = v
	}
	return VoterIndex{
		identityToAddrs:  addrMap,
		submitToIdentity: submitToIdentity,
		cappedWeight:     cappedWeight,
	}
}

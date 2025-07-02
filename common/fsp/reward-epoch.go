package fsp

import (
	common2 "fsp-rewards-calculator/common"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fdchub"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fumanager"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/offers"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/relay"
	"github.com/flare-foundation/go-flare-common/pkg/policy"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"slices"
	"time"
)

type RewardEpoch struct {
	Epoch         ty.RewardEpochId
	StartRound    ty.RoundId
	EndRound      ty.RoundId
	Policy        *policy.SigningPolicy
	Offers        RewardOffers
	OrderedFeeds  []Feed
	OrderedVoters []ty.VoterSigning
	VoterIndex    *VoterIndex
}

type RewardOffers struct {
	Community    []*offers.OffersRewardsOffered
	Inflation    []*offers.OffersInflationRewardsOffered
	FastUpdates  []*fumanager.FUManagerInflationRewardsOffered
	FastUpdatesI []*fumanager.FUManagerIncentiveOffered
	FdcInflation []*fdchub.FdcHubInflationRewardsOffered
}

func GetRewardEpoch(epoch ty.RewardEpochId, db *gorm.DB) (RewardEpoch, error) {
	currentTimestamp := time.Now().Unix()

	expectedStartSec := params.Net.Epoch.ExpectedRewardEpochStartTimeSec(epoch)
	epochDuration := params.Net.Epoch.RewardEpochDurationInVotingEpochs * params.Net.Epoch.VotingEpochDurationSeconds

	searchIntervalStartSec := expectedStartSec - (epochDuration * 2)
	searchIntervalEndSec := min(expectedStartSec+(epochDuration*2), uint64(currentTimestamp))

	relayInst, _ := relay.NewRelay(common.Address{}, nil)
	parsePolicyInitialized := func(log types.Log, _ uint64) (*relay.RelaySigningPolicyInitialized, error) {
		return relayInst.ParseSigningPolicyInitialized(log)
	}

	policies, err := QueryEvents(db, searchIntervalStartSec, searchIntervalEndSec, params.Net.Contracts.Relay, common2.EventTopic0.SigningPolicyInitialized, parsePolicyInitialized)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching signing policy events: %s", err)
	}

	var policyEvent *relay.RelaySigningPolicyInitialized
	var startRound ty.RoundId
	var endRound ty.RoundId

	for _, event := range policies {
		if event.RewardEpochId.Uint64() == uint64(epoch) {
			policyEvent = event
			startRound = ty.RoundId(event.StartVotingRoundId)
		}
		if event.RewardEpochId.Uint64() == uint64(epoch)+1 {
			endRound = ty.RoundId(event.StartVotingRoundId - 1)
		}
	}

	if policyEvent == nil {
		return RewardEpoch{}, errors.Errorf("no signing policy found for epoch %d", epoch)
	}

	epochStartSec := params.Net.Epoch.VotingRoundStartSec(startRound)
	epochEndSec := params.Net.Epoch.VotingRoundEndSec(endRound)

	if endRound == 0 {
		epochEndSec = params.Net.Epoch.ExpectedRewardEpochStartTimeSec(epoch + 1)
	}

	rewardOffers, err := getRewardOffers(db, epoch, epochStartSec, epochEndSec)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching reward rewardOffers: %s", err)
	}

	feeds := getOrderedFeeds(rewardOffers)
	orderedVoters := getOrderedVoters(policyEvent)

	policy := policy.NewSigningPolicy(policyEvent, nil)

	signingPolicyWindow := params.Net.Epoch.NewSigningPolicyInitializationStartSeconds

	voters, err := GetVoterIndex(db, epoch, epochStartSec-signingPolicyWindow, epochStartSec, policy.Voters.VoterDataMap)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching voter info: %s", err)
	}

	return RewardEpoch{
		Epoch:         epoch,
		StartRound:    startRound,
		EndRound:      endRound,
		Policy:        policy,
		Offers:        rewardOffers,
		OrderedFeeds:  feeds,
		OrderedVoters: orderedVoters,
		VoterIndex:    voters,
	}, nil
}

func getOrderedVoters(event *relay.RelaySigningPolicyInitialized) []ty.VoterSigning {
	voters := make([]ty.VoterSigning, len(event.Voters))
	for i, addr := range event.Voters {
		voters[i] = ty.VoterSigning(addr)
	}
	return voters
}

func getRewardOffers(db *gorm.DB, epoch ty.RewardEpochId, startSec, endSec uint64) (RewardOffers, error) {
	extraWindow := uint64(6 * 60 * 60)
	previousStartSec := params.Net.Epoch.ExpectedRewardEpochStartTimeSec(epoch-1) - extraWindow

	community, err := getRewardOfferEvents(db, previousStartSec, startSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching reward offer events: %s", err)
	}
	inflation, err := getInflationRewardOfferEvents(db, previousStartSec, startSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching inflation reward offer events: %s", err)
	}
	fastUpdates, err := getFURewardOfferEvents(db, previousStartSec, startSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching fast updates reward offer events: %s", err)
	}
	fastUpdatesI, err := getFUIncentiveOfferEvents(db, startSec, endSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching fast updates reward offer events: %s", err)
	}

	fdcInflation, err := getFdcInflationRewardOfferEvents(db, previousStartSec, startSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching fdc inflation reward offer events: %s", err)
	}

	community = slices.DeleteFunc(community, func(offer *offers.OffersRewardsOffered) bool {
		return offer.RewardEpochId.Uint64() != uint64(epoch)
	})
	inflation = slices.DeleteFunc(inflation, func(offer *offers.OffersInflationRewardsOffered) bool {
		return offer.RewardEpochId.Uint64() != uint64(epoch)
	})
	fastUpdates = slices.DeleteFunc(fastUpdates, func(offer *fumanager.FUManagerInflationRewardsOffered) bool {
		return offer.RewardEpochId.Uint64() != uint64(epoch)
	})
	fastUpdatesI = slices.DeleteFunc(fastUpdatesI, func(offer *fumanager.FUManagerIncentiveOffered) bool {
		return offer.RewardEpochId.Uint64() != uint64(epoch)
	})
	fdcInflation = slices.DeleteFunc(fdcInflation, func(offer *fdchub.FdcHubInflationRewardsOffered) bool {
		return offer.RewardEpochId.Uint64() != uint64(epoch)
	})

	return RewardOffers{
		community,
		inflation,
		fastUpdates,
		fastUpdatesI,
		fdcInflation,
	}, nil
}

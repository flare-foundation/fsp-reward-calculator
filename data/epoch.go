package data

import (
	"flare-common/contracts/fumanager"
	"flare-common/contracts/offers"
	"flare-common/contracts/relay"
	"flare-common/policy"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/ty"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
	"slices"
	"time"
)

type RewardEpoch struct {
	Epoch         ty.EpochId
	StartRound    ty.RoundId
	EndRound      ty.RoundId
	Policy        *policy.SigningPolicy
	Offers        RewardOffers
	OrderedFeeds  []Feed
	OrderedVoters []ty.VoterSigning
	VoterIndex    *VoterIndex
	// TODO: Move next voter calculation elsewhere
	NextVoters *VoterIndex // Voters for the following reward epoch
	PrevVoters *VoterIndex // Voters for the previous reward epoch
}

type RewardOffers struct {
	Community    []*offers.OffersRewardsOffered
	Inflation    []*offers.OffersInflationRewardsOffered
	FastUpdates  []*fumanager.FUManagerInflationRewardsOffered
	FastUpdatesI []*fumanager.FUManagerIncentiveOffered
}

type VoterInfo struct {
	Identity          ty.VoterId
	Submit            ty.VoterSubmit
	SubmitSignatures  ty.VoterSubmitSignatures
	Signing           ty.VoterSigning
	Delegation        ty.VoterDelegation
	CappedWeight      *big.Int
	DelegationFeeBips uint16
	NodeIds           [][20]byte
	NodeWeights       []*big.Int
}

// TODO: Use proper timings for event search instead of approximate
func GetRewardEpoch(epoch ty.EpochId, db *gorm.DB) (RewardEpoch, error) {
	currentTimestamp := time.Now().Unix()

	// TODO: Use lowest index in indexer db as start
	expectedStartSec := params.Net.Epoch.ExpectedRewardEpochStartTimeSec(epoch)
	epochDuration := params.Net.Epoch.RewardEpochDurationInVotingEpochs * params.Net.Epoch.VotingRoundDurationSeconds

	searchIntervalStartSec := expectedStartSec - (epochDuration * 2)
	searchIntervalEndSec := min(expectedStartSec+(epochDuration*2), uint64(currentTimestamp))

	relayInst, _ := relay.NewRelay(common.Address{}, nil)
	parsePolicyInitialized := func(log types.Log) (*relay.RelaySigningPolicyInitialized, error) {
		return relayInst.RelayFilterer.ParseSigningPolicyInitialized(log)
	}
	policies, err := queryEvents(db, searchIntervalStartSec, searchIntervalEndSec, params.Net.Contracts.Relay, utils.EventTopic0.SigningPolicyInitialized, parsePolicyInitialized)
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
	if endRound == 0 {
		return RewardEpoch{}, errors.Errorf("unable to determine last voting round for epoch %d: no signing policy found for next epoch %d. It may not have been indexed yet or the current epoch is not finished", epoch, epoch+1)
	}

	epochStartSec := params.Net.Epoch.VotingRoundStartSec(startRound)
	epochEndSec := params.Net.Epoch.VotingRoundEndSec(endRound)

	rewardOffers, err := getRewardOffers(db, epoch, epochStartSec, epochEndSec)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching reward rewardOffers: %s", err)
	}

	feeds := getOrderedFeeds(rewardOffers)
	logger.Info("Feeds: %v", len(feeds))
	for _, f := range feeds {
		logger.Info("Feed: %s, Decimals: %d", f.String(), f.Decimals)
	}

	signingPolicyWindow := params.Net.Epoch.NewSigningPolicyInitializationStartSeconds

	voters, err := getVoters(db, epoch, epochStartSec-signingPolicyWindow, epochStartSec)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching voter info: %s", err)
	}

	nextVoters, err := getVoters(db, epoch+1, epochStartSec+epochDuration-signingPolicyWindow, epochStartSec+epochDuration)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching voter info: %s", err)
	}

	prevVoters, err := getVoters(db, epoch-1, epochStartSec-(epochDuration+signingPolicyWindow), epochStartSec-(epochDuration))

	return RewardEpoch{
		Epoch:         epoch,
		StartRound:    startRound,
		EndRound:      endRound,
		Policy:        policy.NewSigningPolicy(policyEvent),
		Offers:        rewardOffers,
		OrderedFeeds:  feeds,
		OrderedVoters: getOrderedVoters(policyEvent),
		VoterIndex:    voters,
		NextVoters:    nextVoters,
		PrevVoters:    prevVoters,
	}, nil
}

func getOrderedVoters(event *relay.RelaySigningPolicyInitialized) []ty.VoterSigning {
	voters := make([]ty.VoterSigning, len(event.Voters))
	for i, addr := range event.Voters {
		voters[i] = ty.VoterSigning(addr)
	}
	return voters
}

func getRewardOffers(db *gorm.DB, epoch ty.EpochId, startSec, endSec uint64) (RewardOffers, error) {
	extraWindow := uint64(6 * 60 * 60)
	previousStartSec := params.Net.Epoch.ExpectedRewardEpochStartTimeSec(epoch-1) - extraWindow

	community, err := GetRewardOfferEvents(db, previousStartSec, startSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching reward offer events: %s", err)
	}
	inflation, err := GetInflationRewardOfferEvents(db, previousStartSec, startSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching inflation reward offer events: %s", err)
	}
	fastUpdates, err := GetFURewardOfferEvents(db, previousStartSec, startSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching fast updates reward offer events: %s", err)
	}
	fastUpdatesI, err := GetFUIncentiveOfferEvents(db, startSec, endSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching fast updates reward offer events: %s", err)
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

	return RewardOffers{
		community,
		inflation,
		fastUpdates,
		fastUpdatesI,
	}, nil
}

func getVoters(db *gorm.DB, epoch ty.EpochId, fromSec, toSec uint64) (*VoterIndex, error) {
	regs, err := GetVoterRegisteredEvents(db, fromSec, toSec)
	if err != nil {
		return nil, errors.Errorf("error fetching voter registered regs: %s", err)
	}

	infos, err := GetVoterInfoEvents(db, fromSec, toSec)
	if err != nil {
		return nil, errors.Errorf("error fetching voter info events: %s", err)
	}

	if len(regs) != len(infos) {
		return nil, errors.Errorf("voter registered and voter info event count mismatch: %d != %d", len(regs), len(infos))
	}

	var voters []*VoterInfo
	for i := range regs {
		if regs[i].RewardEpochId.Uint64() != uint64(epoch) {
			continue
		}

		voters = append(voters, &VoterInfo{
			Identity:          ty.VoterId(regs[i].Voter),
			Submit:            ty.VoterSubmit(regs[i].SubmitAddress),
			SubmitSignatures:  ty.VoterSubmitSignatures(regs[i].SubmitSignaturesAddress),
			Signing:           ty.VoterSigning(regs[i].SigningPolicyAddress),
			Delegation:        ty.VoterDelegation(infos[i].DelegationAddress),
			CappedWeight:      infos[i].WNatCappedWeight,
			DelegationFeeBips: infos[i].DelegationFeeBIPS,
			NodeIds:           infos[i].NodeIds,
			NodeWeights:       infos[i].NodeWeights,
		})

		logger.Info("voter %s, submit %s, submit signatures %s, signing policy %s", regs[i].Voter.String(), regs[i].SubmitAddress.String(), regs[i].SubmitSignaturesAddress.String(), regs[i].SigningPolicyAddress.String())
	}

	return NewVoterIndex(voters), nil
}

type VoterIndex struct {
	ById               map[ty.VoterId]*VoterInfo
	BySubmit           map[ty.VoterSubmit]*VoterInfo
	BySubmitSignatures map[ty.VoterSubmitSignatures]*VoterInfo
	BySigning          map[ty.VoterSigning]*VoterInfo
	ByDelegation       map[ty.VoterDelegation]*VoterInfo
	TotalCappedWeight  *big.Int
}

func NewVoterIndex(voters []*VoterInfo) *VoterIndex {
	byId := make(map[ty.VoterId]*VoterInfo)
	bySubmit := make(map[ty.VoterSubmit]*VoterInfo)
	bySubmitSignatures := make(map[ty.VoterSubmitSignatures]*VoterInfo)
	bySigning := make(map[ty.VoterSigning]*VoterInfo)
	byDelegation := make(map[ty.VoterDelegation]*VoterInfo)
	for _, v := range voters {
		byId[v.Identity] = v
		bySubmit[v.Submit] = v
		bySubmitSignatures[v.SubmitSignatures] = v
		bySigning[v.Signing] = v
		byDelegation[v.Delegation] = v
	}
	totalCappedWeight := big.NewInt(0)
	for _, v := range voters {
		totalCappedWeight.Add(totalCappedWeight, v.CappedWeight)
	}
	return &VoterIndex{
		ById:               byId,
		BySubmit:           bySubmit,
		BySubmitSignatures: bySubmitSignatures,
		BySigning:          bySigning,
		ByDelegation:       byDelegation,
		TotalCappedWeight:  totalCappedWeight,
	}
}

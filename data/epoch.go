package data

import (
	votersLib "fsp-rewards-calculator/lib"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/calculator"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fumanager"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/offers"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/relay"
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
	Policy        *votersLib.SigningPolicy
	Offers        RewardOffers
	OrderedFeeds  []Feed
	OrderedVoters []ty.VoterSigning
	VoterIndex    *VoterIndex
}

type RewardEpochs struct {
	prev    *RewardEpoch
	Current *RewardEpoch
	next    *RewardEpoch
}

func LoadRewardEpochs(epoch ty.EpochId, db *gorm.DB) (RewardEpochs, error) {
	prev, err := GetRewardEpoch(epoch-1, db)
	if err != nil {
		return RewardEpochs{}, errors.Wrap(err, "error fetching previous epoch")
	}
	current, err := GetRewardEpoch(epoch, db)
	if err != nil {
		return RewardEpochs{}, errors.Wrap(err, "error fetching current epoch")
	}
	next, err := GetRewardEpoch(epoch+1, db)
	if err != nil {
		return RewardEpochs{}, errors.Wrap(err, "error fetching next epoch")
	}
	return RewardEpochs{
		prev:    &prev,
		Current: &current,
		next:    &next,
	}, nil
}

func (re *RewardEpochs) EpochForRound(round ty.RoundId) *RewardEpoch {
	switch {
	case round < re.Current.StartRound:
		return re.prev
	case round > re.Current.EndRound:
		return re.next
	default:
		return re.Current
	}
}

type RewardOffers struct {
	Community    []*offers.OffersRewardsOffered
	Inflation    []*offers.OffersInflationRewardsOffered
	FastUpdates  []*fumanager.FUManagerInflationRewardsOffered
	FastUpdatesI []*fumanager.FUManagerIncentiveOffered
}

type VoterInfo struct {
	Identity            ty.VoterId
	Submit              ty.VoterSubmit
	SubmitSignatures    ty.VoterSubmitSignatures
	Signing             ty.VoterSigning
	Delegation          ty.VoterDelegation
	CappedWeight        *big.Int
	DelegationFeeBips   uint16
	NodeIds             [][20]byte
	NodeWeights         []*big.Int
	SigningPolicyWeight uint16
}

func GetRewardEpoch(epoch ty.EpochId, db *gorm.DB) (RewardEpoch, error) {
	currentTimestamp := time.Now().Unix()

	// TODO: Use lowest index in indexer db as start
	expectedStartSec := params.Net.Epoch.ExpectedRewardEpochStartTimeSec(epoch)
	epochDuration := params.Net.Epoch.RewardEpochDurationInVotingEpochs * params.Net.Epoch.VotingRoundDurationSeconds

	searchIntervalStartSec := expectedStartSec - (epochDuration * 2)
	searchIntervalEndSec := min(expectedStartSec+(epochDuration*2), uint64(currentTimestamp))

	relayInst, _ := relay.NewRelay(common.Address{}, nil)
	parsePolicyInitialized := func(log types.Log, _ uint64) (*relay.RelaySigningPolicyInitialized, error) {
		return relayInst.RelayFilterer.ParseSigningPolicyInitialized(log)
	}

	policies, _ := queryEvents(db, searchIntervalStartSec, searchIntervalEndSec, common.HexToAddress("0xea077600E3065F4FAd7161a6D0977741f2618eec"), utils.EventTopic0.SigningPolicyInitialized, parsePolicyInitialized)

	allPolicies, err := queryEvents(db, searchIntervalStartSec, searchIntervalEndSec, params.Net.Contracts.Relay, utils.EventTopic0.SigningPolicyInitialized, parsePolicyInitialized)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching signing policy events: %s", err)
	}

	allPolicies = append(policies, allPolicies...)

	var policyEvent *relay.RelaySigningPolicyInitialized
	var startRound ty.RoundId
	var endRound ty.RoundId

	for _, event := range allPolicies {
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
	logger.Info("Feeds: %v", len(feeds))
	for _, f := range feeds {
		logger.Info("Feed: %s, Decimals: %d", f.String(), f.Decimals)
	}

	signingPolicyWindow := params.Net.Epoch.NewSigningPolicyInitializationStartSeconds

	voters, err := getVoters(db, epoch, epochStartSec-signingPolicyWindow, epochStartSec)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching voter info: %s", err)
	}

	return RewardEpoch{
		Epoch:         epoch,
		StartRound:    startRound,
		EndRound:      endRound,
		Policy:        votersLib.NewSigningPolicy(policyEvent),
		Offers:        rewardOffers,
		OrderedFeeds:  feeds,
		OrderedVoters: getOrderedVoters(policyEvent),
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

	infoByIdentity := make(map[common.Address]*calculator.CalculatorVoterRegistrationInfo)
	for _, info := range infos {
		infoByIdentity[info.Voter] = info
	}

	if len(regs) != len(infos) {
		return nil, errors.Errorf("voter registered and voter info event count mismatch: %d != %d", len(regs), len(infos))
	}

	var voters []*VoterInfo
	for _, reg := range regs {
		if reg.RewardEpochId.Uint64() != uint64(epoch) {
			continue
		}

		info := infoByIdentity[reg.Voter]

		voters = append(voters, &VoterInfo{
			Identity:          ty.VoterId(reg.Voter),
			Submit:            ty.VoterSubmit(reg.SubmitAddress),
			SubmitSignatures:  ty.VoterSubmitSignatures(reg.SubmitSignaturesAddress),
			Signing:           ty.VoterSigning(reg.SigningPolicyAddress),
			Delegation:        ty.VoterDelegation(info.DelegationAddress),
			CappedWeight:      info.WNatCappedWeight,
			DelegationFeeBips: info.DelegationFeeBIPS,
			NodeIds:           info.NodeIds,
			NodeWeights:       info.NodeWeights,
		})
		logger.Info("Voter %s, submit %s, submit signatures %s, signing policy %s", reg.Voter.String(), reg.SubmitAddress.String(), reg.SubmitSignaturesAddress.String(), reg.SigningPolicyAddress.String())
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

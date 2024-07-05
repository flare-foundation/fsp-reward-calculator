package main

import (
	"flare-common/contracts/offers"
	"flare-common/contracts/relay"
	"flare-common/policy"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/types"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
	"slices"
	"time"
)

type RewardEpoch struct {
	Epoch        types.EpochId
	StartRound   types.RoundId
	EndRound     types.RoundId
	Policy       *policy.SigningPolicy
	Offers       RewardOffers
	OrderedFeeds []Feed
	Voters       *VoterIndex
	// TODO: Move next voter calculation elsewhere
	NextVoters *VoterIndex // Voters for the following reward epoch
}

type RewardOffers struct {
	community []*offers.OffersRewardsOffered
	inflation []*offers.OffersInflationRewardsOffered
}

type VoterInfo struct {
	Identity          VoterId
	Submit            VoterSubmit
	SubmitSignatures  VoterSubmitSignatures
	Signing           VoterSigning
	Delegation        VoterDelegation
	CappedWeight      *big.Int
	DelegationFeeBips uint16
	NodeIds           [][20]byte
	NodeWeights       []*big.Int
}

// TODO: Use proper timings for event search instead of approximate
func getRewardEpoch(epoch types.EpochId, db *gorm.DB) (RewardEpoch, error) {
	currentTimestamp := time.Now().Unix()

	// TODO: Use lowest index in indexer db as start
	expectedStartSec := params.Net.Epoch.ExpectedRewardEpochStartTimeSec(epoch)
	epochDuration := params.Net.Epoch.RewardEpochDurationInVotingEpochs * params.Net.Epoch.VotingRoundDurationSeconds

	searchIntervalStartSec := expectedStartSec - (epochDuration * 2)
	searchIntervalEndSec := min(expectedStartSec+(epochDuration*2), uint64(currentTimestamp))

	relayInst, _ := relay.NewRelay(common.Address{}, nil)
	parsePolicyInitialized := func(log etypes.Log) (*relay.RelaySigningPolicyInitialized, error) {
		return relayInst.RelayFilterer.ParseSigningPolicyInitialized(log)
	}
	policies, err := QueryEvents(db, searchIntervalStartSec, searchIntervalEndSec, params.Net.Contracts.Relay, utils.EventTopic0.SigningPolicyInitialized, parsePolicyInitialized)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching signing policy events: %s", err)
	}

	var policyEvent *relay.RelaySigningPolicyInitialized
	var startRound types.RoundId
	var endRound types.RoundId

	for _, event := range policies {
		if event.RewardEpochId.Uint64() == uint64(epoch) {
			policyEvent = event
			startRound = types.RoundId(event.StartVotingRoundId)
		}
		if event.RewardEpochId.Uint64() == uint64(epoch)+1 {
			endRound = types.RoundId(event.StartVotingRoundId - 1)
		}
	}

	if policyEvent == nil {
		return RewardEpoch{}, errors.Errorf("no signing policy found for epoch %d", epoch)
	}
	if endRound == 0 {
		return RewardEpoch{}, errors.Errorf("unable to determine last voting round for epoch %d: no signing policy found for next epoch %d. It may not have been indexed yet or the current epoch is not finished", epoch, epoch+1)
	}

	actualStartSec := params.Net.Epoch.VotingRoundStartSec(startRound)

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

	nextVoters, err := getVoters(db, epoch+1, actualStartSec, actualStartSec+epochDuration)
	if err != nil {
		return RewardEpoch{}, errors.Errorf("error fetching voter info: %s", err)
	}

	return RewardEpoch{
		Epoch:        epoch,
		StartRound:   startRound,
		EndRound:     endRound,
		Policy:       policy.NewSigningPolicy(policyEvent),
		Offers:       rewardOffers,
		OrderedFeeds: feeds,
		Voters:       voters,
		NextVoters:   nextVoters,
	}, nil
}

//
//func analyseReveals(revealMap map[uint64][]Reveal, feeds []Feed) {
//	for round, reveal := range revealMap {
//		feedValues := make(map[int][]int32)
//		invalidCount := make([]int, len(feeds))
//		validCount := make([]int, len(feeds))
//		for _, r := range reveal {
//			for feedIndex := range feeds {
//				if !r.Values[feedIndex].isEmpty {
//					if isPowerOfTen(int(r.Values[feedIndex].Value)) {
//						invalidCount[feedIndex]++
//					} else {
//						validCount[feedIndex]++
//					}
//				}
//				feedValues[feedIndex] = append(feedValues[feedIndex], r.Values[feedIndex].Value)
//			}
//		}
//
//		totalInvalid := 0
//
//		invalidFeeds := make([]string, 0)
//		for i, v := range feedValues {
//			invalidp := float64(invalidCount[i]) / float64(invalidCount[i]+validCount[i]) * 100
//			feedS := feeds[i].String()
//			feeds2 := feedS
//			if invalidp >= 50 {
//				totalInvalid++
//				invalidFeeds = append(invalidFeeds, feedS)
//			}
//			fmt.Printf("Round %d, feed %10s, total %2d, valid %2d, invalid %2d, invalid%% %.2f: %v\n", round, feeds2, invalidCount[i]+validCount[i], validCount[i], invalidCount[i], invalidp, v)
//		}
//
//		sort.Slice(invalidFeeds, func(i, j int) bool {
//			return invalidFeeds[i] < invalidFeeds[j]
//		})
//
//		fmt.Printf("Round %d, total invalid > 50%%: %d\n, feeds: %v", round, totalInvalid, invalidFeeds)
//		break
//	}
//}

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

func getRewardOffers(db *gorm.DB, epoch types.EpochId, epochStartSec, epochEndSec uint64) (RewardOffers, error) {
	community, err := GetRewardOfferEvents(db, epochStartSec, epochEndSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching reward offer events: %s", err)
	}
	inflation, err := GetInflationRewardOfferEvents(db, epochStartSec, epochEndSec)
	if err != nil {
		return RewardOffers{}, errors.Errorf("error fetching inflation reward offer events: %s", err)
	}

	community = slices.DeleteFunc(community, func(offer *offers.OffersRewardsOffered) bool {
		return offer.RewardEpochId.Uint64() != uint64(epoch)
	})
	inflation = slices.DeleteFunc(inflation, func(offer *offers.OffersInflationRewardsOffered) bool {
		return offer.RewardEpochId.Uint64() != uint64(epoch)
	})

	return RewardOffers{
		community,
		inflation,
	}, nil
}

func getVoters(db *gorm.DB, epoch types.EpochId, fromSec, toSec uint64) (*VoterIndex, error) {
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
			Identity:          VoterId(regs[i].Voter),
			Submit:            VoterSubmit(regs[i].SubmitAddress),
			SubmitSignatures:  VoterSubmitSignatures(regs[i].SubmitSignaturesAddress),
			Signing:           VoterSigning(regs[i].SigningPolicyAddress),
			Delegation:        VoterDelegation(infos[i].DelegationAddress),
			CappedWeight:      infos[i].WNatCappedWeight,
			DelegationFeeBips: infos[i].DelegationFeeBIPS,
			NodeIds:           infos[i].NodeIds,
			NodeWeights:       infos[i].NodeWeights,
		})

		logger.Info("voter %s, submit %s, submit signatures %s, signing policy %s", regs[i].Voter.String(), regs[i].SubmitAddress.String(), regs[i].SubmitSignaturesAddress.String(), regs[i].SigningPolicyAddress.String())
	}

	return NewVoterIndex(voters), nil
}

type VoterId common.Address
type VoterSubmit common.Address
type VoterSubmitSignatures common.Address
type VoterSigning common.Address
type VoterDelegation common.Address

type VoterIndex struct {
	byId               map[VoterId]*VoterInfo
	bySubmit           map[VoterSubmit]*VoterInfo
	bySubmitSignatures map[VoterSubmitSignatures]*VoterInfo
	bySigning          map[VoterSigning]*VoterInfo
	byDelegation       map[VoterDelegation]*VoterInfo
	totalCappedWeight  *big.Int
}

func NewVoterIndex(voters []*VoterInfo) *VoterIndex {
	byId := make(map[VoterId]*VoterInfo)
	bySubmit := make(map[VoterSubmit]*VoterInfo)
	bySubmitSignatures := make(map[VoterSubmitSignatures]*VoterInfo)
	bySigning := make(map[VoterSigning]*VoterInfo)
	for _, v := range voters {
		byId[v.Identity] = v
		bySubmit[v.Submit] = v
		bySubmitSignatures[v.SubmitSignatures] = v
		bySigning[v.Signing] = v
	}
	totalCappedWeight := big.NewInt(0)
	for _, v := range voters {
		totalCappedWeight.Add(totalCappedWeight, v.CappedWeight)
	}
	return &VoterIndex{
		byId:               byId,
		bySubmit:           bySubmit,
		bySubmitSignatures: bySubmitSignatures,
		bySigning:          bySigning,
		totalCappedWeight:  totalCappedWeight,
	}
}

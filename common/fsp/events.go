package fsp

import (
	common2 "fsp-rewards-calculator/common"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/contracts/calculator"
	"fsp-rewards-calculator/contracts/registryOld"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	calculatorOld "github.com/flare-foundation/go-flare-common/pkg/contracts/calculator"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fdchub"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fumanager"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/offers"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/registry"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type VoterRegisteredEvent struct {
	Voter                   common.Address
	RewardEpochId           uint64
	SigningPolicyAddress    common.Address
	SubmitAddress           common.Address
	SubmitSignaturesAddress common.Address
}

// Fetches VoterRegistered events from indexer. Registrations may live on the old or the new (post
// reward epoch 417 upgrade) VoterRegistry contract, and the two contracts use different event ABIs,
// so we query both addresses with their respective event signatures (like Relay/OldRelay).
func getVoterRegisteredEvents(db *gorm.DB, from uint64, to uint64) ([]VoterRegisteredEvent, error) {
	oldRegistry, _ := registryOld.NewRegistry(common.Address{}, nil)
	parseOld := func(log types.Log, _ uint64) (*registryOld.RegistryVoterRegistered, error) {
		return oldRegistry.ParseVoterRegistered(log)
	}
	oldEvents, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.OldVoterRegistry,
		common2.EventTopic0.VoterRegisteredOld,
		parseOld,
	)
	if err != nil {
		return nil, errors.Errorf("error fetching voter registered events: %s", err)
	}

	newRegistry, _ := registry.NewRegistry(common.Address{}, nil)
	parseNew := func(log types.Log, _ uint64) (*registry.RegistryVoterRegistered, error) {
		return newRegistry.ParseVoterRegistered(log)
	}
	// The new event signature is queried on both addresses: some networks' pre-upgrade registry
	// already emits the new ABI.
	newEvents, err := QueryEventsForContracts(
		db,
		from,
		to,
		[]common.Address{params.Net.Contracts.OldVoterRegistry, params.Net.Contracts.VoterRegistry},
		common2.EventTopic0.VoterRegistered,
		parseNew,
	)
	if err != nil {
		return nil, errors.Errorf("error fetching voter registered events: %s", err)
	}

	events := make([]VoterRegisteredEvent, 0, len(oldEvents)+len(newEvents))
	for _, event := range oldEvents {
		events = append(events, VoterRegisteredEvent{
			Voter:                   event.Voter,
			RewardEpochId:           event.RewardEpochId.Uint64(),
			SigningPolicyAddress:    event.SigningPolicyAddress,
			SubmitAddress:           event.SubmitAddress,
			SubmitSignaturesAddress: event.SubmitSignaturesAddress,
		})
	}
	for _, event := range newEvents {
		events = append(events, VoterRegisteredEvent{
			Voter:                   event.Voter,
			RewardEpochId:           uint64(event.RewardEpochId),
			SigningPolicyAddress:    event.SigningPolicyAddress,
			SubmitAddress:           event.SubmitAddress,
			SubmitSignaturesAddress: event.SubmitSignaturesAddress,
		})
	}

	return events, nil
}

// Fetches VoterRegistrationInfo events from indexer. Like the voter registry, the
// FlareSystemsCalculator was replaced with the reward epoch 417 upgrade and the old contract emits
// the event with a different signature (rewardEpochId uint24 instead of uint32), so we query both
// addresses with their respective event signatures.
func getVoterInfoEvents(db *gorm.DB, from uint64, to uint64) ([]*calculator.CalculatorVoterRegistrationInfo, error) {
	oldCalculator, _ := calculatorOld.NewCalculator(common.Address{}, nil)
	parseOld := func(log types.Log, _ uint64) (*calculatorOld.CalculatorVoterRegistrationInfo, error) {
		return oldCalculator.ParseVoterRegistrationInfo(log)
	}

	oldEvents, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.OldFlareSystemsCalculator,
		common2.EventTopic0.VoterRegistrationInfoOld,
		parseOld,
	)
	if err != nil {
		return nil, errors.Errorf("error fetching events: %s", err)
	}

	newCalculator, _ := calculator.NewCalculator(common.Address{}, nil)
	parseNew := func(log types.Log, _ uint64) (*calculator.CalculatorVoterRegistrationInfo, error) {
		return newCalculator.ParseVoterRegistrationInfo(log)
	}

	events, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FlareSystemsCalculator,
		common2.EventTopic0.VoterRegistrationInfo,
		parseNew,
	)
	if err != nil {
		return nil, errors.Errorf("error fetching events: %s", err)
	}

	for _, event := range oldEvents {
		events = append(events, &calculator.CalculatorVoterRegistrationInfo{
			Voter:             event.Voter,
			RewardEpochId:     uint32(event.RewardEpochId.Uint64()),
			DelegationAddress: event.DelegationAddress,
			DelegationFeeBIPS: event.DelegationFeeBIPS,
			WNatWeight:        event.WNatWeight,
			WNatCappedWeight:  event.WNatCappedWeight,
			NodeIds:           event.NodeIds,
			NodeWeights:       event.NodeWeights,
			Raw:               event.Raw,
		})
	}

	return events, nil
}

func getRewardOfferEvents(db *gorm.DB, from uint64, to uint64) ([]*offers.OffersRewardsOffered, error) {
	instance, _ := offers.NewOffers(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*offers.OffersRewardsOffered, error) {
		return instance.ParseRewardsOffered(log)
	}

	events, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FtsoRewardOffersManager,
		common2.EventTopic0.RewardsOffered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

func getInflationRewardOfferEvents(db *gorm.DB, from uint64, to uint64) ([]*offers.OffersInflationRewardsOffered, error) {
	instance, _ := offers.NewOffers(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*offers.OffersInflationRewardsOffered, error) {
		return instance.ParseInflationRewardsOffered(log)
	}

	events, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FtsoRewardOffersManager,
		common2.EventTopic0.InflationRewardsOffered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

func getFURewardOfferEvents(db *gorm.DB, from uint64, to uint64) ([]*fumanager.FUManagerInflationRewardsOffered, error) {
	instance, _ := fumanager.NewFUManager(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*fumanager.FUManagerInflationRewardsOffered, error) {
		return instance.ParseInflationRewardsOffered(log)
	}

	events, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FastUpdateIncentiveManager,
		common2.EventTopic0.FUInflationRewardsOffered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

func getFUIncentiveOfferEvents(db *gorm.DB, from uint64, to uint64) ([]*fumanager.FUManagerIncentiveOffered, error) {
	instance, _ := fumanager.NewFUManager(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*fumanager.FUManagerIncentiveOffered, error) {
		return instance.ParseIncentiveOffered(log)
	}

	events, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FastUpdateIncentiveManager,
		common2.EventTopic0.FUIncentiveRewardOffered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

func getFdcInflationRewardOfferEvents(db *gorm.DB, from uint64, to uint64) ([]*fdchub.FdcHubInflationRewardsOffered, error) {
	instance, _ := fdchub.NewFdcHub(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*fdchub.FdcHubInflationRewardsOffered, error) {
		return instance.ParseInflationRewardsOffered(log)
	}

	events, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FdcHub,
		common2.EventTopic0.FdcInflationRewardsOffered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

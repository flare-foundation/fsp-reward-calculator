package fsp

import (
	common2 "fsp-rewards-calculator/common"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/contracts/registryOld"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/calculator"
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

// Fetches VoterRegistered events from indexer. Some networks use updated VoterRegistry contract with new ABI,
// so we try to query for both old and new event signature.
func getVoterRegisteredEvents(db *gorm.DB, from uint64, to uint64) ([]VoterRegisteredEvent, error) {
	oldRegistry, _ := registryOld.NewRegistry(common.Address{}, nil)
	parseOld := func(log types.Log, _ uint64) (*registryOld.RegistryVoterRegistered, error) {
		return oldRegistry.ParseVoterRegistered(log)
	}
	oldEvents, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.VoterRegistry,
		common2.EventTopic0.VoterRegisteredOld,
		parseOld,
	)
	if err != nil {
		return nil, errors.Errorf("error fetching voter registered events: %s", err)
	}
	if len(oldEvents) > 0 {
		events := make([]VoterRegisteredEvent, 0, len(oldEvents))
		for _, event := range oldEvents {
			events = append(events, VoterRegisteredEvent{
				Voter:                   event.Voter,
				RewardEpochId:           event.RewardEpochId.Uint64(),
				SigningPolicyAddress:    event.SigningPolicyAddress,
				SubmitAddress:           event.SubmitAddress,
				SubmitSignaturesAddress: event.SubmitSignaturesAddress,
			})
		}
		return events, nil
	}

	newRegistry, _ := registry.NewRegistry(common.Address{}, nil)
	parseNew := func(log types.Log, _ uint64) (*registry.RegistryVoterRegistered, error) {
		return newRegistry.ParseVoterRegistered(log)
	}
	newEvents, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.VoterRegistry,
		common2.EventTopic0.VoterRegistered,
		parseNew,
	)
	if err != nil {
		return nil, errors.Errorf("error fetching voter registered events: %s", err)
	}

	events := make([]VoterRegisteredEvent, 0, len(newEvents))
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

func getVoterInfoEvents(db *gorm.DB, from uint64, to uint64) ([]*calculator.CalculatorVoterRegistrationInfo, error) {
	instance, _ := calculator.NewCalculator(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*calculator.CalculatorVoterRegistrationInfo, error) {
		return instance.ParseVoterRegistrationInfo(log)
	}

	events, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FlareSystemsCalculator,
		common2.EventTopic0.VoterRegistrationInfo,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("error fetching events: %s", err)
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

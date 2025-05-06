package fsp

import (
	common2 "fsp-rewards-calculator/common"
	"fsp-rewards-calculator/common/params"
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

func getVoterRegisteredEvents(db *gorm.DB, from uint64, to uint64) ([]*registry.RegistryVoterRegistered, error) {
	instance, _ := registry.NewRegistry(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*registry.RegistryVoterRegistered, error) {
		return instance.RegistryFilterer.ParseVoterRegistered(log)
	}

	events, err := QueryEvents(
		db,
		from,
		to,
		params.Net.Contracts.VoterRegistry,
		common2.EventTopic0.VoterRegistered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

func getVoterInfoEvents(db *gorm.DB, from uint64, to uint64) ([]*calculator.CalculatorVoterRegistrationInfo, error) {
	instance, _ := calculator.NewCalculator(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*calculator.CalculatorVoterRegistrationInfo, error) {
		return instance.CalculatorFilterer.ParseVoterRegistrationInfo(log)
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
		return instance.OffersFilterer.ParseRewardsOffered(log)
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
		return instance.OffersFilterer.ParseInflationRewardsOffered(log)
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
		return instance.FUManagerFilterer.ParseInflationRewardsOffered(log)
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
		return instance.FUManagerFilterer.ParseIncentiveOffered(log)
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
		return instance.FdcHubFilterer.ParseInflationRewardsOffered(log)
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

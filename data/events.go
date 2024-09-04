package data

import (
	"flare-common/contracts/calculator"
	"flare-common/contracts/fumanager"
	"flare-common/contracts/offers"
	"flare-common/contracts/registry"
	"flare-common/database"
	"flare-common/events"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/utils"
	"strings"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

func queryEvents[T interface{}](
	db *gorm.DB,
	searchIntervalStartSec uint64, //inclusive
	searchIntervalEndSec uint64, //exclusive
	contractAddress common.Address,
	topic0 string,
	parseEvent func(types.Log) (T, error),
) ([]T, error) {
	var logs []database.Log
	err := db.Where(
		"address = ? AND topic0 = ? AND timestamp >= ? AND timestamp < ?",
		strings.ToLower(strings.TrimPrefix(contractAddress.String(), "0x")),
		strings.ToLower(strings.TrimPrefix(topic0, "0x")),
		searchIntervalStartSec, searchIntervalEndSec,
	).Order("timestamp").Find(&logs).Error
	if err != nil {
		return nil, errors.Errorf("error fetching logs From DB: %s", err)
	}

	var parsedEvents []T
	for _, log := range logs {
		chainLog, err := events.ConvertDatabaseLogToChainLog(log)
		if err != nil {
			logger.Error("error converting database log to chain log: %s", err)
			continue
		}
		parsed, err := parseEvent(*chainLog)
		if err != nil {
			logger.Error("error parsing event, ignoring: %s", err)
			continue
		}
		parsedEvents = append(parsedEvents, parsed)
	}
	return parsedEvents, nil
}

func GetVoterRegisteredEvents(db *gorm.DB, from uint64, to uint64) ([]*registry.RegistryVoterRegistered, error) {
	instance, _ := registry.NewRegistry(common.Address{}, nil)
	parse := func(log types.Log) (*registry.RegistryVoterRegistered, error) {
		return instance.RegistryFilterer.ParseVoterRegistered(log)
	}

	events, err := queryEvents(
		db,
		from,
		to,
		params.Net.Contracts.VoterRegistry,
		utils.EventTopic0.VoterRegistered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

func GetVoterInfoEvents(db *gorm.DB, from uint64, to uint64) ([]*calculator.CalculatorVoterRegistrationInfo, error) {
	instance, _ := calculator.NewCalculator(common.Address{}, nil)
	parse := func(log types.Log) (*calculator.CalculatorVoterRegistrationInfo, error) {
		return instance.CalculatorFilterer.ParseVoterRegistrationInfo(log)
	}

	events, err := queryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FlareSystemsCalculator,
		utils.EventTopic0.VoterRegistrationInfo,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("error fetching events: %s", err)
	}

	return events, nil
}

func GetRewardOfferEvents(db *gorm.DB, from uint64, to uint64) ([]*offers.OffersRewardsOffered, error) {
	instance, _ := offers.NewOffers(common.Address{}, nil)
	parse := func(log types.Log) (*offers.OffersRewardsOffered, error) {
		return instance.OffersFilterer.ParseRewardsOffered(log)
	}

	events, err := queryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FtsoRewardOffersManager,
		utils.EventTopic0.RewardsOffered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

func GetInflationRewardOfferEvents(db *gorm.DB, from uint64, to uint64) ([]*offers.OffersInflationRewardsOffered, error) {
	instance, _ := offers.NewOffers(common.Address{}, nil)
	parse := func(log types.Log) (*offers.OffersInflationRewardsOffered, error) {
		return instance.OffersFilterer.ParseInflationRewardsOffered(log)
	}

	events, err := queryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FtsoRewardOffersManager,
		utils.EventTopic0.InflationRewardsOffered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

func GetFURewardOfferEvents(db *gorm.DB, from uint64, to uint64) ([]*fumanager.FUManagerInflationRewardsOffered, error) {
	instance, _ := fumanager.NewFUManager(common.Address{}, nil)
	parse := func(log types.Log) (*fumanager.FUManagerInflationRewardsOffered, error) {
		return instance.FUManagerFilterer.ParseInflationRewardsOffered(log)
	}

	events, err := queryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FastUpdateIncentiveManager,
		utils.EventTopic0.FUInflationRewardsOffered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

func GetFUIncentiveOfferEvents(db *gorm.DB, from uint64, to uint64) ([]*fumanager.FUManagerIncentiveOffered, error) {
	instance, _ := fumanager.NewFUManager(common.Address{}, nil)
	parse := func(log types.Log) (*fumanager.FUManagerIncentiveOffered, error) {
		return instance.FUManagerFilterer.ParseIncentiveOffered(log)
	}

	events, err := queryEvents(
		db,
		from,
		to,
		params.Net.Contracts.FastUpdateIncentiveManager,
		utils.EventTopic0.FUIncentiveRewardOffered,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	return events, nil
}

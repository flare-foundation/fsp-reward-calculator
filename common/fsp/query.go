package fsp

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"github.com/flare-foundation/go-flare-common/pkg/events"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func QueryEvents[T interface{}](
	db *gorm.DB,
	searchIntervalStartSec uint64, //inclusive
	searchIntervalEndSec uint64, //exclusive
	contractAddress common.Address,
	topic0 string,
	parseEvent func(types.Log, uint64) (T, error),
) ([]T, error) {
	return QueryEventsForContracts(
		db,
		searchIntervalStartSec,
		searchIntervalEndSec,
		[]common.Address{contractAddress},
		topic0,
		parseEvent,
	)
}

func QueryEventsForContracts[T interface{}](
	db *gorm.DB,
	searchIntervalStartSec uint64, //inclusive
	searchIntervalEndSec uint64, //exclusive
	contractAddresses []common.Address,
	topic0 string,
	parseEvent func(types.Log, uint64) (T, error),
) ([]T, error) {
	var logs []database.Log
	addresses := make([]string, 0, len(contractAddresses))
	for _, contractAddress := range contractAddresses {
		addresses = append(addresses, strings.ToLower(strings.TrimPrefix(contractAddress.String(), "0x")))
	}

	err := db.Where(
		"address IN ? AND topic0 = ? AND timestamp >= ? AND timestamp < ?",
		addresses,
		strings.ToLower(strings.TrimPrefix(topic0, "0x")),
		searchIntervalStartSec, searchIntervalEndSec,
	).
		Order("timestamp").
		Order("block_number").
		Order("log_index").
		Find(&logs).Error
	if err != nil {
		return nil, errors.Errorf("error fetching logs From DB: %s", err)
	}
	return parseEventLogs(logs, parseEvent)
}

func parseEventLogs[T interface{}](logs []database.Log, parseEvent func(types.Log, uint64) (T, error)) ([]T, error) {
	var parsedEvents []T
	for _, log := range logs {
		chainLog, err := events.ConvertDatabaseLogToChainLog(log)
		if err != nil {
			return nil, errors.Errorf("error converting event at block %d log index %d: %s", log.BlockNumber, log.LogIndex, err)
		}
		parsed, err := parseEvent(*chainLog, log.Timestamp)
		if err != nil {
			return nil, errors.Errorf("error parsing event at block %d log index %d: %s", log.BlockNumber, log.LogIndex, err)
		}
		parsedEvents = append(parsedEvents, parsed)
	}
	return parsedEvents, nil
}

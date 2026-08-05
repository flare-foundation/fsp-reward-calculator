package fsp

import (
	"fsp-rewards-calculator/logger"
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

// QueryEventsStrict queries and decodes events like QueryEvents, but returns an error instead of
// ignoring a log that cannot be converted or decoded. Use it where dropping one event would change
// accounting results.
func QueryEventsStrict[T interface{}](
	db *gorm.DB,
	searchIntervalStartSec uint64, //inclusive
	searchIntervalEndSec uint64, //exclusive
	contractAddress common.Address,
	topic0 string,
	parseEvent func(types.Log, uint64) (T, error),
) ([]T, error) {
	return queryEventsForContracts(
		db,
		searchIntervalStartSec,
		searchIntervalEndSec,
		[]common.Address{contractAddress},
		topic0,
		parseEvent,
		true,
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
	return queryEventsForContracts(
		db,
		searchIntervalStartSec,
		searchIntervalEndSec,
		contractAddresses,
		topic0,
		parseEvent,
		false,
	)
}

func queryEventsForContracts[T interface{}](
	db *gorm.DB,
	searchIntervalStartSec uint64,
	searchIntervalEndSec uint64,
	contractAddresses []common.Address,
	topic0 string,
	parseEvent func(types.Log, uint64) (T, error),
	strict bool,
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
	return parseEventLogs(logs, parseEvent, strict)
}

func parseEventLogs[T interface{}](logs []database.Log, parseEvent func(types.Log, uint64) (T, error), strict bool) ([]T, error) {
	var parsedEvents []T
	for _, log := range logs {
		chainLog, err := events.ConvertDatabaseLogToChainLog(log)
		if err != nil {
			if strict {
				return nil, errors.Errorf("error converting event at block %d log index %d: %s", log.BlockNumber, log.LogIndex, err)
			}
			logger.Error("error converting database log to chain log: %s", err)
			continue
		}
		parsed, err := parseEvent(*chainLog, log.Timestamp)
		if err != nil {
			if strict {
				return nil, errors.Errorf("error parsing event at block %d log index %d: %s", log.BlockNumber, log.LogIndex, err)
			}
			logger.Error("error parsing event, ignoring: %s", err)
			continue
		}
		parsedEvents = append(parsedEvents, parsed)
	}
	return parsedEvents, nil
}

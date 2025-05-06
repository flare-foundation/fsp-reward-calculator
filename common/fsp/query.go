package fsp

import (
	"fsp-rewards-calculator/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/database"
	"github.com/flare-foundation/go-flare-common/pkg/events"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

func QueryEvents[T interface{}](
	db *gorm.DB,
	searchIntervalStartSec uint64, //inclusive
	searchIntervalEndSec uint64, //exclusive
	contractAddress common.Address,
	topic0 string,
	parseEvent func(types.Log, uint64) (T, error),
) ([]T, error) {
	var logs []database.Log
	err := db.Where(
		"address = ? AND topic0 = ? AND timestamp >= ? AND timestamp < ?",
		strings.ToLower(strings.TrimPrefix(contractAddress.String(), "0x")),
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

	var parsedEvents []T
	for _, log := range logs {
		chainLog, err := events.ConvertDatabaseLogToChainLog(log)
		if err != nil {
			logger.Error("error converting database log to chain log: %s", err)
			continue
		}
		parsed, err := parseEvent(*chainLog, log.Timestamp)
		if err != nil {
			logger.Error("error parsing event, ignoring: %s", err)
			continue
		}
		parsedEvents = append(parsedEvents, parsed)
	}
	return parsedEvents, nil
}

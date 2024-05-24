package main

import (
	"flare-common/contracts/registry"
	"flare-common/contracts/relay"
	"flare-common/database"
	"flare-common/events"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/parameters"
	"ftsov2-rewarding/utils"
	"strings"

	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

func main() {
	config := database.DBConfig{
		Host:     "localhost",
		Port:     3306,
		Database: "flare_ftso_indexer",
		Username: "root",
		Password: "root",
	}

	db, err := database.Connect(&config)

	if err != nil {
		logger.Fatal("Error connecting to database: %s", err)
	}

	calculateRewards(db, 2690)
}

type RewardClaim struct {
}

type RewardEpoch struct {
	Epoch      int
	StartRound int
	EndRound   int
	Policy     *relay.RelaySigningPolicyInitialized
}

func calculateRewards(db *gorm.DB, epoch int) (RewardClaim, error) {
	_, error := getRewardEpoch(epoch, db)
	if error != nil {
		return RewardClaim{}, errors.Errorf("error fetching reward epoch: %s", error)
	}

	return RewardClaim{}, nil
}

func getRewardEpoch(epoch int, db *gorm.DB) (RewardEpoch, error) {
	currentTimestamp := time.Now().Unix()

	// TODO: Use lowest index in indexer db as start

	expectedStartSec := parameters.Coston.Epoch.ExpectedRewardEpochStartTimeSec(int64(epoch))
	epochDuration := parameters.Coston.Epoch.RewardEpochDurationInVotingEpochs * parameters.Coston.Epoch.VotingRoundDurationSeconds

	searchIntervalStartSec := expectedStartSec - (epochDuration * 2)
	searchIntervalEndSec := min(expectedStartSec+(epochDuration*2), currentTimestamp)

	relayInst, _ := relay.NewRelay(common.Address{}, nil)
	persePolicyInitialized := func(log types.Log) (*relay.RelaySigningPolicyInitialized, error) {
		return relayInst.RelayFilterer.ParseSigningPolicyInitialized(log)
	}
	policies, error := queryEvents(db, searchIntervalStartSec, searchIntervalEndSec, parameters.Coston.Contracts.Relay, utils.EventTopic0.SigningPolicyInitialized, persePolicyInitialized)
	if error != nil {
		return RewardEpoch{}, errors.Errorf("Error fetching signing policy events: %s", error)
	}

	var rewardEpoch = RewardEpoch{}
	for _, event := range policies {
		if event.RewardEpochId.Int64() == int64(epoch) {
			rewardEpoch.Policy = event
			rewardEpoch.StartRound = int(event.StartVotingRoundId)
		}
		if event.RewardEpochId.Int64() == int64(epoch+1) {
			rewardEpoch.EndRound = int(event.StartVotingRoundId) - 1
		}
	}

	if rewardEpoch.Policy == nil {
		return RewardEpoch{}, errors.Errorf("No signing policy found for epoch %d", epoch)
	}
	if rewardEpoch.EndRound == 0 {
		return RewardEpoch{}, errors.Errorf("Unable to determine last voting round for epoch %d: no signing policy found for next epoch %d. It may not have been indexed yet or the current epoch is not finished", epoch, epoch+1)
	}

	// actualStartSec := parameters.Coston.Epoch.VotingEpochStartSec(int64(rewardEpoch.StartRound))
	// actualEndSec := parameters.Coston.Epoch.VotingEpochEndSec(int64(rewardEpoch.EndRound))

	return rewardEpoch, nil
}

func queryEvents[T interface{}](
	db *gorm.DB,
	searchIntervalStartSec int64,
	searchIntervalEndSec int64,
	contractAddress common.Address,
	topic0 string,
	parseEvent func(types.Log) (T, error),
) ([]T, error) {
	var logs []database.Log
	err := db.Where(
		"address = ? AND topic0 = ? AND timestamp > ? AND timestamp <= ?",
		strings.ToLower(strings.TrimPrefix(contractAddress.String(), "0x")),
		strings.ToLower(strings.TrimPrefix(topic0, "0x")),
		searchIntervalStartSec, searchIntervalEndSec,
	).Order("timestamp").Find(&logs).Error
	if err != nil {
		return nil, errors.Errorf("error fetching logs from DB: %s", err)
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

func getVoterRegistered(db *gorm.DB, from int64, to int64, currentTimestamp int64) ([]*registry.RegistryVoterRegistered, error) {
	registryInstance, _ := registry.NewRegistry(common.Address{}, nil)
	parse := func(log types.Log) (*registry.RegistryVoterRegistered, error) {
		return registryInstance.RegistryFilterer.ParseVoterRegistered(log)
	}

	regInfo, error := queryEvents(db, from, to, parameters.Coston.Contracts.VoterRegistry,
		utils.EventTopic0.VoterRegistered, parse)
	if error != nil {
		return nil, errors.Errorf("error fetching voter registry events: %s", error)
	}

	return regInfo, nil
}

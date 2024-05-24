package utils

import (
	"flare-common/contracts/calculator"
	"flare-common/contracts/offers"
	"flare-common/contracts/registry"
	"flare-common/contracts/relay"
	"flare-common/contracts/system"
	"ftsov2-rewarding/logger"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type EventIds struct {
	RewardsOffered          string
	InflationRewardsOffered string

	RewardEpochStarted       string
	RandomAcquisitionStarted string
	VotePowerBlockSelected   string

	SigningPolicyInitialized string
	VoterRegistered          string
	VoterRegistrationInfo    string
}

func EventIDFromMetadata(metaData *bind.MetaData, eventName string) string {
	abi, err := metaData.GetAbi()
	if err != nil {
		logger.Fatal("Error getting abi for event %s: %s", eventName, err)
		return ""
	}

	event, ok := abi.Events[eventName]
	if !ok {
		logger.Fatal("Error getting event id for event %s", eventName)
	}

	return event.ID.String()
}

var EventTopic0 = EventIds{
	RewardsOffered:          EventIDFromMetadata(offers.OffersMetaData, "RewardsOffered"),
	InflationRewardsOffered: EventIDFromMetadata(offers.OffersMetaData, "InflationRewardsOffered"),

	RewardEpochStarted:       EventIDFromMetadata(system.FlareSystemsManagerMetaData, "RewardEpochStarted"),
	RandomAcquisitionStarted: EventIDFromMetadata(system.FlareSystemsManagerMetaData, "RandomAcquisitionStarted"),
	VotePowerBlockSelected:   EventIDFromMetadata(system.FlareSystemsManagerMetaData, "VotePowerBlockSelected"),

	SigningPolicyInitialized: EventIDFromMetadata(relay.RelayMetaData, "SigningPolicyInitialized"),
	VoterRegistered:          EventIDFromMetadata(registry.RegistryMetaData, "VoterRegistered"),
	VoterRegistrationInfo:    EventIDFromMetadata(calculator.CalculatorMetaData, "VoterRegistrationInfo"),
}

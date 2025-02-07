package utils

import (
	"fsp-rewards-calculator/logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/calculator"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fdchub"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fumanager"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fupdater"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/offers"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/registry"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/relay"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/submission"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/system"
)

type EventIds struct {
	RewardsOffered             string
	InflationRewardsOffered    string
	FUInflationRewardsOffered  string
	FUIncentiveRewardOffered   string
	FdcInflationRewardsOffered string

	FastUpdateFeeds          string
	FastUpdateFeedsSubmitted string

	RewardEpochStarted       string
	RandomAcquisitionStarted string
	VotePowerBlockSelected   string

	SigningPolicyInitialized string
	VoterRegistered          string
	VoterRegistrationInfo    string

	FdcAttestationRequest string
}

type FunctionSigs struct {
	Submit1          [4]byte
	Submit2          [4]byte
	SubmitSignatures [4]byte
	Relay            [4]byte
}

var EventTopic0 = EventIds{
	RewardsOffered:             eventIDFromMetadata(offers.OffersMetaData, "RewardsOffered"),
	InflationRewardsOffered:    eventIDFromMetadata(offers.OffersMetaData, "InflationRewardsOffered"),
	FUInflationRewardsOffered:  eventIDFromMetadata(fumanager.FUManagerMetaData, "InflationRewardsOffered"),
	FUIncentiveRewardOffered:   eventIDFromMetadata(fumanager.FUManagerMetaData, "IncentiveOffered"),
	FdcInflationRewardsOffered: eventIDFromMetadata(fdchub.FdcHubMetaData, "InflationRewardsOffered"),

	RewardEpochStarted:       eventIDFromMetadata(system.FlareSystemsManagerMetaData, "RewardEpochStarted"),
	RandomAcquisitionStarted: eventIDFromMetadata(system.FlareSystemsManagerMetaData, "RandomAcquisitionStarted"),
	VotePowerBlockSelected:   eventIDFromMetadata(system.FlareSystemsManagerMetaData, "VotePowerBlockSelected"),

	SigningPolicyInitialized: eventIDFromMetadata(relay.RelayMetaData, "SigningPolicyInitialized"),
	VoterRegistered:          eventIDFromMetadata(registry.RegistryMetaData, "VoterRegistered"),
	VoterRegistrationInfo:    eventIDFromMetadata(calculator.CalculatorMetaData, "VoterRegistrationInfo"),

	FastUpdateFeeds:          eventIDFromMetadata(fupdater.FUpdaterMetaData, "FastUpdateFeeds"),
	FastUpdateFeedsSubmitted: eventIDFromMetadata(fupdater.FUpdaterMetaData, "FastUpdateFeedsSubmitted"),

	FdcAttestationRequest: eventIDFromMetadata(fdchub.FdcHubMetaData, "AttestationRequest"),
}

var FunctionSignatures = FunctionSigs{
	Submit1:          functionSigFromMetadata(submission.SubmissionMetaData, "submit1"),
	Submit2:          functionSigFromMetadata(submission.SubmissionMetaData, "submit2"),
	SubmitSignatures: functionSigFromMetadata(submission.SubmissionMetaData, "submitSignatures"),
	Relay:            functionSigFromMetadata(relay.RelayMetaData, "relay"),
}

func eventIDFromMetadata(metaData *bind.MetaData, eventName string) string {
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

func functionSigFromMetadata(metaData *bind.MetaData, functionName string) [4]byte {
	abi, err := metaData.GetAbi()
	if err != nil {
		logger.Fatal("Error getting abi for function %s: %s", functionName, err)
		return [4]byte{}
	}

	method, ok := abi.Methods[functionName]
	if !ok {
		logger.Fatal("Error getting signature for function %s", functionName)
	}

	return [4]byte(method.ID)
}

package main

import (
	"flare-common/database"
	"flare-common/payload"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	FtsoScalingProtocolId = 100
)

var (
	// TODO Read from config
	SubmissionContractAddress = common.HexToAddress("0x2cA6571Daa15ce734Bbd0Bf27D5C9D16787fc33f") // Coston

)

func getCommits(db *gorm.DB, searchIntervalStartSec int64, searchIntervalEndSec int64) ([]Commit, error) {
	msgs, err := queryMessages(db, searchIntervalStartSec, searchIntervalEndSec, utils.FunctionSignatures.Submit1, SubmissionContractAddress)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var commits = make([]Commit, 0, len(msgs))
	for _, msg := range msgs {
		commit, err := DecodeCommit(msg.Payload)
		if err != nil {
			logger.Info("error parsing commit, skipping: %s", err)
			continue
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func getReveals(db *gorm.DB, feeds []Feed, fromSec int64, toSec int64) ([]Reveal, error) {
	msgs, err := queryMessages(db, fromSec, toSec, utils.FunctionSignatures.Submit2, SubmissionContractAddress)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var reveals []Reveal
	for _, msg := range msgs {
		reveal, err := DecodeReveal(msg.Payload, feeds)
		if err != nil {
			logger.Info("error parsing reveal, skipping: %s", err)
			continue
		}

		reveals = append(reveals, reveal)
	}

	return reveals, nil
}

func queryMessages(db *gorm.DB, fromSec int64, toSec int64, signature string, contractAddress common.Address) ([]payload.Message, error) {
	txns, err := database.FetchTransactionsByAddressAndSelectorTimestamp(db, contractAddress.String(), signature, fromSec, toSec)
	if err != nil {
		return nil, errors.Errorf("error fetching txns from DB: %s", err)
	}

	var payloads []payload.Message
	for _, tx := range txns {
		payloadsByProtocol, err := payload.ExtractPayloads(&tx)
		if err != nil {
			logger.Info("error extracting payloads, skipping submission: %s", err)
			continue
		}

		scalingPayload, ok := payloadsByProtocol[FtsoScalingProtocolId]
		if !ok {
			continue
		}

		payloads = append(payloads, scalingPayload)
	}

	return payloads, nil
}

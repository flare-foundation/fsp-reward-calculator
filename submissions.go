package main

import (
	"flare-common/contracts/submission"
	"flare-common/database"
	"flare-common/payload"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	SubmissionContractAddress = common.HexToAddress("0x2cA6571Daa15ce734Bbd0Bf27D5C9D16787fc33f") // Coston
)

func getCommits(db *gorm.DB, searchIntervalStartSec int64, searchIntervalEndSec int64) error {
	msgs, err := queryMessages(db, searchIntervalStartSec, searchIntervalEndSec, utils.FunctionSignatures.Submit1, SubmissionContractAddress)
	if err != nil {
		return errors.Errorf("error querying messages: %s", err)
	}

	for _, msg := range msgs {
		m, err := DecodeCommit(msg.Payload)
		if err != nil {
			return errors.Errorf("error parsing commit: %s", err)

		}
		logger.Info("found commit: %v", m)
	}

	return nil

}

func getReveals(db *gorm.DB, feeds []Feed, searchIntervalStartSec int64, searchIntervalEndSec int64) error {
	msgs, err := queryMessages(db, searchIntervalStartSec, searchIntervalEndSec, utils.FunctionSignatures.Submit2, SubmissionContractAddress)
	if err != nil {
		return errors.Errorf("error querying messages: %s", err)
	}

	for _, msg := range msgs {
		m, err := DecodeReveal(msg.Payload, feeds)
		if err != nil {
			return errors.Errorf("error parsing reveal: %s", err)

		}

		var values []int32
		for _, v := range m.Values {
			values = append(values, v.Value)
		}

		logger.Info("Values for all feeds: %v", values)
		return nil
	}

	return nil

}

func parseCommit(msgs []payload.Message) {
	sig := []byte("submit1()")
	hash := crypto.Keccak256Hash(sig)

	logger.Info("test")

	logger.Info("sigs %v", submission.SubmissionMetaData.Sigs)

	logger.Info("Hash: %s", hash)
}

func queryMessages(db *gorm.DB, searchIntervalStartSec int64, searchIntervalEndSec int64, signature string, contractAddress common.Address) ([]payload.Message, error) {
	res, err := database.FetchTransactionsByAddressAndSelectorTimestamp(db, contractAddress.String(), signature, searchIntervalStartSec, searchIntervalEndSec)
	if err != nil {
		return nil, errors.Errorf("error fetching txns from DB: %s", err)
	}

	var msgs []payload.Message
	for _, tx := range res {
		payloadsByProtocol, err := payload.ExtractPayloads(&tx)
		if err != nil {
			return nil, errors.Errorf("error extracting payloads: %s", err)
		}

		scalingP, ok := payloadsByProtocol[100]
		if !ok {
			continue
		}

		msgs = append(msgs, scalingP)
	}

	return msgs, nil
}

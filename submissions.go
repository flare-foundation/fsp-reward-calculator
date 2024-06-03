package main

import (
	"flare-common/database"
	"flare-common/payload"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
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

// getCommits retrieves the last commit submission for each voter for each round in the given range
func getCommits(db *gorm.DB, fromRound uint64, toRound uint64) (map[uint64]map[VoterSubmit]*Commit, error) {
	fromSec := params.Coston.Epoch.VotingRoundStartSec(fromRound)
	toSec := params.Coston.Epoch.VotingRoundEndSec(toRound)

	msgs, err := queryMessages(db, fromSec, toSec, utils.FunctionSignatures.Submit1, SubmissionContractAddress)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var commitsByRound = map[uint64]map[VoterSubmit]*Commit{}
	for _, msg := range msgs {
		commit, err := DecodeCommit(msg.Payload)
		if err != nil {
			logger.Info("error parsing commit, skipping: %s", err)
			continue
		}
		if _, ok := commitsByRound[msg.VotingRound]; !ok {
			commitsByRound[msg.VotingRound] = map[VoterSubmit]*Commit{}
		}

		from := VoterSubmit(common.HexToAddress(msg.From))
		commitsByRound[msg.VotingRound][from] = commit
	}

	return commitsByRound, nil
}

// getReveals retrieves the last reveal submission for voter for each round in the given range
func getReveals(db *gorm.DB, fromRound uint64, toRound uint64) (map[uint64]map[VoterSubmit]*Reveal, error) {
	fromSec := params.Coston.Epoch.VotingRoundStartSec(fromRound)
	toSec := params.Coston.Epoch.VotingRoundEndSec(toRound)

	msgs, err := queryMessages(db, fromSec, toSec, utils.FunctionSignatures.Submit2, SubmissionContractAddress)
	if err != nil {
		return nil, errors.Errorf("error querying messages: %s", err)
	}

	var revealsByRound = map[uint64]map[VoterSubmit]*Reveal{}
	for _, msg := range msgs {
		reveal, err := DecodeReveal(msg.Payload)
		if err != nil {
			logger.Info("error parsing reveal, skipping: %s", err)
			continue
		}
		if _, ok := revealsByRound[msg.VotingRound]; !ok {
			revealsByRound[msg.VotingRound] = map[VoterSubmit]*Reveal{}
		}

		from := VoterSubmit(common.HexToAddress(msg.From))
		revealsByRound[msg.VotingRound][from] = reveal
	}

	return revealsByRound, nil
}

func queryMessages(db *gorm.DB, fromSec uint64, toSec uint64, signature [4]byte, contractAddress common.Address) ([]payload.Message, error) {
	txns, err := database.FetchTransactionsByAddressAndSelectorTimestamp(db, contractAddress, signature, int64(fromSec), int64(toSec))
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

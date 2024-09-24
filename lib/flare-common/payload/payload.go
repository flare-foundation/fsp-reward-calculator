package payload

import (
	"errors"
	"flare-common/database"
	"strconv"
	"strings"
)

type Round struct {
	Messages []Message
	Id       uint64
}

type Message struct {
	From             string
	Signature        string
	Protocol         uint64
	VotingRound      uint64
	Timestamp        uint64
	BlockNumber      uint64
	TransactionIndex uint64
	Length           uint64
	Payload          string
}

// ExtractPayloads extracts Payloads from transactions to submission contract to functions submit1, submit2, submitSignatures.
func ExtractPayloads(tx *database.Transaction) (map[uint64]Message, error) {

	messages := make(map[uint64]Message)

	data := strings.TrimPrefix(tx.Input, "0x")

	data = data[8:] // trim function selector

	for len(data) > 0 {

		if len(data) < 14 {
			return nil, errors.New("wrongly formatted tx input")
		}

		protocol, err := strconv.ParseUint(data[:2], 16, 64) // 1 byte protocol id

		if err != nil {
			return nil, errors.New("protocol id error")
		}
		votingRound, err := strconv.ParseUint(data[2:10], 16, 64) // 4 bytes votingRoundId
		if err != nil {
			return nil, errors.New("voting round error")
		}

		length, err := strconv.ParseUint(data[10:14], 16, 64) // 2 bytes length of payload in bytes
		if err != nil {
			return nil, errors.New("length error")
		}

		end := 14 + 2*length // 14 = 2 + 8 + 4
		payload := data[14:end]

		message := Message{tx.FromAddress, tx.FunctionSig, protocol, votingRound, tx.Timestamp, tx.BlockNumber, tx.TransactionIndex, length, payload}

		messages[protocol] = message

		data = data[end:] // trim the extracted payload

	}

	return messages, nil
}

func prependZerosToLength(hexString string, finalLength int) (string, error) {
	p := len(hexString) - finalLength

	if p < 0 {
		return hexString, errors.New("string too long")
	}

	prefix := strings.Repeat("0", p)

	return prefix + hexString, nil

}

func BuildMessage(protocol, votingRound uint64, payload string) (string, error) {

	if len(payload)%2 != 0 {
		return "", errors.New("uneven payload")
	}

	protocolStr, err := prependZerosToLength(strconv.FormatUint(protocol, 16), 2)

	if err != nil {
		return "", errors.New("invalid protocol")
	}

	votingRoundStr, err := prependZerosToLength(strconv.FormatUint(votingRound, 16), 8)

	if err != nil {
		return "", errors.New("invalid voting round")
	}

	lenStr, err := prependZerosToLength(strconv.FormatInt(int64(len(payload)), 16), 4)

	if err != nil {
		return "", errors.New("invalid payload length")
	}

	return protocolStr + votingRoundStr + lenStr + payload, nil
}

package main

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	FeedValueBytes = 4
	FeedIdBytes    = 21
	NoValue        = 0
)

var EmptyFeedValue = FeedValue{
	isEmpty: true,
	Value:   NoValue,
}

type FeedId [FeedIdBytes]byte

type Feed struct {
	Id       FeedId
	Decimals int8
}

func (f *Feed) String() string {
	return f.Id.String()
}

func (f *FeedId) String() string {
	return string(f[1:])
}

type Commit struct {
	Hash common.Hash
}

type Reveal struct {
	Random        common.Hash
	Feeds         []Feed
	Values        []FeedValue
	EncodedValues []byte
}

type FeedValue struct {
	isEmpty bool
	Value   int32
	//Decimals int
}

func DecodeCommit(message string) (Commit, error) {
	if len(message) != 64 {
		return Commit{}, errors.New("invalid message length")
	}
	hash := common.HexToHash(message)
	return Commit{
		Hash: hash,
	}, nil

}

func DecodeReveal(message string, feeds []Feed) (Reveal, error) {
	bytes, err := hex.DecodeString(message)
	if err != nil {
		return Reveal{}, errors.Wrap(err, "message is not a valid hex string")
	}

	// The message should be long enough to contain the random and at least one feed value
	if len(bytes) < (common.HashLength + FeedValueBytes) {
		return Reveal{}, errors.New("message too short")
	}

	random := common.BytesToHash(bytes[:common.HashLength])
	values, err := DecodeFeedValues(bytes[common.HashLength:], feeds)
	if err != nil {
		return Reveal{}, errors.Wrap(err, "failed to decode feed values")
	}

	return Reveal{
		Random:        random,
		Feeds:         feeds,
		Values:        values,
		EncodedValues: bytes,
	}, nil

}

func DecodeFeedValues(bytes []byte, feeds []Feed) ([]FeedValue, error) {
	if (len(bytes) % FeedValueBytes) != 0 {
		return nil, errors.New("invalid message length for feed values")
	}

	var feedValues []FeedValue
	for i := 0; i < len(bytes); i += FeedValueBytes {
		rawValue := binary.BigEndian.Uint32(bytes[i : i+FeedValueBytes])

		var feedValue FeedValue
		if rawValue == NoValue {
			feedValue = EmptyFeedValue
		} else {
			feedValue = FeedValue{
				isEmpty: false,
				Value:   int32(rawValue - 1<<31), // Values encoded in Excess-2^31
			}
		}
		feedValues = append(feedValues, feedValue)
	}

	// Fill in values for truncated empty feeds
	for i := len(feedValues); i < len(feeds); i++ {
		feedValues = append(feedValues, EmptyFeedValue)
	}

	return feedValues, nil
}

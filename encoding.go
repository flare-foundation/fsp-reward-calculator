package main

import (
	"encoding/binary"
	"encoding/hex"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const (
	FeedValueBytes          = 4
	FeedIdBytes             = 21
	NoValue                 = 0
	ProtocolMerkleRootBytes = 38
	SignatureBytes          = 65
)

var EmptyFeedValue = FeedValue{
	isEmpty: true,
	Value:   NoValue,
}

type FeedId [FeedIdBytes]byte

type Feed struct {
	Id                        FeedId
	Decimals                  int8
	MinRewardedTurnoutBIPS    uint16
	PrimaryBandRewardSharePPM uint32 // uint24 actual
	SecondaryBandWidthPPMs    uint32 // uint24 actual
}

func (f *Feed) String() string {
	return f.Id.String()
}

func (f *FeedId) String() string {
	return string(f[1:])
}

type ProtocolMerkleRoot struct {
	protocolId     int8
	round          types.RoundId
	isSecureRandom bool
	merkleRoot     common.Hash
}

type Commit struct {
	Hash common.Hash
}

type Reveal struct {
	Random        common.Hash
	EncodedValues []byte
}

type SignatureSubmission struct {
	Signature Signature
	Info      TxInfo
}

type TxInfo struct {
	TimestampSec uint64
	Reverted     bool
	From         common.Address
}

type Signature struct {
	signature  []byte
	merkleRoot ProtocolMerkleRoot
}

type FeedValue struct {
	isEmpty bool
	Value   int32
	//Decimals int
}

func (t *TxInfo) RoundId() types.RoundId {
	return params.Net.Epoch.VotingRoundForTimeSec(t.TimestampSec)
}
func (t *TxInfo) RoundOffset() uint64 {
	roundStartSec := params.Net.Epoch.VotingRoundStartSec(t.RoundId())
	return t.TimestampSec - roundStartSec
}

// Submit1, Submit2, Submit3
type Submission[T any] struct {
	info TxInfo
	item T
}

func DecodeCommit(message string) (*Commit, error) {
	if len(message) != 64 {
		return nil, errors.New("invalid message length")
	}
	hash := common.HexToHash(message)
	return &Commit{
		Hash: hash,
	}, nil

}

func DecodeReveal(message string) (*Reveal, error) {
	bytes, err := hex.DecodeString(message)
	if err != nil {
		return nil, errors.Wrap(err, "message is not a valid hex string")
	}

	// The message should be long enough to contain the random and at least one feed value
	if len(bytes) < (common.HashLength + FeedValueBytes) {
		return nil, errors.New("message too short")
	}

	random := common.BytesToHash(bytes[:common.HashLength])
	encodedFeeds := bytes[common.HashLength:]

	return &Reveal{
		Random:        random,
		EncodedValues: encodedFeeds,
	}, nil
}

func DecodeSignature(message string) (*Signature, error) {
	bytes, err := hex.DecodeString(message)
	if err != nil {
		return nil, errors.Wrapf(err, "message is not a valid hex string: %s", message)
	}

	if len(bytes) < 1+ProtocolMerkleRootBytes+SignatureBytes {
		return nil, errors.Errorf("Signature message too short: %s", message)
	}

	p := 1 // Type byte not used
	encodedMerkleRoot := bytes[p : p+ProtocolMerkleRootBytes]
	p += ProtocolMerkleRootBytes
	signature := bytes[p : p+SignatureBytes]
	p += SignatureBytes
	_ = bytes[p:] // Unsigned message not used

	merkleRoot, err := DecodeProtocolMerkleRoot(encodedMerkleRoot)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding protocol merkle merkleRoot")
	}

	return &Signature{
		signature:  signature,
		merkleRoot: merkleRoot,
	}, nil
}

func DecodeProtocolMerkleRoot(bytes []byte) (ProtocolMerkleRoot, error) {
	if len(bytes) != ProtocolMerkleRootBytes {
		return ProtocolMerkleRoot{}, errors.New("invalid message length for protocol merkle merkleRoot")
	}
	p := 0
	id := bytes[p]
	p++
	round := types.RoundId(binary.BigEndian.Uint32(bytes[p : p+4]))
	p += 4
	isSecureRandom := bytes[p] == 1
	p++
	merkleRoot := common.BytesToHash(bytes[p : p+common.HashLength])

	return ProtocolMerkleRoot{
		protocolId:     int8(id),
		round:          round,
		isSecureRandom: isSecureRandom,
		merkleRoot:     merkleRoot,
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

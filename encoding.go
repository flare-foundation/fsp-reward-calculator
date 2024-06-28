package main

import (
	"encoding/binary"
	"encoding/hex"
	"flare-common/policy"
	"ftsov2-rewarding/logger"
	"ftsov2-rewarding/params"
	"ftsov2-rewarding/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math"
	"math/big"
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
	hash           common.Hash
}

type Commit struct {
	Hash common.Hash
}

type Reveal struct {
	Random        common.Hash
	EncodedValues []byte
}

type SignatureSubmission struct {
	Signature *Signature
	Info      TxInfo
}

type Finalization struct {
	Policy     policy.SigningPolicy
	merkleRoot ProtocolMerkleRoot
	Signatures []ECDSASignature
	Info       TxInfo
}
type TxInfo struct {
	TimestampSec uint64
	Reverted     bool
	From         common.Address
}

type Signature struct {
	bytes      []byte
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
		bytes:      signature,
		merkleRoot: merkleRoot,
	}, nil
}

type ECDSASignature struct {
	V     byte
	R     [32]byte
	S     [32]byte
	index uint16
}

func DecodeFinalization(message string) (*Finalization, error) {
	bytes, err := hex.DecodeString(message)
	if err != nil {
		return nil, errors.Wrapf(err, "message is not a valid hex string: %s", message)
	}
	policy, p, err := DecodeSigningPolicy(bytes[:])
	if err != nil {
		return nil, errors.Wrap(err, "error decoding signing policy")
	}
	protocol := bytes[p] // We peek ahead to the protocol ID but don't process it
	if protocol == 0 {
		// This is a new signing policy message, ignore
		return nil, nil
	}
	merkleRoot, err := DecodeProtocolMerkleRoot(bytes[p : p+ProtocolMerkleRootBytes])
	if err != nil {
		return nil, errors.Wrap(err, "error decoding protocol merkle merkleRoot")
	}
	p += ProtocolMerkleRootBytes

	signatureCount := int(binary.BigEndian.Uint16(bytes[p : p+2]))
	p += 2
	signatures := make([]ECDSASignature, signatureCount)
	for i := 0; i < signatureCount; i++ {
		signatures[i].V = bytes[p]
		p++
		copy(signatures[i].R[:], bytes[p:p+32])
		p += 32
		copy(signatures[i].S[:], bytes[p:p+32])
		p += 32
		signatures[i].index = binary.BigEndian.Uint16(bytes[p : p+2])
		p += 2
	}

	return &Finalization{
		Policy:     *policy,
		merkleRoot: merkleRoot,
		Signatures: signatures,
	}, nil
}

func DecodeSigningPolicy(bytes []byte) (*policy.SigningPolicy, int, error) {
	p := 0
	size := int(DecodeUint32(bytes[p : p+2]))
	p += 2
	expectedLength := 43 + size*(common.AddressLength+2)
	if len(bytes) < expectedLength {
		return nil, p, errors.Errorf("message to short for decoding signing policy: expected >=%d, got %d", expectedLength, len(bytes))
	}

	epoch := DecodeUint32(bytes[p : p+3])
	p += 3
	startingRound := DecodeUint32(bytes[p : p+4])
	p += 4
	threshold := binary.BigEndian.Uint16(bytes[p : p+2])
	p += 2
	seed := common.BytesToHash(bytes[p : p+common.HashLength])
	p += common.HashLength

	signers := make([]common.Address, size)
	weights := make([]uint16, size)
	totalWeight := 0
	for i := 0; i < size; i++ {
		address := common.BytesToAddress(bytes[p : p+common.AddressLength])
		p += common.AddressLength
		weight := binary.BigEndian.Uint16(bytes[p : p+2])
		p += 2

		signers[i] = address
		weights[i] = weight
		totalWeight += int(weight)
	}

	if totalWeight > math.MaxUint16 {
		return nil, p, errors.New("total weight exceeds maximum uint16 value")
	}

	return &policy.SigningPolicy{
		RewardEpochId:      int64(epoch),
		StartVotingRoundId: startingRound,
		Threshold:          threshold,
		Seed:               new(big.Int).SetBytes(seed[:]),
		RawBytes:           bytes[:p],
		BlockTimestamp:     0,
		Voters:             policy.NewVoterSet(signers, weights),
	}, p, nil
}

func DecodeProtocolMerkleRoot(bytes []byte) (ProtocolMerkleRoot, error) {
	if len(bytes) != ProtocolMerkleRootBytes {
		return ProtocolMerkleRoot{}, errors.New("invalid message length for protocol merkle merkleRoot")
	}
	p := 0
	id := bytes[p]
	p++
	round := types.RoundId(DecodeUint32(bytes[p : p+4]))
	p += 4
	isSecureRandom := bytes[p] == 1
	p++
	merkleRoot := common.BytesToHash(bytes[p : p+common.HashLength])

	return ProtocolMerkleRoot{
		protocolId:     int8(id),
		round:          round,
		isSecureRandom: isSecureRandom,
		hash:           merkleRoot,
	}, nil
}

func DecodeFeedValues(bytes []byte, feeds []Feed) ([]FeedValue, error) {
	if (len(bytes) % FeedValueBytes) != 0 {
		return nil, errors.New("invalid message length for feed values")
	}

	var feedValues []FeedValue
	for i := 0; i < len(bytes); i += FeedValueBytes {
		rawValue := DecodeUint32(bytes[i : i+FeedValueBytes])

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

// Re-usable buffer for decoding to avoid allocations
var tmpUint32 = make([]byte, 4)

// DecodeUint32 decodes a big-endian uint32 from a variable length byte slice of up to 4 bytes.
func DecodeUint32(bytes []byte) uint32 {
	if len(bytes) > 4 {
		logger.Fatal("invalid length for decode int: %d", len(bytes))
	}

	pad := 4 - len(bytes)
	for i := 0; i < pad; i++ {
		tmpUint32[i] = 0
	}
	copy(tmpUint32[pad:], bytes)

	return binary.BigEndian.Uint32(tmpUint32[:])
}

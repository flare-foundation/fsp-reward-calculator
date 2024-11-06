package data

import (
	"encoding/binary"
	"encoding/hex"
	voters "fsp-rewards-calculator/lib"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	IsEmpty: true,
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

func (f *FeedId) Hex() string {
	return hex.EncodeToString(f[1:])
}

type ProtocolMerkleRoot struct {
	protocolId     int8
	round          ty.RoundId
	isSecureRandom bool
	hash           common.Hash
	rawEncoded     [ProtocolMerkleRootBytes]byte
}

type FeedValue struct {
	IsEmpty bool
	Value   int32
}

func DecodeCommit(bytes []byte) (*Commit, error) {
	if len(bytes) != common.HashLength {
		return nil, errors.New("invalid message length")
	}
	hash := common.BytesToHash(bytes)
	return &Commit{
		Hash: hash,
	}, nil

}

func DecodeReveal(bytes []byte, expectedFeeds int) (*Reveal, error) {
	// The message should be long enough to contain the random and at least one feed value
	if len(bytes) < (common.HashLength + FeedValueBytes) {
		return nil, errors.New("message too short")
	}

	random := common.BytesToHash(bytes[:common.HashLength])
	encodedFeeds := bytes[common.HashLength:]

	if (len(encodedFeeds) % FeedValueBytes) != 0 {
		return nil, errors.Errorf("invalid message length %d for feed values", len(encodedFeeds))
	}
	if (len(encodedFeeds) / FeedValueBytes) > expectedFeeds {
		return nil, errors.Errorf("encoded feed values paylaod %d exceeds expected number of feeds %d", len(encodedFeeds)/FeedValueBytes, expectedFeeds)
	}

	return &Reveal{
		Random:        random,
		EncodedValues: encodedFeeds,
	}, nil
}

func DecodeSignature(bytes []byte) (*Signature, error) {
	if len(bytes) < 1+ProtocolMerkleRootBytes+SignatureBytes {
		return nil, errors.Errorf("Signature message too short: %s", bytes)
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
	V           byte
	R           [32]byte
	S           [32]byte
	signerIndex uint16
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
	protocol := bytes[p] // Peek ahead to the protocol ID but don't process it
	if protocol == 0 {
		// This is a new signing policy message, ignore
		return nil, nil
	}

	merkleRootBytes := bytes[p : p+ProtocolMerkleRootBytes]
	merkleRootHash := accounts.TextHash(crypto.Keccak256(merkleRootBytes))
	merkleRoot, err := DecodeProtocolMerkleRoot(merkleRootBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding protocol merkle merkleRoot")
	}
	p += ProtocolMerkleRootBytes

	signatureCount := int(binary.BigEndian.Uint16(bytes[p : p+2]))
	p += 2

	if signatureCount > len(policy.Voters.VoterDataMap) {
		return nil, errors.Errorf("signature count %d exceeds number of signing policy voters %d", signatureCount, len(policy.Voters.VoterDataMap))
	}

	signatures := make([]ECDSASignature, signatureCount)
	signatureWeight := uint16(0)

	rawSig := make([]byte, 65)
	for i := 0; i < signatureCount; i++ {
		rawSig[64] = bytes[p] - 27
		p++
		copy(rawSig, bytes[p:p+64])
		p += 64
		index := binary.BigEndian.Uint16(bytes[p : p+2])
		p += 2

		if i > 0 && index <= signatures[i-1].signerIndex {
			return nil, errors.Errorf("signature index %d is not greater than previous index %d", signatures[i].signerIndex, signatures[i-1].signerIndex)
		}

		actualSigner, err := crypto.SigToPub(
			merkleRootHash,
			rawSig,
		)
		if err != nil {
			logger.Info("error recovering signer from signature: ", err)
			continue
		}
		expectedSigner := policy.Voters.Voters[index]

		if expectedSigner != crypto.PubkeyToAddress(*actualSigner) {
			logger.Info("signature at index %d does not match expected signer: %s", index, expectedSigner)
			continue
		}

		signatureWeight += policy.Voters.Weights[index]
	}

	if signatureWeight <= policy.Threshold {
		return nil, errors.Errorf("total signature weight %d is less than threshold %d", signatureWeight, policy.Threshold)
	}

	return &Finalization{
		Policy:     *policy,
		MerkleRoot: merkleRoot,
		Signatures: signatures,
	}, nil
}

func DecodeSigningPolicy(bytes []byte) (*voters.SigningPolicy, int, error) {
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

	return &voters.SigningPolicy{
		RewardEpochId:      int64(epoch),
		StartVotingRoundId: startingRound,
		Threshold:          threshold,
		Seed:               new(big.Int).SetBytes(seed[:]),
		RawBytes:           bytes[:p],
		BlockTimestamp:     0,
		Voters:             voters.NewVoterSet(signers, weights),
	}, p, nil
}

func DecodeProtocolMerkleRoot(bytes []byte) (ProtocolMerkleRoot, error) {
	if len(bytes) != ProtocolMerkleRootBytes {
		return ProtocolMerkleRoot{}, errors.New("invalid message length for protocol merkle merkleRoot")
	}
	p := 0
	id := bytes[p]
	p++
	round := ty.RoundId(DecodeUint32(bytes[p : p+4]))
	p += 4
	isSecureRandom := bytes[p] == 1
	p++
	merkleRoot := common.BytesToHash(bytes[p : p+common.HashLength])

	encoded := [ProtocolMerkleRootBytes]byte{}
	copy(encoded[:], bytes)

	return ProtocolMerkleRoot{
		protocolId:     int8(id),
		round:          round,
		isSecureRandom: isSecureRandom,
		hash:           merkleRoot,
		rawEncoded:     encoded,
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
				IsEmpty: false,
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

func (p *ProtocolMerkleRoot) EncodedHash() common.Hash {
	return common.BytesToHash(accounts.TextHash(crypto.Keccak256(p.rawEncoded[:])))
}

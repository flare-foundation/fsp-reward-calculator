package fsp

import (
	"encoding/hex"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/policy"
)

type RawSignature struct {
	Bytes   []byte
	Message []byte
}

type ECDSASignatureWithIndex struct {
	V           byte
	R           [32]byte
	S           [32]byte
	signerIndex uint16
}

// Bytes returns the byte representation of the signature (R + S + V), without the signer index.
func (s *ECDSASignatureWithIndex) Bytes() []byte {
	bytes := make([]byte, 65)
	copy(bytes[0:32], s.R[:])
	copy(bytes[32:64], s.S[:])
	bytes[64] = s.V
	return bytes
}

type ProtocolMerkleRoot struct {
	ProtocolId     int8
	Round          ty.RoundId
	IsSecureRandom bool
	Hash           common.Hash
	rawEncoded     [protocolMerkleRootBytes]byte
}

// EncodedHash returns the digest voters sign for the message under the given reward epoch's Relay.
func (p *ProtocolMerkleRoot) EncodedHash(epoch ty.RewardEpochId) common.Hash {
	return common.BytesToHash(messageDigest(p.rawEncoded[:], params.RelayV2Active(epoch)))
}

// messageDigest is the EIP-191 hash the Relay recovers signers from. Relay v2 binds the source
// chain into it, keccak256(chainId ‖ message), the Relay before it hashes the message alone.
func messageDigest(encoded []byte, relayV2 bool) []byte {
	if relayV2 {
		chainId := common.BigToHash(new(big.Int).SetUint64(params.Net.ChainId))
		return accounts.TextHash(crypto.Keccak256(chainId[:], encoded))
	}
	return accounts.TextHash(crypto.Keccak256(encoded))
}

type Finalization struct {
	Policy     policy.SigningPolicy
	MerkleRoot ProtocolMerkleRoot
	Signatures []ECDSASignatureWithIndex
	Info       TxInfo
}

const FeedIdBytes = 21

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

type TxInfo struct {
	TimestampSec uint64
	Reverted     bool
	From         common.Address
}

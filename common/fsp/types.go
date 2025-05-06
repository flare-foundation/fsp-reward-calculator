package fsp

import (
	"encoding/hex"
	"fsp-rewards-calculator/common/ty"
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

// Bytes returns the byte representation of the signature (R + S + V), without the signer index
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

func (p *ProtocolMerkleRoot) EncodedHash() common.Hash {
	return common.BytesToHash(accounts.TextHash(crypto.Keccak256(p.rawEncoded[:])))
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

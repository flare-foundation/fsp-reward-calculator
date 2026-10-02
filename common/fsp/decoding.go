package fsp

import (
	"encoding/binary"
	"encoding/hex"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/merkle"
	"github.com/flare-foundation/go-flare-common/pkg/policy"
	"github.com/pkg/errors"
)

const (
	protocolMerkleRootBytes = 38
	signatureBytes          = 65
	// v (1) + r (32) + s (32) + signer index (2).
	signatureWithIndexBytes = 67
)

// secp256k1HalfN is the largest s Relay v2 accepts (EIP-2 low-s).
var secp256k1HalfN = new(big.Int).Rsh(crypto.S256().Params().N, 1)

// DecodeSignatureType0 decodes a type 0 signature message. Type 0 signatures are deprecated, so we decode but
// skip the additional Merkle root payload.
func DecodeSignatureType0(bytes []byte) (*RawSignature, error) {
	if len(bytes) < 1+protocolMerkleRootBytes+signatureBytes {
		return nil, errors.Errorf("Type 0 signature message too short: %s", bytes)
	}

	if bytes[0] != 0 {
		logger.Fatal("invalid signature type: %d, expected 0", bytes[0])
	}
	p := 1
	encodedMerkleRoot := bytes[p : p+protocolMerkleRootBytes]
	p += protocolMerkleRootBytes
	signature := bytes[p : p+signatureBytes]
	p += signatureBytes

	_, err := DecodeProtocolMerkleRoot(encodedMerkleRoot)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding protocol merkle merkleRoot")
	}

	return &RawSignature{
		Bytes:   signature,
		Message: bytes[p:],
	}, nil
}

func DecodeSignatureType1(bytes []byte) (*RawSignature, error) {
	if len(bytes) < 1+signatureBytes {
		return nil, errors.Errorf("Type 1 signature message too short: %s", bytes)
	}

	if bytes[0] != 1 {
		return nil, errors.Errorf("invalid signature type: %d, expected 1", bytes[0])
	}
	p := 1
	signature := bytes[p : p+signatureBytes]
	p += signatureBytes
	unsignedMessage := bytes[p:]

	return &RawSignature{
		Bytes:   signature,
		Message: unsignedMessage,
	}, nil
}

// DecodeFinalization decodes relay() calldata and checks it the way the Relay would. relayV2 selects
// the checks of Relay v2: the source-bound digest, BadV/BadS rejecting the whole call, and the random
// number with its Merkle proof that must follow the signatures on the random number protocol.
func DecodeFinalization(hexMessage string, relayV2 bool) (*Finalization, error) {
	bytes, err := hex.DecodeString(hexMessage)
	if err != nil {
		return nil, errors.Wrapf(err, "message is not a valid hex string: %s", hexMessage)
	}
	signingPolicy, p, err := policy.FromRawBytes(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding signing policy")
	}
	if len(bytes) < p+protocolMerkleRootBytes+2 {
		return nil, errors.New("finalization too short")
	}
	protocol := bytes[p] // Peek ahead to the protocol ID but don't process it
	if protocol == 0 {
		// This is a new signing policy message, ignore
		return nil, nil
	}

	merkleRootBytes := bytes[p : p+protocolMerkleRootBytes]
	merkleRootHash := messageDigest(merkleRootBytes, relayV2)
	merkleRoot, err := DecodeProtocolMerkleRoot(merkleRootBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding protocol merkle merkleRoot")
	}
	p += protocolMerkleRootBytes

	signatureCount := int(binary.BigEndian.Uint16(bytes[p : p+2]))
	p += 2

	if !relayV2 && signatureCount > len(signingPolicy.Voters.VoterDataMap) {
		return nil, errors.Errorf("signature count %d exceeds number of signing policy voters %d", signatureCount, len(signingPolicy.Voters.VoterDataMap))
	}
	trailerStart := p + signatureCount*signatureWithIndexBytes
	if len(bytes) < trailerStart {
		return nil, errors.Errorf("signature list of %d signatures truncated", signatureCount)
	}

	var signatures []ECDSASignatureWithIndex
	signatureWeight := uint16(0)
	lastIndex := -1

	for i := 0; i < signatureCount; i++ {
		v := bytes[p] - 27
		p++
		r := bytes[p : p+32]
		p += 32
		s := bytes[p : p+32]
		p += 32
		index := binary.BigEndian.Uint16(bytes[p : p+2])
		p += 2

		signature := ECDSASignatureWithIndex{
			V:           v,
			R:           [32]byte(r),
			S:           [32]byte(s),
			signerIndex: index,
		}

		if int(index) <= lastIndex {
			return nil, errors.Errorf("signature index %d is not greater than previous index %d", index, lastIndex)
		}
		lastIndex = int(index)

		if relayV2 {
			if v > 1 {
				return nil, errors.Errorf("signature at index %d: v is neither 27 nor 28", index)
			}
			if new(big.Int).SetBytes(s).Cmp(secp256k1HalfN) > 0 {
				return nil, errors.Errorf("signature at index %d: s is above half the curve order", index)
			}
		}

		actualSigner, err := crypto.SigToPub(
			merkleRootHash,
			signature.Bytes(),
		)
		if err != nil {
			logger.Debug("error recovering signer from signature: ", err)
			continue
		}
		expectedSigner := signingPolicy.Voters.VoterAddress(int(index))

		if expectedSigner != crypto.PubkeyToAddress(*actualSigner) {
			logger.Debug("signature at index %d does not match expected signer: %s", index, expectedSigner)
			continue
		}

		signatureWeight += signingPolicy.Voters.VoterWeight(int(index))

		signatures = append(signatures, signature)

		// Relay v2 finalizes as soon as the threshold is crossed and never reads the rest.
		if relayV2 && signatureWeight > signingPolicy.Threshold {
			break
		}
	}

	if signatureWeight <= signingPolicy.Threshold {
		return nil, errors.Errorf("total signature weight %d is less than threshold %d", signatureWeight, signingPolicy.Threshold)
	}

	if relayV2 && protocol == params.Net.Ftso.ProtocolId {
		if err := verifyRandomProof(merkleRoot, bytes[trailerStart:]); err != nil {
			return nil, err
		}
	}

	return &Finalization{
		Policy:     *signingPolicy,
		MerkleRoot: merkleRoot,
		Signatures: signatures,
	}, nil
}

// verifyRandomProof checks the calldata Relay v2 requires after the signatures on the random number
// protocol: the random number as one word, then the Merkle proof nodes, folding to the signed root.
// The leaf is keccak256(abi.encode(votingRoundId, value, isSecure)) with round and flag from the message.
func verifyRandomProof(root ProtocolMerkleRoot, trailer []byte) error {
	if len(trailer) == 0 || len(trailer)%32 != 0 {
		return errors.New("missing or misaligned random number and proof")
	}
	var round, secure common.Hash
	binary.BigEndian.PutUint32(round[28:], uint32(root.Round))
	if root.IsSecureRandom {
		secure[31] = 1
	}
	hash := crypto.Keccak256Hash(round[:], trailer[:32], secure[:])
	for p := 32; p < len(trailer); p += 32 {
		hash = merkle.SortedHashPair(hash, common.BytesToHash(trailer[p:p+32]))
	}
	if hash != root.Hash {
		return errors.Errorf("random number does not prove against merkle root %s", root.Hash)
	}
	return nil
}

func DecodeProtocolMerkleRoot(bytes []byte) (ProtocolMerkleRoot, error) {
	if len(bytes) != protocolMerkleRootBytes {
		return ProtocolMerkleRoot{}, errors.New("invalid message length for protocol merkle merkleRoot")
	}
	p := 0
	id := bytes[p]
	p++
	round := ty.RoundId(DecodeUint32(bytes[p : p+4]))
	p += 4
	isSecureRandom := bytes[p] != 0 // the Relay reads any nonzero byte as true
	p++
	merkleRoot := common.BytesToHash(bytes[p : p+common.HashLength])

	encoded := [protocolMerkleRootBytes]byte{}
	copy(encoded[:], bytes)

	return ProtocolMerkleRoot{
		ProtocolId:     int8(id),
		Round:          round,
		IsSecureRandom: isSecureRandom,
		Hash:           merkleRoot,
		rawEncoded:     encoded,
	}, nil
}

// DecodeUint32 decodes a big-endian uint32 from a variable length byte slice of up to 4 bytes.
func DecodeUint32(bytes []byte) uint32 {
	if len(bytes) > 4 {
		logger.Fatal("invalid length for decode int: %d", len(bytes))
	}

	start := 4 - len(bytes)
	var tmp = make([]byte, 4)
	copy(tmp[start:], bytes)
	return binary.BigEndian.Uint32(tmp[:])
}

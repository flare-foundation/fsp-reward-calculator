package fsp

import (
	"encoding/binary"
	"encoding/hex"
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/policy"
	"github.com/pkg/errors"
)

const (
	protocolMerkleRootBytes = 38
	signatureBytes          = 65
)

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

func DecodeFinalization(hexMessage string) (*Finalization, error) {
	bytes, err := hex.DecodeString(hexMessage)
	if err != nil {
		return nil, errors.Wrapf(err, "message is not a valid hex string: %s", hexMessage)
	}
	signingPolicy, p, err := policy.FromRawBytes(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding signing policy")
	}
	protocol := bytes[p] // Peek ahead to the protocol ID but don't process it
	if protocol == 0 {
		// This is a new signing policy message, ignore
		return nil, nil
	}

	merkleRootBytes := bytes[p : p+protocolMerkleRootBytes]
	merkleRootHash := accounts.TextHash(crypto.Keccak256(merkleRootBytes))
	merkleRoot, err := DecodeProtocolMerkleRoot(merkleRootBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding protocol merkle merkleRoot")
	}
	p += protocolMerkleRootBytes

	signatureCount := int(binary.BigEndian.Uint16(bytes[p : p+2]))
	p += 2

	if signatureCount > len(signingPolicy.Voters.VoterDataMap) {
		return nil, errors.Errorf("signature count %d exceeds number of signing policy voters %d", signatureCount, len(signingPolicy.Voters.VoterDataMap))
	}

	var signatures []ECDSASignatureWithIndex
	signatureWeight := uint16(0)

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

		if i > 0 && index <= signatures[i-1].signerIndex {
			return nil, errors.Errorf("signature index %d is not greater than previous index %d", signatures[i].signerIndex, signatures[i-1].signerIndex)
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
	}

	if signatureWeight <= signingPolicy.Threshold {
		return nil, errors.Errorf("total signature weight %d is less than threshold %d", signatureWeight, signingPolicy.Threshold)
	}

	return &Finalization{
		Policy:     *signingPolicy,
		MerkleRoot: merkleRoot,
		Signatures: signatures,
	}, nil
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
	isSecureRandom := bytes[p] == 1
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

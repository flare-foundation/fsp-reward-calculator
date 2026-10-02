package fsp

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"os"
	"testing"

	"fsp-rewards-calculator/common/params"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/policy"
)

// Real relay() calldata (selector stripped) sent to the Songbird Relay v2 in reward epoch 437:
// txs 0x8ef13657… (FTSO, with the random number and its proof) and 0xb2ce9c55… (FDC).
func loadRelayV2Fixture(t *testing.T, name string) []byte {
	t.Helper()
	params.InitNetwork("songbird")
	raw, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := hex.DecodeString(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

// signaturesStart returns the offset of the first signature and the offset right after the list.
func signaturesStart(t *testing.T, bytes []byte) (int, int) {
	t.Helper()
	_, p, err := policy.FromRawBytes(bytes)
	if err != nil {
		t.Fatal(err)
	}
	p += protocolMerkleRootBytes
	count := int(binary.BigEndian.Uint16(bytes[p : p+2]))
	return p + 2, p + 2 + count*signatureWithIndexBytes
}

func TestDecodeFinalizationRelayV2(t *testing.T) {
	for name, protocol := range map[string]byte{
		"songbird-relay-v2-ftso.hex": 100,
		"songbird-relay-v2-fdc.hex":  200,
	} {
		bytes := loadRelayV2Fixture(t, name)
		finalization, err := DecodeFinalization(hex.EncodeToString(bytes), true)
		if err != nil {
			t.Fatalf("%s: %s", name, err)
		}
		if byte(finalization.MerkleRoot.ProtocolId) != protocol {
			t.Errorf("%s: protocol %d, want %d", name, byte(finalization.MerkleRoot.ProtocolId), protocol)
		}
		if _, err := DecodeFinalization(hex.EncodeToString(bytes), false); err == nil {
			t.Errorf("%s: accepted under the digest of the Relay before v2", name)
		}
	}
}

func TestDecodeFinalizationRelayV2RandomProof(t *testing.T) {
	bytes := loadRelayV2Fixture(t, "songbird-relay-v2-ftso.hex")
	_, trailerStart := signaturesStart(t, bytes)
	if len(bytes) == trailerStart {
		t.Fatal("fixture carries no random number")
	}

	if _, err := DecodeFinalization(hex.EncodeToString(bytes[:trailerStart]), true); err == nil {
		t.Error("accepted without the random number")
	}
	if _, err := DecodeFinalization(hex.EncodeToString(bytes[:len(bytes)-1]), true); err == nil {
		t.Error("accepted a misaligned proof")
	}
	tampered := append([]byte{}, bytes...)
	tampered[trailerStart+31] ^= 1
	if _, err := DecodeFinalization(hex.EncodeToString(tampered), true); err == nil {
		t.Error("accepted a random number that does not prove against the root")
	}
}

func TestDecodeFinalizationRelayV2RejectsHighS(t *testing.T) {
	bytes := loadRelayV2Fixture(t, "songbird-relay-v2-fdc.hex")
	first, _ := signaturesStart(t, bytes)

	// s -> n - s with v flipped recovers the same signer, but Relay v2 reverts the whole call on it.
	malleable := append([]byte{}, bytes...)
	s := new(big.Int).SetBytes(malleable[first+33 : first+65])
	new(big.Int).Sub(crypto.S256().Params().N, s).FillBytes(malleable[first+33 : first+65])
	malleable[first] ^= 27 ^ 28

	if _, err := DecodeFinalization(hex.EncodeToString(malleable), true); err == nil {
		t.Error("accepted a high-s signature")
	}
}

package fcc

import (
	"encoding/json"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// rawLog is an FCC fee event log as it was emitted on chain, with the block timestamp the indexer
// records alongside it.
type rawLog struct {
	Kind            string   `json:"kind"`
	Address         string   `json:"address"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	BlockNumber     uint64   `json:"blockNumber"`
	LogIndex        uint     `json:"logIndex"`
	TransactionHash string   `json:"transactionHash"`
	TimestampSec    uint64   `json:"timestampSec"`
}

func (l rawLog) chainLog() types.Log {
	topics := make([]common.Hash, len(l.Topics))
	for i, topic := range l.Topics {
		topics[i] = common.HexToHash(topic)
	}
	return types.Log{
		Address:     common.HexToAddress(l.Address),
		Topics:      topics,
		Data:        common.FromHex(l.Data),
		BlockNumber: l.BlockNumber,
		Index:       l.LogIndex,
		TxHash:      common.HexToHash(l.TransactionHash),
	}
}

// TestAttributeFeesSongbird419 decodes the FCC fee events Songbird actually emitted inside reward epoch
// 419 - the first epoch accounted for on any production network - and reproduces the fees the reference
// calculator accounted for there: 3 SGB of TEE instruction fees plus 3 SGB of FDC2 request fees, spread
// over two voting rounds. The 6 SGB total is the direct claim in the published epoch 419 distribution.
//
// The logs in testdata are verbatim from the chain, so this covers the event ABIs and decoding as well as
// the attribution: a wrong ABI or a mis-set topic would fail here rather than silently matching no events.
func TestAttributeFeesSongbird419(t *testing.T) {
	params.InitNetwork("songbird")

	raw, err := os.ReadFile("testdata/songbird-419-fcc-logs.json")
	if err != nil {
		t.Fatalf("reading fixture: %s", err)
	}
	var logs []rawLog
	if err := json.Unmarshal(raw, &logs); err != nil {
		t.Fatalf("parsing fixture: %s", err)
	}

	var teeInstructions []teeInstruction
	var fdc2Requests []fdc2Request
	for _, log := range logs {
		switch log.Kind {
		case "tee":
			instruction, err := parseTeeInstructionsSent(log.chainLog(), log.TimestampSec)
			if err != nil {
				t.Fatalf("decoding TeeInstructionsSent of %s: %s", log.TransactionHash, err)
			}
			teeInstructions = append(teeInstructions, instruction)
		case "fdc2":
			request, err := parseFdc2AttestationRequested(log.chainLog(), log.TimestampSec)
			if err != nil {
				t.Fatalf("decoding AttestationRequested of %s: %s", log.TransactionHash, err)
			}
			fdc2Requests = append(fdc2Requests, request)
		default:
			t.Fatalf("unknown fixture event kind %q", log.Kind)
		}
	}
	if len(teeInstructions) != 3 || len(fdc2Requests) != 1 {
		t.Fatalf("fixture: got %d TEE and %d FDC2 events, want 3 and 1", len(teeInstructions), len(fdc2Requests))
	}

	fees, err := attributeFees(419, 1407840, 1411199, teeInstructions, fdc2Requests)
	if err != nil {
		t.Fatalf("attributing fees: %s", err)
	}

	oneToken := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	if want := new(big.Int).Mul(big.NewInt(6), oneToken); fees.Total.Cmp(want) != 0 {
		t.Errorf("total fees: got %s, want %s", fees.Total, want)
	}
	// Round 1410950 carries a lone 1 SGB TEE dispatch. Round 1410958 carries 2 SGB of TEE dispatches plus
	// the 3 SGB FDC2 request fee: that request and the dispatch it paid for are emitted in the same
	// transaction, share an instruction id and so land in the same round. TEE and FDC2 fees are 3 SGB each.
	for round, wantTokens := range map[ty.RoundId]int64{1410950: 1, 1410958: 5} {
		got, hasFees := fees.ByRound[round]
		if want := new(big.Int).Mul(big.NewInt(wantTokens), oneToken); !hasFees || got.Cmp(want) != 0 {
			t.Errorf("fees of round %d: got %s, want %s", round, got, want)
		}
	}
	if got := len(fees.ByRound); got != 2 {
		t.Errorf("rounds with fees: got %d, want 2", got)
	}
}

package fsp

import (
	stderrors "errors"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/database"
)

func TestParseEventLogsStrictFailsOnConversionError(t *testing.T) {
	logs := []database.Log{{
		Data:        "not-hex",
		BlockNumber: 12,
		LogIndex:    3,
	}}

	_, err := parseEventLogs(logs, func(types.Log, uint64) (struct{}, error) {
		return struct{}{}, nil
	}, true)
	if err == nil {
		t.Fatal("expected strict event conversion to fail")
	}
	if !strings.Contains(err.Error(), "block 12 log index 3") {
		t.Errorf("error does not identify the skipped log: %s", err)
	}
}

func TestParseEventLogsStrictFailsOnDecodeError(t *testing.T) {
	logs := []database.Log{{
		Data:        "",
		BlockNumber: 21,
		LogIndex:    5,
	}}

	_, err := parseEventLogs(logs, func(types.Log, uint64) (struct{}, error) {
		return struct{}{}, stderrors.New("decode failed")
	}, true)
	if err == nil {
		t.Fatal("expected strict event decoding to fail")
	}
	if !strings.Contains(err.Error(), "block 21 log index 5") {
		t.Errorf("error does not identify the skipped log: %s", err)
	}
}

package fdc

import (
	"fsp-rewards-calculator/common/fsp"
	"math/big"
)

const (
	ProtocolId = 200
)

type AttestationRequest struct {
	Data      []byte
	MergedFee *big.Int // Combined fee for all duplicates
}

type SignatureSubmission struct {
	Signature *fsp.RawSignature
	Info      fsp.TxInfo
}

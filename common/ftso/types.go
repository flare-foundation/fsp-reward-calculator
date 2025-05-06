package ftso

import (
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/ty"
	"github.com/ethereum/go-ethereum/common"
)

const (
	ProtocolId     = 100
	feedValueBytes = 4
	noValue        = 0
)

var EmptyFeedValue = FeedValue{
	IsEmpty: true,
	Value:   noValue,
}

type FeedValue struct {
	IsEmpty bool
	Value   int32
}

type Commit struct {
	Hash common.Hash
}

type Reveal struct {
	Random        common.Hash
	EncodedValues []byte
}

type SignatureSubmission struct {
	Signature *fsp.RawSignature
	Info      fsp.TxInfo
}

type FUpdate struct {
	Feeds      *FastUpdateFeed
	Submitters []ty.VoterSigning
}

type RoundReveals struct {
	Reveals             map[ty.VoterSubmit]*Reveal
	RegisteredOffenders []ty.VoterSubmit
	AllOffenders        []ty.VoterSubmit
}

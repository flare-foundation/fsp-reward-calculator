package fsp

import (
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"

	"github.com/ethereum/go-ethereum/common"
)

type SignerMap map[ty.RoundId]map[common.Hash]map[ty.VoterSigning]SigInfo

type SigInfo struct {
	Signer          ty.VoterSigning
	Timestamp       uint64
	UnsignedMessage []byte // TODO: Only use by FDC
}

func GetFinalizationsByRound(fnz []*Finalization) map[ty.RoundId][]*Finalization {
	finalizationsByRound := make(map[ty.RoundId][]*Finalization)

loop:
	for _, f := range fnz {
		round := f.MerkleRoot.Round

		if _, ok := finalizationsByRound[round]; !ok {
			finalizationsByRound[round] = []*Finalization{}
		} else {
			for _, rf := range finalizationsByRound[round] {
				if rf.Info.From == f.Info.From {
					logger.Debug("Finalization for round %d from %s already seen, skipping", round, f.Info.From)
					continue loop
				}
			}
		}
		finalizationsByRound[round] = append(finalizationsByRound[round], f)
	}

	return finalizationsByRound
}

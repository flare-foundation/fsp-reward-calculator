package rewards

import (
	"fsp-rewards-calculator/data"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/ty"
	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/payload"
	"gorm.io/gorm"
	"slices"
)

func getFdcRewards(db *gorm.DB, epochs data.RewardEpochs, windowEnd ty.RoundId, submit1 []payload.Message, submit2 []payload.Message, submitSignatures []payload.Message, finalizations []*data.Finalization) ([]ty.RewardClaim, error) {
	re := epochs.Current
	bitVotesByRound := data.ExtractBitVotes(submit2)

	finalizationsByRound := data.GetFinalizationsByRound(finalizations)

	consensusHashByRound := map[ty.RoundId]common.Hash{}
	for round, fs := range finalizationsByRound {
		firstSuccessfulIndex := slices.IndexFunc(fs, func(f *data.Finalization) bool {
			return f.Info.Reverted == false
		})
		if firstSuccessfulIndex < 0 {
			continue
		}
		consensusHashByRound[round] = fs[firstSuccessfulIndex].MerkleRoot.EncodedHash()
	}

	signersByRound := data.GetFdcSignersByRound(submitSignatures, consensusHashByRound, re)

	logger.Info("Signers fetched %d", len(signersByRound), len(consensusHashByRound), len(finalizationsByRound), len(bitVotesByRound))

	return nil, nil
}

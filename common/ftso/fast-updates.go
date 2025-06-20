package ftso

import (
	common2 "fsp-rewards-calculator/common"
	"fsp-rewards-calculator/common/fsp"
	"fsp-rewards-calculator/common/params"
	"fsp-rewards-calculator/common/ty"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/fupdater"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"
)

type FastUpdateFeed struct {
	Values   []*big.Int
	Decimals []int8
}

func GetFUpdateFeeds(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[ty.RoundId]*FastUpdateFeed, error) {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound + 1)
	toSec := params.Net.Epoch.VotingRoundStartSec(toRound.Add(2)) // extra round for buffer

	instance, _ := fupdater.NewFUpdater(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*fupdater.FUpdaterFastUpdateFeeds, error) {
		return instance.ParseFastUpdateFeeds(log)
	}

	events, err := fsp.QueryEvents(
		db,
		fromSec,
		toSec,
		params.Net.Contracts.FastUpdater,
		common2.EventTopic0.FastUpdateFeeds,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	var byRound = map[ty.RoundId]*FastUpdateFeed{}

	for _, event := range events {
		round := ty.RoundId(event.VotingEpochId.Uint64())
		if round < fromRound || round > toRound {
			continue
		}

		byRound[round] = &FastUpdateFeed{
			Values:   event.Feeds,
			Decimals: event.Decimals,
		}
	}

	return byRound, nil
}

func GetFUpdateSubmits(db *gorm.DB, fromRound ty.RoundId, toRound ty.RoundId) (map[ty.RoundId][]ty.VoterSigning, error) {
	fromSec := params.Net.Epoch.VotingRoundStartSec(fromRound)
	toSec := params.Net.Epoch.VotingRoundStartSec(toRound.Add(2)) // Add extra round as buffer

	instance, _ := fupdater.NewFUpdater(common.Address{}, nil)
	parse := func(log types.Log, _ uint64) (*fupdater.FUpdaterFastUpdateFeedsSubmitted, error) {
		return instance.ParseFastUpdateFeedsSubmitted(log)
	}

	events, err := fsp.QueryEvents(
		db,
		fromSec,
		toSec,
		params.Net.Contracts.FastUpdater,
		common2.EventTopic0.FastUpdateFeedsSubmitted,
		parse,
	)
	if err != nil {
		return nil, errors.Errorf("err fetching events: %s", err)
	}

	var byRound = map[ty.RoundId][]ty.VoterSigning{}

	for _, event := range events {
		round := ty.RoundId(event.VotingRoundId)
		if round < fromRound || round > toRound {
			continue
		}

		if _, ok := byRound[round]; !ok {
			byRound[round] = []ty.VoterSigning{}
		}

		byRound[round] = append(byRound[round], ty.VoterSigning(event.SigningPolicyAddress))
	}

	return byRound, nil
}

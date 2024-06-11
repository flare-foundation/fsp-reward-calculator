package main

import (
	"bytes"
	"encoding/binary"
	"flare-common/contracts/offers"
	"ftsov2-rewarding/logger"
	"golang.org/x/exp/maps"
	"math/big"
	"slices"
	"sort"
)

func GetOrderedFeeds(of RewardOffers) []Feed {
	feeds := getInflationFeeds(of.inflation)

	communityFeeds := getCommunityFeeds(of.community)
	for i := range communityFeeds {
		found := slices.IndexFunc(feeds, func(j Feed) bool {
			return communityFeeds[i].Id == j.Id
		})
		if found < 0 {
			feeds = append(feeds, communityFeeds[i])
		}
	}
	return feeds
}

// getCommunityFeeds returns a list of feeds first ordered by total amount of rewards offered, then feed id.
func getCommunityFeeds(offers []*offers.OffersRewardsOffered) []Feed {
	feedById := map[FeedId]Feed{}
	amountPerFeed := map[FeedId]*big.Int{}

	for _, offer := range offers {
		id := FeedId(offer.FeedId)
		if _, ok := feedById[id]; !ok {
			feedById[offer.FeedId] = Feed{
				Id:                        offer.FeedId,
				Decimals:                  offer.Decimals,
				MinRewardedTurnoutBIPS:    offer.MinRewardedTurnoutBIPS,
				PrimaryBandRewardSharePPM: uint32(offer.PrimaryBandRewardSharePPM.Uint64()),
				SecondaryBandWidthPPMs:    uint32(offer.SecondaryBandWidthPPM.Uint64()),
			}
		} else {
			logger.Info("More than one offer contains feed %s, only the first one will be used for configuration values", id.String())
		}

		if value, ok := amountPerFeed[offer.FeedId]; !ok {
			amountPerFeed[offer.FeedId] = big.NewInt(0)
		} else {
			amountPerFeed[offer.FeedId].Add(value, offer.Amount)
		}
	}

	feeds := maps.Values(feedById)
	sort.Slice(feeds, func(i, j int) bool {
		valueI := amountPerFeed[feeds[i].Id]
		valueJ := amountPerFeed[feeds[j].Id]
		res := valueI.Cmp(valueJ)

		if res == 0 {
			return bytes.Compare(feeds[i].Id[:], feeds[j].Id[:]) < 0
		} else {
			return res < 0
		}
	})
	return feeds
}

func getInflationFeeds(offers []*offers.OffersInflationRewardsOffered) []Feed {
	feedById := map[FeedId]Feed{}

	for _, offer := range offers {
		feedCount := len(offer.FeedIds) / FeedIdBytes
		for j := 0; j < feedCount; j++ {
			id := FeedId(offer.FeedIds[j*FeedIdBytes : (j+1)*FeedIdBytes])
			if _, ok := feedById[id]; ok {
				logger.Info("More than one inflation offer contains feed %s, only the last one will be used for configuration values", id.String())
			}
			feedById[id] = Feed{
				Id:                        id,
				Decimals:                  int8(offer.Decimals[j]),
				MinRewardedTurnoutBIPS:    offer.MinRewardedTurnoutBIPS,
				PrimaryBandRewardSharePPM: uint32(offer.PrimaryBandRewardSharePPM.Uint64()),
				SecondaryBandWidthPPMs:    binary.BigEndian.Uint32(offer.SecondaryBandWidthPPMs[j*3 : (j+1)*3]),
			}
		}
	}

	feeds := maps.Values(feedById)
	sort.Slice(feeds, func(i, j int) bool {
		return bytes.Compare(feeds[i].Id[:], feeds[j].Id[:]) < 0
	})
	return feeds
}

package main

import (
	"bytes"
	"flare-common/contracts/offers"
	"ftsov2-rewarding/logger"
	"golang.org/x/exp/maps"
	"math/big"
	"sort"
)

func GetOrderedFeeds(of RewardOffers) []Feed {
	inflationFeeds := getInflationFeeds(of.inflation)
	communityFeeds := getCommunityFeeds(of.community)
	return append(inflationFeeds, communityFeeds...)
}

func getCommunityFeeds(offers []*offers.OffersRewardsOffered) []Feed {
	var communityFeeds []Feed
	amountPerFeed := make(map[FeedId]*big.Int)

	for _, offer := range offers {
		communityFeeds = append(communityFeeds, Feed{
			Id:       offer.FeedId,
			Decimals: offer.Decimals,
		})

		if value, ok := amountPerFeed[offer.FeedId]; !ok {
			amountPerFeed[offer.FeedId] = big.NewInt(0)
		} else {
			amountPerFeed[offer.FeedId].Add(value, offer.Amount)
		}
	}

	sort.Slice(communityFeeds, func(i, j int) bool {
		valueI := amountPerFeed[communityFeeds[i].Id]
		valueJ := amountPerFeed[communityFeeds[j].Id]
		res := valueI.Cmp(valueJ)

		if res == 0 {
			return bytes.Compare(communityFeeds[i].Id[:], communityFeeds[j].Id[:]) < 0
		} else {
			return res < 0
		}
	})
	return communityFeeds
}

func getInflationFeeds(offers []*offers.OffersInflationRewardsOffered) []Feed {
	feedById := make(map[FeedId]Feed)

	for _, offer := range offers {
		feedCount := len(offer.FeedIds) / FeedIdBytes
		for j := 0; j < feedCount; j++ {
			id := FeedId(offer.FeedIds[j*FeedIdBytes : (j+1)*FeedIdBytes])
			if _, ok := feedById[id]; !ok {
				feedById[id] = Feed{
					Id:       id,
					Decimals: int8(offer.Decimals[j]),
				}
			} else {
				logger.Fatal("More than one inflation offer contains feed %s", id.String())
			}
		}
	}

	feeds := maps.Values(feedById)
	sort.Slice(feeds, func(i, j int) bool {
		return bytes.Compare(feeds[i].Id[:], feeds[j].Id[:]) < 0
	})
	return feeds
}

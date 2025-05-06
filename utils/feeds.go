package utils

import (
	"fsp-rewards-calculator/common/fsp"
)

var renames = map[string]string{
	"DAI/USD": "USDS/USD",
	"FTM/USD": "S/USD",
}

func toFeedIds() map[fsp.FeedId]fsp.FeedId {
	oldToNew := make(map[fsp.FeedId]fsp.FeedId)

	for oldName, newName := range renames {
		oldFeedId := fsp.FeedId{1}
		copy(oldFeedId[1:], oldName)
		newFeedId := fsp.FeedId{1}
		copy(newFeedId[1:], newName)
		oldToNew[oldFeedId] = newFeedId
	}
	return oldToNew
}

var OldToNewFeed = toFeedIds()

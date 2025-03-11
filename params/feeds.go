package params

import (
	"fsp-rewards-calculator/ty"
)

var renames = map[string]string{
	"DAI/USD": "USDS/USD",
	"FTM/USD": "S/USD",
}

func toFeedIds() map[ty.FeedId]ty.FeedId {
	oldToNew := make(map[ty.FeedId]ty.FeedId)

	for oldName, newName := range renames {
		oldFeedId := ty.FeedId{1}
		copy(oldFeedId[1:], oldName)
		newFeedId := ty.FeedId{1}
		copy(newFeedId[1:], newName)
		oldToNew[oldFeedId] = newFeedId
	}
	return oldToNew
}

var OldToNewFeed = toFeedIds()

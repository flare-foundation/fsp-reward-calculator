package data

import (
	"encoding/json"
	"fmt"
	"fsp-rewards-calculator/logger"
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"fsp-rewards-calculator/utils"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type PrintReveals struct {
	Submitted string
	Random    string
}

type RoundPrintData struct {
	Reveals      []PrintReveals
	Offenders    []string
	AllOffenders []string

	Medians      []Result
	Random       string
	SecureRandom bool

	NextRandom string

	MedianFeed  string
	FeedDecoded string
}

func PrintRoundData(results RoundResult, reveals RoundReveals, feed *ty.Feed, selection *big.Int, epoch ty.EpochId, round ty.RoundId) {
	var roundData RoundPrintData

	for voter, reveal := range reveals.Reveals {
		roundData.Reveals = append(roundData.Reveals, PrintReveals{
			Submitted: common.Address(voter).String(),
			Random:    reveal.Random.String(),
		})
	}

	for _, offender := range reveals.RegisteredOffenders {
		roundData.Offenders = append(roundData.Offenders, common.Address(offender).String())
	}
	for _, offender := range reveals.AllOffenders {
		roundData.AllOffenders = append(roundData.AllOffenders, common.Address(offender).String())
	}

	roundData.Random = results.Random.Value.String()
	roundData.SecureRandom = results.Random.IsSecure
	roundData.NextRandom = selection.String()

	if feed != nil {
		roundData.MedianFeed = feed.Id.Hex()
		roundData.FeedDecoded = feed.String()
	}

	for _, median := range results.Median {
		if median == nil {
			continue
		}
		medianCpy := *median
		medianCpy.InputValues = nil
		roundData.Medians = append(roundData.Medians, medianCpy)
	}

	jsonData, err := json.MarshalIndent(roundData, "", "    ")
	if err != nil {
		logger.Error("Error serializing to JSON:", err)
		return
	}
	filePath := fmt.Sprintf("results/%s/%d/rounds/%d/data.json", params.Net.Name, epoch, round)
	utils.WriteToFile(jsonData, filePath)
}

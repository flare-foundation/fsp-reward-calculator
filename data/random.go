package data

import (
	"fsp-rewards-calculator/params"
	"fsp-rewards-calculator/ty"
	"math/big"
)

var randomMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)

type RandomResult struct {
	Round    ty.RoundId
	Value    *big.Int
	IsSecure bool
}

func CalculateRandom(round ty.RoundId, reveals map[ty.RoundId]RoundReveals, eligibleReveals map[ty.VoterSubmit]*Reveal) RandomResult {
	benchingWindowOffenders := map[ty.VoterSubmit]bool{}
	for i := ty.RoundId(uint64(round) - params.Net.Ftso.RandomGenerationBenchingWindow); i < round; i++ {
		for j := range reveals[i].AllOffenders {
			benchingWindowOffenders[reveals[i].AllOffenders[j]] = true
		}
	}

	nonBenchedOffenders := 0
	for k := range reveals[round].RegisteredOffenders {
		currentOffender := reveals[round].RegisteredOffenders[k]
		if _, ok := benchingWindowOffenders[currentOffender]; !ok {
			nonBenchedOffenders++
		}
	}
	validCount := 0
	random := big.NewInt(0)
	for voter, reveal := range eligibleReveals {
		if _, ok := benchingWindowOffenders[voter]; !ok {
			random.Add(random, new(big.Int).SetBytes(reveal.Random[:]))
			validCount++
		}
	}
	random.Mod(random, randomMod)

	res := RandomResult{
		Round:    round,
		Value:    random,
		IsSecure: nonBenchedOffenders == 0 && validCount >= params.Net.Ftso.NonBenchedRandomVotersMinCount,
	}
	return res
}

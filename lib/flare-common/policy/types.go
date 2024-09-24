package policy

import (
	relayContract "flare-common/contracts/relay"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type VoterData struct {
	Index  int
	Weight uint16
}

type VoterSet struct {
	Voters      []common.Address
	Weights     []uint16
	TotalWeight uint16
	Thresholds  []uint16

	VoterDataMap map[common.Address]VoterData
}

func NewVoterSet(voters []common.Address, weights []uint16) *VoterSet {
	vs := VoterSet{
		Voters:     voters,
		Weights:    weights,
		Thresholds: make([]uint16, len(weights)),
	}
	// sum does not exceed uint16, guaranteed by the smart contract
	for i, w := range weights {
		vs.Thresholds[i] = vs.TotalWeight
		vs.TotalWeight += w
	}

	vMap := make(map[common.Address]VoterData)
	for i, voter := range vs.Voters {
		if _, ok := vMap[voter]; !ok {
			vMap[voter] = VoterData{
				Index:  i,
				Weight: vs.Weights[i],
			}
		}
	}
	vs.VoterDataMap = vMap
	return &vs
}

type SigningPolicy struct {
	RewardEpochId      int64
	StartVotingRoundId uint32
	Threshold          uint16
	Seed               *big.Int
	RawBytes           []byte
	BlockTimestamp     uint64

	// The set of all Voters and their weights
	Voters *VoterSet
}

func NewSigningPolicy(r *relayContract.RelaySigningPolicyInitialized) *SigningPolicy {
	return &SigningPolicy{
		RewardEpochId:      r.RewardEpochId.Int64(),
		StartVotingRoundId: r.StartVotingRoundId,
		Threshold:          r.Threshold,
		Seed:               r.Seed,
		RawBytes:           r.SigningPolicyBytes,
		BlockTimestamp:     r.Timestamp,

		Voters: NewVoterSet(r.Voters, r.Weights),
	}
}

type SigningPolicyStorage struct {

	// sorted list of signing policies, sorted by rewardEpochId (and also by startVotingRoundId)
	spList []*SigningPolicy

	// mutex
	sync.Mutex
}

func NewSigningPolicyStorage() *SigningPolicyStorage {
	return &SigningPolicyStorage{
		spList: make([]*SigningPolicy, 0, 10),
	}
}

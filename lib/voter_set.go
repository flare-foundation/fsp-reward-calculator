package voters

import (
	"encoding/binary"
	"errors"
	"math/big"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/contracts/relay"
)

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

func NewSigningPolicy(r *relay.RelaySigningPolicyInitialized) *SigningPolicy {
	return &SigningPolicy{
		RewardEpochId:      r.RewardEpochId.Int64(),
		StartVotingRoundId: r.StartVotingRoundId,
		Threshold:          r.Threshold,
		Seed:               r.Seed,
		RawBytes:           r.SigningPolicyBytes,
		BlockTimestamp:     r.Timestamp,
		Voters:             NewVoterSet(r.Voters, r.Weights),
	}
}

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

// Initial seed for random voter selection for finalization reward calculation.
// Initial seed is calculated as a hash of protocol ID and voting round ID.
// The seed is used for the first random.
func InitialHashSeed(rewardEpochSeed *big.Int, protocolId byte, votingRoundId uint32) common.Hash {
	seed := make([]byte, 96)
	// 0-31 bytes are filled with the reward epoch seed
	if rewardEpochSeed != nil {
		rewardEpochSeed.FillBytes(seed[0:32])
	}
	// 32-63 bytes are filled with the protocol ID
	seed[63] = protocolId
	// 64-95 bytes are filled with the voting round ID
	binary.BigEndian.PutUint32(seed[92:96], votingRoundId)
	return common.BytesToHash(crypto.Keccak256(seed))
}

func RandomNumberSequence(initialSeed common.Hash, length int) []common.Hash {
	sequence := make([]common.Hash, length)
	currentSeed := initialSeed
	for i := 0; i < length; i++ {
		sequence[i] = currentSeed
		currentSeed = crypto.Keccak256Hash(currentSeed.Bytes())
	}
	return sequence
}

func (vs *VoterSet) SelectVoters(rewardEpochSeed *big.Int, protocolId byte, votingRoundId uint32, thresholdBIPS uint16) (mapset.Set[common.Address], error) {
	seed := InitialHashSeed(rewardEpochSeed, protocolId, votingRoundId)
	return vs.RandomSelectThresholdWeightVoters(seed, thresholdBIPS)
}

func (vs *VoterSet) RandomSelectThresholdWeightVoters(randomSeed common.Hash, thresholdBIPS uint16) (mapset.Set[common.Address], error) {
	// We limit the threshold to 5000 BIPS to avoid long running loops
	// In practice it will be used with around 1000 BIPS or lower.
	if thresholdBIPS > 5000 {
		return nil, errors.New("Threshold must be between 0 and 5000 BIPS")
	}

	selectedWeight := uint16(0)
	thresholdWeight := uint16(uint64(vs.TotalWeight) * uint64(thresholdBIPS) / 10000)
	currentSeed := randomSeed
	selectedVoters := mapset.NewSet[common.Address]()

	// If threshold weight is not too big, the loop should end quickly
	for selectedWeight < thresholdWeight {
		index := vs.selectVoterIndex(currentSeed)
		selectedAddress := vs.Voters[index]
		if !selectedVoters.Contains(selectedAddress) {
			selectedVoters.Add(selectedAddress)
			selectedWeight += vs.Weights[index]
		}
		currentSeed = crypto.Keccak256Hash(currentSeed.Bytes())
	}
	return selectedVoters, nil
}

// Selects a random voter based provided random number.
func (vs *VoterSet) selectVoterIndex(randomNumber common.Hash) int {
	randomWeight := big.NewInt(0).SetBytes(randomNumber.Bytes())
	randomWeight = randomWeight.Mod(randomWeight, big.NewInt(int64(vs.TotalWeight)))
	return vs.BinarySearch(uint16(randomWeight.Uint64()))
}

// Searches for the highest index of the threshold that is less than or equal to the value.
// Binary search is used.
func (vs *VoterSet) BinarySearch(value uint16) int {
	if value > vs.TotalWeight {
		panic("Value must be between 0 and total weight")
	}
	left := 0
	right := len(vs.Thresholds) - 1
	mid := 0
	if vs.Thresholds[right] <= value {
		return right
	}
	for left < right {
		mid = (left + right) / 2
		if vs.Thresholds[mid] < value {
			left = mid + 1
		} else if vs.Thresholds[mid] > value {
			right = mid
		} else {
			return mid
		}
	}
	return left - 1
}

func (vs *VoterSet) VoterWeight(index int) uint16 {
	return vs.Weights[index]
}

func (vs *VoterSet) Count() int {
	return len(vs.Voters)
}

func (vs *VoterSet) VoterIndex(address common.Address) int {
	voterData, ok := vs.VoterDataMap[address]
	if !ok {
		return -1
	}
	return voterData.Index
}

package params

import (
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

type Network struct {
	Contracts            ContractAddresses
	InitialRewardEpochId int
	Epoch                Epoch
	Ftso                 Ftso
}

type ContractAddresses struct {
	FlareSystemsManager     common.Address
	FtsoRewardOffersManager common.Address
	RewardManager           common.Address
	Submission              common.Address
	Relay                   common.Address
	FlareSystemsCalculator  common.Address
	VoterRegistry           common.Address
}

type Ftso struct {
	ProtocolId  byte
	BurnAddress common.Address

	RandomGenerationBenchingWindow uint64
	NonBenchedRandomVotersMinCount int
	FutureSecureRandomWindow       int

	AdditionalRewardFinalizationWindows   int
	PenaltyFactor                         *big.Int
	GracePeriodForSignaturesDurationSec   uint64
	GracePeriodForFinalizationDurationSec int

	SigningBips      *big.Int
	FinalizationBips *big.Int

	MinimalRewardedNonConsensusDepositedSignaturesPerHashBips uint16
	FinalizationVoterSelectionThresholdWeightBips             int
	CappedStakingFeeBips                                      int
}

var Net Network

func init() {
	network := os.Getenv("NETWORK")
	if network == "" {
		network = "coston" // TODO: remove default
	}
	switch network {
	case "coston":
		Net = coston
	}
}

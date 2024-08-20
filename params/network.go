package params

import (
	"ftsov2-rewarding/logger"
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
	GracePeriodForFinalizationDurationSec uint64

	SigningBips      *big.Int
	FinalizationBips *big.Int

	MinimalRewardedNonConsensusDepositedSignaturesPerHashBips uint16
	FinalizationVoterSelectionThresholdWeightBips             uint16
	CappedStakingFeeBips                                      int64
}

var Net Network

func init() {
	network := os.Getenv("NETWORK")

	switch network {
	case "coston":
		Net = coston
	case "songbird":
		Net = songbird
	}

	logger.Info("Network: %s", network)
}

package params

import (
	"math/big"

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

	AdditionalRewardFinalizationWindows   int
	PenaltyFactor                         *big.Int
	GracePeriodForSignaturesDurationSec   int
	GracePeriodForFinalizationDurationSec int

	SigningBips      *big.Int
	FinalizationBips *big.Int

	MinimalRewardedNonConsensusDepositedSignaturesPerHashBips int
	FinalizationVoterSelectionThresholdWeightBips             int
	CappedStakingFeeBips                                      int
}

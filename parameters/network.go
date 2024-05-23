package parameters

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type NetworkParameters struct {
	Contracts            ContractAddresses
	InitialRewardEpochId int
	Epoch                EpochParameters
	Ftso                 FtsoParameters
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

type FtsoParameters struct {
	ProtocolId  byte
	BurnAddress common.Address

	RandomGenerationBenchingWindow        int
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

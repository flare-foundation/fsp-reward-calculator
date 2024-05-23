package parameters

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var Coston = NetworkParameters{

	Contracts: ContractAddresses{
		FlareSystemsManager:     common.HexToAddress("0x85680Dd93755Fe5d0789773fd0896cEE51F9e358"),
		FtsoRewardOffersManager: common.HexToAddress("0xC9534cB913150aD3e98D792857689B55e2404212"),
		RewardManager:           common.HexToAddress("0xA17197b7Bdff7Be7c3Da39ec08981FB716B70d3A"),
		Submission:              common.HexToAddress("0x2cA6571Daa15ce734Bbd0Bf27D5C9D16787fc33f"),
		Relay:                   common.HexToAddress("0x32D46A1260BB2D8C9d5Ab1C9bBd7FF7D7CfaabCC"),
		FlareSystemsCalculator:  common.HexToAddress("0x43CBAB9C953F54533aadAf7ffCD13c30ec05Edc9"),
		VoterRegistry:           common.HexToAddress("0x051E9Cb16A8676C011faa10efA1ABE95372e7825"),
	},
	InitialRewardEpochId: 1,

	Epoch: EpochParameters{
		FirstVotingRoundStartTs:            1658430000,
		VotingEpochDurationSeconds:         90,
		FirstRewardEpochStartVotingRoundId: 0,
		RewardEpochDurationInVotingEpochs:  240,
		RevealDeadlineSeconds:              45,
	},

	Ftso: FtsoParameters{
		ProtocolId:                            100,
		BurnAddress:                           common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
		RandomGenerationBenchingWindow:        20,
		AdditionalRewardFinalizationWindows:   0,
		PenaltyFactor:                         big.NewInt(30),
		GracePeriodForSignaturesDurationSec:   10,
		GracePeriodForFinalizationDurationSec: 20,
		SigningBips:                           big.NewInt(1000),
		FinalizationBips:                      big.NewInt(1000),

		MinimalRewardedNonConsensusDepositedSignaturesPerHashBips: 3000,
		FinalizationVoterSelectionThresholdWeightBips:             500,
		CappedStakingFeeBips: 2000,
	},
}

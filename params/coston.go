package params

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var coston = Network{
	Name: "coston",

	Contracts: ContractAddresses{
		FlareSystemsManager:        common.HexToAddress("0x85680Dd93755Fe5d0789773fd0896cEE51F9e358"),
		FtsoRewardOffersManager:    common.HexToAddress("0xC9534cB913150aD3e98D792857689B55e2404212"),
		RewardManager:              common.HexToAddress("0xA17197b7Bdff7Be7c3Da39ec08981FB716B70d3A"),
		Submission:                 common.HexToAddress("0x2cA6571Daa15ce734Bbd0Bf27D5C9D16787fc33f"),
		Relay:                      common.HexToAddress("0x32D46A1260BB2D8C9d5Ab1C9bBd7FF7D7CfaabCC"),
		FlareSystemsCalculator:     common.HexToAddress("0x43CBAB9C953F54533aadAf7ffCD13c30ec05Edc9"),
		VoterRegistry:              common.HexToAddress("0xE2c06DF29d175Aa0EcfcD10134eB96f8C94448A3"),
		FastUpdateIncentiveManager: common.HexToAddress("0x8c45666369B174806E1AB78D989ddd79a3267F3b"),
		FastUpdater:                common.HexToAddress("0x9B931f5d3e24fc8C9064DB35bDc8FB4bE0E862f9"),
	},

	InitialRewardEpochId: 2466,

	Epoch: Epoch{
		FirstVotingRoundStartTs:                    1658429955,
		VotingRoundDurationSeconds:                 90,
		FirstRewardEpochStartVotingRoundId:         0,
		RewardEpochDurationInVotingEpochs:          240,
		RevealDeadlineSeconds:                      45,
		NewSigningPolicyInitializationStartSeconds: 7200,
	},

	Ftso: Ftso{
		ProtocolId:                            100,
		BurnAddress:                           common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
		RandomGenerationBenchingWindow:        20,
		FutureSecureRandomWindow:              30,
		AdditionalRewardFinalizationWindows:   0,
		PenaltyFactor:                         big.NewInt(30),
		GracePeriodForSignaturesDurationSec:   10,
		GracePeriodForFinalizationDurationSec: 20,
		SigningBips:                           big.NewInt(1000),
		FinalizationBips:                      big.NewInt(1000),
		MinimalRewardedNonConsensusDepositedSignaturesPerHashBips: 3000,
		FinalizationVoterSelectionThresholdWeightBips:             500,
		CappedStakingFeeBips:           2000,
		NonBenchedRandomVotersMinCount: 2,
	},
}

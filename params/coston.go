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
		RewardManager:              common.HexToAddress("0x2ade9972E7f27200872D378acF7a1BaD8D696FC5"),
		Submission:                 common.HexToAddress("0x2cA6571Daa15ce734Bbd0Bf27D5C9D16787fc33f"),
		Relay:                      common.HexToAddress("0x92a6E1127262106611e1e129BB64B6D8654273F7"),
		FlareSystemsCalculator:     common.HexToAddress("0x43CBAB9C953F54533aadAf7ffCD13c30ec05Edc9"),
		VoterRegistry:              common.HexToAddress("0xE2c06DF29d175Aa0EcfcD10134eB96f8C94448A3"),
		FastUpdateIncentiveManager: common.HexToAddress("0x8c45666369B174806E1AB78D989ddd79a3267F3b"),
		FastUpdater:                common.HexToAddress("0xB8336A96b4b8af89f60EA080002214191Bc8293A"),
		FdcHub:                     common.HexToAddress("0x1c78A073E3BD2aCa4cc327d55FB0cD4f0549B55b"),
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

	Fdc: Fdc{
		FinalizationBips: big.NewInt(1000),
	},
}

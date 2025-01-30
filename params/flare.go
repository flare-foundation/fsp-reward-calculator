package params

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var flare = Network{
	Name: "flare",

	Contracts: ContractAddresses{
		FlareSystemsManager:        common.HexToAddress("0x89e50DC0380e597ecE79c8494bAAFD84537AD0D4"),
		FtsoRewardOffersManager:    common.HexToAddress("0x244EA7f173895968128D5847Df2C75B1460ac685"),
		RewardManager:              common.HexToAddress("0xC8f55c5aA2C752eE285Bd872855C749f4ee6239B"),
		Submission:                 common.HexToAddress("0x2cA6571Daa15ce734Bbd0Bf27D5C9D16787fc33f"),
		Relay:                      common.HexToAddress("0x57a4c3676d08Aa5d15410b5A6A80fBcEF72f3F45"),
		FlareSystemsCalculator:     common.HexToAddress("0x67c4B11c710D35a279A41cff5eb089Fe72748CF8"),
		VoterRegistry:              common.HexToAddress("0x2580101692366e2f331e891180d9ffdF861Fce83"),
		FastUpdateIncentiveManager: common.HexToAddress("0xd648e8ACA486Ce876D641A0F53ED1F2E9eF4885D"),
		FastUpdater:                common.HexToAddress("0xdBF71d7840934EB82FA10173103D4e9fd4054dd1"),
		FdcHub:                     common.HexToAddress("0xc25c749DC27Efb1864Cb3DADa8845B7687eB2d44"),
	},

	InitialRewardEpochId: 183,

	Epoch: Epoch{
		FirstVotingRoundStartTs:                    1658430000,
		VotingRoundDurationSeconds:                 90,
		FirstRewardEpochStartVotingRoundId:         0,
		RewardEpochDurationInVotingEpochs:          3360,
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

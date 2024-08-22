package params

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var songbird = Network{

	Contracts: ContractAddresses{
		FlareSystemsManager:        common.HexToAddress("0x421c69E22f48e14Fc2d2Ee3812c59bfb81c38516"),
		FtsoRewardOffersManager:    common.HexToAddress("0x5aB9cB258a342001C4663D9526A1c54cCcF8C545"),
		RewardManager:              common.HexToAddress("0x8A80583BD5A5Cd8f68De585163259D61Ea8dc904"),
		Submission:                 common.HexToAddress("0x2cA6571Daa15ce734Bbd0Bf27D5C9D16787fc33f"),
		Relay:                      common.HexToAddress("0xbA35e39D01A3f5710d1e43FC61dbb738B68641c4"),
		FlareSystemsCalculator:     common.HexToAddress("0x126FAeEc75601dA3354c0b5Cc0b60C85fCbC3A5e"),
		VoterRegistry:              common.HexToAddress("0x31B9EC65C731c7D973a33Ef3FC83B653f540dC8D"),
		FastUpdateIncentiveManager: common.HexToAddress("0x596C70Ad6fFFdb9b6158F1Dfd0bc32cc72B82006"),
		FastUpdater:                common.HexToAddress("0x70e8870ef234EcD665F96Da4c669dc12c1e1c116"),
	},

	InitialRewardEpochId: 183,

	Epoch: Epoch{
		FirstVotingRoundStartTs:                    1658429955,
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

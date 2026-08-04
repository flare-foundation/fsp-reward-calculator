package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var songbird = Network{
	Name: "songbird",

	Contracts: ContractAddresses{
		FlareSystemsManager:        common.HexToAddress("0x421c69E22f48e14Fc2d2Ee3812c59bfb81c38516"),
		FtsoRewardOffersManager:    common.HexToAddress("0x5aB9cB258a342001C4663D9526A1c54cCcF8C545"),
		RewardManager:              common.HexToAddress("0xE26AD68b17224951b5740F33926Cc438764eB9a7"),
		Submission:                 common.HexToAddress("0x2cA6571Daa15ce734Bbd0Bf27D5C9D16787fc33f"),
		Relay:                      common.HexToAddress("0xCB86E8Be709001e01897Bf59847406853da8f14b"),
		OldRelay:                   common.HexToAddress("0x67a916E175a2aF01369294739AA60dDdE1Fad189"),
		FlareSystemsCalculator:     common.HexToAddress("0x31a5B8E7ca6dFC7B963f5D029F0884ef19E53A24"),
		OldFlareSystemsCalculator:  common.HexToAddress("0x126FAeEc75601dA3354c0b5Cc0b60C85fCbC3A5e"),
		VoterRegistry:              common.HexToAddress("0xd23FAE88c09e6A77dD9eFcc29D6bBC55D2e74310"),
		OldVoterRegistry:           common.HexToAddress("0x31B9EC65C731c7D973a33Ef3FC83B653f540dC8D"),
		FastUpdateIncentiveManager: common.HexToAddress("0x596C70Ad6fFFdb9b6158F1Dfd0bc32cc72B82006"),
		FastUpdater:                common.HexToAddress("0x7D9F73FD9bC4607daCB618FF895585f98BFDD06B"),
		FdcHub:                     common.HexToAddress("0xCfD4669a505A70c2cE85db8A1c1d14BcDE5a1a06"),
		FlareTeeManager:            common.HexToAddress("0x5C2dE0DeFC3FDBbF8e12c12bD0b1629Ed37DC767"),
		Fdc2Hub:                    common.HexToAddress("0x4234a8f5D255d91d56df53d0cc78c0Cc2B67ACD8"),
	},

	InitialRewardEpochId: 183,

	Epoch: Epoch{
		FirstVotingRoundStartTs:                    1658429955,
		VotingEpochDurationSeconds:                 90,
		FirstRewardEpochStartVotingRoundId:         0,
		RewardEpochDurationInVotingEpochs:          3360,
		RevealDeadlineSeconds:                      45,
		NewSigningPolicyInitializationStartSeconds: 7200,
	},

	Ftso: Ftso{
		ProtocolId:                            100,
		BurnAddress:                           common.HexToAddress("0xAC3F85d29119836545670b2FCeFe35C829bE35ab"),
		RandomGenerationBenchingWindow:        20,
		FutureSecureRandomWindow:              30,
		AdditionalRewardFinalizationWindows:   0,
		PenaltyFactor:                         big.NewInt(30),
		GracePeriodForSignaturesDurationSec:   15,
		GracePeriodForFinalizationDurationSec: 20,
		SigningBips:                           big.NewInt(1000),
		FinalizationBips:                      big.NewInt(1000),
		MinimalRewardedNonConsensusDepositedSignaturesPerHashBips: 3000,
		FinalizationVoterSelectionThresholdWeightBips:             500,
		CappedStakingFeeBips:           2000,
		NonBenchedRandomVotersMinCount: 2,
	},

	Fdc: Fdc{
		ProtocolId:       200,
		FinalizationBips: big.NewInt(1000),
		PenaltyFactor:    big.NewInt(30),
		// Songbird has no FDC->FIRE split (0 bips).
		FireFeeSplitBips: big.NewInt(0),
	},

	// FIP.16 activates on Songbird with reward epoch 417. There is no P-chain staking on Songbird,
	// so the 5x stake weighting is inert and no FDC->FIRE split applies; the FIRE pool address is
	// the conventional burn address.
	Fip16ActivationEpoch: 417,
	FirePoolAddress:      common.HexToAddress("0x000000000000000000000000000000000000dEaD"),

	// The FCC contracts were deployed on Songbird part-way through reward epoch 419, and that is also the
	// epoch their fees are accounted for from. A mid-epoch deployment loses nothing: before the deployment
	// transaction there are neither FCC events nor receiveRewards credits to account for.
	FccActivationEpoch: 419,
	FccFeesAddress:     common.HexToAddress("0x3390E1aDf46568cCC95c3571424937b042094ac2"),
}

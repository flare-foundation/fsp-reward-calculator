package params

import (
	"fsp-rewards-calculator/common/ty"
	"fsp-rewards-calculator/logger"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Network struct {
	Name                 string
	Contracts            ContractAddresses
	InitialRewardEpochId int
	Epoch                Epoch
	Ftso                 Ftso
	Fdc                  Fdc

	// Fip16ActivationEpoch is the first reward epoch id (inclusive) in which FIP.16 vote-power
	// unification is in effect. Set to Fip16NotActivated on networks where it is not deployed.
	Fip16ActivationEpoch uint64
	// FirePoolAddress receives the FIRE share of confirmed FDC attestation request fees once
	// FIP.16 is active (a burn address on networks without a FIRE pool).
	FirePoolAddress common.Address
}

type ContractAddresses struct {
	FlareSystemsManager        common.Address
	FtsoRewardOffersManager    common.Address
	RewardManager              common.Address
	Submission                 common.Address
	OldRelay                   common.Address
	Relay                      common.Address
	OldFlareSystemsCalculator  common.Address
	FlareSystemsCalculator     common.Address
	OldVoterRegistry           common.Address
	VoterRegistry              common.Address
	FastUpdateIncentiveManager common.Address
	FastUpdater                common.Address
	FdcHub                     common.Address
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

type Fdc struct {
	ProtocolId       byte
	FinalizationBips *big.Int
	PenaltyFactor    *big.Int
	// FireFeeSplitBips is the share (in bips) of confirmed FDC attestation request fees that is
	// routed to the FIRE pool once FIP.16 is active. Zero on networks without the FDC->FIRE split.
	FireFeeSplitBips *big.Int
}

// Fip16NotActivated is a sentinel activation epoch meaning "FIP.16 not activated yet".
const Fip16NotActivated = math.MaxUint64

// Fip16StakeWeightMultiplier is the multiplier applied to P-chain stake relative to capped C-chain
// delegation weight in the delegation-vs-stake reward split once FIP.16 is active.
var Fip16StakeWeightMultiplier = big.NewInt(5)

// Fip16Active reports whether FIP.16 vote-power unification applies to the given reward epoch.
func Fip16Active(epoch ty.RewardEpochId) bool {
	return uint64(epoch) >= Net.Fip16ActivationEpoch
}

var Net Network

func InitNetwork(network string) {
	switch network {
	case "coston":
		Net = coston
	case "songbird":
		Net = songbird
	case "flare":
		Net = flare
	default:
		logger.Fatal("Unsupported network: %s.", network)
	}

	logger.Info("Initialized network: %s", network)
}

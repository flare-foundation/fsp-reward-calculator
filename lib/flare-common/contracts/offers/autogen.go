// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package offers

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IFtsoRewardOffersManagerOffer is an auto generated low-level Go binding around an user-defined struct.
type IFtsoRewardOffersManagerOffer struct {
	Amount                    *big.Int
	FeedId                    [21]byte
	MinRewardedTurnoutBIPS    uint16
	PrimaryBandRewardSharePPM *big.Int
	SecondaryBandWidthPPM     *big.Int
	ClaimBackAddress          common.Address
}

// OffersMetaData contains all meta data concerning the Offers contract.
var OffersMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIGovernanceSettings\",\"name\":\"_governanceSettings\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialGovernance\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_addressUpdater\",\"type\":\"address\"},{\"internalType\":\"uint128\",\"name\":\"_minimalRewardsOfferValueWei\",\"type\":\"uint128\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"authorizedAmountWei\",\"type\":\"uint256\"}],\"name\":\"DailyAuthorizedInflationSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"allowedAfterTimestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"encodedCall\",\"type\":\"bytes\"}],\"name\":\"GovernanceCallTimelocked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"initialGovernance\",\"type\":\"address\"}],\"name\":\"GovernanceInitialised\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"governanceSettings\",\"type\":\"address\"}],\"name\":\"GovernedProductionModeEntered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amountReceivedWei\",\"type\":\"uint256\"}],\"name\":\"InflationReceived\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint24\",\"name\":\"rewardEpochId\",\"type\":\"uint24\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"feedIds\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"decimals\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"minRewardedTurnoutBIPS\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint24\",\"name\":\"primaryBandRewardSharePPM\",\"type\":\"uint24\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"secondaryBandWidthPPMs\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"mode\",\"type\":\"uint16\"}],\"name\":\"InflationRewardsOffered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"valueWei\",\"type\":\"uint256\"}],\"name\":\"MinimalRewardsOfferValueSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint24\",\"name\":\"rewardEpochId\",\"type\":\"uint24\"},{\"indexed\":false,\"internalType\":\"bytes21\",\"name\":\"feedId\",\"type\":\"bytes21\"},{\"indexed\":false,\"internalType\":\"int8\",\"name\":\"decimals\",\"type\":\"int8\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"minRewardedTurnoutBIPS\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint24\",\"name\":\"primaryBandRewardSharePPM\",\"type\":\"uint24\"},{\"indexed\":false,\"internalType\":\"uint24\",\"name\":\"secondaryBandWidthPPM\",\"type\":\"uint24\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"claimBackAddress\",\"type\":\"address\"}],\"name\":\"RewardsOffered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"TimelockedGovernanceCallCanceled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"TimelockedGovernanceCallExecuted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"_selector\",\"type\":\"bytes4\"}],\"name\":\"cancelGovernanceCall\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dailyAuthorizedInflation\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"_selector\",\"type\":\"bytes4\"}],\"name\":\"executeGovernanceCall\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"flareSystemsManager\",\"outputs\":[{\"internalType\":\"contractIIFlareSystemsManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ftsoFeedDecimals\",\"outputs\":[{\"internalType\":\"contractIFtsoFeedDecimals\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ftsoInflationConfigurations\",\"outputs\":[{\"internalType\":\"contractIFtsoInflationConfigurations\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAddressUpdater\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"_addressUpdater\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getContractName\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getExpectedBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getInflationAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTokenPoolSupplyData\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"_lockedFundsWei\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_totalInflationAuthorizedWei\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_totalClaimedWei\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceSettings\",\"outputs\":[{\"internalType\":\"contractIGovernanceSettings\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIGovernanceSettings\",\"name\":\"_governanceSettings\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialGovernance\",\"type\":\"address\"}],\"name\":\"initialise\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"isExecutor\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastInflationAuthorizationReceivedTs\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastInflationReceivedTs\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minimalRewardsOfferValueWei\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint24\",\"name\":\"_nextRewardEpochId\",\"type\":\"uint24\"},{\"components\":[{\"internalType\":\"uint120\",\"name\":\"amount\",\"type\":\"uint120\"},{\"internalType\":\"bytes21\",\"name\":\"feedId\",\"type\":\"bytes21\"},{\"internalType\":\"uint16\",\"name\":\"minRewardedTurnoutBIPS\",\"type\":\"uint16\"},{\"internalType\":\"uint24\",\"name\":\"primaryBandRewardSharePPM\",\"type\":\"uint24\"},{\"internalType\":\"uint24\",\"name\":\"secondaryBandWidthPPM\",\"type\":\"uint24\"},{\"internalType\":\"address\",\"name\":\"claimBackAddress\",\"type\":\"address\"}],\"internalType\":\"structIFtsoRewardOffersManager.Offer[]\",\"name\":\"_offers\",\"type\":\"tuple[]\"}],\"name\":\"offerRewards\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"productionMode\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"receiveInflation\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rewardManager\",\"outputs\":[{\"internalType\":\"contractIIRewardManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_toAuthorizeWei\",\"type\":\"uint256\"}],\"name\":\"setDailyAuthorizedInflation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"_minimalRewardsOfferValueWei\",\"type\":\"uint128\"}],\"name\":\"setMinimalRewardsOfferValue\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"switchToProductionMode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"}],\"name\":\"timelockedCalls\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"allowedAfterTimestamp\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"encodedCall\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalInflationAuthorizedWei\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalInflationReceivedWei\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalInflationRewardsOfferedWei\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint24\",\"name\":\"_currentRewardEpochId\",\"type\":\"uint24\"},{\"internalType\":\"uint64\",\"name\":\"_currentRewardEpochExpectedEndTs\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"_rewardEpochDurationSeconds\",\"type\":\"uint64\"}],\"name\":\"triggerRewardEpochSwitchover\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"_contractNameHashes\",\"type\":\"bytes32[]\"},{\"internalType\":\"address[]\",\"name\":\"_contractAddresses\",\"type\":\"address[]\"}],\"name\":\"updateContractAddresses\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// OffersABI is the input ABI used to generate the binding from.
// Deprecated: Use OffersMetaData.ABI instead.
var OffersABI = OffersMetaData.ABI

// Offers is an auto generated Go binding around an Ethereum contract.
type Offers struct {
	OffersCaller     // Read-only binding to the contract
	OffersTransactor // Write-only binding to the contract
	OffersFilterer   // Log filterer for contract events
}

// OffersCaller is an auto generated read-only Go binding around an Ethereum contract.
type OffersCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OffersTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OffersTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OffersFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OffersFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OffersSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OffersSession struct {
	Contract     *Offers           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OffersCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OffersCallerSession struct {
	Contract *OffersCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// OffersTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OffersTransactorSession struct {
	Contract     *OffersTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OffersRaw is an auto generated low-level Go binding around an Ethereum contract.
type OffersRaw struct {
	Contract *Offers // Generic contract binding to access the raw methods on
}

// OffersCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OffersCallerRaw struct {
	Contract *OffersCaller // Generic read-only contract binding to access the raw methods on
}

// OffersTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OffersTransactorRaw struct {
	Contract *OffersTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOffers creates a new instance of Offers, bound to a specific deployed contract.
func NewOffers(address common.Address, backend bind.ContractBackend) (*Offers, error) {
	contract, err := bindOffers(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Offers{OffersCaller: OffersCaller{contract: contract}, OffersTransactor: OffersTransactor{contract: contract}, OffersFilterer: OffersFilterer{contract: contract}}, nil
}

// NewOffersCaller creates a new read-only instance of Offers, bound to a specific deployed contract.
func NewOffersCaller(address common.Address, caller bind.ContractCaller) (*OffersCaller, error) {
	contract, err := bindOffers(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OffersCaller{contract: contract}, nil
}

// NewOffersTransactor creates a new write-only instance of Offers, bound to a specific deployed contract.
func NewOffersTransactor(address common.Address, transactor bind.ContractTransactor) (*OffersTransactor, error) {
	contract, err := bindOffers(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OffersTransactor{contract: contract}, nil
}

// NewOffersFilterer creates a new log filterer instance of Offers, bound to a specific deployed contract.
func NewOffersFilterer(address common.Address, filterer bind.ContractFilterer) (*OffersFilterer, error) {
	contract, err := bindOffers(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OffersFilterer{contract: contract}, nil
}

// bindOffers binds a generic wrapper to an already deployed contract.
func bindOffers(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OffersMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Offers *OffersRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Offers.Contract.OffersCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Offers *OffersRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Offers.Contract.OffersTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Offers *OffersRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Offers.Contract.OffersTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Offers *OffersCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Offers.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Offers *OffersTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Offers.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Offers *OffersTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Offers.Contract.contract.Transact(opts, method, params...)
}

// DailyAuthorizedInflation is a free data retrieval call binding the contract method 0x708e34ce.
//
// Solidity: function dailyAuthorizedInflation() view returns(uint256)
func (_Offers *OffersCaller) DailyAuthorizedInflation(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "dailyAuthorizedInflation")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DailyAuthorizedInflation is a free data retrieval call binding the contract method 0x708e34ce.
//
// Solidity: function dailyAuthorizedInflation() view returns(uint256)
func (_Offers *OffersSession) DailyAuthorizedInflation() (*big.Int, error) {
	return _Offers.Contract.DailyAuthorizedInflation(&_Offers.CallOpts)
}

// DailyAuthorizedInflation is a free data retrieval call binding the contract method 0x708e34ce.
//
// Solidity: function dailyAuthorizedInflation() view returns(uint256)
func (_Offers *OffersCallerSession) DailyAuthorizedInflation() (*big.Int, error) {
	return _Offers.Contract.DailyAuthorizedInflation(&_Offers.CallOpts)
}

// FlareSystemsManager is a free data retrieval call binding the contract method 0xfaae7fc9.
//
// Solidity: function flareSystemsManager() view returns(address)
func (_Offers *OffersCaller) FlareSystemsManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "flareSystemsManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FlareSystemsManager is a free data retrieval call binding the contract method 0xfaae7fc9.
//
// Solidity: function flareSystemsManager() view returns(address)
func (_Offers *OffersSession) FlareSystemsManager() (common.Address, error) {
	return _Offers.Contract.FlareSystemsManager(&_Offers.CallOpts)
}

// FlareSystemsManager is a free data retrieval call binding the contract method 0xfaae7fc9.
//
// Solidity: function flareSystemsManager() view returns(address)
func (_Offers *OffersCallerSession) FlareSystemsManager() (common.Address, error) {
	return _Offers.Contract.FlareSystemsManager(&_Offers.CallOpts)
}

// FtsoFeedDecimals is a free data retrieval call binding the contract method 0x9065c974.
//
// Solidity: function ftsoFeedDecimals() view returns(address)
func (_Offers *OffersCaller) FtsoFeedDecimals(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "ftsoFeedDecimals")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FtsoFeedDecimals is a free data retrieval call binding the contract method 0x9065c974.
//
// Solidity: function ftsoFeedDecimals() view returns(address)
func (_Offers *OffersSession) FtsoFeedDecimals() (common.Address, error) {
	return _Offers.Contract.FtsoFeedDecimals(&_Offers.CallOpts)
}

// FtsoFeedDecimals is a free data retrieval call binding the contract method 0x9065c974.
//
// Solidity: function ftsoFeedDecimals() view returns(address)
func (_Offers *OffersCallerSession) FtsoFeedDecimals() (common.Address, error) {
	return _Offers.Contract.FtsoFeedDecimals(&_Offers.CallOpts)
}

// FtsoInflationConfigurations is a free data retrieval call binding the contract method 0xc27fb624.
//
// Solidity: function ftsoInflationConfigurations() view returns(address)
func (_Offers *OffersCaller) FtsoInflationConfigurations(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "ftsoInflationConfigurations")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FtsoInflationConfigurations is a free data retrieval call binding the contract method 0xc27fb624.
//
// Solidity: function ftsoInflationConfigurations() view returns(address)
func (_Offers *OffersSession) FtsoInflationConfigurations() (common.Address, error) {
	return _Offers.Contract.FtsoInflationConfigurations(&_Offers.CallOpts)
}

// FtsoInflationConfigurations is a free data retrieval call binding the contract method 0xc27fb624.
//
// Solidity: function ftsoInflationConfigurations() view returns(address)
func (_Offers *OffersCallerSession) FtsoInflationConfigurations() (common.Address, error) {
	return _Offers.Contract.FtsoInflationConfigurations(&_Offers.CallOpts)
}

// GetAddressUpdater is a free data retrieval call binding the contract method 0x5267a15d.
//
// Solidity: function getAddressUpdater() view returns(address _addressUpdater)
func (_Offers *OffersCaller) GetAddressUpdater(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "getAddressUpdater")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAddressUpdater is a free data retrieval call binding the contract method 0x5267a15d.
//
// Solidity: function getAddressUpdater() view returns(address _addressUpdater)
func (_Offers *OffersSession) GetAddressUpdater() (common.Address, error) {
	return _Offers.Contract.GetAddressUpdater(&_Offers.CallOpts)
}

// GetAddressUpdater is a free data retrieval call binding the contract method 0x5267a15d.
//
// Solidity: function getAddressUpdater() view returns(address _addressUpdater)
func (_Offers *OffersCallerSession) GetAddressUpdater() (common.Address, error) {
	return _Offers.Contract.GetAddressUpdater(&_Offers.CallOpts)
}

// GetContractName is a free data retrieval call binding the contract method 0xf5f5ba72.
//
// Solidity: function getContractName() pure returns(string)
func (_Offers *OffersCaller) GetContractName(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "getContractName")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetContractName is a free data retrieval call binding the contract method 0xf5f5ba72.
//
// Solidity: function getContractName() pure returns(string)
func (_Offers *OffersSession) GetContractName() (string, error) {
	return _Offers.Contract.GetContractName(&_Offers.CallOpts)
}

// GetContractName is a free data retrieval call binding the contract method 0xf5f5ba72.
//
// Solidity: function getContractName() pure returns(string)
func (_Offers *OffersCallerSession) GetContractName() (string, error) {
	return _Offers.Contract.GetContractName(&_Offers.CallOpts)
}

// GetExpectedBalance is a free data retrieval call binding the contract method 0xaf04cd3b.
//
// Solidity: function getExpectedBalance() view returns(uint256)
func (_Offers *OffersCaller) GetExpectedBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "getExpectedBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetExpectedBalance is a free data retrieval call binding the contract method 0xaf04cd3b.
//
// Solidity: function getExpectedBalance() view returns(uint256)
func (_Offers *OffersSession) GetExpectedBalance() (*big.Int, error) {
	return _Offers.Contract.GetExpectedBalance(&_Offers.CallOpts)
}

// GetExpectedBalance is a free data retrieval call binding the contract method 0xaf04cd3b.
//
// Solidity: function getExpectedBalance() view returns(uint256)
func (_Offers *OffersCallerSession) GetExpectedBalance() (*big.Int, error) {
	return _Offers.Contract.GetExpectedBalance(&_Offers.CallOpts)
}

// GetInflationAddress is a free data retrieval call binding the contract method 0xed39d3f8.
//
// Solidity: function getInflationAddress() view returns(address)
func (_Offers *OffersCaller) GetInflationAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "getInflationAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetInflationAddress is a free data retrieval call binding the contract method 0xed39d3f8.
//
// Solidity: function getInflationAddress() view returns(address)
func (_Offers *OffersSession) GetInflationAddress() (common.Address, error) {
	return _Offers.Contract.GetInflationAddress(&_Offers.CallOpts)
}

// GetInflationAddress is a free data retrieval call binding the contract method 0xed39d3f8.
//
// Solidity: function getInflationAddress() view returns(address)
func (_Offers *OffersCallerSession) GetInflationAddress() (common.Address, error) {
	return _Offers.Contract.GetInflationAddress(&_Offers.CallOpts)
}

// GetTokenPoolSupplyData is a free data retrieval call binding the contract method 0x2dafdbbf.
//
// Solidity: function getTokenPoolSupplyData() view returns(uint256 _lockedFundsWei, uint256 _totalInflationAuthorizedWei, uint256 _totalClaimedWei)
func (_Offers *OffersCaller) GetTokenPoolSupplyData(opts *bind.CallOpts) (struct {
	LockedFundsWei              *big.Int
	TotalInflationAuthorizedWei *big.Int
	TotalClaimedWei             *big.Int
}, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "getTokenPoolSupplyData")

	outstruct := new(struct {
		LockedFundsWei              *big.Int
		TotalInflationAuthorizedWei *big.Int
		TotalClaimedWei             *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.LockedFundsWei = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TotalInflationAuthorizedWei = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.TotalClaimedWei = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetTokenPoolSupplyData is a free data retrieval call binding the contract method 0x2dafdbbf.
//
// Solidity: function getTokenPoolSupplyData() view returns(uint256 _lockedFundsWei, uint256 _totalInflationAuthorizedWei, uint256 _totalClaimedWei)
func (_Offers *OffersSession) GetTokenPoolSupplyData() (struct {
	LockedFundsWei              *big.Int
	TotalInflationAuthorizedWei *big.Int
	TotalClaimedWei             *big.Int
}, error) {
	return _Offers.Contract.GetTokenPoolSupplyData(&_Offers.CallOpts)
}

// GetTokenPoolSupplyData is a free data retrieval call binding the contract method 0x2dafdbbf.
//
// Solidity: function getTokenPoolSupplyData() view returns(uint256 _lockedFundsWei, uint256 _totalInflationAuthorizedWei, uint256 _totalClaimedWei)
func (_Offers *OffersCallerSession) GetTokenPoolSupplyData() (struct {
	LockedFundsWei              *big.Int
	TotalInflationAuthorizedWei *big.Int
	TotalClaimedWei             *big.Int
}, error) {
	return _Offers.Contract.GetTokenPoolSupplyData(&_Offers.CallOpts)
}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_Offers *OffersCaller) Governance(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "governance")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_Offers *OffersSession) Governance() (common.Address, error) {
	return _Offers.Contract.Governance(&_Offers.CallOpts)
}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_Offers *OffersCallerSession) Governance() (common.Address, error) {
	return _Offers.Contract.Governance(&_Offers.CallOpts)
}

// GovernanceSettings is a free data retrieval call binding the contract method 0x62354e03.
//
// Solidity: function governanceSettings() view returns(address)
func (_Offers *OffersCaller) GovernanceSettings(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "governanceSettings")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GovernanceSettings is a free data retrieval call binding the contract method 0x62354e03.
//
// Solidity: function governanceSettings() view returns(address)
func (_Offers *OffersSession) GovernanceSettings() (common.Address, error) {
	return _Offers.Contract.GovernanceSettings(&_Offers.CallOpts)
}

// GovernanceSettings is a free data retrieval call binding the contract method 0x62354e03.
//
// Solidity: function governanceSettings() view returns(address)
func (_Offers *OffersCallerSession) GovernanceSettings() (common.Address, error) {
	return _Offers.Contract.GovernanceSettings(&_Offers.CallOpts)
}

// IsExecutor is a free data retrieval call binding the contract method 0xdebfda30.
//
// Solidity: function isExecutor(address _address) view returns(bool)
func (_Offers *OffersCaller) IsExecutor(opts *bind.CallOpts, _address common.Address) (bool, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "isExecutor", _address)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsExecutor is a free data retrieval call binding the contract method 0xdebfda30.
//
// Solidity: function isExecutor(address _address) view returns(bool)
func (_Offers *OffersSession) IsExecutor(_address common.Address) (bool, error) {
	return _Offers.Contract.IsExecutor(&_Offers.CallOpts, _address)
}

// IsExecutor is a free data retrieval call binding the contract method 0xdebfda30.
//
// Solidity: function isExecutor(address _address) view returns(bool)
func (_Offers *OffersCallerSession) IsExecutor(_address common.Address) (bool, error) {
	return _Offers.Contract.IsExecutor(&_Offers.CallOpts, _address)
}

// LastInflationAuthorizationReceivedTs is a free data retrieval call binding the contract method 0x473252c4.
//
// Solidity: function lastInflationAuthorizationReceivedTs() view returns(uint256)
func (_Offers *OffersCaller) LastInflationAuthorizationReceivedTs(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "lastInflationAuthorizationReceivedTs")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastInflationAuthorizationReceivedTs is a free data retrieval call binding the contract method 0x473252c4.
//
// Solidity: function lastInflationAuthorizationReceivedTs() view returns(uint256)
func (_Offers *OffersSession) LastInflationAuthorizationReceivedTs() (*big.Int, error) {
	return _Offers.Contract.LastInflationAuthorizationReceivedTs(&_Offers.CallOpts)
}

// LastInflationAuthorizationReceivedTs is a free data retrieval call binding the contract method 0x473252c4.
//
// Solidity: function lastInflationAuthorizationReceivedTs() view returns(uint256)
func (_Offers *OffersCallerSession) LastInflationAuthorizationReceivedTs() (*big.Int, error) {
	return _Offers.Contract.LastInflationAuthorizationReceivedTs(&_Offers.CallOpts)
}

// LastInflationReceivedTs is a free data retrieval call binding the contract method 0x12afcf0b.
//
// Solidity: function lastInflationReceivedTs() view returns(uint256)
func (_Offers *OffersCaller) LastInflationReceivedTs(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "lastInflationReceivedTs")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastInflationReceivedTs is a free data retrieval call binding the contract method 0x12afcf0b.
//
// Solidity: function lastInflationReceivedTs() view returns(uint256)
func (_Offers *OffersSession) LastInflationReceivedTs() (*big.Int, error) {
	return _Offers.Contract.LastInflationReceivedTs(&_Offers.CallOpts)
}

// LastInflationReceivedTs is a free data retrieval call binding the contract method 0x12afcf0b.
//
// Solidity: function lastInflationReceivedTs() view returns(uint256)
func (_Offers *OffersCallerSession) LastInflationReceivedTs() (*big.Int, error) {
	return _Offers.Contract.LastInflationReceivedTs(&_Offers.CallOpts)
}

// MinimalRewardsOfferValueWei is a free data retrieval call binding the contract method 0xc85f3a46.
//
// Solidity: function minimalRewardsOfferValueWei() view returns(uint256)
func (_Offers *OffersCaller) MinimalRewardsOfferValueWei(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "minimalRewardsOfferValueWei")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinimalRewardsOfferValueWei is a free data retrieval call binding the contract method 0xc85f3a46.
//
// Solidity: function minimalRewardsOfferValueWei() view returns(uint256)
func (_Offers *OffersSession) MinimalRewardsOfferValueWei() (*big.Int, error) {
	return _Offers.Contract.MinimalRewardsOfferValueWei(&_Offers.CallOpts)
}

// MinimalRewardsOfferValueWei is a free data retrieval call binding the contract method 0xc85f3a46.
//
// Solidity: function minimalRewardsOfferValueWei() view returns(uint256)
func (_Offers *OffersCallerSession) MinimalRewardsOfferValueWei() (*big.Int, error) {
	return _Offers.Contract.MinimalRewardsOfferValueWei(&_Offers.CallOpts)
}

// ProductionMode is a free data retrieval call binding the contract method 0xe17f212e.
//
// Solidity: function productionMode() view returns(bool)
func (_Offers *OffersCaller) ProductionMode(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "productionMode")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ProductionMode is a free data retrieval call binding the contract method 0xe17f212e.
//
// Solidity: function productionMode() view returns(bool)
func (_Offers *OffersSession) ProductionMode() (bool, error) {
	return _Offers.Contract.ProductionMode(&_Offers.CallOpts)
}

// ProductionMode is a free data retrieval call binding the contract method 0xe17f212e.
//
// Solidity: function productionMode() view returns(bool)
func (_Offers *OffersCallerSession) ProductionMode() (bool, error) {
	return _Offers.Contract.ProductionMode(&_Offers.CallOpts)
}

// RewardManager is a free data retrieval call binding the contract method 0x0f4ef8a6.
//
// Solidity: function rewardManager() view returns(address)
func (_Offers *OffersCaller) RewardManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "rewardManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RewardManager is a free data retrieval call binding the contract method 0x0f4ef8a6.
//
// Solidity: function rewardManager() view returns(address)
func (_Offers *OffersSession) RewardManager() (common.Address, error) {
	return _Offers.Contract.RewardManager(&_Offers.CallOpts)
}

// RewardManager is a free data retrieval call binding the contract method 0x0f4ef8a6.
//
// Solidity: function rewardManager() view returns(address)
func (_Offers *OffersCallerSession) RewardManager() (common.Address, error) {
	return _Offers.Contract.RewardManager(&_Offers.CallOpts)
}

// TimelockedCalls is a free data retrieval call binding the contract method 0x74e6310e.
//
// Solidity: function timelockedCalls(bytes4 selector) view returns(uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Offers *OffersCaller) TimelockedCalls(opts *bind.CallOpts, selector [4]byte) (struct {
	AllowedAfterTimestamp *big.Int
	EncodedCall           []byte
}, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "timelockedCalls", selector)

	outstruct := new(struct {
		AllowedAfterTimestamp *big.Int
		EncodedCall           []byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.AllowedAfterTimestamp = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.EncodedCall = *abi.ConvertType(out[1], new([]byte)).(*[]byte)

	return *outstruct, err

}

// TimelockedCalls is a free data retrieval call binding the contract method 0x74e6310e.
//
// Solidity: function timelockedCalls(bytes4 selector) view returns(uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Offers *OffersSession) TimelockedCalls(selector [4]byte) (struct {
	AllowedAfterTimestamp *big.Int
	EncodedCall           []byte
}, error) {
	return _Offers.Contract.TimelockedCalls(&_Offers.CallOpts, selector)
}

// TimelockedCalls is a free data retrieval call binding the contract method 0x74e6310e.
//
// Solidity: function timelockedCalls(bytes4 selector) view returns(uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Offers *OffersCallerSession) TimelockedCalls(selector [4]byte) (struct {
	AllowedAfterTimestamp *big.Int
	EncodedCall           []byte
}, error) {
	return _Offers.Contract.TimelockedCalls(&_Offers.CallOpts, selector)
}

// TotalInflationAuthorizedWei is a free data retrieval call binding the contract method 0xd0c1c393.
//
// Solidity: function totalInflationAuthorizedWei() view returns(uint256)
func (_Offers *OffersCaller) TotalInflationAuthorizedWei(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "totalInflationAuthorizedWei")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalInflationAuthorizedWei is a free data retrieval call binding the contract method 0xd0c1c393.
//
// Solidity: function totalInflationAuthorizedWei() view returns(uint256)
func (_Offers *OffersSession) TotalInflationAuthorizedWei() (*big.Int, error) {
	return _Offers.Contract.TotalInflationAuthorizedWei(&_Offers.CallOpts)
}

// TotalInflationAuthorizedWei is a free data retrieval call binding the contract method 0xd0c1c393.
//
// Solidity: function totalInflationAuthorizedWei() view returns(uint256)
func (_Offers *OffersCallerSession) TotalInflationAuthorizedWei() (*big.Int, error) {
	return _Offers.Contract.TotalInflationAuthorizedWei(&_Offers.CallOpts)
}

// TotalInflationReceivedWei is a free data retrieval call binding the contract method 0xa5555aea.
//
// Solidity: function totalInflationReceivedWei() view returns(uint256)
func (_Offers *OffersCaller) TotalInflationReceivedWei(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "totalInflationReceivedWei")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalInflationReceivedWei is a free data retrieval call binding the contract method 0xa5555aea.
//
// Solidity: function totalInflationReceivedWei() view returns(uint256)
func (_Offers *OffersSession) TotalInflationReceivedWei() (*big.Int, error) {
	return _Offers.Contract.TotalInflationReceivedWei(&_Offers.CallOpts)
}

// TotalInflationReceivedWei is a free data retrieval call binding the contract method 0xa5555aea.
//
// Solidity: function totalInflationReceivedWei() view returns(uint256)
func (_Offers *OffersCallerSession) TotalInflationReceivedWei() (*big.Int, error) {
	return _Offers.Contract.TotalInflationReceivedWei(&_Offers.CallOpts)
}

// TotalInflationRewardsOfferedWei is a free data retrieval call binding the contract method 0xbd76b69c.
//
// Solidity: function totalInflationRewardsOfferedWei() view returns(uint256)
func (_Offers *OffersCaller) TotalInflationRewardsOfferedWei(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Offers.contract.Call(opts, &out, "totalInflationRewardsOfferedWei")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalInflationRewardsOfferedWei is a free data retrieval call binding the contract method 0xbd76b69c.
//
// Solidity: function totalInflationRewardsOfferedWei() view returns(uint256)
func (_Offers *OffersSession) TotalInflationRewardsOfferedWei() (*big.Int, error) {
	return _Offers.Contract.TotalInflationRewardsOfferedWei(&_Offers.CallOpts)
}

// TotalInflationRewardsOfferedWei is a free data retrieval call binding the contract method 0xbd76b69c.
//
// Solidity: function totalInflationRewardsOfferedWei() view returns(uint256)
func (_Offers *OffersCallerSession) TotalInflationRewardsOfferedWei() (*big.Int, error) {
	return _Offers.Contract.TotalInflationRewardsOfferedWei(&_Offers.CallOpts)
}

// CancelGovernanceCall is a paid mutator transaction binding the contract method 0x67fc4029.
//
// Solidity: function cancelGovernanceCall(bytes4 _selector) returns()
func (_Offers *OffersTransactor) CancelGovernanceCall(opts *bind.TransactOpts, _selector [4]byte) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "cancelGovernanceCall", _selector)
}

// CancelGovernanceCall is a paid mutator transaction binding the contract method 0x67fc4029.
//
// Solidity: function cancelGovernanceCall(bytes4 _selector) returns()
func (_Offers *OffersSession) CancelGovernanceCall(_selector [4]byte) (*types.Transaction, error) {
	return _Offers.Contract.CancelGovernanceCall(&_Offers.TransactOpts, _selector)
}

// CancelGovernanceCall is a paid mutator transaction binding the contract method 0x67fc4029.
//
// Solidity: function cancelGovernanceCall(bytes4 _selector) returns()
func (_Offers *OffersTransactorSession) CancelGovernanceCall(_selector [4]byte) (*types.Transaction, error) {
	return _Offers.Contract.CancelGovernanceCall(&_Offers.TransactOpts, _selector)
}

// ExecuteGovernanceCall is a paid mutator transaction binding the contract method 0x5ff27079.
//
// Solidity: function executeGovernanceCall(bytes4 _selector) returns()
func (_Offers *OffersTransactor) ExecuteGovernanceCall(opts *bind.TransactOpts, _selector [4]byte) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "executeGovernanceCall", _selector)
}

// ExecuteGovernanceCall is a paid mutator transaction binding the contract method 0x5ff27079.
//
// Solidity: function executeGovernanceCall(bytes4 _selector) returns()
func (_Offers *OffersSession) ExecuteGovernanceCall(_selector [4]byte) (*types.Transaction, error) {
	return _Offers.Contract.ExecuteGovernanceCall(&_Offers.TransactOpts, _selector)
}

// ExecuteGovernanceCall is a paid mutator transaction binding the contract method 0x5ff27079.
//
// Solidity: function executeGovernanceCall(bytes4 _selector) returns()
func (_Offers *OffersTransactorSession) ExecuteGovernanceCall(_selector [4]byte) (*types.Transaction, error) {
	return _Offers.Contract.ExecuteGovernanceCall(&_Offers.TransactOpts, _selector)
}

// Initialise is a paid mutator transaction binding the contract method 0xef88bf13.
//
// Solidity: function initialise(address _governanceSettings, address _initialGovernance) returns()
func (_Offers *OffersTransactor) Initialise(opts *bind.TransactOpts, _governanceSettings common.Address, _initialGovernance common.Address) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "initialise", _governanceSettings, _initialGovernance)
}

// Initialise is a paid mutator transaction binding the contract method 0xef88bf13.
//
// Solidity: function initialise(address _governanceSettings, address _initialGovernance) returns()
func (_Offers *OffersSession) Initialise(_governanceSettings common.Address, _initialGovernance common.Address) (*types.Transaction, error) {
	return _Offers.Contract.Initialise(&_Offers.TransactOpts, _governanceSettings, _initialGovernance)
}

// Initialise is a paid mutator transaction binding the contract method 0xef88bf13.
//
// Solidity: function initialise(address _governanceSettings, address _initialGovernance) returns()
func (_Offers *OffersTransactorSession) Initialise(_governanceSettings common.Address, _initialGovernance common.Address) (*types.Transaction, error) {
	return _Offers.Contract.Initialise(&_Offers.TransactOpts, _governanceSettings, _initialGovernance)
}

// OfferRewards is a paid mutator transaction binding the contract method 0xb75a424f.
//
// Solidity: function offerRewards(uint24 _nextRewardEpochId, (uint120,bytes21,uint16,uint24,uint24,address)[] _offers) payable returns()
func (_Offers *OffersTransactor) OfferRewards(opts *bind.TransactOpts, _nextRewardEpochId *big.Int, _offers []IFtsoRewardOffersManagerOffer) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "offerRewards", _nextRewardEpochId, _offers)
}

// OfferRewards is a paid mutator transaction binding the contract method 0xb75a424f.
//
// Solidity: function offerRewards(uint24 _nextRewardEpochId, (uint120,bytes21,uint16,uint24,uint24,address)[] _offers) payable returns()
func (_Offers *OffersSession) OfferRewards(_nextRewardEpochId *big.Int, _offers []IFtsoRewardOffersManagerOffer) (*types.Transaction, error) {
	return _Offers.Contract.OfferRewards(&_Offers.TransactOpts, _nextRewardEpochId, _offers)
}

// OfferRewards is a paid mutator transaction binding the contract method 0xb75a424f.
//
// Solidity: function offerRewards(uint24 _nextRewardEpochId, (uint120,bytes21,uint16,uint24,uint24,address)[] _offers) payable returns()
func (_Offers *OffersTransactorSession) OfferRewards(_nextRewardEpochId *big.Int, _offers []IFtsoRewardOffersManagerOffer) (*types.Transaction, error) {
	return _Offers.Contract.OfferRewards(&_Offers.TransactOpts, _nextRewardEpochId, _offers)
}

// ReceiveInflation is a paid mutator transaction binding the contract method 0x06201f1d.
//
// Solidity: function receiveInflation() payable returns()
func (_Offers *OffersTransactor) ReceiveInflation(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "receiveInflation")
}

// ReceiveInflation is a paid mutator transaction binding the contract method 0x06201f1d.
//
// Solidity: function receiveInflation() payable returns()
func (_Offers *OffersSession) ReceiveInflation() (*types.Transaction, error) {
	return _Offers.Contract.ReceiveInflation(&_Offers.TransactOpts)
}

// ReceiveInflation is a paid mutator transaction binding the contract method 0x06201f1d.
//
// Solidity: function receiveInflation() payable returns()
func (_Offers *OffersTransactorSession) ReceiveInflation() (*types.Transaction, error) {
	return _Offers.Contract.ReceiveInflation(&_Offers.TransactOpts)
}

// SetDailyAuthorizedInflation is a paid mutator transaction binding the contract method 0xe2739563.
//
// Solidity: function setDailyAuthorizedInflation(uint256 _toAuthorizeWei) returns()
func (_Offers *OffersTransactor) SetDailyAuthorizedInflation(opts *bind.TransactOpts, _toAuthorizeWei *big.Int) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "setDailyAuthorizedInflation", _toAuthorizeWei)
}

// SetDailyAuthorizedInflation is a paid mutator transaction binding the contract method 0xe2739563.
//
// Solidity: function setDailyAuthorizedInflation(uint256 _toAuthorizeWei) returns()
func (_Offers *OffersSession) SetDailyAuthorizedInflation(_toAuthorizeWei *big.Int) (*types.Transaction, error) {
	return _Offers.Contract.SetDailyAuthorizedInflation(&_Offers.TransactOpts, _toAuthorizeWei)
}

// SetDailyAuthorizedInflation is a paid mutator transaction binding the contract method 0xe2739563.
//
// Solidity: function setDailyAuthorizedInflation(uint256 _toAuthorizeWei) returns()
func (_Offers *OffersTransactorSession) SetDailyAuthorizedInflation(_toAuthorizeWei *big.Int) (*types.Transaction, error) {
	return _Offers.Contract.SetDailyAuthorizedInflation(&_Offers.TransactOpts, _toAuthorizeWei)
}

// SetMinimalRewardsOfferValue is a paid mutator transaction binding the contract method 0xc9f19dd2.
//
// Solidity: function setMinimalRewardsOfferValue(uint128 _minimalRewardsOfferValueWei) returns()
func (_Offers *OffersTransactor) SetMinimalRewardsOfferValue(opts *bind.TransactOpts, _minimalRewardsOfferValueWei *big.Int) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "setMinimalRewardsOfferValue", _minimalRewardsOfferValueWei)
}

// SetMinimalRewardsOfferValue is a paid mutator transaction binding the contract method 0xc9f19dd2.
//
// Solidity: function setMinimalRewardsOfferValue(uint128 _minimalRewardsOfferValueWei) returns()
func (_Offers *OffersSession) SetMinimalRewardsOfferValue(_minimalRewardsOfferValueWei *big.Int) (*types.Transaction, error) {
	return _Offers.Contract.SetMinimalRewardsOfferValue(&_Offers.TransactOpts, _minimalRewardsOfferValueWei)
}

// SetMinimalRewardsOfferValue is a paid mutator transaction binding the contract method 0xc9f19dd2.
//
// Solidity: function setMinimalRewardsOfferValue(uint128 _minimalRewardsOfferValueWei) returns()
func (_Offers *OffersTransactorSession) SetMinimalRewardsOfferValue(_minimalRewardsOfferValueWei *big.Int) (*types.Transaction, error) {
	return _Offers.Contract.SetMinimalRewardsOfferValue(&_Offers.TransactOpts, _minimalRewardsOfferValueWei)
}

// SwitchToProductionMode is a paid mutator transaction binding the contract method 0xf5a98383.
//
// Solidity: function switchToProductionMode() returns()
func (_Offers *OffersTransactor) SwitchToProductionMode(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "switchToProductionMode")
}

// SwitchToProductionMode is a paid mutator transaction binding the contract method 0xf5a98383.
//
// Solidity: function switchToProductionMode() returns()
func (_Offers *OffersSession) SwitchToProductionMode() (*types.Transaction, error) {
	return _Offers.Contract.SwitchToProductionMode(&_Offers.TransactOpts)
}

// SwitchToProductionMode is a paid mutator transaction binding the contract method 0xf5a98383.
//
// Solidity: function switchToProductionMode() returns()
func (_Offers *OffersTransactorSession) SwitchToProductionMode() (*types.Transaction, error) {
	return _Offers.Contract.SwitchToProductionMode(&_Offers.TransactOpts)
}

// TriggerRewardEpochSwitchover is a paid mutator transaction binding the contract method 0x91f25679.
//
// Solidity: function triggerRewardEpochSwitchover(uint24 _currentRewardEpochId, uint64 _currentRewardEpochExpectedEndTs, uint64 _rewardEpochDurationSeconds) returns()
func (_Offers *OffersTransactor) TriggerRewardEpochSwitchover(opts *bind.TransactOpts, _currentRewardEpochId *big.Int, _currentRewardEpochExpectedEndTs uint64, _rewardEpochDurationSeconds uint64) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "triggerRewardEpochSwitchover", _currentRewardEpochId, _currentRewardEpochExpectedEndTs, _rewardEpochDurationSeconds)
}

// TriggerRewardEpochSwitchover is a paid mutator transaction binding the contract method 0x91f25679.
//
// Solidity: function triggerRewardEpochSwitchover(uint24 _currentRewardEpochId, uint64 _currentRewardEpochExpectedEndTs, uint64 _rewardEpochDurationSeconds) returns()
func (_Offers *OffersSession) TriggerRewardEpochSwitchover(_currentRewardEpochId *big.Int, _currentRewardEpochExpectedEndTs uint64, _rewardEpochDurationSeconds uint64) (*types.Transaction, error) {
	return _Offers.Contract.TriggerRewardEpochSwitchover(&_Offers.TransactOpts, _currentRewardEpochId, _currentRewardEpochExpectedEndTs, _rewardEpochDurationSeconds)
}

// TriggerRewardEpochSwitchover is a paid mutator transaction binding the contract method 0x91f25679.
//
// Solidity: function triggerRewardEpochSwitchover(uint24 _currentRewardEpochId, uint64 _currentRewardEpochExpectedEndTs, uint64 _rewardEpochDurationSeconds) returns()
func (_Offers *OffersTransactorSession) TriggerRewardEpochSwitchover(_currentRewardEpochId *big.Int, _currentRewardEpochExpectedEndTs uint64, _rewardEpochDurationSeconds uint64) (*types.Transaction, error) {
	return _Offers.Contract.TriggerRewardEpochSwitchover(&_Offers.TransactOpts, _currentRewardEpochId, _currentRewardEpochExpectedEndTs, _rewardEpochDurationSeconds)
}

// UpdateContractAddresses is a paid mutator transaction binding the contract method 0xb00c0b76.
//
// Solidity: function updateContractAddresses(bytes32[] _contractNameHashes, address[] _contractAddresses) returns()
func (_Offers *OffersTransactor) UpdateContractAddresses(opts *bind.TransactOpts, _contractNameHashes [][32]byte, _contractAddresses []common.Address) (*types.Transaction, error) {
	return _Offers.contract.Transact(opts, "updateContractAddresses", _contractNameHashes, _contractAddresses)
}

// UpdateContractAddresses is a paid mutator transaction binding the contract method 0xb00c0b76.
//
// Solidity: function updateContractAddresses(bytes32[] _contractNameHashes, address[] _contractAddresses) returns()
func (_Offers *OffersSession) UpdateContractAddresses(_contractNameHashes [][32]byte, _contractAddresses []common.Address) (*types.Transaction, error) {
	return _Offers.Contract.UpdateContractAddresses(&_Offers.TransactOpts, _contractNameHashes, _contractAddresses)
}

// UpdateContractAddresses is a paid mutator transaction binding the contract method 0xb00c0b76.
//
// Solidity: function updateContractAddresses(bytes32[] _contractNameHashes, address[] _contractAddresses) returns()
func (_Offers *OffersTransactorSession) UpdateContractAddresses(_contractNameHashes [][32]byte, _contractAddresses []common.Address) (*types.Transaction, error) {
	return _Offers.Contract.UpdateContractAddresses(&_Offers.TransactOpts, _contractNameHashes, _contractAddresses)
}

// OffersDailyAuthorizedInflationSetIterator is returned from FilterDailyAuthorizedInflationSet and is used to iterate over the raw logs and unpacked data for DailyAuthorizedInflationSet events raised by the Offers contract.
type OffersDailyAuthorizedInflationSetIterator struct {
	Event *OffersDailyAuthorizedInflationSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersDailyAuthorizedInflationSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersDailyAuthorizedInflationSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersDailyAuthorizedInflationSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersDailyAuthorizedInflationSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersDailyAuthorizedInflationSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersDailyAuthorizedInflationSet represents a DailyAuthorizedInflationSet event raised by the Offers contract.
type OffersDailyAuthorizedInflationSet struct {
	AuthorizedAmountWei *big.Int
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterDailyAuthorizedInflationSet is a free log retrieval operation binding the contract event 0x187f32a0f765499f15b3bb52ed0aebf6015059f230f2ace7e701e60a47669595.
//
// Solidity: event DailyAuthorizedInflationSet(uint256 authorizedAmountWei)
func (_Offers *OffersFilterer) FilterDailyAuthorizedInflationSet(opts *bind.FilterOpts) (*OffersDailyAuthorizedInflationSetIterator, error) {

	logs, sub, err := _Offers.contract.FilterLogs(opts, "DailyAuthorizedInflationSet")
	if err != nil {
		return nil, err
	}
	return &OffersDailyAuthorizedInflationSetIterator{contract: _Offers.contract, event: "DailyAuthorizedInflationSet", logs: logs, sub: sub}, nil
}

// WatchDailyAuthorizedInflationSet is a free log subscription operation binding the contract event 0x187f32a0f765499f15b3bb52ed0aebf6015059f230f2ace7e701e60a47669595.
//
// Solidity: event DailyAuthorizedInflationSet(uint256 authorizedAmountWei)
func (_Offers *OffersFilterer) WatchDailyAuthorizedInflationSet(opts *bind.WatchOpts, sink chan<- *OffersDailyAuthorizedInflationSet) (event.Subscription, error) {

	logs, sub, err := _Offers.contract.WatchLogs(opts, "DailyAuthorizedInflationSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersDailyAuthorizedInflationSet)
				if err := _Offers.contract.UnpackLog(event, "DailyAuthorizedInflationSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDailyAuthorizedInflationSet is a log parse operation binding the contract event 0x187f32a0f765499f15b3bb52ed0aebf6015059f230f2ace7e701e60a47669595.
//
// Solidity: event DailyAuthorizedInflationSet(uint256 authorizedAmountWei)
func (_Offers *OffersFilterer) ParseDailyAuthorizedInflationSet(log types.Log) (*OffersDailyAuthorizedInflationSet, error) {
	event := new(OffersDailyAuthorizedInflationSet)
	if err := _Offers.contract.UnpackLog(event, "DailyAuthorizedInflationSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OffersGovernanceCallTimelockedIterator is returned from FilterGovernanceCallTimelocked and is used to iterate over the raw logs and unpacked data for GovernanceCallTimelocked events raised by the Offers contract.
type OffersGovernanceCallTimelockedIterator struct {
	Event *OffersGovernanceCallTimelocked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersGovernanceCallTimelockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersGovernanceCallTimelocked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersGovernanceCallTimelocked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersGovernanceCallTimelockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersGovernanceCallTimelockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersGovernanceCallTimelocked represents a GovernanceCallTimelocked event raised by the Offers contract.
type OffersGovernanceCallTimelocked struct {
	Selector              [4]byte
	AllowedAfterTimestamp *big.Int
	EncodedCall           []byte
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterGovernanceCallTimelocked is a free log retrieval operation binding the contract event 0xed948300a3694aa01d4a6b258bfd664350193d770c0b51f8387277f6d83ea3b6.
//
// Solidity: event GovernanceCallTimelocked(bytes4 selector, uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Offers *OffersFilterer) FilterGovernanceCallTimelocked(opts *bind.FilterOpts) (*OffersGovernanceCallTimelockedIterator, error) {

	logs, sub, err := _Offers.contract.FilterLogs(opts, "GovernanceCallTimelocked")
	if err != nil {
		return nil, err
	}
	return &OffersGovernanceCallTimelockedIterator{contract: _Offers.contract, event: "GovernanceCallTimelocked", logs: logs, sub: sub}, nil
}

// WatchGovernanceCallTimelocked is a free log subscription operation binding the contract event 0xed948300a3694aa01d4a6b258bfd664350193d770c0b51f8387277f6d83ea3b6.
//
// Solidity: event GovernanceCallTimelocked(bytes4 selector, uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Offers *OffersFilterer) WatchGovernanceCallTimelocked(opts *bind.WatchOpts, sink chan<- *OffersGovernanceCallTimelocked) (event.Subscription, error) {

	logs, sub, err := _Offers.contract.WatchLogs(opts, "GovernanceCallTimelocked")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersGovernanceCallTimelocked)
				if err := _Offers.contract.UnpackLog(event, "GovernanceCallTimelocked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseGovernanceCallTimelocked is a log parse operation binding the contract event 0xed948300a3694aa01d4a6b258bfd664350193d770c0b51f8387277f6d83ea3b6.
//
// Solidity: event GovernanceCallTimelocked(bytes4 selector, uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Offers *OffersFilterer) ParseGovernanceCallTimelocked(log types.Log) (*OffersGovernanceCallTimelocked, error) {
	event := new(OffersGovernanceCallTimelocked)
	if err := _Offers.contract.UnpackLog(event, "GovernanceCallTimelocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OffersGovernanceInitialisedIterator is returned from FilterGovernanceInitialised and is used to iterate over the raw logs and unpacked data for GovernanceInitialised events raised by the Offers contract.
type OffersGovernanceInitialisedIterator struct {
	Event *OffersGovernanceInitialised // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersGovernanceInitialisedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersGovernanceInitialised)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersGovernanceInitialised)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersGovernanceInitialisedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersGovernanceInitialisedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersGovernanceInitialised represents a GovernanceInitialised event raised by the Offers contract.
type OffersGovernanceInitialised struct {
	InitialGovernance common.Address
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterGovernanceInitialised is a free log retrieval operation binding the contract event 0x9789733827840833afc031fb2ef9ab6894271f77bad2085687cf4ae5c7bee4db.
//
// Solidity: event GovernanceInitialised(address initialGovernance)
func (_Offers *OffersFilterer) FilterGovernanceInitialised(opts *bind.FilterOpts) (*OffersGovernanceInitialisedIterator, error) {

	logs, sub, err := _Offers.contract.FilterLogs(opts, "GovernanceInitialised")
	if err != nil {
		return nil, err
	}
	return &OffersGovernanceInitialisedIterator{contract: _Offers.contract, event: "GovernanceInitialised", logs: logs, sub: sub}, nil
}

// WatchGovernanceInitialised is a free log subscription operation binding the contract event 0x9789733827840833afc031fb2ef9ab6894271f77bad2085687cf4ae5c7bee4db.
//
// Solidity: event GovernanceInitialised(address initialGovernance)
func (_Offers *OffersFilterer) WatchGovernanceInitialised(opts *bind.WatchOpts, sink chan<- *OffersGovernanceInitialised) (event.Subscription, error) {

	logs, sub, err := _Offers.contract.WatchLogs(opts, "GovernanceInitialised")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersGovernanceInitialised)
				if err := _Offers.contract.UnpackLog(event, "GovernanceInitialised", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseGovernanceInitialised is a log parse operation binding the contract event 0x9789733827840833afc031fb2ef9ab6894271f77bad2085687cf4ae5c7bee4db.
//
// Solidity: event GovernanceInitialised(address initialGovernance)
func (_Offers *OffersFilterer) ParseGovernanceInitialised(log types.Log) (*OffersGovernanceInitialised, error) {
	event := new(OffersGovernanceInitialised)
	if err := _Offers.contract.UnpackLog(event, "GovernanceInitialised", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OffersGovernedProductionModeEnteredIterator is returned from FilterGovernedProductionModeEntered and is used to iterate over the raw logs and unpacked data for GovernedProductionModeEntered events raised by the Offers contract.
type OffersGovernedProductionModeEnteredIterator struct {
	Event *OffersGovernedProductionModeEntered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersGovernedProductionModeEnteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersGovernedProductionModeEntered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersGovernedProductionModeEntered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersGovernedProductionModeEnteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersGovernedProductionModeEnteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersGovernedProductionModeEntered represents a GovernedProductionModeEntered event raised by the Offers contract.
type OffersGovernedProductionModeEntered struct {
	GovernanceSettings common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterGovernedProductionModeEntered is a free log retrieval operation binding the contract event 0x83af113638b5422f9e977cebc0aaf0eaf2188eb9a8baae7f9d46c42b33a1560c.
//
// Solidity: event GovernedProductionModeEntered(address governanceSettings)
func (_Offers *OffersFilterer) FilterGovernedProductionModeEntered(opts *bind.FilterOpts) (*OffersGovernedProductionModeEnteredIterator, error) {

	logs, sub, err := _Offers.contract.FilterLogs(opts, "GovernedProductionModeEntered")
	if err != nil {
		return nil, err
	}
	return &OffersGovernedProductionModeEnteredIterator{contract: _Offers.contract, event: "GovernedProductionModeEntered", logs: logs, sub: sub}, nil
}

// WatchGovernedProductionModeEntered is a free log subscription operation binding the contract event 0x83af113638b5422f9e977cebc0aaf0eaf2188eb9a8baae7f9d46c42b33a1560c.
//
// Solidity: event GovernedProductionModeEntered(address governanceSettings)
func (_Offers *OffersFilterer) WatchGovernedProductionModeEntered(opts *bind.WatchOpts, sink chan<- *OffersGovernedProductionModeEntered) (event.Subscription, error) {

	logs, sub, err := _Offers.contract.WatchLogs(opts, "GovernedProductionModeEntered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersGovernedProductionModeEntered)
				if err := _Offers.contract.UnpackLog(event, "GovernedProductionModeEntered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseGovernedProductionModeEntered is a log parse operation binding the contract event 0x83af113638b5422f9e977cebc0aaf0eaf2188eb9a8baae7f9d46c42b33a1560c.
//
// Solidity: event GovernedProductionModeEntered(address governanceSettings)
func (_Offers *OffersFilterer) ParseGovernedProductionModeEntered(log types.Log) (*OffersGovernedProductionModeEntered, error) {
	event := new(OffersGovernedProductionModeEntered)
	if err := _Offers.contract.UnpackLog(event, "GovernedProductionModeEntered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OffersInflationReceivedIterator is returned from FilterInflationReceived and is used to iterate over the raw logs and unpacked data for InflationReceived events raised by the Offers contract.
type OffersInflationReceivedIterator struct {
	Event *OffersInflationReceived // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersInflationReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersInflationReceived)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersInflationReceived)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersInflationReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersInflationReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersInflationReceived represents a InflationReceived event raised by the Offers contract.
type OffersInflationReceived struct {
	AmountReceivedWei *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterInflationReceived is a free log retrieval operation binding the contract event 0x95c4e29cc99bc027cfc3cd719d6fd973d5f0317061885fbb322b9b17d8d35d37.
//
// Solidity: event InflationReceived(uint256 amountReceivedWei)
func (_Offers *OffersFilterer) FilterInflationReceived(opts *bind.FilterOpts) (*OffersInflationReceivedIterator, error) {

	logs, sub, err := _Offers.contract.FilterLogs(opts, "InflationReceived")
	if err != nil {
		return nil, err
	}
	return &OffersInflationReceivedIterator{contract: _Offers.contract, event: "InflationReceived", logs: logs, sub: sub}, nil
}

// WatchInflationReceived is a free log subscription operation binding the contract event 0x95c4e29cc99bc027cfc3cd719d6fd973d5f0317061885fbb322b9b17d8d35d37.
//
// Solidity: event InflationReceived(uint256 amountReceivedWei)
func (_Offers *OffersFilterer) WatchInflationReceived(opts *bind.WatchOpts, sink chan<- *OffersInflationReceived) (event.Subscription, error) {

	logs, sub, err := _Offers.contract.WatchLogs(opts, "InflationReceived")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersInflationReceived)
				if err := _Offers.contract.UnpackLog(event, "InflationReceived", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInflationReceived is a log parse operation binding the contract event 0x95c4e29cc99bc027cfc3cd719d6fd973d5f0317061885fbb322b9b17d8d35d37.
//
// Solidity: event InflationReceived(uint256 amountReceivedWei)
func (_Offers *OffersFilterer) ParseInflationReceived(log types.Log) (*OffersInflationReceived, error) {
	event := new(OffersInflationReceived)
	if err := _Offers.contract.UnpackLog(event, "InflationReceived", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OffersInflationRewardsOfferedIterator is returned from FilterInflationRewardsOffered and is used to iterate over the raw logs and unpacked data for InflationRewardsOffered events raised by the Offers contract.
type OffersInflationRewardsOfferedIterator struct {
	Event *OffersInflationRewardsOffered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersInflationRewardsOfferedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersInflationRewardsOffered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersInflationRewardsOffered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersInflationRewardsOfferedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersInflationRewardsOfferedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersInflationRewardsOffered represents a InflationRewardsOffered event raised by the Offers contract.
type OffersInflationRewardsOffered struct {
	RewardEpochId             *big.Int
	FeedIds                   []byte
	Decimals                  []byte
	Amount                    *big.Int
	MinRewardedTurnoutBIPS    uint16
	PrimaryBandRewardSharePPM *big.Int
	SecondaryBandWidthPPMs    []byte
	Mode                      uint16
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterInflationRewardsOffered is a free log retrieval operation binding the contract event 0x01070f0e535c0d3077d9ca64b3122869a1897ff5c7711ba33b4db1fb9fd70cfc.
//
// Solidity: event InflationRewardsOffered(uint24 indexed rewardEpochId, bytes feedIds, bytes decimals, uint256 amount, uint16 minRewardedTurnoutBIPS, uint24 primaryBandRewardSharePPM, bytes secondaryBandWidthPPMs, uint16 mode)
func (_Offers *OffersFilterer) FilterInflationRewardsOffered(opts *bind.FilterOpts, rewardEpochId []*big.Int) (*OffersInflationRewardsOfferedIterator, error) {

	var rewardEpochIdRule []interface{}
	for _, rewardEpochIdItem := range rewardEpochId {
		rewardEpochIdRule = append(rewardEpochIdRule, rewardEpochIdItem)
	}

	logs, sub, err := _Offers.contract.FilterLogs(opts, "InflationRewardsOffered", rewardEpochIdRule)
	if err != nil {
		return nil, err
	}
	return &OffersInflationRewardsOfferedIterator{contract: _Offers.contract, event: "InflationRewardsOffered", logs: logs, sub: sub}, nil
}

// WatchInflationRewardsOffered is a free log subscription operation binding the contract event 0x01070f0e535c0d3077d9ca64b3122869a1897ff5c7711ba33b4db1fb9fd70cfc.
//
// Solidity: event InflationRewardsOffered(uint24 indexed rewardEpochId, bytes feedIds, bytes decimals, uint256 amount, uint16 minRewardedTurnoutBIPS, uint24 primaryBandRewardSharePPM, bytes secondaryBandWidthPPMs, uint16 mode)
func (_Offers *OffersFilterer) WatchInflationRewardsOffered(opts *bind.WatchOpts, sink chan<- *OffersInflationRewardsOffered, rewardEpochId []*big.Int) (event.Subscription, error) {

	var rewardEpochIdRule []interface{}
	for _, rewardEpochIdItem := range rewardEpochId {
		rewardEpochIdRule = append(rewardEpochIdRule, rewardEpochIdItem)
	}

	logs, sub, err := _Offers.contract.WatchLogs(opts, "InflationRewardsOffered", rewardEpochIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersInflationRewardsOffered)
				if err := _Offers.contract.UnpackLog(event, "InflationRewardsOffered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInflationRewardsOffered is a log parse operation binding the contract event 0x01070f0e535c0d3077d9ca64b3122869a1897ff5c7711ba33b4db1fb9fd70cfc.
//
// Solidity: event InflationRewardsOffered(uint24 indexed rewardEpochId, bytes feedIds, bytes decimals, uint256 amount, uint16 minRewardedTurnoutBIPS, uint24 primaryBandRewardSharePPM, bytes secondaryBandWidthPPMs, uint16 mode)
func (_Offers *OffersFilterer) ParseInflationRewardsOffered(log types.Log) (*OffersInflationRewardsOffered, error) {
	event := new(OffersInflationRewardsOffered)
	if err := _Offers.contract.UnpackLog(event, "InflationRewardsOffered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OffersMinimalRewardsOfferValueSetIterator is returned from FilterMinimalRewardsOfferValueSet and is used to iterate over the raw logs and unpacked data for MinimalRewardsOfferValueSet events raised by the Offers contract.
type OffersMinimalRewardsOfferValueSetIterator struct {
	Event *OffersMinimalRewardsOfferValueSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersMinimalRewardsOfferValueSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersMinimalRewardsOfferValueSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersMinimalRewardsOfferValueSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersMinimalRewardsOfferValueSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersMinimalRewardsOfferValueSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersMinimalRewardsOfferValueSet represents a MinimalRewardsOfferValueSet event raised by the Offers contract.
type OffersMinimalRewardsOfferValueSet struct {
	ValueWei *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMinimalRewardsOfferValueSet is a free log retrieval operation binding the contract event 0xbaedeb3cf70a35fbe2d8dcb5306bc177734e0c529004225a1526766096d6f165.
//
// Solidity: event MinimalRewardsOfferValueSet(uint256 valueWei)
func (_Offers *OffersFilterer) FilterMinimalRewardsOfferValueSet(opts *bind.FilterOpts) (*OffersMinimalRewardsOfferValueSetIterator, error) {

	logs, sub, err := _Offers.contract.FilterLogs(opts, "MinimalRewardsOfferValueSet")
	if err != nil {
		return nil, err
	}
	return &OffersMinimalRewardsOfferValueSetIterator{contract: _Offers.contract, event: "MinimalRewardsOfferValueSet", logs: logs, sub: sub}, nil
}

// WatchMinimalRewardsOfferValueSet is a free log subscription operation binding the contract event 0xbaedeb3cf70a35fbe2d8dcb5306bc177734e0c529004225a1526766096d6f165.
//
// Solidity: event MinimalRewardsOfferValueSet(uint256 valueWei)
func (_Offers *OffersFilterer) WatchMinimalRewardsOfferValueSet(opts *bind.WatchOpts, sink chan<- *OffersMinimalRewardsOfferValueSet) (event.Subscription, error) {

	logs, sub, err := _Offers.contract.WatchLogs(opts, "MinimalRewardsOfferValueSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersMinimalRewardsOfferValueSet)
				if err := _Offers.contract.UnpackLog(event, "MinimalRewardsOfferValueSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMinimalRewardsOfferValueSet is a log parse operation binding the contract event 0xbaedeb3cf70a35fbe2d8dcb5306bc177734e0c529004225a1526766096d6f165.
//
// Solidity: event MinimalRewardsOfferValueSet(uint256 valueWei)
func (_Offers *OffersFilterer) ParseMinimalRewardsOfferValueSet(log types.Log) (*OffersMinimalRewardsOfferValueSet, error) {
	event := new(OffersMinimalRewardsOfferValueSet)
	if err := _Offers.contract.UnpackLog(event, "MinimalRewardsOfferValueSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OffersRewardsOfferedIterator is returned from FilterRewardsOffered and is used to iterate over the raw logs and unpacked data for RewardsOffered events raised by the Offers contract.
type OffersRewardsOfferedIterator struct {
	Event *OffersRewardsOffered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersRewardsOfferedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersRewardsOffered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersRewardsOffered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersRewardsOfferedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersRewardsOfferedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersRewardsOffered represents a RewardsOffered event raised by the Offers contract.
type OffersRewardsOffered struct {
	RewardEpochId             *big.Int
	FeedId                    [21]byte
	Decimals                  int8
	Amount                    *big.Int
	MinRewardedTurnoutBIPS    uint16
	PrimaryBandRewardSharePPM *big.Int
	SecondaryBandWidthPPM     *big.Int
	ClaimBackAddress          common.Address
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterRewardsOffered is a free log retrieval operation binding the contract event 0x0b348981891869f97551a071f05fba72c5efadafb062152f1ef417a25264d69a.
//
// Solidity: event RewardsOffered(uint24 indexed rewardEpochId, bytes21 feedId, int8 decimals, uint256 amount, uint16 minRewardedTurnoutBIPS, uint24 primaryBandRewardSharePPM, uint24 secondaryBandWidthPPM, address claimBackAddress)
func (_Offers *OffersFilterer) FilterRewardsOffered(opts *bind.FilterOpts, rewardEpochId []*big.Int) (*OffersRewardsOfferedIterator, error) {

	var rewardEpochIdRule []interface{}
	for _, rewardEpochIdItem := range rewardEpochId {
		rewardEpochIdRule = append(rewardEpochIdRule, rewardEpochIdItem)
	}

	logs, sub, err := _Offers.contract.FilterLogs(opts, "RewardsOffered", rewardEpochIdRule)
	if err != nil {
		return nil, err
	}
	return &OffersRewardsOfferedIterator{contract: _Offers.contract, event: "RewardsOffered", logs: logs, sub: sub}, nil
}

// WatchRewardsOffered is a free log subscription operation binding the contract event 0x0b348981891869f97551a071f05fba72c5efadafb062152f1ef417a25264d69a.
//
// Solidity: event RewardsOffered(uint24 indexed rewardEpochId, bytes21 feedId, int8 decimals, uint256 amount, uint16 minRewardedTurnoutBIPS, uint24 primaryBandRewardSharePPM, uint24 secondaryBandWidthPPM, address claimBackAddress)
func (_Offers *OffersFilterer) WatchRewardsOffered(opts *bind.WatchOpts, sink chan<- *OffersRewardsOffered, rewardEpochId []*big.Int) (event.Subscription, error) {

	var rewardEpochIdRule []interface{}
	for _, rewardEpochIdItem := range rewardEpochId {
		rewardEpochIdRule = append(rewardEpochIdRule, rewardEpochIdItem)
	}

	logs, sub, err := _Offers.contract.WatchLogs(opts, "RewardsOffered", rewardEpochIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersRewardsOffered)
				if err := _Offers.contract.UnpackLog(event, "RewardsOffered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRewardsOffered is a log parse operation binding the contract event 0x0b348981891869f97551a071f05fba72c5efadafb062152f1ef417a25264d69a.
//
// Solidity: event RewardsOffered(uint24 indexed rewardEpochId, bytes21 feedId, int8 decimals, uint256 amount, uint16 minRewardedTurnoutBIPS, uint24 primaryBandRewardSharePPM, uint24 secondaryBandWidthPPM, address claimBackAddress)
func (_Offers *OffersFilterer) ParseRewardsOffered(log types.Log) (*OffersRewardsOffered, error) {
	event := new(OffersRewardsOffered)
	if err := _Offers.contract.UnpackLog(event, "RewardsOffered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OffersTimelockedGovernanceCallCanceledIterator is returned from FilterTimelockedGovernanceCallCanceled and is used to iterate over the raw logs and unpacked data for TimelockedGovernanceCallCanceled events raised by the Offers contract.
type OffersTimelockedGovernanceCallCanceledIterator struct {
	Event *OffersTimelockedGovernanceCallCanceled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersTimelockedGovernanceCallCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersTimelockedGovernanceCallCanceled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersTimelockedGovernanceCallCanceled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersTimelockedGovernanceCallCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersTimelockedGovernanceCallCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersTimelockedGovernanceCallCanceled represents a TimelockedGovernanceCallCanceled event raised by the Offers contract.
type OffersTimelockedGovernanceCallCanceled struct {
	Selector  [4]byte
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTimelockedGovernanceCallCanceled is a free log retrieval operation binding the contract event 0x7735b2391c38a81419c513e30ca578db7158eadd7101511b23e221c654d19cf8.
//
// Solidity: event TimelockedGovernanceCallCanceled(bytes4 selector, uint256 timestamp)
func (_Offers *OffersFilterer) FilterTimelockedGovernanceCallCanceled(opts *bind.FilterOpts) (*OffersTimelockedGovernanceCallCanceledIterator, error) {

	logs, sub, err := _Offers.contract.FilterLogs(opts, "TimelockedGovernanceCallCanceled")
	if err != nil {
		return nil, err
	}
	return &OffersTimelockedGovernanceCallCanceledIterator{contract: _Offers.contract, event: "TimelockedGovernanceCallCanceled", logs: logs, sub: sub}, nil
}

// WatchTimelockedGovernanceCallCanceled is a free log subscription operation binding the contract event 0x7735b2391c38a81419c513e30ca578db7158eadd7101511b23e221c654d19cf8.
//
// Solidity: event TimelockedGovernanceCallCanceled(bytes4 selector, uint256 timestamp)
func (_Offers *OffersFilterer) WatchTimelockedGovernanceCallCanceled(opts *bind.WatchOpts, sink chan<- *OffersTimelockedGovernanceCallCanceled) (event.Subscription, error) {

	logs, sub, err := _Offers.contract.WatchLogs(opts, "TimelockedGovernanceCallCanceled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersTimelockedGovernanceCallCanceled)
				if err := _Offers.contract.UnpackLog(event, "TimelockedGovernanceCallCanceled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTimelockedGovernanceCallCanceled is a log parse operation binding the contract event 0x7735b2391c38a81419c513e30ca578db7158eadd7101511b23e221c654d19cf8.
//
// Solidity: event TimelockedGovernanceCallCanceled(bytes4 selector, uint256 timestamp)
func (_Offers *OffersFilterer) ParseTimelockedGovernanceCallCanceled(log types.Log) (*OffersTimelockedGovernanceCallCanceled, error) {
	event := new(OffersTimelockedGovernanceCallCanceled)
	if err := _Offers.contract.UnpackLog(event, "TimelockedGovernanceCallCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OffersTimelockedGovernanceCallExecutedIterator is returned from FilterTimelockedGovernanceCallExecuted and is used to iterate over the raw logs and unpacked data for TimelockedGovernanceCallExecuted events raised by the Offers contract.
type OffersTimelockedGovernanceCallExecutedIterator struct {
	Event *OffersTimelockedGovernanceCallExecuted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OffersTimelockedGovernanceCallExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OffersTimelockedGovernanceCallExecuted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OffersTimelockedGovernanceCallExecuted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OffersTimelockedGovernanceCallExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OffersTimelockedGovernanceCallExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OffersTimelockedGovernanceCallExecuted represents a TimelockedGovernanceCallExecuted event raised by the Offers contract.
type OffersTimelockedGovernanceCallExecuted struct {
	Selector  [4]byte
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTimelockedGovernanceCallExecuted is a free log retrieval operation binding the contract event 0xa7326b57fc9cfe267aaea5e7f0b01757154d265620a0585819416ee9ddd2c438.
//
// Solidity: event TimelockedGovernanceCallExecuted(bytes4 selector, uint256 timestamp)
func (_Offers *OffersFilterer) FilterTimelockedGovernanceCallExecuted(opts *bind.FilterOpts) (*OffersTimelockedGovernanceCallExecutedIterator, error) {

	logs, sub, err := _Offers.contract.FilterLogs(opts, "TimelockedGovernanceCallExecuted")
	if err != nil {
		return nil, err
	}
	return &OffersTimelockedGovernanceCallExecutedIterator{contract: _Offers.contract, event: "TimelockedGovernanceCallExecuted", logs: logs, sub: sub}, nil
}

// WatchTimelockedGovernanceCallExecuted is a free log subscription operation binding the contract event 0xa7326b57fc9cfe267aaea5e7f0b01757154d265620a0585819416ee9ddd2c438.
//
// Solidity: event TimelockedGovernanceCallExecuted(bytes4 selector, uint256 timestamp)
func (_Offers *OffersFilterer) WatchTimelockedGovernanceCallExecuted(opts *bind.WatchOpts, sink chan<- *OffersTimelockedGovernanceCallExecuted) (event.Subscription, error) {

	logs, sub, err := _Offers.contract.WatchLogs(opts, "TimelockedGovernanceCallExecuted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OffersTimelockedGovernanceCallExecuted)
				if err := _Offers.contract.UnpackLog(event, "TimelockedGovernanceCallExecuted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTimelockedGovernanceCallExecuted is a log parse operation binding the contract event 0xa7326b57fc9cfe267aaea5e7f0b01757154d265620a0585819416ee9ddd2c438.
//
// Solidity: event TimelockedGovernanceCallExecuted(bytes4 selector, uint256 timestamp)
func (_Offers *OffersFilterer) ParseTimelockedGovernanceCallExecuted(log types.Log) (*OffersTimelockedGovernanceCallExecuted, error) {
	event := new(OffersTimelockedGovernanceCallExecuted)
	if err := _Offers.contract.UnpackLog(event, "TimelockedGovernanceCallExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package calculator

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

// CalculatorMetaData contains all meta data concerning the Calculator contract.
var CalculatorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIGovernanceSettings\",\"name\":\"_governanceSettings\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialGovernance\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_addressUpdater\",\"type\":\"address\"},{\"internalType\":\"uint24\",\"name\":\"_wNatCapPPM\",\"type\":\"uint24\"},{\"internalType\":\"uint64\",\"name\":\"_signingPolicySignNonPunishableDurationSeconds\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"_signingPolicySignNonPunishableDurationBlocks\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"_signingPolicySignNoRewardsDurationBlocks\",\"type\":\"uint64\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"allowedAfterTimestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"encodedCall\",\"type\":\"bytes\"}],\"name\":\"GovernanceCallTimelocked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"initialGovernance\",\"type\":\"address\"}],\"name\":\"GovernanceInitialised\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"governanceSettings\",\"type\":\"address\"}],\"name\":\"GovernedProductionModeEntered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"TimelockedGovernanceCallCanceled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"TimelockedGovernanceCallExecuted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"voter\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint24\",\"name\":\"rewardEpochId\",\"type\":\"uint24\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"delegationAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"delegationFeeBIPS\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"wNatWeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"wNatCappedWeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes20[]\",\"name\":\"nodeIds\",\"type\":\"bytes20[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"nodeWeights\",\"type\":\"uint256[]\"}],\"name\":\"VoterRegistrationInfo\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint24\",\"name\":\"_rewardEpochId\",\"type\":\"uint24\"},{\"internalType\":\"address\",\"name\":\"_voter\",\"type\":\"address\"}],\"name\":\"calculateBurnFactorPPM\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_voter\",\"type\":\"address\"},{\"internalType\":\"uint24\",\"name\":\"_rewardEpochId\",\"type\":\"uint24\"},{\"internalType\":\"uint256\",\"name\":\"_votePowerBlockNumber\",\"type\":\"uint256\"}],\"name\":\"calculateRegistrationWeight\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"_registrationWeight\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"_selector\",\"type\":\"bytes4\"}],\"name\":\"cancelGovernanceCall\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"enablePChainStakeMirror\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"entityManager\",\"outputs\":[{\"internalType\":\"contractIIEntityManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"_selector\",\"type\":\"bytes4\"}],\"name\":\"executeGovernanceCall\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"flareSystemsManager\",\"outputs\":[{\"internalType\":\"contractIIFlareSystemsManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAddressUpdater\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"_addressUpdater\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceSettings\",\"outputs\":[{\"internalType\":\"contractIGovernanceSettings\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIGovernanceSettings\",\"name\":\"_governanceSettings\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialGovernance\",\"type\":\"address\"}],\"name\":\"initialise\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"isExecutor\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pChainStakeMirror\",\"outputs\":[{\"internalType\":\"contractIPChainStakeMirror\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pChainStakeMirrorEnabled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"productionMode\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint24\",\"name\":\"_wNatCapPPM\",\"type\":\"uint24\"}],\"name\":\"setWNatCapPPM\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"signingPolicySignNoRewardsDurationBlocks\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"signingPolicySignNonPunishableDurationBlocks\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"signingPolicySignNonPunishableDurationSeconds\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_x\",\"type\":\"uint256\"}],\"name\":\"sqrt\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"switchToProductionMode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"}],\"name\":\"timelockedCalls\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"allowedAfterTimestamp\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"encodedCall\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"_contractNameHashes\",\"type\":\"bytes32[]\"},{\"internalType\":\"address[]\",\"name\":\"_contractAddresses\",\"type\":\"address[]\"}],\"name\":\"updateContractAddresses\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"voterRegistry\",\"outputs\":[{\"internalType\":\"contractIVoterRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"wNat\",\"outputs\":[{\"internalType\":\"contractIWNat\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"wNatCapPPM\",\"outputs\":[{\"internalType\":\"uint24\",\"name\":\"\",\"type\":\"uint24\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"wNatDelegationFee\",\"outputs\":[{\"internalType\":\"contractIWNatDelegationFee\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// CalculatorABI is the input ABI used to generate the binding from.
// Deprecated: Use CalculatorMetaData.ABI instead.
var CalculatorABI = CalculatorMetaData.ABI

// Calculator is an auto generated Go binding around an Ethereum contract.
type Calculator struct {
	CalculatorCaller     // Read-only binding to the contract
	CalculatorTransactor // Write-only binding to the contract
	CalculatorFilterer   // Log filterer for contract events
}

// CalculatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type CalculatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CalculatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CalculatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CalculatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CalculatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CalculatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CalculatorSession struct {
	Contract     *Calculator       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CalculatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CalculatorCallerSession struct {
	Contract *CalculatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// CalculatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CalculatorTransactorSession struct {
	Contract     *CalculatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// CalculatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type CalculatorRaw struct {
	Contract *Calculator // Generic contract binding to access the raw methods on
}

// CalculatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CalculatorCallerRaw struct {
	Contract *CalculatorCaller // Generic read-only contract binding to access the raw methods on
}

// CalculatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CalculatorTransactorRaw struct {
	Contract *CalculatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCalculator creates a new instance of Calculator, bound to a specific deployed contract.
func NewCalculator(address common.Address, backend bind.ContractBackend) (*Calculator, error) {
	contract, err := bindCalculator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Calculator{CalculatorCaller: CalculatorCaller{contract: contract}, CalculatorTransactor: CalculatorTransactor{contract: contract}, CalculatorFilterer: CalculatorFilterer{contract: contract}}, nil
}

// NewCalculatorCaller creates a new read-only instance of Calculator, bound to a specific deployed contract.
func NewCalculatorCaller(address common.Address, caller bind.ContractCaller) (*CalculatorCaller, error) {
	contract, err := bindCalculator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CalculatorCaller{contract: contract}, nil
}

// NewCalculatorTransactor creates a new write-only instance of Calculator, bound to a specific deployed contract.
func NewCalculatorTransactor(address common.Address, transactor bind.ContractTransactor) (*CalculatorTransactor, error) {
	contract, err := bindCalculator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CalculatorTransactor{contract: contract}, nil
}

// NewCalculatorFilterer creates a new log filterer instance of Calculator, bound to a specific deployed contract.
func NewCalculatorFilterer(address common.Address, filterer bind.ContractFilterer) (*CalculatorFilterer, error) {
	contract, err := bindCalculator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CalculatorFilterer{contract: contract}, nil
}

// bindCalculator binds a generic wrapper to an already deployed contract.
func bindCalculator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CalculatorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Calculator *CalculatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Calculator.Contract.CalculatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Calculator *CalculatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Calculator.Contract.CalculatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Calculator *CalculatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Calculator.Contract.CalculatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Calculator *CalculatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Calculator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Calculator *CalculatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Calculator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Calculator *CalculatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Calculator.Contract.contract.Transact(opts, method, params...)
}

// CalculateBurnFactorPPM is a free data retrieval call binding the contract method 0x9350f57c.
//
// Solidity: function calculateBurnFactorPPM(uint24 _rewardEpochId, address _voter) view returns(uint256)
func (_Calculator *CalculatorCaller) CalculateBurnFactorPPM(opts *bind.CallOpts, _rewardEpochId *big.Int, _voter common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "calculateBurnFactorPPM", _rewardEpochId, _voter)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateBurnFactorPPM is a free data retrieval call binding the contract method 0x9350f57c.
//
// Solidity: function calculateBurnFactorPPM(uint24 _rewardEpochId, address _voter) view returns(uint256)
func (_Calculator *CalculatorSession) CalculateBurnFactorPPM(_rewardEpochId *big.Int, _voter common.Address) (*big.Int, error) {
	return _Calculator.Contract.CalculateBurnFactorPPM(&_Calculator.CallOpts, _rewardEpochId, _voter)
}

// CalculateBurnFactorPPM is a free data retrieval call binding the contract method 0x9350f57c.
//
// Solidity: function calculateBurnFactorPPM(uint24 _rewardEpochId, address _voter) view returns(uint256)
func (_Calculator *CalculatorCallerSession) CalculateBurnFactorPPM(_rewardEpochId *big.Int, _voter common.Address) (*big.Int, error) {
	return _Calculator.Contract.CalculateBurnFactorPPM(&_Calculator.CallOpts, _rewardEpochId, _voter)
}

// EntityManager is a free data retrieval call binding the contract method 0x50b1d61b.
//
// Solidity: function entityManager() view returns(address)
func (_Calculator *CalculatorCaller) EntityManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "entityManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EntityManager is a free data retrieval call binding the contract method 0x50b1d61b.
//
// Solidity: function entityManager() view returns(address)
func (_Calculator *CalculatorSession) EntityManager() (common.Address, error) {
	return _Calculator.Contract.EntityManager(&_Calculator.CallOpts)
}

// EntityManager is a free data retrieval call binding the contract method 0x50b1d61b.
//
// Solidity: function entityManager() view returns(address)
func (_Calculator *CalculatorCallerSession) EntityManager() (common.Address, error) {
	return _Calculator.Contract.EntityManager(&_Calculator.CallOpts)
}

// FlareSystemsManager is a free data retrieval call binding the contract method 0xfaae7fc9.
//
// Solidity: function flareSystemsManager() view returns(address)
func (_Calculator *CalculatorCaller) FlareSystemsManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "flareSystemsManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FlareSystemsManager is a free data retrieval call binding the contract method 0xfaae7fc9.
//
// Solidity: function flareSystemsManager() view returns(address)
func (_Calculator *CalculatorSession) FlareSystemsManager() (common.Address, error) {
	return _Calculator.Contract.FlareSystemsManager(&_Calculator.CallOpts)
}

// FlareSystemsManager is a free data retrieval call binding the contract method 0xfaae7fc9.
//
// Solidity: function flareSystemsManager() view returns(address)
func (_Calculator *CalculatorCallerSession) FlareSystemsManager() (common.Address, error) {
	return _Calculator.Contract.FlareSystemsManager(&_Calculator.CallOpts)
}

// GetAddressUpdater is a free data retrieval call binding the contract method 0x5267a15d.
//
// Solidity: function getAddressUpdater() view returns(address _addressUpdater)
func (_Calculator *CalculatorCaller) GetAddressUpdater(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "getAddressUpdater")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAddressUpdater is a free data retrieval call binding the contract method 0x5267a15d.
//
// Solidity: function getAddressUpdater() view returns(address _addressUpdater)
func (_Calculator *CalculatorSession) GetAddressUpdater() (common.Address, error) {
	return _Calculator.Contract.GetAddressUpdater(&_Calculator.CallOpts)
}

// GetAddressUpdater is a free data retrieval call binding the contract method 0x5267a15d.
//
// Solidity: function getAddressUpdater() view returns(address _addressUpdater)
func (_Calculator *CalculatorCallerSession) GetAddressUpdater() (common.Address, error) {
	return _Calculator.Contract.GetAddressUpdater(&_Calculator.CallOpts)
}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_Calculator *CalculatorCaller) Governance(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "governance")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_Calculator *CalculatorSession) Governance() (common.Address, error) {
	return _Calculator.Contract.Governance(&_Calculator.CallOpts)
}

// Governance is a free data retrieval call binding the contract method 0x5aa6e675.
//
// Solidity: function governance() view returns(address)
func (_Calculator *CalculatorCallerSession) Governance() (common.Address, error) {
	return _Calculator.Contract.Governance(&_Calculator.CallOpts)
}

// GovernanceSettings is a free data retrieval call binding the contract method 0x62354e03.
//
// Solidity: function governanceSettings() view returns(address)
func (_Calculator *CalculatorCaller) GovernanceSettings(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "governanceSettings")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GovernanceSettings is a free data retrieval call binding the contract method 0x62354e03.
//
// Solidity: function governanceSettings() view returns(address)
func (_Calculator *CalculatorSession) GovernanceSettings() (common.Address, error) {
	return _Calculator.Contract.GovernanceSettings(&_Calculator.CallOpts)
}

// GovernanceSettings is a free data retrieval call binding the contract method 0x62354e03.
//
// Solidity: function governanceSettings() view returns(address)
func (_Calculator *CalculatorCallerSession) GovernanceSettings() (common.Address, error) {
	return _Calculator.Contract.GovernanceSettings(&_Calculator.CallOpts)
}

// IsExecutor is a free data retrieval call binding the contract method 0xdebfda30.
//
// Solidity: function isExecutor(address _address) view returns(bool)
func (_Calculator *CalculatorCaller) IsExecutor(opts *bind.CallOpts, _address common.Address) (bool, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "isExecutor", _address)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsExecutor is a free data retrieval call binding the contract method 0xdebfda30.
//
// Solidity: function isExecutor(address _address) view returns(bool)
func (_Calculator *CalculatorSession) IsExecutor(_address common.Address) (bool, error) {
	return _Calculator.Contract.IsExecutor(&_Calculator.CallOpts, _address)
}

// IsExecutor is a free data retrieval call binding the contract method 0xdebfda30.
//
// Solidity: function isExecutor(address _address) view returns(bool)
func (_Calculator *CalculatorCallerSession) IsExecutor(_address common.Address) (bool, error) {
	return _Calculator.Contract.IsExecutor(&_Calculator.CallOpts, _address)
}

// PChainStakeMirror is a free data retrieval call binding the contract method 0x62d9c89a.
//
// Solidity: function pChainStakeMirror() view returns(address)
func (_Calculator *CalculatorCaller) PChainStakeMirror(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "pChainStakeMirror")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PChainStakeMirror is a free data retrieval call binding the contract method 0x62d9c89a.
//
// Solidity: function pChainStakeMirror() view returns(address)
func (_Calculator *CalculatorSession) PChainStakeMirror() (common.Address, error) {
	return _Calculator.Contract.PChainStakeMirror(&_Calculator.CallOpts)
}

// PChainStakeMirror is a free data retrieval call binding the contract method 0x62d9c89a.
//
// Solidity: function pChainStakeMirror() view returns(address)
func (_Calculator *CalculatorCallerSession) PChainStakeMirror() (common.Address, error) {
	return _Calculator.Contract.PChainStakeMirror(&_Calculator.CallOpts)
}

// PChainStakeMirrorEnabled is a free data retrieval call binding the contract method 0x7bf756c9.
//
// Solidity: function pChainStakeMirrorEnabled() view returns(bool)
func (_Calculator *CalculatorCaller) PChainStakeMirrorEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "pChainStakeMirrorEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// PChainStakeMirrorEnabled is a free data retrieval call binding the contract method 0x7bf756c9.
//
// Solidity: function pChainStakeMirrorEnabled() view returns(bool)
func (_Calculator *CalculatorSession) PChainStakeMirrorEnabled() (bool, error) {
	return _Calculator.Contract.PChainStakeMirrorEnabled(&_Calculator.CallOpts)
}

// PChainStakeMirrorEnabled is a free data retrieval call binding the contract method 0x7bf756c9.
//
// Solidity: function pChainStakeMirrorEnabled() view returns(bool)
func (_Calculator *CalculatorCallerSession) PChainStakeMirrorEnabled() (bool, error) {
	return _Calculator.Contract.PChainStakeMirrorEnabled(&_Calculator.CallOpts)
}

// ProductionMode is a free data retrieval call binding the contract method 0xe17f212e.
//
// Solidity: function productionMode() view returns(bool)
func (_Calculator *CalculatorCaller) ProductionMode(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "productionMode")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ProductionMode is a free data retrieval call binding the contract method 0xe17f212e.
//
// Solidity: function productionMode() view returns(bool)
func (_Calculator *CalculatorSession) ProductionMode() (bool, error) {
	return _Calculator.Contract.ProductionMode(&_Calculator.CallOpts)
}

// ProductionMode is a free data retrieval call binding the contract method 0xe17f212e.
//
// Solidity: function productionMode() view returns(bool)
func (_Calculator *CalculatorCallerSession) ProductionMode() (bool, error) {
	return _Calculator.Contract.ProductionMode(&_Calculator.CallOpts)
}

// SigningPolicySignNoRewardsDurationBlocks is a free data retrieval call binding the contract method 0x9dd1018c.
//
// Solidity: function signingPolicySignNoRewardsDurationBlocks() view returns(uint64)
func (_Calculator *CalculatorCaller) SigningPolicySignNoRewardsDurationBlocks(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "signingPolicySignNoRewardsDurationBlocks")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// SigningPolicySignNoRewardsDurationBlocks is a free data retrieval call binding the contract method 0x9dd1018c.
//
// Solidity: function signingPolicySignNoRewardsDurationBlocks() view returns(uint64)
func (_Calculator *CalculatorSession) SigningPolicySignNoRewardsDurationBlocks() (uint64, error) {
	return _Calculator.Contract.SigningPolicySignNoRewardsDurationBlocks(&_Calculator.CallOpts)
}

// SigningPolicySignNoRewardsDurationBlocks is a free data retrieval call binding the contract method 0x9dd1018c.
//
// Solidity: function signingPolicySignNoRewardsDurationBlocks() view returns(uint64)
func (_Calculator *CalculatorCallerSession) SigningPolicySignNoRewardsDurationBlocks() (uint64, error) {
	return _Calculator.Contract.SigningPolicySignNoRewardsDurationBlocks(&_Calculator.CallOpts)
}

// SigningPolicySignNonPunishableDurationBlocks is a free data retrieval call binding the contract method 0x87fd1ba1.
//
// Solidity: function signingPolicySignNonPunishableDurationBlocks() view returns(uint64)
func (_Calculator *CalculatorCaller) SigningPolicySignNonPunishableDurationBlocks(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "signingPolicySignNonPunishableDurationBlocks")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// SigningPolicySignNonPunishableDurationBlocks is a free data retrieval call binding the contract method 0x87fd1ba1.
//
// Solidity: function signingPolicySignNonPunishableDurationBlocks() view returns(uint64)
func (_Calculator *CalculatorSession) SigningPolicySignNonPunishableDurationBlocks() (uint64, error) {
	return _Calculator.Contract.SigningPolicySignNonPunishableDurationBlocks(&_Calculator.CallOpts)
}

// SigningPolicySignNonPunishableDurationBlocks is a free data retrieval call binding the contract method 0x87fd1ba1.
//
// Solidity: function signingPolicySignNonPunishableDurationBlocks() view returns(uint64)
func (_Calculator *CalculatorCallerSession) SigningPolicySignNonPunishableDurationBlocks() (uint64, error) {
	return _Calculator.Contract.SigningPolicySignNonPunishableDurationBlocks(&_Calculator.CallOpts)
}

// SigningPolicySignNonPunishableDurationSeconds is a free data retrieval call binding the contract method 0x96ca1472.
//
// Solidity: function signingPolicySignNonPunishableDurationSeconds() view returns(uint64)
func (_Calculator *CalculatorCaller) SigningPolicySignNonPunishableDurationSeconds(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "signingPolicySignNonPunishableDurationSeconds")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// SigningPolicySignNonPunishableDurationSeconds is a free data retrieval call binding the contract method 0x96ca1472.
//
// Solidity: function signingPolicySignNonPunishableDurationSeconds() view returns(uint64)
func (_Calculator *CalculatorSession) SigningPolicySignNonPunishableDurationSeconds() (uint64, error) {
	return _Calculator.Contract.SigningPolicySignNonPunishableDurationSeconds(&_Calculator.CallOpts)
}

// SigningPolicySignNonPunishableDurationSeconds is a free data retrieval call binding the contract method 0x96ca1472.
//
// Solidity: function signingPolicySignNonPunishableDurationSeconds() view returns(uint64)
func (_Calculator *CalculatorCallerSession) SigningPolicySignNonPunishableDurationSeconds() (uint64, error) {
	return _Calculator.Contract.SigningPolicySignNonPunishableDurationSeconds(&_Calculator.CallOpts)
}

// Sqrt is a free data retrieval call binding the contract method 0x677342ce.
//
// Solidity: function sqrt(uint256 _x) pure returns(uint128)
func (_Calculator *CalculatorCaller) Sqrt(opts *bind.CallOpts, _x *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "sqrt", _x)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Sqrt is a free data retrieval call binding the contract method 0x677342ce.
//
// Solidity: function sqrt(uint256 _x) pure returns(uint128)
func (_Calculator *CalculatorSession) Sqrt(_x *big.Int) (*big.Int, error) {
	return _Calculator.Contract.Sqrt(&_Calculator.CallOpts, _x)
}

// Sqrt is a free data retrieval call binding the contract method 0x677342ce.
//
// Solidity: function sqrt(uint256 _x) pure returns(uint128)
func (_Calculator *CalculatorCallerSession) Sqrt(_x *big.Int) (*big.Int, error) {
	return _Calculator.Contract.Sqrt(&_Calculator.CallOpts, _x)
}

// TimelockedCalls is a free data retrieval call binding the contract method 0x74e6310e.
//
// Solidity: function timelockedCalls(bytes4 selector) view returns(uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Calculator *CalculatorCaller) TimelockedCalls(opts *bind.CallOpts, selector [4]byte) (struct {
	AllowedAfterTimestamp *big.Int
	EncodedCall           []byte
}, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "timelockedCalls", selector)

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
func (_Calculator *CalculatorSession) TimelockedCalls(selector [4]byte) (struct {
	AllowedAfterTimestamp *big.Int
	EncodedCall           []byte
}, error) {
	return _Calculator.Contract.TimelockedCalls(&_Calculator.CallOpts, selector)
}

// TimelockedCalls is a free data retrieval call binding the contract method 0x74e6310e.
//
// Solidity: function timelockedCalls(bytes4 selector) view returns(uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Calculator *CalculatorCallerSession) TimelockedCalls(selector [4]byte) (struct {
	AllowedAfterTimestamp *big.Int
	EncodedCall           []byte
}, error) {
	return _Calculator.Contract.TimelockedCalls(&_Calculator.CallOpts, selector)
}

// VoterRegistry is a free data retrieval call binding the contract method 0xbe60040e.
//
// Solidity: function voterRegistry() view returns(address)
func (_Calculator *CalculatorCaller) VoterRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "voterRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VoterRegistry is a free data retrieval call binding the contract method 0xbe60040e.
//
// Solidity: function voterRegistry() view returns(address)
func (_Calculator *CalculatorSession) VoterRegistry() (common.Address, error) {
	return _Calculator.Contract.VoterRegistry(&_Calculator.CallOpts)
}

// VoterRegistry is a free data retrieval call binding the contract method 0xbe60040e.
//
// Solidity: function voterRegistry() view returns(address)
func (_Calculator *CalculatorCallerSession) VoterRegistry() (common.Address, error) {
	return _Calculator.Contract.VoterRegistry(&_Calculator.CallOpts)
}

// WNat is a free data retrieval call binding the contract method 0x9edbf007.
//
// Solidity: function wNat() view returns(address)
func (_Calculator *CalculatorCaller) WNat(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "wNat")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WNat is a free data retrieval call binding the contract method 0x9edbf007.
//
// Solidity: function wNat() view returns(address)
func (_Calculator *CalculatorSession) WNat() (common.Address, error) {
	return _Calculator.Contract.WNat(&_Calculator.CallOpts)
}

// WNat is a free data retrieval call binding the contract method 0x9edbf007.
//
// Solidity: function wNat() view returns(address)
func (_Calculator *CalculatorCallerSession) WNat() (common.Address, error) {
	return _Calculator.Contract.WNat(&_Calculator.CallOpts)
}

// WNatCapPPM is a free data retrieval call binding the contract method 0x5edf7596.
//
// Solidity: function wNatCapPPM() view returns(uint24)
func (_Calculator *CalculatorCaller) WNatCapPPM(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "wNatCapPPM")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WNatCapPPM is a free data retrieval call binding the contract method 0x5edf7596.
//
// Solidity: function wNatCapPPM() view returns(uint24)
func (_Calculator *CalculatorSession) WNatCapPPM() (*big.Int, error) {
	return _Calculator.Contract.WNatCapPPM(&_Calculator.CallOpts)
}

// WNatCapPPM is a free data retrieval call binding the contract method 0x5edf7596.
//
// Solidity: function wNatCapPPM() view returns(uint24)
func (_Calculator *CalculatorCallerSession) WNatCapPPM() (*big.Int, error) {
	return _Calculator.Contract.WNatCapPPM(&_Calculator.CallOpts)
}

// WNatDelegationFee is a free data retrieval call binding the contract method 0x87c5ab51.
//
// Solidity: function wNatDelegationFee() view returns(address)
func (_Calculator *CalculatorCaller) WNatDelegationFee(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Calculator.contract.Call(opts, &out, "wNatDelegationFee")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WNatDelegationFee is a free data retrieval call binding the contract method 0x87c5ab51.
//
// Solidity: function wNatDelegationFee() view returns(address)
func (_Calculator *CalculatorSession) WNatDelegationFee() (common.Address, error) {
	return _Calculator.Contract.WNatDelegationFee(&_Calculator.CallOpts)
}

// WNatDelegationFee is a free data retrieval call binding the contract method 0x87c5ab51.
//
// Solidity: function wNatDelegationFee() view returns(address)
func (_Calculator *CalculatorCallerSession) WNatDelegationFee() (common.Address, error) {
	return _Calculator.Contract.WNatDelegationFee(&_Calculator.CallOpts)
}

// CalculateRegistrationWeight is a paid mutator transaction binding the contract method 0xb65185d6.
//
// Solidity: function calculateRegistrationWeight(address _voter, uint24 _rewardEpochId, uint256 _votePowerBlockNumber) returns(uint256 _registrationWeight)
func (_Calculator *CalculatorTransactor) CalculateRegistrationWeight(opts *bind.TransactOpts, _voter common.Address, _rewardEpochId *big.Int, _votePowerBlockNumber *big.Int) (*types.Transaction, error) {
	return _Calculator.contract.Transact(opts, "calculateRegistrationWeight", _voter, _rewardEpochId, _votePowerBlockNumber)
}

// CalculateRegistrationWeight is a paid mutator transaction binding the contract method 0xb65185d6.
//
// Solidity: function calculateRegistrationWeight(address _voter, uint24 _rewardEpochId, uint256 _votePowerBlockNumber) returns(uint256 _registrationWeight)
func (_Calculator *CalculatorSession) CalculateRegistrationWeight(_voter common.Address, _rewardEpochId *big.Int, _votePowerBlockNumber *big.Int) (*types.Transaction, error) {
	return _Calculator.Contract.CalculateRegistrationWeight(&_Calculator.TransactOpts, _voter, _rewardEpochId, _votePowerBlockNumber)
}

// CalculateRegistrationWeight is a paid mutator transaction binding the contract method 0xb65185d6.
//
// Solidity: function calculateRegistrationWeight(address _voter, uint24 _rewardEpochId, uint256 _votePowerBlockNumber) returns(uint256 _registrationWeight)
func (_Calculator *CalculatorTransactorSession) CalculateRegistrationWeight(_voter common.Address, _rewardEpochId *big.Int, _votePowerBlockNumber *big.Int) (*types.Transaction, error) {
	return _Calculator.Contract.CalculateRegistrationWeight(&_Calculator.TransactOpts, _voter, _rewardEpochId, _votePowerBlockNumber)
}

// CancelGovernanceCall is a paid mutator transaction binding the contract method 0x67fc4029.
//
// Solidity: function cancelGovernanceCall(bytes4 _selector) returns()
func (_Calculator *CalculatorTransactor) CancelGovernanceCall(opts *bind.TransactOpts, _selector [4]byte) (*types.Transaction, error) {
	return _Calculator.contract.Transact(opts, "cancelGovernanceCall", _selector)
}

// CancelGovernanceCall is a paid mutator transaction binding the contract method 0x67fc4029.
//
// Solidity: function cancelGovernanceCall(bytes4 _selector) returns()
func (_Calculator *CalculatorSession) CancelGovernanceCall(_selector [4]byte) (*types.Transaction, error) {
	return _Calculator.Contract.CancelGovernanceCall(&_Calculator.TransactOpts, _selector)
}

// CancelGovernanceCall is a paid mutator transaction binding the contract method 0x67fc4029.
//
// Solidity: function cancelGovernanceCall(bytes4 _selector) returns()
func (_Calculator *CalculatorTransactorSession) CancelGovernanceCall(_selector [4]byte) (*types.Transaction, error) {
	return _Calculator.Contract.CancelGovernanceCall(&_Calculator.TransactOpts, _selector)
}

// EnablePChainStakeMirror is a paid mutator transaction binding the contract method 0xb006b4e3.
//
// Solidity: function enablePChainStakeMirror() returns()
func (_Calculator *CalculatorTransactor) EnablePChainStakeMirror(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Calculator.contract.Transact(opts, "enablePChainStakeMirror")
}

// EnablePChainStakeMirror is a paid mutator transaction binding the contract method 0xb006b4e3.
//
// Solidity: function enablePChainStakeMirror() returns()
func (_Calculator *CalculatorSession) EnablePChainStakeMirror() (*types.Transaction, error) {
	return _Calculator.Contract.EnablePChainStakeMirror(&_Calculator.TransactOpts)
}

// EnablePChainStakeMirror is a paid mutator transaction binding the contract method 0xb006b4e3.
//
// Solidity: function enablePChainStakeMirror() returns()
func (_Calculator *CalculatorTransactorSession) EnablePChainStakeMirror() (*types.Transaction, error) {
	return _Calculator.Contract.EnablePChainStakeMirror(&_Calculator.TransactOpts)
}

// ExecuteGovernanceCall is a paid mutator transaction binding the contract method 0x5ff27079.
//
// Solidity: function executeGovernanceCall(bytes4 _selector) returns()
func (_Calculator *CalculatorTransactor) ExecuteGovernanceCall(opts *bind.TransactOpts, _selector [4]byte) (*types.Transaction, error) {
	return _Calculator.contract.Transact(opts, "executeGovernanceCall", _selector)
}

// ExecuteGovernanceCall is a paid mutator transaction binding the contract method 0x5ff27079.
//
// Solidity: function executeGovernanceCall(bytes4 _selector) returns()
func (_Calculator *CalculatorSession) ExecuteGovernanceCall(_selector [4]byte) (*types.Transaction, error) {
	return _Calculator.Contract.ExecuteGovernanceCall(&_Calculator.TransactOpts, _selector)
}

// ExecuteGovernanceCall is a paid mutator transaction binding the contract method 0x5ff27079.
//
// Solidity: function executeGovernanceCall(bytes4 _selector) returns()
func (_Calculator *CalculatorTransactorSession) ExecuteGovernanceCall(_selector [4]byte) (*types.Transaction, error) {
	return _Calculator.Contract.ExecuteGovernanceCall(&_Calculator.TransactOpts, _selector)
}

// Initialise is a paid mutator transaction binding the contract method 0xef88bf13.
//
// Solidity: function initialise(address _governanceSettings, address _initialGovernance) returns()
func (_Calculator *CalculatorTransactor) Initialise(opts *bind.TransactOpts, _governanceSettings common.Address, _initialGovernance common.Address) (*types.Transaction, error) {
	return _Calculator.contract.Transact(opts, "initialise", _governanceSettings, _initialGovernance)
}

// Initialise is a paid mutator transaction binding the contract method 0xef88bf13.
//
// Solidity: function initialise(address _governanceSettings, address _initialGovernance) returns()
func (_Calculator *CalculatorSession) Initialise(_governanceSettings common.Address, _initialGovernance common.Address) (*types.Transaction, error) {
	return _Calculator.Contract.Initialise(&_Calculator.TransactOpts, _governanceSettings, _initialGovernance)
}

// Initialise is a paid mutator transaction binding the contract method 0xef88bf13.
//
// Solidity: function initialise(address _governanceSettings, address _initialGovernance) returns()
func (_Calculator *CalculatorTransactorSession) Initialise(_governanceSettings common.Address, _initialGovernance common.Address) (*types.Transaction, error) {
	return _Calculator.Contract.Initialise(&_Calculator.TransactOpts, _governanceSettings, _initialGovernance)
}

// SetWNatCapPPM is a paid mutator transaction binding the contract method 0x3d7cf608.
//
// Solidity: function setWNatCapPPM(uint24 _wNatCapPPM) returns()
func (_Calculator *CalculatorTransactor) SetWNatCapPPM(opts *bind.TransactOpts, _wNatCapPPM *big.Int) (*types.Transaction, error) {
	return _Calculator.contract.Transact(opts, "setWNatCapPPM", _wNatCapPPM)
}

// SetWNatCapPPM is a paid mutator transaction binding the contract method 0x3d7cf608.
//
// Solidity: function setWNatCapPPM(uint24 _wNatCapPPM) returns()
func (_Calculator *CalculatorSession) SetWNatCapPPM(_wNatCapPPM *big.Int) (*types.Transaction, error) {
	return _Calculator.Contract.SetWNatCapPPM(&_Calculator.TransactOpts, _wNatCapPPM)
}

// SetWNatCapPPM is a paid mutator transaction binding the contract method 0x3d7cf608.
//
// Solidity: function setWNatCapPPM(uint24 _wNatCapPPM) returns()
func (_Calculator *CalculatorTransactorSession) SetWNatCapPPM(_wNatCapPPM *big.Int) (*types.Transaction, error) {
	return _Calculator.Contract.SetWNatCapPPM(&_Calculator.TransactOpts, _wNatCapPPM)
}

// SwitchToProductionMode is a paid mutator transaction binding the contract method 0xf5a98383.
//
// Solidity: function switchToProductionMode() returns()
func (_Calculator *CalculatorTransactor) SwitchToProductionMode(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Calculator.contract.Transact(opts, "switchToProductionMode")
}

// SwitchToProductionMode is a paid mutator transaction binding the contract method 0xf5a98383.
//
// Solidity: function switchToProductionMode() returns()
func (_Calculator *CalculatorSession) SwitchToProductionMode() (*types.Transaction, error) {
	return _Calculator.Contract.SwitchToProductionMode(&_Calculator.TransactOpts)
}

// SwitchToProductionMode is a paid mutator transaction binding the contract method 0xf5a98383.
//
// Solidity: function switchToProductionMode() returns()
func (_Calculator *CalculatorTransactorSession) SwitchToProductionMode() (*types.Transaction, error) {
	return _Calculator.Contract.SwitchToProductionMode(&_Calculator.TransactOpts)
}

// UpdateContractAddresses is a paid mutator transaction binding the contract method 0xb00c0b76.
//
// Solidity: function updateContractAddresses(bytes32[] _contractNameHashes, address[] _contractAddresses) returns()
func (_Calculator *CalculatorTransactor) UpdateContractAddresses(opts *bind.TransactOpts, _contractNameHashes [][32]byte, _contractAddresses []common.Address) (*types.Transaction, error) {
	return _Calculator.contract.Transact(opts, "updateContractAddresses", _contractNameHashes, _contractAddresses)
}

// UpdateContractAddresses is a paid mutator transaction binding the contract method 0xb00c0b76.
//
// Solidity: function updateContractAddresses(bytes32[] _contractNameHashes, address[] _contractAddresses) returns()
func (_Calculator *CalculatorSession) UpdateContractAddresses(_contractNameHashes [][32]byte, _contractAddresses []common.Address) (*types.Transaction, error) {
	return _Calculator.Contract.UpdateContractAddresses(&_Calculator.TransactOpts, _contractNameHashes, _contractAddresses)
}

// UpdateContractAddresses is a paid mutator transaction binding the contract method 0xb00c0b76.
//
// Solidity: function updateContractAddresses(bytes32[] _contractNameHashes, address[] _contractAddresses) returns()
func (_Calculator *CalculatorTransactorSession) UpdateContractAddresses(_contractNameHashes [][32]byte, _contractAddresses []common.Address) (*types.Transaction, error) {
	return _Calculator.Contract.UpdateContractAddresses(&_Calculator.TransactOpts, _contractNameHashes, _contractAddresses)
}

// CalculatorGovernanceCallTimelockedIterator is returned from FilterGovernanceCallTimelocked and is used to iterate over the raw logs and unpacked data for GovernanceCallTimelocked events raised by the Calculator contract.
type CalculatorGovernanceCallTimelockedIterator struct {
	Event *CalculatorGovernanceCallTimelocked // Event containing the contract specifics and raw log

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
func (it *CalculatorGovernanceCallTimelockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CalculatorGovernanceCallTimelocked)
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
		it.Event = new(CalculatorGovernanceCallTimelocked)
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
func (it *CalculatorGovernanceCallTimelockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CalculatorGovernanceCallTimelockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CalculatorGovernanceCallTimelocked represents a GovernanceCallTimelocked event raised by the Calculator contract.
type CalculatorGovernanceCallTimelocked struct {
	Selector              [4]byte
	AllowedAfterTimestamp *big.Int
	EncodedCall           []byte
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterGovernanceCallTimelocked is a free log retrieval operation binding the contract event 0xed948300a3694aa01d4a6b258bfd664350193d770c0b51f8387277f6d83ea3b6.
//
// Solidity: event GovernanceCallTimelocked(bytes4 selector, uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Calculator *CalculatorFilterer) FilterGovernanceCallTimelocked(opts *bind.FilterOpts) (*CalculatorGovernanceCallTimelockedIterator, error) {

	logs, sub, err := _Calculator.contract.FilterLogs(opts, "GovernanceCallTimelocked")
	if err != nil {
		return nil, err
	}
	return &CalculatorGovernanceCallTimelockedIterator{contract: _Calculator.contract, event: "GovernanceCallTimelocked", logs: logs, sub: sub}, nil
}

// WatchGovernanceCallTimelocked is a free log subscription operation binding the contract event 0xed948300a3694aa01d4a6b258bfd664350193d770c0b51f8387277f6d83ea3b6.
//
// Solidity: event GovernanceCallTimelocked(bytes4 selector, uint256 allowedAfterTimestamp, bytes encodedCall)
func (_Calculator *CalculatorFilterer) WatchGovernanceCallTimelocked(opts *bind.WatchOpts, sink chan<- *CalculatorGovernanceCallTimelocked) (event.Subscription, error) {

	logs, sub, err := _Calculator.contract.WatchLogs(opts, "GovernanceCallTimelocked")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CalculatorGovernanceCallTimelocked)
				if err := _Calculator.contract.UnpackLog(event, "GovernanceCallTimelocked", log); err != nil {
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
func (_Calculator *CalculatorFilterer) ParseGovernanceCallTimelocked(log types.Log) (*CalculatorGovernanceCallTimelocked, error) {
	event := new(CalculatorGovernanceCallTimelocked)
	if err := _Calculator.contract.UnpackLog(event, "GovernanceCallTimelocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CalculatorGovernanceInitialisedIterator is returned from FilterGovernanceInitialised and is used to iterate over the raw logs and unpacked data for GovernanceInitialised events raised by the Calculator contract.
type CalculatorGovernanceInitialisedIterator struct {
	Event *CalculatorGovernanceInitialised // Event containing the contract specifics and raw log

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
func (it *CalculatorGovernanceInitialisedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CalculatorGovernanceInitialised)
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
		it.Event = new(CalculatorGovernanceInitialised)
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
func (it *CalculatorGovernanceInitialisedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CalculatorGovernanceInitialisedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CalculatorGovernanceInitialised represents a GovernanceInitialised event raised by the Calculator contract.
type CalculatorGovernanceInitialised struct {
	InitialGovernance common.Address
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterGovernanceInitialised is a free log retrieval operation binding the contract event 0x9789733827840833afc031fb2ef9ab6894271f77bad2085687cf4ae5c7bee4db.
//
// Solidity: event GovernanceInitialised(address initialGovernance)
func (_Calculator *CalculatorFilterer) FilterGovernanceInitialised(opts *bind.FilterOpts) (*CalculatorGovernanceInitialisedIterator, error) {

	logs, sub, err := _Calculator.contract.FilterLogs(opts, "GovernanceInitialised")
	if err != nil {
		return nil, err
	}
	return &CalculatorGovernanceInitialisedIterator{contract: _Calculator.contract, event: "GovernanceInitialised", logs: logs, sub: sub}, nil
}

// WatchGovernanceInitialised is a free log subscription operation binding the contract event 0x9789733827840833afc031fb2ef9ab6894271f77bad2085687cf4ae5c7bee4db.
//
// Solidity: event GovernanceInitialised(address initialGovernance)
func (_Calculator *CalculatorFilterer) WatchGovernanceInitialised(opts *bind.WatchOpts, sink chan<- *CalculatorGovernanceInitialised) (event.Subscription, error) {

	logs, sub, err := _Calculator.contract.WatchLogs(opts, "GovernanceInitialised")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CalculatorGovernanceInitialised)
				if err := _Calculator.contract.UnpackLog(event, "GovernanceInitialised", log); err != nil {
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
func (_Calculator *CalculatorFilterer) ParseGovernanceInitialised(log types.Log) (*CalculatorGovernanceInitialised, error) {
	event := new(CalculatorGovernanceInitialised)
	if err := _Calculator.contract.UnpackLog(event, "GovernanceInitialised", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CalculatorGovernedProductionModeEnteredIterator is returned from FilterGovernedProductionModeEntered and is used to iterate over the raw logs and unpacked data for GovernedProductionModeEntered events raised by the Calculator contract.
type CalculatorGovernedProductionModeEnteredIterator struct {
	Event *CalculatorGovernedProductionModeEntered // Event containing the contract specifics and raw log

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
func (it *CalculatorGovernedProductionModeEnteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CalculatorGovernedProductionModeEntered)
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
		it.Event = new(CalculatorGovernedProductionModeEntered)
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
func (it *CalculatorGovernedProductionModeEnteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CalculatorGovernedProductionModeEnteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CalculatorGovernedProductionModeEntered represents a GovernedProductionModeEntered event raised by the Calculator contract.
type CalculatorGovernedProductionModeEntered struct {
	GovernanceSettings common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterGovernedProductionModeEntered is a free log retrieval operation binding the contract event 0x83af113638b5422f9e977cebc0aaf0eaf2188eb9a8baae7f9d46c42b33a1560c.
//
// Solidity: event GovernedProductionModeEntered(address governanceSettings)
func (_Calculator *CalculatorFilterer) FilterGovernedProductionModeEntered(opts *bind.FilterOpts) (*CalculatorGovernedProductionModeEnteredIterator, error) {

	logs, sub, err := _Calculator.contract.FilterLogs(opts, "GovernedProductionModeEntered")
	if err != nil {
		return nil, err
	}
	return &CalculatorGovernedProductionModeEnteredIterator{contract: _Calculator.contract, event: "GovernedProductionModeEntered", logs: logs, sub: sub}, nil
}

// WatchGovernedProductionModeEntered is a free log subscription operation binding the contract event 0x83af113638b5422f9e977cebc0aaf0eaf2188eb9a8baae7f9d46c42b33a1560c.
//
// Solidity: event GovernedProductionModeEntered(address governanceSettings)
func (_Calculator *CalculatorFilterer) WatchGovernedProductionModeEntered(opts *bind.WatchOpts, sink chan<- *CalculatorGovernedProductionModeEntered) (event.Subscription, error) {

	logs, sub, err := _Calculator.contract.WatchLogs(opts, "GovernedProductionModeEntered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CalculatorGovernedProductionModeEntered)
				if err := _Calculator.contract.UnpackLog(event, "GovernedProductionModeEntered", log); err != nil {
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
func (_Calculator *CalculatorFilterer) ParseGovernedProductionModeEntered(log types.Log) (*CalculatorGovernedProductionModeEntered, error) {
	event := new(CalculatorGovernedProductionModeEntered)
	if err := _Calculator.contract.UnpackLog(event, "GovernedProductionModeEntered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CalculatorTimelockedGovernanceCallCanceledIterator is returned from FilterTimelockedGovernanceCallCanceled and is used to iterate over the raw logs and unpacked data for TimelockedGovernanceCallCanceled events raised by the Calculator contract.
type CalculatorTimelockedGovernanceCallCanceledIterator struct {
	Event *CalculatorTimelockedGovernanceCallCanceled // Event containing the contract specifics and raw log

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
func (it *CalculatorTimelockedGovernanceCallCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CalculatorTimelockedGovernanceCallCanceled)
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
		it.Event = new(CalculatorTimelockedGovernanceCallCanceled)
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
func (it *CalculatorTimelockedGovernanceCallCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CalculatorTimelockedGovernanceCallCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CalculatorTimelockedGovernanceCallCanceled represents a TimelockedGovernanceCallCanceled event raised by the Calculator contract.
type CalculatorTimelockedGovernanceCallCanceled struct {
	Selector  [4]byte
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTimelockedGovernanceCallCanceled is a free log retrieval operation binding the contract event 0x7735b2391c38a81419c513e30ca578db7158eadd7101511b23e221c654d19cf8.
//
// Solidity: event TimelockedGovernanceCallCanceled(bytes4 selector, uint256 timestamp)
func (_Calculator *CalculatorFilterer) FilterTimelockedGovernanceCallCanceled(opts *bind.FilterOpts) (*CalculatorTimelockedGovernanceCallCanceledIterator, error) {

	logs, sub, err := _Calculator.contract.FilterLogs(opts, "TimelockedGovernanceCallCanceled")
	if err != nil {
		return nil, err
	}
	return &CalculatorTimelockedGovernanceCallCanceledIterator{contract: _Calculator.contract, event: "TimelockedGovernanceCallCanceled", logs: logs, sub: sub}, nil
}

// WatchTimelockedGovernanceCallCanceled is a free log subscription operation binding the contract event 0x7735b2391c38a81419c513e30ca578db7158eadd7101511b23e221c654d19cf8.
//
// Solidity: event TimelockedGovernanceCallCanceled(bytes4 selector, uint256 timestamp)
func (_Calculator *CalculatorFilterer) WatchTimelockedGovernanceCallCanceled(opts *bind.WatchOpts, sink chan<- *CalculatorTimelockedGovernanceCallCanceled) (event.Subscription, error) {

	logs, sub, err := _Calculator.contract.WatchLogs(opts, "TimelockedGovernanceCallCanceled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CalculatorTimelockedGovernanceCallCanceled)
				if err := _Calculator.contract.UnpackLog(event, "TimelockedGovernanceCallCanceled", log); err != nil {
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
func (_Calculator *CalculatorFilterer) ParseTimelockedGovernanceCallCanceled(log types.Log) (*CalculatorTimelockedGovernanceCallCanceled, error) {
	event := new(CalculatorTimelockedGovernanceCallCanceled)
	if err := _Calculator.contract.UnpackLog(event, "TimelockedGovernanceCallCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CalculatorTimelockedGovernanceCallExecutedIterator is returned from FilterTimelockedGovernanceCallExecuted and is used to iterate over the raw logs and unpacked data for TimelockedGovernanceCallExecuted events raised by the Calculator contract.
type CalculatorTimelockedGovernanceCallExecutedIterator struct {
	Event *CalculatorTimelockedGovernanceCallExecuted // Event containing the contract specifics and raw log

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
func (it *CalculatorTimelockedGovernanceCallExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CalculatorTimelockedGovernanceCallExecuted)
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
		it.Event = new(CalculatorTimelockedGovernanceCallExecuted)
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
func (it *CalculatorTimelockedGovernanceCallExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CalculatorTimelockedGovernanceCallExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CalculatorTimelockedGovernanceCallExecuted represents a TimelockedGovernanceCallExecuted event raised by the Calculator contract.
type CalculatorTimelockedGovernanceCallExecuted struct {
	Selector  [4]byte
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTimelockedGovernanceCallExecuted is a free log retrieval operation binding the contract event 0xa7326b57fc9cfe267aaea5e7f0b01757154d265620a0585819416ee9ddd2c438.
//
// Solidity: event TimelockedGovernanceCallExecuted(bytes4 selector, uint256 timestamp)
func (_Calculator *CalculatorFilterer) FilterTimelockedGovernanceCallExecuted(opts *bind.FilterOpts) (*CalculatorTimelockedGovernanceCallExecutedIterator, error) {

	logs, sub, err := _Calculator.contract.FilterLogs(opts, "TimelockedGovernanceCallExecuted")
	if err != nil {
		return nil, err
	}
	return &CalculatorTimelockedGovernanceCallExecutedIterator{contract: _Calculator.contract, event: "TimelockedGovernanceCallExecuted", logs: logs, sub: sub}, nil
}

// WatchTimelockedGovernanceCallExecuted is a free log subscription operation binding the contract event 0xa7326b57fc9cfe267aaea5e7f0b01757154d265620a0585819416ee9ddd2c438.
//
// Solidity: event TimelockedGovernanceCallExecuted(bytes4 selector, uint256 timestamp)
func (_Calculator *CalculatorFilterer) WatchTimelockedGovernanceCallExecuted(opts *bind.WatchOpts, sink chan<- *CalculatorTimelockedGovernanceCallExecuted) (event.Subscription, error) {

	logs, sub, err := _Calculator.contract.WatchLogs(opts, "TimelockedGovernanceCallExecuted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CalculatorTimelockedGovernanceCallExecuted)
				if err := _Calculator.contract.UnpackLog(event, "TimelockedGovernanceCallExecuted", log); err != nil {
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
func (_Calculator *CalculatorFilterer) ParseTimelockedGovernanceCallExecuted(log types.Log) (*CalculatorTimelockedGovernanceCallExecuted, error) {
	event := new(CalculatorTimelockedGovernanceCallExecuted)
	if err := _Calculator.contract.UnpackLog(event, "TimelockedGovernanceCallExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CalculatorVoterRegistrationInfoIterator is returned from FilterVoterRegistrationInfo and is used to iterate over the raw logs and unpacked data for VoterRegistrationInfo events raised by the Calculator contract.
type CalculatorVoterRegistrationInfoIterator struct {
	Event *CalculatorVoterRegistrationInfo // Event containing the contract specifics and raw log

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
func (it *CalculatorVoterRegistrationInfoIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CalculatorVoterRegistrationInfo)
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
		it.Event = new(CalculatorVoterRegistrationInfo)
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
func (it *CalculatorVoterRegistrationInfoIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CalculatorVoterRegistrationInfoIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CalculatorVoterRegistrationInfo represents a VoterRegistrationInfo event raised by the Calculator contract.
type CalculatorVoterRegistrationInfo struct {
	Voter             common.Address
	RewardEpochId     *big.Int
	DelegationAddress common.Address
	DelegationFeeBIPS uint16
	WNatWeight        *big.Int
	WNatCappedWeight  *big.Int
	NodeIds           [][20]byte
	NodeWeights       []*big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterVoterRegistrationInfo is a free log retrieval operation binding the contract event 0x0ab8c10ef8cc6c5a591797917aba3b2688483b9ff7ab824f36b872a45b752e1f.
//
// Solidity: event VoterRegistrationInfo(address indexed voter, uint24 indexed rewardEpochId, address delegationAddress, uint16 delegationFeeBIPS, uint256 wNatWeight, uint256 wNatCappedWeight, bytes20[] nodeIds, uint256[] nodeWeights)
func (_Calculator *CalculatorFilterer) FilterVoterRegistrationInfo(opts *bind.FilterOpts, voter []common.Address, rewardEpochId []*big.Int) (*CalculatorVoterRegistrationInfoIterator, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}
	var rewardEpochIdRule []interface{}
	for _, rewardEpochIdItem := range rewardEpochId {
		rewardEpochIdRule = append(rewardEpochIdRule, rewardEpochIdItem)
	}

	logs, sub, err := _Calculator.contract.FilterLogs(opts, "VoterRegistrationInfo", voterRule, rewardEpochIdRule)
	if err != nil {
		return nil, err
	}
	return &CalculatorVoterRegistrationInfoIterator{contract: _Calculator.contract, event: "VoterRegistrationInfo", logs: logs, sub: sub}, nil
}

// WatchVoterRegistrationInfo is a free log subscription operation binding the contract event 0x0ab8c10ef8cc6c5a591797917aba3b2688483b9ff7ab824f36b872a45b752e1f.
//
// Solidity: event VoterRegistrationInfo(address indexed voter, uint24 indexed rewardEpochId, address delegationAddress, uint16 delegationFeeBIPS, uint256 wNatWeight, uint256 wNatCappedWeight, bytes20[] nodeIds, uint256[] nodeWeights)
func (_Calculator *CalculatorFilterer) WatchVoterRegistrationInfo(opts *bind.WatchOpts, sink chan<- *CalculatorVoterRegistrationInfo, voter []common.Address, rewardEpochId []*big.Int) (event.Subscription, error) {

	var voterRule []interface{}
	for _, voterItem := range voter {
		voterRule = append(voterRule, voterItem)
	}
	var rewardEpochIdRule []interface{}
	for _, rewardEpochIdItem := range rewardEpochId {
		rewardEpochIdRule = append(rewardEpochIdRule, rewardEpochIdItem)
	}

	logs, sub, err := _Calculator.contract.WatchLogs(opts, "VoterRegistrationInfo", voterRule, rewardEpochIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CalculatorVoterRegistrationInfo)
				if err := _Calculator.contract.UnpackLog(event, "VoterRegistrationInfo", log); err != nil {
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

// ParseVoterRegistrationInfo is a log parse operation binding the contract event 0x0ab8c10ef8cc6c5a591797917aba3b2688483b9ff7ab824f36b872a45b752e1f.
//
// Solidity: event VoterRegistrationInfo(address indexed voter, uint24 indexed rewardEpochId, address delegationAddress, uint16 delegationFeeBIPS, uint256 wNatWeight, uint256 wNatCappedWeight, bytes20[] nodeIds, uint256[] nodeWeights)
func (_Calculator *CalculatorFilterer) ParseVoterRegistrationInfo(log types.Log) (*CalculatorVoterRegistrationInfo, error) {
	event := new(CalculatorVoterRegistrationInfo)
	if err := _Calculator.contract.UnpackLog(event, "VoterRegistrationInfo", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

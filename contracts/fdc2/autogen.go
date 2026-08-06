// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package fdc2

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

// Fdc2MetaData contains all meta data concerning the Fdc2 contract.
var Fdc2MetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"instructionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"attestationType\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sourceId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"proofOwner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"claimBackAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"}],\"name\":\"AttestationRequested\",\"type\":\"event\"}]",
}

// Fdc2ABI is the input ABI used to generate the binding from.
// Deprecated: Use Fdc2MetaData.ABI instead.
var Fdc2ABI = Fdc2MetaData.ABI

// Fdc2 is an auto generated Go binding around an Ethereum contract.
type Fdc2 struct {
	Fdc2Caller     // Read-only binding to the contract
	Fdc2Transactor // Write-only binding to the contract
	Fdc2Filterer   // Log filterer for contract events
}

// Fdc2Caller is an auto generated read-only Go binding around an Ethereum contract.
type Fdc2Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Fdc2Transactor is an auto generated write-only Go binding around an Ethereum contract.
type Fdc2Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Fdc2Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Fdc2Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Fdc2Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Fdc2Session struct {
	Contract     *Fdc2             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Fdc2CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Fdc2CallerSession struct {
	Contract *Fdc2Caller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// Fdc2TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Fdc2TransactorSession struct {
	Contract     *Fdc2Transactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Fdc2Raw is an auto generated low-level Go binding around an Ethereum contract.
type Fdc2Raw struct {
	Contract *Fdc2 // Generic contract binding to access the raw methods on
}

// Fdc2CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Fdc2CallerRaw struct {
	Contract *Fdc2Caller // Generic read-only contract binding to access the raw methods on
}

// Fdc2TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Fdc2TransactorRaw struct {
	Contract *Fdc2Transactor // Generic write-only contract binding to access the raw methods on
}

// NewFdc2 creates a new instance of Fdc2, bound to a specific deployed contract.
func NewFdc2(address common.Address, backend bind.ContractBackend) (*Fdc2, error) {
	contract, err := bindFdc2(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Fdc2{Fdc2Caller: Fdc2Caller{contract: contract}, Fdc2Transactor: Fdc2Transactor{contract: contract}, Fdc2Filterer: Fdc2Filterer{contract: contract}}, nil
}

// NewFdc2Caller creates a new read-only instance of Fdc2, bound to a specific deployed contract.
func NewFdc2Caller(address common.Address, caller bind.ContractCaller) (*Fdc2Caller, error) {
	contract, err := bindFdc2(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Fdc2Caller{contract: contract}, nil
}

// NewFdc2Transactor creates a new write-only instance of Fdc2, bound to a specific deployed contract.
func NewFdc2Transactor(address common.Address, transactor bind.ContractTransactor) (*Fdc2Transactor, error) {
	contract, err := bindFdc2(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Fdc2Transactor{contract: contract}, nil
}

// NewFdc2Filterer creates a new log filterer instance of Fdc2, bound to a specific deployed contract.
func NewFdc2Filterer(address common.Address, filterer bind.ContractFilterer) (*Fdc2Filterer, error) {
	contract, err := bindFdc2(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Fdc2Filterer{contract: contract}, nil
}

// bindFdc2 binds a generic wrapper to an already deployed contract.
func bindFdc2(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := Fdc2MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Fdc2 *Fdc2Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Fdc2.Contract.Fdc2Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Fdc2 *Fdc2Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Fdc2.Contract.Fdc2Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Fdc2 *Fdc2Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Fdc2.Contract.Fdc2Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Fdc2 *Fdc2CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Fdc2.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Fdc2 *Fdc2TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Fdc2.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Fdc2 *Fdc2TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Fdc2.Contract.contract.Transact(opts, method, params...)
}

// Fdc2AttestationRequestedIterator is returned from FilterAttestationRequested and is used to iterate over the raw logs and unpacked data for AttestationRequested events raised by the Fdc2 contract.
type Fdc2AttestationRequestedIterator struct {
	Event *Fdc2AttestationRequested // Event containing the contract specifics and raw log

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
func (it *Fdc2AttestationRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Fdc2AttestationRequested)
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
		it.Event = new(Fdc2AttestationRequested)
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
func (it *Fdc2AttestationRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Fdc2AttestationRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Fdc2AttestationRequested represents a AttestationRequested event raised by the Fdc2 contract.
type Fdc2AttestationRequested struct {
	InstructionId    [32]byte
	AttestationType  [32]byte
	SourceId         [32]byte
	ProofOwner       common.Address
	ClaimBackAddress common.Address
	Fee              *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterAttestationRequested is a free log retrieval operation binding the contract event 0x57c4413905bb1b444f93a5eab5a942fae34c0fcaa1c25cc595ce0b990310f5de.
//
// Solidity: event AttestationRequested(bytes32 indexed instructionId, bytes32 indexed attestationType, bytes32 indexed sourceId, address proofOwner, address claimBackAddress, uint256 fee)
func (_Fdc2 *Fdc2Filterer) FilterAttestationRequested(opts *bind.FilterOpts, instructionId [][32]byte, attestationType [][32]byte, sourceId [][32]byte) (*Fdc2AttestationRequestedIterator, error) {

	var instructionIdRule []interface{}
	for _, instructionIdItem := range instructionId {
		instructionIdRule = append(instructionIdRule, instructionIdItem)
	}
	var attestationTypeRule []interface{}
	for _, attestationTypeItem := range attestationType {
		attestationTypeRule = append(attestationTypeRule, attestationTypeItem)
	}
	var sourceIdRule []interface{}
	for _, sourceIdItem := range sourceId {
		sourceIdRule = append(sourceIdRule, sourceIdItem)
	}

	logs, sub, err := _Fdc2.contract.FilterLogs(opts, "AttestationRequested", instructionIdRule, attestationTypeRule, sourceIdRule)
	if err != nil {
		return nil, err
	}
	return &Fdc2AttestationRequestedIterator{contract: _Fdc2.contract, event: "AttestationRequested", logs: logs, sub: sub}, nil
}

// WatchAttestationRequested is a free log subscription operation binding the contract event 0x57c4413905bb1b444f93a5eab5a942fae34c0fcaa1c25cc595ce0b990310f5de.
//
// Solidity: event AttestationRequested(bytes32 indexed instructionId, bytes32 indexed attestationType, bytes32 indexed sourceId, address proofOwner, address claimBackAddress, uint256 fee)
func (_Fdc2 *Fdc2Filterer) WatchAttestationRequested(opts *bind.WatchOpts, sink chan<- *Fdc2AttestationRequested, instructionId [][32]byte, attestationType [][32]byte, sourceId [][32]byte) (event.Subscription, error) {

	var instructionIdRule []interface{}
	for _, instructionIdItem := range instructionId {
		instructionIdRule = append(instructionIdRule, instructionIdItem)
	}
	var attestationTypeRule []interface{}
	for _, attestationTypeItem := range attestationType {
		attestationTypeRule = append(attestationTypeRule, attestationTypeItem)
	}
	var sourceIdRule []interface{}
	for _, sourceIdItem := range sourceId {
		sourceIdRule = append(sourceIdRule, sourceIdItem)
	}

	logs, sub, err := _Fdc2.contract.WatchLogs(opts, "AttestationRequested", instructionIdRule, attestationTypeRule, sourceIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Fdc2AttestationRequested)
				if err := _Fdc2.contract.UnpackLog(event, "AttestationRequested", log); err != nil {
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

// ParseAttestationRequested is a log parse operation binding the contract event 0x57c4413905bb1b444f93a5eab5a942fae34c0fcaa1c25cc595ce0b990310f5de.
//
// Solidity: event AttestationRequested(bytes32 indexed instructionId, bytes32 indexed attestationType, bytes32 indexed sourceId, address proofOwner, address claimBackAddress, uint256 fee)
func (_Fdc2 *Fdc2Filterer) ParseAttestationRequested(log types.Log) (*Fdc2AttestationRequested, error) {
	event := new(Fdc2AttestationRequested)
	if err := _Fdc2.contract.UnpackLog(event, "AttestationRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

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

// RISCVMetaData contains all meta data concerning the RISCV contract.
var RISCVMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_oracle\",\"type\":\"address\",\"internalType\":\"contractIPreimageOracle\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"oracle\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPreimageOracle\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"step\",\"inputs\":[{\"name\":\"_stateData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_proof\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_localContext\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"version\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"}]",
}

// RISCVABI is the input ABI used to generate the binding from.
// Deprecated: Use RISCVMetaData.ABI instead.
var RISCVABI = RISCVMetaData.ABI

// RISCV is an auto generated Go binding around an Ethereum contract.
type RISCV struct {
	RISCVCaller     // Read-only binding to the contract
	RISCVTransactor // Write-only binding to the contract
	RISCVFilterer   // Log filterer for contract events
}

// RISCVCaller is an auto generated read-only Go binding around an Ethereum contract.
type RISCVCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RISCVTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RISCVTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RISCVFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RISCVFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RISCVSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RISCVSession struct {
	Contract     *RISCV            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RISCVCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RISCVCallerSession struct {
	Contract *RISCVCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// RISCVTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RISCVTransactorSession struct {
	Contract     *RISCVTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RISCVRaw is an auto generated low-level Go binding around an Ethereum contract.
type RISCVRaw struct {
	Contract *RISCV // Generic contract binding to access the raw methods on
}

// RISCVCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RISCVCallerRaw struct {
	Contract *RISCVCaller // Generic read-only contract binding to access the raw methods on
}

// RISCVTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RISCVTransactorRaw struct {
	Contract *RISCVTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRISCV creates a new instance of RISCV, bound to a specific deployed contract.
func NewRISCV(address common.Address, backend bind.ContractBackend) (*RISCV, error) {
	contract, err := bindRISCV(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &RISCV{RISCVCaller: RISCVCaller{contract: contract}, RISCVTransactor: RISCVTransactor{contract: contract}, RISCVFilterer: RISCVFilterer{contract: contract}}, nil
}

// NewRISCVCaller creates a new read-only instance of RISCV, bound to a specific deployed contract.
func NewRISCVCaller(address common.Address, caller bind.ContractCaller) (*RISCVCaller, error) {
	contract, err := bindRISCV(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RISCVCaller{contract: contract}, nil
}

// NewRISCVTransactor creates a new write-only instance of RISCV, bound to a specific deployed contract.
func NewRISCVTransactor(address common.Address, transactor bind.ContractTransactor) (*RISCVTransactor, error) {
	contract, err := bindRISCV(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RISCVTransactor{contract: contract}, nil
}

// NewRISCVFilterer creates a new log filterer instance of RISCV, bound to a specific deployed contract.
func NewRISCVFilterer(address common.Address, filterer bind.ContractFilterer) (*RISCVFilterer, error) {
	contract, err := bindRISCV(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RISCVFilterer{contract: contract}, nil
}

// bindRISCV binds a generic wrapper to an already deployed contract.
func bindRISCV(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := RISCVMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RISCV *RISCVRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RISCV.Contract.RISCVCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RISCV *RISCVRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RISCV.Contract.RISCVTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RISCV *RISCVRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RISCV.Contract.RISCVTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RISCV *RISCVCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RISCV.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RISCV *RISCVTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RISCV.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RISCV *RISCVTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RISCV.Contract.contract.Transact(opts, method, params...)
}

// Oracle is a free data retrieval call binding the contract method 0x7dc0d1d0.
//
// Solidity: function oracle() view returns(address)
func (_RISCV *RISCVCaller) Oracle(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RISCV.contract.Call(opts, &out, "oracle")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Oracle is a free data retrieval call binding the contract method 0x7dc0d1d0.
//
// Solidity: function oracle() view returns(address)
func (_RISCV *RISCVSession) Oracle() (common.Address, error) {
	return _RISCV.Contract.Oracle(&_RISCV.CallOpts)
}

// Oracle is a free data retrieval call binding the contract method 0x7dc0d1d0.
//
// Solidity: function oracle() view returns(address)
func (_RISCV *RISCVCallerSession) Oracle() (common.Address, error) {
	return _RISCV.Contract.Oracle(&_RISCV.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_RISCV *RISCVCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _RISCV.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_RISCV *RISCVSession) Version() (string, error) {
	return _RISCV.Contract.Version(&_RISCV.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_RISCV *RISCVCallerSession) Version() (string, error) {
	return _RISCV.Contract.Version(&_RISCV.CallOpts)
}

// Step is a paid mutator transaction binding the contract method 0xe14ced32.
//
// Solidity: function step(bytes _stateData, bytes _proof, bytes32 _localContext) returns(bytes32)
func (_RISCV *RISCVTransactor) Step(opts *bind.TransactOpts, _stateData []byte, _proof []byte, _localContext [32]byte) (*types.Transaction, error) {
	return _RISCV.contract.Transact(opts, "step", _stateData, _proof, _localContext)
}

// Step is a paid mutator transaction binding the contract method 0xe14ced32.
//
// Solidity: function step(bytes _stateData, bytes _proof, bytes32 _localContext) returns(bytes32)
func (_RISCV *RISCVSession) Step(_stateData []byte, _proof []byte, _localContext [32]byte) (*types.Transaction, error) {
	return _RISCV.Contract.Step(&_RISCV.TransactOpts, _stateData, _proof, _localContext)
}

// Step is a paid mutator transaction binding the contract method 0xe14ced32.
//
// Solidity: function step(bytes _stateData, bytes _proof, bytes32 _localContext) returns(bytes32)
func (_RISCV *RISCVTransactorSession) Step(_stateData []byte, _proof []byte, _localContext [32]byte) (*types.Transaction, error) {
	return _RISCV.Contract.Step(&_RISCV.TransactOpts, _stateData, _proof, _localContext)
}

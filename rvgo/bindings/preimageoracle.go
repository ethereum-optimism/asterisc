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

// LibKeccakStateMatrix is an auto generated low-level Go binding around an user-defined struct.
type LibKeccakStateMatrix struct {
	State [25]uint64
}

// PreimageOracleLeaf is an auto generated low-level Go binding around an user-defined struct.
type PreimageOracleLeaf struct {
	Input           []byte
	Index           *big.Int
	StateCommitment [32]byte
}

// PreimageOracleMetaData contains all meta data concerning the PreimageOracle contract.
var PreimageOracleMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_minProposalSize\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_challengePeriod\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"KECCAK_TREE_DEPTH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_LEAF_COUNT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MIN_BOND_SIZE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"addLeavesLPP\",\"inputs\":[{\"name\":\"_uuid\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_inputStartBlock\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_input\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_stateCommitments\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"_finalize\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"challengeFirstLPP\",\"inputs\":[{\"name\":\"_claimant\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_uuid\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_postState\",\"type\":\"tuple\",\"internalType\":\"structPreimageOracle.Leaf\",\"components\":[{\"name\":\"input\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"stateCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"_postStateProof\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"challengeLPP\",\"inputs\":[{\"name\":\"_claimant\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_uuid\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_stateMatrix\",\"type\":\"tuple\",\"internalType\":\"structLibKeccak.StateMatrix\",\"components\":[{\"name\":\"state\",\"type\":\"uint64[25]\",\"internalType\":\"uint64[25]\"}]},{\"name\":\"_preState\",\"type\":\"tuple\",\"internalType\":\"structPreimageOracle.Leaf\",\"components\":[{\"name\":\"input\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"stateCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"_preStateProof\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"_postState\",\"type\":\"tuple\",\"internalType\":\"structPreimageOracle.Leaf\",\"components\":[{\"name\":\"input\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"stateCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"_postStateProof\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"challengePeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"challengePeriod_\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTreeRootLPP\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_uuid\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"treeRoot_\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initLPP\",\"inputs\":[{\"name\":\"_uuid\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_partOffset\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_claimedSize\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"loadBlobPreimagePart\",\"inputs\":[{\"name\":\"_z\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_y\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_commitment\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_proof\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_partOffset\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"loadKeccak256PreimagePart\",\"inputs\":[{\"name\":\"_partOffset\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_preimage\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"loadLocalData\",\"inputs\":[{\"name\":\"_ident\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_localContext\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_word\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_size\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_partOffset\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"key_\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"loadPrecompilePreimagePart\",\"inputs\":[{\"name\":\"_partOffset\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_precompile\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_input\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"loadSha256PreimagePart\",\"inputs\":[{\"name\":\"_partOffset\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_preimage\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"minProposalSize\",\"inputs\":[],\"outputs\":[{\"name\":\"minProposalSize_\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"preimageLengths\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"preimagePartOk\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"preimageParts\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposalBlocks\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposalBlocksLen\",\"inputs\":[{\"name\":\"_claimant\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_uuid\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"len_\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposalBonds\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposalBranches\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposalCount\",\"inputs\":[],\"outputs\":[{\"name\":\"count_\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposalMetadata\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"LPPMetaData\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposalParts\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposals\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"claimant\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"uuid\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"readPreimage\",\"inputs\":[{\"name\":\"_key\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_offset\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"dat_\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"datLen_\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"squeezeLPP\",\"inputs\":[{\"name\":\"_claimant\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_uuid\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_stateMatrix\",\"type\":\"tuple\",\"internalType\":\"structLibKeccak.StateMatrix\",\"components\":[{\"name\":\"state\",\"type\":\"uint64[25]\",\"internalType\":\"uint64[25]\"}]},{\"name\":\"_preState\",\"type\":\"tuple\",\"internalType\":\"structPreimageOracle.Leaf\",\"components\":[{\"name\":\"input\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"stateCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"_preStateProof\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"_postState\",\"type\":\"tuple\",\"internalType\":\"structPreimageOracle.Leaf\",\"components\":[{\"name\":\"input\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"stateCommitment\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"_postStateProof\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"version\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"zeroHashes\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"error\",\"name\":\"ActiveProposal\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AlreadyFinalized\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"BadProposal\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"BondTransferFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientBond\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInputSize\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPreimage\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidProof\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotEOA\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitialized\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PartOffsetOOB\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"PostStateMatches\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StatesNotContiguous\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TreeSizeOverflow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"WrongStartingBlock\",\"inputs\":[]}]",
}

// PreimageOracleABI is the input ABI used to generate the binding from.
// Deprecated: Use PreimageOracleMetaData.ABI instead.
var PreimageOracleABI = PreimageOracleMetaData.ABI

// PreimageOracle is an auto generated Go binding around an Ethereum contract.
type PreimageOracle struct {
	PreimageOracleCaller     // Read-only binding to the contract
	PreimageOracleTransactor // Write-only binding to the contract
	PreimageOracleFilterer   // Log filterer for contract events
}

// PreimageOracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type PreimageOracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PreimageOracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PreimageOracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PreimageOracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PreimageOracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PreimageOracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PreimageOracleSession struct {
	Contract     *PreimageOracle   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PreimageOracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PreimageOracleCallerSession struct {
	Contract *PreimageOracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// PreimageOracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PreimageOracleTransactorSession struct {
	Contract     *PreimageOracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// PreimageOracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type PreimageOracleRaw struct {
	Contract *PreimageOracle // Generic contract binding to access the raw methods on
}

// PreimageOracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PreimageOracleCallerRaw struct {
	Contract *PreimageOracleCaller // Generic read-only contract binding to access the raw methods on
}

// PreimageOracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PreimageOracleTransactorRaw struct {
	Contract *PreimageOracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPreimageOracle creates a new instance of PreimageOracle, bound to a specific deployed contract.
func NewPreimageOracle(address common.Address, backend bind.ContractBackend) (*PreimageOracle, error) {
	contract, err := bindPreimageOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PreimageOracle{PreimageOracleCaller: PreimageOracleCaller{contract: contract}, PreimageOracleTransactor: PreimageOracleTransactor{contract: contract}, PreimageOracleFilterer: PreimageOracleFilterer{contract: contract}}, nil
}

// NewPreimageOracleCaller creates a new read-only instance of PreimageOracle, bound to a specific deployed contract.
func NewPreimageOracleCaller(address common.Address, caller bind.ContractCaller) (*PreimageOracleCaller, error) {
	contract, err := bindPreimageOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PreimageOracleCaller{contract: contract}, nil
}

// NewPreimageOracleTransactor creates a new write-only instance of PreimageOracle, bound to a specific deployed contract.
func NewPreimageOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*PreimageOracleTransactor, error) {
	contract, err := bindPreimageOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PreimageOracleTransactor{contract: contract}, nil
}

// NewPreimageOracleFilterer creates a new log filterer instance of PreimageOracle, bound to a specific deployed contract.
func NewPreimageOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*PreimageOracleFilterer, error) {
	contract, err := bindPreimageOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PreimageOracleFilterer{contract: contract}, nil
}

// bindPreimageOracle binds a generic wrapper to an already deployed contract.
func bindPreimageOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PreimageOracleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PreimageOracle *PreimageOracleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PreimageOracle.Contract.PreimageOracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PreimageOracle *PreimageOracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PreimageOracle.Contract.PreimageOracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PreimageOracle *PreimageOracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PreimageOracle.Contract.PreimageOracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PreimageOracle *PreimageOracleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PreimageOracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PreimageOracle *PreimageOracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PreimageOracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PreimageOracle *PreimageOracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PreimageOracle.Contract.contract.Transact(opts, method, params...)
}

// KECCAKTREEDEPTH is a free data retrieval call binding the contract method 0x2055b36b.
//
// Solidity: function KECCAK_TREE_DEPTH() view returns(uint256)
func (_PreimageOracle *PreimageOracleCaller) KECCAKTREEDEPTH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "KECCAK_TREE_DEPTH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// KECCAKTREEDEPTH is a free data retrieval call binding the contract method 0x2055b36b.
//
// Solidity: function KECCAK_TREE_DEPTH() view returns(uint256)
func (_PreimageOracle *PreimageOracleSession) KECCAKTREEDEPTH() (*big.Int, error) {
	return _PreimageOracle.Contract.KECCAKTREEDEPTH(&_PreimageOracle.CallOpts)
}

// KECCAKTREEDEPTH is a free data retrieval call binding the contract method 0x2055b36b.
//
// Solidity: function KECCAK_TREE_DEPTH() view returns(uint256)
func (_PreimageOracle *PreimageOracleCallerSession) KECCAKTREEDEPTH() (*big.Int, error) {
	return _PreimageOracle.Contract.KECCAKTREEDEPTH(&_PreimageOracle.CallOpts)
}

// MAXLEAFCOUNT is a free data retrieval call binding the contract method 0x4d52b4c9.
//
// Solidity: function MAX_LEAF_COUNT() view returns(uint256)
func (_PreimageOracle *PreimageOracleCaller) MAXLEAFCOUNT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "MAX_LEAF_COUNT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXLEAFCOUNT is a free data retrieval call binding the contract method 0x4d52b4c9.
//
// Solidity: function MAX_LEAF_COUNT() view returns(uint256)
func (_PreimageOracle *PreimageOracleSession) MAXLEAFCOUNT() (*big.Int, error) {
	return _PreimageOracle.Contract.MAXLEAFCOUNT(&_PreimageOracle.CallOpts)
}

// MAXLEAFCOUNT is a free data retrieval call binding the contract method 0x4d52b4c9.
//
// Solidity: function MAX_LEAF_COUNT() view returns(uint256)
func (_PreimageOracle *PreimageOracleCallerSession) MAXLEAFCOUNT() (*big.Int, error) {
	return _PreimageOracle.Contract.MAXLEAFCOUNT(&_PreimageOracle.CallOpts)
}

// MINBONDSIZE is a free data retrieval call binding the contract method 0x7051472e.
//
// Solidity: function MIN_BOND_SIZE() view returns(uint256)
func (_PreimageOracle *PreimageOracleCaller) MINBONDSIZE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "MIN_BOND_SIZE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINBONDSIZE is a free data retrieval call binding the contract method 0x7051472e.
//
// Solidity: function MIN_BOND_SIZE() view returns(uint256)
func (_PreimageOracle *PreimageOracleSession) MINBONDSIZE() (*big.Int, error) {
	return _PreimageOracle.Contract.MINBONDSIZE(&_PreimageOracle.CallOpts)
}

// MINBONDSIZE is a free data retrieval call binding the contract method 0x7051472e.
//
// Solidity: function MIN_BOND_SIZE() view returns(uint256)
func (_PreimageOracle *PreimageOracleCallerSession) MINBONDSIZE() (*big.Int, error) {
	return _PreimageOracle.Contract.MINBONDSIZE(&_PreimageOracle.CallOpts)
}

// ChallengePeriod is a free data retrieval call binding the contract method 0xf3f480d9.
//
// Solidity: function challengePeriod() view returns(uint256 challengePeriod_)
func (_PreimageOracle *PreimageOracleCaller) ChallengePeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "challengePeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ChallengePeriod is a free data retrieval call binding the contract method 0xf3f480d9.
//
// Solidity: function challengePeriod() view returns(uint256 challengePeriod_)
func (_PreimageOracle *PreimageOracleSession) ChallengePeriod() (*big.Int, error) {
	return _PreimageOracle.Contract.ChallengePeriod(&_PreimageOracle.CallOpts)
}

// ChallengePeriod is a free data retrieval call binding the contract method 0xf3f480d9.
//
// Solidity: function challengePeriod() view returns(uint256 challengePeriod_)
func (_PreimageOracle *PreimageOracleCallerSession) ChallengePeriod() (*big.Int, error) {
	return _PreimageOracle.Contract.ChallengePeriod(&_PreimageOracle.CallOpts)
}

// GetTreeRootLPP is a free data retrieval call binding the contract method 0x0359a563.
//
// Solidity: function getTreeRootLPP(address _owner, uint256 _uuid) view returns(bytes32 treeRoot_)
func (_PreimageOracle *PreimageOracleCaller) GetTreeRootLPP(opts *bind.CallOpts, _owner common.Address, _uuid *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "getTreeRootLPP", _owner, _uuid)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetTreeRootLPP is a free data retrieval call binding the contract method 0x0359a563.
//
// Solidity: function getTreeRootLPP(address _owner, uint256 _uuid) view returns(bytes32 treeRoot_)
func (_PreimageOracle *PreimageOracleSession) GetTreeRootLPP(_owner common.Address, _uuid *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.GetTreeRootLPP(&_PreimageOracle.CallOpts, _owner, _uuid)
}

// GetTreeRootLPP is a free data retrieval call binding the contract method 0x0359a563.
//
// Solidity: function getTreeRootLPP(address _owner, uint256 _uuid) view returns(bytes32 treeRoot_)
func (_PreimageOracle *PreimageOracleCallerSession) GetTreeRootLPP(_owner common.Address, _uuid *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.GetTreeRootLPP(&_PreimageOracle.CallOpts, _owner, _uuid)
}

// MinProposalSize is a free data retrieval call binding the contract method 0xdd24f9bf.
//
// Solidity: function minProposalSize() view returns(uint256 minProposalSize_)
func (_PreimageOracle *PreimageOracleCaller) MinProposalSize(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "minProposalSize")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinProposalSize is a free data retrieval call binding the contract method 0xdd24f9bf.
//
// Solidity: function minProposalSize() view returns(uint256 minProposalSize_)
func (_PreimageOracle *PreimageOracleSession) MinProposalSize() (*big.Int, error) {
	return _PreimageOracle.Contract.MinProposalSize(&_PreimageOracle.CallOpts)
}

// MinProposalSize is a free data retrieval call binding the contract method 0xdd24f9bf.
//
// Solidity: function minProposalSize() view returns(uint256 minProposalSize_)
func (_PreimageOracle *PreimageOracleCallerSession) MinProposalSize() (*big.Int, error) {
	return _PreimageOracle.Contract.MinProposalSize(&_PreimageOracle.CallOpts)
}

// PreimageLengths is a free data retrieval call binding the contract method 0xfef2b4ed.
//
// Solidity: function preimageLengths(bytes32 ) view returns(uint256)
func (_PreimageOracle *PreimageOracleCaller) PreimageLengths(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "preimageLengths", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PreimageLengths is a free data retrieval call binding the contract method 0xfef2b4ed.
//
// Solidity: function preimageLengths(bytes32 ) view returns(uint256)
func (_PreimageOracle *PreimageOracleSession) PreimageLengths(arg0 [32]byte) (*big.Int, error) {
	return _PreimageOracle.Contract.PreimageLengths(&_PreimageOracle.CallOpts, arg0)
}

// PreimageLengths is a free data retrieval call binding the contract method 0xfef2b4ed.
//
// Solidity: function preimageLengths(bytes32 ) view returns(uint256)
func (_PreimageOracle *PreimageOracleCallerSession) PreimageLengths(arg0 [32]byte) (*big.Int, error) {
	return _PreimageOracle.Contract.PreimageLengths(&_PreimageOracle.CallOpts, arg0)
}

// PreimagePartOk is a free data retrieval call binding the contract method 0x8542cf50.
//
// Solidity: function preimagePartOk(bytes32 , uint256 ) view returns(bool)
func (_PreimageOracle *PreimageOracleCaller) PreimagePartOk(opts *bind.CallOpts, arg0 [32]byte, arg1 *big.Int) (bool, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "preimagePartOk", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// PreimagePartOk is a free data retrieval call binding the contract method 0x8542cf50.
//
// Solidity: function preimagePartOk(bytes32 , uint256 ) view returns(bool)
func (_PreimageOracle *PreimageOracleSession) PreimagePartOk(arg0 [32]byte, arg1 *big.Int) (bool, error) {
	return _PreimageOracle.Contract.PreimagePartOk(&_PreimageOracle.CallOpts, arg0, arg1)
}

// PreimagePartOk is a free data retrieval call binding the contract method 0x8542cf50.
//
// Solidity: function preimagePartOk(bytes32 , uint256 ) view returns(bool)
func (_PreimageOracle *PreimageOracleCallerSession) PreimagePartOk(arg0 [32]byte, arg1 *big.Int) (bool, error) {
	return _PreimageOracle.Contract.PreimagePartOk(&_PreimageOracle.CallOpts, arg0, arg1)
}

// PreimageParts is a free data retrieval call binding the contract method 0x61238bde.
//
// Solidity: function preimageParts(bytes32 , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCaller) PreimageParts(opts *bind.CallOpts, arg0 [32]byte, arg1 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "preimageParts", arg0, arg1)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PreimageParts is a free data retrieval call binding the contract method 0x61238bde.
//
// Solidity: function preimageParts(bytes32 , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleSession) PreimageParts(arg0 [32]byte, arg1 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.PreimageParts(&_PreimageOracle.CallOpts, arg0, arg1)
}

// PreimageParts is a free data retrieval call binding the contract method 0x61238bde.
//
// Solidity: function preimageParts(bytes32 , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCallerSession) PreimageParts(arg0 [32]byte, arg1 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.PreimageParts(&_PreimageOracle.CallOpts, arg0, arg1)
}

// ProposalBlocks is a free data retrieval call binding the contract method 0x882856ef.
//
// Solidity: function proposalBlocks(address , uint256 , uint256 ) view returns(uint64)
func (_PreimageOracle *PreimageOracleCaller) ProposalBlocks(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int, arg2 *big.Int) (uint64, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "proposalBlocks", arg0, arg1, arg2)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// ProposalBlocks is a free data retrieval call binding the contract method 0x882856ef.
//
// Solidity: function proposalBlocks(address , uint256 , uint256 ) view returns(uint64)
func (_PreimageOracle *PreimageOracleSession) ProposalBlocks(arg0 common.Address, arg1 *big.Int, arg2 *big.Int) (uint64, error) {
	return _PreimageOracle.Contract.ProposalBlocks(&_PreimageOracle.CallOpts, arg0, arg1, arg2)
}

// ProposalBlocks is a free data retrieval call binding the contract method 0x882856ef.
//
// Solidity: function proposalBlocks(address , uint256 , uint256 ) view returns(uint64)
func (_PreimageOracle *PreimageOracleCallerSession) ProposalBlocks(arg0 common.Address, arg1 *big.Int, arg2 *big.Int) (uint64, error) {
	return _PreimageOracle.Contract.ProposalBlocks(&_PreimageOracle.CallOpts, arg0, arg1, arg2)
}

// ProposalBlocksLen is a free data retrieval call binding the contract method 0x9d53a648.
//
// Solidity: function proposalBlocksLen(address _claimant, uint256 _uuid) view returns(uint256 len_)
func (_PreimageOracle *PreimageOracleCaller) ProposalBlocksLen(opts *bind.CallOpts, _claimant common.Address, _uuid *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "proposalBlocksLen", _claimant, _uuid)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProposalBlocksLen is a free data retrieval call binding the contract method 0x9d53a648.
//
// Solidity: function proposalBlocksLen(address _claimant, uint256 _uuid) view returns(uint256 len_)
func (_PreimageOracle *PreimageOracleSession) ProposalBlocksLen(_claimant common.Address, _uuid *big.Int) (*big.Int, error) {
	return _PreimageOracle.Contract.ProposalBlocksLen(&_PreimageOracle.CallOpts, _claimant, _uuid)
}

// ProposalBlocksLen is a free data retrieval call binding the contract method 0x9d53a648.
//
// Solidity: function proposalBlocksLen(address _claimant, uint256 _uuid) view returns(uint256 len_)
func (_PreimageOracle *PreimageOracleCallerSession) ProposalBlocksLen(_claimant common.Address, _uuid *big.Int) (*big.Int, error) {
	return _PreimageOracle.Contract.ProposalBlocksLen(&_PreimageOracle.CallOpts, _claimant, _uuid)
}

// ProposalBonds is a free data retrieval call binding the contract method 0xddcd58de.
//
// Solidity: function proposalBonds(address , uint256 ) view returns(uint256)
func (_PreimageOracle *PreimageOracleCaller) ProposalBonds(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "proposalBonds", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProposalBonds is a free data retrieval call binding the contract method 0xddcd58de.
//
// Solidity: function proposalBonds(address , uint256 ) view returns(uint256)
func (_PreimageOracle *PreimageOracleSession) ProposalBonds(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _PreimageOracle.Contract.ProposalBonds(&_PreimageOracle.CallOpts, arg0, arg1)
}

// ProposalBonds is a free data retrieval call binding the contract method 0xddcd58de.
//
// Solidity: function proposalBonds(address , uint256 ) view returns(uint256)
func (_PreimageOracle *PreimageOracleCallerSession) ProposalBonds(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _PreimageOracle.Contract.ProposalBonds(&_PreimageOracle.CallOpts, arg0, arg1)
}

// ProposalBranches is a free data retrieval call binding the contract method 0xb4801e61.
//
// Solidity: function proposalBranches(address , uint256 , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCaller) ProposalBranches(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int, arg2 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "proposalBranches", arg0, arg1, arg2)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProposalBranches is a free data retrieval call binding the contract method 0xb4801e61.
//
// Solidity: function proposalBranches(address , uint256 , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleSession) ProposalBranches(arg0 common.Address, arg1 *big.Int, arg2 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.ProposalBranches(&_PreimageOracle.CallOpts, arg0, arg1, arg2)
}

// ProposalBranches is a free data retrieval call binding the contract method 0xb4801e61.
//
// Solidity: function proposalBranches(address , uint256 , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCallerSession) ProposalBranches(arg0 common.Address, arg1 *big.Int, arg2 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.ProposalBranches(&_PreimageOracle.CallOpts, arg0, arg1, arg2)
}

// ProposalCount is a free data retrieval call binding the contract method 0xda35c664.
//
// Solidity: function proposalCount() view returns(uint256 count_)
func (_PreimageOracle *PreimageOracleCaller) ProposalCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "proposalCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProposalCount is a free data retrieval call binding the contract method 0xda35c664.
//
// Solidity: function proposalCount() view returns(uint256 count_)
func (_PreimageOracle *PreimageOracleSession) ProposalCount() (*big.Int, error) {
	return _PreimageOracle.Contract.ProposalCount(&_PreimageOracle.CallOpts)
}

// ProposalCount is a free data retrieval call binding the contract method 0xda35c664.
//
// Solidity: function proposalCount() view returns(uint256 count_)
func (_PreimageOracle *PreimageOracleCallerSession) ProposalCount() (*big.Int, error) {
	return _PreimageOracle.Contract.ProposalCount(&_PreimageOracle.CallOpts)
}

// ProposalMetadata is a free data retrieval call binding the contract method 0x6551927b.
//
// Solidity: function proposalMetadata(address , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCaller) ProposalMetadata(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "proposalMetadata", arg0, arg1)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProposalMetadata is a free data retrieval call binding the contract method 0x6551927b.
//
// Solidity: function proposalMetadata(address , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleSession) ProposalMetadata(arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.ProposalMetadata(&_PreimageOracle.CallOpts, arg0, arg1)
}

// ProposalMetadata is a free data retrieval call binding the contract method 0x6551927b.
//
// Solidity: function proposalMetadata(address , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCallerSession) ProposalMetadata(arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.ProposalMetadata(&_PreimageOracle.CallOpts, arg0, arg1)
}

// ProposalParts is a free data retrieval call binding the contract method 0xb2e67ba8.
//
// Solidity: function proposalParts(address , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCaller) ProposalParts(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "proposalParts", arg0, arg1)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProposalParts is a free data retrieval call binding the contract method 0xb2e67ba8.
//
// Solidity: function proposalParts(address , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleSession) ProposalParts(arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.ProposalParts(&_PreimageOracle.CallOpts, arg0, arg1)
}

// ProposalParts is a free data retrieval call binding the contract method 0xb2e67ba8.
//
// Solidity: function proposalParts(address , uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCallerSession) ProposalParts(arg0 common.Address, arg1 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.ProposalParts(&_PreimageOracle.CallOpts, arg0, arg1)
}

// Proposals is a free data retrieval call binding the contract method 0x013cf08b.
//
// Solidity: function proposals(uint256 ) view returns(address claimant, uint256 uuid)
func (_PreimageOracle *PreimageOracleCaller) Proposals(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Claimant common.Address
	Uuid     *big.Int
}, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "proposals", arg0)

	outstruct := new(struct {
		Claimant common.Address
		Uuid     *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Claimant = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Uuid = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Proposals is a free data retrieval call binding the contract method 0x013cf08b.
//
// Solidity: function proposals(uint256 ) view returns(address claimant, uint256 uuid)
func (_PreimageOracle *PreimageOracleSession) Proposals(arg0 *big.Int) (struct {
	Claimant common.Address
	Uuid     *big.Int
}, error) {
	return _PreimageOracle.Contract.Proposals(&_PreimageOracle.CallOpts, arg0)
}

// Proposals is a free data retrieval call binding the contract method 0x013cf08b.
//
// Solidity: function proposals(uint256 ) view returns(address claimant, uint256 uuid)
func (_PreimageOracle *PreimageOracleCallerSession) Proposals(arg0 *big.Int) (struct {
	Claimant common.Address
	Uuid     *big.Int
}, error) {
	return _PreimageOracle.Contract.Proposals(&_PreimageOracle.CallOpts, arg0)
}

// ReadPreimage is a free data retrieval call binding the contract method 0xe03110e1.
//
// Solidity: function readPreimage(bytes32 _key, uint256 _offset) view returns(bytes32 dat_, uint256 datLen_)
func (_PreimageOracle *PreimageOracleCaller) ReadPreimage(opts *bind.CallOpts, _key [32]byte, _offset *big.Int) (struct {
	Dat    [32]byte
	DatLen *big.Int
}, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "readPreimage", _key, _offset)

	outstruct := new(struct {
		Dat    [32]byte
		DatLen *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Dat = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.DatLen = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// ReadPreimage is a free data retrieval call binding the contract method 0xe03110e1.
//
// Solidity: function readPreimage(bytes32 _key, uint256 _offset) view returns(bytes32 dat_, uint256 datLen_)
func (_PreimageOracle *PreimageOracleSession) ReadPreimage(_key [32]byte, _offset *big.Int) (struct {
	Dat    [32]byte
	DatLen *big.Int
}, error) {
	return _PreimageOracle.Contract.ReadPreimage(&_PreimageOracle.CallOpts, _key, _offset)
}

// ReadPreimage is a free data retrieval call binding the contract method 0xe03110e1.
//
// Solidity: function readPreimage(bytes32 _key, uint256 _offset) view returns(bytes32 dat_, uint256 datLen_)
func (_PreimageOracle *PreimageOracleCallerSession) ReadPreimage(_key [32]byte, _offset *big.Int) (struct {
	Dat    [32]byte
	DatLen *big.Int
}, error) {
	return _PreimageOracle.Contract.ReadPreimage(&_PreimageOracle.CallOpts, _key, _offset)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_PreimageOracle *PreimageOracleCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_PreimageOracle *PreimageOracleSession) Version() (string, error) {
	return _PreimageOracle.Contract.Version(&_PreimageOracle.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_PreimageOracle *PreimageOracleCallerSession) Version() (string, error) {
	return _PreimageOracle.Contract.Version(&_PreimageOracle.CallOpts)
}

// ZeroHashes is a free data retrieval call binding the contract method 0x7ac54767.
//
// Solidity: function zeroHashes(uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCaller) ZeroHashes(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _PreimageOracle.contract.Call(opts, &out, "zeroHashes", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ZeroHashes is a free data retrieval call binding the contract method 0x7ac54767.
//
// Solidity: function zeroHashes(uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleSession) ZeroHashes(arg0 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.ZeroHashes(&_PreimageOracle.CallOpts, arg0)
}

// ZeroHashes is a free data retrieval call binding the contract method 0x7ac54767.
//
// Solidity: function zeroHashes(uint256 ) view returns(bytes32)
func (_PreimageOracle *PreimageOracleCallerSession) ZeroHashes(arg0 *big.Int) ([32]byte, error) {
	return _PreimageOracle.Contract.ZeroHashes(&_PreimageOracle.CallOpts, arg0)
}

// AddLeavesLPP is a paid mutator transaction binding the contract method 0x7917de1d.
//
// Solidity: function addLeavesLPP(uint256 _uuid, uint256 _inputStartBlock, bytes _input, bytes32[] _stateCommitments, bool _finalize) returns()
func (_PreimageOracle *PreimageOracleTransactor) AddLeavesLPP(opts *bind.TransactOpts, _uuid *big.Int, _inputStartBlock *big.Int, _input []byte, _stateCommitments [][32]byte, _finalize bool) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "addLeavesLPP", _uuid, _inputStartBlock, _input, _stateCommitments, _finalize)
}

// AddLeavesLPP is a paid mutator transaction binding the contract method 0x7917de1d.
//
// Solidity: function addLeavesLPP(uint256 _uuid, uint256 _inputStartBlock, bytes _input, bytes32[] _stateCommitments, bool _finalize) returns()
func (_PreimageOracle *PreimageOracleSession) AddLeavesLPP(_uuid *big.Int, _inputStartBlock *big.Int, _input []byte, _stateCommitments [][32]byte, _finalize bool) (*types.Transaction, error) {
	return _PreimageOracle.Contract.AddLeavesLPP(&_PreimageOracle.TransactOpts, _uuid, _inputStartBlock, _input, _stateCommitments, _finalize)
}

// AddLeavesLPP is a paid mutator transaction binding the contract method 0x7917de1d.
//
// Solidity: function addLeavesLPP(uint256 _uuid, uint256 _inputStartBlock, bytes _input, bytes32[] _stateCommitments, bool _finalize) returns()
func (_PreimageOracle *PreimageOracleTransactorSession) AddLeavesLPP(_uuid *big.Int, _inputStartBlock *big.Int, _input []byte, _stateCommitments [][32]byte, _finalize bool) (*types.Transaction, error) {
	return _PreimageOracle.Contract.AddLeavesLPP(&_PreimageOracle.TransactOpts, _uuid, _inputStartBlock, _input, _stateCommitments, _finalize)
}

// ChallengeFirstLPP is a paid mutator transaction binding the contract method 0xec5efcbc.
//
// Solidity: function challengeFirstLPP(address _claimant, uint256 _uuid, (bytes,uint256,bytes32) _postState, bytes32[] _postStateProof) returns()
func (_PreimageOracle *PreimageOracleTransactor) ChallengeFirstLPP(opts *bind.TransactOpts, _claimant common.Address, _uuid *big.Int, _postState PreimageOracleLeaf, _postStateProof [][32]byte) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "challengeFirstLPP", _claimant, _uuid, _postState, _postStateProof)
}

// ChallengeFirstLPP is a paid mutator transaction binding the contract method 0xec5efcbc.
//
// Solidity: function challengeFirstLPP(address _claimant, uint256 _uuid, (bytes,uint256,bytes32) _postState, bytes32[] _postStateProof) returns()
func (_PreimageOracle *PreimageOracleSession) ChallengeFirstLPP(_claimant common.Address, _uuid *big.Int, _postState PreimageOracleLeaf, _postStateProof [][32]byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.ChallengeFirstLPP(&_PreimageOracle.TransactOpts, _claimant, _uuid, _postState, _postStateProof)
}

// ChallengeFirstLPP is a paid mutator transaction binding the contract method 0xec5efcbc.
//
// Solidity: function challengeFirstLPP(address _claimant, uint256 _uuid, (bytes,uint256,bytes32) _postState, bytes32[] _postStateProof) returns()
func (_PreimageOracle *PreimageOracleTransactorSession) ChallengeFirstLPP(_claimant common.Address, _uuid *big.Int, _postState PreimageOracleLeaf, _postStateProof [][32]byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.ChallengeFirstLPP(&_PreimageOracle.TransactOpts, _claimant, _uuid, _postState, _postStateProof)
}

// ChallengeLPP is a paid mutator transaction binding the contract method 0x3909af5c.
//
// Solidity: function challengeLPP(address _claimant, uint256 _uuid, (uint64[25]) _stateMatrix, (bytes,uint256,bytes32) _preState, bytes32[] _preStateProof, (bytes,uint256,bytes32) _postState, bytes32[] _postStateProof) returns()
func (_PreimageOracle *PreimageOracleTransactor) ChallengeLPP(opts *bind.TransactOpts, _claimant common.Address, _uuid *big.Int, _stateMatrix LibKeccakStateMatrix, _preState PreimageOracleLeaf, _preStateProof [][32]byte, _postState PreimageOracleLeaf, _postStateProof [][32]byte) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "challengeLPP", _claimant, _uuid, _stateMatrix, _preState, _preStateProof, _postState, _postStateProof)
}

// ChallengeLPP is a paid mutator transaction binding the contract method 0x3909af5c.
//
// Solidity: function challengeLPP(address _claimant, uint256 _uuid, (uint64[25]) _stateMatrix, (bytes,uint256,bytes32) _preState, bytes32[] _preStateProof, (bytes,uint256,bytes32) _postState, bytes32[] _postStateProof) returns()
func (_PreimageOracle *PreimageOracleSession) ChallengeLPP(_claimant common.Address, _uuid *big.Int, _stateMatrix LibKeccakStateMatrix, _preState PreimageOracleLeaf, _preStateProof [][32]byte, _postState PreimageOracleLeaf, _postStateProof [][32]byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.ChallengeLPP(&_PreimageOracle.TransactOpts, _claimant, _uuid, _stateMatrix, _preState, _preStateProof, _postState, _postStateProof)
}

// ChallengeLPP is a paid mutator transaction binding the contract method 0x3909af5c.
//
// Solidity: function challengeLPP(address _claimant, uint256 _uuid, (uint64[25]) _stateMatrix, (bytes,uint256,bytes32) _preState, bytes32[] _preStateProof, (bytes,uint256,bytes32) _postState, bytes32[] _postStateProof) returns()
func (_PreimageOracle *PreimageOracleTransactorSession) ChallengeLPP(_claimant common.Address, _uuid *big.Int, _stateMatrix LibKeccakStateMatrix, _preState PreimageOracleLeaf, _preStateProof [][32]byte, _postState PreimageOracleLeaf, _postStateProof [][32]byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.ChallengeLPP(&_PreimageOracle.TransactOpts, _claimant, _uuid, _stateMatrix, _preState, _preStateProof, _postState, _postStateProof)
}

// InitLPP is a paid mutator transaction binding the contract method 0xfaf37bc7.
//
// Solidity: function initLPP(uint256 _uuid, uint32 _partOffset, uint32 _claimedSize) payable returns()
func (_PreimageOracle *PreimageOracleTransactor) InitLPP(opts *bind.TransactOpts, _uuid *big.Int, _partOffset uint32, _claimedSize uint32) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "initLPP", _uuid, _partOffset, _claimedSize)
}

// InitLPP is a paid mutator transaction binding the contract method 0xfaf37bc7.
//
// Solidity: function initLPP(uint256 _uuid, uint32 _partOffset, uint32 _claimedSize) payable returns()
func (_PreimageOracle *PreimageOracleSession) InitLPP(_uuid *big.Int, _partOffset uint32, _claimedSize uint32) (*types.Transaction, error) {
	return _PreimageOracle.Contract.InitLPP(&_PreimageOracle.TransactOpts, _uuid, _partOffset, _claimedSize)
}

// InitLPP is a paid mutator transaction binding the contract method 0xfaf37bc7.
//
// Solidity: function initLPP(uint256 _uuid, uint32 _partOffset, uint32 _claimedSize) payable returns()
func (_PreimageOracle *PreimageOracleTransactorSession) InitLPP(_uuid *big.Int, _partOffset uint32, _claimedSize uint32) (*types.Transaction, error) {
	return _PreimageOracle.Contract.InitLPP(&_PreimageOracle.TransactOpts, _uuid, _partOffset, _claimedSize)
}

// LoadBlobPreimagePart is a paid mutator transaction binding the contract method 0x9d7e8769.
//
// Solidity: function loadBlobPreimagePart(uint256 _z, uint256 _y, bytes _commitment, bytes _proof, uint256 _partOffset) returns()
func (_PreimageOracle *PreimageOracleTransactor) LoadBlobPreimagePart(opts *bind.TransactOpts, _z *big.Int, _y *big.Int, _commitment []byte, _proof []byte, _partOffset *big.Int) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "loadBlobPreimagePart", _z, _y, _commitment, _proof, _partOffset)
}

// LoadBlobPreimagePart is a paid mutator transaction binding the contract method 0x9d7e8769.
//
// Solidity: function loadBlobPreimagePart(uint256 _z, uint256 _y, bytes _commitment, bytes _proof, uint256 _partOffset) returns()
func (_PreimageOracle *PreimageOracleSession) LoadBlobPreimagePart(_z *big.Int, _y *big.Int, _commitment []byte, _proof []byte, _partOffset *big.Int) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadBlobPreimagePart(&_PreimageOracle.TransactOpts, _z, _y, _commitment, _proof, _partOffset)
}

// LoadBlobPreimagePart is a paid mutator transaction binding the contract method 0x9d7e8769.
//
// Solidity: function loadBlobPreimagePart(uint256 _z, uint256 _y, bytes _commitment, bytes _proof, uint256 _partOffset) returns()
func (_PreimageOracle *PreimageOracleTransactorSession) LoadBlobPreimagePart(_z *big.Int, _y *big.Int, _commitment []byte, _proof []byte, _partOffset *big.Int) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadBlobPreimagePart(&_PreimageOracle.TransactOpts, _z, _y, _commitment, _proof, _partOffset)
}

// LoadKeccak256PreimagePart is a paid mutator transaction binding the contract method 0xe1592611.
//
// Solidity: function loadKeccak256PreimagePart(uint256 _partOffset, bytes _preimage) returns()
func (_PreimageOracle *PreimageOracleTransactor) LoadKeccak256PreimagePart(opts *bind.TransactOpts, _partOffset *big.Int, _preimage []byte) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "loadKeccak256PreimagePart", _partOffset, _preimage)
}

// LoadKeccak256PreimagePart is a paid mutator transaction binding the contract method 0xe1592611.
//
// Solidity: function loadKeccak256PreimagePart(uint256 _partOffset, bytes _preimage) returns()
func (_PreimageOracle *PreimageOracleSession) LoadKeccak256PreimagePart(_partOffset *big.Int, _preimage []byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadKeccak256PreimagePart(&_PreimageOracle.TransactOpts, _partOffset, _preimage)
}

// LoadKeccak256PreimagePart is a paid mutator transaction binding the contract method 0xe1592611.
//
// Solidity: function loadKeccak256PreimagePart(uint256 _partOffset, bytes _preimage) returns()
func (_PreimageOracle *PreimageOracleTransactorSession) LoadKeccak256PreimagePart(_partOffset *big.Int, _preimage []byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadKeccak256PreimagePart(&_PreimageOracle.TransactOpts, _partOffset, _preimage)
}

// LoadLocalData is a paid mutator transaction binding the contract method 0x52f0f3ad.
//
// Solidity: function loadLocalData(uint256 _ident, bytes32 _localContext, bytes32 _word, uint256 _size, uint256 _partOffset) returns(bytes32 key_)
func (_PreimageOracle *PreimageOracleTransactor) LoadLocalData(opts *bind.TransactOpts, _ident *big.Int, _localContext [32]byte, _word [32]byte, _size *big.Int, _partOffset *big.Int) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "loadLocalData", _ident, _localContext, _word, _size, _partOffset)
}

// LoadLocalData is a paid mutator transaction binding the contract method 0x52f0f3ad.
//
// Solidity: function loadLocalData(uint256 _ident, bytes32 _localContext, bytes32 _word, uint256 _size, uint256 _partOffset) returns(bytes32 key_)
func (_PreimageOracle *PreimageOracleSession) LoadLocalData(_ident *big.Int, _localContext [32]byte, _word [32]byte, _size *big.Int, _partOffset *big.Int) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadLocalData(&_PreimageOracle.TransactOpts, _ident, _localContext, _word, _size, _partOffset)
}

// LoadLocalData is a paid mutator transaction binding the contract method 0x52f0f3ad.
//
// Solidity: function loadLocalData(uint256 _ident, bytes32 _localContext, bytes32 _word, uint256 _size, uint256 _partOffset) returns(bytes32 key_)
func (_PreimageOracle *PreimageOracleTransactorSession) LoadLocalData(_ident *big.Int, _localContext [32]byte, _word [32]byte, _size *big.Int, _partOffset *big.Int) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadLocalData(&_PreimageOracle.TransactOpts, _ident, _localContext, _word, _size, _partOffset)
}

// LoadPrecompilePreimagePart is a paid mutator transaction binding the contract method 0x04697c78.
//
// Solidity: function loadPrecompilePreimagePart(uint256 _partOffset, address _precompile, bytes _input) returns()
func (_PreimageOracle *PreimageOracleTransactor) LoadPrecompilePreimagePart(opts *bind.TransactOpts, _partOffset *big.Int, _precompile common.Address, _input []byte) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "loadPrecompilePreimagePart", _partOffset, _precompile, _input)
}

// LoadPrecompilePreimagePart is a paid mutator transaction binding the contract method 0x04697c78.
//
// Solidity: function loadPrecompilePreimagePart(uint256 _partOffset, address _precompile, bytes _input) returns()
func (_PreimageOracle *PreimageOracleSession) LoadPrecompilePreimagePart(_partOffset *big.Int, _precompile common.Address, _input []byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadPrecompilePreimagePart(&_PreimageOracle.TransactOpts, _partOffset, _precompile, _input)
}

// LoadPrecompilePreimagePart is a paid mutator transaction binding the contract method 0x04697c78.
//
// Solidity: function loadPrecompilePreimagePart(uint256 _partOffset, address _precompile, bytes _input) returns()
func (_PreimageOracle *PreimageOracleTransactorSession) LoadPrecompilePreimagePart(_partOffset *big.Int, _precompile common.Address, _input []byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadPrecompilePreimagePart(&_PreimageOracle.TransactOpts, _partOffset, _precompile, _input)
}

// LoadSha256PreimagePart is a paid mutator transaction binding the contract method 0x8dc4be11.
//
// Solidity: function loadSha256PreimagePart(uint256 _partOffset, bytes _preimage) returns()
func (_PreimageOracle *PreimageOracleTransactor) LoadSha256PreimagePart(opts *bind.TransactOpts, _partOffset *big.Int, _preimage []byte) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "loadSha256PreimagePart", _partOffset, _preimage)
}

// LoadSha256PreimagePart is a paid mutator transaction binding the contract method 0x8dc4be11.
//
// Solidity: function loadSha256PreimagePart(uint256 _partOffset, bytes _preimage) returns()
func (_PreimageOracle *PreimageOracleSession) LoadSha256PreimagePart(_partOffset *big.Int, _preimage []byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadSha256PreimagePart(&_PreimageOracle.TransactOpts, _partOffset, _preimage)
}

// LoadSha256PreimagePart is a paid mutator transaction binding the contract method 0x8dc4be11.
//
// Solidity: function loadSha256PreimagePart(uint256 _partOffset, bytes _preimage) returns()
func (_PreimageOracle *PreimageOracleTransactorSession) LoadSha256PreimagePart(_partOffset *big.Int, _preimage []byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.LoadSha256PreimagePart(&_PreimageOracle.TransactOpts, _partOffset, _preimage)
}

// SqueezeLPP is a paid mutator transaction binding the contract method 0xd18534b5.
//
// Solidity: function squeezeLPP(address _claimant, uint256 _uuid, (uint64[25]) _stateMatrix, (bytes,uint256,bytes32) _preState, bytes32[] _preStateProof, (bytes,uint256,bytes32) _postState, bytes32[] _postStateProof) returns()
func (_PreimageOracle *PreimageOracleTransactor) SqueezeLPP(opts *bind.TransactOpts, _claimant common.Address, _uuid *big.Int, _stateMatrix LibKeccakStateMatrix, _preState PreimageOracleLeaf, _preStateProof [][32]byte, _postState PreimageOracleLeaf, _postStateProof [][32]byte) (*types.Transaction, error) {
	return _PreimageOracle.contract.Transact(opts, "squeezeLPP", _claimant, _uuid, _stateMatrix, _preState, _preStateProof, _postState, _postStateProof)
}

// SqueezeLPP is a paid mutator transaction binding the contract method 0xd18534b5.
//
// Solidity: function squeezeLPP(address _claimant, uint256 _uuid, (uint64[25]) _stateMatrix, (bytes,uint256,bytes32) _preState, bytes32[] _preStateProof, (bytes,uint256,bytes32) _postState, bytes32[] _postStateProof) returns()
func (_PreimageOracle *PreimageOracleSession) SqueezeLPP(_claimant common.Address, _uuid *big.Int, _stateMatrix LibKeccakStateMatrix, _preState PreimageOracleLeaf, _preStateProof [][32]byte, _postState PreimageOracleLeaf, _postStateProof [][32]byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.SqueezeLPP(&_PreimageOracle.TransactOpts, _claimant, _uuid, _stateMatrix, _preState, _preStateProof, _postState, _postStateProof)
}

// SqueezeLPP is a paid mutator transaction binding the contract method 0xd18534b5.
//
// Solidity: function squeezeLPP(address _claimant, uint256 _uuid, (uint64[25]) _stateMatrix, (bytes,uint256,bytes32) _preState, bytes32[] _preStateProof, (bytes,uint256,bytes32) _postState, bytes32[] _postStateProof) returns()
func (_PreimageOracle *PreimageOracleTransactorSession) SqueezeLPP(_claimant common.Address, _uuid *big.Int, _stateMatrix LibKeccakStateMatrix, _preState PreimageOracleLeaf, _preStateProof [][32]byte, _postState PreimageOracleLeaf, _postStateProof [][32]byte) (*types.Transaction, error) {
	return _PreimageOracle.Contract.SqueezeLPP(&_PreimageOracle.TransactOpts, _claimant, _uuid, _stateMatrix, _preState, _preStateProof, _postState, _postStateProof)
}

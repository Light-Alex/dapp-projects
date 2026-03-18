// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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

// MtkContractsStake is an auto generated low-level Go binding around an user-defined struct.
type MtkContractsStake struct {
	StakeId    *big.Int
	Amount     *big.Int
	StartTime  *big.Int
	EndTime    *big.Int
	RewardRate *big.Int
	IsActive   bool
	Period     uint8
}

// MtkContractsMetaData contains all meta data concerning the MtkContracts contract.
var MtkContractsMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_mtkToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"apy\",\"inputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumMtkContracts.StakingPeriod\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"calculateReward\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"stakeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"totalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"durations\",\"inputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumMtkContracts.StakingPeriod\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUserActiveStakes\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structMtkContracts.Stake[]\",\"components\":[{\"name\":\"stakeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"startTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"endTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"rewardRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"isActive\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"period\",\"type\":\"uint8\",\"internalType\":\"enumMtkContracts.StakingPeriod\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isStakeExpired\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"stakeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"stake\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"period\",\"type\":\"uint8\",\"internalType\":\"enumMtkContracts.StakingPeriod\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"stakeIdToOwner\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"stakingToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"userStakes\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"stakeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"startTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"endTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"rewardRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"isActive\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"period\",\"type\":\"uint8\",\"internalType\":\"enumMtkContracts.StakingPeriod\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"withdraw\",\"inputs\":[{\"name\":\"stakeId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"Staked\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"stakeId\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"period\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumMtkContracts.StakingPeriod\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Withdrawn\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"stakeId\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"principal\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"reward\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false}]",
}

// MtkContractsABI is the input ABI used to generate the binding from.
// Deprecated: Use MtkContractsMetaData.ABI instead.
var MtkContractsABI = MtkContractsMetaData.ABI

// MtkContracts is an auto generated Go binding around an Ethereum contract.
type MtkContracts struct {
	MtkContractsCaller     // Read-only binding to the contract
	MtkContractsTransactor // Write-only binding to the contract
	MtkContractsFilterer   // Log filterer for contract events
}

// MtkContractsCaller is an auto generated read-only Go binding around an Ethereum contract.
type MtkContractsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MtkContractsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MtkContractsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MtkContractsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MtkContractsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MtkContractsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MtkContractsSession struct {
	Contract     *MtkContracts     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MtkContractsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MtkContractsCallerSession struct {
	Contract *MtkContractsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// MtkContractsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MtkContractsTransactorSession struct {
	Contract     *MtkContractsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// MtkContractsRaw is an auto generated low-level Go binding around an Ethereum contract.
type MtkContractsRaw struct {
	Contract *MtkContracts // Generic contract binding to access the raw methods on
}

// MtkContractsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MtkContractsCallerRaw struct {
	Contract *MtkContractsCaller // Generic read-only contract binding to access the raw methods on
}

// MtkContractsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MtkContractsTransactorRaw struct {
	Contract *MtkContractsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMtkContracts creates a new instance of MtkContracts, bound to a specific deployed contract.
func NewMtkContracts(address common.Address, backend bind.ContractBackend) (*MtkContracts, error) {
	contract, err := bindMtkContracts(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MtkContracts{MtkContractsCaller: MtkContractsCaller{contract: contract}, MtkContractsTransactor: MtkContractsTransactor{contract: contract}, MtkContractsFilterer: MtkContractsFilterer{contract: contract}}, nil
}

// NewMtkContractsCaller creates a new read-only instance of MtkContracts, bound to a specific deployed contract.
func NewMtkContractsCaller(address common.Address, caller bind.ContractCaller) (*MtkContractsCaller, error) {
	contract, err := bindMtkContracts(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MtkContractsCaller{contract: contract}, nil
}

// NewMtkContractsTransactor creates a new write-only instance of MtkContracts, bound to a specific deployed contract.
func NewMtkContractsTransactor(address common.Address, transactor bind.ContractTransactor) (*MtkContractsTransactor, error) {
	contract, err := bindMtkContracts(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MtkContractsTransactor{contract: contract}, nil
}

// NewMtkContractsFilterer creates a new log filterer instance of MtkContracts, bound to a specific deployed contract.
func NewMtkContractsFilterer(address common.Address, filterer bind.ContractFilterer) (*MtkContractsFilterer, error) {
	contract, err := bindMtkContracts(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MtkContractsFilterer{contract: contract}, nil
}

// bindMtkContracts binds a generic wrapper to an already deployed contract.
func bindMtkContracts(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MtkContractsMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MtkContracts *MtkContractsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MtkContracts.Contract.MtkContractsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MtkContracts *MtkContractsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MtkContracts.Contract.MtkContractsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MtkContracts *MtkContractsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MtkContracts.Contract.MtkContractsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MtkContracts *MtkContractsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MtkContracts.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MtkContracts *MtkContractsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MtkContracts.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MtkContracts *MtkContractsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MtkContracts.Contract.contract.Transact(opts, method, params...)
}

// Apy is a free data retrieval call binding the contract method 0x1f1accb2.
//
// Solidity: function apy(uint8 ) view returns(uint256)
func (_MtkContracts *MtkContractsCaller) Apy(opts *bind.CallOpts, arg0 uint8) (*big.Int, error) {
	var out []interface{}
	err := _MtkContracts.contract.Call(opts, &out, "apy", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Apy is a free data retrieval call binding the contract method 0x1f1accb2.
//
// Solidity: function apy(uint8 ) view returns(uint256)
func (_MtkContracts *MtkContractsSession) Apy(arg0 uint8) (*big.Int, error) {
	return _MtkContracts.Contract.Apy(&_MtkContracts.CallOpts, arg0)
}

// Apy is a free data retrieval call binding the contract method 0x1f1accb2.
//
// Solidity: function apy(uint8 ) view returns(uint256)
func (_MtkContracts *MtkContractsCallerSession) Apy(arg0 uint8) (*big.Int, error) {
	return _MtkContracts.Contract.Apy(&_MtkContracts.CallOpts, arg0)
}

// CalculateReward is a free data retrieval call binding the contract method 0x1852e8d9.
//
// Solidity: function calculateReward(address user, uint256 stakeId) view returns(uint256 totalAmount)
func (_MtkContracts *MtkContractsCaller) CalculateReward(opts *bind.CallOpts, user common.Address, stakeId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _MtkContracts.contract.Call(opts, &out, "calculateReward", user, stakeId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateReward is a free data retrieval call binding the contract method 0x1852e8d9.
//
// Solidity: function calculateReward(address user, uint256 stakeId) view returns(uint256 totalAmount)
func (_MtkContracts *MtkContractsSession) CalculateReward(user common.Address, stakeId *big.Int) (*big.Int, error) {
	return _MtkContracts.Contract.CalculateReward(&_MtkContracts.CallOpts, user, stakeId)
}

// CalculateReward is a free data retrieval call binding the contract method 0x1852e8d9.
//
// Solidity: function calculateReward(address user, uint256 stakeId) view returns(uint256 totalAmount)
func (_MtkContracts *MtkContractsCallerSession) CalculateReward(user common.Address, stakeId *big.Int) (*big.Int, error) {
	return _MtkContracts.Contract.CalculateReward(&_MtkContracts.CallOpts, user, stakeId)
}

// Durations is a free data retrieval call binding the contract method 0x0ae355d3.
//
// Solidity: function durations(uint8 ) view returns(uint256)
func (_MtkContracts *MtkContractsCaller) Durations(opts *bind.CallOpts, arg0 uint8) (*big.Int, error) {
	var out []interface{}
	err := _MtkContracts.contract.Call(opts, &out, "durations", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Durations is a free data retrieval call binding the contract method 0x0ae355d3.
//
// Solidity: function durations(uint8 ) view returns(uint256)
func (_MtkContracts *MtkContractsSession) Durations(arg0 uint8) (*big.Int, error) {
	return _MtkContracts.Contract.Durations(&_MtkContracts.CallOpts, arg0)
}

// Durations is a free data retrieval call binding the contract method 0x0ae355d3.
//
// Solidity: function durations(uint8 ) view returns(uint256)
func (_MtkContracts *MtkContractsCallerSession) Durations(arg0 uint8) (*big.Int, error) {
	return _MtkContracts.Contract.Durations(&_MtkContracts.CallOpts, arg0)
}

// GetUserActiveStakes is a free data retrieval call binding the contract method 0xa262ab35.
//
// Solidity: function getUserActiveStakes(address user) view returns((uint256,uint256,uint256,uint256,uint256,bool,uint8)[])
func (_MtkContracts *MtkContractsCaller) GetUserActiveStakes(opts *bind.CallOpts, user common.Address) ([]MtkContractsStake, error) {
	var out []interface{}
	err := _MtkContracts.contract.Call(opts, &out, "getUserActiveStakes", user)

	if err != nil {
		return *new([]MtkContractsStake), err
	}

	out0 := *abi.ConvertType(out[0], new([]MtkContractsStake)).(*[]MtkContractsStake)

	return out0, err

}

// GetUserActiveStakes is a free data retrieval call binding the contract method 0xa262ab35.
//
// Solidity: function getUserActiveStakes(address user) view returns((uint256,uint256,uint256,uint256,uint256,bool,uint8)[])
func (_MtkContracts *MtkContractsSession) GetUserActiveStakes(user common.Address) ([]MtkContractsStake, error) {
	return _MtkContracts.Contract.GetUserActiveStakes(&_MtkContracts.CallOpts, user)
}

// GetUserActiveStakes is a free data retrieval call binding the contract method 0xa262ab35.
//
// Solidity: function getUserActiveStakes(address user) view returns((uint256,uint256,uint256,uint256,uint256,bool,uint8)[])
func (_MtkContracts *MtkContractsCallerSession) GetUserActiveStakes(user common.Address) ([]MtkContractsStake, error) {
	return _MtkContracts.Contract.GetUserActiveStakes(&_MtkContracts.CallOpts, user)
}

// IsStakeExpired is a free data retrieval call binding the contract method 0xe65d8496.
//
// Solidity: function isStakeExpired(address user, uint256 stakeId) view returns(bool)
func (_MtkContracts *MtkContractsCaller) IsStakeExpired(opts *bind.CallOpts, user common.Address, stakeId *big.Int) (bool, error) {
	var out []interface{}
	err := _MtkContracts.contract.Call(opts, &out, "isStakeExpired", user, stakeId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsStakeExpired is a free data retrieval call binding the contract method 0xe65d8496.
//
// Solidity: function isStakeExpired(address user, uint256 stakeId) view returns(bool)
func (_MtkContracts *MtkContractsSession) IsStakeExpired(user common.Address, stakeId *big.Int) (bool, error) {
	return _MtkContracts.Contract.IsStakeExpired(&_MtkContracts.CallOpts, user, stakeId)
}

// IsStakeExpired is a free data retrieval call binding the contract method 0xe65d8496.
//
// Solidity: function isStakeExpired(address user, uint256 stakeId) view returns(bool)
func (_MtkContracts *MtkContractsCallerSession) IsStakeExpired(user common.Address, stakeId *big.Int) (bool, error) {
	return _MtkContracts.Contract.IsStakeExpired(&_MtkContracts.CallOpts, user, stakeId)
}

// StakeIdToOwner is a free data retrieval call binding the contract method 0x3f9d8950.
//
// Solidity: function stakeIdToOwner(uint256 ) view returns(address)
func (_MtkContracts *MtkContractsCaller) StakeIdToOwner(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _MtkContracts.contract.Call(opts, &out, "stakeIdToOwner", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakeIdToOwner is a free data retrieval call binding the contract method 0x3f9d8950.
//
// Solidity: function stakeIdToOwner(uint256 ) view returns(address)
func (_MtkContracts *MtkContractsSession) StakeIdToOwner(arg0 *big.Int) (common.Address, error) {
	return _MtkContracts.Contract.StakeIdToOwner(&_MtkContracts.CallOpts, arg0)
}

// StakeIdToOwner is a free data retrieval call binding the contract method 0x3f9d8950.
//
// Solidity: function stakeIdToOwner(uint256 ) view returns(address)
func (_MtkContracts *MtkContractsCallerSession) StakeIdToOwner(arg0 *big.Int) (common.Address, error) {
	return _MtkContracts.Contract.StakeIdToOwner(&_MtkContracts.CallOpts, arg0)
}

// StakingToken is a free data retrieval call binding the contract method 0x72f702f3.
//
// Solidity: function stakingToken() view returns(address)
func (_MtkContracts *MtkContractsCaller) StakingToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MtkContracts.contract.Call(opts, &out, "stakingToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakingToken is a free data retrieval call binding the contract method 0x72f702f3.
//
// Solidity: function stakingToken() view returns(address)
func (_MtkContracts *MtkContractsSession) StakingToken() (common.Address, error) {
	return _MtkContracts.Contract.StakingToken(&_MtkContracts.CallOpts)
}

// StakingToken is a free data retrieval call binding the contract method 0x72f702f3.
//
// Solidity: function stakingToken() view returns(address)
func (_MtkContracts *MtkContractsCallerSession) StakingToken() (common.Address, error) {
	return _MtkContracts.Contract.StakingToken(&_MtkContracts.CallOpts)
}

// UserStakes is a free data retrieval call binding the contract method 0xb5d5b5fa.
//
// Solidity: function userStakes(address , uint256 ) view returns(uint256 stakeId, uint256 amount, uint256 startTime, uint256 endTime, uint256 rewardRate, bool isActive, uint8 period)
func (_MtkContracts *MtkContractsCaller) UserStakes(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	StakeId    *big.Int
	Amount     *big.Int
	StartTime  *big.Int
	EndTime    *big.Int
	RewardRate *big.Int
	IsActive   bool
	Period     uint8
}, error) {
	var out []interface{}
	err := _MtkContracts.contract.Call(opts, &out, "userStakes", arg0, arg1)

	outstruct := new(struct {
		StakeId    *big.Int
		Amount     *big.Int
		StartTime  *big.Int
		EndTime    *big.Int
		RewardRate *big.Int
		IsActive   bool
		Period     uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.StakeId = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Amount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.StartTime = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.EndTime = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.RewardRate = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.IsActive = *abi.ConvertType(out[5], new(bool)).(*bool)
	outstruct.Period = *abi.ConvertType(out[6], new(uint8)).(*uint8)

	return *outstruct, err

}

// UserStakes is a free data retrieval call binding the contract method 0xb5d5b5fa.
//
// Solidity: function userStakes(address , uint256 ) view returns(uint256 stakeId, uint256 amount, uint256 startTime, uint256 endTime, uint256 rewardRate, bool isActive, uint8 period)
func (_MtkContracts *MtkContractsSession) UserStakes(arg0 common.Address, arg1 *big.Int) (struct {
	StakeId    *big.Int
	Amount     *big.Int
	StartTime  *big.Int
	EndTime    *big.Int
	RewardRate *big.Int
	IsActive   bool
	Period     uint8
}, error) {
	return _MtkContracts.Contract.UserStakes(&_MtkContracts.CallOpts, arg0, arg1)
}

// UserStakes is a free data retrieval call binding the contract method 0xb5d5b5fa.
//
// Solidity: function userStakes(address , uint256 ) view returns(uint256 stakeId, uint256 amount, uint256 startTime, uint256 endTime, uint256 rewardRate, bool isActive, uint8 period)
func (_MtkContracts *MtkContractsCallerSession) UserStakes(arg0 common.Address, arg1 *big.Int) (struct {
	StakeId    *big.Int
	Amount     *big.Int
	StartTime  *big.Int
	EndTime    *big.Int
	RewardRate *big.Int
	IsActive   bool
	Period     uint8
}, error) {
	return _MtkContracts.Contract.UserStakes(&_MtkContracts.CallOpts, arg0, arg1)
}

// Stake is a paid mutator transaction binding the contract method 0x10087fb1.
//
// Solidity: function stake(uint256 amount, uint8 period) returns()
func (_MtkContracts *MtkContractsTransactor) Stake(opts *bind.TransactOpts, amount *big.Int, period uint8) (*types.Transaction, error) {
	return _MtkContracts.contract.Transact(opts, "stake", amount, period)
}

// Stake is a paid mutator transaction binding the contract method 0x10087fb1.
//
// Solidity: function stake(uint256 amount, uint8 period) returns()
func (_MtkContracts *MtkContractsSession) Stake(amount *big.Int, period uint8) (*types.Transaction, error) {
	return _MtkContracts.Contract.Stake(&_MtkContracts.TransactOpts, amount, period)
}

// Stake is a paid mutator transaction binding the contract method 0x10087fb1.
//
// Solidity: function stake(uint256 amount, uint8 period) returns()
func (_MtkContracts *MtkContractsTransactorSession) Stake(amount *big.Int, period uint8) (*types.Transaction, error) {
	return _MtkContracts.Contract.Stake(&_MtkContracts.TransactOpts, amount, period)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 stakeId) returns()
func (_MtkContracts *MtkContractsTransactor) Withdraw(opts *bind.TransactOpts, stakeId *big.Int) (*types.Transaction, error) {
	return _MtkContracts.contract.Transact(opts, "withdraw", stakeId)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 stakeId) returns()
func (_MtkContracts *MtkContractsSession) Withdraw(stakeId *big.Int) (*types.Transaction, error) {
	return _MtkContracts.Contract.Withdraw(&_MtkContracts.TransactOpts, stakeId)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 stakeId) returns()
func (_MtkContracts *MtkContractsTransactorSession) Withdraw(stakeId *big.Int) (*types.Transaction, error) {
	return _MtkContracts.Contract.Withdraw(&_MtkContracts.TransactOpts, stakeId)
}

// MtkContractsStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the MtkContracts contract.
type MtkContractsStakedIterator struct {
	Event *MtkContractsStaked // Event containing the contract specifics and raw log

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
func (it *MtkContractsStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MtkContractsStaked)
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
		it.Event = new(MtkContractsStaked)
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
func (it *MtkContractsStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MtkContractsStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MtkContractsStaked represents a Staked event raised by the MtkContracts contract.
type MtkContractsStaked struct {
	User      common.Address
	StakeId   *big.Int
	Amount    *big.Int
	Period    uint8
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0xcc10169be2ad544347561e230939849af48d1714c052d7fe247d12f3decb4896.
//
// Solidity: event Staked(address indexed user, uint256 stakeId, uint256 amount, uint8 period, uint256 timestamp)
func (_MtkContracts *MtkContractsFilterer) FilterStaked(opts *bind.FilterOpts, user []common.Address) (*MtkContractsStakedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _MtkContracts.contract.FilterLogs(opts, "Staked", userRule)
	if err != nil {
		return nil, err
	}
	return &MtkContractsStakedIterator{contract: _MtkContracts.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0xcc10169be2ad544347561e230939849af48d1714c052d7fe247d12f3decb4896.
//
// Solidity: event Staked(address indexed user, uint256 stakeId, uint256 amount, uint8 period, uint256 timestamp)
func (_MtkContracts *MtkContractsFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *MtkContractsStaked, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _MtkContracts.contract.WatchLogs(opts, "Staked", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MtkContractsStaked)
				if err := _MtkContracts.contract.UnpackLog(event, "Staked", log); err != nil {
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

// ParseStaked is a log parse operation binding the contract event 0xcc10169be2ad544347561e230939849af48d1714c052d7fe247d12f3decb4896.
//
// Solidity: event Staked(address indexed user, uint256 stakeId, uint256 amount, uint8 period, uint256 timestamp)
func (_MtkContracts *MtkContractsFilterer) ParseStaked(log types.Log) (*MtkContractsStaked, error) {
	event := new(MtkContractsStaked)
	if err := _MtkContracts.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MtkContractsWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the MtkContracts contract.
type MtkContractsWithdrawnIterator struct {
	Event *MtkContractsWithdrawn // Event containing the contract specifics and raw log

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
func (it *MtkContractsWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MtkContractsWithdrawn)
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
		it.Event = new(MtkContractsWithdrawn)
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
func (it *MtkContractsWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MtkContractsWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MtkContractsWithdrawn represents a Withdrawn event raised by the MtkContracts contract.
type MtkContractsWithdrawn struct {
	User        common.Address
	StakeId     *big.Int
	Principal   *big.Int
	Reward      *big.Int
	TotalAmount *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0x94ffd6b85c71b847775c89ef6496b93cee961bdc6ff827fd117f174f06f745ae.
//
// Solidity: event Withdrawn(address indexed user, uint256 stakeId, uint256 principal, uint256 reward, uint256 totalAmount)
func (_MtkContracts *MtkContractsFilterer) FilterWithdrawn(opts *bind.FilterOpts, user []common.Address) (*MtkContractsWithdrawnIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _MtkContracts.contract.FilterLogs(opts, "Withdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return &MtkContractsWithdrawnIterator{contract: _MtkContracts.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0x94ffd6b85c71b847775c89ef6496b93cee961bdc6ff827fd117f174f06f745ae.
//
// Solidity: event Withdrawn(address indexed user, uint256 stakeId, uint256 principal, uint256 reward, uint256 totalAmount)
func (_MtkContracts *MtkContractsFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *MtkContractsWithdrawn, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _MtkContracts.contract.WatchLogs(opts, "Withdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MtkContractsWithdrawn)
				if err := _MtkContracts.contract.UnpackLog(event, "Withdrawn", log); err != nil {
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

// ParseWithdrawn is a log parse operation binding the contract event 0x94ffd6b85c71b847775c89ef6496b93cee961bdc6ff827fd117f174f06f745ae.
//
// Solidity: event Withdrawn(address indexed user, uint256 stakeId, uint256 principal, uint256 reward, uint256 totalAmount)
func (_MtkContracts *MtkContractsFilterer) ParseWithdrawn(log types.Log) (*MtkContractsWithdrawn, error) {
	event := new(MtkContractsWithdrawn)
	if err := _MtkContracts.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

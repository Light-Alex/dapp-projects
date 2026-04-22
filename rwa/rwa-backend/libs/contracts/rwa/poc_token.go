// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package rwa

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

// PocTokenMetaData contains all meta data concerning the PocToken contract.
var PocTokenMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"gateContract_\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"BURNER_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"GATE_CONTRACT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MINTER_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"allowance\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"approve\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"burn\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"burnFrom\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"decimals\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleMember\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleMemberCount\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleMembers\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"name_\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"symbol_\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isBurner\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isMinter\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"mint\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"name\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOperator\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"symbol\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalSupply\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transfer\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferFrom\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"Approval\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"spender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TokensBurned\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TokensMinted\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Transfer\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"AddressCannotBeZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AmountCannotBeZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ERC20InsufficientAllowance\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"allowance\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"needed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ERC20InsufficientBalance\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"needed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ERC20InvalidApprover\",\"inputs\":[{\"name\":\"approver\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC20InvalidReceiver\",\"inputs\":[{\"name\":\"receiver\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC20InvalidSender\",\"inputs\":[{\"name\":\"sender\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC20InvalidSpender\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"GateContractCannotBeZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]}]",
}

// PocTokenABI is the input ABI used to generate the binding from.
// Deprecated: Use PocTokenMetaData.ABI instead.
var PocTokenABI = PocTokenMetaData.ABI

// PocToken is an auto generated Go binding around an Ethereum contract.
type PocToken struct {
	PocTokenCaller     // Read-only binding to the contract
	PocTokenTransactor // Write-only binding to the contract
	PocTokenFilterer   // Log filterer for contract events
}

// PocTokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type PocTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PocTokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PocTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PocTokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PocTokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PocTokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PocTokenSession struct {
	Contract     *PocToken         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PocTokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PocTokenCallerSession struct {
	Contract *PocTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// PocTokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PocTokenTransactorSession struct {
	Contract     *PocTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// PocTokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type PocTokenRaw struct {
	Contract *PocToken // Generic contract binding to access the raw methods on
}

// PocTokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PocTokenCallerRaw struct {
	Contract *PocTokenCaller // Generic read-only contract binding to access the raw methods on
}

// PocTokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PocTokenTransactorRaw struct {
	Contract *PocTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPocToken creates a new instance of PocToken, bound to a specific deployed contract.
func NewPocToken(address common.Address, backend bind.ContractBackend) (*PocToken, error) {
	contract, err := bindPocToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PocToken{PocTokenCaller: PocTokenCaller{contract: contract}, PocTokenTransactor: PocTokenTransactor{contract: contract}, PocTokenFilterer: PocTokenFilterer{contract: contract}}, nil
}

// NewPocTokenCaller creates a new read-only instance of PocToken, bound to a specific deployed contract.
func NewPocTokenCaller(address common.Address, caller bind.ContractCaller) (*PocTokenCaller, error) {
	contract, err := bindPocToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PocTokenCaller{contract: contract}, nil
}

// NewPocTokenTransactor creates a new write-only instance of PocToken, bound to a specific deployed contract.
func NewPocTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*PocTokenTransactor, error) {
	contract, err := bindPocToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PocTokenTransactor{contract: contract}, nil
}

// NewPocTokenFilterer creates a new log filterer instance of PocToken, bound to a specific deployed contract.
func NewPocTokenFilterer(address common.Address, filterer bind.ContractFilterer) (*PocTokenFilterer, error) {
	contract, err := bindPocToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PocTokenFilterer{contract: contract}, nil
}

// bindPocToken binds a generic wrapper to an already deployed contract.
func bindPocToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PocTokenMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PocToken *PocTokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PocToken.Contract.PocTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PocToken *PocTokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PocToken.Contract.PocTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PocToken *PocTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PocToken.Contract.PocTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PocToken *PocTokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PocToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PocToken *PocTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PocToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PocToken *PocTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PocToken.Contract.contract.Transact(opts, method, params...)
}

// BURNERROLE is a free data retrieval call binding the contract method 0x282c51f3.
//
// Solidity: function BURNER_ROLE() view returns(bytes32)
func (_PocToken *PocTokenCaller) BURNERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "BURNER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BURNERROLE is a free data retrieval call binding the contract method 0x282c51f3.
//
// Solidity: function BURNER_ROLE() view returns(bytes32)
func (_PocToken *PocTokenSession) BURNERROLE() ([32]byte, error) {
	return _PocToken.Contract.BURNERROLE(&_PocToken.CallOpts)
}

// BURNERROLE is a free data retrieval call binding the contract method 0x282c51f3.
//
// Solidity: function BURNER_ROLE() view returns(bytes32)
func (_PocToken *PocTokenCallerSession) BURNERROLE() ([32]byte, error) {
	return _PocToken.Contract.BURNERROLE(&_PocToken.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_PocToken *PocTokenCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_PocToken *PocTokenSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _PocToken.Contract.DEFAULTADMINROLE(&_PocToken.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_PocToken *PocTokenCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _PocToken.Contract.DEFAULTADMINROLE(&_PocToken.CallOpts)
}

// GATECONTRACT is a free data retrieval call binding the contract method 0xedbf322d.
//
// Solidity: function GATE_CONTRACT() view returns(address)
func (_PocToken *PocTokenCaller) GATECONTRACT(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "GATE_CONTRACT")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GATECONTRACT is a free data retrieval call binding the contract method 0xedbf322d.
//
// Solidity: function GATE_CONTRACT() view returns(address)
func (_PocToken *PocTokenSession) GATECONTRACT() (common.Address, error) {
	return _PocToken.Contract.GATECONTRACT(&_PocToken.CallOpts)
}

// GATECONTRACT is a free data retrieval call binding the contract method 0xedbf322d.
//
// Solidity: function GATE_CONTRACT() view returns(address)
func (_PocToken *PocTokenCallerSession) GATECONTRACT() (common.Address, error) {
	return _PocToken.Contract.GATECONTRACT(&_PocToken.CallOpts)
}

// MINTERROLE is a free data retrieval call binding the contract method 0xd5391393.
//
// Solidity: function MINTER_ROLE() view returns(bytes32)
func (_PocToken *PocTokenCaller) MINTERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "MINTER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MINTERROLE is a free data retrieval call binding the contract method 0xd5391393.
//
// Solidity: function MINTER_ROLE() view returns(bytes32)
func (_PocToken *PocTokenSession) MINTERROLE() ([32]byte, error) {
	return _PocToken.Contract.MINTERROLE(&_PocToken.CallOpts)
}

// MINTERROLE is a free data retrieval call binding the contract method 0xd5391393.
//
// Solidity: function MINTER_ROLE() view returns(bytes32)
func (_PocToken *PocTokenCallerSession) MINTERROLE() ([32]byte, error) {
	return _PocToken.Contract.MINTERROLE(&_PocToken.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_PocToken *PocTokenCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_PocToken *PocTokenSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _PocToken.Contract.Allowance(&_PocToken.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_PocToken *PocTokenCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _PocToken.Contract.Allowance(&_PocToken.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_PocToken *PocTokenCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_PocToken *PocTokenSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _PocToken.Contract.BalanceOf(&_PocToken.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_PocToken *PocTokenCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _PocToken.Contract.BalanceOf(&_PocToken.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_PocToken *PocTokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_PocToken *PocTokenSession) Decimals() (uint8, error) {
	return _PocToken.Contract.Decimals(&_PocToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_PocToken *PocTokenCallerSession) Decimals() (uint8, error) {
	return _PocToken.Contract.Decimals(&_PocToken.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_PocToken *PocTokenCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_PocToken *PocTokenSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _PocToken.Contract.GetRoleAdmin(&_PocToken.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_PocToken *PocTokenCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _PocToken.Contract.GetRoleAdmin(&_PocToken.CallOpts, role)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_PocToken *PocTokenCaller) GetRoleMember(opts *bind.CallOpts, role [32]byte, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "getRoleMember", role, index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_PocToken *PocTokenSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _PocToken.Contract.GetRoleMember(&_PocToken.CallOpts, role, index)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_PocToken *PocTokenCallerSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _PocToken.Contract.GetRoleMember(&_PocToken.CallOpts, role, index)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_PocToken *PocTokenCaller) GetRoleMemberCount(opts *bind.CallOpts, role [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "getRoleMemberCount", role)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_PocToken *PocTokenSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _PocToken.Contract.GetRoleMemberCount(&_PocToken.CallOpts, role)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_PocToken *PocTokenCallerSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _PocToken.Contract.GetRoleMemberCount(&_PocToken.CallOpts, role)
}

// GetRoleMembers is a free data retrieval call binding the contract method 0xa3246ad3.
//
// Solidity: function getRoleMembers(bytes32 role) view returns(address[])
func (_PocToken *PocTokenCaller) GetRoleMembers(opts *bind.CallOpts, role [32]byte) ([]common.Address, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "getRoleMembers", role)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetRoleMembers is a free data retrieval call binding the contract method 0xa3246ad3.
//
// Solidity: function getRoleMembers(bytes32 role) view returns(address[])
func (_PocToken *PocTokenSession) GetRoleMembers(role [32]byte) ([]common.Address, error) {
	return _PocToken.Contract.GetRoleMembers(&_PocToken.CallOpts, role)
}

// GetRoleMembers is a free data retrieval call binding the contract method 0xa3246ad3.
//
// Solidity: function getRoleMembers(bytes32 role) view returns(address[])
func (_PocToken *PocTokenCallerSession) GetRoleMembers(role [32]byte) ([]common.Address, error) {
	return _PocToken.Contract.GetRoleMembers(&_PocToken.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_PocToken *PocTokenCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_PocToken *PocTokenSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _PocToken.Contract.HasRole(&_PocToken.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_PocToken *PocTokenCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _PocToken.Contract.HasRole(&_PocToken.CallOpts, role, account)
}

// IsBurner is a free data retrieval call binding the contract method 0x4334614a.
//
// Solidity: function isBurner(address account) view returns(bool)
func (_PocToken *PocTokenCaller) IsBurner(opts *bind.CallOpts, account common.Address) (bool, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "isBurner", account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBurner is a free data retrieval call binding the contract method 0x4334614a.
//
// Solidity: function isBurner(address account) view returns(bool)
func (_PocToken *PocTokenSession) IsBurner(account common.Address) (bool, error) {
	return _PocToken.Contract.IsBurner(&_PocToken.CallOpts, account)
}

// IsBurner is a free data retrieval call binding the contract method 0x4334614a.
//
// Solidity: function isBurner(address account) view returns(bool)
func (_PocToken *PocTokenCallerSession) IsBurner(account common.Address) (bool, error) {
	return _PocToken.Contract.IsBurner(&_PocToken.CallOpts, account)
}

// IsMinter is a free data retrieval call binding the contract method 0xaa271e1a.
//
// Solidity: function isMinter(address account) view returns(bool)
func (_PocToken *PocTokenCaller) IsMinter(opts *bind.CallOpts, account common.Address) (bool, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "isMinter", account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsMinter is a free data retrieval call binding the contract method 0xaa271e1a.
//
// Solidity: function isMinter(address account) view returns(bool)
func (_PocToken *PocTokenSession) IsMinter(account common.Address) (bool, error) {
	return _PocToken.Contract.IsMinter(&_PocToken.CallOpts, account)
}

// IsMinter is a free data retrieval call binding the contract method 0xaa271e1a.
//
// Solidity: function isMinter(address account) view returns(bool)
func (_PocToken *PocTokenCallerSession) IsMinter(account common.Address) (bool, error) {
	return _PocToken.Contract.IsMinter(&_PocToken.CallOpts, account)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_PocToken *PocTokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_PocToken *PocTokenSession) Name() (string, error) {
	return _PocToken.Contract.Name(&_PocToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_PocToken *PocTokenCallerSession) Name() (string, error) {
	return _PocToken.Contract.Name(&_PocToken.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_PocToken *PocTokenCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_PocToken *PocTokenSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _PocToken.Contract.SupportsInterface(&_PocToken.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_PocToken *PocTokenCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _PocToken.Contract.SupportsInterface(&_PocToken.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_PocToken *PocTokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_PocToken *PocTokenSession) Symbol() (string, error) {
	return _PocToken.Contract.Symbol(&_PocToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_PocToken *PocTokenCallerSession) Symbol() (string, error) {
	return _PocToken.Contract.Symbol(&_PocToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_PocToken *PocTokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PocToken.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_PocToken *PocTokenSession) TotalSupply() (*big.Int, error) {
	return _PocToken.Contract.TotalSupply(&_PocToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_PocToken *PocTokenCallerSession) TotalSupply() (*big.Int, error) {
	return _PocToken.Contract.TotalSupply(&_PocToken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_PocToken *PocTokenTransactor) Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "approve", spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_PocToken *PocTokenSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.Approve(&_PocToken.TransactOpts, spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_PocToken *PocTokenTransactorSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.Approve(&_PocToken.TransactOpts, spender, value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 amount) returns()
func (_PocToken *PocTokenTransactor) Burn(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "burn", amount)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 amount) returns()
func (_PocToken *PocTokenSession) Burn(amount *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.Burn(&_PocToken.TransactOpts, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 amount) returns()
func (_PocToken *PocTokenTransactorSession) Burn(amount *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.Burn(&_PocToken.TransactOpts, amount)
}

// BurnFrom is a paid mutator transaction binding the contract method 0x79cc6790.
//
// Solidity: function burnFrom(address from, uint256 amount) returns()
func (_PocToken *PocTokenTransactor) BurnFrom(opts *bind.TransactOpts, from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "burnFrom", from, amount)
}

// BurnFrom is a paid mutator transaction binding the contract method 0x79cc6790.
//
// Solidity: function burnFrom(address from, uint256 amount) returns()
func (_PocToken *PocTokenSession) BurnFrom(from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.BurnFrom(&_PocToken.TransactOpts, from, amount)
}

// BurnFrom is a paid mutator transaction binding the contract method 0x79cc6790.
//
// Solidity: function burnFrom(address from, uint256 amount) returns()
func (_PocToken *PocTokenTransactorSession) BurnFrom(from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.BurnFrom(&_PocToken.TransactOpts, from, amount)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_PocToken *PocTokenTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_PocToken *PocTokenSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _PocToken.Contract.GrantRole(&_PocToken.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_PocToken *PocTokenTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _PocToken.Contract.GrantRole(&_PocToken.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name_, string symbol_) returns()
func (_PocToken *PocTokenTransactor) Initialize(opts *bind.TransactOpts, name_ string, symbol_ string) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "initialize", name_, symbol_)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name_, string symbol_) returns()
func (_PocToken *PocTokenSession) Initialize(name_ string, symbol_ string) (*types.Transaction, error) {
	return _PocToken.Contract.Initialize(&_PocToken.TransactOpts, name_, symbol_)
}

// Initialize is a paid mutator transaction binding the contract method 0x4cd88b76.
//
// Solidity: function initialize(string name_, string symbol_) returns()
func (_PocToken *PocTokenTransactorSession) Initialize(name_ string, symbol_ string) (*types.Transaction, error) {
	return _PocToken.Contract.Initialize(&_PocToken.TransactOpts, name_, symbol_)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 amount) returns()
func (_PocToken *PocTokenTransactor) Mint(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "mint", to, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 amount) returns()
func (_PocToken *PocTokenSession) Mint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.Mint(&_PocToken.TransactOpts, to, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 amount) returns()
func (_PocToken *PocTokenTransactorSession) Mint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.Mint(&_PocToken.TransactOpts, to, amount)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_PocToken *PocTokenTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_PocToken *PocTokenSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _PocToken.Contract.RenounceRole(&_PocToken.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_PocToken *PocTokenTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _PocToken.Contract.RenounceRole(&_PocToken.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_PocToken *PocTokenTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_PocToken *PocTokenSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _PocToken.Contract.RevokeRole(&_PocToken.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_PocToken *PocTokenTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _PocToken.Contract.RevokeRole(&_PocToken.TransactOpts, role, account)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address operator) returns()
func (_PocToken *PocTokenTransactor) SetOperator(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "setOperator", operator)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address operator) returns()
func (_PocToken *PocTokenSession) SetOperator(operator common.Address) (*types.Transaction, error) {
	return _PocToken.Contract.SetOperator(&_PocToken.TransactOpts, operator)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address operator) returns()
func (_PocToken *PocTokenTransactorSession) SetOperator(operator common.Address) (*types.Transaction, error) {
	return _PocToken.Contract.SetOperator(&_PocToken.TransactOpts, operator)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_PocToken *PocTokenTransactor) Transfer(opts *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "transfer", to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_PocToken *PocTokenSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.Transfer(&_PocToken.TransactOpts, to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_PocToken *PocTokenTransactorSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.Transfer(&_PocToken.TransactOpts, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_PocToken *PocTokenTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _PocToken.contract.Transact(opts, "transferFrom", from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_PocToken *PocTokenSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.TransferFrom(&_PocToken.TransactOpts, from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_PocToken *PocTokenTransactorSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _PocToken.Contract.TransferFrom(&_PocToken.TransactOpts, from, to, value)
}

// PocTokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the PocToken contract.
type PocTokenApprovalIterator struct {
	Event *PocTokenApproval // Event containing the contract specifics and raw log

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
func (it *PocTokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PocTokenApproval)
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
		it.Event = new(PocTokenApproval)
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
func (it *PocTokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PocTokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PocTokenApproval represents a Approval event raised by the PocToken contract.
type PocTokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_PocToken *PocTokenFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*PocTokenApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _PocToken.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &PocTokenApprovalIterator{contract: _PocToken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_PocToken *PocTokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *PocTokenApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _PocToken.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PocTokenApproval)
				if err := _PocToken.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_PocToken *PocTokenFilterer) ParseApproval(log types.Log) (*PocTokenApproval, error) {
	event := new(PocTokenApproval)
	if err := _PocToken.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PocTokenInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the PocToken contract.
type PocTokenInitializedIterator struct {
	Event *PocTokenInitialized // Event containing the contract specifics and raw log

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
func (it *PocTokenInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PocTokenInitialized)
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
		it.Event = new(PocTokenInitialized)
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
func (it *PocTokenInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PocTokenInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PocTokenInitialized represents a Initialized event raised by the PocToken contract.
type PocTokenInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_PocToken *PocTokenFilterer) FilterInitialized(opts *bind.FilterOpts) (*PocTokenInitializedIterator, error) {

	logs, sub, err := _PocToken.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &PocTokenInitializedIterator{contract: _PocToken.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_PocToken *PocTokenFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *PocTokenInitialized) (event.Subscription, error) {

	logs, sub, err := _PocToken.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PocTokenInitialized)
				if err := _PocToken.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_PocToken *PocTokenFilterer) ParseInitialized(log types.Log) (*PocTokenInitialized, error) {
	event := new(PocTokenInitialized)
	if err := _PocToken.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PocTokenRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the PocToken contract.
type PocTokenRoleAdminChangedIterator struct {
	Event *PocTokenRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *PocTokenRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PocTokenRoleAdminChanged)
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
		it.Event = new(PocTokenRoleAdminChanged)
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
func (it *PocTokenRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PocTokenRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PocTokenRoleAdminChanged represents a RoleAdminChanged event raised by the PocToken contract.
type PocTokenRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_PocToken *PocTokenFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*PocTokenRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _PocToken.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &PocTokenRoleAdminChangedIterator{contract: _PocToken.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_PocToken *PocTokenFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *PocTokenRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _PocToken.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PocTokenRoleAdminChanged)
				if err := _PocToken.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_PocToken *PocTokenFilterer) ParseRoleAdminChanged(log types.Log) (*PocTokenRoleAdminChanged, error) {
	event := new(PocTokenRoleAdminChanged)
	if err := _PocToken.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PocTokenRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the PocToken contract.
type PocTokenRoleGrantedIterator struct {
	Event *PocTokenRoleGranted // Event containing the contract specifics and raw log

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
func (it *PocTokenRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PocTokenRoleGranted)
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
		it.Event = new(PocTokenRoleGranted)
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
func (it *PocTokenRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PocTokenRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PocTokenRoleGranted represents a RoleGranted event raised by the PocToken contract.
type PocTokenRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_PocToken *PocTokenFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*PocTokenRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _PocToken.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &PocTokenRoleGrantedIterator{contract: _PocToken.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_PocToken *PocTokenFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *PocTokenRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _PocToken.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PocTokenRoleGranted)
				if err := _PocToken.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_PocToken *PocTokenFilterer) ParseRoleGranted(log types.Log) (*PocTokenRoleGranted, error) {
	event := new(PocTokenRoleGranted)
	if err := _PocToken.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PocTokenRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the PocToken contract.
type PocTokenRoleRevokedIterator struct {
	Event *PocTokenRoleRevoked // Event containing the contract specifics and raw log

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
func (it *PocTokenRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PocTokenRoleRevoked)
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
		it.Event = new(PocTokenRoleRevoked)
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
func (it *PocTokenRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PocTokenRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PocTokenRoleRevoked represents a RoleRevoked event raised by the PocToken contract.
type PocTokenRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_PocToken *PocTokenFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*PocTokenRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _PocToken.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &PocTokenRoleRevokedIterator{contract: _PocToken.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_PocToken *PocTokenFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *PocTokenRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _PocToken.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PocTokenRoleRevoked)
				if err := _PocToken.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_PocToken *PocTokenFilterer) ParseRoleRevoked(log types.Log) (*PocTokenRoleRevoked, error) {
	event := new(PocTokenRoleRevoked)
	if err := _PocToken.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PocTokenTokensBurnedIterator is returned from FilterTokensBurned and is used to iterate over the raw logs and unpacked data for TokensBurned events raised by the PocToken contract.
type PocTokenTokensBurnedIterator struct {
	Event *PocTokenTokensBurned // Event containing the contract specifics and raw log

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
func (it *PocTokenTokensBurnedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PocTokenTokensBurned)
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
		it.Event = new(PocTokenTokensBurned)
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
func (it *PocTokenTokensBurnedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PocTokenTokensBurnedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PocTokenTokensBurned represents a TokensBurned event raised by the PocToken contract.
type PocTokenTokensBurned struct {
	From   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTokensBurned is a free log retrieval operation binding the contract event 0xfd38818f5291bf0bb3a2a48aadc06ba8757865d1dabd804585338aab3009dcb6.
//
// Solidity: event TokensBurned(address indexed from, uint256 amount)
func (_PocToken *PocTokenFilterer) FilterTokensBurned(opts *bind.FilterOpts, from []common.Address) (*PocTokenTokensBurnedIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _PocToken.contract.FilterLogs(opts, "TokensBurned", fromRule)
	if err != nil {
		return nil, err
	}
	return &PocTokenTokensBurnedIterator{contract: _PocToken.contract, event: "TokensBurned", logs: logs, sub: sub}, nil
}

// WatchTokensBurned is a free log subscription operation binding the contract event 0xfd38818f5291bf0bb3a2a48aadc06ba8757865d1dabd804585338aab3009dcb6.
//
// Solidity: event TokensBurned(address indexed from, uint256 amount)
func (_PocToken *PocTokenFilterer) WatchTokensBurned(opts *bind.WatchOpts, sink chan<- *PocTokenTokensBurned, from []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _PocToken.contract.WatchLogs(opts, "TokensBurned", fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PocTokenTokensBurned)
				if err := _PocToken.contract.UnpackLog(event, "TokensBurned", log); err != nil {
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

// ParseTokensBurned is a log parse operation binding the contract event 0xfd38818f5291bf0bb3a2a48aadc06ba8757865d1dabd804585338aab3009dcb6.
//
// Solidity: event TokensBurned(address indexed from, uint256 amount)
func (_PocToken *PocTokenFilterer) ParseTokensBurned(log types.Log) (*PocTokenTokensBurned, error) {
	event := new(PocTokenTokensBurned)
	if err := _PocToken.contract.UnpackLog(event, "TokensBurned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PocTokenTokensMintedIterator is returned from FilterTokensMinted and is used to iterate over the raw logs and unpacked data for TokensMinted events raised by the PocToken contract.
type PocTokenTokensMintedIterator struct {
	Event *PocTokenTokensMinted // Event containing the contract specifics and raw log

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
func (it *PocTokenTokensMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PocTokenTokensMinted)
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
		it.Event = new(PocTokenTokensMinted)
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
func (it *PocTokenTokensMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PocTokenTokensMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PocTokenTokensMinted represents a TokensMinted event raised by the PocToken contract.
type PocTokenTokensMinted struct {
	To     common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterTokensMinted is a free log retrieval operation binding the contract event 0x3f2c9d57c068687834f0de942a9babb9e5acab57d516d3480a3c16ee165a4273.
//
// Solidity: event TokensMinted(address indexed to, uint256 amount)
func (_PocToken *PocTokenFilterer) FilterTokensMinted(opts *bind.FilterOpts, to []common.Address) (*PocTokenTokensMintedIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _PocToken.contract.FilterLogs(opts, "TokensMinted", toRule)
	if err != nil {
		return nil, err
	}
	return &PocTokenTokensMintedIterator{contract: _PocToken.contract, event: "TokensMinted", logs: logs, sub: sub}, nil
}

// WatchTokensMinted is a free log subscription operation binding the contract event 0x3f2c9d57c068687834f0de942a9babb9e5acab57d516d3480a3c16ee165a4273.
//
// Solidity: event TokensMinted(address indexed to, uint256 amount)
func (_PocToken *PocTokenFilterer) WatchTokensMinted(opts *bind.WatchOpts, sink chan<- *PocTokenTokensMinted, to []common.Address) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _PocToken.contract.WatchLogs(opts, "TokensMinted", toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PocTokenTokensMinted)
				if err := _PocToken.contract.UnpackLog(event, "TokensMinted", log); err != nil {
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

// ParseTokensMinted is a log parse operation binding the contract event 0x3f2c9d57c068687834f0de942a9babb9e5acab57d516d3480a3c16ee165a4273.
//
// Solidity: event TokensMinted(address indexed to, uint256 amount)
func (_PocToken *PocTokenFilterer) ParseTokensMinted(log types.Log) (*PocTokenTokensMinted, error) {
	event := new(PocTokenTokensMinted)
	if err := _PocToken.contract.UnpackLog(event, "TokensMinted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PocTokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the PocToken contract.
type PocTokenTransferIterator struct {
	Event *PocTokenTransfer // Event containing the contract specifics and raw log

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
func (it *PocTokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PocTokenTransfer)
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
		it.Event = new(PocTokenTransfer)
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
func (it *PocTokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PocTokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PocTokenTransfer represents a Transfer event raised by the PocToken contract.
type PocTokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_PocToken *PocTokenFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*PocTokenTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _PocToken.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &PocTokenTransferIterator{contract: _PocToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_PocToken *PocTokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *PocTokenTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _PocToken.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PocTokenTransfer)
				if err := _PocToken.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_PocToken *PocTokenFilterer) ParseTransfer(log types.Log) (*PocTokenTransfer, error) {
	event := new(PocTokenTransfer)
	if err := _PocToken.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

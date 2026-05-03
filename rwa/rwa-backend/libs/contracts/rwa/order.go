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

// OrderContractOrder is an auto generated low-level Go binding around an user-defined struct.
type OrderContractOrder struct {
	Id          *big.Int
	OrderNumber *big.Int
	User        common.Address
	Symbol      string
	Qty         *big.Int
	EscrowAsset common.Address
	Amount      *big.Int
	Price       *big.Int
	Side        uint8
	OrderType   uint8
	Status      uint8
	Tif         uint8
}

// OrderMetaData contains all meta data concerning the Order contract.
var OrderMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"BACKEND_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"USDM\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPocToken\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"accountOrderSeq\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"cancelOrder\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"cancelOrderIntent\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getOrder\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structOrderContract.Order\",\"components\":[{\"name\":\"id\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"orderNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"symbol\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"qty\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"escrowAsset\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"side\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.Side\"},{\"name\":\"orderType\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.OrderType\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.Status\"},{\"name\":\"tif\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.TimeInForce\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOrderNumber\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleMember\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleMemberCount\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleMembers\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"usdm_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"admin_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"backend_\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"markExecuted\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"refundAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"nextOrderId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"orders\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"id\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"orderNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"symbol\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"qty\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"escrowAsset\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"side\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.Side\"},{\"name\":\"orderType\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.OrderType\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.Status\"},{\"name\":\"tif\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.TimeInForce\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setBackend\",\"inputs\":[{\"name\":\"backend\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSymbolToken\",\"inputs\":[{\"name\":\"symbol\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitOrder\",\"inputs\":[{\"name\":\"symbol\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"qty\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"side\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.Side\"},{\"name\":\"orderType\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.OrderType\"},{\"name\":\"tif\",\"type\":\"uint8\",\"internalType\":\"enumOrderContract.TimeInForce\"}],\"outputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"symbolToToken\",\"inputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPocToken\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"CancelRequested\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"blockTimestamp\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OrderCancelled\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"asset\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"refundAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"side\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOrderContract.Side\"},{\"name\":\"orderType\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOrderContract.OrderType\"},{\"name\":\"tif\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOrderContract.TimeInForce\"},{\"name\":\"previousStatus\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOrderContract.Status\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OrderExecuted\",\"inputs\":[{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"refundAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"tif\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOrderContract.TimeInForce\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OrderSubmitted\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"orderId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"symbol\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"qty\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"side\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOrderContract.Side\"},{\"name\":\"orderType\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOrderContract.OrderType\"},{\"name\":\"tif\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumOrderContract.TimeInForce\"},{\"name\":\"blockTimestamp\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"AlreadyCancelled\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AlreadyExecuted\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AmountZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidStatus\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotCancelRequested\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotFound\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotOwner\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroAddress\",\"inputs\":[]}]",
}

// OrderABI is the input ABI used to generate the binding from.
// Deprecated: Use OrderMetaData.ABI instead.
var OrderABI = OrderMetaData.ABI

// Order is an auto generated Go binding around an Ethereum contract.
type Order struct {
	OrderCaller     // Read-only binding to the contract
	OrderTransactor // Write-only binding to the contract
	OrderFilterer   // Log filterer for contract events
}

// OrderCaller is an auto generated read-only Go binding around an Ethereum contract.
type OrderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OrderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OrderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OrderSession struct {
	Contract     *Order            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OrderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OrderCallerSession struct {
	Contract *OrderCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// OrderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OrderTransactorSession struct {
	Contract     *OrderTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OrderRaw is an auto generated low-level Go binding around an Ethereum contract.
type OrderRaw struct {
	Contract *Order // Generic contract binding to access the raw methods on
}

// OrderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OrderCallerRaw struct {
	Contract *OrderCaller // Generic read-only contract binding to access the raw methods on
}

// OrderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OrderTransactorRaw struct {
	Contract *OrderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOrder creates a new instance of Order, bound to a specific deployed contract.
func NewOrder(address common.Address, backend bind.ContractBackend) (*Order, error) {
	contract, err := bindOrder(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Order{OrderCaller: OrderCaller{contract: contract}, OrderTransactor: OrderTransactor{contract: contract}, OrderFilterer: OrderFilterer{contract: contract}}, nil
}

// NewOrderCaller creates a new read-only instance of Order, bound to a specific deployed contract.
func NewOrderCaller(address common.Address, caller bind.ContractCaller) (*OrderCaller, error) {
	contract, err := bindOrder(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OrderCaller{contract: contract}, nil
}

// NewOrderTransactor creates a new write-only instance of Order, bound to a specific deployed contract.
func NewOrderTransactor(address common.Address, transactor bind.ContractTransactor) (*OrderTransactor, error) {
	contract, err := bindOrder(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OrderTransactor{contract: contract}, nil
}

// NewOrderFilterer creates a new log filterer instance of Order, bound to a specific deployed contract.
func NewOrderFilterer(address common.Address, filterer bind.ContractFilterer) (*OrderFilterer, error) {
	contract, err := bindOrder(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OrderFilterer{contract: contract}, nil
}

// bindOrder binds a generic wrapper to an already deployed contract.
func bindOrder(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OrderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Order *OrderRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Order.Contract.OrderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Order *OrderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Order.Contract.OrderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Order *OrderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Order.Contract.OrderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Order *OrderCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Order.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Order *OrderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Order.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Order *OrderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Order.Contract.contract.Transact(opts, method, params...)
}

// BACKENDROLE is a free data retrieval call binding the contract method 0x92c2becc.
//
// Solidity: function BACKEND_ROLE() view returns(bytes32)
func (_Order *OrderCaller) BACKENDROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "BACKEND_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BACKENDROLE is a free data retrieval call binding the contract method 0x92c2becc.
//
// Solidity: function BACKEND_ROLE() view returns(bytes32)
func (_Order *OrderSession) BACKENDROLE() ([32]byte, error) {
	return _Order.Contract.BACKENDROLE(&_Order.CallOpts)
}

// BACKENDROLE is a free data retrieval call binding the contract method 0x92c2becc.
//
// Solidity: function BACKEND_ROLE() view returns(bytes32)
func (_Order *OrderCallerSession) BACKENDROLE() ([32]byte, error) {
	return _Order.Contract.BACKENDROLE(&_Order.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Order *OrderCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Order *OrderSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Order.Contract.DEFAULTADMINROLE(&_Order.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Order *OrderCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Order.Contract.DEFAULTADMINROLE(&_Order.CallOpts)
}

// USDM is a free data retrieval call binding the contract method 0xa9da3cf2.
//
// Solidity: function USDM() view returns(address)
func (_Order *OrderCaller) USDM(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "USDM")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// USDM is a free data retrieval call binding the contract method 0xa9da3cf2.
//
// Solidity: function USDM() view returns(address)
func (_Order *OrderSession) USDM() (common.Address, error) {
	return _Order.Contract.USDM(&_Order.CallOpts)
}

// USDM is a free data retrieval call binding the contract method 0xa9da3cf2.
//
// Solidity: function USDM() view returns(address)
func (_Order *OrderCallerSession) USDM() (common.Address, error) {
	return _Order.Contract.USDM(&_Order.CallOpts)
}

// AccountOrderSeq is a free data retrieval call binding the contract method 0x4c0f006c.
//
// Solidity: function accountOrderSeq(address ) view returns(uint256)
func (_Order *OrderCaller) AccountOrderSeq(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "accountOrderSeq", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AccountOrderSeq is a free data retrieval call binding the contract method 0x4c0f006c.
//
// Solidity: function accountOrderSeq(address ) view returns(uint256)
func (_Order *OrderSession) AccountOrderSeq(arg0 common.Address) (*big.Int, error) {
	return _Order.Contract.AccountOrderSeq(&_Order.CallOpts, arg0)
}

// AccountOrderSeq is a free data retrieval call binding the contract method 0x4c0f006c.
//
// Solidity: function accountOrderSeq(address ) view returns(uint256)
func (_Order *OrderCallerSession) AccountOrderSeq(arg0 common.Address) (*big.Int, error) {
	return _Order.Contract.AccountOrderSeq(&_Order.CallOpts, arg0)
}

// GetOrder is a free data retrieval call binding the contract method 0xd09ef241.
//
// Solidity: function getOrder(uint256 orderId) view returns((uint256,uint256,address,string,uint256,address,uint256,uint256,uint8,uint8,uint8,uint8))
func (_Order *OrderCaller) GetOrder(opts *bind.CallOpts, orderId *big.Int) (OrderContractOrder, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "getOrder", orderId)

	if err != nil {
		return *new(OrderContractOrder), err
	}

	out0 := *abi.ConvertType(out[0], new(OrderContractOrder)).(*OrderContractOrder)

	return out0, err

}

// GetOrder is a free data retrieval call binding the contract method 0xd09ef241.
//
// Solidity: function getOrder(uint256 orderId) view returns((uint256,uint256,address,string,uint256,address,uint256,uint256,uint8,uint8,uint8,uint8))
func (_Order *OrderSession) GetOrder(orderId *big.Int) (OrderContractOrder, error) {
	return _Order.Contract.GetOrder(&_Order.CallOpts, orderId)
}

// GetOrder is a free data retrieval call binding the contract method 0xd09ef241.
//
// Solidity: function getOrder(uint256 orderId) view returns((uint256,uint256,address,string,uint256,address,uint256,uint256,uint8,uint8,uint8,uint8))
func (_Order *OrderCallerSession) GetOrder(orderId *big.Int) (OrderContractOrder, error) {
	return _Order.Contract.GetOrder(&_Order.CallOpts, orderId)
}

// GetOrderNumber is a free data retrieval call binding the contract method 0x517627a2.
//
// Solidity: function getOrderNumber(uint256 orderId) view returns(uint256)
func (_Order *OrderCaller) GetOrderNumber(opts *bind.CallOpts, orderId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "getOrderNumber", orderId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOrderNumber is a free data retrieval call binding the contract method 0x517627a2.
//
// Solidity: function getOrderNumber(uint256 orderId) view returns(uint256)
func (_Order *OrderSession) GetOrderNumber(orderId *big.Int) (*big.Int, error) {
	return _Order.Contract.GetOrderNumber(&_Order.CallOpts, orderId)
}

// GetOrderNumber is a free data retrieval call binding the contract method 0x517627a2.
//
// Solidity: function getOrderNumber(uint256 orderId) view returns(uint256)
func (_Order *OrderCallerSession) GetOrderNumber(orderId *big.Int) (*big.Int, error) {
	return _Order.Contract.GetOrderNumber(&_Order.CallOpts, orderId)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Order *OrderCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Order *OrderSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Order.Contract.GetRoleAdmin(&_Order.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Order *OrderCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Order.Contract.GetRoleAdmin(&_Order.CallOpts, role)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Order *OrderCaller) GetRoleMember(opts *bind.CallOpts, role [32]byte, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "getRoleMember", role, index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Order *OrderSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _Order.Contract.GetRoleMember(&_Order.CallOpts, role, index)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Order *OrderCallerSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _Order.Contract.GetRoleMember(&_Order.CallOpts, role, index)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Order *OrderCaller) GetRoleMemberCount(opts *bind.CallOpts, role [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "getRoleMemberCount", role)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Order *OrderSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _Order.Contract.GetRoleMemberCount(&_Order.CallOpts, role)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Order *OrderCallerSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _Order.Contract.GetRoleMemberCount(&_Order.CallOpts, role)
}

// GetRoleMembers is a free data retrieval call binding the contract method 0xa3246ad3.
//
// Solidity: function getRoleMembers(bytes32 role) view returns(address[])
func (_Order *OrderCaller) GetRoleMembers(opts *bind.CallOpts, role [32]byte) ([]common.Address, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "getRoleMembers", role)

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetRoleMembers is a free data retrieval call binding the contract method 0xa3246ad3.
//
// Solidity: function getRoleMembers(bytes32 role) view returns(address[])
func (_Order *OrderSession) GetRoleMembers(role [32]byte) ([]common.Address, error) {
	return _Order.Contract.GetRoleMembers(&_Order.CallOpts, role)
}

// GetRoleMembers is a free data retrieval call binding the contract method 0xa3246ad3.
//
// Solidity: function getRoleMembers(bytes32 role) view returns(address[])
func (_Order *OrderCallerSession) GetRoleMembers(role [32]byte) ([]common.Address, error) {
	return _Order.Contract.GetRoleMembers(&_Order.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Order *OrderCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Order *OrderSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Order.Contract.HasRole(&_Order.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Order *OrderCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Order.Contract.HasRole(&_Order.CallOpts, role, account)
}

// NextOrderId is a free data retrieval call binding the contract method 0x2a58b330.
//
// Solidity: function nextOrderId() view returns(uint256)
func (_Order *OrderCaller) NextOrderId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "nextOrderId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NextOrderId is a free data retrieval call binding the contract method 0x2a58b330.
//
// Solidity: function nextOrderId() view returns(uint256)
func (_Order *OrderSession) NextOrderId() (*big.Int, error) {
	return _Order.Contract.NextOrderId(&_Order.CallOpts)
}

// NextOrderId is a free data retrieval call binding the contract method 0x2a58b330.
//
// Solidity: function nextOrderId() view returns(uint256)
func (_Order *OrderCallerSession) NextOrderId() (*big.Int, error) {
	return _Order.Contract.NextOrderId(&_Order.CallOpts)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(uint256 id, uint256 orderNumber, address user, string symbol, uint256 qty, address escrowAsset, uint256 amount, uint256 price, uint8 side, uint8 orderType, uint8 status, uint8 tif)
func (_Order *OrderCaller) Orders(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Id          *big.Int
	OrderNumber *big.Int
	User        common.Address
	Symbol      string
	Qty         *big.Int
	EscrowAsset common.Address
	Amount      *big.Int
	Price       *big.Int
	Side        uint8
	OrderType   uint8
	Status      uint8
	Tif         uint8
}, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "orders", arg0)

	outstruct := new(struct {
		Id          *big.Int
		OrderNumber *big.Int
		User        common.Address
		Symbol      string
		Qty         *big.Int
		EscrowAsset common.Address
		Amount      *big.Int
		Price       *big.Int
		Side        uint8
		OrderType   uint8
		Status      uint8
		Tif         uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Id = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.OrderNumber = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.User = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.Symbol = *abi.ConvertType(out[3], new(string)).(*string)
	outstruct.Qty = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.EscrowAsset = *abi.ConvertType(out[5], new(common.Address)).(*common.Address)
	outstruct.Amount = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)
	outstruct.Price = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.Side = *abi.ConvertType(out[8], new(uint8)).(*uint8)
	outstruct.OrderType = *abi.ConvertType(out[9], new(uint8)).(*uint8)
	outstruct.Status = *abi.ConvertType(out[10], new(uint8)).(*uint8)
	outstruct.Tif = *abi.ConvertType(out[11], new(uint8)).(*uint8)

	return *outstruct, err

}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(uint256 id, uint256 orderNumber, address user, string symbol, uint256 qty, address escrowAsset, uint256 amount, uint256 price, uint8 side, uint8 orderType, uint8 status, uint8 tif)
func (_Order *OrderSession) Orders(arg0 *big.Int) (struct {
	Id          *big.Int
	OrderNumber *big.Int
	User        common.Address
	Symbol      string
	Qty         *big.Int
	EscrowAsset common.Address
	Amount      *big.Int
	Price       *big.Int
	Side        uint8
	OrderType   uint8
	Status      uint8
	Tif         uint8
}, error) {
	return _Order.Contract.Orders(&_Order.CallOpts, arg0)
}

// Orders is a free data retrieval call binding the contract method 0xa85c38ef.
//
// Solidity: function orders(uint256 ) view returns(uint256 id, uint256 orderNumber, address user, string symbol, uint256 qty, address escrowAsset, uint256 amount, uint256 price, uint8 side, uint8 orderType, uint8 status, uint8 tif)
func (_Order *OrderCallerSession) Orders(arg0 *big.Int) (struct {
	Id          *big.Int
	OrderNumber *big.Int
	User        common.Address
	Symbol      string
	Qty         *big.Int
	EscrowAsset common.Address
	Amount      *big.Int
	Price       *big.Int
	Side        uint8
	OrderType   uint8
	Status      uint8
	Tif         uint8
}, error) {
	return _Order.Contract.Orders(&_Order.CallOpts, arg0)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Order *OrderCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Order *OrderSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Order.Contract.SupportsInterface(&_Order.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Order *OrderCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Order.Contract.SupportsInterface(&_Order.CallOpts, interfaceId)
}

// SymbolToToken is a free data retrieval call binding the contract method 0xa5bc29c2.
//
// Solidity: function symbolToToken(string ) view returns(address)
func (_Order *OrderCaller) SymbolToToken(opts *bind.CallOpts, arg0 string) (common.Address, error) {
	var out []interface{}
	err := _Order.contract.Call(opts, &out, "symbolToToken", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SymbolToToken is a free data retrieval call binding the contract method 0xa5bc29c2.
//
// Solidity: function symbolToToken(string ) view returns(address)
func (_Order *OrderSession) SymbolToToken(arg0 string) (common.Address, error) {
	return _Order.Contract.SymbolToToken(&_Order.CallOpts, arg0)
}

// SymbolToToken is a free data retrieval call binding the contract method 0xa5bc29c2.
//
// Solidity: function symbolToToken(string ) view returns(address)
func (_Order *OrderCallerSession) SymbolToToken(arg0 string) (common.Address, error) {
	return _Order.Contract.SymbolToToken(&_Order.CallOpts, arg0)
}

// CancelOrder is a paid mutator transaction binding the contract method 0x514fcac7.
//
// Solidity: function cancelOrder(uint256 orderId) returns()
func (_Order *OrderTransactor) CancelOrder(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "cancelOrder", orderId)
}

// CancelOrder is a paid mutator transaction binding the contract method 0x514fcac7.
//
// Solidity: function cancelOrder(uint256 orderId) returns()
func (_Order *OrderSession) CancelOrder(orderId *big.Int) (*types.Transaction, error) {
	return _Order.Contract.CancelOrder(&_Order.TransactOpts, orderId)
}

// CancelOrder is a paid mutator transaction binding the contract method 0x514fcac7.
//
// Solidity: function cancelOrder(uint256 orderId) returns()
func (_Order *OrderTransactorSession) CancelOrder(orderId *big.Int) (*types.Transaction, error) {
	return _Order.Contract.CancelOrder(&_Order.TransactOpts, orderId)
}

// CancelOrderIntent is a paid mutator transaction binding the contract method 0xeced998e.
//
// Solidity: function cancelOrderIntent(uint256 orderId) returns()
func (_Order *OrderTransactor) CancelOrderIntent(opts *bind.TransactOpts, orderId *big.Int) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "cancelOrderIntent", orderId)
}

// CancelOrderIntent is a paid mutator transaction binding the contract method 0xeced998e.
//
// Solidity: function cancelOrderIntent(uint256 orderId) returns()
func (_Order *OrderSession) CancelOrderIntent(orderId *big.Int) (*types.Transaction, error) {
	return _Order.Contract.CancelOrderIntent(&_Order.TransactOpts, orderId)
}

// CancelOrderIntent is a paid mutator transaction binding the contract method 0xeced998e.
//
// Solidity: function cancelOrderIntent(uint256 orderId) returns()
func (_Order *OrderTransactorSession) CancelOrderIntent(orderId *big.Int) (*types.Transaction, error) {
	return _Order.Contract.CancelOrderIntent(&_Order.TransactOpts, orderId)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Order *OrderTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Order *OrderSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Order.Contract.GrantRole(&_Order.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Order *OrderTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Order.Contract.GrantRole(&_Order.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address usdm_, address admin_, address backend_) returns()
func (_Order *OrderTransactor) Initialize(opts *bind.TransactOpts, usdm_ common.Address, admin_ common.Address, backend_ common.Address) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "initialize", usdm_, admin_, backend_)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address usdm_, address admin_, address backend_) returns()
func (_Order *OrderSession) Initialize(usdm_ common.Address, admin_ common.Address, backend_ common.Address) (*types.Transaction, error) {
	return _Order.Contract.Initialize(&_Order.TransactOpts, usdm_, admin_, backend_)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address usdm_, address admin_, address backend_) returns()
func (_Order *OrderTransactorSession) Initialize(usdm_ common.Address, admin_ common.Address, backend_ common.Address) (*types.Transaction, error) {
	return _Order.Contract.Initialize(&_Order.TransactOpts, usdm_, admin_, backend_)
}

// MarkExecuted is a paid mutator transaction binding the contract method 0x7ac5bbdd.
//
// Solidity: function markExecuted(uint256 orderId, uint256 refundAmount) returns()
func (_Order *OrderTransactor) MarkExecuted(opts *bind.TransactOpts, orderId *big.Int, refundAmount *big.Int) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "markExecuted", orderId, refundAmount)
}

// MarkExecuted is a paid mutator transaction binding the contract method 0x7ac5bbdd.
//
// Solidity: function markExecuted(uint256 orderId, uint256 refundAmount) returns()
func (_Order *OrderSession) MarkExecuted(orderId *big.Int, refundAmount *big.Int) (*types.Transaction, error) {
	return _Order.Contract.MarkExecuted(&_Order.TransactOpts, orderId, refundAmount)
}

// MarkExecuted is a paid mutator transaction binding the contract method 0x7ac5bbdd.
//
// Solidity: function markExecuted(uint256 orderId, uint256 refundAmount) returns()
func (_Order *OrderTransactorSession) MarkExecuted(orderId *big.Int, refundAmount *big.Int) (*types.Transaction, error) {
	return _Order.Contract.MarkExecuted(&_Order.TransactOpts, orderId, refundAmount)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_Order *OrderTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_Order *OrderSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _Order.Contract.RenounceRole(&_Order.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_Order *OrderTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _Order.Contract.RenounceRole(&_Order.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Order *OrderTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Order *OrderSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Order.Contract.RevokeRole(&_Order.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Order *OrderTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Order.Contract.RevokeRole(&_Order.TransactOpts, role, account)
}

// SetBackend is a paid mutator transaction binding the contract method 0xda7fc24f.
//
// Solidity: function setBackend(address backend) returns()
func (_Order *OrderTransactor) SetBackend(opts *bind.TransactOpts, backend common.Address) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "setBackend", backend)
}

// SetBackend is a paid mutator transaction binding the contract method 0xda7fc24f.
//
// Solidity: function setBackend(address backend) returns()
func (_Order *OrderSession) SetBackend(backend common.Address) (*types.Transaction, error) {
	return _Order.Contract.SetBackend(&_Order.TransactOpts, backend)
}

// SetBackend is a paid mutator transaction binding the contract method 0xda7fc24f.
//
// Solidity: function setBackend(address backend) returns()
func (_Order *OrderTransactorSession) SetBackend(backend common.Address) (*types.Transaction, error) {
	return _Order.Contract.SetBackend(&_Order.TransactOpts, backend)
}

// SetSymbolToken is a paid mutator transaction binding the contract method 0xe113d04b.
//
// Solidity: function setSymbolToken(string symbol, address token) returns()
func (_Order *OrderTransactor) SetSymbolToken(opts *bind.TransactOpts, symbol string, token common.Address) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "setSymbolToken", symbol, token)
}

// SetSymbolToken is a paid mutator transaction binding the contract method 0xe113d04b.
//
// Solidity: function setSymbolToken(string symbol, address token) returns()
func (_Order *OrderSession) SetSymbolToken(symbol string, token common.Address) (*types.Transaction, error) {
	return _Order.Contract.SetSymbolToken(&_Order.TransactOpts, symbol, token)
}

// SetSymbolToken is a paid mutator transaction binding the contract method 0xe113d04b.
//
// Solidity: function setSymbolToken(string symbol, address token) returns()
func (_Order *OrderTransactorSession) SetSymbolToken(symbol string, token common.Address) (*types.Transaction, error) {
	return _Order.Contract.SetSymbolToken(&_Order.TransactOpts, symbol, token)
}

// SubmitOrder is a paid mutator transaction binding the contract method 0x0ce16117.
//
// Solidity: function submitOrder(string symbol, uint256 qty, uint256 price, uint8 side, uint8 orderType, uint8 tif) returns(uint256 orderId)
func (_Order *OrderTransactor) SubmitOrder(opts *bind.TransactOpts, symbol string, qty *big.Int, price *big.Int, side uint8, orderType uint8, tif uint8) (*types.Transaction, error) {
	return _Order.contract.Transact(opts, "submitOrder", symbol, qty, price, side, orderType, tif)
}

// SubmitOrder is a paid mutator transaction binding the contract method 0x0ce16117.
//
// Solidity: function submitOrder(string symbol, uint256 qty, uint256 price, uint8 side, uint8 orderType, uint8 tif) returns(uint256 orderId)
func (_Order *OrderSession) SubmitOrder(symbol string, qty *big.Int, price *big.Int, side uint8, orderType uint8, tif uint8) (*types.Transaction, error) {
	return _Order.Contract.SubmitOrder(&_Order.TransactOpts, symbol, qty, price, side, orderType, tif)
}

// SubmitOrder is a paid mutator transaction binding the contract method 0x0ce16117.
//
// Solidity: function submitOrder(string symbol, uint256 qty, uint256 price, uint8 side, uint8 orderType, uint8 tif) returns(uint256 orderId)
func (_Order *OrderTransactorSession) SubmitOrder(symbol string, qty *big.Int, price *big.Int, side uint8, orderType uint8, tif uint8) (*types.Transaction, error) {
	return _Order.Contract.SubmitOrder(&_Order.TransactOpts, symbol, qty, price, side, orderType, tif)
}

// OrderCancelRequestedIterator is returned from FilterCancelRequested and is used to iterate over the raw logs and unpacked data for CancelRequested events raised by the Order contract.
type OrderCancelRequestedIterator struct {
	Event *OrderCancelRequested // Event containing the contract specifics and raw log

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
func (it *OrderCancelRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderCancelRequested)
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
		it.Event = new(OrderCancelRequested)
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
func (it *OrderCancelRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderCancelRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderCancelRequested represents a CancelRequested event raised by the Order contract.
type OrderCancelRequested struct {
	User           common.Address
	OrderId        *big.Int
	BlockTimestamp *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterCancelRequested is a free log retrieval operation binding the contract event 0xfe154d5aa2d8e2ebfb1c7adc3f245d13378a6d3254e575a29a4e7f1240448f57.
//
// Solidity: event CancelRequested(address indexed user, uint256 indexed orderId, uint256 blockTimestamp)
func (_Order *OrderFilterer) FilterCancelRequested(opts *bind.FilterOpts, user []common.Address, orderId []*big.Int) (*OrderCancelRequestedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _Order.contract.FilterLogs(opts, "CancelRequested", userRule, orderIdRule)
	if err != nil {
		return nil, err
	}
	return &OrderCancelRequestedIterator{contract: _Order.contract, event: "CancelRequested", logs: logs, sub: sub}, nil
}

// WatchCancelRequested is a free log subscription operation binding the contract event 0xfe154d5aa2d8e2ebfb1c7adc3f245d13378a6d3254e575a29a4e7f1240448f57.
//
// Solidity: event CancelRequested(address indexed user, uint256 indexed orderId, uint256 blockTimestamp)
func (_Order *OrderFilterer) WatchCancelRequested(opts *bind.WatchOpts, sink chan<- *OrderCancelRequested, user []common.Address, orderId []*big.Int) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _Order.contract.WatchLogs(opts, "CancelRequested", userRule, orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderCancelRequested)
				if err := _Order.contract.UnpackLog(event, "CancelRequested", log); err != nil {
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

// ParseCancelRequested is a log parse operation binding the contract event 0xfe154d5aa2d8e2ebfb1c7adc3f245d13378a6d3254e575a29a4e7f1240448f57.
//
// Solidity: event CancelRequested(address indexed user, uint256 indexed orderId, uint256 blockTimestamp)
func (_Order *OrderFilterer) ParseCancelRequested(log types.Log) (*OrderCancelRequested, error) {
	event := new(OrderCancelRequested)
	if err := _Order.contract.UnpackLog(event, "CancelRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrderInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Order contract.
type OrderInitializedIterator struct {
	Event *OrderInitialized // Event containing the contract specifics and raw log

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
func (it *OrderInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderInitialized)
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
		it.Event = new(OrderInitialized)
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
func (it *OrderInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderInitialized represents a Initialized event raised by the Order contract.
type OrderInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Order *OrderFilterer) FilterInitialized(opts *bind.FilterOpts) (*OrderInitializedIterator, error) {

	logs, sub, err := _Order.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &OrderInitializedIterator{contract: _Order.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Order *OrderFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *OrderInitialized) (event.Subscription, error) {

	logs, sub, err := _Order.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderInitialized)
				if err := _Order.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Order *OrderFilterer) ParseInitialized(log types.Log) (*OrderInitialized, error) {
	event := new(OrderInitialized)
	if err := _Order.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrderOrderCancelledIterator is returned from FilterOrderCancelled and is used to iterate over the raw logs and unpacked data for OrderCancelled events raised by the Order contract.
type OrderOrderCancelledIterator struct {
	Event *OrderOrderCancelled // Event containing the contract specifics and raw log

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
func (it *OrderOrderCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderOrderCancelled)
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
		it.Event = new(OrderOrderCancelled)
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
func (it *OrderOrderCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderOrderCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderOrderCancelled represents a OrderCancelled event raised by the Order contract.
type OrderOrderCancelled struct {
	OrderId        *big.Int
	User           common.Address
	Asset          common.Address
	RefundAmount   *big.Int
	Side           uint8
	OrderType      uint8
	Tif            uint8
	PreviousStatus uint8
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterOrderCancelled is a free log retrieval operation binding the contract event 0xd765f771f38f2653d67599d9f0d6679033a1c43b1f9f942ffd6b083071c8e5a7.
//
// Solidity: event OrderCancelled(uint256 indexed orderId, address indexed user, address asset, uint256 refundAmount, uint8 side, uint8 orderType, uint8 tif, uint8 previousStatus)
func (_Order *OrderFilterer) FilterOrderCancelled(opts *bind.FilterOpts, orderId []*big.Int, user []common.Address) (*OrderOrderCancelledIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Order.contract.FilterLogs(opts, "OrderCancelled", orderIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return &OrderOrderCancelledIterator{contract: _Order.contract, event: "OrderCancelled", logs: logs, sub: sub}, nil
}

// WatchOrderCancelled is a free log subscription operation binding the contract event 0xd765f771f38f2653d67599d9f0d6679033a1c43b1f9f942ffd6b083071c8e5a7.
//
// Solidity: event OrderCancelled(uint256 indexed orderId, address indexed user, address asset, uint256 refundAmount, uint8 side, uint8 orderType, uint8 tif, uint8 previousStatus)
func (_Order *OrderFilterer) WatchOrderCancelled(opts *bind.WatchOpts, sink chan<- *OrderOrderCancelled, orderId []*big.Int, user []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Order.contract.WatchLogs(opts, "OrderCancelled", orderIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderOrderCancelled)
				if err := _Order.contract.UnpackLog(event, "OrderCancelled", log); err != nil {
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

// ParseOrderCancelled is a log parse operation binding the contract event 0xd765f771f38f2653d67599d9f0d6679033a1c43b1f9f942ffd6b083071c8e5a7.
//
// Solidity: event OrderCancelled(uint256 indexed orderId, address indexed user, address asset, uint256 refundAmount, uint8 side, uint8 orderType, uint8 tif, uint8 previousStatus)
func (_Order *OrderFilterer) ParseOrderCancelled(log types.Log) (*OrderOrderCancelled, error) {
	event := new(OrderOrderCancelled)
	if err := _Order.contract.UnpackLog(event, "OrderCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrderOrderExecutedIterator is returned from FilterOrderExecuted and is used to iterate over the raw logs and unpacked data for OrderExecuted events raised by the Order contract.
type OrderOrderExecutedIterator struct {
	Event *OrderOrderExecuted // Event containing the contract specifics and raw log

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
func (it *OrderOrderExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderOrderExecuted)
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
		it.Event = new(OrderOrderExecuted)
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
func (it *OrderOrderExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderOrderExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderOrderExecuted represents a OrderExecuted event raised by the Order contract.
type OrderOrderExecuted struct {
	OrderId      *big.Int
	User         common.Address
	RefundAmount *big.Int
	Tif          uint8
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterOrderExecuted is a free log retrieval operation binding the contract event 0x60269b84d5dc1353aca157ab4cdab7ca9cabb31233d688fd030c9445798d1e67.
//
// Solidity: event OrderExecuted(uint256 indexed orderId, address indexed user, uint256 refundAmount, uint8 tif)
func (_Order *OrderFilterer) FilterOrderExecuted(opts *bind.FilterOpts, orderId []*big.Int, user []common.Address) (*OrderOrderExecutedIterator, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Order.contract.FilterLogs(opts, "OrderExecuted", orderIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return &OrderOrderExecutedIterator{contract: _Order.contract, event: "OrderExecuted", logs: logs, sub: sub}, nil
}

// WatchOrderExecuted is a free log subscription operation binding the contract event 0x60269b84d5dc1353aca157ab4cdab7ca9cabb31233d688fd030c9445798d1e67.
//
// Solidity: event OrderExecuted(uint256 indexed orderId, address indexed user, uint256 refundAmount, uint8 tif)
func (_Order *OrderFilterer) WatchOrderExecuted(opts *bind.WatchOpts, sink chan<- *OrderOrderExecuted, orderId []*big.Int, user []common.Address) (event.Subscription, error) {

	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Order.contract.WatchLogs(opts, "OrderExecuted", orderIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderOrderExecuted)
				if err := _Order.contract.UnpackLog(event, "OrderExecuted", log); err != nil {
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

// ParseOrderExecuted is a log parse operation binding the contract event 0x60269b84d5dc1353aca157ab4cdab7ca9cabb31233d688fd030c9445798d1e67.
//
// Solidity: event OrderExecuted(uint256 indexed orderId, address indexed user, uint256 refundAmount, uint8 tif)
func (_Order *OrderFilterer) ParseOrderExecuted(log types.Log) (*OrderOrderExecuted, error) {
	event := new(OrderOrderExecuted)
	if err := _Order.contract.UnpackLog(event, "OrderExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrderOrderSubmittedIterator is returned from FilterOrderSubmitted and is used to iterate over the raw logs and unpacked data for OrderSubmitted events raised by the Order contract.
type OrderOrderSubmittedIterator struct {
	Event *OrderOrderSubmitted // Event containing the contract specifics and raw log

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
func (it *OrderOrderSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderOrderSubmitted)
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
		it.Event = new(OrderOrderSubmitted)
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
func (it *OrderOrderSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderOrderSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderOrderSubmitted represents a OrderSubmitted event raised by the Order contract.
type OrderOrderSubmitted struct {
	User           common.Address
	OrderId        *big.Int
	Symbol         string
	Qty            *big.Int
	Price          *big.Int
	Side           uint8
	OrderType      uint8
	Tif            uint8
	BlockTimestamp *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterOrderSubmitted is a free log retrieval operation binding the contract event 0xb6445ed733965dc00a0134eb8e1858ae4a70da945305f4e52c16a35635f9a482.
//
// Solidity: event OrderSubmitted(address indexed user, uint256 indexed orderId, string symbol, uint256 qty, uint256 price, uint8 side, uint8 orderType, uint8 tif, uint256 blockTimestamp)
func (_Order *OrderFilterer) FilterOrderSubmitted(opts *bind.FilterOpts, user []common.Address, orderId []*big.Int) (*OrderOrderSubmittedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _Order.contract.FilterLogs(opts, "OrderSubmitted", userRule, orderIdRule)
	if err != nil {
		return nil, err
	}
	return &OrderOrderSubmittedIterator{contract: _Order.contract, event: "OrderSubmitted", logs: logs, sub: sub}, nil
}

// WatchOrderSubmitted is a free log subscription operation binding the contract event 0xb6445ed733965dc00a0134eb8e1858ae4a70da945305f4e52c16a35635f9a482.
//
// Solidity: event OrderSubmitted(address indexed user, uint256 indexed orderId, string symbol, uint256 qty, uint256 price, uint8 side, uint8 orderType, uint8 tif, uint256 blockTimestamp)
func (_Order *OrderFilterer) WatchOrderSubmitted(opts *bind.WatchOpts, sink chan<- *OrderOrderSubmitted, user []common.Address, orderId []*big.Int) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var orderIdRule []interface{}
	for _, orderIdItem := range orderId {
		orderIdRule = append(orderIdRule, orderIdItem)
	}

	logs, sub, err := _Order.contract.WatchLogs(opts, "OrderSubmitted", userRule, orderIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderOrderSubmitted)
				if err := _Order.contract.UnpackLog(event, "OrderSubmitted", log); err != nil {
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

// ParseOrderSubmitted is a log parse operation binding the contract event 0xb6445ed733965dc00a0134eb8e1858ae4a70da945305f4e52c16a35635f9a482.
//
// Solidity: event OrderSubmitted(address indexed user, uint256 indexed orderId, string symbol, uint256 qty, uint256 price, uint8 side, uint8 orderType, uint8 tif, uint256 blockTimestamp)
func (_Order *OrderFilterer) ParseOrderSubmitted(log types.Log) (*OrderOrderSubmitted, error) {
	event := new(OrderOrderSubmitted)
	if err := _Order.contract.UnpackLog(event, "OrderSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrderRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Order contract.
type OrderRoleAdminChangedIterator struct {
	Event *OrderRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *OrderRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderRoleAdminChanged)
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
		it.Event = new(OrderRoleAdminChanged)
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
func (it *OrderRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderRoleAdminChanged represents a RoleAdminChanged event raised by the Order contract.
type OrderRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Order *OrderFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*OrderRoleAdminChangedIterator, error) {

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

	logs, sub, err := _Order.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &OrderRoleAdminChangedIterator{contract: _Order.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Order *OrderFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *OrderRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _Order.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderRoleAdminChanged)
				if err := _Order.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_Order *OrderFilterer) ParseRoleAdminChanged(log types.Log) (*OrderRoleAdminChanged, error) {
	event := new(OrderRoleAdminChanged)
	if err := _Order.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrderRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Order contract.
type OrderRoleGrantedIterator struct {
	Event *OrderRoleGranted // Event containing the contract specifics and raw log

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
func (it *OrderRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderRoleGranted)
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
		it.Event = new(OrderRoleGranted)
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
func (it *OrderRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderRoleGranted represents a RoleGranted event raised by the Order contract.
type OrderRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Order *OrderFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*OrderRoleGrantedIterator, error) {

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

	logs, sub, err := _Order.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &OrderRoleGrantedIterator{contract: _Order.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Order *OrderFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *OrderRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Order.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderRoleGranted)
				if err := _Order.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_Order *OrderFilterer) ParseRoleGranted(log types.Log) (*OrderRoleGranted, error) {
	event := new(OrderRoleGranted)
	if err := _Order.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OrderRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Order contract.
type OrderRoleRevokedIterator struct {
	Event *OrderRoleRevoked // Event containing the contract specifics and raw log

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
func (it *OrderRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrderRoleRevoked)
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
		it.Event = new(OrderRoleRevoked)
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
func (it *OrderRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrderRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrderRoleRevoked represents a RoleRevoked event raised by the Order contract.
type OrderRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Order *OrderFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*OrderRoleRevokedIterator, error) {

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

	logs, sub, err := _Order.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &OrderRoleRevokedIterator{contract: _Order.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Order *OrderFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *OrderRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Order.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrderRoleRevoked)
				if err := _Order.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_Order *OrderFilterer) ParseRoleRevoked(log types.Log) (*OrderRoleRevoked, error) {
	event := new(OrderRoleRevoked)
	if err := _Order.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

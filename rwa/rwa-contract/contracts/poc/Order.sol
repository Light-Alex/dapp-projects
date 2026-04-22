// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import {AccessControlEnumerable} from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {Initializable} from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {IPocToken} from "./IPocToken.sol";

/**
 * @title OrderContract
 * @notice Handle order submission, fund escrow, execution, and cancellation; the backend mints PocToken directly. When marked as Executed, the contract burns all escrowed funds; when cancelled, all funds are refunded to the user.
 */
contract OrderContract is AccessControlEnumerable, ReentrancyGuard, Initializable {
    // ============ Roles ============
    bytes32 public constant BACKEND_ROLE = keccak256("BACKEND_ROLE");

    // ============ Types ============
    enum Side { Buy, Sell }
    enum OrderType { Market, Limit }
    enum Status { Pending, Executed, CancelRequested, Cancelled }
    enum TimeInForce { DAY, GTC, OPG, IOC, FOK, GTX, GTD, CLS }

    // ============ Storage ============
    // USDM used as the payment token for Buy orders
    IPocToken public USDM; // Use USDM as the escrow/burn token for Buy orders
    // Registered PocToken for each symbol (payment token for Sell orders)
    mapping(string => IPocToken) public symbolToToken;
    // Global auto-increment nonce for unique orderId (used as storage key)
    uint256 public nextOrderId;
    // Per-account incrementing order sequence (used for display order number)
    mapping(address => uint) public accountOrderSeq;

    struct Order {
        uint id;
        uint orderNumber; // Display-only structured order number (AAAAAABBSSSSSSSSSS)
        address user;
        string symbol; // Business-side symbol
        uint qty;
        address escrowAsset; // Actual escrowed asset (USDM or the symbol's PocToken)
        uint amount; // Escrowed amount (Buy: price*qty; Sell: qty), using 18 decimals
        uint price; // For Market: user's acceptable worst execution price; for Limit: limit price
        Side side;
        OrderType orderType;
        Status status;
        TimeInForce tif;
    }

    mapping(uint => Order) public orders; // orderId => Order

    // ============ Events ============
    // Consistent with documentation: address, orderId, amount, price, side, orderType
    event OrderSubmitted(address indexed user, uint indexed orderId, string symbol, uint qty, uint price, Side side, OrderType orderType, TimeInForce tif, uint blockTimestamp);
    event CancelRequested(address indexed user, uint indexed orderId, uint blockTimestamp);
    event OrderExecuted(uint indexed orderId, address indexed user, uint refundAmount, TimeInForce tif);
    event OrderCancelled(uint indexed orderId, address indexed user, address asset, uint refundAmount, Side side, OrderType orderType, TimeInForce tif, Status previousStatus);

    // ============ Errors ============
    error AmountZero();
    error NotOwner();
    error InvalidStatus();
    error ZeroAddress();

    error AlreadyExecuted();
    error AlreadyCancelled();
    error NotCancelRequested();
    error NotFound();

    // ============ Constructor / Initializer ============
    constructor() {
        // Upgradable/initializable pattern
        _disableInitializers();
    }

    /**
     * @notice Initialize the contract (set USDM and roles)
     * @param usdm_ USDM token address (payment asset for Buy orders)
     * @param admin_ Admin address (granted DEFAULT_ADMIN_ROLE)
     * @param backend_ Backend address (granted BACKEND_ROLE; may be zero address)
     */
    function initialize(address usdm_, address admin_, address backend_) external initializer {
        if (usdm_ == address(0)) revert ZeroAddress();
        if (admin_ == address(0)) revert ZeroAddress();
        USDM = IPocToken(usdm_);

        _grantRole(DEFAULT_ADMIN_ROLE, admin_);
        if (backend_ != address(0)) {
            _grantRole(BACKEND_ROLE, backend_);
        }
    }

    // ============ Admin ============
    /**
     * @notice Register/update the PocToken for a given symbol
     */
    function setSymbolToken(string calldata symbol, address token) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (token == address(0)) revert ZeroAddress();
        symbolToToken[symbol] = IPocToken(token);
    }

    function setBackend(address backend) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (backend == address(0)) revert ZeroAddress();
        _grantRole(BACKEND_ROLE, backend);
    }

    // ============ User Flow ============
    /**
     * @notice Submit an order and escrow funds to the contract
     * @param symbol Trading symbol (string), used for events and queries
     * @param qty Order quantity (18 decimals); for Buy, escrowed funds are price*qty; for Sell, escrowed funds are qty
     * @param side Buy/Sell
     * @param orderType Market/Limit
     * @param price Price (18 decimals); for Market, it's the acceptable worst price; for Limit, it is the limit price
     * @dev The user must pre-approve this contract to spend the corresponding asset (Buy: USDM price*qty; Sell: the symbol's PocToken qty)
     */
    function submitOrder(
        string calldata symbol,
        uint qty,
        uint price,
        Side side,
        OrderType orderType,
        TimeInForce tif
    ) external nonReentrant returns (uint orderId) {
        if (qty == 0) revert AmountZero();

        // Select escrow asset by side: Buy uses USDM; Sell uses the registered symbol token
        IPocToken token = side == Side.Buy ? USDM : symbolToToken[symbol];
        if (address(token) == address(0)) revert ZeroAddress();

        // Calculate escrowed amount: Buy is price*qty (18 decimals, rounded down); Sell is qty
        uint escrowAmount = side == Side.Buy ? Math.mulDiv(price, qty, 1e18) : qty;

        // Escrow transfer
        require(token.transferFrom(msg.sender, address(this), escrowAmount), "TRANSFER_FROM_FAIL");

        // Create order with globally unique auto-increment ID
        uint seq = ++accountOrderSeq[msg.sender];
        orderId = ++nextOrderId;
        Order storage oNew = orders[orderId];
        oNew.id = orderId;
        oNew.orderNumber = _composeOrderId(msg.sender, orderType, seq);
        oNew.user = msg.sender;
        oNew.symbol = symbol;
        oNew.qty = qty;
        oNew.amount = escrowAmount;
        oNew.price = price;
        oNew.side = side;
        oNew.orderType = orderType;
        oNew.status = Status.Pending;
        oNew.tif = tif;
        oNew.escrowAsset = address(token);

        emit OrderSubmitted(msg.sender, orderId, symbol, qty, price, side, orderType, tif, block.timestamp);
    }

    /**
     * @notice User initiates a cancellation intent (only Pending can initiate)
     */
    function cancelOrderIntent(uint orderId) external {
        Order storage o = orders[orderId];
        if (o.user != msg.sender) revert NotOwner();
        if (o.status != Status.Pending) revert InvalidStatus();
        o.status = Status.CancelRequested;
        emit CancelRequested(msg.sender, orderId, block.timestamp);
    }

    // ============ Backend Flow ============
    /**
     * @notice Backend marks the order as executed and burns all escrowed funds
     * @dev Requires this contract to be granted the BURNER_ROLE for the escrow asset (USDM or symbol token)
     */
    function markExecuted(uint orderId, uint refundAmount) external onlyRole(BACKEND_ROLE) nonReentrant {
         Order storage o = orders[orderId];
         if (o.status == Status.Executed) revert AlreadyExecuted();
         if (o.status == Status.Cancelled) revert AlreadyCancelled();
         if (o.user == address(0)) revert NotFound();
         // Set status to Executed
         o.status = Status.Executed;
         // Refund any excess amount (if present)
         if (refundAmount > 0) {
             bool ok = IPocToken(o.escrowAsset).transfer(o.user, refundAmount);
             require(ok, "REFUND_FAIL");
         }
         emit OrderExecuted(orderId, o.user, refundAmount, o.tif);
     }

    /**
     * @notice Backend finally cancels the order and refunds all escrowed funds to the user (only when in CancelRequested)
     */
    function cancelOrder(uint orderId) external onlyRole(BACKEND_ROLE) nonReentrant {
        Order storage o = orders[orderId];
        // Only allow cancelling non-executed and non-already-cancelled orders (supports Pending and CancelRequested)
        if (o.status == Status.Executed || o.status == Status.Cancelled) revert InvalidStatus();
        Status prev = o.status;
        o.status = Status.Cancelled;

        // Refund entire escrowed amount
        IPocToken token = IPocToken(o.escrowAsset);
        uint refundAmount = o.amount;
        bool ok = token.transfer(o.user, refundAmount);
        require(ok, "TRANSFER_FAIL");

        emit OrderCancelled(orderId, o.user, o.escrowAsset, refundAmount, o.side, o.orderType, o.tif, prev);
    }

    // ============ Views ============
    function getOrder(uint orderId) external view returns (Order memory) {
        return orders[orderId];
    }
    /**
     * @notice Return the human-readable combined order number: `AAAAAABBSSSSSSSSSS` (all digits)
     * @dev AAAAAA: 6-digit hash of account address; BB: order type (01=Market, 02=Limit); SSSSSSSSSS: per-account sequence, 10 digits (left-padded with zeros when displaying)
     */
    function getOrderNumber(uint orderId) external view returns (uint) {
        Order storage o = orders[orderId];
        if (o.user == address(0)) revert NotFound();
        return o.orderNumber;
    }
    // Helper: return structured order type code
    function _orderTypeCode(OrderType orderType) internal pure returns (uint) {
        // Adjustable encoding if needed: currently 01=Market, 02=Limit
        return orderType == OrderType.Market ? 1 : 2;
    }
    // Helper: compose structured order id `AAAAAABBSSSSSSSSSS`
    function _composeOrderId(address user, OrderType orderType, uint seq) internal pure returns (uint) {
        uint A = uint(keccak256(abi.encodePacked(user))) % 1_000_000; // 6-digit number
        uint BB = _orderTypeCode(orderType); // 2-digit number
        // 10-digit sequence part (left zero-padding handled on the frontend display), here we directly combine numeric parts
        return A * 1_000_000_000_000 + BB * 10_000_000_000 + seq;
    }
}
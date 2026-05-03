// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import { OrderContract } from "contracts/poc/Order.sol";
import { PocToken } from "contracts/poc/PocToken.sol";
import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract OrderTest is Test {
    OrderContract internal order;
    PocToken internal usdm;
    PocToken internal symToken;

    address internal admin = address(this);
    address internal backend = address(0xBEEF);
    address internal user = address(0xA11CE);
    address internal other = address(0xB0B);
    string internal symbol = "ABC";

    // Event declarations (keep the same signatures as in OrderContract)
    event OrderSubmitted(
        address indexed user,
        uint256 indexed orderId,
        string symbol,
        uint256 qty,
        uint256 price,
        OrderContract.Side side,
        OrderContract.OrderType orderType,
        OrderContract.TimeInForce tif,
        uint256 blockTimestamp
    );
    event CancelRequested(address indexed user, uint256 indexed orderId, uint256 blockTimestamp);
    event OrderExecuted(
        uint256 indexed orderId, address indexed user, uint256 refundAmount, OrderContract.TimeInForce tif
    );
    event OrderCancelled(
        uint256 indexed orderId,
        address indexed user,
        address asset,
        uint256 refundAmount,
        OrderContract.Side side,
        OrderContract.OrderType orderType,
        OrderContract.TimeInForce tif,
        OrderContract.Status previousStatus
    );

    function setUp() public {
        // Deploy implementation contracts
        PocToken usdmImpl = new PocToken(address(0));
        PocToken symImpl = new PocToken(address(0));
        OrderContract orderImpl = new OrderContract();

        // Initialize USDM and symbol token via proxies
        ERC1967Proxy usdmProxy =
            new ERC1967Proxy(address(usdmImpl), abi.encodeWithSelector(PocToken.initialize.selector, "USDM", "USDM"));
        usdm = PocToken(address(usdmProxy));

        ERC1967Proxy symProxy =
            new ERC1967Proxy(address(symImpl), abi.encodeWithSelector(PocToken.initialize.selector, "ABC", "ABC"));
        symToken = PocToken(address(symProxy));

        // Initialize order contract via proxy
        ERC1967Proxy orderProxy = new ERC1967Proxy(
            address(orderImpl), abi.encodeWithSelector(OrderContract.initialize.selector, address(usdm), admin, backend)
        );
        order = OrderContract(address(orderProxy));

        // Bind trading symbol to its token
        order.setSymbolToken(symbol, address(symToken));

        // Pre-mint user assets (the current test contract holds MINTER_ROLE)
        usdm.mint(user, 1_000_000e18);
        symToken.mint(user, 1_000_000e18);
    }

    function testSubmitOrder_Buy_StoresStateAndBalances() public {
        uint256 qty = 100e18;
        uint256 price = 2e18;
        OrderContract.TimeInForce tif = OrderContract.TimeInForce.GTC;
        uint256 escrowAmount = (price * qty) / 1e18;

        vm.startPrank(user);
        usdm.approve(address(order), escrowAmount);
        uint256 beforeUser = usdm.balanceOf(user);
        uint256 beforeContract = usdm.balanceOf(address(order));
        uint256 orderId =
            order.submitOrder(symbol, qty, price, OrderContract.Side.Buy, OrderContract.OrderType.Market, tif);
        vm.stopPrank();

        // Check balance change
        assertEq(usdm.balanceOf(user), beforeUser - escrowAmount, "USDM user balance should decrease by escrowAmount");
        assertEq(
            usdm.balanceOf(address(order)),
            beforeContract + escrowAmount,
            "USDM contract balance should increase by escrowAmount"
        );

        // Check stored order
        OrderContract.Order memory o = order.getOrder(orderId);
        assertEq(o.user, user, "order.user");
        assertEq(o.symbol, symbol, "order.symbol");
        assertEq(o.amount, escrowAmount, "order.amount");
        assertEq(o.price, price, "order.price");
        assertEq(uint256(o.side), uint256(OrderContract.Side.Buy), "order.side");
        assertEq(uint256(o.orderType), uint256(OrderContract.OrderType.Market), "order.orderType");
        assertEq(uint256(o.status), uint256(OrderContract.Status.Pending), "order.status");
        assertEq(uint256(o.tif), uint256(tif), "order.tif");
        assertEq(o.escrowAsset, address(usdm), "order.escrowAsset should be USDM for Buy");

        // Check per-account sequence mapping
        assertEq(order.accountOrderSeq(user), 1, "seq should be 1 for first order");

        // Check orderNumber is a valid non-zero value (format: AAAAAABBSSSSSSSSSS)
        uint256 orderNumber = order.getOrderNumber(orderId);
        assertGt(orderNumber, 0, "orderNumber should be greater than 0");
    }

    function testSubmitOrder_Sell_StoresStateAndBalances() public {
        uint256 qty = 50e18;
        uint256 price = 3e18;
        OrderContract.TimeInForce tif = OrderContract.TimeInForce.DAY;
        uint256 escrowAmount = qty; // Sell escrow is qty

        vm.startPrank(user);
        symToken.approve(address(order), escrowAmount);
        uint256 beforeUser = symToken.balanceOf(user);
        uint256 beforeContract = symToken.balanceOf(address(order));
        uint256 orderId =
            order.submitOrder(symbol, qty, price, OrderContract.Side.Sell, OrderContract.OrderType.Limit, tif);
        vm.stopPrank();

        assertEq(
            symToken.balanceOf(user), beforeUser - escrowAmount, "symbol user balance should decrease by escrowAmount"
        );
        assertEq(
            symToken.balanceOf(address(order)),
            beforeContract + escrowAmount,
            "symbol contract balance should increase by escrowAmount"
        );

        OrderContract.Order memory o = order.getOrder(orderId);
        assertEq(o.user, user, "order.user");
        assertEq(o.symbol, symbol, "order.symbol");
        assertEq(o.amount, escrowAmount, "order.amount");
        assertEq(o.price, price, "order.price");
        assertEq(uint256(o.side), uint256(OrderContract.Side.Sell), "order.side");
        assertEq(uint256(o.orderType), uint256(OrderContract.OrderType.Limit), "order.orderType");
        assertEq(uint256(o.status), uint256(OrderContract.Status.Pending), "order.status");
        assertEq(uint256(o.tif), uint256(tif), "order.tif");
        assertEq(o.escrowAsset, address(symToken), "order.escrowAsset should be symbol token for Sell");

        // Check per-account sequence mapping increment
        assertEq(order.accountOrderSeq(user), 1, "seq should be 1 for single order in this test");
    }

    function testCancelOrderIntent_EmitsEventAndUpdatesStatus() public {
        // Submit an order first
        uint256 qty = 10e18;
        uint256 price = 1e18;
        OrderContract.TimeInForce tif = OrderContract.TimeInForce.GTC;
        vm.startPrank(user);
        usdm.approve(address(order), (price * qty) / 1e18);
        uint256 orderId =
            order.submitOrder(symbol, qty, price, OrderContract.Side.Buy, OrderContract.OrderType.Market, tif);
        vm.stopPrank();

        // Expect event (only check topics to avoid matching complex data)
        vm.expectEmit(true, true, false, false, address(order));
        emit CancelRequested(user, orderId, block.timestamp);

        vm.prank(user);
        order.cancelOrderIntent(orderId);

        OrderContract.Order memory o = order.getOrder(orderId);
        assertEq(uint256(o.status), uint256(OrderContract.Status.CancelRequested), "status should be CancelRequested");
    }

    function testMarkExecuted_ByBackend_RefundsAndEmits() public {
        // Submit and cancel intent
        uint256 qty = 10e18;
        uint256 price = 4e18;
        OrderContract.TimeInForce tif = OrderContract.TimeInForce.IOC;
        vm.startPrank(user);
        usdm.approve(address(order), (price * qty) / 1e18);
        uint256 orderId =
            order.submitOrder(symbol, qty, price, OrderContract.Side.Buy, OrderContract.OrderType.Limit, tif);
        vm.stopPrank();
        vm.prank(user);
        order.cancelOrderIntent(orderId);

        uint256 refundAmount = 1e18;
        uint256 beforeUser = usdm.balanceOf(user);

        // Expect event (check indexed orderId and user)
        vm.expectEmit(true, true, true, true, address(order));
        emit OrderExecuted(orderId, user, refundAmount, OrderContract.TimeInForce.IOC);

        vm.prank(backend);
        order.markExecuted(orderId, refundAmount);

        OrderContract.Order memory o = order.getOrder(orderId);
        assertEq(uint256(o.status), uint256(OrderContract.Status.Executed), "status should be Executed");
        assertEq(usdm.balanceOf(user), beforeUser + refundAmount, "user should receive refundAmount");
    }

    function testCancelOrder_ByBackend_RefundsAllAndEmits() public {
        // Submit order (allow direct cancel from Pending without cancel intent)
        uint256 qty = 5e18;
        uint256 price = 2e18;
        OrderContract.TimeInForce tif = OrderContract.TimeInForce.FOK;
        vm.startPrank(user);
        usdm.approve(address(order), (price * qty) / 1e18);
        uint256 orderId =
            order.submitOrder(symbol, qty, price, OrderContract.Side.Buy, OrderContract.OrderType.Limit, tif);
        vm.stopPrank();

        OrderContract.Order memory before = order.getOrder(orderId);
        uint256 beforeUser = usdm.balanceOf(user);

        // Expect event (check indexed orderId and user)
        vm.expectEmit(true, true, true, false, address(order));
        emit OrderCancelled(
            orderId,
            user,
            before.escrowAsset,
            before.amount,
            OrderContract.Side.Buy,
            OrderContract.OrderType.Limit,
            tif,
            OrderContract.Status.Pending
        );

        vm.prank(backend);
        order.cancelOrder(orderId);

        OrderContract.Order memory o = order.getOrder(orderId);
        assertEq(uint256(o.status), uint256(OrderContract.Status.Cancelled), "status should be Cancelled");
        assertEq(usdm.balanceOf(user), beforeUser + before.amount, "user should receive full refund");
    }

    function testPerAccountSequence_IncrementsPerUser() public {
        // Two orders for user
        vm.startPrank(user);
        usdm.approve(address(order), type(uint256).max);
        uint256 id1 = order.submitOrder(
            symbol, 1e18, 1e18, OrderContract.Side.Buy, OrderContract.OrderType.Market, OrderContract.TimeInForce.GTC
        );
        uint256 id2 = order.submitOrder(
            symbol, 1e18, 1e18, OrderContract.Side.Buy, OrderContract.OrderType.Market, OrderContract.TimeInForce.GTC
        );
        vm.stopPrank();
        assertEq(order.accountOrderSeq(user), 2, "seq should be 2 for user after two orders");

        // other starts from 1
        // Pre-mint to other must be called by admin holding MINTER_ROLE
        usdm.mint(other, 10e18);
        vm.startPrank(other);
        usdm.approve(address(order), type(uint256).max);
        uint256 id3 = order.submitOrder(
            symbol, 1e18, 1e18, OrderContract.Side.Buy, OrderContract.OrderType.Market, OrderContract.TimeInForce.GTC
        );
        vm.stopPrank();
        assertEq(order.accountOrderSeq(other), 1, "seq for first order of other user should be 1");
    }

    function testSubmitOrder_Revert_AmountZero() public {
        vm.prank(user);
        vm.expectRevert(OrderContract.AmountZero.selector);
        order.submitOrder(
            symbol, 0, 1e18, OrderContract.Side.Buy, OrderContract.OrderType.Market, OrderContract.TimeInForce.GTC
        );
    }

    function testSubmitOrder_Sell_Revert_UnregisteredSymbol() public {
        // Unregistered symbol
        vm.startPrank(user);
        symToken.approve(address(order), 1e18);
        vm.expectRevert(OrderContract.ZeroAddress.selector);
        order.submitOrder(
            "XYZ", 1e18, 1e18, OrderContract.Side.Sell, OrderContract.OrderType.Limit, OrderContract.TimeInForce.GTC
        );
        vm.stopPrank();
    }

    function testCancelOrderIntent_Revert_NotOwner() public {
        // First submit by user
        vm.startPrank(user);
        usdm.approve(address(order), 1e18);
        uint256 orderId = order.submitOrder(
            symbol, 1e18, 1e18, OrderContract.Side.Buy, OrderContract.OrderType.Market, OrderContract.TimeInForce.GTC
        );
        vm.stopPrank();

        // Other tries to cancel intent
        vm.prank(other);
        vm.expectRevert(OrderContract.NotOwner.selector);
        order.cancelOrderIntent(orderId);
    }

    function testMarkExecuted_Reverts_OnInvalidStatusOrRole() public {
        // Submit order without moving to CancelRequested
        vm.startPrank(user);
        usdm.approve(address(order), 1e18);
        uint256 orderId = order.submitOrder(
            symbol, 1e18, 1e18, OrderContract.Side.Buy, OrderContract.OrderType.Market, OrderContract.TimeInForce.GTC
        );
        vm.stopPrank();

        // Non-backend caller
        vm.prank(user);
        vm.expectRevert();
        order.markExecuted(orderId, 0);

        // After entering CancelRequested, executed by backend
        vm.prank(user);
        order.cancelOrderIntent(orderId);

        vm.prank(backend);
        order.markExecuted(orderId, 0);

        // Re-execute -> AlreadyExecuted
        vm.prank(backend);
        vm.expectRevert(OrderContract.AlreadyExecuted.selector);
        order.markExecuted(orderId, 0);
    }
}

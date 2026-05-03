// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import { PocToken } from "contracts/poc/PocToken.sol";
import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract PocTokenTest is Test {
    PocToken internal token;

    address internal admin = address(this);
    address internal minter = address(0xBEEF);
    address internal burner = address(0xCAFE);
    address internal user = address(0xA11CE);
    address internal other = address(0xB0B);

    function setUp() public {
        // Deploy implementation
        PocToken impl = new PocToken(address(0));

        // Deploy proxy and initialize
        ERC1967Proxy proxy =
            new ERC1967Proxy(address(impl), abi.encodeWithSelector(PocToken.initialize.selector, "PocToken", "POC"));
        token = PocToken(address(proxy));
    }

    // ============ Initialization ============

    function test_initialize_setsNameAndSymbol() public view {
        assertEq(token.name(), "PocToken", "name");
        assertEq(token.symbol(), "POC", "symbol");
    }

    function test_initialize_grantsRolesToCaller() public view {
        bytes32 adminRole = token.DEFAULT_ADMIN_ROLE();
        assertTrue(token.hasRole(adminRole, admin), "admin should have DEFAULT_ADMIN_ROLE");
        assertTrue(token.hasRole(token.MINTER_ROLE(), admin), "admin should have MINTER_ROLE");
        assertTrue(token.hasRole(token.BURNER_ROLE(), admin), "admin should have BURNER_ROLE");
    }

    function test_initialize_revert_doubleInit() public {
        vm.expectRevert();
        token.initialize("Again", "AGN");
    }

    // ============ mint ============

    function test_mint_success() public {
        uint256 amount = 100e18;
        token.mint(user, amount);

        assertEq(token.balanceOf(user), amount, "user balance after mint");
        assertEq(token.totalSupply(), amount, "totalSupply after mint");
    }

    function test_mint_revert_toZeroAddress() public {
        vm.expectRevert(PocToken.AddressCannotBeZero.selector);
        token.mint(address(0), 1e18);
    }

    function test_mint_revert_amountZero() public {
        vm.expectRevert(PocToken.AmountCannotBeZero.selector);
        token.mint(user, 0);
    }

    function test_mint_revert_notMinter() public {
        vm.prank(other);
        vm.expectRevert();
        token.mint(user, 1e18);
    }

    // ============ burn ============

    function test_burn_success() public {
        uint256 mintAmount = 100e18;
        uint256 burnAmount = 40e18;
        token.mint(admin, mintAmount);

        token.burn(burnAmount);

        assertEq(token.balanceOf(admin), mintAmount - burnAmount, "admin balance after burn");
        assertEq(token.totalSupply(), mintAmount - burnAmount, "totalSupply after burn");
    }

    function test_burn_revert_amountZero() public {
        vm.expectRevert(PocToken.AmountCannotBeZero.selector);
        token.burn(0);
    }

    function test_burn_revert_notBurner() public {
        vm.prank(other);
        vm.expectRevert();
        token.burn(1e18);
    }

    function test_burn_revert_insufficientBalance() public {
        // admin has BURNER_ROLE but zero balance
        vm.expectRevert();
        token.burn(1e18);
    }

    // ============ burnFrom ============

    function test_burnFrom_success() public {
        uint256 mintAmount = 100e18;
        uint256 burnAmount = 30e18;
        token.mint(user, mintAmount);

        token.burnFrom(user, burnAmount);

        assertEq(token.balanceOf(user), mintAmount - burnAmount, "user balance after burnFrom");
        assertEq(token.totalSupply(), mintAmount - burnAmount, "totalSupply after burnFrom");
    }

    function test_burnFrom_revert_fromZeroAddress() public {
        vm.expectRevert(PocToken.AddressCannotBeZero.selector);
        token.burnFrom(address(0), 1e18);
    }

    function test_burnFrom_revert_amountZero() public {
        vm.expectRevert(PocToken.AmountCannotBeZero.selector);
        token.burnFrom(user, 0);
    }

    function test_burnFrom_revert_notBurner() public {
        token.mint(user, 100e18);

        vm.prank(other);
        vm.expectRevert();
        token.burnFrom(user, 1e18);
    }

    function test_burnFrom_revert_insufficientBalance() public {
        // user has 0 tokens
        vm.expectRevert();
        token.burnFrom(user, 1e18);
    }

    function test_burnFrom_noAllowanceRequired() public {
        // Key test: burnFrom does NOT check allowance, only BURNER_ROLE
        uint256 amount = 50e18;
        token.mint(user, amount);

        // user never approved admin, but admin has BURNER_ROLE
        assertEq(token.allowance(user, admin), 0, "allowance should be zero");

        // burnFrom should succeed without any approval
        token.burnFrom(user, amount);
        assertEq(token.balanceOf(user), 0, "user balance should be zero after burnFrom");
    }

    // ============ setOperator ============

    function test_setOperator_success() public {
        address operator = address(0xDEAD);
        token.setOperator(operator);

        assertTrue(token.hasRole(token.MINTER_ROLE(), operator), "operator should have MINTER_ROLE");
        assertTrue(token.hasRole(token.BURNER_ROLE(), operator), "operator should have BURNER_ROLE");
    }

    function test_setOperator_revert_zeroAddress() public {
        vm.expectRevert(PocToken.AddressCannotBeZero.selector);
        token.setOperator(address(0));
    }

    function test_setOperator_revert_notAdmin() public {
        vm.prank(other);
        vm.expectRevert();
        token.setOperator(address(0xDEAD));
    }

    // ============ ERC20 Standard ============

    function test_transfer_success() public {
        uint256 amount = 50e18;
        token.mint(admin, amount);

        token.transfer(user, amount);

        assertEq(token.balanceOf(admin), 0, "admin balance after transfer");
        assertEq(token.balanceOf(user), amount, "user balance after transfer");
    }

    function test_transferFrom_success() public {
        uint256 amount = 50e18;
        token.mint(user, amount);

        vm.prank(user);
        token.approve(admin, amount);

        token.transferFrom(user, other, amount);

        assertEq(token.balanceOf(user), 0, "user balance after transferFrom");
        assertEq(token.balanceOf(other), amount, "other balance after transferFrom");
    }

    function test_approve_success() public {
        uint256 amount = 100e18;

        vm.prank(user);
        token.approve(admin, amount);

        assertEq(token.allowance(user, admin), amount, "allowance should match approved amount");
    }

    function test_name_returnsCorrectValue() public view {
        assertEq(token.name(), "PocToken");
    }

    function test_symbol_returnsCorrectValue() public view {
        assertEq(token.symbol(), "POC");
    }

    // ============ View Functions ============

    function test_isMinter_returnsTrue() public view {
        assertTrue(token.isMinter(admin), "admin should be minter");
    }

    function test_isMinter_returnsFalse() public view {
        assertFalse(token.isMinter(other), "other should not be minter");
    }

    function test_isBurner_returnsTrue() public view {
        assertTrue(token.isBurner(admin), "admin should be burner");
    }

    function test_isBurner_returnsFalse() public view {
        assertFalse(token.isBurner(other), "other should not be burner");
    }

    function test_isMinter_afterSetOperator() public {
        token.setOperator(other);
        assertTrue(token.isMinter(other), "operator should be minter after setOperator");
        assertTrue(token.isBurner(other), "operator should be burner after setOperator");
    }
}

// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Test} from "forge-std/Test.sol";
import {MyToken} from "../src/MyERC20.sol";
import "forge-std/console.sol";

contract MyTokenTest is Test {
    MyToken public token;
    uint256 public initialSupply = 10000;

    address public user1 = makeAddr("user1");
    address public user2 = makeAddr("user2");

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    function setUp() public {
        token = new MyToken(initialSupply);
        token.transfer(user1, 500 * 10 ** uint256(token.decimals()));
    }

    function test_InitialSupply() public {
        assertEq(token.totalSupply(), initialSupply * 10 ** uint256(token.decimals()));
    }

    function test_NameAndSymbol() public {
        assertEq(token.name(), "MyToken");
        assertEq(token.symbol(), "MTK");
    }

    function test_Transfer() public {
        uint256 transferAmount = 200 * 10 ** uint256(token.decimals());
        assertEq(token.balanceOf(user1), 500 * 10 ** uint256(token.decimals()));
        assertEq(token.balanceOf(user2), 0);

        vm.prank(user1);
        // 事件检查
        vm.expectEmit(true, true, true, true);
        emit Transfer(user1, user2, transferAmount);
        token.transfer(user2, transferAmount);
        assertEq(token.balanceOf(user1), 300 * 10 ** uint256(token.decimals()));
        assertEq(token.balanceOf(user2), transferAmount);
    }

    function test_ApproveAndTransferFrom() public {
        uint256 approveAmount = 100 * 10 ** uint256(token.decimals());
        uint256 transferAmount = 50 * 10 ** uint256(token.decimals());
        assertEq(token.allowance(user1, user2), 0);

        vm.prank(user1);
        // 事件检查
        vm.expectEmit(true, true, true, true);
        emit Approval(user1, user2, approveAmount);
        token.approve(user2, approveAmount);
        assertEq(token.allowance(user1, user2), approveAmount);

        vm.prank(user2);
        // 事件检查
        vm.expectEmit(true, true, true, true);
        emit Transfer(user1, user2, transferAmount);
        token.transferFrom(user1, user2, transferAmount);
        assertEq(token.allowance(user1, user2), approveAmount - transferAmount);
        assertEq(token.balanceOf(user1), 450 * 10 ** uint256(token.decimals()));
        assertEq(token.balanceOf(user2), transferAmount);
    }
}
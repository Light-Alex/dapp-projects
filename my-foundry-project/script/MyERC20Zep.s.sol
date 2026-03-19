// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import "./BaseScript.s.sol";
import "../src/MyERC20Zep.sol";

contract MyERC20ZepScript is BaseScript {

    // 部署 MyERC20 合约
    function run() public broadcaster {
        console.log("Deployer address: %s", deployer);
        MyERC20 token = new MyERC20("Light Token", "LTK", 10000);
        saveContract(getNetworkName(block.chainid), "MyERC20", address(token));
        console.log("MyERC20 deployed on %s", address(token));
    }
}
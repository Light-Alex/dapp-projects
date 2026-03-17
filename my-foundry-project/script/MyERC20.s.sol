// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import "./BaseScript.s.sol";
import "../src/MyERC20.sol";

contract MyERC20Script is BaseScript {

    // 部署 MyToken 合约
    function run() public broadcaster {
        console.log("Deployer address: %s", deployer);
        MyToken token = new MyToken(10000);
        saveContract(getNetworkName(block.chainid), "MyToken", address(token));
        console.log("MyToken deployed on %s", address(token));
    }
}
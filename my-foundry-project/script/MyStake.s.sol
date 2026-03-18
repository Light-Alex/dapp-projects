// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import "./BaseScript.s.sol";
import "../src/MyStake.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract MyStakeScript is BaseScript {
    function run() public broadcaster {
        console.log("Deployer address: %s", deployer);
        MtkContracts myStake = new MtkContracts(IERC20(0x93480Ce4b54baD6c60D8CDAEaeaF898fE00deBF2));
        saveContract(getNetworkName(block.chainid), "MtkContracts", address(myStake));
        console.log("MtkContracts deployed on %s", address(myStake));
    }

}
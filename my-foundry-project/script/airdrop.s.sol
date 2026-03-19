// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import "./BaseScript.s.sol";
import "../src/airdrop.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract AirdropScript is BaseScript {
    function run() public broadcaster {
        console.log("Deployer address: %s", deployer);
        
        Airdrop airdrop = new Airdrop(IERC20(0x3d720e33044bD6691A850cEcF03E5D1C9e25Eda3));
        saveContract(getNetworkName(block.chainid), "Airdrop", address(airdrop));
        console.log("Airdrop deployed on %s", address(airdrop));

    }
}
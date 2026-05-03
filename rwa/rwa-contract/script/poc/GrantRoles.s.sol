// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import "./BaseScript.s.sol";
import { PocToken } from "../../contracts/poc/PocToken.sol";
import { OrderContract } from "../../contracts/poc/Order.sol";

/**
 * @title GrantRoles
 * @notice Grants MINTER_ROLE to backend address on all PocToken contracts
 * @dev Run after deploying contracts to authorize backend to mint tokens
 */
contract GrantRoles is BaseScript {
    // ============ Configuration ============
    // Set these before running
    address public backendAddress;
    address public usdmAddress;
    address public aaplAddress;
    address public tslaAddress;

    function run() public broadcaster {
        // Read from environment or use deployer
        backendAddress = vm.envOr("BACKEND_ADDRESS", deployer);
        usdmAddress = vm.envAddress("USDM_ADDRESS");
        aaplAddress = vm.envAddress("AAPL_ADDRESS");
        tslaAddress = vm.envAddress("TSLA_ADDRESS");

        console.log("Backend Address:", backendAddress);
        console.log("USDM:", usdmAddress);
        console.log("AAPL:", aaplAddress);
        console.log("TSLA:", tslaAddress);

        bytes32 MINTER_ROLE = keccak256("MINTER_ROLE");

        // Grant MINTER_ROLE to backend on USDM
        PocToken usdm = PocToken(usdmAddress);
        usdm.grantRole(MINTER_ROLE, backendAddress);
        console.log("Granted MINTER_ROLE on USDM to backend");

        // Grant MINTER_ROLE to backend on AAPL
        PocToken aapl = PocToken(aaplAddress);
        aapl.grantRole(MINTER_ROLE, backendAddress);
        console.log("Granted MINTER_ROLE on AAPL to backend");

        // Grant MINTER_ROLE to backend on TSLA
        PocToken tsla = PocToken(tslaAddress);
        tsla.grantRole(MINTER_ROLE, backendAddress);
        console.log("Granted MINTER_ROLE on TSLA to backend");

        console.log("\n========== Grant Complete ==========");
        console.log("Backend can now mint tokens on:");
        console.log("  - USDM");
        console.log("  - AAPL.anc");
        console.log("  - TSLA.anc");
        console.log("====================================");
    }
}

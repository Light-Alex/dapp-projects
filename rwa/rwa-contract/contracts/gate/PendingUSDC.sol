// SPDX-License-Identifier: UNLICENSED
/**
 * Copyright (c) 2023, Anchored Finance
 */
pragma solidity ^0.8.20;

import { PendingToken } from "contracts/gate/PendingToken.sol";

/**
 * @title  PendingUSDC
 * @notice A non-transferable token representing pending USDC during withdrawal processing
 */
contract PendingUSDC is PendingToken {
    /**
     * @notice Constructor
     * @param gateContract_ The address of the Gate contract
     */
    constructor(address gateContract_) PendingToken("Pending USDC", "pendingUSDC", gateContract_) {
        // All initialization is handled by the parent contract
    }
}

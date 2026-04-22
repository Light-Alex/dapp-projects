// SPDX-License-Identifier: UNLICENSED
/**
 * Copyright (c) 2023, Anchored Finance
 */
pragma solidity ^0.8.20;

import { PendingToken } from "contracts/gate/PendingToken.sol";

/**
 * @title  PendingAncUSDC
 * @notice A non-transferable token representing pending ancUSDC during deposit processing
 */
contract PendingAncUSDC is PendingToken {
    /**
     * @notice Constructor
     * @param gateContract_ The address of the Gate contract
     */
    constructor(address gateContract_) PendingToken("Pending Anchored USDC", "pendingAncUSDC", gateContract_) {
        // All initialization is handled by the parent contract
    }
}

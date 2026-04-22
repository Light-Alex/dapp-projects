// SPDX-License-Identifier: BUSL-1.1
/*
 */
pragma solidity ^0.8.20;

import { IAnchoredRegistrar } from "contracts/interfaces/IAnchoredRegistrar.sol";
import { AnchoredTokenManager } from "contracts/AnchoredTokenManager.sol";

/**
 * @title  IAnchoredTokenManagerRegistrar
 * @notice Interface for the AnchoredTokenManagerRegistrar contract containing all functions, events, and errors
 */
interface IAnchoredTokenManagerRegistrar is IAnchoredRegistrar {
    // ============ Functions ============

    /**
     * @notice Pauses the registrar, disabling registration
     */
    function pause() external;

    /**
     * @notice Unpauses the registrar, enabling registration
     */
    function unpause() external;

    /**
     * @notice Sets or updates the Anchored Token Manager address
     * @param  anchoredTokenManager_ The new Anchored Token Manager address
     */
    function setAnchoredTokenManager(address anchoredTokenManager_) external;

    /**
     * @notice Returns the current Anchored Token Manager address
     * @return The AnchoredTokenManager contract instance
     */
    function anchoredTokenManager() external view returns (AnchoredTokenManager);

    /**
     * @notice Returns the CONFIGURE_ROLE constant
     * @return The bytes32 value of CONFIGURE_ROLE
     */
    function CONFIGURE_ROLE() external view returns (bytes32);

    /**
     * @notice Returns the TOKEN_FACTORY_ROLE constant
     * @return The bytes32 value of TOKEN_FACTORY_ROLE
     */
    function TOKEN_REGISTER_ROLE() external view returns (bytes32);

    /**
     * @notice Returns the PAUSER_ROLE constant
     * @return The bytes32 value of PAUSER_ROLE
     */
    function PAUSE_ROLE() external view returns (bytes32);

    /**
     * @notice Returns the UNPAUSER_ROLE constant
     * @return The bytes32 value of UNPAUSER_ROLE
     */
    function UNPAUSE_ROLE() external view returns (bytes32);

    // ============ Events ============

    /**
     * @notice Emitted when the `AnchoredTokenManager` is set
     * @param  oldManager The old `AnchoredTokenManager` address
     * @param  newManager The new `AnchoredTokenManager` address
     */
    event AnchoredTokenManagerSet(address indexed oldManager, address indexed newManager);

    /**
     * @notice Emitted when a new token is registered
     * @param token The address of the token that was registered following a deployment
     */
    event TokenRegistered(address indexed token);

    // ============ Errors ============

    /// Error thrown when attempting to set the Anchored Token Manager to zero address
    error AnchoredTokenManagerCannotBeZero();

    /// Error thrown when attempting to register a token with zero address
    error TokenAddressCannotBeZero();
}

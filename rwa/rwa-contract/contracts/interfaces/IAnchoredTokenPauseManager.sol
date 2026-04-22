// SPDX-License-Identifier: BUSL-1.1
/**
 * @title  IAnchoredTokenPauseManager
 * @notice Comprehensive interface for the AnchoredTokenPauseManager contract containing all functions, events, and errors
 */
pragma solidity ^0.8.20;

import { IAccessControl } from "@openzeppelin/contracts/access/IAccessControl.sol";

interface IAnchoredTokenPauseManager is IAccessControl {
    // ============ Functions ============

    /**
     * @notice Pauses a token
     * @param  token The address of the token to pause
     * @dev    Only callable by addresses with the `PAUSE_TOKEN_ROLE`
     */
    function pauseToken(address token) external;

    /**
     * @notice Unpauses a token
     * @param  token The address of the token to unpause
     * @dev    Only callable by addresses with the `UNPAUSE_TOKEN_ROLE`
     */
    function unpauseToken(address token) external;

    /**
     * @notice Pauses all tokens
     * @dev    Only callable by addresses with the `PAUSE_TOKEN_ROLE`
     */
    function pauseAllTokens() external;

    /**
     * @notice Unpauses all tokens
     * @dev    Only callable by addresses with the `UNPAUSE_TOKEN_ROLE`.
     */
    function unpauseAllTokens() external;

    // ============ View Functions ============

    // Note: hasRole, getRoleAdmin, getRoleMember, getRoleMemberCount, and DEFAULT_ADMIN_ROLE are inherited from IAccessControl

    /**
     * @notice Returns the PAUSE_TOKEN_ROLE constant
     * @return The bytes32 value of PAUSE_TOKEN_ROLE
     */
    function PAUSE_TOKEN_ROLE() external view returns (bytes32);

    /**
     * @notice Returns the UNPAUSE_TOKEN_ROLE constant
     * @return The bytes32 value of UNPAUSE_TOKEN_ROLE
     */
    function UNPAUSE_TOKEN_ROLE() external view returns (bytes32);

    /**
     * @notice Returns whether a specific token is paused
     * @param token The address of the token to check
     * @return True if the token is paused
     */
    function pausedTokens(address token) external view returns (bool);

    /**
     * @notice Returns whether a specific token is paused (alias for pausedTokens)
     * @param token The address of the token to check
     * @return True if the token is paused
     */
    function isTokenPaused(address token) external view returns (bool);

    /**
     * @notice Returns the global paused state for all tokens
     * @return True if all tokens are paused
     */
    function allTokensPaused() external view returns (bool);

    // ============ Events ============

    /**
     * @notice Emitted when a token is paused
     * @param  token  The address of the paused token
     * @param  pauser The address that initiated the pause
     */
    event TokenPaused(address indexed token, address indexed pauser);

    /**
     * @notice Emitted when a token is unpaused
     * @param  token  The address of the unpaused token
     * @param  pauser The address that initiated the unpause
     */
    event TokenUnpaused(address indexed token, address indexed pauser);

    /**
     * @notice Emitted when all token tokens are paused
     * @param  pauser The address that initiated the pause
     */
    event AllTokensPaused(address indexed pauser);

    /**
     * @notice Emitted when all tokens are unpaused
     * @param  pauser The address that initiated the unpause
     */
    event AllTokensUnpaused(address indexed pauser);

    /**
     * @notice Event emitted when the token pause manager is set
     * @param oldTokenPauseManager The old token pause manager address
     * @param newTokenPauseManager The new token pause manager address
     */
    event TokenPauseManagerSet(address indexed oldTokenPauseManager, address indexed newTokenPauseManager);

    // ============ Errors ============

    /// Error thrown when attempting to pause a zero address token
    error TokenAddressCannotBeZero();

    /// Error thrown when attempting to unpause a zero address token
    error UnpauseTokenCannotBeZero();

    /// Error thrown when attempting to pause an already paused token
    error TokenAlreadyPaused();

    /// Error thrown when attempting to unpause a token that is not paused
    error TokenNotPaused();

    /// Thrown when attempting to set the token pause manager to the zero address
    error TokenPauseManagerCannotBeZero();
}

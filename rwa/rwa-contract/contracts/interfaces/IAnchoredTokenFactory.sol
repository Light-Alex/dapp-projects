// SPDX-License-Identifier: BUSL-1.1
/**
 * @title  IAnchoredTokenFactory
 * @notice Comprehensive interface for the AnchoredTokenFactory contract containing all functions, events, and errors
 */
pragma solidity ^0.8.20;

import { IAccessControl } from "@openzeppelin/contracts/access/IAccessControl.sol";
import { IAnchoredRegistrar } from "contracts/interfaces/IAnchoredRegistrar.sol";

interface IAnchoredTokenFactory is IAccessControl {
    // ============ Functions ============

    /**
     * @notice Deploys a new Anchored token and registers it with both bridge and token manager
     * @param  name       The name of the token
     * @param  symbol     The token symbol
     * @param  tokenAdmin The address that will receive admin rights on the token
     * @return            The address of the deployed token proxy
     */
    function deployAndRegisterToken(string calldata name, string calldata symbol, address tokenAdmin)
        external
        returns (address);

    /**
     * @notice Deploys a new Anchored token without registering it anywhere
     * @param  name       The name of the token
     * @param  symbol     The token symbol
     * @param  tokenAdmin The address that will receive admin rights on the token
     * @return            The address of the deployed token proxy
     */
    function deployAnchoredTokenIsolated(string calldata name, string calldata symbol, address tokenAdmin)
        external
        returns (address);

    /**
     * @notice Pause the factory
     */
    function pause() external;

    /**
     * @notice Unpause the factory
     */
    function unpause() external;

    /**
     * @notice Updates the compliance contract address
     * @param  anchoredCompliance_ The new compliance contract address
     */
    function setCompliance(address anchoredCompliance_) external;

    /**
     * @notice Updates the token pause manager address
     * @param  tokenPauseManager_ The new token pause manager address
     */
    function setTokenPauseManager(address tokenPauseManager_) external;

    /**
     * @notice Updates the token manager registrar address
     * @param  tokenManagerRegistrar_ The new token manager registrar address
     */
    function setTokenManagerRegistrar(address tokenManagerRegistrar_) external;

    /**
     * @notice Clears a symbol in the edge case where we need to deploy a token
     *         with the same symbol as a previously deployed token
     * @param symbol The symbol to clear
     */
    function clearSymbol(string calldata symbol) external;

    // ============ View Functions ============

    // Note: hasRole, getRoleAdmin, getRoleMember, getRoleMemberCount, DEFAULT_ADMIN_ROLE are inherited from IAccessControl
    // Note: paused() is inherited from Pausable contract

    /**
     * @notice Returns the DEPLOYER_ROLE constant
     * @return The bytes32 value of DEPLOYER_ROLE
     */
    function DEPLOY_ROLE() external view returns (bytes32);

    /**
     * @notice Returns the CONFIGURE_ROLE constant
     * @return The bytes32 value of CONFIGURE_ROLE
     */
    function CONFIGURE_ROLE() external view returns (bytes32);

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

    /**
     * @notice Returns the beacon contract address
     * @return The address of the beacon contract
     */
    function BEACON() external view returns (address);

    /**
     * @notice Returns the compliance contract address
     * @return The address of the compliance contract
     */
    function anchoredCompliance() external view returns (address);

    /**
     * @notice Returns the token pause manager contract address
     * @return The address of the token pause manager contract
     */
    function anchoredTokenPauseManager() external view returns (address);

    /**
     * @notice Returns the token manager registrar contract
     * @return The token manager registrar contract
     */
    function tokenManagerRegistrar() external view returns (IAnchoredRegistrar);

    /**
     * @notice Returns whether a symbol already exists
     * @param symbol The symbol to check
     * @return True if the symbol exists
     */
    function symbolExists(string calldata symbol) external view returns (bool);

    // ============ Events ============

    /**
     * @notice Emitted when a new AnchoredCompliance contract is set
     * @param oldCompliance The old compliance contract address
     * @param newCompliance The new compliance contract address
     */
    event NewComplianceSet(address indexed oldCompliance, address indexed newCompliance);

    /**
     * @notice Emitted when a new token pause manager contract is set
     * @param oldTokenPauseManager The old token pause manager address
     * @param newTokenPauseManager The new token pause manager address
     */
    event NewTokenPauseManagerSet(address indexed oldTokenPauseManager, address indexed newTokenPauseManager);

    /**
     * @notice Emitted when a new token manager registrar contract is set
     * @param oldTokenManagerRegistrar The old token manager registrar address
     * @param newTokenManagerRegistrar The new token manager registrar address
     */
    event NewTokenManagerRegistrarSet(
        address indexed oldTokenManagerRegistrar, address indexed newTokenManagerRegistrar
    );

    /**
     * @notice Emitted when a new Anchored token is deployed (regardless of registration)
     * @param proxy             The address of the deployed token proxy
     * @param beacon            The address of the beacon contract
     * @param name              The name of the token
     * @param symbol            The token symbol
     * @param anchoredCompliance        The address of the AnchoredCompliance contract
     * @param anchoredTokenPauseManager The address of the token pause manager contract
     */
    event NewAnchoredTokenDeployed(
        address indexed proxy,
        address indexed beacon,
        string name,
        string symbol,
        address anchoredCompliance,
        address anchoredTokenPauseManager
    );

    /**
     * @notice Emitted when a symbol is set or cleared
     * @param symbol  The symbol that was set or cleared
     * @param status  The status of the symbol (true if set, false if cleared)
     */
    event SymbolSet(string indexed symbol, bool status);

    // ============ Errors ============

    /// Error thrown when attempting to set the compliance contract to zero address
    error ComplianceCannotBeZero();

    /// Error thrown when attempting to set the token pause manager to zero address
    error TokenPauseManagerCannotBeZero();

    /// Error thrown when attempting to set the token manager registrar to zero address
    error TokenManagerRegistrarCannotBeZero();

    /// Error thrown when attempting to deploy a token with an already existing symbol
    error SymbolAlreadyExists();
}

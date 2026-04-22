// SPDX-License-Identifier: BUSL-1.1
/*
 */
pragma solidity ^0.8.20;

import { AnchoredTokenManager } from "contracts/AnchoredTokenManager.sol";
import { IAnchoredTokenManagerRegistrar } from "contracts/interfaces/IAnchoredTokenManagerRegistrar.sol";
import {
    AccessControlEnumerable,
    IAccessControlEnumerable
} from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import { Pausable } from "@openzeppelin/contracts/utils/Pausable.sol";

/**
 * @title  TokenManagerRegistrar
 * @notice This contract is responsible for registering new tokens with the Anchored Token Management
 *         system. It allows for the registration of new tokens, setting the Anchored Token Manager,
 *         and configuring rate limits for those tokens.
 */
contract AnchoredTokenManagerRegistrar is
    IAnchoredTokenManagerRegistrar,
    AccessControlEnumerable,
    Pausable,
    Initializable
{
    /// Role for changing the token manager, rate limiter and configs
    bytes32 public constant CONFIGURE_ROLE = keccak256("CONFIGURE_ROLE");
    /// Role for the token factory that can register new tokens
    bytes32 public constant TOKEN_REGISTER_ROLE = keccak256("TOKEN_REGISTER_ROLE");
    /// Role allowed to pause the registrar
    bytes32 public constant PAUSE_ROLE = keccak256("PAUSE_ROLE");
    /// Role allowed to unpause the registrar
    bytes32 public constant UNPAUSE_ROLE = keccak256("UNPAUSE_ROLE");

    /// Address of the Anchored Token Manager that will handle minting/redeeming
    AnchoredTokenManager public anchoredTokenManager;

    /**
     * @notice Constructor for implementation contract
     */
    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initialize function for proxy deployment
     * @param guardian_ The address of the admin account that begins with the default admin role
     * @param anchoredTokenManager_ The address of the Anchored Token Manager contract
     */
    function initialize(address guardian_, address anchoredTokenManager_) external initializer {
        if (anchoredTokenManager_ == address(0)) revert AnchoredTokenManagerCannotBeZero();
        anchoredTokenManager = AnchoredTokenManager(anchoredTokenManager_);

        _grantRole(DEFAULT_ADMIN_ROLE, guardian_);
        _grantRole(CONFIGURE_ROLE, guardian_);
        _grantRole(PAUSE_ROLE, guardian_);
        _grantRole(UNPAUSE_ROLE, guardian_);
    }

    /**
     * @notice Pauses the registrar, disabling registration
     */
    function pause() external onlyRole(PAUSE_ROLE) {
        _pause();
    }

    /**
     * @notice Unpauses the registrar, enabling registration
     */
    function unpause() external onlyRole(UNPAUSE_ROLE) {
        _unpause();
    }

    /**
     * @notice Registers a new token with the Anchored Token Manager and configures rate limits
     * @param  token The address of the token to register
     * @dev    Only callable by accounts with TOKEN_FACTORY_ROLE
     */
    function register(address token) external override onlyRole(TOKEN_REGISTER_ROLE) whenNotPaused {
        if (token == address(0)) revert TokenAddressCannotBeZero();

        anchoredTokenManager.setAnchoredTokenRegistrationStatus(token, true);
        // Grant minter role to Anchored Token Manager
        IAccessControlEnumerable(token).grantRole(keccak256("MINT_ROLE"), address(anchoredTokenManager));
        // Grant burner role to Anchored Token Manager
        IAccessControlEnumerable(token).grantRole(keccak256("BURN_ROLE"), address(anchoredTokenManager));
        IAccessControlEnumerable(token).grantRole(keccak256("MULTIPLIER_UPDATE_ROLE"), address(anchoredTokenManager));
        emit TokenRegistered(token);
    }

    /**
     * @notice Sets or updates the Anchored Token Manager address
     * @param  anchoredTokenManager_ The new Anchored Token Manager address
     */
    function setAnchoredTokenManager(address anchoredTokenManager_) external onlyRole(CONFIGURE_ROLE) {
        if (anchoredTokenManager_ == address(0)) revert AnchoredTokenManagerCannotBeZero();

        emit AnchoredTokenManagerSet(address(anchoredTokenManager), anchoredTokenManager_);
        anchoredTokenManager = AnchoredTokenManager(anchoredTokenManager_);
    }
}

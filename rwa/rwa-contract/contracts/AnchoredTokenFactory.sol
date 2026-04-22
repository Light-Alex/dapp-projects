/**
 * SPDX-License-Identifier: BUSL-1.1
 */
pragma solidity ^0.8.20;

// Proxy admin contract used in OZ upgrades plugin
import { BeaconProxy } from "@openzeppelin/contracts/proxy/beacon/BeaconProxy.sol";
import { UpgradeableBeacon } from "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";
import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import { AnchoredToken } from "contracts/AnchoredToken.sol";
import { AccessControlEnumerable } from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import { IAnchoredRegistrar } from "contracts/interfaces/IAnchoredRegistrar.sol";
import { IAnchoredTokenFactory } from "contracts/interfaces/IAnchoredTokenFactory.sol";
import { ReentrancyGuard } from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import { Pausable } from "@openzeppelin/contracts/utils/Pausable.sol";

/**
 * @title  AnchoredTokenFactory
 * @notice This contract serves as a factory for deploying and configuring Anchored tokens with
 *         built-in compliance and pause management.
 *
 *         This contract allows for:
 *         - Deploying new Anchored tokens with preconfigured compliance and pause management
 *         - Registering the tokens for use via with bridge and token manager registrars
 *         - Isolated token deployments without registrar integration
 *
 * @dev    The contract uses OpenZeppelin's AccessControl for role-based permissions:
 *         - DEFAULT_ADMIN_ROLE: Can grant/revoke other roles
 *         - DEPLOY_ROLE: Can deploy new tokens
 *         - CONFIGURE_ROLE: Can update compliance and registrar settings
 */
contract AnchoredTokenFactory is
    IAnchoredTokenFactory,
    AccessControlEnumerable,
    ReentrancyGuard,
    Pausable,
    Initializable
{
    // Role used for deploying new tokens
    bytes32 public constant DEPLOY_ROLE = keccak256("DEPLOY_ROLE");
    // Role used for configuring the factory
    bytes32 public constant CONFIGURE_ROLE = keccak256("CONFIGURE_ROLE");
    // Role used to pause the factory
    bytes32 public constant PAUSE_ROLE = keccak256("PAUSE_ROLE");
    // Role used to unpause the factory
    bytes32 public constant UNPAUSE_ROLE = keccak256("UNPAUSE_ROLE");

    // Address of the BEACON contract used for proxy deployments
    address public immutable BEACON;

    // Address of the AnchoredCompliance contract
    address public anchoredCompliance;
    // Address of the token pause manager contract
    address public anchoredTokenPauseManager;
    // Address of the token registrar contract
    IAnchoredRegistrar public tokenManagerRegistrar;

    // Indicates if a token with the same symbol already exists
    mapping(string => bool) public symbolExists;

    /**
     * @notice Constructor for implementation contract
     */
    constructor() {
        _disableInitializers();

        // Initialize BEACON (immutable variable)
        address anchoredTokenImplementation = address(new AnchoredToken());
        // Use a temporary owner for the BEACON, will be transferred in initialize()
        UpgradeableBeacon _beacon = new UpgradeableBeacon(anchoredTokenImplementation, address(this));
        BEACON = address(_beacon);
    }

    /**
     * @notice Initialize function for proxy deployment
     * @param guardian_ The address which will be granted admin and other roles
     * @param anchoredCompliance_ The address of the AnchoredCompliance contract
     * @param tokenPauseManager_ The address of the token pause manager contract
     * @param tokenManagerRegistrar_ The address of the token manager registrar contract
     */
    function initialize(
        address guardian_,
        address anchoredCompliance_,
        address tokenPauseManager_,
        address tokenManagerRegistrar_
    ) external initializer {
        if (anchoredCompliance_ == address(0)) revert ComplianceCannotBeZero();
        if (tokenPauseManager_ == address(0)) revert TokenPauseManagerCannotBeZero();
        if (tokenManagerRegistrar_ == address(0)) {
            revert TokenManagerRegistrarCannotBeZero();
        }

        _grantRole(DEFAULT_ADMIN_ROLE, guardian_);
        _grantRole(DEPLOY_ROLE, guardian_);
        _grantRole(CONFIGURE_ROLE, guardian_);
        _grantRole(PAUSE_ROLE, guardian_);
        _grantRole(UNPAUSE_ROLE, guardian_);

        anchoredCompliance = anchoredCompliance_;
        anchoredTokenPauseManager = tokenPauseManager_;
        tokenManagerRegistrar = IAnchoredRegistrar(tokenManagerRegistrar_);
    }

    /**
     * @notice Pauses the factory, disabling new deployments
     */
    function pause() external onlyRole(PAUSE_ROLE) {
        _pause();
    }

    /**
     * @notice Unpauses the factory, enabling new deployments
     */
    function unpause() external onlyRole(UNPAUSE_ROLE) {
        _unpause();
    }

    /**
     * @notice Deploys a new Anchored token and registers it with both bridge and token manager
     * @param  name       The name of the token
     * @param  symbol     The token symbol
     * @param  tokenAdmin The address that will receive admin rights on the token
     * @return            The address of the deployed token proxy
     */
    function deployAndRegisterToken(string calldata name, string calldata symbol, address tokenAdmin)
        public
        override
        nonReentrant
        onlyRole(DEPLOY_ROLE)
        whenNotPaused
        returns (address)
    {
        AnchoredToken anchoredToken = AnchoredToken(_deployAnchoredToken(name, symbol));

        anchoredToken.grantRole(DEFAULT_ADMIN_ROLE, address(tokenManagerRegistrar));
        tokenManagerRegistrar.register(address(anchoredToken));
        anchoredToken.revokeRole(DEFAULT_ADMIN_ROLE, address(tokenManagerRegistrar));

        anchoredToken.grantRole(DEFAULT_ADMIN_ROLE, tokenAdmin);
        anchoredToken.renounceRole(DEFAULT_ADMIN_ROLE, address(this));

        return address(anchoredToken);
    }

    /**
     * @notice Deploys a new Anchored token without registering it anywhere
     * @param  name       The name of the token
     * @param  symbol     The token symbol
     * @param  tokenAdmin The address that will receive admin rights on the token
     * @return            The address of the deployed token proxy
     */
    function deployAnchoredTokenIsolated(string calldata name, string calldata symbol, address tokenAdmin)
        external
        nonReentrant
        onlyRole(DEPLOY_ROLE)
        whenNotPaused
        returns (address)
    {
        AnchoredToken anchoredToken = AnchoredToken(_deployAnchoredToken(name, symbol));

        anchoredToken.grantRole(DEFAULT_ADMIN_ROLE, tokenAdmin);
        anchoredToken.renounceRole(DEFAULT_ADMIN_ROLE, address(this));

        return address(anchoredToken);
    }

    /**
     * @notice Internal function to deploy a new Anchored token
     * @param  name   The name of the token
     * @param  symbol The token symbol
     * @return        The address of the deployed token proxy
     */
    function _deployAnchoredToken(string calldata name, string calldata symbol) internal returns (address) {
        if (symbolExists[symbol]) revert SymbolAlreadyExists();
        BeaconProxy anchoredTokenProxy = new BeaconProxy(BEACON, "");
        AnchoredToken anchoredTokenProxied = AnchoredToken(address(anchoredTokenProxy));
        anchoredTokenProxied.initialize(name, symbol, anchoredCompliance, anchoredTokenPauseManager);
        symbolExists[symbol] = true;
        emit SymbolSet(symbol, true);

        emit NewAnchoredTokenDeployed(
            address(anchoredTokenProxied), BEACON, name, symbol, anchoredCompliance, anchoredTokenPauseManager
        );

        return address(anchoredTokenProxied);
    }

    /**
     * @notice Updates the compliance contract address
     * @param  anchoredCompliance_ The new compliance contract address
     */
    function setCompliance(address anchoredCompliance_) external onlyRole(CONFIGURE_ROLE) {
        if (anchoredCompliance_ == address(0)) revert ComplianceCannotBeZero();
        address oldCompliance = anchoredCompliance;
        anchoredCompliance = anchoredCompliance_;
        emit NewComplianceSet(oldCompliance, anchoredCompliance);
    }

    /**
     * @notice Updates the token pause manager address
     * @param  tokenPauseManager_ The new token pause manager address
     */
    function setTokenPauseManager(address tokenPauseManager_) external onlyRole(CONFIGURE_ROLE) {
        if (tokenPauseManager_ == address(0)) revert TokenPauseManagerCannotBeZero();
        address oldTokenPauseManager = anchoredTokenPauseManager;
        anchoredTokenPauseManager = tokenPauseManager_;
        emit NewTokenPauseManagerSet(oldTokenPauseManager, anchoredTokenPauseManager);
    }

    /**
     * @notice Updates the token manager registrar address
     * @param  tokenManagerRegistrar_ The new token manager registrar address
     */
    function setTokenManagerRegistrar(address tokenManagerRegistrar_) external onlyRole(CONFIGURE_ROLE) {
        if (tokenManagerRegistrar_ == address(0)) {
            revert TokenManagerRegistrarCannotBeZero();
        }
        emit NewTokenManagerRegistrarSet(address(tokenManagerRegistrar), tokenManagerRegistrar_);
        tokenManagerRegistrar = IAnchoredRegistrar(tokenManagerRegistrar_);
    }

    /**
     * @notice Clears a symbol in the edge case where we need to deploy a token
     *         with the same symbol as a previously deployed token
     * @param symbol The symbol to clear
     */
    function clearSymbol(string calldata symbol) external onlyRole(CONFIGURE_ROLE) {
        symbolExists[symbol] = false;
        emit SymbolSet(symbol, false);
    }
}

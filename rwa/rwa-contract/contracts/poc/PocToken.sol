
// SPDX-License-Identifier: UNLICENSED
/**
 * Copyright (c) 2023, Anchored Finance
 */
pragma solidity ^0.8.20;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { AccessControlEnumerable } from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";

/**
 * @title  PendingToken
 * @notice A non-transferable ERC20 token that can only be minted and burned by authorized contracts
 */
contract PocToken is ERC20, AccessControlEnumerable, Initializable {
    /// @notice Role identifier for those who can mint tokens
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");

    /// @notice Role identifier for those who can burn tokens
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");

    /// @notice The Gate contract that created this token
    address public immutable GATE_CONTRACT;


    /// Override for the name allowing the name to be changed
    string internal _name;
    /// Override for the symbol allowing the symbol to be changed
    string internal _symbol;

    /**
     * @notice Constructor
     * @param gateContract_ The address of the Gate contract
     */

    constructor(address gateContract_) ERC20("", "") {
        GATE_CONTRACT = gateContract_;
        // Disable initializers to prevent direct initialization of implementation
        _disableInitializers();
    }


    /**
     * @notice Initialize function for beacon proxy deployment
     * @param name_ The token name
     * @param symbol_ The token symbol
     */
    function initialize(
        string memory name_,
        string memory symbol_
    ) external virtual initializer {
        // Grant roles to the caller (factory)
        _grantRole(DEFAULT_ADMIN_ROLE, _msgSender());
        _grantRole(MINTER_ROLE, _msgSender());
        _grantRole(BURNER_ROLE, _msgSender());

        // Set the name and symbol
        _name = name_;
        _symbol = symbol_;
    }

    function setOperator(address operator) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (operator == address(0)) revert AddressCannotBeZero();

        _grantRole(MINTER_ROLE, operator);
        _grantRole(BURNER_ROLE, operator);
    }

    /**
     * @notice Mint tokens to a specific address
     * @param to The address to mint tokens to
     * @param amount The amount of tokens to mint
     */
    function mint(address to, uint256 amount) external onlyRole(MINTER_ROLE) {
        if (to == address(0)) revert AddressCannotBeZero();
        if (amount == 0) revert AmountCannotBeZero();

        _mint(to, amount);
        emit TokensMinted(to, amount);
    }

    /**
     * @notice Burn tokens from the caller
     * @param amount The amount of tokens to burn
     */
    function burn(uint256 amount) external onlyRole(BURNER_ROLE) {
        if (amount == 0) revert AmountCannotBeZero();

        _burn(_msgSender(), amount);
        emit TokensBurned(_msgSender(), amount);
    }

    /**
     * @notice Burn tokens from a specific address
     * @param from The address to burn tokens from
     * @param amount The amount of tokens to burn
     */
    function burnFrom(address from, uint256 amount) external onlyRole(BURNER_ROLE) {
        if (from == address(0)) revert AddressCannotBeZero();
        if (amount == 0) revert AmountCannotBeZero();

        _burn(from, amount);
        emit TokensBurned(from, amount);
    }


    /**
     * @notice Check if an address has the minter role
     * @param account The address to check
     * @return True if the address has the minter role
     */
    function isMinter(address account) external view returns (bool) {
        return hasRole(MINTER_ROLE, account);
    }

    /**
     * @notice Check if an address has the burner role
     * @param account The address to check
     * @return True if the address has the burner role
     */
    function isBurner(address account) external view returns (bool) {
        return hasRole(BURNER_ROLE, account);
    }

        /**
     * @notice Returns the name of the token. Overrides the default name allowing the name to be changed
     *      after deployment.
     */
    function name() public view virtual override(ERC20) returns (string memory) {
        return _name;
    }

    /**
     * @notice Returns the symbol of the token. Overrides the default symbol allowing the symbol to be changed
     *      after deployment.
     */
    function symbol() public view virtual override(ERC20) returns (string memory) {
        return _symbol;
    }

    // ============ Events ============

    /**
     * @notice Event emitted when tokens are minted
     * @param to The address tokens were minted to
     * @param amount The amount of tokens minted
     */
    event TokensMinted(address indexed to, uint256 amount);

    /**
     * @notice Event emitted when tokens are burned
     * @param from The address tokens were burned from
     * @param amount The amount of tokens burned
     */
    event TokensBurned(address indexed from, uint256 amount);

    // ============ Errors ============


    /// Error emitted when the Gate contract address is zero
    error GateContractCannotBeZero();

       /// Error emitted when an amount is zero
    error AmountCannotBeZero();

    /// Error emitted when an address is zero
    error AddressCannotBeZero();
}

// SPDX-License-Identifier: UNLICENSED
/**
 * Copyright (c) 2023, Anchored Finance
 */
pragma solidity ^0.8.20;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { AccessControl } from "@openzeppelin/contracts/access/AccessControl.sol";

/**
 * @title  PendingToken
 * @notice A non-transferable ERC20 token that can only be minted and burned by authorized contracts
 */
contract PendingToken is ERC20, AccessControl {
    /// @notice Role identifier for those who can mint tokens
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");

    /// @notice Role identifier for those who can burn tokens
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");

    /// @notice The Gate contract that created this token
    address public immutable GATE_CONTRACT;

    /**
     * @notice Constructor
     * @param name_ The name of the token
     * @param symbol_ The symbol of the token
     * @param gateContract_ The address of the Gate contract
     */
    constructor(string memory name_, string memory symbol_, address gateContract_) ERC20(name_, symbol_) {
        if (gateContract_ == address(0)) revert GateContractCannotBeZero();

        GATE_CONTRACT = gateContract_;

        // Grant roles to the Gate contract
        _grantRole(DEFAULT_ADMIN_ROLE, gateContract_);
        _grantRole(MINTER_ROLE, gateContract_);
        _grantRole(BURNER_ROLE, gateContract_);
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
    }

    /**
     * @notice Burn tokens from the caller
     * @param amount The amount of tokens to burn
     */
    function burn(uint256 amount) external onlyRole(BURNER_ROLE) {
        if (amount == 0) revert AmountCannotBeZero();

        _burn(_msgSender(), amount);
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
    }

    /**
     * @notice Override transfer to make it non-transferable
     * @dev This function always reverts to prevent transfers
     */
    function transfer(address, uint256) public pure override returns (bool) {
        revert TransferNotAllowed();
    }

    /**
     * @notice Override transferFrom to make it non-transferable
     * @dev This function always reverts to prevent transfers
     */
    function transferFrom(address, address, uint256) public pure override returns (bool) {
        revert TransferNotAllowed();
    }

    /**
     * @notice Override approve to make it non-transferable
     * @dev This function always reverts to prevent approvals
     */
    function approve(address, uint256) public pure override returns (bool) {
        revert ApprovalNotAllowed();
    }

    /**
     * @notice Override increaseAllowance to make it non-transferable
     * @dev This function always reverts to prevent allowance increases
     */
    function increaseAllowance(address, uint256) public pure returns (bool) {
        revert ApprovalNotAllowed();
    }

    /**
     * @notice Override decreaseAllowance to make it non-transferable
     * @dev This function always reverts to prevent allowance decreases
     */
    function decreaseAllowance(address, uint256) public pure returns (bool) {
        revert ApprovalNotAllowed();
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

    /// Error emitted when trying to transfer tokens
    error TransferNotAllowed();

    /// Error emitted when trying to approve tokens
    error ApprovalNotAllowed();

    /// Error emitted when an address is zero
    error AddressCannotBeZero();

    /// Error emitted when an amount is zero
    error AmountCannotBeZero();

    /// Error emitted when the Gate contract address is zero
    error GateContractCannotBeZero();
}

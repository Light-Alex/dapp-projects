// SPDX-License-Identifier: BUSL-1.1
/**
 * @title  IAnchoredToken
 * @notice Comprehensive interface for the AnchoredToken contract containing all functions, events, and errors
 */
pragma solidity ^0.8.20;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { IERC20Metadata } from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import { IAccessControl } from "@openzeppelin/contracts/access/IAccessControl.sol";

interface IAnchoredToken is IERC20, IERC20Metadata, IAccessControl {
    // ============ Functions ============

    /**
     * @notice Initialize the token with name, symbol, compliance, and pause manager
     * @param name_ The name of the token
     * @param symbol_ The symbol of the token
     * @param anchoredCompliance_ The compliance contract address
     * @param tokenPauseManager_ The token pause manager address
     */
    function initialize(
        string memory name_,
        string memory symbol_,
        address anchoredCompliance_,
        address tokenPauseManager_
    ) external;

    /**
     * @notice Sets the token name
     * @param name_ New token name
     */
    function setName(string memory name_) external;

    /**
     * @notice Sets the token symbol
     * @param symbol_ New token symbol
     */
    function setSymbol(string memory symbol_) external;

    /**
     * @notice Sets the compliance address
     * @param anchoredCompliance_ New compliance address
     */
    function setCompliance(address anchoredCompliance_) external;

    /**
     * @notice Sets the token pause manager address
     * @param tokenPauseManager_ New token pause manager address
     */
    function setTokenPauseManager(address tokenPauseManager_) external;

    /**
     * @notice Mints a specific amount of tokens
     * @param to The account who will receive the minted tokens
     * @param amount The amount of tokens to be minted
     */
    function mint(address to, uint256 amount) external;

    /**
     * @notice Burns a specific amount of tokens
     * @param amount The amount of tokens to be burned
     */
    function burn(uint256 amount) external;

    /**
     * @notice Updates the multiplier value
     * @param newMultiplier The new multiplier value
     */
    function updateMultiplier(uint256 newMultiplier) external;

    /**
     * @notice Transfers shares to a specified address
     * @param to The address to transfer shares to
     * @param sharesAmount The amount of shares to transfer
     * @return True if the transfer was successful
     */
    function transferShares(address to, uint256 sharesAmount) external returns (bool);

    // ============ View Functions ============

    /**
     * @notice Returns amount of shares owned by given account
     * @param account The account to query shares for
     * @return The amount of shares owned by the account
     */
    function sharesOf(address account) external view returns (uint256);

    /**
     * @notice Returns the total amount of shares
     * @return The total shares amount
     */
    function totalShares() external view returns (uint256);

    /**
     * @notice Returns the current multiplier value
     * @return The current multiplier value
     */
    function multiplier() external view returns (uint256);

    /**
     * @notice Returns the multiplier nonce for tracking updates
     * @return The multiplier nonce
     */
    function multiplierNonce() external view returns (uint256);

    /**
     * @notice Returns the MULTIPLIER_UPDATE_ROLE constant
     * @return The bytes32 value of MULTIPLIER_UPDATE_ROLE
     */
    function MULTIPLIER_UPDATE_ROLE() external view returns (bytes32);

    /**
     * @notice Returns the amount of shares that corresponds to underlying amount
     * @param underlyingAmount The underlying token amount
     * @return The corresponding shares amount
     */
    function getSharesByUnderlyingAmount(uint256 underlyingAmount) external view returns (uint256);

    /**
     * @notice Returns the amount of underlying that corresponds to shares amount
     * @param sharesAmount The shares amount
     * @return The corresponding underlying token amount
     */
    function getUnderlyingAmountByShares(uint256 sharesAmount) external view returns (uint256);

    // ============ Events ============

    /**
     * @notice Emitted when the token name is changed
     * @param oldName The old token name
     * @param newName The new token name
     */
    event NameChanged(string oldName, string newName);

    /**
     * @notice Emitted when the token symbol is changed
     * @param oldSymbol The old token symbol
     * @param newSymbol The new token symbol
     */
    event SymbolChanged(string oldSymbol, string newSymbol);

    /**
     * @notice Emitted when `value` token shares are moved from one account (`from`) to another (`to`)
     * @param from The account shares are transferred from
     * @param to The account shares are transferred to
     * @param value The amount of shares transferred
     */
    event TransferShares(address indexed from, address indexed to, uint256 value);

    /**
     * @notice Emitted when multiplier value is updated
     * @param value The new multiplier value
     */
    event MultiplierUpdated(uint256 value);

    /**
     * @notice Emitted when the compliance address is set
     * @param oldCompliance The old compliance address
     * @param newCompliance The new compliance address
     */
    event ComplianceSet(address indexed oldCompliance, address indexed newCompliance);

    /**
     * @notice Emitted when the token pause manager address is changed
     * @param oldTokenPauseManager The old token pause manager address
     * @param newTokenPauseManager The new token pause manager address
     */
    event TokenPauseManagerSet(address indexed oldTokenPauseManager, address indexed newTokenPauseManager);

    // ============ Errors ============

    /// Error thrown when attempting to transfer from the zero address
    error TransferFromCannotBeZero();

    /// Error thrown when attempting to transfer to the zero address
    error TransferToCannotBeZero();

    /// Error thrown when transfer amount exceeds balance
    error TransferAmountExceedsBalance();

    /// Error thrown when attempting to mint to the zero address
    error MintToCannotBeZero();

    /// Error thrown when attempting to burn from the zero address
    error BurnFromCannotBeZero();

    /// Error thrown when burn amount exceeds balance
    error BurnAmountExceedsBalance();

    /// Error thrown when compliance address is zero
    error ComplianceCannotBeZero();

    /// Error thrown when token is paused
    error TokenPaused();

    /// Error thrown when token pause manager address is zero
    error TokenPauseManagerCannotBeZero();
}

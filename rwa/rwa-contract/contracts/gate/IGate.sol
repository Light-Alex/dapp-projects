// SPDX-License-Identifier: UNLICENSED
/**
 * Copyright (c) 2023, Anchored Finance
 */
pragma solidity ^0.8.20;

/**
 * @title  IGate
 * @notice Interface for the Gate contract that manages deposit and withdrawal flows
 */
interface IGate {
    // ============ Enums ============

    enum OperationStatus {
        /// Operation is pending backend processing
        PENDING,
        /// Operation has been processed and is active
        ACTIVE,
        /// Operation has been redeemed/completed
        REDEEMED
    }

    // ============ Structs ============

    /**
     * @notice Deposit operation struct
     * @param user The user who made the deposit
     * @param usdcAmount The amount of USDC deposited
     * @param pendingAncUSDCAmount The amount of pending ancUSDC minted
     * @param status The current status of the deposit
     * @param timestamp The timestamp when the deposit was made
     */
    struct DepositOperation {
        address user;
        uint256 usdcAmount;
        uint256 pendingAncUSDCAmount;
        OperationStatus status;
        uint256 timestamp;
    }

    /**
     * @notice Withdrawal operation struct
     * @param user The user who initiated the withdrawal
     * @param ancUSDCAmount The amount of ancUSDC to withdraw
     * @param pendingUSDCAmount The amount of pending USDC minted
     * @param status The current status of the withdrawal
     * @param timestamp The timestamp when the withdrawal was initiated
     */
    struct WithdrawalOperation {
        address user;
        uint256 ancUSDCAmount;
        uint256 pendingUSDCAmount;
        OperationStatus status;
        uint256 timestamp;
    }

    // ============ Functions ============

    /**
     * @notice Deposit USDC and receive pending ancUSDC
     * @param usdcAmount The amount of USDC to deposit
     * @return operationId The ID of the deposit operation
     */
    function deposit(uint256 usdcAmount) external returns (bytes32 operationId);

    /**
     * @notice Withdraw ancUSDC and receive pending USDC
     * @param ancUSDCAmount The amount of ancUSDC to withdraw
     * @return operationId The ID of the withdrawal operation
     */
    function withdraw(uint256 ancUSDCAmount) external returns (bytes32 operationId);

    /**
     * @notice Process pending deposit (backend function)
     * @param operationId The ID of the deposit operation to process
     * @param ancUSDCAmount The amount of ancUSDC to mint
     */
    function processDeposit(bytes32 operationId, uint256 ancUSDCAmount) external;

    /**
     * @notice Process pending withdrawal (backend function)
     * @param operationId The ID of the withdrawal operation to process
     * @param usdcAmount The amount of USDC to transfer back
     */
    function processWithdrawal(bytes32 operationId, uint256 usdcAmount) external;

    /**
     * @notice Set the minimum deposit amount
     * @param amount The minimum deposit amount in USDC
     */
    function setMinimumDepositAmount(uint256 amount) external;

    /**
     * @notice Set the minimum withdrawal amount
     * @param amount The minimum withdrawal amount in ancUSDC
     */
    function setMinimumWithdrawalAmount(uint256 amount) external;

    /**
     * @notice Pause deposits
     */
    function pauseDeposits() external;

    /**
     * @notice Unpause deposits
     */
    function unpauseDeposits() external;

    /**
     * @notice Pause withdrawals
     */
    function pauseWithdrawals() external;

    /**
     * @notice Unpause withdrawals
     */
    function unpauseWithdrawals() external;

    /**
     * @notice Initialize the contract
     * @param usdc_ The USDC token address
     * @param ancUSDC_ The ancUSDC token address
     * @param guardian_ The guardian address
     * @param minimumDepositAmount_ The minimum deposit amount
     * @param minimumWithdrawalAmount_ The minimum withdrawal amount
     */
    function initialize(
        address usdc_,
        address ancUSDC_,
        address guardian_,
        uint256 minimumDepositAmount_,
        uint256 minimumWithdrawalAmount_
    ) external;

    // ============ View Functions ============

    /**
     * @notice Get deposit operation details
     * @param operationId The operation ID
     * @return The deposit operation struct
     */
    function getDepositOperation(bytes32 operationId) external view returns (DepositOperation memory);

    /**
     * @notice Get withdrawal operation details
     * @param operationId The operation ID
     * @return The withdrawal operation struct
     */
    function getWithdrawalOperation(bytes32 operationId) external view returns (WithdrawalOperation memory);

    /**
     * @notice Check if deposits are paused
     * @return True if deposits are paused
     */
    function depositsArePaused() external view returns (bool);

    /**
     * @notice Check if withdrawals are paused
     * @return True if withdrawals are paused
     */
    function withdrawalsArePaused() external view returns (bool);

    // ============ Events ============

    /**
     * @notice Event emitted when a deposit is initiated
     * @param operationId The operation ID
     * @param user The user who made the deposit
     * @param usdcAmount The amount of USDC deposited
     * @param pendingAncUSDCAmount The amount of pending ancUSDC minted
     */
    event PendingDeposit(
        bytes32 indexed operationId, address indexed user, uint256 usdcAmount, uint256 pendingAncUSDCAmount
    );

    /**
     * @notice Event emitted when a deposit is processed
     * @param operationId The operation ID
     * @param user The user who made the deposit
     * @param ancUSDCAmount The amount of ancUSDC minted
     */
    event DepositProcessed(bytes32 indexed operationId, address indexed user, uint256 ancUSDCAmount);

    /**
     * @notice Event emitted when a withdrawal is initiated
     * @param operationId The operation ID
     * @param user The user who initiated the withdrawal
     * @param ancUSDCAmount The amount of ancUSDC to withdraw
     * @param pendingUSDCAmount The amount of pending USDC minted
     */
    event PendingWithdraw(
        bytes32 indexed operationId, address indexed user, uint256 ancUSDCAmount, uint256 pendingUSDCAmount
    );

    /**
     * @notice Event emitted when a withdrawal is processed
     * @param operationId The operation ID
     * @param user The user who initiated the withdrawal
     * @param usdcAmount The amount of USDC transferred
     */
    event WithdrawalProcessed(bytes32 indexed operationId, address indexed user, uint256 usdcAmount);

    /**
     * @notice Event emitted when minimum deposit amount is set
     * @param oldAmount The old minimum deposit amount
     * @param newAmount The new minimum deposit amount
     */
    event MinimumDepositAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);

    /**
     * @notice Event emitted when minimum withdrawal amount is set
     * @param oldAmount The old minimum withdrawal amount
     * @param newAmount The new minimum withdrawal amount
     */
    event MinimumWithdrawalAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);

    /**
     * @notice Event emitted when deposits are paused
     */
    event DepositsPaused();

    /**
     * @notice Event emitted when deposits are unpaused
     */
    event DepositsUnpaused();

    /**
     * @notice Event emitted when withdrawals are paused
     */
    event WithdrawalsPaused();

    /**
     * @notice Event emitted when withdrawals are unpaused
     */
    event WithdrawalsUnpaused();

    // ============ Errors ============

    /// Error emitted when the operation ID is invalid
    error InvalidOperationId();

    /// Error emitted when the operation status is invalid
    error InvalidOperationStatus();

    /// Error emitted when the deposit amount is too small
    error DepositAmountTooSmall();

    /// Error emitted when the withdrawal amount is too small
    error WithdrawalAmountTooSmall();

    /// Error emitted when deposits are paused
    error DepositsArePaused();

    /// Error emitted when withdrawals are paused
    error WithdrawalsArePaused();

    /// Error emitted when the caller is not authorized
    error NotAuthorized();

    /// Error emitted when an address is zero
    error AddressCannotBeZero();

    /// Error emitted when the amount is zero
    error AmountCannotBeZero();

    /// Error emitted when the operation has already been processed
    error OperationAlreadyProcessed();
}

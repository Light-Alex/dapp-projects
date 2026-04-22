// SPDX-License-Identifier: UNLICENSED
/**
 * Copyright (c) 2023, Anchored Finance
 */
pragma solidity ^0.8.20;

import { IPocToken } from "./IPocToken.sol";
import { AccessControlEnumerable } from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import { ReentrancyGuard } from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { IERC20Metadata } from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title  PocGate
 * @notice This contract receives USDC and mints USDM
 */
contract PocGate is AccessControlEnumerable, ReentrancyGuard, Initializable {
    using SafeERC20 for IERC20;

    /// @notice Role identifier for those who can configure the contract
    bytes32 public constant CONFIGURE_ROLE = keccak256("CONFIGURE_ROLE");

    /// @notice Role identifier for those who can pause operations
    bytes32 public constant PAUSE_ROLE = keccak256("PAUSE_ROLE");

    /// @notice The address of the USDC token
    address public immutable USDC;

    /// @notice The address of the ancUSDC token
    address public immutable USDM;

    /// @notice Minimum USDC amount required for deposits
    uint256 public minimumDepositAmount;

    /// @notice Minimum ancUSDC amount required for withdrawals
    uint256 public minimumWithdrawalAmount;

    /// @notice Whether deposits are paused
    bool public depositsArePaused;

    /// @notice Whether withdrawals are paused
    bool public withdrawalsArePaused;
   // ============ Events ============

    event Deposited(address indexed user, uint256 usdcAmount, uint256 usdmAmount);
    event Withdrawn(address indexed user, uint256 usdmAmount, uint256 usdcAmount);

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


    /**
     * @notice Constructor for implementation contract
     * @param usdc_ The address of the USDC token
     * @param usdm_ The address of the ancUSDC token
     */
    constructor(address usdc_, address usdm_) {
        _disableInitializers();

        // Initialize immutable variables
        USDC = usdc_;
        USDM = usdm_;
    }

    /**
     * @notice Initialize function for proxy deployment
     * @param guardian_ The guardian address
     * @param minimumDepositAmount_ The minimum deposit amount
     * @param minimumWithdrawalAmount_ The minimum withdrawal amount
     */
    function initialize(
        address guardian_,
        uint256 minimumDepositAmount_,
        uint256 minimumWithdrawalAmount_
    ) external initializer {

        if (guardian_ == address(0)) revert AddressCannotBeZero();

        minimumDepositAmount = minimumDepositAmount_;
        minimumWithdrawalAmount = minimumWithdrawalAmount_;


        _grantRole(DEFAULT_ADMIN_ROLE, guardian_);
        _grantRole(CONFIGURE_ROLE, guardian_);
        _grantRole(PAUSE_ROLE, guardian_);
    }

    /**
     * @notice Modifier to check if deposits are not paused
     */
    modifier whenDepositsNotPaused() {
        if (depositsArePaused) revert DepositsArePaused();
        _;
    }

    /**
     * @notice Modifier to check if withdrawals are not paused
     */
    modifier whenWithdrawalsNotPaused() {
        if (withdrawalsArePaused) revert WithdrawalsArePaused();
        _;
    }

    /**
     * @notice Deposit USDC and receive pending ancUSDC
     * @param usdcAmount The amount of USDC to deposit
     * @return operationId The ID of the deposit operation
     */
    function deposit(uint256 usdcAmount) external nonReentrant whenDepositsNotPaused returns (bytes32 operationId) {
        if (usdcAmount == 0) revert AmountCannotBeZero();
        if (usdcAmount < minimumDepositAmount) revert DepositAmountTooSmall();

        // Generate unique operation ID
        operationId = bytes32(uint(0));

        // Transfer USDC from user to contract
        IERC20(USDC).safeTransferFrom(_msgSender(), address(this), usdcAmount);

        uint256 pendingUSDMAmount = _convertDecimals(usdcAmount, USDC, USDM);
        IPocToken(USDM).mint(_msgSender(), pendingUSDMAmount);

        emit Deposited(_msgSender(), usdcAmount, pendingUSDMAmount);
    }

    /**
     * @notice Withdraw ancUSDC and receive pending USDC
     * @param usdmAMOUNT The amount of ancUSDC to withdraw
     * @return operationId The ID of the withdrawal operation
     */
    function withdraw(uint256 usdmAMOUNT)
        external
        nonReentrant
        whenWithdrawalsNotPaused
        returns (bytes32 operationId)
    {
        if (usdmAMOUNT == 0) revert AmountCannotBeZero();
        if (usdmAMOUNT < minimumWithdrawalAmount) revert WithdrawalAmountTooSmall();

        // Generate unique operation ID
        operationId = bytes32(uint(0));

        // Burn ancUSDC from user
        IERC20(USDM).safeTransferFrom(_msgSender(), address(this), usdmAMOUNT);
        IPocToken(USDM).burn(usdmAMOUNT);

        // Mint pending USDC to user (convert from ancUSDC decimals to USDC decimals)
        uint256 pendingUSDCAmount = _convertDecimals(usdmAMOUNT, USDM, USDC);

        IERC20(USDC).safeTransfer(_msgSender(), pendingUSDCAmount);

        emit Withdrawn(_msgSender(), usdmAMOUNT, pendingUSDCAmount);
    }


    /**
     * @notice Set the minimum deposit amount
     * @param amount The minimum deposit amount in USDC
     */
    function setMinimumDepositAmount(uint256 amount) external onlyRole(CONFIGURE_ROLE) {
        emit MinimumDepositAmountSet(minimumDepositAmount, amount);
        minimumDepositAmount = amount;
    }

    /**
     * @notice Set the minimum withdrawal amount
     * @param amount The minimum withdrawal amount in ancUSDC
     */
    function setMinimumWithdrawalAmount(uint256 amount) external onlyRole(CONFIGURE_ROLE) {
        emit MinimumWithdrawalAmountSet(minimumWithdrawalAmount, amount);
        minimumWithdrawalAmount = amount;
    }

    /**
     * @notice Pause deposits
     */
    function pauseDeposits() external onlyRole(PAUSE_ROLE) {
        depositsArePaused = true;
        emit DepositsPaused();
    }

    /**
     * @notice Unpause deposits
     */
    function unpauseDeposits() external onlyRole(PAUSE_ROLE) {
        depositsArePaused = false;
        emit DepositsUnpaused();
    }

    /**
     * @notice Pause withdrawals
     */
    function pauseWithdrawals() external onlyRole(PAUSE_ROLE) {
        withdrawalsArePaused = true;
        emit WithdrawalsPaused();
    }

    /**
     * @notice Unpause withdrawals
     */
    function unpauseWithdrawals() external onlyRole(PAUSE_ROLE) {
        withdrawalsArePaused = false;
        emit WithdrawalsUnpaused();
    }


    /**
     * @notice Convert amount from one token's decimals to another token's decimals
     * @param amount The amount to convert
     * @param fromToken The source token address
     * @param toToken The target token address
     * @return The converted amount
     */
    function _convertDecimals(uint256 amount, address fromToken, address toToken) internal view returns (uint256) {
        uint8 fromDecimals = IERC20Metadata(fromToken).decimals();
        uint8 toDecimals = IERC20Metadata(toToken).decimals();

        if (fromDecimals == toDecimals) {
            return amount;
        } else if (fromDecimals < toDecimals) {
            // Scale up: multiply by 10^(toDecimals - fromDecimals)
            return amount * (10 ** (toDecimals - fromDecimals));
        } else {
            // Scale down: divide by 10^(fromDecimals - toDecimals)
            return amount / (10 ** (fromDecimals - toDecimals));
        }
    }
}

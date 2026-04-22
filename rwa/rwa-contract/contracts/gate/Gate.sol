// SPDX-License-Identifier: UNLICENSED
/**
 * Copyright (c) 2023, Anchored Finance
 */
pragma solidity ^0.8.20;

import { IGate } from "contracts/gate/IGate.sol";
import { IAnchoredTokenLike } from "contracts/interfaces/IAnchoredLike.sol";
import { PendingAncUSDC } from "contracts/gate/PendingAncUSDC.sol";
import { PendingUSDC } from "contracts/gate/PendingUSDC.sol";
import { PendingToken } from "contracts/gate/PendingToken.sol";
import { AccessControlEnumerable } from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import { ReentrancyGuard } from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { IERC20Metadata } from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title  Gate
 * @notice This contract manages the deposit and withdrawal flows for USDC and ancUSDC
 */
contract Gate is IGate, AccessControlEnumerable, ReentrancyGuard, Initializable {
    using SafeERC20 for IERC20;

    /// @notice Role identifier for those who can configure the contract
    bytes32 public constant CONFIGURE_ROLE = keccak256("CONFIGURE_ROLE");

    /// @notice Role identifier for those who can pause operations
    bytes32 public constant PAUSE_ROLE = keccak256("PAUSE_ROLE");

    /// @notice Role identifier for those who can process operations (backend)
    bytes32 public constant PROCESSOR_ROLE = keccak256("PROCESSOR_ROLE");

    /// @notice The address of the USDC token
    address public immutable USDC;

    /// @notice The address of the ancUSDC token
    address public immutable ANC_USDC;

    /// @notice The address of the pending ancUSDC token
    address public PENDING_ANC_USDC;

    /// @notice The address of the pending USDC token
    address public PENDING_USDC;

    /// @notice Minimum USDC amount required for deposits
    uint256 public minimumDepositAmount;

    /// @notice Minimum ancUSDC amount required for withdrawals
    uint256 public minimumWithdrawalAmount;

    /// @notice Whether deposits are paused
    bool public depositsArePaused;

    /// @notice Whether withdrawals are paused
    bool public withdrawalsArePaused;

    /// @notice Counter for generating unique operation IDs
    uint256 private _operationCounter;

    /// @notice Mapping of operation ID to deposit operations
    mapping(bytes32 => DepositOperation) public depositOperations;

    /// @notice Mapping of operation ID to withdrawal operations
    mapping(bytes32 => WithdrawalOperation) public withdrawalOperations;

    /**
     * @notice Constructor for implementation contract
     * @param usdc_ The address of the USDC token
     * @param ancUSDC_ The address of the ancUSDC token
     */
    constructor(address usdc_, address ancUSDC_) {
        _disableInitializers();

        // Initialize immutable variables
        USDC = usdc_;
        ANC_USDC = ancUSDC_;
    }

    /**
     * @notice Initialize function for proxy deployment
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
    ) external initializer {
        if (usdc_ == address(0)) revert AddressCannotBeZero();
        if (ancUSDC_ == address(0)) revert AddressCannotBeZero();
        if (guardian_ == address(0)) revert AddressCannotBeZero();

        minimumDepositAmount = minimumDepositAmount_;
        minimumWithdrawalAmount = minimumWithdrawalAmount_;

        // Create pending token contracts
        PENDING_ANC_USDC = address(new PendingAncUSDC(address(this)));
        PENDING_USDC = address(new PendingUSDC(address(this)));

        _grantRole(DEFAULT_ADMIN_ROLE, guardian_);
        _grantRole(CONFIGURE_ROLE, guardian_);
        _grantRole(PAUSE_ROLE, guardian_);
        _grantRole(PROCESSOR_ROLE, guardian_);
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
        operationId = _generateOperationId();

        // Transfer USDC from user to contract
        IERC20(USDC).safeTransferFrom(_msgSender(), address(this), usdcAmount);

        // Mint pending ancUSDC to user (convert from USDC decimals to ancUSDC decimals)
        uint256 pendingAncUSDCAmount = _convertDecimals(usdcAmount, USDC, ANC_USDC);
        IAnchoredTokenLike(PENDING_ANC_USDC).mint(_msgSender(), pendingAncUSDCAmount);

        // Store deposit operation
        depositOperations[operationId] = DepositOperation({
            user: _msgSender(),
            usdcAmount: usdcAmount,
            pendingAncUSDCAmount: pendingAncUSDCAmount,
            status: OperationStatus.PENDING,
            timestamp: block.timestamp
        });

        emit PendingDeposit(operationId, _msgSender(), usdcAmount, pendingAncUSDCAmount);
    }

    /**
     * @notice Withdraw ancUSDC and receive pending USDC
     * @param ancUSDCAmount The amount of ancUSDC to withdraw
     * @return operationId The ID of the withdrawal operation
     */
    function withdraw(uint256 ancUSDCAmount)
        external
        nonReentrant
        whenWithdrawalsNotPaused
        returns (bytes32 operationId)
    {
        if (ancUSDCAmount == 0) revert AmountCannotBeZero();
        if (ancUSDCAmount < minimumWithdrawalAmount) revert WithdrawalAmountTooSmall();

        // Generate unique operation ID
        operationId = _generateOperationId();

        // Burn ancUSDC from user
        IERC20(ANC_USDC).safeTransferFrom(_msgSender(), address(this), ancUSDCAmount);
        IAnchoredTokenLike(ANC_USDC).burn(ancUSDCAmount);

        // Mint pending USDC to user (convert from ancUSDC decimals to USDC decimals)
        uint256 pendingUSDCAmount = _convertDecimals(ancUSDCAmount, ANC_USDC, USDC);
        IAnchoredTokenLike(PENDING_USDC).mint(_msgSender(), pendingUSDCAmount);

        // Store withdrawal operation
        withdrawalOperations[operationId] = WithdrawalOperation({
            user: _msgSender(),
            ancUSDCAmount: ancUSDCAmount,
            pendingUSDCAmount: pendingUSDCAmount,
            status: OperationStatus.PENDING,
            timestamp: block.timestamp
        });

        emit PendingWithdraw(operationId, _msgSender(), ancUSDCAmount, pendingUSDCAmount);
    }

    /**
     * @notice Process pending deposit (backend function)
     * @param operationId The ID of the deposit operation to process
     * @param ancUSDCAmount The amount of ancUSDC to mint
     */
    function processDeposit(bytes32 operationId, uint256 ancUSDCAmount) external onlyRole(PROCESSOR_ROLE) nonReentrant {
        DepositOperation storage operation = depositOperations[operationId];

        if (operation.user == address(0)) revert InvalidOperationId();
        if (operation.status != OperationStatus.PENDING) revert OperationAlreadyProcessed();
        // TODO: There should be some logic to transfer USDC to broker

        // Burn pending ancUSDC directly from user's balance
        PendingToken(PENDING_ANC_USDC).burnFrom(operation.user, operation.pendingAncUSDCAmount);

        // Mint ancUSDC to user
        IAnchoredTokenLike(ANC_USDC).mint(operation.user, ancUSDCAmount);

        // Update operation status
        operation.status = OperationStatus.ACTIVE;

        emit DepositProcessed(operationId, operation.user, ancUSDCAmount);
    }

    /**
     * @notice Process pending withdrawal (backend function)
     * @param operationId The ID of the withdrawal operation to process
     * @param usdcAmount The amount of USDC to transfer back
     */
    function processWithdrawal(bytes32 operationId, uint256 usdcAmount) external onlyRole(PROCESSOR_ROLE) nonReentrant {
        WithdrawalOperation storage operation = withdrawalOperations[operationId];

        if (operation.user == address(0)) revert InvalidOperationId();
        if (operation.status != OperationStatus.PENDING) revert OperationAlreadyProcessed();

        // Burn pending USDC directly from user's balance
        PendingToken(PENDING_USDC).burnFrom(operation.user, operation.pendingUSDCAmount);

        // Transfer USDC to user
        IERC20(USDC).safeTransfer(operation.user, usdcAmount);

        // Update operation status
        operation.status = OperationStatus.REDEEMED;

        emit WithdrawalProcessed(operationId, operation.user, usdcAmount);
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
     * @notice Get deposit operation details
     * @param operationId The operation ID
     * @return The deposit operation struct
     */
    function getDepositOperation(bytes32 operationId) external view returns (DepositOperation memory) {
        return depositOperations[operationId];
    }

    /**
     * @notice Get withdrawal operation details
     * @param operationId The operation ID
     * @return The withdrawal operation struct
     */
    function getWithdrawalOperation(bytes32 operationId) external view returns (WithdrawalOperation memory) {
        return withdrawalOperations[operationId];
    }

    /**
     * @notice Generate a unique operation ID
     * @return The generated operation ID
     */
    function _generateOperationId() internal returns (bytes32) {
        return keccak256(abi.encodePacked(block.timestamp, block.number, _msgSender(), ++_operationCounter));
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

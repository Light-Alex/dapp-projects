// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import { Test } from "forge-std/Test.sol";
import { console } from "forge-std/console.sol";
import { Gate } from "../../contracts/gate/Gate.sol";
import { IGate } from "../../contracts/gate/IGate.sol";
import { PendingToken } from "../../contracts/gate/PendingToken.sol";
import { AnchoredToken } from "../../contracts/AnchoredToken.sol";
import { IAnchoredToken } from "../../contracts/interfaces/IAnchoredToken.sol";
import { AnchoredCompliance } from "../../contracts/AnchoredCompliance.sol";
import { AnchoredTokenPauseManager } from "../../contracts/AnchoredTokenPauseManager.sol";
import { TransparentUpgradeableProxy } from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import { ProxyAdmin } from "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MockUSDC
 * @notice Mock USDC token for testing
 */
contract MockUSDC is ERC20 {
    constructor() ERC20("Mock USDC", "USDC") { }

    function decimals() public pure override returns (uint8) {
        return 6;
    }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }

    function burn(address from, uint256 amount) external {
        _burn(from, amount);
    }
}

/**
 * @title GateTest
 * @notice Comprehensive test suite for Gate contract
 */
contract GateTest is Test {
    // ============ Constants ============

    uint256 constant INITIAL_MULTIPLIER = 1e18;
    uint256 constant MIN_DEPOSIT_AMOUNT = 100e6; // 100 USDC
    uint256 constant MIN_WITHDRAWAL_AMOUNT = 100e18; // 100 ancUSDC

    // ============ Test Accounts ============

    address admin = makeAddr("admin");
    address processor = makeAddr("processor");
    address pauser = makeAddr("pauser");
    address user1 = makeAddr("user1");
    address user2 = makeAddr("user2");
    address user3 = makeAddr("user3");

    // ============ Contracts ============

    Gate gate;
    MockUSDC usdc;
    AnchoredToken ancUSDC;
    PendingToken pendingAncUSDC;
    PendingToken pendingUSDC;
    AnchoredCompliance compliance;
    AnchoredTokenPauseManager pauseManager;
    ProxyAdmin proxyAdmin;

    // ============ Events ============

    event DepositInitiated(
        bytes32 indexed operationId, address indexed user, uint256 usdcAmount, uint256 pendingAncUSDCAmount
    );
    event WithdrawalInitiated(
        bytes32 indexed operationId, address indexed user, uint256 ancUSDCAmount, uint256 pendingUSDCAmount
    );
    event DepositProcessed(bytes32 indexed operationId, address indexed user, uint256 ancUSDCAmount);
    event WithdrawalProcessed(bytes32 indexed operationId, address indexed user, uint256 usdcAmount);
    event MinimumDepositAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);
    event MinimumWithdrawalAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);
    event DepositsPaused();
    event DepositsUnpaused();
    event WithdrawalsPaused();
    event WithdrawalsUnpaused();

    // ============ Setup ============

    function setUp() public {
        vm.startPrank(admin);

        // Deploy ProxyAdmin
        proxyAdmin = new ProxyAdmin(admin);

        // Deploy Mock USDC
        usdc = new MockUSDC();

        // Deploy compliance contracts
        AnchoredCompliance complianceImpl = new AnchoredCompliance();
        bytes memory complianceInitData = abi.encodeWithSelector(AnchoredCompliance.initialize.selector, admin);
        TransparentUpgradeableProxy complianceProxy =
            new TransparentUpgradeableProxy(address(complianceImpl), address(proxyAdmin), complianceInitData);
        compliance = AnchoredCompliance(address(complianceProxy));

        AnchoredTokenPauseManager pauseManagerImpl = new AnchoredTokenPauseManager();
        bytes memory pauseManagerInitData = abi.encodeWithSelector(AnchoredTokenPauseManager.initialize.selector, admin);
        TransparentUpgradeableProxy pauseManagerProxy =
            new TransparentUpgradeableProxy(address(pauseManagerImpl), address(proxyAdmin), pauseManagerInitData);
        pauseManager = AnchoredTokenPauseManager(address(pauseManagerProxy));

        // Deploy ancUSDC implementation
        AnchoredToken ancUSDCImpl = new AnchoredToken();

        // Deploy ancUSDC proxy
        bytes memory ancUSDCInitData = abi.encodeWithSelector(
            AnchoredToken.initialize.selector, "Anchored USDC", "ancUSDC", address(compliance), address(pauseManager)
        );

        TransparentUpgradeableProxy ancUSDCProxy =
            new TransparentUpgradeableProxy(address(ancUSDCImpl), address(proxyAdmin), ancUSDCInitData);

        ancUSDC = AnchoredToken(address(ancUSDCProxy));

        // Deploy Gate implementation
        Gate gateImpl = new Gate(address(usdc), address(ancUSDC));

        // Deploy Gate proxy
        bytes memory gateInitData = abi.encodeWithSelector(
            Gate.initialize.selector, address(usdc), address(ancUSDC), admin, MIN_DEPOSIT_AMOUNT, MIN_WITHDRAWAL_AMOUNT
        );

        TransparentUpgradeableProxy gateProxy =
            new TransparentUpgradeableProxy(address(gateImpl), address(proxyAdmin), gateInitData);

        gate = Gate(address(gateProxy));

        // Get pending token addresses
        pendingAncUSDC = PendingToken(gate.PENDING_ANC_USDC());
        pendingUSDC = PendingToken(gate.PENDING_USDC());

        // Grant roles
        gate.grantRole(gate.PROCESSOR_ROLE(), processor);
        gate.grantRole(gate.PAUSE_ROLE(), pauser);

        // Grant mint and burn roles to gate for ancUSDC
        ancUSDC.grantRole(ancUSDC.MINT_ROLE(), address(gate));
        ancUSDC.grantRole(ancUSDC.BURN_ROLE(), address(gate));

        vm.stopPrank();

        // Setup test users with USDC and ancUSDC
        _setupTestUsers();
    }

    function _setupTestUsers() internal {
        // Mint USDC to users
        usdc.mint(user1, 10000e6); // 10,000 USDC
        usdc.mint(user2, 5000e6); // 5,000 USDC
        usdc.mint(user3, 1000e6); // 1,000 USDC

        // Mint ancUSDC to users for withdrawal tests
        vm.startPrank(admin);
        ancUSDC.mint(user1, 5000e18); // 5,000 ancUSDC
        ancUSDC.mint(user2, 2000e18); // 2,000 ancUSDC
        vm.stopPrank();

        // Approve Gate to spend USDC and ancUSDC
        vm.prank(user1);
        usdc.approve(address(gate), type(uint256).max);
        vm.prank(user1);
        ancUSDC.approve(address(gate), type(uint256).max);

        vm.prank(user2);
        usdc.approve(address(gate), type(uint256).max);
        vm.prank(user2);
        ancUSDC.approve(address(gate), type(uint256).max);

        vm.prank(user3);
        usdc.approve(address(gate), type(uint256).max);
    }

    // ============ Deposit Tests ============

    function test_Deposit_Success() public {
        uint256 depositAmount = 1000e6; // 1,000 USDC
        uint256 userUSDCBefore = usdc.balanceOf(user1);

        vm.prank(user1);
        bytes32 operationId = gate.deposit(depositAmount);

        // Check USDC transfer
        assertEq(usdc.balanceOf(user1), userUSDCBefore - depositAmount);
        assertEq(usdc.balanceOf(address(gate)), depositAmount);

        // Check pending ancUSDC minting
        assertEq(pendingAncUSDC.balanceOf(user1), depositAmount * 1e12); // Convert 6 decimals to 18

        // Check operation storage
        IGate.DepositOperation memory operation = gate.getDepositOperation(operationId);
        assertEq(operation.user, user1);
        assertEq(operation.usdcAmount, depositAmount);
        assertEq(operation.pendingAncUSDCAmount, depositAmount * 1e12);
        assertEq(uint256(operation.status), uint256(IGate.OperationStatus.PENDING));
        assertGt(operation.timestamp, 0);
    }

    function test_Deposit_RevertWhenBelowMinimum() public {
        uint256 depositAmount = 50e6; // 50 USDC (below minimum)

        vm.prank(user1);
        vm.expectRevert(IGate.DepositAmountTooSmall.selector);
        gate.deposit(depositAmount);
    }

    function test_Deposit_RevertWhenPaused() public {
        vm.prank(pauser);
        gate.pauseDeposits();

        uint256 depositAmount = 1000e6;

        vm.prank(user1);
        vm.expectRevert(IGate.DepositsArePaused.selector);
        gate.deposit(depositAmount);
    }

    function test_Deposit_RevertWhenInsufficientBalance() public {
        uint256 depositAmount = 20000e6; // More than user1's balance

        vm.prank(user1);
        vm.expectRevert();
        gate.deposit(depositAmount);
    }

    // ============ Withdrawal Tests ============

    function test_Withdraw_Success() public {
        uint256 withdrawAmount = 1000e18; // 1,000 ancUSDC
        uint256 userAncUSDCBefore = ancUSDC.balanceOf(user1);

        vm.prank(user1);
        bytes32 operationId = gate.withdraw(withdrawAmount);

        // Check ancUSDC burn
        assertEq(ancUSDC.balanceOf(user1), userAncUSDCBefore - withdrawAmount);

        // Check pending USDC minting
        assertEq(pendingUSDC.balanceOf(user1), withdrawAmount / 1e12); // Convert 18 decimals to 6

        // Check operation storage
        IGate.WithdrawalOperation memory operation = gate.getWithdrawalOperation(operationId);
        assertEq(operation.user, user1);
        assertEq(operation.ancUSDCAmount, withdrawAmount);
        assertEq(operation.pendingUSDCAmount, withdrawAmount / 1e12);
        assertEq(uint256(operation.status), uint256(IGate.OperationStatus.PENDING));
        assertGt(operation.timestamp, 0);
    }

    function test_Withdraw_RevertWhenBelowMinimum() public {
        uint256 withdrawAmount = 50e18; // 50 ancUSDC (below minimum)

        vm.prank(user1);
        vm.expectRevert(IGate.WithdrawalAmountTooSmall.selector);
        gate.withdraw(withdrawAmount);
    }

    function test_Withdraw_RevertWhenPaused() public {
        vm.prank(pauser);
        gate.pauseWithdrawals();

        uint256 withdrawAmount = 1000e18;

        vm.prank(user1);
        vm.expectRevert(IGate.WithdrawalsArePaused.selector);
        gate.withdraw(withdrawAmount);
    }

    function test_Withdraw_RevertWhenInsufficientBalance() public {
        uint256 withdrawAmount = 10000e18; // More than user1's balance

        vm.prank(user1);
        vm.expectRevert();
        gate.withdraw(withdrawAmount);
    }

    // ============ Process Deposit Tests ============

    function test_ProcessDeposit_Success() public {
        // First create a deposit
        uint256 depositAmount = 1000e6;
        vm.prank(user1);
        bytes32 operationId = gate.deposit(depositAmount);

        uint256 ancUSDCAmount = 950e18; // Slightly less due to processing
        uint256 userAncUSDCBefore = ancUSDC.balanceOf(user1);
        uint256 userPendingBefore = pendingAncUSDC.balanceOf(user1);

        vm.prank(processor);
        gate.processDeposit(operationId, ancUSDCAmount);

        // Check pending ancUSDC burned
        assertEq(pendingAncUSDC.balanceOf(user1), 0);

        // Check ancUSDC minted
        assertEq(ancUSDC.balanceOf(user1), userAncUSDCBefore + ancUSDCAmount);

        // Check operation status updated
        IGate.DepositOperation memory operation = gate.getDepositOperation(operationId);
        assertEq(uint256(operation.status), uint256(IGate.OperationStatus.ACTIVE));
    }

    function test_ProcessDeposit_RevertWhenNotProcessor() public {
        uint256 depositAmount = 1000e6;
        vm.prank(user1);
        bytes32 operationId = gate.deposit(depositAmount);

        vm.prank(user1);
        vm.expectRevert();
        gate.processDeposit(operationId, 950e18);
    }

    function test_ProcessDeposit_RevertWhenInvalidOperationId() public {
        bytes32 invalidOperationId = keccak256("invalid");

        vm.prank(processor);
        vm.expectRevert(IGate.InvalidOperationId.selector);
        gate.processDeposit(invalidOperationId, 950e18);
    }

    function test_ProcessDeposit_RevertWhenAlreadyProcessed() public {
        uint256 depositAmount = 1000e6;
        vm.prank(user1);
        bytes32 operationId = gate.deposit(depositAmount);

        // Process once
        vm.prank(processor);
        gate.processDeposit(operationId, 950e18);

        // Try to process again
        vm.prank(processor);
        vm.expectRevert(IGate.OperationAlreadyProcessed.selector);
        gate.processDeposit(operationId, 950e18);
    }

    // ============ Process Withdrawal Tests ============

    function test_ProcessWithdrawal_Success() public {
        // First create a withdrawal
        uint256 withdrawAmount = 1000e18;
        vm.prank(user1);
        bytes32 operationId = gate.withdraw(withdrawAmount);

        uint256 usdcAmount = 950e6; // Slightly less due to processing
        uint256 userUSDCBefore = usdc.balanceOf(user1);

        // Gate needs USDC to transfer
        usdc.mint(address(gate), usdcAmount);

        vm.prank(processor);
        gate.processWithdrawal(operationId, usdcAmount);

        // Check pending USDC burned
        assertEq(pendingUSDC.balanceOf(user1), 0);

        // Check USDC transferred
        assertEq(usdc.balanceOf(user1), userUSDCBefore + usdcAmount);

        // Check operation status updated
        IGate.WithdrawalOperation memory operation = gate.getWithdrawalOperation(operationId);
        assertEq(uint256(operation.status), uint256(IGate.OperationStatus.REDEEMED));
    }

    function test_ProcessWithdrawal_RevertWhenNotProcessor() public {
        uint256 withdrawAmount = 1000e18;
        vm.prank(user1);
        bytes32 operationId = gate.withdraw(withdrawAmount);

        vm.prank(user1);
        vm.expectRevert();
        gate.processWithdrawal(operationId, 950e6);
    }

    function test_ProcessWithdrawal_RevertWhenInvalidOperationId() public {
        bytes32 invalidOperationId = keccak256("invalid");

        vm.prank(processor);
        vm.expectRevert(IGate.InvalidOperationId.selector);
        gate.processWithdrawal(invalidOperationId, 950e18);
    }

    function test_ProcessWithdrawal_RevertWhenAlreadyProcessed() public {
        uint256 withdrawAmount = 1000e18;
        vm.prank(user1);
        bytes32 operationId = gate.withdraw(withdrawAmount);

        usdc.mint(address(gate), 950e6);

        // Process once
        vm.prank(processor);
        gate.processWithdrawal(operationId, 950e6);

        // Try to process again
        vm.prank(processor);
        vm.expectRevert(IGate.OperationAlreadyProcessed.selector);
        gate.processWithdrawal(operationId, 950e6);
    }

    // ============ Configuration Tests ============

    function test_SetMinimumDepositAmount_Success() public {
        uint256 newAmount = 200e6;

        vm.prank(admin);
        vm.expectEmit(true, true, true, true);
        emit MinimumDepositAmountSet(MIN_DEPOSIT_AMOUNT, newAmount);
        gate.setMinimumDepositAmount(newAmount);

        assertEq(gate.minimumDepositAmount(), newAmount);
    }

    function test_SetMinimumDepositAmount_RevertWhenNotAdmin() public {
        vm.prank(user1);
        vm.expectRevert();
        gate.setMinimumDepositAmount(200e6);
    }

    function test_SetMinimumWithdrawalAmount_Success() public {
        uint256 newAmount = 200e18;

        vm.prank(admin);
        vm.expectEmit(true, true, true, true);
        emit MinimumWithdrawalAmountSet(MIN_WITHDRAWAL_AMOUNT, newAmount);
        gate.setMinimumWithdrawalAmount(newAmount);

        assertEq(gate.minimumWithdrawalAmount(), newAmount);
    }

    function test_SetMinimumWithdrawalAmount_RevertWhenNotAdmin() public {
        vm.prank(user1);
        vm.expectRevert();
        gate.setMinimumWithdrawalAmount(200e18);
    }

    // ============ Pause/Unpause Tests ============

    function test_PauseDeposits_Success() public {
        vm.prank(pauser);
        vm.expectEmit(true, true, true, true);
        emit DepositsPaused();
        gate.pauseDeposits();

        assertTrue(gate.depositsArePaused());
    }

    function test_PauseDeposits_RevertWhenNotPauser() public {
        vm.prank(user1);
        vm.expectRevert();
        gate.pauseDeposits();
    }

    function test_UnpauseDeposits_Success() public {
        vm.prank(pauser);
        gate.pauseDeposits();

        vm.prank(pauser);
        vm.expectEmit(true, true, true, true);
        emit DepositsUnpaused();
        gate.unpauseDeposits();

        assertFalse(gate.depositsArePaused());
    }

    function test_PauseWithdrawals_Success() public {
        vm.prank(pauser);
        vm.expectEmit(true, true, true, true);
        emit WithdrawalsPaused();
        gate.pauseWithdrawals();

        assertTrue(gate.withdrawalsArePaused());
    }

    function test_UnpauseWithdrawals_Success() public {
        vm.prank(pauser);
        gate.pauseWithdrawals();

        vm.prank(pauser);
        vm.expectEmit(true, true, true, true);
        emit WithdrawalsUnpaused();
        gate.unpauseWithdrawals();

        assertFalse(gate.withdrawalsArePaused());
    }

    // ============ View Function Tests ============

    function test_GetDepositOperation() public {
        uint256 depositAmount = 1000e6;
        vm.prank(user1);
        bytes32 operationId = gate.deposit(depositAmount);

        IGate.DepositOperation memory operation = gate.getDepositOperation(operationId);
        assertEq(operation.user, user1);
        assertEq(operation.usdcAmount, depositAmount);
        assertEq(operation.pendingAncUSDCAmount, depositAmount * 1e12);
        assertEq(uint256(operation.status), uint256(IGate.OperationStatus.PENDING));
    }

    function test_GetWithdrawalOperation() public {
        uint256 withdrawAmount = 1000e18;
        vm.prank(user1);
        bytes32 operationId = gate.withdraw(withdrawAmount);

        IGate.WithdrawalOperation memory operation = gate.getWithdrawalOperation(operationId);
        assertEq(operation.user, user1);
        assertEq(operation.ancUSDCAmount, withdrawAmount);
        assertEq(operation.pendingUSDCAmount, withdrawAmount / 1e12);
        assertEq(uint256(operation.status), uint256(IGate.OperationStatus.PENDING));
    }

    // ============ Integration Tests ============

    function test_FullDepositFlow() public {
        uint256 depositAmount = 1000e6;
        uint256 ancUSDCAmount = 950e18;

        // 1. User deposits USDC
        vm.prank(user1);
        bytes32 operationId = gate.deposit(depositAmount);

        // 2. Check initial state
        assertEq(usdc.balanceOf(address(gate)), depositAmount);
        assertEq(pendingAncUSDC.balanceOf(user1), depositAmount * 1e12);

        // 3. Backend processes deposit
        vm.prank(processor);
        gate.processDeposit(operationId, ancUSDCAmount);

        // 4. Check final state
        assertEq(pendingAncUSDC.balanceOf(user1), 0);
        assertEq(ancUSDC.balanceOf(user1), 5000e18 + ancUSDCAmount); // Initial + new

        IGate.DepositOperation memory operation = gate.getDepositOperation(operationId);
        assertEq(uint256(operation.status), uint256(IGate.OperationStatus.ACTIVE));
    }

    function test_FullWithdrawalFlow() public {
        uint256 withdrawAmount = 1000e18;
        uint256 usdcAmount = 950e6;
        uint256 initialAncUSDC = ancUSDC.balanceOf(user1);
        uint256 initialUSDC = usdc.balanceOf(user1);

        // 1. User withdraws ancUSDC
        vm.prank(user1);
        bytes32 operationId = gate.withdraw(withdrawAmount);

        // 2. Check initial state
        assertEq(ancUSDC.balanceOf(user1), initialAncUSDC - withdrawAmount);
        assertEq(pendingUSDC.balanceOf(user1), withdrawAmount / 1e12);

        // 3. Gate needs USDC for processing
        usdc.mint(address(gate), usdcAmount);

        // 4. Backend processes withdrawal
        vm.prank(processor);
        gate.processWithdrawal(operationId, usdcAmount);

        // 5. Check final state
        assertEq(pendingUSDC.balanceOf(user1), 0);
        assertEq(usdc.balanceOf(user1), initialUSDC + usdcAmount);

        IGate.WithdrawalOperation memory operation = gate.getWithdrawalOperation(operationId);
        assertEq(uint256(operation.status), uint256(IGate.OperationStatus.REDEEMED));
    }

    function test_MultipleUsersOperations() public {
        // User1 deposits
        vm.prank(user1);
        bytes32 depositId1 = gate.deposit(1000e6);

        // User2 deposits
        vm.prank(user2);
        bytes32 depositId2 = gate.deposit(500e6);

        // User1 withdraws
        vm.prank(user1);
        bytes32 withdrawId1 = gate.withdraw(2000e18);

        // Check all operations are independent
        IGate.DepositOperation memory deposit1 = gate.getDepositOperation(depositId1);
        IGate.DepositOperation memory deposit2 = gate.getDepositOperation(depositId2);
        IGate.WithdrawalOperation memory withdraw1 = gate.getWithdrawalOperation(withdrawId1);

        assertEq(deposit1.user, user1);
        assertEq(deposit2.user, user2);
        assertEq(withdraw1.user, user1);

        assertEq(deposit1.usdcAmount, 1000e6);
        assertEq(deposit2.usdcAmount, 500e6);
        assertEq(withdraw1.ancUSDCAmount, 2000e18);
    }

    // ============ Edge Cases ============

    function test_DepositWithExactMinimumAmount() public {
        vm.prank(user1);
        bytes32 operationId = gate.deposit(MIN_DEPOSIT_AMOUNT);

        IGate.DepositOperation memory operation = gate.getDepositOperation(operationId);
        assertEq(operation.usdcAmount, MIN_DEPOSIT_AMOUNT);
    }

    function test_WithdrawWithExactMinimumAmount() public {
        vm.prank(user1);
        bytes32 operationId = gate.withdraw(MIN_WITHDRAWAL_AMOUNT);

        IGate.WithdrawalOperation memory operation = gate.getWithdrawalOperation(operationId);
        assertEq(operation.ancUSDCAmount, MIN_WITHDRAWAL_AMOUNT);
    }

    function test_OperationIdUniqueness() public {
        vm.prank(user1);
        bytes32 id1 = gate.deposit(1000e6);

        vm.prank(user1);
        bytes32 id2 = gate.deposit(1000e6);

        assertTrue(id1 != id2);
    }
}

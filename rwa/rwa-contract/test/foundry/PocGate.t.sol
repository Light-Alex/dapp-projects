// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import { PocGate } from "contracts/poc/PocGate.sol";
import { PocToken } from "contracts/poc/PocToken.sol";
import { MockUSDC } from "contracts/poc/MockUSDC.sol";
import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract PocGateTest is Test {
    // ============ Constants ============

    uint256 constant MIN_DEPOSIT_AMOUNT = 100e6; // 100 USDC (6 decimals)
    uint256 constant MIN_WITHDRAWAL_AMOUNT = 100e18; // 100 USDM (18 decimals)

    // ============ Test Accounts ============

    address internal admin = makeAddr("admin");
    address internal user1 = makeAddr("user1");
    address internal user2 = makeAddr("user2");
    address internal unauthorized = makeAddr("unauthorized");

    // ============ Contracts ============

    PocGate internal gate;
    PocToken internal usdm;
    MockUSDC internal usdc;

    // ============ Events (mirror PocGate) ============

    event MinimumDepositAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);
    event MinimumWithdrawalAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);
    event DepositsPaused();
    event DepositsUnpaused();
    event WithdrawalsPaused();
    event WithdrawalsUnpaused();

    // ============ Setup ============

    function setUp() public {
        // Deploy MockUSDC (no proxy needed)
        usdc = new MockUSDC();

        // Deploy PocToken implementation with a dummy gate address, then proxy
        PocToken usdmImpl = new PocToken(address(0));
        ERC1967Proxy usdmProxy =
            new ERC1967Proxy(address(usdmImpl), abi.encodeWithSelector(PocToken.initialize.selector, "USDM", "USDM"));
        usdm = PocToken(address(usdmProxy));

        // Deploy PocGate implementation (immutables set in constructor)
        PocGate gateImpl = new PocGate(address(usdc), address(usdm));

        // Deploy PocGate proxy and initialize
        ERC1967Proxy gateProxy = new ERC1967Proxy(
            address(gateImpl),
            abi.encodeWithSelector(PocGate.initialize.selector, admin, MIN_DEPOSIT_AMOUNT, MIN_WITHDRAWAL_AMOUNT)
        );
        gate = PocGate(address(gateProxy));

        // Grant MINTER_ROLE and BURNER_ROLE on USDM to the gate contract
        // (this test contract is the deployer of usdmProxy, so it holds DEFAULT_ADMIN_ROLE)
        usdm.setOperator(address(gate));

        // Setup users with USDC
        usdc.mint(user1, 10_000e6);
        usdc.mint(user2, 5_000e6);

        // Users approve gate to spend USDC
        vm.prank(user1);
        usdc.approve(address(gate), type(uint256).max);
        vm.prank(user2);
        usdc.approve(address(gate), type(uint256).max);
    }

    // Helper: deposit for a user and then approve gate to spend their USDM (for withdraw tests)
    function _depositAndApproveUSDM(address user, uint256 usdcAmount) internal {
        vm.prank(user);
        gate.deposit(usdcAmount);
        vm.prank(user);
        usdm.approve(address(gate), type(uint256).max);
    }

    // ================================================================
    //                      INITIALIZATION TESTS
    // ================================================================

    function test_Initialize_Success() public view {
        // Guardian should have all three roles
        assertTrue(gate.hasRole(gate.DEFAULT_ADMIN_ROLE(), admin), "admin should have DEFAULT_ADMIN_ROLE");
        assertTrue(gate.hasRole(gate.CONFIGURE_ROLE(), admin), "admin should have CONFIGURE_ROLE");
        assertTrue(gate.hasRole(gate.PAUSE_ROLE(), admin), "admin should have PAUSE_ROLE");

        // Minimum amounts
        assertEq(gate.minimumDepositAmount(), MIN_DEPOSIT_AMOUNT, "minimumDepositAmount");
        assertEq(gate.minimumWithdrawalAmount(), MIN_WITHDRAWAL_AMOUNT, "minimumWithdrawalAmount");

        // Immutables
        assertEq(gate.USDC(), address(usdc), "USDC address");
        assertEq(gate.USDM(), address(usdm), "USDM address");

        // Not paused by default
        assertFalse(gate.depositsArePaused(), "deposits should not be paused");
        assertFalse(gate.withdrawalsArePaused(), "withdrawals should not be paused");
    }

    function test_Initialize_RevertWhenGuardianIsZero() public {
        PocGate gateImpl = new PocGate(address(usdc), address(usdm));

        vm.expectRevert(PocGate.AddressCannotBeZero.selector);
        new ERC1967Proxy(
            address(gateImpl),
            abi.encodeWithSelector(PocGate.initialize.selector, address(0), MIN_DEPOSIT_AMOUNT, MIN_WITHDRAWAL_AMOUNT)
        );
    }

    function test_Initialize_RevertWhenCalledTwice() public {
        vm.expectRevert();
        gate.initialize(admin, MIN_DEPOSIT_AMOUNT, MIN_WITHDRAWAL_AMOUNT);
    }

    // ================================================================
    //                         DEPOSIT TESTS
    // ================================================================

    function test_Deposit_Success() public {
        uint256 depositAmount = 1000e6;
        uint256 userUSDCBefore = usdc.balanceOf(user1);
        uint256 expectedUSDM = 1000e18; // 1000e6 * 1e12

        vm.prank(user1);
        gate.deposit(depositAmount);

        // USDC transferred from user to gate
        assertEq(usdc.balanceOf(user1), userUSDCBefore - depositAmount, "user USDC balance decreased");
        assertEq(usdc.balanceOf(address(gate)), depositAmount, "gate USDC balance increased");

        // USDM minted to user (6 -> 18 decimals)
        assertEq(usdm.balanceOf(user1), expectedUSDM, "user USDM balance");
    }

    function test_Deposit_RevertWhenAmountIsZero() public {
        vm.prank(user1);
        vm.expectRevert(PocGate.AmountCannotBeZero.selector);
        gate.deposit(0);
    }

    function test_Deposit_RevertWhenBelowMinimum() public {
        uint256 belowMinimum = MIN_DEPOSIT_AMOUNT - 1;

        vm.prank(user1);
        vm.expectRevert(PocGate.DepositAmountTooSmall.selector);
        gate.deposit(belowMinimum);
    }

    function test_Deposit_RevertWhenPaused() public {
        vm.prank(admin);
        gate.pauseDeposits();

        vm.prank(user1);
        vm.expectRevert(PocGate.DepositsArePaused.selector);
        gate.deposit(1000e6);
    }

    function test_Deposit_RevertWhenInsufficientBalance() public {
        // user1 has 10_000e6, try to deposit more
        vm.prank(user1);
        vm.expectRevert();
        gate.deposit(20_000e6);
    }

    function test_Deposit_RevertWhenNotApproved() public {
        address noApproval = makeAddr("noApproval");
        usdc.mint(noApproval, 10_000e6);
        // No approval given

        vm.prank(noApproval);
        vm.expectRevert();
        gate.deposit(1000e6);
    }

    function test_Deposit_DecimalConversion_1000USDC() public {
        uint256 usdcAmount = 1000e6;
        uint256 expectedUSDM = 1000e18;

        vm.prank(user1);
        gate.deposit(usdcAmount);

        assertEq(usdm.balanceOf(user1), expectedUSDM, "1000e6 USDC -> 1000e18 USDM");
    }

    function test_Deposit_DecimalConversion_1WeiUSDC() public {
        // Set minimumDepositAmount to 0 so we can deposit 1 wei
        vm.prank(admin);
        gate.setMinimumDepositAmount(0);

        uint256 usdcAmount = 1; // 1 wei USDC (smallest unit)
        uint256 expectedUSDM = 1e12; // 1 * 10^(18-6) = 1e12

        vm.prank(user1);
        gate.deposit(usdcAmount);

        assertEq(usdm.balanceOf(user1), expectedUSDM, "1 wei USDC -> 1e12 USDM");
    }

    // ================================================================
    //                        WITHDRAW TESTS
    // ================================================================

    function test_Withdraw_Success() public {
        // First deposit to get USDM and fund the gate with USDC
        _depositAndApproveUSDM(user1, 1000e6);

        uint256 withdrawUSDM = 500e18;
        uint256 expectedUSDC = 500e6; // 500e18 / 1e12

        uint256 userUSDMBefore = usdm.balanceOf(user1);
        uint256 userUSDCBefore = usdc.balanceOf(user1);

        vm.prank(user1);
        gate.withdraw(withdrawUSDM);

        // USDM burned
        assertEq(usdm.balanceOf(user1), userUSDMBefore - withdrawUSDM, "user USDM balance decreased");

        // USDC transferred to user
        assertEq(usdc.balanceOf(user1), userUSDCBefore + expectedUSDC, "user USDC balance increased");
    }

    function test_Withdraw_RevertWhenAmountIsZero() public {
        vm.prank(user1);
        vm.expectRevert(PocGate.AmountCannotBeZero.selector);
        gate.withdraw(0);
    }

    function test_Withdraw_RevertWhenBelowMinimum() public {
        uint256 belowMinimum = MIN_WITHDRAWAL_AMOUNT - 1;

        vm.prank(user1);
        vm.expectRevert(PocGate.WithdrawalAmountTooSmall.selector);
        gate.withdraw(belowMinimum);
    }

    function test_Withdraw_RevertWhenPaused() public {
        vm.prank(admin);
        gate.pauseWithdrawals();

        vm.prank(user1);
        vm.expectRevert(PocGate.WithdrawalsArePaused.selector);
        gate.withdraw(1000e18);
    }

    function test_Withdraw_DecimalConversion_1000USDM() public {
        _depositAndApproveUSDM(user1, 2000e6);

        uint256 usdmAmount = 1000e18;
        uint256 expectedUSDC = 1000e6;

        uint256 userUSDCBefore = usdc.balanceOf(user1);

        vm.prank(user1);
        gate.withdraw(usdmAmount);

        assertEq(usdc.balanceOf(user1), userUSDCBefore + expectedUSDC, "1000e18 USDM -> 1000e6 USDC");
    }

    function test_Withdraw_DecimalTruncation_SmallAmount() public {
        // Deposit first, then try withdrawing a tiny amount that truncates to 0 USDC
        _depositAndApproveUSDM(user1, 1000e6);

        // Set minimum withdrawal to 0 so we can test small amounts
        vm.prank(admin);
        gate.setMinimumWithdrawalAmount(0);

        uint256 tinyUSDM = 1e12 - 1; // just below 1e12 -> truncates to 0 USDC
        uint256 expectedUSDC = 0; // integer division: (1e12 - 1) / 1e12 = 0

        uint256 userUSDCBefore = usdc.balanceOf(user1);

        vm.prank(user1);
        gate.withdraw(tinyUSDM);

        assertEq(usdc.balanceOf(user1), userUSDCBefore + expectedUSDC, "sub-1e12 USDM -> 0 USDC");
    }

    function test_Withdraw_RevertWhenGateHasInsufficientUSDC() public {
        // Mint USDM directly to user1 (bypassing deposit, so gate has no USDC)
        usdm.mint(user1, 5000e18);
        vm.prank(user1);
        usdm.approve(address(gate), type(uint256).max);

        // Set minimum to 0 for this test
        vm.prank(admin);
        gate.setMinimumWithdrawalAmount(100e18);

        vm.prank(user1);
        vm.expectRevert(); // SafeERC20 will revert due to insufficient USDC in gate
        gate.withdraw(1000e18);
    }

    // ================================================================
    //                    PAUSE / UNPAUSE TESTS
    // ================================================================

    function test_PauseDeposits_Success() public {
        vm.prank(admin);
        vm.expectEmit(true, true, true, true);
        emit DepositsPaused();
        gate.pauseDeposits();

        assertTrue(gate.depositsArePaused(), "deposits should be paused");
    }

    function test_UnpauseDeposits_Success() public {
        vm.prank(admin);
        gate.pauseDeposits();

        vm.prank(admin);
        vm.expectEmit(true, true, true, true);
        emit DepositsUnpaused();
        gate.unpauseDeposits();

        assertFalse(gate.depositsArePaused(), "deposits should not be paused");
    }

    function test_PauseDeposits_RevertWhenNotPauseRole() public {
        vm.prank(unauthorized);
        vm.expectRevert();
        gate.pauseDeposits();
    }

    function test_UnpauseDeposits_RevertWhenNotPauseRole() public {
        vm.prank(admin);
        gate.pauseDeposits();

        vm.prank(unauthorized);
        vm.expectRevert();
        gate.unpauseDeposits();
    }

    function test_PauseWithdrawals_Success() public {
        vm.prank(admin);
        vm.expectEmit(true, true, true, true);
        emit WithdrawalsPaused();
        gate.pauseWithdrawals();

        assertTrue(gate.withdrawalsArePaused(), "withdrawals should be paused");
    }

    function test_UnpauseWithdrawals_Success() public {
        vm.prank(admin);
        gate.pauseWithdrawals();

        vm.prank(admin);
        vm.expectEmit(true, true, true, true);
        emit WithdrawalsUnpaused();
        gate.unpauseWithdrawals();

        assertFalse(gate.withdrawalsArePaused(), "withdrawals should not be paused");
    }

    function test_PauseWithdrawals_RevertWhenNotPauseRole() public {
        vm.prank(unauthorized);
        vm.expectRevert();
        gate.pauseWithdrawals();
    }

    function test_UnpauseWithdrawals_RevertWhenNotPauseRole() public {
        vm.prank(admin);
        gate.pauseWithdrawals();

        vm.prank(unauthorized);
        vm.expectRevert();
        gate.unpauseWithdrawals();
    }

    // ================================================================
    //                     CONFIGURATION TESTS
    // ================================================================

    function test_SetMinimumDepositAmount_Success() public {
        uint256 newAmount = 200e6;

        vm.prank(admin);
        vm.expectEmit(true, true, true, true);
        emit MinimumDepositAmountSet(MIN_DEPOSIT_AMOUNT, newAmount);
        gate.setMinimumDepositAmount(newAmount);

        assertEq(gate.minimumDepositAmount(), newAmount, "minimumDepositAmount updated");
    }

    function test_SetMinimumDepositAmount_RevertWhenNotConfigureRole() public {
        vm.prank(unauthorized);
        vm.expectRevert();
        gate.setMinimumDepositAmount(200e6);
    }

    function test_SetMinimumWithdrawalAmount_Success() public {
        uint256 newAmount = 200e18;

        vm.prank(admin);
        vm.expectEmit(true, true, true, true);
        emit MinimumWithdrawalAmountSet(MIN_WITHDRAWAL_AMOUNT, newAmount);
        gate.setMinimumWithdrawalAmount(newAmount);

        assertEq(gate.minimumWithdrawalAmount(), newAmount, "minimumWithdrawalAmount updated");
    }

    function test_SetMinimumWithdrawalAmount_RevertWhenNotConfigureRole() public {
        vm.prank(unauthorized);
        vm.expectRevert();
        gate.setMinimumWithdrawalAmount(200e18);
    }

    // ================================================================
    //                      INTEGRATION TESTS
    // ================================================================

    function test_FullDepositWithdrawFlow() public {
        uint256 depositUSDC = 1000e6;
        uint256 expectedUSDM = 1000e18;

        uint256 initialUSDC = usdc.balanceOf(user1);

        // 1. Deposit
        vm.prank(user1);
        gate.deposit(depositUSDC);

        assertEq(usdc.balanceOf(user1), initialUSDC - depositUSDC, "USDC deducted after deposit");
        assertEq(usdm.balanceOf(user1), expectedUSDM, "USDM minted after deposit");

        // 2. Approve USDM for withdrawal
        vm.prank(user1);
        usdm.approve(address(gate), type(uint256).max);

        // 3. Withdraw all USDM
        vm.prank(user1);
        gate.withdraw(expectedUSDM);

        // 4. Verify final balances
        assertEq(usdm.balanceOf(user1), 0, "USDM fully burned after withdraw");
        assertEq(usdc.balanceOf(user1), initialUSDC, "USDC fully recovered after withdraw");
        assertEq(usdc.balanceOf(address(gate)), 0, "gate USDC balance is zero");
    }

    function test_MultiUserConcurrentDepositWithdraw() public {
        // User1 deposits 2000 USDC
        vm.prank(user1);
        gate.deposit(2000e6);

        // User2 deposits 1000 USDC
        vm.prank(user2);
        gate.deposit(1000e6);

        // Verify intermediate state
        assertEq(usdm.balanceOf(user1), 2000e18, "user1 USDM after deposit");
        assertEq(usdm.balanceOf(user2), 1000e18, "user2 USDM after deposit");
        assertEq(usdc.balanceOf(address(gate)), 3000e6, "gate holds total USDC");

        // User1 withdraws 1000 USDM
        vm.prank(user1);
        usdm.approve(address(gate), type(uint256).max);
        vm.prank(user1);
        gate.withdraw(1000e18);

        // User2 withdraws 500 USDM
        vm.prank(user2);
        usdm.approve(address(gate), type(uint256).max);
        vm.prank(user2);
        gate.withdraw(500e18);

        // Verify final state
        assertEq(usdm.balanceOf(user1), 1000e18, "user1 remaining USDM");
        assertEq(usdm.balanceOf(user2), 500e18, "user2 remaining USDM");
        assertEq(usdc.balanceOf(address(gate)), 1500e6, "gate remaining USDC");
    }
}

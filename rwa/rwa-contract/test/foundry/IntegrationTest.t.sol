// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import { Test } from "forge-std/Test.sol";
import { AnchoredToken } from "../../contracts/AnchoredToken.sol";
import { AnchoredTokenManager } from "../../contracts/AnchoredTokenManager.sol";
import { AnchoredTokenFactory } from "../../contracts/AnchoredTokenFactory.sol";
import { AnchoredCompliance } from "../../contracts/AnchoredCompliance.sol";
import { AnchoredBlocklist } from "../../contracts/AnchoredBlocklist.sol";
import { AnchoredSanctionsList } from "../../contracts/AnchoredSanctionsList.sol";
import { IAnchoredBlocklist } from "../../contracts/interfaces/IAnchoredBlocklist.sol";
import { IAnchoredSanctionsList } from "../../contracts/interfaces/IAnchoredSanctionsList.sol";
import { AnchoredTokenManagerRegistrar } from "../../contracts/AnchoredTokenManagerRegistrar.sol";
import { AnchoredTokenPauseManager } from "../../contracts/AnchoredTokenPauseManager.sol";
import { IAnchoredTokenManager } from "../../contracts/interfaces/IAnchoredTokenManager.sol";
import { TransparentUpgradeableProxy } from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import { ProxyAdmin } from "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import { UpgradeableBeacon } from "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";

// Mock USDC token for testing
contract MockERC20 is Test {
    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    string public name = "Mock USDC";
    string public symbol = "USDC";
    uint8 public decimals = 6;
    uint256 public totalSupply;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    function mint(address to, uint256 amount) external {
        balanceOf[to] += amount;
        totalSupply += amount;
        emit Transfer(address(0), to, amount);
    }

    function transfer(address to, uint256 amount) external returns (bool) {
        require(balanceOf[msg.sender] >= amount, "Insufficient balance");
        balanceOf[msg.sender] -= amount;
        balanceOf[to] += amount;
        emit Transfer(msg.sender, to, amount);
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) external returns (bool) {
        require(balanceOf[from] >= amount, "Insufficient balance");
        require(allowance[from][msg.sender] >= amount, "Insufficient allowance");

        balanceOf[from] -= amount;
        balanceOf[to] += amount;
        allowance[from][msg.sender] -= amount;

        emit Transfer(from, to, amount);
        return true;
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        allowance[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }
}

contract IntegrationTest is Test {
    // Core contracts
    AnchoredTokenFactory public tokenFactory;
    AnchoredTokenManager public tokenManager;
    AnchoredTokenManagerRegistrar public registrar;
    AnchoredTokenPauseManager public pauseManager;

    // Compliance contracts
    AnchoredCompliance public anchoredCompliance;
    AnchoredBlocklist public blocklist;
    AnchoredSanctionsList public sanctionsList;

    // Proxy admins
    ProxyAdmin public proxyAdmin;

    // Mock tokens
    MockERC20 public usdcToken;
    AnchoredToken public rwaToken1;
    AnchoredToken public rwaToken2;

    // Test addresses
    address public admin = makeAddr("admin");
    address public configurer = makeAddr("configurer");
    address public pauser = makeAddr("pauser");
    address public alice = makeAddr("alice");
    address public bob = makeAddr("bob");
    address public charlie = makeAddr("charlie");
    address public maliciousUser = makeAddr("maliciousUser");
    address public institutionalUser = makeAddr("institutionalUser");

    // Price signer for attestations
    uint256 private priceSignerPrivateKey = 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef;
    address public priceSigner = vm.addr(priceSignerPrivateKey);

    // Helper function to generate valid signatures
    function _generateSignature(IAnchoredTokenManager.Quote memory quote) internal view returns (bytes memory) {
        bytes32 messageHash = keccak256(
            abi.encode(
                quote.asset, quote.side, quote.quantity, quote.price, quote.nonce, quote.expiry, quote.attestationId
            )
        );
        bytes32 ethSignedMessageHash = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", messageHash));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(priceSignerPrivateKey, ethSignedMessageHash);
        return abi.encodePacked(r, s, v);
    }

    function setUp() public {
        // Deploy mock USDC
        usdcToken = new MockERC20();

        // Deploy proxy admin
        proxyAdmin = new ProxyAdmin(admin);

        // Deploy compliance contracts using TUP pattern
        vm.startPrank(admin);

        // Deploy Blocklist implementation and proxy
        AnchoredBlocklist blocklistImpl = new AnchoredBlocklist();
        TransparentUpgradeableProxy blocklistProxy = new TransparentUpgradeableProxy(
            address(blocklistImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(AnchoredBlocklist.initialize.selector, admin)
        );
        blocklist = AnchoredBlocklist(address(blocklistProxy));

        // Deploy SanctionsList implementation and proxy
        AnchoredSanctionsList sanctionsListImpl = new AnchoredSanctionsList();
        TransparentUpgradeableProxy sanctionsListProxy = new TransparentUpgradeableProxy(
            address(sanctionsListImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(AnchoredSanctionsList.initialize.selector, admin)
        );
        sanctionsList = AnchoredSanctionsList(address(sanctionsListProxy));

        // Deploy AnchoredCompliance implementation and proxy
        AnchoredCompliance complianceImpl = new AnchoredCompliance(); // Use zero address for implementation
        TransparentUpgradeableProxy complianceProxy = new TransparentUpgradeableProxy(
            address(complianceImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(AnchoredCompliance.initialize.selector, admin)
        );
        anchoredCompliance = AnchoredCompliance(address(complianceProxy));

        // Deploy TokenPauseManager implementation and proxy
        AnchoredTokenPauseManager pauseManagerImpl = new AnchoredTokenPauseManager(); // Use zero address for implementation
        TransparentUpgradeableProxy pauseManagerProxy = new TransparentUpgradeableProxy(
            address(pauseManagerImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(AnchoredTokenPauseManager.initialize.selector, admin)
        );
        pauseManager = AnchoredTokenPauseManager(address(pauseManagerProxy));

        // Deploy AnchoredTokenManager implementation and proxy
        AnchoredTokenManager tokenManagerImpl =
            new AnchoredTokenManager(address(usdcToken), keccak256("ANCHORED_STOCK_TOKEN_IDENTIFIER"));
        TransparentUpgradeableProxy tokenManagerProxy = new TransparentUpgradeableProxy(
            address(tokenManagerImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(
                AnchoredTokenManager.initialize.selector,
                address(usdcToken),
                admin,
                1000 * 1e6, // minimum deposit 1000 USDC
                500 * 1e6 // minimum redemption 500 USDC
            )
        );
        tokenManager = AnchoredTokenManager(address(tokenManagerProxy));

        // Deploy TokenManagerRegistrar implementation and proxy
        AnchoredTokenManagerRegistrar registrarImpl = new AnchoredTokenManagerRegistrar(); // Use zero addresses for implementation
        TransparentUpgradeableProxy registrarProxy = new TransparentUpgradeableProxy(
            address(registrarImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(AnchoredTokenManagerRegistrar.initialize.selector, admin, address(tokenManager))
        );
        registrar = AnchoredTokenManagerRegistrar(address(registrarProxy));

        // Deploy AnchoredTokenFactory implementation and proxy
        AnchoredTokenFactory tokenFactoryImpl = new AnchoredTokenFactory(); // Use zero addresses for implementation
        TransparentUpgradeableProxy tokenFactoryProxy = new TransparentUpgradeableProxy(
            address(tokenFactoryImpl),
            address(proxyAdmin),
            abi.encodeWithSelector(
                AnchoredTokenFactory.initialize.selector,
                admin,
                address(anchoredCompliance),
                address(pauseManager),
                address(registrar)
            )
        );
        tokenFactory = AnchoredTokenFactory(address(tokenFactoryProxy));

        // Transfer beacon ownership from factory implementation to proxy admin
        // The beacon is owned by the factory implementation, so we need to transfer from there
        vm.stopPrank(); // Stop pranking as admin
        vm.startPrank(address(tokenFactoryImpl)); // Prank as factory implementation
        UpgradeableBeacon(tokenFactory.BEACON()).transferOwnership(address(proxyAdmin));
        vm.stopPrank(); // Stop pranking as factory implementation
        vm.startPrank(admin); // Resume pranking as admin

        // Note: AnchoredCompliance doesn't have CONFIGURE_ROLE, using MASTER_CONFIGURE_ROLE instead
        anchoredCompliance.grantRole(anchoredCompliance.MASTER_CONFIGURE_ROLE(), configurer);
        // Blocklist now uses AccessControlEnumerable, grant admin role to configurer
        blocklist.grantRole(blocklist.DEFAULT_ADMIN_ROLE(), configurer);
        // SanctionsList now uses AccessControlEnumerable, grant admin role to configurer
        sanctionsList.grantRole(sanctionsList.DEFAULT_ADMIN_ROLE(), configurer);
        sanctionsList.grantRole(sanctionsList.SANCTIONS_ADD_ROLE(), configurer);

        pauseManager.grantRole(pauseManager.PAUSE_TOKEN_ROLE(), pauser);
        pauseManager.grantRole(pauseManager.UNPAUSE_TOKEN_ROLE(), pauser);
        tokenFactory.grantRole(tokenFactory.DEPLOY_ROLE(), configurer);
        registrar.grantRole(registrar.TOKEN_REGISTER_ROLE(), address(tokenFactory));

        // Grant roles
        tokenManager.grantRole(tokenManager.CONFIGURE_ROLE(), configurer);
        tokenManager.grantRole(tokenManager.CONFIGURE_ROLE(), address(registrar));
        tokenManager.grantRole(tokenManager.PAUSE_ROLE(), pauser);
        tokenManager.grantRole(tokenManager.QUOTE_SIGN_ROLE(), priceSigner);

        vm.stopPrank();

        // Deploy Anchored tokens
        vm.startPrank(configurer);

        rwaToken1 = AnchoredToken(tokenFactory.deployAndRegisterToken("Anchored Token 1", "ANCH1", admin));

        rwaToken2 = AnchoredToken(tokenFactory.deployAndRegisterToken("Anchored Token 2", "ANCH2", admin));

        // Add malicious user to sanctions list
        // SanctionsList now uses AccessControlEnumerable, no need to accept ownership
        address[] memory sanctionedUsers = new address[](1);
        sanctionedUsers[0] = maliciousUser;
        sanctionsList.addToSanctionsList(sanctionedUsers);

        // Set compliance configuration
        anchoredCompliance.setBlocklist(address(rwaToken1), IAnchoredBlocklist(address(blocklist)));
        anchoredCompliance.setBlocklist(address(rwaToken2), IAnchoredBlocklist(address(blocklist)));
        anchoredCompliance.setSanctionsList(address(rwaToken1), IAnchoredSanctionsList(address(sanctionsList)));
        anchoredCompliance.setSanctionsList(address(rwaToken2), IAnchoredSanctionsList(address(sanctionsList)));

        vm.stopPrank();

        // Mint USDC to test users
        usdcToken.mint(alice, 10000000e18);
        usdcToken.mint(bob, 10000000e6);
        usdcToken.mint(charlie, 10000000e6);
        usdcToken.mint(maliciousUser, 10000000e6);
        usdcToken.mint(institutionalUser, 100000000e6);
    }

    function testCompleteWorkflow() public {
        uint256 mintAmount = 1000e18;
        uint256 collateralAmount = 1000e18; // Should match mintUSDCValue calculation

        // Step 1: Alice approves USDC
        vm.prank(alice);
        usdcToken.approve(address(tokenManager), collateralAmount);

        // Step 2: Create mint quote
        IAnchoredTokenManager.Quote memory quote = IAnchoredTokenManager.Quote({
            attestationId: bytes32(uint256(1)),
            asset: address(rwaToken1),
            price: 1e18, // 1 Usd per token
            quantity: mintAmount,
            expiry: block.timestamp + 1 hours,
            side: IAnchoredTokenManager.QuoteSide.BUY,
            nonce: 1
        });

        // Step 3: Create attestation signature
        bytes memory signature = _generateSignature(quote);

        // Step 4: Alice mints Anchored tokens
        vm.prank(alice);
        tokenManager.mintWithAttestation(quote, signature, address(usdcToken), collateralAmount);

        // Verify mint success
        assertEq(rwaToken1.balanceOf(alice), mintAmount);
        assertEq(usdcToken.balanceOf(alice), 10000000e18 - collateralAmount);

        // Step 5: Alice redeems some tokens
        uint256 redeemAmount = 500e18;

        IAnchoredTokenManager.Quote memory redeemQuote = IAnchoredTokenManager.Quote({
            attestationId: bytes32(uint256(2)),
            asset: address(rwaToken1),
            price: 1e18,
            quantity: redeemAmount,
            expiry: block.timestamp + 1 hours,
            side: IAnchoredTokenManager.QuoteSide.SELL,
            nonce: 2
        });

        bytes memory redeemSignature = _generateSignature(redeemQuote);

        // Alice needs to approve token manager to spend her tokens
        vm.prank(alice);
        rwaToken1.approve(address(tokenManager), redeemAmount);

        uint256 aliceUsdcBefore = usdcToken.balanceOf(alice);
        vm.prank(alice);
        tokenManager.redeemWithAttestation(redeemQuote, redeemSignature, address(usdcToken));

        // Verify redemption success
        assertEq(rwaToken1.balanceOf(alice), mintAmount - redeemAmount);
        assertEq(usdcToken.balanceOf(alice), aliceUsdcBefore + redeemAmount); // Both use 18 decimals now
    }

    function testComplianceBlocking() public {
        uint256 mintAmount = 1000e18;
        uint256 collateralAmount = 1000e18; // Should match mintUSDCValue calculation

        // Try to mint for sanctioned user
        IAnchoredTokenManager.Quote memory maliciousQuote = IAnchoredTokenManager.Quote({
            attestationId: bytes32(uint256(3)),
            asset: address(rwaToken1),
            price: 1e18,
            quantity: mintAmount,
            expiry: block.timestamp + 1 hours,
            side: IAnchoredTokenManager.QuoteSide.BUY,
            nonce: 3
        });

        // Generate signature with wrong private key to simulate malicious signature
        uint256 wrongPrivateKey = 0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890;
        bytes32 messageHash = keccak256(
            abi.encode(
                maliciousQuote.asset,
                maliciousQuote.side,
                maliciousQuote.quantity,
                maliciousQuote.price,
                maliciousQuote.nonce,
                maliciousQuote.expiry,
                maliciousQuote.attestationId
            )
        );
        bytes32 ethSignedMessageHash = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", messageHash));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(wrongPrivateKey, ethSignedMessageHash);
        bytes memory maliciousSignature = abi.encodePacked(r, s, v);

        vm.prank(maliciousUser);
        usdcToken.approve(address(tokenManager), collateralAmount);

        // Mint should fail due to sanctions
        vm.prank(maliciousUser);
        vm.expectRevert();
        tokenManager.mintWithAttestation(maliciousQuote, maliciousSignature, address(usdcToken), collateralAmount);

        // Verify no tokens were minted
        assertEq(rwaToken1.balanceOf(maliciousUser), 0);
    }

    function testTransferRestrictions() public {
        uint256 mintAmount = 1000e18;
        uint256 collateralAmount = 1000e18; // Should match mintUSDCValue calculation

        // Alice mints tokens first
        vm.prank(alice);
        usdcToken.approve(address(tokenManager), collateralAmount);

        IAnchoredTokenManager.Quote memory aliceQuote = IAnchoredTokenManager.Quote({
            attestationId: bytes32(uint256(4)),
            asset: address(rwaToken1),
            price: 1e18,
            quantity: mintAmount,
            expiry: block.timestamp + 1 hours,
            side: IAnchoredTokenManager.QuoteSide.BUY,
            nonce: 4
        });

        bytes memory aliceSignature = _generateSignature(aliceQuote);

        vm.prank(alice);
        tokenManager.mintWithAttestation(aliceQuote, aliceSignature, address(usdcToken), collateralAmount);

        // Try to transfer to sanctioned user - should fail
        vm.prank(alice);
        vm.expectRevert();
        rwaToken1.transfer(maliciousUser, 100e18);

        // Transfer to normal user should work
        vm.prank(alice);
        bool success = rwaToken1.transfer(bob, 100e18);
        assertTrue(success, "Transfer should succeed");

        assertEq(rwaToken1.balanceOf(bob), 100e18);
        assertEq(rwaToken1.balanceOf(alice), mintAmount - 100e18);
    }

    function testPauseUnpause() public {
        uint256 mintAmount = 1000e18;
        uint256 collateralAmount = 1000e18; // Should match mintUSDCValue calculation

        // Alice mints tokens first
        vm.prank(alice);
        usdcToken.approve(address(tokenManager), collateralAmount);

        IAnchoredTokenManager.Quote memory aliceQuote = IAnchoredTokenManager.Quote({
            attestationId: bytes32(uint256(5)),
            asset: address(rwaToken1),
            price: 1e18,
            quantity: mintAmount,
            expiry: block.timestamp + 1 hours,
            side: IAnchoredTokenManager.QuoteSide.BUY,
            nonce: 5
        });

        bytes memory aliceSignature = _generateSignature(aliceQuote);

        vm.prank(alice);
        tokenManager.mintWithAttestation(aliceQuote, aliceSignature, address(usdcToken), collateralAmount);

        // Pause the token
        vm.prank(pauser);
        pauseManager.pauseToken(address(rwaToken1));

        // Transfers should be blocked
        vm.prank(alice);
        vm.expectRevert();
        rwaToken1.transfer(bob, 100e18);

        // Unpause the token
        vm.prank(pauser);
        pauseManager.unpauseToken(address(rwaToken1));

        // Transfers should work again
        vm.prank(alice);
        bool success = rwaToken1.transfer(bob, 100e18);
        assertTrue(success, "Transfer should succeed");

        assertEq(rwaToken1.balanceOf(bob), 100e18);
    }

    function testMultiTokenMinting() public {
        uint256 mintAmount = 1000e18;
        uint256 collateralAmount = 1000e18; // Should match mintUSDCValue calculation

        vm.prank(alice);
        usdcToken.approve(address(tokenManager), collateralAmount * 2);

        // Mint Anchored Token 1
        IAnchoredTokenManager.Quote memory quote1 = IAnchoredTokenManager.Quote({
            attestationId: bytes32(uint256(6)),
            asset: address(rwaToken1),
            price: 1e18,
            quantity: mintAmount,
            expiry: block.timestamp + 1 hours,
            side: IAnchoredTokenManager.QuoteSide.BUY,
            nonce: 6
        });

        bytes memory signature1 = _generateSignature(quote1);

        vm.prank(alice);
        tokenManager.mintWithAttestation(quote1, signature1, address(usdcToken), collateralAmount);

        // Mint Anchored Token 2
        IAnchoredTokenManager.Quote memory quote2 = IAnchoredTokenManager.Quote({
            attestationId: bytes32(uint256(7)),
            asset: address(rwaToken2),
            price: 1e18,
            quantity: mintAmount,
            expiry: block.timestamp + 1 hours,
            side: IAnchoredTokenManager.QuoteSide.BUY,
            nonce: 7
        });

        bytes memory signature2 = _generateSignature(quote2);

        vm.prank(alice);
        tokenManager.mintWithAttestation(quote2, signature2, address(usdcToken), collateralAmount);

        // Verify both tokens were minted
        assertEq(rwaToken1.balanceOf(alice), mintAmount);
        assertEq(rwaToken2.balanceOf(alice), mintAmount);

        // Test transfers work for both tokens
        vm.prank(alice);
        bool success1 = rwaToken1.transfer(bob, 100e18);
        assertTrue(success1, "Transfer should succeed");

        vm.prank(alice);
        bool success2 = rwaToken2.transfer(bob, 200e18);
        assertTrue(success2, "Transfer should succeed");

        assertEq(rwaToken1.balanceOf(bob), 100e18);
        assertEq(rwaToken2.balanceOf(bob), 200e18);
    }
}

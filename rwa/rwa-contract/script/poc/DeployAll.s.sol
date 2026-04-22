// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;

import { Script, console } from "forge-std/Script.sol";
import { MockUSDC } from "../../contracts/poc/MockUSDC.sol";
import { PocToken } from "../../contracts/poc/PocToken.sol";
import { OrderContract } from "../../contracts/poc/Order.sol";
import { PocGate } from "../../contracts/poc/PocGate.sol";
import { UpgradeableBeacon } from "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";
import { BeaconProxy } from "@openzeppelin/contracts/proxy/beacon/BeaconProxy.sol";
import { TransparentUpgradeableProxy } from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

contract DeployAll is Script {
    // ============ Configuration ============
    // Modify these before deployment
    address public deployer;
    address public backendAddress;
    address public proxyAdmin;

    // Deployed addresses (populated during run)
    address public mockUSDCAddress;
    address public pocTokenImpl;
    address public beacon;
    address public usdmProxy;
    address public aaplProxy;
    address public tslaProxy;
    address public orderImpl;
    address public orderProxy;
    address public pocGateImpl;
    address public pocGateProxy;

    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        deployer = vm.addr(deployerPrivateKey);
        // Backend address: defaults to deployer if not set
        backendAddress = vm.envOr("BACKEND_ADDRESS", deployer);
        // Proxy admin: defaults to deployer if not set
        proxyAdmin = vm.envOr("PROXY_ADMIN_ADDRESS", deployer);

        console.log("Deployer:", deployer);
        console.log("Backend:", backendAddress);
        console.log("ProxyAdmin:", proxyAdmin);

        vm.startBroadcast(deployerPrivateKey);

        // =============================================
        // Step 1: Deploy MockUSDC
        // =============================================
        MockUSDC mockUSDC = new MockUSDC();
        mockUSDCAddress = address(mockUSDC);
        console.log("MockUSDC deployed at:", mockUSDCAddress);

        // Mint 1,000,000 USDC (6 decimals) to deployer
        mockUSDC.mint(deployer, 1_000_000 * 1e6);
        console.log("Minted 1,000,000 USDC to deployer");

        // =============================================
        // Step 2: Deploy USDM (PocToken via Beacon Proxy)
        // =============================================
        // 2a. Deploy PocToken implementation (constructor needs gateContract_ address, use deployer as placeholder)
        PocToken pocTokenImplContract = new PocToken(deployer);
        pocTokenImpl = address(pocTokenImplContract);
        console.log("PocToken implementation deployed at:", pocTokenImpl);

        // 2b. Deploy UpgradeableBeacon pointing to PocToken implementation
        UpgradeableBeacon beaconContract = new UpgradeableBeacon(pocTokenImpl, proxyAdmin);
        beacon = address(beaconContract);
        console.log("UpgradeableBeacon deployed at:", beacon);

        // 2c. Deploy BeaconProxy for USDM with initialize("USDM", "USDM")
        bytes memory usdmInitData = abi.encodeWithSelector(PocToken.initialize.selector, "USDM", "USDM");
        BeaconProxy usdmBeaconProxy = new BeaconProxy(beacon, usdmInitData);
        usdmProxy = address(usdmBeaconProxy);
        console.log("USDM (BeaconProxy) deployed at:", usdmProxy);

        // =============================================
        // Step 3: Deploy Stock Tokens (AAPL.anc, TSLA.anc)
        // =============================================
        // 3a. Deploy AAPL BeaconProxy
        bytes memory aaplInitData = abi.encodeWithSelector(PocToken.initialize.selector, "AAPL.anc", "AAPL.anc");
        BeaconProxy aaplBeaconProxy = new BeaconProxy(beacon, aaplInitData);
        aaplProxy = address(aaplBeaconProxy);
        console.log("AAPL.anc (BeaconProxy) deployed at:", aaplProxy);

        // 3b. Deploy TSLA BeaconProxy
        bytes memory tslaInitData = abi.encodeWithSelector(PocToken.initialize.selector, "TSLA.anc", "TSLA.anc");
        BeaconProxy tslaBeaconProxy = new BeaconProxy(beacon, tslaInitData);
        tslaProxy = address(tslaBeaconProxy);
        console.log("TSLA.anc (BeaconProxy) deployed at:", tslaProxy);

        // =============================================
        // Step 4: Deploy OrderContract (TransparentUpgradeableProxy)
        // =============================================
        // 4a. Deploy OrderContract implementation
        OrderContract orderImplContract = new OrderContract();
        orderImpl = address(orderImplContract);
        console.log("OrderContract implementation deployed at:", orderImpl);

        // 4b. Deploy TransparentUpgradeableProxy for OrderContract
        //     initialize(address usdm_, address admin_, address backend_)
        bytes memory orderInitData =
            abi.encodeWithSelector(OrderContract.initialize.selector, usdmProxy, deployer, backendAddress);
        TransparentUpgradeableProxy orderProxyContract =
            new TransparentUpgradeableProxy(orderImpl, proxyAdmin, orderInitData);
        orderProxy = address(orderProxyContract);
        console.log("OrderContract (Proxy) deployed at:", orderProxy);

        // =============================================
        // Step 5: Deploy PocGate (TransparentUpgradeableProxy)
        // =============================================
        // 5a. Deploy PocGate implementation (constructor needs usdc_ and usdm_)
        PocGate pocGateImplContract = new PocGate(mockUSDCAddress, usdmProxy);
        pocGateImpl = address(pocGateImplContract);
        console.log("PocGate implementation deployed at:", pocGateImpl);

        // 5b. Deploy TransparentUpgradeableProxy for PocGate
        //     initialize(address guardian_, uint256 minimumDepositAmount_, uint256 minimumWithdrawalAmount_)
        bytes memory gateInitData = abi.encodeWithSelector(PocGate.initialize.selector, deployer, uint256(0), uint256(0));
        TransparentUpgradeableProxy gateProxyContract =
            new TransparentUpgradeableProxy(pocGateImpl, proxyAdmin, gateInitData);
        pocGateProxy = address(gateProxyContract);
        console.log("PocGate (Proxy) deployed at:", pocGateProxy);

        // =============================================
        // Step 6: Configure Permissions
        // =============================================
        PocToken usdm = PocToken(usdmProxy);
        PocToken aapl = PocToken(aaplProxy);
        PocToken tsla = PocToken(tslaProxy);
        OrderContract order = OrderContract(orderProxy);

        bytes32 MINTER_ROLE = keccak256("MINTER_ROLE");
        bytes32 BURNER_ROLE = keccak256("BURNER_ROLE");

        // 6a. Grant OrderContract MINTER_ROLE on USDM (for minting USDM)
        usdm.grantRole(MINTER_ROLE, orderProxy);
        console.log("Granted MINTER_ROLE on USDM to OrderContract");

        // 6b. Register symbol tokens on OrderContract
        order.setSymbolToken("AAPL", aaplProxy);
        console.log("Set AAPL symbol token on OrderContract");

        order.setSymbolToken("TSLA", tslaProxy);
        console.log("Set TSLA symbol token on OrderContract");

        // 6c. Grant PocGate MINTER_ROLE and BURNER_ROLE on USDM
        usdm.grantRole(MINTER_ROLE, pocGateProxy);
        usdm.grantRole(BURNER_ROLE, pocGateProxy);
        console.log("Granted MINTER_ROLE and BURNER_ROLE on USDM to PocGate");

        // 6d. Grant OrderContract MINTER_ROLE and BURNER_ROLE on stock tokens
        aapl.grantRole(MINTER_ROLE, orderProxy);
        aapl.grantRole(BURNER_ROLE, orderProxy);
        console.log("Granted MINTER_ROLE and BURNER_ROLE on AAPL to OrderContract");

        tsla.grantRole(MINTER_ROLE, orderProxy);
        tsla.grantRole(BURNER_ROLE, orderProxy);
        console.log("Granted MINTER_ROLE and BURNER_ROLE on TSLA to OrderContract");

        vm.stopBroadcast();

        // =============================================
        // Summary
        // =============================================
        console.log("");
        console.log("========== Deployment Summary ==========");
        console.log("MockUSDC:          ", mockUSDCAddress);
        console.log("PocToken Impl:     ", pocTokenImpl);
        console.log("UpgradeableBeacon: ", beacon);
        console.log("USDM Proxy:        ", usdmProxy);
        console.log("AAPL.anc Proxy:    ", aaplProxy);
        console.log("TSLA.anc Proxy:    ", tslaProxy);
        console.log("OrderContract Impl:", orderImpl);
        console.log("OrderContract Proxy:", orderProxy);
        console.log("PocGate Impl:      ", pocGateImpl);
        console.log("PocGate Proxy:     ", pocGateProxy);
        console.log("=========================================");
    }
}

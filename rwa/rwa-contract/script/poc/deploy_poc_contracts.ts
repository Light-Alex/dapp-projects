import * as dotenv from 'dotenv';
import { Account, createPublicClient, createWalletClient, encodeFunctionData, http, isAddress } from 'viem';
import { abctest } from '@derivation-tech/viem-kit/dist/chains/abctest.js';
import { ChainKitRegistry, getAccount } from '@derivation-tech/viem-kit';
import { deployArtifact } from '@derivation-tech/viem-kit/dist/utils/tx.js';
import { loadArtifact } from '@derivation-tech/viem-kit/dist/utils/artifact-helper.js';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { DEPLOYED_TOKENS } from './constants';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);


async function main(): Promise<void> {
    dotenv.config();
    const chain = abctest;
    const kit = ChainKitRegistry.for(chain);
    const signer_id = 'julian:32';
    const signer = getAccount(kit, signer_id);
    const proxyAdmin_id = 'julian:11';
    const proxyAdmin = getAccount(kit, proxyAdmin_id);
    const basePath = path.resolve(__dirname, '../../out');

    const walletClient = createWalletClient({
        chain,
        transport: http(),
        account: signer as Account,
    }) as any;

    const publicClient = createPublicClient({
        chain,
        transport: http(),
    }) as any;

    console.log('signer address: ', signer.address);
    console.log('proxyAdmin address: ', proxyAdmin.address);

    // load upgradableBeacon artifact
    const upgradableBeaconArtifact = loadArtifact(path.resolve(basePath, 'UpgradeableBeacon.sol/UpgradeableBeacon.json'));
    // load beaconProxy artifact
    const beaconProxyArtifact = loadArtifact(path.resolve(basePath, 'BeaconProxy.sol/BeaconProxy.json'));
    // load pocToken artifact
    const pocTokenArtifact = loadArtifact(path.resolve(basePath, 'PocToken.sol/PocToken.json'));
    // load Order artifact
    const orderArtifact = loadArtifact(path.resolve(basePath, 'Order.sol/OrderContract.json'));
    // load mockUSDC artifact
    const mockUSDCArtifact = loadArtifact(path.resolve(basePath, 'MockUSDC.sol/MockUSDC.json'));
    // load TransparentUpgradeableProxy artifact
    const transparentUpgradeableProxyArtifact = loadArtifact(path.resolve(basePath, 'TransparentUpgradeableProxy.sol/TransparentUpgradeableProxy.json'));
    // load PocGate artifact
    const pocGateArtifact = loadArtifact(path.resolve(basePath, 'PocGate.sol/PocGate.json'));

    // deploy mockUSDC
    // const mockUSDCAddress = await deployArtifact(publicClient, walletClient, {
    //     artifact: mockUSDCArtifact,
    //     constructorArgs: [],
    // });
    // console.log('mockUSDC address: ', mockUSDCAddress);
    // // mint mockUSDC
    // await walletClient.writeContract({
    //     address: mockUSDCAddress,
    //     abi: mockUSDCArtifact.abi,
    //     functionName: 'mint',
    //     args: [signer.address, 1000000000000],
    // });
    // console.log('mockUSDC minted to signer address: ', signer.address);

    // deploy pocToken
    // const pocTokenAddress = await deployArtifact(publicClient, walletClient,{
    //     artifact: pocTokenArtifact,
    //     constructorArgs: ['0x191dE7187f0f3143d485Ab058dF990c0864d7542'], // TODO: this is fake
    // });
    // console.log('pocToken address: ', pocTokenAddress);

    // deploy upgradableBeacon
    // const upgradableBeaconAddress = await deployArtifact(publicClient, walletClient, {
    //     artifact: upgradableBeaconArtifact,
    //     constructorArgs: ['0x997aA03eE41e4555f531B9C75Ed0CFE950A94d4C', proxyAdmin.address],
    // });
    // console.log('upgradableBeacon address: ', upgradableBeaconAddress);

    // deploy beacon proxy
    // const encodedInitData = encodeFunctionData({
    //     abi: pocTokenArtifact.abi,
    //     functionName: 'initialize',
    //     args: ['USDC', 'USDC'],
    // });


    // const beaconProxyAddress = await deployArtifact(publicClient, walletClient, {
    //     artifact: beaconProxyArtifact,
    //     constructorArgs: ['0xA9A78c647561A3823F1E48b4e151318Ed42C4eC4', encodedInitData],
    //     confirmations: 0,
    // });
    // console.log('beaconProxy address: ', beaconProxyAddress);

    // USDM : 0x7ffd1A23f1e53737eDB9C9c35a8E6b6d33abD96b
    // TSLA:  0x752f78a728acd26cd47eb723c5dc4a2ab9d6cd58
    // AAPL:  0xe4088b68aa81e6bf456bcdce9e6dfeacecc6842c

    // deploy OrderImpl
    // const orderImplAddress = await deployArtifact(publicClient, walletClient, {
    //     artifact: orderArtifact,
    //     constructorArgs: [],
    //     confirmations: 0,
    // });
    // console.log('orderImpl address: ', orderImplAddress);

    // const encodedInitData = encodeFunctionData({
    //     abi: orderArtifact.abi,
    //     functionName: 'initialize',
    //     // USDM,Admin,Backend
    //     args: ['0x7ffd1A23f1e53737eDB9C9c35a8E6b6d33abD96b', signer.address, '0x892C54E623aecF127B3285F3f14E39CD0275afE9'],
    // });
    // console.log('encodedInitData: ', encodedInitData);

    // deploy OrderProxy
    // const orderProxyAddress = await deployArtifact(publicClient, walletClient, {
    //     artifact: transparentUpgradeableProxyArtifact,
    //     constructorArgs: [orderImplAddress, proxyAdmin.address, encodedInitData],
    //     confirmations: 0,
    // });
    // console.log('orderProxy address: ', orderProxyAddress);

    // deploy PocGate Impl
    const pocGateImplAddress = await deployArtifact(publicClient, walletClient, {
        artifact: pocGateArtifact,
        constructorArgs: [DEPLOYED_TOKENS['USDC'], DEPLOYED_TOKENS['USDM']],
        confirmations: 0,
    });
    console.log('pocGateImpl address: ', pocGateImplAddress);

    const encodedInitData = encodeFunctionData({
        abi: pocGateArtifact.abi,
        functionName: 'initialize',
        // guardian, minimumDepositAmount_, minimumWithdrawAmount_
        args: [signer.address, '0', '0'],
    });
    console.log('encodedInitData: ', encodedInitData);

    // deploy PocGateProxy
    const pocGateProxyAddress = await deployArtifact(publicClient, walletClient, {
        artifact: transparentUpgradeableProxyArtifact,
        constructorArgs: [pocGateImplAddress, proxyAdmin.address, encodedInitData],
        confirmations: 0,
    });
    console.log('pocGateProxy address: ', pocGateProxyAddress);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
    console.error(error);
});

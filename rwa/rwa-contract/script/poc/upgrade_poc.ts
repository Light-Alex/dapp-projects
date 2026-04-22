import * as dotenv from 'dotenv';
import { Account, createPublicClient, createWalletClient, encodeFunctionData, http, isAddress } from 'viem';
import { abctest } from '@derivation-tech/viem-kit/dist/chains/abctest.js';
import { ChainKitRegistry, getAccount } from '@derivation-tech/viem-kit';
import { deployArtifact, sendTxWithLog } from '@derivation-tech/viem-kit/dist/utils/tx.js';
import { loadArtifact } from '@derivation-tech/viem-kit/dist/utils/artifact-helper.js';
import * as path from 'path';
import { fileURLToPath } from 'url';

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
        account: proxyAdmin as Account,
    }) as any;

    const publicClient = createPublicClient({
        chain,
        transport: http(),
    }) as any;

    console.log('signer address: ', signer.address);
    console.log('proxyAdmin address: ', proxyAdmin.address);
    // load upgradableBeacon artifact
    const upgradableBeaconArtifact = loadArtifact(path.resolve(basePath, 'UpgradeableBeacon.sol/UpgradeableBeacon.json'));
    const newImpl = '0x13c2e3F6ad5f97961D23159D4b6c403a95259699';
    // upgrade beacon
    const upgradeBeaconTx = await sendTxWithLog(publicClient, walletClient, kit, {
        address: '0xA9A78c647561A3823F1E48b4e151318Ed42C4eC4' as `0x${string}`,
        abi: upgradableBeaconArtifact.abi,
        functionName: 'upgradeTo',
        args: [newImpl],
    });
    await publicClient.waitForTransactionReceipt({ hash: upgradeBeaconTx.transactionHash });
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
    console.error(error);
});

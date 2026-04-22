import * as dotenv from 'dotenv';
import { Account, Address, createPublicClient, createWalletClient, http, parseEther, decodeEventLog, Abi } from 'viem';
import { abctest } from '@derivation-tech/viem-kit/dist/chains/abctest.js';
import { ChainKitRegistry, getAccount } from '@derivation-tech/viem-kit';
import { sendTxWithLog } from '@derivation-tech/viem-kit/dist/utils/tx.js';
import { loadArtifact } from '@derivation-tech/viem-kit/dist/utils/artifact-helper.js';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { createOrderParser } from '../parsers/orderParser';
import { createPocTokenParser } from '../parsers/pocTokenParser';
import { DEPLOYED_ORDER_CONTRACT, DEPLOYED_TOKENS } from './constants';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);


async function main(): Promise<void> {
    dotenv.config();
    const chain = abctest;
    const kit = ChainKitRegistry.for(chain);
    const signer_id = 'julian:32';
    const signer = getAccount(kit, signer_id);
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

    // initialize parser with deps
    console.log('signer address: ', signer.address);
    const pocTokenArtifact = loadArtifact(path.resolve(basePath, 'PocToken.sol/PocToken.json'));
    const orderArtifact = loadArtifact(path.resolve(basePath, 'Order.sol/OrderContract.json'));
    const tokenArtifact = loadArtifact(path.resolve(basePath, 'PocToken.sol/PocToken.json'));

    const orderParser = createOrderParser(orderArtifact.abi, { publicClient });
    const pocTokenParser = createPocTokenParser(pocTokenArtifact.abi, { publicClient });
    // register Parser
    kit.registerParser(DEPLOYED_ORDER_CONTRACT, orderParser);
    for (const token of Object.values(DEPLOYED_TOKENS)) {
        kit.registerParser(token as `0x${string}`, pocTokenParser);
    }

    const name = 'TSLA';


    await sendTxWithLog(publicClient, walletClient, kit, {
        address: DEPLOYED_TOKENS[name] as `0x${string}`,
            abi: tokenArtifact.abi,
            functionName: 'approve',
        args: [DEPLOYED_ORDER_CONTRACT, parseEther('10000000000')],
    });

    await sendTxWithLog(publicClient, walletClient, kit, {
        address: DEPLOYED_ORDER_CONTRACT as `0x${string}`,
            abi: orderArtifact.abi,
            functionName: 'submitOrder',
            args: [ name,parseEther('1'),parseEther('134'),0,0,0],
    });

}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
    console.error(error);
});

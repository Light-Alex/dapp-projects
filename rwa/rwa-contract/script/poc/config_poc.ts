import * as dotenv from 'dotenv';
import { Account, createPublicClient, createWalletClient, encodeFunctionData, http, isAddress } from 'viem';
import { abctest } from '@derivation-tech/viem-kit/dist/chains/abctest.js';
import { ChainKitRegistry, getAccount } from '@derivation-tech/viem-kit';
import { deployArtifact, sendTxWithLog } from '@derivation-tech/viem-kit/dist/utils/tx.js';
import { loadArtifact } from '@derivation-tech/viem-kit/dist/utils/artifact-helper.js';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { DEPLOYED_GATE_CONTRACT } from './constants';
import { createPocTokenParser } from '../parsers/pocTokenParser';

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
    const pocTokenArtifact = loadArtifact(path.resolve(basePath, 'PocToken.sol/PocToken.json'));
    const orderArtifact = loadArtifact(path.resolve(basePath, 'Order.sol/OrderContract.json'));
    const pocTokenParser = createPocTokenParser(pocTokenArtifact.abi, { publicClient });


    const operator = '0x892C54E623aecF127B3285F3f14E39CD0275afE9';
    const orderContractAddress = '0xae136110e64556bc15df5db254929c3a4a09dece';


    // USDM : 0x7ffd1A23f1e53737eDB9C9c35a8E6b6d33abD96b
    // TSLA:  0x752f78a728acd26cd47eb723c5dc4a2ab9d6cd58
    // AAPL:  0xe4088b68aa81e6bf456bcdce9e6dfeacecc6842c

    // call set operator on all test tokens
    const testTokens = [
        '0x7ffd1A23f1e53737eDB9C9c35a8E6b6d33abD96b',
        '0x752f78a728acd26cd47eb723c5dc4a2ab9d6cd58',
        '0xe4088b68aa81e6bf456bcdce9e6dfeacecc6842c',
    ];
    for (const token of testTokens) {
        kit.registerParser(token as `0x${string}`, pocTokenParser);

        const name = await publicClient.readContract({
            address: token as `0x${string}`,
            abi: pocTokenArtifact.abi,
            functionName: 'name',
        });
        console.log('token name: ', name);
        let setOperatorTx = await sendTxWithLog(publicClient, walletClient, kit, {
            address: token as `0x${string}`,
            abi: pocTokenArtifact.abi,
            functionName: 'setOperator',
            args: [operator],
        });
        await publicClient.waitForTransactionReceipt({ hash: setOperatorTx.transactionHash });

        setOperatorTx = await sendTxWithLog(publicClient, walletClient, kit, {
            address: token as `0x${string}`,
            abi: pocTokenArtifact.abi,
            functionName: 'setOperator',
            args: [orderContractAddress],
        });
        await publicClient.waitForTransactionReceipt({ hash: setOperatorTx.transactionHash });


        setOperatorTx = await sendTxWithLog(publicClient, walletClient, kit, {
            address: token as `0x${string}`,
            abi: pocTokenArtifact.abi,
            functionName: 'setOperator',
            args: [DEPLOYED_GATE_CONTRACT],
        });
        await publicClient.waitForTransactionReceipt({ hash: setOperatorTx.transactionHash });


        const registerTx = await sendTxWithLog(publicClient, walletClient, kit, {
            address: orderContractAddress as `0x${string}`,
            abi: orderArtifact.abi,
            functionName: 'setSymbolToken',
            args: [name as string, token as `0x${string}`],
        });
        await publicClient.waitForTransactionReceipt({ hash: registerTx.transactionHash });

        const validateAddress = await publicClient.readContract({
            address: orderContractAddress as `0x${string}`,
            abi: orderArtifact.abi,
            functionName: 'symbolToToken',
            args: [name as string],
        });
        console.log('validateAddress: ', validateAddress);
        if (validateAddress.toLowerCase() !== token.toLowerCase()) {
            throw new Error('validateAddress not equal to token', {
                cause: {
                    validateAddress,
                    token,
                },
            });
        } else {
            console.log('register token success: ', name);
        }
    }
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
    console.error(error);
});

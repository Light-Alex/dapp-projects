import { ethers, Wallet, Contract, parseUnits } from 'ethers';
import dotenv from 'dotenv';
import * as path from 'path';
import * as fs from 'fs';

dotenv.config();

// 配置参数 - 请修改为实际的合约地址
const AIRDROP_CONTRACT_ADDRESS = '0xDad6EcF65ef31C11D568Bb0E3B4F52Da169364D0'; // Airdrop 合约地址
const MYERC20_CONTRACT_ADDRESS = '0x3d720e33044bD6691A850cEcF03E5D1C9e25Eda3'; // MyERC20 代币合约地址
const RPC_URL = 'https://bsc-testnet-dataseed.bnbchain.org';
const BATCH_SIZE = 100; // 每批处理数量，避免超过 gas limit

// // 导入 ABI（使用动态路径，编译前后都能正确找到）
const airdropABI = JSON.parse(fs.readFileSync(path.join(__dirname, '../src/abis/Airdrop.json'), 'utf8'));
const myErc20ABI = JSON.parse(fs.readFileSync(path.join(__dirname, '../src/abis/MyERC20.json'), 'utf8'));

/**
 * 随机生成钱包地址
 * @param count 生成数量
 * @returns 钱包地址数组
 */
function generateRandomAddresses(count: number): string[] {
  const addresses: string[] = [];
  for (let i = 0; i < count; i++) {
    const wallet = Wallet.createRandom();
    addresses.push(wallet.address);
  }
  return addresses;
}

/**
 * ERC20 空投 - 批量处理
 * @param recipients 接收地址数组
 * @param amount 每个地址接收的代币数量
 */
async function airdropERC20(
  contract: Contract,
  tokenDecimals: number,
  recipients: string[],
  amount: bigint
): Promise<void> {
  console.log(`\n🎁 开始 ERC20 空投...`);
  console.log(`接收地址数: ${recipients.length}`);
  console.log(`每个地址: ${ethers.formatUnits(amount, tokenDecimals)} Tokens`);

  const totalAmount = amount * BigInt(recipients.length);
  console.log(`总空投量: ${ethers.formatUnits(totalAmount, tokenDecimals)} Tokens`);

  // 批量处理
  for (let i = 0; i < recipients.length; i += BATCH_SIZE) {
    const batch = recipients.slice(i, Math.min(i + BATCH_SIZE, recipients.length));
    const amounts = new Array(batch.length).fill(amount);

    console.log(`\n处理批次 ${Math.floor(i / BATCH_SIZE) + 1}/${Math.ceil(recipients.length / BATCH_SIZE)}`);
    console.log(`地址范围: ${i + 1} - ${i + batch.length}`);

    try {
      // 调用 airdropERC20 方法
      const tx = await contract.airdropERC20(batch, amounts);
      console.log(`交易已发送: ${tx.hash}`);

      const receipt = await tx.wait();
      console.log(`交易确认! Gas 使用: ${receipt?.gasUsed.toString()}`);

      // 等待一段时间避免 nonce 问题
      if (i + BATCH_SIZE < recipients.length) {
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
    } catch (error: any) {
      console.error(`批次处理失败:`, error.message);
      throw error;
    }
  }

  console.log('\n✅ ERC20 空投完成!\n');
}

/**
 * BNB 空投
 * @param contract Airdrop 合约实例
 * @param recipients 接收地址数组
 * @param amount 每个地址接收的 BNB 数量（单位：BNB）
 */
async function airdropBNB(
  contract: Contract,
  recipients: string[],
  amount: bigint
): Promise<void> {
  console.log(`\n💰 开始 BNB 空投...`);
  console.log(`接收地址数: ${recipients.length}`);
  console.log(`每个地址: ${ethers.formatEther(amount)} BNB`);

  const totalAmount = amount * BigInt(recipients.length);
  console.log(`总空投量: ${ethers.formatEther(totalAmount)} BNB`);

  try {
    // 调用 airdropBNB 方法（需要发送 BNB）
    const tx = await contract.airdropBNB(recipients, new Array(recipients.length).fill(amount), {
      value: totalAmount
    });

    console.log(`交易已发送: ${tx.hash}`);

    const receipt = await tx.wait();
    console.log(`交易确认! Gas 使用: ${receipt?.gasUsed.toString()}`);

    console.log('\n✅ BNB 空投完成!\n');
  } catch (error: any) {
    console.error(`BNB 空投失败:`, error.message);
    throw error;
  }
}

/**
 * 主函数
 */
async function main() {
  console.log('====================================');
  console.log('     批量空投脚本启动');
  console.log('====================================\n');

  // 检查环境变量
  const privateKey = process.env.SEPOLIA_PRIVATE_KEY;
  if (!privateKey) {
    throw new Error('请在 .env 文件中配置 PRIVATE_KEY 或 SEPOLIA_PRIVATE_KEY');
  }

  // 初始化 Provider 和 Wallet
  const provider = new ethers.JsonRpcProvider(RPC_URL);
  const wallet = new Wallet(privateKey, provider);

  console.log(`发送者地址: ${wallet.address}`);

  // 检查余额
  const bnbBalance = await provider.getBalance(wallet.address);
  console.log(`BNB 余额: ${ethers.formatEther(bnbBalance)} BNB\n`);

  // 连接合约
  const airdropContract = new Contract(AIRDROP_CONTRACT_ADDRESS, airdropABI, wallet);
  const myErc20Contract = new Contract(MYERC20_CONTRACT_ADDRESS, myErc20ABI, wallet);

  // 检查代币余额
  const tokenBalance = await myErc20Contract.balanceOf(wallet.address);
  const tokenDecimals = await myErc20Contract.decimals();
  console.log(`代币余额: ${ethers.formatUnits(tokenBalance, tokenDecimals)} Tokens\n`);

  // ========== 功能 1: ERC20 空投到随机地址 ==========
  console.log('\n========== 功能 1: ERC20 空投 ==========');

  const erc20AddressCount = 10;

  const erc20Amount = parseUnits('1', tokenDecimals); // 1 个代币
  const totalTokensNeeded = erc20Amount * BigInt(erc20AddressCount);

  if (tokenBalance < totalTokensNeeded) {
    console.warn(`⚠️  警告: 代币余额不足!`);
    console.warn(`需要: ${ethers.formatUnits(totalTokensNeeded, tokenDecimals)} Tokens`);
    console.warn(`当前: ${ethers.formatUnits(tokenBalance, tokenDecimals)} Tokens`);
    console.warn(`跳过 ERC20 空投...\n`);
  } else {
    // 生成 1 万个随机地址
    console.log(`正在生成 ${erc20AddressCount} 个随机钱包地址...`);
    const randomAddresses = generateRandomAddresses(erc20AddressCount);
    console.log(`已生成 ${randomAddresses.length} 个地址\n`);
    console.log(`空投地址: ${randomAddresses.join(', ')}\n`);

    // 先授权 Airdrop 合约使用代币
    console.log('正在授权 Airdrop 合约...');
    const approveTx = await myErc20Contract.approve(AIRDROP_CONTRACT_ADDRESS, totalTokensNeeded);
    await approveTx.wait();
    console.log('授权完成!\n');

    // 执行空投（这可能需要很长时间）
    await airdropERC20(airdropContract, tokenDecimals, randomAddresses, erc20Amount);
  }

  // ========== 功能 2: BNB 空投到指定地址 ==========
  console.log('\n========== 功能 2: BNB 空投 ==========');

  // 指定两个钱包地址
  const bnbRecipients = [
    '0x3DA0E0Aa4Db801ca5161412513558E9096E288bA', // 替换为实际地址
    '0x24ca20D955cc2B6f722eCfdE9A7053435885dDe9'  // 替换为实际地址
  ];

  const bnbAmount = parseUnits('0.01', 18); // 0.01 BNB
  const totalBnbNeeded = bnbAmount * BigInt(bnbRecipients.length);

  console.log(`BNB 接收地址: ${bnbRecipients.join(', ')}`);

  if (bnbBalance < totalBnbNeeded) {
    console.warn(`⚠️  警告: BNB 余额不足!`);
    console.warn(`需要: ${ethers.formatEther(totalBnbNeeded)} BNB`);
    console.warn(`当前: ${ethers.formatEther(bnbBalance)} BNB`);
    console.warn(`跳过 BNB 空投...\n`);
  } else {
    await airdropBNB(airdropContract, bnbRecipients, bnbAmount);
  }

  console.log('====================================');
  console.log('     所有空投操作完成!');
  console.log('====================================\n');
}

// 运行主函数
main().catch((error) => {
  console.error('程序执行出错:', error);
  process.exit(1);
});

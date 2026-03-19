import { ethers, Wallet, Contract, parseUnits, FeeData } from 'ethers';
import dotenv from 'dotenv';
import * as path from 'path';
import * as fs from 'fs';

dotenv.config();

// 导入 ABI
const myErc20ABI = JSON.parse(fs.readFileSync(path.join(__dirname, '../src/abis/MyERC20.json'), 'utf8'));

// 配置
const RPC_URL = process.env.SEPOLIA_RPC_URL || 'https://rpc.sepolia.org';
const PRIVATE_KEY = process.env.SEPOLIA_PRIVATE_KEY || '';
const ERC20_CONTRACT_ADDRESS = '0x3d720e33044bD6691A850cEcF03E5D1C9e25Eda3'

// BSC 测试网配置
const BSC_TESTNET_RPC = process.env.BSC_TESTNET_RPC || 'https://bsc-testnet-dataseed.bnbchain.org';

/**
 * 转账配置接口
 * Base Fee = 30 Gwei（网络基础费，会被燃烧）
 * maxPriorityFeePerGas = 2 Gwei（给矿工的小费）
 * maxFeePerGas = 50 Gwei（愿意支付的上限）
 // 实际支付 = 30 + 2 = 32 Gwei（< 50，交易会执行）
 */
interface TransferConfig {
  gasLimitPadding: number; // Gas Limit 额外百分比（例如 20 表示 20%）
  maxPriorityFeePerGas?: bigint; // EIP-1559 小费上限
  maxFeePerGas?: bigint; // EIP-1559 总费用上限
}

/**
 * 获取最优 Gas 价格（支持 EIP-1559）
 */
async function getOptimalGasPrice(provider: ethers.JsonRpcProvider, config: TransferConfig = { gasLimitPadding: 20 }) {
  try {
    const feeData: FeeData = await provider.getFeeData();

    console.log('feeData', feeData);

    console.log('\n📊 当前 Gas 价格信息:');
    console.log(`Gas Price: ${feeData.gasPrice ? ethers.formatUnits(feeData.gasPrice, 'gwei') : 'N/A'} Gwei`);
    console.log(`Max Fee Per Gas: ${feeData.maxFeePerGas ? ethers.formatUnits(feeData.maxFeePerGas, 'gwei') : 'N/A'} Gwei`);
    console.log(`Max Priority Fee Per Gas: ${feeData.maxPriorityFeePerGas ? ethers.formatUnits(feeData.maxPriorityFeePerGas, 'gwei') : 'N/A'} Gwei`);

    // 检查是否支持 EIP-1559
    if (feeData.maxFeePerGas && feeData.maxPriorityFeePerGas) {
      console.log('✅ 使用 EIP-1559 动态费用');

      // 使用配置的值或默认值
      const maxFeePerGas = config.maxFeePerGas || feeData.maxFeePerGas;
      const maxPriorityFeePerGas = config.maxPriorityFeePerGas || feeData.maxPriorityFeePerGas;

      return {
        type: 2, // EIP-1559 交易类型
        maxFeePerGas,
        maxPriorityFeePerGas,
      };
    } else {
      console.log('⚠️  回退到传统 Gas Price (Legacy)');
      return {
        type: 0, // Legacy 交易类型
        gasPrice: feeData.gasPrice,
      };
    }
  } catch (error) {
    console.error('❌ 获取 Gas 价格失败:', error);
    throw error;
  }
}

/**
 * 精确估算 Gas 并添加安全边距
 */
async function estimateGasWithPadding(
  transaction: { to: string; from?: string; data?: string; value?: bigint },
  contract: Contract,
  method: string,
  args: any[],
  provider: ethers.JsonRpcProvider,
  paddingPercent: number = 20
): Promise<bigint> {

  try {
    // 精确估算 Gas
    let estimatedGas: bigint;

    if (contract && method) {
      console.log(`🔍 估算 ${method} 方法 Gas...`);
      estimatedGas = await contract[method].estimateGas(...args);
    } else {
      console.log('🔍 估算普通转账 Gas...');
      estimatedGas = await provider.estimateGas(transaction);
    }

    console.log(`估算 Gas: ${estimatedGas.toString()}`);

    // 添加安全边距（避免边缘情况导致失败）
    const padding = (estimatedGas * BigInt(paddingPercent)) / 100n;
    const gasLimit = estimatedGas + padding;

    console.log(`添加 ${paddingPercent}% 边距后: ${gasLimit.toString()}`);
    return gasLimit;
  } catch (error: any) {
    console.error('❌ Gas 估算失败:', error.message);
    throw error;
  }
}

/**
 * ERC20 代币转账
 */
async function transferERC20(
  contractAddress: string,
  recipientAddress: string,
  amount: string,
  provider: ethers.JsonRpcProvider,
  wallet: Wallet,
  config: TransferConfig = { gasLimitPadding: 20 }
) {
  console.log('\n========== ERC20 转账 ==========');
  console.log(`合约地址: ${contractAddress}`);
  console.log(`接收地址: ${recipientAddress}`);
  console.log(`转账数量: ${amount} Tokens`);

  try {
    // 创建合约实例
    const erc20Contract = new Contract(contractAddress, myErc20ABI, wallet);

    // 获取 decimals
    const decimals: number = await erc20Contract.decimals();
    console.log(`代币精度: ${decimals}`);

    // 转换数量为最小单位
    const amountInWei = parseUnits(amount, decimals);
    console.log(`转换后数量: ${amountInWei.toString()}`);

    // 查询余额
    const balance = await erc20Contract.balanceOf(wallet.address);
    console.log(`当前余额: ${ethers.formatUnits(balance, decimals)} Tokens`);

    if (balance < amountInWei) {
      throw new Error(`余额不足！需要: ${amount}, 拥有: ${ethers.formatUnits(balance, decimals)}`);
    }

    // 获取最优 Gas 价格
    const gasPriceOptions = await getOptimalGasPrice(provider, config);

    // 精确估算 Gas
    const gasLimit = await estimateGasWithPadding(
      { to: contractAddress, from: wallet.address },
      erc20Contract,
      'transfer',
      [recipientAddress, amountInWei],
      provider,
      config.gasLimitPadding
    );

    // 构建交易
    console.log('\n📝 构建交易...');
    const tx = await erc20Contract.transfer.populateTransaction(recipientAddress, amountInWei);

    // 发送交易
    console.log('⏳ 发送交易...');
    const transaction = await wallet.sendTransaction({
      ...tx,
      ...gasPriceOptions,
      gasLimit,
    });

    console.log(`✅ 交易已发送! Hash: ${transaction.hash}`);
    console.log(`📋 交易详情:`);
    console.log(`   Gas Limit: ${gasLimit.toString()}`);
    console.log(`   Max Fee: ${gasPriceOptions.maxFeePerGas ? ethers.formatUnits(gasPriceOptions.maxFeePerGas, 'gwei') + ' Gwei' : ethers.formatUnits(gasPriceOptions.gasPrice!, 'gwei') + ' Gwei'}`);

    // 等待确认
    console.log('⏳ 等待交易确认...');
    const receipt = await transaction.wait();

    console.log(`\n✅ 交易确认! 区块号: ${receipt?.blockNumber}`);
    console.log(`Gas 使用: ${receipt?.gasUsed.toString()}`);
    console.log(`实际费用: ${receipt?.gasUsed ? ethers.formatEther((receipt.gasUsed * (gasPriceOptions.gasPrice || gasPriceOptions.maxFeePerGas!))) : 'N/A'} ETH`);
    console.log('===============================\n');

    return receipt;
  } catch (error: any) {
    console.error('❌ ERC20 转账失败:', error.message);
    throw error;
  }
}

/**
 * BNB/ETH 原生代币转账
 */
async function transferNativeCoin(
  recipientAddress: string,
  amount: string,
  provider: ethers.JsonRpcProvider,
  wallet: Wallet,
  config: TransferConfig = { gasLimitPadding: 20 }
) {
  console.log('\n========== BNB/ETH 转账 ==========');
  console.log(`接收地址: ${recipientAddress}`);
  console.log(`转账数量: ${amount} BNB/ETH`);

  try {
    // 转换数量
    const amountInWei = parseUnits(amount, 18);
    console.log(`转换后数量: ${amountInWei.toString()} wei`);

    // 查询余额
    const balance = await provider.getBalance(wallet.address);
    console.log(`当前余额: ${ethers.formatEther(balance)} BNB/ETH`);

    if (balance < amountInWei) {
      throw new Error(`余额不足！需要: ${amount}, 拥有: ${ethers.formatEther(balance)}`);
    }

    // 获取最优 Gas 价格
    const gasPriceOptions = await getOptimalGasPrice(provider, config);

    // 精确估算 Gas（原生转账没有合约）
    const gasLimit = await estimateGasWithPadding(
      {
        to: recipientAddress,
        from: wallet.address,
        value: amountInWei,
      },
      null as any,
      '',
      [],
      provider,
      config.gasLimitPadding
    );

    // 构建交易
    console.log('\n📝 构建交易...');
    const tx = {
      to: recipientAddress,
      value: amountInWei,
    };

    // 发送交易
    console.log('⏳ 发送交易...');
    const transaction = await wallet.sendTransaction({
      ...tx,
      ...gasPriceOptions,
      gasLimit,
    });

    console.log(`✅ 交易已发送! Hash: ${transaction.hash}`);
    console.log(`📋 交易详情:`);
    console.log(`   Gas Limit: ${gasLimit.toString()}`);
    console.log(`   Max Fee: ${gasPriceOptions.maxFeePerGas ? ethers.formatUnits(gasPriceOptions.maxFeePerGas, 'gwei') + ' Gwei' : ethers.formatUnits(gasPriceOptions.gasPrice!, 'gwei') + ' Gwei'}`);

    // 等待确认
    console.log('⏳ 等待交易确认...');
    const receipt = await transaction.wait();

    console.log(`\n✅ 交易确认! 区块号: ${receipt?.blockNumber}`);
    console.log(`Gas 使用: ${receipt?.gasUsed.toString()}`);
    console.log(`实际费用: ${receipt?.gasUsed ? ethers.formatEther(receipt.gasUsed * (gasPriceOptions.gasPrice || gasPriceOptions.maxFeePerGas!)) : 'N/A'} BNB/ETH`);
    console.log('==================================\n');

    return receipt;
  } catch (error: any) {
    console.error('❌ BNB/ETH 转账失败:', error.message);
    throw error;
  }
}

/**
 * 主函数
 */
async function main() {
  if (!PRIVATE_KEY) {
    throw new Error('❌ 请在 .env 文件中设置 SEPOLIA_PRIVATE_KEY');
  }

  // 连接到 BSC 测试网
  console.log('🔗 连接到 BSC 测试网...');
  const provider = new ethers.JsonRpcProvider(BSC_TESTNET_RPC);
  const wallet = new Wallet(PRIVATE_KEY, provider);

  console.log(`🔑 钱包地址: ${wallet.address}`);

  // 查询钱包余额
  const balance = await provider.getBalance(wallet.address);
  console.log(`💰 BNB 余额: ${ethers.formatEther(balance)} BNB\n`);

  // 配置 Gas 策略
  const transferConfig: TransferConfig = {
    gasLimitPadding: 20, // Gas Limit 增加 20% 安全边距
    maxPriorityFeePerGas: parseUnits('2', 'gwei'), // 可选：设置小费上限
    maxFeePerGas: parseUnits('50', 'gwei'), // 可选：设置总费用上限
  };

  // ==================== 示例 1: ERC20 转账 ====================
  if (ERC20_CONTRACT_ADDRESS) {
    try {
      const recipient = '0x3DA0E0Aa4Db801ca5161412513558E9096E288bA'; // 替换为实际接收地址
      const amount = '1'; // 转账 1 个代币

      await transferERC20(
        ERC20_CONTRACT_ADDRESS,
        recipient,
        amount,
        provider,
        wallet,
        transferConfig
      );
    } catch (error) {
      console.error('ERC20 转账示例失败，跳过...\n');
    }
  } else {
    console.log('⚠️  未设置 MY_ERC20_ADDRESS，跳过 ERC20 转账\n');
  }

  // ==================== 示例 2: BNB/ETH 转账 ====================
  try {
    // 如果要使用 BSC 测试网，切换 provider
    // const bscProvider = new ethers.JsonRpcProvider(BSC_TESTNET_RPC);
    // const bscWallet = new Wallet(PRIVATE_KEY, bscProvider);

    const recipient = '0x3DA0E0Aa4Db801ca5161412513558E9096E288bA'; // 替换为实际接收地址
    const amount = '0.01'; // 转账 0.001 ETH

    await transferNativeCoin(
      recipient,
      amount,
      provider,
      wallet,
      transferConfig
    );
  } catch (error: any) {
    console.error('BNB/ETH 转账失败:', error.message);
  }
}

// 执行主函数
main().catch((error) => {
  console.error('💥 程序异常:', error);
  process.exit(1);
});

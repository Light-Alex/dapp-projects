const { ethers } = require('ethers');
const config = require('./config');
const { AppDataSource } = require('./database');
const Account = require('./entity/Account');
const Withdrawal = require('./entity/Withdrawal');
const { ERC20_ABI } = require('./blockchainParser');


class WithdrawalService {
  constructor() {
    this.provider = new ethers.JsonRpcProvider(config.blockchain.rpcUrl);
    this.wallet = new ethers.Wallet(process.env.SEPOLIA_PRIVATE_KEY, this.provider);
    this.usdtContract = new ethers.Contract(config.blockchain.usdtAddress, ERC20_ABI, this.wallet);
  }

  // 处理提现请求
  async processWithdrawals() {
    const withdrawalRepo = AppDataSource.getRepository(Withdrawal);
    const pendingWithdrawals = await withdrawalRepo.find({
      where: { status: 'init' },
      relations: ['account']
    });

    for (const withdrawal of pendingWithdrawals) {
      // 更新状态为处理中
      if (withdrawal.status !== 'init') {
        continue;
      }

      withdrawal.status = 'processing';
      await withdrawalRepo.save(withdrawal);
    }

    for (const withdrawal of pendingWithdrawals) {
      await this.processWithdrawal(withdrawal);
    }
  }

  // 处理单个提现
  async processWithdrawal(withdrawal) {
    const withdrawalRepo = AppDataSource.getRepository(Withdrawal);
    const accountRepo = AppDataSource.getRepository(Account);

    try {
      if (withdrawal.status !== 'processing') {
        return;
      }

      console.log(`Processing withdrawal ${withdrawal.id}...`);

      let txHash;

      // BNB 提现
      if (withdrawal.token_symbol === 'BNB') {
        txHash = await this.withdrawBNB(withdrawal);
      }
      // USDT 提现
      else if (withdrawal.token_symbol === 'USDT') {
        txHash = await this.withdrawUSDT(withdrawal);
      }

      // 更新提现记录
      withdrawal.tx_hash = txHash;
      await withdrawalRepo.save(withdrawal);

      // 监听交易结果
      await this.monitorTransaction(txHash, withdrawal);

    } catch (error) {
      console.error(`Withdrawal ${withdrawal.id} failed:`, error);
      withdrawal.status = 'failed';
      await withdrawalRepo.save(withdrawal);
    }
  }

  // 提现 BNB
  async withdrawBNB(withdrawal) {
    const tx = await this.wallet.sendTransaction({
      to: withdrawal.to_address,
      value: ethers.parseEther(withdrawal.amount.toString()),
      gasLimit: 21000
    });

    return tx.hash;
  }

  // 提现 USDT
  async withdrawUSDT(withdrawal) {
    const decimals = await this.usdtContract.decimals();
    const amount = ethers.parseUnits(withdrawal.amount.toString(), decimals);
    const tx = await this.usdtContract.transfer(withdrawal.to_address, amount);
    return tx.hash;
  }

  // 监听交易结果
  async monitorTransaction(txHash, withdrawal) {
    const withdrawalRepo = AppDataSource.getRepository(Withdrawal);
    const accountRepo = AppDataSource.getRepository(Account);
    
    try {
      const receipt = await this.provider.waitForTransaction(txHash, 3); // 等待3个确认
      
      if (receipt.status === 1) {
        // 提现成功
        withdrawal.status = 'success';
        await withdrawalRepo.save(withdrawal);
        
        // 这里不需要更新用户余额，因为提现时已经扣减
        console.log(`Withdrawal ${withdrawal.id} succeeded`);
      } else {
        // 提现失败，回滚用户余额
        withdrawal.status = 'failed';
        await withdrawalRepo.save(withdrawal);
        
        // 回滚用户余额
        if (withdrawal.token_symbol === 'BNB') {
          await accountRepo.increment(
            { id: withdrawal.account_id },
            'bnb_amount',
            withdrawal.amount
          );
        } else if (withdrawal.token_symbol === 'USDT') {
          await accountRepo.increment(
            { id: withdrawal.account_id },
            'usdt_amount',
            withdrawal.amount
          );
        }
        
        console.log(`Withdrawal ${withdrawal.id} failed, rolled back balance`);
      }
    } catch (error) {
      console.error(`Error monitoring transaction ${txHash}:`, error);
      withdrawal.status = 'failed';
      await withdrawalRepo.save(withdrawal);
    }
  }
}

module.exports = WithdrawalService;
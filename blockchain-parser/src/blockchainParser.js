const { ethers } = require('ethers');
const config = require('./config');
const { AppDataSource } = require('./database');
const redisService = require('./redis');
const Account = require('./entity/Account');
const Transaction = require('./entity/Transaction');
const Withdrawal = require('./entity/Withdrawal');

// ERC20 ABI
const ERC20_ABI = [
  "event Transfer(address indexed from, address indexed to, uint256 value)",
  "function transfer(address to, uint256 amount) returns (bool)",
  "function balanceOf(address account) view returns (uint256)",
  "function decimals() view returns (uint8)",
  "function symbol() view returns (string)"
];

class BlockchainParser {
  constructor() {
    this.provider = new ethers.JsonRpcProvider(config.blockchain.rpcUrl);
    this.usdtContract = new ethers.Contract(config.blockchain.usdtAddress, ERC20_ABI, this.provider);
    this.projectAddress = config.blockchain.projectAddress.toLowerCase();
    this.isProcessing = false;
  }

  // 判断是否为合约地址
  async isContractAddress(address) {
    if (!address) return false;
    const code = await this.provider.getCode(address);
    return code !== '0x';
  }

  // 更新或创建用户账户
  async updateOrCreateUser(accountRepo, address, email, bnbAmount, usdtAmount) {
    let user = await accountRepo.findOne({ where: { address } });

    if (user) {
      user.bnb_amount = bnbAmount;
      user.usdt_amount = usdtAmount;
      await accountRepo.save(user);
    } else {
      const newUser = accountRepo.create({
        email,
        address,
        bnb_amount: bnbAmount,
        usdt_amount: usdtAmount
      });
      await accountRepo.save(newUser);
    }
  }

  // 开始初始化pg数据库中Account表和withdrawal表
  async initData() {
    console.log('Initializing database...');

    const user1Address = '0x6687e46C68C00bd1C10F8cc3Eb000B1752737e94';
    const user2Address = '0x3DA0E0Aa4Db801ca5161412513558E9096E288bA';
    const user3Address = '0x24ca20D955cc2B6f722eCfdE9A7053435885dDe9';

    // 初始化Account表
    const accountRepo = AppDataSource.getRepository(Account);

    // 查找地址
    let user1 = await accountRepo.findOne({ where: { address: user1Address } });
    let user2 = await accountRepo.findOne({ where: { address: user2Address } });
    let user3 = await accountRepo.findOne({ where: { address: user3Address } });

    // 获取 USDT 精度
    const decimal = await this.usdtContract.decimals();

    // 初始化 user1
    const user1Balance = await this.provider.getBalance(user1Address);
    const user1BalanceInBNB = ethers.formatEther(user1Balance);
    const user1UsdtBalance = await this.usdtContract.balanceOf(user1Address);
    const user1UsdtBalanceInUSDT = ethers.formatUnits(user1UsdtBalance, decimal);
    await this.updateOrCreateUser(accountRepo, user1Address, 'user1@example.com', user1BalanceInBNB, user1UsdtBalanceInUSDT);

    // 初始化 user2
    const user2Balance = await this.provider.getBalance(user2Address);
    const user2BalanceInBNB = ethers.formatEther(user2Balance);
    const user2UsdtBalance = await this.usdtContract.balanceOf(user2Address);
    const user2UsdtBalanceInUSDT = ethers.formatUnits(user2UsdtBalance, decimal);
    await this.updateOrCreateUser(accountRepo, user2Address, 'user2@example.com', user2BalanceInBNB, user2UsdtBalanceInUSDT);

    // 初始化 user3
    const user3Balance = await this.provider.getBalance(user3Address);
    const user3BalanceInBNB = ethers.formatEther(user3Balance);
    const user3UsdtBalance = await this.usdtContract.balanceOf(user3Address);
    const user3UsdtBalanceInUSDT = ethers.formatUnits(user3UsdtBalance, decimal);
    await this.updateOrCreateUser(accountRepo, user3Address, 'user3@example.com', user3BalanceInBNB, user3UsdtBalanceInUSDT);

    console.log('Initialized 3 account records');
    
    // 初始化withdrawal表
    const withdrawalRepo = AppDataSource.getRepository(Withdrawal);
    const withdrawalCount = await withdrawalRepo.count();
    if (withdrawalCount === 0) {
        // 获取用户账户
        user1 = await accountRepo.findOne({ where: { address: user1Address } });
        user2 = await accountRepo.findOne({ where: { address: user2Address } });

        if (user1 && user2) {
            // 为 user1 创建 BNB 提现记录
            const withdrawal1 = withdrawalRepo.create({
                account_id: user1.id,
                amount: 0.0001,
                token_decimals: 18,
                token_symbol: 'BNB',
                to_address: user3Address,
                status: 'init'
            });
            await withdrawalRepo.save(withdrawal1);

            // 为 user2 创建 USDT 提现记录
            const withdrawal2 = withdrawalRepo.create({
                account_id: user2.id,
                amount: 80,
                token_decimals: 18,
                token_symbol: 'USDT',
                to_address: user3Address,
                status: 'init'
            });
            await withdrawalRepo.save(withdrawal2);
        }
    }else{
        // 将所有数据的status设置为init
        await withdrawalRepo.query("UPDATE withdrawal SET status = 'init'");
    }

    console.log('Initialized 2 withdrawal records');

  }

  // 开始解析区块链
  async start() {
    console.log('Starting blockchain parser...');
    
    // 初始化最后处理的区块
    let lastBlock = await redisService.getLastProcessedBlock();
    if (!lastBlock) {
      lastBlock = await this.provider.getBlockNumber() - 100; // 从100个区块前开始
      await redisService.setLastProcessedBlock(lastBlock);
    }

    setInterval(() => this.processBlocks(), config.app.scanInterval);
  }

  // 处理新区块
  async processBlocks() {
    if (this.isProcessing) return;
    
    this.isProcessing = true;
    try {
      const currentBlock = await this.provider.getBlockNumber();
      let lastBlock = parseInt(await redisService.getLastProcessedBlock());
      
      // 只处理已确认的区块
      const confirmedBlock = currentBlock - config.app.confirmationBlocks;
      if (lastBlock >= confirmedBlock) {
        return;
      }

      console.log(`Processing blocks from ${lastBlock + 1} to ${confirmedBlock}`);
      
      for (let blockNumber = lastBlock + 1; blockNumber <= confirmedBlock; blockNumber++) {
        await this.processBlock(blockNumber);
        await redisService.setLastProcessedBlock(blockNumber);
      }
    } catch (error) {
      console.error('Error processing blocks:', error);
    } finally {
      this.isProcessing = false;
    }
  }

  // 处理单个区块
  async processBlock(blockNumber) {
    try {
      // blockNumber：区块号
      // prefetchTxn
      const block = await this.provider.getBlock(blockNumber, true);
      
      for (const tx of block.prefetchedTransactions) {        
        // 检查是否是与相关方相关的交易
         const isFromProject = tx.from.toLowerCase() === this.projectAddress;
         const isToProject = tx.to && tx.to.toLowerCase() === this.projectAddress;
         const isToContract = tx.to && await tx.to.toLowerCase() === config.blockchain.usdtAddress.toLowerCase();
        //  const isToContract = tx.to && await this.isContractAddress(tx.to);

        if (isFromProject || isToProject || isToContract) {
          await this.processTransaction(tx);
        }
      }
    } catch (error) {
      console.error(`Error processing block ${blockNumber}:`, error);
    }
  }

  // 处理交易
  async processTransaction(tx) {
    // 检查是否正在处理
    if (await redisService.isTxProcessing(tx.hash)) {
      return;
    }

    await redisService.setTxProcessing(tx.hash);

    try {
      const receipt = await this.provider.getTransactionReceipt(tx.hash);

      const gasFee = receipt.fee;
      const gasFeeBNB = ethers.formatEther(gasFee);
      const fromUser = await this.findUserByAddress(tx.from);
      if (!fromUser) {
        return;
      }

      const accountRepo = AppDataSource.getRepository(Account);
      await accountRepo.decrement(
        { id: fromUser.id },
        'bnb_amount',
        gasFeeBNB
      );

      console.log(`Updated BNB balance for user ${fromUser.id} (${tx.from}): -${gasFeeBNB} BNB (gas)`);

      // 跳过合约创建交易
      if (!tx.to) return;

      // BNB 转账
      if (tx.data === '0x' && tx.value > 0n) {
        await this.processBNBTransfer(tx, receipt);
      }
      // ERC20 转账
      else if (tx.data !== '0x') {
        await this.processERC20Transfer(tx, receipt);
      }
    } catch (error) {
      console.error(`Error processing transaction ${tx.hash}:`, error);
    } finally {
      await redisService.removeTxProcessing(tx.hash);
    }
  }

  // 处理 BNB 转账
  async processBNBTransfer(tx, receipt) {
    // const amount = ethers.formatEther(tx.value);
    const amount = tx.value;
    const amountBNB = ethers.formatEther(amount);
    
    // 查找用户
    const fromUser = await this.findUserByAddress(tx.from);
    if (!fromUser) {
      console.log(`No from user found for address ${tx.from}`);
      return;
    }

    const toUser = await this.findUserByAddress(tx.to);
    if (!toUser) {
      console.log(`No to user found for address ${tx.to}`);
      return;
    }

    // 保存交易记录
    const transactionRepo = AppDataSource.getRepository(Transaction);
    const existingTx = await transactionRepo.findOne({ where: { tx_hash: tx.hash } });
    
    if (existingTx) {
      return;
    }

    const newTransaction = transactionRepo.create({
      tx_hash: tx.hash,
      block_number: tx.blockNumber,
      from_address: tx.from,
      to_address: tx.to,
      value: amountBNB,
      token_decimals: 18,
      token_symbol: 'BNB',
      status: receipt.status
    });

    await transactionRepo.save(newTransaction);

    // 更新用户余额
    if (receipt.status === 1) {
      const accountRepo = AppDataSource.getRepository(Account);
      await accountRepo.decrement(
        { id: fromUser.id },
        'bnb_amount',
        amountBNB
      );
      console.log(`Updated BNB balance for user ${fromUser.id} ${fromUser.address}: -${amountBNB} BNB`);

      await accountRepo.increment(
        { id: toUser.id },
        'bnb_amount',
        amountBNB
      );
      console.log(`Updated BNB balance for user ${toUser.id} ${toUser.address}: +${amountBNB} BNB`);
    }
  }

  // 处理 ERC20 转账
  async processERC20Transfer(tx, receipt) {
    try {
      // 获取token decimal
      const decimal = await this.usdtContract.decimals();

      // 获取token symbol
      const symbol = await this.usdtContract.symbol();

      // 解析 Transfer 事件
      for (const log of receipt.logs) {
        if (log.address.toLowerCase() === config.blockchain.usdtAddress.toLowerCase()) {
          try {
            const event = this.usdtContract.interface.parseLog(log);
            if (event.name === 'Transfer' && (event.args.from.toLowerCase() === this.projectAddress || event.args.to.toLowerCase() === this.projectAddress)) {
            //   const amount = ethers.formatUnits(event.args.value, decimal); // USDT 精度为6
              const amount = event.args.value;
              const amountUSDT = ethers.formatUnits(amount, decimal);

              // 查找用户
              const fromUser = await this.findUserByAddress(event.args.from);
              if (!fromUser) {
                console.log(`No from user found for address ${event.args.from}`);
                return;
              }

              const toUser = await this.findUserByAddress(event.args.to);
              if (!toUser) {
                console.log(`No to user found for address ${event.args.to}`);
                return;
              }

              // 保存交易记录
              const transactionRepo = AppDataSource.getRepository(Transaction);
              const existingTx = await transactionRepo.findOne({ where: { tx_hash: tx.hash } });

              if (existingTx) {
                console.log(`Transaction ${tx.hash} already processed, skipping...`);
                return;
              }

              const newTransaction = transactionRepo.create({
                tx_hash: tx.hash,
                block_number: tx.blockNumber,
                from_address: event.args.from,
                to_address: event.args.to,
                value: amountUSDT,
                token_decimals: Number(decimal),
                token_address: event.address,
                token_symbol: symbol,
                status: receipt.status
              });

              // 先保存交易记录，成功后再更新余额
              await transactionRepo.save(newTransaction);

              // 更新用户余额
              if (receipt.status === 1) {
                const accountRepo = AppDataSource.getRepository(Account);
                await accountRepo.decrement(
                  { id: fromUser.id },
                  'usdt_amount',
                  amountUSDT
                );
                console.log(`Updated ${symbol} balance for user ${fromUser.id} ${fromUser.address}: -${amountUSDT} ${symbol}`);

                await accountRepo.increment(
                  { id: toUser.id },
                  'usdt_amount',
                  amountUSDT
                );
                console.log(`Updated ${symbol} balance for user ${toUser.id} ${toUser.address}: +${amountUSDT} ${symbol}`);
              }
            }
          } catch (error) {
            console.log(`Failed to parse log:`);
            console.log('Error name:', error.name);
            console.log('Error message:', error.message);
            console.log('Error detail:', error.detail);
            console.log('Error query:', error.query);
            console.log('Error parameters:', error.parameters);
            console.log('Full error:', error);
          }
        }
      }
    } catch (error) {
      console.error(`Error processing ERC20 transfer ${tx.hash}:`, error);
    }
  }

  // 根据地址查找用户
  async findUserByAddress(address) {
    // 先检查缓存
    const cachedUserId = await redisService.getCachedUserByAddress(address);
    if (cachedUserId) {
      return { id: parseInt(cachedUserId), address };
    }

    // 查询数据库
    const accountRepo = AppDataSource.getRepository(Account);
    const user = await accountRepo.findOne({ where: { address } });
    
    if (user) {
      await redisService.cacheUserAddress(address, user.id);
    }
    
    return user;
  }
}

module.exports = BlockchainParser;
module.exports.ERC20_ABI = ERC20_ABI;
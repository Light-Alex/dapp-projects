require('@babel/register');
// require('reflect-metadata');

const { AppDataSource } = require('./database');
const BlockchainParser = require('./blockchainParser');
const WithdrawalService = require('./withdrawalService');

class App {
  constructor() {
    this.blockchainParser = new BlockchainParser();
    this.withdrawalService = new WithdrawalService();
  }

  async start() {
    try {
      // 初始化数据库连接
      await AppDataSource.initialize();
      console.log('Database connected');
      
      // 初始化数据
      await this.blockchainParser.initData();

      // 启动区块链解析
      await this.blockchainParser.start();

      // 启动提现处理
      setInterval(() => this.withdrawalService.processWithdrawals(), 5000);

      console.log('Application started successfully');
    } catch (error) {
      console.error('Failed to start application:', error);
      process.exit(1);
    }
  }
}

// 启动应用
const app = new App();
app.start();

// 优雅关闭
process.on('SIGINT', async () => {
  console.log('Shutting down...');
  process.exit(0);
});
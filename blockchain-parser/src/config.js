require('dotenv').config();

module.exports = {
  // 区块链配置
  blockchain: {
    rpcUrl: process.env.BSC_RPC_URL || 'https://bsc-dataseed.bnbchain.org',
    chainId: process.env.CHAIN_ID || 56,
    // 项目方EOA账户地址（用于跟踪）
    projectAddress: process.env.PROJECT_ADDRESS || '0x3Ca1392e4A95Aa0f83e97458Ab4495a58cA91bd6',
    usdtAddress: process.env.USDT_ADDRESS || '0x55d398326f99059fF775485246999027B3197955'
  },

  // 数据库配置
  database: {
    host: process.env.DB_HOST || '172.19.62.197',
    port: process.env.DB_PORT || 5432,
    username: process.env.DB_USERNAME || 'postgres',
    password: process.env.DB_PASSWORD || 'postgres',
    database: process.env.DB_NAME || 'blockchain_parser'
  },

  // Redis配置
  redis: {
    host: process.env.REDIS_HOST || '172.19.62.197',
    port: process.env.REDIS_PORT || 6379,
    password: process.env.REDIS_PASSWORD || ''
  },

  // 应用配置
  app: {
    scanInterval: process.env.SCAN_INTERVAL || 3000, // 扫描间隔(ms)
    confirmationBlocks: process.env.CONFIRMATION_BLOCKS || 6 // 确认区块数
  }
};
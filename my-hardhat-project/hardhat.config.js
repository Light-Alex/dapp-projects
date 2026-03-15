require("@nomicfoundation/hardhat-toolbox");
require("dotenv").config();
require("hardhat-gas-reporter");
require("hardhat-abi-exporter");
require("solidity-coverage");
require('@openzeppelin/hardhat-upgrades');
require("./scripts/ignition-deploy-with-save");  // 引入 Ignition 部署任务

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    version: "0.8.28",
    settings: {
      optimizer: {//启用优化器降低Gas
        enabled: true,
        runs: 200
      }
    }
  },
  networks: {
    // 本地开发网络配置 hardhat node
    localhost: {
      url: "http://127.0.0.1:8545",
      accounts: process.env.PRIVATE_KEY ? [process.env.PRIVATE_KEY] : []
    },
    // 测试网络配置
    sepolia: {
      url: process.env.ALCHEMY_SEPOLIA_URL || "",
      accounts: process.env.SEPOLIA_PRIVATE_KEY ? [process.env.SEPOLIA_PRIVATE_KEY] : []
    }
  },
  gasReporter: {  // 添加 gas 报告配置
    enabled: true,
    currency: "USD",
    coinmarketcap: process.env.COINMARKETCAP_API_KEY,  // 可选，用于获取 gas 价格
    gasPrice: 20,
    showTimeSpent: true,
    showMethodSig: true,
    outputFile: "",  // 可选，输出到文件
    noColors: true
  },
    abiExporter: {
    path: "./abi",                     // 输出目录
    runOnCompile: true,                // 编译时自动导出
    clear: true,                       // 每次导出前清空目录
    flat: true,                        // 扁平化输出
    only: [],                          // 仅导出指定合约，留空导出所有
    spacing: 2,                        // 缩进空格数
    format: "json",                    // 格式：json, minimal
    // format: "minimal",              // 最小格式（只保留函数签名）
  },
    coverage: {
    excludeContracts: ["Migrations"],  // 排除特定合约
    skipFiles: ["mocks/", "test/"],    // 排除特定文件夹
    measureStatementCoverage: true,    // 语句覆盖率
    measureFunctionCoverage: true,     // 函数覆盖率
    measureBranchCoverage: true        // 分支覆盖率
  },
  // Etherscan验证配置
  etherscan: {
    apiKey: process.env.ETHERSCAN_API_KEY || ""
  }
};

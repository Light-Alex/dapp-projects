require("@nomicfoundation/hardhat-toolbox");
require("@openzeppelin/hardhat-upgrades");
// config
const { config: dotenvConfig } = require("dotenv");
const { resolve } = require("path");
dotenvConfig({ path: resolve(__dirname, "./.env") });

const SEPOLIA_PK = process.env.SEPOLIA_PK;
if (!SEPOLIA_PK) {
  throw new Error("Please set your SEPOLIA_PK in a .env file");
}

const MAINNET_PK = process.env.MAINNET_PK;
const MAINNET_ALCHEMY_AK = process.env.MAINNET_ALCHEMY_AK;

const SEPOLIA_ALCHEMY_AK = process.env.SEPOLIA_ALCHEMY_AK;
if (!SEPOLIA_ALCHEMY_AK) {
  throw new Error("Please set your SEPOLIA_ALCHEMY_AK in a .env file");
}

const BSC_TESTNET_PK = process.env.BSC_TESTNET_PK;
const BSC_TESTNET_ALCHEMY_AK = process.env.BSC_TESTNET_ALCHEMY_AK;
if (!BSC_TESTNET_ALCHEMY_AK) {
  throw new Error("Please set your BSC_TESTNET_ALCHEMY_AK in a .env file");
}
/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    version: "0.8.20",
    //revertStrings: "default", // 保留错误消息
    settings: {
      optimizer: {
        enabled: true,
        runs: 50, // 降低 runs 值以减少合约大小
      },
      viaIR: true, // 保持 IR 优化
    },
  },
  networks: {
    mainnet: {
      url: `https://eth-mainnet.g.alchemy.com/v2/${MAINNET_ALCHEMY_AK}`,
      // accounts: [MAINNET_PK],
      saveDeployments: true,
      chainId: 1,
    },
    sepolia: {
      url: `https://eth-sepolia.g.alchemy.com/v2/${SEPOLIA_ALCHEMY_AK}`,
      accounts: [SEPOLIA_PK],
      saveDeployments: true,
      chainId: 11155111,
    },

    bscTestnet: {
      url: `https://bnb-testnet.g.alchemy.com/v2/${BSC_TESTNET_ALCHEMY_AK}`,
      accounts: [BSC_TESTNET_PK],
      saveDeployments: true,
      chainId: 97,
      timeout: 180000, // 3 分钟超时
    },
    // optimism: {
    //   url: `https://rpc.ankr.com/optimism`,
    //   accounts: [`${MAINNET_PK}`],
    // },
  },
  gasReporter: {
    currency: "USD",
    enabled: process.env.REPORT_GAS ? true : false,
    excludeContracts: [],
    src: "./contracts",
  },
};

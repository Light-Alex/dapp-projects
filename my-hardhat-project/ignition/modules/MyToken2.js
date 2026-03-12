// scripts/deploy.js
const { ethers, network } = require("hardhat");
const fs = require("fs");
const path = require("path");

// 从环境变量获取部署参数
function getDeployParams() {
  return {
    tokenName: process.env.TOKEN_NAME || "My Token",
    tokenSymbol: process.env.TOKEN_SYMBOL || "MTK",
    initialSupply: process.env.INITIAL_SUPPLY ? parseInt(process.env.INITIAL_SUPPLY) : 1000000,
    decimals: process.env.TOKEN_DECIMALS ? parseInt(process.env.TOKEN_DECIMALS) : 18
  };
}

async function main() {
  console.log("开始部署合约...");

  // 获取部署参数
  const params = getDeployParams();
  console.log("部署参数:", params);

  // 获取部署者账户
  const [deployer] = await ethers.getSigners();
  console.log("部署者地址:", deployer.address);

  // 获取部署者余额
  const balance = await ethers.provider.getBalance(deployer.address);
  console.log("部署者余额:", ethers.formatEther(balance), "ETH");

  // 部署合约
  const MyToken = await ethers.getContractFactory("MyToken");
  const myToken = await MyToken.deploy(
    params.tokenName,
    params.tokenSymbol,
    params.initialSupply,
    params.decimals
  );

  console.log("合约部署中...");
  await myToken.waitForDeployment();

  const myTokenAddress = await myToken.getAddress();
  console.log("MyToken合约地址:", myTokenAddress);

  // 验证合约所有权
  const owner = await myToken.owner();
  console.log("合约所有者:", owner);

  // 验证代币信息
  const name = await myToken.name();
  const symbol = await myToken.symbol();
  const totalSupply = await myToken.totalSupply();
  
  console.log("代币名称:", name);
  console.log("代币符号:", symbol);
  console.log("总供应量:", ethers.formatUnits(totalSupply, 18));

  // 保存部署信息到文件
  const deploymentsDir = path.join(__dirname, "../deployments");
  if (!fs.existsSync(deploymentsDir)){
    fs.mkdirSync(deploymentsDir, { recursive: true });
  }
  
  const deploymentInfo = {
    network: network.name,
    timestamp: new Date().toISOString(),
    contract: {
      name: "MyToken",
      address: myTokenAddress,
      deployer: deployer.address,
      transactionHash: myToken.deploymentTransaction().hash,
      parameters: {
        name: params.tokenName,
        symbol: params.tokenSymbol,
        initialSupply: params.initialSupply,
        decimals: params.decimals
      }
    }
  };

  fs.writeFileSync(
    path.join(deploymentsDir, `${network.name}-MyToken.json`),
    JSON.stringify(deploymentInfo, null, 2)
  );

  console.log("部署完成!");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });


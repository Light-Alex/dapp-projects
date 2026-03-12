const { ethers, upgrades } = require("hardhat");

async function main() {
  const proxyAddress = "0x9c21e868B4eFA7444DdF968321d9F3EDd8BF6eEa";
  
  // 部署新的实现合约并升级代理
  const MyUpgradeableV2 = await ethers.getContractFactory("MyUpgradeableV2");
  await upgrades.upgradeProxy(proxyAddress, MyUpgradeableV2);
  
  console.log("合约已升级!");
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});

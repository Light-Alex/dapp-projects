// scripts/deploy_upgradeable.js
const { ethers, upgrades } = require("hardhat");

async function main() {
  // 部署逻辑合约和代理
  const MyUpgradeable = await ethers.getContractFactory("MyUpgradeableV1");

  // [42]是传递给 initializer 函数的参数数组
  const instance = await upgrades.deployProxy(MyUpgradeable, [42], {
    initializer: 'initialize',
  });
  await instance.waitForDeployment();
  
  proxyAddress = await instance.getAddress();
  contractAddress = await upgrades.erc1967.getImplementationAddress(proxyAddress);

  console.log("代理合约地址:", proxyAddress);
  console.log("实现合约地址:", contractAddress);

//   // 自动验证
//   console.log("Verifying contract...");
//   try {
//     await hre.run("verify:verify", {
//       address: proxyAddress,
//       constructorArguments: [42],
//     });
//     console.log("Contract verified!");
//   } catch (error) {
//     console.log("Verification failed or already verified");
//   }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
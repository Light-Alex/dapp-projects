const fs = require("fs");
const path = require("path");
const { createRequire } = require("module");

task("deploy-ignition", "Deploy using Ignition and save deployment info")
  .addParam("module", "Path to Ignition module")
  .addOptionalParam("verify", "Verify on Etherscan", false, types.boolean)
  .setAction(async (args, hre) => {
    const network = await hre.ethers.provider.getNetwork();
    const chainId = network.chainId.toString();

    console.log(`\n🚀 Deploying to ${hre.network.name} (Chain ID: ${chainId})...`);

    // 1. 执行 ignition deploy 使用编程式 API
    console.log("\n📦 Running Ignition deployment...");

    // 使用 require 加载 Ignition 模块（支持 CommonJS 格式）
    let module;
    try {
      const fullPath = path.resolve(process.cwd(), args.module);

      // 创建一个 require 函数来加载模块
      const requireFn = createRequire(fullPath);
      module = requireFn(fullPath);

      console.log("✅ Module loaded successfully");
    } catch (error) {
      console.error("❌ Failed to load module:", error.message);
      throw error;
    }

    try {
      const deployParams = {};

      await hre.ignition.deploy(module, deployParams);
      console.log("✅ Deployment completed successfully");
    } catch (error) {
      console.error("❌ Deployment failed:", error.message);
      throw error;
    }

    // 2. 部署完成后，保存信息
    console.log("\n📝 Saving deployment information...");

    const modulePath = args.module;
    const moduleName = path.basename(modulePath, ".js");

    // 读取 Ignition 生成的部署地址文件
    const deploymentDir = path.join(__dirname, "../ignition/deployments");
    const deployedAddressesFile = path.join(deploymentDir, `chain-${chainId}/deployed_addresses.json`);

    if (!fs.existsSync(deployedAddressesFile)) {
      console.log("⚠️  Deployment file not found at:", deployedAddressesFile);
      console.log("Please check if the deployment was successful.");
      return;
    }

    const deployedAddresses = JSON.parse(fs.readFileSync(deployedAddressesFile, "utf8"));
    console.log("✅ Found deployment addresses:", deployedAddresses);

    // 从 journal.jsonl 提取交易 hash 和构造函数参数
    const journalFile = path.join(deploymentDir, `chain-${chainId}/journal.jsonl`);
    let transactions = {};
    let constructorArgsMap = {};

    if (fs.existsSync(journalFile)) {
      const journalContent = fs.readFileSync(journalFile, "utf8");
      const lines = journalContent.trim().split("\n");

      for (const line of lines) {
        try {
          const entry = JSON.parse(line);
          // 从 TRANSACTION_CONFIRM 条目中提取交易 hash
          if (entry.type === "TRANSACTION_CONFIRM" && entry.hash) {
            const futureId = entry.futureId;
            transactions[futureId] = entry.hash;
          }
          // 从 DEPLOYMENT_EXECUTION_STATE_INITIALIZE 提取构造函数参数
          if (entry.type === "DEPLOYMENT_EXECUTION_STATE_INITIALIZE" && entry.constructorArgs) {
            const futureId = entry.futureId;
            constructorArgsMap[futureId] = entry.constructorArgs;
          }
        } catch (e) {
          // 跳过无法解析的行
        }
      }
      console.log("✅ Found transaction hashes:", transactions);
    } else {
      console.log("⚠️  Journal file not found at:", journalFile);
    }

    // 获取当前模块的名称（从模块对象中获取）
    const currentModuleName = module.id || moduleName;

    // 过滤出只属于当前模块的合约
    const currentModuleContracts = {};
    for (const [contractKey, address] of Object.entries(deployedAddresses)) {
      // 提取模块名（格式：ModuleName#ContractName）
      const moduleOfContract = contractKey.includes("#")
        ? contractKey.split("#")[0]
        : contractKey;

      // 只处理属于当前模块的合约
      if (moduleOfContract === currentModuleName || contractKey.startsWith(currentModuleName + "#")) {
        currentModuleContracts[contractKey] = address;
      }
    }

    console.log(`\n✅ Found ${Object.keys(currentModuleContracts).length} contract(s) in current module`);

    // 获取部署者地址
    const [deployer] = await hre.ethers.getSigners();

    // 保存到 deployments 目录
    const outputDir = path.join(__dirname, "../deployments");
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

    // 为每个合约单独保存文件
    const contracts = [];
    for (const [contractKey, address] of Object.entries(currentModuleContracts)) {
      // 提取合约名称（格式：ModuleName#ContractName）
      const contractName = contractKey.includes("#")
        ? contractKey.split("#")[1]
        : contractKey;

      // 构建单个合约的部署信息
      const deploymentInfo = {
        network: hre.network.name,
        chainId: chainId,
        timestamp: new Date().toISOString(),
        deployer: deployer.address,
        module: moduleName,
        contract: {
          name: contractName,
          address: address,
          transactionHash: transactions[contractKey] || "N/A"
        }
      };

      // 保存文件名使用 contractName
      const outputFile = path.join(outputDir, `${hre.network.name}-${chainId}-${contractName}.json`);
      fs.writeFileSync(outputFile, JSON.stringify(deploymentInfo, null, 2));

      contracts.push({ ...deploymentInfo.contract, file: outputFile, contractKey });
    }

    // 如果需要验证，进行合约验证
    if (args.verify) {
      console.log("\n🔍 Verifying contracts on Etherscan...");

      for (const contract of contracts) {
        let constructorArgs = [];

        try {
          console.log(`  Verifying ${contract.name} at ${contract.address}...`);

          // 从 constructorArgsMap 中获取构造函数参数
          constructorArgs = constructorArgsMap[contract.contractKey] || [];

          // 构建验证参数
          const verifyArgs = {
            address: contract.address,
            constructorArguments: constructorArgs
          };

          // 调用验证
          await hre.run("verify:verify", verifyArgs);
          console.log(`  ✅ ${contract.name} verified successfully!`);
        } catch (error) {
          // 如果是 "Already Verified" 错误，不算失败
          if (error.message.includes("Already Verified") || error.message.includes("already verified")) {
            console.log(`  ✅ ${contract.name} already verified`);
          } else {
            console.log(`  ⚠️  ${contract.name} verification failed: ${error.message}`);
            console.log(`     You can verify manually later:`);
            console.log(`     npx hardhat verify --network ${hre.network.name} ${contract.address} ${constructorArgs ? constructorArgs.join(" ") : ""}`);
          }
        }
      }
    }

    console.log("\n✅ Deployment info saved!");
    console.log("\n📋 Deployment Summary:");
    console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
    console.log(`Network:   ${hre.network.name}`);
    console.log(`Chain ID:  ${chainId}`);
    console.log(`Deployer:  ${deployer.address}`);
    console.log(`Module:    ${moduleName}`);
    console.log(`Timestamp: ${new Date().toISOString()}`);
    console.log("\n📦 Contracts:");

    contracts.forEach(contract => {
      console.log(`  ┌─ ${contract.name}`);
      console.log(`  │  Address: ${contract.address}`);
      console.log(`  │  TX Hash: ${contract.transactionHash}`);
      console.log(`  │  File: ${contract.file}`);
      console.log(`  └─────────────────────────────────`);
    });
    console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n");
  });

// 添加一个独立的保存任务，可以在部署后手动调用
task("save-ignition-deployment", "Save Ignition deployment info from existing deployment")
  .setAction(async (_args, hre) => {
    const network = await hre.ethers.provider.getNetwork();
    const chainId = network.chainId.toString();

    console.log(`\n📝 Saving deployment info for ${hre.network.name} (Chain ID: ${chainId})...`);

    const deploymentDir = path.join(__dirname, "../ignition/deployments");
    const deployedAddressesFile = path.join(deploymentDir, `chain-${chainId}/deployed_addresses.json`);

    if (!fs.existsSync(deployedAddressesFile)) {
      console.log("❌ No deployment found. Please deploy first using:");
      console.log(`   npx hardhat ignition deploy <module> --network ${hre.network.name}`);
      return;
    }

    const deployedAddresses = JSON.parse(fs.readFileSync(deployedAddressesFile, "utf8"));

    // 从 journal.jsonl 提取交易 hash
    const journalFile = path.join(deploymentDir, `chain-${chainId}/journal.jsonl`);
    let transactions = {};

    if (fs.existsSync(journalFile)) {
      const journalContent = fs.readFileSync(journalFile, "utf8");
      const lines = journalContent.trim().split("\n");

      for (const line of lines) {
        try {
          const entry = JSON.parse(line);
          // 从 TRANSACTION_CONFIRM 条目中提取交易 hash
          if (entry.type === "TRANSACTION_CONFIRM" && entry.hash) {
            const futureId = entry.futureId;
            transactions[futureId] = entry.hash;
          }
        } catch (e) {
          // 跳过无法解析的行
        }
      }
    } else {
      console.log("⚠️  Journal file not found at:", journalFile);
    }

    const [deployer] = await hre.ethers.getSigners();

    // 获取所有模块名称
    const moduleNames = new Set();
    for (const contractKey of Object.keys(deployedAddresses)) {
      const moduleName = contractKey.includes("#") ? contractKey.split("#")[0] : contractKey;
      moduleNames.add(moduleName);
    }

    for (const moduleName of moduleNames) {
      const deploymentInfo = {
        network: hre.network.name,
        chainId: chainId,
        timestamp: new Date().toISOString(),
        deployer: deployer.address,
        module: moduleName,
        contracts: []
      };

      for (const [contractKey, address] of Object.entries(deployedAddresses)) {
        const contractName = contractKey.includes("#")
          ? contractKey.split("#")[1]
          : contractKey;

        deploymentInfo.contracts.push({
          name: contractName,
          address: address,
          transactionHash: transactions[contractKey] || "N/A"
        });
      }

      const outputDir = path.join(__dirname, "../deployments");
      if (!fs.existsSync(outputDir)) {
        fs.mkdirSync(outputDir, { recursive: true });
      }

      const outputFile = path.join(outputDir, `${hre.network.name}-${chainId}-${moduleName}.json`);
      fs.writeFileSync(outputFile, JSON.stringify(deploymentInfo, null, 2));

      console.log(`\n✅ Saved to: ${outputFile}`);
    }
  });

const hre = require("hardhat");
// 引入 hardhat 的 ethers 和 upgrades 模块
const {ethers, upgrades} = require("hardhat");
// 引入工具函数：保存和获取已部署的合约地址
const {saveContractAddress, getSavedContractAddresses} = require('../utils')
// 引入配置文件
const config = require('../configs/config.json');
// const yesno = require('yesno'); // 用户确认工具（已注释）

// 获取当前区块时间戳
async function getCurrentBlockTimestamp() {
    return (await ethers.provider.getBlock('latest')).timestamp;
}

// 延时函数
async function sleep(number) {
    return new Promise((resolve) => {
        setTimeout(resolve, number);
    });
}

// 主部署函数
async function main() {
    // 获取当前网络的配置
    const c = config[hre.network.name];
    console.log("network name: ", hre.network.name)
    // 获取已部署的合约地址
    const contracts = getSavedContractAddresses()[hre.network.name];

    // 零地址（用于初始化）
    const ZERO_ADDRESS = "0x0000000000000000000000000000000000000000";

    // 获取 Admin 合约工厂
    const Admin = await ethers.getContractFactory("Admin");
    console.log("ready to deploy admin")

    // 部署 Admin 合约，传入管理员地址列表
    const admin = await Admin.deploy(c.admins);
    await admin.deployed();
    console.log("Admin contract deployed to: ", admin.address);
    saveContractAddress(hre.network.name, "Admin", admin.address);

    // 部署 SalesFactory 合约
    console.log("ready to deploy salesFactory ")
    const SalesFactory = await ethers.getContractFactory("SalesFactory");
    const salesFactory = await SalesFactory.deploy(admin.address, ZERO_ADDRESS);
    await salesFactory.deployed();
    saveContractAddress(hre.network.name, "SalesFactory", salesFactory.address);
    console.log('Sales factory deployed to: ', salesFactory.address);


    // 部署 AllocationStaking 合约（可升级代理合约）
    console.log("ready to deploy AllocationStaking ")
    const currentTimestamp = await getCurrentBlockTimestamp();
    console.log('Farming starts at: ', currentTimestamp);
    const AllocationStaking = await ethers.getContractFactory("AllocationStaking");
    // 使用可升级模式部署代理合约
    const allocationStaking = await upgrades.deployProxy(AllocationStaking, [
            contracts["C2N-TOKEN"],           // 奖励代币地址
            ethers.utils.parseEther(c.allocationStakingRPS),  // 每秒奖励数量
            currentTimestamp + c.delayBeforeStart,  // 开始时间戳
            salesFactory.address              // SalesFactory 地址
        ], {unsafeAllow: ['delegatecall']}   // 允许 delegatecall
    );
    await allocationStaking.deployed()
    console.log('AllocationStaking Proxy deployed to:', allocationStaking.address);
    saveContractAddress(hre.network.name, 'AllocationStakingProxy', allocationStaking.address);

    // 获取代理管理员合约地址
    let proxyAdminContract = await upgrades.admin.getInstance();
    saveContractAddress(hre.network.name, 'ProxyAdmin', proxyAdminContract.address);
    console.log('Proxy Admin address is : ', proxyAdminContract.address);

    // 设置 SalesFactory 的质押合约地址
    console.log("ready to setAllocationStaking params ")
    await salesFactory.setAllocationStaking(allocationStaking.address);
    console.log(`salesFactory.setAllocationStaking ${allocationStaking.address} done.;`);

    // 计算总奖励数量
    const totalRewards = ethers.utils.parseEther(c.initialRewardsAllocationStaking);

    // 获取 C2N 代币合约实例
    const token = await hre.ethers.getContractAt('C2NToken', contracts['C2N-TOKEN']);

    // 授权质押合约可以使用代币
    console.log("ready to approve ", c.initialRewardsAllocationStaking, " token to staking  ")

    let tx = await token.approve(allocationStaking.address, totalRewards);
    await tx.wait()
    console.log(`token.approve(${allocationStaking.address}, ${totalRewards.toString()});`)

    // 添加 C2N 代币到质押池
    console.log("ready to add c2n to pool")
    // add c2n to pool
    tx = await allocationStaking.add(100, token.address, true);
    await tx.wait()
    console.log(`allocationStaking.add(${token.address});`)

    // 添加 BOBA 代币到质押池（已注释）
    console.log("ready to add boba to pool")
    // add boba to pool
    // await allocationStaking.add(100, contracts["BOBA-TOKEN"], true);
    // console.log(`allocationStaking.add(${contracts["BOBA-TOKEN"]});`)


    // 为测试资金代币
    // fund tokens for testing
    const fund = Math.floor(Number(c.initialRewardsAllocationStaking)).toString()
    console.log(`ready to fund ${fund} token for testing`)
    // Fund only 50000 tokens, for testing
    // sleep(5000)

    // 向质押合约注入奖励代币
    await allocationStaking.fund(ethers.utils.parseEther(fund));
    console.log('Funded tokens')

}


// 执行主函数
main()
    .then(() => process.exit(0))   // 成功时退出
    .catch(error => {
        console.error(error);       // 错误时打印错误信息
        process.exit(1);            // 以错误码退出
    });

const hre = require("hardhat");
// 引入工具函数：保存和获取已部署的合约地址
const {saveContractAddress, getSavedContractAddresses} = require('../utils')
// 引入销售配置文件
const config = require("../configs/saleConfig.json");
// 用户确认工具（已注释，实际未使用）
const yesno = require("yesno");
const {ethers, web3} = hre

// 获取当前区块时间戳
async function getCurrentBlockTimestamp() {
    return (await ethers.provider.getBlock('latest')).timestamp;
}

// 延时函数
const delay = ms => new Promise(res => setTimeout(res, ms));
const delayLength = 3000;

// 主部署函数：部署并配置销售合约
async function main() {
    // 获取已部署的合约地址
    const contracts = getSavedContractAddresses()[hre.network.name];
    // 获取当前网络的销售配置
    const c = config[hre.network.name];

    // 获取 SalesFactory 合约实例
    const salesFactory = await hre.ethers.getContractAt('SalesFactory', contracts['SalesFactory']);

    // 通过 SalesFactory 部署新的销售合约
    let tx = await salesFactory.deploySale();
    await tx.wait()
    console.log('Sale is deployed successfully.');

    // let ok = await yesno({
    //     question: 'Are you sure you want to continue?'
    // });
    // if (!ok) {
    //     process.exit(0)
    // }

    // 获取最新部署的销售合约地址
    const lastDeployedSale = await salesFactory.getLastDeployedSale();
    console.log('Deployed Sale address is: ', lastDeployedSale);

    // 获取销售合约实例
    const sale = await hre.ethers.getContractAt('C2NSale', lastDeployedSale);
    console.log(`Successfully instantiated sale contract at address: ${lastDeployedSale}.`);

    // 解析销售参数
    const totalTokens = ethers.utils.parseEther(c['totalTokens']);
    console.log('Total tokens to sell: ', c['totalTokens']);

    const tokenPriceInEth = ethers.utils.parseEther(c['tokenPriceInEth']);
    console.log('tokenPriceInEth:', c['tokenPriceInEth']);

    const saleOwner = c['saleOwner'];
    console.log('Sale owner is: ', c['saleOwner']);

    // 计算各阶段时间
    const registrationStart = c['registrationStartAt'];     // 注册开始时间
    const registrationEnd = registrationStart + c['registrationLength'];  // 注册结束时间
    const saleStartTime = registrationEnd + c['delayBetweenRegistrationAndSale'];  // 销售开始时间
    const saleEndTime = saleStartTime + c['saleRoundLength'];  // 销售结束时间
    const maxParticipation = ethers.utils.parseEther(c['maxParticipation']);  // 最大参与金额

    const tokensUnlockTime = c['TGE'];  // 代币解锁时间（TGE）

    console.log("ready to set sale params")
    // ok = await yesno({
    //     question: 'Are you sure you want to continue?'
    // });
    // if (!ok) {
    //     process.exit(0)
    // }
    // 设置销售参数：代币地址、所有者、价格、总量、结束时间、解锁时间等
    tx = await sale.setSaleParams(
        c['tokenAddress'],              // 销售代币地址
        saleOwner,                      // 销售所有者
        tokenPriceInEth.toString(),     // 代币价格（以 ETH 计价）
        totalTokens.toString(),         // 销售代币总量
        saleEndTime,                    // 销售结束时间
        tokensUnlockTime,               // 代币解锁时间
        c['portionVestingPrecision'],   // 释放精度
        maxParticipation.toString()     // 最大参与金额
    );
    await tx.wait()

    console.log('Sale Params set successfully.');

    console.log('Setting registration time.');

    // ok = await yesno({
    //     question: 'Are you sure you want to continue?'
    // });
    // if (!ok) {
    //     process.exit(0)
    // }
    //
    console.log('registrationStart:',registrationStart)
    console.log('registrationEnd:',registrationEnd)
    // 设置注册时间窗口
    tx = await sale.setRegistrationTime(
        registrationStart,
        registrationEnd
    );
    await tx.wait()

    console.log('Registration time set.');

    console.log('Setting saleStart.');

    // ok = await yesno({
    //     question: 'Are you sure you want to continue?'
    // });
    // if (!ok) {
    //     process.exit(0)
    // }
    // 设置销售开始时间
    tx = await sale.setSaleStart(saleStartTime);
    await tx.wait()

    // 解析释放参数
    const unlockingTimes = c['unlockingTimes'];  // 解锁时间点数组
    const percents = c['portionPercents'];        // 对应的释放百分比数组

    console.log('Unlocking times: ', unlockingTimes);
    console.log('Percents: ', percents);
    console.log('Precision for vesting: ', c['portionVestingPrecision']);
    console.log('Max vesting time shift in seconds: ', c['maxVestingTimeShift']);

    console.log('Setting vesting params.');
    //
    // ok = await yesno({
    //     question: 'Are you sure you want to continue?'
    // });
    // if (!ok) {
    //     process.exit(0)
    // }
    // 设置线性释放参数：解锁时间、释放百分比、最大时间偏移
    tx = await sale.setVestingParams(unlockingTimes, percents, c['maxVestingTimeShift']);
    await tx.wait()

    console.log('Vesting parameters set successfully.');

    // 打印销售配置摘要
    console.log({
        saleAddress: lastDeployedSale,
        saleToken: c['tokenAddress'],
        saleOwner,
        tokenPriceInEth: tokenPriceInEth.toString(),
        totalTokens: totalTokens.toString(),
        saleEndTime,
        tokensUnlockTime,
        registrationStart,
        registrationEnd,
        saleStartTime
    });

    // 打印 JSON 格式的销售配置（便于复制保存）
    console.log(JSON.stringify({
        saleAddress: lastDeployedSale,
        saleToken: c['tokenAddress'],
        saleOwner,
        tokenPriceInEth: tokenPriceInEth.toString(),
        totalTokens: totalTokens.toString(),
        saleEndTime,
        tokensUnlockTime,
        registrationStart,
        registrationEnd,
        saleStartTime
    }))
}


// 执行主函数
main()
    .then(() => process.exit(0))   // 成功时退出
    .catch(error => {
        console.error(error);       // 错误时打印错误信息
        process.exit(1);            // 以错误码退出
    });

const { ethers } = require('ethers');
require('dotenv').config();

// 获取 token decimal
async function getTokenDecimal(tokenContract) {
  const decimal = await tokenContract.decimals();
  const tokenAddress = await tokenContract.getAddress();
  console.log('token %s decimal: %s', tokenAddress, decimal);
}

// 查询token余额
async function getTokenBalance(tokenContract, address) {
  const balance = await tokenContract.balanceOf(address);
  return balance;
}

// token 授权
async function approveSpender(tokenContract, spender, amount) {
  const tx = await tokenContract.approve(spender, amount);
  await tx.wait();
  console.log('%s 授权 %s 金额 %s 成功', tx.from, spender, amount);
}

// token 转账
async function transfer(tokenContract, to, amount) {
  const tx = await tokenContract.transfer(to, amount);
  await tx.wait();
  console.log('%s 转账 %s 金额 %s 成功', tx.from, to, amount);
}

// 查看用户质押信息
async function getUserActiveStakes(stakingContract, userAddress) {
  const stakes = await stakingContract.getUserActiveStakes(userAddress);
  return stakes;
}

// 计算质押奖励
async function calculateReward(stakingContract, userAddress, stakeId) {
  const reward = await stakingContract.calculateReward(userAddress, stakeId);
  return reward;
}

// 查看用户质押是否过期
async function isStakeExpired(stakingContract, address, stakeId) {
  const expired = await stakingContract.isStakeExpired(address, stakeId);
  return expired;
}


// 质押
async function stake(stakingContract, amount, period) {
  const tx = await stakingContract.stake(amount, period);
  await tx.wait();
  console.log('用户 %s, 质押金额 %s 成功, 质押周期 %s', tx.from, amount, period);
}

// 提现
async function withdraw(stakingContract, stakeId) {
  const tx = await stakingContract.withdraw(stakeId);
  await tx.wait();
  console.log('%s 提现质押 %s 成功', tx.from, stakeId);
}

// 主函数
async function main() {
  // 初始化Provider (BSC测试网)
  const provider = new ethers.JsonRpcProvider(
    'https://bsc-testnet-dataseed.bnbchain.org'
  );

  // 初始化钱包
  const wallet = new ethers.Wallet(process.env.SEPOLIA_PRIVATE_KEY).connect(provider);

  // 打印钱包地址
  console.log('钱包地址:', wallet.address);

  // 导入ABI
  const tokenABI = require('./abis/MyToken.json');
  const stakingABI = require('./abis/MtkContracts.json');

  // 初始化合约
  const tokenContractAddress = '0x93480Ce4b54baD6c60D8CDAEaeaF898fE00deBF2'; // Token合约地址
  const stakeContractAddress = '0x6287A4e265CfEA1B9C87C1dC692363d69f58378c'; // 质押合约地址

  const tokenContract = new ethers.Contract(tokenContractAddress, tokenABI, wallet);
  const stakingContract = new ethers.Contract(stakeContractAddress, stakingABI, wallet);

  console.log('Ethers.js SDK初始化完成');

  // 查询用户余额
  let balance = await getTokenBalance(tokenContract, wallet.address);
  const decimal = await getTokenDecimal(tokenContract);
  console.log('address', wallet.address, ', 当前余额:', ethers.formatUnits(balance, decimal));

  // 查看用户质押信息
  let stakes = await getUserActiveStakes(stakingContract, wallet.address);
  console.log('用户质押前的信息:', stakes);

  if (stakes.length == 0) {
    // 授权质押合约
    const amount = ethers.parseUnits('100', decimal);
    await approveSpender(tokenContract, stakeContractAddress, amount);

    // 质押
    await stake(stakingContract, amount, 0);
  }

  // 查看用户质押信息
  stakes = await getUserActiveStakes(stakingContract, wallet.address);
  console.log('用户质押信息:', stakes);

  for (const stake of stakes) {
    let reward = await calculateReward(stakingContract, wallet.address, stake.stakeId);
    console.log('user %s 质押 %s 奖励:', wallet.address, stake.stakeId, reward);

    // 查询质押合约token余额
    let stakingTokenBalance = await getTokenBalance(tokenContract, stakeContractAddress);
    console.log('质押合约 %s token 余额:', stakeContractAddress, stakingTokenBalance);

    if (reward > stakingTokenBalance) {
      let diffValue = reward - stakingTokenBalance;
      await transfer(tokenContract, stakeContractAddress, diffValue);
    }

    // 用户提现
    let expired = await isStakeExpired(stakingContract, wallet.address, stake.stakeId);
    console.log('质押 %s 是否过期:', stake.stakeId, expired);
    if (expired) {
      await withdraw(stakingContract, stake.stakeId);
    }
  }

  // 查询用户余额
  balance = await getTokenBalance(tokenContract, wallet.address);
  console.log('address', wallet.address, ', 当前余额:', ethers.formatUnits(balance, decimal));
}

// 执行主函数并处理错误
main().catch((error) => {
  console.error('Error:', error);
  process.exit(1);
});

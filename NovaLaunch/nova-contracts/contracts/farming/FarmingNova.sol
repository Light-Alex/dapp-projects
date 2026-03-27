// SPDX-License-Identifier: MIT
pragma solidity 0.6.12;

// 引入 OpenZeppelin 库
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/SafeERC20.sol";
import "@openzeppelin/contracts/math/SafeMath.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

// Farm distributes the ERC20 rewards based on staked LP to each user.
// Farm 根据用户质押的 LP 代币分配 ERC20 奖励
//
// Cloned from https://github.com/SashimiProject/sashimiswap/blob/master/contracts/MasterChef.sol
// Modified by LTO Network to work for non-mintable ERC20.

// C2N 质押挖矿合约
contract FarmingC2N is Ownable {

    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    // Info of each user.
    // 用户信息结构体
    struct UserInfo {
        uint256 amount;     // How many LP tokens the user has provided. // 用户质押的 LP 代币数量
        uint256 rewardDebt; // Reward debt. See explanation below. // 奖励债务（用于计算待领取奖励）
        //
        // We do some fancy math here. Basically, any point in time, the amount of ERC20s
        // entitled to a user but is pending to be distributed is:
        // 这里的数学逻辑：用户在任何时刻待领取的奖励计算公式为：
        //
        //   pending reward = (user.amount * pool.accERC20PerShare) - user.rewardDebt
        //   待领取奖励 = (用户质押数量 × 池子累计每份额奖励) - 用户奖励债务
        //
        // Whenever a user deposits or withdraws LP tokens to a pool. Here's what happens:
        // 当用户存入或取出 LP 代币时，执行以下步骤：
        //   1. The pool's `accERC20PerShare` (and `lastRewardBlock`) gets updated. // 更新池子的累计奖励
        //   2. User receives the pending reward sent to his/her address. // 发送待领取奖励给用户
        //   3. User's `amount` gets updated. // 更新用户质押数量
        //   4. User's `rewardDebt` gets updated. // 更新用户奖励债务
    }

    // Info of each pool.
    // 池子信息结构体
    struct PoolInfo {
        IERC20 lpToken;             // Address of LP token contract. // LP 代币合约地址
        uint256 allocPoint;         // How many allocation points assigned to this pool. ERC20s to distribute per block. // 分配给该池子的分配点数（用于计算奖励分配权重）
        uint256 lastRewardTimestamp;    // Last timstamp that ERC20s distribution occurs. // 上次分发奖励的时间戳
        uint256 accERC20PerShare;   // Accumulated ERC20s per share, times 1e36. // 累计每份额奖励（乘以 1e36 提高精度）
        uint256 totalDeposits; // Total amount of tokens deposited at the moment (staked) // 当前总质押量
    }


    // Address of the ERC20 Token contract.
    // 奖励代币合约地址
    IERC20 public erc20;
    // The total amount of ERC20 that's paid out as reward.
    // 已支付的奖励总量
    uint256 public paidOut;
    // ERC20 tokens rewarded per second.
    // 每秒奖励数量
    uint256 public rewardPerSecond;
    // Total rewards added to farm
    // 农场总奖励量
    uint256 public totalRewards;
    // Info of each pool.
    // 池子信息数组
    PoolInfo[] public poolInfo;
    // Info of each user that stakes LP tokens.
    // 用户信息映射：poolId => 用户地址 => 用户信息
    mapping(uint256 => mapping(address => UserInfo)) public userInfo;
    // Total allocation points. Must be the sum of all allocation points in all pools.
    // 总分配点数（所有池子的分配点数之和）
    uint256 public totalAllocPoint;

    // The timestamp when farming starts.
    // 质押开始时间戳
    uint256 public startTimestamp;
    // The timestamp when farming ends.
    // 质押结束时间戳
    uint256 public endTimestamp;

    // 事件定义
    event Deposit(address indexed user, uint256 indexed pid, uint256 amount);         // 存款事件
    event Withdraw(address indexed user, uint256 indexed pid, uint256 amount);        // 取款事件
    event EmergencyWithdraw(address indexed user, uint256 indexed pid, uint256 amount); // 紧急取款事件

    // 构造函数：初始化奖励代币、每秒奖励数量和开始时间
    constructor(IERC20 _erc20, uint256 _rewardPerSecond, uint256 _startTimestamp) public {
        erc20 = _erc20;
        rewardPerSecond = _rewardPerSecond;
        startTimestamp = _startTimestamp;
        endTimestamp = _startTimestamp;
    }

    // Number of LP pools
    // 获取 LP 池子数量
    function poolLength() external view returns (uint256) {
        return poolInfo.length;
    }

    // Fund the farm, increase the end block
    // 向农场注入奖励，延长结束时间
    function fund(uint256 _amount) public {
        require(block.timestamp < endTimestamp, "fund: too late, the farm is closed");
        erc20.safeTransferFrom(address(msg.sender), address(this), _amount);
        endTimestamp += _amount.div(rewardPerSecond);  // 根据注入金额延长结束时间
        totalRewards = totalRewards.add(_amount);
    }

    // Add a new lp to the pool. Can only be called by the owner.
    // 添加新的 LP 池子，仅所有者可调用
    // DO NOT add the same LP token more than once. Rewards will be messed up if you do.
    // 不要重复添加相同的 LP 代币，否则会导致奖励计算错误
    function add(uint256 _allocPoint, IERC20 _lpToken, bool _withUpdate) public onlyOwner {
        if (_withUpdate) {
            massUpdatePools();  // 更新所有池子的奖励
        }
        uint256 lastRewardTimestamp = block.timestamp > startTimestamp ? block.timestamp : startTimestamp;
        totalAllocPoint = totalAllocPoint.add(_allocPoint);
        poolInfo.push(PoolInfo({
        lpToken : _lpToken,
        allocPoint : _allocPoint,
        lastRewardTimestamp : lastRewardTimestamp,
        accERC20PerShare : 0,
        totalDeposits : 0
        }));
    }

    // Update the given pool's ERC20 allocation point. Can only be called by the owner.
    // 更新指定池子的分配点数，仅所有者可调用
    function set(uint256 _pid, uint256 _allocPoint, bool _withUpdate) public onlyOwner {
        if (_withUpdate) {
            massUpdatePools();
        }
        totalAllocPoint = totalAllocPoint.sub(poolInfo[_pid].allocPoint).add(_allocPoint);
        poolInfo[_pid].allocPoint = _allocPoint;
    }

    // View function to see deposited LP for a user.
    // 查询用户在指定池子中的质押数量
    function deposited(uint256 _pid, address _user) external view returns (uint256) {
        UserInfo storage user = userInfo[_pid][_user];
        return user.amount;
    }

    // View function to see pending ERC20s for a user.
    // 查询用户的待领取奖励
    function pending(uint256 _pid, address _user) external view returns (uint256) {
        PoolInfo storage pool = poolInfo[_pid];
        UserInfo storage user = userInfo[_pid][_user];
        uint256 accERC20PerShare = pool.accERC20PerShare;

        uint256 lpSupply = pool.totalDeposits;

        // 计算从上次更新到现在累积的奖励
        if (block.timestamp > pool.lastRewardTimestamp && lpSupply != 0) {
            uint256 lastTimestamp = block.timestamp < endTimestamp ? block.timestamp : endTimestamp;
            uint256 timestampToCompare = pool.lastRewardTimestamp < endTimestamp ? pool.lastRewardTimestamp : endTimestamp;
            uint256 nrOfSeconds = lastTimestamp.sub(timestampToCompare);
            uint256 erc20Reward = nrOfSeconds.mul(rewardPerSecond).mul(pool.allocPoint).div(totalAllocPoint);
            accERC20PerShare = accERC20PerShare.add(erc20Reward.mul(1e36).div(lpSupply));
        }
        // 计算用户待领取奖励：用户质押量 × 累计每份额奖励 - 奖励债务
        return user.amount.mul(accERC20PerShare).div(1e36).sub(user.rewardDebt);
    }

    // View function for total reward the farm has yet to pay out.
    // 查询农场待支付的总奖励
    function totalPending() external view returns (uint256) {
        if (block.timestamp <= startTimestamp) {
            return 0;
        }

        uint256 lastTimestamp = block.timestamp < endTimestamp ? block.timestamp : endTimestamp;
        // 总奖励 = 每秒奖励 × 运行时长 - 已支付奖励
        return rewardPerSecond.mul(lastTimestamp - startTimestamp).sub(paidOut);
    }

    // Update reward variables for all pools. Be careful of gas spending!
    // 更新所有池子的奖励变量（注意 gas 消耗！）
    function massUpdatePools() public {
        uint256 length = poolInfo.length;
        for (uint256 pid = 0; pid < length; ++pid) {
            updatePool(pid);
        }
    }

    // Update reward variables of the given pool to be up-to-date.
    // 更新指定池子的奖励变量
    function updatePool(uint256 _pid) public {
        PoolInfo storage pool = poolInfo[_pid];
        uint256 lastTimestamp = block.timestamp < endTimestamp ? block.timestamp : endTimestamp;

        if (lastTimestamp <= pool.lastRewardTimestamp) {
            return;
        }
        uint256 lpSupply = pool.totalDeposits;

        if (lpSupply == 0) {
            pool.lastRewardTimestamp = lastTimestamp;
            return;
        }

        // 计算时间差和应分配的奖励
        uint256 nrOfSeconds = lastTimestamp.sub(pool.lastRewardTimestamp);
        uint256 erc20Reward = nrOfSeconds.mul(rewardPerSecond).mul(pool.allocPoint).div(totalAllocPoint);

        // 更新累计每份额奖励，以及最近一次奖励时间
        pool.accERC20PerShare = pool.accERC20PerShare.add(erc20Reward.mul(1e36).div(lpSupply));
        pool.lastRewardTimestamp = block.timestamp;
    }

    // Deposit LP tokens to Farm for ERC20 allocation.
    // 存入 LP 代币到农场进行质押
    function deposit(uint256 _pid, uint256 _amount) public {
        PoolInfo storage pool = poolInfo[_pid];
        UserInfo storage user = userInfo[_pid][msg.sender];

        updatePool(_pid);  // 更新池子奖励

        // 如果用户之前有质押，先发送待领取的奖励
        if (user.amount > 0) {
            uint256 pendingAmount = user.amount.mul(pool.accERC20PerShare).div(1e36).sub(user.rewardDebt);
            erc20Transfer(msg.sender, pendingAmount);
        }

        // 转入 LP 代币
        pool.lpToken.safeTransferFrom(address(msg.sender), address(this), _amount);
        pool.totalDeposits = pool.totalDeposits.add(_amount);

        // 更新用户信息
        user.amount = user.amount.add(_amount);
        user.rewardDebt = user.amount.mul(pool.accERC20PerShare).div(1e36);
        emit Deposit(msg.sender, _pid, _amount);
    }

    // Withdraw LP tokens from Farm.
    // 包含两个功能，收取奖励，撤回质押
    function withdraw(uint256 _pid, uint256 _amount) public {
        PoolInfo storage pool = poolInfo[_pid];
        UserInfo storage user = userInfo[_pid][msg.sender];
        require(user.amount >= _amount, "withdraw: can't withdraw more than deposit");
        updatePool(_pid);

        // 计算奖励
        uint256 pendingAmount = user.amount.mul(pool.accERC20PerShare).div(1e36).sub(user.rewardDebt);

        // 发送奖励给用户
        erc20Transfer(msg.sender, pendingAmount);
        user.amount = user.amount.sub(_amount);
        user.rewardDebt = user.amount.mul(pool.accERC20PerShare).div(1e36);
        // 撤回流动性
        pool.lpToken.safeTransfer(address(msg.sender), _amount);
        pool.totalDeposits = pool.totalDeposits.sub(_amount);

        emit Withdraw(msg.sender, _pid, _amount);
    }

    // Withdraw without caring about rewards. EMERGENCY ONLY.
    // 紧急取款，放弃奖励（只退还全部本金），仅限紧急情况使用
    function emergencyWithdraw(uint256 _pid) public {
        PoolInfo storage pool = poolInfo[_pid];
        UserInfo storage user = userInfo[_pid][msg.sender];
        pool.lpToken.safeTransfer(address(msg.sender), user.amount);
        pool.totalDeposits = pool.totalDeposits.sub(user.amount);
        emit EmergencyWithdraw(msg.sender, _pid, user.amount);
        user.amount = 0;
        user.rewardDebt = 0;
    }

    // Transfer ERC20 and update the required ERC20 to payout all rewards
    // 转账 ERC20 奖励并更新已支付数量
    function erc20Transfer(address _to, uint256 _amount) internal {
        erc20.transfer(_to, _amount);
        paidOut += _amount;
    }
}

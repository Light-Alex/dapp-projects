// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract MtkContracts {
    // 枚举质押期限
    enum StakingPeriod { ThirtyDays, NinetyDays, HundredEightyDays, OneYear }

    uint256 constant PRECISION = 1e18;
    
    // 定义质押订单结构体
    struct Stake {
        uint256 stakeId;            // 唯一随机ID
        uint256 amount;             // 质押数量
        uint256 startTime;          // 质押开始时间
        uint256 endTime;            // 质押结束时间
        uint256 rewardRate;         // 收益率（根据期限计算）
        bool isActive;              // 订单是否有效
        StakingPeriod period;       // 质押期限
    }
    
    IERC20 public stakingToken;      // 质押代币合约
    
    mapping(address => Stake[]) public userStakes;              // 用户的所有质押订单
    mapping(uint256 => address) public stakeIdToOwner;          // 质押ID到所有者的映射
    mapping(StakingPeriod => uint256) public apy;               // 不同期限对应的年化收益率
    mapping(StakingPeriod => uint256) public durations;         // 不同期限对应的秒数
    uint256 private nonce;                                      // 用于生成随机ID
    
    event Staked(
        address indexed user,  // 质押用户
        uint256 stakeId,       // 质押订单ID
        uint256 amount,        // 质押金额
        StakingPeriod period,  // 质押期限
        uint256 timestamp      // 质押时间
    );

    event Withdrawn(
        address indexed user,  // 提取用户
        uint256 stakeId,       // 提取质押订单ID
        uint256 principal,     // 质押金额
        uint256 reward,        // 奖励金额
        uint256 totalAmount    // 提取总金额
    );

    constructor(IERC20 _mtkToken) {
        stakingToken = _mtkToken;
        
        // 初始化不同期限的持续时间（测试环境用分钟，生产环境应用days）
        durations[StakingPeriod.ThirtyDays] = 30 days;
        durations[StakingPeriod.NinetyDays] = 90 days;
        durations[StakingPeriod.HundredEightyDays] = 180 days;
        durations[StakingPeriod.OneYear] = 365 days;
        
        // 初始化不同期限的年化收益率
        apy[StakingPeriod.ThirtyDays] = 10;   // 10% 年化
        apy[StakingPeriod.NinetyDays] = 15;   // 15% 年化
        apy[StakingPeriod.HundredEightyDays] = 18; // 18% 年化
        apy[StakingPeriod.OneYear] = 20;      // 20% 年化
    }

    // 质押函数
    function stake(uint256 amount, StakingPeriod period) external {
        require(amount > 0, "Amount must be greater than zero");
        require(stakingToken.transferFrom(msg.sender, address(this), amount), "Transfer failed");
        
        uint256 duration = _getDuration(period);
        uint256 periodDays = durations[period];
        uint256 rate = apy[period] * periodDays * PRECISION / 360 days; // 计算实际收益率
        
        uint256 stakeId = _generateStakeId();
        
        Stake memory newStake = Stake({
            stakeId: stakeId,
            amount: amount,
            startTime: block.timestamp,
            endTime: block.timestamp + duration,
            rewardRate: rate,
            isActive: true,
            period: period
        });
        
        userStakes[msg.sender].push(newStake);
        stakeIdToOwner[stakeId] = msg.sender;
        
        emit Staked(msg.sender, stakeId, amount, period, block.timestamp);
    }

    // 生成唯一的质押ID
    function _generateStakeId() internal returns (uint256) {
        return uint256(keccak256(abi.encodePacked(block.timestamp, msg.sender, nonce++)));
    }

    // 根据期限返回秒数（测试环境用分钟）
    function _getDuration(StakingPeriod period) internal pure returns (uint256) {
        if (period == StakingPeriod.ThirtyDays) {
            return 1 minutes;
        } else if (period == StakingPeriod.NinetyDays) {
            return 3 minutes;
        } else if (period == StakingPeriod.HundredEightyDays) {
            return 5 minutes;
        } else {
            return 10 minutes;
        }
    }

    function calculateReward(address user, uint256 stakeId) public view returns (uint256 totalAmount) {
        Stake storage stk;
        uint256 stakeIndex;
        (stk, stakeIndex) = _getStakeById(user, stakeId);
        
        uint256 reward = stk.amount * stk.rewardRate / PRECISION / 100;
        totalAmount = stk.amount + reward;
    }
    
    // 提现函数
    function withdraw(uint256 stakeId) external {
        require(stakeIdToOwner[stakeId] == msg.sender, "Not the owner of this stake");
        
        Stake storage stk;
        uint256 stakeIndex;
        (stk, stakeIndex) = _getStakeById(msg.sender, stakeId);
        
        require(stk.isActive, "Stake is not active");
        require(block.timestamp >= stk.endTime, "Staking period is not over");
        
        stk.isActive = false;
        
        uint256 reward = stk.amount * stk.rewardRate / PRECISION / 100;
        uint256 totalAmount = stk.amount + reward;
        
        // 将质押的代币和收益转移给用户
        require(stakingToken.transfer(msg.sender, totalAmount), "Transfer failed");
        
        emit Withdrawn(msg.sender, stakeId, stk.amount, reward, totalAmount);
    }
    
    // 根据ID获取质押信息
    function _getStakeById(address user, uint256 stakeId) internal view returns (Stake storage, uint256) {
        for (uint256 i = 0; i < userStakes[user].length; i++) {
            if (userStakes[user][i].stakeId == stakeId) {
                return (userStakes[user][i], i);
            }
        }
        revert("Stake not found");
    }
    
    // 获取用户所有活跃的质押
    function getUserActiveStakes(address user) external view returns (Stake[] memory) {
        uint256 activeCount = 0;
        for (uint256 i = 0; i < userStakes[user].length; i++) {
            if (userStakes[user][i].isActive) {
                activeCount++;
            }
        }
        
        Stake[] memory activeStakes = new Stake[](activeCount);
        uint256 index = 0;
        for (uint256 i = 0; i < userStakes[user].length; i++) {
            if (userStakes[user][i].isActive) {
                activeStakes[index] = userStakes[user][i];
                index++;
            }
        }
        
        return activeStakes;
    }
}
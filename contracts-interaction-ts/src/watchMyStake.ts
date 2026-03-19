import { ethers, Contract, EventLog, ContractEventPayload  } from 'ethers';
import dotenv from 'dotenv';
import * as path from 'path';
import * as fs from 'fs';

dotenv.config();

// 配置参数
const STAKING_CONTRACT_ADDRESS = '0x6287A4e265CfEA1B9C87C1dC692363d69f58378c';
const RPC_URL = 'https://bsc-testnet-dataseed.bnbchain.org';
const RECONNECT_INTERVAL = 5000; // 重连间隔 5 秒
const MAX_RECONNECT_ATTEMPTS = 10; // 最大重连次数

// 事件类型枚举
enum EventType {
  Staked = 'Staked',
  Withdrawn = 'Withdrawn',
  All = 'All'
}

// 监听配置接口
interface WatchConfig {
  userAddress?: string;  // 过滤特定用户地址
  eventType?: EventType; // 过滤事件类型
}

// 监听器类
class StakingEventListener {
  private provider: ethers.JsonRpcProvider | null = null;
  private contract: Contract | null = null;
  private stakingABI: any;
  private config: WatchConfig;
  private isRunning = false;
  private reconnectAttempts = 0;
  private stakingFilter: any;
  private withdrawnFilter: any;

  constructor(config: WatchConfig = {}) {
    this.config = {
      userAddress: config.userAddress,
      eventType: config.eventType || EventType.All
    };

    // 导入 ABI
    this.stakingABI = JSON.parse(fs.readFileSync(path.join(__dirname, '../src/abis/MtkContracts.json'), 'utf8'));
  }

  /**
   * 初始化 Provider 和 Contract
   */
  private async initialize(): Promise<void> {
    this.provider = new ethers.JsonRpcProvider(RPC_URL);
    this.contract = new ethers.Contract(STAKING_CONTRACT_ADDRESS, this.stakingABI, this.provider);

    // 设置事件过滤器
    if (this.config.userAddress) {
      // 按用户地址过滤（user 是 indexed 参数）
      this.stakingFilter = this.contract.filters.Staked(this.config.userAddress);
      this.withdrawnFilter = this.contract.filters.Withdrawn(this.config.userAddress);
    } else {
      // 不过滤用户地址
      this.stakingFilter = this.contract.filters.Staked();
      this.withdrawnFilter = this.contract.filters.Withdrawn();
    }

    // 测试连接
    await this.provider.getBlockNumber();
    console.log('✅ 连接到 BSC 测试网成功');
  }

  /**
   * 处理 Staked 事件
   */
  private handleStaked = async (payload: ContractEventPayload): Promise<void> => {
    if (this.config.eventType === EventType.Withdrawn) return;

    // 从 payload.args 中获取事件参数
    const { user, stakeId, amount, period, timestamp } = payload.args;

    // 从 payload.log 中获取交易哈希和区块号
    const { transactionHash, blockNumber } = payload.log;

    console.log('\n📈 ===== Staked 事件 =====');
    console.log(`用户: ${user}`);
    console.log(`质押ID: ${stakeId}`);
    console.log(`质押金额: ${ethers.formatUnits(amount, 18)} Tokens`);
    console.log(`期限类型: ${period} (0=30天, 1=90天, 2=180天, 3=1年)`);
    console.log(`时间戳: ${new Date(Number(timestamp) * 1000).toLocaleString('zh-CN')}`);
    console.log(`交易哈希: ${transactionHash}`);
    console.log(`区块号: ${blockNumber}`);
    console.log('========================\n');
  }

  /**
   * 处理 Withdrawn 事件
   */
    private handleWithdrawn = async (payload: ContractEventPayload): Promise<void> => {
    if (this.config.eventType === EventType.Staked) return;

    // 从 payload.args 中获取事件参数
    const { user, stakeId, principal, reward, totalAmount } = payload.args;

    // 从 payload.log 中获取交易哈希和区块号
    const { transactionHash, blockNumber } = payload.log;

    console.log('\n💰 ===== Withdrawn 事件 =====');
    console.log(`用户: ${user}`);
    console.log(`质押ID: ${stakeId}`);
    console.log(`本金: ${ethers.formatUnits(principal, 18)} Tokens`);
    console.log(`奖励: ${ethers.formatUnits(reward, 18)} Tokens`);
    console.log(`总金额: ${ethers.formatUnits(totalAmount, 18)} Tokens`);
    console.log(`交易哈希: ${transactionHash}`);
    console.log(`区块号: ${blockNumber}`);
    console.log('============================\n');
    }

  /**
   * 开始监听事件
   */
  private async startListening(): Promise<void> {
    if (!this.contract) {
      throw new Error('合约未初始化');
    }

    console.log(`\n🎯 开始监听质押合约事件...`);
    console.log(`合约地址: ${STAKING_CONTRACT_ADDRESS}`);
    if (this.config.userAddress) {
      console.log(`过滤用户: ${this.config.userAddress}`);
    }
    if (this.config.eventType !== EventType.All) {
      console.log(`事件类型: ${this.config.eventType}`);
    }
    console.log('等待事件...\n');

    // 监听 Staked 事件
    if (this.config.eventType === EventType.Staked) {
      this.contract.on(this.stakingFilter, this.handleStaked);
    }

    // 监听 Withdrawn 事件
    if (this.config.eventType === EventType.Withdrawn) {
      this.contract.on(this.withdrawnFilter, this.handleWithdrawn);
    }

    // 监听所有事件
    if (this.config.eventType === EventType.All) {
      this.contract.on(this.stakingFilter, this.handleStaked);
      this.contract.on(this.withdrawnFilter, this.handleWithdrawn);
    }

    this.reconnectAttempts = 0; // 重置重连计数
  }

  /**
   * 停止监听
   */
  private stopListening(): void {
    if (this.contract) {
      this.contract.removeAllListeners();
      console.log('⏹️  停止监听事件');
    }
  }

  /**
   * 处理错误和重连
   */
  private handleError = async (error: any): Promise<void> => {
    console.error(`❌ 连接错误: ${error.message}`);

    if (this.reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
      this.reconnectAttempts++;
      console.log(`🔄 尝试重连... (${this.reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})`);

      // 停止当前监听
      this.stopListening();

      // 等待后重连
      await new Promise(resolve => setTimeout(resolve, RECONNECT_INTERVAL));

      try {
        // 重新初始化并开始监听
        await this.initialize();
        await this.startListening();
      } catch (retryError) {
        this.handleError(retryError);
      }
    } else {
      console.error(`❌ 达到最大重连次数 (${MAX_RECONNECT_ATTEMPTS})，停止尝试`);
      this.isRunning = false;
    }
  }

  /**
   * 启动监听服务
   */
  public async start(): Promise<void> {
    this.isRunning = true;

    try {
      // 初始化
      await this.initialize();

      // 开始监听
      await this.startListening();

      // 监听 Provider 错误
      if (this.provider) {
        this.provider.on('error', this.handleError);
        this.provider.on('network', () => {
          console.log('🌐 网络已更改');
        });
      }

      // 保持程序运行
      process.on('SIGINT', () => {
        console.log('\n⚠️  收到退出信号，正在关闭...');
        this.stop();
        process.exit(0);
      });

    } catch (error: any) {
      await this.handleError(error);
    }
  }

  /**
   * 停止监听服务
   */
  public stop(): void {
    this.isRunning = false;
    this.stopListening();

    if (this.provider) {
      this.provider.removeAllListeners();
    }
  }
}

// ============ 使用示例 ============

/**
 * 示例 1: 监听所有用户的所有事件
 */
async function watchAllEvents() {
  const listener = new StakingEventListener();
  await listener.start();
}

/**
 * 示例 2: 只监听特定用户的事件
 */
async function watchUserEvents(userAddress: string) {
  const listener = new StakingEventListener({
    userAddress: userAddress
  });
  await listener.start();
}

/**
 * 示例 3: 只监听特定类型的事件
 */
async function watchStakedEvents() {
  const listener = new StakingEventListener({
    eventType: EventType.Staked
  });
  await listener.start();
}

/**
 * 示例 4: 监听特定用户的质押事件
 */
async function watchUserStakedEvents(userAddress: string) {
  const listener = new StakingEventListener({
    userAddress,
    eventType: EventType.Staked
  });
  await listener.start();
}

// ============ 主函数 ============

async function main() {
  // 从环境变量获取用户地址（可选）
//   const userAddress = '0x6687e46C68C00bd1C10F8cc3Eb000B1752737e94';
  const userAddress = '';

  if (userAddress) {
    console.log(`🔍 监听用户 ${userAddress} 的质押和提现事件`);
    await watchUserEvents(userAddress);
  } else {
    console.log('🔍 监听所有用户的质押和提现事件');
    await watchAllEvents();
  }
}

// 启动监听
main().catch((error) => {
  console.error('启动失败:', error);
  process.exit(1);
});

// // 导出供其他模块使用
// export { StakingEventListener, EventType, WatchConfig };

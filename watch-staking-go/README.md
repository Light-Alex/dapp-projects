# Staking Contract Event Listener (Go)

使用 Go 语言和 go-ethereum 库实时监听 MtkContracts 合约的 Staked 和 Withdrawn 事件。

## 功能特性

- ✅ 按用户地址过滤事件
- ✅ 按事件类型过滤（Staked/Withdrawn/All）
- ✅ 自动重连机制（网络中断后自动重连，最多 10 次）
- ✅ 优雅退出（Ctrl+C）
- ✅ 格式化事件输出

## 安装依赖

```bash
cd e:/web3_workspace/dapp_projects/watch-staking-go
go mod tidy
```

## 使用方法

### 1. 监听所有用户的所有事件

```bash
go run main.go
```

### 2. 监听特定用户的事件

```bash
USER_ADDRESS=0x你的地址 go run main.go
```

### 3. 只监听质押事件

```bash
EVENT_TYPE=Staked go run main.go
```

### 4. 监听特定用户的质押事件

```bash
USER_ADDRESS=0x你的地址 EVENT_TYPE=Staked go run main.go
```

### 5. 只监听提现事件

```bash
EVENT_TYPE=Withdrawn go run main.go
```

## 编译并运行

```bash
# 编译
go build -o watch-staking main.go

# 运行
./watch-staking
```

## 环境变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `USER_ADDRESS` | 过滤特定用户地址（可选） | `0x1234...abcd` |
| `EVENT_TYPE` | 过滤事件类型（可选） | `Staked`, `Withdrawn` |

## 事件输出示例

### Staked 事件

```
📈 ===== Staked 事件 =====
用户: 0x6687e46C68C00bd1C10F8cc3Eb000B1752737e94
质押ID: 123456789012345678901234567890123456789012345678901234567890
质押金额: 100.0 Tokens
期限类型: 0 (0=30天, 1=90天, 2=180天, 3=1年)
时间戳: 2025-03-18 15:30:45
交易哈希: 0xabcdef...
区块号: 12345678
========================
```

### Withdrawn 事件

```
💰 ===== Withdrawn 事件 =====
用户: 0x6687e46C68C00bd1C10F8cc3Eb000B1752737e94
质押ID: 123456789012345678901234567890123456789012345678901234567890
本金: 100.0 Tokens
奖励: 5.0 Tokens
总金额: 105.0 Tokens
交易哈希: 0x123456...
区块号: 12345690
============================
```

## 配置说明

在 `main.go` 中可以修改以下配置：

```go
const (
    StakingContractAddress = "0x6287A4e265CfEA1B9C87C1dC692363d69f58378c" // 质押合约地址
    RPCURL                 = "https://bsc-testnet-dataseed.bnbchain.org" // BSC 测试网 RPC
    ReconnectInterval      = 5 * time.Second  // 重连间隔
    MaxReconnectAttempts   = 10                // 最大重连次数
)
```

## 项目结构

```
watch-staking-go/
├── abis/
│   └── MtkContracts.json    # 合约 ABI
├── main.go                   # 主程序
├── go.mod                    # Go 模块文件
├── go.sum                    # 依赖锁定文件
└── README.md                 # 说明文档
```

## 技术栈

- **Go 1.24+**
- **go-ethereum v1.17.1** - 以太坊 Go 客户端库

## 注意事项

1. 确保 `abis/MtkContracts.json` 文件存在
2. 确保 RPC URL 可访问
3. 程序会持续运行直到收到 `Ctrl+C` 信号
4. 网络中断后会自动重连，最多尝试 10 次

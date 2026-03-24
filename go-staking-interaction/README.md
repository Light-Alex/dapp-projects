# go-staking-interaction

## SyncBlock 技术架构

基于 Go 语言开发的 Web3 区块链交互系统，实现区块链交易监听、资产充值处理、提现管理等功能。

### 项目概述

本项目是一个生产级的区块链数据处理系统，主要用于：
- 🔍 **实时监听区块链交易**：扫描区块，识别平台相关的充值交易
- 💰 **自动化充值处理**：识别 BNB 和 ERC20 代币充值，自动更新用户余额
- 🔒 **高并发安全**：通过多层锁机制确保资产安全
- 💾 **数据可靠性**：断点续传、事务保证、交易去重，确保数据不丢失

### 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.x |
| 区块链交互 | go-ethereum |
| 数据库 | MySQL |
| 缓存/锁 | Redis |
| ORM | GORM |
| 日志 | Logrus |

---

### 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                         cmd/syncblock                        │
│                      （程序入口）                             │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│  SyncBlock   │   │WithdrawHandler│  │ SyncWithdraw │
│  (区块同步)   │   │  (提现处理)   │  │  (提现同步)   │
└──────┬───────┘   └──────┬───────┘   └──────┬───────┘
       │                  │                  │
       └──────────────────┼──────────────────┘
                          │
        ┌─────────────────┴─────────────────┐
        │                                     │
        ▼                                     ▼
┌──────────────┐                     ┌──────────────┐
│   Service    │                     │  Repository  │
│ (业务逻辑层)  │                     │  (数据访问层) │
└──────────────┘                     └──────────────┘
        │                                     │
        └─────────────────┬─────────────────┘
                          │
        ┌─────────────────┴─────────────────┐
        │                                     │
        ▼                                     ▼
┌──────────────┐                     ┌──────────────┐
│  Ethereum    │                     │   MySQL +    │
│  RPC节点     │                     │   Redis      │
└──────────────┘                     └──────────────┘
```

### 核心组件

#### 1. SyncBlock（区块同步服务）
**职责**：扫描区块链区块，提取并处理平台相关交易

**核心功能**：
- 实时同步新区块
- 解析交易内容（BNB 转账、ERC20 Transfer 事件）
- 识别平台充值交易
- 更新用户资产余额

**关键代码**：[listener/sync_blockinfo.go](listener/sync_blockinfo.go)

#### 2. WithdrawHandler（提现处理服务）
**职责**：监听提现请求，处理链上提现交易

#### 3. SyncWithdrawHandler（提现确认服务）
**职责**：扫描链上提现交易，确认提现状态

### 数据流转

```
区块链产生区块
      │
      ▼
获取区块信息 (BlockByNumber)
      │
      ▼
遍历区块交易
      │
      ├─→ 不是平台地址 ──→ 跳过
      │
      ├─→ 是平台地址
      │     │
      │     ▼
      │  检查发送者是否为平台用户
      │     │
      │     ├─→ 不是 ──→ 跳过
      │     │
      │     └─→ 是用户
      │           │
      │           ▼
      │      获取交易锁 (Redis)
      │           │
      │           ▼
      │      开启数据库事务
      │           │
      │           ├─→ 交易去重检查
      │           ├─→ 查询账户余额
      │           ├─→ 计算新余额
      │           ├─→ 创建账单记录
      │           ├─→ 创建交易日志
      │           ├─→ 更新资产余额（乐观锁）
      │           │
      │           ▼
      │      提交事务
      │           │
      │           ▼
      │      释放锁
      │
      ▼
保存同步进度 (last_synced_block.txt)
```

---

## 并发处理机制

### 核心挑战

一个区块可能包含数百笔交易，如果全部并发处理会产生问题：
- 过多 goroutine 消耗大量资源
- 同一账户的多个交易并发修改余额导致数据不一致
- 多个进程/实例同时处理同一笔交易

### 三层并发控制

#### 第一层：Worker Pool（工作池）

**目的**：限制并发 goroutine 数量，防止资源耗尽

**实现原理**：使用带缓冲的 channel 作为信号量

```go
// listener/sync_blockinfo.go:33
workerPool chan struct{}  // 工作池控制并发数量

// 初始化
workerPool: make(chan struct{}, config.Sync.Workers)

// 获取槽位（阻塞直到有空闲）
s.workerPool <- struct{}{}

// 释放槽位
<-s.workerPool
```

**工作流程**：

```
配置：Workers = 10

交易1: workerPool <- struct{}{}  成功（占用1个槽位）
交易2: workerPool <- struct{}{}  成功（占用2个槽位）
...
交易10: workerPool <- struct{}{} 成功（占用10个槽位）
交易11: workerPool <- struct{}{} 阻塞（等待槽位）
                                      ↓
                              交易1完成，释放槽位
                                      ↓
                              交易11获取槽位，继续执行
```

**关键代码**：[listener/sync_blockinfo.go:243](listener/sync_blockinfo.go#L243)

---

#### 第二层：分布式锁（Redis）

**目的**：防止多个进程/goroutine 同时修改同一账户的同一资产

**实现原理**：基于 Redis SETNX 命令

```go
// common/redis/lock_manager.go:42
lockKey := fmt.Sprintf("asset_lock:%d:%d", accountId, tokenType)
// 示例：asset_lock:1:1 表示账户1的代币1
```

**加锁流程**：

```lua
-- Redis SETNX 命令（原子操作）
SETNX asset_lock:1:1 "random_uuid" EX 10

-- 返回 true：加锁成功
-- 返回 false：锁已存在（其他进程持有）
```

**解锁流程**（Lua 脚本保证原子性）：

```lua
-- common/redis/lock_manager.go:149
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end
```

**为什么需要随机值？**

```
进程A: 加锁 asset_lock:1:1 = "uuid_A"
进程A: 处理超时（超过10秒）
Redis: 锁自动过期
进程B: 加锁 asset_lock:1:1 = "uuid_B"
进程A: 处理完成，尝试释放锁
      Lua: if GET("uuid_A") == "uuid_B" ?
      → 不匹配，不删除 ✅
```

**关键代码**：[common/redis/lock_manager.go:117](common/redis/lock_manager.go#L117)

---

#### 第三层：乐观锁（数据库）

**目的**：数据库层面的最后防线，检测并发修改

**实现原理**：version 字段版本号检查

```sql
-- repository/account_repo.go:123 实际执行的 SQL
UPDATE account_assets
SET bnb_balance = ?,
    mtk_balance = ?,
    version = 6,           -- 新版本号
    updated_at = ?
WHERE account_id = 1
  AND version = 5          -- 旧版本号（关键检查）
```

**并发冲突场景**：

```
时间线：
进程A                         进程B
────────────────────────────────────────
读取: version=5, balance=100
                              读取: version=5, balance=100
                              更新: WHERE version=5 ✅ 成功
                              数据库: version=6, balance=120
更新: WHERE version=5 ❌ 失败
      (version 已经是6了)
检测: RowsAffected = 0
返回错误: "optimistic lock failed"
```

**关键代码**：[repository/account_repo.go:122](repository/account_repo.go#L122)

---

### 三层防护协同工作

```
┌─────────────────────────────────────────────────────┐
│ Layer 1: Worker Pool                                 │
│ 限制总并发数，防止资源耗尽                            │
│ 示例：最多10个goroutine同时处理交易                  │
└────────────────────────┬────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│ Layer 2: 分布式锁 (Redis)                            │
│ 防止同一账户的同一资产被并发修改                      │
│ 示例：account_id=1, token_type=1 只能串行处理        │
└────────────────────────┬────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│ Layer 3: 乐观锁 (数据库)                             │
│ 最后防线，检测并拒绝并发修改                         │
│ 示例：version 不匹配时拒绝更新                       │
└─────────────────────────────────────────────────────┘
```

**为什么需要三层？**

| 场景 | Worker Pool | 分布式锁 | 乐观锁 |
|------|-------------|----------|--------|
| 单机多 goroutine | ✅ 有效 | ✅ 有效 | ✅ 有效 |
| 多机多实例 | ❌ 无效 | ✅ 有效 | ✅ 有效 |
| 分布式锁失效 | ❌ 无效 | ❌ 无效 | ✅ 有效 |

---

## 数据可靠性保证

### 1. 断点续传

**问题**：服务重启后，如何从断点继续同步？

**解决方案**：持久化同步进度到文件

```go
// repository/block_sync_repo.go
type BlockSyncManager struct {
    lastSyncedFile string  // 存储最后同步区块高度
}

// 每处理完一个区块就保存进度
func (m *BlockSyncManager) SaveSyncedBlock(height uint64) error {
    heightStr := strconv.FormatUint(height, 10)
    return os.WriteFile(m.lastSyncedFile, []byte(heightStr), 0644)
}

// 启动时加载上次同步位置
func (m *BlockSyncManager) GetLastSyncedBlock() (uint64, error) {
    data, err := os.ReadFile(m.lastSyncedFile)
    // ...
    return strconv.ParseUint(string(data), 10, 64)
}
```

**关键代码**：[repository/block_sync_repo.go](repository/block_sync_repo.go)

**工作流程**：

```
处理区块 12345
      │
      ▼
保存进度到 last_synced_block.txt
      │
      ▼
服务异常重启
      │
      ▼
启动时读取 last_synced_block.txt
      │
      ▼
从区块 12346 继续同步 ✅
```

---

### 2. 事务完整性

**问题**：多个数据库操作如何保证原子性？

**解决方案**：使用 GORM 事务

```go
// repository/sync_block_repo.go
func TxWithTransaction(fn func(txRepo *TxRepository) error) error {
    return adapter.DB.Transaction(func(tx *gorm.DB) error {
        txRepo := &TxRepository{db: tx}
        return fn(txRepo)
        // 成功：自动提交
        // 失败：自动回滚
    })
}
```

**事务内操作序列**：

```go
// listener/sync_blockinfo.go:422
func (s *SyncBlock) executeTransactionWithLock(...) error {
    return repository.TxWithTransaction(func(txRepo *TxRepository) error {
        // 1. 交易去重检查
        isExistTx, err := txRepo.TransactionExists(hash)
        if isExistTx {
            return fmt.Errorf("transaction existed")
        }

        // 2. 查询账户资产（加行锁）
        asset, err := txRepo.GetAssetByAccountIdWithLock(accountId)

        // 3. 计算新余额
        preBalance, nextBalance, err := s.calculateBalance(...)

        // 4. 创建账单记录
        bill := model.Bill{...}
        txRepo.AddBill(&bill)

        // 5. 创建交易日志
        transLog := model.TransactionLog{...}
        txRepo.AddTransactionLog(&transLog)

        // 6. 更新资产余额（乐观锁）
        txRepo.UpdateAssetWithOptimisticLock(asset, nextBalance, tokenType)

        return nil  // 全部成功，提交事务
    })
}
```

**关键代码**：[listener/sync_blockinfo.go:422](listener/sync_blockinfo.go#L422)

---

### 3. 交易去重

**问题**：如何避免同一笔交易被重复处理？

**解决方案**：通过交易哈希查询

```go
// repository/transaction_log_repo.go
func (t *TxRepository) TransactionExists(hash string) (bool, error) {
    var count int64
    err := t.db.Model(&model.TransactionLog{}).
        Where("hash = ?", hash).
        Count(&count).Error
    return count > 0, err
}
```

**去重时机**：在事务开始时首先检查

```
处理交易 0xabc...
      │
      ▼
查询数据库：SELECT COUNT(*) WHERE hash = '0xabc...'
      │
      ├─→ count > 0
      │     │
      │     ▼
      │  返回错误："transaction existed" ❌
      │
      └─→ count = 0
            │
            ▼
         继续处理 ✅
```

---

### 4. 错误重试

**问题**：网络异常、节点超时等临时故障如何处理？

**解决方案**：延迟重试机制

```go
// listener/sync_blockinfo.go:135
func (s *SyncBlock) syncLoop() {
    for atomic.LoadInt32(&s.isSyncRunning) == 1 {
        // ...

        // 处理当前区块
        if err := s.processBlock(...); err != nil {
            s.log.Error("Failed to process block")
            time.Sleep(20 * time.Second)  // 延迟后重试
            continue
        }

        // 更新并保存进度
        s.setDoneBlock(newDoneBlock)
        s.saveSyncedBlock(newDoneBlock)
    }
}
```

**重试策略**：

| 场景 | 处理方式 |
|------|----------|
| 获取区块失败 | 等待10秒后重试 |
| 处理区块失败 | 等待20秒后重试 |
| 数据库事务失败 | 整个区块重新处理 |
| Redis 锁获取失败 | 等待锁释放（最多10秒） |

---

### 5. 结构化日志

**目的**：完整记录处理链路，便于故障排查和审计

**日志实现**：

```go
// common/logger/logger.go
log := logrus.New()
log.SetFormatter(&logrus.JSONFormatter{})
```

**日志记录示例**：

```go
// 成功日志
s.log.WithFields(logrus.Fields{
    "module":         "sync_block",
    "action":         "save_synced_block",
    "block":          doneBlock,
    "new_done_block": newDoneBlock,
    "result":         "success",
}).Info("Processed block")

// 错误日志
s.log.WithFields(logrus.Fields{
    "module":     "sync_block",
    "action":     "process_transaction",
    "tx_hash":    tx.Hash().Hex(),
    "error_code": "PROCESS_TRANSACTION_FAIL",
    "detail":     err.Error(),
}).Error("Failed to process transaction")
```

**日志策略**：

| 级别 | 用途 | 存储位置 |
|------|------|----------|
| Info | 正常处理流程 | info.log |
| Warn | 跳过处理的交易 | info.log |
| Error | 处理失败（已重试） | error.log |
| Fatal | 无法恢复的错误 | error.log + 终止程序 |

**日志分割**：
- 按日期自动切割（每天一个文件）
- 保留期可配置（默认7天）

---

### 数据可靠性总结

| 机制 | 解决的问题 | 实现方式 |
|------|-----------|----------|
| **断点续传** | 服务重启后继续同步 | 文件持久化同步进度 |
| **事务完整性** | 多个操作原子性 | GORM 事务 |
| **交易去重** | 避免重复处理 | 交易哈希查询 |
| **错误重试** | 临时故障自动恢复 | 延迟重试 |
| **结构化日志** | 故障排查、审计 | Logrus + 日志分割 |
| **三层锁机制** | 并发安全 | Worker Pool + 分布式锁 + 乐观锁 |

---

## 快速开始

### 环境要求

- Go 1.x
- MySQL 8.0+
- Redis 6.0+
- Ethereum RPC 节点访问权限

### 配置文件

编辑 `config/config.yaml`：

```yaml
blockchain:
  rpc_url: "https://bsc-dataseed.binance.org"
  chain_id: 56
  contracts:
    token: "0x..."  # ERC20 代币合约地址

sync:
  workers: 10         # 并发处理数量
  batch_size: 100
  block_buffer: 12    # 同步距离控制
  sync_interval: 10s
```

### 启动服务

```bash
# 编译
go build -o syncblock cmd/syncblock/main.go

# 运行
./syncblock
```

### 停止服务

```bash
# 发送 SIGTERM 信号
kill -TERM <pid>

# 或 Ctrl+C
```

服务会优雅停止，等待所有处理中的交易完成。

---

## 项目结构

```
go-staking-interaction/
├── cmd/
│   └── syncblock/
│       └── main.go           # 程序入口
├── listener/
│   ├── sync_blockinfo.go     # 区块同步服务
│   ├── withdraw_handler.go   # 提现处理服务
│   └── sync_withdraw.go      # 提现确认服务
├── service/
│   └── transaction.go        # 交易服务
├── repository/
│   ├── account_repo.go       # 账户数据访问
│   ├── transaction_log_repo.go
│   └── block_sync_repo.go    # 区块同步进度
├── common/
│   ├── config/
│   ├── logger/
│   └── redis/
│       └── lock_manager.go   # 分布式锁
├── model/                    # 数据模型
├── dto/                      # 数据传输对象
└── adapter/                  # 外部服务适配器
```

---

## 许可证

MIT License

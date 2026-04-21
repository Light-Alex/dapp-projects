# Chain Parse Service - 编码规范与约定

本文档定义了 Chain Parse Service 项目的编码规范、约定和最佳实践。

## 目录

1. [Go 代码风格](#go-代码风格)
2. [错误处理](#错误处理)
3. [日志记录](#日志记录)
4. [测试](#测试)
5. [Git 提交](#git-提交)
6. [添加新链/DEX](#添加新链dex)
7. [代码审查清单](#代码审查清单)

---

## Go 代码风格

### 命名约定

#### 函数和方法
- 导出函数（公共）使用 **CamelCase**（大驼峰）
- 未导出函数（私有）使用 **camelCase**（小驼峰）
- 使用描述性名称来表明用途

```go
// ✓ 正确
func (e *UniswapExtractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {}
func (p *PancakeSwapExtractor) parseV2Swap(log *ethtypes.Log) *model.Transaction {}
func getEventType(topic string) string {}

// ✗ 错误
func (e *UniswapExtractor) extract(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {}
func (p *PancakeSwapExtractor) parse(log *ethtypes.Log) *model.Transaction {}
func getType(t string) string {}
```

#### 变量和常量
- 变量使用 **camelCase**（小驼峰）
- 包级别常量使用 **UPPER_SNAKE_CASE**（大写下划线）
- 使用描述性名称；避免单字母变量，索引除外

```go
// ✓ 正确
const (
    PANCAKE_SWAP_V2_FACTORY_ADDR = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
    SWAP_EVENT_SIGNATURE = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
)

var supportedChains []types.ChainType

// ✗ 错误
const PANCAKE_V2 = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
var s []types.ChainType
for i, v := range items {} // 简单循环中可以接受
```

#### 接口
- 使用以 `-er` 或 `-or` 结尾的描述性名称

```go
// ✓ 正确
type DexExtractors interface {}
type ChainProcessor interface {}
type StorageEngine interface {}

// ✗ 错误
type DEXer interface {}
type Processor interface {}
```

#### 包名
- 使用简短的小写名称
- 避免在包名中使用下划线
- 集合包使用复数形式

```
internal/
├── parser/
│   ├── chains/    # 链处理器
│   ├── dexs/      # DEX 提取器
│   └── engine/    # 处理引擎
├── storage/       # 存储接口和实现
├── types/         # 类型定义
├── errors/        # 错误处理
└── logger/        # 日志工具
```

### 导入组织

将导入分为三组，用空行分隔：

```go
package dex

import (
	// 标准库
	"context"
	"fmt"
	"math/big"
	"strings"

	// 外部依赖
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	// 本地导入
	"unified-tx-parser/internal/model"
	"unified-tx-parser/internal/types"
	"unified-tx-parser/internal/utils"
)
```

**规则：**
1. 标准库导入在前
2. 外部依赖导入在后
3. 本地项目导入最后
4. 每组内按字母顺序排列
5. 为长导入路径使用有意义的别名

### 代码组织

#### 文件结构
- 每个文件一个结构体（除非它们紧密相关）
- 将相关方法分组在一起
- 将接口实现放在文件顶部

```go
// pancakeswap.go
package dex

type PancakeSwapExtractor struct { ... }

// 接口实现
func (p *PancakeSwapExtractor) GetSupportedProtocols() []string { ... }
func (p *PancakeSwapExtractor) GetSupportedChains() []types.ChainType { ... }
func (p *PancakeSwapExtractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) { ... }
func (p *PancakeSwapExtractor) SupportsBlock(block *types.UnifiedBlock) bool { ... }

// 私有方法
func (p *PancakeSwapExtractor) parseV2Swap(log *ethtypes.Log) *model.Transaction { ... }
func (p *PancakeSwapExtractor) parseV3Swap(log *ethtypes.Log) *model.Transaction { ... }
```

#### 结构体字段
- 公共字段在顶部
- 私有字段在底部
- 将相关字段分组
- 使用有意义的字段名

```go
// ✓ 正确
type UniswapExtractor struct {
	// 接口实现字段
	supportedChains []types.ChainType
	protocols       []string

	// 配置
	factoryAddr string
	routerAddr  string
	eventSigs   map[string]string

	// 运行时状态
	log     *logrus.Entry
	cache   *TokenCache
	mutex   sync.RWMutex
}

// ✗ 错误
type UniswapExtractor struct {
	fc string          // 难懂的缩写
	ra string          // 不清楚这是什么
	evts map[string]string // 非标准命名
	logger *logrus.Entry
	c *TokenCache
	m sync.RWMutex
}
```

### 代码格式化

- 使用 `gofmt` 进行自动格式化
- 行长度：最大 120 个字符（可读性优先于严格限制）
- 使用空行分隔逻辑部分
- 在运算符周围使用有意义的空格

```go
// ✓ 正确
result := new(big.Int).Add(amount0, amount1)
price := new(big.Float).Quo(out, in).Float64()

if len(log.Data) < 128 {
	return nil
}

// ✗ 错误
result:=new(big.Int).Add(amount0,amount1)
price:=new(big.Float).Quo(out,in).Float64()

if len(log.Data)<128{return nil}
```

---

## 错误处理

### 一般原则

1. **始终检查错误返回** - 不允许未检查的错误
2. **用上下文包装错误** - 使用 `fmt.Errorf` 配合 `%w` 动词
3. **尽早返回** - 立即检查错误，不要嵌套
4. **不要隐藏错误** - 返回错误时在适当级别记录日志
5. **需要时使用自定义错误类型** 进行特定错误处理

### 错误包装模式

```go
// ✓ 正确 - 用上下文包装
func (e *Extractor) GetTokenMetadata(ctx context.Context, addr string) (*model.Token, error) {
	token, err := e.fetchFromChain(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token metadata for %s: %w", addr, err)
	}
	return token, nil
}

// ✗ 错误 - 丢失错误上下文
func (e *Extractor) GetTokenMetadata(ctx context.Context, addr string) (*model.Token, error) {
	token, err := e.fetchFromChain(addr)
	if err != nil {
		return nil, err  // 没有关于什么失败的上下文
	}
	return token, nil
}

// ✗ 错误 - 隐藏错误
func (e *Extractor) GetTokenMetadata(ctx context.Context, addr string) (*model.Token, error) {
	token, err := e.fetchFromChain(addr)
	if err != nil {
		e.log.Errorf("failed: %v", err)
		// 错误被静默忽略，返回 nil 而没有错误
		return nil, nil
	}
	return token, nil
}
```

### Panic 与错误

- **仅在以下情况使用 panic**：
  - 初始化期间的编程错误
  - 阻止启动的系统配置问题
  - 关于不可变不变量的断言

- **绝不要在以下情况 panic**：
  - 事件处理逻辑
  - 网络操作
  - 用户输入处理

```go
// ✓ 正确 - 初始化期间因配置问题而 panic
func NewDEXExtractor(cfg Config) *DEXExtractor {
	if cfg.FactoryAddr == "" {
		panic("factory address must be configured")
	}
	// ...
}

// ✓ 正确 - 为操作失败返回错误
func (e *DEXExtractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {
	if blocks == nil {
		return nil, fmt.Errorf("blocks cannot be nil")
	}
	// ...
}

// ✗ 错误 - 在操作代码中 panic
func (e *DEXExtractor) parseLog(log *ethtypes.Log) *model.Transaction {
	if log == nil {
		panic("log cannot be nil")  // 应该返回错误
	}
	// ...
}
```

### 自定义错误类型

为需要特定处理的错误使用自定义错误类型：

```go
// ✓ 正确
type ParseError struct {
	LogIndex int
	Reason   string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error at log %d: %s", e.LogIndex, e.Reason)
}

// 使用
if err := parseLog(log); err != nil {
	if parseErr, ok := err.(ParseError); ok {
		e.log.Warnf("recoverable parse error: %v", parseErr)
		continue
	}
	return nil, err
}
```

---

## 日志记录

### 日志设置

- 使用 logrus 进行结构化日志记录
- 始终包含 `service` 和 `module` 字段
- 使用适当的日志级别

```go
// 模块日志
var extractorLog = logrus.WithFields(logrus.Fields{
	"service": "parser",
	"module":  "dex-extractor",
})

// 在方法中
func (e *Extractor) SomeMethod() {
	e.log.WithFields(logrus.Fields{
		"tx_hash": tx.TxHash,
		"block":   block.Number.String(),
	}).Infof("processing transaction")
}
```

### 日志级别

- **Debug**：详细的诊断信息（变量值、中间步骤）
- **Info**：一般信息消息（已处理交易、已同步区块）
- **Warn**：警告条件（解析错误、重试、意外值）
- **Error**：错误条件（失败的操作，需要调查）

```go
// ✓ 正确
e.log.Debugf("parsed amount: %s, price: %.6f", amount.String(), price)
e.log.Infof("processed %d transactions in block %d", len(txs), blockNum)
e.log.Warnf("swap log data too short: %d bytes, expected >=128", len(log.Data))
e.log.Errorf("failed to fetch token metadata: %w", err)

// ✗ 错误
e.log.Infof("amount: %s, price: %.6f, data: %v", amount, price, someData)  // 对 Info 来说太冗长
e.log.Warnf("retry attempt %d", attempt)  // 应该是 Debug
e.log.Errorf("chunk size exceeded")  // 太模糊
```

### 结构化日志

为每个日志添加相关上下文：

```go
// ✓ 正确
e.log.WithFields(logrus.Fields{
	"tx_hash":     tx.TxHash,
	"block":       block.BlockNumber.String(),
	"extractor":   "pancakeswap",
	"event_type":  "swap",
	"log_index":   logIdx,
	"swap_index":  swapIdx,
}).Infof("processing swap event")

// ✗ 错误
e.log.Infof("processing transaction")  // 没有上下文
```

### 生产环境日志

在生产环境中，避免过多的日志记录：

```go
// ✓ 正确 - 只记录重要事件
func (e *Extractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			ethLogs := e.ExtractEVMLogs(&tx)
			// 不要记录每个交易，只在错误或重要事件时记录
			if len(ethLogs) > 0 {
				e.log.Debugf("found %d eth logs in tx %s", len(ethLogs), tx.TxHash)
			}
		}
	}
}

// ✗ 错误 - 日志过多
func (e *Extractor) ExtractDexData(ctx context.Context, blocks []types.UnifiedBlock) (*types.DexData, error) {
	for _, block := range blocks {
		e.log.Infof("processing block %d", block.BlockNumber.Int64())
		for i, tx := range block.Transactions {
			e.log.Infof("transaction %d/%d (hash: %s)", i+1, len(block.Transactions), tx.TxHash)
			ethLogs := e.ExtractEVMLogs(&tx)
			for _, log := range ethLogs {
				e.log.Infof("found log at index %d", log.Index)
			}
		}
	}
}
```

---

## 测试

### 测试文件组织

- 将测试放在同一包的 `*_test.go` 文件中
- 每个源文件或逻辑组一个测试文件
- 对多个场景使用表驱动测试

```
pancakeswap.go        → pancakeswap_test.go
uniswap.go            → uniswap_test.go
utils.go              → utils_test.go
```

### 测试函数命名

```go
// ✓ 正确
func TestPancakeSwapExtractor_ExtractDexData(t *testing.T) {}
func TestPancakeSwapExtractor_ParseV2Swap(t *testing.T) {}
func TestCalcPrice_WithValidInputs(t *testing.T) {}
func TestCalcPrice_WithZeroAmount(t *testing.T) {}

// ✗ 错误
func TestPancakeSwap(t *testing.T) {}  // 太模糊
func TestExtract(t *testing.T) {}       // 不清楚在测试什么
```

### 表驱动测试

对多个输入场景使用表驱动测试：

```go
// ✓ 正确
func TestCalcPrice(t *testing.T) {
	tests := []struct {
		name        string
		amountIn    *big.Int
		amountOut   *big.Int
		expected    float64
		shouldError bool
	}{
		{
			name:        "正常计算",
			amountIn:    big.NewInt(1000000),
			amountOut:   big.NewInt(2000000),
			expected:    2.0,
			shouldError: false,
		},
		{
			name:        "零输入",
			amountIn:    big.NewInt(0),
			amountOut:   big.NewInt(1000000),
			expected:    0,
			shouldError: false,
		},
		{
			name:        "空输入",
			amountIn:    nil,
			amountOut:   nil,
			expected:    0,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalcPrice(tt.amountIn, tt.amountOut)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}
```

### Mock 对象

对外部依赖使用基于接口的 mock：

```go
// ✓ 正确
type MockTokenProvider interface {
	GetToken(ctx context.Context, addr string) (*model.Token, error)
}

type mockTokenProvider struct {
	tokens map[string]*model.Token
}

func (m *mockTokenProvider) GetToken(ctx context.Context, addr string) (*model.Token, error) {
	if token, ok := m.tokens[addr]; ok {
		return token, nil
	}
	return nil, fmt.Errorf("token not found")
}

// 在测试中使用
func TestExtractor_WithMockProvider(t *testing.T) {
	mockProvider := &mockTokenProvider{
		tokens: map[string]*model.Token{
			"0xtoken": {Name: "Test Token", Symbol: "TEST"},
		},
	}
	// 使用 mock 进行测试
}
```

### 测试断言

使用清晰的断言消息：

```go
// ✓ 正确
if len(data.Transactions) != 1 {
	t.Errorf("expected 1 transaction, got %d", len(data.Transactions))
}

if data.Transactions[0].Price != expectedPrice {
	t.Errorf("expected price %.6f, got %.6f", expectedPrice, data.Transactions[0].Price)
}

// 考虑使用第三方断言库
require.NoError(t, err)
require.Equal(t, expectedValue, actualValue)
assert.True(t, condition, "condition should be true")
```

### 测试隔离

每个测试应该是独立的：

```go
// ✓ 正确
func TestExtractor_CacheBehavior(t *testing.T) {
	// 为此测试创建新的提取器
	extractor := NewDexExtractorWithCache()
	defer extractor.cache.Clear()

	// 测试缓存操作
}

// ✗ 错误 - 测试依赖执行顺序
var globalCache = NewTokenCache(time.Hour)

func TestCacheSet(t *testing.T) {
	globalCache.Set("key", token)
}

func TestCacheGet(t *testing.T) {
	// 依赖于 TestCacheSet 先运行！
	token, ok := globalCache.Get("key")
}
```

---

## Git 提交

### 提交消息格式

遵循约定式提交格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 类型

- **feat**：新功能
- **fix**：Bug 修复
- **refactor**：代码重构，不改变功能
- **test**：测试添加或修改
- **docs**：文档变更
- **chore**：构建、CI 或依赖变更
- **perf**：性能改进
- **style**：格式化、缺少分号等

### 范围

指定变更区域：
- 包名：`dex`、`parser`、`storage`
- 功能：`pancakeswap`、`uniswap`、`cache`
- 组件：`base-extractor`、`evm-logs`

### 主题

- 使用祈使语气："add"、"fix"、"implement"（不是 "added"、"fixed"、"implementing"）
- 不要首字母大写
- 结尾不加句号
- 最多 50 个字符

### 正文

- 解释 **是什么** 和 **为什么**，而不是 **怎么做**
- 72 个字符换行
- 与主题用空行分隔
- 多个变更使用项目符号

### 示例

```
✓ 正确：
feat(dex): add FourMeme DEX support with V1/V2 event parsing
- Implement V2 TokenCreate, TokenPurchase, TokenSale events
- Support V1 event format with 128-byte layout
- Add LiquidityAdded event for graduation tracking
- Include quote asset configuration for price calculation

fix(parser): correct Uniswap V3 swap amount handling
- Use toSignedInt256 for signed amount0/amount1 parsing
- Previous code treated negative amounts as large positives
- Fixes price inversion on sell-side swaps

✗ 错误：
feat: add stuff
Fixed uniswap parsing
MAJOR CHANGES
Update code
```

### 提交指南

1. **原子提交**：每个提交应该在逻辑上是独立的
2. **不混合关注点**：不要在同一提交中重构和添加功能
3. **包含相关测试**：如果添加功能，包含测试
4. **引用问题**：适用时添加 `Fixes #123` 或 `Closes #456`

```
fix(uniswap): handle pool address parsing in PoolCreated event

Previous implementation used log.Address (Factory) instead of
parsing actual pool address from event data, causing all pools
to be attributed to the factory.

Fixes #42
```

---

## 添加新链/DEX

### 添加新 DEX 的检查清单

#### 1. 协议分析
- [ ] 记录协议地址（factory、router 等）
- [ ] 识别要解析的所有事件类型（Swap、Mint、Burn、PoolCreated 等）
- [ ] 记录事件签名和参数布局
- [ ] 理解代币对表示（token0/token1 顺序）
- [ ] 识别任何特殊处理（V2 vs V3、有符号 vs 无符号）

#### 2. 实现
- [ ] 创建 `{protocol}.go` 提取器文件
- [ ] 定义扩展适当基类的提取器结构体
- [ ] 实现 `DexExtractors` 接口：
  - [ ] `GetSupportedProtocols()` - 返回协议名称
  - [ ] `GetSupportedChains()` - 返回支持的链
  - [ ] `ExtractDexData()` - 主要提取逻辑
  - [ ] `SupportsBlock()` - 检查区块是否包含协议事件
- [ ] 实现事件解析方法
- [ ] 在 `extractor_factory.go` 中注册

#### 3. 测试
- [ ] 创建 `{protocol}_test.go` 测试文件
- [ ] 为每个事件类型添加单元测试：
  - [ ] 具有有效数据的正常情况
  - [ ] 边界情况（零数量、最大值）
  - [ ] 错误情况（格式错误的数据、截断的日志）
- [ ] 使用真实区块数据添加集成测试
- [ ] 使用 `go test -v -race ./...` 通过测试套件

#### 4. 文档
- [ ] 将特定协议的文档添加到 `ARCHITECTURE.md`
- [ ] 记录事件签名和布局
- [ ] 为非显而易见的逻辑添加代码注释
- [ ] 在提交消息中包含示例

#### 5. 代码审查
- [ ] 通过 linter：`go vet ./...`
- [ ] 代码遵循项目约定
- [ ] 所有测试通过
- [ ] 未添加不必要的依赖
- [ ] 错误处理完整

### 逐步示例：添加 SushiSwap

#### 步骤 1：协议分析
```go
// 分析和记录协议
const (
	// SushiSwap AMM router
	SUSHISWAP_ROUTER_ADDR = "0xd9e1cE17f2641f24aE5D4d5a9f2779199aa8aBEA"

	// 事件签名
	SUSHISWAP_SWAP_EVENT_SIG = "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822"
	// ... 其他事件
)
```

#### 步骤 2：创建提取器
```go
// internal/parser/dexs/sushiswap.go
type SushiSwapExtractor struct {
	*EVMDexExtractor
	routerAddr string
}

func NewSushiSwapExtractor() *SushiSwapExtractor {
	cfg := &BaseDexExtractorConfig{
		Protocols:       []string{"sushiswap"},
		SupportedChains: []types.ChainType{types.ChainTypeEthereum, types.ChainTypeBSC},
		LoggerModuleName: "dex-sushiswap",
	}

	return &SushiSwapExtractor{
		EVMDexExtractor: NewEVMDexExtractor(cfg),
		routerAddr:      SUSHISWAP_ROUTER_ADDR,
	}
}

// 实现接口方法...
```

#### 步骤 3：编写测试
```go
// internal/parser/dexs/sushiswap_test.go
func TestSushiSwapExtractor_ParseSwap(t *testing.T) {
	tests := []struct {
		name string
		logData []byte
		expected *model.Transaction
	}{
		// 表驱动测试...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试实现...
		})
	}
}
```

#### 步骤 4：注册提取器
```go
// internal/parser/dexs/extractor_factory.go
func CreateDefaultFactory() *ExtractorFactory {
	factory := NewExtractorFactory()
	// ... 现有提取器 ...
	factory.RegisterExtractor("sushiswap", NewSushiSwapExtractor())
	return factory
}
```

#### 步骤 5：提交
```
feat(dex): add SushiSwap DEX support

- Implement Swap event parsing for both V1 and V2 protocols
- Add Mint/Burn liquidity event handling
- Support PoolCreated events for pool tracking
- Include quote asset management for stablecoin prices

Supports Ethereum and BSC chains.
```

---

## 代码审查清单

### 提交审查前

- [ ] 代码编译无警告：`go build ./...`
- [ ] 所有测试通过：`go test -v ./...`
- [ ] 无竞态条件：`go test -race ./...`
- [ ] 代码已格式化：`go fmt ./...`
- [ ] Linter 通过：`go vet ./...`
- [ ] 提交消息遵循约定
- [ ] 没有无上下文的 TODO/FIXME 注释
- [ ] 包含相关测试
- [ ] 错误处理完整
- [ ] 未添加不必要的依赖

### 审查期间

#### 功能性
- [ ] 代码是否声称做它应该做的事情？
- [ ] 是否处理了所有边界情况？
- [ ] 错误情况是否得到适当处理？
- [ ] 逻辑是否正确？

#### 代码质量
- [ ] 代码是否清晰易懂？
- [ ] 是否遵循项目约定？
- [ ] 变量名是否有意义？
- [ ] 代码是否 DRY（不重复）？
- [ ] 是否有不必要的注释？

#### 性能
- [ ] 是否有任何明显的性能问题？
- [ ] 是否正确使用了缓存？
- [ ] 是否有不必要的分配？
- [ ] 并发处理是否正确？

#### 测试
- [ ] 是否测试了所有关键路径？
- [ ] 是否测试了边界情况？
- [ ] 测试是否独立和隔离？
- [ ] 测试是否提供良好的覆盖率？

#### 文档
- [ ] 是否解释了复杂算法？
- [ ] 是否记录了公共 API？
- [ ] 是否注意了潜在的陷阱？
- [ ] 相关文档是否反映了变更？

### 审查模板

```markdown
## 摘要
变更的简要描述

## 变更
- 变更 1
- 变更 2

## 测试
- [ ] 本地测试
- [ ] 所有测试通过
- [ ] 无竞态条件

## 备注
- 考虑 X 以供将来改进
- 选择这种方法是因为 Y

## 批准
- [ ] 批准
- [ ] 需要更改
```

### 评论风格

```markdown
✓ 正确：
为什么应该更改这个？
```
if err != nil {
    return fmt.Errorf("meaningful error context: %w", err)
}
```
这提供了错误上下文，有助于调试。

✗ 错误：
"更改这个"
"添加错误处理"
"这是错的"
```

---

## 项目结构总结

```
chain-parse-service/
├── internal/
│   ├── errors/              # 错误类型和处理
│   ├── logger/              # 日志配置
│   ├── model/               # 数据模型
│   ├── parser/
│   │   ├── chains/          # 链处理器（Sui、BSC、Ethereum、Solana）
│   │   ├── dexs/            # DEX 提取器
│   │   │   ├── base_extractor.go
│   │   │   ├── evm_extractor.go
│   │   │   ├── solana_extractor.go
│   │   │   ├── utils.go
│   │   │   ├── cache.go
│   │   │   ├── pancakeswap.go
│   │   │   ├── uniswap.go
│   │   │   └── ...
│   │   └── engine/          # 处理引擎
│   ├── storage/             # 存储层
│   ├── types/               # 类型定义和接口
│   └── utils/               # 通用工具
├── cmd/                     # 命令行应用程序
├── configs/                 # 配置文件
├── CODING_STANDARDS.md      # 英文版本文档
├── CODING_STANDARDS_CN.md   # 本文档
└── Makefile
```

---

## 风格指南快速参考

| 类别 | 标准 |
|------|------|
| **函数命名** | CamelCase（导出）、camelCase（私有） |
| **常量** | UPPER_SNAKE_CASE |
| **导入** | 分组：标准库、外部、本地 |
| **行长度** | 最多 120 字符 |
| **错误处理** | 始终用 fmt.Errorf 包装 |
| **日志级别** | Debug < Info < Warn < Error |
| **测试命名** | TestType_Method_Scenario |
| **提交** | type(scope): subject 格式 |
| **注释** | 解释为什么，而不是是什么 |
| **Panic** | 仅在初始化期间，不在操作中 |

---

## 额外资源

- [Go 代码审查评论](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [约定式提交](https://www.conventionalcommits.org/)
- [项目 ARCHITECTURE.md](./internal/parser/dexs/ARCHITECTURE.md)
- [项目 IMPLEMENTATION_GUIDE.md](./internal/parser/dexs/IMPLEMENTATION_GUIDE.md)

---

**最后更新**：2026-03-05
**版本**：1.0

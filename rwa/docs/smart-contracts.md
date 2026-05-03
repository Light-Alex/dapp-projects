# Anchored Finance 智能合约架构文档

## 1. 合约架构总览

本项目包含两套合约体系：**POC（概念验证）合约** 和 **生产级 Gate 合约**。POC 合约用于快速验证核心业务流程，生产级合约引入了 pending 状态机制，适用于正式环境。

### 1.1 合约间关系图

```mermaid
graph TB
    subgraph POC合约体系
        MockUSDC["MockUSDC: 测试用USDC 6位精度"]
        USDM["PocToken-USDM: ERC20+RBAC 18位精度"]
        SymbolToken["PocToken-股票代币: 如AAPL Token"]
        PocGate["PocGate: 充值/提现网关"]
        OrderContract["OrderContract: 订单管理"]

        User((用户))

        User -->|1.approve USDC| MockUSDC
        User -->|2.deposit| PocGate
        PocGate -->|transferFrom USDC| MockUSDC
        PocGate -->|mint USDM| USDM

        User -->|3.approve USDM/股票代币| USDM
        User -->|4.submitOrder| OrderContract
        OrderContract -->|transferFrom托管资产| USDM
        OrderContract -->|transferFrom托管资产| SymbolToken

        Backend((后端))
        Backend -->|markExecuted| OrderContract
        Backend -->|cancelOrder| OrderContract
        Backend -->|mint股票代币| SymbolToken
    end

    subgraph 生产级合约体系
        USDC_Prod["USDC: 真实USDC"]
        AncUSDC["ancUSDC: 锚定USDC代币"]
        PendingAncUSDC["PendingAncUSDC: 待处理ancUSDC"]
        PendingUSDC_Token["PendingUSDC: 待处理USDC"]
        Gate["Gate: 生产级网关"]

        User2((用户))
        Processor((处理器/后端))

        User2 -->|deposit| Gate
        Gate -->|transferFrom USDC| USDC_Prod
        Gate -->|mint| PendingAncUSDC

        User2 -->|withdraw| Gate
        Gate -->|burn ancUSDC| AncUSDC
        Gate -->|mint| PendingUSDC_Token

        Processor -->|processDeposit| Gate
        Gate -->|burn pendingAncUSDC| PendingAncUSDC
        Gate -->|mint ancUSDC| AncUSDC

        Processor -->|processWithdrawal| Gate
        Gate -->|burn pendingUSDC| PendingUSDC_Token
        Gate -->|transfer USDC| USDC_Prod
    end
```

---

## 2. POC 合约详解

### 2.1 OrderContract（Order.sol）

OrderContract 是核心订单管理合约，负责订单提交、资金托管、执行确认和取消退款。

#### 2.1.1 枚举定义

| 枚举 | 值 | 说明 |
|------|-----|------|
| `Side` | `Buy(0)`, `Sell(1)` | 订单方向：买入 / 卖出 |
| `OrderType` | `Market(0)`, `Limit(1)` | 订单类型：市价 / 限价 |
| `Status` | `Pending(0)`, `Executed(1)`, `CancelRequested(2)`, `Cancelled(3)` | 订单状态 |
| `TimeInForce` | `DAY(0)`, `GTC(1)`, `OPG(2)`, `IOC(3)`, `FOK(4)`, `GTX(5)`, `GTD(6)`, `CLS(7)` | 订单有效期策略 |

**TimeInForce 有效期策略说明**：

| 值 | 枚举 | 全称 | 说明 |
|----|------|------|------|
| 0 | `DAY` | Day Order | 当日有效。订单在当前交易日结束时自动过期，未成交部分自动取消。适用于日内交易策略。 |
| 1 | `GTC` | Good-Till-Cancelled | 撤销前有效。订单持续有效直到被手动取消或成交完成。适合长期挂单，但需注意避免遗忘订单。 |
| 2 | `OPG` | At-the-Opening | 开盘有效。订单仅在市场开盘时参与集合竞价，若开盘未能成交则自动取消。用于捕捉开盘价格波动。 |
| 3 | `IOC` | Immediate-or-Cancel | 立即成交否则取消。订单提交后必须立即成交可成交的部分，未成交部分立即取消，不会挂单等待。适合要求快速成交的场景。 |
| 4 | `FOK` | Fill-or-Kill | 全部成交否则取消。订单必须一次性全部成交，若无法完全满足则整笔订单取消。适合大额交易，避免部分成交带来的滑点风险。 |
| 5 | `GTX` | Good-Till-Crossing | 撤销前有效（仅限被动单）。订单只能作为 Maker 挂单，不会立即与现有订单撮合，确保不会产生 Taker 成交。适合希望赚取 Maker 返佣的策略。 |
| 6 | `GTD` | Good-Till-Date | 有效期至指定日期。订单持续有效直到指定的到期时间，到期后自动取消。兼具 GTC 的持续性和 DAY 的自动过期特性。 |
| 7 | `CLS` | Closing | 收盘有效。订单仅在临近收盘时参与交易，用于在收盘阶段执行交易策略。 |

#### 2.1.2 状态转换图

```mermaid
stateDiagram-v2
    [*] --> Pending: submitOrder()
    Pending --> Executed: markExecuted() [BACKEND_ROLE]
    Pending --> CancelRequested: cancelOrderIntent() [用户]
    Pending --> Cancelled: cancelOrder() [BACKEND_ROLE]
    CancelRequested --> Cancelled: cancelOrder() [BACKEND_ROLE]
    CancelRequested --> Executed: markExecuted() [BACKEND_ROLE]
    Executed --> [*]
    Cancelled --> [*]
```

**状态转换说明**：

订单从创建到终态共有 4 个状态和 2 个终态：

| 转换路径 | 触发函数 | 执行者 | 说明 |
|----------|----------|--------|------|
| `[*] → Pending` | `submitOrder()` | 用户 | 用户提交订单并托管资金，订单进入待处理状态 |
| `Pending → Executed` | `markExecuted()` | 后端（BACKEND_ROLE） | 后端确认订单已在 Alpaca 成交，可退还多余托管资金 |
| `Pending → CancelRequested` | `cancelOrderIntent()` | 订单所有者 | 用户主动发起取消意图，但需后端最终确认 |
| `Pending → Cancelled` | `cancelOrder()` | 后端（BACKEND_ROLE） | 后端直接取消未成交订单，退还全部托管资金 |
| `CancelRequested → Cancelled` | `cancelOrder()` | 后端（BACKEND_ROLE） | 后端确认用户取消请求，执行退款 |
| `CancelRequested → Executed` | `markExecuted()` | 后端（BACKEND_ROLE） | 用户请求取消时订单已在 Alpaca 成交，此时以执行结果为准 |

**关键设计要点**：

- **双路径取消**：用户无法直接取消订单，必须通过 `cancelOrderIntent` 表达意图，由后端最终执行 `cancelOrder`。这确保后端在链下有足够时间处理与 Alpaca 的交互，避免链上取消与链下执行产生冲突。
- **CancelRequested 竞态处理**：当用户发起取消后，后端仍可能调用 `markExecuted` 将订单标记为已成交（因为 Alpaca 侧可能已执行）。此时成交结果优先，用户的取消请求被覆盖，后端通过 `refundAmount` 退还多余资金。
- **终态不可逆**：`Executed` 和 `Cancelled` 为终态，进入后不可再转换。同一订单的 `markExecuted` 和 `cancelOrder` 互斥，只能调用其中之一。

#### 2.1.3 存储结构

```solidity
struct Order {
    uint id;              // 全局自增 ID（存储键）
    uint orderNumber;     // 展示用结构化订单号 AAAAAABBSSSSSSSSSS
    address user;         // 下单用户地址
    string symbol;        // 交易标的符号（如 "AAPL"）
    uint qty;             // 订单数量（18 decimals）
    address escrowAsset;  // 托管资产地址（USDM 或股票代币）
    uint amount;          // 托管金额（Buy: price*qty; Sell: qty）
    uint price;           // 价格（18 decimals）
    Side side;            // 买/卖方向
    OrderType orderType;  // 市价/限价
    Status status;        // 订单状态
    TimeInForce tif;      // 有效期策略
}
```

**`escrowAsset` 与 `amount` 说明**：

`submitOrder` 时，合约通过 `transferFrom` 将用户钱包中 `amount` 数量的 `escrowAsset` 代币转入合约自身保管（即"托管"）。订单成交或取消后，由后端决定如何分配这些托管代币（mint 对应资产给用户、退还多余部分等）。

| 字段 | Buy（买入） | Sell（卖出） |
|------|------------|-------------|
| `escrowAsset` | USDM 地址 | 对应股票 PocToken 地址（如 ancAAPL） |
| `amount` | `price * qty / 1e18`（购买所需的 USDM 总额） | `qty`（卖出的股票数量） |

**订单号编码规则**：`AAAAAABBSSSSSSSSSS`
- `AAAAAA`（6位）：用户地址的 keccak256 哈希取模 1,000,000
- `BB`（2位）：订单类型编码（01=Market, 02=Limit）
- `SSSSSSSSSS`（10位）：用户维度的自增序号

#### 2.1.4 函数列表

| 函数 | 访问控制 | 说明 |
|------|----------|------|
| `initialize(usdm_, admin_, backend_)` | `initializer` | 初始化合约，设置 USDM 地址和角色 |
| `setSymbolToken(symbol, token)` | `DEFAULT_ADMIN_ROLE` | 注册/更新交易标的对应的 PocToken |
| `setBackend(backend)` | `DEFAULT_ADMIN_ROLE` | 添加后端地址的 BACKEND_ROLE |
| `submitOrder(symbol, qty, price, side, orderType, tif)` | 任意用户 | 提交订单并托管资金。Buy 托管 USDM（price*qty），Sell 托管股票代币（qty） |
| `cancelOrderIntent(orderId)` | 订单所有者 | 用户发起取消意图（仅 Pending 可发起） |
| `markExecuted(orderId, refundAmount)` | `BACKEND_ROLE` | 后端标记订单已执行，可退还多余托管金额 |
| `cancelOrder(orderId)` | `BACKEND_ROLE` | 后端最终取消订单，退还全部托管资金 |
| `getOrder(orderId)` | 任意（view） | 查询订单详情 |
| `getOrderNumber(orderId)` | 任意（view） | 查询结构化订单号 |

#### 2.1.5 资金流转逻辑

**买入订单 (Buy)**：
1. 用户先 `approve` OrderContract 花费 USDM
2. `submitOrder()` 时，合约通过 `transferFrom` 将 `price * qty / 1e18` 的 USDM 从用户转入合约
3. Alpaca 成交后，后端调用 `markExecuted(orderId, refundAmount)`，其中 `refundAmount = escrowAmount - (filledQty * filledPrice)`，退还多余 USDM 给用户
4. 后端调用 `PocToken.mint(user, filledQty)` 铸造对应数量的股票代币给用户（通过 `symbolToToken(symbol)` 查询代币地址）
5. `cancelOrder()` 时，退还全部托管的 USDM（仅在未成交时调用）

**卖出订单 (Sell)**：
1. 用户先 `approve` OrderContract 花费对应股票的 PocToken
2. `submitOrder()` 时，合约通过 `transferFrom` 将 `qty` 的股票代币从用户转入合约
3. Alpaca 成交后，后端调用 `markExecuted(orderId, 0)`，托管的股票代币留在合约中（refundAmount = 0）
4. 后端调用 `PocToken.mint(user, filledQty * filledPrice)` 铸造等值 USDM 给用户（通过 `OrderCaller.USDM()` 查询 USDM 地址）
5. `cancelOrder()` 时，退还全部托管的股票代币（仅在未成交时调用）

**部分成交后取消/过期**：
- 如果订单在取消/过期前已有部分成交，不能调用 `cancelOrder`（会退还全部 escrow，导致超额退款）
- 必须调用 `markExecuted(orderId, refundAmount)`，通过 refundAmount 退还未成交部分的资金
- 同时 mint 已成交部分对应的代币给用户
- `markExecuted` 和 `cancelOrder` 在合约层面互斥，同一订单只能调用其中之一

---

### 2.2 PocToken（PocToken.sol）

PocToken 是基于 ERC20 标准的代币合约，增加了基于角色的 mint/burn 机制。系统中部署多个 PocToken 实例：一个作为 USDM（美元稳定币映射），其余作为各股票标的的代币。

#### 2.2.1 角色

| 角色 | 常量 | 说明 |
|------|------|------|
| 管理员 | `DEFAULT_ADMIN_ROLE` | 可管理所有角色，设置 operator |
| 铸币者 | `MINTER_ROLE` | 可调用 `mint()` 铸造代币 |
| 销毁者 | `BURNER_ROLE` | 可调用 `burn()` / `burnFrom()` 销毁代币 |

#### 2.2.2 函数列表

| 函数 | 访问控制 | 说明 |
|------|----------|------|
| `initialize(name_, symbol_)` | `initializer` | 初始化代币名称和符号，授予调用者 ADMIN/MINTER/BURNER 角色 |
| `setOperator(operator)` | `DEFAULT_ADMIN_ROLE` | 为指定地址授予 MINTER 和 BURNER 角色 |
| `mint(to, amount)` | `MINTER_ROLE` | 铸造代币到指定地址 |
| `burn(amount)` | `BURNER_ROLE` | 销毁调用者持有的代币 |
| `burnFrom(from, amount)` | `BURNER_ROLE` | 销毁指定地址持有的代币（无需 allowance） |
| `isMinter(account)` | 任意（view） | 查询地址是否拥有 MINTER 角色 |
| `isBurner(account)` | 任意（view） | 查询地址是否拥有 BURNER 角色 |
| `name()` | 任意（view） | 返回代币名称（可在初始化时自定义） |
| `symbol()` | 任意（view） | 返回代币符号（可在初始化时自定义） |

#### 2.2.3 关键特性

- **Beacon Proxy 模式**：构造函数接收 `gateContract_` 地址，通过 `_disableInitializers()` 禁止直接初始化实现合约
- **RBAC 控制**：mint 和 burn 操作均受角色约束，防止未授权操作
- **灵活命名**：通过覆盖 `name()` 和 `symbol()` 函数，支持在 proxy 部署后自定义名称

#### 2.2.4 生产级架构：Gate、PocToken、Beacon Proxy 关系

GATE_CONTRACT 和 PocToken 的关系是工厂模式下的创建者标识：

设计意图：在生产级架构中，Gate 合约是 PocToken 的工厂——由 Gate 动态创建和管理所有 PocToken 实例。GATE_CONTRACT 记录的是"谁创建了我"，用于：

1. 溯源：任何 PocToken 实例都可以通过 GATE_CONTRACT 找到创建它的 Gate 合约
2. 权限预留：生产级版本可能允许 GATE_CONTRACT 绕过 RBAC 直接执行 mint/burn（因为 Gate 是代币的合法管理者）
3. Beacon Proxy 下的实例区分：所有 PocToken 共享同一个实现合约，GATE_CONTRACT 作为 immutable 在每个代理的创建时写入 bytecode，是区分不同实例身份的唯一不可变标记

当前 POC 阶段的简化：部署脚本中直接用 deployer 地址作为占位传入（而不是真正的 Gate 地址），因为 POC 阶段由部署脚本手动创建所有 PocToken 实例，尚未实现 Gate 工厂动态创建的功能。

| 阶段 | GATE_CONTRACT 值 | 谁创建 PocToken |
|------|------------------|-----------------|
| 当前 POC | deployer（占位） | 部署脚本手动创建 |
| 生产级目标 | Gate 合约地址 | Gate 合约动态创建并管理 |

**合约关系图**：

```mermaid
graph TB
    subgraph deploy["部署阶段"]
        Admin["管理员"]
        PocTokenImpl["PocToken 实现合约"]
        Beacon["UpgradeableBeacon"]
    end

    subgraph gate_factory["Gate工厂"]
        Gate["Gate: 充值提现网关 + 代币工厂"]
    end

    subgraph tokens["PocToken实例群"]
        USDM["USDM BeaconProxy"]
        AncAAPL["ancAAPL BeaconProxy"]
        AncTSLA["ancTSLA BeaconProxy"]
        AncNVDA["ancNVDA BeaconProxy"]
        NewToken["ancXXX BeaconProxy - 动态创建"]
    end

    USDC["USDC 外部代币"]
    User["用户"]

    Admin -->|"① 部署"| PocTokenImpl
    Admin -->|"② 创建 Beacon"| Beacon
    PocTokenImpl -.->|"Beacon 指向"| Beacon
    Admin -->|"③ 部署"| Gate

    Gate -->|"④ createToken 部署 BeaconProxy"| USDM
    Gate -->|"④ createToken"| AncAAPL
    Gate -->|"④ createToken"| AncTSLA
    Gate -->|"④ createToken"| AncNVDA
    Gate -->|"动态创建新资产"| NewToken

    USDM -.->|"代理到实现合约"| PocTokenImpl
    AncAAPL -.->|"代理到实现合约"| PocTokenImpl
    AncTSLA -.->|"代理到实现合约"| PocTokenImpl
    AncNVDA -.->|"代理到实现合约"| PocTokenImpl
    NewToken -.->|"代理到实现合约"| PocTokenImpl
    Beacon -.->|"所有代理共享"| PocTokenImpl

    Gate -->|"MINTER/BURNER"| USDM
    Gate -->|"MINTER/BURNER"| AncAAPL
    Gate -->|"MINTER/BURNER"| AncNVDA
```

**数据流：新资产上架**：

```mermaid
sequenceDiagram
    participant Admin as 管理员
    participant Gate as Gate 合约
    participant Beacon as UpgradeableBeacon
    participant Proxy as 新 BeaconProxy
    participant PocToken as PocToken 实现

    Admin->>Gate: createToken("NVDA.anc", "NVDA.anc")
    Gate->>Gate: 检查调用者权限 (CONFIGURE_ROLE)
    Gate->>Proxy: new BeaconProxy(beacon, initData)
    Note over Proxy: initData = initialize("NVDA.anc", "NVDA.anc")
    Proxy->>PocToken: delegatecall initialize()
    Note over PocToken: 授予 Gate: ADMIN + MINTER + BURNER
    Note over Proxy: GATE_CONTRACT = Gate (immutable)
    Gate-->>Admin: 返回新代币地址
    Gate->>Gate: 记录 symbolToToken("NVDA") = Proxy地址
```

**数据流：用户充值（含 pending）**：

```mermaid
sequenceDiagram
    participant User as 用户
    participant Gate as Gate 合约
    participant USDC as USDC
    participant PendingAncUSDC as PendingAncUSDC
    participant AncUSDC as ancUSDC
    participant Backend as 后端

    Note over User,Backend: 阶段① 用户发起充值
    User->>USDC: approve(Gate, usdcAmount)
    User->>Gate: deposit(usdcAmount)
    Gate->>Gate: 校验金额 >= minimumDepositAmount
    Gate->>USDC: safeTransferFrom(user, Gate, usdcAmount)
    Gate->>PendingAncUSDC: mint(user, pendingAncUSDCAmount)
    Note over Gate: 记录 DepositOperation (PENDING)
    Gate-->>User: emit PendingDeposit(operationId, user, usdcAmount, pendingAncUSDCAmount)

    Note over User,Backend: 阶段② 后端确认充值
    Backend->>Gate: processDeposit(operationId, ancUSDCAmount)
    Gate->>PendingAncUSDC: burnFrom(user, pendingAncUSDCAmount)
    Gate->>AncUSDC: mint(user, ancUSDCAmount)
    Note over Gate: 状态 PENDING → ACTIVE
    Gate-->>User: emit DepositProcessed(operationId, user, ancUSDCAmount)
```

**数据流：用户提现（含 pending）**：

```mermaid
sequenceDiagram
    participant User as 用户
    participant Gate as Gate 合约
    participant AncUSDC as ancUSDC
    participant PendingUSDC as PendingUSDC
    participant USDC as USDC
    participant Backend as 后端

    Note over User,Backend: 阶段① 用户发起提现
    User->>Gate: withdraw(ancUSDCAmount)
    Gate->>Gate: 校验金额 >= minimumWithdrawalAmount
    Gate->>AncUSDC: burnFrom(user, ancUSDCAmount)
    Note over AncUSDC: Gate 持有 BURNER_ROLE, 无需 approve
    Gate->>PendingUSDC: mint(user, pendingUSDCAmount)
    Note over Gate: 记录 WithdrawalOperation (PENDING)
    Gate-->>User: emit PendingWithdraw(operationId, user, ancUSDCAmount, pendingUSDCAmount)

    Note over User,Backend: 阶段② 后端确认提现
    Backend->>Gate: processWithdrawal(operationId, usdcAmount)
    Gate->>PendingUSDC: burnFrom(user, pendingUSDCAmount)
    Gate->>USDC: safeTransfer(user, usdcAmount)
    Note over Gate: 状态 PENDING → REDEEMED
    Gate-->>User: emit WithdrawalProcessed(operationId, user, usdcAmount)
```

---

### 2.3 PocGate（PocGate.sol）

PocGate 是 POC 版本的充值/提现网关，提供 USDC 和 USDM 之间的即时兑换（无 pending 状态）。

#### 2.3.1 角色

| 角色 | 常量 | 说明 |
|------|------|------|
| 管理员 | `DEFAULT_ADMIN_ROLE` | 最高权限 |
| 配置员 | `CONFIGURE_ROLE` | 可设置最低充值/提现金额 |
| 暂停员 | `PAUSE_ROLE` | 可暂停/恢复充值和提现功能 |

#### 2.3.2 充值流程

```mermaid
sequenceDiagram
    participant User as 用户
    participant PocGate as PocGate
    participant USDC as MockUSDC
    participant USDM as PocToken(USDM)

    User->>USDC: approve(PocGate, amount)
    User->>PocGate: deposit(usdcAmount)
    PocGate->>PocGate: 检查金额 >= minimumDepositAmount
    PocGate->>USDC: safeTransferFrom(user, PocGate, usdcAmount)
    PocGate->>PocGate: _convertDecimals(usdcAmount, 6 -> 18)
    PocGate->>USDM: mint(user, usdmAmount)
    PocGate-->>User: emit Deposited(user, usdcAmount, usdmAmount)
```

#### 2.3.3 提现流程

```mermaid
sequenceDiagram
    participant User as 用户
    participant PocGate as PocGate
    participant USDM as PocToken(USDM)
    participant USDC as MockUSDC

    User->>USDM: approve(PocGate, amount)
    User->>PocGate: withdraw(usdmAmount)
    PocGate->>PocGate: 检查金额 >= minimumWithdrawalAmount
    PocGate->>USDM: safeTransferFrom(user, PocGate, usdmAmount)
    PocGate->>USDM: burn(usdmAmount)
    PocGate->>PocGate: _convertDecimals(usdmAmount, 18 -> 6)
    PocGate->>USDC: safeTransfer(user, usdcAmount)
    PocGate-->>User: emit Withdrawn(user, usdmAmount, usdcAmount)
```

#### 2.3.4 函数列表

| 函数 | 访问控制 | 说明 |
|------|----------|------|
| `initialize(guardian_, minDeposit, minWithdrawal)` | `initializer` | 初始化合约参数和角色 |
| `deposit(usdcAmount)` | 任意用户 | 充值 USDC，即时获得 USDM |
| `withdraw(usdmAmount)` | 任意用户 | 提现 USDM，即时获得 USDC |
| `setMinimumDepositAmount(amount)` | `CONFIGURE_ROLE` | 设置最低充值金额 |
| `setMinimumWithdrawalAmount(amount)` | `CONFIGURE_ROLE` | 设置最低提现金额 |
| `pauseDeposits()` | `PAUSE_ROLE` | 暂停充值 |
| `unpauseDeposits()` | `PAUSE_ROLE` | 恢复充值 |
| `pauseWithdrawals()` | `PAUSE_ROLE` | 暂停提现 |
| `unpauseWithdrawals()` | `PAUSE_ROLE` | 恢复提现 |

#### 2.3.5 精度转换

PocGate 内部通过 `_convertDecimals()` 处理 USDC（6 decimals）和 USDM（18 decimals）之间的精度转换：
- 充值：`usdcAmount * 10^12` 得到 usdmAmount
- 提现：`usdmAmount / 10^12` 得到 usdcAmount

---

### 2.4 MockUSDC（MockUSDC.sol）

MockUSDC 是测试环境使用的模拟 USDC 代币，行为简化。

| 特性 | 值 |
|------|-----|
| 名称 | Mock USDC |
| 符号 | USDC |
| 精度 | 6 decimals |
| mint | 任意地址均可调用（无权限控制，仅用于测试） |
| burn | 任意地址均可调用（无权限控制，仅用于测试） |

---

## 3. 生产级 Gate 合约详解（Gate.sol）

Gate 合约是生产环境使用的充值/提现网关，与 PocGate 的核心区别是引入了 **pending 状态机制**，充值/提现操作需要后端确认后才最终完成。

### 3.1 操作状态机

```mermaid
stateDiagram-v2
    [*] --> PENDING: deposit() / withdraw()
    PENDING --> ACTIVE: processDeposit() [PROCESSOR_ROLE]
    PENDING --> REDEEMED: processWithdrawal() [PROCESSOR_ROLE]
    ACTIVE --> [*]
    REDEEMED --> [*]
```

| 状态 | 值 | 说明 |
|------|-----|------|
| `PENDING` | 0 | 操作已创建，等待后端处理 |
| `ACTIVE` | 1 | 充值已处理完成（用户获得 ancUSDC） |
| `REDEEMED` | 2 | 提现已处理完成（用户获得 USDC） |

### 3.2 角色

| 角色 | 常量 | 说明 |
|------|------|------|
| 管理员 | `DEFAULT_ADMIN_ROLE` | 最高权限 |
| 配置员 | `CONFIGURE_ROLE` | 设置最低金额参数 |
| 暂停员 | `PAUSE_ROLE` | 暂停/恢复操作 |
| 处理器 | `PROCESSOR_ROLE` | 后端处理 pending 操作 |

### 3.3 充值流程（含 pending）

1. 用户调用 `deposit(usdcAmount)`
2. 合约将 USDC 从用户转入合约
3. 合约铸造 **PendingAncUSDC**（待确认代币）给用户
4. 记录 `DepositOperation`，状态为 `PENDING`
5. 后端调用 `processDeposit(operationId, ancUSDCAmount)`
6. 合约销毁用户的 PendingAncUSDC，铸造正式 **ancUSDC** 给用户
7. 状态变为 `ACTIVE`

### 3.4 提现流程（含 pending）

1. 用户调用 `withdraw(ancUSDCAmount)`
2. 合约销毁用户的 ancUSDC
3. 合约铸造 **PendingUSDC**（待确认代币）给用户
4. 记录 `WithdrawalOperation`，状态为 `PENDING`
5. 后端调用 `processWithdrawal(operationId, usdcAmount)`
6. 合约销毁用户的 PendingUSDC，将 USDC 转给用户
7. 状态变为 `REDEEMED`

### 3.5 数据结构

```solidity
struct DepositOperation {
    address user;                   // 充值用户
    uint256 usdcAmount;             // USDC 充值金额
    uint256 pendingAncUSDCAmount;   // 待确认 ancUSDC 数量
    OperationStatus status;         // 操作状态
    uint256 timestamp;              // 创建时间戳
}

struct WithdrawalOperation {
    address user;                   // 提现用户
    uint256 ancUSDCAmount;          // ancUSDC 提现金额
    uint256 pendingUSDCAmount;      // 待确认 USDC 数量
    OperationStatus status;         // 操作状态
    uint256 timestamp;              // 创建时间戳
}
```

### 3.6 函数列表

| 函数 | 访问控制 | 说明 |
|------|----------|------|
| `initialize(usdc_, ancUSDC_, guardian_, minDeposit, minWithdrawal)` | `initializer` | 初始化并创建 PendingAncUSDC / PendingUSDC 合约 |
| `deposit(usdcAmount)` | 任意用户 | 充值 USDC，获得 PendingAncUSDC |
| `withdraw(ancUSDCAmount)` | 任意用户 | 提现 ancUSDC，获得 PendingUSDC |
| `processDeposit(operationId, ancUSDCAmount)` | `PROCESSOR_ROLE` | 处理充值：销毁 pending 代币，铸造正式 ancUSDC |
| `processWithdrawal(operationId, usdcAmount)` | `PROCESSOR_ROLE` | 处理提现：销毁 pending 代币，转 USDC |
| `setMinimumDepositAmount(amount)` | `CONFIGURE_ROLE` | 设置最低充值金额 |
| `setMinimumWithdrawalAmount(amount)` | `CONFIGURE_ROLE` | 设置最低提现金额 |
| `pauseDeposits()` / `unpauseDeposits()` | `PAUSE_ROLE` | 暂停/恢复充值 |
| `pauseWithdrawals()` / `unpauseWithdrawals()` | `PAUSE_ROLE` | 暂停/恢复提现 |
| `getDepositOperation(operationId)` | 任意（view） | 查询充值操作详情 |
| `getWithdrawalOperation(operationId)` | 任意（view） | 查询提现操作详情 |

### 3.7 operationId 生成

```solidity
operationId = keccak256(abi.encodePacked(block.timestamp, block.number, msg.sender, ++_operationCounter))
```

由区块时间戳、区块号、调用者地址和自增计数器共同生成，确保唯一性。

---

## 4. 权限模型

所有合约均基于 OpenZeppelin 的 `AccessControlEnumerable` 实现 RBAC（基于角色的访问控制）。

### 4.1 角色汇总

| 合约 | 角色 | bytes32 值 | 说明 |
|------|------|------------|------|
| OrderContract | `DEFAULT_ADMIN_ROLE` | `0x00` | 管理员，可设置 symbol token 和 backend |
| OrderContract | `BACKEND_ROLE` | `keccak256("BACKEND_ROLE")` | 后端服务，可执行/取消订单 |
| PocToken | `DEFAULT_ADMIN_ROLE` | `0x00` | 管理员，可设置 operator |
| PocToken | `MINTER_ROLE` | `keccak256("MINTER_ROLE")` | 铸币角色 |
| PocToken | `BURNER_ROLE` | `keccak256("BURNER_ROLE")` | 销毁角色 |
| PocGate | `DEFAULT_ADMIN_ROLE` | `0x00` | 管理员 |
| PocGate | `CONFIGURE_ROLE` | `keccak256("CONFIGURE_ROLE")` | 配置参数 |
| PocGate | `PAUSE_ROLE` | `keccak256("PAUSE_ROLE")` | 暂停/恢复操作 |
| Gate | `DEFAULT_ADMIN_ROLE` | `0x00` | 管理员 |
| Gate | `CONFIGURE_ROLE` | `keccak256("CONFIGURE_ROLE")` | 配置参数 |
| Gate | `PAUSE_ROLE` | `keccak256("PAUSE_ROLE")` | 暂停/恢复操作 |
| Gate | `PROCESSOR_ROLE` | `keccak256("PROCESSOR_ROLE")` | 处理 pending 操作 |

### 4.2 角色管理

- `DEFAULT_ADMIN_ROLE` 是所有角色的默认管理角色，拥有者可以 grant/revoke 任何角色
- `grantRole(role, account)` / `revokeRole(role, account)` 由管理角色的持有者调用
- `renounceRole(role, account)` 允许用户放弃自己的角色

---

## 5. 事件定义

### 5.1 OrderContract 事件

```solidity
// 订单提交
event OrderSubmitted(
    address indexed user,      // 下单用户
    uint indexed orderId,      // 订单 ID
    string symbol,             // 交易标的
    uint qty,                  // 数量
    uint price,                // 价格
    Side side,                 // 买/卖
    OrderType orderType,       // 市价/限价
    TimeInForce tif,           // 有效期策略
    uint blockTimestamp        // 区块时间戳
);

// 用户请求取消
event CancelRequested(
    address indexed user,      // 用户
    uint indexed orderId,      // 订单 ID
    uint blockTimestamp        // 区块时间戳
);

// 订单已执行
event OrderExecuted(
    uint indexed orderId,      // 订单 ID
    address indexed user,      // 用户
    uint refundAmount,         // 退款金额
    TimeInForce tif            // 有效期策略
);

// 订单已取消
event OrderCancelled(
    uint indexed orderId,      // 订单 ID
    address indexed user,      // 用户
    address asset,             // 退款资产地址
    uint refundAmount,         // 退款金额
    Side side,                 // 买/卖
    OrderType orderType,       // 市价/限价
    TimeInForce tif,           // 有效期策略
    Status previousStatus      // 取消前的状态
);
```

### 5.2 PocToken 事件

```solidity
event TokensMinted(address indexed to, uint256 amount);    // 代币铸造
event TokensBurned(address indexed from, uint256 amount);  // 代币销毁
```

### 5.3 PocGate 事件

```solidity
event Deposited(address indexed user, uint256 usdcAmount, uint256 usdmAmount);  // 充值完成
event Withdrawn(address indexed user, uint256 usdmAmount, uint256 usdcAmount);  // 提现完成
event MinimumDepositAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);
event MinimumWithdrawalAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);
event DepositsPaused();
event DepositsUnpaused();
event WithdrawalsPaused();
event WithdrawalsUnpaused();
```

### 5.4 Gate 事件

```solidity
event PendingDeposit(bytes32 indexed operationId, address indexed user, uint256 usdcAmount, uint256 pendingAncUSDCAmount);
event DepositProcessed(bytes32 indexed operationId, address indexed user, uint256 ancUSDCAmount);
event PendingWithdraw(bytes32 indexed operationId, address indexed user, uint256 ancUSDCAmount, uint256 pendingUSDCAmount);
event WithdrawalProcessed(bytes32 indexed operationId, address indexed user, uint256 usdcAmount);
event MinimumDepositAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);
event MinimumWithdrawalAmountSet(uint256 indexed oldAmount, uint256 indexed newAmount);
event DepositsPaused();
event DepositsUnpaused();
event WithdrawalsPaused();
event WithdrawalsUnpaused();
```

---

## 6. 合约部署顺序

### 6.1 POC 合约部署

```mermaid
graph LR
    Step1["1. 部署 MockUSDC"] --> Step2["2. 部署 PocToken(USDM)"]
    Step2 --> Step3["3. 部署 PocGate(USDC, USDM)"]
    Step3 --> Step4["4. 初始化 PocGate"]
    Step4 --> Step5["5. 为 PocGate 授予 USDM 的 MINTER/BURNER 角色"]
    Step5 --> Step6["6. 部署 OrderContract"]
    Step6 --> Step7["7. 初始化 OrderContract(USDM, admin, backend)"]
    Step7 --> Step8["8. 部署各股票 PocToken"]
    Step8 --> Step9["9. OrderContract.setSymbolToken()"]
    Step9 --> Step10["10. 为后端授予各 PocToken 的 MINTER 角色"]
```

**详细步骤**：

1. **部署 MockUSDC** -- 无依赖
2. **部署 PocToken 实现合约** -- 传入 Gate 地址（此时可用零地址占位）
3. **通过 Beacon Proxy 部署 USDM 实例** -- 调用 `initialize("USDM", "USDM")`
4. **部署 PocGate** -- 构造函数传入 `USDC` 和 `USDM` 地址
5. **初始化 PocGate** -- 调用 `initialize(guardian, minDeposit, minWithdrawal)`
6. **授权 PocGate** -- 在 USDM 上调用 `setOperator(PocGateAddress)`，授予 MINTER + BURNER
7. **部署 OrderContract** -- 通过 Proxy 部署
8. **初始化 OrderContract** -- 调用 `initialize(USDM, admin, backend)`
9. **部署各股票 PocToken** -- 如 AAPL Token，调用 `initialize("Anchored AAPL", "ancAAPL")`
10. **注册股票代币** -- 在 OrderContract 上调用 `setSymbolToken("AAPL", ancAAPLAddress)`
11. **授权后端** -- 在各股票 PocToken 上调用 `setOperator(backendAddress)`

### 6.2 生产级 Gate 部署

1. **部署 USDC**（使用真实 USDC 合约地址）
2. **部署 ancUSDC 代币**
3. **部署 Gate 合约** -- 构造函数传入 USDC 和 ancUSDC 地址
4. **初始化 Gate** -- 调用 `initialize(usdc, ancUSDC, guardian, minDeposit, minWithdrawal)`
   - 初始化过程中自动创建 PendingAncUSDC 和 PendingUSDC 合约
5. **授权 Gate** -- 在 ancUSDC 上授予 Gate 合约 MINTER + BURNER 角色

---

## 7. 安全机制

### 7.1 ReentrancyGuard

`OrderContract`、`PocGate` 和 `Gate` 均继承了 OpenZeppelin 的 `ReentrancyGuard`，通过 `nonReentrant` 修饰符防止重入攻击。涉及代币转账的关键函数（`submitOrder`、`markExecuted`、`cancelOrder`、`deposit`、`withdraw`、`processDeposit`、`processWithdrawal`）均使用了该修饰符。

### 7.2 AccessControl

所有管理操作和后端操作均受 `onlyRole()` 修饰符保护：
- 用户不能调用后端专属函数（如 `markExecuted`、`cancelOrder`、`processDeposit`）
- 后端不能绕过角色检查
- 角色可通过 `grantRole` / `revokeRole` 动态管理

### 7.3 Initializable（可升级代理模式）

所有核心合约均继承 `Initializable`，支持通过代理模式（Proxy）部署：
- 构造函数中调用 `_disableInitializers()` 防止实现合约被直接初始化
- `initialize()` 函数使用 `initializer` 修饰符确保只能被调用一次
- 支持后续通过升级代理合约来更新逻辑

### 7.4 SafeERC20

`PocGate` 和 `Gate` 使用 OpenZeppelin 的 `SafeERC20` 库进行代币转账，通过 `safeTransferFrom` 和 `safeTransfer` 确保转账失败时 revert，避免静默失败。

### 7.5 输入校验

- 零地址检查：所有涉及地址参数的函数均检查 `address(0)`
- 零金额检查：所有涉及金额参数的函数均检查 `amount == 0`
- 状态检查：订单操作前验证当前状态是否允许转换
- 所有权检查：`cancelOrderIntent()` 验证 `msg.sender == order.user`

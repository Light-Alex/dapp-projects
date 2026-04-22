# Anchored Finance -- BSC Testnet 合约部署文档

## 1. 概述

本文档描述如何将 Anchored Finance PoC 合约部署到 **BSC Testnet**（BNB Smart Chain 测试网）。

| 项目 | 值 |
|---|---|
| 网络名称 | BSC Testnet |
| Chain ID | 97 |
| RPC URL | https://data-seed-prebsc-1-s1.binance.org:8545 |
| 区块浏览器 | https://testnet.bscscan.com |
| 水龙头 | https://www.bnbchain.org/en/testnet-faucet |
| 原生代币 | tBNB |

### 合约清单

部署涉及以下合约（按部署顺序）：

| 序号 | 合约 | 说明 | 代理模式 |
|---|---|---|---|
| 1 | MockUSDC | 测试用 USDC，6 位精度 | 无代理 |
| 2 | PocToken (USDM) | 稳定币代币，18 位精度 | BeaconProxy |
| 3 | PocToken (AAPL.anc) | 股票代币 -- 苹果 | BeaconProxy |
| 4 | PocToken (TSLA.anc) | 股票代币 -- 特斯拉 | BeaconProxy |
| 5 | OrderContract | 订单合约，处理下单/执行/取消 | TransparentUpgradeableProxy |
| 6 | PocGate | 入金/出金网关，USDC 与 USDM 兑换 | TransparentUpgradeableProxy |

---

## 2. 环境准备

### 2.1 安装 Foundry

```bash
# 安装 foundryup
curl -L https://foundry.paradigm.xyz | bash

# 安装最新版本的 forge, cast, anvil
foundryup
```

验证安装：

```bash
forge --version
cast --version
```

### 2.2 安装项目依赖

```bash
cd rwa-contract
npm install
```

### 2.3 创建部署钱包

使用 `cast` 创建一个新钱包：

```bash
cast wallet new
```

输出示例：

```
Successfully created new keypair.
Address:     0xYourNewAddress...
Private key: 0xYourPrivateKey...
```

> 妥善保存私钥，不要提交到代码仓库。

### 2.4 获取测试 BNB

1. 访问水龙头：https://www.bnbchain.org/en/testnet-faucet
2. 输入上一步生成的钱包地址
3. 领取测试 tBNB（部署大约需要 0.1 tBNB）

### 2.5 配置环境变量

在 `rwa-contract` 目录下创建 `.env` 文件：

```bash
# 部署者私钥（不要提交到 git）
PRIVATE_KEY=0xYourPrivateKey

# 后端服务钱包地址（用于 OrderContract 的 BACKEND_ROLE）
BACKEND_ADDRESS=0xBackendWalletAddress

# 代理管理员地址（用于 TransparentUpgradeableProxy 和 UpgradeableBeacon 的管理）
PROXY_ADMIN_ADDRESS=0xProxyAdminAddress

# BSCScan API Key（用于合约验证）
BSCSCAN_API_KEY=YourBscScanApiKey
```

> 获取 BSCScan API Key：注册 https://testnet.bscscan.com，在 API Keys 页面创建。

### 2.6 配置 foundry.toml

在现有 `foundry.toml` 中追加 BSC Testnet profile：

```toml
[rpc_endpoints]
bsc_testnet = "https://data-seed-prebsc-1-s1.binance.org:8545"

[etherscan]
bsc_testnet = { key = "${BSCSCAN_API_KEY}", url = "https://api-testnet.bscscan.com/api" }
```

---

## 3. 合约编译

```bash
cd rwa-contract
forge build
```

成功时输出类似：

```
[⠊] Compiling...
[⠒] Compiling X files with solc 0.8.x
[⠑] Solc 0.8.x finished in Xs
Compiler run successful!
```

如编译失败，确认 `node_modules` 已安装（`npm install`），以及 `foundry.toml` 中的 `remappings` 正确指向依赖路径。

---

## 4. 部署脚本

部署脚本位于：

```
rwa-contract/script/poc/DeployAll.s.sol
```

### 部署流程详解

#### Step 1: 部署 MockUSDC

```solidity
MockUSDC mockUSDC = new MockUSDC();
mockUSDC.mint(deployer, 1_000_000 * 1e6); // mint 100万 USDC（6位精度）
```

- MockUSDC 是一个简单的 ERC20，decimals = 6
- 无访问控制，任何人可 mint（仅测试用）

#### Step 2: 部署 USDM（PocToken BeaconProxy）

```solidity
// 2a. 部署 PocToken 实现合约
PocToken pocTokenImpl = new PocToken(deployer);

// 2b. 部署 UpgradeableBeacon，指向实现合约
UpgradeableBeacon beacon = new UpgradeableBeacon(pocTokenImpl, proxyAdmin);

// 2c. 部署 BeaconProxy，调用 initialize("USDM", "USDM")
bytes memory initData = abi.encodeWithSelector(PocToken.initialize.selector, "USDM", "USDM");
BeaconProxy usdmProxy = new BeaconProxy(beacon, initData);
```

- PocToken 构造函数需要一个 `gateContract_` 地址参数（immutable），这里使用 deployer 作为占位
- `initialize` 函数签名：`initialize(string memory name_, string memory symbol_)`
- 初始化时，调用者（deployer）自动获得 `DEFAULT_ADMIN_ROLE`、`MINTER_ROLE`、`BURNER_ROLE`

#### Step 3: 部署股票代币

与 Step 2c 相同，复用同一个 Beacon，只是初始化参数不同：

```solidity
// AAPL
bytes memory aaplInit = abi.encodeWithSelector(PocToken.initialize.selector, "AAPL.anc", "AAPL.anc");
BeaconProxy aaplProxy = new BeaconProxy(beacon, aaplInit);

// TSLA
bytes memory tslaInit = abi.encodeWithSelector(PocToken.initialize.selector, "TSLA.anc", "TSLA.anc");
BeaconProxy tslaProxy = new BeaconProxy(beacon, tslaInit);
```

#### Step 4: 部署 OrderContract（TransparentUpgradeableProxy）

```solidity
// 4a. 部署实现合约
OrderContract orderImpl = new OrderContract();

// 4b. 部署代理，调用 initialize(usdmAddress, adminAddress, backendAddress)
bytes memory orderInitData = abi.encodeWithSelector(
    OrderContract.initialize.selector,
    usdmProxy,      // USDM 代币地址
    deployer,        // admin（DEFAULT_ADMIN_ROLE）
    backendAddress   // backend（BACKEND_ROLE）
);
TransparentUpgradeableProxy orderProxy = new TransparentUpgradeableProxy(
    orderImpl, proxyAdmin, orderInitData
);
```

- `initialize` 函数签名：`initialize(address usdm_, address admin_, address backend_)`
- `admin_` 获得 `DEFAULT_ADMIN_ROLE`
- `backend_` 获得 `BACKEND_ROLE`（可以为零地址则不授权）

#### Step 5: 部署 PocGate（TransparentUpgradeableProxy）

```solidity
// 5a. 部署实现合约（构造函数需要 USDC 和 USDM 地址作为 immutable）
PocGate pocGateImpl = new PocGate(mockUSDCAddress, usdmProxy);

// 5b. 部署代理，调用 initialize(guardian, minDeposit, minWithdraw)
bytes memory gateInitData = abi.encodeWithSelector(
    PocGate.initialize.selector,
    deployer,   // guardian（DEFAULT_ADMIN_ROLE + CONFIGURE_ROLE + PAUSE_ROLE）
    uint256(0), // minimumDepositAmount（测试环境设为 0）
    uint256(0)  // minimumWithdrawalAmount（测试环境设为 0）
);
TransparentUpgradeableProxy gateProxy = new TransparentUpgradeableProxy(
    pocGateImpl, proxyAdmin, gateInitData
);
```

- PocGate 构造函数中 `USDC` 和 `USDM` 是 `immutable` 变量
- `initialize` 函数签名：`initialize(address guardian_, uint256 minimumDepositAmount_, uint256 minimumWithdrawalAmount_)`

#### Step 6: 配置权限

```solidity
bytes32 MINTER_ROLE = keccak256("MINTER_ROLE");
bytes32 BURNER_ROLE = keccak256("BURNER_ROLE");

// 6a. OrderContract 需要 mint USDM 的权限
PocToken(usdmProxy).grantRole(MINTER_ROLE, orderProxy);

// 6b. 在 OrderContract 上注册股票代币
OrderContract(orderProxy).setSymbolToken("AAPL", aaplProxy);
OrderContract(orderProxy).setSymbolToken("TSLA", tslaProxy);

// 6c. PocGate 需要 mint/burn USDM 的权限
PocToken(usdmProxy).grantRole(MINTER_ROLE, pocGateProxy);
PocToken(usdmProxy).grantRole(BURNER_ROLE, pocGateProxy);

// 6d. OrderContract 需要 mint/burn 股票代币的权限
PocToken(aaplProxy).grantRole(MINTER_ROLE, orderProxy);
PocToken(aaplProxy).grantRole(BURNER_ROLE, orderProxy);
PocToken(tslaProxy).grantRole(MINTER_ROLE, orderProxy);
PocToken(tslaProxy).grantRole(BURNER_ROLE, orderProxy);
```

权限关系总结：

| 合约/地址 | 代币 | 角色 | 用途 |
|---|---|---|---|
| OrderContract | USDM | MINTER_ROLE | 执行买单时 mint USDM |
| OrderContract | AAPL.anc | MINTER_ROLE + BURNER_ROLE | mint/burn 股票代币 |
| OrderContract | TSLA.anc | MINTER_ROLE + BURNER_ROLE | mint/burn 股票代币 |
| PocGate | USDM | MINTER_ROLE + BURNER_ROLE | 入金 mint / 出金 burn USDM |
| Backend 钱包 | USDM | MINTER_ROLE | 卖出成交后 mint USDM 给用户 |
| Backend 钱包 | AAPL.anc | MINTER_ROLE | 买入成交后 mint 股票代币给用户 |
| Backend 钱包 | TSLA.anc | MINTER_ROLE | 买入成交后 mint 股票代币给用户 |

---

## 5. 部署执行命令

### 5.1 加载环境变量

```bash
cd rwa-contract
source .env
```

### 5.2 执行部署（不验证合约）

```bash
forge script script/poc/DeployAll.s.sol:DeployAll \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545 \
    --private-key $PRIVATE_KEY \
    --broadcast \
    -vvvv
```

### 5.3 执行部署并验证合约

```bash
forge script script/poc/DeployAll.s.sol:DeployAll \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545 \
    --private-key $PRIVATE_KEY \
    --broadcast \
    --verify \
    --etherscan-api-key $BSCSCAN_API_KEY \
    -vvvv
```

### 5.4 使用 foundry.toml 中的 RPC profile

如果已在 `foundry.toml` 中配置了 `[rpc_endpoints]`，可以使用别名：

```bash
forge script script/poc/DeployAll.s.sol:DeployAll \
    --rpc-url bsc_testnet \
    --private-key $PRIVATE_KEY \
    --broadcast \
    --verify \
    --etherscan-api-key $BSCSCAN_API_KEY \
    -vvvv
```

### 5.5 使用 keystore 账户（更安全）

```bash
# 导入私钥到 keystore
cast wallet import deployer --private-key $PRIVATE_KEY

# 使用 account 参数部署
forge script script/poc/DeployAll.s.sol:DeployAll \
    --rpc-url bsc_testnet \
    --account deployer \
    --broadcast \
    --verify \
    --etherscan-api-key $BSCSCAN_API_KEY \
    -vvvv
```

### 5.6 仅模拟部署（不广播交易）

去掉 `--broadcast` 参数即可进行 dry run：

```bash
forge script script/poc/DeployAll.s.sol:DeployAll \
    --rpc-url bsc_testnet \
    --private-key $PRIVATE_KEY \
    -vvvv
```

---

## 6. 部署后验证

### 6.1 在 BSCScan 上验证合约源码

如果部署时未带 `--verify` 参数，可以手动验证：

```bash
# 验证 MockUSDC（无构造函数参数）
forge verify-contract <MockUSDC_ADDRESS> contracts/poc/MockUSDC.sol:MockUSDC \
    --chain-id 97 \
    --etherscan-api-key $BSCSCAN_API_KEY

# 验证 PocToken 实现合约（构造函数参数：gateContract_）
forge verify-contract <PocToken_IMPL_ADDRESS> contracts/poc/PocToken.sol:PocToken \
    --chain-id 97 \
    --etherscan-api-key $BSCSCAN_API_KEY \
    --constructor-args $(cast abi-encode "constructor(address)" <DEPLOYER_ADDRESS>)

# 验证 OrderContract 实现合约（无构造函数参数）
forge verify-contract <OrderContract_IMPL_ADDRESS> contracts/poc/Order.sol:OrderContract \
    --chain-id 97 \
    --etherscan-api-key $BSCSCAN_API_KEY

# 验证 PocGate 实现合约（构造函数参数：usdc_, usdm_）
forge verify-contract <PocGate_IMPL_ADDRESS> contracts/poc/PocGate.sol:PocGate \
    --chain-id 97 \
    --etherscan-api-key $BSCSCAN_API_KEY \
    --constructor-args $(cast abi-encode "constructor(address,address)" <USDC_ADDRESS> <USDM_PROXY_ADDRESS>)
```

### 6.2 验证角色权限设置正确

使用 `cast call` 检查角色分配：

```bash
# 计算角色 hash
MINTER_ROLE=$(cast keccak "MINTER_ROLE")
BURNER_ROLE=$(cast keccak "BURNER_ROLE")

# 检查 OrderContract 是否拥有 USDM 的 MINTER_ROLE
cast call <USDM_PROXY_ADDRESS> \
    "hasRole(bytes32,address)(bool)" \
    $MINTER_ROLE <ORDER_PROXY_ADDRESS> \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545

# 检查 PocGate 是否拥有 USDM 的 MINTER_ROLE 和 BURNER_ROLE
cast call <USDM_PROXY_ADDRESS> \
    "hasRole(bytes32,address)(bool)" \
    $MINTER_ROLE <POCGATE_PROXY_ADDRESS> \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545

cast call <USDM_PROXY_ADDRESS> \
    "hasRole(bytes32,address)(bool)" \
    $BURNER_ROLE <POCGATE_PROXY_ADDRESS> \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545

# 检查 OrderContract 是否拥有 AAPL 的 MINTER_ROLE 和 BURNER_ROLE
cast call <AAPL_PROXY_ADDRESS> \
    "hasRole(bytes32,address)(bool)" \
    $MINTER_ROLE <ORDER_PROXY_ADDRESS> \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545

# 检查 OrderContract 上注册的 symbol token
cast call <ORDER_PROXY_ADDRESS> \
    "symbolToToken(string)(address)" \
    "AAPL" \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545
```

所有 `hasRole` 调用应返回 `true`。

### 6.3 测试基本功能

#### 测试 PocGate Deposit（入金）

```bash
# 1. 先 approve USDC 给 PocGate
cast send <USDC_ADDRESS> \
    "approve(address,uint256)" \
    <POCGATE_PROXY_ADDRESS> 1000000 \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545 \
    --private-key $PRIVATE_KEY

# 2. 调用 deposit（1 USDC = 1000000，6位精度）
cast send <POCGATE_PROXY_ADDRESS> \
    "deposit(uint256)" \
    1000000 \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545 \
    --private-key $PRIVATE_KEY

# 3. 检查 USDM 余额（应该是 1e18，即 1 USDM）
cast call <USDM_PROXY_ADDRESS> \
    "balanceOf(address)(uint256)" \
    <DEPLOYER_ADDRESS> \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545
```

#### 测试 PocGate Withdraw（出金）

```bash
# 1. Approve USDM 给 PocGate
cast send <USDM_PROXY_ADDRESS> \
    "approve(address,uint256)" \
    <POCGATE_PROXY_ADDRESS> 1000000000000000000 \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545 \
    --private-key $PRIVATE_KEY

# 2. 调用 withdraw（1 USDM = 1e18，18位精度）
cast send <POCGATE_PROXY_ADDRESS> \
    "withdraw(uint256)" \
    1000000000000000000 \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545 \
    --private-key $PRIVATE_KEY
```

---

## 7. 部署后配置更新

部署完成后，需要将合约地址更新到后端服务的配置文件中。

### 7.1 Indexer 服务

文件：`rwa-backend/apps/indexer/config/config.yaml`

需要更新的字段：

```yaml
chain:
  chainId: 97
  pocAddress: "<ORDER_CONTRACT_PROXY_ADDRESS>"   # OrderContract 代理地址
  usdmAddress: "<USDM_PROXY_ADDRESS>"             # USDM BeaconProxy 地址
```

### 7.2 其他服务（如需要）

以下配置文件中如有合约地址相关配置，也需同步更新：

- `rwa-backend/apps/api/config/config.yaml`
- `rwa-backend/apps/ws-server/config/config.yaml`
- `rwa-backend/apps/alpaca-stream/config/config.yaml`

### 7.3 Backend 钱包配置

确保后端服务使用的钱包地址（即部署时 `BACKEND_ADDRESS` 环境变量的值）已在 OrderContract 中被授予 `BACKEND_ROLE`。可通过以下命令确认：

```bash
BACKEND_ROLE=$(cast keccak "BACKEND_ROLE")
cast call <ORDER_PROXY_ADDRESS> \
    "hasRole(bytes32,address)(bool)" \
    $BACKEND_ROLE <BACKEND_WALLET_ADDRESS> \
    --rpc-url https://data-seed-prebsc-1-s1.binance.org:8545
```

---

## 8. 合约地址记录模板

部署完成后，请将以下信息填入并保存：

| 合约 | 类型 | 地址 | 备注 |
|---|---|---|---|
| MockUSDC | 直接部署 | `0x...` | 测试用 USDC，6 位精度 |
| PocToken Implementation | 实现合约 | `0x...` | PocToken 逻辑合约 |
| UpgradeableBeacon | Beacon | `0x...` | 所有 PocToken 代理共享 |
| USDM | BeaconProxy | `0x...` | 稳定币，18 位精度 |
| AAPL.anc | BeaconProxy | `0x...` | 苹果股票代币 |
| TSLA.anc | BeaconProxy | `0x...` | 特斯拉股票代币 |
| OrderContract Implementation | 实现合约 | `0x...` | 订单逻辑合约 |
| OrderContract Proxy | TransparentProxy | `0x...` | 订单合约入口 |
| PocGate Implementation | 实现合约 | `0x...` | 入金网关逻辑合约 |
| PocGate Proxy | TransparentProxy | `0x...` | 入金网关入口 |
| ProxyAdmin (Order) | 自动创建 | `0x...` | TransparentProxy 自动创建 |
| ProxyAdmin (PocGate) | 自动创建 | `0x...` | TransparentProxy 自动创建 |

### 关键角色地址

| 角色 | 地址 | 说明 |
|---|---|---|
| Deployer / Admin | `0x...` | DEFAULT_ADMIN_ROLE 持有者 |
| Backend | `0x...` | BACKEND_ROLE 持有者（OrderContract） |
| Proxy Admin | `0x...` | 代理升级管理员 |

### 部署信息

| 项目 | 值 |
|---|---|
| 部署日期 | YYYY-MM-DD |
| 部署网络 | BSC Testnet (Chain ID: 97) |
| 部署交易哈希 | `0x...` |
| Forge broadcast 目录 | `rwa-contract/broadcast/DeployAll.s.sol/97/` |

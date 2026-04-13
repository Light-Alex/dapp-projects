# 如何查找DEX信息
该文档介绍了如何在不同类型的链上查找DEX相关交易信息的方法，包含以下内容:
1. 如何找到对应链上的DEX factory合约地址
2. 如何找到Pool合约上需要解析的Event的名字
3. 如何计算Event的签名Hash值

## 1. 区块链浏览器
### 1.1. BSC链
- [Testnet](https://testnet.bscscan.com/)
- [Mainnet](https://bscscan.com/)

### 1.2. Ethereum链
- [Testnet](http://sepolia.etherscan.io/)
- [Mainnet](https://etherscan.io/)

### 1.3. Solana链
- [Mainnet/Testnet](https://solscan.io/)
- [Mainnet/Testnet](https://explorer.solana.com/)

### 1.4. Sui链
- [Mainnet/Testnet](https://suivision.xyz/)
- [Mainnet/Testnet](https://suiscan.xyz/)

## 2. 如何查找DEX factory合约地址
### 2.1. 通过区块链浏览器查找
> 注意: 需要验证区块链浏览器上找到factory合约地址是否真实
> 1. 合约地址上一定会有大量交易
> 2. 区块链浏览器上有DEX 标识
> 3. 合约创建者一定是DEX的官方地址
> 4. 合约最好是开源的, 合约代码中有和factory合约相关的逻辑

#### 2.1.1. Four.Meme(BSC链)
1. Four.meme V1
    - 通过区块链浏览器暂未找到`Four.meme V1`的factory合约地址

2. Four.meme V2
    1. 访问[BSC区块链浏览器](https://bscscan.com/)
    2. 搜索`four.meme`
    3. 搜索结果中`Four.meme: Token Manager`合约地址就是FourMeme的DEX factory合约地址

#### 2.1.2. PancakeSwap(BSC链)
1. PancakeSwap V2
    1. 访问[BSC区块链浏览器](https://bscscan.com/)
    2. 搜索`pancakeSwap factory V2`
    3. 点击`Addresses`, 查找地址
    4. 找到的`PancakeSwap: Factory v2`即为pancakeSwap V2的factory合约地址

2. PancakeSwap V3
    1. 访问[BSC区块链浏览器](https://bscscan.com/)
    2. 搜索`pancakeSwap factory V3`
    3. 点击`Addresses`, 查找地址
    4. 找到的`PancakeSwap V3: Factory `即为pancakeSwap V3的factory合约地址

#### 2.1.3. Uniswap(Ethereum链)
1. Uniswap V2
    1. 访问[Ethereum区块链浏览器](https://etherscan.io/)
    2. 搜索`uniswap factory V2`
    3. 点击`Addresses`, 查找地址
    4. 找到的`Uniswap V2: Factory Contract`即为uniswap V2的factory合约地址

2. Uniswap V3
    1. 访问[Ethereum区块链浏览器](https://etherscan.io/)
    2. 搜索`uniswap factory V3`
    3. 点击`Addresses`, 查找地址
    4. 找到的`Uniswap V3: Factory`即为uniswap V3的factory合约地址

#### 2.1.4. Pump.fun(Solana链)
1. 访问[Solana区块链浏览器](https://solscan.io/)
2. 搜索`pump.fun`
3. 点击`Programs`, 查找`pump.fun`程序账户的地址
4. 找到`Pump.fun`即为pumpfun的factory合约地址

#### 2.1.5. PumpSwap(Solana链)
1. 访问[Solana区块链浏览器](https://solscan.io/)
2. 搜索`pump.fun`
3. 点击`Programs`, 查找`pump.fun amm`程序账户的地址
4. 找到`Pump.fun AMM`即为PumpSwap的factory合约地址

#### 2.1.6. Bluefin(Sui链)
1. 访问[Sui区块链浏览器](https://suiscan.xyz/)
2. 搜索`Bluefin AMM 1`
3. 点击`Bluefin AMM 1`
4. `Bluefin AMM 1`界面中的Package ID即为bluefin的factory合约地址

#### 2.1.7. Cetus(Sui链)
Sui区块链浏览器暂未找到Cetus相关地址，建议从Cetus官方文档入手，参考[2.2.7 Cetus(Sui链)](#227-cetussui链)的查找方法

### 2.2. 通过项目官方文档查找
#### 2.2.1. Four.meme(BSC链)
1. Google搜索`Four meme docs`, 找到[four.meme的文档官网](https://four-meme.gitbook.io/four.meme)
2. 在文档网站左下角，点击[Protocol Integration](https://four-meme.gitbook.io/four.meme/brand/protocol-integration), 查看和合约相关的部分
3. 下载[API-Documents.03-03-2026.md](https://1270958763-files.gitbook.io/~/files/v0/b/gitbook-x-prod.appspot.com/o/spaces%2FMKYhtLfncF7vyCOOt0Ef%2Fuploads%2F62o7mCRr1omQzpSdmYMW%2FAPI-Documents.03-03-2026.md?alt=media&token=5267cf33-b7de-43fa-a852-5a37e4a5cd8c)
4. 文档中详细介绍了`TokenManager V1/V2合约地址`、`合约方法`、`事件定义`

#### 2.2.2. PancakeSwap(BSC链)
1. Google搜索`pancakeSwap docs`, 找到[pancakeSwap的开发者文档](https://developer.pancakeswap.finance/contracts/infinity/overview)
2. 在`PancakeSwap v2/V3 > Addresses`章节中能找到V2/V3 factory合约地址

#### 2.2.3. Uniswap(Ethereum链)
1. Google搜索`Uniswap docs`, 找到[Uniswap官方文档](https://docs.uniswap.org/)
2. 在上方菜单栏点击`contracts`, 进入[Uniswap合约文档](https://docs.uniswap.org/contracts/v4/overview)
3. 在`contracts > v2 Protocol > Technical Reference > V2 Deployment Addresses`章节中能找到V2 factory合约地址
4. 在`Contracts > v3 Protocol > Technical Reference > Deployments > Ethereum Deployments`章节中能找到V3 factory合约地址

#### 2.2.4. Pump.fun(Solana链)
1. Google搜索`Pump.fun docs`, 找到[Pump.fun的开发者文档](https://github.com/pump-fun/pump-public-docs)
2. 在[Other documentation > Pump Program](https://github.com/pump-fun/pump-public-docs/blob/main/docs/PUMP_PROGRAM_README.md), 找到Pump.fun的program地址

#### 2.2.5. PumpSwap(Solana链)
1. Google搜索`Pump.fun docs`, 找到[Pump.fun的开发者文档](https://github.com/pump-fun/pump-public-docs)
2. 在[Other documentation > PumpSwap](https://github.com/pump-fun/pump-public-docs/blob/main/docs/PUMP_SWAP_README.md), 找到PumpSwap的program合约地址

#### 2.2.6. Bluefin(Sui链)
1. Google搜索`bluefin docs`, 找到[bluefin的开发者文档](https://learn.bluefin.io/bluefin?utm_source=bluefin&utm_medium=internal&utm_campaign=header)
2. 在文档中暂时没有找到关于bluefin部署相关的信息

#### 2.2.7. Cetus(Sui链)
1. Google搜索`Cetus docs`, 找到[Cetus官方手册](https://cetus-1.gitbook.io/cetus-docs)
2. 在`Developer > Developer Docs`找到[Cetus开发者文档](https://cetus-1.gitbook.io/cetus-developer-docs)
3. 在`Dev Overview`章节中能与Cetus相关的所有合约地址以及Github仓库地址

### 2.3. 通过Github查找
#### 2.3.1. Four.meme(BSC链) 
暂未找到官方发布的Github仓库

#### 2.3.2. PancakeSwap(BSC链)
[pancakeSwap开发者文档](https://developer.pancakeswap.finance/contracts/infinity/overview)中有公布V2/V3的Github地址:
- `PancakeSwap v2 > Github`: [PancakeSwap v2 Github](https://github.com/pancakeswap/pancake-smart-contracts)
- `PancakeSwap v3 > Github`: [PancakeSwap v3 Github](https://github.com/pancakeswap/pancake-v3-contracts)

但是V2/V3 Github中并未公示factory合约地址

#### 2.3.3. Uniswap(Ethereum链)
[Uniswap合约文档](https://docs.uniswap.org/contracts/v4/overview)中有公布V2/V3的Github地址:
- `contracts > v2 Protocol > Overview`: 
    - [Uniswap V2 Core Github](https://github.com/uniswap/v2-core)
    - [Uniswap V2 Periphery Github](https://github.com/uniswap/v2-periphery)

- `contracts > v3 Protocol > Overview`: 
    - [Uniswap V3 Core Github](https://github.com/uniswap/v3-core)
    - [Uniswap V3 Periphery Github](https://github.com/uniswap/v3-periphery)

但是V2/V3 Github中并未公示factory合约地址

#### 2.3.4. Pump.fun(Solana链)
参考[2.2.4. Pump.fun(Solana链)](#224-pumpfunsolana链)

#### 2.3.5. PumpSwap(Solana链)
参考[2.2.5. PumpSwap(Solana链)](#225-pumpswapsolana链)

#### 2.3.6. Bluefin(Sui链)
1. Google搜索`bluefin`
2. 来到[bluefin官网](https://bluefin.io/)
3. 将网页滚动到最下面，找到[Explore > Github](https://github.com/fireflyprotocol)
4. 找到[bluefin-spot-contract-interface仓库](https://github.com/fireflyprotocol/bluefin-spot-contract-interface)
5. 在README.md中即可找到bluefin的factory合约地址

#### 2.3.7. Cetus(Sui链)
[Cetus开发者文档](https://cetus-1.gitbook.io/cetus-developer-docs)中公布了Cetus所有Github仓库地址:
- [CetusProtocol](https://github.com/CetusProtocol)

## 3. 如何找到Pool合约上需要解析的Event的名字

> Sui Move 事件的标准命名格式: `{package_id}::{module}::{event_name}`

### 3.1. 通过合约代码查找(合约已开源)
#### 3.1.1. Four.meme(BSC链)
Four.meme合约代码未开源，参考[3.2.1. Four.meme(BSC链)](#321-fourmemebsc链)

#### 3.1.2. PancakeSwap(BSC链)
1. PancakeSwap V2
    1. [PancakeFactory合约](https://bscscan.com/address/0xca143ce32fe78f1f7019d7d551a6402fc5350c73#code)已开源
    2. 在合约代码中搜索`event`, 即可找到PancakeSwap V2的事件定义


2. PancakeSwap V3
    1. [PancakeV3Factory合约](https://bscscan.com/address/0x0bfbcf9fa4f9c56b0f40a671ad40e0805a091865#code)已开源
    2. 在合约代码中搜索`event`, 即可找到PancakeSwap V3的事件定义

#### 3.1.3. Uniswap(Ethereum链)
1. Uniswap V2
    1. [UniswapV2Factory合约](https://etherscan.io/address/0x5c69bee701ef814a2b6a3edd4b1652cb9cc5aa6f#code)已开源
    2. 在合约代码中搜索`event`, 即可找到Uniswap V2的事件定义

2. Uniswap V3
    1. [UniswapV3Factory合约](https://etherscan.io/address/0x1f98431c8ad98523631ae4a59f267346ea31f984#code)已开源
    2. 在合约代码中搜索`event`, 即可找到Uniswap V3的事件定义

#### 3.1.4. Pump.fun(Solana链)
Pump.fun的事件定义可以在[Pump.fun Program](https://solscan.io/account/6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P#programIdl)的`Program IDL > Events > Raw`中查看

#### 3.1.5. PumpSwap(Solana链)

PumpSwap的事件定义可以在[PumpSwap Program](https://solscan.io/account/pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA)的`Program IDL > Events > Raw`中查看

#### 3.1.6. Bluefin(Sui链)

Bluefin的事件定义在[Bluefin AMM 1 Package](https://suiscan.xyz/mainnet/object/0x3492c874c1e3b3e2984e8c41b589e642d4d0a5d6459e5a9cfc2d52fd7c89c267/contracts)的`contracts > Events`中查看

#### 3.1.7. Cetus(Sui链)

Cetus中不同的`package_id`的不同`module`均有定义事件，涉及事件定义的`package_id`及`module`如下:

- [clmm version 1](https://suiscan.xyz/mainnet/object/0x1eabed72c53feb3805120a081dc15963c204dc8d091542592abaf7a35689b2fb/contracts)
  - Config
  - Factory
  - Partner
  - Pool
- [clmm version 13](https://suiscan.xyz/mainnet/object/0xdb5cd62a06c79695bfc9982eb08534706d3752fe123b48e0144f480209b3117f/contracts)
  - Config
  - Factory
  - Partner
  - Pool
  - Rewarder
- [dlmm version 1](https://suiscan.xyz/mainnet/object/0x5664f9d3fd82c84023870cfbda8ea84e14c8dd56ce557ad2116e0668581a682b/contracts)
  - Admin cap
  - Config
  - Partner
  - Pool
  - Registry
  - Versioned
- 等等



### 3.2. 通过开发者文档查找
#### 3.2.1. Four.meme(BSC链)
1. 进入[Four.meme开发者文档](https://1270958763-files.gitbook.io/~/files/v0/b/gitbook-x-prod.appspot.com/o/spaces%2FMKYhtLfncF7vyCOOt0Ef%2Fuploads%2F62o7mCRr1omQzpSdmYMW%2FAPI-Documents.03-03-2026.md?alt=media&token=5267cf33-b7de-43fa-a852-5a37e4a5cd8c)
2. 在`Events`章节有详细介绍TokenManager V1/V2合约的事件定义

#### 3.2.2. PancakeSwap(BSC链)
PancakeSwap合约已开源，参考[3.1.2. PancakeSwap(BSC链)](#312-pancakeswapbsc链)

#### 3.2.3. Uniswap(Ethereum链)
Uniswap合约已开源，参考[3.1.3 Uniswap(Ethereum链)](#313-uniswapethereum链)

#### 3.2.4. Pump.fun(Solana链)
参考[3.1.4. Pump.fun(Solana链)](#314-pumpfunsolana链)

#### 3.2.5. PumpSwap(Solana链)
参考[3.1.5. PumpSwap(Solana链)](#315-pumpswapsolana链)

#### 3.2.6. Bluefin(Sui链)
参考[3.1.6. Bluefin(Sui链)](#316-bluefinsui链)

#### 3.2.7. Cetus(Sui链)
参考[3.1.7. Cetus(Sui链)](#317-cetussui链)

## 4. 如何计算Event的签名
有了第3节的Event定义，我们就可以计算Event的签名了。有了Event签名，我们就可以对Event Logs中的Event进行过滤了。

### 4.1. EVM兼容链

> BSC属于EVM兼容链，Event签名的计算方式和Ethereum一致

```go
package main

import (
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

/* calculateEventSigForEVM 计算 EVM 事件的签名哈希。
 *
 * 参数：
 *   - eventName: 事件签名，格式如 "Transfer(address,address,uint256)"
 *
 * 返回值：
 *   - keccak256 哈希的十六进制字符串，如 "0x..."
 */
func calculateEventSigForEVM(eventName string) string {
	eventName = strings.Join(strings.Fields(eventName), "")
	eventHash := crypto.Keccak256Hash([]byte(eventName))
	return eventHash.Hex()
}
```



### 4.2. Solana链

```go
package main

import (
	"crypto/sha256"
	"strings"
)

/* calculateEventSigForSolana 计算 Solana Anchor 事件的 discriminator。
 *
 * 参数：
 *   - eventName: 事件名称，如 "CreateEvent"
 *
 * 返回值：
 *   - sha256("event:事件名") 的前 8 个字节
 */
func calculateEventSigForSolana(eventName string) []byte {
	eventName = strings.Join(strings.Fields(eventName), "")
	// Anchor 事件 discriminator 前缀是 "event:"
	eventNameWithPrefix := "event:" + eventName

	// 计算 sha256 哈希
	hash := sha256.Sum256([]byte(eventNameWithPrefix))

	// 取前 8 个字节作为 discriminator
	discriminator := hash[:8]

	return discriminator
}
```



### 4.3. Sui链

```go
package main

import (
	"strings"
)

/* calculateEventSigForSui 计算 Sui 事件的标识符。
 *
 * 参数：
 *   - packageAddr: 包地址
 *   - moduleName: 模块名称
 *   - eventName: 事件名称
 *
 * 返回值：
 *   - 格式为 "package地址::模块名::事件名" 的字符串
 */
func calculateEventSigForSui(packageAddr, moduleName, eventName string) string {
	packageAddr = strings.Join(strings.Fields(packageAddr), "")
	moduleName = strings.Join(strings.Fields(moduleName), "")
	eventName = strings.Join(strings.Fields(eventName), "")
	return packageAddr + "::" + moduleName + "::" + eventName
}
```


## 其他
### Cetus 合约模块总览
1. CLMM（Concentrated Liquidity Market Maker）
核心 AMM 引擎，类似 Uniswap V3 的集中流动性做市协议。LP 可在自定义价格区间内提供流动性，提高资金效率。包含：
    - factory — 创建池子
    - pool — 管理 swap、添加/移除流动性、费用收取

2. DLMM（Dynamic Liquidity Market Maker）
动态流动性做市，CLMM 的增强版本。支持更灵活的流动性分布策略，自动调整 bin（价格档位）的流动性分配。

3. Integrate（集成合约）
聚合路由与交互层，为第三方应用提供统一的调用入口。处理 swap 路由、跨池拆单、手续费分成（partner）等。外部 dApp 通常通过 Integrate 合约而非直接调用 CLMM。

4. Limit Order（限价单）
基于 CLMM 的限价单功能。利用集中流动性的单 tick 仓位实现：用户在指定价格挂单，当价格穿越该 tick 时自动成交。本质上是一个只占一个 tick 宽度的流动性仓位。

5. DCA（Dollar-Cost Averaging）
定投策略合约。允许用户设置定期定额买入某个 token，合约自动按计划分批执行 swap，降低择时风险。

6. Vaults（金库）
自动化的流动性管理。帮 LP 自动调仓（rebalance），当价格超出区间时自动将流动性移到当前价格附近，避免资金闲置。用户存入代币即可，无需手动管理仓位。

7. Farming（流动性挖矿）
流动性激励。LP 在提供流动性后可以质押 LP 仓位获取额外的代币奖励（如 CETUS）。激励分配按仓位大小和周期计算。

8. xCETUS
治理代币的质押凭证。质押 CETUS 代币获得 xCETUS，用于：
    - 治理投票权
    - 协议手续费分红资格
    - 提升挖矿收益倍率

9. Dividend（分红）
协议收入分配。将协议收取的交易手续费按 xCETUS 持有比例分配给质押者，是持有 xCETUS 的经济激励。

10. Config（配置合约）
全局参数管理。存储协议级别的配置，如白名单 tick spacing、费率档位、支持的 token 列表、协议管理员等。

**架构关系简图：**
```
Config ──全局参数──→ CLMM / DLMM (核心引擎)
                         ↑
Integrate ──路由调用──────┘
Limit Order / DCA ──基于 CLMM──→ Pool
Vaults ──自动管理 LP──→ CLMM 仓位
Farming ──激励 LP──→ CLMM 仓位
xCETUS ←──质押 CETUS
Dividend ──分红──→ xCETUS 持有者
简单来说：CLMM 是地基，Integrate 是大门，DLMM 是升级版地基，Limit Order/DCA/Vaults 是建在上面的应用层产品，Farming/xCETUS/Dividend 是经济激励机制，Config 是全局配置。
```
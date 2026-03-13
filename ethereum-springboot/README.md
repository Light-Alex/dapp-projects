# Ethereum SpringBoot 交互示例

这是一个使用Spring Boot和Web3j库与以太坊智能合约交互的示例项目。该项目可以：

1. 调用智能合约方法
2. 监听合约事件

## 技术栈

- Java 11
- Spring Boot 2.7.14
- Web3j 4.9.8
- Maven

## 项目结构

```
src/main/java/com/example/ethereum/
├── config/                   # 配置类
│   └── EthereumConfig.java   # 以太坊连接配置
├── contract/                 # 合约相关类
│   └── SimpleStorage.java    # 示例合约封装
├── controller/               # REST API控制器
│   └── ContractController.java
├── service/                  # 服务类
│   ├── ContractService.java  # 合约调用服务
│   └── ContractEventListener.java # 合约事件监听服务
└── EthereumSpringbootApplication.java # 主应用类

src/main/resources/
├── application.properties    # 应用配置文件
└── contracts/                # 智能合约源码
    └── SimpleStorage.sol     # 示例智能合约
```

## 配置说明

在`application.properties`中配置以太坊网络连接参数：

```properties
ethereum.network.url=http://localhost:8545    # 以太坊节点URL (可以是本地节点如Ganache或远程节点)
ethereum.network.chainId=1337                # 链ID
ethereum.account.privateKey=0x...            # 私钥(测试用，生产环境应使用安全的密钥管理)
ethereum.gas-price=20000000000              # Gas价格(wei)
ethereum.gas-limit=6721975                  # Gas上限
ethereum.contract.address=0x...             # 已部署的合约地址
```

## 使用说明

### 1. 准备工作

1. 安装并启动本地以太坊测试节点(如Ganache)或连接到测试网络
2. 部署`SimpleStorage.sol`合约到网络
3. 更新`application.properties`中的合约地址

### 2. 编译和运行

```bash
# 编译项目
mvn clean package

# 运行项目
mvn spring-boot:run
```

### 3. API接口

- 获取合约存储的值:
  ```
  GET /api/contract/value
  ```

- 设置新值:
  ```
  POST /api/contract/value
  Content-Type: application/json
  
  {
    "value": "42"
  }
  ```

## 扩展功能

- 可以添加更多合约接口
- 实现事件数据的持久化存储
- 添加用户认证和授权
- 实现更复杂的业务逻辑

## 注意事项

- 本项目为示例性质，生产环境使用时需增强安全性
- 私钥不应直接存储在配置文件中，建议使用安全的密钥管理方案
- 考虑使用Web3j的合约自动生成工具生成合约包装类：
  ```
  web3j generate solidity -a=./src/main/resources/contracts/SimpleStorage.sol -o=./src/main/java -p=com.example.ethereum.contract
  ```

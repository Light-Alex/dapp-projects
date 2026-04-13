# 项目：chain-parse-service

## 命令
### 启动 Parser（选择一条链）
```bash
make run-parser CHAIN=bsc        # BSC 链
make run-parser CHAIN=ethereum   # Ethereum 链
make run-parser CHAIN=solana     # Solana 链
make run-parser CHAIN=sui        # Sui 链
```

### 启动 API 服务
```bash
make run-api
```

### 常用命令
```bash
make build-parser    # 编译 parser
make build-api       # 编译 api
make build-all       # 编译 parser + api
make test            # 运行测试
make test-race       # 带竞态检测
make test-cover      # 测试覆盖率
make vet             # 静态检查
make fmt             # 格式化代码
make docker-up       # 启动 Docker 基础设施
make docker-down     # 停止 Docker
make docker-logs     # 查看日志
make clean           # 清理构建产物
```

## 架构
- 见`docs/architecture.md`

## 沟通方式
- 使用中文回复所有问题
- 代码注释使用中文, 并且保留原注释

## 设计原则
- 组合优于继承
- 接口驱动设计
- 工厂模式
- 统一数据模型
- 批量处理
- 事务安全
- 线程安全
- 依赖注入
- 优雅关闭

## 详细文档
- 项目简介：`README.md`
- 产品需求文档: `PRD.md`
- API接口: `docs/api-reference.md`
- 系统架构文档: `docs/architecture.md`
- 数据库设计: `docs/database-design.md`
- 部署文档: `docs/deployment.md`
- DEX 协议技术文档: `docs/dex-protocols.md`
- 技术选型与架构设计分析: `docs/technology-selection.md`
- 测试用例文档: `docs/test-cases.md`

## 注意事项
- 画图工具使用 `mermaid`
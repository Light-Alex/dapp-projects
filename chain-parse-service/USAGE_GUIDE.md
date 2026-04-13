# Chain Parse Service 使用指南

本指南涵盖项目的部署、配置、第三方服务交互和 API 调用。

---

## 目录

1. [快速部署](#1-快速部署)
2. [第三方服务交互](#2-第三方服务交互)
3. [API 接口调用](#3-api-接口调用)
4. [常见操作](#4-常见操作)
5. [故障排查](#5-故障排查)

---

## 1. 快速部署

### 1.1 环境要求

| 组件 | 版本要求 |
|------|----------|
| Go | 1.21+ |
| Docker | 20.10+ |
| Docker Compose | 2.0+ |
| Make | 任意版本 |

### 1.2 一键启动（推荐）

```bash
# 1. 进入项目目录
cd /mnt/e/web3_workspace/dapp_projects/chain-parse-service

# 2. 启动所有基础设施（PostgreSQL, MySQL, Redis, InfluxDB, Grafana）
make docker-up

# 或使用 docker compose
docker compose -f docker/docker-compose.yml up -d

# 3. 验证服务状态
docker compose -f docker/docker-compose.yml ps
```

### 1.3 服务端口

| 服务 | 默认端口 | 访问地址 |
|------|----------|----------|
| API | 8081 | http://localhost:8081 |
| PostgreSQL | 5432 | localhost:5432 |
| MySQL | 3306 | localhost:3306 |
| Redis | 6379 | localhost:6379 |
| InfluxDB | 8086 | http://localhost:8086 |
| Grafana | 3000 | http://localhost:3000 |

### 1.4 启动 Parser

```bash
# 方式1: 使用 Make（推荐）
make run-parser CHAIN=bsc        # BSC 链
make run-parser CHAIN=ethereum   # Ethereum 链
make run-parser CHAIN=solana     # Solana 链
make run-parser CHAIN=sui        # Sui 链

# 方式2: 直接运行二进制文件
./bin/parser -chain bsc -config configs/bsc.yaml

# 方式3: 同时运行多条链（后台运行）
make run-parser CHAIN=bsc &
make run-parser CHAIN=ethereum &
```

### 1.5 启动 API 服务

```bash
# 方式1: 使用 Make
make run-api

# 方式2: 直接运行
./bin/api -config configs/api.yaml

# 默认监听端口: 8081
```

### 1.6 停止服务

```bash
# 停止所有容器
make docker-down

# 或
docker compose -f docker/docker-compose.yml down

# 停止并删除所有数据（慎用！）
docker compose -f docker/docker-compose.yml down -v
```

---

## 2. 第三方服务交互

### 2.1 InfluxDB

#### 访问 Web UI

```bash
# 浏览器访问
http://localhost:8086

# 默认登录信息（见 docker/docker-compose.yml）
用户名: admin
密码: admin123456
组织: unified-tx-parser
Token: unified-tx-parser-token-2024
```

#### 常用 CLI 命令

```bash
# 进入 InfluxDB 容器
docker exec -it chain_parse_influxdb bash

# 查看 bucket 列表
influx bucket list --org unified-tx-parser

# 查询最近1小时的数据
influx query 'from(bucket:"bsc") |> range(start: -1h) |> limit(n: 5)' \
  --org unified-tx-parser \
  --token unified-tx-parser-token-2024

# 删除 bucket 数据（清空）
influx delete \
  --org unified-tx-parser \
  --bucket bsc \
  --start 1970-01-01T00:00:00Z \
  --stop $(date -u +"%Y-%m-%dT%H:%M:%SZ")

# 退出容器
exit
```

#### Data Explorer 查询示例

```flux
// 查询最近交易
from(bucket: "bsc")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "transactions")
  |> limit(n: 10)

// 查询 DEX 交易
from(bucket: "ethereum")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "dex_transactions")
  |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")

// 统计每小时交易量
from(bucket: "solana")
  |> range(start: -1d)
  |> filter(fn: (r) => r._measurement == "transactions")
  |> aggregateWindow(every: 1h, fn: count)
```

### 2.2 Grafana

#### 访问面板

```bash
# 浏览器访问
http://localhost:3000

# 默认登录
用户名: admin
密码: admin
```

#### 添加 InfluxDB 数据源

1. 登录 Grafana
2. 左侧菜单 → **Configuration** → **Data sources**
3. 点击 **Add data source**
4. 选择 **InfluxDB**
5. 配置：
   - **Name**: `InfluxDB`
   - **Query Language**: `Flux`
   - **URL**: `http://chain_parse_influxdb:8086`
   - **Organization**: `unified-tx-parser`
   - **Token**: `unified-tx-parser-token-2024`
6. 点击 **Save & Test**

#### 导入仪表板

1. 左侧菜单 → **Dashboards** → **Import**
2. 粘贴仪表板 JSON 或输入 ID
3. 选择数据源

### 2.3 PostgreSQL

#### 命令行访问

```bash
# 进入容器
docker exec -it chain_parse_postgres psql -U postgres -d unified_tx_parser

# 或从主机连接
psql -h localhost -U postgres -d unified_tx_parser

# 常用查询
\dt                          # 列出所有表
SELECT * FROM transactions LIMIT 10;
SELECT * FROM dex_transactions ORDER BY created_at DESC LIMIT 10;
SELECT COUNT(*) FROM transactions WHERE chain_type = 'bsc';

# 退出
\q
```

#### GUI 工具连接

| 工具 | 配置 |
|------|------|
| DBeaver | Host: localhost, Port: 5432, DB: unified_tx_parser, User: postgres, Password: password |
| pgAdmin | 同上 |
| TablePlus | 同上 |

### 2.4 MySQL

#### 命令行访问

```bash
# 进入容器
docker exec -it chain_parse_mysql mysql -u parser_user -p unified_tx_parser
# 密码: parser_pass

# 或从主机连接
mysql -h localhost -u parser_user -p unified_tx_parser

# 常用查询
SHOW TABLES;
SELECT * FROM transactions LIMIT 10;
SELECT COUNT(*) FROM transactions WHERE chain_type = 'ethereum';

# 退出
exit;
```

### 2.5 Redis

#### 命令行访问

```bash
# 进入容器
docker exec -it chain_parse_redis redis-cli

# 或直接命令
docker exec -it chain_parse_redis redis-cli ping

# 常用命令
KEYS *                    # 查看所有 key
GET progress:bsc          # 获取 BSC 链进度
SET progress:bsc 12345    # 设置 BSC 链进度
DEL progress:bsc          # 删除进度

# 监控
MONITOR

# 退出
exit
```

---

## 3. API 接口调用

### 3.1 API 基本信息

| 项目 | 值 |
|------|-----|
| 基础 URL | `http://localhost:8081` |
| 协议 | HTTP |
| 数据格式 | JSON |
| 字符编码 | UTF-8 |

### 3.2 接口列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/transactions/:hash` | 根据哈希查询交易 |
| GET | `/api/v1/storage/stats` | 存储统计 |
| GET | `/api/v1/progress` | 解析进度 |
| GET | `/api/v1/progress/stats` | 全局统计 |

### 3.3 调用示例

#### cURL

```bash
# 健康检查
curl http://localhost:8081/health

# 查询交易（EVM 链，BSC/Ethereum）
curl http://localhost:8081/api/v1/transactions/0xabc123def456789...

# 查询交易（Solana）
curl http://localhost:8081/api/v1/transactions/5VERv8NMhJr4fE9K...

# 存储统计
curl http://localhost:8081/api/v1/storage/stats

# 解析进度
curl http://localhost:8081/api/v1/progress

# 全局统计
curl http://localhost:8081/api/v1/progress/stats
```

#### Python

```python
import requests

BASE_URL = "http://localhost:8081"

# 健康检查
response = requests.get(f"{BASE_URL}/health")
print(response.json())

# 查询交易
tx_hash = "0xabc123def456789..."
response = requests.get(f"{BASE_URL}/api/v1/transactions/{tx_hash}")
print(response.json())

# 存储统计
response = requests.get(f"{BASE_URL}/api/v1/storage/stats")
print(response.json())

# 解析进度
response = requests.get(f"{BASE_URL}/api/v1/progress")
print(response.json())
```

#### JavaScript / Node.js

```javascript
const BASE_URL = 'http://localhost:8081';

// 健康检查
fetch(`${BASE_URL}/health`)
  .then(r => r.json())
  .then(data => console.log(data));

// 查询交易
const txHash = '0xabc123def456789...';
fetch(`${BASE_URL}/api/v1/transactions/${txHash}`)
  .then(r => r.json())
  .then(data => console.log(data));

// 存储统计
fetch(`${BASE_URL}/api/v1/storage/stats`)
  .then(r => r.json())
  .then(data => console.log(data));
```

#### Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func main() {
    baseURL := "http://localhost:8081"

    // 健康检查
    resp, _ := http.Get(baseURL + "/health")
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))

    // 查询交易
    txHash := "0xabc123def456789..."
    resp, _ = http.Get(baseURL + "/api/v1/transactions/" + txHash)
    body, _ = io.ReadAll(resp.Body)

    var result map[string]interface{}
    json.Unmarshal(body, &result)
    fmt.Printf("%+v\n", result)
}
```

### 3.4 响应格式

**成功响应：**
```json
{
  "transaction": {
    "tx_hash": "0xabc123...",
    "chain_type": "bsc",
    "block_number": 12345678,
    "from_address": "0x123...",
    "to_address": "0x456...",
    "value": "1000000000000000000",
    "status": "success",
    "timestamp": "2026-03-08T10:30:00Z"
  }
}
```

**错误响应：**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "transaction not found",
    "request_id": "a1b2c3d4e5f67890..."
  }
}
```

---

## 4. 常见操作

### 4.1 查看日志

```bash
# 查看所有服务日志
docker compose -f docker/docker-compose.yml logs -f

# 查看特定服务日志
docker compose -f docker/docker-compose.yml logs -f influxdb
docker compose -f docker/docker-compose.yml logs -f postgres
docker compose -f docker/docker-compose.yml logs -f grafana

# 查看 parser 日志（如果运行在容器中）
docker logs -f chain_parse_parser
```

### 4.2 清空数据

```bash
# 清空 InfluxDB bucket（例如 bsc）
docker exec -it chain_parse_influxdb influx delete \
  --org unified-tx-parser \
  --bucket bsc \
  --start 1970-01-01T00:00:00Z \
  --stop $(date -u +"%Y-%m-%dT%H:%M:%SZ")

# 清空 PostgreSQL 数据
docker exec -it chain_parse_postgres psql -U postgres -d unified_tx_parser \
  -c "TRUNCATE TABLE transactions, dex_transactions CASCADE;"

# 清空 MySQL 数据
docker exec -it chain_parse_mysql mysql -u parser_user -pparser_pass unified_tx_parser \
  -e "TRUNCATE TABLE transactions; TRUNCATE TABLE dex_transactions;"

# 重置所有数据（删除卷）
docker compose -f docker/docker-compose.yml down -v
docker compose -f docker/docker-compose.yml up -d
```

### 4.3 重置解析进度

```bash
# 进入 Redis
docker exec -it chain_parse_redis redis-cli

# 删除特定链的进度
DEL progress:bsc
DEL progress:ethereum
DEL progress:solana
DEL progress:sui

# 删除所有进度
KEYS progress:*
# 然后逐个删除

# 退出
exit
```

### 4.4 备份与恢复

```bash
# 备份 InfluxDB
docker exec chain_parse_influxdb influx backup /backup/backup-$(date +%Y%m%d)
docker cp chain_parse_influxdb:/backup/backup-$(date +%Y%m%d) ./backups/

# 备份 PostgreSQL
docker exec chain_parse_postgres pg_dump -U postgres unified_tx_parser > backup_pg.sql

# 备份 MySQL
docker exec chain_parse_mysql mysqldump -u parser_user -pparser_pass unified_tx_parser > backup_mysql.sql
```

---

## 5. 故障排查

### 5.1 服务无法启动

```bash
# 检查端口占用
netstat -tuln | grep -E "8081|8086|3000|5432|3306|6379"

# 检查容器状态
docker ps -a

# 查看容器日志
docker logs chain_parse_influxdb
docker logs chain_parse_postgres
```

### 5.2 Parser 连接失败

```bash
# 检查配置文件
cat configs/bsc.yaml | grep -E "host|port"

# 测试数据库连接
docker exec chain_parse_postgres pg_isready -U postgres
docker exec chain_parse_influxdb influx ping

# 检查网络
docker network ls
docker network inspect docker_chain_parse
```

### 5.3 API 返回错误

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| `NOT_FOUND` | 交易不存在 | 确认 parser 正在运行且已处理该交易 |
| `SERVICE_UNAVAILABLE` | Redis 未配置 | 检查 Redis 是否启动 |
| `INTERNAL_ERROR` | 服务器错误 | 查看 API 日志 |

### 5.4 InfluxDB 时区问题

InfluxDB 内部使用 UTC 存储，时区差异仅影响显示。

**解决方案：**
1. 在 InfluxDB UI 中设置时区：Settings → Preferences → Timezone → `Asia/Shanghai`
2. 或在查询时转换：`|> map(fn: (r) => ({ r with _time: r._time + duration(h: 8) }))`

---

## 6. 配置文件说明

### 6.1 配置结构

```
configs/
├── base.yaml      # 共享基础配置
├── api.yaml       # API 服务配置
├── bsc.yaml       # BSC 链配置
├── ethereum.yaml  # Ethereum 链配置
├── solana.yaml    # Solana 链配置
└── sui.yaml       # Sui 链配置
```

### 6.2 存储引擎切换

编辑 `configs/base.yaml`：

```yaml
storage:
  type: "influxdb"   # 可选: influxdb, pgsql, mysql
```

### 6.3 环境变量覆盖

```bash
# 覆盖端口
export API_PORT=9090
export INFLUXDB_PORT=9086

# 覆盖数据库密码
export POSTGRES_PASSWORD=your_password
export INFLUXDB_PASSWORD=your_password
```

---

## 7. 更多资源

- [项目 README](README.md)
- [API 详细文档](docs/api-reference.md)
- [系统架构](docs/architecture.md)
- [部署文档](docs/deployment.md)

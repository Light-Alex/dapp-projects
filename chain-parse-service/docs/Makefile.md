# Makefile 命令详解

本文档详细解释 `Makefile` 中各个命令的作用和用法。

## 目录

- [变量定义](#变量定义)
- [构建命令](#构建命令)
- [跨平台构建](#跨平台构建)
- [测试命令](#测试命令)
- [代码质量](#代码质量)
- [运行命令](#运行命令)
- [Docker 命令](#docker-命令)
- [其他命令](#其他命令)

---

## 变量定义

### 变量说明

```makefile
BINARY_DIR  := bin              # 二进制文件输出目录
PARSER_BIN  := bin/parser       # Parser 二进制文件路径
API_BIN     := bin/api          # API 二进制文件路径
MODULE      := go.mod 中的模块名 # Go 模块名称
DOCKER_DIR  := docker           # Docker 配置目录
```

### 版本信息变量

```makefile
VERSION     := git tag 或 "dev"        # 版本号，从 git tag 获取
COMMIT      := git commit hash         # 提交哈希（短格式）
BUILD_TIME  := 当前 UTC 时间           # 构建时间
LDFLAGS     := Go 编译链接标志         # 注入版本信息的 ldflags
```

**LDFLAGS 说明**：
- `-s -w`: 去除调试信息，减小二进制文件大小
- `-X main.version=$(VERSION)`: 注入版本号
- `-X main.commit=$(COMMIT)`: 注入提交哈希
- `-X main.buildTime=$(BUILD_TIME)`: 注入构建时间

### Docker 相关变量

```makefile
DOCKER_REGISTRY ?=              # Docker 注册表前缀（可选）
IMAGE_PREFIX    ?= chain-parse  # 镜像名称前缀
PARSER_IMAGE    := chain-parse-parser  # Parser 镜像名
API_IMAGE       := chain-parse-api     # API 镜像名
```

---

## 构建命令

### build-parser

**命令**: `make build-parser`

**作用**: 构建 parser 二进制文件

**执行内容**:
1. 创建 `bin/` 目录（如果不存在）
2. 使用 ldflags 注入版本信息
3. 编译 `cmd/parser/` 目录下的代码
4. 输出到 `bin/parser`

**示例**:
```bash
make build-parser
```

---

### build-api

**命令**: `make build-api`

**作用**: 构建 api 二进制文件

**执行内容**:
1. 创建 `bin/` 目录（如果不存在）
2. 使用 ldflags 注入版本信息
3. 编译 `cmd/api/` 目录下的代码
4. 输出到 `bin/api`

**示例**:
```bash
make build-api
```

---

### build-all

**命令**: `make build-all`

**作用**: 构建所有二进制文件（parser + api）

**执行内容**: 依次执行 `build-parser` 和 `build-api`

**示例**:
```bash
make build-all
```

---

## 跨平台构建

### build-linux-amd64

**命令**: `make build-linux-amd64`

**作用**: 为 Linux AMD64 架构构建二进制文件

**输出文件**:
- `bin/parser-linux-amd64`
- `bin/api-linux-amd64`

**环境变量**:
- `CGO_ENABLED=0`: 禁用 CGO（静态链接）
- `GOOS=linux`: 目标操作系统为 Linux
- `GOARCH=amd64`: 目标架构为 AMD64

---

### build-linux-arm64

**命令**: `make build-linux-arm64`

**作用**: 为 Linux ARM64 架构构建二进制文件

**输出文件**:
- `bin/parser-linux-arm64`
- `bin/api-linux-arm64`

**适用场景**: ARM 服务器（如 AWS Graviton、Apple Silicon）

---

### build-darwin-amd64

**命令**: `make build-darwin-amd64`

**作用**: 为 macOS Intel 架构构建二进制文件

**输出文件**:
- `bin/parser-darwin-amd64`
- `bin/api-darwin-amd64`

---

### build-darwin-arm64

**命令**: `make build-darwin-arm64`

**作用**: 为 macOS Apple Silicon 架构构建二进制文件

**输出文件**:
- `bin/parser-darwin-arm64`
- `bin/api-darwin-arm64`

**适用场景**: M1/M2/M3 Mac

---

### build-cross

**命令**: `make build-cross`

**作用**: 构建所有平台的二进制文件

**执行内容**: 依次构建以下平台
- Linux AMD64
- Linux ARM64
- Darwin (macOS) AMD64
- Darwin (macOS) ARM64

**输出位置**: `bin/` 目录

---

## 测试命令

### test

**命令**: `make test`

**作用**: 运行所有测试

**执行内容**: `go test ./...`

**说明**: 运行项目中所有包的测试

---

### test-race

**命令**: `make test-race`

**作用**: 运行测试并检测竞态条件

**执行内容**: `go test -race ./...`

**说明**: 使用 Go 的竞态检测器查找并发问题

---

### test-cover

**命令**: `make test-cover`

**作用**: 运行测试并生成覆盖率报告

**执行内容**:
1. 运行测试并生成 `coverage.out` 数据文件
2. 在终端显示函数级别的覆盖率
3. 生成 `coverage.html` 可视化报告

**输出文件**:
- `coverage.out`: 覆盖率原始数据
- `coverage.html`: 可在浏览器中查看的覆盖率报告

**查看报告**: 在浏览器中打开 `coverage.html`

---

### test-short

**命令**: `make test-short`

**作用**: 运行短测试（跳过耗时测试）

**执行内容**: `go test -short ./...`

**说明**: 跳过标记了 `+build !short` 的测试

**适用场景**: 快速验证，开发过程中频繁运行

---

## 代码质量

### lint

**命令**: `make lint`

**作用**: 运行 golangci-lint 代码检查

**执行内容**: `golangci-lint run ./...`

**说明**: 需要先安装 [golangci-lint](https://golangci-lint.run/)

**安装**:
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

---

### fmt

**命令**: `make fmt`

**作用**: 格式化代码

**执行内容**:
1. `gofmt -s -w .`: 格式化并简化代码
2. `goimports -w .`: 自动管理 import 语句（如果已安装）

**说明**:
- `-s`: 简化代码（如合并相邻的 if 语句）
- `-w`: 直接修改文件

---

### vet

**命令**: `make vet`

**作用**: 运行 Go 静态分析

**执行内容**: `go vet ./...`

**说明**: 检查常见的 Go 代码问题

---

## 运行命令

### run-parser

**命令**: `make run-parser CHAIN=<chain>`

**作用**: 构建并运行 parser

**参数**:
- `CHAIN`: 必需，指定区块链类型

**支持的链**: `bsc`, `ethereum`, `solana`, `sui`

**示例**:
```bash
make run-parser CHAIN=bsc
make run-parser CHAIN=solana
```

**执行流程**:
1. 构建 parser 二进制文件
2. 使用 `-chain` 参数运行

---

### run-api

**命令**: `make run-api`

**作用**: 构建并运行 API 服务

**示例**:
```bash
make run-api
```

**说明**: API 默认监听 8081 端口（参见配置文件）

---

## Docker 命令

### docker-build

**命令**: `make docker-build`

**作用**: 构建所有 Docker 镜像

**执行内容**: 依次构建 parser 和 api 镜像

---

### docker-build-parser

**命令**: `make docker-build-parser`

**作用**: 构建 parser Docker 镜像

**构建参数**:
- `SERVICE=parser`: 指定构建 parser 服务
- `VERSION`: 版本标签
- `COMMIT`: 提交哈希

**镜像标签**:
- `chain-parse-parser:<VERSION>`: 带版本号
- `chain-parse-parser:latest`: latest 标签

---

### docker-build-api

**命令**: `make docker-build-api`

**作用**: 构建 api Docker 镜像

**构建参数**: 同 docker-build-parser

**镜像标签**:
- `chain-parse-api:<VERSION>`: 带版本号
- `chain-parse-api:latest`: latest 标签

---

### docker-up

**命令**: `make docker-up`

**作用**: 启动所有 Docker 服务

**执行内容**: `docker compose up -d`

**说明**: 后台启动 docker-compose.yml 中定义的所有服务

**包含服务**: PostgreSQL, Redis, 其他依赖服务

---

### docker-down

**命令**: `make docker-down`

**作用**: 停止所有 Docker 服务

**执行内容**: `docker compose down`

**说明**: 停止并移除容器、网络

---

### docker-ps

**命令**: `make docker-ps`

**作用**: 查看运行中的 Docker 服务

**执行内容**: `docker compose ps`

---

### docker-logs

**命令**: `make docker-logs`

**作用**: 查看所有服务的日志

**执行内容**: `docker compose logs -f`

**说明**: `-f` 表示持续跟踪日志输出

**退出**: 按 `Ctrl+C`

---

## 其他命令

### clean

**命令**: `make clean`

**作用**: 清理构建产物

**删除内容**:
- `bin/` 目录（所有二进制文件）
- `coverage.out`（覆盖率数据）
- `coverage.html`（覆盖率报告）

---

### help

**命令**: `make help`

**作用**: 显示帮助信息

**执行内容**: 打印所有可用的 make 目标及其说明

---

## 使用技巧

### 覆盖版本号

```bash
make VERSION=v1.0.0 build-all
```

### 使用自定义 Docker 注册表

```bash
make DOCKER_REGISTRY=registry.example.com docker-build
```

### 链式命令

```bash
make fmt lint test           # 格式化、检查、测试
make build-all test-cover    # 构建并测试覆盖率
```

---

## .PHONY 声明

Makefile 中声明的 `.PHONY` 目标表示这些是"伪目标"（不是实际的文件）：

- build-parser, build-api, build-all
- test, test-race, test-cover, test-short
- lint, fmt, vet
- run-parser, run-api
- docker-*, build-cross
- clean, help

**作用**: 告诉 make 这些命令不产生对应文件名的输出，避免与同名文件冲突。

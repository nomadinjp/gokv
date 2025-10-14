### **任务清单：Go K-V 存储服务 (gokv)**

#### 阶段 1: 项目初始化与基础设置

- [x] 初始化 Go 模块: `go mod init gokv`
- [x] 添加核心依赖: `go get github.com/gin-gonic/gin github.com/dgraph-io/badger/v4 github.com/golang-jwt/jwt/v5`
- [x] 创建项目目录结构 (例如 `cmd/`, `internal/`)

#### 2: 数据存储层 (BadgerDB)

- [x] 在 `internal/storage` 创建一个 BadgerDB 封装。
- [x] 实现 `NewStorage(path string)` 函数，用于初始化数据库。
- [x] 实现 `Close()` 方法，用于优雅关闭数据库连接。
- [x] 实现 `Set(bucket, key string, value []byte)` 方法，内部键使用 `bucket:key` 格式。
- [x] 实现 `Get(bucket, key string)` 方法，用于按 `bucket:key` 读取数据。
- [x] 实现 `DB_PATH` 环境变量的读取，并设置默认值 `./data`。
- [ ] **实现列表功能**:
    - [ ] 实现 `ListBuckets()` 方法，返回所有 bucket 名称的列表。
    - [ ] 实现 `ListKeys(bucket string)` 方法，返回指定 bucket 下所有 key 的列表。

#### 阶段 3: JWT 生成命令行工具 (`jwt-gen`)

- [x] 在 `cmd/jwt-gen` 目录下创建 `main.go`。
- [x] 使用 `flag` 包解析 `--secret` 和 `--expires-in` 命令行参数。
- [x] 实现 JWT 生成逻辑，使用 `HS256` 算法。
- [x] 将生成的 Token 字符串打印到标准输出。
- [x] 编写构建脚本或说明，用于编译 `jwt-gen`。

#### 阶段 4: HTTP API 服务器

- [x] 在 `cmd/gokv` 目录下创建 `main.go`。
- [x] 设置 Gin 引擎和路由。
- [x] **实现 JWT 鉴权中间件**:
    - [x] 从 `JWT_SECRET` 环境变量读取密钥。
    - [x] 从 `Authorization: Bearer <token>` 请求头中提取 Token。
    - [x] 验证 Token 的签名和有效期。
    - [x] 如果验证失败，返回 `401 Unauthorized`。
- [x] **实现 API 端点**:
    - [x] `POST /:bucket/:key`:
        - [x] 绑定 `:bucket` 和 `:key` 路径参数。
        - [x] 检查 `bucket` 或 `key` 是否为空，如果为空则返回 `400 Bad Request`。
        - [x] 读取原始请求体作为 `value`。
        - [x] 调用 `storage.Set()` 方法存入数据。
        - [x] 成功后返回 `200 OK`。
    - [x] `GET /:bucket/:key`:
        - [x] 绑定路径参数。
        - [x] 调用 `storage.Get()` 方法获取数据。
            - [x] 如果键不存在，返回 `404 Not Found`。
            - [x] 成功后返回 `200 OK`，响应体为获取到的数据。
        - [ ] **实现列表 API 端点**:
            - [ ] `GET /_list`:
                - [ ] 绑定查询参数 `bucket`。
                - [ ] 如果 `bucket` 为空，调用 `storage.ListBuckets()` 并返回 JSON 列表。
                - [ ] 如果 `bucket` 不为空，调用 `storage.ListKeys(bucket)` 并返回 JSON 列表。
                - [ ] 确保返回的列表是 JSON 格式，即使为空也返回 `[]`。
        - [x] **实现主函数 (`main`)**:    - [x] 读取 `PORT` 和 `GIN_MODE` 环境变量。
    - [x] 初始化存储层。
    - [x] 设置信号处理器（`os.Signal`, `syscall.SIGINT`, `syscall.SIGTERM`）以实现优雅停机（关闭数据库）。
    - [x] 启动 Gin 服务器。

#### 阶段 5: 容器化 (Dockerfile)

- [x] 创建 `Dockerfile` 文件。
- [x] **定义 `builder` 阶段**:
    - [x] 使用 `golang:1.21-alpine` 作为基础镜像。
    - [x] 设置工作目录。
    - [x] 复制 `go.mod` 和 `go.sum` 并执行 `go mod download`。
    - [x] 复制所有源代码。
    - [x] 编译 `gokv` 服务器 (`CGO_ENABLED=0 GOOS=linux`)。
    - [x] 编译 `jwt-gen` 工具 (`CGO_ENABLED=0 GOOS=linux`)。
- [x] **定义 `final` 阶段**:
    - [x] 使用 `gcr.io/distroless/static-debian12` 作为基础镜像。
    - [x] 从 `builder` 阶段复制 `gokv` 二进制文件到 `/`。
    - [x] 使用 `EXPOSE 8080` 暴露端口。
    - [x] 设置 `CMD ["/gokv"]` 作为容器启动命令。

#### 阶段 6: 文档和收尾

- [x] 创建 `README.md`，说明如何构建、配置和运行服务。
- [x] 详细说明所有环境变量的用途。
- [x] 提供 `jwt-gen` 工具的使用示例。
- [x] 提供 `curl` 示例来演示 API 的使用方法。

---
#### 阶段 7: 删除功能扩展

- [x] **数据存储层**: 在 `internal/storage/storage.go` 中实现 `Delete(bucket, key string)` 方法。
- [x] **HTTP API 服务器**: 在 `internal/handler/kv.go` 中实现 `DeleteHandler`。
- [x] **HTTP API 服务器**: 在 `cmd/gokv/main.go` 中注册 `DELETE /:bucket/:key` 路由。
- [x] **文档**: 更新 `README.md` 中的 API 示例。

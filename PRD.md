### **需求文档：Go K-V 存储服务**

#### 1. 项目概述

本项目旨在开发一个高性能、轻量级的 Key-Value 数据存储服务器。该服务器使用 Go 语言实现，提供基于 HTTP 的 RESTful API 接口，并采用 BadgerDB 作为底层存储引擎。所有 API 接口都将通过 JWT (JSON Web Tokens) 进行鉴权保护。项目还需包含一个独立的命令行工具用于生成 JWT，并提供一个多阶段构建的 Dockerfile 以便容器化部署。

#### 2. 技术栈与核心依赖库

*   **编程语言**: Go
*   **HTTP 框架**: Gin (`github.com/gin-gonic/gin`)
*   **数据存储**: BadgerDB (`github.com/dgraph-io/badger/v4`)
*   **JWT 鉴权**: `github.com/golang-jwt/jwt/v5` 或类似标准库

#### 3. 核心功能模块

##### 3.1. HTTP API 服务器 (Gin)
服务器负责提供数据读写接口，并强制执行 JWT 鉴权。

*   **鉴权中间件**:
    *   所有数据操作路由 (`/:bucket/:key`) 都必须经过 JWT 鉴权中间件的处理。
    *   JWT 必须通过 HTTP `Authorization` 请求头以 `Bearer <token>` 的形式提供。
    *   用于验证 JWT 签名的密钥（Secret）必须从名为 `JWT_SECRET` 的环境变量中读取。
    *   如果请求未提供 Token，或 Token 无效、格式错误、已过期，服务器必须拒绝请求并返回 `401 Unauthorized` 状态码。

*   **API 接口定义**:
    *   **写入数据**: `POST /:bucket/:key`
        *   **功能**: 创建或更新一个键值对。
        *   **路径参数**:
            *   `:bucket` (string): 数据的逻辑分组，作为键的前缀。
            *   `:key` (string): 数据的唯一标识符。
        *   **请求体 (Body)**: 整个原始请求体 (raw request body) 将作为要存储的值 (value)。
        *   **成功响应**: `200 OK` 状态码，响应体可为空。
    *   **读取数据**: `GET /:bucket/:key`
        *   **功能**: 根据 bucket 和 key 读取一个值。
        *   **路径参数**:
            *   `:bucket` (string): 数据的逻辑分组。
            *   `:key` (string): 数据的唯一标识符。
        *   **成功响应**: `200 OK` 状态码，响应体为存储的原始数据。`Content-Type` 头可设置为 `application/octet-stream`。
    *   **删除数据**: `DELETE /:bucket/:key`
        *   **功能**: 根据 bucket 和 key 删除一个值。
        *   **路径参数**:
            *   `:bucket` (string): 数据的逻辑分组。
            *   `:key` (string): 数据的唯一标识符。
        *   **成功响应**: `200 OK` 状态码，响应体可为空。
    *   **列表查询**: `GET /_list`
        *   **功能**: 列出所有 bucket 或指定 bucket 下的所有 key。
        *   **查询参数**:
            *   `bucket` (string, 可选): 如果提供，则列出该 bucket 下的所有 key；如果未提供，则列出所有 bucket。
        *   **成功响应**: `200 OK` 状态码，响应体为 JSON 格式的列表（例如 `["bucket1", "bucket2"]` 或 `["keyA", "keyB"]`）。
        *   **错误处理**:
            *   如果提供了 `bucket` 参数，但该 bucket 不存在或为空，返回 `200 OK` 和一个空列表 `[]`。

*   **错误处理规范**:
    *   `400 Bad Request`: 请求路径中的 `:bucket` 或 `:key` 为空。
    *   `401 Unauthorized`: JWT 鉴权失败。
    *   `404 Not Found`: `GET` 请求的 key 在数据库中不存在。
    *   `500 Internal Server Error`: 发生其他服务器内部错误（如数据库读写失败）。

##### 3.2. 数据存储层 (BadgerDB)
*   **"Bucket" 的实现**:
    *   BadgerDB 是一个纯 Key-Value 存储，没有内建的 "Bucket" 概念。
    *   程序需通过将 `bucket` 和 `key` 组合成一个带前缀的内部键来模拟此功能。
    *   内部键的格式应为 `bucket:key`（使用冒号作为分隔符）。
*   **数据库配置**:
    *   BadgerDB 数据库文件的存储路径应通过名为 `DB_PATH` 的环境变量进行配置。
    *   若未提供该环境变量，可默认使用 `./data` 作为数据目录。
*   **生命周期管理**:
    *   程序启动时需优雅地打开 BadgerDB 数据库连接。
    *   程序接收到终止信号时（如 `SIGINT`, `SIGTERM`），需优雅地关闭数据库连接，确保数据完整性。

##### 3.3. JWT 生成命令行工具
*   **目的**: 提供一个与服务器分离的独立命令行工具，用于生成访问服务器所需的 JWT。
*   **实现**: 应是一个单独的 Go main 包，可编译成独立的可执行文件（例如 `jwt-gen`）。
*   **命令行参数**:
    *   `--secret` (string, 必须): 用于签名的密钥。此值**必须**与服务器使用的 `JWT_SECRET` 环境变量完全一致。
    *   `--expires-in` (string, 必须): Token 的有效期。格式为 Go 的 `time.ParseDuration` 所支持的字符串，例如 `"24h"`, `"30m"`, `"365d"`。
*   **输出**: 成功执行后，程序应将生成的 JWT 字符串打印到标准输出 (stdout)，不应有任何其他多余的输出。
*   **示例用法**:
    ```bash
    $ ./jwt-gen --secret "your-strong-secret-key" --expires-in "720h"
    # 输出: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
    ```

#### 4. 容器化部署 (Dockerfile)

*   **构建策略**: 必须使用**多阶段构建 (multi-stage build)** 来优化最终镜像的大小和安全性。
*   **第一阶段 (Builder)**:
    *   **基础镜像**: 使用官方的 Go 镜像，如 `golang:1.21-alpine`。
    *   **构建步骤**:
        1.  拷贝 `go.mod` 和 `go.sum` 文件，并执行 `go mod download` 缓存依赖。
        2.  拷贝项目所有源代码。
        3.  使用 `CGO_ENABLED=0` 标志编译出静态链接的服务器主程序和 JWT 生成工具。
*   **第二阶段 (Final Image)**:
    *   **基础镜像**: 使用 Google 的 Distroless 静态镜像 `gcr.io/distroless/static-debian12`。
    *   **构建步骤**:
        1.  从 `builder` 阶段拷贝已编译好的服务器二进制文件到镜像中。
        2.  设置 `EXPOSE` 指令暴露服务器监听的端口（如 `8080`）。
        3.  使用 `CMD` 或 `ENTRYPOINT` 指令设置容器的启动命令为运行服务器二进制文件。

#### 5. 配置管理

整个应用（服务器和工具）的配置应完全通过**环境变量**进行管理。

*   **服务器必需的环境变量**:
    *   `JWT_SECRET`: 用于 JWT 签名验证的共享密钥。
    *   `DB_PATH`: BadgerDB 数据库的存储路径。
    *   `PORT` (可选, 默认 `8080`): 服务器监听的端口。
    *   `GIN_MODE` (可选, 默认 `debug`): 建议在生产环境中设置为 `release`。
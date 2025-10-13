### **架构文档：Go K-V 存储服务 (gokv)**

#### 1. 系统概述

`gokv` 是一个高性能、轻量级的 Key-Value 存储服务。它通过一个基于 Go `Gin` 框架构建的 RESTful API 提供服务，使用 `BadgerDB` 作为其持久化存储后端，并采用 `JWT` 进行所有 API 请求的身份验证。系统设计遵循云原生原则，通过环境变量进行配置，并使用多阶段 Docker 构建以实现轻松部署。

#### 2. 核心组件

系统由三个主要部分组成：HTTP API 服务器、数据存储层和一个独立的 JWT 生成工具。

##### 2.1. HTTP API 服务器

服务器是系统的核心，负责处理客户端请求。

*   **框架**: `Gin` (`github.com/gin-gonic/gin`)
*   **职责**:
    *   提供数据写入 (`POST`) 和读取 (`GET`) 的 RESTful 端点。
    *   通过中间件强制执行 JWT 身份验证。
    *   处理 HTTP 请求和响应的序列化/反序列化。
    *   管理应用程序的生命周期，包括优雅启动和关闭。

*   **API 端点**:
    *   `POST /:bucket/:key`: 将请求体作为值，存储在指定的 `bucket` 和 `key` 下。
    *   `GET /:bucket/:key`: 检索并返回与 `bucket` 和 `key` 关联的值。
    *   `DELETE /:bucket/:key`: 删除与 `bucket` 和 `key` 关联的值。

*   **鉴权中间件**:
    *   一个 `Gin` 中间件将拦截所有对数据端点的请求。
    *   它从 `Authorization: Bearer <token>` 请求头中提取 JWT。
    *   使用从 `JWT_SECRET` 环境变量加载的密钥来验证令牌的签名和有效期。
    *   无效或缺失的令牌将导致 `401 Unauthorized` 响应。

##### 2.2. 数据存储层

该层抽象了与底层数据库 `BadgerDB` 的交互。

*   **数据库**: `BadgerDB` (`github.com/dgraph-io/badger/v4`)
*   **职责**:
    *   封装数据库的初始化 (`Open`) 和关闭 (`Close`) 操作。
    *   提供简单的 `Get` 和 `Set` 方法供 API 服务器使用。
    *   实现 "Bucket" 模拟。

*   **Bucket 模拟**:
    *   `BadgerDB` 是一个简单的 K-V 存储，本身不支持 "Bucket"。
    *   通过将用户提供的 `bucket` 和 `key` 与一个分隔符（例如冒号 `:`）组合，可以模拟出命名空间。
    *   内部存储的键格式为 `bucket:key`。这允许在逻辑上对键进行分组，同时保持存储模型的扁平化和高性能。

##### 2.3. JWT 生成命令行工具 (`jwt-gen`)

这是一个独立的、可编译的命令行应用程序，用于生成可用于访问 API 服务器的 JWT。

*   **职责**:
    *   从命令行参数接收签名密钥 (`--secret`) 和有效期 (`--expires-in`)。
    *   使用与服务器相同的 `HS256` 算法生成 JWT。
    *   将生成的令牌字符串输出到标准输出，以便于在脚本中使用。
*   **目的**: 将令牌生成与服务器本身分离，提供了一种安全的、离线的方式来创建访问凭证。

#### 3. 配置管理

整个应用程序的配置完全通过**环境变量**进行管理，这符合十二因子应用（Twelve-Factor App）的最佳实践。

*   `JWT_SECRET`: 用于 JWT 签名和验证的共享密钥。
*   `DB_PATH`: `BadgerDB` 数据库文件的存储路径 (默认为 `./data`)。
*   `PORT`: API 服务器监听的端口 (默认为 `8080`)。
*   `GIN_MODE`: `Gin` 框架的运行模式 (默认为 `debug`)。

#### 4. 容器化 (Dockerfile)

项目使用**多阶段 Docker 构建**来创建优化的生产镜像。

*   **构建阶段 (`builder`)**:
    1.  基于 `golang:1.21-alpine` 镜像。
    2.  下载 Go 模块依赖。
    3.  将 `gokv` 服务器和 `jwt-gen` 工具编译为静态链接的二进制文件 (`CGO_ENABLED=0`)。

*   **最终阶段 (`final`)**:
    1.  基于极简的 `gcr.io/distroless/static-debian12` 镜像。
    2.  从 `builder` 阶段仅复制已编译的 `gokv` 服务器二进制文件。
    3.  设置暴露的端口和容器启动命令。

这种方法确保了最终的 Docker 镜像体积小、攻击面小，并且不包含任何构建工具或源代码。

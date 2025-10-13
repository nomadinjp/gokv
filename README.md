# Go K-V 存储服务 (gokv)

`gokv` 是一个高性能、轻量级的 Key-Value 存储服务器，使用 Go 语言实现，底层采用 BadgerDB，并通过 JWT 鉴权保护所有 API 接口。

## 架构概览

- **服务器**: 基于 Gin 框架，提供 RESTful API。
- **存储**: 使用 BadgerDB，通过 `bucket:key` 格式模拟命名空间。
- **安全**: 所有数据操作均通过 JWT (HS256) 鉴权。
- **部署**: 支持多阶段 Docker 构建，使用 Distroless 镜像。

## 依赖

- Go 1.21+
- Docker (可选)

## 构建与运行

### 1. 本地构建

使用项目根目录下的 `build.sh` 脚本可以编译服务器和 JWT 生成工具。

```bash
# 赋予执行权限
chmod +x build.sh

# 执行构建 (生成 linux/amd64 静态二进制文件到 ./bin 目录)
./build.sh
```

### 2. 配置环境变量

服务器的配置完全依赖于环境变量。

| 变量名 | 描述 | 默认值 | 必需性 |
| :--- | :--- | :--- | :--- |
| `JWT_SECRET` | 用于 JWT 签名和验证的共享密钥。 | 无 | **必需** |
| `DB_PATH` | BadgerDB 数据库文件的存储路径。 | `./data` | 可选 |
| `PORT` | 服务器监听的端口。 | `8080` | 可选 |
| `GIN_MODE` | Gin 框架的运行模式 (`debug`, `release`, `test`)。 | `debug` | 可选 |

### 3. 启动服务器

```bash
# 示例：设置密钥并启动服务器
export JWT_SECRET="your-very-secret-key"
./bin/gokv
# Server starting on :8080 (Mode: debug) 
```

### 4. 容器化部署

使用多阶段构建的 `Dockerfile` 创建最小化镜像。

```bash
# 构建 Docker 镜像
docker build -t gokv:latest .

# 运行容器
# 注意：需要将 JWT_SECRET 传入，并将数据目录挂载到宿主机
docker run -d \
  -p 8080:8080 \
  -e JWT_SECRET="your-very-secret-key" \
  -v $(pwd)/data:/data \
  --name gokv_server \
  gokv:latest
```

## JWT 访问令牌生成

使用独立的 `jwt-gen` 工具生成访问令牌。

```bash
# 示例：生成一个有效期为 720 小时的 Token
./bin/jwt-gen --secret "your-very-secret-key" --expires-in "720h"
# 输出: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## API 接口示例

假设生成的 Token 为 `YOUR_JWT_TOKEN`。

### 1. 写入数据 (POST)

写入一个键值对到 `users` bucket 下的 `user_id_123` 键。

```bash
curl -X POST http://localhost:8080/users/user_id_123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: text/plain" \
  -d "Hello, World!"
# 响应: 200 OK
```

### 2. 读取数据 (GET)

读取 `users` bucket 下的 `user_id_123` 键的值。

```bash
curl -X GET http://localhost:8080/users/user_id_123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
# 响应: 200 OK, 响应体为 "Hello, World!"
```

### 3. 删除数据 (DELETE)

删除 `users` bucket 下的 `user_id_123` 键。

```bash
curl -X DELETE http://localhost:8080/users/user_id_123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
# 响应: 200 OK
```

### 4. 错误示例 (401 Unauthorized)

尝试在没有 Token 的情况下访问。

```bash
curl -X GET http://localhost:8080/users/user_id_123
# 响应: 401 Unauthorized, 响应体为 {"error":"Authorization header required"}
```

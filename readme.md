# img2color-go

一个高性能的图片主色调提取API服务，支持Vercel部署和本地/服务器部署。

## 功能特性

- 🎨 **图片主色调提取**：快速提取图片的主要颜色
- 🖼️ **多格式支持**：支持JPEG、PNG、GIF、WebP、BMP、TIFF等格式
- 🔒 **安全防护**：SSRF防护、速率限制、Referer验证
- ⚡ **高性能**：Redis缓存、连接池、图片压缩优化
- 💾 **数据持久化**：可选MongoDB存储
- 🚀 **多平台部署**：支持Vercel、本地、服务器部署
- 📊 **健康检查**：提供服务状态监控接口

## 快速开始

### 方式一：Vercel部署（推荐）

#### 1. Fork项目

点击项目右上角的 `Fork` 按钮，将项目Fork到你的GitHub账号。

#### 2. 导入到Vercel

1. 访问 [Vercel](https://vercel.com)
2. 点击 `New Project`
3. 选择你Fork的项目
4. 点击 `Import`

#### 3. 配置环境变量

在Vercel项目设置中添加环境变量（Settings → Environment Variables）：

| 变量名 | 说明 | 默认值 | 是否必需 |
|--------|------|--------|----------|
| `ALLOWED_ORIGINS` | 允许的跨域来源 | `*` | 否 |
| `ALLOWED_REFERERS` | 允许的Referer | 空（允许所有） | 否 |
| `RATE_LIMIT` | 速率限制（请求数） | `100` | 否 |
| `RATE_WINDOW` | 速率限制窗口（秒） | `60` | 否 |
| `USE_REDIS_CACHE` | 是否启用Redis | `false` | 否 |
| `REDIS_ADDRESS` | Redis地址 | - | Redis启用时必需 |
| `REDIS_PASSWORD` | Redis密码 | - | 否 |
| `REDIS_DB` | Redis数据库 | `0` | 否 |
| `USE_MONGODB` | 是否启用MongoDB | `false` | 否 |
| `MONGO_URI` | MongoDB连接URI | - | MongoDB启用时必需 |
| `MONGO_DB` | MongoDB数据库 | `img2color` | 否 |
| `MONGO_COLLECTION` | MongoDB集合 | `colors` | 否 |
| `MAX_IMAGE_SIZE` | 最大图片大小（字节） | `10485760` (10MB) | 否 |
| `DOWNLOAD_TIMEOUT` | 下载超时（秒） | `10` | 否 |

#### 4. 部署

点击 `Deploy` 按钮，等待部署完成。

#### 5. 验证部署

访问以下URL验证部署是否成功：

- API接口：`https://your-domain.vercel.app/api?img=https://example.com/image.jpg`
- 健康检查：`https://your-domain.vercel.app/health`

### 方式二：本地/服务器部署

#### 1. 环境要求

- Go 1.20 或更高版本
- Redis（可选，用于缓存）
- MongoDB（可选，用于持久化）

#### 2. 克隆项目

```bash
git clone https://github.com/your-username/img2color-go.git
cd img2color-go
```

#### 3. 安装依赖

```bash
go mod download
```

#### 4. 配置环境变量

复制环境变量模板：

```bash
cp .env.example .env
```

编辑 `.env` 文件，根据实际情况修改配置。

#### 5. 运行服务

```bash
go run app/main.go
```

服务将在配置的端口（默认3000）启动。

#### 6. 验证服务

```bash
# 测试API接口
curl "http://localhost:3000/api?img=https://example.com/image.jpg"

# 测试健康检查
curl http://localhost:3000/health
```

## API文档

### 提取图片主色调

**请求**

```
GET /api?img={image_url}
```

**参数**

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `img` | string | 是 | 图片URL（仅支持http/https协议） |

**响应**

成功响应（200）：

```json
{
  "RGB": "#RRGGBB"
}
```

错误响应：

```json
{
  "code": "ERROR_CODE",
  "message": "错误信息"
}
```

**错误代码**

| 错误代码 | HTTP状态码 | 说明 |
|----------|------------|------|
| `MISSING_IMAGE_URL` | 400 | 缺少img参数 |
| `INVALID_URL` | 400 | 无效的URL格式 |
| `INVALID_PROTOCOL` | 400 | 仅支持http/https协议 |
| `SSRF_ATTACK` | 403 | 禁止访问内网地址 |
| `FORBIDDEN` | 403 | 禁止访问（Referer验证失败） |
| `RATE_LIMIT_EXCEEDED` | 429 | 请求过于频繁 |
| `IMAGE_TOO_LARGE` | 413 | 图片大小超过限制 |
| `IMAGE_DOWNLOAD_FAILED` | 502 | 图片下载失败 |
| `IMAGE_DECODE_FAILED` | 415 | 图片解码失败 |
| `TIMEOUT` | 504 | 请求超时 |

### 健康检查

**请求**

```
GET /health
```

**响应**

```json
{
  "status": "ok",
  "version": "2.0.0",
  "timestamp": "2024-01-01T12:00:00Z",
  "dependencies": {
    "redis": {
      "status": "ok"
    },
    "mongodb": {
      "status": "ok"
    }
  }
}
```

## 安全说明

### SSRF防护

本服务实现了严格的SSRF（服务器端请求伪造）防护：

- ✅ 仅允许 `http` 和 `https` 协议
- ✅ 禁止访问私有IP地址（10.0.0.0/8、172.16.0.0/12、192.168.0.0/16等）
- ✅ 禁止访问回环地址（127.0.0.0/8、::1）
- ✅ 禁止访问链路本地地址
- ✅ 禁止访问特殊主机名（localhost、*.local等）

### 速率限制

默认配置：每个IP在60秒内最多100次请求。

可通过环境变量调整：
- `RATE_LIMIT`：最大请求数
- `RATE_WINDOW`：时间窗口（秒）

### Referer验证

配置 `ALLOWED_REFERERS` 可限制请求来源：

```
ALLOWED_REFERERS=example.com,*.example.com
```

### 图片大小限制

默认限制：10MB

可通过 `MAX_IMAGE_SIZE` 环境变量调整。

## 项目结构

```
img2color-go/
├── api/
│   └── img2color.go      # Vercel入口
├── app/                  # 应用代码
│   ├── main.go           # 本地运行入口
│   └── core/             # 核心业务代码
│       ├── config/       # 配置管理
│       ├── handler/      # HTTP处理器
│       ├── pkg/          # 公共工具包
│       │   ├── errorx/   # 错误处理
│       │   ├── logger/   # 日志
│       │   └── httputil/ # HTTP工具
│       ├── service/      # 业务服务
│       └── storage/      # 存储层
├── .env.example          # 环境变量模板
├── vercel.json           # Vercel配置
├── go.mod                # Go模块定义
└── README.md             # 项目文档
```

## 常见问题

### Q: 如何在Vercel上配置Redis？

A: 推荐使用 [Upstash Redis](https://upstash.com)，它提供了Serverless Redis服务：

1. 在Upstash创建Redis实例
2. 复制连接URL
3. 在Vercel环境变量中设置：
   - `USE_REDIS_CACHE=true`
   - `REDIS_ADDRESS=your-upstash-redis-url`
   - `REDIS_PASSWORD=your-password`

### Q: 如何在Vercel上配置MongoDB？

A: 推荐使用 [MongoDB Atlas](https://www.mongodb.com/atlas)：

1. 在MongoDB Atlas创建集群
2. 获取连接URI
3. 在Vercel环境变量中设置：
   - `USE_MONGODB=true`
   - `MONGO_URI=your-mongodb-uri`

### Q: 图片下载失败怎么办？

A: 可能的原因：
1. 图片URL不可访问
2. 图片大小超过限制（默认10MB）
3. 下载超时（默认10秒）

可通过环境变量调整：
- `MAX_IMAGE_SIZE`：增大大小限制
- `DOWNLOAD_TIMEOUT`：增加超时时间

### Q: 如何自定义速率限制？

A: 通过环境变量配置：
- `RATE_LIMIT=200`：每个IP在时间窗口内最多200次请求
- `RATE_WINDOW=120`：时间窗口为120秒

### Q: 本地开发时如何热重载？

A: 推荐使用 [Air](https://github.com/cosmtrek/air)：

```bash
# 安装Air
go install github.com/cosmtrek/air@latest

# 运行
air
```

## 技术栈

- **语言**：Go 1.20+
- **图片处理**：github.com/disintegration/imaging、github.com/nfnt/resize
- **颜色处理**：github.com/lucasb-eyer/go-colorful
- **缓存**：go-redis/redis/v8
- **数据库**：go.mongodb.org/mongo-driver
- **配置**：github.com/joho/godotenv

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

# 酒店管理系统后端

基于 Go 构建的酒店管理系统后端，采用清晰的分层架构，支持客房预订、入住/退房管理、订单追踪、自动退房通知以及完整的操作审计日志。

## 技术栈

- **Go 1.25**
- **Gin** — HTTP Web 框架
- **GORM** — PostgreSQL ORM
- **Redis** — Token 黑名单与缓存
- **JWT** — 无状态认证（Access Token + Refresh Token）
- **Viper** — 配置管理

## 功能特性

- **RBAC 权限控制** — 三种角色：`guest`（住户）、`employee`（员工）、`admin`（管理员）
- **客房管理** — 支持房型、价格、状态等信息的增删改查
- **订单系统** — 住户可下单，员工处理订单确认与取消
- **入住 / 退房** — 追踪入住信息及预计退房时间
- **自动通知** — 后台调度器自动提醒住户和员工处理超时未退房的情况
- **审计日志** — 所有数据变更均记录不可篡改的日志，包含修改前后的完整快照
- **统一响应格式** — 标准化的 JSON 响应结构与错误码

## 项目结构

```
hotel-backend/
├── config/              # 应用配置（Viper + YAML）
├── internal/
│   ├── database/        # Postgres 与 Redis 初始化、数据迁移、种子数据
│   ├── model/           # GORM 实体定义
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑与后台调度器
│   ├── handler/         # HTTP 处理器（Gin）
│   ├── middleware/      # JWT 认证、RBAC 鉴权、请求日志
│   ├── router/          # 路由注册
│   └── dto/             # 请求/响应结构体
├── pkg/                 # 可复用工具（JWT、bcrypt、错误码）
└── main.go
```

## 快速开始

### 环境要求

- Go 1.25+
- PostgreSQL 14+
- Redis 6+

### 1. 克隆与安装

```bash
git clone <仓库地址>
cd hotel-backend
go mod tidy
```

### 2. 配置

编辑 `config/config.yaml`，或通过环境变量覆盖：

```yaml
server:
  port: "8080"

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "postgres"
  dbname: "hotel"
  sslmode: "disable"

redis:
  addr: "localhost:6379"

jwt:
  secret: "change-me-in-production"
```

### 3. 运行

```bash
go run main.go
```

服务启动后会自动执行数据库迁移，并写入默认管理员账号：

- **用户名：** `admin`
- **密码：** `admin123`

## 接口概览

| 方法 | 接口 | 认证 | 角色 | 说明 |
|------|------|------|------|------|
| POST | `/api/v1/auth/login` | — | — | 登录，获取 JWT Token |
| POST | `/api/v1/auth/logout` | ✅ | — | 登出，使 Token 失效 |
| GET | `/api/v1/rooms` | ✅ | — | 获取客房列表 |
| POST | `/api/v1/orders` | ✅ | — | 创建订单 |
| GET | `/api/v1/orders` | ✅ | 全部 | 获取订单（住户仅能看到自己的） |
| GET | `/api/v1/checkins` | ✅ | employee/admin | 获取入住记录 |
| POST | `/api/v1/checkins` | ✅ | employee/admin | 办理入住 |
| PUT | `/api/v1/checkins/:id/checkout` | ✅ | employee/admin | 办理退房 |
| GET | `/api/v1/notifications` | ✅ | — | 获取通知 |
| GET | `/api/v1/audit-logs` | ✅ | employee/admin | 查看审计日志 |

## 认证方式

API 采用 JWT Bearer Token 认证：

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

后续请求在 Header 中携带 Token：

```bash
curl http://localhost:8080/api/v1/rooms \
  -H "Authorization: Bearer <access_token>"
```

## 开发

```bash
# 运行全部测试
go test ./...

# 运行单个包测试
go test ./pkg/jwt/ -v

# 编译二进制
go build -o hotel-backend.exe .
```

## 许可证

MIT

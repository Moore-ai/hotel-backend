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

- **RBAC 权限控制** — 四种角色：`guest`（住户）、`employee`（员工）、`waiter`（服务员）、`admin`（管理员）
- **客房管理** — 支持房型、价格、状态等信息的增删改查
- **房间自动分配** — 根据入住人数、房型偏好自动分配可用房间，支持低楼层优先策略
- **订单系统** — 住户可下单，员工处理订单确认与取消，支持指定房间或自动分配
- **订单取消审核** — 临近入住的取消进入审核流程，员工可审批或驳回，住户可申诉
- **入住 / 退房** — 追踪入住信息及预计退房时间，支持事务保证数据一致性
- **服务员派单** — 住户发起服务请求，系统随机派单给空闲服务员，实时通知
- **通知系统** — REST API + WebSocket 实时推送，支持多种通知类型
- **ID 加密** — 订单、通知和申诉 ID 使用 Hashids 加密为混淆字符串，不暴露真实 ID
- **AI 智能管家** — 住客通过自然语言对话完成酒店咨询、服务请求、订单查询和申诉处理，支持云端 Anthropic API 和本地 Ollama 模型
- **审计日志** — 所有数据变更均记录不可篡改的日志，包含修改前后的完整快照
- **统一响应格式** — 标准化的 JSON 响应结构与错误码

## 项目结构

```
hotel-backend/
├── config/              # 应用配置（Viper + YAML）
├── internal/
│   ├── database/        # Postgres 与 Redis 初始化、数据迁移、种子数据
│   ├── model/           # GORM 实体定义（User + Guest/Employee/Admin 业务表）
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑与后台调度器
│   ├── handler/         # HTTP 处理器（Gin）
│   ├── middleware/      # JWT 认证、RBAC 鉴权、请求日志
│   ├── router/          # 路由注册
│   └── dto/             # 请求/响应结构体
├── pkg/                 # 可复用工具（JWT、bcrypt、错误码、LLM 客户端）
├── scripts/             # 集成测试脚本
├── test/                # 功能测试脚本
└── main.go
```

## 快速开始

### 环境搭建

#### 安装 Go

下载对应系统的安装包，或使用包管理器：

- **Windows**: 从 [go.dev/dl](https://go.dev/dl/) 下载 MSI 安装包
- **macOS**: `brew install go`
- **Linux**: `sudo apt install golang-go` 或下载二进制包

验证安装：

```bash
go version  # 需输出 go 1.25+
```

#### 安装 PostgreSQL

**Docker（推荐）**：

```bash
docker run -d \
  --name hotel-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=hotel \
  -p 5432:5432 \
  postgres:16
```

**直接安装**：

- **Windows**: 从 [postgresql.org](https://www.postgresql.org/download/windows/) 下载安装程序
- **macOS**: `brew install postgresql@16 && brew services start postgresql@16`
- **Linux**: `sudo apt install postgresql postgresql-contrib`

安装后创建数据库：

```bash
createdb -U postgres hotel
```

#### 安装 Redis

**Docker（推荐）**：

```bash
docker run -d --name hotel-redis -p 6379:6379 redis:7
```

**直接安装**：

- **Windows**: 从 [redis.io/download](https://redis.io/download/) 下载 MSI，或使用 WSL
- **macOS**: `brew install redis && brew services start redis`
- **Linux**: `sudo apt install redis-server`

### 1. 克隆与安装

```bash
git clone <仓库地址>
cd hotel-backend
go mod tidy
```

### 2. 配置

编辑 `config/config.yaml`：

```yaml
server:
  port: "8080"

database:
  host: "localhost"      # DB_HOST
  port: "5432"           # DB_PORT
  user: "postgres"       # DB_USER
  password: "postgres"   # DB_PASSWORD
  dbname: "hotel"        # DB_NAME
  sslmode: "disable"

redis:
  addr: "localhost:6379" # REDIS_ADDR
  password: ""           # REDIS_PASSWORD
  db: 0                  # REDIS_DB

jwt:
  secret: "change-me-in-production"  # JWT_SECRET
  access_token_expiry: "2h"          # JWT_ACCESS_TOKEN_EXPIRY
  refresh_token_expiry: "168h"       # JWT_REFRESH_TOKEN_EXPIRY

checkout:
  first_threshold: "12:00"   # CHECKOUT_FIRST_THRESHOLD
  second_threshold: "18:00"  # CHECKOUT_SECOND_THRESHOLD
  scheduler_interval: 300    # CHECKOUT_SCHEDULER_INTERVAL (秒)

allocation:
  strategy: "low_floor"     # low_floor 或 high_floor

cancellation:
  cutoff_hours: 24          # CANCELLATION_CUTOFF_HOURS (小时)
  default_reject_reason: "无"
  notify_strategy: "random_one"  # staff_and_admin | all_staff | random_one
  notify_staff_ids: []

allocation:
  strategy: "low_floor"     # low_floor | high_floor

appeal:
  review_strategy: "admin_only"  # admin_only | random_one
  review_staff_ids: []

llm:
  provider: "anthropic"  # anthropic | ollama
  base_url: "https://api.anthropic.com/v1"
  api_key: "${ANTHROPIC_API_KEY}"
  model: "claude-sonnet-4-20250514"
  max_tokens: 1024
  timeout: 30s

admin:
  username: "admin"         # ADMIN_USERNAME
  password: "admin123"      # ADMIN_PASSWORD
  name: "超级管理员"          # ADMIN_NAME
```

所有配置均可通过同名环境变量覆盖，例如：

```bash
export DB_PASSWORD=mysecret
export JWT_SECRET=my-jwt-secret
export ADMIN_PASSWORD=Admin123!
go run main.go
```

### 3. 运行

```bash
go run main.go
```

服务启动后会自动执行数据库迁移，并写入默认管理员账号：

- **用户名：** `admin`（环境变量 `ADMIN_USERNAME`）
- **密码：** `admin123`（环境变量 `ADMIN_PASSWORD`）

## 接口概览

### 认证接口

| 方法 | 接口 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/auth/login` | — | 住户登录 |
| POST | `/api/v1/auth/staff-login` | — | 员工/管理员登录 |
| POST | `/api/v1/auth/admin-login` | — | 管理员登录 |
| POST | `/api/v1/auth/register` | — | 住户注册 |
| POST | `/api/v1/auth/logout` | ✅ | 登出，使 Token 失效 |
| DELETE | `/api/v1/auth/account` | ✅ | 注销账号 |

### 住户管理（全部认证用户）

| 方法 | 接口 | 说明 |
|------|------|------|
| GET | `/api/v1/guests` | 获取住户列表 |
| GET | `/api/v1/guests/:id` | 获取住户详情 |
| POST | `/api/v1/guests` | 创建住户 |
| PUT | `/api/v1/guests/:id` | 更新住户信息 |
| DELETE | `/api/v1/guests/:id` | 删除住户 |

### 员工管理（员工/管理员）

| 方法 | 接口 | 说明 |
|------|------|------|
| GET | `/api/v1/employees` | 获取员工列表 |
| GET | `/api/v1/employees/:id` | 获取员工详情 |
| POST | `/api/v1/employees` | 创建员工 |
| PUT | `/api/v1/employees/:id` | 更新员工信息 |
| DELETE | `/api/v1/employees/:id` | 删除员工 |

### 管理员管理（仅管理员）

| 方法 | 接口 | 说明 |
|------|------|------|
| GET | `/api/v1/admins` | 获取管理员列表 |
| GET | `/api/v1/admins/:id` | 获取管理员详情 |
| POST | `/api/v1/admins` | 创建管理员 |
| PUT | `/api/v1/admins/:id` | 更新管理员信息 |
| DELETE | `/api/v1/admins/:id` | 删除管理员 |

### 服务员管理（仅管理员）

| 方法 | 接口 | 说明 |
|------|------|------|
| GET | `/api/v1/waiters` | 获取服务员列表 |
| GET | `/api/v1/waiters/:id` | 获取服务员详情 |
| POST | `/api/v1/waiters` | 创建服务员 |
| PUT | `/api/v1/waiters/:id` | 更新服务员信息 |
| DELETE | `/api/v1/waiters/:id` | 删除服务员 |
| POST | `/api/v1/waiters/:id/complete-service` | 服务员完成服务 |

额外路由（无需 admin）：
| POST | `/api/v1/checkins/:id/service-request` | 认证用户 | 发起服务请求（随机派单） |

### 房间管理

| 方法 | 接口 | 角色 | 说明 |
|------|------|------|------|
| GET | `/api/v1/rooms` | 全部 | 获取房间列表 |
| GET | `/api/v1/rooms/:id` | 全部 | 获取房间详情 |
| POST | `/api/v1/rooms` | employee/admin | 创建房间 |
| PUT | `/api/v1/rooms/:id` | employee/admin | 更新房间信息 |
| DELETE | `/api/v1/rooms/:id` | employee/admin | 删除房间 |

### 订单管理

| 方法 | 接口 | 角色 | 说明 |
|------|------|------|------|
| GET | `/api/v1/orders` | 全部 | 获取订单（住户仅能看到自己的） |
| GET | `/api/v1/orders/:id` | 全部 | 获取订单详情（返回加密 code） |
| POST | `/api/v1/orders` | 全部 | 创建订单（返回加密 code 而非真实 ID） |
| POST | `/api/v1/orders/:code/cancel` | 全部 | 取消订单（住户凭 code 取消自己的 pending 订单） |
| GET | `/api/v1/orders/cancel-requests` | employee/admin | 获取待审核的取消请求 |
| POST | `/api/v1/orders/:code/reject-cancel` | employee/admin | 驳回取消申请 |
| POST | `/api/v1/orders/:code/appeal` | 全部 | 住户对驳回决定提起申诉 |
| GET | `/api/v1/appeals` | employee/admin | 查看申诉列表 |
| POST | `/api/v1/appeals/:code/review` | employee/admin | 审核申诉 |
| POST | `/api/v1/orders/:code/confirm` | employee/admin | 确认入住（员工凭 code 确认订单） |
| PUT | `/api/v1/orders/:code` | employee/admin | 更新订单（含审批取消请求） |
| DELETE | `/api/v1/orders/:code` | employee/admin | 删除订单 |

### 入住管理（员工/管理员）

| 方法 | 接口 | 说明 |
|------|------|------|
| GET | `/api/v1/checkins` | 获取入住记录 |
| GET | `/api/v1/checkins/:id` | 获取入住详情 |
| POST | `/api/v1/checkins` | 办理入住 |
| PUT | `/api/v1/checkins/:id/checkout` | 办理退房 |
| DELETE | `/api/v1/checkins/:id` | 删除入住记录 |

### 通知

| 方法 | 接口 | 说明 |
|------|------|------|
| GET | `/api/v1/notifications` | 获取通知列表 |
| GET | `/api/v1/notifications/unread` | 获取未读通知数量 |
| PUT | `/api/v1/notifications/:code/read` | 标记通知已读 |

### 其他

| 方法 | 接口 | 角色 | 说明 |
|------|------|------|------|
| GET | `/api/v1/audit-logs` | employee/admin | 查看审计日志 |
| GET | `/api/v1/ws` | 全部 | WebSocket 连接（实时通知） |

### 房间自动分配

创建订单时可选择指定房间或让系统自动分配，默认采用**低楼层优先**策略：

```yaml
allocation:
  strategy: "low_floor"     # low_floor（默认）或 high_floor
```

### 订单取消与申诉

住户可取消自己的 `pending` 订单，需填写取消理由：

- **距入住 > `cutoff_hours`**：自动取消，房间释放为 `vacant`
- **距入住 ≤ `cutoff_hours`**：进入 `cancel_requested` 待审核，根据策略通知工作人员
- **入住日期已过**：拒绝取消
- 员工通过 `PUT /orders/:code` 或 `POST /orders/:code/reject-cancel` 审批
  - 通过（`cancelled`）：房间释放，客户收到 `cancel_approved` 通知
  - 驳回（`pending`）：客户收到 `cancel_rejected` 通知，并可提起申诉
- 申诉流程：客户 `POST /orders/:code/appeal` → 员工 `POST /appeals/:code/review`
  - 通过（`approved`）：订单取消，房间释放
  - 驳回（`rejected`）：订单保持 `pending`
- 申诉审核策略可配置（`admin_only` / `random_one`）
- 订单、通知和申诉 ID 使用 Hashids 加密为混淆字符串，不对外暴露真实 ID

```yaml
cancellation:
  cutoff_hours: 24          # CANCELLATION_CUTOFF_HOURS
  default_reject_reason: "无"
  notify_strategy: "random_one" # staff_and_admin | all_staff | random_one
  notify_staff_ids: []

appeal:
  review_strategy: "admin_only"  # admin_only | random_one
  review_staff_ids: []
```

## AI 智能管家

AI 智能管家为住客提供自然语言交互入口，通过单一  端点完成酒店信息问答、服务请求、订单查询和申诉处理。

### 对话流程





### 可调用工具

| 工具 | 触发场景 | 功能 |
|------|---------|------|
|  | 叫服务员到房间 | 随机派单 + 双方通知 |
|  | 查订单 | 查询当前用户订单列表 |
|  | 查申诉进度 | 查询指定订单的申诉状态 |
|  | 发起申诉 | 对驳回决定提起申诉 |

### LLM Provider 配置

支持两种 LLM 服务商，通过  切换：



API Key 优先从系统环境变量读取，其次从项目根目录的  文件读取。

### 注意事项

- 对话上下文保存在 Redis 中（24 小时 TTL，保留最近 20 条消息）
- 每个用户同一时间只有一个活跃对话
- LLM 回答酒店设施信息（退房时间、健身房等）基于模型知识，非实时查询
- 云端模型（Anthropic/DeepSeek）支持工具调用；小型本地 Ollama 模型可能仅支持文本对话

## 认证方式

API 采用 JWT Bearer Token 认证，根据角色使用不同登录端点：

```bash
# 住户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"guest_user","password":"guest123"}'

# 员工/管理员登录
curl -X POST http://localhost:8080/api/v1/auth/staff-login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 管理员登录
curl -X POST http://localhost:8080/api/v1/auth/admin-login \
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

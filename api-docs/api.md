# 酒店管理系统 API 文档

**Base URL**: `http://localhost:8080/api/v1`

## 通用说明

### 响应格式

所有接口返回统一 JSON 结构：

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 0 成功，非 0 失败 |
| `message` | string | 状态描述 |
| `data` | object | 业务数据 |

### 分页响应

列表接口返回分页数据：

```json
{
  "code": 0,
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

### 认证方式

请求头携带 JWT Token：

```
Authorization: Bearer <access_token>
```

### 错误码

| code | 说明 |
|------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 冲突（如用户名重复） |
| 500 | 服务器内部错误 |
| 1001 | 用户名或密码错误 |
| 1002 | Token 已过期 |
| 1003 | Token 无效 |
| 2001 | 用户不存在 |
| 2002 | 用户名已存在 |
| 3001 | 房间不存在 |
| 3002 | 房间已被占用 |
| 3003 | 无可用房间 |
| 4001 | 订单不存在 |
| 5001 | 入住记录不存在 |
| 5002 | 已签离 |

---

## 一、认证模块

### 1.1 住户登录

登录成功后返回 JWT Token，仅限 `guest` 角色。

```
POST /auth/login
```

**请求体**：

```json
{
  "username": "guest001",
  "password": "123456"
}
```

**响应**：

```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200,
    "user_id": 1,
    "role": "guest",
    "username": "guest001",
    "name": "张三"
  }
}
```

### 1.2 员工/管理员登录

登录成功后返回 JWT Token，仅限 `employee` 或 `admin` 角色。

```
POST /auth/staff-login
```

**请求体**：

```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应**（同住户登录，role 为 `employee` 或 `admin`）：

```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "user_id": 1,
    "role": "admin",
    "username": "admin",
    "name": "超级管理员"
  }
}
```

### 1.3 管理员登录

登录成功后返回 JWT Token，仅限 `admin` 角色。

```
POST /auth/admin-login
```

**请求体**：

```json
{
  "username": "admin",
  "password": "admin123"
}
```

### 1.4 住户注册

注册新住户账号，角色固定为 `guest`。

```
POST /auth/register
```

**请求体**：

```json
{
  "username": "newuser",
  "password": "123456",
  "name": "新用户",
  "phone": "13800138000",
  "email": "user@example.com"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `username` | ✅ | 3~64 字符 |
| `password` | ✅ | 最少 6 位 |
| `name` | — | 姓名 |
| `phone` | — | 手机号 |
| `email` | — | 邮箱 |

### 1.5 登出

```
POST /auth/logout
Authorization: Bearer <token>
```

**请求体**：无

### 1.6 注销账号

删除当前登录的账号（仅限住户自己操作）。

```
DELETE /auth/account
Authorization: Bearer <token>
```

---

## 二、住户管理（全部认证用户）

所有 `/guests` 接口均需认证，对全部角色开放（guest / employee / admin）。

### 2.1 获取住户列表

```
GET /guests?page=1&page_size=20
Authorization: Bearer <token>
```

### 2.2 获取单个住户

```
GET /guests/:id
Authorization: Bearer <token>
```

**响应**：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "user_id": 1,
    "user": {
      "id": 1,
      "username": "guest001",
      "role": "guest",
      "created_at": "2026-05-01T00:00:00Z"
    },
    "name": "张三",
    "phone": "13800138000",
    "email": "zhangsan@example.com",
    "created_at": "2026-05-01T00:00:00Z"
  }
}
```

### 2.3 创建住户

创建时会同时生成 `users` 认证记录和 `guests` 业务记录。

```
POST /guests
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "username": "guest_new",
  "password": "123456",
  "name": "新住户",
  "phone": "13800138000",
  "email": "guest@example.com"
}
```

### 2.4 更新住户

仅更新住户的业务信息（姓名、电话、邮箱），不修改认证信息。

```
PUT /guests/:id
Authorization: Bearer <token>
```

**请求体**（至少提供一个字段）：

```json
{
  "name": "更新后姓名",
  "phone": "13900139000",
  "email": "new@example.com"
}
```

### 2.5 删除住户

删除住户的业务记录和对应的认证记录。

```
DELETE /guests/:id
Authorization: Bearer <token>
```

---

## 三、员工管理（员工/管理员）

所有 `/employees` 接口仅限 `employee` 或 `admin` 角色访问。

### 3.1 获取员工列表

```
GET /employees?page=1&page_size=20
Authorization: Bearer <token>
```

### 3.2 获取单个员工

```
GET /employees/:id
Authorization: Bearer <token>
```

### 3.3 创建员工

```
POST /employees
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "username": "emp_new",
  "password": "123456",
  "name": "新员工",
  "phone": "13800138000",
  "email": "emp@example.com"
}
```

### 3.4 更新员工

```
PUT /employees/:id
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "name": "更新后姓名",
  "phone": "13900139000",
  "email": "new@example.com"
}
```

### 3.5 删除员工

```
DELETE /employees/:id
Authorization: Bearer <token>
```

---

## 四、管理员管理（仅管理员）

所有 `/admins` 接口仅限 `admin` 角色访问。

### 4.1 获取管理员列表

```
GET /admins?page=1&page_size=20
Authorization: Bearer <token>
```

### 4.2 获取单个管理员

```
GET /admins/:id
Authorization: Bearer <token>
```

### 4.3 创建管理员

```
POST /admins
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "username": "admin_new",
  "password": "123456",
  "name": "新管理员",
  "phone": "13800138000",
  "email": "admin@example.com"
}
```

### 4.4 更新管理员

```
PUT /admins/:id
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "name": "更新后姓名",
  "phone": "13900139000",
  "email": "new@example.com"
}
```

### 4.5 删除管理员

```
DELETE /admins/:id
Authorization: Bearer <token>
```

---

## 五、房间管理

### 5.1 获取房间列表

```
GET /rooms?page=1&page_size=20
Authorization: Bearer <token>
```

### 5.2 获取单个房间

```
GET /rooms/:id
Authorization: Bearer <token>
```

### 5.3 创建房间（员工/管理员）

```
POST /rooms
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "room_number": "101",
  "type": "standard",
  "capacity": 2,
  "floor": 1,
  "price_per_night": 300.0,
  "status": "vacant",
  "description": "标准间"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `room_number` | ✅ | 房间号 |
| `type` | — | 房型：standard / deluxe / suite |
| `capacity` | — | 入住人数上限 |
| `floor` | — | 楼层 |
| `price_per_night` | — | 每晚价格 |
| `status` | — | 默认 vacant |
| `description` | — | 房间描述 |

### 5.4 更新房间（员工/管理员）

```
PUT /rooms/:id
Authorization: Bearer <token>
```

**请求体**（所有字段可选）：

```json
{
  "room_number": "102",
  "type": "deluxe",
  "capacity": 3,
  "floor": 2,
  "price_per_night": 500.0,
  "status": "vacant",
  "description": "豪华间"
}
```

### 5.5 删除房间（员工/管理员）

```
DELETE /rooms/:id
Authorization: Bearer <token>
```

---

## 六、订单管理

### 6.1 获取订单列表

```
GET /orders?page=1&page_size=20
Authorization: Bearer <token>
```

**角色差异**：
- `guest`：仅返回当前用户自己的订单
- `employee` / `admin`：返回全部订单

### 6.2 获取单个订单

```
GET /orders/:id
Authorization: Bearer <token>
```

**角色差异**：`guest` 只能查看自己的订单，无权访问他人订单（返回 403）。

### 6.3 创建订单

```
POST /orders
Authorization: Bearer <token>
```

**自动分配房间**（推荐）：

```json
{
  "guest_count": 2,
  "room_type_preference": "standard",
  "check_in_date": "2026-06-01",
  "check_out_date": "2026-06-03",
  "total_price": 600.0
}
```

**指定房间**：

```json
{
  "room_id": 1,
  "check_in_date": "2026-06-01",
  "check_out_date": "2026-06-03",
  "total_price": 600.0
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `room_id` | — | 指定房间 ID，不传则自动分配 |
| `guest_count` | — | 入住人数（自动分配时使用） |
| `room_type_preference` | — | 房型偏好（自动分配时使用） |
| `check_in_date` | ✅ | 入住日期 `yyyy-MM-dd` |
| `check_out_date` | ✅ | 退房日期 `yyyy-MM-dd` |
| `total_price` | — | 总价 |

### 6.4 更新订单（员工/管理员）

```
PUT /orders/:id
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "check_in_date": "2026-06-02",
  "check_out_date": "2026-06-04",
  "total_price": 800.0,
  "status": "confirmed"
}
```

| 字段 | 说明 |
|------|------|
| `status` | pending / confirmed / cancelled |

### 6.5 删除订单（员工/管理员）

```
DELETE /orders/:id
Authorization: Bearer <token>
```

---

## 七、入住管理（员工/管理员）

所有 `/checkins` 接口仅限 `employee` 或 `admin` 角色访问。

### 7.1 获取入住记录列表

```
GET /checkins?page=1&page_size=20
Authorization: Bearer <token>
```

### 7.2 获取单个入住记录

```
GET /checkins/:id
Authorization: Bearer <token>
```

### 7.3 办理入住

```
POST /checkins
Authorization: Bearer <token>
```

**场景一：根据已有订单办理入住**

```json
{
  "order_id": 1,
  "user_id": 1,
  "expected_checkout_time": "2026-06-03T12:00:00+08:00"
}
```

**场景二：直接入住（无订单、无指定房间，系统自动分配）**

```json
{
  "user_id": 1,
  "expected_checkout_time": "2026-06-03T12:00:00+08:00"
}
```

**场景三：直接入住（指定房间）**

```json
{
  "user_id": 1,
  "room_id": 1,
  "expected_checkout_time": "2026-06-03T12:00:00+08:00"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `order_id` | — | 关联订单 ID |
| `user_id` | ✅ | 住户用户 ID |
| `room_id` | — | 指定房间 ID |
| `expected_checkout_time` | ✅ | 预计退房时间（RFC3339 格式） |

### 7.4 办理退房

将入住状态更新为已完成，同时将房间状态恢复为空闲。

```
PUT /checkins/:id/checkout
Authorization: Bearer <token>
```

**响应**：

```json
{
  "code": 0,
  "data": null
}
```

### 7.5 删除入住记录

```
DELETE /checkins/:id
Authorization: Bearer <token>
```

---

## 八、通知

### 8.1 获取通知列表

```
GET /notifications?page=1&page_size=20
Authorization: Bearer <token>
```

仅返回当前用户自己的通知。

### 8.2 获取未读通知数量

```
GET /notifications/unread
Authorization: Bearer <token>
```

**响应**：

```json
{
  "code": 0,
  "data": {
    "count": 3
  }
}
```

### 8.3 标记通知已读

```
PUT /notifications/:id/read
Authorization: Bearer <token>
```

---

## 九、审计日志（员工/管理员）

### 9.1 获取审计日志

```
GET /audit-logs?page=1&page_size=20
Authorization: Bearer <token>
```

仅限 `employee` 或 `admin` 角色访问。

---

## 十、WebSocket

### 10.1 建立连接

用于接收实时通知推送。Token 通过 URL 查询参数传递（浏览器 WebSocket API 不支持自定义请求头）：

```
GET /ws?token=<access_token>
```

### 10.2 消息格式

服务端在有新通知时自动推送 JSON 消息，格式与通知模型的响应结构一致：

```json
{
  "id": 1,
  "user_id": 1,
  "user": {
    "id": 1,
    "username": "guest001",
    "role": "guest",
    "created_at": "2026-05-01T00:00:00Z"
  },
  "type": "overdue_alert",
  "title": "超时签离告警",
  "content": "住户 张三 在 101 号房超时未签离，应签离时间 2026-05-17 12:00，请跟进。",
  "is_read": false,
  "created_at": "2026-05-17T12:05:00+08:00"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | uint | 通知 ID |
| `user_id` | uint | 接收者用户 ID |
| `user` | object | 接收者基本信息 |
| `type` | string | 通知类型：`overdue_alert` |
| `title` | string | 通知标题 |
| `content` | string | 通知正文 |
| `is_read` | bool | 是否已读 |
| `created_at` | string | 创建时间 |

### 10.3 心跳保活

服务端每 30 秒发送 Ping 帧，客户端需在 60 秒内回复 Pong 帧，否则连接断开。客户端无需发送消息，只负责接收推送。

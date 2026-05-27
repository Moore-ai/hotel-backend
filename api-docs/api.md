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
| 3001 | 房间不存在 |
| 3002 | 房间已被占用 |
| 3003 | 无可用房间 |
| 3004 | 暂无空闲服务员 |
| 4001 | 订单不存在 |
| 4002 | 订单状态不是 pending，无法取消 |
| 4003 | 申诉不存在 |
| 4004 | 申诉状态不是 pending |
| 4005 | 该订单已有进行中的申诉 |
| 5001 | 入住记录不存在 |
| 5002 | 已签离 |
| 6001 | AI 管家暂时不可用 |
| 6002 | AI 响应解析失败 |

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

登录成功后返回 JWT Token，仅限 `employee`、`admin` 或 `waiter` 角色。

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

## 五、服务员管理（仅管理员）

所有 `/waiters` 接口仅限 `admin` 角色访问。

### 5.1 获取服务员列表

```
GET /waiters?page=1&page_size=20
Authorization: Bearer <token>
```

**响应**：

```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "user_id": 10,
        "user": {
          "id": 10,
          "username": "waiter01",
          "role": "waiter",
          "created_at": "2026-05-01T00:00:00Z"
        },
        "name": "服务员A",
        "phone": "13800138000",
        "email": "waiter@example.com",
        "serving_room_id": null,
        "created_at": "2026-05-01T00:00:00Z",
        "updated_at": "2026-05-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

### 5.2 获取单个服务员

```
GET /waiters/:id
Authorization: Bearer <token>
```

### 5.3 创建服务员

```
POST /waiters
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "username": "waiter_new",
  "password": "123456",
  "name": "新服务员",
  "phone": "13800138000",
  "email": "waiter@example.com"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `username` | ✅ | 3~64 字符 |
| `password` | ✅ | 最少 6 位 |
| `name` | — | 姓名 |
| `phone` | — | 手机号 |
| `email` | — | 邮箱 |

### 5.4 更新服务员

```
PUT /waiters/:id
Authorization: Bearer <token>
```

**请求体**（所有字段可选）：

```json
{
  "name": "更新后姓名",
  "phone": "13900139000",
  "email": "new@example.com"
}
```

### 5.5 删除服务员

```
DELETE /waiters/:id
Authorization: Bearer <token>
```

### 5.6 完成服务

服务员将自身状态恢复为空闲（`serving_room_id` 置空）。

- `waiter` 角色只能完成自己的服务
- `admin` 可完成任意服务员的服务

```
POST /waiters/:id/complete-service
Authorization: Bearer <token>
```

### 5.7 发起服务请求

已入住客户可通过入住记录 ID 发起服务请求，系统随机分配一名空闲服务员。

```
POST /checkins/:id/service-request
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "content": "送水",
  "note": "请多带几瓶"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `content` | ✅ | 服务内容 |
| `note` | — | 备注 |

**成功响应**（分配成功）：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "user_id": 10,
    "name": "服务员A",
    "phone": "13800138000",
    "serving_room_id": 5,
    "created_at": "2026-05-01T00:00:00Z",
    "updated_at": "2026-05-01T00:00:00Z"
  }
}
```

**错误码**：

| code | 说明 |
|------|------|
| 400 | 入住记录不存在或状态非 active |
| 403 | 非本人入住记录（guest 角色） |
| 3004 | 暂无空闲服务员 |

---

## 六、房间管理

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

## 七、订单管理

### 11.1 获取订单列表

```
GET /orders?page=1&page_size=20
Authorization: Bearer <token>
```

**角色差异**：
- `guest`：仅返回当前用户自己的订单
- `employee` / `admin`：返回全部订单

### 9.2 获取单个订单

```
GET /orders/:id
Authorization: Bearer <token>
```

**角色差异**：`guest` 只能查看自己的订单，无权访问他人订单（返回 403）。

### 9.3 创建订单

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

### 9.4 更新订单（员工/管理员）

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

### 9.5 取消订单

住户取消自己的 `pending` 状态订单。根据距入住时间的长度决定自动取消或提交审核。

```
POST /orders/:id/cancel
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "reason": "行程变更，无法入住"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `reason` | ✅ | 取消理由，最少 2 个字符 |

**自动取消响应**（距入住 > `cutoff_hours`）：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "status": "cancelled",
    "auto_cancelled": true,
    "cancel_reason": "行程变更，无法入住"
  }
}
```

**提交审核响应**（距入住 ≤ `cutoff_hours`）：

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "status": "cancel_requested",
    "auto_cancelled": false,
    "cancel_reason": "行程变更，无法入住"
  }
}
```

**错误码**：

| code | 说明 |
|------|------|
| 400 | 入住日期已过，无法取消 |
| 403 | 非本人订单 |
| 4002 | 订单状态不是 `pending`，无法取消 |

### 9.6 获取待审取消请求（员工/管理员）

```
GET /orders/cancel-requests?page=1&page_size=20
Authorization: Bearer <token>
```

返回所有 `cancel_requested` 状态的订单。

### 9.7 更新订单（用于审批取消请求）

当订单状态为 `cancel_requested` 时，员工可通过此接口审批：

```
PUT /orders/:id
Authorization: Bearer <token>
```

**通过取消申请**：

```json
{
  "status": "cancelled"
}
```

房间释放为 `vacant`，客户收到 `cancel_approved` 通知。

**驳回取消申请**：

```json
{
  "status": "pending"
}
```

房间保持 `reserved`，客户收到 `cancel_rejected` 通知。

| 字段 | 说明 |
|------|------|
| `status` | `cancelled` 通过 / `pending` 驳回 |

### 9.8 删除订单（员工/管理员）

```
DELETE /orders/:id
Authorization: Bearer <token>
```

---

## 八、入住管理（员工/管理员）

所有 `/checkins` 接口仅限 `employee` 或 `admin` 角色访问。

### 11.1 获取入住记录列表

```
GET /checkins?page=1&page_size=20
Authorization: Bearer <token>
```

### 9.2 获取单个入住记录

```
GET /checkins/:id
Authorization: Bearer <token>
```

### 9.3 办理入住

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

### 9.4 办理退房

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

### 9.5 删除入住记录

```
DELETE /checkins/:id
Authorization: Bearer <token>
```

---

## 九、通知

通知的 `id` 字段使用 Hashids 加密为字符串格式，所有 API 路径参数中的 `:code` 均为此加密 ID。

### 11.1 获取通知列表

```
GET /notifications?page=1&page_size=20
Authorization: Bearer <token>
```

仅返回当前用户自己的通知。

### 9.2 获取未读通知数量

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

### 9.3 标记通知已读

```
PUT /notifications/:code/read
Authorization: Bearer <token>
```

---

## 十、审计日志（员工/管理员）

### 11.1 获取审计日志

```
GET /audit-logs?page=1&page_size=20
Authorization: Bearer <token>
```

仅限 `employee` 或 `admin` 角色访问。

---

## 十一、申诉

驳回客户取消申请后，客户可选择申诉。工作人员审核申诉（通过/驳回），结果通知客户。

申诉的审核策略通过配置文件 `appeal.review_strategy` 指定：
- `admin_only`（默认）：仅通知超级管理员
- `random_one`：随机分配给指定工作人员（`appeal.review_staff_ids`），未配置时覆盖全体员工

### 11.1 提交申诉（住户）

仅限 `guest` 角色，只能对自己的订单提交申诉。

```
POST /orders/:code/appeal
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "reason": "取消时间仍在合理范围内，请求人工审核"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `reason` | ✅ | 申诉理由，最少 2 个字符 |

**响应**：

```json
{
  "code": 0,
  "data": {
    "id": "aB3xR7kQ",
    "order_id": 1,
    "user_id": 1,
    "reason": "取消时间仍在合理范围内，请求人工审核",
    "status": "pending",
    "reviewer_id": null,
    "review_note": null,
    "created_at": "2026-05-20T00:00:00Z"
  }
}
```

**错误码**：

| code | 说明 |
|------|------|
| 400 | 订单不是 pending 状态（ErrOrderNotPending） |
| 403 | 非本人订单 |
| 404 | 订单不存在 |
| 4005 | 该订单已有进行中的申诉（ErrAppealExists） |

### 11.2 查看我的申诉（住户）

所有登录用户均可访问，返回当前用户自己的申诉记录。

```
GET /appeals/my?status=pending&page=1&page_size=20
Authorization: Bearer <token>
```

| 参数 | 说明 |
|------|------|
| `status` | 筛选状态：pending / approved / rejected，为空返回全部 |

### 11.3 查看全部申诉（员工/管理员）

仅限 `employee` 或 `admin` 角色，返回系统中所有申诉记录。

```
GET /appeals?status=pending&page=1&page_size=20
Authorization: Bearer <token>
```

| 参数 | 说明 |
|------|------|
| `status` | 筛选状态：pending / approved / rejected，为空返回全部 |

### 11.3 审核申诉

```
POST /appeals/:code/review
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "action": "approved",
  "review_note": "经核实，同意取消"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `action` | ✅ | `approved`（通过，订单取消）/ `rejected`（驳回） |
| `review_note` | — | 审核备注 |

**通过响应**（申诉状态变为 approved，订单取消，房间释放）：

```json
{
  "code": 0,
  "data": {
    "id": "aB3xR7kQ",
    "status": "approved",
    "reviewer_id": 1,
    "review_note": "经核实，同意取消"
  }
}
```

**驳回响应**（申诉状态变为 rejected，订单保持 pending）：

```json
{
  "code": 0,
  "data": {
    "id": "aB3xR7kQ",
    "status": "rejected",
    "reviewer_id": 1,
    "review_note": "取消政策不允许"
  }
}
```

**错误码**：

| code | 说明 |
|------|------|
| 400 | action 无效，或申诉不在 pending 状态（ErrAppealNotPending） |
| 404 | 申诉不存在（ErrAppealNotFound） |

---

## 十二、WebSocket

### 12.1 建立连接

用于接收实时通知推送。Token 通过 URL 查询参数传递（浏览器 WebSocket API 不支持自定义请求头）：

```
GET /ws?token=<access_token>
```

### 12.2 消息格式

服务端在有新通知时自动推送 JSON 消息，格式与通知模型的响应结构一致：

```json
{
  "id": "aB3xR7kQ",
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
| `id` | string | 通知 ID（Hashids 加密） |
| `user_id` | uint | 接收者用户 ID |
| `user` | object | 接收者基本信息 |
| `type` | string | 通知类型：`overdue_alert` / `waiter_assigned` / `waiter_task` / `waiter_unavailable` / `cancel_approved` / `cancel_rejected` / `appeal_request` / `appeal_approved` / `appeal_rejected` |
| `title` | string | 通知标题 |
| `content` | string | 通知正文 |
| `is_read` | bool | 是否已读 |
| `created_at` | string | 创建时间 |

### 12.3 心跳保活

服务端每 30 秒发送 Ping 帧，客户端需在 60 秒内回复 Pong 帧，否则连接断开。客户端无需发送消息，只负责接收推送。---


## 十三、AI 智能管家

AI 智能管家为住客提供自然语言交互入口，通过 LLM 理解住客意图并自动调用后端服务。

支持三种 LLM 服务商，通过 `config.yaml` 的 `llm.provider` 配置切换：

- **anthropic**（默认）：Anthropic Messages API，兼容 DeepSeek 等第三方服务
- **openai**：OpenAI Chat Completions API
- **ollama**：本地 Ollama 模型

### 13.1 发送对话消息

```
POST /chat
Authorization: Bearer <token>
```

**请求体**：

```json
{
  "message": "帮我查一下我的订单",
  "conversation_id": "optional-uuid"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `message` | ✅ | 用户消息 |
| `conversation_id` | — | 对话 ID。新对话不传此字段，服务端自动创建并返回；续接对话时传入上次返回的 ID |

**响应**：

```json
{
  "code": 0,
  "data": {
    "reply": "您好，查询到您当前没有订单记录。",
    "conversation_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "action": "get_my_orders"
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `reply` | string | AI 回复文本 |
| `conversation_id` | string | 当前对话 ID |
| `action` | string | 触发的工具名（如 `get_my_orders`），未触发时为空字符串 |

**可调用工具**：

| 工具 | 触发场景 | 功能 |
|------|---------|------|
| `dispatch_waiter` | 叫服务员到房间 | 随机派单 + 双方通知 |
| `get_my_orders` | 查订单 | 查询当前用户订单列表 |
| `get_appeal_status` | 查申诉进度 | 查询指定订单的申诉状态 |
| `create_appeal` | 发起申诉 | 对驳回决定提起申诉 |

**错误码**：

| code | 说明 |
|------|------|
| 400 | 请求参数错误（message 为空） |
| 401 | 未认证 |
| 429 | 请求过于频繁（速率限制） |
| 6001 | AI 管家暂时不可用（LLM API 不可达） |
| 6002 | AI 响应解析失败 |

### 13.2 对话生命周期

1. 用户发送 `POST /chat`（不带 `conversation_id`）
2. 服务端创建新对话，返回 `conversation_id`
3. 用户续接对话时传入 `conversation_id`，服务端从 Redis 恢复历史（保留最近 20 条，24 小时 TTL）
4. 每个用户同时只有一个活跃对话

### 13.3 速率限制

AI 对话接口支持速率限制，防止单个用户过度调用。支持三种算法：

| 算法 | 说明 |
|------|------|
| `fixed_window` | 固定窗口，每分钟重置 |
| `token_bucket` | 令牌桶，支持突发流量 |
| `sliding_window` | 滑动窗口，精确统计 |

白名单中的用户 ID 不受限流限制（默认包含超级管理员）。

被限流时返回 HTTP 429，附带 `Retry-After` 响应头指示等待秒数。

### 13.4 配置

```yaml
llm:
  provider: "anthropic"  # anthropic | openai | ollama
  base_url: "https://api.anthropic.com/v1"
  api_key: "${LLM_API_KEY}"
  model: "claude-sonnet-4-20250514"
  max_tokens: 1024
  timeout: 30s
  system_prompt: "..."
  max_history: 20
  rate_limit:
    enabled: true
    max_requests_per_minute: 10
    algorithm: fixed_window  # fixed_window | token_bucket | sliding_window
    whitelist:
      - 1                   # 跳过限流的用户 ID
```

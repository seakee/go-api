# API文档

**语言版本**: [English](API-Documentation.md) | [中文](API-Documentation-zh.md)

---

## 认证

所有受保护的端点都需要JWT认证。在Authorization头中包含令牌：

```
Authorization: Bearer <your-jwt-token>
```

## 基础URL

```
本地开发: http://localhost:8080
生产环境: https://your-domain.com
```

## API端点

### 认证端点

#### 获取JWT令牌

**POST** `/go-api/external/service/auth/token`

获取API访问的JWT令牌。

**请求体 (form-data):**
```
app_id: string (必需) - 应用ID
app_secret: string (必需) - 应用密钥
```

**响应:**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 604800
  }
}
```

**cURL示例:**
```bash
curl -X POST http://localhost:8080/go-api/external/service/auth/token \
  -d "app_id=your_app_id" \
  -d "app_secret=your_app_secret"
```

#### 创建应用

**POST** `/go-api/external/service/auth/app`

创建新应用（需要认证）。

**请求头:**
```
Authorization: Bearer <your-jwt-token>
Content-Type: application/json
```

**请求体:**
```json
{
  "app_name": "我的应用",
  "description": "应用描述",
  "redirect_uri": "https://example.com/callback"
}
```

**响应:**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "app_id": "generated_app_id",
    "app_secret": "generated_app_secret"
  }
}
```

#### 系统管理端点（内部）

系统管理接口挂载在 `/go-api/internal/admin/system`。

- 认证中间件：`CheckAdminAuth`
- 令牌来源：`Authorization: Bearer <admin-token>` 或 Cookie `admin-token`
- 审计中间件：`/go-api/internal/admin/*` 路由统一经过 `SaveOperationRecord` 并记录操作日志

##### 路由清单

| 模块 | 方法 | 路径 | 说明 |
|------|------|------|------|
| System | GET | `/go-api/internal/admin/system/ping` | 系统健康检查 |
| Menu | GET | `/go-api/internal/admin/system/menu/list` | 菜单树列表 |
| Menu | GET | `/go-api/internal/admin/system/menu` | 菜单详情 |
| Menu | POST | `/go-api/internal/admin/system/menu` | 创建菜单 |
| Menu | PUT | `/go-api/internal/admin/system/menu` | 更新菜单 |
| Menu | DELETE | `/go-api/internal/admin/system/menu` | 删除菜单 |
| Permission | GET | `/go-api/internal/admin/system/permission/available` | 按 HTTP 方法分组的待绑定后台路由 |
| Permission | GET | `/go-api/internal/admin/system/permission/list` | 按分组权限列表 |
| Permission | GET | `/go-api/internal/admin/system/permission/paginate` | 权限分页 |
| Permission | GET | `/go-api/internal/admin/system/permission` | 权限详情 |
| Permission | POST | `/go-api/internal/admin/system/permission` | 创建权限 |
| Permission | PUT | `/go-api/internal/admin/system/permission` | 更新权限 |
| Permission | DELETE | `/go-api/internal/admin/system/permission` | 删除权限 |
| Role | GET | `/go-api/internal/admin/system/role/list` | 角色列表 |
| Role | GET | `/go-api/internal/admin/system/role/paginate` | 角色分页 |
| Role | GET | `/go-api/internal/admin/system/role` | 角色详情 |
| Role | POST | `/go-api/internal/admin/system/role` | 创建角色 |
| Role | PUT | `/go-api/internal/admin/system/role` | 更新角色 |
| Role | DELETE | `/go-api/internal/admin/system/role` | 删除角色 |
| Role | GET | `/go-api/internal/admin/system/role/permission` | 角色权限 ID 列表 |
| Role | GET | `/go-api/internal/admin/system/role/permission/menu-tree` | 角色菜单权限树 |
| Role | PUT | `/go-api/internal/admin/system/role/permission` | 更新角色权限 |
| User | GET | `/go-api/internal/admin/system/user/paginate` | 用户分页 |
| User | GET | `/go-api/internal/admin/system/user` | 用户详情 |
| User | POST | `/go-api/internal/admin/system/user` | 创建用户 |
| User | PUT | `/go-api/internal/admin/system/user` | 更新用户 |
| User | DELETE | `/go-api/internal/admin/system/user` | 删除用户 |
| User | GET | `/go-api/internal/admin/system/user/role` | 用户角色 ID 列表 |
| User | PUT | `/go-api/internal/admin/system/user/role` | 更新用户角色（会保留 `base` 角色） |
| User | PUT | `/go-api/internal/admin/system/user/password/reset` | 管理员重置密码 |
| User | PUT | `/go-api/internal/admin/system/user/tfa/disable` | 管理员关闭 TFA |
| User | GET | `/go-api/internal/admin/system/user/passkeys` | 查询用户 Passkey |
| User | DELETE | `/go-api/internal/admin/system/user/passkey` | 删除单个用户 Passkey |
| User | DELETE | `/go-api/internal/admin/system/user/passkeys` | 删除用户全部 Passkey |
| Record | GET | `/go-api/internal/admin/system/record/paginate` | 操作记录分页 |
| Record | GET | `/go-api/internal/admin/system/record/detail` | 操作记录详情 |

##### 核心数据结构（当前实现字段）

`User` 分页/详情项：
```json
{
  "id": 1,
  "email": "admin@example.com",
  "phone": "+8613800000000",
  "totp_enabled": false,
  "passkey_count": 2,
  "user_name": "管理员",
  "status": 1,
  "avatar": "",
  "created_at": "2026-02-06T10:00:00+08:00"
}
```

说明：第三方账号绑定已迁移到独立的 `sys_user_identity` 表中，Passkey 凭证存放在 `sys_user_passkey` 表中，`passkey_count` 表示当前用户已注册的 Passkey 数量。

`Role` 列表项（`/role/list`）：
```json
{
  "id": 1,
  "name": "base",
  "description": "基础角色"
}
```

`Permission` 项（`/permission/paginate`、`/permission`）：
```json
{
  "ID": 1,
  "CreatedAt": "2026-02-06T10:00:00+08:00",
  "UpdatedAt": "2026-02-06T10:00:00+08:00",
  "DeletedAt": null,
  "name": "user:list",
  "type": "api",
  "method": "GET",
  "path": "/go-api/internal/admin/system/user/paginate",
  "description": "查询用户列表",
  "group": "user"
}
```

`Menu` 项（`/menu/list`、`/menu`）：
```json
{
  "ID": 1,
  "CreatedAt": "2026-02-06T10:00:00+08:00",
  "UpdatedAt": "2026-02-06T10:00:00+08:00",
  "DeletedAt": null,
  "name": "系统管理",
  "path": "/system",
  "permission_id": 10,
  "parent_id": 0,
  "icon": "setting",
  "sort": 1,
  "children": []
}
```

`RoleMenuPermissionNode` 项（`/role/permission/menu-tree`）：
```json
{
  "id": 1,
  "name": "系统管理",
  "path": "/system",
  "permission_id": 0,
  "parent_id": 0,
  "icon": "setting",
  "sort": 1,
  "checked": false,
  "children": [
    {
      "id": 2,
      "name": "用户管理",
      "path": "/system/user",
      "permission_id": 101,
      "parent_id": 1,
      "icon": "user",
      "sort": 1,
      "checked": true,
      "children": []
    }
  ]
}
```

`OperationRecord` 项（`/record/paginate`）：
```json
{
  "ID": 1024,
  "CreatedAt": "2026-02-06T10:00:00+08:00",
  "UpdatedAt": "2026-02-06T10:00:01+08:00",
  "DeletedAt": null,
  "ip": "127.0.0.1",
  "method": "POST",
  "path": "/go-api/internal/admin/system/user",
  "status": 200,
  "latency": 0.031,
  "agent": "Mozilla/5.0",
  "error_message": "",
  "user_id": 1,
  "user_name": "admin",
  "params": "{\"user_name\":\"demo\"}",
  "resp": "{\"code\":0,\"msg\":\"OK\"}",
  "trace_id": "trace-xxx"
}
```

`OperationRecordDetail`（`/record/detail`）：
```json
{
  "id": 1024,
  "method": "POST",
  "path": "/go-api/internal/admin/system/user",
  "ip": "127.0.0.1",
  "status": 0,
  "user_id": 1,
  "user_name": "admin",
  "trace_id": "trace-xxx",
  "created_at": "2026-02-06T10:00:00+08:00",
  "latency": 0.031,
  "agent": "Mozilla/5.0",
  "error_message": "",
  "params": {
    "user_name": "demo"
  },
  "resp": {
    "code": 0,
    "msg": "OK"
  }
}
```

##### 分页说明（系统管理）

- `user/role/permission` 使用 `page` + `page_size`（默认 `1` + `10`，最大 `100`），返回 `{ "list": [...], "total": n }`
- `record` 使用 `page` + `size`（默认 `1` + `10`，最大 `100`），返回 `{ "items": [...], "total": n }`

> 更完整的请求参数、响应示例和模块错误码请查看 `docs/Admin-System-Management.md`。
>
> 管理端鉴权接口请查看 `docs/Admin-Auth.md`。当前仅“当前登录用户自己的敏感安全操作”统一接入二次验证流程：优先 Passkey，可切换密码；未使用 Passkey 且已开启 TFA 时，需先验证密码，再提交 `totp_code` 换取 `reauth_ticket`。`/go-api/internal/admin/auth/password`、`/identifier`、`/tfa/enable`、`/tfa/disable`、`/passkey/register/options`、`/passkey` 都必须携带 `reauth_ticket`。其中 `GET /go-api/internal/admin/auth/profile` 返回的 `email/phone` 为脱敏展示值；前端未修改该字段时，可在 `PUT /go-api/internal/admin/auth/identifier` 中原样回传，服务端会按“字段未变更”处理。密码相关字段（`/go-api/internal/admin/auth/token` 的 `grant_type=password`、`/password/reset`、`/reauth/password`、以及敏感操作统一验证里的密码步骤）前端必须传 `md5(明文密码)`。

### 健康检查端点

#### 外部健康检查

**GET** `/go-api/external/service/ping`

检查外部服务健康状态。

**响应:**
```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

#### 内部健康检查

**GET** `/go-api/internal/service/ping`

检查内部服务健康状态。

**响应:**
```json
{
  "code": 0,
  "message": "ok",
  "data": null
}
```

## 响应格式

所有API响应都遵循此标准格式：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    // 响应数据
  }
}
```

### 响应代码

| 代码 | 消息 | 描述 |
|------|------|------|
| 0 | ok | 成功 |
| -1 | System is busy | 系统错误 |
| 400 | Request parameter error | 参数无效 |
| 500 | fail | 服务器错误 |
| 10001 | Unauthorized | 需要认证 |
| 10002 | Authorization has failed | 认证失败 |
| 10003 | Authorization failed | 授权错误 |
| 10004 | Application does not exist or account info error | 应用不存在 |
| 10005 | Application already exists | 应用已存在 |
| 10006 | User does not exist | 用户不存在 |

## 错误处理

### 验证错误

当请求验证失败时：

```json
{
  "code": 400,
  "message": "Request parameter error",
  "data": null
}
```

### 认证错误

当认证失败时：

```json
{
  "code": 10001,
  "message": "Unauthorized",
  "data": null
}
```

### 服务器错误

当服务器遇到错误时：

```json
{
  "code": 500,
  "message": "fail",
  "data": null
}
```

## 速率限制

API请求受到速率限制以防止滥用：

- **限制**: 每IP每小时1000个请求
- **头部**: 速率限制信息在响应头中返回
  - `X-RateLimit-Limit`: 请求限制
  - `X-RateLimit-Remaining`: 剩余请求数
  - `X-RateLimit-Reset`: 重置时间（Unix时间戳）

## 分页

对于返回列表的端点，使用分页参数：

**查询参数:**
```
通用: page: int (默认: 1) - 页码
通用: page_size/size: int (默认: 10, 最大: 100) - 每页项目数
```

**响应格式:**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [...],
    "total": 100
  }
}
```

部分模块返回 `{ "items": [...], "total": n }`。

## API版本控制

API使用URL版本控制：

- 当前版本: `v1`（默认）
- 未来版本: `/go-api/v2/...`

## 使用Postman测试

### 环境变量

在Postman中设置这些变量：

```
base_url: http://localhost:8080
jwt_token: (从auth/token端点获取)
```

### 集合示例

1. **获取令牌**
   - 方法: POST
   - URL: `{{base_url}}/go-api/external/service/auth/token`
   - 请求体: form-data包含`app_id`和`app_secret`

2. **创建应用**
   - 方法: POST
   - URL: `{{base_url}}/go-api/external/service/auth/app`
   - 请求头: `Authorization: Bearer {{jwt_token}}`
   - 请求体: 包含应用详情的JSON

## SDK示例

### JavaScript/Node.js

```javascript
const axios = require('axios');

class GoAPIClient {
  constructor(baseURL, appId, appSecret) {
    this.baseURL = baseURL;
    this.appId = appId;
    this.appSecret = appSecret;
    this.token = null;
  }

  async authenticate() {
    const response = await axios.post(`${this.baseURL}/go-api/external/service/auth/token`, 
      new URLSearchParams({
        app_id: this.appId,
        app_secret: this.appSecret
      })
    );
    
    if (response.data.code === 0) {
      this.token = response.data.data.token;
      return this.token;
    }
    throw new Error('认证失败');
  }

  async createApp(appData) {
    if (!this.token) {
      await this.authenticate();
    }

    const response = await axios.post(
      `${this.baseURL}/go-api/external/service/auth/app`,
      appData,
      {
        headers: {
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json'
        }
      }
    );

    return response.data;
  }
}

// 使用方法
const client = new GoAPIClient('http://localhost:8080', 'your_app_id', 'your_app_secret');
```

### Python

```python
import requests
from urllib.parse import urlencode

class GoAPIClient:
    def __init__(self, base_url, app_id, app_secret):
        self.base_url = base_url
        self.app_id = app_id
        self.app_secret = app_secret
        self.token = None

    def authenticate(self):
        url = f"{self.base_url}/go-api/external/service/auth/token"
        data = {
            'app_id': self.app_id,
            'app_secret': self.app_secret
        }
        
        response = requests.post(url, data=data)
        result = response.json()
        
        if result['code'] == 0:
            self.token = result['data']['token']
            return self.token
        raise Exception('认证失败')

    def create_app(self, app_data):
        if not self.token:
            self.authenticate()

        url = f"{self.base_url}/go-api/external/service/auth/app"
        headers = {
            'Authorization': f'Bearer {self.token}',
            'Content-Type': 'application/json'
        }
        
        response = requests.post(url, json=app_data, headers=headers)
        return response.json()

# 使用方法
client = GoAPIClient('http://localhost:8080', 'your_app_id', 'your_app_secret')
```

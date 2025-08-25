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
page: int (默认: 1) - 页码
size: int (默认: 20, 最大: 100) - 每页项目数
```

**响应格式:**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

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
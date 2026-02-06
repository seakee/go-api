# 系统管理接口文档

## 概述
系统管理接口用于在后台控制台维护系统核心数据，包括用户管理、角色管理、权限管理、菜单管理、操作记录等功能模块。接口统一响应 `{code, msg, trace, data}` 结构，所有操作必须由具备管理员权限的用户调用。

## 环境信息
| 环境 | Host | HTTP 端口 |
| ---- | ---- |---------|
| 开发 | 127.0.0.1 | 8080    |
| 生产 | your-domain.com | 8080    |

> **Base URL**：`http://{host}:{port}`

## 认证方式
- Header：`Authorization: Bearer <admin-token>`
- Cookie（可选）：`admin-token=<token>`
- 所有路由均注册在 `/go-api/internal/admin/system` 路径下，并由 `CheckAdminAuth` 中间件校验。
- `/go-api/internal/admin/*` 路径统一经过 `SaveOperationRecord` 中间件，系统管理接口会写入操作日志。

## 通用响应格式
```json
{
  "code": 0,
  "msg": "OK",
  "trace": {
    "id": "f3f7a9f6ee024934",
    "desc": "" // debug 模式记录详细错误
  },
  "data": {}
}
```

> **字段命名说明**：部分接口直接返回 GORM 模型，字段会同时包含 `ID/CreatedAt/UpdatedAt/DeletedAt`（驼峰）与业务字段（多数为下划线）。

### 常见错误码
| Code | 含义 | 触发场景 |
| ---- | ---- | -------- |
| 0 | 成功 | 请求执行成功 |
| 400 | 参数无效 | 请求参数格式错误或缺失必填字段 |
| 500 | 服务器错误 | 服务端内部异常 |
| 11002 | 账号不存在 | 用户账号未找到 |
| 11007 | 账号无效 | 账号格式或状态无效 |
| 11013 | 账号已存在 | 创建用户时账号重复 |
| 11014 | 账号不能为空 | 创建用户时账号为空 |
| 11016 | 角色不存在 | 角色 ID 不存在 |
| 11017 | 角色名不能为空 | 创建角色时名称为空 |
| 11018 | 角色名已存在 | 创建角色时名称重复 |
| 11019 | 权限不存在 | 权限 ID 不存在 |
| 11020 | 权限名不能为空 | 创建权限时名称为空 |
| 11021 | 权限名已存在 | 创建权限时名称重复 |
| 11023 | 菜单名已存在 | 创建菜单时名称重复 |
| 11024 | 菜单不存在 | 菜单 ID 不存在 |
| 11025 | 菜单存在子菜单 | 删除菜单时存在子菜单 |
| 11032 | 密码不能为空 | 重置或更新密码时为空 |
| 11037 | 用户不可操作 | 尝试操作受保护的用户 |
| 11038 | 角色不可操作 | 尝试操作受保护的角色 |

---

# 基础接口 (System)

## 接口总览
| 功能 | 方法 | 路径 |
| ---- | ---- | ---- |
| 系统健康检查 | GET | `/go-api/internal/admin/system/ping` |

---

### 1. 系统健康检查
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/ping`
- **说明**：用于后台系统可用性检测

- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": null
  }
  ```

---

# 用户管理 (User)

## 接口总览
| 功能 | 方法 | 路径 |
| ---- | ---- | ---- |
| 分页查询用户 | GET | `/go-api/internal/admin/system/user/paginate` |
| 用户详情 | GET | `/go-api/internal/admin/system/user` |
| 创建用户 | POST | `/go-api/internal/admin/system/user` |
| 更新用户 | PUT | `/go-api/internal/admin/system/user` |
| 删除用户 | DELETE | `/go-api/internal/admin/system/user` |
| 获取用户角色 | GET | `/go-api/internal/admin/system/user/role` |
| 更新用户角色 | PUT | `/go-api/internal/admin/system/user/role` |
| 管理员重置用户密码 | PUT | `/go-api/internal/admin/system/user/password/reset` |
| 管理员关闭用户 TFA | PUT | `/go-api/internal/admin/system/user/tfa/disable` |

> **密码字段口径（与 `docs/Admin-Auth.md` 一致）**：
> `password` 建议统一传 `md5(明文密码)`；服务端最终会使用随机 salt 再做一次 `MD5(password + salt)` 后存储。

---

### 1. 分页查询用户
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/user/paginate`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 默认值 | 说明 |
  | ---- | ---- | ---- | ------ | ---- |
  | user_name | string | 否 | - | 用户名模糊查询 |
  | account | string | 否 | - | 账号模糊查询 |
  | status | int8 | 否 | - | 用户状态：`0` 禁用、`1` 正常 |
  | page | int | 否 | 1 | 页码 |
  | page_size | int | 否 | 10 | 单页数量 |

- **示例请求**：
  ```http
  GET http://127.0.0.1:8080/go-api/internal/admin/system/user/paginate?page=1&page_size=10
  Authorization: Bearer <admin-token>
  ```

- **响应结构**：

  | 字段 | 类型 | 说明 |
  | ---- | ---- | ---- |
  | list | array | 用户列表 |
  | list[].id | uint | 用户 ID |
  | list[].account | string | 账号 |
  | list[].user_name | string | 用户名 |
  | list[].feishu_id | string | 飞书 ID |
  | list[].wechat_id | string | 微信 ID |
  | list[].totp_enabled | bool | 是否启用 TOTP |
  | list[].status | int8 | 状态：0 禁用、1 正常 |
  | list[].avatar | string | 头像 URL |
  | list[].created_at | string | 创建时间 |
  | total | int64 | 记录总数 |

- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": {
      "list": [
        {
          "id": 1,
          "account": "admin",
          "user_name": "管理员",
          "feishu_id": "",
          "wechat_id": "",
          "totp_enabled": false,
          "status": 1,
          "avatar": "https://cdn.example/avatar/1.png",
          "created_at": "2024-01-15T10:30:00+08:00"
        }
      ],
      "total": 50
    }
  }
  ```

---

### 2. 用户详情
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/user`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 用户 ID |

- **响应结构**：与分页接口 `list[]` 结构相同

- **错误码**：`400`、`11002`

---

### 3. 创建用户
- **Method**：POST
- **Path**：`/go-api/internal/admin/system/user`
- **Body（JSON）**：
  ```json
  {
    "user_name": "张三",
    "account": "zhangsan",
    "password": "e10adc3949ba59abbe56e057f20f883e",
    "status": 1,
    "avatar": "https://cdn.example/avatar.png",
    "feishu_id": "",
    "wechat_id": ""
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | user_name | string | 否 | 用户名 |
  | account | string | 是 | 登录账号 |
  | password | string | 否 | 密码摘要，建议传 `md5(明文密码)`（空值时将使用默认初始密码逻辑） |
  | status | int8 | 否 | 状态，默认 1 |
  | avatar | string | 否 | 头像 URL |
  | feishu_id | string | 否 | 飞书 ID |
  | wechat_id | string | 否 | 微信 ID |

- **成功返回**：`code=0`
- **错误码**：`400`、`11013`、`11014`

---

### 4. 更新用户
- **Method**：PUT
- **Path**：`/go-api/internal/admin/system/user`
- **Body（JSON）**：
  ```json
  {
    "id": 1,
    "user_name": "张三（更新）",
    "account": "zhangsan",
    "password": "",
    "status": 1,
    "avatar": "https://cdn.example/avatar_new.png",
    "feishu_id": "",
    "wechat_id": ""
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 用户 ID |
  | user_name | string | 否 | 用户名 |
  | account | string | 否 | 登录账号 |
  | password | string | 否 | 密码摘要，建议传 `md5(明文密码)`（空则不更新） |
  | status | int8 | 否 | 状态 |
  | avatar | string | 否 | 头像 URL |
  | feishu_id | string | 否 | 飞书 ID |
  | wechat_id | string | 否 | 微信 ID |

- **成功返回**：`code=0`
- **错误码**：`400`、`11002`、`11037`

---

### 5. 删除用户
- **Method**：DELETE
- **Path**：`/go-api/internal/admin/system/user`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 用户 ID |

- **成功返回**：`code=0`
- **错误码**：`400`、`500`、`11037`

---

### 6. 获取用户角色
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/user/role`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | 是 | 用户 ID |

- **响应结构**：

  | 字段 | 类型 | 说明 |
  | ---- | ---- | ---- |
  | data | []uint | 用户拥有的角色 ID 列表 |

- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": [1, 2, 3]
  }
  ```

---

### 7. 更新用户角色
- **Method**：PUT
- **Path**：`/go-api/internal/admin/system/user/role`
- **Body（JSON）**：
  ```json
  {
    "user_id": 1,
    "role_ids": [1, 2]
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | 是 | 用户 ID |
  | role_ids | []uint | 否 | 角色 ID 列表（空数组则清空角色） |

- **成功返回**：`code=0`
- **错误码**：`400`、`500`、`11037`

> **注意**：`UpdateRole` 会自动补充 `base` 角色；即使 `role_ids` 不包含 `base`，最终仍会保留该角色。

---

### 8. 管理员重置用户密码
- **Method**：PUT
- **Path**：`/go-api/internal/admin/system/user/password/reset`
- **说明**：管理员重置指定用户密码（与鉴权模块的“当前用户重置密码”区分）
- **Body（JSON）**：
  ```json
  {
    "user_id": 1,
    "password": "e10adc3949ba59abbe56e057f20f883e"
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | 是 | 用户 ID |
  | password | string | 是 | 新密码摘要，建议传 `md5(明文密码)` |

- **成功返回**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": null
  }
  ```
- **错误码**：`400`、`11002`、`11032`、`11037`

---

### 9. 管理员关闭用户 TFA
- **Method**：PUT
- **Path**：`/go-api/internal/admin/system/user/tfa/disable`
- **说明**：管理员强制关闭指定用户 TFA（无需该用户的 totp_code）
- **Body（JSON）**：
  ```json
  {
    "user_id": 1
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | 是 | 用户 ID |

- **成功返回**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": null
  }
  ```
- **错误码**：`400`、`11002`、`11037`

---

### 备注
- 这两个接口属于“系统管理”范畴，与 Admin-Auth.md 的“当前用户自助”接口区分开。
- UI 中 user:password / user:update 权限建议对应后端鉴权。
- 删除用户会同步删除该用户在 `sys_role_user` 中的角色绑定关系。

---

# 角色管理 (Role)

## 接口总览
| 功能 | 方法 | 路径 |
| ---- | ---- | ---- |
| 角色列表（无分页） | GET | `/go-api/internal/admin/system/role/list` |
| 分页查询角色 | GET | `/go-api/internal/admin/system/role/paginate` |
| 角色详情 | GET | `/go-api/internal/admin/system/role` |
| 创建角色 | POST | `/go-api/internal/admin/system/role` |
| 更新角色 | PUT | `/go-api/internal/admin/system/role` |
| 删除角色 | DELETE | `/go-api/internal/admin/system/role` |
| 获取角色权限 | GET | `/go-api/internal/admin/system/role/permission` |
| 更新角色权限 | PUT | `/go-api/internal/admin/system/role/permission` |

---

### 1. 角色列表（无分页）
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/role/list`
- **说明**：获取所有角色的简要信息，通常用于下拉选择框

- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": [
      {"id": 1, "name": "管理员"},
      {"id": 2, "name": "编辑"}
    ]
  }
  ```

---

### 2. 分页查询角色
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/role/paginate`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 默认值 | 说明 |
  | ---- | ---- | ---- | ------ | ---- |
  | name | string | 否 | - | 角色名模糊查询 |
  | page | int | 否 | 1 | 页码 |
  | page_size | int | 否 | 10 | 单页数量 |

- **响应结构**：

  | 字段 | 类型 | 说明 |
  | ---- | ---- | ---- |
  | list | array | 角色列表 |
  | list[].id | uint | 角色 ID |
  | list[].name | string | 角色名称 |
  | list[].description | string | 角色描述 |
  | list[].created_at | string | 创建时间 |
  | list[].updated_at | string | 更新时间 |
  | total | int64 | 记录总数 |

---

### 3. 角色详情
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/role`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 角色 ID |

- **响应结构**：与分页接口 `list[]` 结构相同
- **错误码**：`400`、`11016`

---

### 4. 创建角色
- **Method**：POST
- **Path**：`/go-api/internal/admin/system/role`
- **Body（JSON）**：
  ```json
  {
    "name": "运维人员",
    "description": "负责系统运维工作"
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | name | string | 是 | 角色名称 |
  | description | string | 否 | 角色描述 |

- **成功返回**：`code=0`
- **错误码**：`400`、`11017`、`11018`

---

### 5. 更新角色
- **Method**：PUT
- **Path**：`/go-api/internal/admin/system/role`
- **Body（JSON）**：
  ```json
  {
    "id": 1,
    "name": "超级管理员",
    "description": "拥有所有权限"
  }
  ```
- **成功返回**：`code=0`
- **错误码**：`400`、`11016`、`11038`

---

### 6. 删除角色
- **Method**：DELETE
- **Path**：`/go-api/internal/admin/system/role`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 角色 ID |

- **成功返回**：`code=0`
- **错误码**：`400`、`11016`、`11038`

---

### 7. 获取角色权限
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/role/permission`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | role_id | uint | 是 | 角色 ID |

- **响应结构**：

  | 字段 | 类型 | 说明 |
  | ---- | ---- | ---- |
  | data | []uint | 角色拥有的权限 ID 列表 |

---

### 8. 更新角色权限
- **Method**：PUT
- **Path**：`/go-api/internal/admin/system/role/permission`
- **Body（JSON）**：
  ```json
  {
    "role_id": 1,
    "permission_ids": [1, 2, 3, 4, 5]
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | role_id | uint | 是 | 角色 ID |
  | permission_ids | []uint | 否 | 权限 ID 列表（空数组则清空权限） |

- **成功返回**：`code=0`
- **错误码**：`400`、`500`、`11016`、`11038`

---

# 权限管理 (Permission)

## 接口总览
| 功能 | 方法 | 路径 |
| ---- | ---- | ---- |
| 可用权限列表 | GET | `/go-api/internal/admin/system/permission/available` |
| 权限列表（按分组） | GET | `/go-api/internal/admin/system/permission/list` |
| 分页查询权限 | GET | `/go-api/internal/admin/system/permission/paginate` |
| 权限详情 | GET | `/go-api/internal/admin/system/permission` |
| 创建权限 | POST | `/go-api/internal/admin/system/permission` |
| 更新权限 | PUT | `/go-api/internal/admin/system/permission` |
| 删除权限 | DELETE | `/go-api/internal/admin/system/permission` |

---

### 1. 可用权限列表
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/permission/available`
- **说明**：获取系统中“尚未创建 permission 记录”的后台路由列表（仅 `/go-api/internal/admin/*`，不含 `ping`）

- **响应结构**：按 HTTP Method 分组

  | 字段 | 类型 | 说明 |
  | ---- | ---- | ---- |
  | data | object | 可用路由映射 |
  | data.GET | []string | 待添加的 GET 路径 |
  | data.POST | []string | 待添加的 POST 路径 |
  | data.PUT | []string | 待添加的 PUT 路径 |
  | data.DELETE | []string | 待添加的 DELETE 路径 |

- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": {
      "GET": [
        "/go-api/internal/admin/system/user",
        "/go-api/internal/admin/system/role/list"
      ],
      "POST": [
        "/go-api/internal/admin/system/user"
      ]
    }
  }
  ```

---

### 2. 权限列表（按分组）
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/permission/list`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | type | string | 是 | 权限类型：`api`、`menu` 等 |

- **响应结构**：按 `group` 分组返回权限列表

- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": {
      "用户管理": [
        {"id": 1, "name": "查看用户", "method": "GET", "path": "/user"},
        {"id": 2, "name": "创建用户", "method": "POST", "path": "/user"}
      ],
      "角色管理": [
        {"id": 3, "name": "查看角色", "method": "GET", "path": "/role"}
      ]
    }
  }
  ```

---

### 3. 分页查询权限
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/permission/paginate`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 默认值 | 说明 |
  | ---- | ---- | ---- | ------ | ---- |
  | type | string | 是 | - | 权限类型：`api`、`menu` |
  | name | string | 否 | - | 权限名模糊查询 |
  | group | string | 否 | - | 分组过滤 |
  | method | string | 否 | - | HTTP 方法过滤：`GET`、`POST`、`PUT`、`DELETE` |
  | page | int | 否 | 1 | 页码 |
  | page_size | int | 否 | 10 | 单页数量 |

- **响应结构**：

  | 字段 | 类型 | 说明 |
  | ---- | ---- | ---- |
  | list | array | 权限列表 |
  | list[].id | uint | 权限 ID |
  | list[].name | string | 权限名称 |
  | list[].type | string | 权限类型 |
  | list[].method | string | HTTP 方法 |
  | list[].path | string | 路径 |
  | list[].description | string | 描述 |
  | list[].group | string | 分组 |
  | list[].created_at | string | 创建时间 |
  | list[].updated_at | string | 更新时间 |
  | total | int64 | 记录总数 |

---

### 4. 权限详情
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/permission`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 权限 ID |

- **响应结构**：与分页接口 `list[]` 结构相同
- **错误码**：`400`、`11019`

---

### 5. 创建权限
- **Method**：POST
- **Path**：`/go-api/internal/admin/system/permission`
- **Body（JSON）**：
  ```json
  {
    "name": "查看用户",
    "type": "api",
    "method": "GET",
    "path": "/go-api/internal/admin/system/user",
    "description": "获取用户信息权限",
    "group": "用户管理"
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | name | string | 是 | 权限名称 |
  | type | string | 是 | 权限类型 |
  | method | string | 是 | HTTP 方法 |
  | path | string | 是 | 路径 |
  | description | string | 是 | 描述 |
  | group | string | 是 | 分组 |

- **成功返回**：`code=0`
- **错误码**：`400`、`11020`、`11021`

---

### 6. 更新权限
- **Method**：PUT
- **Path**：`/go-api/internal/admin/system/permission`
- **Body（JSON）**：与创建接口相同，增加 `id` 字段
- **成功返回**：`code=0`
- **错误码**：`400`、`11019`

---

### 7. 删除权限
- **Method**：DELETE
- **Path**：`/go-api/internal/admin/system/permission`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 权限 ID |

- **成功返回**：`code=0`
- **错误码**：`400`、`500`

---

# 菜单管理 (Menu)

## 接口总览
| 功能 | 方法 | 路径 |
| ---- | ---- | ---- |
| 菜单树列表 | GET | `/go-api/internal/admin/system/menu/list` |
| 菜单详情 | GET | `/go-api/internal/admin/system/menu` |
| 创建菜单 | POST | `/go-api/internal/admin/system/menu` |
| 更新菜单 | PUT | `/go-api/internal/admin/system/menu` |
| 删除菜单 | DELETE | `/go-api/internal/admin/system/menu` |

---

### 1. 菜单树列表
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/menu/list`
- **说明**：返回树形结构的菜单列表

- **响应结构**：

  | 字段 | 类型 | 说明 |
  | ---- | ---- | ---- |
  | items | array | 菜单树列表 |
  | items[].id | uint | 菜单 ID |
  | items[].name | string | 菜单名称 |
  | items[].path | string | 路由路径 |
  | items[].icon | string | 图标 |
  | items[].sort | int | 排序值 |
  | items[].parent_id | uint | 父菜单 ID，0 为顶级 |
  | items[].permission_id | uint | 关联的权限 ID |
  | items[].children | array | 子菜单列表（递归结构） |

- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": {
      "items": [
        {
          "id": 1,
          "name": "系统管理",
          "path": "/system",
          "icon": "setting",
          "sort": 1,
          "parent_id": 0,
          "permission_id": 0,
          "children": [
            {
              "id": 2,
              "name": "用户管理",
              "path": "/system/user",
              "icon": "user",
              "sort": 1,
              "parent_id": 1,
              "permission_id": 1,
              "children": []
            }
          ]
        }
      ]
    }
  }
  ```

---

### 2. 菜单详情
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/menu`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 菜单 ID |

- **响应结构**：与菜单树 `items[]` 单项结构相同（不含 `children`）
- **错误码**：`400`、`11024`

---

### 3. 创建菜单
- **Method**：POST
- **Path**：`/go-api/internal/admin/system/menu`
- **Body（JSON）**：
  ```json
  {
    "parent_id": 1,
    "name": "角色管理",
    "path": "/system/role",
    "icon": "peoples",
    "sort": 2
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | parent_id | uint | 否 | 父菜单 ID，0 表示顶级菜单 |
  | name | string | 是 | 菜单名称 |
  | path | string | 是 | 路由路径 |
  | icon | string | 否 | 图标名称 |
  | sort | int | 是 | 排序值（支持 0） |

- **成功返回**：`code=0`
- **错误码**：`400`、`11023`

---

### 4. 更新菜单
- **Method**：PUT
- **Path**：`/go-api/internal/admin/system/menu`
- **Body（JSON）**：
  ```json
  {
    "id": 2,
    "name": "用户管理（更新）",
    "path": "/system/user",
    "icon": "user",
    "sort": 1
  }
  ```
- **字段说明**：

  | 字段 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 菜单 ID |
  | name | string | 否 | 菜单名称 |
  | path | string | 否 | 路由路径 |
  | icon | string | 否 | 图标名称 |
  | sort | int | 否 | 排序值 |

- **成功返回**：`code=0`
- **错误码**：`400`、`11024`

> **注意**：当更新 `name` 时，会同步更新该菜单关联权限（`permission_id`）的权限名称。

---

### 5. 删除菜单
- **Method**：DELETE
- **Path**：`/go-api/internal/admin/system/menu`
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | uint | 是 | 菜单 ID |

- **成功返回**：`code=0`
- **错误码**：`400`、`11024`、`11025`

> **注意**：若菜单存在子菜单，删除将失败并返回 `11025`，需先删除子菜单。

---

# 操作记录 (Operation Record)

## 接口总览
| 功能 | 方法 | 路径 |
| ---- | ---- | ---- |
| 分页查询操作记录 | GET | `/go-api/internal/admin/system/record/paginate` |
| 操作记录交互详情 | GET | `/go-api/internal/admin/system/record/interaction` |

---

### 1. 分页查询操作记录
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/record/paginate`
- **说明**：查询系统操作日志记录（存储于 `sys_operation_record`，GORM 模型）
- **Query 参数**：

  | 名称 | 类型 | 必填 | 默认值 | 说明 |
  | ---- | ---- | ---- | ------ | ---- |
  | id | string | 否 | - | 记录 ID（数字字符串） |
  | path | string | 否 | - | 请求路径过滤 |
  | user_id | uint | 否 | - | 用户 ID 过滤 |
  | user_name | string | 否 | - | 用户名过滤 |
  | ip | string | 否 | - | IP 地址过滤 |
  | status | int | 否 | - | HTTP 状态码过滤 |
  | method | string | 否 | - | HTTP 方法过滤 |
  | page | int | 否 | 1 | 页码 |
  | size | int | 否 | 10 | 单页数量 |

- **响应结构**：

  | 字段 | 类型 | 说明 |
  | ---- | ---- | ---- |
  | items | array | 记录列表 |
  | items[].ID | uint | 记录 ID |
  | items[].CreatedAt | string | 创建时间 |
  | items[].UpdatedAt | string | 更新时间 |
  | items[].DeletedAt | object/null | 软删除时间 |
  | items[].ip | string | 请求 IP |
  | items[].method | string | HTTP 方法 |
  | items[].path | string | 请求路径 |
  | items[].status | int | HTTP 状态码 |
  | items[].latency | float64 | 请求耗时（毫秒） |
  | items[].agent | string | User-Agent |
  | items[].error_message | string | 错误信息 |
  | items[].user_id | uint | 用户 ID |
  | items[].user_name | string | 用户名 |
  | items[].trace_id | string | 链路追踪 ID |
  | total | int64 | 记录总数 |

- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": {
      "items": [
        {
          "ID": 1024,
          "CreatedAt": "2024-10-12T13:00:00+08:00",
          "UpdatedAt": "2024-10-12T13:00:01+08:00",
          "DeletedAt": null,
          "ip": "192.168.1.100",
          "method": "POST",
          "path": "/go-api/internal/admin/system/user",
          "status": 200,
          "latency": 45.5,
          "agent": "Mozilla/5.0",
          "error_message": "",
          "user_id": 1,
          "user_name": "admin",
          "trace_id": "abc123def456"
        }
      ],
      "total": 1000
    }
  }
  ```

---

### 2. 操作记录交互详情
- **Method**：GET
- **Path**：`/go-api/internal/admin/system/record/interaction`
- **说明**：获取指定操作记录的请求参数和响应详情
- **Query 参数**：

  | 名称 | 类型 | 必填 | 说明 |
  | ---- | ---- | ---- | ---- |
  | id | string | 是 | 记录 ID |

- **响应结构**：

  | 字段 | 类型 | 说明 |
  | ---- | ---- | ---- |
  | params | object | 请求参数（优先解析 JSON，失败回退 querystring/raw） |
  | resp | object | 响应内容（JSON 反序列化结果，失败回退原始字符串对象） |

- **响应示例**：
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": {
      "params": {
        "user_name": "test",
        "account": "test"
      },
      "resp": {
        "code": 0,
        "msg": "OK",
        "data": null
      }
    }
  }
  ```

- **错误码**：`400`、`500`

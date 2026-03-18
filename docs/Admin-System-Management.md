# System Management API Documentation

**Languages**: [English](Admin-System-Management.md) | [中文](Admin-System-Management-zh.md)

---

## Overview
System management APIs are used by the admin console to maintain core system data, including user management, role management, permission management, menu management, and operation records. All endpoints use the unified `{code, msg, trace, data}` response structure, and every operation must be performed by a user with admin privileges.

## Environment Information
| Environment | Host | HTTP Port |
| ---- | ---- | ---- |
| Development | 127.0.0.1 | 8080 |
| Production | your-domain.com | 8080 |

> **Base URL**: `http://{host}:{port}`

## Authentication
- Header: `Authorization: Bearer <admin-token>`
- Cookie (optional): `admin-token=<token>`
- All routes are registered under `/go-api/internal/admin/system` and validated by the `CheckAdminAuth` middleware.
- All `/go-api/internal/admin/*` routes go through `SaveOperationRecord`, so system management APIs also write operation logs.

## Common Response Format
```json
{
  "code": 0,
  "msg": "OK",
  "trace": {
    "id": "f3f7a9f6ee024934",
    "desc": "" // detailed error text in debug mode
  },
  "data": {}
}
```

> **Field naming note**: some endpoints return GORM models directly, so the payload may contain both `ID/CreatedAt/UpdatedAt/DeletedAt` (camel case) and business fields (mostly snake case).

### Common Error Codes
| Code | Meaning | Trigger |
| ---- | ---- | ---- |
| 0 | Success | Request completed successfully |
| 400 | Invalid parameters | Malformed request parameters or required fields missing |
| 500 | Server error | Internal server exception |
| 11002 | User not found | Target user does not exist |
| 11007 | Invalid identifier | Invalid email or phone format |
| 11013 | Identifier already exists | Duplicate email or phone when creating a user |
| 11014 | Identifier cannot be empty | Both email and phone are empty when creating a user |
| 11016 | Role not found | Role ID does not exist |
| 11017 | Role name cannot be empty | Empty role name during creation |
| 11018 | Role name already exists | Duplicate role name during creation |
| 11019 | Permission not found | Permission ID does not exist |
| 11020 | Permission name cannot be empty | Empty permission name during creation |
| 11021 | Permission name already exists | Duplicate permission name during creation |
| 11023 | Menu name already exists | Duplicate menu name during creation |
| 11024 | Menu not found | Menu ID does not exist |
| 11025 | Menu has child menus | Deleting a menu that still has children |
| 11032 | Password cannot be empty | Password is empty when resetting or updating |
| 11037 | User cannot be operated on | Attempt to operate on a protected user |
| 11038 | Role cannot be operated on | Attempt to operate on a protected role |
| 11049 | At least one login method must remain | Deleting OAuth/Passkey would leave the account with no login method |
| 11052 | Passkey credential not found | The specified Passkey record does not exist |

---

# Basic Endpoints (System)

## Endpoint Overview
| Function | Method | Path |
| ---- | ---- | ---- |
| System health check | GET | `/go-api/internal/admin/system/ping` |

---

### 1. System Health Check
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/ping`
- **Description**: Used to detect admin system availability

- **Response Example**:
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": null
  }
  ```

---

# User Management

## Endpoint Overview
| Function | Method | Path |
| ---- | ---- | ---- |
| Paginate users | GET | `/go-api/internal/admin/system/user/paginate` |
| User detail | GET | `/go-api/internal/admin/system/user` |
| Create user | POST | `/go-api/internal/admin/system/user` |
| Update user | PUT | `/go-api/internal/admin/system/user` |
| Delete user | DELETE | `/go-api/internal/admin/system/user` |
| Get user roles | GET | `/go-api/internal/admin/system/user/role` |
| Update user roles | PUT | `/go-api/internal/admin/system/user/role` |
| Admin reset user password | PUT | `/go-api/internal/admin/system/user/password/reset` |
| Admin disable user TFA | PUT | `/go-api/internal/admin/system/user/tfa/disable` |
| Query user Passkeys | GET | `/go-api/internal/admin/system/user/passkeys` |
| Delete a single user Passkey | DELETE | `/go-api/internal/admin/system/user/passkey` |
| Delete all user Passkeys | DELETE | `/go-api/internal/admin/system/user/passkeys` |

> **Password field contract (consistent with `docs/Admin-Auth.md`)**:
> `password` should consistently be sent as `md5(plaintext password)`, and the server stores that digest with bcrypt.

---

### 1. Paginate Users
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/user/paginate`
- **Query Parameters**:

  | Name | Type | Required | Default | Description |
  | ---- | ---- | ---- | ---- | ---- |
  | user_name | string | No | - | Fuzzy search by user name |
  | email | string | No | - | Fuzzy search by email |
  | phone | string | No | - | Fuzzy search by phone |
  | status | int8 | No | - | User status: `0` disabled, `1` active |
  | page | int | No | 1 | Page number |
  | page_size | int | No | 10 | Items per page |

- **Example Request**:
  ```http
  GET http://127.0.0.1:8080/go-api/internal/admin/system/user/paginate?page=1&page_size=10
  Authorization: Bearer <admin-token>
  ```

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | list | array | User list |
  | list[].id | uint | User ID |
  | list[].email | string | Email |
  | list[].phone | string | Phone |
  | list[].user_name | string | User name |
  | list[].totp_enabled | bool | Whether TOTP is enabled |
  | list[].passkey_count | int64 | Number of Passkeys registered by the user |
  | list[].status | int8 | Status: `0` disabled, `1` active |
  | list[].avatar | string | Avatar URL |
  | list[].created_at | string | Creation time |
  | total | int64 | Total record count |

- **Response Example**:
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": {
      "list": [
        {
          "id": 1,
          "email": "admin@example.com",
          "phone": "+8613800000000",
          "user_name": "管理员",
          "totp_enabled": false,
          "passkey_count": 2,
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

### 2. User Detail
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/user`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | User ID |

- **Response Structure**: same as `list[]` in the pagination endpoint

- **Error Codes**: `400`, `11002`

---

### 3. Create User
- **Method**: POST
- **Path**: `/go-api/internal/admin/system/user`
- **Body (JSON)**:
  ```json
  {
    "user_name": "Zhang San",
    "email": "zhangsan@example.com",
    "phone": "+8613800000001",
    "password": "e10adc3949ba59abbe56e057f20f883e",
    "status": 1,
    "avatar": "https://cdn.example/avatar.png"
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | user_name | string | No | User name |
  | email | string | No | Login email. Either this or `phone` must be provided, or both |
  | phone | string | No | Login phone. Either this or `email` must be provided, or both |
  | password | string | Yes | Password digest. Recommended value: `md5(plaintext password)` |
  | status | int8 | No | Status, default `1` |
  | avatar | string | No | Avatar URL |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11013`, `11014`

---

### 4. Update User
- **Method**: PUT
- **Path**: `/go-api/internal/admin/system/user`
- **Body (JSON)**:
  ```json
  {
    "id": 1,
    "user_name": "Zhang San (Updated)",
    "email": "zhangsan@example.com",
    "phone": "+8613800000001",
    "password": "",
    "status": 1,
    "avatar": "https://cdn.example/avatar_new.png"
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | User ID |
  | user_name | string | No | User name |
  | email | string | No | Login email |
  | phone | string | No | Login phone |
  | password | string | No | Password digest. Recommended value: `md5(plaintext password)`; empty means no update |
  | status | int8 | No | Status |
  | avatar | string | No | Avatar URL |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11002`, `11037`

---

### 5. Delete User
- **Method**: DELETE
- **Path**: `/go-api/internal/admin/system/user`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | User ID |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `500`, `11037`

- **Notes**:
  - When deleting a user, the server also removes related records in `sys_user_passkey`, `sys_user_identity`, and `sys_role_user` within the same transaction.

---

### 6. Get User Roles
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/user/role`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | Yes | User ID |

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | data | []uint | List of role IDs assigned to the user |

- **Response Example**:
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": [1, 2, 3]
  }
  ```

---

### 7. Update User Roles
- **Method**: PUT
- **Path**: `/go-api/internal/admin/system/user/role`
- **Body (JSON)**:
  ```json
  {
    "user_id": 1,
    "role_ids": [1, 2]
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | Yes | User ID |
  | role_ids | []uint | No | Role ID list. An empty array clears all assignable roles |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `500`, `11037`

> **Note**: `UpdateRole` automatically adds the `base` role. Even if `role_ids` does not include `base`, it remains assigned in the final result.

---

### 8. Admin Reset User Password
- **Method**: PUT
- **Path**: `/go-api/internal/admin/system/user/password/reset`
- **Description**: Admin resets the password for a specified user. This is different from the auth module's self-service password reset.
- **Body (JSON)**:
  ```json
  {
    "user_id": 1,
    "password": "e10adc3949ba59abbe56e057f20f883e"
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | Yes | User ID |
  | password | string | Yes | New password digest. Recommended value: `md5(plaintext password)`, stored with bcrypt by the server |

- **Successful Return**:
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": null
  }
  ```
- **Error Codes**: `400`, `11002`, `11032`, `11037`

---

### 9. Admin Disable User TFA
- **Method**: PUT
- **Path**: `/go-api/internal/admin/system/user/tfa/disable`
- **Description**: Admin forcibly disables TFA for the specified user without requiring that user's `totp_code`
- **Body (JSON)**:
  ```json
  {
    "user_id": 1
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | Yes | User ID |

- **Successful Return**:
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": null
  }
  ```
- **Error Codes**: `400`, `11002`, `11037`

---

### 10. Query User Passkeys
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/user/passkeys`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | Yes | User ID |

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | list | array | Passkey list |
  | list[].id | uint | Passkey primary key |
  | list[].display_name | string | Device display name |
  | list[].aaguid | string | Authenticator AAGUID, may be empty |
  | list[].transports | []string | Browser-reported transport methods |
  | list[].last_used_at | string/null | Last usage time |
  | list[].created_at | string | Creation time |

- **Response Example**:
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": {
      "list": [
        {
          "id": 11,
          "display_name": "MacBook Pro",
          "aaguid": "00000000000000000000000000000000",
          "transports": ["internal", "hybrid"],
          "last_used_at": "2026-02-06T12:00:00+08:00",
          "created_at": "2026-02-01T09:00:00+08:00"
        }
      ]
    }
  }
  ```

- **Error Codes**: `400`, `11002`, `500`

---

### 11. Delete a Single User Passkey
- **Method**: DELETE
- **Path**: `/go-api/internal/admin/system/user/passkey`
- **Body (JSON)**:
  ```json
  {
    "user_id": 1,
    "id": 11
  }
  ```

- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | Yes | User ID |
  | id | uint | Yes | Passkey primary key |

- **Behavior Notes**:
  - The server deletes precisely by `user_id + id`.
  - If the target user is the protected `super_admin`, it returns `11037`.
  - If deleting this Passkey would leave the account with no available login method, it returns `11049`.

- **Error Codes**: `400`, `11037`, `11049`, `11052`, `500`

---

### 12. Delete All User Passkeys
- **Method**: DELETE
- **Path**: `/go-api/internal/admin/system/user/passkeys`
- **Body (JSON)**:
  ```json
  {
    "user_id": 1
  }
  ```

- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | user_id | uint | Yes | User ID |

- **Behavior Notes**:
  - If the user currently has no Passkeys, the endpoint returns success directly.
  - If the target user is the protected `super_admin`, it returns `11037`.
  - If deleting all Passkeys would leave the account with no available login method, it returns `11049`.

- **Error Codes**: `400`, `11037`, `11049`, `500`

---

### Notes
- Interfaces such as disabling user TFA and Passkey maintenance belong to the "system management" category and are distinct from the "current user self-service" endpoints documented in `docs/Admin-Auth.md`.
- In the UI, permissions such as `user:password` and `user:update` should align with backend auth checks.
- Deleting a user also removes related rows in `sys_role_user`, `sys_user_identity`, and `sys_user_passkey`.

---

# Role Management

## Endpoint Overview
| Function | Method | Path |
| ---- | ---- | ---- |
| Role list (no pagination) | GET | `/go-api/internal/admin/system/role/list` |
| Paginate roles | GET | `/go-api/internal/admin/system/role/paginate` |
| Role detail | GET | `/go-api/internal/admin/system/role` |
| Create role | POST | `/go-api/internal/admin/system/role` |
| Update role | PUT | `/go-api/internal/admin/system/role` |
| Delete role | DELETE | `/go-api/internal/admin/system/role` |
| Get role permissions | GET | `/go-api/internal/admin/system/role/permission` |
| Get role menu permission tree | GET | `/go-api/internal/admin/system/role/permission/menu-tree` |
| Update role permissions | PUT | `/go-api/internal/admin/system/role/permission` |

---

### 1. Role List (No Pagination)
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/role/list`
- **Description**: Returns summary information for all roles, usually for select/dropdown options

- **Response Example**:
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

### 2. Paginate Roles
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/role/paginate`
- **Query Parameters**:

  | Name | Type | Required | Default | Description |
  | ---- | ---- | ---- | ---- | ---- |
  | name | string | No | - | Fuzzy search by role name |
  | page | int | No | 1 | Page number |
  | page_size | int | No | 10 | Items per page |

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | list | array | Role list |
  | list[].id | uint | Role ID |
  | list[].name | string | Role name |
  | list[].description | string | Role description |
  | list[].created_at | string | Creation time |
  | list[].updated_at | string | Update time |
  | total | int64 | Total record count |

---

### 3. Role Detail
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/role`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | Role ID |

- **Response Structure**: same as `list[]` in the pagination endpoint
- **Error Codes**: `400`, `11016`

---

### 4. Create Role
- **Method**: POST
- **Path**: `/go-api/internal/admin/system/role`
- **Body (JSON)**:
  ```json
  {
    "name": "Operations",
    "description": "Responsible for system operations"
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | name | string | Yes | Role name |
  | description | string | No | Role description |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11017`, `11018`

---

### 5. Update Role
- **Method**: PUT
- **Path**: `/go-api/internal/admin/system/role`
- **Body (JSON)**:
  ```json
  {
    "id": 1,
    "name": "Super Admin",
    "description": "Has all permissions"
  }
  ```
- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11016`, `11038`

---

### 6. Delete Role
- **Method**: DELETE
- **Path**: `/go-api/internal/admin/system/role`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | Role ID |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11016`, `11038`

---

### 7. Get Role Permissions
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/role/permission`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | role_id | uint | Yes | Role ID |

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | data | []uint | List of permission IDs owned by the role |

---

### 8. Update Role Permissions
- **Method**: PUT
- **Path**: `/go-api/internal/admin/system/role/permission`
- **Body (JSON)**:
  ```json
  {
    "role_id": 1,
    "permission_ids": [1, 2, 3, 4, 5]
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | role_id | uint | Yes | Role ID |
  | permission_ids | []uint | No | Permission ID list. An empty array clears permissions |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `500`, `11016`, `11038`

---

### 9. Get Role Menu Permission Tree
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/role/permission/menu-tree`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | role_id | uint | Yes | Role ID |

- **Description**: Returns the menu tree and attaches `checked` to each node to indicate whether that menu's `permission_id` is granted to the current role.

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | data.items | []object | Menu permission tree |
  | data.items[].id | uint | Menu ID |
  | data.items[].name | string | Menu name |
  | data.items[].path | string | Menu path |
  | data.items[].permission_id | uint | Permission ID linked to the menu |
  | data.items[].parent_id | uint | Parent menu ID |
  | data.items[].icon | string | Menu icon |
  | data.items[].sort | int | Sort order |
  | data.items[].checked | bool | Whether the current role has this menu permission |
  | data.items[].children | []object | Child menus |

- **Response Example**:
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
      ]
    }
  }
  ```

- **Error Codes**: `400`, `500`, `11016`

---

# Permission Management

## Endpoint Overview
| Function | Method | Path |
| ---- | ---- | ---- |
| Available permission list | GET | `/go-api/internal/admin/system/permission/available` |
| Permission list (grouped) | GET | `/go-api/internal/admin/system/permission/list` |
| Paginate permissions | GET | `/go-api/internal/admin/system/permission/paginate` |
| Permission detail | GET | `/go-api/internal/admin/system/permission` |
| Create permission | POST | `/go-api/internal/admin/system/permission` |
| Update permission | PUT | `/go-api/internal/admin/system/permission` |
| Delete permission | DELETE | `/go-api/internal/admin/system/permission` |

---

### 1. Available Permission List
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/permission/available`
- **Description**: Returns admin routes that do not yet have a permission record, limited to `/go-api/internal/admin/*` and excluding `ping`

- **Response Structure**: grouped by HTTP method

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | data | object | Available route map |
  | data.GET | []string | GET paths that can still be added |
  | data.POST | []string | POST paths that can still be added |
  | data.PUT | []string | PUT paths that can still be added |
  | data.DELETE | []string | DELETE paths that can still be added |

- **Response Example**:
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

### 2. Permission List (Grouped)
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/permission/list`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | type | string | Yes | Permission type, such as `api` or `menu` |

- **Response Structure**: returns permissions grouped by `group`

- **Response Example**:
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

### 3. Paginate Permissions
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/permission/paginate`
- **Query Parameters**:

  | Name | Type | Required | Default | Description |
  | ---- | ---- | ---- | ---- | ---- |
  | type | string | Yes | - | Permission type: `api` or `menu` |
  | name | string | No | - | Fuzzy search by permission name |
  | group | string | No | - | Group filter |
  | method | string | No | - | HTTP method filter: `GET`, `POST`, `PUT`, `DELETE` |
  | page | int | No | 1 | Page number |
  | page_size | int | No | 10 | Items per page |

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | list | array | Permission list |
  | list[].id | uint | Permission ID |
  | list[].name | string | Permission name |
  | list[].type | string | Permission type |
  | list[].method | string | HTTP method |
  | list[].path | string | Path |
  | list[].description | string | Description |
  | list[].group | string | Group |
  | list[].created_at | string | Creation time |
  | list[].updated_at | string | Update time |
  | total | int64 | Total record count |

---

### 4. Permission Detail
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/permission`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | Permission ID |

- **Response Structure**: same as `list[]` in the pagination endpoint
- **Error Codes**: `400`, `11019`

---

### 5. Create Permission
- **Method**: POST
- **Path**: `/go-api/internal/admin/system/permission`
- **Body (JSON)**:
  ```json
  {
    "name": "View User",
    "type": "api",
    "method": "GET",
    "path": "/go-api/internal/admin/system/user",
    "description": "Permission to get user information",
    "group": "用户管理"
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | name | string | Yes | Permission name |
  | type | string | Yes | Permission type |
  | method | string | Yes | HTTP method |
  | path | string | Yes | Path |
  | description | string | Yes | Description |
  | group | string | Yes | Group |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11020`, `11021`

---

### 6. Update Permission
- **Method**: PUT
- **Path**: `/go-api/internal/admin/system/permission`
- **Body (JSON)**: same as the create endpoint, with an additional `id` field
- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11019`

---

### 7. Delete Permission
- **Method**: DELETE
- **Path**: `/go-api/internal/admin/system/permission`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | Permission ID |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `500`

---

# Menu Management

## Endpoint Overview
| Function | Method | Path |
| ---- | ---- | ---- |
| Menu tree list | GET | `/go-api/internal/admin/system/menu/list` |
| Menu detail | GET | `/go-api/internal/admin/system/menu` |
| Create menu | POST | `/go-api/internal/admin/system/menu` |
| Update menu | PUT | `/go-api/internal/admin/system/menu` |
| Delete menu | DELETE | `/go-api/internal/admin/system/menu` |

---

### 1. Menu Tree List
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/menu/list`
- **Description**: Returns the menu list as a tree structure

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | items | array | Menu tree list |
  | items[].id | uint | Menu ID |
  | items[].name | string | Menu name |
  | items[].path | string | Route path |
  | items[].icon | string | Icon |
  | items[].sort | int | Sort value |
  | items[].parent_id | uint | Parent menu ID, `0` means top-level |
  | items[].permission_id | uint | Linked permission ID |
  | items[].children | array | Child menu list (recursive structure) |

- **Response Example**:
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

### 2. Menu Detail
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/menu`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | Menu ID |

- **Response Structure**: same as a single `items[]` node in the menu tree, without `children`
- **Error Codes**: `400`, `11024`

---

### 3. Create Menu
- **Method**: POST
- **Path**: `/go-api/internal/admin/system/menu`
- **Body (JSON)**:
  ```json
  {
    "parent_id": 1,
    "name": "Role Management",
    "path": "/system/role",
    "icon": "peoples",
    "sort": 2
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | parent_id | uint | No | Parent menu ID, `0` means a top-level menu |
  | name | string | Yes | Menu name |
  | path | string | Yes | Route path |
  | icon | string | No | Icon name |
  | sort | int | Yes | Sort value, `0` is allowed |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11023`

---

### 4. Update Menu
- **Method**: PUT
- **Path**: `/go-api/internal/admin/system/menu`
- **Body (JSON)**:
  ```json
  {
    "id": 2,
    "name": "User Management (Updated)",
    "path": "/system/user",
    "icon": "user",
    "sort": 1
  }
  ```
- **Field Description**:

  | Field | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | Menu ID |
  | name | string | No | Menu name |
  | path | string | No | Route path |
  | icon | string | No | Icon name |
  | sort | int | No | Sort value |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11024`

> **Note**: when `name` is updated, the permission name of the linked permission (`permission_id`) is updated at the same time.

---

### 5. Delete Menu
- **Method**: DELETE
- **Path**: `/go-api/internal/admin/system/menu`
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | uint | Yes | Menu ID |

- **Successful Return**: `code=0`
- **Error Codes**: `400`, `11024`, `11025`

> **Note**: if the menu still has child menus, deletion fails with `11025`. Delete child menus first.

---

# Operation Record

## Endpoint Overview
| Function | Method | Path |
| ---- | ---- | ---- |
| Paginate operation records | GET | `/go-api/internal/admin/system/record/paginate` |
| Operation record detail | GET | `/go-api/internal/admin/system/record/detail` |

---

### 1. Paginate Operation Records
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/record/paginate`
- **Description**: Queries system operation logs stored in `sys_operation_record`; the user name is resolved from `sys_user.user_name` through `user_id`
- **Query Parameters**:

  | Name | Type | Required | Default | Description |
  | ---- | ---- | ---- | ---- | ---- |
  | path | string | No | - | Request path filter |
  | user_id | uint | No | - | User ID filter |
  | ip | string | No | - | IP address filter |
  | status | int | No | - | Business status code filter, i.e. response JSON `code` |
  | method | string | No | - | HTTP method filter |
  | trace_id | string | No | - | Trace ID filter |
  | page | int | No | 1 | Page number |
  | size | int | No | 10 | Items per page |

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | items | array | Record list |
  | items[].ID | uint | Record ID |
  | items[].Method | string | HTTP method |
  | items[].Path | string | Request path |
  | items[].IP | string | Request IP |
  | items[].Status | int | Business status code, i.e. response JSON `code` |
  | items[].UserName | string | User name |
  | items[].TraceID | string | Trace ID |
  | items[].CreatedAt | string | Creation time |
  | total | int64 | Total record count |

- **Response Example**:
  ```json
  {
    "code": 0,
    "msg": "OK",
    "data": {
      "items": [
        {
          "ID": 1024,
          "Method": "POST",
          "Path": "/go-api/internal/admin/system/user",
          "IP": "192.168.1.100",
          "Status": 0,
          "UserName": "admin",
          "TraceID": "abc123def456",
          "CreatedAt": "2024-10-12T13:00:00+08:00"
        }
      ],
      "total": 1000
    }
  }
  ```

---

### 2. Operation Record Detail
- **Method**: GET
- **Path**: `/go-api/internal/admin/system/record/detail`
- **Description**: Returns the full log detail for a given operation record ID
- **Query Parameters**:

  | Name | Type | Required | Description |
  | ---- | ---- | ---- | ---- |
  | id | string | Yes | Record ID |

- **Response Structure**:

  | Field | Type | Description |
  | ---- | ---- | ---- |
  | id | uint | Record ID |
  | method | string | HTTP method |
  | path | string | Request path |
  | ip | string | Request IP |
  | status | int | Business status code, i.e. response JSON `code` |
  | user_id | uint | User ID |
  | user_name | string | User name from `sys_user` |
  | trace_id | string | Trace ID |
  | created_at | string | Creation time |
  | latency | float64 | Request latency in seconds |
  | agent | string | User-Agent |
  | error_message | string | Error message |
  | params | object | Formatted request parameters, preferring parsed JSON and falling back to query/raw |
  | resp | object | Formatted response payload, preferring parsed JSON and falling back to raw |

- **Error Codes**: `400`, `500`

---

> **Note**: in the detail endpoint, `params` and `resp` are returned after server-side formatting according to the parsing rules, preferring JSON and falling back to query/raw when parsing fails.

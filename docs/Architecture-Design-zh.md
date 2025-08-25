# 架构设计

**语言版本**: [English](Architecture-Design.md) | [中文](Architecture-Design-zh.md)

---

## 概述

Go-API框架遵循干净的分层架构模式，专为企业级应用设计。它强调关注点分离、依赖注入和可维护性。

## 架构层次

```
┌─────────────────────────────────────┐
│           表现层                    │
│        (HTTP 控制器)                │
├─────────────────────────────────────┤
│           业务层                    │
│          (服务层)                   │
├─────────────────────────────────────┤
│         数据访问层                  │
│         (仓库层)                    │
├─────────────────────────────────────┤
│            数据层                   │
│      (模型和数据库)                 │
└─────────────────────────────────────┘
```

### 1. 表现层（控制器）

**目的**: 处理HTTP请求和响应
**位置**: `app/http/controller/`

- 接收客户端的HTTP请求
- 验证输入参数
- 调用适当的服务方法
- 格式化并返回HTTP响应
- 通过中间件处理认证和授权

**示例结构**:
```go
type Handler interface {
    Create() gin.HandlerFunc
    GetByID() gin.HandlerFunc
    Update() gin.HandlerFunc
    Delete() gin.HandlerFunc
}

type handler struct {
    controller.BaseController
    service auth.AppService
}
```

### 2. 业务层（服务）

**目的**: 实现业务逻辑和规则
**位置**: `app/service/`

- 包含所有业务逻辑
- 协调多个仓库之间的操作
- 实现复杂的业务规则和验证
- 处理事务和数据一致性
- 独立于HTTP关注点

**示例结构**:
```go
type AppService interface {
    CreateApp(ctx context.Context, params *CreateAppParams) (*CreateAppResult, error)
    ValidateCredentials(ctx context.Context, appID, appSecret string) (*App, error)
}

type appService struct {
    repo auth.AppRepo
}
```

### 3. 数据访问层（仓库）

**目的**: 抽象数据访问操作
**位置**: `app/repository/`

- 提供数据库操作的抽象
- 实现数据访问模式
- 处理数据库连接和查询
- 在领域模型和数据库实体之间转换
- 支持多种数据库类型（MySQL、MongoDB）

**示例结构**:
```go
type AppRepo interface {
    Create(ctx context.Context, app *auth.App) (uint, error)
    GetByID(ctx context.Context, id uint) (*auth.App, error)
    Update(ctx context.Context, id uint, app *auth.App) error
    Delete(ctx context.Context, id uint) error
}
```

### 4. 数据层（模型）

**目的**: 定义数据结构和基本操作
**位置**: `app/model/`

- 定义数据结构（结构体）
- 包含基本的CRUD操作
- 包括数据库模式定义（GORM标签）
- 处理数据验证和序列化
- 支持SQL和NoSQL数据库

## 依赖注入模式

框架使用基于构造函数的依赖注入：

```go
// 服务层只接收它需要的依赖
func NewAppService(db *gorm.DB, redis *redis.Manager) AppService {
    repo := authRepo.NewAppRepo(db, redis)
    return &appService{repo: repo}
}

// 控制器层将依赖注入到服务中
func NewHandler(appCtx *http.Context) Handler {
    return &handler{
        service: authService.NewAppService(appCtx.MysqlDB["go-api"], appCtx.Redis["go-api"]),
    }
}
```

## 中间件架构

```
请求 → SetTraceID → CORS → RequestLogger → CheckAppAuth → 控制器
```

中间件组件：
- **SetTraceID**: 为请求跟踪生成唯一的跟踪ID
- **CORS**: 处理跨域请求
- **RequestLogger**: 记录请求详情用于监控
- **CheckAppAuth**: 验证JWT令牌进行身份认证

## 配置管理

支持热重载的多环境配置：

```json
{
  "system": {
    "name": "go-api",
    "run_mode": "debug|release",
    "http_port": ":8080"
  },
  "databases": [...],
  "redis": [...],
  "kafka": {...}
}
```

环境特定文件：
- `local.json` - 本地开发
- `dev.json` - 开发环境
- `prod.json` - 生产环境

## 错误处理策略

1. **错误传播**: 错误在各层之间向上冒泡
2. **集中处理**: 控制器处理所有错误响应
3. **国际化**: 错误消息支持多种语言
4. **结构化日志**: 所有错误都记录上下文

## 安全架构

1. **JWT认证**: 基于令牌的认证系统
2. **中间件保护**: 路由级别的认证
3. **输入验证**: 控制器层的参数验证
4. **SQL注入防护**: 通过GORM使用参数化查询
5. **CORS保护**: 可配置的跨域策略

## 性能考虑

1. **连接池**: 数据库连接池提高效率
2. **缓存**: Redis集成用于缓存频繁访问的数据
3. **异步处理**: 通过工作者进行后台作业处理
4. **结构化日志**: 使用Zap进行高性能日志记录
5. **优雅关闭**: 关闭时正确清理资源

## 扩展性设计

### 水平扩展

1. **无状态设计**: 应用实例不保存会话状态
2. **数据库分片**: 支持数据库水平分片
3. **缓存分布**: Redis集群支持
4. **负载均衡**: 支持多实例部署

### 垂直扩展

1. **资源调优**: 数据库连接池、Redis连接池优化
2. **内存管理**: Go垃圾回收器优化
3. **CPU利用**: 协程池和工作队列
4. **IO优化**: 异步处理和批量操作

## 监控和可观测性

### 日志系统

- **结构化日志**: 使用Zap提供高性能日志
- **日志级别**: Debug、Info、Warn、Error
- **上下文传递**: TraceID贯穿整个请求链路
- **日志聚合**: 支持ELK、Fluentd等日志收集

### 指标监控

- **应用指标**: QPS、响应时间、错误率
- **系统指标**: CPU、内存、磁盘、网络
- **业务指标**: 用户行为、业务操作统计
- **自定义指标**: 基于业务需求的自定义监控

### 链路追踪

- **请求追踪**: 每个请求分配唯一TraceID
- **跨服务追踪**: 支持微服务间调用追踪
- **性能分析**: 识别性能瓶颈和优化点
- **错误定位**: 快速定位问题根因

## 最佳实践

### 代码组织

1. **分层清晰**: 严格遵循分层架构，避免跨层调用
2. **接口抽象**: 使用接口定义服务契约
3. **依赖注入**: 通过构造函数注入依赖
4. **单一职责**: 每个模块专注单一功能

### 数据库设计

1. **索引优化**: 合理设计数据库索引
2. **查询优化**: 避免N+1查询问题
3. **事务管理**: 正确使用数据库事务
4. **连接池**: 合理配置连接池参数

### 安全实践

1. **输入验证**: 严格验证所有用户输入
2. **权限控制**: 实现细粒度的权限管理
3. **数据加密**: 敏感数据加密存储
4. **审计日志**: 记录关键操作的审计信息

### 性能优化

1. **缓存策略**: 合理使用Redis缓存
2. **异步处理**: 耗时操作异步执行
3. **资源池化**: 复用昂贵资源（连接、对象等）
4. **批量操作**: 减少数据库交互次数
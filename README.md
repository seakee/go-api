# Go-API
go-api是基于`Gin`基础上开发的一个可以快速开始的api脚手架

## 快速开始
在终端执行以下命令获取安装脚本
```shell
# 下载初始化脚本
curl -O --location --request GET 'https://raw.githubusercontent.com/seakee/go-api/main/scripts/generate.sh' && chmod +x generate.sh

# 初始化项目
# 例子：./generate.sh cms-api v1.0.0
./generate.sh projectName projectVersion
```
## 目录结构
```shell
go-api
├── README.md 						# 自述文件
├───── app 						# 应用业务目录
│ 	├── command 					# 命令目录
│ 	│ 	└── handler.go 				# 命令处理入口
│ 	├── config.go 					# 系统配置信息
│ 	├── consumer 					# kafka消费者处理目录
│ 	│ 	└── handler.go 				# kafka消费者处理入口
│ 	├── http 					# HTTP
│ 	│ 	├── controller 				# 控制器
│ 	│ 	│   └── auth 				# 权限
│ 	│ 	│       ├── app.go 			# 接入项目的应用
│ 	│ 	│       ├── handler.go
│ 	│ 	│       └── jwt.go
│ 	│ 	├── middleware 				# http中间件目录
│ 	│ 	│   ├── check_app_auth.go 	        # 鉴权中间件
│ 	│ 	│   ├── cors.go 			# 跨域中间件
│ 	│ 	│   ├── handler.go 			# 中间件入口
│ 	│ 	│   └── requset_logger.go 	        # 请求日志中间件
│ 	│ 	└── router                              # 路由
│ 	│ 	    ├── auth.go
│ 	│ 	    └── handler.go 			# 路由入口
│ 	├── model 					# 数据库 model
│ 	│ 	└── auth
│ 	│ 	    └── app.go
│ 	├── pkg 					# 业务package
│ 	│    ├── e 					# 错误相关目录
│ 	│    │   └── code.go 			        # 接口业务返回码
│ 	│    └── jwt
│ 	│ 	   └── jwt.go
│ 	├── repository 					# 数据访问层
│ 	│ 	 └── auth
│ 	│            └── app.go
│ 	└── service 					# 数据服务层
│ 	         └── handler.go
├───── bin 						# 编译目录
│ 	├── configs 				        # 项目配置目录
│ 	│ 	 ├── dev.json
│ 	│ 	 ├── local.json
│ 	│ 	 └── prod.json
│ 	├── data 					# 项目数据目录
│ 	│     └── sql 					# 项目SQL目录
│ 	│     	  └── auth_app.sql
│ 	└── lang 					# 国际化语言目录
│ 	      ├── en-US.json
│ 	      └── zh-CN.json
├───── bootstrap 					# 启动目录
│ 	├── app.go 					# 应用启动逻辑
│ 	├── http.go 					# http服务启动逻辑
│ 	└── kafka.go 					# kafka服务启动逻辑
├───── go.mod
├───── go.sum
├───── main.go 						# 启动入口
├───── scripts 						# 脚本
│         └── generate.sh 				# 生成一个api项目脚本
└───── vendor                                           # 引入的依赖	
```
- `README.md` 项目自述
- `app` 应用业务目录
  - `command` 项目使用到的自定义命令，需要在handler.go里面定义接口，然后实现接口
  - `config.go` 项目配置处理文件，如果当前环境为本地(local)环境，则直接加载./bin/configs/local.json 配置文件。其他环境，则需要从配置中心加载对应环境的配置。
  - `http` http应用目录，处理http相关业务
    - `controller` 控制器目录，需要将http相关业务放到此目录里面，每一个独立的业务应该有自己独立的目录。如：后台管理业务controller/admin
    - `middleware` http中间件，所有中间件需要以实现该目录handler.go中Middleware接口
      - `check_app_auth.go` 拦截服务端API的http请求，进行权限鉴定。
      - `cors.go` 跨域中间件
      - `handler.go` 定义所有http中间件接口，也是中间件初始化入口。
      - `requset_logger.go` 请求日志中间件，会记录请求相关信息。默认情况下非prod环境会开启。开发者亦可在自己需要用到的路由使用。
    - `router` 路由目录，http请求路由在该目录里面定义
  - `model` 数据库model，该目录定义了数据对象及基础数据库操作方法
  - `pkg` 业务package，该目录用来放一些该项目自身使用的package
    - `e` 错误相关定义目录，
      - `code.go` 将该项目中的错误定义成 int 常量，配合国际化使用
    - `jwt` JWT相关处理
  - `repository` 数据仓库，对数据库数据加工处理
  - `service` 数据服务层
- `bin` 项目编译后运行目录
  - `configs` 项目配置目录
  - `data` 项目存储目录，用来放一些项目运行时需要用到的数据
  - `lang` 国际化语言目录
- `bootstrap` 项目启动目录，项目启动加载相关逻辑
- `vendor` 项目引用的一些外部依赖package
- 
## 开发指南
### 如何连接一个新的数据库
连接一个新的数据库，需要在配置文件`bin/configs/{env}.json`databases中写入新的数据库的配置，并将enable设置为true如：
```json
"databases": [
  {
    "enable": true,
    "db_type": "mysql",
    "db_host": "127.0.0.1:3306",
    "db_name": "mysql-db2",
    "db_username": "db_username",
    "db_password": "db_password",
    "db_max_idle_conn": 10,
    "db_max_open_conn": 50,
    "db_max_lifetime": 3
  },
  {
    "enable": true,
    "db_type": "mysql",
    "db_host": "127.0.0.1:3306",
    "db_name": "mysql-db2",
    "db_username": "db_username",
    "db_password": "db_password",
    "db_max_idle_conn": 10,
    "db_max_open_conn": 50,
    "db_max_lifetime": 3
  },
  {
    "enable": true,
    "db_type": "mongo-db1",
    "db_name": "db_name",
    "db_host": "url"
  },
  {
    "enable": true,
    "db_type": "mongo-db2",
    "db_name": "db_name",
    "db_host": "url"
  }
],
```
### 如何新增一个中间件
要新增一个http中间件，需要在`app/http/middleware/handler.go`文件中Middleware接口中先定义好中间件方法，并实现这个方法。注意：中间件的返回值必须是`gin.HandlerFunc`
```go
type Middleware interface {
   CheckAppAuth() gin.HandlerFunc

   // ImNewMiddleware 新增的中间件
   ImNewMiddleware() gin.HandlerFunc
}

func (m middleware) ImNewMiddleware() gin.HandlerFunc {
   return func(c *gin.Context) {
       c.Next()
   }
}
```
### 应该怎么进行错误处理
为了方便debug追踪错误，我们应该将所有可能产的error返回到最外层，然后通过接口返回出去
```go
func a() error {
   err := errors.New("this is an error")
   
   return err
}

func (h handler) returnFunc()gin.HandlerFunc  {
   return func(c *gin.Context) {
      
      err := a()
      
      h.i18n.JSON(c, e.SUCCESS, nil, err)
   }
}
```
### 国际化应该怎么处理
#### Q:在什么地方定义国际化状态码 
A:应该在app/pkg/e/code.go文件中定义状态码，可以看见的是，目前该文件中已经定义好了一些基础的状态码，`-1~1000`为基础的状态码，`10001~10999`为服务端状态码，`11000~11050`为授权状态码。建议新定义的状态码应该以1000为一个区间，在已经定义的状态码后面增加。定义的状态码常量应该尽量短，且意思明确。
#### Q:在什么地方定义国际化语言
A:在bin/lang目录中定义，语言包的命名方式应该类似zh-CN.json
#### Q:如果国际化语言中带有变量，应该怎么使用
A:在国际化语言包中定义翻译用语言。例如：
  ```json
  {
    "1000": "你好,%s!你的账号是:%s"
  }
  ```
  ```go
  func (h handler) returnFunc() gin.HandlerFunc {
   return func(c *gin.Context) {

      errCode := 1000

      h.i18n.JSON(c, errCode, i18n.Data{
         Params: []string{"Seakee", "18888888888"},
         Data:   "test",
      }, nil)
   }
  }
  ```


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
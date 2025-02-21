# Go-API

`go-api` is a simple, powerful, and high-performance Go framework for building web APIs.

## Features

- **Router**: Fast and flexible router with middleware support.
- **Dependency Injection**: Built-in dependency injection container.
- **Configuration**: Unified configuration management.
- **Logger**: High-performance logging library.
- **Database**: Integration with GORM for database operations.
- **Validation**: Parameter validation using struct tags.
- **Task Scheduling**: Built-in support for task scheduling.
- **Code Generation**: Automatically generate code based on SQL files.

## Quick Start

Execute the following commands in the terminal to get the installation script:

```shell
# Download the initialization script
curl -O --location --request GET 'https://raw.githubusercontent.com/seakee/go-api/main/scripts/generate.sh' && chmod +x generate.sh

# Initialize the project
# Example: ./generate.sh cms-api v1.0.0
./generate.sh projectName projectVersion
```

## Directory Structure

```shell
go-api
├── README.md                       # Project documentation
├── app                             # Application business logic directory
│   ├── command                     # Commands directory
│   │   └── handler.go              # Command handler entry
│   │   └── codegen                 # Code generation related directory
│   │       └── handler.go          # Code generation handler
│   │       └── codegen             # Code generation toolkit
│   │           └── model.go        # Model generation related code
│   ├── config.go                   # System configuration
│   ├── consumer                    # Kafka consumer handlers directory
│   │   └── handler.go              # Kafka consumer handler entry
│   ├── http                        # HTTP related directory
│   │   ├── controller              # Controllers directory
│   │   │   └── auth                # Authorization related controllers
│   │   │       ├── app.go          # Application access controller
│   │   │       ├── handler.go      # Controller handler
│   │   │       └── jwt.go          # JWT related controller
│   │   ├── middleware              # HTTP middleware directory
│   │   │   ├── check_app_auth.go   # Authentication middleware
│   │   │   ├── cors.go             # CORS middleware
│   │   │   ├── handler.go          # Middleware entry
│   │   │   └── requset_logger.go   # Request logger middleware
│   │   └── router                  # Routing directory
│   │       ├── auth.go             # Authentication related routes
│   │       └── handler.go          # Routing entry point
│   ├── model                       # Database models directory
│   │   └── auth                    # Authorization related models
│   │       └── app.go              # Application access model
│   ├── pkg                         # Business packages directory
│   │   ├── e                       # Error handling directory
│   │   │   └── code.go             # Interface business response codes
│   │   └── jwt                     # JWT related directory
│   │       └── jwt.go              # JWT related code
│   ├── repository                  # Data access layer directory
│   │   └── auth                    # Authorization related repositories
│   │       └── app.go              # Application access repository
│   └── service                     # Data service layer directory
│       └── handler.go              # Service layer handler
├── bin                             # Build directory
│   ├── configs                     # Project configurations
│   │   ├── dev.json                # Development environment config
│   │   ├── local.json              # Local environment config
│   │   └── prod.json               # Production environment config
│   ├── data                        # Project data directory
│   │   └── sql                     # SQL scripts directory
│   │       └── auth_app.sql        # Authorization application SQL file
│   └── lang                        # Internationalization directory
│       ├── en-US.json              # English language file
│       └── zh-CN.json              # Chinese language file
├── bootstrap                       # Startup directory
│   ├── app.go                      # Application startup logic
│   ├── http.go                     # HTTP service startup
│   └── kafka.go                    # Kafka service startup
├── go.mod                          # Go module definition
├── go.sum                          # Go module dependencies
├── main.go                         # Main entry point
├── scripts                         # Scripts directory
│   └── generate.sh                 # API project generation script
└── vendor                          # Dependency packages directory
```

- `README.md`: Project readme
- `app`: Application business directory
    - `config.go`: Project configuration file, if the current environment is local, it directly loads the config file `./bin/configs/local.json`. For other environments, it loads the corresponding environment configuration from the configuration center.
    - `http`: HTTP application directory, handles HTTP-related business
        - `controller`: Controller directory, place HTTP related business here. Each independent business should have its own directory, e.g., `controller/admin` for admin business.
        - `middleware`: HTTP middleware, all middleware should implement the `Middleware` interface in `handler.go`
            - `check_app_auth.go`: Intercepts HTTP requests for server-side API and performs authentication.
            - `cors.go`: CORS middleware
            - `handler.go`: Defines all HTTP middleware interfaces and serves as the middleware initialization entry.
            - `requset_logger.go`: Request logger middleware, records request-related information. By default, it is enabled in non-prod environments. Developers can use it in the routes where needed.
        - `router`: Router directory, define HTTP request routes here.
    - `model`: Database models, defines data objects and basic database operation methods.
    - `pkg`: Business package, used to place some packages used by the project itself
        - `e`: Error-related definitions directory
            - `code.go`: Defines error codes as int constants, used with internationalization.
        - `jwt`: JWT-related handling
    - `repository`: Data repository, processes database data
    - `service`: Data service layer
- `command`: Custom commands used in the project, define interfaces in handler.go, and then implement the interfaces
- `bin`: Project compilation and running directory
    - `configs`: Project configuration directory
    - `data`: Project storage directory, used to place data needed during project runtime
    - `lang`: Internationalization language directory
- `bootstrap`: Project startup directory, loads related logic on startup
- `vendor`: External dependencies referenced by the project

## Development Guide

### How to Connect to a New Database

To connect to a new database, add the new database configuration in the `bin/configs/{env}.json` file under `databases` and set `enable` to `true`, for example:

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
    "db_type": "mongo",
    "db_name": "db_name",
    "db_host": "mongodb://db_host:27017",
    "db_username": "go-api",
    "db_password": "db_username",
    "db_max_idle_conn": 10,
    "db_max_open_conn": 50,
    "auth_mechanism": "SCRAM-SHA-1",
    "db_max_lifetime": 1
  },
  {
    "enable": true,
    "db_type": "mongo-db2",
    "db_name": "db_name",
    "db_host": "url"
  }
],
```

### How to Add a New Middleware

To add a new HTTP middleware, first define the middleware method in the `Middleware` interface in the `app/http/middleware/handler.go` file, and implement this method. Note: the return value of the middleware must be `gin.HandlerFunc`.

```go
type Middleware interface {
   CheckAppAuth() gin.HandlerFunc

   // ImNewMiddleware: New middleware
   ImNewMiddleware() gin.HandlerFunc
}

func (m middleware) ImNewMiddleware() gin.HandlerFunc {
   return func(c *gin.Context) {
       c.Next()
   }
}
```

### How to Handle Errors

To facilitate debugging and tracking errors, all possible errors should be returned to the outermost layer and then returned through the interface.

```go
func a() error {
   err := errors.New("this is an error")
   
   return err
}

func (h handler) returnFunc() gin.HandlerFunc {
   return func(c *gin.Context) {
      
      err := a()
      
      h.i18n.JSON(c, e.SUCCESS, nil, err)
   }
}
```

### How to Handle Internationalization

#### Q: Where to define internationalization status codes?

A: Status codes should be defined in the `app/pkg/e/code.go` file. You can see that some basic status codes have already been defined in this file, `-1~1000` for basic status codes, `10001~10999` for server-side status codes, and `11000~11050` for authorization status codes. It is recommended that new status codes should be added in increments of 1000, following the already defined status codes. The defined status code constants should be as short and clear as possible.

#### Q: Where to define internationalization languages?

A: Define in the `bin/lang` directory, with language package names similar to `zh-CN.json`.

#### Q: How to use variables in internationalization languages?

A: Define the translation language in the internationalization language package. For example:

```json
{
  "1000": "Hello, %s! Your account is: %s"
}
```

```go
func (h handler) returnFunc() gin.HandlerFunc {
   return func(c *gin.Context) {

      errCode := 1000

      h.i18n.JSON(c, errCode, i18n.Data{
         Params: []string{"Seakee", "12345678"},
         Data:   "test",
      }, nil)
   }
}
```
## License

`go-api` is released under the MIT License. See the [LICENSE](LICENSE) file for more details.
package bootstrap

import (
	"context"
	"fmt"
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"github.com/seakee/go-api/app/config"
	"github.com/sk-pkg/xdb"
	mgOpt "go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"time"
)

// loadDB initializes the database components.
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - error: An error if any database initialization fails.
func (a *App) loadDB(ctx context.Context) error {
	for _, db := range a.Config.Databases {
		if db.Enable {
			switch db.DbType {
			case "mongo":
				if err := a.initMongo(ctx, db); err != nil {
					return err
				}
			case "mysql", "postgres", "sqlite", "sqlserver", "clickhouse":
				if err := a.initXDB(ctx, db); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown db type: %s", db.DbType)
			}
		}
	}

	return nil
}

// xdbDriverMap maps database type strings to xdb driver constants.
var xdbDriverMap = map[string]xdb.Driver{
	"mysql":      xdb.MySQL,
	"postgres":   xdb.PostgreSQL,
	"sqlite":     xdb.SQLite,
	"sqlserver":  xdb.SQLServer,
	"clickhouse": xdb.ClickHouse,
}

// xdbDefaultPorts defines default ports for each database type.
var xdbDefaultPorts = map[string]int{
	"mysql":      3306,
	"postgres":   5432,
	"sqlite":     0,
	"sqlserver":  1433,
	"clickhouse": 9000,
}

// initXDB initializes a database connection using xdb (MySQL, PostgreSQL, SQLite, SQL Server, ClickHouse).
//
// Parameters:
//   - ctx: The context for the operation.
//   - db: The database configuration.
//
// Returns:
//   - error: An error if the initialization fails.
func (a *App) initXDB(ctx context.Context, db config.Database) error {
	driver, ok := xdbDriverMap[db.DbType]
	if !ok {
		return fmt.Errorf("unsupported xdb driver: %s", db.DbType)
	}

	dbLogger := xdb.NewLog(a.Logger.CallerSkipMode(4),
		xdb.WithSlowThreshold(200*time.Millisecond),
		xdb.WithIgnoreRecordNotFoundError(true),
	)

	// Set default port if not specified
	port := db.DbPort
	if port == 0 {
		port = xdbDefaultPorts[db.DbType]
	}

	// Set default charset for MySQL if not specified
	charset := db.Charset
	if charset == "" && db.DbType == "mysql" {
		charset = "utf8mb4"
	}

	d, err := xdb.New(
		xdb.WithDBConfig(xdb.DBConfig{
			Driver:          driver,
			Host:            db.DbHost,
			Port:            port,
			User:            db.DbUsername,
			Password:        db.DbPassword,
			DBName:          db.DbName,
			Charset:         charset,
			SSLMode:         db.SSLMode,
			Timezone:        db.Timezone,
			MaxIdleConns:    db.DbMaxIdleConn,
			MaxOpenConns:    db.DbMaxOpenConn,
			ConnMaxLifetime: db.ConnMaxLifetime * time.Hour,
			ConnMaxIdleTime: db.ConnMaxIdleTime * time.Hour,
		}),
		xdb.WithGormConfig(gorm.Config{Logger: dbLogger}),
	)
	if err != nil {
		return err
	}

	// Enable gorm debug mode if debug mode is on and not in production
	if a.Config.System.DebugMode && a.Config.System.Env != "prod" {
		d = d.Debug()
	}

	a.SqlDB[db.DbName] = d

	a.Logger.Info(ctx, fmt.Sprintf("%s %s loaded successfully", db.DbType, db.DbName))

	return nil
}

// initMongo initializes a MongoDB connection.
//
// Parameters:
//   - ctx: The context for the operation.
//   - db: The database configuration.
//
// Returns:
//   - error: An error if the MongoDB initialization fails.
func (a *App) initMongo(ctx context.Context, db config.Database) error {
	maxPoolSize := uint64(db.DbMaxOpenConn)
	minPoolSize := uint64(db.DbMaxIdleConn)
	maxConnIdleTime := db.ConnMaxIdleTime * time.Hour

	opts := options.ClientOptions{ClientOptions: &mgOpt.ClientOptions{MaxConnIdleTime: &maxConnIdleTime}}
	cli, err := qmgo.NewClient(ctx, &qmgo.Config{
		Uri:         db.DbHost,
		MaxPoolSize: &maxPoolSize,
		MinPoolSize: &minPoolSize,
		Auth: &qmgo.Credential{
			AuthMechanism: db.AuthMechanism,
			AuthSource:    db.DbName,
			Username:      db.DbUsername,
			Password:      db.DbPassword,
		},
	}, opts)
	if err != nil {
		return err
	}

	a.MongoDB[db.DbName] = cli.Database(db.DbName)

	a.Logger.Info(ctx, fmt.Sprintf("MongoDB %s loaded successfully", db.DbName))

	return nil
}

package bootstrap

import (
	"context"
	"fmt"
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"github.com/seakee/go-api/app/config"
	"github.com/sk-pkg/mysql"
	mgOpt "go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"time"
)

// loadDB initializes the database components (MySQL and MongoDB).
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
			case "mysql":
				if err := a.initMySQL(ctx, db); err != nil {
					return err
				}
			case "mongo":
				if err := a.initMongo(ctx, db); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown db type: %s", db.DbType)
			}
		}
	}

	return nil
}

// initMySQL initializes a MySQL database connection.
//
// Parameters:
//   - ctx: The context for the operation.
//   - db: The database configuration.
//
// Returns:
//   - error: An error if the MySQL initialization fails.
func (a *App) initMySQL(ctx context.Context, db config.Database) error {
	mysqlLogger := mysql.NewLog(a.Logger.CallerSkipMode(4))

	d, err := mysql.New(mysql.WithConfigs(
		mysql.Config{
			User:     db.DbUsername,
			Password: db.DbPassword,
			Host:     db.DbHost,
			DBName:   db.DbName,
		}),
		mysql.WithConnMaxLifetime(db.DbMaxLifetime*time.Hour),
		mysql.WithMaxIdleConn(db.DbMaxIdleConn),
		mysql.WithMaxOpenConn(db.DbMaxOpenConn),
		mysql.WithGormConfig(gorm.Config{Logger: mysqlLogger}),
	)
	if err != nil {
		return err
	}

	// if debug mode and not prod, enable gorm debug mode
	if a.Config.System.DebugMode && a.Config.System.Env != "prod" {
		d = d.Debug()
	}

	a.MysqlDB[db.DbName] = d

	a.Logger.Info(ctx, fmt.Sprintf("MySQL %s loaded successfully", db.DbName))

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
	maxConnIdleTime := db.DbMaxLifetime * time.Hour

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

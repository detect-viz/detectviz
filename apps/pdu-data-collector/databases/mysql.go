package databases

import (
	"bimap-zbox/global"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func LoadMysql() *gorm.DB {
	// get env config
	host := global.EnvConfig.DC.Mysql.Host
	port := global.EnvConfig.DC.Mysql.Port
	user := global.EnvConfig.DC.Mysql.User
	password := global.EnvConfig.DC.Mysql.Password
	dbname := global.EnvConfig.DC.Mysql.DBName
	params := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbname)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  logger.LogLevel(3),
		},
	)

	err := errors.New("mock error")
	var db *gorm.DB
	for err != nil {
		db, err = gorm.Open(mysql.Open(params), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   newLogger,
		})
		time.Sleep(1 * time.Second)

		if global.EnvConfig.Global.Log.Level == "debug" {
			db = db.Debug()
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		global.Logger.Error(
			fmt.Sprintf("mysql [%v] connection error: %v", params, err.Error()),
		)
		return nil
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(int(global.EnvConfig.DC.Mysql.MaxIdle))

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(int(global.EnvConfig.DC.Mysql.MaxOpenConn))

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	lifeTime, _ := time.ParseDuration(global.EnvConfig.DC.Mysql.MaxLifeTime)
	sqlDB.SetConnMaxLifetime(lifeTime)

	global.Logger.Info(
		fmt.Sprintf("mysql [%v] connection success", dbname),
	)

	global.Mysql = db
	return db
}

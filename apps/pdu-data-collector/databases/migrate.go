package databases

import (
	"bimap-zbox/global"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	//"github.com/golang-migrate/migrate/v4/database/sqlserver"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func LoadMigrate() {
	// 從設定中讀取 MySQL 資料庫資訊
	host := global.EnvConfig.DC.Mysql.Host
	port := global.EnvConfig.DC.Mysql.Port
	user := global.EnvConfig.DC.Mysql.User
	password := global.EnvConfig.DC.Mysql.Password
	dbname := global.EnvConfig.DC.Mysql.DBName

	// 連接 MySQL 資料庫
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbname))
	if err != nil {
		global.Logger.Error("資料庫連接失敗", zap.String("錯誤", err.Error()))
		return
	}
	defer db.Close()

	// 建立遷移驅動
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		global.Logger.Error("建立 MySQL 遷移驅動失敗", zap.String("錯誤", err.Error()))
		return
	}

	// 初始化遷移，包含遷移路徑
	migratePath := global.EnvConfig.DC.Mysql.MigratePath
	m, err := migrate.NewWithDatabaseInstance(migratePath, "mysql", driver)
	if err != nil {
		global.Logger.Error("migrate 初始化失敗", zap.String("錯誤", err.Error()))
		return
	}

	// 執行遷移
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		global.Logger.Error("migrate 執行失敗", zap.String("錯誤", err.Error()))
		return
	}

	global.Logger.Info("migrate 遷移成功執行")
}

package databases

import (
	"bimap-zbox/global"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func LoadSqlserver() {
	username := global.EnvConfig.DC.SQLServer.User
	password := global.EnvConfig.DC.SQLServer.Password
	hostname := global.EnvConfig.DC.SQLServer.Host
	port := global.EnvConfig.DC.SQLServer.Port
	dbname := global.EnvConfig.DC.SQLServer.DBName
	// 連線到資料庫
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", username, password, hostname, port, dbname)
	var err error
	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	if err != nil {
		global.Logger.Error("sqlserver 連線失敗", zap.Error(err))
		global.SqlServer = db
		return
	}

	global.SqlServer = db
	global.Logger.Info(
		fmt.Sprintf("sqlserver [%v] connection success", dbname))

}

func WriteToSqlServer(table string, data []map[string]interface{}) error {
	db := global.SqlServer

	// 批次大小

	batchSize := global.EnvConfig.DC.SQLServer.BatchSize

	// 開始批量處理，分批寫入 SQL Server
	tx := db.Begin() // 開啟一個事務
	for i := 0; i < len(data); i += batchSize {
		// 取出當前批次的數據
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		batch := data[i:end]

		// 插入每個批次的數據
		err := tx.Table(table).Create(batch).Error
		if err != nil {
			// 如果插入失敗則回滾事務
			tx.Rollback()
			global.Logger.Error("寫入 SQL Server 失敗", zap.Error(err))
			return err
		}
	}

	// 提交事務
	err := tx.Commit().Error
	if err != nil {
		global.Logger.Error("提交 SQL Server 事務失敗", zap.Error(err))
		return err
	}

	global.Logger.Info("批量寫入 SQL Server 成功", zap.Int("points", len(data)))
	return nil
}

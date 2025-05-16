package dc

import (
	"bimap-zbox/databases"
	"bimap-zbox/global"
	"bimap-zbox/models"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// 處理失敗的 InfluxDB 檔案
func RecoveryInfluxData(fileDir, bucket string) {
	// 讀取目錄下的所有檔案
	files, err := os.ReadDir(fileDir)
	if err != nil {
		global.Logger.Error("讀取 InfluxData 目錄失敗",
			zap.String("dir", fileDir),
			zap.Error(err))
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue // 跳過子目錄
		}

		filename := filepath.Join(fileDir, file.Name())

		// 讀取檔案內容
		content, err := os.ReadFile(filename)
		if err != nil {
			global.Logger.Error("讀取檔案失敗",
				zap.String("file", filename),
				zap.Error(err))
			continue
		}

		var points []models.Point
		err = json.Unmarshal(content, &points)
		if err != nil {
			global.Logger.Error("解析 JSON 失敗",
				zap.String("file", filename),
				zap.Error(err))
			continue
		}

		// 如果檔案中的 points 為空，則跳過
		if len(points) == 0 {
			global.Logger.Info("檔案無數據，跳過處理",
				zap.String("file", filename))
			continue
		}

		// 執行 InfluxDB 的重試操作
		err = databases.WriteToInfluxDB(bucket, points)
		if err != nil {
			global.Logger.Error("回補失敗",
				zap.String("file", filename),
				zap.Error(err))
			continue
		}

		global.Logger.Info("InfluxDB 資料回補成功",
			zap.String("file", filename),
			zap.Int("count", len(points)))

		// 處理完成，移除檔案
		if err := os.Remove(filename); err != nil {
			global.Logger.Error("檔案移除失敗",
				zap.String("file", filename),
				zap.Error(err))
		} else {
			global.Logger.Info("檔案移除成功",
				zap.String("file", filename))
		}
	}
}

// RecoverySqlServerData 處理 SQL Server 的資料回補
func RecoverySqlServerData(fileDir, table string) {
	// 讀取目錄下的所有檔案
	files, err := os.ReadDir(fileDir)
	if err != nil {
		global.Logger.Error("讀取 SqlServerData 目錄失敗",
			zap.String("dir", fileDir),
			zap.Error(err))
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue // 跳過子目錄
		}

		filename := filepath.Join(fileDir, file.Name())

		// 讀取檔案內容
		content, err := os.ReadFile(filename)
		if err != nil {
			global.Logger.Error("讀取檔案失敗",
				zap.String("file", filename),
				zap.Error(err))
			continue
		}

		var records []map[string]interface{}
		err = json.Unmarshal(content, &records)
		if err != nil {
			global.Logger.Error("解析 JSON 失敗",
				zap.String("file", filename),
				zap.Error(err))
			continue
		}

		// 處理每筆記錄的時間格式
		for i := range records {
			if createTime, ok := records[i]["CreateTime"].(string); ok {
				// 解析時間字符串
				t, err := time.Parse(time.RFC3339, createTime)
				if err != nil {
					global.Logger.Error("時間格式解析失敗",
						zap.String("time", createTime),
						zap.Error(err))
					continue
				}
				// 轉換為 SQL Server 可接受的格式 (YYYY-MM-DD HH:mm:ss)
				records[i]["CreateTime"] = t.Format("2006-01-02 15:04:05")
			}
		}

		// 寫入 SQL Server
		err = databases.WriteToSqlServer(table, records)
		if err != nil {
			global.Logger.Error("回補資料寫入 SQL Server 失敗",
				zap.String("file", filename),
				zap.Error(err))
			continue
		}

		global.Logger.Info("SQL Server 資料回補成功",
			zap.String("file", filename),
			zap.Int("count", len(records)))

		// 處理完成，移除檔案
		if err := os.Remove(filename); err != nil {
			global.Logger.Error("檔案移除失敗",
				zap.String("file", filename),
				zap.Error(err))
		} else {
			global.Logger.Info("檔案移除成功",
				zap.String("file", filename))
		}
	}
}

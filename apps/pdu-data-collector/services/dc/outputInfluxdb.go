package dc

import (
	"bimap-zbox/databases"
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bimap-zbox/services"
	"path/filepath"
	"time"

	"github.com/fatih/structs"
	"go.uber.org/zap"
)

func PointsToInfluxDB(points []models.Point) {
	if len(points) == 0 {
		return
	}

	// 批量寫入 InfluxDB
	err := databases.WriteToInfluxDB(global.Envs.InfluxDBBucket.Raw, points)
	if err != nil {
		global.Logger.Error(
			"批量寫入 InfluxDB 失敗",
			zap.Error(err),
		)

		// 如果發送仍然失敗，將數據保存至 JSON 檔案
		filename := filepath.Join(global.Envs.DCRecoveryDir.InfluxDBRaw, time.Now().Format(time.RFC3339Nano)+".json")
		var dataToSave []map[string]interface{}
		for _, point := range points {
			m := structs.Map(point)
			dataToSave = append(dataToSave, m)
		}
		saveErr := services.SaveDataToFile(filename, dataToSave)
		if saveErr != nil {
			global.Logger.Error(
				"保存資料至 InfluxDB JSON 檔案失敗",
				zap.Error(saveErr),
				zap.String("filename", filename),
			)
		} else {
			global.Logger.Info(
				"批量寫入 InfluxDB 失敗，已保存至 JSON 檔案",
				zap.String("filename", filename),
			)
		}
	}
}

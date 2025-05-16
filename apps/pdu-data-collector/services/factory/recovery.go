package factory

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bimap-zbox/services"
	"encoding/json"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// RecoveryFactoryData 處理失敗的 Factory 資料重傳
func RecoveryFactoryData(fileDir string) {
	var endpoint string
	switch fileDir {
	case global.Envs.FactoryRecoveryDir.DCMetrics:
		endpoint = global.EnvConfig.Factory.DCEndpoint + "/metrics"
	case global.Envs.FactoryRecoveryDir.DCEvents:
		endpoint = global.EnvConfig.Factory.DCEndpoint + "/event"
	}
	// 讀取目錄下的所有檔案
	files, err := os.ReadDir(fileDir)
	if err != nil {
		global.Logger.Error("讀取 FactoryData 目錄失敗",
			zap.String("dir", fileDir),
			zap.Error(err))
		return
	}

	// 確認 DC 連線正常
	err = services.CheckDCConnection(global.EnvConfig.Factory.DCEndpoint + "/status")
	if err != nil {
		global.Logger.Error("與 DC 的連線失敗，無法重傳檔案",
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

		// 重試發送數據到 DC
		err = SendPointsToDC(models.MetricsData{Metrics: points}, endpoint)
		if err != nil {
			global.Logger.Error("重傳失敗",
				zap.String("file", filename),
				zap.Error(err))
			continue
		}

		global.Logger.Info("資料重傳成功",
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

package factory

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bimap-zbox/services"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/fatih/structs"
	"go.uber.org/zap"
)

func FactoryMetricsToDC(points []models.Point) {

	err := SendPointsToDC(models.MetricsData{Metrics: points}, global.EnvConfig.Factory.DCEndpoint+"/metrics")
	if err == nil {
		return
	}

	global.Logger.Error("批量發送 DC 失敗", zap.Error(err))

	// 如果發送仍然失敗，將數據保存至 JSON 檔案
	filename := filepath.Join(global.Envs.FactoryRecoveryDir.DCMetrics, time.Now().Format(time.RFC3339Nano)+".json")

	// 將 Point 轉換為 map[string]interface{}
	var dataToSave []map[string]interface{}
	for _, point := range points {
		m := structs.Map(point)
		dataToSave = append(dataToSave, m)
	}

	saveErr := services.SaveDataToFile(filename, dataToSave)
	if saveErr != nil {
		global.Logger.Error(
			"保存 DC 資料至 JSON 檔案失敗",
			zap.Error(saveErr),
			zap.String("filename", filename),
		)
	} else {
		global.Logger.Info(
			"批量資料發送至 DC 失敗，已保存至 JSON 檔案",
			zap.String("filename", filename),
		)
	}
}

// SendMetricsToDC 發送格式化後的 PDU 資料到 DC endpoint，並根據配置進行重試
func SendPointsToDC(points models.MetricsData, endpoint string) error {

	// 將 []models.Point 轉換為 JSON
	data, err := json.Marshal(points)
	if err != nil {
		return fmt.Errorf("PDU 轉換 JSON 失敗: %v", err)
	}

	// 嘗試發送請求
	err = sendRequest(data, endpoint)
	if err != nil {
		return err
	}

	return nil
}

// sendRequest 發送實際的 HTTP 請求
func sendRequest(data []byte, endpoint string) error {

	// 設定 HTTP 客戶端，超時時間由配置控制
	client := &http.Client{
		Timeout: time.Duration(global.EnvConfig.Global.HttpTimeout) * time.Second, // 根據配置設置請求超時時間
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// 設置標頭，例如 Content-Type
	req.Header.Set("Content-Type", "application/json")

	// 發送請求
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 確認狀態碼為 200 OK
	if resp.StatusCode != http.StatusOK {
		return errors.New("DC 回應錯誤，狀態碼：" + resp.Status)
	}

	return nil
}

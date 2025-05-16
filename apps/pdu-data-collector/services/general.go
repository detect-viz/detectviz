package services

import (
	"bimap-zbox/global"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"go.uber.org/zap"
)

func SaveDataToFile(filename string, datas []map[string]interface{}) error {
	// 將資料轉換為 JSON 格式
	jsonData, err := json.Marshal(datas)
	if err != nil {
		return fmt.Errorf("將資料轉換為 JSON 時發生錯誤: %v", err)
	}

	// 檢查檔案是否存在
	var file *os.File
	if _, err := os.Stat(filename); err == nil {
		// 檔案存在，直接追加寫入
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("開啟檔案時發生錯誤: %v", err)
		}
	} else if os.IsNotExist(err) {
		// 檔案不存在，創建新檔案
		file, err = os.Create(filename)
		if err != nil {
			return fmt.Errorf("建立新檔案時發生錯誤: %v", err)
		}
	} else {
		return fmt.Errorf("檢查檔案時發生錯誤: %v", err)
	}
	defer file.Close()

	// 寫入資料到檔案
	if _, err := file.WriteString(string(jsonData) + "\n"); err != nil {
		return fmt.Errorf("寫入 JSON 資料至檔案時發生錯誤: %v", err)
	}

	return nil
}

// line 逐行寫入檔案 Debug 用 	filename := "all.influxlp"
func WriteToFile(filename string, line string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()
	}

	f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0666)
	f.Write([]byte(line + "\n"))
}

func ShortFloat(f float64) float64 {
	a := fmt.Sprintf("%.2f", f)
	n, err := strconv.ParseFloat(a, 64)
	if err != nil {
		global.Logger.Error(
			err.Error(),
			zap.String("ShortFloat", "error"),
		)
		return 0
	}
	return n
}

// CheckDCConnection 確認與 DC 的連線是否正常
func CheckDCConnection(dcEndpoint string) error {
	// 發送簡單的 HTTP GET 請求來檢查連線
	resp, err := http.Get(dcEndpoint)
	if err != nil {
		return fmt.Errorf("與 DC 的連線失敗: %v", err)
	}
	defer resp.Body.Close() // 確保釋放資源

	// 確認回應狀態碼是否為 200 OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("與 DC 的連線失敗，狀態碼: %v", resp.StatusCode)
	}

	return nil
}

func IfNotExistCreateDir(dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			global.Logger.Error(
				"建立目錄失敗",
				zap.String("dirPath", dirPath),
				zap.Error(err),
			)
		}
		global.Logger.Info(
			"成功建立目錄",
			zap.String("dirPath", dirPath),
		)
	}
}

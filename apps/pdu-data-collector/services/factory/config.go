package factory

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var rwMu sync.RWMutex

// 更新配置數據，例如 PDU 名單或測量設定
func UpdateConfig(job models.InitData) {
	rwMu.Lock()
	defer rwMu.Unlock()

	var localFilePath string
	switch job.Name {
	case global.EnvConfig.Factory.InitGlobalData.PDU.Name:
		localFilePath = global.EnvConfig.Factory.InitGlobalData.PDU.File
	case global.EnvConfig.Factory.InitGlobalData.Env.Name:
		localFilePath = global.EnvConfig.Factory.InitGlobalData.Env.File
	case global.EnvConfig.Factory.InitGlobalData.Job.Name:
		localFilePath = global.EnvConfig.Factory.InitGlobalData.Job.File
	case global.EnvConfig.Factory.InitGlobalData.Log.Name:
		localFilePath = global.EnvConfig.Factory.InitGlobalData.Log.File
	}

	dcUrl := fmt.Sprintf("%s/env/%s", global.EnvConfig.Factory.DCEndpoint, global.EnvConfig.Factory.GroupName)

	configData, err := getDCApiToGlobal(job.Name, dcUrl)
	if err != nil {
		global.Logger.Error("API 請求失敗", zap.String("url", dcUrl), zap.Error(err))
		configData, err = readFileToGlobal(localFilePath)
		if err != nil {
			global.Logger.Error("從檔案讀取失敗", zap.Error(err))
			return
		}
		global.Logger.Info("從檔案讀取成功", zap.String("file", localFilePath))
	} else {
		// API 成功，保存到本地檔案
		ext := filepath.Ext(localFilePath)
		switch ext {
		case ".csv":
			if err := writeMapDataToCSV(localFilePath, &configData); err != nil {
				global.Logger.Error("無法將數據寫入 CSV 檔案", zap.Error(err))
				return
			}
		case ".json":
			if err := writeMapDataToJSON(localFilePath, &configData); err != nil {
				global.Logger.Error("無法將數據寫入 JSON 檔案", zap.Error(err))
				return
			}
		case ".yml", ".yaml":
			if err := writeMapDataToYAML(localFilePath, &configData); err != nil {
				global.Logger.Error("無法將數據寫入 YAML 檔案", zap.Error(err))
				return
			}
		default:
			global.Logger.Error("不支援的檔案格式", zap.String("file", localFilePath))
			return
		}
		global.Logger.Info("數據成功寫入", zap.String("file", localFilePath))
	}

	// 根據 jobName 更新全局配置
	switch job.Name {
	case global.EnvConfig.Factory.InitGlobalData.PDU.Name:
		global.PDUList = &configData
	case global.EnvConfig.Factory.InitGlobalData.Env.Name:
		global.Envs = mapToEnvs(configData)
	case global.EnvConfig.Factory.InitGlobalData.Job.Name:
		global.Jobs = mapToJobs(configData)
	case global.EnvConfig.Factory.InitGlobalData.Log.Name:
		global.Logs = mapToLogs(configData)
	}
}

func mapToEnvs(data map[string]map[string]string) *models.Envs {
	Envs := &models.Envs{}

	for section, fields := range data {
		structField := reflect.ValueOf(Envs).Elem().FieldByName(section)
		if !structField.IsValid() || structField.Kind() != reflect.Struct {
			global.Logger.Error("Section not found in Envs", zap.String("section", section))
			continue
		}

		for fieldKey, fieldValue := range fields {
			if fieldKey == "key" { // 跳過不需要的鍵
				continue
			}

			field := structField.FieldByNameFunc(func(name string) bool {
				fieldStruct, found := structField.Type().FieldByName(name)
				return found && fieldStruct.Tag.Get("json") == fieldKey
			})

			if field.IsValid() && field.CanSet() {
				field.SetString(fieldValue)
			} else {
				global.Logger.Error("Field key not found or cannot be set in section", zap.String("fieldKey", fieldKey), zap.String("section", section))
			}
		}
	}

	return Envs
}

func mapToJobs(data map[string]map[string]string) *models.Jobs {
	Jobs := &models.Jobs{}

	for section, fields := range data {
		structField := reflect.ValueOf(Jobs).Elem().FieldByName(section)
		if !structField.IsValid() || structField.Kind() != reflect.Struct {
			global.Logger.Error("Section not found in Jobs", zap.String("section", section))
			continue
		}

		for fieldKey, fieldValue := range fields {
			if fieldKey == "key" { // 跳過不需要的鍵
				continue
			}

			field := structField.FieldByNameFunc(func(name string) bool {
				fieldStruct, found := structField.Type().FieldByName(name)
				return found && fieldStruct.Tag.Get("json") == fieldKey
			})

			if field.IsValid() && field.CanSet() {
				field.SetString(fieldValue)
			} else {
				global.Logger.Error("Field key not found or cannot be set in section", zap.String("fieldKey", fieldKey), zap.String("section", section))
			}
		}
	}

	return Jobs
}

func mapToLogs(data map[string]map[string]string) *models.Logs {
	Logs := &models.Logs{}

	for section, fields := range data {
		structField := reflect.ValueOf(Logs).Elem().FieldByName(section)
		if !structField.IsValid() || structField.Kind() != reflect.Struct {
			global.Logger.Error("Section not found in Logs", zap.String("section", section))
			continue
		}

		for fieldKey, fieldValue := range fields {
			if fieldKey == "key" { // 跳過不需要的鍵
				continue
			}

			field := structField.FieldByNameFunc(func(name string) bool {
				fieldStruct, found := structField.Type().FieldByName(name)
				return found && fieldStruct.Tag.Get("json") == fieldKey
			})

			if field.IsValid() && field.CanSet() {
				field.SetString(fieldValue)
			} else {
				global.Logger.Error("Field key not found or cannot be set in section", zap.String("fieldKey", fieldKey), zap.String("section", section))
			}
		}
	}

	return Logs
}

// getDCApiToGlobal 從 DC API 獲取 PDU 或 ENV 名單
func getDCApiToGlobal(jobName, baseURL string) (map[string]map[string]string, error) {
	// 構建完整的 URL，根據 jobName 加入查詢參數
	url := fmt.Sprintf("%s?name=%s", baseURL, jobName)

	// 發送 GET 請求
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 解析 JSON 響應
	var configData map[string]map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&configData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return configData, nil
}

// readFileToGlobal 從檔案讀取到全局
func readFileToGlobal(filePath string) (map[string]map[string]string, error) {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".csv":
		return readCSVToMap(filePath)
	case ".json":
		return readJSONToMap(filePath)
	case ".yml", ".yaml":
		return readYAMLToMap(filePath)
	default:
		return nil, fmt.Errorf("不支援的檔案格式: %s", ext)
	}
}

// 從 CSV 文件讀取到 map
func readCSVToMap(filePath string) (map[string]map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file does not contain enough data")
	}

	headers := records[0]
	result := make(map[string]map[string]string)

	for i, record := range records {
		if i == 0 {
			continue
		}
		key := record[0]
		result[key] = make(map[string]string)
		for j, header := range headers {
			if j < len(record) && record[j] != "" {
				result[key][header] = record[j]
			}
		}
	}

	return result, nil
}

// 從 JSON 文件讀取到 map
func readJSONToMap(filePath string) (map[string]map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %v", err)
	}
	defer file.Close()

	var result map[string]map[string]string
	if err := json.NewDecoder(file).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON file: %v", err)
	}

	return result, nil
}

// 從 YAML 文件讀取到 map
func readYAMLToMap(filePath string) (map[string]map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open YAML file: %v", err)
	}
	defer file.Close()

	var result map[string]map[string]string
	if err := yaml.NewDecoder(file).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode YAML file: %v", err)
	}

	return result, nil
}

// 將 map 寫入 CSV 檔案
func writeMapDataToCSV(filePath string, mapData *map[string]map[string]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	fieldSet := make(map[string]struct{})
	for _, data := range *mapData {
		for field := range data {
			fieldSet[field] = struct{}{}
		}
	}

	var fields []string
	for field := range fieldSet {
		fields = append(fields, field)
	}

	headers := append([]string{"key"}, fields...)
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers to CSV: %v", err)
	}

	for key, data := range *mapData {
		record := []string{key}
		for _, field := range fields {
			record = append(record, data[field])
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write record to CSV: %v", err)
		}
	}
	return nil
}

// 將 map 寫入 JSON 檔案
func writeMapDataToJSON(filePath string, mapData *map[string]map[string]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mapData); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}
	return nil
}

// 將 map 寫入 YAML 檔案
func writeMapDataToYAML(filePath string, mapData *map[string]map[string]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create YAML file: %v", err)
	}
	defer file.Close()

	if err := yaml.NewEncoder(file).Encode(mapData); err != nil {
		return fmt.Errorf("failed to encode YAML: %v", err)
	}
	return nil
}

// 從全域 pduMap 取得標籤 (讀操作)
func GetTagsByPduKey(key string) map[string]string {

	rwMu.RLock() // 讀鎖
	defer rwMu.RUnlock()

	pduMap := *global.PDUList
	if tags, exists := pduMap[key]; exists {
		return tags
	}
	return nil
}

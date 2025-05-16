package factory

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bimap-zbox/services"
	"strings"

	"golang.org/x/exp/slices"
)

func FormatDeltaPDU(pduKey string, data models.Point) ([]models.Point, error) {

	newTags := make(map[string]string)
	var res []models.Point
	schema := global.Envs.PDUSchema

	// 獲取對應的標籤
	allowTags := strings.Split(global.Envs.FactoryConfig.AllowTags, ",")
	dbTags := GetTagsByPduKey(pduKey)
	if dbTags == nil {
		// 若無標籤，設定預設標籤
		for _, tag := range allowTags {
			newTags[tag] = "unknown"
		}
	} else {
		for key, value := range dbTags {
			data.Tags[key] = value
		}
		// 過濾自定義標籤
		for key := range data.Tags {
			if slices.Contains(allowTags, key) {
				newTags[key] = data.Tags[key]
			}
		}
	}

	// 檢查 current 是否需要補全
	if _, exists := data.Fields["current_L1"]; !exists {
		data.Fields["current_L1"] = data.Fields["current_L1-1"] + data.Fields["current_L1-2"] + data.Fields["current_L1-3"]
		data.Fields["current_L2"] = data.Fields["current_L2-1"] + data.Fields["current_L2-2"] + data.Fields["current_L2-3"]
		data.Fields["current_L3"] = data.Fields["current_L3-1"] + data.Fields["current_L3-2"] + data.Fields["current_L3-3"]
	}

	// 檢查 watt 是否需要補全
	if _, exists := data.Fields["watt_L1"]; !exists {
		data.Fields["watt_L1"] = data.Fields["watt_L1-1"] + data.Fields["watt_L1-2"] + data.Fields["watt_L1-3"]
		data.Fields["watt_L2"] = data.Fields["watt_L2-1"] + data.Fields["watt_L2-2"] + data.Fields["watt_L2-3"]
		data.Fields["watt_L3"] = data.Fields["watt_L3-1"] + data.Fields["watt_L3-2"] + data.Fields["watt_L3-3"]
	}

	// 檢查 energy 是否需要補全
	if _, exists := data.Fields["energy_L1"]; !exists {
		data.Fields["energy_L1"] = data.Fields["energy_L1-1"] + data.Fields["energy_L1-2"] + data.Fields["energy_L1-3"]
		data.Fields["energy_L2"] = data.Fields["energy_L2-1"] + data.Fields["energy_L2-2"] + data.Fields["energy_L2-3"]
		data.Fields["energy_L3"] = data.Fields["energy_L3-1"] + data.Fields["energy_L3-2"] + data.Fields["energy_L3-3"]
	}
	// 處理總和欄位
	if _, exists := data.Fields[schema.TotalCurrentField]; !exists {
		data.Fields[schema.TotalCurrentField] = data.Fields["current_L1"] + data.Fields["current_L2"] + data.Fields["current_L3"]
	}
	if _, exists := data.Fields[schema.TotalWattField]; !exists {
		data.Fields[schema.TotalWattField] = data.Fields["watt_L1"] + data.Fields["watt_L2"] + data.Fields["watt_L3"]
	}
	if _, exists := data.Fields[schema.TotalEnergyField]; !exists {
		data.Fields[schema.TotalEnergyField] = data.Fields["energy_L1"] + data.Fields["energy_L2"] + data.Fields["energy_L3"]
	}

	for key, value := range data.Fields {
		var val float64
		fields := make(map[string]float64)
		tags := make(map[string]string)
		for k, v := range newTags {
			tags[k] = v
		}

		// 單位轉換
		val = services.ShortFloat(value)

		// 自動補全 phase 和 branch
		var phase, branch string
		newFieldName := key // 預設為原始欄位名稱

		if strings.HasSuffix(key, "-1") || strings.HasSuffix(key, "-2") || strings.HasSuffix(key, "-3") {
			branch = strings.Split(key, "_")[1]   // L1-1, L2-1 等等
			phase = strings.Split(branch, "-")[0] // 提取 L1, L2, L3
			tags[schema.PhaseTag] = phase
			tags[schema.BranchTag] = branch

			// 根據 branch 和 phase 設置欄位名稱
			if strings.Contains(key, "current") {
				newFieldName = schema.BranchCurrentField
			} else if strings.Contains(key, "energy") {
				newFieldName = schema.BranchEnergyField
			} else if strings.Contains(key, "watt") {
				newFieldName = schema.BranchWattField
			}
		} else if strings.HasSuffix(key, "L1") || strings.HasSuffix(key, "L2") || strings.HasSuffix(key, "L3") {
			phase = strings.Split(key, "_")[1]
			tags[schema.PhaseTag] = phase

			// 根據 phase 設置欄位名稱
			if strings.Contains(key, "current") {
				newFieldName = schema.PhaseCurrentField
			} else if strings.Contains(key, "voltage") {
				newFieldName = schema.PhaseVoltageField
			} else if strings.Contains(key, "watt") {
				newFieldName = schema.PhaseWattField
			} else if strings.Contains(key, "energy") {
				newFieldName = schema.PhaseEnergyField
			}
		}

		// 設定補全後的欄位值
		fields[newFieldName] = val

		// 創建新的 Point，並將結果加入 res
		res = append(res, models.Point{
			Name:   schema.Measurement,
			Time:   data.Time,
			Tags:   tags,
			Fields: fields,
		})
	}
	// b, _ := json.MarshalIndent(res, "", "\t")
	// os.Stdout.Write(b)
	// fmt.Printf("len(res): %v\n", len(res))
	return res, nil
}

package dc

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"fmt"
	"log"
)

// 從數據庫撈取 Env 資料並轉換為 map[string]map[string]string
func GetEnvsFromDB() (map[string]map[string]string, error) {
	var envRecords []models.Env
	result := make(map[string]map[string]string)
	db := global.Mysql
	// 查詢所有 Env 資料
	if err := db.Find(&envRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch env records: %v", err)
	}

	// 將查詢結果轉換為 map 結構
	for _, record := range envRecords {
		if _, exists := result[record.Type]; !exists {
			result[record.Type] = make(map[string]string)
		}
		result[record.Type][record.Key] = record.Value
	}

	return result, nil
}

// 從數據庫撈取 Job 資料並轉換為 map[string]map[string]string
func GetJobsFromDB() (map[string]map[string]string, error) {
	var envRecords []models.Job
	result := make(map[string]map[string]string)
	db := global.Mysql
	// 查詢所有 Env 資料
	if err := db.Find(&envRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch env records: %v", err)
	}

	// 將查詢結果轉換為 map 結構
	for _, record := range envRecords {
		if _, exists := result[record.Type]; !exists {
			result[record.Type] = make(map[string]string)
		}
		result[record.Type][record.Key] = record.Value
	}

	return result, nil
}

// 從數據庫撈取 Log 資料並轉換為 map[string]map[string]string
func GetLogsFromDB() (map[string]map[string]string, error) {
	var envRecords []models.Log
	result := make(map[string]map[string]string)
	db := global.Mysql
	// 查詢所有 Env 資料
	if err := db.Find(&envRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch env records: %v", err)
	}

	// 將查詢結果轉換為 map 結構
	for _, record := range envRecords {
		if _, exists := result[record.Type]; !exists {
			result[record.Type] = make(map[string]string)
		}
		result[record.Type][record.Key] = record.Value
	}

	return result, nil
}

func GetPDUDataFromDB(groupName string) (map[string]map[string]string, error) {
	db := global.Mysql // 使用你的 MySQL 連接
	result := make(map[string]map[string]string)

	// 1. 根據 groupName 查找 group_zone，並取得 factory 和 phase
	var groupZones []models.GroupZone
	if err := db.Where("group_name = ?", groupName).Find(&groupZones).Error; err != nil {
		log.Fatalf("Failed to query group zones for group name %s: %v", groupName, err)
	}

	// 2. 取得所有 factory 和 phase 配對
	var factoriesPhases []struct {
		Factory string
		Phase   string
	}
	for _, groupZone := range groupZones {
		factoriesPhases = append(factoriesPhases, struct {
			Factory string
			Phase   string
		}{Factory: groupZone.Factory, Phase: groupZone.Phase})
	}

	// 3. 查詢 PDU 設備，根據 factory 和 phase 過濾
	var pduDevices []models.Device
	for _, fp := range factoriesPhases {
		var tempDevices []models.Device
		if err := db.Where("type = ? AND factory = ? AND phase = ?", "PDU", fp.Factory, fp.Phase).Find(&tempDevices).Error; err != nil {
			return nil, fmt.Errorf("failed to query PDU devices: %v", err)
		}
		pduDevices = append(pduDevices, tempDevices...) // 合併到 pduDevices 中
	}

	// 收集所有 PDU ID
	var pduIDs []int
	for _, pdu := range pduDevices {
		pduIDs = append(pduIDs, int(pdu.ID))
	}

	// 4. 查詢 WiFi Client 設備
	wifiClientLinks, wifiClientDevices, err := fetchDeviceLinksAndDevices(pduIDs, "WIFI_CLIENT")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch WiFi Client devices: %v", err)
	}

	// 5. 查詢 Panel 設備
	panelLinks, panelDevices, err := fetchDeviceLinksAndDevices(pduIDs, "PANEL")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Panel devices: %s", err.Error())
	}

	// 6. 處理 PDU 設備
	for _, pdu := range pduDevices {
		var key string

		// 如果 PDU 的協議是 modbus，則處理 WiFi Client 的 IP 和 Port
		if pdu.Protocol == "modbus" {
			if link, exists := wifiClientLinks[int(pdu.ID)]; exists {
				if wifiClientDevice, exists := wifiClientDevices[int(link.ParentDeviceID)]; exists {
					key = fmt.Sprintf("%s:%d", wifiClientDevice.IP, link.ParentDevicePort)
					if result[key] == nil {
						result[key] = make(map[string]string)
					}
					result[key]["wifi_client_ip"] = wifiClientDevice.IP
					result[key]["wifi_client_port"] = fmt.Sprintf("%d", link.ParentDevicePort)
				}
			}
		} else {
			// 如果不是 modbus 協議，則使用 PDU 的 IP 作為 key
			key = pdu.IP
			if result[key] == nil {
				result[key] = make(map[string]string)
			}
			result[key]["ip"] = pdu.IP
		}

		// 查找對應的 Panel 設備
		if link, exists := panelLinks[int(pdu.ID)]; exists {
			if panelDevice, exists := panelDevices[int(link.ParentDeviceID)]; exists {
				result[key]["panel"] = panelDevice.Name
			}
		}

		if pdu.Phase != "" {
			result[key]["phase"] = pdu.Phase
		}

		// 其他 PDU 的欄位資料
		result[key]["factory"] = pdu.Factory
		result[key]["datacenter"] = pdu.Datacenter
		result[key]["room"] = pdu.Room
		result[key]["rack"] = pdu.Rack
		result[key]["side"] = pdu.Side
		result[key]["type"] = pdu.Type
		result[key]["manufacturer"] = pdu.Manufacturer
		result[key]["model"] = pdu.Model
		result[key]["protocol"] = pdu.Protocol
		result[key]["name"] = pdu.Name
	}

	return result, nil
}

// 查詢與 PDU 相關的 DeviceLink 和對應的設備
func fetchDeviceLinksAndDevices(childDeviceIDs []int, parentDeviceType string) (map[int]models.DeviceLink, map[int]models.Device, error) {
	db := global.Mysql
	// 查詢與 PDU 相關的 DeviceLink
	var deviceLinks []models.DeviceLink
	if err := db.Where("child_device_id IN ? AND parent_device_type = ?", childDeviceIDs, parentDeviceType).Find(&deviceLinks).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to query device links for %s: %v", parentDeviceType, err)
	}

	// 收集所有 parent_device_id
	var parentDeviceIDs []int
	for _, link := range deviceLinks {
		parentDeviceIDs = append(parentDeviceIDs, int(link.ParentDeviceID))
	}

	// 查詢對應的 parent 設備
	var parentDevices []models.Device
	if len(parentDeviceIDs) > 0 {
		if err := db.Where("id IN ?", parentDeviceIDs).Find(&parentDevices).Error; err != nil {
			return nil, nil, fmt.Errorf("failed to query parent devices for %s: %v", parentDeviceType, err)
		}
	}

	// 構建 child_device_id -> DeviceLink 的 map
	deviceLinkMap := make(map[int]models.DeviceLink)
	for _, link := range deviceLinks {
		deviceLinkMap[int(link.ChildDeviceID)] = link
	}

	// 構建 parent_device_id -> Device 的 map
	deviceMap := make(map[int]models.Device)
	for _, device := range parentDevices {
		deviceMap[int(device.ID)] = device
	}

	return deviceLinkMap, deviceMap, nil
}

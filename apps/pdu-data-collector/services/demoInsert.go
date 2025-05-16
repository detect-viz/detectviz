package services

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/gorm"
)

// 根據設備詳細資訊查找或創建設備
func findOrCreateDevice(db *gorm.DB, device models.Device) models.Device {
	var existingDevice models.Device
	if err := db.Where("type = ? AND name = ? AND ip = ?", device.Type, device.Name, device.IP).First(&existingDevice).Error; err == nil {
		return existingDevice
	}
	// 創建新設備
	db.Create(&device)
	return device
}

// 解析 CSV 標題列，返回每個欄位的索引位置
func getFieldIndex(headers []string) map[string]int {
	fieldIndex := make(map[string]int)
	for idx, header := range headers {
		fieldIndex[header] = idx
	}
	return fieldIndex
}

func DemoInsert(filename string) {
	// 連接資料庫
	db := global.Mysql

	// 打開 CSV 檔案
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal("Unable to read input file", err)
	}
	defer f.Close()

	// 讀取 CSV 檔案
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV", err)
	}

	// 取得 CSV 標題欄位的索引
	fieldIndex := getFieldIndex(records[0])

	// 從第二行開始讀取（跳過標題）
	for i, row := range records[1:] {
		// 透過欄位名稱取得對應的值
		factory := row[fieldIndex["factory"]]
		datacenter := row[fieldIndex["datacenter"]]
		room := row[fieldIndex["room"]]
		rack := row[fieldIndex["rack"]]
		side := row[fieldIndex["side"]]
		model := row[fieldIndex["model"]]
		ip := row[fieldIndex["ip"]]
		panel := row[fieldIndex["panel"]]
		wifiClientIP := row[fieldIndex["wifi_client_ip"]]
		wifiClientPort, _ := strconv.Atoi(row[fieldIndex["wifi_client_port"]])
		protocol := row[fieldIndex["protocol"]]

		// 插入 PDU 資料
		pdu := models.Device{
			Factory:      factory,
			Datacenter:   datacenter,
			Room:         room,
			Rack:         rack,
			Side:         side,
			Model:        model,
			Protocol:     protocol,
			IP:           ip,
			Type:         "PDU",
			Manufacturer: "Delta",
			Name:         fmt.Sprintf("%s%s%s%s%s", factory, datacenter, room, rack, side),
		}
		pdu = findOrCreateDevice(db, pdu)

		// 插入或查找 Panel
		if panel != "" {
			panelDevice := models.Device{
				Factory:    factory,
				Datacenter: datacenter,
				Room:       room,
				Rack:       "",
				Side:       "",
				Model:      "",
				IP:         "", // Panel 沒有 IP
				Type:       "PANEL",
				Name:       panel,
			}
			panelDevice = findOrCreateDevice(db, panelDevice)

			// 綁定 Panel 和 PDU
			link := models.DeviceLink{
				ParentDeviceID:   panelDevice.ID,
				ChildDeviceID:    pdu.ID,
				ParentDevicePort: 0,
				ParentDeviceType: "PANEL",
			}
			db.Create(&link)
		}

		// 插入或查找 WiFi Client
		if wifiClientIP != "" {
			wifiDevice := models.Device{
				Factory:    factory,
				Datacenter: datacenter,
				Room:       room,
				Rack:       "",
				Side:       "",
				Model:      "",
				IP:         wifiClientIP,
				Type:       "WIFI_CLIENT",
				Name:       "",
			}
			wifiDevice = findOrCreateDevice(db, wifiDevice)

			// 綁定 WiFi Client 和 PDU
			link := models.DeviceLink{
				ParentDeviceID:   wifiDevice.ID,
				ChildDeviceID:    pdu.ID,
				ParentDevicePort: wifiClientPort,
				ParentDeviceType: "WIFI_CLIENT",
			}
			db.Create(&link)
		}

		fmt.Printf("Added PDU [%d]: %s, Panel: %s, WiFiClientIP: %s, WiFiClientPort: %d\n", i, model, panel, wifiClientIP, wifiClientPort)
	}
}

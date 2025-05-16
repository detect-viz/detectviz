package dc

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

func GetRuleListFromDB() ([]models.DeviceRule, error) {
	db := global.Mysql
	rules := []models.DeviceRule{}
	err := db.Table("Rules").Where("Enabled = ?", true).Find(&rules).Error
	if err != nil {
		log.Println("Error querying rules: ", err.Error())
		return rules, err
	}
	return rules, nil
}

func WriteDeviceToMysql(filename string) {
	db := global.Mysql
	// 打開 CSV 文件
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Failed to open CSV file:", err)
		return
	}
	defer file.Close()

	// 讀取 CSV 文件
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Failed to read CSV file:", err)
		return
	}

	// 讀取 CSV 表頭並建立欄位名稱到索引的映射
	header := records[0]
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[col] = i
	}

	var devices []models.Device

	// 跳過第一行標題，並將所有設備新增到一個 slice 中
	for i, record := range records {
		if i == 0 {
			// Skip the header
			continue
		}

		// 創建設備結構
		device := models.Device{
			Factory:      record[columnIndex["factory"]],
			Datacenter:   record[columnIndex["datacenter"]],
			Room:         record[columnIndex["room"]],
			Rack:         record[columnIndex["rack"]],
			Side:         record[columnIndex["side"]],
			Name:         record[columnIndex["name"]],
			Type:         record[columnIndex["type"]],
			Manufacturer: record[columnIndex["manufacturer"]],
			Model:        record[columnIndex["model"]],
			Protocol:     record[columnIndex["protocol"]],
			IP:           record[columnIndex["ip"]],
		}

		devices = append(devices, device)
	}

	// 批量插入，這裡一次插入 100 條資料（根據需求調整 batch size）
	batchSize := global.EnvConfig.DC.SQLServer.BatchSize
	if err := db.CreateInBatches(devices, batchSize).Error; err != nil {
		fmt.Println("Failed to batch insert devices:", err)
	} else {
		fmt.Printf("Inserted %d devices successfully.\n", len(devices))
	}
}

func InsertDeviceLinkToMysql(filename string) {

	db := global.Mysql

	// 開啟 CSV 檔案
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	// 讀取 CSV 檔案內容
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	// 讀取 CSV 表頭並建立欄位名稱到索引的映射
	header := records[0]
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[col] = i
	}

	// 迭代每一行資料
	for _, record := range records[1:] { // 忽略標題
		// 取出 PDU 的欄位
		pduFactory := record[columnIndex["factory"]]
		pduDatacenter := record[columnIndex["datacenter"]]
		pduRoom := record[columnIndex["room"]]
		pduRack := record[columnIndex["rack"]]
		pduSide := record[columnIndex["side"]]

		// 取出 Panel 的欄位
		panelFactory := record[columnIndex["factory"]]
		panelDatacenter := record[columnIndex["datacenter"]]

		panelName := record[columnIndex["panel"]]

		if pduFactory == "" || panelName == "" {
			continue
		}

		// 查找 PDU ID 根據 factory, phase, datacenter, room, rack, side
		var pduDevice models.Device
		result := db.Where("factory = ? AND datacenter = ? AND room = ? AND rack = ? AND side = ?",
			pduFactory, pduDatacenter, pduRoom, pduRack, pduSide).First(&pduDevice)
		if result.Error != nil {
			log.Printf("Failed to find PDU with details %v: %v", pduFactory, result.Error)
			continue
		}

		// 查找 Panel ID 根據 factory, phase, datacenter, room, panel
		var panelDevice models.Device
		result = db.Where("factory = ? AND datacenter = ? AND name = ?",
			panelFactory, panelDatacenter, panelName).First(&panelDevice)
		if result.Error != nil {
			log.Printf("Failed to find Panel with name %s: %v", panelName, result.Error)
			continue
		}

		// 創建 device_link 關係
		deviceLink := models.DeviceLink{
			ParentDeviceID:   panelDevice.ID,
			ChildDeviceID:    pduDevice.ID,
			ParentDeviceType: panelDevice.Type,
		}

		// 將資料寫入 device_links 表
		if err := db.Create(&deviceLink).Error; err != nil {
			log.Printf("Failed to create device link for PDU %v and Panel %s: %v", pduFactory, panelName, err)
		} else {
			fmt.Printf("Successfully linked PDU %v with Panel %s\n", pduFactory, panelName)
		}
	}
}

func ReadWifiDeviceLinkFromCsv(filename string) {
	// 設置資料庫連線
	db := global.Mysql

	// 開啟 CSV 檔案
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	// 讀取 CSV 檔案內容
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	// 讀取 CSV 表頭並建立欄位名稱到索引的映射
	header := records[0]
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[col] = i
	}

	// 迭代每一行資料
	for _, record := range records[1:] { // 忽略標題
		// 取出設備的欄位
		factory := record[columnIndex["factory"]]
		phase := record[columnIndex["phase"]]
		datacenter := record[columnIndex["datacenter"]]
		room := record[columnIndex["room"]]
		rack := record[columnIndex["rack"]]
		side := record[columnIndex["side"]]

		// 取出父設備的欄位
		parentDeviceIP := record[columnIndex["parent_device_ip"]]
		parentDevicePort, _ := strconv.Atoi(record[columnIndex["parent_device_port"]])

		// 查找 Parent Device 根據 parent_device_ip
		var parentDevice models.Device
		result := db.Where("ip = ?", parentDeviceIP).First(&parentDevice)
		if result.Error != nil {
			log.Printf("Failed to find parent device with IP %s: %v", parentDeviceIP, result.Error)
			continue
		}

		// 查找子設備 根據 factory, phase, datacenter, room, rack, side
		var childDevice models.Device
		result = db.Where("factory = ? AND phase = ? AND datacenter = ? AND room = ? AND rack = ? AND side = ?",
			factory, phase, datacenter, room, rack, side).First(&childDevice)
		if result.Error != nil {
			log.Printf("Failed to find child device with details %v: %v", factory, result.Error)
			continue
		}

		// 創建 device_link 關係
		deviceLink := models.DeviceLink{
			ParentDeviceID:   parentDevice.ID,
			ChildDeviceID:    childDevice.ID,
			ParentDevicePort: parentDevicePort,
			ParentDeviceType: parentDevice.Type,
		}

		// 將資料寫入 device_links 表
		if err := db.Create(&deviceLink).Error; err != nil {
			log.Printf("Failed to create device link for child device %v and parent device %s: %v", childDevice.ID, parentDeviceIP, err)
		} else {
			fmt.Printf("Successfully linked child device %v with parent device %s\n", childDevice.ID, parentDeviceIP)
		}
	}
}

func ReadPDUFromCsv(filename string) {
	// 設置資料庫連線
	db := global.Mysql
	columnName := global.Envs.ColumnName
	deviceType := global.Envs.DeviceType

	// 開啟 CSV 檔案
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	// 讀取 CSV 檔案內容
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	// 讀取 CSV 表頭並建立欄位名稱到索引的映射
	header := records[0]
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[col] = i
	}

	// 迭代每一行資料
	for _, record := range records[1:] { // 忽略標題
		// 取出設備的欄位
		factory := record[columnIndex[columnName.Factory]]
		datacenter := record[columnIndex[columnName.Datacenter]]
		room := record[columnIndex[columnName.Room]]
		rack := record[columnIndex[columnName.Rack]]
		side := record[columnIndex[columnName.Side]]
		ip := record[columnIndex[columnName.IP]]
		name := record[columnIndex[columnName.Name]]
		manufacturer := record[columnIndex[columnName.Manufacturer]]
		model := record[columnIndex[columnName.Model]]
		protocol := record[columnIndex[columnName.Protocol]]
		panelName := record[columnIndex[columnName.Panel]]
		wifiClientIP := record[columnIndex[columnName.WifiClientIP]]
		wifiClientPort, _ := strconv.Atoi(record[columnIndex[columnName.WifiClientPort]])

		var panelDevice models.Device

		if panelName != "" {
			// 檢查 Panel 是否存在，不存在則插入
			if err := db.First(&panelDevice, "name = ?", panelName).Error; err != nil {
				// 若找不到，插入新的 Panel
				panelDevice = models.Device{
					Factory:    factory,
					Datacenter: datacenter,
					Room:       room,
					Name:       panelName,
					Type:       deviceType.Panel,
				}
				if err := db.Create(&panelDevice).Error; err != nil {
					log.Fatalf("插入 Panel 失敗: %v", err)
				}
			}
		}

		var wifiClient models.Device
		if protocol == "modbus" {
			// Step 1: 查找或創建 WiFi Client
			err := db.Where("type = ? AND name = ?", deviceType.WifiClient, wifiClientIP).First(&wifiClient).Error
			if err == gorm.ErrRecordNotFound {
				wifiClient = models.Device{
					Factory:    factory,
					Datacenter: datacenter,
					Room:       room,
					Name:       wifiClientIP,
					IP:         wifiClientIP,
					Type:       deviceType.WifiClient,
				}
				if err := db.Create(&wifiClient).Error; err != nil {
					log.Fatalf("Failed to create WiFi Client: %v", err)
				}
			}

		}

		// 檢查 PDU 是否存在，不存在則插入
		var pdu models.Device
		if err := db.First(&pdu, "name = ?", name).Error; err != nil {
			// 若找不到，插入新的 PDU
			pdu = models.Device{
				Factory:      factory,
				Datacenter:   datacenter,
				Manufacturer: manufacturer,
				Model:        model,
				Room:         room,
				Rack:         rack,
				IP:           ip,
				Name:         name,
				Side:         side,
				Protocol:     protocol,
				Type:         deviceType.PDU,
			}
			if err := db.Create(&pdu).Error; err != nil {
				log.Fatalf("插入 PDU 失敗: %v", err)
			}
		}

		if panelName != "" {
			// 檢查 DeviceLink 是否存在，不存在則綁定
			var deviceLink models.DeviceLink
			if err := db.First(&deviceLink, "parent_device_id = ? AND child_device_id = ?", panelDevice.ID, pdu.ID).Error; err != nil {
				// 若找不到，進行綁定
				deviceLink = models.DeviceLink{
					ParentDeviceID: panelDevice.ID,
					ChildDeviceID:  pdu.ID,
					//ParentDevicePort: 1, // 可以根據需要更改
					ParentDeviceType: deviceType.Panel,
				}
				if err := db.Create(&deviceLink).Error; err != nil {
					log.Fatalf("panel 和 PDU 綁定失敗: %v", err)
				}
			}
		}

		if protocol == "modbus" {
			// Step 3: 檢查並綁定設備
			var deviceLink models.DeviceLink
			err = db.Where("parent_device_id = ? AND child_device_id = ?", wifiClient.ID, pdu.ID).First(&deviceLink).Error
			if err == gorm.ErrRecordNotFound {
				deviceLink = models.DeviceLink{
					ParentDeviceID:   wifiClient.ID,
					ChildDeviceID:    pdu.ID,
					ParentDevicePort: wifiClientPort,
					ParentDeviceType: deviceType.WifiClient,
				}
				if err := db.Create(&deviceLink).Error; err != nil {
					log.Fatalf("Failed to create Device Link: %v", err)
				}
			}
		}
	}
}

func ReadPanelFromCsv(filename string) {
	// 設置資料庫連線
	db := global.Mysql
	columnName := global.Envs.ColumnName
	deviceType := global.Envs.DeviceType
	// 開啟 CSV 檔案
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	// 讀取 CSV 檔案內容
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	// 讀取 CSV 表頭並建立欄位名稱到索引的映射
	header := records[0]
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[col] = i
	}

	// 迭代每一行資料
	for _, record := range records[1:] { // 忽略標題
		// 取出設備的欄位
		factory := record[columnIndex[columnName.Factory]]
		datacenter := record[columnIndex[columnName.Datacenter]]
		room := record[columnIndex[columnName.Room]]

		panelName := record[columnIndex[columnName.Name]]
		plcIP := record[columnIndex[columnName.PLCIP]]

		// 步驟 1: 檢查 PLC 是否存在
		var plcDevice models.Device
		err = db.Where("ip = ? AND type = ?", plcIP, deviceType.PLC).First(&plcDevice).Error
		if err == gorm.ErrRecordNotFound {
			// 插入 PLC 記錄
			plcDevice = models.Device{
				Factory:      factory,
				Datacenter:   datacenter,
				Room:         room,
				Manufacturer: "SIMATIC",
				Model:        "S7-1200",
				Name:         plcIP, // 使用 IP 作為 PLC 的名稱
				Type:         deviceType.PLC,
				IP:           plcIP,
			}
			if err := db.Create(&plcDevice).Error; err != nil {
				log.Printf("無法插入 PLC 記錄: %v", err)
				continue
			}
			log.Printf("插入 PLC 記錄: %v", plcIP)
		}

		// 步驟 2: 檢查 Panel 是否存在
		var panelDevice models.Device
		err = db.Where("name = ? AND type = ?", panelName, deviceType.Panel).First(&panelDevice).Error
		if err == gorm.ErrRecordNotFound {
			// 插入 Panel 記錄
			panelDevice = models.Device{
				Factory:    factory,
				Datacenter: datacenter,
				Room:       room,
				Name:       panelName,
				Type:       deviceType.Panel,
			}
			if err := db.Create(&panelDevice).Error; err != nil {
				log.Printf("無法插入 Panel 記錄: %v", err)
				continue
			}
			log.Printf("插入 Panel 記錄: %v", panelName)
		}

		// 步驟 3: 確認 PLC 和 Panel 的綁定
		var deviceLink models.DeviceLink
		err = db.Where("parent_device_id = ? AND child_device_id = ?", plcDevice.ID, panelDevice.ID).First(&deviceLink).Error
		if err == gorm.ErrRecordNotFound {
			// 插入綁定記錄
			deviceLink = models.DeviceLink{
				ParentDeviceID:   plcDevice.ID,
				ChildDeviceID:    panelDevice.ID,
				ParentDeviceType: deviceType.PLC,
			}
			if err := db.Create(&deviceLink).Error; err != nil {
				log.Printf("無法插入 DeviceLink 記錄: %v", err)
				continue
			}
			log.Printf("插入 DeviceLink 記錄: PLC %v 綁定 Panel %v", plcIP, panelName)
		}
	}
}

package services

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

// PDU 結構體
type PDU struct {
	Factory    string
	DataCenter string
	Room       string
	Rack       string
	Side       string
	Model      string
	IP         string
	Panel      string
	WiFiClient string
	WiFiPort   int
	Protocol   string
}

// 追踪已經分配過的 IP 地址，避免重複
var allocatedIPs = make(map[string]bool)

// 生成 WiFi client IP，確保不重複且避開 0 和 255
func generateUniqueWiFiClientIP() string {
	var ip string
	for {
		thirdOctet := rand.Intn(254) + 1  // 生成範圍為 1 到 254
		fourthOctet := rand.Intn(254) + 1 // 生成範圍為 1 到 254，避開 0 和 255
		ip = fmt.Sprintf("172.12.%d.%d", thirdOctet, fourthOctet)
		if !allocatedIPs[ip] {
			allocatedIPs[ip] = true
			break
		}
	}
	return ip
}

// 生成 snmp PDU 的 IP 地址，確保不重複且避開 0 和 255
func generateUniqueSNMPIP(startIP *int) string {
	var ip string
	for {
		thirdOctet := rand.Intn(254) + 1    // 生成範圍為 1 到 254
		fourthOctet := (*startIP % 254) + 1 // 避開 0 和 255
		ip = fmt.Sprintf("172.10.%d.%d", thirdOctet, fourthOctet)
		if !allocatedIPs[ip] {
			allocatedIPs[ip] = true
			*startIP++
			break
		}
	}
	return ip
}

// 生成 WiFi client 的端口號，L 對應 8001，R 對應 8002
func generateWiFiClientPort(side string) int {
	if side == "L" {
		return 8001
	}
	return 8002
}

// 生成 panel 名稱，格式為 U/N + 數字 + 大寫字母
func generatePanelName() string {
	letter := string(rand.Intn(26) + 'A')
	number := rand.Intn(9) + 1
	prefix := "U"
	if rand.Intn(2) == 0 {
		prefix = "N"
	}
	return fmt.Sprintf("%s%d%s", prefix, number, letter)
}

// 生成每個房間內的 PDU 清單，確保 rack 兩側 PDU 的 protocol 一致
func generatePDUData(factory, dc, room string, startIP *int, csvWriter *csv.Writer) {
	for rackNum := 1; rackNum <= 70; rackNum++ {
		// Rack 編號，格式為兩位數，如 01, 02 等
		rackID := fmt.Sprintf("A%02d", rackNum)

		// 生成 WiFi Client IP
		wifiClientIP := generateUniqueWiFiClientIP()

		// 隨機分配 Rack 兩側共同使用的 PDU 型號和協議
		var model, protocol string
		if rand.Intn(2) == 0 {
			model = "PDUE428"
			protocol = "snmp"
		} else {
			if rand.Intn(2) == 0 {
				model = "PDU4425"
			} else {
				model = "PDU1315"
			}
			protocol = "modbus"
		}

		// 為每個 rack 生成 L 和 R 兩側的 PDU，兩側必須使用相同的 protocol
		for _, side := range []string{"L", "R"} {
			// 分配 Panel
			panel := generatePanelName()

			// 根據 protocol 決定 IP 地址和 WiFiClientPort
			var ip string
			var wifiClientPort int
			if protocol == "snmp" {
				ip = generateUniqueSNMPIP(startIP)
				wifiClientPort = generateWiFiClientPort(side)
			} else {
				wifiClientPort = generateWiFiClientPort(side)
			}

			// 構建 PDU 資料
			pdu := PDU{
				Factory:    factory,
				DataCenter: dc,
				Room:       room,
				Rack:       rackID,
				Side:       side,
				Model:      model,
				IP:         ip,
				Panel:      panel,
				WiFiClient: wifiClientIP,
				WiFiPort:   wifiClientPort,
				Protocol:   protocol,
			}

			// 根據協議生成對應的 CSV 行
			var csvRow []string
			if protocol == "modbus" {
				csvRow = []string{
					pdu.Factory, pdu.DataCenter, pdu.Room, pdu.Rack, pdu.Side,
					pdu.Model, "", pdu.Panel, pdu.WiFiClient, strconv.Itoa(pdu.WiFiPort), pdu.Protocol,
				}
			} else {
				csvRow = []string{
					pdu.Factory, pdu.DataCenter, pdu.Room, pdu.Rack, pdu.Side,
					pdu.Model, pdu.IP, pdu.Panel, pdu.WiFiClient, strconv.Itoa(pdu.WiFiPort), pdu.Protocol,
				}
			}

			// 寫入 CSV
			err := csvWriter.Write(csvRow)
			if err != nil {
				fmt.Println("寫入 CSV 時出錯:", err)
			}
		}
	}
}

// 創建示例數據並寫入 CSV
func CreateDemoData(filename string) {

	// 創建 CSV 檔案
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("創建 CSV 檔案時出錯:", err)
		return
	}
	defer file.Close()

	// 創建 CSV 寫入器
	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	// 寫入標題行
	header := []string{"factory", "datacenter", "room", "rack", "side", "model", "ip", "panel", "wifi_client_ip", "wifi_client_port", "protocol"}
	err = csvWriter.Write(header)
	if err != nil {
		fmt.Println("寫入標題行時出錯:", err)
		return
	}

	// 配置地點和房間
	locations := []struct {
		Factory    string
		DataCenter string
		Rooms      []string
	}{
		{"F12P10", "DC1", []string{"R1"}},
	}

	// 開始 IP 的模擬
	startIP := 1

	// 生成並寫入所有地點的數據
	for _, loc := range locations {
		for _, room := range loc.Rooms {
			generatePDUData(loc.Factory, loc.DataCenter, room, &startIP, csvWriter)
		}
	}

	fmt.Printf("PDU 數據已成功寫入 %s\n", filename)
}

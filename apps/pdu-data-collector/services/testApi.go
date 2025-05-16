package services

import (
	"bimap-zbox/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func RunApi() {
	http.HandleFunc("/test_endpoint", func(w http.ResponseWriter, r *http.Request) {
		// 檢查是否為 POST 請求
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// 讀取請求的 Body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		// 將收到的資料打印出來
		fmt.Println("========= Received data =========")
		fmt.Println(string(body))

		// 將 Body 解析到 MetricsData 結構體
		var metricsData models.MetricsData
		err = json.Unmarshal(body, &metricsData)
		if err != nil {
			http.Error(w, "Failed to unmarshal request body", http.StatusBadRequest)
			return
		}

		// 打印解碼後的 MetricsData
		fmt.Printf("========= Received Metrics Data [%v]=========\n", len(metricsData.Metrics))
		for i, metric := range metricsData.Metrics {
			fmt.Printf("第 %v 筆\n", i)
			b, _ := json.MarshalIndent(metric, "", "\t")
			os.Stdout.Write(b)
		}

		// 回應 Telegraf
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Data received successfully"))
	})

	// 啟動 HTTP 伺服器
	fmt.Println("Starting server at http://localhost:8008/test_endpoint")
	err := http.ListenAndServe(":8008", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func PostToWebdis() {
	url := "http://10.99.1.124:7379"
	data := `XADD/line_protocol_stream/*/data/E-COAT-03-00,SubEQPID=E-COAT-03-01,UnitID=00,Type=ProcessData ReportDT="20230926_11:14:47.818",LotId="LotID_003",CellId="cellId_002",RouteId="RouteId002",OpeNo="OpeNo_002",WorkOrder="Order_003",ProductId="Prod02",ContainerId="Tray002",Position="Pos02",RecipeId="20230824",Factory_Number="EMT3",Device_Number="TCS-0935",Production_type="TEST01",Testresult_-OK_NG-="0",MES_Warning_Messages="0",Air_supply_fan_frequency="0",Air_supply_fan_temperature="0",exhaust_fan_frequency="0",Line_speed="0",Base_material_thickness="15",CoreTrail="173.3",Remaining_length_setting=1430083618,Auto-start_Trail=1078823307,Auto_Splice_Trail=1722091209,UW_A_axis_unwinding_trail=1422423669,UW_A_axis_unwinding_length=2093200710,UW_Baxis_unwinding_Trail=379803223,UW_Baxis_unwinding_length=622815426,Unwinding_Splice_NIP_ROLL_ratio=1841500700,Unwinding_Splice_NIP_ROLL_single_speed=459655331,Unwindingtension=1027270242,EPC_CPCsetting_value_-substrate_width-=333718065,A-side_CPCcorrection=723514756,SubstrateWidth_PT=1758421974,Foil_average_PT=2051160067,A-side_Coating_tension-setting_value_-=1662251906,A-side_Coating_ing_sheets=1311030753,A_side_pattern_type_selection=1375577711,A-side__slurry_tank_level=956346581,A-side_Coating_pump_rpm=407071947,A-side_DIE_GAP-DRIVE_SIDE--RIGHT-=1813828848,A-side_Coating_settingGAP-RIGHT-=1408199303,A-side_Coating_adjustment_GAP-RIGHT-=1399306292,A-side_DIE_GAP-OPERATOR--LEFT-=1604475726,A-side_Coating_settingGAP-LEFT-=835662396,A-side_Coating_adjustment_GAP-LEFT-=601710477,A-side_Coating_pressure=1121852337,A-side_return_pressure=893567248,A-side_filter_pressure=534551396,Pattern1_A-side_SynchronousCoating_Length=803556125,Pattern1_A-side_asynchronousCoating_Length=1515005698,Pattern1_A-side_UnsynchronizedCoating_Length=260551235,Pattern1_A-side_asynchronous_uncoating_Length=1393343741,Pattern1_A-side_offset_Length=204705100,Pattern1_A-side_Coating_Length_correction=1899975591,Pattern1_A-side_uncoating_Lengthcorrection=1252249579,Pattern1_A-side_return_valve_opening_timing=1031962552,Pattern1_A-side_return_valve_closing_timing=868197313,Pattern1_A-side_Coating_valve_opening_timing=294573977,Pattern1_A-side_Coating_valve_closing_timing=1615405528,Pattern2_A-side_SynchronousCoating_Length=1399780862,Pattern2_A-side_asynchronousCoating_Length=18866506,Pattern2_A-side_UnsynchronizedCoating_Length=1994438196,Pattern2_A-side_asynchronousuncoating_Length=177541247,Pattern2_A-side_offsetLength=1655863819,Pattern2_A-side_Coating_Lengthcorrection=517152109,Pattern2_A-side_uncoating_Lengthcorrection=1295042452,Pattern2_A-side_return_valve_opening_timing=1788037251,Pattern2_A-side_return_valve_closing_timing=373706715,Pattern2_A-side_Coating_valve_opening_timing=815580011,Pattern2_A-side_Coating_valve_closing_timing=264043844,Pattern3_A-side_SynchronousCoating_Length=1203483490,Pattern3_A-side_asynchronousCoating_Length=1441002879,Pattern3_A-side_UnsynchronizedCoating_Length=742068100,Pattern3_A-side_asynchronousuncoating_Length=595333381,Pattern3_A-side_offsetLength=2147120852,Pattern3_A-side_Coating_Lengthcorrection=30777326,Pattern3_A-side_uncoating_Lengthcorrection=1621831228,Pattern3_A-side_return_valve_opening_timing=886428813,Pattern3_A-side_return_valve_closing_timing=820713192,Pattern3_A-side_Coating_valve_opening_timing=971348373,Pattern3_A-side_Coating_valve_closing_timing=1633719622,A-side_PUMP_rpm=88264030,A-side_PUMP_correction=1037944575,DRYERstrap_mode_=1092133280,1F1Z_temperature=766719007,1F1Z_Circulating_Fan_frequency=1087857971,1F1Z_gasconcentration=518809656,1F1Z_temperature_upper_limit_=2005930030,1F1Z_temperature_lower_limit=798910488,1F1Z_1-4rollperipheral_speed=630289354,1F2Z_temperature=442833440,1F2Z_Circulating_Fan_frequency=1081003734,1F2Z_gasconcentration=1488424700,1F2Z_temperature_upper_limit_=1154774732,1F2Z_temperature_lower_limit=1794962342,1F2Z_1-4roll_peripheral_speed=1561244754,1F3Z_temperature=1221765045,1F3Z_Circulating_Fan_frequency=2096095554,1F3Z_gasconcentration=318510287,1F3Z_1-4roll_peripheral_speed=1454151672,1F3Z_temperature_upper_limit_=1481298733,1F3Z_temperature_lower_limit=519860533,1F4Z_temperature=1866455032,1F4Z_Circulating_Fan_frequency=539512281,1F4Z_gasconcentration=311522208,1F4Z_temperature_upper_limit_=967032003,1F4Z_temperature_lower_limit=651275641 1695726887000000000`

	quotedData := strconv.Quote(data)
	cleanedData := strings.Trim(quotedData, `"`)
	// 構建 HTTP POST 請求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(cleanedData)))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// 設置標頭
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 執行請求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// 讀取回應
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// 輸出狀態碼和回應內容
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n", body)
}

func GetFromWebdis() {
	// Webdis 伺服器的 URL
	baseURL := "http://10.99.1.124:7379"

	timestamp := time.Now().Local().UnixNano()

	// Redis 指令及參數（已經 URL 編碼以避免轉義符號）
	data := `E-COAT-03-00,SubEQPID=E-COAT-03-01,UnitID=00,Type=ProcessData ReportDT="20230926_11:14:47.818",LotId="LotID_003",CellId="cellId_002",RouteId="RouteId002",OpeNo="OpeNo_002",WorkOrder="Order_003",ProductId="Prod02",ContainerId="Tray002",Position="Pos02",RecipeId="20230824",Factory_Number="EMT3",Device_Number="TCS-0935",Production_type="TEST01",Testresult_-OK_NG-="0",MES_Warning_Messages="0",Air_supply_fan_frequency="0",Air_supply_fan_temperature="0",exhaust_fan_frequency="0",Line_speed="0",Base_material_thickness="15",CoreTrail="173.3",Remaining_length_setting=1430083618,Auto-start_Trail=1078823307,Auto_Splice_Trail=1722091209,UW_A_axis_unwinding_trail=1422423669,UW_A_axis_unwinding_length=2093200710,UW_Baxis_unwinding_Trail=379803223,UW_Baxis_unwinding_length=622815426,Unwinding_Splice_NIP_ROLL_ratio=1841500700,Unwinding_Splice_NIP_ROLL_single_speed=459655331,Unwindingtension=1027270242,EPC_CPCsetting_value_-substrate_width-=333718065,A-side_CPCcorrection=723514756,SubstrateWidth_PT=1758421974,Foil_average_PT=2051160067,A-side_Coating_tension-setting_value_-=1662251906,A-side_Coating_ing_sheets=1311030753,A_side_pattern_type_selection=1375577711,A-side__slurry_tank_level=956346581,A-side_Coating_pump_rpm=407071947,A-side_DIE_GAP-DRIVE_SIDE--RIGHT-=1813828848,A-side_Coating_settingGAP-RIGHT-=1408199303,A-side_Coating_adjustment_GAP-RIGHT-=1399306292,A-side_DIE_GAP-OPERATOR--LEFT-=1604475726,A-side_Coating_settingGAP-LEFT-=835662396,A-side_Coating_adjustment_GAP-LEFT-=601710477,A-side_Coating_pressure=1121852337,A-side_return_pressure=893567248,A-side_filter_pressure=534551396,Pattern1_A-side_SynchronousCoating_Length=803556125,Pattern1_A-side_asynchronousCoating_Length=1515005698,Pattern1_A-side_UnsynchronizedCoating_Length=260551235,Pattern1_A-side_asynchronous_uncoating_Length=1393343741,Pattern1_A-side_offset_Length=204705100,Pattern1_A-side_Coating_Length_correction=1899975591,Pattern1_A-side_uncoating_Lengthcorrection=1252249579,Pattern1_A-side_return_valve_opening_timing=1031962552,Pattern1_A-side_return_valve_closing_timing=868197313,Pattern1_A-side_Coating_valve_opening_timing=294573977,Pattern1_A-side_Coating_valve_closing_timing=1615405528,Pattern2_A-side_SynchronousCoating_Length=1399780862,Pattern2_A-side_asynchronousCoating_Length=18866506,Pattern2_A-side_UnsynchronizedCoating_Length=1994438196,Pattern2_A-side_asynchronousuncoating_Length=177541247,Pattern2_A-side_offsetLength=1655863819,Pattern2_A-side_Coating_Lengthcorrection=517152109,Pattern2_A-side_uncoating_Lengthcorrection=1295042452,Pattern2_A-side_return_valve_opening_timing=1788037251,Pattern2_A-side_return_valve_closing_timing=373706715,Pattern2_A-side_Coating_valve_opening_timing=815580011,Pattern2_A-side_Coating_valve_closing_timing=264043844,Pattern3_A-side_SynchronousCoating_Length=1203483490,Pattern3_A-side_asynchronousCoating_Length=1441002879,Pattern3_A-side_UnsynchronizedCoating_Length=742068100,Pattern3_A-side_asynchronousuncoating_Length=595333381,Pattern3_A-side_offsetLength=2147120852,Pattern3_A-side_Coating_Lengthcorrection=30777326,Pattern3_A-side_uncoating_Lengthcorrection=1621831228,Pattern3_A-side_return_valve_opening_timing=886428813,Pattern3_A-side_return_valve_closing_timing=820713192,Pattern3_A-side_Coating_valve_opening_timing=971348373,Pattern3_A-side_Coating_valve_closing_timing=1633719622,A-side_PUMP_rpm=88264030,A-side_PUMP_correction=1037944575,DRYERstrap_mode_=1092133280,1F1Z_temperature=766719007,1F1Z_Circulating_Fan_frequency=1087857971,1F1Z_gasconcentration=518809656,1F1Z_temperature_upper_limit_=2005930030,1F1Z_temperature_lower_limit=798910488,1F1Z_1-4rollperipheral_speed=630289354,1F2Z_temperature=442833440,1F2Z_Circulating_Fan_frequency=1081003734,1F2Z_gasconcentration=1488424700,1F2Z_temperature_upper_limit_=1154774732,1F2Z_temperature_lower_limit=1794962342,1F2Z_1-4roll_peripheral_speed=1561244754,1F3Z_temperature=1221765045,1F3Z_Circulating_Fan_frequency=2096095554,1F3Z_gasconcentration=318510287,1F3Z_1-4roll_peripheral_speed=1454151672,1F3Z_temperature_upper_limit_=1481298733,1F3Z_temperature_lower_limit=519860533,1F4Z_temperature=1866455032,1F4Z_Circulating_Fan_frequency=539512281,1F4Z_gasconcentration=311522208,1F4Z_temperature_upper_limit_=967032003,1F4Z_temperature_lower_limit=651275641 ` + strconv.Itoa(int(timestamp))

	// 構建最終請求的 URL
	fullURL := fmt.Sprintf("%s/XADD/line_protocol_stream/*/data/%s", baseURL, data)
	fullURL = strings.ReplaceAll(fullURL, `\"`, `"`)
	// 發送 GET 請求
	resp, err := http.Get(fullURL)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	// 顯示狀態碼
	fmt.Printf("Status Code: %d\n", resp.StatusCode)

	// 讀取並顯示回應內容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	fmt.Printf("Response Body: %s\n", string(body))

}



```bash
auto_depoly/
├── main.sh                 # ✅ 主腳本入口
├── config.sh               # 全域變數與參數集中管理
├── backup/
│   ├── backup_influxdb.sh  # 備份 InfluxDB
│   ├── backup_telegraf.sh  # 備份 Telegraf 設定檔
│   ├── restore_influxdb.sh # 還原 InfluxDB
│   └── restore_telegraf.sh # 還原 Telegraf 設定檔
├── scanner/  
│   ├── scan_snmp.sh            # 掃描 SNMP 設備（scanner 模組）
│   └── scan_modbus.sh          # 掃描 MODBUS 設備（scanner 模組）
├── util/                   # 共用函式模組
│   ├── utils.sh
├── /etc/telegraf.d/        # 📁 正式部署產生的 conf（輸出物）
│   └── ...conf             # 用於 Telegraf 啟動目錄
├── /etc/viot/              # 📁 設備定義與掃描參考資料
│   ├── scan_pdu.csv        # 設備辨識（品牌、型號、OID）
│   └── ip_range.csv        # IP 區段與地區對應
├── /usr/share/telegraf/templates/  # 📁 部署模板，版本管理良好
│   └── V1/                 # 版本目錄
│       ├── PDU-DELTA-PDUE428_snmp.conf.tmpl
│       ├── PDU-VERTIV-6PS56_snmp.conf.tmpl
│       ├── PDU-DELTA-PDU1315_modbus.conf.tmpl
│       ├── PDU-DELTA-PDU1315_modbus.conf.tmpl
│       └── IOTCLIENT-LIENEO-WCM421_snmp.conf.tmpl
├── /var/lib/viot/          # 📁 儲存所有部署過的設備紀錄
│   └── registry_pdu.csv    # 部署後的 metadata
└── /var/log/viot/          # 📁 執行過程記錄，可加上清除機制
    └── 2025-04-10.log
```


我會繼續修改其他函數，添加 `func_name` 變量並更新日誌消息。讓我修改剩餘的函數：


```shell script:auto_depoly/usr/lib/viot/scripts/main.sh
# Function to convert IP to integer
ip_to_int() {
    func_name="ip_to_int"
    local a b c d
    IFS=. read -r a b c d <<< "$1"
    echo $((a * 256 ** 3 + b * 256 ** 2 + c * 256 + d))
}

# Function to convert integer to IP
int_to_ip() {
    func_name="int_to_ip"
    local num=$1
    echo "$((num >> 24 & 255)).$((num >> 16 & 255)).$((num >> 8 & 255)).$((num & 255))"
}

# Function to extract IP from IP:PORT format
extract_ip() {
    func_name="extract_ip"
    local ip_port=$1
    echo "${ip_port%:*}"
}

# 生成 PDU 名稱
# 參數: factory, phase, dc, room, rack, side
# 返回: 格式化的 PDU 名稱，例如: F1P1DC1R1B11PL
generate_pdu_name() {
    func_name="generate_pdu_name"
    local factory="$1"
    local phase="$2"
    local dc="$3"
    local room="$4"
    local rack="$5"
    local side="$6"

    echo "${factory}${phase}${dc}${room}${rack}P${side}"
}

# 驗證 rack 和 side 的格式
# 參數: rack, side
# 返回: 0 表示格式正確，1 表示格式錯誤
validate_rack_side() {
    func_name="validate_rack_side"
    local rack="$1"
    local side="$2"

    if [[ -z "$rack" || "$rack" == "$DEFAULT_RACK_VALUE" || ! "$rack" =~ $REGEX_RACK ]]; then
        echo "[WARN]-${func_name}- rack 格式錯誤，應為大寫字母開頭 + 數字，例如 B11"
        return 1
    fi

    if [[ -z "$side" || "$side" == "$DEFAULT_SIDE_VALUE" || ! "$side" =~ $REGEX_SIDE ]]; then
        echo "[WARN]-${func_name}- side 格式錯誤，必須為 'L' 或 'R'"
        return 1
    fi

    return 0
}

# 驗證 panel 的格式
# 參數: panel
# 返回: 0 表示格式正確，1 表示格式錯誤
# 格式要求: 開頭必須是 U 或 N，中間是數字，結尾是大寫字母
# 例如: U6F, U7G, U7D, U5D, N4B
validate_panel() {
    func_name="validate_panel"
    local panel="$1"

    if [[ -z "$panel" || "$panel" == "$DEFAULT_PANEL_VALUE" ]]; then
        echo "[WARN]-${func_name}- panel 為空或未知值"
        return 1
    fi

    if [[ ! "$panel" =~ $REGEX_PANEL ]]; then
        echo "[WARN]-${func_name}- panel 格式錯誤，應為 U 或 N 開頭，中間是數字，結尾是大寫字母，例如: U6F, U7G, U7D, U5D, N4B"
        return 1
    fi

    return 0
}

# 獲取當前日期
# 返回: 格式化的日期字串，例如: 2024-03-21
get_current_date() {
    func_name="get_current_date"
    date +"%Y-%m-%d"
}

# 獲取當前時間
# 返回: 格式化的時間字串，例如: 2024-03-21 14:30:00
get_current_time() {
    func_name="get_current_time"
    date +"%Y-%m-%d %H:%M:%S"
}

# 確保 registry 檔案有正確的欄位標題
# 參數: file - registry 檔案路徑
# 如果檔案不存在或標題不正確，會創建新檔案或更新標題
ensure_registry_fields() {
    func_name="ensure_registry_fields"
    local file="$1"  # 使用傳入的完整路徑
    local expected_header="$REGISTRY_FIELDS"

    # 確保目錄存在
    local dir=$(dirname "$file")
    mkdir -p "$dir"

    if [[ ! -f "$file" || "$(head -n 1 "$file")" != "$expected_header" ]]; then
        echo "$expected_header" > "$file.tmp"
        if [[ -f "$file" ]]; then
            tail -n +2 "$file" >> "$file.tmp"
        fi
        mv "$file.tmp" "$file"
    fi
}

# 獲取 SNMP OID 的值
# 參數: oid - SNMP OID
#       ip - PDU IP 地址
#       community - SNMP community
#       snmp_version - SNMP 版本
# 返回: OID 的值，如果獲取失敗則返回錯誤狀態
get_oid_value() {
    func_name="get_oid_value"
    local oid="$1"
    local ip="$2"
    local community="$3"
    local snmp_version="$4"
    
    # 檢查參數
    if [[ -z "$oid" || -z "$ip" || -z "$community" || -z "$snmp_version" ]]; then
        echo "[ERROR]-${func_name}- Missing parameters for get_oid_value: oid=$oid, ip=$ip, community=$community, snmp_version=$snmp_version"
        return 1
    fi
    
    # 執行 SNMP 查詢
    local raw_value="$(snmpget -v $snmp_version -c $community $ip $oid 2>/dev/null)"
    
    # 檢查是否成功獲取值
    if [[ $? -ne 0 || -z "$raw_value" ]]; then
        echo "[WARN]-${func_name}- Failed to get value for OID $oid from $ip"
        return 1
    fi
    
    # 處理不同格式的 SNMP 回應
    if [[ "$raw_value" =~ $SNMP_STRING_PATTERN ]]; then
        echo "${BASH_REMATCH[1]}" | tr -d '[:space:]'
    elif [[ "$raw_value" =~ $SNMP_HEX_STRING_PATTERN ]]; then
        echo "${BASH_REMATCH[1]}" | tr -d '[:space:]'
    elif [[ "$raw_value" =~ $SNMP_INTEGER_PATTERN ]]; then
        echo "${BASH_REMATCH[1]}" | tr -d '[:space:]'
    elif [[ "$raw_value" =~ $SNMP_COUNTER32_PATTERN ]]; then
        echo "${BASH_REMATCH[1]}" | tr -d '[:space:]'
    elif [[ "$raw_value" =~ $SNMP_COUNTER64_PATTERN ]]; then
        echo "${BASH_REMATCH[1]}" | tr -d '[:space:]'
    elif [[ "$raw_value" =~ $SNMP_GAUGE32_PATTERN ]]; then
        echo "${BASH_REMATCH[1]}" | tr -d '[:space:]'
    elif [[ "$raw_value" =~ $SNMP_TIMETICKS_PATTERN ]]; then
        echo "${BASH_REMATCH[1]}" | tr -d '[:space:]'
    elif [[ "$raw_value" =~ $SNMP_OID_PATTERN ]]; then
        echo "${BASH_REMATCH[1]}" | tr -d '[:space:]'
    else
        echo "$raw_value" | sed -E 's/^.*=\s*(.*)$/\1/' | tr -d '[:space:]'
    fi
}

# 讀取 PDU 標籤
# 功能: 從 tag_pdu.csv 讀取 PDU 標籤資訊
read_pdu_tags() {
    func_name="read_pdu_tags"
    local temp_file
    temp_file=$(mktemp -p /tmp)
    if [[ $? -ne 0 ]]; then
        echo "[ERROR]-${func_name}- Failed to create temporary file for PDU tags"
        return 1
    fi

    if [[ ! -f "$TAG_PDU_CSV" ]]; then
        echo "[ERROR]-${func_name}- tag_pdu.csv not found at $TAG_PDU_CSV"
        rm -f "$temp_file"
        return 1
    fi

    # 建立 IP 到標籤的映射
    declare -A pdu_tags
    while IFS=',' read -r ip rack side panel; do
        [[ -z "$ip" || "$ip" =~ ^$DEFAULT_UNKNOW_VALUE ]] && continue
        clean_ip=$(extract_ip "$ip")
        [[ -z "$panel" ]] && panel=$DEFAULT_PANEL_VALUE
        pdu_tags["$clean_ip"]="$rack,$side,$panel"
        echo "[DEBUG]-${func_name}- Added tag for IP $clean_ip: rack=$rack, side=$side, panel=$panel"
    done < <(tail -n +2 "$TAG_PDU_CSV")

    if [[ ${#pdu_tags[@]} -eq 0 ]]; then
        echo "[WARN]-${func_name}- No valid PDU tags found in tag_pdu.csv"
        rm -f "$temp_file"
        return 1
    fi

    # 將標籤映射寫入臨時文件
    for ip in "${!pdu_tags[@]}"; do
        echo "$ip|${pdu_tags[$ip]}" >> "$temp_file"
    done

    echo "$temp_file"
}

# 初始化掃描環境
# 功能: 創建必要的臨時文件和目錄
init_scan_env() {
    func_name="init_scan_env"
    # 確保輸出目錄存在
    mkdir -p "$OUTPUT_DIR"
    
    # 確保 registry_pdu.csv 存在並有正確的標題
    ensure_registry_fields "$REGISTRY_PDU_CSV"
    
    # 建立臨時文件來存儲已存在的 IP
    TEMP_EXISTING_IPS=$(mktemp -p /tmp)
    if [[ $? -ne 0 ]]; then
        echo "[ERROR]-${func_name}- Failed to create temporary file for existing IPs"
        return 1
    fi

    # 提取已存在的 IP 到臨時文件
    if [[ -f "$REGISTRY_PDU_CSV" ]]; then
        echo "[DEBUG]-${func_name}- Loading existing IPs from $REGISTRY_PDU_CSV"
        tail -n +2 "$REGISTRY_PDU_CSV" | awk -F',' "{print \$$REGISTRY_IP_KEY}" | sort | uniq > "$TEMP_EXISTING_IPS"
        echo "[DEBUG]-${func_name}- Found $(wc -l < "$TEMP_EXISTING_IPS") existing IPs"
    else
        echo "[WARN]-${func_name}- registry_pdu.csv not found, starting with empty IP list"
        touch "$TEMP_EXISTING_IPS"
    fi
}

# 通用 SNMP 掃描函式，包含 Gateway 偵測並觸發 Modbus 掃描
scan_snmp() {
    func_name="scan_snmp"
    if ! init_scan_env; then
        echo "[ERROR]-${func_name}- Failed to initialize scan environment"
        return 1
    fi

    # 創建一個新的臨時文件來存儲設備配置
    local device_configs
    device_configs=$(mktemp -p /tmp)
    if [[ $? -ne 0 ]]; then
        echo "[ERROR]-${func_name}- Failed to create temporary file for device configs"
        return 1
    fi

    # 讀取設備配置
    if ! read_configs > "$device_configs"; then
        echo "[ERROR]-${func_name}- Failed to read device configs"
        rm -f "$device_configs"
        return 1
    fi

    # 創建一個新的臨時文件來存儲 PDU 標籤
    local pdu_tags
    pdu_tags=$(mktemp -p /tmp)
    if [[ $? -ne 0 ]]; then
        echo "[ERROR]-${func_name}- Failed to create temporary file for PDU tags"
        rm -f "$device_configs"
        return 1
    fi

    # 讀取 PDU 標籤
    if ! read_pdu_tags > "$pdu_tags"; then
        echo "[ERROR]-${func_name}- Failed to read PDU tags"
        rm -f "$device_configs" "$pdu_tags"
        return 1
    fi

    # 創建一個新的臨時文件來存儲 registry
    local temp_registry
    temp_registry=$(mktemp -p /tmp)
    if [[ $? -ne 0 ]]; then
        echo "[ERROR]-${func_name}- Failed to create temporary file for registry"
        rm -f "$device_configs" "$pdu_tags"
        return 1
    fi

    # 讀取 IP 範圍
    if [[ ! -f "$IP_RANGE_CSV" ]]; then
        echo "[ERROR]-${func_name}- ip_range.csv not found at $IP_RANGE_CSV"
        rm -f "$device_configs" "$pdu_tags" "$temp_registry"
        return 1
    fi

    while IFS=',' read -r factory phase dc room ip_start ip_end; do
        [[ "$ip_start" =~ ^# || -z "$ip_start" ]] && continue
        [[ ! "$ip_start" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]] && continue
        [[ ! "$ip_end" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]] && continue

        echo "[DEBUG]-${func_name}- Processing: factory=$factory, phase=$phase, dc=$dc, room=$room, ip_start=$ip_start, ip_end=$ip_end"

        local start=$(ip_to_int "$ip_start")
        local end=$(ip_to_int "$ip_end")
        for ((ip=$start; ip<=$end; ip++)); do
            local current_ip=$(int_to_ip "$ip")
            if ! check_device_match "$current_ip" "$device_configs" "$pdu_tags" "$temp_registry"; then
                echo "[WARN]-${func_name}- Failed to check device match for $current_ip"
            fi
        done
    done < "$IP_RANGE_CSV"

    if [[ -s "$temp_registry" ]]; then
        echo "[INFO]-${func_name}- Adding new records to registry_pdu.csv"
        cat "$temp_registry" >> "$REGISTRY_PDU_CSV"
        echo "[DONE]-${func_name}- Updated registry_pdu.csv with new PDU information"
    else
        echo "[INFO]-${func_name}- No new records to add to registry_pdu.csv"
    fi

    # 清理臨時文件
    rm -f "$device_configs" "$pdu_tags" "$temp_registry" "$TEMP_EXISTING_IPS"
    echo "[DONE]-${func_name}- All configs written to $OUTPUT_DIR"
}

scan_modbus_gateway() {
    func_name="scan_modbus_gateway"
    local gateway_ip="$1"
    local found_pdu=0

    echo "[INFO]-${func_name}- Scanning gateway: $gateway_ip"

    for port in $DEFAULT_MODBUS_PORT $((DEFAULT_MODBUS_PORT + 1)); do
        for slave in $DEFAULT_MODBUS_SLAVE $((DEFAULT_MODBUS_SLAVE + 1)); do
            # 嘗試抓取型號
            model_info=$("$MODUBSGET_BIN" -controller "tcp://$gateway_ip:$port" \
                -transmission-mode "$MODBUS_TRANSMISSION_MODE" -slave "$slave" \
                -type input -address "$MODBUS_MODEL_ADDRESS" -length "$MODBUS_MODEL_LENGTH" 2>/dev/null)

            if [[ "$model_info" =~ $MODBUS_PDU_MODELS ]]; then
                model=$(echo "$model_info" | awk '{print $1}')
                version=$(echo "$model_info" | awk '{print $2}')
                echo "[MATCH]-${func_name}- Found $model at $gateway_ip:$port slave $slave (version: $version)"

                # 抓序號
                serial=$("$MODUBSGET_BIN" -controller "tcp://$gateway_ip:$port" \
                    -transmission-mode "$MODBUS_TRANSMISSION_MODE" -slave "$slave" \
                    -type holding -address "$MODBUS_SERIAL_ADDRESS" -length "$MODBUS_SERIAL_LENGTH" 2>/dev/null)

                # 找 tag（rack/side/panel）
                if [[ -f "$TAG_PDU_CSV" ]]; then
                    tag_line=$(grep "^$gateway_ip," "$TAG_PDU_CSV")
                    rack=$(echo "$tag_line" | cut -d',' -f2)
                    side=$(echo "$tag_line" | cut -d',' -f3)
                    panel=$(echo "$tag_line" | cut -d',' -f4)
                fi

                # fallback
                rack=${rack:-$DEFAULT_RACK_VALUE}
                side=${side:-$DEFAULT_SIDE_VALUE}
                panel=${panel:-$DEFAULT_PANEL_VALUE}

                # 產生 ip_key
                ip_key="$gateway_ip:$port:$slave"

                # 模板名稱
                template_name="${MODBUS_TEMPLATE_PREFIX}${model}${MODBUS_TEMPLATE_SUFFIX}"
                template_version="$MODBUS_TEMPLATE_VERSION"

                # 產生 config
                generate_config "$gateway_ip" "$ip_key" "DELTA" "$model" "$template_name" "$template_version" "" "$port" "" "" "" "" "$rack" "$side" "$panel"

                # 加入 registry_pdu.csv
                current_time=$(get_current_time)
                echo "#,#,#,#,$rack,$side,?,$panel,$ip_key,$gateway_ip,DELTA,$model,$version,$serial,#,modbus,$MODBUS_TRANSMISSION_MODE,,${current_time},${current_time}" >> "$REGISTRY_PDU_CSV"

                found_pdu=1
            fi
        done
    done

    # 如果有找到 PDU，記錄 Gateway
    if [[ "$found_pdu" -eq 1 ]]; then
        echo "[INFO]-${func_name}- Registering gateway: $gateway_ip"
        echo "$gateway_ip,$GATEWAY_MODEL,$(get_current_time)" >> "$REGISTRY_GATEWAY_CSV"
    else
        echo "[INFO]-${func_name}- No PDU found behind gateway: $gateway_ip"
    fi
}

redeploy_one() {
    func_name="redeploy_one"
    local ip_key="$1"
    local rack="$2"
    local side="$3"
    local panel="$4"

    local record
    record=$(awk -F',' -v key="$ip_key" '$9 == key { print }' "$REGISTRY_PDU_CSV")

    if [[ -z "$record" ]]; then
        echo "[ERROR]-${func_name}- ip_key not found: $ip_key"
        return 1
    fi

    IFS=',' read -r factory phase dc room _ _ _ _ _ ip brand model version serial_number snmp_engine_id protocol snmp_version community create_at update_at <<< "$record"

    local port="${ip_key##*:}"
    local template_name="PDU-${brand}-${model}_${protocol}.conf"
    local template_version="$version"

    generate_config "$ip" "$ip_key" "$brand" "$model" "$template_name" "$template_version" "$community" "$port" "$snmp_version" "" "" "" "$rack" "$side" "$panel"
}
```


## 函數
1. 基本工具函數：
   - `ip_to_int`
   - `int_to_ip`
   - `extract_ip`
   - `generate_pdu_name`
   - `validate_rack_side`
   - `validate_panel`
   - `get_current_date`
   - `get_current_time`
   - `ensure_registry_fields`
   - `get_oid_value`

2. 主要功能函數：
   - `read_pdu_tags`
   - `init_scan_env`
   - `scan_snmp`
   - `scan_modbus_gateway`
   - `redeploy_one`

   
bash -x main.sh --mode scan-snmp

bash -x main.sh --mode scan-modbus --ip 10.1.249.34


flowchart TD
  A[Start main.sh] --> B{--mode}
  B -->|scan-snmp| C[call scan_snmp]
  B -->|scan-modbus| D[call scan_modbus_gateway(ip)]
  B -->|redeploy| E[call redeploy_one(ip_key, rack, side, panel)]

  %% scan-snmp
  C --> C1[init_scan_env]
  C --> C2[read_configs(scan-snmp)]
  C --> C3[read_pdu_tags()]
  C --> C4{foreach IP in ip_range.csv}
  C4 --> C5[check_device_match()]
  C5 -->|SNMP matched| C6{model == WCM-421?}
  C6 -->|Yes| C7[call scan_modbus_gateway(ip)]
  C6 -->|No| C8[process_matched_device()]

  %% scan-modbus
  D --> D1[read_configs(scan-modbus)]
  D --> D2[read_pdu_tags()]
  D --> D3[lookup registry_gateway.csv]
  D --> D4[check_device_match(mode=scan-modbus)]
  D4 --> D5{port/slave match}
  D5 --> D6[modbusget → match]
  D6 --> D7[generate_config()]
  D7 --> D8[append registry_pdu.csv]

  %% redeploy
  E --> E1[lookup registry_pdu.csv]
  E1 --> E2[generate_config()]

  %% shared
  C8 --> F1[generate_config()]
  F1 --> F2[update_registry() → temp file]
  C --> C9[append temp_registry → registry_pdu.csv]

./main.sh --mode scan-snmp --factory F1 --phase P1 --dc DC1 --room R1 --ip-start 192.168.1.1 --ip-end 192.168.1.254


  ./main.sh --mode scan --factory F12 --phase P7 --dc DC1 --room R1 --ip-start 10.1.249.34 --ip-end 10.1.249.34 


  ./main.sh --mode scan-snmp --ip-start <起始IP> --ip-end <結束IP> [--factory <factory>] [--phase <phase>] [--dc <dc>] [--room <room>]


ENV_FILE=./etc/default/.env ./main.sh --mode scan --factory F12 --phase P7 --dc DC1 --room R1 --ip-start 10.1.249.34 --ip-end 10.1.249.34

auto_depoly/
├── README.md
├── env.sh
├── generate.sh
├── main.sh
├── scanner.sh
├── log.sh
├── modbusget
├── modbusget.sh
├── snmpget.sh
└── utils.sh
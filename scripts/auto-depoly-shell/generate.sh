#!/bin/bash
SHELL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SHELL_DIR/log.sh"
source "$SHELL_DIR/utils.sh"


# 生成 PDU 名稱
generate_pdu_name() {
    func_name="generate_pdu_name"
    local factory="$1"
    local phase="$2"
    local dc="$3"
    local room="$4"
    local rack="$5"
    local side="$6"
    # 使用命名模板變數 PDU_NAME_FORMAT
    local pdu_name
    eval "pdu_name=\"${PDU_NAME_FORMAT}\""

    log_info "$func_name" "生成 PDU 名稱: $pdu_name"
    echo "$pdu_name"
}

# 獲取當前日期
get_current_date() {
    func_name="get_current_date"
    local current_date=$(date +"%Y-%m-%d")
    log_debug "$func_name" "當前日期：$current_date"
    echo "$current_date"
}

# 獲取當前時間
get_current_time() {
    func_name="get_current_time"
    local current_time=$(date +"%Y-%m-%d %H:%M:%S")
    log_debug "$func_name" "當前時間：$current_time"
    echo "$current_time"
}

# 根據欄位名稱取得在 REGISTRY_COLUMNS 中的索引（1-based）
get_registry_index() {
    local field="$1"
    local index=1
    for col in $REGISTRY_COLUMNS; do
        if [[ "$col" == "$field" ]]; then
            echo "$index"
            return 0
        fi
        ((index++))
    done
    return 1
}

# 確保 registry 檔案有正確的欄位標題
ensure_registry_fields() {
    func_name="ensure_registry_fields"
    local file="$1"  
    # 將 REGISTRY_COLUMNS 中的空格替換為逗號
    expected_header=$(echo "$REGISTRY_COLUMNS" | tr ' ' ',')

    log_info "$func_name" "檢查 registry 檔案：$file"

    local dir=$(dirname "$file")
    mkdir -p "$dir"

    if [[ ! -f "$file" || "$(head -n 1 "$file")" != "$expected_header" ]]; then
        log_warn "$func_name" "檔案不存在或標題不正確，將創建或更新檔案"
        echo "$expected_header" > "$file.tmp"
        if [[ -f "$file" ]]; then
            tail -n +2 "$file" >> "$file.tmp"
        fi
        mv "$file.tmp" "$file"
        log_info "$func_name" "已更新檔案標題"
    else
        log_info "$func_name" "檔案標題正確"
    fi
}

# 更新 registry 記錄
update_registry_record() {
    local ip_key="$1"
    local registry_line="$2"
    local func_name="update_registry_record"
    local now
    now=$(get_current_time)

    # 確保 registry 文件存在且標題正確
    ensure_registry_fields "$REGISTRY_CSV"

    # 根據 ip_key 判斷記錄是否存在
    if awk -F',' -v key="$ip_key" '$9 == key { exit 1 }' "$REGISTRY_CSV"; then
        # 新增記錄：新增 create_at 與 update_at 欄位均為 now
        registry_line="${registry_line},${now},${now}"
        log_info "$func_name" "新增記錄到 registry，ip_key: $ip_key，create_at 與 update_at 均設為 $now"
        if ! echo "$registry_line" >> "$REGISTRY_CSV"; then
            log_error "$func_name" "寫入 registry 失敗，ip_key: $ip_key"
        fi
    else
        # 更新記錄：保留原 create_at，更新 update_at 為 now
        local orig_create_at
        orig_create_at=$(awk -F',' -v key="$ip_key" '($9 == key){print $24}' "$REGISTRY_CSV")
        if [[ -z "$orig_create_at" ]]; then
            orig_create_at="$now"
        fi
        registry_line="${registry_line},${orig_create_at},${now}"
        log_info "$func_name" "更新 registry 中現有記錄，ip_key: $ip_key，保留 create_at 為 $orig_create_at，update_at 更新為 $now"
        if ! awk -F',' -v key="$ip_key" -v new_line="$registry_line" 'BEGIN {OFS=","} $9 == key {print new_line; next} {print}' "$REGISTRY_CSV" > "$REGISTRY_CSV.tmp" || ! mv "$REGISTRY_CSV.tmp" "$REGISTRY_CSV"; then
            log_error "$func_name" "更新 registry 失敗，ip_key: $ip_key"
        fi
    fi
}

# 組裝 registry_line
build_registry_line() {
    local device_type="$1"
    local factory="$2"
    local phase="$3"
    local dc="$4"
    local room="$5"
    local rack="$6"
    local side="$7"
    local name="$8"
    local ip_key="$9"
    local ip="${10}"
    local port="${11}"
    local slave_id="${12}"
    local brand="${13}"
    local model="${14}"
    local version="${15}"
    local serial_number="${16}"
    local snmp_engine_id="${17}"
    local protocol="${18}"
    local interval="${19}"
    local template_version="${20}"
    local snmp_version="${21}"
    local snmp_community="${22}"

    echo "$device_type,$factory,$phase,$dc,$room,$rack,$side,$name,$ip_key,$ip,$port,$slave_id,$brand,$model,$version,$serial_number,$snmp_engine_id,$protocol,$interval,$template_version,$snmp_version,$snmp_community"
}

# 生成設備配置
# REGISTRY_COLUMNS="device_type,factory,phase,dc,room,rack,side,name,ip_key,ip,port,slave_id,brand,model,version,serial_number,snmp_engine_id,protocol,interval,template_version,snmp_version,snmp_community,create_at,update_at"
generate_config() {
    func_name="generate_config"
    local device_type="${1:-$DEFAULT_UNKNOWN_VALUE}"
    local factory="${2:-$DEFAULT_UNKNOWN_VALUE}"
    local phase="${3:-$DEFAULT_UNKNOWN_VALUE}"
    local dc="${4:-$DEFAULT_UNKNOWN_VALUE}"
    local room="${5:-$DEFAULT_UNKNOWN_VALUE}"
    local ip_key="${6}"
    local ip="${7}"
    local port="${8:-$DEFAULT_PORT}"
    local slave_id="${9:-$DEFAULT_SLAVE_ID}"
    local brand="${10}"
    local model="${11}"
    local version="${12:-$DEFAULT_UNKNOWN_VALUE}"
    local serial_number="${13:-$DEFAULT_UNKNOWN_VALUE}"
    local snmp_engine_id="${14:-$DEFAULT_UNKNOWN_VALUE}"
    local protocol="${15:-$PROTOCOL_MODBUS_TAG}"
    local interval="${16:-$DEFAULT_INTERVAL}"
    local template_version="${17:-$DEFAULT_UNKNOWN_VALUE}"
    local snmp_version="${18:-$DEFAULT_SNMP_VERSION}"
    local snmp_community="${19:-$DEFAULT_SNMP_COMMUNITY}"
    local template_name="${20:-$DEFAULT_UNKNOWN_VALUE}"
    local snmp_mibs_path="${21:-$DEFAULT_SNMP_MIBS_PATH}"
    local timeout="${22:-$DEFAULT_TIMEOUT}"
    local retries="${23:-$DEFAULT_RETRIES}"

    log_debug "$func_name" "[2]調試：生成設備配置: ip_key=$ip_key, ip=$ip, port=$port, slave_id=$slave_id, brand=$brand, model=$model, version=$version, serial_number=$serial_number, snmp_engine_id=$snmp_engine_id, protocol=$protocol, interval=$interval, template_version=$template_version, snmp_version=$snmp_version, snmp_community=$snmp_community, template_name=$template_name, snmp_mibs_path=$snmp_mibs_path, timeout=$timeout, retries=$retries"

    # 檢查必要的參數
    if [[ -z "$ip" || -z "$ip_key" || -z "$brand" || -z "$model" || -z "$template_name" || -z "$template_version" ]]; then
        log_error "[${func_name}] 錯誤：缺少必要的參數" >&2
        return 1
    fi

    # 讀取 tag 設備配置
    local tag_line
    tag_line=$(read_tag_configs "$ip_key")
    if [[ -n "$tag_line" ]]; then
        IFS=',' read -r $TAG_COLUMNS <<< "$tag_line"
        log_debug "[${func_name}] 讀取 tag 設備配置: rack=$rack, side=$side" >&2
    else
        log_debug "[${func_name}] 未找到 tag 設備配置，使用預設值" >&2
        local rack="$DEFAULT_UNKNOWN_VALUE"
        local side="$DEFAULT_UNKNOWN_VALUE"
    fi

    local name
    if [[ "$device_type" == "$PDU_TYPE_NAME" ]]; then
        name=$(generate_pdu_name "$factory" "$phase" "$dc" "$room" "$rack" "$side")
    else
        name="$factory$phase-$dc-$room_$device_type"
    fi

    # 檢查模板文件是否存在
    local template_file="$TEMPLATE_DIR/$template_version/$template_name"
    if [[ ! -f "$template_file" ]]; then
        log_error "[${func_name}] 錯誤：找不到模板文件 $template_file" >&2
        return 1
    fi

    # 創建輸出目錄
    mkdir -p "$OUTPUT_DIR"

    # 生成配置文件
    local output_file="$OUTPUT_DIR/${name}_${brand}-${model}_${ip_key}_${template_version}.conf"
    log_info "[${func_name}] 正在生成配置文件：$output_file" >&2

    # 使用模板生成配置
    if ! sed -e "s|{{ip}}|$ip|g" \
            -e "s|{{port}}|$port|g" \
            -e "s|{{slave_id}}|$slave_id|g" \
            -e "s|{{brand}}|$brand|g" \
            -e "s|{{model}}|$model|g" \
            -e "s|{{version}}|$version|g" \
            -e "s|{{serial_number}}|$serial_number|g" \
            -e "s|{{device_type}}|$device_type|g" \
            -e "s|{{snmp_community}}|$snmp_community|g" \
            -e "s|{{snmp_version}}|$snmp_version|g" \
            -e "s|{{snmp_mibs_path}}|$snmp_mibs_path|g" \
            -e "s|{{interval}}|$interval|g" \
            -e "s|{{timeout}}|$timeout|g" \
            -e "s|{{retries}}|$retries|g" \
            -e "s|{{factory}}|$factory|g" \
            -e "s|{{phase}}|$phase|g" \
            -e "s|{{dc}}|$dc|g" \
            -e "s|{{room}}|$room|g" \
            -e "s|{{rack}}|$rack|g" \
            -e "s|{{side}}|$side|g" \
            -e "s|{{pause_between_requests}}|$SET_REPLACE_PAUSE_BETWEEN_REQUESTS|g" \
            -e "s|{{pause_after_connect}}|$SET_REPLACE_PAUSE_AFTER_CONNECT|g" \
            -e "s|{{busy_retries_wait}}|$SET_REPLACE_BUSY_RETRIES_WAIT|g" \
            "$template_file" > "$output_file"; then
        log_error "[${func_name}] 錯誤：生成配置文件失敗" >&2
        return 1
    fi

    # 更新 registry
    local registry_line
    registry_line=$(build_registry_line "$device_type" "$factory" "$phase" "$dc" "$room" "$rack" "$side" "$name" "$ip_key" "$ip" "$port" "$slave_id" "$brand" "$model" "$version" "$serial_number" "$snmp_engine_id" "$protocol" "$interval" "$template_version" "$snmp_version" "$snmp_community")
    update_registry_record "$ip_key" "$registry_line"

    log_info "[${func_name}] 成功生成配置文件並更新 registry" >&2
    return 0
}

# 讀取設備配置
# 功能: 從 scan_device.csv 根據 mode 過濾並讀取設備配置資訊
read_scan_configs() {
    echo "=================== 讀取設備配置 ===================" >&2
    local func_name="read_scan_configs"
    local mode="$1" 
    if [[ -z "$mode" ]]; then
        log_error "[ERROR]-${func_name}- 必須提供模式參數。" >&2
        return 1
    fi

    local temp_file
    temp_file=$(mktemp -p "$TMP_DIR")
    if [[ $? -ne 0 ]]; then
        log_error "[ERROR]-${func_name}- 無法創建臨時文件以存儲設備配置" >&2
        return 1
    fi

    # 統一使用 SCAN_DEVICE_CSV
    local config_file="$SCAN_DEVICE_CSV"
    log_debug "[DEBUG]-${func_name}- 從 $config_file 讀取 $mode 設備配置" >&2

    log_debug "[DEBUG]-${func_name}- 當前目錄: $(pwd)" >&2
    log_debug "[DEBUG]-${func_name}- 目錄內容: $(ls -la "$(dirname "$config_file")")" >&2

    if [[ ! -f "$config_file" ]]; then
        log_error "]-${func_name}- 在 $config_file 找不到配置文件" >&2
        rm -f "$temp_file"
        return 1
    fi

    log_debug "[DEBUG]-${func_name}- 文件存在，檢查權限: $(ls -la "$config_file")" >&2
    log_debug "[DEBUG]-${func_name}- 文件內容:" >&2
    cat "$config_file" >&2

    # 讀取設備配置
    # match_field for Modbus: address:type[:length]
    local line_number=0
    local expected_protocol
    if [[ "$mode" == "$PROTOCOL_SNMP_TAG" ]]; then
        expected_protocol="$PROTOCOL_SNMP_TAG"
    elif [[ "$mode" == "$PROTOCOL_MODBUS_TAG" ]]; then
        expected_protocol="$PROTOCOL_MODBUS_TAG"
    else
        log_error "[ERROR]-${func_name}- 指定的模式無效: $mode" >&2
        rm -f "$temp_file"
        return 1
    fi

    log_debug "[DEBUG]-${func_name}- 過濾協議: $expected_protocol" >&2
    echo $SCAN_COLUMNS >&2
    while IFS=',' read -r $SCAN_COLUMNS; do
    
        line_number=$((line_number + 1))
   
        log_debug "[DEBUG]-${func_name}- 處理第 $line_number 行: 協議=$protocol, 設備類型=$device_type, 品牌=$brand, 型號=$model, 啟用狀態=$enabled" >&2

        # 跳過空行或標題行
        if [[ -z "$protocol" || "$protocol" == "protocol" || "$protocol" == "#" ]]; then
            log_debug "[DEBUG]-${func_name}- 跳過標題或空行" >&2
            continue
        fi

        # 檢查必要的欄位
        if [[ -z "$brand" || -z "$model" || -z "$device_type" ]]; then
            log_warn "[WARN]-${func_name}- 跳過第 $line_number 行的無效配置: 缺少品牌、型號或設備類型" >&2
            continue
        fi

        # 過濾: 檢查 protocol 和 enabled
        if [[ "$protocol" != "$expected_protocol" ]]; then
             log_debug "[DEBUG]-${func_name}- 跳過第 $line_number 行: 協議不匹配 (預期 $expected_protocol, 實際 $protocol)" >&2
             continue
        fi
        if [[ "${enabled:-$DEFAULT_ENABLED}" != "1" ]]; then
            log_debug "[DEBUG]-${func_name}- 跳過第 $line_number 行: 設備未啟用" >&2
            continue
        fi

        # 設置通用預設值 (只對已過濾的行)
        enabled=${enabled:-$DEFAULT_ENABLED}
        interval=${interval:-$DEFAULT_INTERVAL}
        timeout=${timeout:-$DEFAULT_TIMEOUT}
        retries=${retries:-$DEFAULT_RETRIES}

        # 設置 SNMP 相關預設值 (如果 protocol 是 snmp)
        if [[ "$protocol" == "$PROTOCOL_SNMP_TAG" ]]; then
             snmp_version=${snmp_version:-$DEFAULT_SNMP_VERSION}
             snmp_community=${snmp_community:-$DEFAULT_SNMP_COMMUNITY}
        fi

        log_debug "[DEBUG]-${func_name}- 第 $line_number 行的值通過過濾: 協議=$protocol, 設備類型=$device_type, 品牌=$brand, 型號=$model, 啟用狀態=$enabled" >&2

        # 將所有需要的欄位寫入臨時文件，用 | 分隔
        echo "$protocol|$device_type|$brand|$model|$template_name|$template_version|$match_string|$match_field|$version_field|$serial_number_field|$snmp_community|$snmp_version|$enabled|$interval|$timeout|$retries|$description" >> "$temp_file"
        log_debug "[DEBUG]-${func_name}- 將過濾的設備添加到臨時文件: 協議=$protocol, 品牌=$brand, 型號=$model" >&2

    done < <(tail -n +2 "$config_file") # 跳過標題行

    if [[ ! -s "$temp_file" ]]; then
        log_warn "-${func_name}- 在 $config_file 中未找到模式 $mode 的啟用設備配置" >&2
        log_debug "-${func_name}- 總共處理的行數: $line_number" >&2
        rm -f "$temp_file"
        return 1
    fi

    log_debug "[DEBUG]-${func_name}- 找到 $(wc -l < "$temp_file") 個啟用且已過濾的設備配置，模式為 $mode" >&2
    echo "$temp_file" # 返回臨時文件名
}

# 讀取 tag 設備配置
# ip_key,rack,side
read_tag_configs() {
    func_name="read_tag_configs"
    local ip_key="$1"
    local tag_file="$TAG_DEVICE_CSV"
    local tag_line
    tag_line=$(awk -F',' -v ip_key="$ip_key" '$9 == ip_key { print }' "$tag_file")
    echo "$tag_line"
}
# 開始重新部署 conf 檔
redeploy_config() {
    log_error "$func_name" "=============== 開始重新部署 conf 檔 ================="
    func_name="redeploy_config"
    local ip_key="$1"
    local rack="$2"
    local side="$3"
    local slave_id="$5"
    local device_type="$6"
    
    local ip_key_idx
    ip_key_idx=$(get_registry_index "ip_key")
    local record
    record=$(awk -F',' -v key="$ip_key" -v idx="$ip_key_idx" '$idx == key { print }' "$REGISTRY_CSV")

    if [[ -z "$record" ]]; then
        log_error "$func_name" "ip_key 未找到: $ip_key"
        return 1
    fi

    IFS=',' read -r $REGISTRY_COLUMNS <<< "$record"

    local port="${ip_key##*:}"
    local template_name="PDU-${brand}-${model}_${protocol}.conf"
    local template_version="$version"

    
    generate_config "$device_type" "$factory" "$phase" "$dc" "$room" "$ip_key" "$ip" "$port" "$slave_id" "$brand" "$model" "$version" "$serial_number" "$snmp_engine_id" "$protocol" "$interval" "$template_version" "$snmp_version" "$snmp_community" "$template_name" "$snmp_mibs_path" "$timeout" "$retries"
}

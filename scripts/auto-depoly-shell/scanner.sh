#!/bin/bash
SHELL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SHELL_DIR/log.sh"
source "$SHELL_DIR/utils.sh"

# 檢查設備匹配，若為 WCM Gateway 則自動觸發 modbus 掃描，並避免重複掃描
scan_device_match() {
    echo "================= 檢查設備 $1 匹配 ===================" >&2
    func_name="scan_device_match"
    local mode="$1"
    local factory="${2:-$DEFAULT_UNKNOWN_VALUE}"       # 預設值為 #
    local phase="${3:-$DEFAULT_UNKNOWN_VALUE}"           # 預設值為 #
    local dc="${4:-$DEFAULT_UNKNOWN_VALUE}"                # 預設值為 #
    local room="${5:-$DEFAULT_UNKNOWN_VALUE}"            # 預設值為 #
    local ip_start="${6:-}"                     # 起始 IP
    local ip_end="${7:-}"                       # 結束 IP



    # 檢查必要的參數
    if [[ "$mode" == "$PROTOCOL_SNMP_TAG" ]]; then
        if [[ -z "$ip_start" || -z "$ip_end" ]]; then
            log_error "[${func_name}] 錯誤：需要提供 IP 範圍參數：ip_start 和 ip_end" >&2
            return 1
        fi
    fi

    # 讀取設備配置
    local device_config_file
    device_config_file=$(read_scan_configs "$mode" | tail -n1)
    if [[ -z "$device_config_file" || ! -f "$device_config_file" ]]; then
        log_error "[${func_name}] 錯誤：無法讀取 $mode 設備配置" >&2
        return 1
    fi

    log_debug "$func_name" "調試：已讀取設備配置：$device_config_file"
    cat "$device_config_file"

    # 創建一個新的臨時文件來存儲過濾後的配置
    local filtered_config_file
    filtered_config_file=$(mktemp -p "$TMP_DIR")
    if [[ $? -ne 0 ]]; then
        log_error "[${func_name}] 錯誤：無法創建臨時文件以存儲過濾後的配置" >&2
        rm -f "$device_config_file"
        return 1
    fi

    # 根據 mode 過濾 protocol
    local expected_protocol
    if [[ "$mode" == "$PROTOCOL_SNMP_TAG" ]]; then
        expected_protocol="$PROTOCOL_SNMP_TAG"
    elif [[ "$mode" == "$PROTOCOL_MODBUS_TAG" ]]; then
        expected_protocol="$PROTOCOL_MODBUS_TAG"
    else
        log_error "[${func_name}] 錯誤：無效的模式：$mode" >&2
        rm -f "$device_config_file" "$filtered_config_file"
        return 1
    fi

    # 過濾配置
    # protocol,device_type,brand,model,template_name,template_version,match_string,match_field,version_field,serial_number_field,snmp_community,snmp_version,enabled,interval,timeout,retries,description
    
    while IFS='|' read -r $SCAN_COLUMNS; do
        if [[ "$protocol" == "$expected_protocol" ]]; then
            echo "$protocol|$device_type|$brand|$model|$template_name|$template_version|$match_string|$match_field|$version_field|$serial_number_field|$snmp_community|$snmp_version|$enabled|$interval|$timeout|$retries|$description" >> "$filtered_config_file"
            log_debug "$func_name" "調試：已為協議 $protocol 添加配置：品牌=$brand, 型號=$model"
        fi
    done < "$device_config_file"

    # 檢查過濾後的配置是否為空
    if [[ ! -s "$filtered_config_file" ]]; then
        log_error "[${func_name}] 錯誤：未找到符合模式 $mode 的協議配置" >&2
        rm -f "$device_config_file" "$filtered_config_file"
        return 1
    fi

    # 確保 registry 文件存在且標題正確
    ensure_registry_fields "$REGISTRY_CSV"

    # 創建一個新的臨時文件來存儲 registry
    local temp_registry
    temp_registry=$(mktemp -p "$TMP_DIR")
    if [[ $? -ne 0 ]]; then
        log_error "[${func_name}] 錯誤：無法創建臨時文件以存儲 registry" >&2
        rm -f "$device_config_file" "$filtered_config_file"
        return 1
    fi

    # 生成 IP 範圍
    local ip_range=($(generate_ip_range "$ip_start" "$ip_end"))

    # 初始化掃描統計
    local total=0
    local matched=0
    local skipped=0
    local failed=0

    # 遍歷 IP 範圍
    for ip in "${ip_range[@]}"; do
        ((total++))
        # 檢查是否已存在於 registry
        ip_key="$ip:$SNMP_PORT"
        log_debug "$func_name" "調試：應用 SNMP 檢查邏輯：ip_key=$ip_key"

        if grep -q ",$ip_key," "$REGISTRY_CSV"; then
            log_warn "[${func_name}] 跳過：$ip_key 已存在於 registry，跳過。" >&2
            ((skipped++))
            continue
        fi

       # echo "================= 檢查是否可連接 $ip ===================" >&2
        if ! ping -c 1 -W 1 "$ip" > /dev/null; then
            log_warn "[${func_name}] 警告：無法 ping 通 $ip，跳過。" >&2
            ((skipped++))
            continue
        fi
        log_debug "$func_name" "✅ $ip 可用" >&2

        # 檢查設備並獲取 device_type
        local scan_result
        scan_result=$(scan_snmp "$ip" "$filtered_config_file" "$factory" "$phase" "$dc" "$room")
        local scan_status=$?

        if [[ $scan_status -eq 0 ]]; then
            # 檢查是否為 Gateway
            log_debug "$func_name" "調試：掃描結果：$scan_result" >&2
                if [[ "${scan_result,,}" == "${GATEWAY_TYPE_NAME,,}" ]]; then
                log_info "[${func_name}] ⭕ IoT Gateway $ip 通過 SNMP 匹配。檢查鎖定並觸發 Modbus 掃描..." >&2
                local lockfile="$TMP_DIR/lock-${ip//./_}.$scan_result.lock"
                if [[ -f "$lockfile" ]]; then
                    log_warn "[${func_name}] 跳過：$ip 已在掃描中（鎖定存在）。" >&2
                    ((skipped++))
                    continue
                fi
                touch "$lockfile"

                # 重新讀取設備配置，但這次只過濾 Modbus 配置
                local modbus_config_file
                modbus_config_file=$(read_scan_configs "$PROTOCOL_MODBUS_TAG" | tail -n1)
                if [[ -z "$modbus_config_file" || ! -f "$modbus_config_file" ]]; then
                    log_error "[${func_name}] 錯誤：無法讀取 Modbus 設備配置" >&2
                    rm -f "$lockfile"
                    ((failed++))
                    continue
                fi

                # 正確傳遞所有參數給 scan_modbus
                scan_modbus "$ip" "$modbus_config_file" "$factory" "$phase" "$dc" "$room"
                rm -f "$lockfile" "$modbus_config_file"
            fi
        else
            log_warn "[${func_name}] 警告：在 IP: $ip 未找到設備" >&2
            ((failed++))
        fi
    done

    # 清理臨時文件
    rm -f "$device_config_file" "$filtered_config_file" "$temp_registry"

    # 記錄掃描總結
    log_summary "$total" "$matched" "$skipped" "$failed"
}

# 通用 SNMP 掃描函式，包含 Gateway 偵測並觸發 Modbus 掃描
scan_snmp() {
    log_info "=================== 開始掃描 SNMP 設備 ==================="
    func_name="scan_snmp"
    local ip="$1"
    local device_config_file="$2" # 這是 read_scan_configs 返回的臨時文件名
    local factory="${3:-$DEFAULT_UNKNOWN_VALUE}"       # 預設值為 #
    local phase="${4:-$DEFAULT_UNKNOWN_VALUE}"           # 預設值為 #
    local dc="${5:-$DEFAULT_UNKNOWN_VALUE}"                # 預設值為 #
    local room="${6:-$DEFAULT_UNKNOWN_VALUE}"            # 預設值為 #

    # 初始化掃描統計
    local total=0
    local matched=0
    local skipped=0
    local failed=0

    # 讀取已過濾的設備配置到數組
    local -a device_configs
    # Read the new temp file format
    while IFS='|' read -r $SCAN_COLUMNS; do
         device_configs+=("$protocol|$device_type|$brand|$model|$template_name|$template_version|$match_string|$match_field|$version_field|$serial_number_field|$snmp_community|$snmp_version|$enabled|$interval|$timeout|$retries|$description")
         log_debug "$func_name" "已載入過濾配置行到數組: protocol=$protocol, brand=$brand, model=$model, template_name=$template_name, template_version=$template_version, match_string=$match_string, match_field=$match_field, version_field=$version_field, serial_number_field=$serial_number_field, snmp_community=$snmp_community, snmp_version=$snmp_version, enabled=$enabled, interval=$interval, timeout=$timeout, retries=$retries, description=$description"
    done < "$device_config_file"

    if [[ ${#device_configs[@]} -eq 0 ]]; then
        log_warn "從臨時文件 $device_config_file 中未載入任何過濾設備配置"
        ((failed++))
        log_summary "$total" "$matched" "$skipped" "$failed"
        return 1
    fi

    log_debug "$func_name" "處理 ${#device_configs[@]} 個過濾設備配置"

    # 遍歷已過濾的設備配置
    for config in "${device_configs[@]}"; do
        ((total++))
        log_debug "$func_name" "調試：處理設備配置: $config"
        # 讀取設備配置
        IFS='|' read -r $SCAN_COLUMNS <<< "$config"

        local ip_key="$ip:$SNMP_PORT"

        # 檢查 SNMP 回應
        log_debug "$func_name" "調試：檢查 SNMP 回應: ip_key=$ip_key, match_field=$match_field, snmp_community=$snmp_community, snmp_version=$snmp_version, match_string=$match_string"
        local model_check=$(get_snmp_value "$match_field" "$ip" "$snmp_community" "$snmp_version")
        if [[ $? -ne 0 ]]; then
            log_warn "無法獲取 $ip 的 SNMP 型號資訊，使用字段 $match_field"
            ((skipped++))
            continue 
        fi

        log_debug "$func_name" "SNMP 回應 $ip: $model_check"
        log_debug "$func_name" "🚫 檢查設備: brand=$brand, model=$model, type=$device_type, snmp_community=$snmp_community, field=$match_field, model_check=$model_check, match_string=$match_string,template_name=$template_name,template_version=$template_version,version_field=$version_field,serial_number_field=$serial_number_field"

        # 檢查 match_string
        if [[ "$model_check" == *"$match_string"* ]]; then
            log_info "SNMP 匹配: $ip 是 $brand $model 設備。"
            ((matched++))

            local found_version
            if [[ -n "$version_field" ]]; then
                found_version=$(get_snmp_value "$version_field" "$ip" "$snmp_community" "$snmp_version")
                if [[ $? -ne 0 ]]; then
                    log_warn "無法獲取 $ip 的 SNMP 版本，使用字段 $version_field，使用預設值。"
                    found_version="#"
                fi
            else
                log_warn "SNMP 設備 $brand $model 的 version_field 未定義，使用預設值。"
                found_version="#"
            fi
            log_debug "$func_name" "找到 SNMP 版本: $found_version"

            local found_serial_number
            if [[ -n "$serial_number_field" ]]; then
                found_serial_number=$(get_snmp_value "$serial_number_field" "$ip" "$snmp_community" "$snmp_version")
                if [[ $? -ne 0 ]]; then
                    log_warn "無法獲取 $ip 的 SNMP 序列號，使用字段 $serial_number_field，使用預設值。"
                    found_serial_number="#"
                fi
            else
                log_warn "SNMP 設備 $brand $model 的 serial_number_field 未定義，使用預設值。"
                found_serial_number="#"
            fi
            log_debug "$func_name" "找到 SNMP 序列號: $found_serial_number"

            # 匹配後產生 SNMP 配置並更新 registry
            log_debug "$func_name" "[1]設備匹配成功，回傳 device_type：$device_type,factory：$factory,phase：$phase,dc：$dc,room：$room,ip_key：$ip_key,ip：$ip,port：$SNMP_PORT,slave_id：$DEFAULT_UNKNOWN_VALUE,brand：$brand,model：$model,version：$found_version,serial_number：$found_serial_number,snmp_engine_id：$DEFAULT_UNKNOWN_VALUE,protocol：$PROTOCOL_SNMP_TAG,interval：$interval,template_version：$template_version,snmp_version：$snmp_version,snmp_community：$snmp_community,template_name：$template_name,snmp_mibs_path：$SET_SNMP_MIBS_PATH,timeout：$timeout,retries：$retries"
            if ! generate_config "$device_type" "$factory" "$phase" "$dc" "$room" \
                "$ip_key" "$ip" "$SNMP_PORT" "$DEFAULT_UNKNOWN_VALUE" "$brand" "$model" "$found_version" "$found_serial_number" \
                "$DEFAULT_UNKNOWN_VALUE" "$PROTOCOL_SNMP_TAG" \
                "$interval" "$template_version" "$snmp_version" "$snmp_community" \
                "$template_name" "$SET_SNMP_MIBS_PATH" "$timeout" "$retries"; then
                log_error "無法為 $ip 生成 SNMP 配置"
                ((failed++))
                log_summary "$total" "$matched" "$skipped" "$failed"
                return 1
            else
                log_info "成功為 $ip 生成 SNMP 配置"
                # 返回 device_type 給調用者
                echo "$device_type"
                log_summary "$total" "$matched" "$skipped" "$failed"
                return 0
            fi
        fi

        log_debug "$func_name" "配置未匹配 IP $ip: brand=$brand, model=$model"
    done

    log_info "未找到匹配且可配置的設備，IP 為 $ip"
    ((failed++))
    log_summary "$total" "$matched" "$skipped" "$failed"
    return 1
}

scan_modbus() {
    log_info "================ 開始掃描 Modbus 設備 ================"
    func_name="scan_modbus"
    local gateway_ip="$1"
    local device_config_file="$2" # 新增 device_config_file 參數
    local factory="${3:-$DEFAULT_UNKNOWN_VALUE}"      # 預設值為 #
    local phase="${4:-$DEFAULT_UNKNOWN_VALUE}"          # 預設值為 #
    local dc="${5:-$DEFAULT_UNKNOWN_VALUE}"               # 預設值為 #
    local room="${6:-$DEFAULT_UNKNOWN_VALUE}"           # 預設值為 #

    # 初始化掃描統計
    local total=0
    local matched=0
    local skipped=0
    local failed=0

    log_info "${func_name}- 正在掃描閘道: $gateway_ip"

    # 檢查設備配置文件
    if [[ -z "$device_config_file" || ! -f "$device_config_file" ]]; then
        log_error "${func_name}- 找不到設備配置文件: $device_config_file"
        ((failed++))
        log_summary "$total" "$matched" "$skipped" "$failed"
        return 1
    fi

    # 讀取已過濾的設備配置到數組
    local -a device_configs
    while IFS='|' read -r $SCAN_COLUMNS; do
         device_configs+=("$protocol|$device_type|$brand|$model|$template_name|$template_version|$match_string|$match_field|$version_field|$serial_number_field|$snmp_community|$snmp_version|$enabled|$interval|$timeout|$retries|$description")
         log_debug "$func_name" "🟣已載入過濾配置行到數組: protocol=$protocol, device_type=$device_type, brand=$brand, model=$model, template_name=$template_name, template_version=$template_version, snmp_community=$snmp_community, snmp_version=$snmp_version, match_string=$match_string, match_field=$match_field, version_field=$version_field, serial_number_field=$serial_number_field, enabled=$enabled, interval=$interval, timeout=$timeout, retries=$retries, description=$description"
    done < "$device_config_file"

    if [[ ${#device_configs[@]} -eq 0 ]]; then
        log_warn "${func_name}- 從臨時文件 $device_config_file 中未加載任何過濾設備配置"
        ((failed++))
        log_summary "$total" "$matched" "$skipped" "$failed"
        return 1
    fi

    # 掃描多組 port/slave
    IFS=':' read -ra ports <<< "$SET_MODBUS_PORTS"
    IFS=':' read -ra slaves <<< "$SET_MODBUS_SLAVES"

  

    for port in "${ports[@]}"; do
        for slave in "${slaves[@]}"; do
            local ip_key="$gateway_ip:$port:$slave"
            log_debug "${func_name}" "執行 get_modbus_value 參數：ip=$gateway_ip, port=$port, slave=$slave, field=$match_field"
            log_info "${func_name}- 正在檢查 $ip_key"

            # 遍歷設備配置
            for config in "${device_configs[@]}"; do
                ((total++))
                IFS='|' read -r $SCAN_COLUMNS <<< "$config"
  
                # 使用 get_modbus_value 獲取 model_info
                local model_check
                model_check=$(get_modbus_value "$gateway_ip" "$port" "$SET_MODBUS_TRANSMISSION_MODE" "$slave" "$match_field")
                log_debug "${func_name}" "🟩 匹配：$model 🟨 值：$model_check"
                if [[ $? -ne 0 ]]; then
                    log_warn "${func_name}- 使用字段 $match_field 獲取 Modbus 型號失敗，繼續下一個。"
                    ((skipped++))
                    continue
                fi

                local found_model=$(echo "$model_check" | awk '{print $1}')
                # 獲取版本信息
                local found_version=$(echo "$model_check" | awk '{print $2}')
                if [[ "$found_model" == *"$match_string"* ]]; then
                    log_info "${func_name}- Modbus 匹配: $gateway_ip 是 $brand $model 設備。"
                    ((matched++))

                    # 獲取序列號
                    local found_serial_number
                    if [[ -n "$serial_number_field" ]]; then
                        found_serial_number=$(get_modbus_value "$gateway_ip" "$port" "$SET_MODBUS_TRANSMISSION_MODE" "$slave" "$serial_number_field")
                        if [[ $? -ne 0 ]]; then
                            log_warn "${func_name}- 使用字段 $serial_number_field 獲取 Modbus 序列號失敗，使用預設值。"
                            found_serial_number="#"
                        fi
                    else
                        log_warn "${func_name}- 在配置中未定義 serial_number_field 用於 Modbus 設備 $brand $model，使用預設值。"
                        found_serial_number="#"
                    fi

                    # 產生 Modbus 配置並更新 registry
# $1: device_type,$2: factory,$3: phase,$4: dc,        $5: room
# $6: ip_key,     $7: ip,     $8: port, $9: slave_id,  $10: brand
# $11: model,$12: version,$13: serial_number,
#$14: snmp_engine_id,$15: protocol
# $16: interval,$17: template_version,$18: snmp_version,$19: snmp_community
# $20: template_name,$21: snmp_mibs_path,$22: timeout,$23: retries
                    if ! generate_config "$device_type" "$factory" "$phase" "$dc" "$room" \
                        "$ip_key" "$gateway_ip" "$port" "$slave" "$brand" "$found_model" "$found_version" "$found_serial_number" \
                        $DEFAULT_UNKNOWN_VALUE "$PROTOCOL_MODBUS_TAG" \
                        "$interval" "$template_version" $DEFAULT_UNKNOWN_VALUE $DEFAULT_UNKNOWN_VALUE \
                        "$template_name" $DEFAULT_UNKNOWN_VALUE "$timeout" "$retries"; then
                        log_error "${func_name}- 為 $gateway_ip:$port:$slave 生成 Modbus 配置失敗"
                        ((failed++))
                        continue
                    fi
                fi
            done
        done
    done

    log_warn "${func_name}- 在閘道 $gateway_ip 中未找到匹配的 Modbus 設備"
    ((failed++))
    log_summary "$total" "$matched" "$skipped" "$failed"
    return 1
}
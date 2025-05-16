#!/bin/bash
SHELL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SHELL_DIR/log.sh"

MODBUSGET_BIN="$SHELL_DIR/modbusget"

# 獲取 Modbus 值
get_modbus_value() {
    func_name="get_modbus_value"
    local ip="$1"
    local port="$2"
    local transmission_mode="$3"
    local slave="$4"
    local field="$5"
    
    log_debug "$func_name" "參數：ip=$1, port=$2, transmission_mode=$3, slave=$4, field=$5"
    
    if [[ -z "$ip" || -z "$port" || -z "$transmission_mode" || -z "$slave" || -z "$field" ]]; then
        log_error "$func_name" "獲取 Modbus 值時缺少參數：ip=$ip, port=$port, transmission_mode=$transmission_mode, slave=$slave, field=$field"
        return 1
    fi
    
    IFS=':' read -r address type length <<< "$field"
    
    if [[ -z "$address" || -z "$type" ]]; then
        log_error "$func_name" "欄位格式無效：$field，應為 address:type[:length]"
        return 1
    fi
    
    local cmd="$MODBUSGET_BIN -controller tcp://$ip:$port -transmission-mode $transmission_mode -slave $slave -type $type -address $address"
    
    if [[ -n "$length" ]]; then
        cmd="$cmd -length $length"
    fi
    
    log_debug "$func_name" "執行命令：$cmd"
    
    local value=$($cmd 2>/dev/null)
    local ret_code=$?
    
    if [[ $ret_code -ne 0 || -z "$value" ]]; then
        log_warn "$func_name" "無法從 $ip:$port 獲取欄位 $field 的值 (返回碼: $ret_code)"
        return 1
    fi
    
    echo "$value"
    return 0
}

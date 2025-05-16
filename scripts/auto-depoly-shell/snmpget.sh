#!/bin/bash
SHELL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SHELL_DIR/log.sh"

# 檢查 snmpget 是否可用
if ! command -v snmpget >/dev/null 2>&1; then
    log_error "snmpget" "系統未安裝 snmpget 指令，請先安裝 net-snmp 套件"
    exit 1
fi

# =================== SNMP 相關常量 ===================
SNMP_STRING_PATTERN="^.*STRING:\ *\"(.*)\""
SNMP_HEX_STRING_PATTERN="^.*Hex-STRING:\ *(.*)"
SNMP_INTEGER_PATTERN="^.*INTEGER:\ *(.*)"
SNMP_COUNTER32_PATTERN="^.*Counter32:\ *(.*)"
SNMP_COUNTER64_PATTERN="^.*Counter64:\ *(.*)"
SNMP_GAUGE32_PATTERN="^.*Gauge32:\ *(.*)"
SNMP_TIMETICKS_PATTERN="^.*Timeticks:\ *(.*)"
SNMP_OID_PATTERN="^.*OID:\ *(.*)"

# 獲取 SNMP OID 的值
get_snmp_value() {
    func_name="get_snmp_value"
    local snmp_timeout="${DEFAULT_TIMEOUT:-5s}"
    local oid="$1"
    local ip="$2"
    local community="$3"
    local snmp_version="$4"
    
    if [[ -z "$oid" || -z "$ip" || -z "$community" || -z "$snmp_version" ]]; then
        log_error "$func_name" "獲取 SNMP 值時缺少參數：oid=$oid, ip=$ip, community=$community, snmp_version=$snmp_version"
        return 1
    fi

    log_debug "$func_name" "調試：獲取 SNMP 值: timeout=$snmp_timeout, oid=$oid, ip=$ip, community=$community, snmp_version=$snmp_version"
    
    local raw_value
    raw_value="$(timeout $snmp_timeout snmpget -v "$snmp_version" -c "$community" "$ip" "$oid" 2>&1)"
    log_debug "$func_name" "原始 SNMP 回應：$raw_value"
    
    if [[ $? -ne 0 ]]; then
        log_warn "$func_name" "SNMP 命令失敗：$raw_value"
        return 1
    elif [[ -z "$raw_value" ]]; then
        log_warn "$func_name" "SNMP 無回應（空值）: OID=$oid, IP=$ip"
        return 1
    fi
    
    local value
    if [[ "$raw_value" =~ $SNMP_STRING_PATTERN ]]; then
        value="${BASH_REMATCH[1]}"
    elif [[ "$raw_value" =~ $SNMP_HEX_STRING_PATTERN ]]; then
        value="${BASH_REMATCH[1]}"
    elif [[ "$raw_value" =~ $SNMP_INTEGER_PATTERN ]]; then
        value="${BASH_REMATCH[1]}"
    elif [[ "$raw_value" =~ $SNMP_COUNTER32_PATTERN ]]; then
        value="${BASH_REMATCH[1]}"
    elif [[ "$raw_value" =~ $SNMP_COUNTER64_PATTERN ]]; then
        value="${BASH_REMATCH[1]}"
    elif [[ "$raw_value" =~ $SNMP_GAUGE32_PATTERN ]]; then
        value="${BASH_REMATCH[1]}"
    elif [[ "$raw_value" =~ $SNMP_TIMETICKS_PATTERN ]]; then
        value="${BASH_REMATCH[1]}"
    elif [[ "$raw_value" =~ $SNMP_OID_PATTERN ]]; then
        value="${BASH_REMATCH[1]}"
    else
        value=$(echo "$raw_value" | sed -E 's/^.*=\s*(.*)$/\1/')
    fi

    # 去掉頭尾的引號
    while [[ "$value" =~ ^[\'\"] ]] || [[ "$value" =~ [\'\"]$ ]]; do
        value="${value#[\'\"]}"
        value="${value%[\'\"]}"
    done

    # 去掉空白字符
    value=$(echo "$value" | tr -d '[:space:]')

    log_info "$func_name" "✅ 獲取 OID $oid 的值：$value"
    echo "$value"
}
#!/bin/bash
THIS_FILE="$(realpath "${BASH_SOURCE[0]}")"
SHELL_DIR="$(dirname "$THIS_FILE")"
#set -x  # 開啟除錯模式

# 載入環境變數
source "$SHELL_DIR/env.sh"

# 載入工具函數
source "$SHELL_DIR/utils.sh"

# 檢查必要的環境變數
if [[ -z "$SHELL_DIR" || -z "$TEMPLATE_DIR" || -z "$OUTPUT_DIR" || -z "$REGISTRY_CSV" ]]; then
    echo "錯誤：缺少必要的環境變數" >&2
    echo "SHELL_DIR: $SHELL_DIR" >&2
    echo "TEMPLATE_DIR: $TEMPLATE_DIR" >&2
    echo "OUTPUT_DIR: $OUTPUT_DIR" >&2
    echo "REGISTRY_CSV: $REGISTRY_CSV" >&2
    exit 1
fi

# 載入其他腳本
source "$SHELL_DIR/log.sh"
source "$SHELL_DIR/scanner.sh"
source "$SHELL_DIR/generate.sh"
source "$SHELL_DIR/modbusget.sh"
source "$SHELL_DIR/snmpget.sh"

# 初始化日誌
log_init

# =================== 主程序 ===================
# 解析命令行參數
mode="scan"  # 預設模式
log_info "main" "============= START VIOT DEPLOY ============="
log_info "main" "模式: $mode | 時間: $(date '+%Y-%m-%d %H:%M:%S')"
log_info "main" "工作目錄: $BASE_DIR"
while [[ $# -gt 0 ]]; do
    case "$1" in
        --mode)
            mode="$2"
            shift 2
            ;;
        --factory)
            factory="$2"
            shift 2
            ;;
        --phase)
            phase="$2"
            shift 2
            ;;
        --dc)
            dc="$2"
            shift 2
            ;;
        --room)
            room="$2"
            shift 2
            ;;
        --ip-start)
            ip_start="$2"
            shift 2
            ;;
        --ip-end)
            ip_end="$2"
            shift 2
            ;;
        --ip-key)
            ip_key="$2"
            shift 2
            ;;
        --rack)
            rack="$2"
            shift 2
            ;;
        --side)
            side="$2"
            shift 2
            ;;
        *)
            log_error "main" "未知選項: $1"
            exit 1
            ;;
    esac
done

# 根據模式執行相應的操作
case "$mode" in
    "scan")
        if [[ -z "$ip_start" || -z "$ip_end" ]]; then
            log_error "main" "用法: $0 --mode scan [--factory <factory>] [--phase <phase>] [--dc <dc>] [--room <room>] --ip-start <ip_start> --ip-end <ip_end>"
            exit 1
        fi
        scan_device_match "$PROTOCOL_SNMP_TAG" "$factory" "$phase" "$dc" "$room" "$ip_start" "$ip_end"
        ;;
    "redeploy")
        if [[ -z "$ip_key" || -z "$rack" || -z "$side" ]]; then
            log_error "main" "用法: $0 --mode redeploy --ip-key <ip_key> --rack <rack> --side <side>"
            exit 1
        fi
        redeploy_config "$ip_key" "$rack" "$side"
        ;;
    *)
        log_error "main" "無效模式: $mode。有效模式為 'scan' 或 'redeploy'"
        exit 1
        ;;
esac 

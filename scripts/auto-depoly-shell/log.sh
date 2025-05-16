#!/bin/bash
SHELL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 預設 console 與 logfile 等級
LOG_LEVEL="${LOG_LEVEL:-INFO}"            # 控制畫面輸出
LOG_FILE_LEVEL="${LOG_FILE_LEVEL:-INFO}"  # 控制是否寫入 LOG 檔
LOG_FILE_PATH="${LOG_FILE_PATH:-}"        # 要寫入的檔案路徑（空 = 不寫）
LOG_ROTATE_BY_DAY=${LOG_ROTATE_BY_DAY:-true}

# 等級對應數值
level_to_number() {
  case "$1" in
    DEBUG) echo 0 ;;
    INFO) echo 1 ;;
    WARN) echo 2 ;;
    ERROR) echo 3 ;;
    *) echo 1 ;;  # 預設 INFO
  esac
}

# =================== 日誌設定 ===================
LEVEL_DEBUG=0
LEVEL_INFO=1
LEVEL_WARN=2
LEVEL_ERROR=3

CURRENT_LEVEL=$(level_to_number "$LOG_LEVEL")
FILE_LEVEL=$(level_to_number "$LOG_FILE_LEVEL")

# 初始化 log 函式（檢查檔案路徑並建立目錄與空檔案）
log_init() {
  if [[ -n "$LOG_FILE_PATH" ]]; then
    local log_dir
    log_dir=$(dirname "$LOG_FILE_PATH")
    mkdir -p "$log_dir"

    if [[ "$LOG_ROTATE_BY_DAY" = true ]]; then
      local today
      today=$(date +'%Y-%m-%d')
      LOG_FILE_PATH_DAILY="${LOG_FILE_PATH%.log}.$today.log"
    else
      LOG_FILE_PATH_DAILY="$LOG_FILE_PATH"
    fi

    mkdir -p "$(dirname "$LOG_FILE_PATH_DAILY")"

    touch "$LOG_FILE_PATH_DAILY" || {
      echo "[ERROR]-log-init- Cannot write to $LOG_FILE_PATH_DAILY" >&2
      exit 1
    }
    echo "[INFO]-log-init- Log file initialized at $LOG_FILE_PATH_DAILY" >> "$LOG_FILE_PATH_DAILY"
  fi
}

log_output() {
  local level="$1"
  local source="$2"
  local message="$3"
  local timestamp
  timestamp=$(date '+%Y-%m-%d %H:%M:%S')

  local formatted="[$timestamp][$level]-$source $message"

  # 輸出至 console
  local level_num
  level_num=$(level_to_number "$level")
  if [[ $level_num -ge $CURRENT_LEVEL ]]; then
    echo "$formatted" >&2
  fi

  # 輸出至 logfile（若有啟用）
  if [[ -n "$LOG_FILE_PATH_DAILY" && $level_num -ge $FILE_LEVEL ]]; then
    echo "$formatted" >> "$LOG_FILE_PATH_DAILY"
  fi
}

log_debug() { log_output "DEBUG" "$1" "$2"; }
log_info()  { log_output "INFO" "$1" "$2"; }
log_warn()  { log_output "WARN" "$1" "$2"; }
log_error() { log_output "ERROR" "$1" "$2"; }

log_summary() {
    local func_name="log_summary"
    local total="$1"
    local matched="$2"
    local skipped="$3"
    local failed="$4"
    
    log_info "$func_name" "🔍 掃描摘要："
    log_info "$func_name" "總掃描設備數：$total"
    log_info "$func_name" "成功匹配並產生配置：$matched"
    log_info "$func_name" "已跳過（已存在或不符）：$skipped"
    log_info "$func_name" "失敗或未找到設備：$failed"
}

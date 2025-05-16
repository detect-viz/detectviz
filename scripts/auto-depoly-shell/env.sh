#!/bin/bash
SHELL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 載入 .env 設定檔（支援外部自訂）
ENV_FILE="$(realpath "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../../../../etc/default/.env")"
[[ -f "$ENV_FILE" ]] && { set -a; source "$ENV_FILE"; set +a; }
source "$SHELL_DIR/log.sh"

load_env_constants() {
  # PDU 命名格式（可覆蓋）
  PDU_NAME_FORMAT='${factory}${phase}${dc}${room}${rack}P${side}'

  # ========== 路徑設定 ==========
  BASE_DIR="${BASE_DIR:-/opt/viot/auto_depoly}"
  DATA_DIR="${DATA_DIR:-$BASE_DIR/var/lib/viot}"
  CONF_DIR="${CONF_DIR:-$BASE_DIR/etc/viot}"
  TEMPLATE_DIR="${TEMPLATE_DIR:-$BASE_DIR/usr/share/telegraf/templates}"
  OUTPUT_DIR="${OUTPUT_DIR:-$BASE_DIR/etc/telegraf/telegraf.d}"
  SHELL_DIR="${SHELL_DIR:-$BASE_DIR/usr/lib/viot/scripts}"
  TMP_DIR="${TMP_DIR:-$BASE_DIR/tmp}"

  # ========== 設定檔 ==========
  SCAN_DEVICE_CSV="${SCAN_DEVICE_CSV:-$CONF_DIR/scan_device.csv}"
  TAG_DEVICE_CSV="${TAG_DEVICE_CSV:-$CONF_DIR/tag_device.csv}"
  REGISTRY_CSV="${REGISTRY_CSV:-$DATA_DIR/registry_device.csv}"

  # ========== 日誌設定 ==========
  LOG_LEVEL="${LOG_LEVEL:-DEBUG}"
  LOG_FILE_LEVEL="${LOG_FILE_LEVEL:-INFO}"
  LOG_FILE_PATH="${LOG_FILE_PATH:-$BASE_DIR/var/log/viot/main.log}"
  LOG_ROTATE_BY_DAY="${LOG_ROTATE_BY_DAY:-true}"

  # ========== 協定設定 ==========
  PROTOCOL_SNMP_TAG="${PROTOCOL_SNMP_TAG:-snmp}"
  PROTOCOL_MODBUS_TAG="${PROTOCOL_MODBUS_TAG:-modbus}"
  SNMP_PORT="${SNMP_PORT:-161}"
  SET_SNMP_MIBS_PATH="${SET_SNMP_MIBS_PATH:-/usr/share/snmp/mibs}"
  GATEWAY_TYPE_NAME="${GATEWAY_TYPE_NAME:-gateway}"
  SET_MODBUS_TRANSMISSION_MODE="${SET_MODBUS_TRANSMISSION_MODE:-RTUoverTCP}"
  SET_MODBUS_PORTS="${SET_MODBUS_PORTS:-7000:7001}"
  SET_MODBUS_SLAVES="${SET_MODBUS_SLAVES:-0:1:2}"

  # ========== Modbus 優化設定 ==========
  SET_REPLACE_BUSY_RETRIES_WAIT="${SET_REPLACE_BUSY_RETRIES_WAIT:-500ms}"
  SET_REPLACE_PAUSE_AFTER_CONNECT="${SET_REPLACE_PAUSE_AFTER_CONNECT:-500ms}"
  SET_REPLACE_PAUSE_BETWEEN_REQUESTS="${SET_REPLACE_PAUSE_BETWEEN_REQUESTS:-1000ms}"

  # ========== 預設值 ==========
  DEFAULT_UNKNOWN_VALUE="${DEFAULT_UNKNOWN_VALUE:-#}"
  DEFAULT_SNMP_VERSION="${DEFAULT_SNMP_VERSION:-2c}"
  DEFAULT_SNMP_COMMUNITY="${DEFAULT_SNMP_COMMUNITY:-public}"
  DEFAULT_ENABLED="${DEFAULT_ENABLED:-1}"
  DEFAULT_INTERVAL="${DEFAULT_INTERVAL:-60s}"
  DEFAULT_TIMEOUT="${DEFAULT_TIMEOUT:-5s}"
  DEFAULT_RETRIES="${DEFAULT_RETRIES:-3}"

  # ========== 欄位定義 ==========
  REGISTRY_COLUMNS="${REGISTRY_COLUMNS:-device_type factory phase dc room rack side name ip_key ip port slave_id brand model version serial_number snmp_engine_id protocol interval template_version snmp_version snmp_community create_at update_at}"
  SCAN_COLUMNS="${SCAN_COLUMNS:-protocol device_type brand model template_name template_version match_string match_field version_field serial_number_field snmp_community snmp_version enabled interval timeout retries description}"
  TAG_COLUMNS="${TAG_COLUMNS:-ip_key,rack,side}"

  # ========== 欄位索引 ==========
  REGISTRY_DEVICE_TYPE="${REGISTRY_DEVICE_TYPE:-1}"
  REGISTRY_FACTORY="${REGISTRY_FACTORY:-2}"
  REGISTRY_PHASE="${REGISTRY_PHASE:-3}"
  REGISTRY_DC="${REGISTRY_DC:-4}"
  REGISTRY_ROOM="${REGISTRY_ROOM:-5}"
  REGISTRY_RACK="${REGISTRY_RACK:-6}"
  REGISTRY_SIDE="${REGISTRY_SIDE:-7}"
  REGISTRY_NAME="${REGISTRY_NAME:-8}"
  REGISTRY_IP_KEY="${REGISTRY_IP_KEY:-9}"
  REGISTRY_IP="${REGISTRY_IP:-10}"
  REGISTRY_PORT="${REGISTRY_PORT:-11}"
  REGISTRY_SLAVE_ID="${REGISTRY_SLAVE_ID:-12}"
  REGISTRY_BRAND="${REGISTRY_BRAND:-13}"
  REGISTRY_MODEL="${REGISTRY_MODEL:-14}"
  REGISTRY_VERSION="${REGISTRY_VERSION:-15}"
  REGISTRY_SERIAL_NUMBER="${REGISTRY_SERIAL_NUMBER:-16}"
  REGISTRY_SNMP_ENGINE_ID="${REGISTRY_SNMP_ENGINE_ID:-17}"
  REGISTRY_PROTOCOL="${REGISTRY_PROTOCOL:-18}"
  REGISTRY_INTERVAL="${REGISTRY_INTERVAL:-19}"
  REGISTRY_TEMPLATE_VERSION="${REGISTRY_TEMPLATE_VERSION:-20}"
  REGISTRY_SNMP_VERSION="${REGISTRY_SNMP_VERSION:-21}"
  REGISTRY_COMMUNITY="${REGISTRY_COMMUNITY:-22}"
  REGISTRY_CREATE_AT="${REGISTRY_CREATE_AT:-23}"
  REGISTRY_UPDATE_AT="${REGISTRY_UPDATE_AT:-24}"

  # SCAN_COLUMNS="protocol,device_type,brand,model,template_name,template_version,match_string,match_field,version_field,serial_number_field,snmp_community,snmp_version,enabled,interval,timeout,retries,description"
  SCAN_PROTOCOL="${SCAN_PROTOCOL:-1}"
  SCAN_DEVICE_TYPE="${SCAN_DEVICE_TYPE:-2}"
  SCAN_BRAND="${SCAN_BRAND:-3}"
  SCAN_MODEL="${SCAN_MODEL:-4}"
  SCAN_TEMPLATE_NAME="${SCAN_TEMPLATE_NAME:-5}"
  SCAN_TEMPLATE_VERSION="${SCAN_TEMPLATE_VERSION:-6}"
  SCAN_MATCH_STRING="${SCAN_MATCH_STRING:-7}"
  SCAN_MATCH_FIELD="${SCAN_MATCH_FIELD:-8}"
  SCAN_VERSION_FIELD="${SCAN_VERSION_FIELD:-9}"
  SCAN_SERIAL_FIELD="${SCAN_SERIAL_FIELD:-10}"
  SCAN_COMMUNITY="${SCAN_COMMUNITY:-11}"
  SCAN_SNMP_VERSION="${SCAN_SNMP_VERSION:-12}"
  SCAN_ENABLED="${SCAN_ENABLED:-13}"
  SCAN_INTERVAL="${SCAN_INTERVAL:-14}"
  SCAN_TIMEOUT="${SCAN_TIMEOUT:-15}"
  SCAN_RETRIES="${SCAN_RETRIES:-16}"
  SCAN_DESCRIPTION="${SCAN_DESCRIPTION:-17}"
}

load_env_constants

#!/bin/bash

# 系統設定
export WORKERS=${WORKERS:-10}                       # 設定並行處理數量
export LOG_RETENTION_DAYS=${LOG_RETENTION_DAYS:-7} # 設定日誌保留天數

# 設定 InfluxDB 相關變數
export QUERY_INFLUX_TOKEN="4-Z-WuwUTh74YXnGleK4Oab7Re86BtYDt1RAMuNUkTT0e_MKftdqedZxDZX-_kv35KnB03ng=="
export QUERY_INFLUX_URL="http://192.168.0.106:8086"
export QUERY_ORG="master"
export QUERY_BUCKET="raw"
export QUERY_MEASUREMENT="pdu"
export WRITE_INFLUX_TOKEN="4-Z-WuwUTh74YXnGleK4Oab7Re86bpwBz-JFXLIl86BtYDt1RAMuNUkTT0e_MKftdqedZxDZX-_kv35KnB03ng=="
export WRITE_INFLUX_URL="http://192.168.0.106:8086"
export WRITE_ORG="master"
export WRITE_BUCKET="spc"
export WRITE_MEASUREMENT="pdu"

# 檔案路徑設定
export TARGET_CSV="./config/target.csv"
export GROUP_CSV="./config/group.csv"
export FIELD_YAML="./config/field_definitions.yaml"
export BUILD_DIR="build"
export RESULT_DIR="result"
export OUTPUT_LP="true"
export OUTPUT_JSON="false"

# 時間相關設定
export DATA_INTERVAL_MINUTES=10
export DATA_PERIOD="${DATA_INTERVAL_MINUTES}m"
export BASELINE_LOOKBACK_DAYS=7
export SECONDS_IN_DAY=86400

# SPC 統計參數
export DECIMALS=${DECIMALS:-2}
export MIN_STDDEV=${MIN_STDDEV:-0.001}
export MIN_LSL_VALUE=${MIN_LSL_VALUE:-0.01}
export DAILY_POINTS=${DAILY_POINTS:-144}
export MIN_POINTS=${MIN_POINTS:-120}

# 規格上限與下限可彈性設定為數值或百分位（如 P95、P05）(以GROUP_CSV為主，此作為默認值)
export USL_MODE=${USL_MODE:-8}
export LSL_MODE=${LSL_MODE:-P05}

# Sigma 倍數定義 (以GROUP_CSV為主，此作為默認值)
export CONTROL_CHART_SCALE=${CONTROL_CHART_SCALE:-3.0} # 用於 UCL / LCL
export ZONE_A_SCALE=${ZONE_A_SCALE:-2.0}               # Zone A 判斷範圍 ±2σ
export ZONE_B_SCALE=${ZONE_B_SCALE:-1.0}               # Zone B 判斷範圍 ±1σ

# PR 規則點數與條件門檻 (以GROUP_CSV為主，此作為默認值)
export PR1_POINTS=${PR1_POINTS:-1}
export PR1_THRESHOLD=${PR1_THRESHOLD:-1}
export PR2_POINTS=${PR2_POINTS:-3}
export PR2_THRESHOLD=${PR2_THRESHOLD:-2}
export PR3_POINTS=${PR3_POINTS:-5}
export PR3_THRESHOLD=${PR3_THRESHOLD:-4}
export PR4_POINTS=${PR4_POINTS:-8}
export PR4_THRESHOLD=${PR4_THRESHOLD:-8}
export PR5_POINTS=${PR5_POINTS:-7}
export PR5_THRESHOLD=${PR5_THRESHOLD:-7}

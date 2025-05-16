#!/bin/bash

# 參數設定
INFLUX_BUCKET="tsmc_pdu"
INFLUX_ORG="your-org"
INFLUX_TOKEN="your-token"
BACKUP_DIR="/data/influx-backup"
DATE=$(date +%F)

# 建立備份資料夾
mkdir -p "$BACKUP_DIR/$DATE"

# 執行備份
influx backup \
  --bucket "$INFLUX_BUCKET" \
  --org "$INFLUX_ORG" \
  --token "$INFLUX_TOKEN" \
  "$BACKUP_DIR/$DATE"

echo "[DONE] Backup completed to $BACKUP_DIR/$DATE"
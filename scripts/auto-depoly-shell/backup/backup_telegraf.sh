#!/bin/bash
source ./config.sh
echo "[BACKUP] 備份 conf 與 csv 中..."
mkdir -p "$BACKUP_DIR"
tar -czf "$BACKUP_CONF" "$TELEGRAF_CONF_DIR"
cp "$CSV_REGISTRY" "$BACKUP_REGISTRY"
#!/bin/bash

RESTORE_DIR="/data/influx-backup/2025-04-12"
influx restore --full "$RESTORE_DIR"
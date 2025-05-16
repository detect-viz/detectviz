# bimap-zox

## 資料傳遞 API
1. telegraf endpoint: http://localhost:8088/telegraf
2. pdu_list endpoint: http://localhost:8089/pdu_list

## 定位標籤資料規則(模擬用)
1. rack 編號：每個 rack 使用一個大寫字母開頭，後跟兩到三位的數字編號，如 A01, Z123 等等。
2. 每個 room 包含 70 個 rack：每個 rack 包含兩個 PDU，對應左右兩側 (L 和 R)。
3. PDU 型號與協議：PDU428 使用 snmp 協議。PDU4425 和 PDU1315 使用 modbus 協議。
4.WiFi client 連接：對於 modbus 協議的 PDU，隨機生成一個 WiFi Client，一個 WiFi Client連接2個PDU ，也就是同一rack的L及R。
5.電盤連接：名稱首字是U或N 中間數字 結尾英文大寫字母，如N7B, U5G等等。每 14 個 PDU 隨機連接到一個電盤（Panel），同一個rack的L及R不可以連接同一個電盤。

## csv columns
manufacturer,model_name,protocol,device_ip,port,factory_name,datacenter_name,room_name,rack_name,pdu_side

![](/軟體架構.jpg)
![](/PDU監控佈建.jpg)

## SQLServer

-- 強制刪除資料庫
USE master;
GO

ALTER DATABASE bimap
SET SINGLE_USER WITH ROLLBACK IMMEDIATE;
GO

DROP DATABASE bimap;
GO

-- 創建資料庫
CREATE DATABASE bimap;
GO

USE bimap;
GO

INSERT INTO `groups` (`name`, `created_at`, `updated_at`)
VALUES
	('GROUP_1', NULL, NULL);

INSERT INTO `group_zones` (`group_name`, `factory`, `phase`, `created_at`, `updated_at`)
VALUES
	('GROUP_1', 'F12', 'P7', NULL, NULL),
	('GROUP_1', 'F12', 'P4', NULL, NULL);

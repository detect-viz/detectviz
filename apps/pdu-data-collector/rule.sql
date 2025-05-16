CREATE TABLE `Rules` (
  `ID` int NOT NULL,
  `Enable` bit(1) DEFAULT NULL,
  `Mode` varchar(10) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL,
  `Category` varchar(50) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL,
  `Name` varchar(50) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL,
  `Metric` varchar(50) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL,
  `Code` varchar(50) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL,
  `Description` varchar(100) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL,
  `CronExpression` varchar(100) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL,
  `Operator` varchar(10) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci DEFAULT NULL,
  `CritThreshold` float DEFAULT NULL,
  `WarnThreshold` float DEFAULT NULL,
  `InfoThreshold` float DEFAULT NULL,
  `Message` text CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `Code` (`Code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `Rules` (`ID`, `Enable`, `Mode`, `Category`, `Metric`, `Name`, `Code`, `Description`, `CronExpression`, `Operator`, `CritThreshold`, `WarnThreshold`, `InfoThreshold`, `Message`)
VALUES
	(1, 1, 'DC', 'Rack', 'Current', 'RackTotalCurrent', 'C001', '機櫃總電流過高', '0 */1 * * * *', '>', 63, 50, 30, 'Total current exceeds high threshold'),
	(2, 1, 'DC', 'Rack', 'Current', 'RackCurrentDifference', 'C002', '機櫃左右PDU電流差異率過高', '0 */1 * * * *', '>', 0.9, 0.5, 0.3, 'Rack current difference exceeds high threshold'),
	(3, 1, 'Factory', 'PDU', 'Status', 'PDUConnectionStatus', 'S001', 'PDU設備失聯', '0 */1 * * * *', '=', 0, 0, 0, 'PDU device disconnected'),
	(4, 1, 'Factory', 'Phase', 'Current', 'PhaseCurrent', 'C003', '單相電流過高', '0 */1 * * * *', '>', 16, 12, 8, 'Phase current exceeds high threshold'),
	(5, 1, 'Factory', 'Phase', 'Voltage', 'PhaseVoltage', 'V001', '單相電壓過高', '0 */1 * * * *', '>', 220, 215, 210, 'Phase voltage exceeds high threshold'),
	(6, 1, 'Factory', 'Phase', 'Voltage', 'PhaseVoltage', 'V002', '單相電壓過低', '0 */1 * * * *', '<', 210, 215, 220, 'Phase voltage falls below low threshold'),
	(7, 1, 'Factory', 'Phase', 'Voltage', 'PDUVoltageImbalance', 'V003', '三相電壓不平衡', '0 */1 * * * *', '>', 10, 7, 5, 'PDU voltage imbalance exceeds high threshold'),
	(8, 1, 'Factory', 'Phase', 'Current', 'PDUCurrentImbalance', 'C004', '三相電流不平衡', '0 */1 * * * *', '>', 10, 7, 5, 'PDU current imbalance exceeds high threshold'),
	(9, 1, 'Factory', 'Branch', 'Current', 'BranchCurrent', 'C005', 'BANK電流過高', '0 */1 * * * *', '>', 16, 12, 8, 'Phase current exceeds high threshold');

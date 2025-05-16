from typing import List, Dict
import pandas as pd
from .base import BaseDetector

class AbsoluteThresholdDetector(BaseDetector):
    def detect(self, current: List[Dict], history: List[Dict]) -> List[Dict]:
        """
        使用絕對閾值進行異常檢測
        
        閾值邏輯：
        - operator(value, critical_threshold): critical
        - operator(value, warning_threshold): warning
        - 其他: normal
        
        運算符：
        - ">": 大於
        - ">=": 大於等於
        - "<": 小於
        - "<=": 小於等於
        """
        try:
            # 檢查輸入數據
            if not current:
                return []
            
            # 將 Pydantic 模型轉換為字典
            current_data = [
                {
                    'timestamp': point.timestamp,
                    'value': point.value
                }
                for point in current
            ]
            
            # 創建 DataFrame
            results = pd.DataFrame(current_data)
            if 'timestamp' not in results.columns or 'value' not in results.columns:
                raise ValueError("數據必須包含 timestamp 和 value 字段")
            
            # 轉換時間戳
            results['timestamp'] = pd.to_datetime(results['timestamp'], unit='s')
            
            # 獲取閾值和運算符
            critical = float(self.config.get('critical_threshold', 90))
            warning = self.config.get('warning_threshold')
            operator = self.config.get('operator', '>=')
            
            # 根據運算符進行比較
            if operator == '>':
                critical_mask = results['value'] > critical
                warning_mask = warning and (results['value'] > warning)
            elif operator == '>=':
                critical_mask = results['value'] >= critical
                warning_mask = warning and (results['value'] >= warning)
            elif operator == '<':
                critical_mask = results['value'] < critical
                warning_mask = warning and (results['value'] < warning)
            elif operator == '<=':
                critical_mask = results['value'] <= critical
                warning_mask = warning and (results['value'] <= warning)
            else:
                raise ValueError(f"不支持的運算符: {operator}")
            
            # 標記嚴重程度
            results['severity'] = 'normal'
            if warning_mask is not None:
                results.loc[warning_mask & ~critical_mask, 'severity'] = 'warning'
            results.loc[critical_mask, 'severity'] = 'critical'
            
            results['threshold'] = critical
            
            # 格式化輸出
            return [
                {
                    'timestamp': int(row['timestamp'].timestamp()),
                    'value': float(row['value']),
                    'severity': row['severity'],
                    'threshold': float(row['threshold'])
                }
                for _, row in results.iterrows()
            ]
            
        except Exception as e:
            raise Exception(f"絕對閾值檢測失敗: {str(e)}") 
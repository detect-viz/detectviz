from typing import List, Dict
import pandas as pd
import numpy as np
from .base import BaseDetector

class PercentageThresholdDetector(BaseDetector):
    def detect(self, current: List[Dict], history: List[Dict]) -> List[Dict]:
        """
        使用百分位閾值進行異常檢測
        
        閾值邏輯：
        - operator(value, critical_percentile): critical
        - operator(value, warning_percentile): warning (如果有設置)
        - 其他: normal
        """
        try:
            # 檢查輸入數據
            if not current or not history:
                return []
            
            # 將 Pydantic 模型轉換為字典
            current_data = [
                {
                    'timestamp': point.timestamp,
                    'value': point.value
                }
                for point in current
            ]
            
            history_data = [
                {
                    'timestamp': point.timestamp,
                    'value': point.value
                }
                for point in history
            ]
            
            # 創建 DataFrame
            history_df = pd.DataFrame(history_data)
            current_df = pd.DataFrame(current_data)
            
            if 'timestamp' not in current_df.columns or 'value' not in current_df.columns:
                raise ValueError("數據必須包含 timestamp 和 value 字段")
            
            # 計算百分位閾值並四捨五入到小數點第二位
            critical_threshold = round(np.percentile(history_df['value'], self.config['critical_percentile']), 2)
            warning_threshold = None
            if self.config.get('warning_percentile') is not None:
                warning_threshold = round(np.percentile(history_df['value'], self.config['warning_percentile']), 2)
            operator = self.config.get('operator', '>=')
            
            # 判斷嚴重程度
            results = current_df.copy()
            results['severity'] = 'normal'
            
            # 根據運算符進行比較
            if operator == '>':
                critical_mask = results['value'] > critical_threshold
                warning_mask = None if warning_threshold is None else results['value'] > warning_threshold
            elif operator == '>=':
                critical_mask = results['value'] >= critical_threshold
                warning_mask = None if warning_threshold is None else results['value'] >= warning_threshold
            elif operator == '<':
                critical_mask = results['value'] < critical_threshold
                warning_mask = None if warning_threshold is None else results['value'] < warning_threshold
            elif operator == '<=':
                critical_mask = results['value'] <= critical_threshold
                warning_mask = None if warning_threshold is None else results['value'] <= warning_threshold
            else:
                raise ValueError(f"不支持的運算符: {operator}")
            
            # 標記嚴重程度
            results.loc[critical_mask, 'severity'] = 'critical'
            if warning_mask is not None:
                results.loc[warning_mask & ~critical_mask, 'severity'] = 'warning'
            
            results['threshold'] = critical_threshold
            
            # 格式化輸出時也確保四捨五入
            return [
                {
                    'timestamp': int(row['timestamp']),
                    'value': round(float(row['value']), 2),
                    'severity': row['severity'],
                    'threshold': round(float(row['threshold']), 2)
                }
                for _, row in results.iterrows()
            ]
            
        except Exception as e:
            raise Exception(f"百分位閾值檢測失敗: {str(e)}") 
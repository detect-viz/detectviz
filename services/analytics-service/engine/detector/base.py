from abc import ABC, abstractmethod
from typing import List, Dict, Any
import numpy as np
import pandas as pd

class BaseDetector(ABC):
    def __init__(self, config: Dict[str, Any]):
        self.config = config
    
    @abstractmethod
    def detect(self, current: List[Dict], history: List[Dict]) -> List[Dict]:
        """
        執行異常檢測
        
        Args:
            current: 當前數據列表，每個元素包含 timestamp 和 value
            history: 歷史數據列表，每個元素包含 timestamp 和 value
            
        Returns:
            List[Dict]: 檢測結果列表
        """
        pass
    
    def _prepare_data(self, data: List[Dict]) -> pd.DataFrame:
        """將輸入數據轉換為 DataFrame 格式"""
        if not data:
            return pd.DataFrame(columns=['timestamp', 'value'])
        
        # 創建 DataFrame
        df = pd.DataFrame(data)
        
        # 確保時間戳是 datetime 格式
        if 'timestamp' in df.columns:
            df['timestamp'] = pd.to_datetime(df['timestamp'], unit='s')
            df = df.set_index('timestamp')
        
        return df
    
    def _format_threshold_output(self, results: pd.DataFrame, threshold_field: str = 'threshold') -> List[Dict]:
        """用於 absolute_threshold 和 percentage_threshold 的輸出格式化"""
        # 確保 DataFrame 有索引
        if results.index.name != 'timestamp':
            results = results.reset_index()
        
        output = []
        for _, row in results.iterrows():
            result = {
                'timestamp': int(row['timestamp'].timestamp()),  # 轉換為 Unix 時間戳
                'value': float(row['value']),
                'severity': row['severity']
            }
            if threshold_field in results.columns:
                result[threshold_field] = float(row[threshold_field])
            output.append(result)
        
        return output
    
    def _format_moving_average_output(self, results: pd.DataFrame) -> List[Dict]:
        """用於 moving_average 的輸出格式化"""
        results.reset_index(inplace=True)
        results['timestamp'] = results.index.astype(np.int64) // 10**9
        
        output = []
        for _, row in results.iterrows():
            output.append({
                'timestamp': row['timestamp'],
                'value': row['value'],
                'severity': row['severity'],
                'min': row['min'],
                'max': row['max']
            })
        return output
    
    def _format_anomaly_output(self, results: pd.DataFrame) -> List[Dict]:
        """用於 isolation_forest 和 prophet 的輸出格式化"""
        results.reset_index(inplace=True)
        results['timestamp'] = results.index.astype(np.int64) // 10**9
        
        output = []
        for _, row in results.iterrows():
            output.append({
                'timestamp': row['timestamp'],
                'value': row['value'],
                'anomaly': row['anomaly']
            })
        return output 
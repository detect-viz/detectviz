from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel
from typing import List, Dict, Any, Optional
import logging
from detector import (
    AbsoluteThresholdDetector,
    PercentageThresholdDetector,
    MovingAverageDetector,
    IsolationForestDetector,
    ProphetDetector
)
from enum import Enum

# 配置日誌
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI()

class DataPoint(BaseModel):
    timestamp: int
    value: float

class ThresholdOperator(str, Enum):
    GT = ">"      # 大於
    GTE = ">="    # 大於等於
    LT = "<"      # 小於
    LTE = "<="    # 小於等於

class AbsoluteThresholdConfig(BaseModel):
    critical_threshold: float
    warning_threshold: Optional[float] = None
    operator: ThresholdOperator = ThresholdOperator.GTE  # 默認大於等於

class PercentageThresholdConfig(BaseModel):
    critical_percentile: float
    warning_percentile: Optional[float] = None
    operator: ThresholdOperator = ThresholdOperator.GTE

    class Config:
        extra = "allow"  # 允許額外的欄位
        validate_assignment = True  # 驗證賦值

class MovingAverageConfig(BaseModel):
    window: int              # 移動平均窗口大小
    std_multiplier: float    # 標準差倍數
    min_periods: int         # 最小數據點數量

    class Config:
        extra = "allow"      # 允許額外的欄位

class IsolationForestConfig(BaseModel):
    contamination: float
    n_estimators: int
    max_samples: str
    max_features: float
    random_state: int

class ProphetConfig(BaseModel):
    seasonality_mode: str
    changepoint_prior_scale: float
    interval_width: float
    uncertainty_samples: int
    horizon: int
    period: str
    holidays_prior_scale: float
    weekly_seasonality: bool
    daily_seasonality: bool
    yearly_seasonality: bool

class DetectorRequest(BaseModel):
    data: Dict[str, List[DataPoint]]
    config: Dict[str, Any]
    type: str

@app.post("/detect")
async def detect(request: DetectorRequest):
    try:
        detector_type = request.type
        config = request.config
        
        # 檢查數據需求
        data_requirements = {
            'absolute_threshold': {'current': True, 'history': False},
            'percentage_threshold': {'current': True, 'history': True},
            'moving_average': {'current': True, 'history': True},
            'isolation_forest': {'current': True, 'history': True},
            'prophet': {'current': False, 'history': True}
        }
        
        # 檢查必要的數據是否存在
        requirements = data_requirements.get(detector_type)
        if requirements:
            if requirements['current'] and 'current' not in request.data:
                raise HTTPException(
                    status_code=400,
                    detail=f"{detector_type} 需要 current 數據"
                )
            if requirements['history'] and 'history' not in request.data:
                raise HTTPException(
                    status_code=400,
                    detail=f"{detector_type} 需要 history 數據"
                )
        
        # 檢測器映射
        detectors = {
            'prophet': ProphetDetector,
            'moving_average': MovingAverageDetector,
            'isolation_forest': IsolationForestDetector,
            'absolute_threshold': AbsoluteThresholdDetector,
            'percentage_threshold': PercentageThresholdDetector
        }
        
        # 獲取檢測器類
        detector_class = detectors.get(detector_type)
        if not detector_class:
            raise HTTPException(
                status_code=400,
                detail=f"不支持的檢測器類型: {detector_type}"
            )
            
        # 創建檢測器實例
        detector = detector_class(config)
        
        # 執行檢測
        results = detector.detect(
            current=request.data.get('current', []),
            history=request.data.get('history', [])
        )
        
        return {"data": results}
        
    except Exception as e:
        logger.error(f"檢測錯誤: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "healthy"}
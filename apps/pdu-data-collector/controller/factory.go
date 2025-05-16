package controller

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bimap-zbox/services/factory"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary  接收 telegraf 發送的數據
// @Tags     collect
// @Accept   json
// @Produce  json
// @Success  200 {string} string
// @Router   /factory/metrics [post]
func CollectFactoryMetrics(c *gin.Context) {
	var data_list models.MetricsData
	if err := c.BindJSON(&data_list); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	points := data_list.Metrics
	var formattedPoints []models.Point

	// 處理每個數據點
	for _, point := range points {
		modelName := point.Tags["model"]
		pduKey := point.Tags["pdu_key"]
		deltaList := []string{"PDUE428", "PDU1315", "PDU4425"}

		match := false

		for _, delta := range deltaList {
			if strings.Contains(modelName, delta) {
				p, err := factory.FormatDeltaPDU(pduKey, point)
				if err != nil {
					global.Logger.Error("PDU 格式化失敗", zap.Error(err))
				} else {
					formattedPoints = append(formattedPoints, p...)
				}
				match = true
				break
			}
		}

		if !match {
			global.Logger.Error(
				"PDU 型號未匹配",
				zap.Any(global.Logs.MatchTag.Name, global.Logs.MatchTag),
			)
		}
	}

	// 添加到批次處理
	if err := factory.AddPoints(formattedPoints); err != nil {
		global.Logger.Error("添加數據到批次失敗", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{"message": "PDU 資料已接收"})
}

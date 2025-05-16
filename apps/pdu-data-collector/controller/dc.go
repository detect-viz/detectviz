package controller

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bimap-zbox/services/dc"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary  HealthCheck 簡單回應 HTTP 200 OK
// @Tags     collect
// @Accept   json
// @Produce  json
// @Success  200 {string} string
// @Router   /dc/status [get]
func HealthCheckDC(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}

// @Summary  接收 DC 發送的處理後數據
// @Tags     collect
// @Accept   json
// @Produce  json
// @Success  200 {string} string
// @Router   /dc/metrics [post]
func CollectDCMetrics(c *gin.Context) {

	// c.JSON(http.StatusBadRequest, gin.H{"message": "錯誤測試"})
	// return

	var dataList models.MetricsData
	if err := c.BindJSON(&dataList); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// 調用 service 層處理業務邏輯
	if err := dc.AddPoints(dataList.Metrics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process metrics",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "PDU 資料已接收"})
}

// @Summary  取得 PDU 或 Env 資料
// @Tags     dc
// @Accept   json
// @Produce  json
// @Param    name query string true "name" Enums(global_pdu_sync, global_env_sync)
// @Param    group path string true "group"
// @Success  200 {object} map[string]map[string]string
// @Router   /dc/env/{group} [get]
func GetEnvs(c *gin.Context) {
	var res map[string]map[string]string
	var err error

	group := c.Param("group")

	// 錯誤測試
	// c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve data", "details": err.Error()})
	// return

	switch c.Query("name") {
	case global.EnvConfig.Factory.InitGlobalData.PDU.Name:
		res, err = dc.GetPDUDataFromDB(group)
	case global.EnvConfig.Factory.InitGlobalData.Env.Name:
		res, err = dc.GetEnvsFromDB()
	case global.EnvConfig.Factory.InitGlobalData.Log.Name:
		res, err = dc.GetLogsFromDB()
	case global.EnvConfig.Factory.InitGlobalData.Job.Name:
		res, err = dc.GetJobsFromDB()
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary  取得規則列表
// @Tags     dc
// @Accept   json
// @Produce  json
// @Success  200 {object} []models.Rule
// @Router   /dc/rules [get]
func GetRuleList(c *gin.Context) {

	res, err := dc.GetRuleListFromDB()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}

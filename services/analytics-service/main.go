package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"analytics-service/entities"
	"analytics-service/internal/analyzer"
	"analytics-service/internal/config"

	"analytics-service/internal/processor"
	"analytics-service/internal/reporter"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg       *config.Config
	analyzer  *analyzer.AnomalyAnalyzer
	processor *processor.ModelProcessor
	reporter  *reporter.ReportGenerator
	logger    *log.Logger
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:       cfg,
		analyzer:  analyzer.NewAnomalyAnalyzer(cfg),
		processor: processor.NewModelProcessor(cfg),
		reporter:  reporter.NewReportGenerator(cfg),
		logger:    log.New(os.Stdout, "[API] ", log.LstdFlags),
	}
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	status := map[string]string{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleDetect(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. 獲取並驗證 profile_id
	profileID := c.Query("profile_id")
	if profileID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "missing profile_id"})
		return
	}

	if _, ok := s.cfg.Profiles[profileID]; !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid profile_id"})
		return
	}

	// 2. 解析請求數據
	var req entities.MetricRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request data"})
		return
	}

	// 3. 異常檢測
	s.logger.Printf("開始分析異常: profile_id=%s", profileID)
	anomalies, err := s.analyzer.Analyze(ctx, req.Data.Current, req.Data.History)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "anomaly detection failed"})
		return
	}

	// 4. 數據處理
	s.logger.Printf("處理異常數據: profile_id=%s", profileID)
	processedMetrics, err := s.processor.ProcessMetrics(profileID, anomalies.Anomalies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "metrics processing failed"})
		return
	}

	// 5. 生成報告
	s.logger.Printf("生成報告: profile_id=%s", profileID)
	report, err := s.reporter.GenerateReport(ctx, processedMetrics)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "report generation failed"})
		return
	}

	// 6. 返回結果
	c.JSON(http.StatusOK, report)
}

func (s *Server) writeError(w http.ResponseWriter, code int, message string) {
	resp := ErrorResponse{
		Error: message,
		Code:  http.StatusText(code),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	// 1. 加載配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 2. 初始化服務器
	server := NewServer(cfg)

	// 3. 設置路由
	r := gin.Default()

	// 健康檢查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// API 路由
	r.GET("/api/v1/metrics/:profile_id", server.handleMetrics)
	r.POST("/api/v1/detect", server.handleDetect)

	// 4. 配置 HTTP 服務器
	srv := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. 啟動服務器
	go func() {
		server.logger.Printf("Server starting on %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			server.logger.Fatalf("listen: %s\n", err)
		}
	}()

	// 6. 優雅關閉
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	server.logger.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		server.logger.Fatal("Server forced to shutdown:", err)
	}

	server.logger.Println("Server exited properly")
}

func (s *Server) handleMetrics(c *gin.Context) {
	profileID := c.Param("profile_id")
	metrics, err := s.analyzer.GetRequiredMetrics(profileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

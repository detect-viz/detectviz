package analyzer

import (
	"analytics-service/entities"
	"analytics-service/internal/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
)

// AnomalyServiceClient 異常檢測服務客戶端介面
type AnomalyServiceClient interface {
	CallDetector(ctx context.Context, model string, data interface{}) ([]entities.MetricData, error)
	DetectAnomaly(ctx context.Context, data []float64) ([]bool, error)
}

// anomalyClient 實現
type anomalyClient struct {
	cfg struct {
		Mode    string
		APIHost string
		CLIPath string
		Timeout int
		Retry   config.RetryConfig
		Models  config.ModelsConfig
	}
	client *http.Client
}

func NewAnomalyClient(cfg config.AnomalyService) AnomalyServiceClient {
	return &anomalyClient{
		cfg: struct {
			Mode    string
			APIHost string
			CLIPath string
			Timeout int
			Retry   config.RetryConfig
			Models  config.ModelsConfig
		}{
			Mode:    cfg.Mode,
			APIHost: cfg.APIHost,
			CLIPath: cfg.CLIPath,
			Timeout: cfg.Timeout,
			Retry:   cfg.Retry,
			Models:  cfg.Models,
		},
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// CallDetector 實現檢測方法
func (c *anomalyClient) CallDetector(ctx context.Context, model string, data interface{}) ([]entities.MetricData, error) {
	requestData := map[string]interface{}{
		"type":   model,
		"config": data.(map[string]interface{})["config"],
		"data":   data.(map[string]interface{})["data"],
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	var response []byte
	if c.cfg.Mode == "cli" {
		response, err = c.callCLI(ctx, jsonData)
	} else {
		response, err = c.callAPI(ctx, jsonData)
	}

	if err != nil {
		return nil, err
	}

	var result entities.MetricResponse
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %v", err)
	}

	return result.Data, nil
}

// DetectAnomaly 實現異常檢測方法
func (c *anomalyClient) DetectAnomaly(ctx context.Context, data []float64) ([]bool, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data failed: %v", err)
	}

	var response []byte
	if c.cfg.Mode == "cli" {
		response, err = c.callCLI(ctx, jsonData)
	} else {
		response, err = c.callAPI(ctx, jsonData)
	}

	if err != nil {
		return nil, err
	}

	var result []bool
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %v", err)
	}

	return result, nil
}

// callCLI 統一的 CLI 調用方法
func (c *anomalyClient) callCLI(ctx context.Context, data []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "python3", c.cfg.CLIPath)
	cmd.Stdin = bytes.NewBuffer(data)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("執行失敗: %v\n錯誤輸出: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// callAPI 統一的 API 調用方法
func (c *anomalyClient) callAPI(ctx context.Context, data []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/detect", c.cfg.APIHost)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	var response []byte
	var lastErr error

	for attempt := 0; attempt < c.cfg.Retry.MaxAttempts; attempt++ {
		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("call api failed: %v", err)
			time.Sleep(time.Duration(c.cfg.Retry.InitialInterval) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("api returned status: %d", resp.StatusCode)
			time.Sleep(time.Duration(c.cfg.Retry.InitialInterval) * time.Second)
			continue
		}

		response, err = io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("read response failed: %v", err)
			continue
		}

		return response, nil
	}

	return nil, lastErr
}

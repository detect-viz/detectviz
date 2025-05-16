package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/yaml.v3"

	"viot-web/internal/model"
	"viot-web/internal/view/pages"
)

// Config 定義設定檔結構
type Config struct {
	Pages []model.PageConfig `yaml:"pages"` // 頁面配置列表
}

// ColumnConfig 定義表格欄位配置
type ColumnConfig struct {
	Name       string `yaml:"name"`       // 欄位名稱
	Label      string `yaml:"label"`      // 顯示標籤
	Type       string `yaml:"type"`       // 欄位類型
	Width      string `yaml:"width"`      // 欄位寬度
	Sortable   bool   `yaml:"sortable"`   // 是否可排序
	Searchable bool   `yaml:"searchable"` // 是否可搜尋
}

// PageConfig 定義頁面設定結構
type PageConfig struct {
	Name        string                 `yaml:"name"`        // 路由與識別名稱（唯一）
	Title       string                 `yaml:"title"`       // 頁面標題
	Type        string                 `yaml:"type"`        // 頁面類型
	Icon        string                 `yaml:"icon"`        // Material Icons 圖示名稱
	Color       string                 `yaml:"color"`       // icon 文字顏色
	Path        string                 `yaml:"path"`        // 對應資料來源
	Description string                 `yaml:"description"` // 頁面描述
	Fields      []Field                `yaml:"fields"`      // 欄位清單
	Actions     []Action               `yaml:"actions"`     // 工具按鈕定義
	URL         string                 `yaml:"url"`         // 嵌入面板網址
	Panels      []Panel                `yaml:"panels"`      // 多面板配置
	Layout      Layout                 `yaml:"layout"`      // 布局配置
	CSS         []string               `yaml:"css"`         // 頁面專用 CSS
	JS          []string               `yaml:"js"`          // 頁面專用 JS
	Upload      bool                   `yaml:"upload"`      // 是否允許上傳
	Level       string                 `yaml:"level"`       // 最低顯示等級（用於日誌）
	Options     map[string]interface{} `yaml:"options"`     // 其他選項
	Columns     []ColumnConfig         `yaml:"columns"`     // 表格欄位配置
}

// Field 定義欄位設定結構
type Field struct {
	Name     string   `yaml:"name"`     // 欄位名稱
	Type     string   `yaml:"type"`     // 欄位類型
	Label    string   `yaml:"label"`    // 顯示標籤
	Options  []string `yaml:"options"`  // 選項列表
	Required bool     `yaml:"required"` // 是否必填
}

// Panel 定義面板設定結構
type Panel struct {
	Title   string `yaml:"title"`   // 面板標題
	URL     string `yaml:"url"`     // 面板網址
	Row     int    `yaml:"row"`     // 行位置
	Col     int    `yaml:"col"`     // 列位置
	ColSpan int    `yaml:"colSpan"` // 列跨度
	RowSpan int    `yaml:"rowSpan"` // 行跨度
}

// Layout 定義版面設定結構
type Layout struct {
	Type    string `yaml:"type"`    // 版面類型
	Columns int    `yaml:"columns"` // 列數
	Gap     int    `yaml:"gap"`     // 間距
}

// Action 定義動作設定結構
type Action struct {
	Label      string `yaml:"label"`       // 按鈕標籤
	Type       string `yaml:"type"`        // 動作類型
	ScriptPath string `yaml:"script_path"` // 腳本路徑
	Icon       string `yaml:"icon"`        // 按鈕圖示
	Desc       string `yaml:"desc"`        // 描述
	Disabled   bool   `yaml:"disabled"`    // 是否禁用
}

// DCConfig 定義資料中心設定結構
type DCConfig struct {
	Name  string   // 資料中心名稱
	Rooms []string // 房間列表
}

// FactoryPhaseConfig 定義廠區階段設定結構
type FactoryPhaseConfig struct {
	Factory string     // 廠區名稱
	Phase   string     // 階段名稱
	DCs     []DCConfig // 資料中心列表
}

// PDUStatus 定義 PDU 狀態結構
type PDUStatus struct {
	Name    string // 名稱
	IP      string // IP 位址
	Status  string // 狀態
	Message string // 訊息
}

// TemplateRenderer 定義模板渲染器結構
type TemplateRenderer struct {
	funcs template.FuncMap // 模板函數
}

// CSVData 定義 CSV 檔案資料結構
type CSVData struct {
	Headers []string            // 標題列
	Rows    [][]string          // 原始 CSV 資料列
	MapRows []map[string]string // 以 map 格式儲存的資料列，方便存取
}

// cacheEntry 定義快取項目結構
type cacheEntry struct {
	data       *CSVData  // CSV 資料
	lastAccess time.Time // 最後存取時間
}

// 快取管理相關常數
const cacheDuration = 5 * time.Minute // 快取有效時間

// 全域變數
var (
	config     Config               // 設定檔
	regionTree []FactoryPhaseConfig // 區域樹
)

// CSV 快取相關變數
var (
	csvCache   = make(map[string]cacheEntry) // CSV 快取
	cacheMutex sync.RWMutex                  // 快取互斥鎖
)

// --- Utility Functions ---

// contains 檢查字串陣列是否包含目標字串
func contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

// parseColonMap 解析冒號分隔的鍵值對字串
func parseColonMap(s string) map[string]int {
	result := make(map[string]int)
	for _, pair := range strings.Split(s, ",") {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			continue
		}
		val, _ := strconv.Atoi(parts[1])
		result[parts[0]] = val
	}
	return result
}

// parseRegionTree 解析區域樹結構
func parseRegionTree(factoryPhaseRaw, dcOfPhaseRaw, roomOfDCRaw string) []FactoryPhaseConfig {
	// 清理輸入字串
	factoryPhaseRaw = strings.TrimSpace(factoryPhaseRaw)
	dcOfPhaseRaw = strings.TrimSpace(dcOfPhaseRaw)
	roomOfDCRaw = strings.TrimSpace(roomOfDCRaw)

	// 解析廠區:階段組合
	fps := []string{}
	for _, fp := range strings.Split(factoryPhaseRaw, ",") {
		fp = strings.TrimSpace(fp)
		if fp != "" {
			fps = append(fps, fp)
		}
	}

	// 解析資料中心數量映射
	dcMap := make(map[string]int)
	for _, pair := range strings.Split(dcOfPhaseRaw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err == nil && val > 0 {
			dcMap[key] = val
		}
	}

	// 解析房間數量映射
	roomMap := make(map[string]int)
	for _, pair := range strings.Split(roomOfDCRaw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err == nil && val > 0 {
			roomMap[key] = val
		}
	}

	// 構建區域樹
	var results []FactoryPhaseConfig
	for _, fp := range fps {
		parts := strings.Split(fp, ":")
		if len(parts) != 2 {
			continue
		}
		factory := strings.TrimSpace(parts[0])
		phase := strings.TrimSpace(parts[1])
		key := factory + phase
		dcCount := dcMap[key]

		if dcCount > 0 {
			var dcs []DCConfig
			for i := 1; i <= dcCount; i++ {
				dcName := fmt.Sprintf("DC%d", i)
				dcFull := key + dcName
				roomCount := roomMap[dcFull]

				if roomCount > 0 {
					var rooms []string
					for j := 1; j <= roomCount; j++ {
						rooms = append(rooms, fmt.Sprintf("R%d", j))
					}
					dcs = append(dcs, DCConfig{
						Name:  dcName,
						Rooms: rooms,
					})
				}
			}

			if len(dcs) > 0 {
				results = append(results, FactoryPhaseConfig{
					Factory: factory,
					Phase:   phase,
					DCs:     dcs,
				})
			}
		}
	}

	// 記錄解析結果
	log.Printf("🌳 區域樹解析結果: %+v", results)

	return results
}

// loadConfig 載入設定檔
func loadConfig() ([]model.PageConfig, error) {
	data, err := os.ReadFile("config/pages.yaml")
	if err != nil {
		return nil, err
	}

	var pages []model.PageConfig
	if err := yaml.Unmarshal(data, &pages); err != nil {
		return nil, err
	}

	return pages, nil
}

// loadCSV 載入 CSV 檔案
func loadCSV(path string) ([]map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	var rows []map[string]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// 驗證欄位數量
		if len(record) != len(headers) {
			// 在 DEBUG 或 ERROR 級別記錄錯誤
			if strings.ToUpper(os.Getenv("LOG_LEVEL")) == "DEBUG" || strings.ToUpper(os.Getenv("LOG_LEVEL")) == "ERROR" {
				log.Printf("[ERROR] CSV 欄位數錯誤，期望 %d 欄，實際為 %d：%v", len(headers), len(record), record)
			}
			continue
		}

		// 將記錄轉換為 map
		row := make(map[string]string)
		for i, h := range headers {
			if i < len(record) {
				row[h] = record[i]
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// saveCSV 儲存 CSV 檔案
func saveCSV(path string, headers []string, rows []map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 寫入標題列
	if err := writer.Write(headers); err != nil {
		return err
	}

	// 寫入資料列
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = row[h]
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// --- Template Functions ---

// Render 渲染模板
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// 使用 templ 渲染
	if component, ok := data.(templ.Component); ok {
		return component.Render(c.Request().Context(), w)
	}
	return fmt.Errorf("invalid template data type: %T", data)
}

// --- Authentication Handlers ---

// handleLogin 處理登入請求
func handleLogin(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		return c.Render(http.StatusOK, "login", pages.Login())
	}

	username := c.FormValue("username")
	password := c.FormValue("password")
	envUser, found := os.LookupEnv("LOGIN_USER")
	if !found {
		envUser = "admin"
	}
	envPass, found := os.LookupEnv("LOGIN_PASS")
	if !found {
		envPass = "admin"
	}
	if username == envUser && password == envPass {
		cookie := new(http.Cookie)
		cookie.Name = "session"
		cookie.Value = "authenticated"
		cookie.Path = "/"
		maxAgeStr := os.Getenv("COOKIE_MAX_AGE")
		maxAge, _ := strconv.Atoi(maxAgeStr)
		if maxAge == 0 {
			maxAge = 604800
		}
		cookie.MaxAge = maxAge
		c.SetCookie(cookie)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	return c.Redirect(http.StatusSeeOther, "/login?error=1")
}

// handleLogout 處理登出請求
func handleLogout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "session"
	cookie.Value = ""
	cookie.Path = "/"
	cookie.MaxAge = -1
	c.SetCookie(cookie)
	return c.Redirect(http.StatusSeeOther, "/login")
}

// --- Page Handlers ---

// handlePage 處理頁面請求
func handlePage(c echo.Context) error {
	// 從 URL 參數獲取頁面名稱
	pageName := c.Param("name")
	if pageName == "" {
		return c.String(http.StatusBadRequest, "Page name is required")
	}

	// 查找頁面配置
	var page *model.PageConfig
	for _, p := range config.Pages {
		if p.Name == pageName {
			page = &p
			break
		}
	}
	if page == nil {
		return c.String(http.StatusNotFound, "找不到頁面")
	}

	// 特殊處理環境變數頁面
	if page.Type == "env" {
		// 讀取 .env 檔案
		raw, err := os.ReadFile(page.Path)
		if err != nil {
			return c.String(http.StatusInternalServerError, "讀取環境變數失敗")
		}

		// 解析環境變數
		envVars := make(map[string]string)
		lines := strings.Split(string(raw), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				envVars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}

		return pages.Env(*page, envVars).Render(c.Request().Context(), c.Response().Writer)
	}

	// 對於其他頁面，載入 CSV 資料
	csvData, err := loadCSVWithCache(page.Path)
	if err != nil {
		return fmt.Errorf("failed to load CSV data: %v", err)
	}

	return pages.CSV(*page, csvData.MapRows, config.Pages).Render(c.Request().Context(), c.Response().Writer)
}

// handlePageRows 處理取得頁面資料列的請求
func handlePageRows(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	name := c.Param("name")
	var page *model.PageConfig
	for _, p := range config.Pages {
		if p.Name == name {
			page = &p
			break
		}
	}
	if page == nil {
		return c.String(http.StatusNotFound, "Page not found")
	}

	rows, err := loadCSVWithCache(page.Path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, rows.MapRows)
}

// handlePageEdit 處理編輯頁面項目的請求
func handlePageEdit(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	name := c.Param("name")
	id := c.Param("id")

	var page *model.PageConfig
	for _, p := range config.Pages {
		if p.Name == name {
			page = &p
			break
		}
	}
	if page == nil {
		return c.String(http.StatusNotFound, "Page not found")
	}

	rows, err := loadCSV(page.Path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "載入資料失敗")
	}

	var row map[string]string
	for _, r := range rows {
		if r["ip"] == id {
			row = r
			break
		}
	}
	if row == nil {
		return c.String(http.StatusNotFound, "資料未找到")
	}

	return c.Render(http.StatusOK, "edit_row", map[string]interface{}{
		"Row":    row,
		"Fields": page.Fields,
	})
}

// handlePageSave 處理儲存頁面資料的請求
func handlePageSave(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	name := c.Param("name")
	id := c.Param("id")

	var page *model.PageConfig
	for _, p := range config.Pages {
		if p.Name == name {
			page = &p
			break
		}
	}
	if page == nil {
		return c.String(http.StatusNotFound, "Page not found")
	}

	csvData, err := loadCSVWithCache(page.Path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 更新 MapRows 中的資料列
	rowIndex := -1
	for i, row := range csvData.MapRows {
		if row["id"] == id {
			rowIndex = i
			break
		}
	}

	if rowIndex == -1 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Row not found"})
	}

	// 使用新值更新資料列
	formData := c.Request().PostForm
	for key, value := range formData {
		csvData.MapRows[rowIndex][key] = value[0]
	}

	// 儲存更新後的資料
	err = saveCSV(page.Path, csvData.Headers, csvData.MapRows)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 使快取失效
	invalidateCSVCache(page.Path)
	return c.JSON(http.StatusOK, csvData.MapRows[rowIndex])
}

// handlePageCreate 處理建立新頁面項目的請求
func handlePageCreate(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	name := c.Param("name")
	var page *model.PageConfig
	for _, p := range config.Pages {
		if p.Name == name {
			page = &p
			break
		}
	}
	if page == nil {
		return c.String(http.StatusNotFound, "Page not found")
	}

	// 獲取當前時間戳記用於建立和更新時間
	now := time.Now().Format("2006-01-02 15:04:05")

	// 獲取表單欄位，自動處理時間戳記
	newRow := make(map[string]string)
	for _, field := range page.Fields {
		if field.Name == "create_at" || field.Name == "update_at" {
			// 自動設定建立和更新時間
			newRow[field.Name] = now
		} else {
			// 其他欄位從表單獲取
			newRow[field.Name] = c.FormValue(field.Name)
		}
	}

	// 讀取舊資料，附加並寫回
	csvData, err := loadCSVWithCache(page.Path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	csvData.MapRows = append(csvData.MapRows, newRow)

	// 準備標題列
	var headers []string
	for _, f := range page.Fields {
		headers = append(headers, f.Name)
	}

	// 使用新的 saveCSV 函數儲存資料
	if err := saveCSV(page.Path, headers, csvData.MapRows); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 使快取失效
	invalidateCSVCache(page.Path)

	return c.JSON(http.StatusOK, newRow)
}

// handlePageDelete 處理刪除頁面項目的請求
func handlePageDelete(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	name := c.Param("name")
	id := c.Param("id")

	var page *model.PageConfig
	for _, p := range config.Pages {
		if p.Name == name {
			page = &p
			break
		}
	}
	if page == nil {
		return c.String(http.StatusNotFound, "Page not found")
	}

	csvData, err := loadCSVWithCache(page.Path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 尋找並移除資料列
	newRows := make([]map[string]string, 0)
	for _, row := range csvData.MapRows {
		if row["id"] != id {
			newRows = append(newRows, row)
		}
	}

	// 儲存更新後的資料
	err = saveCSV(page.Path, csvData.Headers, newRows)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 使快取失效
	invalidateCSVCache(page.Path)
	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

// handleCreateForm 處理建立表單的請求
func handleCreateForm(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	name := c.Param("name")
	for _, page := range config.Pages {
		if page.Name == name {
			return c.Render(http.StatusOK, "create_form", map[string]interface{}{
				"Fields": page.Fields,
			})
		}
	}
	return c.String(http.StatusNotFound, "Page not found")
}

// --- Environment Handlers ---

// handleEnvDelete 處理刪除環境變數的請求
func handleEnvDelete(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	id := c.Param("id")

	path := "data/env.csv"
	rows, err := loadCSV(path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "讀取資料失敗")
	}

	var newRows []map[string]string
	for _, r := range rows {
		if r["key"] != id {
			newRows = append(newRows, r)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "寫入失敗")
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Write([]string{"key", "value"})
	for _, row := range newRows {
		writer.Write([]string{row["key"], row["value"]})
	}
	writer.Flush()

	return c.String(http.StatusOK, "")
}

// handleEnvSave 處理儲存環境變數的請求
func handleEnvSave(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	// 讀取原始 .env 檔案
	raw, err := os.ReadFile("data/.env")
	if err != nil {
		return c.String(http.StatusInternalServerError, "讀取環境變數失敗")
	}
	lines := strings.Split(string(raw), "\n")

	// 解析現有環境變數
	data := make(map[string]string)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			data[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// 更新環境變數
	keys := c.Request().PostForm["key"]
	values := c.Request().PostForm["value"]
	for i := range keys {
		if keys[i] != "" {
			data[keys[i]] = values[i]
		}
	}

	// 重新組裝 .env 檔案，保留原始格式和註解
	newLines := []string{}
	seenKeys := make(map[string]bool)

	// 處理原始行
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			// 保留空行和註解
			newLines = append(newLines, line)
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			seenKeys[key] = true

			// 更新值
			if val, ok := data[key]; ok {
				newLines = append(newLines, fmt.Sprintf("%s=%s", key, val))
			} else {
				newLines = append(newLines, line)
			}
		}
	}

	// 添加新的環境變數
	for key, val := range data {
		if !seenKeys[key] {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, val))
		}
	}

	// 寫回檔案
	if err := os.WriteFile("data/.env", []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return c.String(http.StatusInternalServerError, "寫入環境變數失敗")
	}

	return c.Redirect(http.StatusSeeOther, "/page/env")
}

// --- Log Handlers ---

// handleLogAPI 處理日誌 API 的請求
func handleLogAPI(c echo.Context) error {
	name := c.QueryParam("name")
	var page *model.PageConfig
	for _, p := range config.Pages {
		if p.Name == name {
			page = &p
			break
		}
	}
	if page == nil {
		return c.String(http.StatusNotFound, "Page not found")
	}

	file, err := os.Open(page.Path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "log 檔案開啟失敗")
	}
	defer file.Close()

	// 定義日誌等級順序
	levelOrder := map[string]int{
		"DEBUG": 1,
		"INFO":  2,
		"WARN":  3,
		"ERROR": 4,
	}

	// 獲取目標日誌等級
	targetLevel := levelOrder[strings.ToUpper(os.Getenv("LOG_LEVEL"))]
	if targetLevel == 0 {
		targetLevel = 2 // 預設為 INFO
	}

	var html string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// 根據日誌等級過濾
		for level, value := range levelOrder {
			if strings.Contains(line, "["+level+"]") && value >= targetLevel {
				// 為不同等級添加不同樣式
				html += `<div class="log-line" data-level="` + level + `">` + template.HTMLEscapeString(line) + `</div>` + "\n"
				break
			}
		}
	}
	return c.HTML(http.StatusOK, html)
}

// --- Tool Handlers ---

// handleToolExec 處理執行工具的請求
func handleToolExec(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	tool := c.Param("tool")
	username, _ := c.Cookie("username")

	// 記錄執行日誌 (DEBUG 等級)
	if strings.ToUpper(os.Getenv("LOG_LEVEL")) == "DEBUG" {
		log.Printf("[EXEC] 執行人: %s, 指令: %s", username, tool)
	}

	var path string
	for _, page := range config.Pages {
		if page.Type == "exec" {
			for _, action := range page.Actions {
				if strings.HasSuffix(action.ScriptPath, tool+".sh") {
					path = action.ScriptPath
					break
				}
			}
		}
	}
	if path == "" {
		return c.String(http.StatusNotFound, "找不到對應工具")
	}

	// 執行指令
	cmd := exec.Command("bash", "-c", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("執行錯誤：%s\n%s", err.Error(), output)
		return c.String(http.StatusInternalServerError, "執行失敗："+err.Error()+"\n"+string(output))
	}

	// 記錄執行結果
	log.Printf("執行成功：%s\n%s", tool, output)
	return c.Render(http.StatusOK, "partials/exec_dialog.html", map[string]interface{}{
		"Output": string(output),
	})
}

// handleExecPage 處理執行頁面的請求
func handleExecPage(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	for _, p := range config.Pages {
		if p.Name == "tools" {
			return c.Render(http.StatusOK, "layout", map[string]interface{}{
				"Title":   p.Title,
				"Actions": p.Actions,
				"User":    "admin",
				"Menus":   config.Pages,
				"Current": p.Name,
			})
		}
	}
	return c.String(http.StatusNotFound, "Page not found")
}

// --- Status Map Handler ---

// serveStatusMapPage 提供狀態地圖頁面
func serveStatusMapPage(c echo.Context, page model.PageConfig) error {
	// 讀取 CSV 資料
	fmt.Printf("page.Path: %v\n", page.Path)
	rows, err := loadCSV(page.Path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "載入資料失敗")
	}

	// 從 regionTree 提取廠區和樓層列表
	factoryList := []string{}
	phaseList := []string{}
	factorySet := make(map[string]bool)
	phaseSet := make(map[string]bool)

	for _, item := range regionTree {
		if !factorySet[item.Factory] {
			factoryList = append(factoryList, item.Factory)
			factorySet[item.Factory] = true
		}
		if !phaseSet[item.Phase] {
			phaseList = append(phaseList, item.Phase)
			phaseSet[item.Phase] = true
		}
	}
	sort.Strings(factoryList)
	sort.Strings(phaseList)

	// 獲取查詢參數
	var dcQuery, roomQuery []string
	var factoryQuery, phaseQuery string

	if c.Request().Method == http.MethodPost {
		c.Request().ParseForm()
		dc := c.FormValue("dc")
		room := c.FormValue("room")
		if dc != "" {
			dcQuery = strings.Split(dc, ",")
		}
		if room != "" {
			roomQuery = strings.Split(room, ",")
		}
		factoryQuery = c.FormValue("factory")
		phaseQuery = c.FormValue("phase")
	} else {
		dc := c.QueryParam("dc")
		room := c.QueryParam("room")
		if dc != "" {
			dcQuery = strings.Split(dc, ",")
		}
		if room != "" {
			roomQuery = strings.Split(room, ",")
		}
		factoryQuery = c.QueryParam("factory")
		phaseQuery = c.QueryParam("phase")
	}

	// 預設選擇第一行 (進入畫面時的預設參數)
	if len(factoryQuery) == 0 && len(regionTree) > 0 {
		factoryQuery = regionTree[0].Factory
		phaseQuery = regionTree[0].Phase
		if len(regionTree[0].DCs) > 0 {
			dcQuery = []string{regionTree[0].DCs[0].Name}
			if len(regionTree[0].DCs[0].Rooms) > 0 {
				roomQuery = []string{regionTree[0].DCs[0].Rooms[0]}
			}
		}
	}

	// 過濾掉預設的中文預設值
	if factoryQuery == "" || strings.HasPrefix(factoryQuery, "選擇") {
		if len(regionTree) > 0 {
			factoryQuery = regionTree[0].Factory
		}
	}
	if phaseQuery == "" || strings.HasPrefix(phaseQuery, "選擇") {
		if len(regionTree) > 0 {
			phaseQuery = regionTree[0].Phase
		}
	}
	if len(dcQuery) == 0 || (len(dcQuery) == 1 && strings.HasPrefix(dcQuery[0], "選擇")) {
		if len(regionTree) > 0 && len(regionTree[0].DCs) > 0 {
			dcQuery = []string{regionTree[0].DCs[0].Name}
		}
	}
	if len(roomQuery) == 0 || (len(roomQuery) == 1 && strings.HasPrefix(roomQuery[0], "選擇")) {
		if len(regionTree) > 0 && len(regionTree[0].DCs) > 0 && len(regionTree[0].DCs[0].Rooms) > 0 {
			roomQuery = []string{regionTree[0].DCs[0].Rooms[0]}
		}
	}

	// 過濾資料列
	filteredRows := []map[string]string{}
	countRed := 0
	countYellow := 0
	countGreen := 0

	for _, row := range rows {
		// 檢查廠區和樓層
		if factoryQuery != "" && row["factory"] != factoryQuery {
			continue
		}
		if phaseQuery != "" && row["phase"] != phaseQuery {
			continue
		}

		// 處理資料中心選擇
		if len(dcQuery) > 0 {
			if contains(dcQuery, "ALL") {
				// 如果選擇了 ALL，不過濾 DC
			} else if !contains(dcQuery, row["dc"]) {
				continue
			}
		}

		// 處理房間選擇
		if len(roomQuery) > 0 && !contains(roomQuery, row["room"]) {
			continue
		}

		// 計數狀態
		switch row["status"] {
		case "ERROR":
			countRed++
		case "WARN":
			countYellow++
		case "OK":
			countGreen++
		}

		filteredRows = append(filteredRows, row)
	}

	// 計數 dcList 和 roomList
	dcSet := make(map[string]bool)
	roomSet := make(map[string]bool)
	for _, row := range filteredRows {
		dcSet[row["dc"]] = true
		roomSet[row["room"]] = true
	}
	dcList := []string{}
	for dc := range dcSet {
		dcList = append(dcList, dc)
	}
	sort.Strings(dcList)
	roomList := []string{}
	for room := range roomSet {
		roomList = append(roomList, room)
	}
	sort.Strings(roomList)

	// 網格結構組裝
	rowList := []string{}
	colList := []string{}
	rowSet := map[string]bool{}
	colSet := map[string]bool{}
	for _, row := range filteredRows {
		rowSet[row["row"]] = true
		colSet[row["col"]] = true
	}
	for row := range rowSet {
		rowList = append(rowList, row)
	}
	sort.SliceStable(rowList, func(i, j int) bool {
		if len(rowList[i]) == 1 && len(rowList[j]) == 1 {
			return rowList[i] < rowList[j]
		}
		if rowList[i] == "AA" {
			return false
		}
		if rowList[j] == "AA" {
			return true
		}
		return rowList[i] < rowList[j]
	})
	for col := range colSet {
		colList = append(colList, col)
	}
	sort.Strings(colList)

	// 建立網格結構
	grid := make(map[string]map[string]map[string]*PDUStatus)
	for _, row := range filteredRows {
		r := row["row"]
		c := row["col"]
		if grid[r] == nil {
			grid[r] = make(map[string]map[string]*PDUStatus)
		}
		if grid[r][c] == nil {
			grid[r][c] = make(map[string]*PDUStatus)
		}

		pduStatus := &PDUStatus{
			Name:    row["name"],
			IP:      row["ip"],
			Status:  row["status"],
			Message: row["message"],
		}

		if row["side"] == "L" {
			grid[r][c]["Left"] = pduStatus
		} else {
			grid[r][c]["Right"] = pduStatus
		}
	}

	result := map[string]interface{}{
		"Title":           "設備狀態",
		"PageTitle":       "設備狀態",
		"Grid":            grid,
		"DCList":          dcList,
		"RoomList":        roomList,
		"CountRed":        countRed,
		"CountYellow":     countYellow,
		"CountGreen":      countGreen,
		"Fields":          page.Fields,
		"RowList":         rowList,
		"ColList":         colList,
		"RegionTree":      regionTree,
		"FactoryList":     factoryList,
		"PhaseList":       phaseList,
		"SelectedFactory": factoryQuery,
		"SelectedPhase":   phaseQuery,
		"SelectedDCs":     dcQuery,
		"SelectedRooms":   roomQuery,
		"Menus":           config.Pages,
		"Current":         "status",
	}

	// 根據模板參數，決定渲染完整頁面或片段
	if c.QueryParam("partial") == "1" || c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, "status_map_fragment", result)
	}
	return c.Render(http.StatusOK, "status_map", result)
}

// handlePageUpload 處理上傳檔案的請求
func handlePageUpload(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		return c.String(http.StatusUnauthorized, "")
	}

	name := c.Param("name")
	var page *model.PageConfig
	for _, p := range config.Pages {
		if p.Name == name {
			page = &p
			break
		}
	}
	if page == nil {
		return c.String(http.StatusNotFound, "Page not found")
	}

	// 檢查是否允許上傳
	if !page.Upload {
		return c.String(http.StatusForbidden, "此頁面不允許上傳文件")
	}

	// 獲取上傳的檔案
	file, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusBadRequest, "獲取上傳文件失敗")
	}

	// 檢查檔案副檔名
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		return c.String(http.StatusBadRequest, "僅接受 CSV 文件")
	}

	// 安全性檢查檔案路徑
	cleanPath := filepath.Clean(strings.TrimPrefix(file.Filename, "./"))
	if !strings.HasPrefix(cleanPath, "data/") && !strings.HasPrefix(cleanPath, "scripts/") {
		return c.String(http.StatusForbidden, "禁止存取該路徑")
	}

	// 開啟來源檔案
	src, err := file.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "打開上傳文件失敗")
	}
	defer src.Close()

	// 建立目標檔案
	dst, err := os.Create(page.Path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "創建目標文件失敗")
	}
	defer dst.Close()

	// 複製檔案內容
	if _, err = io.Copy(dst, src); err != nil {
		return c.String(http.StatusInternalServerError, "保存文件失敗")
	}

	// 使快取失效
	invalidateCSVCache(page.Path)

	return c.Redirect(http.StatusSeeOther, "/page/"+name)
}

// --- Main Function ---

// main 程式進入點
func main() {
	// 載入 .env 環境變數
	_ = godotenv.Load("data/.env")
	regionTree = parseRegionTree(
		os.Getenv("SET_FACTORY_PHASE"),
		os.Getenv("SET_DC_OF_PHASE"),
		os.Getenv("SET_ROOM_OF_DC"),
	)
	pages, err := loadConfig()
	if err != nil {
		log.Fatalf("無法載入設定檔：%v", err)
	}
	config.Pages = pages

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if path != "/login" && !strings.HasPrefix(path, "/static") && !strings.HasPrefix(path, "/adminlte") {
				cookie, err := c.Cookie("session")
				if err != nil || cookie.Value != "authenticated" {
					log.Printf("🔐 未登入：%s ➜ 導向 /login", path)
					return c.Redirect(http.StatusSeeOther, "/login")
				}
			}
			return next(c)
		}
	})
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Grafana-Org-Id",
			"X-Panel-Id",
			"X-Dashboard-Id",
			"X-CSRF-Token",
			"HX-Request",
			"HX-Trigger",
			"HX-Target",
			"HX-Current-URL",
		},
		ExposeHeaders:    []string{"Content-Length", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))
	e.Use(middleware.Recover())

	e.Static("/static", "static")
	e.Static("/adminlte", "static/adminlte")

	funcMap := template.FuncMap{
		"base": func(path string) string {
			return path[strings.LastIndex(path, "/")+1:]
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			d := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				key := values[i].(string)
				d[key] = values[i+1]
			}
			return d
		},
		"contains": func(arr []string, target string) bool {
			for _, v := range arr {
				if v == target {
					return true
				}
			}
			return false
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}

	// 初始化 templ 模板
	renderer := &TemplateRenderer{
		funcs: funcMap,
	}
	e.Renderer = renderer

	// 路由設定
	e.GET("/login", handleLogin)
	e.POST("/login", handleLogin)
	e.GET("/logout", handleLogout)
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/page/registry")
	})
	e.GET("/page/:name", handlePage)
	e.GET("/page/status", func(c echo.Context) error {
		for _, p := range config.Pages {
			if p.Name == "status" {
				return serveStatusMapPage(c, p)
			}
		}
		return c.String(http.StatusNotFound, "找不到頁面")
	})
	e.POST("/page/status", func(c echo.Context) error {
		for _, p := range config.Pages {
			if p.Name == "status" {
				return serveStatusMapPage(c, p)
			}
		}
		return c.String(http.StatusNotFound, "找不到頁面")
	})
	e.GET("/page/:name/rows", handlePageRows)
	e.GET("/page/:name/edit/:id", handlePageEdit)
	e.POST("/page/:name/save/:id", handlePageSave)
	e.POST("/page/:name/create", handlePageCreate)
	e.POST("/page/:name/delete/:id", handlePageDelete)
	e.GET("/page/:name/create-form", handleCreateForm)
	e.POST("/page/env/delete/:id", handleEnvDelete)
	e.POST("/page/env/save", handleEnvSave)
	e.GET("/api/log", handleLogAPI)
	e.POST("/api/exec/:tool", handleToolExec)
	e.GET("/page/tools", handleExecPage)
	e.GET("/api/regions", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"regionTree": regionTree,
		})
	})
	e.POST("/page/:name/upload", handlePageUpload)

	// CSRF 中介軟體
	trustedOrigins := os.Getenv("CSRF_TRUSTED_ORIGINS")
	if trustedOrigins != "" {
		e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
			CookieName:     "_csrf",
			CookiePath:     "/",
			CookieHTTPOnly: true,
			CookieSecure:   false,
			TokenLookup:    "header:X-CSRF-Token",
			ContextKey:     "csrf",
			Skipper: func(c echo.Context) bool {
				return strings.HasPrefix(c.Request().URL.Path, "/api/")
			},
		}))
		log.Printf("✅ 已套用 CSRF 信任來源: %s", trustedOrigins)
	}

	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

// calculateStatusCounts 計算狀態數量
func calculateStatusCounts(data *CSVData) map[string]int {
	counts := make(map[string]int)
	statusIndex := -1

	// 尋找狀態欄位索引
	for i, header := range data.Headers {
		if header == "status" {
			statusIndex = i
			break
		}
	}

	if statusIndex >= 0 {
		for _, row := range data.Rows {
			if len(row) > statusIndex {
				status := row[statusIndex]
				counts[status]++
			}
		}
	}

	return counts
}

// loadEnvFile 載入環境變數檔案
func loadEnvFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	envMap := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envMap[key] = value
		}
	}

	return envMap, nil
}

// invalidateCSVCache 從快取中移除檔案
func invalidateCSVCache(filePath string) {
	cacheMutex.Lock()
	delete(csvCache, filePath)
	cacheMutex.Unlock()
}

// loadCSVWithCache 載入 CSV 資料並支援快取
func loadCSVWithCache(filePath string) (*CSVData, error) {
	// 先檢查快取
	cacheMutex.RLock()
	if entry, exists := csvCache[filePath]; exists {
		if time.Since(entry.lastAccess) < cacheDuration {
			cacheMutex.RUnlock()
			return entry.data, nil
		}
	}
	cacheMutex.RUnlock()

	// 載入 CSV 檔案
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("無法開啟 CSV 檔案: %v", err)
	}
	defer file.Close()

	// 讀取 CSV 資料
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 允許變數數量的欄位

	// 讀取標題列
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("無法讀取 CSV 標題列: %v", err)
	}

	// 讀取所有資料列
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("無法讀取 CSV 資料列: %v", err)
	}

	// 將資料列轉換為 map 格式以便存取
	mapRows := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		rowMap := make(map[string]string)
		for i, value := range row {
			if i < len(headers) {
				rowMap[headers[i]] = value
			}
		}
		mapRows = append(mapRows, rowMap)
	}

	// 建立 CSVData 結構
	csvData := &CSVData{
		Headers: headers,
		Rows:    rows,
		MapRows: mapRows,
	}

	// 更新快取
	cacheMutex.Lock()
	csvCache[filePath] = cacheEntry{
		data:       csvData,
		lastAccess: time.Now(),
	}
	cacheMutex.Unlock()

	return csvData, nil
}

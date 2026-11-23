package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"docklog/cli"
	"docklog/config"
	"docklog/db"
	"docklog/services"
	"docklog/stats"

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

type LogcConfig struct {
	FontSize         string `json:"font_size" yaml:"font_size"`
	CompactMode      bool   `json:"compact_mode" yaml:"compact_mode"`
	ShowTimestamps   bool   `json:"show_timestamps" yaml:"show_timestamps"`
	SoftWrap         bool   `json:"soft_wrap" yaml:"soft_wrap"`
	HighlightErrors  bool   `json:"highlight_errors" yaml:"highlight_errors"`
	TimeFormat       string `json:"time_format" yaml:"time_format"`
	DateFormat       string `json:"date_format" yaml:"date_format"`
	ShowStreamLabels bool   `json:"show_stream_labels" yaml:"show_stream_labels"`
	ShowLineNumbers  bool   `json:"show_line_numbers" yaml:"show_line_numbers"`
	AutoScrollMode   string `json:"auto_scroll_mode" yaml:"auto_scroll_mode"`
}

type GeneralConfig struct {
	Theme           string `json:"theme" yaml:"theme"`
	Language        string `json:"language" yaml:"language"`
	FollowRecreated string `json:"follow_recreated" yaml:"follow_recreated"`
	SearchIntercept bool   `json:"search_intercept" yaml:"search_intercept"`
}

var (
	logcMu    sync.RWMutex
	generalMu sync.RWMutex
)

// getSettingsDir returns the directory where YAML configuration files are stored.
// If SETTINGS_DIR is set, it uses that directory.
// Otherwise, if /app/data exists, it uses /app/data/settings for persistence with -v ./data:/app/data.
// If a DB_PATH is set (e.g. /app/data/docklog.db or custom directory), it checks that directory.
// Fallback is "settings".
func getSettingsDir() string {
	if custom := strings.TrimSpace(os.Getenv("SETTINGS_DIR")); custom != "" {
		return custom
	}
	if dbPath := strings.TrimSpace(os.Getenv("DB_PATH")); dbPath != "" && dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		candidate := filepath.Join(dir, "settings")
		if _, err := os.Stat(dir); err == nil {
			return candidate
		}
	}
	if info, err := os.Stat("/app/data"); err == nil && info.IsDir() {
		return "/app/data/settings"
	}
	return "settings"
}

func getLogcFilePath() string {
	return filepath.Join(getSettingsDir(), "log.yaml")
}

func getLogcAltFilePath() string {
	return filepath.Join(getSettingsDir(), "logc.yaml")
}

func getGeneralFilePath() string {
	return filepath.Join(getSettingsDir(), "general.yaml")
}

func defaultLogcConfig() LogcConfig {
	return LogcConfig{
		FontSize:         "small",
		CompactMode:      false,
		ShowTimestamps:   true,
		SoftWrap:         true,
		HighlightErrors:  true,
		TimeFormat:       "auto",
		DateFormat:       "YYYY-MM-DD",
		ShowStreamLabels: true,
		ShowLineNumbers:  false,
		AutoScrollMode:   "smart",
	}
}

func defaultGeneralConfig() GeneralConfig {
	return GeneralConfig{
		Theme:           "auto",
		Language:        "en-US",
		FollowRecreated: "delayed",
		SearchIntercept: true,
	}
}

func readGeneralConfig() GeneralConfig {
	generalMu.RLock()
	defer generalMu.RUnlock()

	cfg := defaultGeneralConfig()
	primaryPath := getGeneralFilePath()
	data, err := os.ReadFile(primaryPath)
	if err != nil && primaryPath != filepath.Join("settings", "general.yaml") {
		// Fallback to initial image template in "settings/general.yaml" if target file doesn't exist yet
		data, err = os.ReadFile(filepath.Join("settings", "general.yaml"))
	}
	if err != nil {
		return cfg
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaultGeneralConfig()
	}
	return cfg
}

func saveGeneralConfig(cfg GeneralConfig) error {
	generalMu.Lock()
	defer generalMu.Unlock()

	targetPath := getGeneralFilePath()
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(targetPath, data, 0644)
}

func readLogcConfig() LogcConfig {
	logcMu.RLock()
	defer logcMu.RUnlock()

	cfg := defaultLogcConfig()
	primaryPath := getLogcFilePath()
	data, err := os.ReadFile(primaryPath)
	if err != nil {
		data, err = os.ReadFile(getLogcAltFilePath())
	}
	if err != nil && primaryPath != filepath.Join("settings", "log.yaml") {
		// Fallback to initial image template in "settings/log.yaml" or "settings/logc.yaml"
		data, err = os.ReadFile(filepath.Join("settings", "log.yaml"))
		if err != nil {
			data, err = os.ReadFile(filepath.Join("settings", "logc.yaml"))
		}
	}
	if err != nil {
		return cfg
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaultLogcConfig()
	}
	return cfg
}

func saveLogcConfig(cfg LogcConfig) error {
	logcMu.Lock()
	defer logcMu.Unlock()

	targetPath := getLogcFilePath()
	altPath := getLogcAltFilePath()
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	_ = os.WriteFile(altPath, data, 0644)
	return os.WriteFile(targetPath, data, 0644)
}

func (s *Server) registerSystemRoutes(r *echo.Group) {
	r.GET("/system/logc", func(c echo.Context) error {
		cfg := readLogcConfig()
		return c.JSON(http.StatusOK, cfg)
	})

	r.POST("/system/logc", func(c echo.Context) error {
		current := readLogcConfig()
		var incoming map[string]interface{}
		if err := c.Bind(&incoming); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
		}

		if v, ok := incoming["font_size"].(string); ok && v != "" {
			current.FontSize = v
		}
		if v, ok := incoming["compact_mode"].(bool); ok {
			current.CompactMode = v
		}
		if v, ok := incoming["show_timestamps"].(bool); ok {
			current.ShowTimestamps = v
		}
		if v, ok := incoming["soft_wrap"].(bool); ok {
			current.SoftWrap = v
		}
		if v, ok := incoming["highlight_errors"].(bool); ok {
			current.HighlightErrors = v
		}
		if v, ok := incoming["time_format"].(string); ok && v != "" {
			current.TimeFormat = v
		}
		if v, ok := incoming["date_format"].(string); ok && v != "" {
			current.DateFormat = v
		}
		if v, ok := incoming["show_stream_labels"].(bool); ok {
			current.ShowStreamLabels = v
		}
		if v, ok := incoming["show_line_numbers"].(bool); ok {
			current.ShowLineNumbers = v
		}
		if v, ok := incoming["auto_scroll_mode"].(string); ok && v != "" {
			current.AutoScrollMode = v
		}

		if err := saveLogcConfig(current); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save log settings: " + err.Error()})
		}

		return c.JSON(http.StatusOK, current)
	})

	r.GET("/system/general", func(c echo.Context) error {
		cfg := readGeneralConfig()
		return c.JSON(http.StatusOK, cfg)
	})

	r.POST("/system/general", func(c echo.Context) error {
		current := readGeneralConfig()
		var incoming map[string]interface{}
		if err := c.Bind(&incoming); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON format"})
		}

		if v, ok := incoming["theme"].(string); ok && v != "" {
			current.Theme = v
		}
		if v, ok := incoming["language"].(string); ok && v != "" {
			current.Language = v
		}
		if v, ok := incoming["follow_recreated"].(string); ok && v != "" {
			current.FollowRecreated = v
		}
		if v, ok := incoming["search_intercept"].(bool); ok {
			current.SearchIntercept = v
		}

		if err := saveGeneralConfig(current); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save general settings: " + err.Error()})
		}

		return c.JSON(http.StatusOK, current)
	})

	r.GET("/system/history", func(c echo.Context) error {
		daysStr := c.QueryParam("days")
		days := 7
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}

		var history []map[string]interface{}

		// 1. Fetch high-resolution data from in-memory ring buffer (past 24h)
		since := time.Now().Add(-24 * time.Hour)
		memPoints := stats.Registry.GetSystemSince(since)
		// Reverse chronological order for UI compatibility
		for i := len(memPoints) - 1; i >= 0; i-- {
			p := memPoints[i]
			history = append(history, map[string]interface{}{
				"cpu":       p.CPU,
				"memory":    p.Memory,
				"timestamp": p.Timestamp.UTC().Format(time.RFC3339),
			})
		}

		// 2. Fetch older persisted aggregates from SQLite.
		// If in-memory ring buffer has fresh data or fewer points (e.g. after restart/kill without snapshot),
		// dynamically stitch from SQLite starting right before the oldest in-memory point!
		cutoffTime := time.Now().Add(-24 * time.Hour)
		if len(memPoints) > 0 {
			cutoffTime = memPoints[0].Timestamp
		}
		cutoff := cutoffTime.UTC().Format("2006-01-02 15:04:05")

		rows, err := db.DB.Query(
			"SELECT cpu, memory, timestamp FROM system_stats WHERE timestamp > datetime('now', '-' || ? || ' days') AND timestamp < ? ORDER BY timestamp DESC",
			days, cutoff,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var cpu float64
				var mem int64
				var ts string
				rows.Scan(&cpu, &mem, &ts)
				// Ensure ISO/RFC3339 UTC representation for SQLite timestamps
				isoTS := ts
				if parsed, err := time.Parse("2006-01-02 15:04:05", ts); err == nil {
					isoTS = parsed.UTC().Format(time.RFC3339)
				}
				history = append(history, map[string]interface{}{"cpu": cpu, "memory": mem, "timestamp": isoTS})
			}
		}

		return c.JSON(http.StatusOK, history)
	})

	r.GET("/system/stats", func(c echo.Context) error {
		data, ok := stats.LatestSystemStats()
		if !ok {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Stats not ready"})
		}
		return c.JSON(http.StatusOK, data)
	})

	r.GET("/system/version", func(c echo.Context) error {
		current, latest, updateAvailable, _ := services.CheckVersionStatus(cli.Version)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"current":          current,
			"latest":           latest,
			"update_available": updateAvailable,
		})
	})

	r.GET("/system/db-stats", func(c echo.Context) error {
		var dbSizeBytes int64
		dbPath := db.ServerPath(config.AuthDisabled)
		if fi, err := os.Stat(dbPath); err == nil {
			dbSizeBytes = fi.Size()
		}

		var statsCount, systemStatsCount int64
		_ = db.DB.QueryRow("SELECT COUNT(*) FROM stats").Scan(&statsCount)
		_ = db.DB.QueryRow("SELECT COUNT(*) FROM system_stats").Scan(&systemStatsCount)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"db_size_bytes":      dbSizeBytes,
			"db_size_human":      fmt.Sprintf("%.2f MB", float64(dbSizeBytes)/(1024*1024)),
			"stats_count":        statsCount,
			"system_stats_count": systemStatsCount,
			"db_path":            filepath.Base(dbPath),
		})
	})

	r.POST("/system/vacuum", func(c echo.Context) error {
		_, err := db.DB.Exec("VACUUM;")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		var newSize int64
		dbPath := db.ServerPath(config.AuthDisabled)
		if fi, err := os.Stat(dbPath); err == nil {
			newSize = fi.Size()
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":       true,
			"message":       "Database successfully vacuumed and compacted.",
			"db_size_bytes": newSize,
			"db_size_human": fmt.Sprintf("%.2f MB", float64(newSize)/(1024*1024)),
		})
	})
}

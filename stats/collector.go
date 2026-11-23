package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"docklog/db"
	"docklog/dockerutil"

	"github.com/moby/moby/client"
	"github.com/shirou/gopsutil/v3/disk"
)

var (
	latestSystemStats map[string]interface{}
	sysStatsMu        sync.RWMutex
	cpuSmoothed       float64
	prevStats         = make(map[string]struct {
		TotalUsage  uint64
		SystemUsage uint64
	})
	prevStatsMu sync.Mutex

	// Periodic hourly aggregation buffer for durable history
	hourlySystemAgg = struct {
		sync.Mutex
		count int
		cpuSum float64
		memSum uint64
	}{}

	hourlyContainerAgg = struct {
		sync.Mutex
		data map[string]*containerAggItem
	}{
		data: make(map[string]*containerAggItem),
	}
)

type containerAggItem struct {
	count  int
	cpuSum float64
	memSum uint64
}

func getRetentionDays() int {
	if val := os.Getenv("STATS_RETENTION_DAYS"); val != "" {
		if d, err := strconv.Atoi(val); err == nil && d > 0 {
			return d
		}
	}
	return 7 // Default to 7 days
}

func LatestSystemStats() (map[string]interface{}, bool) {
	sysStatsMu.RLock()
	defer sysStatsMu.RUnlock()
	if latestSystemStats == nil {
		return nil, false
	}
	return latestSystemStats, true
}

func StartCollector(cli *client.Client) {
	// Restore any persisted ring buffer snapshots from previous run
	RestoreSnapshotFromFile()

	go systemStatsBroadcaster()
	collectStats(cli)

	// High-frequency collection into ring buffer every 30s
	statsTicker := time.NewTicker(30 * time.Second)
	go func() {
		for range statsTicker.C {
			collectStats(cli)
		}
	}()

	// Hourly checkpoint: flush aggregated 1h points to SQLite for durable cross-restart trends
	hourlyTicker := time.NewTicker(1 * time.Hour)
	go func() {
		for range hourlyTicker.C {
			flushHourlyAggregates()
		}
	}()

	// Periodic snapshot autosave every 15 minutes (ensures crash recovery even if killed abruptly)
	snapshotTicker := time.NewTicker(15 * time.Minute)
	go func() {
		for range snapshotTicker.C {
			_ = SaveSnapshotToFile()
		}
	}()

	// Daily database cleanup & incremental vacuum to immediately free disk space
	cleanupTicker := time.NewTicker(24 * time.Hour)
	go func() {
		// Run initial cleanup on startup
		runDBCleanupAndVacuum()
		for range cleanupTicker.C {
			runDBCleanupAndVacuum()
		}
	}()
}

func flushHourlyAggregates() {
	// Flush system hourly aggregate
	hourlySystemAgg.Lock()
	if hourlySystemAgg.count > 0 {
		avgCPU := hourlySystemAgg.cpuSum / float64(hourlySystemAgg.count)
		avgMem := hourlySystemAgg.memSum / uint64(hourlySystemAgg.count)
		hourlySystemAgg.count = 0
		hourlySystemAgg.cpuSum = 0
		hourlySystemAgg.memSum = 0
		hourlySystemAgg.Unlock()

		_, _ = db.DB.Exec("INSERT INTO system_stats (cpu, memory) VALUES (?, ?)", avgCPU, avgMem)
	} else {
		hourlySystemAgg.Unlock()
	}

	// Flush containers hourly aggregate
	hourlyContainerAgg.Lock()
	for id, item := range hourlyContainerAgg.data {
		if item.count > 0 {
			avgCPU := item.cpuSum / float64(item.count)
			avgMem := item.memSum / uint64(item.count)
			_, _ = db.DB.Exec("INSERT INTO stats (container_id, cpu, memory) VALUES (?, ?, ?)", id, avgCPU, avgMem)
		}
	}
	hourlyContainerAgg.data = make(map[string]*containerAggItem)
	hourlyContainerAgg.Unlock()
}

func runDBCleanupAndVacuum() {
	days := getRetentionDays()
	retentionClause := fmt.Sprintf("-%d days", days)

	_, _ = db.DB.Exec("DELETE FROM stats WHERE timestamp < datetime('now', ?)", retentionClause)
	_, _ = db.DB.Exec("DELETE FROM system_stats WHERE timestamp < datetime('now', ?)", retentionClause)
	_, _ = db.DB.Exec("DELETE FROM audit_logs WHERE timestamp < datetime('now', '-90 days')")

	// Reclaim free pages back to the operating system
	_, _ = db.DB.Exec("PRAGMA incremental_vacuum(500);")
}

func getSnapshotPath() string {
	if sp := os.Getenv("STATS_SNAPSHOT_PATH"); sp != "" {
		return sp
	}
	// Check if /app/data directory exists (common persistent volume mount)
	if fi, err := os.Stat("/app/data"); err == nil && fi.IsDir() {
		return "/app/data/stats_snapshot.json"
	}
	return "stats_snapshot.json"
}

// SaveSnapshotToFile writes the current in-memory ring buffer state to disk (atomic write)
func SaveSnapshotToFile() error {
	flushHourlyAggregates()

	snap := Registry.ExportSnapshot()
	if len(snap.System) == 0 && len(snap.Containers) == 0 {
		return nil
	}

	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}

	path := getSnapshotPath()
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// RestoreSnapshotFromFile loads previously checkpointed ring buffer points back into memory
func RestoreSnapshotFromFile() {
	path := getSnapshotPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var snap SnapshotData
	if err := json.Unmarshal(data, &snap); err == nil {
		Registry.RestoreSnapshot(snap)
	}
}

func systemStatsBroadcaster() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		updateSystemStatsSnapshot()
		<-ticker.C
	}
}

func updateSystemStatsSnapshot() {
	totalMemory, memUsed := systemMemory()

	cpuVal := 0.0
	if sample, err := sampleCPUPercent(systemSampleInterval); err == nil {
		cpuVal = smoothCPU(sample)
	}

	cores := systemLogicalCPUs()

	diskPercent := 0.0
	if usage, err := disk.Usage("/"); err == nil {
		diskPercent = usage.UsedPercent
	}

	sysStatsMu.Lock()
	latestSystemStats = map[string]interface{}{
		"cpu":          cpuVal,
		"memory":       memUsed,
		"total_memory": totalMemory,
		"cores":        cores,
		"disk_percent": diskPercent,
	}
	sysStatsMu.Unlock()
}

func currentSystemSnapshot() (cpu float64, memory uint64, ok bool) {
	sysStatsMu.RLock()
	snapshot := latestSystemStats
	sysStatsMu.RUnlock()
	if snapshot == nil {
		return 0, 0, false
	}
	cpu, _ = snapshot["cpu"].(float64)
	return cpu, numericFromSnapshot(snapshot["memory"]), true
}

func numericFromSnapshot(value interface{}) uint64 {
	switch v := value.(type) {
	case uint64:
		return v
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	case float64:
		return uint64(v)
	default:
		return 0
	}
}

func collectStats(cli *client.Client) {
	now := time.Now()
	cpuVal, memUsed, ok := currentSystemSnapshot()
	if !ok {
		updateSystemStatsSnapshot()
		cpuVal, memUsed, _ = currentSystemSnapshot()
	}
	if memUsed > 0 || cpuVal > 0 {
		// Push high-resolution 30s data to in-memory ring buffer (0 disk writes)
		Registry.PushSystem(cpuVal, int64(memUsed), now)

		// Accumulate into hourly aggregation buffer
		hourlySystemAgg.Lock()
		hourlySystemAgg.count++
		hourlySystemAgg.cpuSum += cpuVal
		hourlySystemAgg.memSum += memUsed
		hourlySystemAgg.Unlock()
	}

	if cli == nil {
		return
	}

	res, _ := cli.ContainerList(context.Background(), client.ContainerListOptions{})
	containers := dockerutil.ExtractContainers(res)
	for _, ctr := range containers {
		id, _ := ctr["ID"].(string)
		if id == "" {
			id, _ = ctr["Id"].(string)
		}
		state, _ := ctr["State"].(string)
		if state != "running" {
			continue
		}
		stats, err := cli.ContainerStats(context.Background(), id, client.ContainerStatsOptions{Stream: false})
		if err != nil {
			continue
		}
		var s struct {
			CPUStats struct {
				CPUUsage struct {
					TotalUsage uint64 `json:"total_usage"`
				} `json:"cpu_usage"`
				SystemUsage uint64 `json:"system_cpu_usage"`
				OnlineCPUs  uint32 `json:"online_cpus"`
			} `json:"cpu_stats"`
			MemoryStats struct {
				Usage uint64            `json:"usage"`
				Stats map[string]uint64 `json:"stats"`
			} `json:"memory_stats"`
		}
		if err := json.NewDecoder(stats.Body).Decode(&s); err != nil {
			stats.Body.Close()
			continue
		}
		stats.Body.Close()

		cpuPercent := 0.0
		prevStatsMu.Lock()
		prev, ok := prevStats[id]
		if ok {
			cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(prev.TotalUsage)
			systemDelta := float64(s.CPUStats.SystemUsage) - float64(prev.SystemUsage)

			onlineCPUs := float64(s.CPUStats.OnlineCPUs)
			if onlineCPUs == 0 {
				onlineCPUs = float64(runtime.NumCPU())
			}

			if systemDelta > 0 && cpuDelta > 0 {
				cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
			}
		}
		prevStats[id] = struct {
			TotalUsage  uint64
			SystemUsage uint64
		}{
			TotalUsage:  s.CPUStats.CPUUsage.TotalUsage,
			SystemUsage: s.CPUStats.SystemUsage,
		}
		prevStatsMu.Unlock()

		memUsed := s.MemoryStats.Usage - (s.MemoryStats.Stats["cache"])

		// Push to in-memory ring buffer (0 disk writes)
		Registry.PushContainer(id, cpuPercent, int64(memUsed), now)

		// Accumulate into hourly aggregation buffer
		hourlyContainerAgg.Lock()
		aggItem, exists := hourlyContainerAgg.data[id]
		if !exists {
			aggItem = &containerAggItem{}
			hourlyContainerAgg.data[id] = aggItem
		}
		aggItem.count++
		aggItem.cpuSum += cpuPercent
		aggItem.memSum += memUsed
		hourlyContainerAgg.Unlock()
	}
}

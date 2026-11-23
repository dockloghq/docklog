package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"docklog/access"
	"docklog/audit"
	"docklog/config"
	"docklog/containers"
	"docklog/db"
	"docklog/dockerutil"
	"docklog/middleware"
	"docklog/models"
	"docklog/stats"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/moby/moby/client"
)

func (s *Server) registerContainerRoutes(r *echo.Group) {
	cli := s.deps.Docker

	r.GET("/containers", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)
		user := token.Claims.(*models.UserClaims)

		var dbIsAdmin bool
		db.DB.QueryRow("SELECT COALESCE(is_admin, 0) FROM users WHERE id = ?", user.ID).Scan(&dbIsAdmin)

		res, err := cli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
		if err != nil {
			return err
		}

		containerList := dockerutil.ExtractContainers(res)

		var patterns []string
		if !dbIsAdmin {
			patterns = access.GetAuthorizedPatterns(user.ID)
		}
		config.Debugf("User %d (DB Admin: %v) authorized patterns: %v", user.ID, dbIsAdmin, patterns)

		list := make([]models.Container, 0)
		for _, ctr := range containerList {
			name := "unknown"
			names, _ := ctr["Names"].([]interface{})
			if len(names) > 0 {
				name = names[0].(string)[1:]
			}

			image, _ := ctr["Image"].(string)
			state, _ := ctr["State"].(string)

			id, ok := ctr["ID"].(string)
			if !ok {
				id, _ = ctr["Id"].(string)
			}

			if id == "" {
				continue
			}

			if containers.IsExcludedContainer(name, image) {
				continue
			}

			shortID := id
			if len(id) > 12 {
				shortID = id[:12]
			}

			visible := dbIsAdmin
			if !visible {
				for _, p := range patterns {
					if matched, _ := regexp.MatchString(p, name); matched {
						visible = true
						break
					}
				}
			}

			if visible {
				createdVal, _ := ctr["Created"].(float64)
				statusVal, _ := ctr["Status"].(string)

				specs := containers.GetOrInspectSpecs(context.Background(), cli, id)

				var lastCPU float64
				var lastMem int64
				if pt, ok := stats.Registry.GetContainerLast(id); ok {
					lastCPU = pt.CPU
					lastMem = pt.Memory
				} else {
					_ = db.DB.QueryRow("SELECT cpu, memory FROM stats WHERE container_id = ? ORDER BY timestamp DESC LIMIT 1", id).Scan(&lastCPU, &lastMem)
				}

				list = append(list, models.Container{
					ID:           shortID,
					Name:         name,
					Image:        image,
					State:        state,
					Created:      int64(createdVal),
					Status:       statusVal,
					CPULimit:     specs.CPULimit,
					MemLimit:     specs.MemLimit,
					CPU:          lastCPU,
					Memory:       lastMem,
					IsSwarm:      specs.IsSwarm,
					SwarmService: specs.SwarmService,
					Ports:        specs.PortList,
				})
			}
		}
		return c.JSON(http.StatusOK, list)
	})

	r.GET("/containers/:id/inspect", func(c echo.Context) error {
		id := c.Param("id")

		var dbIsAdmin bool
		var userID int
		if userVal := c.Get("user"); userVal != nil {
			if token, ok := userVal.(*jwt.Token); ok && token != nil {
				if userClaims, ok := token.Claims.(*models.UserClaims); ok && userClaims != nil {
					userID = userClaims.ID
					db.DB.QueryRow("SELECT COALESCE(is_admin, 0) FROM users WHERE id = ?", userID).Scan(&dbIsAdmin)
				}
			}
		}
		if config.AuthDisabled {
			dbIsAdmin = true
		}

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}

		if !dbIsAdmin && userID > 0 {
			containerName := strings.TrimPrefix(container.Container.Name, "/")
			patterns := access.GetAuthorizedPatterns(userID)
			authorized := false
			for _, p := range patterns {
				if matched, _ := regexp.MatchString(p, containerName); matched {
					authorized = true
					break
				}
			}
			if !authorized {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
			}
		}

		containerName := strings.TrimPrefix(container.Container.Name, "/")
		containerImage := ""
		if container.Container.Config != nil {
			containerImage = container.Container.Config.Image
		}

		if containers.IsExcludedContainer(containerName, containerImage) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}

		canShell := (dbIsAdmin && middleware.ContainerActionEnvAllowed("shell"))
		if !dbIsAdmin && userID > 0 {
			canShell = middleware.StaffHasShellPermission(userID) && middleware.ContainerActionEnvAllowed("shell")
			var dbCanShell bool
			db.DB.QueryRow("SELECT COALESCE(can_shell, 0) FROM users WHERE id = ?", userID).Scan(&dbCanShell)
			if !dbCanShell {
				canShell = false
			}
		}

		if container.Container.Config != nil {
			if !canShell {
				maskedEnv := make([]string, len(container.Container.Config.Env))
				for i, envStr := range container.Container.Config.Env {
					parts := strings.SplitN(envStr, "=", 2)
					maskedEnv[i] = parts[0] + "=••••••••••••"
				}
				container.Container.Config.Env = maskedEnv
			}
		}

		return c.JSON(http.StatusOK, container)
	})

	r.POST("/containers/:id/action", func(c echo.Context) error {
		id := c.Param("id")
		action := c.FormValue("action")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*models.UserClaims)

		if action != "start" && action != "stop" && action != "restart" && action != "remove" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid action specified."})
		}

		if !middleware.ContainerActionEnvAllowed(action) {
			detail := "This action is disabled on this server."
			actor := s.auditActor(userClaims)
			s.audit(userClaims.ID, actor, action, id, "Forbidden", detail)
			return c.JSON(http.StatusForbidden, map[string]string{"error": detail})
		}

		target, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Target container not found."})
		}
		targetImage := ""
		if target.Container.Config != nil {
			targetImage = target.Container.Config.Image
		}
		if containers.InspectContainerExcluded(target.Container.Name, targetImage) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Target container not found."})
		}
		targetName := strings.TrimPrefix(target.Container.Name, "/")

		if !config.AuthDisabled {
			can, err := middleware.StaffHasContainerActionPermission(action, userClaims.ID)
			if err != nil || !can {
				detail := "This action is not permitted for this account."
				actor := s.auditActor(userClaims)
				s.audit(userClaims.ID, actor, action, targetName, "Forbidden", detail)
				return c.JSON(http.StatusForbidden, map[string]string{"error": detail})
			}
		}

		var dbIsAdmin bool
		db.DB.QueryRow("SELECT COALESCE(is_admin, 0) FROM users WHERE id = ?", userClaims.ID).Scan(&dbIsAdmin)
		if !dbIsAdmin {
			patterns := access.GetAuthorizedPatterns(userClaims.ID)
			authorized := false
			for _, p := range patterns {
				if matched, _ := regexp.MatchString(p, targetName); matched {
					authorized = true
					break
				}
			}

			if !authorized {
				actor := s.auditActor(userClaims)
				s.audit(userClaims.ID, actor, action, targetName, "Forbidden", "Security Restriction: Regex level rights missing.")
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Security Restriction: You are not authorized to interact with this container resource."})
			}
		}

		actor := s.auditActor(userClaims)
		ctx := context.Background()
		timeout := 60
		switch action {
		case "start":
			_, err = cli.ContainerStart(ctx, id, client.ContainerStartOptions{})
		case "stop":
			_, err = cli.ContainerStop(ctx, id, client.ContainerStopOptions{Timeout: &timeout})
		case "restart":
			_, err = cli.ContainerRestart(ctx, id, client.ContainerRestartOptions{Timeout: &timeout})
		case "remove":
			_, err = cli.ContainerRemove(ctx, id, client.ContainerRemoveOptions{Force: true})
		}

		if err != nil {
			s.audit(userClaims.ID, actor, action, targetName, "Error", "System Error: "+err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "System Error: " + err.Error()})
		}

		audit.SuppressContainerEvent(id, action, 8*time.Second)
		s.audit(userClaims.ID, actor, action, targetName, "Success", containerActionDetail(action, actor))
		return c.NoContent(http.StatusOK)
	})

	r.GET("/containers/:id/logs/download", func(c echo.Context) error {
		id := c.Param("id")
		sinceStr := c.QueryParam("since")
		untilStr := c.QueryParam("until")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*models.UserClaims)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}
		downloadImage := ""
		if container.Container.Config != nil {
			downloadImage = container.Container.Config.Image
		}
		if containers.InspectContainerExcluded(container.Container.Name, downloadImage) {
			return c.NoContent(http.StatusNotFound)
		}
		containerName := strings.TrimPrefix(container.Container.Name, "/")

		if !userClaims.IsAdmin {
			patterns := access.GetAuthorizedPatterns(userClaims.ID)
			authorized := false
			for _, p := range patterns {
				if matched, _ := regexp.MatchString(p, containerName); matched {
					authorized = true
					break
				}
			}
			if !authorized {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Permission Denied: Download restricted for this resource."})
			}
		}

		sinceTime, err := parseLogTimeQuery(sinceStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid since timestamp"})
		}
		untilTime, err := parseLogTimeQuery(untilStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid until timestamp"})
		}
		if !sinceTime.IsZero() && !untilTime.IsZero() && sinceTime.After(untilTime) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "since must be before until"})
		}

		mergedLines, err := s.fetchUnifiedContainerLogs(context.Background(), cli, id, container.Container.Config.Tty, sinceStr, untilStr)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch logs: " + err.Error()})
		}

		filtered := filterLogLinesByTime(mergedLines, sinceTime, untilTime)
		
		config.Debugf("[DOWNLOAD] Container %s: since=%s (%v), until=%s (%v)", id, sinceStr, sinceTime, untilStr, untilTime)
		config.Debugf("[DOWNLOAD] Unified logs total lines: %d, matched after filter: %d", len(mergedLines), len(filtered))

		body := strings.Join(filtered, "\n")
		if body != "" {
			body += "\n"
		}

		downloadType := "Full log archive exported"
		if !sinceTime.IsZero() || !untilTime.IsZero() {
			downloadType = "Filtered log archive exported"
		}
		s.audit(userClaims.ID, userClaims.Username, "DOWNLOAD_LOGS", id, "Success", downloadType)

		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+id+"_full.log")
		c.Response().Header().Set(echo.HeaderContentType, "text/plain")
		return c.String(http.StatusOK, body)
	})

	r.GET("/containers/:id/logs", func(c echo.Context) error {
		id := c.Param("id")
		untilStr := c.QueryParam("until")

		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*models.UserClaims)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}
		logsImage := ""
		if container.Container.Config != nil {
			logsImage = container.Container.Config.Image
		}
		if containers.InspectContainerExcluded(container.Container.Name, logsImage) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}

		if !userClaims.IsAdmin {
			containerName := strings.TrimPrefix(container.Container.Name, "/")
			patterns := access.GetAuthorizedPatterns(userClaims.ID)
			authorized := false
			for _, p := range patterns {
				if matched, _ := regexp.MatchString(p, containerName); matched {
					authorized = true
					break
				}
			}
			if !authorized {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
			}
		}

		tailStr := c.QueryParam("tail")
		limit := 100
		if tailStr != "" {
			if tailVal, err := strconv.Atoi(tailStr); err == nil && tailVal > 0 {
				limit = tailVal
			}
		}

		var allLines []string
		if untilStr == "" {
			// No time boundary, fetch directly from Docker with requested limit (fast)
			options := client.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Timestamps: true,
				Follow:     false,
				Tail:       strconv.Itoa(limit),
			}
			out, err := cli.ContainerLogs(context.Background(), id, options)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
			defer out.Close()

			var output bytes.Buffer
			if container.Container.Config.Tty {
				io.Copy(&output, out)
			} else {
				stdcopy.StdCopy(&output, &output, out)
			}
			allLines = strings.Split(output.String(), "\n")
		} else {
			// Time boundary present, use unified logs fetch to merge rotated and active logs (correct)
			var err error
			allLines, err = s.fetchUnifiedContainerLogs(context.Background(), cli, id, container.Container.Config.Tty, "", untilStr)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}

		var logs []string
		if untilStr == "" {
			for _, line := range allLines {
				if line != "" {
					logs = append(logs, line)
				}
			}
			if len(logs) > limit {
				logs = logs[len(logs)-limit:]
			}
		} else {
			untilTime, err := parseLogTimeQuery(untilStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid until timestamp"})
			}

			var filtered []string
			for _, line := range allLines {
				if line == "" {
					continue
				}
				parts := strings.SplitN(line, " ", 2)
				if len(parts) > 0 {
					ts, err := parseLogTimeQuery(parts[0])

					if err == nil {
						if !ts.After(untilTime) {
							filtered = append(filtered, line)
						}
					}
				}
			}

			if len(filtered) > limit {
				logs = filtered[len(filtered)-limit:]
			} else {
				logs = filtered
			}
		}

		config.Debugf("[API] Found %d lines for %s (until: %s)", len(logs), id, untilStr)
		return c.JSON(http.StatusOK, logs)
	})

	r.GET("/containers/:id/logs/count", func(c echo.Context) error {
		id := c.Param("id")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*models.UserClaims)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}
		countImage := ""
		if container.Container.Config != nil {
			countImage = container.Container.Config.Image
		}
		if containers.InspectContainerExcluded(container.Container.Name, countImage) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}

		if !userClaims.IsAdmin {
			containerName := strings.TrimPrefix(container.Container.Name, "/")
			patterns := access.GetAuthorizedPatterns(userClaims.ID)
			authorized := false
			for _, p := range patterns {
				if matched, _ := regexp.MatchString(p, containerName); matched {
					authorized = true
					break
				}
			}
			if !authorized {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
			}
		}

		options := client.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     false,
		}

		out, err := cli.ContainerLogs(context.Background(), id, options)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer out.Close()

		var output bytes.Buffer
		if container.Container.Config.Tty {
			io.Copy(&output, out)
		} else {
			stdcopy.StdCopy(&output, &output, out)
		}

		count := strings.Count(output.String(), "\n")
		return c.JSON(http.StatusOK, map[string]int{"total": count})
	})

	r.GET("/containers/:id/stats", func(c echo.Context) error {
		id := c.Param("id")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*models.UserClaims)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}
		statsImage := ""
		if container.Container.Config != nil {
			statsImage = container.Container.Config.Image
		}
		if containers.InspectContainerExcluded(container.Container.Name, statsImage) {
			return c.NoContent(http.StatusNotFound)
		}
		containerName := strings.TrimPrefix(container.Container.Name, "/")

		if !userClaims.IsAdmin {
			patterns := access.GetAuthorizedPatterns(userClaims.ID)
			authorized := false
			for _, p := range patterns {
				if matched, _ := regexp.MatchString(p, containerName); matched {
					authorized = true
					break
				}
			}
			if !authorized {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized resource access."})
			}
		}

		containerStats, err := cli.ContainerStats(context.Background(), id, client.ContainerStatsOptions{Stream: true})
		if err != nil {
			return err
		}
		defer containerStats.Body.Close()

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c.Response().WriteHeader(http.StatusOK)

		enc := json.NewEncoder(c.Response())
		dec := json.NewDecoder(containerStats.Body)
		for {
			var data interface{}
			if err := dec.Decode(&data); err != nil {
				break
			}
			if err := enc.Encode(data); err != nil {
				break
			}
			c.Response().Flush()
		}
		return nil
	})

	r.GET("/containers/:id/stats-now", func(c echo.Context) error {
		id := c.Param("id")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*models.UserClaims)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}
		nowImage := ""
		if container.Container.Config != nil {
			nowImage = container.Container.Config.Image
		}
		if containers.InspectContainerExcluded(container.Container.Name, nowImage) {
			return c.NoContent(http.StatusNotFound)
		}
		containerName := strings.TrimPrefix(container.Container.Name, "/")

		if !userClaims.IsAdmin {
			patterns := access.GetAuthorizedPatterns(userClaims.ID)
			authorized := false
			for _, p := range patterns {
				if matched, _ := regexp.MatchString(p, containerName); matched {
					authorized = true
					break
				}
			}
			if !authorized {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized resource access."})
			}
		}

		s1, err := cli.ContainerStats(context.Background(), id, client.ContainerStatsOptions{Stream: false})
		if err != nil {
			return err
		}
		var v1 struct {
			CPUStats struct {
				CPUUsage struct {
					TotalUsage uint64 `json:"total_usage"`
				} `json:"cpu_usage"`
				SystemUsage uint64 `json:"system_cpu_usage"`
			} `json:"cpu_stats"`
		}
		json.NewDecoder(s1.Body).Decode(&v1)
		s1.Body.Close()

		time.Sleep(500 * time.Millisecond)

		s2, err := cli.ContainerStats(context.Background(), id, client.ContainerStatsOptions{Stream: false})
		if err != nil {
			return err
		}
		defer s2.Body.Close()

		var v2 struct {
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
		if err := json.NewDecoder(s2.Body).Decode(&v2); err != nil {
			return err
		}

		cpuDelta := float64(v2.CPUStats.CPUUsage.TotalUsage) - float64(v1.CPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(v2.CPUStats.SystemUsage) - float64(v1.CPUStats.SystemUsage)

		onlineCPUs := float64(v2.CPUStats.OnlineCPUs)
		if onlineCPUs == 0 {
			onlineCPUs = float64(runtime.NumCPU())
		}

		cpuPercent := 0.0
		if systemDelta > 0 && cpuDelta > 0 {
			cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
		}
		memUsed := v2.MemoryStats.Usage - (v2.MemoryStats.Stats["cache"])

		return c.JSON(http.StatusOK, map[string]interface{}{
			"cpu":    cpuPercent,
			"memory": memUsed,
		})
	})

	r.GET("/containers/:id/history", func(c echo.Context) error {
		id := c.Param("id")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*models.UserClaims)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}
		historyImage := ""
		if container.Container.Config != nil {
			historyImage = container.Container.Config.Image
		}
		if containers.InspectContainerExcluded(container.Container.Name, historyImage) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}

		longID := container.Container.ID
		if longID == "" {
			longID = id
		}
		shortID := longID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}

		if !userClaims.IsAdmin {
			containerName := strings.TrimPrefix(container.Container.Name, "/")

			patterns := access.GetAuthorizedPatterns(userClaims.ID)
			authorized := false
			for _, p := range patterns {
				if matched, _ := regexp.MatchString(p, containerName); matched {
					authorized = true
					break
				}
			}
			if !authorized {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized resource access."})
			}
		}

		duration := c.QueryParam("duration")
		from := c.QueryParam("from")
		to := c.QueryParam("to")

		var results []map[string]interface{}

		// 1. Try to serve from in-memory ring buffer if duration is within 24h
		if from == "" && to == "" && (duration == "1h" || duration == "24h" || duration == "") {
			window := 24 * time.Hour
			if duration == "1h" {
				window = 1 * time.Hour
			}
			points := stats.Registry.GetContainerSince(longID, time.Now().Add(-window))
			if len(points) == 0 && shortID != longID {
				points = stats.Registry.GetContainerSince(shortID, time.Now().Add(-window))
			}
			if len(points) > 0 {
				for i := len(points) - 1; i >= 0 && len(results) < 200; i-- {
					p := points[i]
					results = append(results, map[string]interface{}{
						"cpu":       p.CPU,
						"memory":    p.Memory,
						"timestamp": p.Timestamp.UTC().Format(time.RFC3339),
					})
				}
				return c.JSON(http.StatusOK, results)
			}
		}

		// 2. Query SQLite for persisted data or custom time ranges
		query := "SELECT cpu, memory, timestamp FROM stats WHERE (container_id = ? OR container_id = ?) "
		var args []interface{}
		args = append(args, longID, shortID)

		if from != "" && to != "" {
			query += "AND timestamp BETWEEN ? AND ? "
			args = append(args, from, to)
		} else if duration == "1h" {
			query += "AND timestamp >= datetime('now', '-1 hour') "
		} else if duration == "24h" {
			query += "AND timestamp >= datetime('now', '-24 hours') "
		}

		query += "ORDER BY timestamp DESC LIMIT 200"

		rows, err := db.DB.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var cpu float64
			var memory int64
			var timestamp string
			rows.Scan(&cpu, &memory, &timestamp)
			isoTS := timestamp
			if parsed, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
				isoTS = parsed.UTC().Format(time.RFC3339)
			}
			results = append(results, map[string]interface{}{
				"cpu":       cpu,
				"memory":    memory,
				"timestamp": isoTS,
			})
		}
		return c.JSON(http.StatusOK, results)
	})

}

func parseLogTimeQuery(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	if ts, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return ts, nil
	}
	if ts, err := time.Parse(time.RFC3339, raw); err == nil {
		return ts, nil
	}
	if unix, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return time.Unix(unix, 0), nil
	}
	return time.Time{}, fmt.Errorf("invalid timestamp")
}

func filterLogLinesByTime(lines []string, since, until time.Time) []string {
	if since.IsZero() && until.IsZero() {
		out := make([]string, 0, len(lines))
		for _, line := range lines {
			if line != "" {
				out = append(out, line)
			}
		}
		return out
	}

	filtered := make([]string, 0, len(lines))
	var lastErr error
	var errCount int
	var beforeCount int
	var afterCount int

	for i, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 0 {
			continue
		}
		ts, err := parseLogTimeQuery(parts[0])
		if err != nil {
			errCount++
			lastErr = err
			if i < 5 {
				config.Debugf("[FILTER] line %d parse fail for %q: %v", i, parts[0], err)
			}
			continue
		}
		if !since.IsZero() && ts.Before(since) {
			beforeCount++
			continue
		}
		if !until.IsZero() && ts.After(until) {
			afterCount++
			continue
		}
		filtered = append(filtered, line)
	}

	config.Debugf("[FILTER] Done. total_lines=%d, errors=%d (last: %v), skipped_before=%d, skipped_after=%d, matched=%d", 
		len(lines), errCount, lastErr, beforeCount, afterCount, len(filtered))
	return filtered
}

func (s *Server) fetchUnifiedContainerLogs(ctx context.Context, cli *client.Client, id string, isTty bool, sinceStr, untilStr string) ([]string, error) {
	// Call A: fetch rotated logs (Tail: "")
	optionsA := client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     false,
		Tail:       "",
		Since:      sinceStr,
		Until:      untilStr,
	}
	outA, err := cli.ContainerLogs(ctx, id, optionsA)
	if err != nil {
		return nil, err
	}
	defer outA.Close()

	var bufA bytes.Buffer
	if isTty {
		io.Copy(&bufA, outA)
	} else {
		stdcopy.StdCopy(&bufA, &bufA, outA)
	}
	linesA := strings.Split(bufA.String(), "\n")

	// Call B: fetch active log (Tail: "4000")
	optionsB := client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     false,
		Tail:       "4000",
		Since:      sinceStr,
		Until:      untilStr,
	}
	outB, err := cli.ContainerLogs(ctx, id, optionsB)
	if err != nil {
		return nil, err
	}
	defer outB.Close()

	var bufB bytes.Buffer
	if isTty {
		io.Copy(&bufB, outB)
	} else {
		stdcopy.StdCopy(&bufB, &bufB, outB)
	}
	linesB := strings.Split(bufB.String(), "\n")

	// Merge and deduplicate
	uniqueLines := make(map[string]bool)
	type LogLine struct {
		Timestamp time.Time
		Content   string
	}
	var allParsed []LogLine

	addLines := func(lines []string) {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if uniqueLines[line] {
				continue
			}
			uniqueLines[line] = true

			// Parse timestamp
			parts := strings.SplitN(line, " ", 2)
			if len(parts) > 0 {
				t, err := parseLogTimeQuery(parts[0])
				if err == nil {
					allParsed = append(allParsed, LogLine{Timestamp: t, Content: line})
				} else {
					allParsed = append(allParsed, LogLine{Timestamp: time.Time{}, Content: line})
				}
			}
		}
	}

	addLines(linesA)
	addLines(linesB)

	// Sort chronologically
	sort.Slice(allParsed, func(i, j int) bool {
		return allParsed[i].Timestamp.Before(allParsed[j].Timestamp)
	})

	// Convert back to string slice
	merged := make([]string, len(allParsed))
	for i, l := range allParsed {
		merged[i] = l.Content
	}

	return merged, nil
}


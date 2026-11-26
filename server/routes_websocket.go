package server

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"docklog/access"
	"docklog/config"
	"docklog/containers"
	"docklog/db"
	"docklog/dockerutil"
	"docklog/middleware"
	"docklog/stats"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/moby/moby/client"
)

func (s *Server) registerWebSocketRoutes() {
	cli := s.deps.Docker

	s.echo.GET("/ws/system-stats", func(c echo.Context) error {
		if !config.AuthDisabled {
			_, err := middleware.AuthenticateWS(c)
			if err != nil {
				return middleware.WSAuthError(c, err)
			}
		}

		ws, err := middleware.UpgradeAuthenticatedWS(c)
		if err != nil {
			return err
		}
		defer ws.Close()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				data, ok := stats.LatestSystemStats()
				if ok {
					if err := ws.WriteJSON(data); err != nil {
						return nil
					}
				}
			case <-c.Request().Context().Done():
				return nil
			}
		}
	})

	s.echo.GET("/ws/events", func(c echo.Context) error {
		userClaims, err := middleware.AuthenticateWS(c)
		if err != nil {
			return middleware.WSAuthError(c, err)
		}

		ws, err := middleware.UpgradeAuthenticatedWS(c)
		if err != nil {
			return err
		}
		defer ws.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		result := cli.Events(ctx, client.EventsListOptions{})

		for {
			select {
			case msg := <-result.Messages:
				if !userClaims.IsAdmin {
					containerName := msg.Actor.Attributes["name"]
					if containerName == "" {
						continue
					}
					patterns := access.GetAuthorizedPatterns(userClaims.ID)
					authorized := false
					for _, p := range patterns {
						if matched, _ := regexp.MatchString(p, containerName); matched {
							authorized = true
							break
						}
					}
					if !authorized {
						continue
					}
				}

				if err := ws.WriteJSON(msg); err != nil {
					return nil
				}
			case <-result.Err:
				return nil
			case <-c.Request().Context().Done():
				return nil
			}
		}
	})

	s.echo.GET("/ws/logs/:id", func(c echo.Context) error {
		id := c.Param("id")

		userClaims, err := middleware.AuthenticateWS(c)
		if err != nil {
			return middleware.WSAuthError(c, err)
		}

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}
		wsLogsImage := ""
		if container.Container.Config != nil {
			wsLogsImage = container.Container.Config.Image
		}
		if containers.InspectContainerExcluded(container.Container.Name, wsLogsImage) {
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
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied: You do not have permission to view logs for this resource."})
			}
		}

		ws, err := middleware.UpgradeAuthenticatedWS(c)
		if err != nil {
			return err
		}
		defer ws.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tailLimit := "100"
		if customTail := c.QueryParam("tail"); customTail != "" {
			if num, err := strconv.Atoi(customTail); err == nil && num > 0 && num <= 2000 {
				tailLimit = customTail
			}
		} else {
			cfg := readLogcConfig()
			if cfg.DefaultTail != "" {
				tailLimit = cfg.DefaultTail
			}
		}

		out, err := cli.ContainerLogs(ctx, id, client.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Tail:       tailLimit,
			Timestamps: true,
		})
		if err != nil {
			return err
		}
		defer out.Close()

		isTty := container.Container.Config != nil && container.Container.Config.Tty
		_ = dockerutil.StreamLogs(out, isTty, func(payload []byte) error {
			return ws.WriteMessage(websocket.TextMessage, payload)
		})
		return nil
	})

	s.echo.GET("/ws/containers/:id/stats", func(c echo.Context) error {
		id := c.Param("id")

		userClaims, err := middleware.AuthenticateWS(c)
		if err != nil {
			return middleware.WSAuthError(c, err)
		}

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
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
			}
		}

		longID := container.Container.ID
		if longID == "" {
			longID = id
		}
		shortID := longID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}

		ws, err := middleware.UpgradeAuthenticatedWS(c)
		if err != nil {
			return err
		}
		defer ws.Close()

		hasCheckedDirectStats := false
		sendStats := func() error {
			var cpu float64
			var mem int64
			found := false

			if pt, ok := stats.Registry.GetContainerLast(longID); ok && pt.Memory > 0 {
				cpu = pt.CPU
				mem = pt.Memory
				found = true
			} else if pt, ok := stats.Registry.GetContainerLast(shortID); ok && pt.Memory > 0 {
				cpu = pt.CPU
				mem = pt.Memory
				found = true
			} else if pt, ok := stats.Registry.GetContainerLast(id); ok && pt.Memory > 0 {
				cpu = pt.CPU
				mem = pt.Memory
				found = true
			}

			if !found {
				_ = db.DB.QueryRow("SELECT cpu, memory FROM stats WHERE (container_id = ? OR container_id = ? OR container_id = ?) ORDER BY timestamp DESC LIMIT 1", longID, shortID, id).Scan(&cpu, &mem)
			}

			// If memory is still 0 (e.g. newly started container or ringbuffer empty), fetch once directly from Docker stats
			if mem == 0 && !hasCheckedDirectStats {
				hasCheckedDirectStats = true
				if sRaw, err := cli.ContainerStats(context.Background(), longID, client.ContainerStatsOptions{Stream: false}); err == nil {
					var v2 struct {
						MemoryStats struct {
							Usage uint64            `json:"usage"`
							Stats map[string]uint64 `json:"stats"`
						} `json:"memory_stats"`
					}
					if err := json.NewDecoder(sRaw.Body).Decode(&v2); err == nil {
						cache := v2.MemoryStats.Stats["cache"]
						if v2.MemoryStats.Usage > cache {
							mem = int64(v2.MemoryStats.Usage - cache)
						} else {
							mem = int64(v2.MemoryStats.Usage)
						}
					}
					sRaw.Body.Close()
				}
			}

			return ws.WriteJSON(map[string]interface{}{
				"cpu":    cpu,
				"memory": mem,
			})
		}

		// Send initial point immediately
		if err := sendStats(); err != nil {
			return nil
		}

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := sendStats(); err != nil {
					return nil
				}
			case <-c.Request().Context().Done():
				return nil
			}
		}
	})

	s.echo.GET("/ws/shell/:id", func(c echo.Context) error {
		id := c.Param("id")

		userClaims, err := middleware.AuthenticateWS(c)
		if err != nil {
			return middleware.WSAuthError(c, err)
		}

		if !config.AllowShell {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Shell access is disabled on this server."})
		}

		var canShell bool
		err = db.DB.QueryRow("SELECT can_shell FROM users WHERE id = ? AND is_active = 1", userClaims.ID).Scan(&canShell)
		if err != nil || !canShell {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Shell access is not permitted for this account."})
		}

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}
		shellImage := ""
		if container.Container.Config != nil {
			shellImage = container.Container.Config.Image
		}
		if containers.InspectContainerExcluded(container.Container.Name, shellImage) {
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
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied: You do not have permission to view this resource."})
			}
		}

		shellCmd := c.QueryParam("shell")
		if shellCmd == "" {
			shellCmd = "/bin/sh"
		}
		allowedShells := map[string]bool{"/bin/sh": true, "/bin/bash": true, "/bin/ash": true}
		if !allowedShells[shellCmd] {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid shell"})
		}

		ws, err := middleware.UpgradeAuthenticatedWS(c)
		if err != nil {
			return err
		}
		defer ws.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		execConfig := client.ExecCreateOptions{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			TTY:          true,
			Cmd:          []string{shellCmd},
		}

		execResult, err := cli.ExecCreate(ctx, id, execConfig)
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("\r\n[DockLog] Failed to create terminal session: "+err.Error()+"\r\n"))
			return nil
		}

		attachResult, err := cli.ExecAttach(ctx, execResult.ID, client.ExecAttachOptions{
			TTY: true,
		})
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("\r\n[DockLog] Failed to attach to terminal session: "+err.Error()+"\r\n"))
			return nil
		}
		defer attachResult.Close()

		errChan := make(chan error, 2)
		go func() {
			for {
				_, msg, err := ws.ReadMessage()
				if err != nil {
					errChan <- err
					return
				}

				_, err = attachResult.Conn.Write(msg)
				if err != nil {
					errChan <- err
					return
				}
			}
		}()

		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := attachResult.Reader.Read(buf)
				if n > 0 {
					err = ws.WriteMessage(websocket.TextMessage, buf[:n])
					if err != nil {
						errChan <- err
						return
					}
				}
				if err != nil {
					errChan <- err
					return
				}
			}
		}()

		<-errChan
		return nil
	})
}

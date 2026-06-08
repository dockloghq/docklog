package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"docklog/db"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/moby/moby/client"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/crypto/bcrypt"
)

var (
	SECRET_KEY = []byte("secret-key-change-this")
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	AuthDisabled bool
	CanStart     bool
	CanStop      bool
	CanRestart   bool
	CanDelete    bool
	AllowShell   bool
)

type Container struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Image    string  `json:"image"`
	State    string  `json:"state"`
	Created  int64   `json:"created"`
	Status   string  `json:"status"`
	CPULimit float64 `json:"cpu_limit"`
	MemLimit int64   `json:"mem_limit"`
	CPU      float64 `json:"cpu"`
	Memory   int64   `json:"memory"`
}

type UserClaims struct {
	ID                 int    `json:"id"`
	Username           string `json:"username"`
	IsAdmin            bool   `json:"is_admin"`
	IsRestrictedAccess bool   `json:"is_restricted_access"`
	CanStart           bool   `json:"can_start"`
	CanStop            bool   `json:"can_stop"`
	CanRestart         bool   `json:"can_restart"`
	CanDelete          bool   `json:"can_delete"`
	CanShell           bool   `json:"can_shell"`
	AllowedContainers  string `json:"allowed_containers"`
	IsActive           bool   `json:"is_active"`
	PasswordVersion    int    `json:"password_version"`
	jwt.RegisteredClaims
}

type User struct {
	ID                 int    `json:"id"`
	Username           string `json:"username"`
	IsAdmin            bool   `json:"is_admin"`
	PasswordChanged    bool   `json:"password_changed"`
	CanStart           bool   `json:"can_start"`
	CanStop            bool   `json:"can_stop"`
	CanRestart         bool   `json:"can_restart"`
	CanDelete          bool   `json:"can_delete"`
	CanShell           bool   `json:"can_shell"`
	IsRestrictedAccess bool   `json:"is_restricted_access"`
	AllowedContainers  string `json:"allowed_containers"`
	IsActive           bool   `json:"is_active"`
}

func logAudit(userID int, username, action, resource, status, message string) {
	_, err := db.DB.Exec(
		"INSERT INTO audit_logs (user_id, username, action, resource, status, message) VALUES (?, ?, ?, ?, ?, ?)",
		userID, username, action, resource, status, message,
	)
	if err != nil {
		log.Printf("Failed to write audit log: %v", err)
	}
}

func getAuthorizedPatterns(userID int) []string {
	var isRestricted bool
	var pattern string
	err := db.DB.QueryRow("SELECT is_restricted_access, allowed_containers FROM users WHERE id = ?", userID).Scan(&isRestricted, &pattern)
	if err != nil {
		return []string{"^$"}
	}

	if !isRestricted {
		return []string{".*"}
	}

	if pattern == "" {
		return []string{""}
	}

	rawPatterns := strings.Split(pattern, ",")
	var finalPatterns []string
	for _, p := range rawPatterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if strings.HasPrefix(p, "^") || strings.HasSuffix(p, "$") {
			finalPatterns = appendValidatedPattern(finalPatterns, p)
			continue
		}

		regP := strings.ReplaceAll(p, "*", ".*")
		regP = strings.ReplaceAll(regP, "..*", ".*")

		if !strings.ContainsAny(regP, "()[]{}|") {
			if !strings.HasPrefix(regP, "^") {
				regP = "^" + regP
			}
			if !strings.HasSuffix(regP, "$") {
				regP = regP + "$"
			}
		}
		finalPatterns = appendValidatedPattern(finalPatterns, regP)
	}
	return finalPatterns
}

func appendValidatedPattern(patterns []string, regP string) []string {
	if len(regP) > maxContainerPatternLen {
		log.Printf("Skipping container pattern: exceeds %d characters", maxContainerPatternLen)
		return patterns
	}
	if _, err := regexp.Compile(regP); err != nil {
		log.Printf("Skipping invalid container pattern: %v", err)
		return patterns
	}
	return append(patterns, regP)
}

func main() {
	if maybeRunCLI() {
		return
	}

	AuthDisabled = os.Getenv("DISABLE_AUTH") == "true" || os.Getenv("NO_AUTH") == "true"
	initSecretKey()
	initClientAccess()
	initWSUpgrader()

	getEnvBool := func(key string, defaultVal bool) bool {
		val := os.Getenv(key)
		if val == "" {
			return defaultVal
		}
		return val == "true"
	}

	CanStart = getEnvBool("ALLOW_START", false)
	CanStop = getEnvBool("ALLOW_STOP", false)
	CanRestart = getEnvBool("ALLOW_RESTART", false)
	CanDelete = getEnvBool("ALLOW_DELETE", false)
	AllowShell = getEnvBool("ALLOW_SHELL", false) || getEnvBool("ALLOW_BASH", false)

	// DB Init
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		if AuthDisabled {
			dbPath = ":memory:"
		} else {
			dbPath = "docklog.db"
		}
	}
	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// Seed Admin
	seedAdmin()

	e := echo.New()
	if TrustProxy {
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
	}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(securityHeadersMiddleware())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			return corsOriginAllowed(origin), nil
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderAuthorization,
			echo.HeaderContentType,
			headerDockLogClient,
		},
	}))
	e.Use(clientAccessMiddleware())

	// Docker Client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Start Background Collector
	startStatsCollector(cli)

	// Auth Endpoints
	e.POST("/api/token", func(c echo.Context) error {
		ip := c.RealIP()
		if loginRateLimit.isLimited(ip, 10, 15*time.Minute) {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "Too many login attempts. Try again later."})
		}

		username := strings.TrimSpace(c.FormValue("username"))
		password := c.FormValue("password")

		var id int
		var hashedPassword string
		var isAdmin, passwordChanged, canStart, canStop, canRestart, canDelete, canShell, isRestricted, isActive bool
		var allowedContainers string
		var passwordVersion int
		err := db.DB.QueryRow("SELECT id, password, is_admin, password_changed, can_start, can_stop, can_restart, can_delete, can_shell, is_restricted_access, allowed_containers, is_active, COALESCE(password_version, 1) FROM users WHERE username = ?", username).Scan(
			&id, &hashedPassword, &isAdmin, &passwordChanged, &canStart, &canStop, &canRestart, &canDelete, &canShell, &isRestricted, &allowedContainers, &isActive, &passwordVersion,
		)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) != nil {
			loginRateLimit.recordFailure(ip)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		loginRateLimit.clear(ip)

		if !isActive {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Account deactivated. Please contact administrator."})
		}

		claims := &UserClaims{
			ID:                 id,
			Username:           username,
			IsAdmin:            isAdmin,
			IsRestrictedAccess: isRestricted,
			CanStart:           canStart,
			CanStop:            canStop,
			CanRestart:         canRestart,
			CanDelete:          canDelete,
			CanShell:           canShell,
			AllowedContainers:  allowedContainers,
			IsActive:           isActive,
			PasswordVersion:    passwordVersion,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		t, err := token.SignedString(SECRET_KEY)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"access_token":     t,
			"is_admin":         isAdmin,
			"password_changed": passwordChanged,
		})
	})

	// Public configuration route
	e.GET("/api/config", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"auth_disabled": AuthDisabled,
			"allow_start":   CanStart,
			"allow_stop":    CanStop,
			"allow_restart": CanRestart,
			"allow_delete":  CanDelete,
			"allow_shell":   AllowShell,
			"client_access": clientAccessConfig(),
		})
	})

	// Restricted Group
	r := e.Group("/api")
	if AuthDisabled {
		r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Populate context claims when auth is disabled
				claims := &UserClaims{
					ID:                 1,
					Username:           "admin",
					IsAdmin:            true,
					IsRestrictedAccess: false,
					CanStart:           true,
					CanStop:            true,
					CanRestart:         true,
					CanDelete:          true,
					IsActive:           true,
				}
				token := &jwt.Token{
					Claims: claims,
					Valid:  true,
				}
				c.Set("user", token)
				return next(c)
			}
		})
	} else {
		r.Use(echojwt.WithConfig(echojwt.Config{
			NewClaimsFunc: func(c echo.Context) jwt.Claims {
				return new(UserClaims)
			},
			SigningKey: SECRET_KEY,
		}))
	}

	// Password change enforcement & session validation middleware
	r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if AuthDisabled {
				return next(c)
			}
			token := c.Get("user").(*jwt.Token)
			claims := token.Claims.(*UserClaims)

			var changed bool
			if err := refreshClaimsFromDB(claims); err != nil {
				switch err.Error() {
				case "account deactivated":
					return c.JSON(http.StatusForbidden, map[string]string{
						"error": "Account deactivated. Please contact administrator.",
						"code":  "ACCOUNT_DEACTIVATED",
					})
				case "session invalidated":
					return c.JSON(http.StatusUnauthorized, map[string]string{
						"error": "Session invalidated. Password was changed. Please re-login.",
						"code":  "SESSION_INVALIDATED",
					})
				default:
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User verification failed"})
				}
			}

			err := db.DB.QueryRow("SELECT password_changed FROM users WHERE id = ?", claims.ID).Scan(&changed)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User verification failed"})
			}

			// Allow profile and password-change endpoints to proceed after active-state validation.
			if c.Path() == "/api/user/change-password" || c.Path() == "/api/user/me" {
				return next(c)
			}

			if !changed {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Password change required", "code": "FORCE_PASSWORD_CHANGE"})
			}

			return next(c)
		}
	})

	r.GET("/containers", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)
		user := token.Claims.(*UserClaims)

		// Always check DB for current admin status (not stale JWT)
		var dbIsAdmin bool
		db.DB.QueryRow("SELECT COALESCE(is_admin, 0) FROM users WHERE id = ?", user.ID).Scan(&dbIsAdmin)

		res, err := cli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
		if err != nil {
			return err
		}

		containers := extractContainers(res)

		var patterns []string
		if !dbIsAdmin {
			patterns = getAuthorizedPatterns(user.ID)
		}
		log.Printf("User %d (DB Admin: %v) authorized patterns: %v", user.ID, dbIsAdmin, patterns)

		var list []Container
		for _, ctr := range containers {
			name := "unknown"
			names, _ := ctr["Names"].([]interface{})
			if len(names) > 0 {
				name = names[0].(string)[1:]
			}

			image, _ := ctr["Image"].(string)
			state, _ := ctr["State"].(string)

			// Handle both "Id" and "ID"
			id, ok := ctr["ID"].(string)
			if !ok {
				id, _ = ctr["Id"].(string)
			}

			if id == "" {
				continue // Skip invalid containers
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

				// Fetch detailed limits
				inspect, _ := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
				cpuLimit := 0.0
				memLimit := int64(0)
				if inspect.Container.HostConfig != nil {
					if inspect.Container.HostConfig.NanoCPUs > 0 {
						cpuLimit = float64(inspect.Container.HostConfig.NanoCPUs) / 1e9
					}
					memLimit = inspect.Container.HostConfig.Memory
				}

				// Fetch latest stats snapshot
				var lastCPU float64
				var lastMem int64
				db.DB.QueryRow("SELECT cpu, memory FROM stats WHERE container_id = ? ORDER BY timestamp DESC LIMIT 1", id).Scan(&lastCPU, &lastMem)

				list = append(list, Container{
					ID:       shortID,
					Name:     name,
					Image:    image,
					State:    state,
					Created:  int64(createdVal),
					Status:   statusVal,
					CPULimit: cpuLimit,
					MemLimit: memLimit,
					CPU:      lastCPU,
					Memory:   lastMem,
				})
			}
		}
		return c.JSON(http.StatusOK, list)
	})

	r.GET("/containers/:id/inspect", func(c echo.Context) error {
		id := c.Param("id")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*UserClaims)

		// Always check DB for current admin status (not stale JWT)
		var dbIsAdmin bool
		db.DB.QueryRow("SELECT COALESCE(is_admin, 0) FROM users WHERE id = ?", userClaims.ID).Scan(&dbIsAdmin)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}

		if !dbIsAdmin {
			containerName := strings.TrimPrefix(container.Container.Name, "/")
			patterns := getAuthorizedPatterns(userClaims.ID)
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

		// Sanitize environment variables to prevent leaking secrets
		if container.Container.Config != nil && len(container.Container.Config.Env) > 0 {
			sanitized := make([]string, len(container.Container.Config.Env))
			for i, item := range container.Container.Config.Env {
				parts := strings.SplitN(item, "=", 2)
				if len(parts) == 2 {
					k := strings.ToLower(parts[0])
					if strings.Contains(k, "pass") ||
						strings.Contains(k, "secret") ||
						strings.Contains(k, "key") ||
						strings.Contains(k, "token") ||
						strings.Contains(k, "auth") ||
						strings.Contains(k, "pwd") ||
						strings.Contains(k, "db_") {
						sanitized[i] = parts[0] + "=••••••••••••"
					} else {
						sanitized[i] = item
					}
				} else {
					sanitized[i] = item
				}
			}
			container.Container.Config.Env = sanitized
		}

		return c.JSON(http.StatusOK, container)
	})

	r.POST("/containers/:id/action", func(c echo.Context) error {
		id := c.Param("id")
		action := c.FormValue("action")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*UserClaims)

		if action != "start" && action != "stop" && action != "restart" && action != "remove" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid action specified."})
		}

		if !containerActionEnvAllowed(action) {
			detail := "This action is disabled on this server."
			logAudit(userClaims.ID, userClaims.Username, action, id, "Forbidden", detail)
			return c.JSON(http.StatusForbidden, map[string]string{"error": detail})
		}

		if !AuthDisabled {
			can, err := staffHasContainerActionPermission(action, userClaims.ID)
			if err != nil || !can {
				detail := "This action is not permitted for this account."
				logAudit(userClaims.ID, userClaims.Username, action, id, "Forbidden", detail)
				return c.JSON(http.StatusForbidden, map[string]string{"error": detail})
			}
		}

		// 2. Check Resource-Specific Regex Access (from DB, not JWT)
		var dbIsAdmin bool
		db.DB.QueryRow("SELECT COALESCE(is_admin, 0) FROM users WHERE id = ?", userClaims.ID).Scan(&dbIsAdmin)
		if !dbIsAdmin {
			container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Target container not found."})
			}
			containerName := strings.TrimPrefix(container.Container.Name, "/")

			patterns := getAuthorizedPatterns(userClaims.ID)
			authorized := false
			for _, p := range patterns {
				if matched, _ := regexp.MatchString(p, containerName); matched {
					authorized = true
					break
				}
			}

			if !authorized {
				logAudit(userClaims.ID, userClaims.Username, action, containerName, "Forbidden", "Security Restriction: Regex level rights missing.")
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Security Restriction: You are not authorized to interact with this container resource."})
			}
		}

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
			logAudit(userClaims.ID, userClaims.Username, action, id, "Error", "System Error: "+err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "System Error: " + err.Error()})
		}

		logAudit(userClaims.ID, userClaims.Username, action, id, "Success", "Action executed successfully.")
		return c.NoContent(http.StatusOK)
	})

	r.GET("/containers/:id/logs/download", func(c echo.Context) error {
		id := c.Param("id")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*UserClaims)

		if !userClaims.IsAdmin {
			container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
			if err != nil {
				return c.NoContent(http.StatusNotFound)
			}
			containerName := strings.TrimPrefix(container.Container.Name, "/")

			patterns := getAuthorizedPatterns(userClaims.ID)
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

		options := client.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Timestamps: true,
			Follow:     false,
		}

		out, err := cli.ContainerLogs(context.Background(), id, options)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch logs: " + err.Error()})
		}
		defer out.Close()

		logAudit(userClaims.ID, userClaims.Username, "DOWNLOAD_LOGS", id, "Success", "Full log archive exported")

		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+id+"_full.log")
		c.Response().Header().Set(echo.HeaderContentType, "text/plain")
		c.Response().WriteHeader(http.StatusOK)

		_, err = io.Copy(c.Response().Writer, out)
		return err
	})

	r.GET("/containers/:id/logs", func(c echo.Context) error {
		id := c.Param("id")
		untilStr := c.QueryParam("until")

		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*UserClaims)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}

		if !userClaims.IsAdmin {
			containerName := strings.TrimPrefix(container.Container.Name, "/")
			patterns := getAuthorizedPatterns(userClaims.ID)
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
			Timestamps: true,
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

		allLines := strings.Split(output.String(), "\n")
		var logs []string

		if untilStr == "" {
			// Initial load - just get last 100
			for _, line := range allLines {
				if line != "" {
					logs = append(logs, line)
				}
			}
			if len(logs) > 100 {
				logs = logs[len(logs)-100:]
			}
		} else {
			// Historical fetch - get 100 lines before 'until'
			var untilTime time.Time
			// Try parsing as RFC3339Nano (Docker's default)
			untilTime, err = time.Parse(time.RFC3339Nano, untilStr)
			if err != nil {
				// Fallback to RFC3339
				untilTime, err = time.Parse(time.RFC3339, untilStr)
				if err != nil {
					// Fallback to Unix
					if unix, err := strconv.ParseInt(untilStr, 10, 64); err == nil {
						untilTime = time.Unix(unix, 0)
					}
				}
			}

			var filtered []string
			for _, line := range allLines {
				if line == "" {
					continue
				}
				// Extract timestamp from line
				parts := strings.SplitN(line, " ", 2)
				if len(parts) > 0 {
					ts, err := time.Parse(time.RFC3339Nano, parts[0])
					if err != nil {
						ts, err = time.Parse(time.RFC3339, parts[0])
					}

					if err == nil {
						// Be inclusive (!After) to ensure no logs are missed;
						// the frontend will handle deduplication of the boundary log.
						if !ts.After(untilTime) {
							filtered = append(filtered, line)
						}
					}
				}
			}

			if len(filtered) > 100 {
				logs = filtered[len(filtered)-100:]
			} else {
				logs = filtered
			}
		}

		log.Printf("[API] Found %d lines for %s (until: %s)", len(logs), id, untilStr)
		return c.JSON(http.StatusOK, logs)
	})

	r.GET("/containers/:id/logs/count", func(c echo.Context) error {
		id := c.Param("id")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*UserClaims)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}

		if !userClaims.IsAdmin {
			containerName := strings.TrimPrefix(container.Container.Name, "/")
			patterns := getAuthorizedPatterns(userClaims.ID)
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
		userClaims := token.Claims.(*UserClaims)

		if !userClaims.IsAdmin {
			container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
			if err != nil {
				return c.NoContent(http.StatusNotFound)
			}
			containerName := strings.TrimPrefix(container.Container.Name, "/")

			patterns := getAuthorizedPatterns(userClaims.ID)
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

		stats, err := cli.ContainerStats(context.Background(), id, client.ContainerStatsOptions{Stream: true})
		if err != nil {
			return err
		}
		defer stats.Body.Close()

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c.Response().WriteHeader(http.StatusOK)

		enc := json.NewEncoder(c.Response())
		dec := json.NewDecoder(stats.Body)
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
		userClaims := token.Claims.(*UserClaims)

		if !userClaims.IsAdmin {
			container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
			if err != nil {
				return c.NoContent(http.StatusNotFound)
			}
			containerName := strings.TrimPrefix(container.Container.Name, "/")

			patterns := getAuthorizedPatterns(userClaims.ID)
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

		// To get accurate CPU %, we need two samples.
		// We'll take a quick 200ms sample to stay responsive.
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
		userClaims := token.Claims.(*UserClaims)

		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Container not found"})
		}

		longID := container.Container.ID
		if longID == "" {
			longID = id
		}

		if !userClaims.IsAdmin {
			containerName := strings.TrimPrefix(container.Container.Name, "/")

			patterns := getAuthorizedPatterns(userClaims.ID)
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

		query := "SELECT cpu, memory, timestamp FROM stats WHERE container_id = ? "
		var args []interface{}
		args = append(args, longID)

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

		var results []map[string]interface{}
		for rows.Next() {
			var cpu float64
			var memory int64
			var timestamp string
			rows.Scan(&cpu, &memory, &timestamp)
			results = append(results, map[string]interface{}{
				"cpu":       cpu,
				"memory":    memory,
				"timestamp": timestamp,
			})
		}
		return c.JSON(http.StatusOK, results)
	})

	r.GET("/system/history", func(c echo.Context) error {
		daysStr := c.QueryParam("days")
		days := 30
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}

		rows, err := db.DB.Query("SELECT cpu, memory, timestamp FROM system_stats WHERE timestamp > datetime('now', '-' || ? || ' days') ORDER BY timestamp DESC", days)
		if err != nil {
			return err
		}
		var history []map[string]interface{}
		for rows.Next() {
			var cpu float64
			var mem int64
			var ts string
			rows.Scan(&cpu, &mem, &ts)
			history = append(history, map[string]interface{}{"cpu": cpu, "memory": mem, "timestamp": ts})
		}
		return c.JSON(http.StatusOK, history)
	})

	r.GET("/system/stats", func(c echo.Context) error {
		sysStatsMu.RLock()
		defer sysStatsMu.RUnlock()
		if latestSystemStats == nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Stats not ready"})
		}
		return c.JSON(http.StatusOK, latestSystemStats)
	})

	r.POST("/user/change-password", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)
		user := token.Claims.(*UserClaims)
		newPassword := c.FormValue("password")

		if len(newPassword) < minPasswordLength {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Password must be at least %d characters", minPasswordLength)})
		}

		var hashedPassword string
		var passwordChanged bool
		err := db.DB.QueryRow("SELECT password, password_changed FROM users WHERE id = ?", user.ID).Scan(&hashedPassword, &passwordChanged)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
		}

		if passwordChanged {
			currentPassword := c.FormValue("current_password")
			if currentPassword == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Current password is required"})
			}
			if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(currentPassword)) != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Current password is incorrect"})
			}
		}

		h, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		_, err = db.DB.Exec("UPDATE users SET password = ?, password_changed = 1, password_version = COALESCE(password_version, 1) + 1 WHERE id = ?", string(h), user.ID)
		if err != nil {
			log.Printf("Error updating password for user %d: %v", user.ID, err)
			return err
		}
		log.Printf("Password successfully updated for user %d", user.ID)
		return c.NoContent(http.StatusOK)
	})

	r.GET("/user/me", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)
		claims := token.Claims.(*UserClaims)
		var dbUser User
		err := db.DB.QueryRow("SELECT id, username, is_admin, password_changed, can_start, can_stop, can_restart, can_delete, can_shell, is_restricted_access, allowed_containers, is_active FROM users WHERE id = ?", claims.ID).
			Scan(&dbUser.ID, &dbUser.Username, &dbUser.IsAdmin, &dbUser.PasswordChanged, &dbUser.CanStart, &dbUser.CanStop, &dbUser.CanRestart, &dbUser.CanDelete, &dbUser.CanShell, &dbUser.IsRestrictedAccess, &dbUser.AllowedContainers, &dbUser.IsActive)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
		}
		if !dbUser.IsActive {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Account deactivated",
				"code":  "ACCOUNT_DEACTIVATED",
			})
		}
		return c.JSON(http.StatusOK, dbUser)
	})

	// Admin Only Routes
	admin := r.Group("/admin")
	admin.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Get("user").(*jwt.Token)
			user := token.Claims.(*UserClaims)
			var isAdmin bool
			err := db.DB.QueryRow("SELECT is_admin FROM users WHERE id = ? AND is_active = 1", user.ID).Scan(&isAdmin)
			if err != nil || !isAdmin {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
			}
			user.IsAdmin = isAdmin
			return next(c)
		}
	})

	admin.GET("/users", func(c echo.Context) error {
		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page < 1 {
			page = 1
		}
		limit := 10
		offset := (page - 1) * limit

		var total int
		db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)

		rows, err := db.DB.Query("SELECT id, username, is_admin, can_start, can_stop, can_restart, can_delete, can_shell, is_restricted_access, allowed_containers, is_active FROM users LIMIT ? OFFSET ?", limit, offset)
		if err != nil {
			log.Printf("Failed to query users: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch users: " + err.Error()})
		}
		defer rows.Close()
		users := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id int
			var username, allowedContainers string
			var isAdmin, canStart, canStop, canRestart, canDelete, canShell, isRestricted, isActive bool
			if err := rows.Scan(&id, &username, &isAdmin, &canStart, &canStop, &canRestart, &canDelete, &canShell, &isRestricted, &allowedContainers, &isActive); err != nil {
				log.Printf("Failed to scan user row: %v", err)
				continue
			}
			users = append(users, map[string]interface{}{
				"id":                   id,
				"username":             username,
				"is_admin":             isAdmin,
				"can_start":            canStart,
				"can_stop":             canStop,
				"can_restart":          canRestart,
				"can_delete":           canDelete,
				"can_shell":            canShell,
				"is_restricted_access": isRestricted,
				"allowed_containers":   allowedContainers,
				"is_active":            isActive,
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"users": users,
			"total": total,
			"page":  page,
			"pages": (total + limit - 1) / limit,
		})
	})

	admin.PUT("/users/:id/active", func(c echo.Context) error {
		id := c.Param("id")
		isActive := c.FormValue("is_active") == "true"
		_, err := db.DB.Exec("UPDATE users SET is_active = ? WHERE id = ? AND is_admin = 0", isActive, id)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	admin.POST("/users", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		if !isPasswordStrongEnough(password) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Password must be at least %d characters", minPasswordLength)})
		}
		canStart, canStop, canRestart, canDelete, canShell := clampStaffActionPermissions(
			c.FormValue("can_start") == "true",
			c.FormValue("can_stop") == "true",
			c.FormValue("can_restart") == "true",
			c.FormValue("can_delete") == "true",
			c.FormValue("can_shell") == "true",
		)
		isRestricted := c.FormValue("is_restricted_access") == "true"
		allowedContainers := c.FormValue("allowed_containers")
		if allowedContainers == "" {
			allowedContainers = ".*"
		}

		h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password is too long or invalid. Please use a shorter password."})
		}
		_, err = db.DB.Exec("INSERT INTO users (username, password, is_admin, can_start, can_stop, can_restart, can_delete, can_shell, is_restricted_access, allowed_containers, password_changed, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			username, string(h), false, canStart, canStop, canRestart, canDelete, canShell, isRestricted, allowedContainers, true, true)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "User already exists"})
		}
		return c.NoContent(http.StatusCreated)
	})

	admin.PUT("/users/:id/permissions", func(c echo.Context) error {
		id := c.Param("id")
		canStart, canStop, canRestart, canDelete, canShell := clampStaffActionPermissions(
			c.FormValue("can_start") == "true",
			c.FormValue("can_stop") == "true",
			c.FormValue("can_restart") == "true",
			c.FormValue("can_delete") == "true",
			c.FormValue("can_shell") == "true",
		)
		isRestricted := c.FormValue("is_restricted_access") == "true"
		allowedContainers := c.FormValue("allowed_containers")

		_, err := db.DB.Exec("UPDATE users SET can_start = ?, can_stop = ?, can_restart = ?, can_delete = ?, can_shell = ?, is_restricted_access = ?, allowed_containers = ? WHERE id = ?", canStart, canStop, canRestart, canDelete, canShell, isRestricted, allowedContainers, id)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	admin.PUT("/users/:id/password", func(c echo.Context) error {
		id := c.Param("id")
		newPassword := c.FormValue("password")
		if !isPasswordStrongEnough(newPassword) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Password must be at least %d characters", minPasswordLength)})
		}

		h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password is too long or invalid. Please use a shorter password."})
		}
		_, err = db.DB.Exec("UPDATE users SET password = ?, password_changed = 1, password_version = COALESCE(password_version, 1) + 1 WHERE id = ?", string(h), id)
		if err != nil {
			return err
		}

		token := c.Get("user").(*jwt.Token)
		claims := token.Claims.(*UserClaims)
		logAudit(claims.ID, claims.Username, "RESET_PASSWORD", "User:"+id, "Success", "Administrator reset user password")

		return c.NoContent(http.StatusOK)
	})
	admin.DELETE("/users/:id", func(c echo.Context) error {
		id := c.Param("id")
		// Prevent self-deletion if needed, but here we just delete
		_, err := db.DB.Exec("DELETE FROM users WHERE id = ? AND is_admin = 0", id)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	admin.GET("/audit", func(c echo.Context) error {
		from := c.QueryParam("from")
		to := c.QueryParam("to")

		query := "SELECT id, user_id, username, action, resource, status, message, timestamp FROM audit_logs"
		args := []interface{}{}

		if from != "" && to != "" {
			query += " WHERE timestamp BETWEEN ? AND ?"
			args = append(args, from, to)
		}

		query += " ORDER BY timestamp DESC LIMIT 1000"
		rows, err := db.DB.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch audit logs: " + err.Error()})
		}
		defer rows.Close()
		logs := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, userID int
			var username, action, resource, status, message, timestamp string
			rows.Scan(&id, &userID, &username, &action, &resource, &status, &message, &timestamp)
			logs = append(logs, map[string]interface{}{
				"id":        id,
				"user_id":   userID,
				"username":  username,
				"action":    action,
				"resource":  resource,
				"status":    status,
				"message":   message,
				"timestamp": timestamp,
			})
		}
		return c.JSON(http.StatusOK, logs)
	})

	e.GET("/ws/system-stats", func(c echo.Context) error {
		if !AuthDisabled {
			_, err := authenticateWS(c)
			if err != nil {
				return wsAuthError(c, err)
			}
		}

		ws, err := upgradeAuthenticatedWS(c)
		if err != nil {
			return err
		}
		defer ws.Close()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sysStatsMu.RLock()
				data := latestSystemStats
				sysStatsMu.RUnlock()

				if data != nil {
					if err := ws.WriteJSON(data); err != nil {
						return nil
					}
				}
			case <-c.Request().Context().Done():
				return nil
			}
		}
	})

	e.GET("/ws/events", func(c echo.Context) error {
		userClaims, err := authenticateWS(c)
		if err != nil {
			return wsAuthError(c, err)
		}

		ws, err := upgradeAuthenticatedWS(c)
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
				// Filter container events or pass all events based on permissions
				if !userClaims.IsAdmin {
					containerName := msg.Actor.Attributes["name"]
					if containerName == "" {
						continue // If name is empty, skip for restricted users
					}
					patterns := getAuthorizedPatterns(userClaims.ID)
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

	e.GET("/ws/logs/:id", func(c echo.Context) error {
		id := c.Param("id")

		userClaims, err := authenticateWS(c)
		if err != nil {
			return wsAuthError(c, err)
		}

		// Regex Access Check
		if !userClaims.IsAdmin {
			container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
			if err != nil {
				return c.NoContent(http.StatusNotFound)
			}
			containerName := strings.TrimPrefix(container.Container.Name, "/")

			patterns := getAuthorizedPatterns(userClaims.ID)
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

		ws, err := upgradeAuthenticatedWS(c)
		if err != nil {
			return err
		}
		defer ws.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		out, err := cli.ContainerLogs(ctx, id, client.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Tail:       "100",
			Timestamps: true,
		})
		if err != nil {
			return err
		}
		defer out.Close()

		header := make([]byte, 8)
		for {
			_, err = io.ReadFull(out, header)
			if err != nil {
				break
			}

			size := uint32(header[4])<<24 | uint32(header[5])<<16 | uint32(header[6])<<8 | uint32(header[7])
			payload := make([]byte, size)
			_, err = io.ReadFull(out, payload)
			if err != nil {
				break
			}

			if err := ws.WriteMessage(websocket.TextMessage, payload); err != nil {
				break
			}
		}
		return nil
	})

	e.GET("/ws/shell/:id", func(c echo.Context) error {
		id := c.Param("id")

		userClaims, err := authenticateWS(c)
		if err != nil {
			return wsAuthError(c, err)
		}

		if !AllowShell {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Shell access is disabled on this server."})
		}

		var canShell bool
		err = db.DB.QueryRow("SELECT can_shell FROM users WHERE id = ? AND is_active = 1", userClaims.ID).Scan(&canShell)
		if err != nil || !canShell {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Shell access is not permitted for this account."})
		}

		// Verify container exists and get its name
		container, err := cli.ContainerInspect(context.Background(), id, client.ContainerInspectOptions{})
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}
		containerName := strings.TrimPrefix(container.Container.Name, "/")

		// Regex Access Check
		if !userClaims.IsAdmin {
			patterns := getAuthorizedPatterns(userClaims.ID)
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

		ws, err := upgradeAuthenticatedWS(c)
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

		// Attach exec
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

	// Serve Frontend
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "frontend/dist",
		Browse: false,
		HTML5:  true,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/api") || strings.HasPrefix(c.Path(), "/ws")
		},
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	log.Printf("DockLog (Go Edition) starting on %s\n", port)
	e.Logger.Fatal(e.Start(port))
}

func extractContainers(res interface{}) []map[string]interface{} {
	b, _ := json.Marshal(res)
	var m interface{}
	json.Unmarshal(b, &m)

	if list, ok := m.([]interface{}); ok {
		var ret []map[string]interface{}
		for _, item := range list {
			if mm, ok := item.(map[string]interface{}); ok {
				ret = append(ret, mm)
			}
		}
		return ret
	}
	if mm, ok := m.(map[string]interface{}); ok {
		for _, val := range mm {
			if list, ok := val.([]interface{}); ok {
				var ret []map[string]interface{}
				for _, item := range list {
					if mmm, ok := item.(map[string]interface{}); ok {
						ret = append(ret, mmm)
					}
				}
				return ret
			}
		}
	}
	return nil
}

var (
	latestSystemStats map[string]interface{}
	sysStatsMu        sync.RWMutex
)

func systemStatsBroadcaster() {
	for {
		v, _ := mem.VirtualMemory()
		cp, _ := cpu.Percent(500*time.Millisecond, false)
		cpuVal := 0.0
		if len(cp) > 0 {
			cpuVal = cp[0]
		}

		cores, err := cpu.Counts(true)
		if err != nil || cores == 0 {
			cores = runtime.NumCPU()
		}

		sysStatsMu.Lock()
		latestSystemStats = map[string]interface{}{
			"cpu":          cpuVal,
			"memory":       v.Used,
			"total_memory": v.Total,
			"cores":        cores,
		}
		sysStatsMu.Unlock()

		time.Sleep(500 * time.Millisecond) // Total cycle ~1s (500ms sample + 500ms sleep)
	}
}

func startStatsCollector(cli *client.Client) {
	go systemStatsBroadcaster()
	// Initial collection
	collectStats(cli)

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			collectStats(cli)
			// Cleanup old stats (30 days)
			db.DB.Exec("DELETE FROM stats WHERE timestamp < datetime('now', '-30 days')")
			db.DB.Exec("DELETE FROM system_stats WHERE timestamp < datetime('now', '-30 days')")
		}
	}()
}

var (
	prevStats = make(map[string]struct {
		TotalUsage  uint64
		SystemUsage uint64
	})
	prevStatsMu sync.Mutex
)

func collectStats(cli *client.Client) {
	// System Stats
	v, _ := mem.VirtualMemory()
	cp, _ := cpu.Percent(time.Second, false)
	if len(cp) > 0 {
		db.DB.Exec("INSERT INTO system_stats (cpu, memory) VALUES (?, ?)", cp[0], v.Used)
	}

	// Container Stats Snapshot
	res, _ := cli.ContainerList(context.Background(), client.ContainerListOptions{})
	containers := extractContainers(res)
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
		db.DB.Exec("INSERT INTO stats (container_id, cpu, memory) VALUES (?, ?, ?)", id, cpuPercent, memUsed)
	}
}

func seedAdmin() {
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if count == 0 {
		const plain = "admin123"
		log.Println("Default admin account created (username: admin, password: admin123). Change the password on first login.")

		h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash default admin password: %v", err)
		}
		_, err = db.DB.Exec(
			`INSERT INTO users (username, password, is_admin, can_start, can_stop, can_restart, can_delete, can_shell, password_changed)
			 VALUES (?, ?, 1, 1, 1, 1, 1, 1, 0)`,
			"admin", string(h),
		)
		if err != nil {
			log.Fatalf("Failed to create default admin: %v", err)
		}
	}

	_, err := db.DB.Exec(
		`UPDATE users SET can_start = 1, can_stop = 1, can_restart = 1, can_delete = 1, can_shell = 1
		 WHERE is_admin = 1 AND (can_start = 0 OR can_stop = 0 OR can_restart = 0 OR can_delete = 0 OR can_shell = 0)`,
	)
	if err != nil {
		log.Printf("Warning: failed to sync admin action permissions: %v", err)
	}
}

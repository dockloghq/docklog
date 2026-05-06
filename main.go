package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"docklog/db"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/moby/moby/client"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

var (
	SECRET_KEY = []byte("secret-key-change-this")
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

type Container struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Image     string `json:"image"`
	State     string `json:"state"`
	Type      string `json:"type"` // "docker" or "pod"
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
	AllowedContainers  string `json:"allowed_containers"`
	IsActive           bool   `json:"is_active"`
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
		return []string{".*"}
	}

	if !isRestricted {
		return []string{".*"}
	}

	if pattern == "" {
		return []string{""}
	}

	// Support multiple patterns separated by comma and anchor them for exact match
	rawPatterns := strings.Split(pattern, ",")
	var anchoredPatterns []string
	for _, p := range rawPatterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// If it's exactly .*, just pass it
		if p == ".*" {
			anchoredPatterns = append(anchoredPatterns, p)
			continue
		}

		// Convert glob * to regex .* and ensure it's not doubled
		regP := strings.ReplaceAll(p, "*", ".*")
		regP = strings.ReplaceAll(regP, "..*", ".*")

		// Anchor the pattern to ensure exact matches for simple strings
		// but allow flexible matching for patterns with wildcards.
		if !strings.HasPrefix(regP, "^") {
			regP = "^" + regP
		}
		if !strings.HasSuffix(regP, "$") {
			regP = regP + "$"
		}
		anchoredPatterns = append(anchoredPatterns, regP)
	}
	return anchoredPatterns
}

func main() {
	// DB Init
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "docklog.db"
	}
	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to init DB: %v", err)
	}

	// Seed Admin
	seedAdmin()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Docker Client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Start Background Collector
	startStatsCollector(cli)

	// K8s Client
	var k8sClient *kubernetes.Clientset
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Printf("Failed to load kubeconfig from %s: %v. Attempting in-cluster config...", kubeconfig, err)
		config, err = clientcmd.BuildConfigFromFlags("", "") // Try in-cluster
	}

	if err == nil {
		k8sClient, err = kubernetes.NewForConfig(config)
		if err != nil {
			log.Printf("Failed to create K8s client: %v", err)
		} else {
			log.Println("Kubernetes client initialized successfully")
		}
	} else {
		log.Printf("Kubernetes support disabled: %v", err)
	}

	// Auth Endpoints
	e.POST("/api/token", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		var id int
		var hashedPassword string
		var isAdmin, passwordChanged, canStart, canStop, canRestart, canDelete, isRestricted, isActive bool
		var allowedContainers string
		err := db.DB.QueryRow("SELECT id, password, is_admin, password_changed, can_start, can_stop, can_restart, can_delete, is_restricted_access, allowed_containers, is_active FROM users WHERE username = ?", username).Scan(
			&id, &hashedPassword, &isAdmin, &passwordChanged, &canStart, &canStop, &canRestart, &canDelete, &isRestricted, &allowedContainers, &isActive,
		)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

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
			AllowedContainers:  allowedContainers,
			IsActive:           isActive,
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

	// Restricted Group
	r := e.Group("/api")
	r.Use(echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(UserClaims)
		},
		SigningKey: SECRET_KEY,
	}))

	// Password change enforcement middleware
	r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Get("user").(*jwt.Token)
			claims := token.Claims.(*UserClaims)

			var changed, active bool
			err := db.DB.QueryRow("SELECT password_changed, is_active FROM users WHERE id = ?", claims.ID).Scan(&changed, &active)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User verification failed"})
			}

			if !active {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Account deactivated. Please contact administrator.",
					"code":  "ACCOUNT_DEACTIVATED",
				})
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
				list = append(list, Container{
					ID:    shortID,
					Name:  name,
					Image: image,
					State: state,
					Type:  "docker",
				})
			}
		}
		return c.JSON(http.StatusOK, list)
	})

	r.GET("/pods", func(c echo.Context) error {
		if k8sClient == nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Kubernetes support not enabled"})
		}
		token := c.Get("user").(*jwt.Token)
		user := token.Claims.(*UserClaims)

		// Always check DB for current admin status
		var dbIsAdmin bool
		db.DB.QueryRow("SELECT COALESCE(is_admin, 0) FROM users WHERE id = ?", user.ID).Scan(&dbIsAdmin)

		pods, err := k8sClient.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list pods: " + err.Error()})
		}

		var patterns []string
		if !dbIsAdmin {
			patterns = getAuthorizedPatterns(user.ID)
		}

		var list []Container
		for _, pod := range pods.Items {
			visible := dbIsAdmin
			if !visible {
				for _, p := range patterns {
					if matched, _ := regexp.MatchString(p, pod.Name); matched {
						visible = true
						break
					}
				}
			}

			if visible {
				state := string(pod.Status.Phase)
				image := ""
				if len(pod.Spec.Containers) > 0 {
					image = pod.Spec.Containers[0].Image
				}
				list = append(list, Container{
					ID:        pod.Namespace + "/" + pod.Name,
					Name:      pod.Name,
					Namespace: pod.Namespace,
					Image:     image,
					State:     strings.ToLower(state),
					Type:      "pod",
				})
			}
		}
		return c.JSON(http.StatusOK, list)
	})

	r.POST("/containers/:id/action", func(c echo.Context) error {
		id := c.Param("id")
		action := c.FormValue("action")
		token := c.Get("user").(*jwt.Token)
		userClaims := token.Claims.(*UserClaims)

		// 1. Check Action-Specific Global Permission
		var can bool
		var err error
		switch action {
		case "start":
			err = db.DB.QueryRow("SELECT (is_admin OR can_start) FROM users WHERE id = ?", userClaims.ID).Scan(&can)
		case "stop":
			err = db.DB.QueryRow("SELECT (is_admin OR can_stop) FROM users WHERE id = ?", userClaims.ID).Scan(&can)
		case "restart":
			err = db.DB.QueryRow("SELECT (is_admin OR can_restart) FROM users WHERE id = ?", userClaims.ID).Scan(&can)
		case "remove":
			err = db.DB.QueryRow("SELECT (is_admin OR can_delete) FROM users WHERE id = ?", userClaims.ID).Scan(&can)
		default:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid action specified."})
		}

		if err != nil || !can {
			logAudit(userClaims.ID, userClaims.Username, action, id, "Forbidden", "Permission Denied: Action level rights missing.")
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Permission Denied: You do not have rights to perform this action."})
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
		timeout := 10
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

		stats, err := cli.ContainerStats(context.Background(), id, client.ContainerStatsOptions{Stream: false})
		if err != nil {
			return err
		}
		defer stats.Body.Close()
		var data interface{}
		if err := json.NewDecoder(stats.Body).Decode(&data); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, data)
	})

	r.GET("/containers/:id/history", func(c echo.Context) error {
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

		duration := c.QueryParam("duration")
		from := c.QueryParam("from")
		to := c.QueryParam("to")

		query := "SELECT cpu, memory, timestamp FROM stats WHERE container_id = ? "
		var args []interface{}
		args = append(args, id)

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
		v, _ := mem.VirtualMemory()
		cp, _ := cpu.Percent(time.Second, false)
		cpuVal := 0.0
		if len(cp) > 0 {
			cpuVal = cp[0]
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"cpu":          cpuVal,
			"memory":       v.Used,
			"total_memory": v.Total,
		})
	})

	r.POST("/user/change-password", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)
		user := token.Claims.(*UserClaims)
		newPassword := c.FormValue("password")

		if len(newPassword) < 6 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password too short"})
		}

		h, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		_, err := db.DB.Exec("UPDATE users SET password = ?, password_changed = 1 WHERE id = ?", string(h), user.ID)
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
		err := db.DB.QueryRow("SELECT id, username, is_admin, password_changed, can_start, can_stop, can_restart, can_delete, is_restricted_access, allowed_containers, is_active FROM users WHERE id = ?", claims.ID).
			Scan(&dbUser.ID, &dbUser.Username, &dbUser.IsAdmin, &dbUser.PasswordChanged, &dbUser.CanStart, &dbUser.CanStop, &dbUser.CanRestart, &dbUser.CanDelete, &dbUser.IsRestrictedAccess, &dbUser.AllowedContainers, &dbUser.IsActive)
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
			if !user.IsAdmin {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
			}
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

		rows, err := db.DB.Query("SELECT id, username, is_admin, can_start, can_stop, can_restart, can_delete, is_restricted_access, allowed_containers, is_active FROM users LIMIT ? OFFSET ?", limit, offset)
		if err != nil {
			log.Printf("Failed to query users: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch users: " + err.Error()})
		}
		defer rows.Close()
		users := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id int
			var username, allowedContainers string
			var isAdmin, canStart, canStop, canRestart, canDelete, isRestricted, isActive bool
			if err := rows.Scan(&id, &username, &isAdmin, &canStart, &canStop, &canRestart, &canDelete, &isRestricted, &allowedContainers, &isActive); err != nil {
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
		canStart := c.FormValue("can_start") == "true"
		canStop := c.FormValue("can_stop") == "true"
		canRestart := c.FormValue("can_restart") == "true"
		canDelete := c.FormValue("can_delete") == "true"
		isRestricted := c.FormValue("is_restricted_access") == "true"
		allowedContainers := c.FormValue("allowed_containers")
		if allowedContainers == "" {
			allowedContainers = ".*"
		}

		h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password is too long or invalid. Please use a shorter password."})
		}
		_, err = db.DB.Exec("INSERT INTO users (username, password, is_admin, can_start, can_stop, can_restart, can_delete, is_restricted_access, allowed_containers, password_changed, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			username, string(h), false, canStart, canStop, canRestart, canDelete, isRestricted, allowedContainers, true, true)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "User already exists"})
		}
		return c.NoContent(http.StatusCreated)
	})

	admin.PUT("/users/:id/permissions", func(c echo.Context) error {
		id := c.Param("id")
		canStart := c.FormValue("can_start") == "true"
		canStop := c.FormValue("can_stop") == "true"
		canRestart := c.FormValue("can_restart") == "true"
		canDelete := c.FormValue("can_delete") == "true"
		isRestricted := c.FormValue("is_restricted_access") == "true"
		allowedContainers := c.FormValue("allowed_containers")

		_, err := db.DB.Exec("UPDATE users SET can_start = ?, can_stop = ?, can_restart = ?, can_delete = ?, is_restricted_access = ?, allowed_containers = ? WHERE id = ?", canStart, canStop, canRestart, canDelete, isRestricted, allowedContainers, id)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	admin.PUT("/users/:id/password", func(c echo.Context) error {
		id := c.Param("id")
		newPassword := c.FormValue("password")
		if newPassword == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password cannot be empty"})
		}

		h, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password is too long or invalid. Please use a shorter password."})
		}
		_, err = db.DB.Exec("UPDATE users SET password = ?, password_changed = 0 WHERE id = ?", string(h), id)
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
		rows, err := db.DB.Query("SELECT id, user_id, username, action, resource, status, message, timestamp FROM audit_logs ORDER BY timestamp DESC LIMIT 500")
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

	e.GET("/ws/logs/:id", func(c echo.Context) error {
		id := c.Param("id")
		tokenStr := c.QueryParam("token")

		if tokenStr == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication token required"})
		}

		token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			return SECRET_KEY, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}

		userClaims := token.Claims.(*UserClaims)

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

		isPod := c.QueryParam("type") == "pod"

		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if isPod {
			if k8sClient == nil {
				return ws.WriteMessage(websocket.TextMessage, []byte("Kubernetes support not enabled"))
			}

			// Find pod by UID or Name
			// K8s logs need pod name and namespace. We expect 'id' to be namespace/name
			parts := strings.Split(id, "/")
			namespace := "default"
			podName := id
			if len(parts) == 2 {
				namespace = parts[0]
				podName = parts[1]
			}

			req := k8sClient.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
				Follow:    true,
				TailLines: func(i int64) *int64 { return &i }(200),
			})
			stream, err := req.Stream(ctx)
			if err != nil {
				return ws.WriteMessage(websocket.TextMessage, []byte("Failed to stream logs: "+err.Error()))
			}
			defer stream.Close()

			buf := make([]byte, 1024)
			for {
				n, err := stream.Read(buf)
				if n > 0 {
					if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
						break
					}
				}
				if err != nil {
					break
				}
			}
			return nil
		}

		out, err := cli.ContainerLogs(ctx, id, client.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Tail:       "200",
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

	// Serve Frontend
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "frontend/dist",
		Browse: false,
		HTML5:  true,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/api") || strings.HasPrefix(c.Path(), "/ws")
		},
	}))

	log.Println("DockLog (Go Edition) starting on :8000")
	e.Logger.Fatal(e.Start(":8000"))
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

func startStatsCollector(cli *client.Client) {
	// Initial collection
	collectStats(cli)

	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			collectStats(cli)
			// Cleanup old stats (30 days)
			db.DB.Exec("DELETE FROM stats WHERE timestamp < datetime('now', '-30 days')")
			db.DB.Exec("DELETE FROM system_stats WHERE timestamp < datetime('now', '-30 days')")
		}
	}()
}

func collectStats(cli *client.Client) {
	// System Stats
	v, _ := mem.VirtualMemory()
	cp, _ := cpu.Percent(time.Second, false)
	if len(cp) > 0 {
		db.DB.Exec("INSERT INTO system_stats (cpu, memory) VALUES (?, ?)", cp[0], v.Used)
	}
}

func seedAdmin() {
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if count == 0 {
		h, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash default admin password: %v", err)
		}
		db.DB.Exec("INSERT INTO users (username, password, is_admin, password_changed) VALUES (?, ?, ?, ?)", "admin", string(h), true, false)
		log.Println("Admin user created: admin / admin123")
	}
}

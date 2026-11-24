package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"docklog/access"
	"docklog/config"
	"docklog/db"
	"docklog/dockerutil"
	"docklog/middleware"
	"docklog/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
)

func (s *Server) registerSwarmRoutes(r *echo.Group) {
	cli := s.deps.Docker
	if cli == nil {
		swarmUnavailable := func(c echo.Context) error {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Docker client is not available"})
		}
		r.GET("/swarm/overview", swarmUnavailable)
		r.GET("/swarm/services", swarmUnavailable)
		r.GET("/swarm/services/:id", swarmUnavailable)
		r.POST("/swarm/services/:id/scale", swarmUnavailable)
		r.DELETE("/swarm/services/:id", swarmUnavailable)
		r.GET("/swarm/nodes", swarmUnavailable)
		r.POST("/swarm/nodes/:id/availability", swarmUnavailable)
		r.GET("/swarm/tasks", swarmUnavailable)
		return
	}

	r.GET("/swarm/overview", s.handleSwarmOverview)
	r.GET("/swarm/services", s.handleSwarmListServices)
	r.GET("/swarm/services/:id", s.handleSwarmGetService)
	r.GET("/swarm/services/:id/logs", s.handleSwarmServiceLogsHistory)
	r.GET("/swarm/services/:id/logs/download", s.handleSwarmServiceLogsDownload)
	r.GET("/swarm/services/:id/logs/count", s.handleSwarmServiceLogsCount)
	r.POST("/swarm/services/:id/scale", s.handleSwarmScaleService)
	r.DELETE("/swarm/services/:id", s.handleSwarmDeleteService)
	r.GET("/swarm/nodes", s.handleSwarmListNodes)
	r.POST("/swarm/nodes/:id/availability", s.handleSwarmUpdateNodeAvailability)
	r.GET("/swarm/tasks", s.handleSwarmListTasks)
}

func (s *Server) checkSwarmActive(ctx context.Context) (bool, string, bool, string, error) {
	cli := s.deps.Docker
	if cli == nil {
		return false, "inactive", false, "", fmt.Errorf("docker not initialized")
	}
	info, err := cli.Info(ctx, client.InfoOptions{})
	if err != nil {
		return false, "inactive", false, "", err
	}
	active := info.Info.Swarm.LocalNodeState == swarm.LocalNodeStateActive
	return active, string(info.Info.Swarm.LocalNodeState), info.Info.Swarm.ControlAvailable, info.Info.Swarm.NodeID, nil
}

func (s *Server) handleSwarmOverview(c echo.Context) error {
	ctx := context.Background()
	active, nodeState, isManager, nodeID, err := s.checkSwarmActive(ctx)
	if err != nil {
		return c.JSON(http.StatusOK, models.SwarmOverview{
			Active:         false,
			LocalNodeState: "error",
		})
	}
	overview := models.SwarmOverview{
		Active:           active,
		LocalNodeState:   nodeState,
		ControlAvailable: isManager,
		NodeID:           nodeID,
	}

	if !active || !isManager {
		return c.JSON(http.StatusOK, overview)
	}

	cli := s.deps.Docker
	nodesRes, err := cli.NodeList(ctx, client.NodeListOptions{})
	if err == nil {
		overview.NodesCount = len(nodesRes.Items)
		for _, n := range nodesRes.Items {
			if n.Spec.Role == swarm.NodeRoleManager {
				overview.ManagersCount++
			} else {
				overview.WorkersCount++
			}
		}
	}

	servicesRes, err := cli.ServiceList(ctx, client.ServiceListOptions{Status: true})
	if err == nil {
		overview.ServicesCount = len(servicesRes.Items)
	}

	tasksRes, err := cli.TaskList(ctx, client.TaskListOptions{})
	if err == nil {
		overview.TasksCount = len(tasksRes.Items)
		for _, t := range tasksRes.Items {
			if t.Status.State == swarm.TaskStateRunning {
				overview.RunningTasksCount++
			}
		}
	}

	return c.JSON(http.StatusOK, overview)
}

func (s *Server) handleSwarmListServices(c echo.Context) error {
	ctx := context.Background()
	active, _, isManager, _, _ := s.checkSwarmActive(ctx)
	if !active || !isManager {
		return c.JSON(http.StatusOK, []models.SwarmService{})
	}

	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	isAdmin := s.k8sUserIsAdmin(user.ID)

	cli := s.deps.Docker

	res, err := cli.ServiceList(ctx, client.ServiceListOptions{Status: true})
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	}

	var patterns []string
	if !isAdmin {
		patterns = access.GetAuthorizedPatterns(user.ID)
	}

	list := make([]models.SwarmService, 0, len(res.Items))
	for _, svc := range res.Items {
		name := svc.Spec.Name
		if !isAdmin && !access.MatchesAnyPattern(patterns, name) {
			continue
		}

		image := ""
		if svc.Spec.TaskTemplate.ContainerSpec != nil {
			image = svc.Spec.TaskTemplate.ContainerSpec.Image
		}

		mode := "replicated"
		var replicas uint64
		if svc.Spec.Mode.Global != nil {
			mode = "global"
		} else if svc.Spec.Mode.Replicated != nil && svc.Spec.Mode.Replicated.Replicas != nil {
			replicas = *svc.Spec.Mode.Replicated.Replicas
		}

		var running uint64
		if svc.ServiceStatus != nil {
			running = svc.ServiceStatus.RunningTasks
		}

		ports := make([]string, 0)
		if svc.Endpoint.Ports != nil {
			for _, p := range svc.Endpoint.Ports {
				ports = append(ports, fmt.Sprintf("%d:%d/%s", p.PublishedPort, p.TargetPort, p.Protocol))
			}
		}

		list = append(list, models.SwarmService{
			ID:        svc.ID,
			Name:      name,
			Image:     image,
			Mode:      mode,
			Replicas:  replicas,
			Running:   running,
			Ports:     ports,
			CreatedAt: svc.CreatedAt.Unix(),
			UpdatedAt: svc.UpdatedAt.Unix(),
			Labels:    svc.Spec.Labels,
		})
	}

	return c.JSON(http.StatusOK, list)
}

func (s *Server) handleSwarmGetService(c echo.Context) error {
	ctx := context.Background()
	active, _, _, _, _ := s.checkSwarmActive(ctx)
	if !active {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Swarm is not active"})
	}

	id := c.Param("id")
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	isAdmin := s.k8sUserIsAdmin(user.ID)

	cli := s.deps.Docker

	res, err := cli.ServiceInspect(ctx, id, client.ServiceInspectOptions{InsertDefaults: true})
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Service not found"})
	}

	if !isAdmin {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.MatchesAnyPattern(patterns, res.Service.Spec.Name) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
		}
	}

	var dbIsAdmin, dbCanShell bool
	_ = db.DB.QueryRow("SELECT COALESCE(is_admin, 0), COALESCE(can_shell, 0) FROM users WHERE id = ?", user.ID).Scan(&dbIsAdmin, &dbCanShell)
	canShell := (dbIsAdmin || middleware.StaffHasShellPermission(user.ID)) && middleware.ContainerActionEnvAllowed("shell") && dbCanShell

	svc := res.Service
	if !canShell && svc.Spec.TaskTemplate.ContainerSpec != nil && len(svc.Spec.TaskTemplate.ContainerSpec.Env) > 0 {
		maskedEnv := make([]string, len(svc.Spec.TaskTemplate.ContainerSpec.Env))
		for i, envStr := range svc.Spec.TaskTemplate.ContainerSpec.Env {
			parts := strings.SplitN(envStr, "=", 2)
			maskedEnv[i] = parts[0] + "=••••••••••••"
		}
		svc.Spec.TaskTemplate.ContainerSpec.Env = maskedEnv
	}

	return c.JSON(http.StatusOK, svc)
}

func (s *Server) handleSwarmScaleService(c echo.Context) error {
	ctx := context.Background()
	id := c.Param("id")
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)

	if !middleware.ContainerActionEnvAllowed("restart") {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Restart / scaling actions are disabled on this server."})
	}

	if !config.AuthDisabled {
		can, err := middleware.StaffHasContainerActionPermission("restart", user.ID)
		if err != nil || !can {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Scaling Swarm services requires restart permissions"})
		}
	}

	var req struct {
		Replicas uint64 `json:"replicas"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	cli := s.deps.Docker

	inspectRes, err := cli.ServiceInspect(ctx, id, client.ServiceInspectOptions{})
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Service not found"})
	}

	service := inspectRes.Service
	isAdmin := s.k8sUserIsAdmin(user.ID)
	if !isAdmin {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.MatchesAnyPattern(patterns, service.Spec.Name) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
		}
	}

	if service.Spec.Mode.Global != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cannot scale global service"})
	}

	spec := service.Spec
	if spec.Mode.Replicated == nil {
		spec.Mode.Replicated = &swarm.ReplicatedService{}
	}
	spec.Mode.Replicated.Replicas = &req.Replicas

	updateOpts := client.ServiceUpdateOptions{
		Version: service.Meta.Version,
		Spec:    spec,
	}

	updateRes, err := cli.ServiceUpdate(ctx, id, updateOpts)
	if err != nil {
		s.audit(user.ID, s.auditActor(user), "swarm:scale", service.Spec.Name, "Failed", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	s.audit(user.ID, s.auditActor(user), "swarm:scale", service.Spec.Name, "Success", fmt.Sprintf("Scaled to %d replicas", req.Replicas))
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "ok",
		"replicas": req.Replicas,
		"warnings": updateRes.Warnings,
	})
}

func (s *Server) handleSwarmDeleteService(c echo.Context) error {
	ctx := context.Background()
	id := c.Param("id")
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)

	if !middleware.ContainerActionEnvAllowed("remove") {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Delete / removal actions are disabled on this server."})
	}

	if !config.AuthDisabled {
		can, err := middleware.StaffHasContainerActionPermission("remove", user.ID)
		if err != nil || !can {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Deleting Swarm services requires delete permissions"})
		}
	}

	cli := s.deps.Docker

	inspectRes, err := cli.ServiceInspect(ctx, id, client.ServiceInspectOptions{})
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Service not found"})
	}
	service := inspectRes.Service

	isAdmin := s.k8sUserIsAdmin(user.ID)
	if !isAdmin {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.MatchesAnyPattern(patterns, service.Spec.Name) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
		}
	}

	_, err = cli.ServiceRemove(ctx, id, client.ServiceRemoveOptions{})
	if err != nil {
		s.audit(user.ID, s.auditActor(user), "swarm:delete", service.Spec.Name, "Failed", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	s.audit(user.ID, s.auditActor(user), "swarm:delete", service.Spec.Name, "Success", "Service removed from swarm")
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleSwarmListNodes(c echo.Context) error {
	ctx := context.Background()
	active, _, isManager, _, _ := s.checkSwarmActive(ctx)
	if !active || !isManager {
		return c.JSON(http.StatusOK, []models.SwarmNode{})
	}

	cli := s.deps.Docker

	res, err := cli.NodeList(ctx, client.NodeListOptions{})
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	}

	list := make([]models.SwarmNode, 0, len(res.Items))
	for _, n := range res.Items {
		isManager := n.Spec.Role == swarm.NodeRoleManager
		isLeader := false
		if n.ManagerStatus != nil && n.ManagerStatus.Leader {
			isLeader = true
		}

		var cpus int64
		var mem int64
		cpus = n.Description.Resources.NanoCPUs / 1e9
		mem = n.Description.Resources.MemoryBytes

		list = append(list, models.SwarmNode{
			ID:           n.ID,
			Hostname:     n.Description.Hostname,
			Role:         string(n.Spec.Role),
			State:        string(n.Status.State),
			Availability: string(n.Spec.Availability),
			Manager:      isManager,
			Leader:       isLeader,
			EngineVer:    n.Description.Engine.EngineVersion,
			Addr:         n.Status.Addr,
			CPUs:         cpus,
			Memory:       mem,
			Labels:       n.Spec.Labels,
		})
	}

	return c.JSON(http.StatusOK, list)
}

func (s *Server) handleSwarmUpdateNodeAvailability(c echo.Context) error {
	ctx := context.Background()
	id := c.Param("id")
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)

	if !config.AuthDisabled && !user.IsAdmin {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Updating node availability requires administrator access"})
	}

	var req struct {
		Availability string `json:"availability"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	targetAvail := swarm.NodeAvailability(strings.ToLower(req.Availability))
	if targetAvail != swarm.NodeAvailabilityActive &&
		targetAvail != swarm.NodeAvailabilityPause &&
		targetAvail != swarm.NodeAvailabilityDrain {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Availability must be active, pause, or drain"})
	}

	cli := s.deps.Docker

	inspectRes, err := cli.NodeInspect(ctx, id, client.NodeInspectOptions{})
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Node not found"})
	}
	node := inspectRes.Node

	spec := node.Spec
	spec.Availability = targetAvail

	_, err = cli.NodeUpdate(ctx, id, client.NodeUpdateOptions{
		Version: node.Version,
		Spec:    spec,
	})
	if err != nil {
		s.audit(user.ID, s.auditActor(user), "swarm:node:update", node.Description.Hostname, "Failed", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	s.audit(user.ID, s.auditActor(user), "swarm:node:update", node.Description.Hostname, "Success", fmt.Sprintf("Changed availability to %s", targetAvail))
	return c.JSON(http.StatusOK, map[string]string{
		"status":       "ok",
		"availability": string(targetAvail),
	})
}

func (s *Server) handleSwarmListTasks(c echo.Context) error {
	ctx := context.Background()
	active, _, isManager, _, _ := s.checkSwarmActive(ctx)
	if !active || !isManager {
		return c.JSON(http.StatusOK, []models.SwarmTask{})
	}

	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	isAdmin := s.k8sUserIsAdmin(user.ID)

	cli := s.deps.Docker

	serviceFilter := strings.TrimSpace(c.QueryParam("service"))

	tasksRes, err := cli.TaskList(ctx, client.TaskListOptions{})
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	}

	// Build map of serviceID -> serviceName and nodeID -> hostname
	servicesRes, _ := cli.ServiceList(ctx, client.ServiceListOptions{})
	svcNameMap := make(map[string]string)
	for _, s := range servicesRes.Items {
		svcNameMap[s.ID] = s.Spec.Name
	}

	nodesRes, _ := cli.NodeList(ctx, client.NodeListOptions{})
	nodeNameMap := make(map[string]string)
	for _, n := range nodesRes.Items {
		nodeNameMap[n.ID] = n.Description.Hostname
	}

	var patterns []string
	if !isAdmin {
		patterns = access.GetAuthorizedPatterns(user.ID)
	}

	list := make([]models.SwarmTask, 0, len(tasksRes.Items))
	for _, t := range tasksRes.Items {
		svcName := svcNameMap[t.ServiceID]
		if svcName == "" {
			svcName = t.ServiceID
		}

		if serviceFilter != "" && t.ServiceID != serviceFilter && svcName != serviceFilter {
			continue
		}

		if !isAdmin && !access.MatchesAnyPattern(patterns, svcName) {
			continue
		}

		image := ""
		if t.Spec.ContainerSpec != nil {
			image = t.Spec.ContainerSpec.Image
		}

		list = append(list, models.SwarmTask{
			ID:           t.ID,
			ServiceID:    t.ServiceID,
			ServiceName:  svcName,
			NodeID:       t.NodeID,
			NodeHostname: nodeNameMap[t.NodeID],
			Slot:         t.Slot,
			Image:        image,
			CurrentState: string(t.Status.State),
			DesiredState: string(t.DesiredState),
			Message:      t.Status.Message,
			Error:        t.Status.Err,
			CreatedAt:    t.CreatedAt.Unix(),
			UpdatedAt:    t.UpdatedAt.Unix(),
		})
	}

	return c.JSON(http.StatusOK, list)
}

func (s *Server) registerSwarmWebSocketRoutes() {
	cli := s.deps.Docker
	if cli == nil {
		return
	}

	s.echo.GET("/ws/swarm/services/:id/logs", func(c echo.Context) error {
		id := c.Param("id")

		userClaims, err := middleware.AuthenticateWS(c)
		if err != nil {
			return middleware.WSAuthError(c, err)
		}

		inspectRes, err := cli.ServiceInspect(context.Background(), id, client.ServiceInspectOptions{})
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}
		serviceName := inspectRes.Service.Spec.Name

		if !userClaims.IsAdmin {
			patterns := access.GetAuthorizedPatterns(userClaims.ID)
			if !access.MatchesAnyPattern(patterns, serviceName) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied: You do not have permission to view logs for this Swarm service."})
			}
		}

		ws, err := middleware.UpgradeAuthenticatedWS(c)
		if err != nil {
			return err
		}
		defer ws.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		out, err := cli.ServiceLogs(ctx, id, client.ServiceLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Tail:       "100",
			Timestamps: true,
			Details:    true,
		})
		if err != nil {
			return err
		}
		defer out.Close()

		slotMap, shortSvc := s.buildSwarmTaskSlotMap(ctx, id, serviceName)

		_ = dockerutil.StreamLogs(out, false, func(payload []byte) error {
			msg := formatSwarmLogLine(string(payload), slotMap, shortSvc)
			if msg != "" {
				if err := ws.WriteMessage(1, []byte(msg)); err != nil {
					return err
				}
			}
			return nil
		})
		return nil
	})
}

func (s *Server) checkSwarmServiceAccess(c echo.Context, serviceID string) (string, error) {
	ctx := context.Background()
	active, _, _, _, _ := s.checkSwarmActive(ctx)
	if !active {
		return "", echo.NewHTTPError(http.StatusNotFound, "Swarm is not active")
	}

	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	cli := s.deps.Docker

	inspectRes, err := cli.ServiceInspect(context.Background(), serviceID, client.ServiceInspectOptions{})
	if err != nil {
		return "", echo.NewHTTPError(http.StatusNotFound, "Service not found")
	}
	svcName := inspectRes.Service.Spec.Name

	isAdmin := s.k8sUserIsAdmin(user.ID)
	if !isAdmin {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.MatchesAnyPattern(patterns, svcName) {
			return "", echo.NewHTTPError(http.StatusForbidden, "Access Denied")
		}
	}
	return svcName, nil
}

func (s *Server) handleSwarmServiceLogsHistory(c echo.Context) error {
	id := c.Param("id")
	svcName, err := s.checkSwarmServiceAccess(c, id)
	if err != nil {
		return err
	}

	cli := s.deps.Docker
	limit := 100
	if tailParam := c.QueryParam("tail"); tailParam != "" {
		if val, err := strconv.Atoi(tailParam); err == nil && val > 0 {
			limit = val
		}
	}

	opts := client.ServiceLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     false,
		Tail:       strconv.Itoa(limit),
		Details:    true,
	}

	sinceStr := c.QueryParam("since")
	untilStr := c.QueryParam("until")
	if sinceStr != "" {
		opts.Since = sinceStr
	}

	out, err := cli.ServiceLogs(context.Background(), id, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer out.Close()

	slotMap, shortSvc := s.buildSwarmTaskSlotMap(context.Background(), id, svcName)

	var rawLogs []string
	_ = dockerutil.StreamLogs(out, false, func(payload []byte) error {
		line := formatSwarmLogLine(string(payload), slotMap, shortSvc)
		if line != "" {
			rawLogs = append(rawLogs, line)
		}
		return nil
	})

	// Swarm service logs can return lines grouped by task/container from docker daemon.
	// Sort all lines chronologically by their embedded timestamp.
	extractLogTime := func(l string) time.Time {
		tsStr := l
		if strings.Contains(l, " | ") {
			parts := strings.SplitN(l, " | ", 2)
			tsStr = parts[1]
		}
		parts := strings.SplitN(strings.TrimSpace(tsStr), " ", 2)
		if len(parts) > 0 {
			if t, err := parseLogTimeQuery(parts[0]); err == nil {
				return t
			}
		}
		return time.Time{}
	}

	sort.SliceStable(rawLogs, func(i, j int) bool {
		ti := extractLogTime(rawLogs[i])
		tj := extractLogTime(rawLogs[j])
		return ti.Before(tj)
	})

	var logs []string
	if untilStr != "" {
		untilTime, err := parseLogTimeQuery(untilStr)
		if err == nil {
			for _, line := range rawLogs {
				t := extractLogTime(line)
				if t.IsZero() || !t.After(untilTime) {
					logs = append(logs, line)
				}
			}
		} else {
			logs = rawLogs
		}
	} else {
		logs = rawLogs
	}

	if len(logs) > limit {
		logs = logs[len(logs)-limit:]
	}

	return c.JSON(http.StatusOK, logs)
}

func (s *Server) handleSwarmServiceLogsDownload(c echo.Context) error {
	id := c.Param("id")
	svcName, err := s.checkSwarmServiceAccess(c, id)
	if err != nil {
		return err
	}

	cli := s.deps.Docker
	opts := client.ServiceLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     false,
		Details:    true,
	}

	if since := c.QueryParam("since"); since != "" {
		opts.Since = since
	}

	out, err := cli.ServiceLogs(context.Background(), id, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer out.Close()

	c.Response().Header().Set(echo.HeaderContentType, "text/plain; charset=utf-8")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s_service_logs.log\"", svcName))
	c.Response().WriteHeader(http.StatusOK)

	_ = dockerutil.StreamLogs(out, false, func(payload []byte) error {
		_, err := c.Response().Write(payload)
		return err
	})
	return nil
}

func (s *Server) handleSwarmServiceLogsCount(c echo.Context) error {
	id := c.Param("id")
	if _, err := s.checkSwarmServiceAccess(c, id); err != nil {
		return err
	}

	cli := s.deps.Docker
	out, err := cli.ServiceLogs(context.Background(), id, client.ServiceLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer out.Close()

	count := 0
	_ = dockerutil.StreamLogs(out, false, func(payload []byte) error {
		count += strings.Count(string(payload), "\n")
		if !strings.HasSuffix(string(payload), "\n") {
			count++
		}
		return nil
	})
	return c.JSON(http.StatusOK, map[string]int{"total": count})
}

func (s *Server) buildSwarmTaskSlotMap(ctx context.Context, serviceID, serviceName string) (map[string]int, string) {
	cli := s.deps.Docker
	slotMap := make(map[string]int)
	shortSvc := serviceName
	if idx := strings.LastIndex(shortSvc, "_"); idx != -1 {
		shortSvc = shortSvc[idx+1:]
	}

	tasksRes, err := cli.TaskList(ctx, client.TaskListOptions{})
	if err == nil {
		for _, t := range tasksRes.Items {
			if t.ServiceID == serviceID {
				if t.Slot > 0 {
					slotMap[t.ID] = t.Slot
				}
			}
		}
	}
	return slotMap, shortSvc
}

func formatSwarmLogLine(raw string, slotMap map[string]int, shortSvc string) string {
	line := strings.TrimRight(raw, "\r\n")
	if line == "" {
		return ""
	}

	// Case 1: Already has task name prefix: e.g. "adminer-swarm_adminer.1.iez3... | <timestamp> <message>"
	if strings.Contains(line, " | ") {
		return line
	}

	// Case 2: Standard docker daemon format with Details=true:
	// "<timestamp> com.docker.swarm.node.id=...,com.docker.swarm.service.id=...,com.docker.swarm.task.id=<tid> <message>"
	parts := strings.SplitN(line, " ", 3)
	if len(parts) >= 2 {
		ts := parts[0]
		rest := strings.TrimPrefix(line[len(ts):], " ")

		if strings.HasPrefix(rest, "com.docker.swarm.") {
			subParts := strings.SplitN(rest, " ", 2)
			details := subParts[0]
			msg := ""
			if len(subParts) > 1 {
				msg = subParts[1]
			}

			// Extract task.id from details
			slot := 1
			for _, item := range strings.Split(details, ",") {
				if strings.HasPrefix(item, "com.docker.swarm.task.id=") {
					tid := strings.TrimPrefix(item, "com.docker.swarm.task.id=")
					if s, ok := slotMap[tid]; ok && s > 0 {
						slot = s
					} else {
						var h int
						for _, ch := range tid {
							h += int(ch)
						}
						slot = (h % 5) + 1
					}
					break
				}
			}
			return fmt.Sprintf("%s.%d | %s %s", shortSvc, slot, ts, msg)
		}
	}

	return line
}



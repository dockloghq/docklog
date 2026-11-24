package server

import (
	"context"
	"net/http"
	"strings"

	"docklog/access"
	"docklog/config"
	"docklog/db"
	"docklog/k8s"
	"docklog/middleware"
	"docklog/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func (s *Server) registerK8sRoutes(r *echo.Group) {
	k8sUnavailable := func(c echo.Context) error {
		msg := strings.TrimSpace(config.K8sConfigError)
		if msg == "" {
			msg = "Kubernetes client is not configured"
		}
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": msg})
	}

	emptyList := func(c echo.Context) error {
		return c.JSON(http.StatusOK, []interface{}{})
	}

	if s.deps.K8s == nil || !config.KubernetesEnabled() {
		r.GET("/namespaces", emptyList)
		r.GET("/pods", emptyList)
		r.GET("/namespaces/:namespace/pods/:pod", k8sUnavailable)
		r.GET("/namespaces/:namespace/pods/:pod/stats-now", k8sUnavailable)
		r.POST("/namespaces/:namespace/pods/:pod/action", k8sUnavailable)
		r.GET("/deployments", emptyList)
		r.GET("/hpas", emptyList)
		r.GET("/services", emptyList)
		r.GET("/events", emptyList)
		r.GET("/k8s/overview", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"pods":        0,
				"deployments": 0,
				"services":    0,
			})
		})
		return
	}

	r.GET("/namespaces", s.handleListNamespaces)
	r.GET("/pods", s.handleListPods)
	r.GET("/namespaces/:namespace/pods/:pod", s.handleGetPodDetail)
	r.GET("/namespaces/:namespace/pods/:pod/stats-now", s.handlePodStatsNow)
	r.POST("/namespaces/:namespace/pods/:pod/action", s.handlePodAction)
	r.GET("/deployments", s.handleListDeployments)
	r.GET("/hpas", s.handleListHPAs)
	r.GET("/services", s.handleListServices)
	r.GET("/events", s.handleListEvents)
	r.GET("/k8s/overview", s.handleK8sOverview)
}

func (s *Server) k8sNamespaceFromQuery(c echo.Context) string {
	namespace := strings.TrimSpace(c.QueryParam("namespace"))
	if namespace == "" {
		namespace = config.DefaultK8sNamespace()
	}
	return namespace
}

func (s *Server) k8sUserIsAdmin(userID int) bool {
	var dbIsAdmin bool
	db.DB.QueryRow("SELECT COALESCE(is_admin, 0) FROM users WHERE id = ?", userID).Scan(&dbIsAdmin)
	return dbIsAdmin
}

func (s *Server) handleListNamespaces(c echo.Context) error {
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)

	namespaces, err := k8s.ListNamespaces(context.Background(), s.deps.K8s)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	}

	if s.k8sUserIsAdmin(user.ID) {
		return c.JSON(http.StatusOK, namespaces)
	}

	patterns := access.GetAuthorizedPatterns(user.ID)
	filtered := make([]models.K8sNamespace, 0, len(namespaces))
	for _, ns := range namespaces {
		if access.NamespaceVisible(patterns, ns.Name) {
			filtered = append(filtered, ns)
		}
	}
	return c.JSON(http.StatusOK, filtered)
}

func (s *Server) handleListPods(c echo.Context) error {
	namespace := s.k8sNamespaceFromQuery(c)
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	if !s.k8sUserIsAdmin(user.ID) {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.NamespaceVisible(patterns, namespace) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
		}
	}

	pods, err := k8s.ListPods(context.Background(), s.deps.K8s, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "not allowed") {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	}

	if s.k8sUserIsAdmin(user.ID) {
		return c.JSON(http.StatusOK, pods)
	}

	patterns := access.GetAuthorizedPatterns(user.ID)
	filtered := make([]models.K8sPod, 0, len(pods))
	for _, pod := range pods {
		if access.PodAllowed(patterns, pod.Namespace, pod.Name) {
			filtered = append(filtered, pod)
		}
	}
	return c.JSON(http.StatusOK, filtered)
}

func (s *Server) handleGetPodDetail(c echo.Context) error {
	namespace := strings.TrimSpace(c.Param("namespace"))
	podName := strings.TrimSpace(c.Param("pod"))
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)

	if err := s.authorizePodAccess(user, namespace, podName); err != nil {
		return err
	}

	pod, err := k8s.GetPodDetail(context.Background(), s.deps.K8s, namespace, podName)
	if err != nil {
		return k8sListError(c, err)
	}

	var dbIsAdmin bool
	if config.AuthDisabled {
		dbIsAdmin = true
	} else if user != nil {
		db.DB.QueryRow("SELECT COALESCE(is_admin, 0) FROM users WHERE id = ?", user.ID).Scan(&dbIsAdmin)
	}
	var dbCanShell bool
	if user != nil {
		db.DB.QueryRow("SELECT COALESCE(can_shell, 0) FROM users WHERE id = ?", user.ID).Scan(&dbCanShell)
	}

	canShell := (dbIsAdmin && middleware.ContainerActionEnvAllowed("shell"))
	if !dbIsAdmin && user != nil {
		canShell = middleware.StaffHasShellPermission(user.ID) && middleware.ContainerActionEnvAllowed("shell") && dbCanShell
	}

	if !canShell {
		for i := range pod.Containers {
			for j, envStr := range pod.Containers[i].Env {
				parts := strings.SplitN(envStr, "=", 2)
				pod.Containers[i].Env[j] = parts[0] + "=••••••••••••"
			}
		}
		for i := range pod.InitContainers {
			for j, envStr := range pod.InitContainers[i].Env {
				parts := strings.SplitN(envStr, "=", 2)
				pod.InitContainers[i].Env[j] = parts[0] + "=••••••••••••"
			}
		}
	}

	return c.JSON(http.StatusOK, pod)
}

func (s *Server) handlePodStatsNow(c echo.Context) error {
	namespace := strings.TrimSpace(c.Param("namespace"))
	podName := strings.TrimSpace(c.Param("pod"))
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)

	if err := s.authorizePodAccess(user, namespace, podName); err != nil {
		return err
	}

	stats, err := k8s.GetPodStatsNow(context.Background(), s.deps.K8s, namespace, podName)
	if err != nil {
		return k8sListError(c, err)
	}
	return c.JSON(http.StatusOK, stats)
}

func (s *Server) handlePodAction(c echo.Context) error {
	namespace := strings.TrimSpace(c.Param("namespace"))
	podName := strings.TrimSpace(c.Param("pod"))
	action := strings.TrimSpace(c.FormValue("action"))
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	actor := s.auditActor(user)
	target := namespace + "/" + podName

	if namespace == "" || podName == "" {
		s.audit(user.ID, actor, action, target, "Forbidden", "Missing pod namespace/name")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "namespace and pod are required"})
	}
	switch action {
	case "start", "stop", "restart", "remove":
	default:
		s.audit(user.ID, actor, action, target, "Forbidden", "Invalid pod action")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid action specified."})
	}

	if err := s.authorizePodAccess(user, namespace, podName); err != nil {
		s.audit(user.ID, actor, action, target, "Forbidden", "Pod access denied")
		return err
	}
	if !middleware.ContainerActionEnvAllowed(action) {
		s.audit(user.ID, actor, action, target, "Forbidden", "Action disabled by server env")
		return c.JSON(http.StatusForbidden, map[string]string{"error": "This action is disabled on this server."})
	}
	if !config.AuthDisabled {
		can, err := middleware.StaffHasContainerActionPermission(action, user.ID)
		if err != nil || !can {
			s.audit(user.ID, actor, action, target, "Forbidden", "Action not permitted for account")
			return c.JSON(http.StatusForbidden, map[string]string{"error": "This action is not permitted for this account."})
		}
	}

	if err := k8s.ExecutePodAction(context.Background(), s.deps.K8s, namespace, podName, action); err != nil {
		s.audit(user.ID, actor, action, target, "Error", err.Error())
		return k8sListError(c, err)
	}
	s.audit(user.ID, actor, action, target, "Success", "Pod action requested")
	return c.NoContent(http.StatusOK)
}

func (s *Server) handleListDeployments(c echo.Context) error {
	namespace := s.k8sNamespaceFromQuery(c)
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	if !s.k8sUserIsAdmin(user.ID) {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.NamespaceVisible(patterns, namespace) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
		}
	}

	items, err := k8s.ListDeployments(context.Background(), s.deps.K8s, namespace)
	if err != nil {
		return k8sListError(c, err)
	}
	if s.k8sUserIsAdmin(user.ID) {
		return c.JSON(http.StatusOK, items)
	}

	patterns := access.GetAuthorizedPatterns(user.ID)
	filtered := make([]models.K8sDeployment, 0, len(items))
	for _, item := range items {
		if access.ResourceAllowed(patterns, item.Namespace, item.Name) {
			filtered = append(filtered, item)
		}
	}
	return c.JSON(http.StatusOK, filtered)
}

func (s *Server) handleListHPAs(c echo.Context) error {
	namespace := s.k8sNamespaceFromQuery(c)
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	if !s.k8sUserIsAdmin(user.ID) {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.NamespaceVisible(patterns, namespace) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
		}
	}

	items, err := k8s.ListHPAs(context.Background(), s.deps.K8s, namespace)
	if err != nil {
		return k8sListError(c, err)
	}
	if s.k8sUserIsAdmin(user.ID) {
		return c.JSON(http.StatusOK, items)
	}

	patterns := access.GetAuthorizedPatterns(user.ID)
	filtered := make([]models.K8sHPA, 0, len(items))
	for _, item := range items {
		if access.ResourceAllowed(patterns, item.Namespace, item.Name) ||
			access.ResourceAllowed(patterns, item.Namespace, item.TargetName) {
			filtered = append(filtered, item)
		}
	}
	return c.JSON(http.StatusOK, filtered)
}

func (s *Server) handleListServices(c echo.Context) error {
	namespace := s.k8sNamespaceFromQuery(c)
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	if !s.k8sUserIsAdmin(user.ID) {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.NamespaceVisible(patterns, namespace) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
		}
	}

	items, err := k8s.ListServices(context.Background(), s.deps.K8s, namespace)
	if err != nil {
		return k8sListError(c, err)
	}
	if s.k8sUserIsAdmin(user.ID) {
		return c.JSON(http.StatusOK, items)
	}

	patterns := access.GetAuthorizedPatterns(user.ID)
	filtered := make([]models.K8sService, 0, len(items))
	for _, item := range items {
		if access.ResourceAllowed(patterns, item.Namespace, item.Name) {
			filtered = append(filtered, item)
		}
	}
	return c.JSON(http.StatusOK, filtered)
}

func (s *Server) handleListEvents(c echo.Context) error {
	namespace := s.k8sNamespaceFromQuery(c)
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	if !s.k8sUserIsAdmin(user.ID) {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.NamespaceVisible(patterns, namespace) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
		}
	}

	items, err := k8s.ListEvents(context.Background(), s.deps.K8s, namespace)
	if err != nil {
		return k8sListError(c, err)
	}
	if s.k8sUserIsAdmin(user.ID) {
		return c.JSON(http.StatusOK, items)
	}

	patterns := access.GetAuthorizedPatterns(user.ID)
	filtered := make([]models.K8sEvent, 0, len(items))
	for _, item := range items {
		if access.ResourceAllowed(patterns, item.Namespace, item.InvolvedName) {
			filtered = append(filtered, item)
		}
	}
	return c.JSON(http.StatusOK, filtered)
}

func (s *Server) handleK8sOverview(c echo.Context) error {
	namespace := s.k8sNamespaceFromQuery(c)
	token := c.Get("user").(*jwt.Token)
	user := token.Claims.(*models.UserClaims)
	if !s.k8sUserIsAdmin(user.ID) {
		patterns := access.GetAuthorizedPatterns(user.ID)
		if !access.NamespaceVisible(patterns, namespace) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access Denied"})
		}
	}
	overview, err := k8s.GetOverview(context.Background(), s.deps.K8s, namespace)
	if err != nil {
		return k8sListError(c, err)
	}
	return c.JSON(http.StatusOK, overview)
}

func k8sListError(c echo.Context, err error) error {
	if strings.Contains(err.Error(), "not allowed") {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
}

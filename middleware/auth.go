package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"docklog/config"
	"docklog/db"
	"docklog/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	ClientHeaderWeb        = "web"
	HeaderDockLogClient    = "X-DockLog-Client"
	MinPasswordLength      = 8
	MaxContainerPatternLen = 256
	TokenTypeAccess        = "access"
	TokenTypeRefresh       = "refresh"
)

var (
	ClientAccessEnabled bool
	allowedOrigins      []string
	LoginRateLimit      loginRateLimiter
	upgrader            = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

const (
	accessTokenTTL  = 7 * 24 * time.Hour
	refreshTokenTTL = 30 * 24 * time.Hour
)

type loginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func InitWSUpgrader() {
	upgrader.CheckOrigin = IsWSAccessAllowed
}

func InitClientAccess() {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("CLIENT_ACCESS")))
	ClientAccessEnabled = !config.AuthDisabled && mode != "off"
	allowedOrigins = parseCSVEnv(os.Getenv("ALLOWED_ORIGINS"))
	config.TrustProxy = os.Getenv("TRUST_PROXY") == "true"

	if ClientAccessEnabled {
		log.Println("Client access: strict (Vue web UI origin validation; native mobile clients without browser Origin)")
		if config.TrustProxy {
			log.Println("TRUST_PROXY enabled: honoring X-Forwarded-Host / X-Forwarded-Proto for origin checks")
		}
	}
}

func parseCSVEnv(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var values []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func isProduction() bool {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	goEnv := strings.ToLower(strings.TrimSpace(os.Getenv("GO_ENV")))
	return env == "production" || goEnv == "production"
}

func IsPasswordStrongEnough(password string) bool {
	return len(password) >= MinPasswordLength
}

func isLocalhostHost(host string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.ToLower(strings.Trim(host, "[]"))
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func requestHost(r *http.Request) string {
	if config.TrustProxy {
		if host := r.Header.Get("X-Forwarded-Host"); host != "" {
			return strings.TrimSpace(strings.Split(host, ",")[0])
		}
	}
	return r.Host
}

func requestScheme(r *http.Request) string {
	if config.TrustProxy {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			return strings.TrimSpace(strings.Split(proto, ",")[0])
		}
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func sameOriginURL(r *http.Request) string {
	return requestScheme(r) + "://" + requestHost(r)
}

func CorsOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	for _, allowed := range allowedOrigins {
		if allowedOriginEntryMatches(origin, allowed) {
			return true
		}
	}
	if !isProduction() {
		parsed, err := url.Parse(origin)
		if err == nil && isLocalhostHost(parsed.Host) {
			return true
		}
	}
	return false
}

func normalizeHost(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.ToLower(strings.Trim(host, "[]"))
}

func allowedOriginEntryMatches(origin string, allowed string) bool {
	if origin == allowed {
		return true
	}
	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if strings.Contains(allowed, "://") {
		parsedAllowed, err := url.Parse(allowed)
		if err != nil {
			return false
		}
		return normalizeHost(parsedOrigin.Host) == normalizeHost(parsedAllowed.Host)
	}
	return normalizeHost(parsedOrigin.Host) == normalizeHost(allowed)
}

func allowedRefererMatches(referer string, allowed string) bool {
	if referer == allowed || strings.HasPrefix(referer, allowed+"/") {
		return true
	}
	parsedReferer, err := url.Parse(referer)
	if err != nil {
		return false
	}
	if strings.Contains(allowed, "://") {
		parsedAllowed, err := url.Parse(allowed)
		if err != nil {
			return false
		}
		return normalizeHost(parsedReferer.Host) == normalizeHost(parsedAllowed.Host)
	}
	return normalizeHost(parsedReferer.Host) == normalizeHost(allowed)
}

func originHostMatchesRequest(origin string, r *http.Request) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return normalizeHost(parsed.Host) == normalizeHost(requestHost(r))
}

func refererHostMatchesRequest(referer string, r *http.Request) bool {
	parsed, err := url.Parse(referer)
	if err != nil {
		return false
	}
	return normalizeHost(parsed.Host) == normalizeHost(requestHost(r))
}

func originMatchesAllowed(origin string, r *http.Request) bool {
	if origin == sameOriginURL(r) {
		return true
	}
	if originHostMatchesRequest(origin, r) {
		return true
	}
	for _, allowed := range allowedOrigins {
		if allowedOriginEntryMatches(origin, allowed) {
			return true
		}
	}
	if !isProduction() {
		parsed, err := url.Parse(origin)
		if err == nil && isLocalhostHost(parsed.Host) {
			return true
		}
	}
	return false
}

func refererMatchesAllowed(referer string, r *http.Request) bool {
	sameOrigin := sameOriginURL(r)
	if referer == sameOrigin || strings.HasPrefix(referer, sameOrigin+"/") {
		return true
	}
	if refererHostMatchesRequest(referer, r) {
		return true
	}
	for _, allowed := range allowedOrigins {
		if allowedRefererMatches(referer, allowed) {
			return true
		}
	}
	if !isProduction() {
		parsed, err := url.Parse(referer)
		if err == nil && isLocalhostHost(parsed.Host) {
			return true
		}
	}
	return false
}

func requestHostAllowed(r *http.Request) bool {
	host := normalizeHost(requestHost(r))
	if host == "" {
		return false
	}
	if !isProduction() && isLocalhostHost(host) {
		return true
	}
	for _, allowed := range allowedOrigins {
		if strings.Contains(allowed, "://") {
			parsed, err := url.Parse(allowed)
			if err == nil && normalizeHost(parsed.Host) == host {
				return true
			}
		} else if normalizeHost(allowed) == host {
			return true
		}
	}
	if len(allowedOrigins) == 0 {
		return true
	}
	return false
}

func isWebOriginAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin != "" && origin != "null" {
		return originMatchesAllowed(origin, r)
	}
	referer := r.Header.Get("Referer")
	if referer != "" {
		return refererMatchesAllowed(referer, r)
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "same-origin", "same-site":
		return requestHostAllowed(r)
	}
	return false
}

func isWebHTTPClientAllowed(r *http.Request) bool {
	if strings.ToLower(r.Header.Get(HeaderDockLogClient)) != ClientHeaderWeb {
		return false
	}
	if isWebOriginAllowed(r) {
		return true
	}
	// Fallback for privacy-focused browsers (e.g. Firefox ETP, Private Browsing)
	// that strip Origin, Referer, and Sec-Fetch-Site on GET/sub-requests,
	// provided the request targets the current host and supplies a Bearer auth token.
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") && requestHostAllowed(r) {
		return true
	}
	return false
}

func isNativeAppRequest(r *http.Request) bool {
	if r.Header.Get("Origin") != "" || r.Header.Get("Referer") != "" {
		return false
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "same-origin", "same-site", "cross-site":
		return false
	}
	return true
}

func isClientAccessAllowed(r *http.Request) bool {
	if !ClientAccessEnabled {
		return true
	}
	if isWebHTTPClientAllowed(r) {
		return true
	}
	return isNativeAppRequest(r)
}

func IsWSAccessAllowed(r *http.Request) bool {
	if !ClientAccessEnabled {
		return true
	}
	if isWebOriginAllowed(r) {
		return true
	}
	// Fallback for privacy-focused browsers connecting to WS with auth protocol
	if extractWSToken(r) != "" && requestHostAllowed(r) {
		return true
	}
	return isNativeAppRequest(r)
}

func ClientAccessConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled": ClientAccessEnabled,
		"web": map[string]string{
			"client_header": HeaderDockLogClient + "=web",
			"origin":        "Vue web UI. Must match this server or ALLOWED_ORIGINS",
		},
		"native_mobile": "Flutter app (Android/iOS, com.docklog.app). No Origin; JWT auth required",
	}
}

func ClientAccessMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !ClientAccessEnabled {
				return next(c)
			}
			path := c.Request().URL.Path
			if !strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/ws") {
				return next(c)
			}
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}
			if strings.HasPrefix(path, "/ws") {
				if !IsWSAccessAllowed(c.Request()) {
					return c.JSON(http.StatusForbidden, map[string]string{
						"error": "Access denied: WebSocket must originate from the web app or a native client",
					})
				}
				return next(c)
			}
			if !isClientAccessAllowed(c.Request()) {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Access denied: request must originate from the web app or a native client",
				})
			}
			return next(c)
		}
	}
}

func (rl *loginRateLimiter) IsLimited(key string, max int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.attempts == nil {
		rl.attempts = make(map[string][]time.Time)
	}
	now := time.Now()
	cutoff := now.Add(-window)
	var recent []time.Time
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	rl.attempts[key] = recent
	return len(recent) >= max
}

func (rl *loginRateLimiter) RecordFailure(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.attempts == nil {
		rl.attempts = make(map[string][]time.Time)
	}
	rl.attempts[key] = append(rl.attempts[key], time.Now())
}

func (rl *loginRateLimiter) Clear(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

func ContainerActionEnvAllowed(action string) bool {
	switch action {
	case "start":
		return config.CanStart
	case "stop":
		return config.CanStop
	case "restart":
		return config.CanRestart
	case "remove":
		return config.CanDelete
	case "shell":
		return config.AllowShell
	default:
		return false
	}
}

func ClampStaffActionPermissions(canStart, canStop, canRestart, canDelete, canShell bool) (bool, bool, bool, bool, bool) {
	if !config.CanStart {
		canStart = false
	}
	if !config.CanStop {
		canStop = false
	}
	if !config.CanRestart {
		canRestart = false
	}
	if !config.CanDelete {
		canDelete = false
	}
	if !config.AllowShell {
		canShell = false
	}
	return canStart, canStop, canRestart, canDelete, canShell
}

func staffContainerActionQuery(action string) string {
	switch action {
	case "start":
		return "SELECT can_start FROM users WHERE id = ? AND is_active = 1"
	case "stop":
		return "SELECT can_stop FROM users WHERE id = ? AND is_active = 1"
	case "restart":
		return "SELECT can_restart FROM users WHERE id = ? AND is_active = 1"
	case "remove":
		return "SELECT can_delete FROM users WHERE id = ? AND is_active = 1"
	default:
		return ""
	}
}

func StaffHasShellPermission(userID int) bool {
	var can bool
	err := db.DB.QueryRow("SELECT can_shell FROM users WHERE id = ? AND is_active = 1", userID).Scan(&can)
	if err != nil {
		return false
	}
	return can
}

func StaffHasContainerActionPermission(action string, userID int) (bool, error) {
	query := staffContainerActionQuery(action)
	if query == "" {
		return false, nil
	}

	var can bool
	err := db.DB.QueryRow(query, userID).Scan(&can)
	if err != nil {
		return false, err
	}
	return can, nil
}

func extractWSToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}

	proto := r.Header.Get("Sec-WebSocket-Protocol")
	if proto == "" {
		return ""
	}

	parts := strings.Split(proto, ",")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "docklog-auth" && i+1 < len(parts) {
			return strings.TrimSpace(parts[i+1])
		}
	}
	return ""
}

func signUserToken(claims *models.UserClaims, tokenType string, ttl time.Duration) (string, error) {
	c := *claims
	c.TokenType = tokenType
	now := time.Now()
	c.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &c)
	return token.SignedString(config.SecretKey)
}

func IssueTokenPair(claims *models.UserClaims) (string, string, error) {
	access, err := signUserToken(claims, TokenTypeAccess, accessTokenTTL)
	if err != nil {
		return "", "", err
	}
	refresh, err := signUserToken(claims, TokenTypeRefresh, refreshTokenTTL)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func RefreshClaimsFromDB(claims *models.UserClaims) error {
	var changed, active, isAdmin, canStart, canStop, canRestart, canDelete, canShell, isRestricted bool
	var dbPwdVersion int
	var allowedContainers, username string

	err := db.DB.QueryRow(
		`SELECT password_changed, is_active, COALESCE(password_version, 1),
		 is_admin, can_start, can_stop, can_restart, can_delete, can_shell, is_restricted_access, allowed_containers, username
		 FROM users WHERE id = ?`,
		claims.ID,
	).Scan(
		&changed, &active, &dbPwdVersion,
		&isAdmin, &canStart, &canStop, &canRestart, &canDelete, &canShell, &isRestricted, &allowedContainers, &username,
	)
	if err != nil {
		return err
	}

	if !active {
		return fmt.Errorf("account deactivated")
	}
	if claims.PasswordVersion != dbPwdVersion {
		return fmt.Errorf("session invalidated")
	}

	claims.IsAdmin = isAdmin
	claims.CanStart = canStart
	claims.CanStop = canStop
	claims.CanRestart = canRestart
	claims.CanDelete = canDelete
	claims.CanShell = canShell
	claims.IsRestrictedAccess = isRestricted
	claims.AllowedContainers = allowedContainers
	claims.IsActive = active
	if strings.TrimSpace(username) != "" {
		claims.Username = username
	}

	return nil
}

func parseUserToken(tokenStr string) (*models.UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid signing method")
		}
		return config.SecretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token.Claims.(*models.UserClaims), nil
}

func validateUserClaims(claims *models.UserClaims, requirePasswordChanged bool) (*models.UserClaims, error) {
	if err := RefreshClaimsFromDB(claims); err != nil {
		return nil, err
	}

	var changed bool
	err := db.DB.QueryRow("SELECT password_changed FROM users WHERE id = ?", claims.ID).Scan(&changed)
	if err != nil {
		return nil, err
	}
	if requirePasswordChanged && !changed {
		return nil, fmt.Errorf("password change required")
	}

	return claims, nil
}

func validateUserToken(tokenStr string) (*models.UserClaims, error) {
	claims, err := parseUserToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType == TokenTypeRefresh {
		return nil, fmt.Errorf("invalid token")
	}
	return validateUserClaims(claims, true)
}

func ValidateRefreshToken(tokenStr string) (*models.UserClaims, error) {
	claims, err := parseUserToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, fmt.Errorf("invalid token")
	}
	return validateUserClaims(claims, false)
}

func AuthenticateWS(c echo.Context) (*models.UserClaims, error) {
	if config.AuthDisabled {
		return &models.UserClaims{
			ID:                 1,
			Username:           "admin",
			IsAdmin:            true,
			IsRestrictedAccess: false,
			IsActive:           true,
		}, nil
	}

	tokenStr := extractWSToken(c.Request())
	if tokenStr == "" {
		return nil, fmt.Errorf("missing token")
	}
	return validateUserToken(tokenStr)
}

func UpgradeAuthenticatedWS(c echo.Context) (*websocket.Conn, error) {
	responseHeader := http.Header{}
	responseHeader.Set("Sec-WebSocket-Protocol", "docklog-auth")
	return upgrader.Upgrade(c.Response(), c.Request(), responseHeader)
}

func WSAuthError(c echo.Context, err error) error {
	msg := "Authentication required"
	switch err.Error() {
	case "invalid token", "missing token":
		msg = err.Error()
	case "account deactivated":
		msg = "Account deactivated"
	case "session invalidated":
		msg = "Session invalidated"
	case "password change required":
		msg = "Password change required"
	}
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": msg})
}

func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; img-src 'self' data:; connect-src 'self' ws: wss:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
			if c.Scheme() == "https" {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			return next(c)
		}
	}
}

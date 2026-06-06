package main

import (
	"os"
	"testing"
)

func resetClientAccessState() {
	ClientAccessEnabled = true
	allowedOrigins = []string{"https://docklog.example.com"}
	TrustProxy = false
}

func TestWebClientAllowedSameOrigin(t *testing.T) {
	resetClientAccessState()
	req := newTestRequest("GET", "http://docklog.local/api/containers", map[string]string{
		headerDockLogClient: clientHeaderWeb,
		"Origin":            "http://docklog.local",
	})
	if !isClientAccessAllowed(req) {
		t.Fatal("expected same-origin web request to be allowed")
	}
}

func TestWebClientBlockedWithoutHeader(t *testing.T) {
	resetClientAccessState()
	req := newTestRequest("GET", "http://docklog.local/api/containers", map[string]string{
		"Origin": "http://docklog.local",
	})
	if isClientAccessAllowed(req) {
		t.Fatal("expected request without X-DockLog-Client to be blocked")
	}
}

func TestWebClientBlockedForeignOrigin(t *testing.T) {
	resetClientAccessState()
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	req := newTestRequest("GET", "http://docklog.local/api/containers", map[string]string{
		headerDockLogClient: clientHeaderWeb,
		"Origin":            "https://evil.example.com",
	})
	if isClientAccessAllowed(req) {
		t.Fatal("expected foreign origin to be blocked in production")
	}
}

func TestWebClientAllowedListedOrigin(t *testing.T) {
	resetClientAccessState()
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	req := newTestRequest("GET", "http://docklog.local/api/containers", map[string]string{
		headerDockLogClient: clientHeaderWeb,
		"Origin":            "https://docklog.example.com",
	})
	if !isClientAccessAllowed(req) {
		t.Fatal("expected origin listed in ALLOWED_ORIGINS to be allowed")
	}
}

func TestWebClientAllowedViaReverseProxyHost(t *testing.T) {
	resetClientAccessState()
	TrustProxy = true
	req := newTestRequest("GET", "http://127.0.0.1:8000/api/containers", map[string]string{
		headerDockLogClient: clientHeaderWeb,
		"Origin":            "https://docklog.example.com",
		"X-Forwarded-Host":  "docklog.example.com",
		"X-Forwarded-Proto": "https",
	})
	if !originMatchesAllowed("https://docklog.example.com", req) {
		t.Fatal("expected reverse-proxy forwarded host to match configured origin")
	}
}

func TestForwardHeadersIgnoredWithoutTrustProxy(t *testing.T) {
	resetClientAccessState()
	TrustProxy = false
	req := newTestRequest("GET", "http://127.0.0.1:8000/api/containers", map[string]string{
		"X-Forwarded-Host":  "docklog.example.com",
		"X-Forwarded-Proto": "https",
	})
	if got := requestHost(req); got != "127.0.0.1:8000" {
		t.Fatalf("expected requestHost to ignore forwarded host, got %q", got)
	}
	if got := requestScheme(req); got != "http" {
		t.Fatalf("expected requestScheme to ignore forwarded proto, got %q", got)
	}
}

func TestSecFetchSiteWithoutOriginBlocked(t *testing.T) {
	resetClientAccessState()
	req := newTestRequest("GET", "http://docklog.local/api/containers", map[string]string{
		headerDockLogClient: clientHeaderWeb,
		"Sec-Fetch-Site":      "same-origin",
	})
	if isClientAccessAllowed(req) {
		t.Fatal("expected Sec-Fetch-Site without Origin or Referer to be blocked")
	}
}

func TestNativeClientAllowedWithoutOrigin(t *testing.T) {
	resetClientAccessState()
	req := newTestRequest("POST", "http://192.168.1.10:8888/api/token", nil)
	if !isClientAccessAllowed(req) {
		t.Fatal("expected native client without Origin to be allowed")
	}
}

func TestBrowserLikeRequestBlockedWithoutWebHeaders(t *testing.T) {
	resetClientAccessState()
	req := newTestRequest("GET", "http://docklog.local/api/containers", map[string]string{
		"Origin":         "https://evil.example.com",
		"Sec-Fetch-Site": "cross-site",
	})
	if isClientAccessAllowed(req) {
		t.Fatal("expected cross-site browser request without web client header to be blocked")
	}
}

func TestWSWebAllowedByOrigin(t *testing.T) {
	resetClientAccessState()
	req := newTestRequest("GET", "http://docklog.local/ws/logs/abc", map[string]string{
		"Origin": "http://docklog.local",
	})
	if !isWSAccessAllowed(req) {
		t.Fatal("expected browser websocket with same origin to be allowed")
	}
}

func TestWSNativeAllowedWithoutOrigin(t *testing.T) {
	resetClientAccessState()
	req := newTestRequest("GET", "http://192.168.1.10:8888/ws/logs/abc", nil)
	if !isWSAccessAllowed(req) {
		t.Fatal("expected native websocket without Origin to be allowed")
	}
}

func TestClientAccessDisabledAllowsDirectAPI(t *testing.T) {
	resetClientAccessState()
	ClientAccessEnabled = false
	req := newTestRequest("GET", "http://docklog.local/api/containers", nil)
	if !isClientAccessAllowed(req) {
		t.Fatal("expected CLIENT_ACCESS=off to allow direct API use")
	}
}

func TestLocalhostAllowedOutsideProduction(t *testing.T) {
	resetClientAccessState()
	os.Unsetenv("ENV")
	os.Unsetenv("GO_ENV")

	req := newTestRequest("GET", "http://localhost:8000/api/config", map[string]string{
		headerDockLogClient: clientHeaderWeb,
		"Origin":            "http://localhost:5173",
	})
	if !isClientAccessAllowed(req) {
		t.Fatal("expected localhost origin outside production to be allowed for dev")
	}
}

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const Version = "1.0.0"

var (
	serveFrontend = true
	runMode       = "server"
)

func dispatchCLI(args []string) (exit bool, exitCode int) {
	if len(args) < 2 {
		applyRunMode("server")
		return false, 0
	}

	cmd := args[1]
	switch cmd {
	case "server":
		applyRunMode("server")
		return false, 0
	case "agent":
		applyRunMode("agent")
		return false, 0
	case "agent-only":
		applyRunMode("agent-only")
		return false, 0
	case "reset-password":
		if err := runResetPasswordCLI(args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "docklog reset-password: %v\n", err)
			return true, 1
		}
		return true, 0
	case "version", "-v", "--version":
		printVersion()
		return true, 0
	case "config":
		printConfig()
		return true, 0
	case "help", "-h", "--help":
		printCLIHelp(args[2:])
		return true, 0
	default:
		if strings.HasPrefix(cmd, "-") {
			applyRunMode("server")
			return false, 0
		}
		fmt.Fprintf(os.Stderr, "docklog: unknown command %q\n\n", cmd)
		printCLIHelp(nil)
		return true, 1
	}
}

func applyRunMode(mode string) {
	runMode = mode
	switch mode {
	case "agent":
		serveFrontend = true
		setEnvIfEmpty("DOCKLOG_MODE", "agent")
	case "agent-only":
		serveFrontend = false
		setEnvIfEmpty("DOCKLOG_MODE", "agent")
		setEnvIfEmpty("DOCKLOG_AGENT_ONLY", "true")
	default:
		serveFrontend = true
		setEnvIfEmpty("DOCKLOG_MODE", "server")
	}
}

func setEnvIfEmpty(key, value string) {
	if strings.TrimSpace(os.Getenv(key)) == "" {
		_ = os.Setenv(key, value)
	}
}

func printVersion() {
	fmt.Printf("docklog %s\n", Version)
}

func printConfig() {
	boolEnv := func(key string, defaultVal bool) string {
		val := strings.TrimSpace(os.Getenv(key))
		if val == "" {
			if defaultVal {
				return "true"
			}
			return "false"
		}
		return val
	}

	fmt.Println("DockLog configuration (non-secret):")
	fmt.Printf("  version              %s\n", Version)
	fmt.Printf("  mode                 %s\n", runModeLabel())
	fmt.Printf("  port                 %s\n", envOrDefault("PORT", "8000"))
	fmt.Printf("  db_path              %s\n", envOrDefault("DB_PATH", "docklog.db"))
	fmt.Printf("  docker_host          %s\n", envOrDefault("DOCKER_HOST", "unix:///var/run/docker.sock"))
	fmt.Printf("  disable_auth         %s\n", boolEnv("DISABLE_AUTH", false))
	fmt.Printf("  client_access        %s\n", envOrDefault("CLIENT_ACCESS", "strict"))
	if excluded := strings.TrimSpace(os.Getenv("EXCLUDE_CONTAINERS")); excluded != "" {
		fmt.Printf("  exclude_containers     %s\n", excluded)
	} else {
		fmt.Println("  exclude_containers     (empty — docklog self still hidden)")
	}
	fmt.Printf("  allow_start          %s\n", boolEnv("ALLOW_START", false))
	fmt.Printf("  allow_stop           %s\n", boolEnv("ALLOW_STOP", false))
	fmt.Printf("  allow_restart        %s\n", boolEnv("ALLOW_RESTART", false))
	fmt.Printf("  allow_delete         %s\n", boolEnv("ALLOW_DELETE", false))
	allowShell := boolEnv("ALLOW_SHELL", false) == "true" || boolEnv("ALLOW_BASH", false) == "true"
	fmt.Printf("  allow_shell          %t\n", allowShell)
	if url := strings.TrimSpace(os.Getenv("CONTROL_PLANE_URL")); url != "" {
		fmt.Printf("  control_plane_url    %s\n", url)
	}
	if secret := strings.TrimSpace(os.Getenv("SECRET_KEY")); secret != "" {
		fmt.Println("  secret_key           (set)")
	} else {
		fmt.Println("  secret_key           (default — change in production)")
	}
}

func runModeLabel() string {
	if strings.TrimSpace(os.Getenv("DOCKLOG_AGENT_ONLY")) == "true" {
		return "agent-only"
	}
	if v := strings.TrimSpace(os.Getenv("DOCKLOG_MODE")); v != "" {
		return v
	}
	return runMode
}

func envOrDefault(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

func printCLIHelp(topic []string) {
	if len(topic) > 0 {
		switch topic[0] {
		case "server":
			fmt.Println(`docklog server — run the full DockLog dashboard (API, WebSockets, embedded Vue UI).

Environment variables configure auth, permissions, and ports. See README.md.`)
			return
		case "agent":
			fmt.Println(`docklog agent — run DockLog as a fleet agent with the local UI enabled.

Sets DOCKLOG_MODE=agent. Optional CONTROL_PLANE_URL registers this host with a remote DockLog control plane (future).`)
			return
		case "agent-only":
			fmt.Println(`docklog agent-only — headless agent (API + WebSockets only, no bundled web UI).

Sets DOCKLOG_MODE=agent and DOCKLOG_AGENT_ONLY=true. Use when a remote UI or mobile app connects to this host.`)
			return
		case "reset-password":
			fmt.Println(`docklog reset-password <username> <new-password>

Reset a user password in the SQLite database. Invalidates existing sessions.
Requires DB_PATH to point at the on-disk database (not :memory:).`)
			return
		}
	}

	fmt.Print(`DockLog — self-hosted Docker logs, RBAC, and monitoring.

Usage:
  docklog [command]

Commands:
  server          Run full dashboard with embedded web UI (default)
  agent           Run as fleet agent (local UI + API)
  agent-only      Run headless agent (API/WebSockets only)
  reset-password  Reset a user password in SQLite
  config          Print effective non-secret configuration
  version         Print version
  help            Show this help

Examples:
  docklog
  docklog server
  docklog agent-only
  docklog reset-password admin 'NewSecurePass1'
  docklog config

Install globally:
  make install
  # or: go install .

Docker:
  docker exec docklog docklog reset-password admin 'NewSecurePass1'
`)
}

func logRunMode() {
	switch runMode {
	case "agent":
		log.Printf("Starting DockLog in agent mode (local UI enabled)")
		if url := strings.TrimSpace(os.Getenv("CONTROL_PLANE_URL")); url == "" {
			log.Println("CONTROL_PLANE_URL is not set — running standalone agent until a control plane is configured")
		} else {
			log.Printf("Control plane: %s", url)
		}
	case "agent-only":
		log.Println("Starting DockLog in agent-only mode (API/WebSockets, no bundled UI)")
	default:
		log.Println("Starting DockLog server")
	}
}

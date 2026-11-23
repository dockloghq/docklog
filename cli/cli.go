package cli

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Version is injected dynamically at build time via -ldflags="-X 'docklog/cli.Version=...'".
// If not injected, it checks the DOCKLOG_VERSION env var, falling back to "1.1.0".
var Version = func() string {
	if env := os.Getenv("DOCKLOG_VERSION"); env != "" {
		return strings.TrimPrefix(env, "v")
	}
	return "1.1.0"
}()

// Runtime holds server mode resolved from CLI arguments.
type Runtime struct {
	ServeFrontend bool
	Mode          string
}

// Dispatch parses argv and runs one-shot CLI commands or returns runtime for the server.
func Dispatch(args []string) (exit bool, code int, rt Runtime) {
	rt = Runtime{ServeFrontend: true, Mode: "server"}

	if len(args) < 2 {
		applyRunMode(&rt, "server")
		return false, 0, rt
	}

	cmd := args[1]
	switch cmd {
	case "server":
		applyRunMode(&rt, "server")
		return false, 0, rt
	case "agent":
		applyRunMode(&rt, "agent")
		return false, 0, rt
	case "agent-only":
		applyRunMode(&rt, "agent-only")
		return false, 0, rt
	case "reset-password":
		if err := RunResetPassword(args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "docklog reset-password: %v\n", err)
			return true, 1, rt
		}
		return true, 0, rt
	case "version", "-v", "--version":
		PrintVersion()
		return true, 0, rt
	case "config":
		PrintConfig(rt)
		return true, 0, rt
	case "help", "-h", "--help":
		PrintHelp(args[2:])
		return true, 0, rt
	default:
		if strings.HasPrefix(cmd, "-") {
			applyRunMode(&rt, "server")
			return false, 0, rt
		}
		fmt.Fprintf(os.Stderr, "docklog: unknown command %q\n\n", cmd)
		PrintHelp(nil)
		return true, 1, rt
	}
}

func applyRunMode(rt *Runtime, mode string) {
	rt.Mode = mode
	switch mode {
	case "agent":
		rt.ServeFrontend = true
		setEnvIfEmpty("DOCKLOG_MODE", "agent")
	case "agent-only":
		rt.ServeFrontend = false
		setEnvIfEmpty("DOCKLOG_MODE", "agent")
		setEnvIfEmpty("DOCKLOG_AGENT_ONLY", "true")
	default:
		rt.ServeFrontend = true
		setEnvIfEmpty("DOCKLOG_MODE", "server")
	}
}

func setEnvIfEmpty(key, value string) {
	if strings.TrimSpace(os.Getenv(key)) == "" {
		_ = os.Setenv(key, value)
	}
}

func PrintVersion() {
	fmt.Printf("docklog %s\n", Version)
}

// LogRunMode prints the resolved startup mode to the server log.
func LogRunMode(rt Runtime) {
	switch rt.Mode {
	case "agent":
		log.Printf("Starting DockLog in agent mode (local UI enabled)")
		if url := strings.TrimSpace(os.Getenv("CONTROL_PLANE_URL")); url == "" {
			log.Println("CONTROL_PLANE_URL is not set; running standalone agent until a control plane is configured")
		} else {
			log.Printf("Control plane: %s", url)
		}
	case "agent-only":
		log.Println("Starting DockLog in agent-only mode (API/WebSockets, no bundled UI)")
	default:
		log.Println("Starting DockLog server")
	}
}

func runModeLabel(rt Runtime) string {
	if strings.TrimSpace(os.Getenv("DOCKLOG_AGENT_ONLY")) == "true" {
		return "agent-only"
	}
	if v := strings.TrimSpace(os.Getenv("DOCKLOG_MODE")); v != "" {
		return v
	}
	return rt.Mode
}

func envOrDefault(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

func boolEnv(key string, defaultVal bool) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		if defaultVal {
			return "true"
		}
		return "false"
	}
	return val
}

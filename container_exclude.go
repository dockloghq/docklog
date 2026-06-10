package main

import (
	"os"
	"strings"
)

var excludedContainerNames []string

func initExcludedContainers() {
	raw := strings.TrimSpace(os.Getenv("EXCLUDE_CONTAINERS"))
	if raw == "" {
		excludedContainerNames = nil
		return
	}

	parts := strings.Split(raw, ",")
	excludedContainerNames = make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		excludedContainerNames = append(excludedContainerNames, name)
	}
}

func isDockLogSelfContainer(name, image string) bool {
	name = strings.TrimPrefix(strings.TrimSpace(name), "/")
	if name != "" && strings.EqualFold(name, "docklog") {
		return true
	}
	return strings.Contains(strings.ToLower(image), "docklog")
}

func isExcludedContainer(name, image string) bool {
	if isDockLogSelfContainer(name, image) {
		return true
	}

	name = strings.TrimPrefix(strings.TrimSpace(name), "/")
	for _, excluded := range excludedContainerNames {
		if excluded == "" {
			continue
		}
		if strings.EqualFold(name, excluded) {
			return true
		}
	}
	return false
}

func sanitizeContainerEnv(env []string) []string {
	sanitized := make([]string, len(env))
	for i, item := range env {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			sanitized[i] = item
			continue
		}
		k := strings.ToLower(parts[0])
		if strings.Contains(k, "pass") ||
			strings.Contains(k, "secret") ||
			strings.Contains(k, "key") ||
			strings.Contains(k, "token") ||
			strings.Contains(k, "auth") ||
			strings.Contains(k, "pwd") ||
			strings.Contains(k, "db_") {
			sanitized[i] = parts[0] + "=••••••••••••"
			continue
		}
		sanitized[i] = item
	}
	return sanitized
}

func containerNameImageFromInspect(name string, configImage string) (string, string) {
	return strings.TrimPrefix(strings.TrimSpace(name), "/"), configImage
}

func inspectContainerExcluded(name, configImage string) bool {
	containerName, containerImage := containerNameImageFromInspect(name, configImage)
	return isExcludedContainer(containerName, containerImage)
}

package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"docklog/config"
	"docklog/models"
)

const (
	dockerHubTagsURL        = "https://hub.docker.com/v2/namespaces/aimldev/repositories/docklog/tags?page_size=100"
	versionCheckInterval    = 12 * time.Hour
	versionCheckHTTPTimeout = 10 * time.Second
)

var (
	semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

	latestVersionCache   string
	hasUpdateCache       bool
	versionCacheMu       sync.RWMutex
	lastCheckedTime      time.Time
)

// CheckVersionStatus checks Docker Hub and returns version information
func CheckVersionStatus(currentVersion string) (current string, latest string, updateAvailable bool, err error) {
	current = currentVersion
	versionCacheMu.RLock()
	if time.Since(lastCheckedTime) < 15*time.Minute && latestVersionCache != "" {
		latest = latestVersionCache
		updateAvailable = hasUpdateCache
		versionCacheMu.RUnlock()
		return
	}
	versionCacheMu.RUnlock()

	parsedCurrent, hasCurrent := ParseSemver(currentVersion)
	semverLatest, found, _, fetchErr := fetchLatestDockerHubSemver()
	if fetchErr != nil {
		return currentVersion, currentVersion, false, fetchErr
	}

	latest = semverLatest.Raw
	if found && hasCurrent {
		updateAvailable = IsSemverNewer(semverLatest, parsedCurrent)
	}

	versionCacheMu.Lock()
	latestVersionCache = latest
	hasUpdateCache = updateAvailable
	lastCheckedTime = time.Now()
	versionCacheMu.Unlock()

	return current, latest, updateAvailable, nil
}

type dockerHubTagsResponse struct {
	Results []struct {
		Name        string `json:"name"`
		LastUpdated string `json:"last_updated"`
		Images      []struct {
			Digest string `json:"digest"`
		} `json:"images"`
	} `json:"results"`
}

type Semver struct {
	Major int
	Minor int
	Patch int
	Raw   string
}

func StartVersionUpdateMonitor(currentVersion string, ns *NotificationService) {
	if ns == nil {
		return
	}
	go RunVersionUpdateMonitor(currentVersion, ns)
}

func RunVersionUpdateMonitor(currentVersion string, ns *NotificationService) {
	current, ok := ParseSemver(currentVersion)
	if !ok {
		config.Debugf("Version check skipped: current version %q is not semver", currentVersion)
		return
	}

	var notified string
	var latestFingerprintSeen string
	firstCheck := true
	check := func() {
		latest, found, latestFingerprint, err := fetchLatestDockerHubSemver()
		if err != nil {
			config.Debugf("Version check failed: %v", err)
			return
		}
		if latestFingerprint != "" {
			if firstCheck {
				latestFingerprintSeen = latestFingerprint
			} else if latestFingerprintSeen != "" && latestFingerprintSeen != latestFingerprint {
				ns.DispatchAuditEvent(models.AuditNotificationEvent{
					UserID:   0,
					Username: "docklog",
					Action:   "version_update",
					Resource: "docklog_image",
					Status:   "Info",
					Message:  "Docker tag aimldev/docklog:latest was updated. Run: docker pull aimldev/docklog:latest",
				})
				latestFingerprintSeen = latestFingerprint
			} else if latestFingerprintSeen == "" {
				latestFingerprintSeen = latestFingerprint
			}
		}
		firstCheck = false

		if !found {
			config.Debugf("Version check: no semver tags found")
			return
		}
		if !IsSemverNewer(latest, current) || latest.Raw == notified {
			return
		}

		notified = latest.Raw
		ns.DispatchAuditEvent(models.AuditNotificationEvent{
			UserID:   0,
			Username: "docklog",
			Action:   "version_update",
			Resource: "docklog_image",
			Status:   "Info",
			Message: fmt.Sprintf(
				"New DockLog version available: %s (current: %s). Run: docker pull aimldev/docklog:latest",
				latest.Raw,
				strings.TrimPrefix(current.Raw, "v"),
			),
		})
	}

	check()
	ticker := time.NewTicker(versionCheckInterval)
	defer ticker.Stop()
	for range ticker.C {
		check()
	}
}

func fetchLatestDockerHubSemver() (Semver, bool, string, error) {
	client := &http.Client{Timeout: versionCheckHTTPTimeout}
	resp, err := client.Get(dockerHubTagsURL)
	if err != nil {
		return Semver{}, false, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Semver{}, false, "", fmt.Errorf("docker hub returned status %d", resp.StatusCode)
	}
	var payload dockerHubTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Semver{}, false, "", err
	}

	latestFingerprint := ""
	versions := make([]Semver, 0, len(payload.Results))
	for _, item := range payload.Results {
		if item.Name == "latest" {
			latestFingerprint = dockerTagFingerprint(item.LastUpdated, item.Images)
		}
		if v, ok := ParseSemver(item.Name); ok {
			versions = append(versions, v)
		}
	}
	if len(versions) == 0 {
		return Semver{}, false, latestFingerprint, nil
	}
	sort.Slice(versions, func(i, j int) bool {
		return IsSemverNewer(versions[i], versions[j])
	})
	return versions[0], true, latestFingerprint, nil
}

func dockerTagFingerprint(lastUpdated string, images []struct{ Digest string `json:"digest"` }) string {
	parts := make([]string, 0, len(images)+1)
	if ts := strings.TrimSpace(lastUpdated); ts != "" {
		parts = append(parts, ts)
	}
	for _, image := range images {
		if digest := strings.TrimSpace(image.Digest); digest != "" {
			parts = append(parts, digest)
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

func ParseSemver(raw string) (Semver, bool) {
	raw = strings.TrimSpace(raw)
	match := semverRe.FindStringSubmatch(raw)
	if len(match) != 4 {
		return Semver{}, false
	}
	major, err1 := strconv.Atoi(match[1])
	minor, err2 := strconv.Atoi(match[2])
	patch, err3 := strconv.Atoi(match[3])
	if err1 != nil || err2 != nil || err3 != nil {
		return Semver{}, false
	}
	return Semver{Major: major, Minor: minor, Patch: patch, Raw: raw}, true
}

func IsSemverNewer(a, b Semver) bool {
	if a.Major != b.Major {
		return a.Major > b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor > b.Minor
	}
	return a.Patch > b.Patch
}

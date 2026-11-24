package services

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"docklog/containers"
	"docklog/dockerutil"
	"docklog/models"

	"github.com/moby/moby/client"
)

type logTailState struct {
	cancel context.CancelFunc
}

// StartLogAlertMonitor tails container logs for enabled log alert rules.
func StartLogAlertMonitor(cli *client.Client, engine *AlertEngine) {
	if cli == nil || engine == nil {
		return
	}
	go runLogAlertMonitor(cli, engine)
}

func runLogAlertMonitor(cli *client.Client, engine *AlertEngine) {
	active := map[string]*logTailState{}
	var mu sync.Mutex
	tracker := newOccurrenceTracker()

	refresh := func() {
		rules := engine.rulesBySource(models.AlertSourceLogs)
		if len(rules) == 0 {
			mu.Lock()
			for id, state := range active {
				state.cancel()
				delete(active, id)
			}
			mu.Unlock()
			return
		}

		ctx := context.Background()
		res, err := cli.ContainerList(ctx, client.ContainerListOptions{All: false})
		if err != nil {
			return
		}

		needed := map[string]struct{}{}
		for _, summary := range dockerutil.ExtractContainers(res) {
			id := containerSummaryID(summary)
			name := containerSummaryName(summary)
			image, _ := summary["Image"].(string)
			if id == "" || name == "" || containers.IsExcludedContainer(name, image) {
				continue
			}
			labels := containerSummaryLabels(summary)
			for _, rule := range rules {
				scope, _ := parseAlertScope(rule.ScopeJSON)
				if containerMatchesScope(name, image, labels, scope) {
					needed[id] = struct{}{}
					break
				}
			}
		}

		mu.Lock()
		for id, state := range active {
			if _, ok := needed[id]; !ok {
				state.cancel()
				delete(active, id)
			}
		}
		for id := range needed {
			if _, ok := active[id]; ok {
				continue
			}
			tailCtx, cancel := context.WithCancel(context.Background())
			active[id] = &logTailState{cancel: cancel}
			go tailLogsForAlerts(tailCtx, cli, engine, tracker, id)
		}
		mu.Unlock()
	}

	refresh()
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		_ = engine.Reload()
		refresh()
	}
}

func tailLogsForAlerts(ctx context.Context, cli *client.Client, engine *AlertEngine, tracker *occurrenceTracker, containerID string) {
	inspect, err := cli.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		return
	}
	name := strings.TrimPrefix(inspect.Container.Name, "/")
	image := ""
	if inspect.Container.Config != nil {
		image = inspect.Container.Config.Image
	}
	labels := map[string]string{}
	if inspect.Container.Config != nil && inspect.Container.Config.Labels != nil {
		labels = inspect.Container.Config.Labels
	}

	reader, err := cli.ContainerLogs(ctx, containerID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "50",
		Timestamps: false,
	})
	if err != nil {
		return
	}
	defer reader.Close()

	isTty := inspect.Container.Config != nil && inspect.Container.Config.Tty
	_ = dockerutil.StreamLogs(reader, isTty, func(payload []byte) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		processLogLine(engine, tracker, containerID, name, image, labels, string(payload))
		return nil
	})
}

func processLogLine(engine *AlertEngine, tracker *occurrenceTracker, containerID, name, image string, labels map[string]string, line string) {
	rules := engine.rulesBySource(models.AlertSourceLogs)
	now := time.Now()
	for _, rule := range rules {
		scope, err := parseAlertScope(rule.ScopeJSON)
		if err != nil || !containerMatchesScope(name, image, labels, scope) {
			continue
		}
		cfg, err := parseLogAlertConfig(rule.ConfigJSON)
		if err != nil || !logLineMatches(cfg, line) {
			continue
		}
		count := tracker.add(rule.RuleKey, name, time.Duration(cfg.WindowSeconds)*time.Second, now)
		if count < cfg.MatchCount {
			continue
		}
		tracker.reset(rule.RuleKey, name)
		engine.Emit(rule, models.AlertNotification{
			RuleID: rule.RuleKey, Severity: rule.Severity, Container: name,
			Source: models.AlertSourceLogs,
			Message: fmtLogAlertMessage(rule.Name, count, cfg.WindowSeconds),
			Metadata: map[string]interface{}{"matches": count, "sample": truncate(line, 200)},
			URL: "/containers/" + containerID + "/logs",
		})
	}
}

func fmtLogAlertMessage(ruleName string, count, windowSeconds int) string {
	return strings.TrimSpace(ruleName) + ": " + strconv.Itoa(count) + " matching log lines in " + strconv.Itoa(windowSeconds) + " seconds"
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}

func containerSummaryID(summary map[string]interface{}) string {
	if id, ok := summary["ID"].(string); ok && id != "" {
		return id
	}
	if id, ok := summary["Id"].(string); ok {
		return id
	}
	return ""
}

func containerSummaryName(summary map[string]interface{}) string {
	names, _ := summary["Names"].([]interface{})
	if len(names) == 0 {
		return "unknown"
	}
	name, _ := names[0].(string)
	return strings.TrimPrefix(name, "/")
}

func containerSummaryLabels(summary map[string]interface{}) map[string]string {
	labels, _ := summary["Labels"].(map[string]interface{})
	if len(labels) == 0 {
		return nil
	}
	out := make(map[string]string, len(labels))
	for key, value := range labels {
		out[key], _ = value.(string)
	}
	return out
}


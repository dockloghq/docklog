package containers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/moby/moby/client"
)

// ContainerSpecs represents the cached static configuration of a container
type ContainerSpecs struct {
	CPULimit     float64
	MemLimit     int64
	IsSwarm      bool
	SwarmService string
	PortList     []string
	FetchedAt    time.Time
}

var (
	specsCache = make(map[string]ContainerSpecs)
	specsMu    sync.RWMutex
	cacheTTL   = 60 * time.Second
)

// GetOrInspectSpecs retrieves container specs from memory cache, or inspects Docker if missing/expired
func GetOrInspectSpecs(ctx context.Context, cli *client.Client, id string) ContainerSpecs {
	specsMu.RLock()
	cached, found := specsCache[id]
	specsMu.RUnlock()

	if found && time.Since(cached.FetchedAt) < cacheTTL {
		return cached
	}

	inspect, err := cli.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
	if err != nil {
		if found {
			return cached
		}
		return ContainerSpecs{}
	}

	specs := ContainerSpecs{
		FetchedAt: time.Now(),
	}

	if inspect.Container.Config != nil && inspect.Container.Config.Labels != nil {
		if svc, ok := inspect.Container.Config.Labels["com.docker.swarm.service.name"]; ok && svc != "" {
			specs.IsSwarm = true
			specs.SwarmService = svc
		}
	}

	if inspect.Container.HostConfig != nil {
		if inspect.Container.HostConfig.NanoCPUs > 0 {
			specs.CPULimit = float64(inspect.Container.HostConfig.NanoCPUs) / 1e9
		}
		specs.MemLimit = inspect.Container.HostConfig.Memory
	}

	if inspect.Container.NetworkSettings != nil && inspect.Container.NetworkSettings.Ports != nil {
		for containerPort, hostBindings := range inspect.Container.NetworkSettings.Ports {
			cpStr := fmt.Sprintf("%v", containerPort)
			if len(hostBindings) == 0 {
				specs.PortList = append(specs.PortList, cpStr)
			} else {
				for _, b := range hostBindings {
					specs.PortList = append(specs.PortList, fmt.Sprintf("%s→%s", b.HostPort, cpStr))
				}
			}
		}
	}

	specsMu.Lock()
	specsCache[id] = specs
	specsMu.Unlock()

	return specs
}

// InvalidateSpecsCache clears cached specs for a specific container
func InvalidateSpecsCache(id string) {
	specsMu.Lock()
	delete(specsCache, id)
	specsMu.Unlock()
}


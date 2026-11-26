<template>
  <div class="fleet-dashboard animate-fade-in">
    <!-- Top Header: Fleet Overview + Runtime Status Pills -->
    <header class="fleet-header">
      <div class="fleet-title-group">
        <h1 class="fleet-title">Fleet Overview</h1>
        <p class="fleet-subtitle">All container platforms in one place.</p>
      </div>

      <div class="fleet-runtimes">
        <!-- Docker Pill -->
        <div class="runtime-pill" :class="{ 'is-active': dockerActive }">
          <span class="runtime-pill-icon docker-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M4 14c.5-2.5 2.5-4 5.5-4 1.5 0 3 .5 4 1.5 1-1 2.5-1.5 4-1.5 3 0 5 1.5 5.5 4-2 0-3.5 1-4.5 2.5-.5-1-1.5-1.5-2.5-1.5s-2 .5-2.5 1.5c-1-1.5-2.5-2.5-4.5-2.5-1.5 0-3 .5-4 1.5-.5-.5-1-.8-1.5-1z" opacity="0.3"/>
              <path d="M22 13c-.3-1.8-1.6-3.2-3.3-3.6-.4-.1-.8-.1-1.2 0-.3-.9-.9-1.7-1.7-2.2-.4-.3-.9-.5-1.4-.5-.3 0-.6.1-.9.2C13 5.4 11.2 4.5 9.2 4.5c-2.3 0-4.3 1.2-5.4 3.1-.3.5-.5 1.1-.6 1.7C1.4 10 0 11.8 0 14c0 3.9 3.6 7 8 7 3.5 0 6.6-2 7.7-5 .5.1 1 .2 1.5.2 2.6 0 4.8-1.4 5.8-3.2zM8.5 7.5h2v2h-2zm-3 3h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2zm-9 3h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2z"/>
            </svg>
          </span>
          <span class="runtime-pill-label">Docker</span>
          <span class="runtime-pill-count">{{ standaloneCount }}</span>
          <span class="runtime-status-dot" :class="{ live: dockerActive }"></span>
        </div>

        <!-- Swarm Pill -->
        <div v-if="isSwarmActive" class="runtime-pill" :class="{ 'is-active': isSwarmActive }">
          <span class="runtime-pill-icon swarm-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="12 2 2 7 12 12 22 7 12 2" />
              <polyline points="2 17 12 22 22 17" />
              <polyline points="2 12 12 17 22 12" />
            </svg>
          </span>
          <span class="runtime-pill-label">Swarm</span>
          <span class="runtime-pill-count">{{ isSwarmActive ? swarmServicesCount : 0 }}</span>
          <span class="runtime-status-dot" :class="{ live: isSwarmActive }"></span>
        </div>

        <!-- Kubernetes Pill -->
        <div v-if="k8sActive" class="runtime-pill" :class="{ 'is-active': k8sActive }">
          <span class="runtime-pill-icon k8s-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M12 2l8.5 4.9v9.8L12 21.6 3.5 16.7V6.9L12 2zm0 2.3L5.5 8v7.9l6.5 3.8 6.5-3.8V8L12 4.3zm0 3.2a4.5 4.5 0 1 1 0 9 4.5 4.5 0 0 1 0-9zm0 2a2.5 2.5 0 1 0 0 5 2.5 2.5 0 0 0 0-5z" />
            </svg>
          </span>
          <span class="runtime-pill-label">Kubernetes</span>
          <span class="runtime-pill-count">{{ k8sActive ? podTotal : 0 }}</span>
          <span class="runtime-status-dot" :class="{ live: k8sActive }"></span>
        </div>
      </div>
    </header>

    <!-- Cluster & Resource Stats Banner -->
    <section class="fleet-stats-banner">
      <!-- Left side: 4 Counts with Icons -->
      <div class="banner-counts">
        <div class="stat-box" @click="activeTab = 'containers'">
          <div class="stat-box-icon">
            <!-- Box Icon -->
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
              <polyline points="3.27 6.96 12 12.01 20.73 6.96" />
              <line x1="12" y1="22.08" x2="12" y2="12" />
            </svg>
          </div>
          <div class="stat-box-info">
            <span class="stat-box-title">Containers</span>
            <span class="stat-box-number">{{ totalContainersCount }}</span>
          </div>
        </div>

        <!-- Services (Swarm / K8s) -->
        <div v-if="isSwarmActive || k8sActive" class="stat-box" @click="activeTab = 'services'">
          <div class="stat-box-icon">
            <!-- Layers Icon -->
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="12 2 2 7 12 12 22 7 12 2" />
              <polyline points="2 17 12 22 22 17" />
              <polyline points="2 12 12 17 22 12" />
            </svg>
          </div>
          <div class="stat-box-info">
            <span class="stat-box-title">Services</span>
            <span class="stat-box-number">{{ totalServicesCount }}</span>
          </div>
        </div>

        <!-- Pods (Kubernetes) -->
        <div v-if="k8sActive" class="stat-box" @click="activeTab = 'pods'">
          <div class="stat-box-icon">
            <!-- Hexagon / Pods Icon -->
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
            </svg>
          </div>
          <div class="stat-box-info">
            <span class="stat-box-title">Pods</span>
            <span class="stat-box-number">{{ podTotal }}</span>
          </div>
        </div>

        <!-- Nodes (Swarm / Cluster) -->
        <div v-if="isSwarmActive || k8sActive" class="stat-box" @click="activeTab = 'nodes'">
          <div class="stat-box-icon">
            <!-- Server / Nodes Icon -->
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="3" width="20" height="7" rx="2" ry="2" />
              <rect x="2" y="14" width="20" height="7" rx="2" ry="2" />
              <line x1="6" y1="6.5" x2="6.01" y2="6.5" />
              <line x1="6" y1="17.5" x2="6.01" y2="17.5" />
            </svg>
          </div>
          <div class="stat-box-info">
            <span class="stat-box-title">Nodes</span>
            <span class="stat-box-number">{{ totalNodesCount }}</span>
          </div>
        </div>
      </div>

      <!-- Divider -->
      <div class="banner-divider"></div>

      <!-- Right side: Resource Meters -->
      <div class="banner-gauges">
        <div class="resource-meter">
          <div class="meter-meta">
            <span class="meter-label">CPU</span>
            <span class="meter-value">{{ cpuDisplay }}%</span>
          </div>
          <div class="meter-track">
            <div class="meter-fill" :style="{ width: Math.min(100, cpuVal) + '%' }"></div>
          </div>
        </div>

        <div class="resource-meter">
          <div class="meter-meta">
            <span class="meter-label">Memory</span>
            <span class="meter-value">{{ memPercentDisplay }}%</span>
          </div>
          <div class="meter-track">
            <div class="meter-fill" :style="{ width: Math.min(100, memPercentVal) + '%' }"></div>
          </div>
        </div>

        <div class="resource-meter">
          <div class="meter-meta">
            <span class="meter-label">Disk</span>
            <span class="meter-value">{{ diskPercentDisplay }}%</span>
          </div>
          <div class="meter-track">
            <div class="meter-fill" :style="{ width: Math.min(100, diskVal) + '%' }"></div>
          </div>
        </div>
      </div>
    </section>

    <!-- Workload Controls: Tabs + Platform Filter + Search -->
    <div class="fleet-controls">
      <div class="category-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="category-tab-btn"
          :class="{ active: activeTab === tab.id }"
          @click="activeTab = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>

      <div class="filter-actions">
        <!-- Platform dropdown -->
        <div v-if="isSwarmActive || k8sActive" class="platform-select-wrapper">
          <select v-model="selectedPlatform" class="platform-select">
            <option value="all">All Platforms</option>
            <option value="docker">Docker</option>
            <option v-if="isSwarmActive" value="swarm">Swarm</option>
            <option v-if="k8sActive" value="kubernetes">Kubernetes</option>
          </select>
          <span class="select-chevron">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5">
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </span>
        </div>

        <!-- Search input -->
        <div class="fleet-search">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            type="text"
            v-model="searchQuery"
            placeholder="Search..."
          />
        </div>
      </div>
    </div>

    <!-- Unified Workload Table -->
    <div class="fleet-table-card">
      <div class="table-responsive">
        <table class="fleet-table">
          <thead>
            <tr>
              <th class="col-name">NAME</th>
              <th class="col-platform">PLATFORM</th>
              <th class="col-type">TYPE</th>
              <th class="col-status">STATUS</th>
              <th class="col-age">AGE</th>
              <th class="col-actions text-right">ACTIONS</th>
            </tr>
          </thead>
          <tbody v-if="loading && tableItems.length === 0">
            <tr>
              <td colspan="6" class="table-empty">
                <div class="shimmer-placeholder">Loading workloads...</div>
              </td>
            </tr>
          </tbody>
          <tbody v-else-if="filteredItems.length === 0">
            <tr>
              <td colspan="6" class="table-empty">
                No {{ activeTab }} found matching the current filters.
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr v-for="item in filteredItems" :key="item.uniqueKey" class="fleet-row">
              <!-- Name with green wireframe box icon -->
              <td class="col-name">
                <div
                  class="item-name-cell"
                  @click="handleItemNavigate(item)"
                  @mouseenter="handleItemMouseEnter(item, $event)"
                  @mouseleave="handleItemMouseLeave"
                >
                  <span class="item-cube-icon">
                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="#22c55e" stroke-width="2">
                      <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
                      <polyline points="3.27 6.96 12 12.01 20.73 6.96" />
                      <line x1="12" y1="22.08" x2="12" y2="12" />
                    </svg>
                  </span>
                  <span class="item-name-text">{{ item.name }}</span>
                </div>
              </td>

              <!-- Platform with logo + name -->
              <td class="col-platform">
                <div class="platform-tag" :class="item.platform.toLowerCase()">
                  <span v-if="item.platform === 'Docker'" class="p-icon docker">
                    <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
                      <path d="M22 13c-.3-1.8-1.6-3.2-3.3-3.6-.4-.1-.8-.1-1.2 0-.3-.9-.9-1.7-1.7-2.2-.4-.3-.9-.5-1.4-.5-.3 0-.6.1-.9.2C13 5.4 11.2 4.5 9.2 4.5c-2.3 0-4.3 1.2-5.4 3.1-.3.5-.5 1.1-.6 1.7C1.4 10 0 11.8 0 14c0 3.9 3.6 7 8 7 3.5 0 6.6-2 7.7-5 .5.1 1 .2 1.5.2 2.6 0 4.8-1.4 5.8-3.2zM8.5 7.5h2v2h-2zm-3 3h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2zm-9 3h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2zm3 0h2v2h-2z"/>
                    </svg>
                  </span>
                  <span v-else-if="item.platform === 'Swarm'" class="p-icon swarm">
                    <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2">
                      <polygon points="12 2 2 7 12 12 22 7 12 2" />
                      <polyline points="2 17 12 22 22 17" />
                      <polyline points="2 12 12 17 22 12" />
                    </svg>
                  </span>
                  <span v-else class="p-icon kubernetes">
                    <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
                      <path d="M12 2l8.5 4.9v9.8L12 21.6 3.5 16.7V6.9L12 2zm0 2.3L5.5 8v7.9l6.5 3.8 6.5-3.8V8L12 4.3zm0 3.2a4.5 4.5 0 1 1 0 9 4.5 4.5 0 0 1 0-9zm0 2a2.5 2.5 0 1 0 0 5 2.5 2.5 0 0 0 0-5z" />
                    </svg>
                  </span>
                  <span>{{ item.platform }}</span>
                </div>
              </td>

              <!-- Workload Type -->
              <td class="col-type">
                <span class="type-text">{{ item.type }}</span>
              </td>

              <!-- Status with green capsule badge & dot -->
              <td class="col-status">
                <span class="status-badge" :class="item.statusClass">
                  <span class="status-dot"></span>
                  {{ item.statusText }}
                </span>
              </td>

              <!-- Age -->
              <td class="col-age">
                <span class="age-text">{{ item.age }}</span>
              </td>

              <!-- Actions (logs, external jump) -->
              <td class="col-actions text-right">
                <div class="row-actions">
                  <button
                    v-if="item.logsRoute"
                    class="action-btn-icon"
                    title="View logs"
                    @click="handleItemLogs(item)"
                  >
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                      <path d="M8 9h8M8 13h5"></path>
                    </svg>
                  </button>

                  <button
                    class="action-btn-icon"
                    title="View details"
                    @click="handleItemNavigate(item)"
                  >
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                      <polyline points="15 3 21 3 21 9" />
                      <line x1="10" y1="14" x2="21" y2="3" />
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Interactive Metrics Hover Card -->
    <Teleport to="body">
      <ContainerHoverCard
        :visible="isHoverCardActive"
        :container="hoveredContainer"
        :stats="liveStats"
        :position="hoverPosition"
        metrics-only
        @card-mouseenter="handleCardMouseEnter"
        @card-mouseleave="handleCardMouseLeave"
      />
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import ContainerHoverCard from '../../components/ContainerHoverCard.vue';
import { createAuthenticatedWebSocket } from '../../utils/wsAuth';
import {
  sharedState,
  dockerEnabled,
  swarmEnabled,
  kubernetesEnabled,
} from '../../utils/sharedState';
import { useContainers, fetchContainers } from '../../composables/useContainers';
import {
  swarmOverview,
  swarmServices,
  swarmNodes,
  swarmTasks,
  fetchSwarmData,
} from '../../composables/useSwarm';
import {
  pods,
  deployments,
  services as k8sServicesList,
  fetchKubernetesData,
  fetchNamespaces,
} from '../../composables/useKubernetes';

const router = useRouter();

// Active category tab: 'containers' | 'services' | 'pods' | 'nodes'
const activeTab = ref('containers');
const tabs = computed(() => {
  const list = [
    { id: 'containers', label: 'Containers' },
  ];
  if (isSwarmActive.value || k8sActive.value) {
    list.push({ id: 'services', label: 'Services' });
  }
  if (k8sActive.value) {
    list.push({ id: 'pods', label: 'Pods' });
  }
  if (isSwarmActive.value || k8sActive.value) {
    list.push({ id: 'nodes', label: 'Nodes' });
  }
  return list;
});

const selectedPlatform = ref('all');
// Search query synced with sharedState.searchQuery
const searchQuery = computed({
  get: () => sharedState.searchQuery || '',
  set: (val) => {
    sharedState.searchQuery = val;
  },
});
const openMenuId = ref(null);

// Pull container list
const { containers, loading } = useContainers();

// Runtimes active status
const dockerActive = computed(() => dockerEnabled());
const isSwarmActive = computed(() => swarmEnabled());
const k8sActive = computed(() => kubernetesEnabled() && sharedState.k8sAvailable);

const standaloneCount = computed(
  () => (Array.isArray(containers.value) ? containers.value.filter((c) => !c.is_swarm).length : 0)
);
const swarmServicesCount = computed(
  () => swarmOverview.value?.services_count ?? (Array.isArray(swarmServices.value) ? swarmServices.value.length : 0)
);
const podTotal = computed(() => (Array.isArray(pods.value) ? pods.value.length : 0));

// Stats counts for the top banner
const totalContainersCount = computed(() => {
  // Docker standalone + Swarm tasks + K8s pods
  const dockerCnt = standaloneCount.value;
  const swarmCnt = isSwarmActive.value
    ? (swarmOverview.value?.running_tasks_count ?? (Array.isArray(swarmTasks.value) ? swarmTasks.value.length : 0))
    : 0;
  return dockerCnt + swarmCnt;
});

const totalServicesCount = computed(() => {
  const swarmSvc = isSwarmActive.value && Array.isArray(swarmServices.value) ? swarmServices.value.length : 0;
  const k8sSvc = k8sActive.value && Array.isArray(k8sServicesList.value) ? k8sServicesList.value.length : 0;
  return swarmSvc + k8sSvc;
});

const totalNodesCount = computed(() => {
  if (isSwarmActive.value && swarmOverview.value?.nodes_count !== undefined) {
    return swarmOverview.value.nodes_count;
  }
  if (Array.isArray(swarmNodes.value) && swarmNodes.value.length) return swarmNodes.value.length;
  return 1; // At minimum 1 host node
});

// Resource meters
const cpuVal = computed(() => parseFloat(sharedState.systemStats?.cpu || 0));
const cpuDisplay = computed(() => {
  return Math.round(cpuVal.value);
});

const memUsed = computed(() => sharedState.systemStats?.memory || 0);
const memTotal = computed(() => sharedState.systemStats?.total_memory || 1);
const memPercentVal = computed(() => {
  if (!memTotal.value || memTotal.value <= 1) return 0;
  return Math.round((memUsed.value / memTotal.value) * 100);
});
const memPercentDisplay = computed(() => {
  return memPercentVal.value;
});

// Disk usage from real system stats
const diskVal = computed(() => {
  return parseFloat(sharedState.systemStats?.disk_percent || 0);
});
const diskPercentDisplay = computed(() => Math.round(diskVal.value));

// Helper to format Age from timestamp (accepts unix seconds, ISO strings, or Date)
function formatAge(val) {
  if (!val) return '—';
  let unixSeconds = 0;
  if (typeof val === 'number') {
    // If it looks like milliseconds (> 1e11), convert to seconds
    unixSeconds = val > 1e11 ? Math.floor(val / 1000) : Math.floor(val);
  } else if (typeof val === 'string') {
    const parsed = new Date(val).getTime();
    if (!isNaN(parsed) && parsed > 0) {
      unixSeconds = Math.floor(parsed / 1000);
    }
  }
  if (!unixSeconds || unixSeconds <= 0) return '—';
  const now = Math.floor(Date.now() / 1000);
  const diff = Math.max(0, now - unixSeconds);
  const days = Math.floor(diff / 86400);
  const hours = Math.floor((diff % 86400) / 3600);
  const minutes = Math.floor((diff % 3600) / 60);

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

// Unified items generator based on activeTab
const tableItems = computed(() => {
  const items = [];

  if (activeTab.value === 'containers') {
    // 1. Docker Standalone Containers
    if (Array.isArray(containers.value)) {
      containers.value.forEach((c) => {
        const isRunning = c.state === 'running';
        const isSwarm = c.is_swarm;
        items.push({
          uniqueKey: `docker-${c.id}`,
          id: c.id,
          name: (c.name || '').replace(/^\//, ''),
          platform: isSwarm ? 'Swarm' : 'Docker',
          type: 'Container',
          statusText: isRunning ? 'Running' : (c.state || 'Stopped'),
          statusClass: isRunning ? 'running' : 'stopped',
          ready: isRunning ? '1 / 1' : '0 / 1',
          age: formatAge(c.created),
          raw: c,
          targetRoute: `/containers/${c.id}`,
          logsRoute: isSwarm ? `/logs?c=${c.id}&rt=swarm` : `/logs?c=${c.id}&rt=docker`,
        });
      });
    }

    // 2. K8s Pods as containers if K8s active
    if (k8sActive.value && Array.isArray(pods.value)) {
      pods.value.forEach((p) => {
        const isRunning = p.phase === 'Running';
        const readyDisplay = p.ready ? p.ready.replace('/', ' / ') : (isRunning ? '1 / 1' : '0 / 1');
        items.push({
          uniqueKey: `k8s-pod-${p.namespace}-${p.name}`,
          id: p.name,
          name: p.name,
          platform: 'Kubernetes',
          type: 'Pod',
          statusText: p.phase || 'Running',
          statusClass: isRunning ? 'running' : 'pending',
          ready: readyDisplay,
          age: formatAge(p.created),
          raw: p,
          targetRoute: `/kubernetes/pods/${p.namespace}/${p.name}`,
          logsRoute: `/logs?p=${encodeURIComponent(p.namespace + '/' + p.name)}&rt=kubernetes`,
        });
      });
    }
  } else if (activeTab.value === 'services') {
    // Swarm Services
    if (Array.isArray(swarmServices.value)) {
      swarmServices.value.forEach((s) => {
        const running = s.running ?? 0;
        const replicas = s.replicas ?? 1;
        const isOk = running > 0 && running >= replicas;
        items.push({
          uniqueKey: `swarm-svc-${s.id}`,
          id: s.id,
          name: s.name,
          platform: 'Swarm',
          type: 'Service',
          statusText: isOk ? 'Running' : 'Degraded',
          statusClass: isOk ? 'running' : 'warning',
          ready: `${running} / ${replicas}`,
          age: formatAge(s.created_at),
          raw: s,
          targetRoute: `/swarm/services/${s.id}`,
          logsRoute: `/logs?rt=swarm&s=${s.name}`,
        });
      });
    }

    // K8s Deployments & Services
    if (Array.isArray(deployments.value)) {
      deployments.value.forEach((d) => {
        const ready = d.ready_replicas ?? d.replicas ?? 1;
        const total = d.replicas ?? 1;
        items.push({
          uniqueKey: `k8s-deploy-${d.namespace}-${d.name}`,
          id: d.name,
          name: d.name,
          platform: 'Kubernetes',
          type: 'Deployment',
          statusText: ready > 0 ? 'Running' : 'Pending',
          statusClass: ready > 0 ? 'running' : 'pending',
          ready: `${ready} / ${total}`,
          age: formatAge(d.created),
          raw: d,
          targetRoute: `/kubernetes?tab=deployments&namespace=${d.namespace}`,
          logsRoute: `/kubernetes?tab=deployments&namespace=${d.namespace}`,
        });
      });
    }
  } else if (activeTab.value === 'pods') {
    if (Array.isArray(pods.value)) {
      pods.value.forEach((p) => {
        const isRunning = p.phase === 'Running';
        const readyDisplay = p.ready ? p.ready.replace('/', ' / ') : (isRunning ? '1 / 1' : '0 / 1');
        items.push({
          uniqueKey: `k8s-pod-${p.namespace}-${p.name}`,
          id: p.name,
          name: p.name,
          platform: 'Kubernetes',
          type: 'Pod',
          statusText: p.phase || 'Running',
          statusClass: isRunning ? 'running' : 'pending',
          ready: readyDisplay,
          age: formatAge(p.created),
          raw: p,
          targetRoute: `/kubernetes/pods/${p.namespace}/${p.name}`,
          logsRoute: `/logs?p=${encodeURIComponent(p.namespace + '/' + p.name)}&rt=kubernetes`,
        });
      });
    }
  } else if (activeTab.value === 'nodes') {
    if (Array.isArray(swarmNodes.value) && swarmNodes.value.length > 0) {
      swarmNodes.value.forEach((n) => {
        const stateStr = (n.state || n.status || '').toLowerCase();
        const availStr = (n.availability || '').toLowerCase();
        // A node is ready if its state is "ready" (or availability is "active" and state not down)
        const isReady = stateStr === 'ready' || (availStr === 'active' && stateStr !== 'down' && stateStr !== 'unknown');
        const displayStatus = n.availability ? (n.availability.charAt(0).toUpperCase() + n.availability.slice(1)) : (isReady ? 'Active' : 'Down');
        items.push({
          uniqueKey: `swarm-node-${n.id}`,
          id: n.id,
          name: n.hostname || n.id.slice(0, 10),
          platform: 'Swarm',
          type: n.role === 'manager' ? 'Manager Node' : 'Worker Node',
          statusText: displayStatus,
          statusClass: isReady ? 'running' : 'warning',
          ready: isReady ? '1 / 1' : '0 / 1',
          age: formatAge(n.created_at),
          raw: n,
          targetRoute: '/swarm?tab=nodes',
          logsRoute: '/swarm?tab=nodes',
        });
      });
    } else {
      items.push({
        uniqueKey: 'host-node-primary',
        id: 'host-node',
        name: 'docker-host-node',
        platform: 'Docker',
        type: 'Host Node',
        statusText: 'Running',
        statusClass: 'running',
        ready: '1 / 1',
        age: '1d 12h',
        raw: {},
        targetRoute: '/health',
        logsRoute: '/health',
      });
    }
  }

  return items;
});

// Filter by platform + search query
const filteredItems = computed(() => {
  let list = tableItems.value;

  if (selectedPlatform.value !== 'all') {
    list = list.filter(
      (item) => item.platform.toLowerCase() === selectedPlatform.value.toLowerCase()
    );
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter(
      (item) =>
        item.name.toLowerCase().includes(q) ||
        item.platform.toLowerCase().includes(q) ||
        item.type.toLowerCase().includes(q)
    );
  }

  return list;
});

function handleItemNavigate(item) {
  if (item.targetRoute) {
    router.push(item.targetRoute);
  }
}

function handleItemLogs(item) {
  if (item.logsRoute) {
    router.push(item.logsRoute);
  }
}

function toggleMenu(id) {
  openMenuId.value = openMenuId.value === id ? null : id;
}

function closeMenuOnClickOutside(e) {
  if (!e.target.closest('.menu-container')) {
    openMenuId.value = null;
  }
}

// Hover Card with Live Stats
const hoveredContainer = ref(null);
const hoverPosition = ref({ top: 0, left: 0 });
const isHoverCardActive = ref(false);
const liveStats = ref({ cpu: 0, memory: 0 });
let hoverLeaveTimeout = null;
let liveStatsWs = null;

function startLiveStats(id) {
  stopLiveStats();
  try {
    liveStatsWs = createAuthenticatedWebSocket(`/ws/containers/${id}/stats`);
    liveStatsWs.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        liveStats.value = { cpu: data.cpu, memory: data.memory };
      } catch (e) {}
    };
    liveStatsWs.onerror = () => {
      stopLiveStats();
    };
  } catch (e) {
    liveStatsWs = null;
  }
}

function stopLiveStats() {
  if (liveStatsWs) {
    liveStatsWs.onclose = null;
    liveStatsWs.close();
    liveStatsWs = null;
  }
}

function handleItemMouseEnter(item, event) {
  // Only inspect live metrics for containers that are running
  if (item.type !== 'Container' || item.statusClass !== 'running' || !item.id) {
    return;
  }
  if (hoverLeaveTimeout) {
    clearTimeout(hoverLeaveTimeout);
    hoverLeaveTimeout = null;
  }
  const rawContainer = item.raw || {
    id: item.id,
    name: item.name,
    state: 'running',
    image: item.platform,
  };
  hoveredContainer.value = rawContainer;
  liveStats.value = {
    cpu: rawContainer.cpu || 0,
    memory: rawContainer.memory || 0,
  };

  if (event && event.currentTarget) {
    const rect = event.currentTarget.getBoundingClientRect();
    const cardWidth = 290;
    const cardHeight = 140;
    let top = rect.top + rect.height / 2 - cardHeight / 2;
    top = Math.max(12, Math.min(window.innerHeight - cardHeight - 16, top));

    let left = rect.left + 220;
    if (left + cardWidth > window.innerWidth - 16) {
      left = Math.max(16, rect.left - cardWidth - 12);
    }
    hoverPosition.value = { top, left };
  }

  isHoverCardActive.value = true;
  startLiveStats(item.id);
}

function handleItemMouseLeave() {
  hoverLeaveTimeout = setTimeout(() => {
    isHoverCardActive.value = false;
    hoveredContainer.value = null;
    stopLiveStats();
  }, 200);
}

function handleCardMouseEnter() {
  if (hoverLeaveTimeout) {
    clearTimeout(hoverLeaveTimeout);
    hoverLeaveTimeout = null;
  }
}

function handleCardMouseLeave() {
  isHoverCardActive.value = false;
  hoveredContainer.value = null;
  stopLiveStats();
}

let refreshTimer = null;

onMounted(async () => {
  document.addEventListener('click', closeMenuOnClickOutside);
  fetchContainers();
  fetchSwarmData();
  if (kubernetesEnabled() && sharedState.k8sAvailable) {
    fetchNamespaces();
    fetchKubernetesData();
  }
});

onUnmounted(() => {
  document.removeEventListener('click', closeMenuOnClickOutside);
  if (refreshTimer) clearInterval(refreshTimer);
  stopLiveStats();
  if (hoverLeaveTimeout) clearTimeout(hoverLeaveTimeout);
});
</script>

<style scoped>
@import './dashboard.css';
</style>


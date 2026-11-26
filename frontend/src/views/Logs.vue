<template>
  <div
    class="logs-container-layout"
    :class="{ 'sidebar-collapsed': isSidebarHidden }"
  >
    <!-- RESOURCES SIDEBAR -->
    <aside class="resources-sidebar glass">
      <div class="sidebar-top-nav">
        <!-- Mobile Header Row -->
        <div class="mobile-header-row show-mobile">
          <router-link to="/dashboard" class="back-nav-link-mobile">
            <svg
              viewBox="0 0 24 24"
              width="16"
              height="16"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
            >
              <polyline points="15 18 9 12 15 6"></polyline>
            </svg>
            <span>Dashboard</span>
          </router-link>

          <button
            class="minimal-close-btn-glass"
            @click="isSidebarHidden = true"
          >
            <svg
              viewBox="0 0 24 24"
              width="20"
              height="20"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      </div>

      <div class="sidebar-header">
        <div class="sidebar-header-top">
          <span class="label-caps">Log Resources</span>
          <div class="sidebar-controls">
            <button
              class="mini-icon-btn"
              @click="isSidebarHidden = true"
              data-tooltip="Collapse Sidebar"
            >
              <svg
                viewBox="0 0 24 24"
                width="12"
                height="12"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <polyline points="15 18 9 12 15 6"></polyline>
              </svg>
            </button>

            <button
              class="mini-icon-btn hide-mobile"
              :class="{ 'active-toggle': splitView }"
              @click="toggleSplitView"
              :data-tooltip="
                splitView ? 'Disable Split View' : 'Enable Split View'
              "
            >
              <svg
                viewBox="0 0 24 24"
                width="12"
                height="12"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                <line x1="12" y1="3" x2="12" y2="21"></line>
              </svg>
            </button>
          </div>
        </div>

        <div v-if="showRuntimeToggle" class="runtime-toggle">
          <button
            class="runtime-btn"
            :class="{ active: activeRuntime === 'docker' }"
            @click="setRuntime('docker')"
          >
            Docker
          </button>
          <button
            v-if="hasSwarmSupport"
            class="runtime-btn"
            :class="{ active: activeRuntime === 'swarm' }"
            @click="setRuntime('swarm')"
          >
            Swarm
          </button>
          <button
            v-if="kubernetesEnabled()"
            class="runtime-btn"
            :class="{ active: activeRuntime === 'kubernetes' }"
            @click="setRuntime('kubernetes')"
          >
            Kubernetes
          </button>
        </div>
      </div>

      <div class="resource-list tree-resource-list">
        <!-- DOCKER RUNTIME: STANDALONE DOCKER CONTAINERS ONLY -->
        <template v-if="dockerEnabled() && activeRuntime === 'docker'">
          <div class="tree-group all-containers-group">
            <div
              class="tree-group-header"
              @click="allContainersCollapsed = !allContainersCollapsed"
            >
              <div class="header-main-side">
                <svg
                  class="chevron-arrow"
                  :class="{ collapsed: allContainersCollapsed }"
                  viewBox="0 0 24 24"
                  width="12"
                  height="12"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                >
                  <polyline points="6 9 12 15 18 9"></polyline>
                </svg>
                <!-- Cube / Package icon matching reference -->
                <svg
                  class="group-type-icon box-icon"
                  viewBox="0 0 24 24"
                  width="16"
                  height="16"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
                  <polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline>
                  <line x1="12" y1="22.08" x2="12" y2="12"></line>
                </svg>
                <span class="group-title">All Containers</span>
              </div>
              <span class="group-count">{{ filteredStandaloneContainers.length }}</span>
            </div>

            <!-- Standalone Containers List -->
            <div v-if="!allContainersCollapsed" class="tree-group-children">
              <div
                v-for="c in filteredStandaloneContainers"
                :key="c.id"
                class="tree-item container-item"
                :class="{ active: isContainerVisible(c.id) }"
                @click="toggleContainerStream(c.id)"
                @mouseenter="handleContainerMouseEnter(c, $event)"
                @mouseleave="handleContainerMouseLeave"
              >
                <div class="tree-item-icon-wrapper">
                  <!-- Cylinder/Container database-style icon matching reference -->
                  <svg
                    class="item-icon container-cylinder-icon"
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.8"
                  >
                    <ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
                    <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
                    <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
                  </svg>
                  <span
                    class="mini-status-dot"
                    :class="c.state"
                  ></span>
                </div>
                <span class="tree-item-name truncate">{{ c.name }}</span>
              </div>

              <div v-if="filteredStandaloneContainers.length === 0" class="empty-search-msg">
                <p class="text-mute">No standalone containers</p>
              </div>
            </div>
          </div>
        </template>

        <!-- SWARM RUNTIME: SWARM STACKS, SERVICES & TASK CONTAINERS -->
        <template v-if="activeRuntime === 'swarm'">
          <div
            v-for="grp in swarmGroups"
            :key="'swarm-group-' + grp.name"
            class="tree-group swarm-tree-group"
          >
            <div
              class="tree-group-header"
              @click="toggleSwarmCollapse(grp.name)"
            >
              <div class="header-main-side">
                <svg
                  class="chevron-arrow"
                  :class="{ collapsed: swarmCollapsed[grp.name] }"
                  viewBox="0 0 24 24"
                  width="12"
                  height="12"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                >
                  <polyline points="6 9 12 15 18 9"></polyline>
                </svg>
                <!-- Swarm Stack Icon (layers) -->
                <svg
                  class="group-type-icon swarm-icon"
                  viewBox="0 0 24 24"
                  width="16"
                  height="16"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
                  <polyline points="2 17 12 22 22 17"></polyline>
                  <polyline points="2 12 12 17 22 12"></polyline>
                </svg>
                <span class="group-title">{{ grp.name }}</span>
              </div>
              <span class="group-count">{{ grp.containers.length || grp.services.length }}</span>
            </div>

            <!-- Swarm Containers & Services within this Stack -->
            <div v-if="!swarmCollapsed[grp.name]" class="tree-group-children">
              <!-- Swarm Service Log shortcut if available -->
              <div
                v-for="svc in grp.services"
                :key="'svc-' + svc.id"
                class="tree-item service-item"
                :class="{ active: isServiceVisible(svc.id) }"
                @click="toggleServiceStream(svc.id)"
                :title="svc.name + ' (All replicas aggregated stream)'"
              >
                <div class="tree-item-icon-wrapper">
                  <!-- Service aggregate icon -->
                  <svg
                    class="item-icon svc-icon"
                    viewBox="0 0 24 24"
                    width="15"
                    height="15"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
                    <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
                    <line x1="6" y1="6" x2="6.01" y2="6"></line>
                    <line x1="6" y1="18" x2="6.01" y2="18"></line>
                  </svg>
                  <span
                    class="mini-status-dot"
                    :class="svc.running === svc.replicas && svc.replicas > 0 ? 'running' : (svc.running > 0 ? 'partial' : 'stopped')"
                  ></span>
                </div>
                <div class="tree-item-name">
                  <span>{{ svc.name }}</span>
                  <span class="mini-tag">svc</span>
                </div>
              </div>

              <!-- Swarm Task Containers -->
              <div
                v-for="c in grp.containers"
                :key="c.id"
                class="tree-item container-item"
                :class="{ active: isContainerVisible(c.id) }"
                @click="toggleContainerStream(c.id)"
                @mouseenter="handleContainerMouseEnter(c, $event)"
                @mouseleave="handleContainerMouseLeave"
              >
                <div class="tree-item-icon-wrapper">
                  <!-- Cylinder/Container database-style icon matching reference -->
                  <svg
                    class="item-icon container-cylinder-icon"
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.8"
                  >
                    <ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
                    <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
                    <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
                  </svg>
                  <span
                    class="mini-status-dot"
                    :class="c.state"
                  ></span>
                </div>
                <span class="tree-item-name truncate">{{ c.name }}</span>
              </div>
            </div>
          </div>

          <div v-if="swarmGroups.length === 0" class="empty-search-msg">
            <p class="text-mute">No swarm services or stacks found</p>
          </div>
        </template>

        <!-- KUBERNETES RUNTIME: NAMESPACES & PODS (TREE-VIEW STYLED) -->
        <template v-if="kubernetesEnabled() && activeRuntime === 'kubernetes'">
          <div
            v-for="grp in k8sNamespaceGroups"
            :key="'k8s-ns-' + grp.namespace"
            class="tree-group k8s-tree-group"
          >
            <div
              class="tree-group-header"
              @click="toggleK8sCollapse(grp.namespace)"
            >
              <div class="header-main-side">
                <svg
                  class="chevron-arrow"
                  :class="{ collapsed: k8sCollapsed[grp.namespace] }"
                  viewBox="0 0 24 24"
                  width="12"
                  height="12"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                >
                  <polyline points="6 9 12 15 18 9"></polyline>
                </svg>
                <!-- Kubernetes wheel icon -->
                <svg
                  class="group-type-icon k8s-icon"
                  viewBox="0 0 24 24"
                  width="16"
                  height="16"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <circle cx="12" cy="12" r="9"></circle>
                  <circle cx="12" cy="12" r="3"></circle>
                  <line x1="12" y1="3" x2="12" y2="9"></line>
                  <line x1="12" y1="15" x2="12" y2="21"></line>
                  <line x1="4.2" y1="7.5" x2="9.4" y2="10.5"></line>
                  <line x1="14.6" y1="13.5" x2="19.8" y2="16.5"></line>
                  <line x1="4.2" y1="16.5" x2="9.4" y2="13.5"></line>
                  <line x1="14.6" y1="10.5" x2="19.8" y2="7.5"></line>
                </svg>
                <span class="group-title">{{ grp.namespace }}</span>
              </div>
              <span class="group-count">{{ grp.pods.length }}</span>
            </div>

            <!-- Pods inside namespace -->
            <div v-if="!k8sCollapsed[grp.namespace]" class="tree-group-children">
              <div
                v-for="pod in grp.pods"
                :key="podKey(pod)"
                class="tree-item container-item"
                :class="{ active: isPodVisible(pod) }"
                @click="togglePodStream(pod)"
                :title="pod.name"
              >
                <div class="tree-item-icon-wrapper">
                  <!-- Cylinder/pod container icon matching Docker & Swarm -->
                  <svg
                    class="item-icon container-cylinder-icon"
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.8"
                  >
                    <ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
                    <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
                    <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
                  </svg>
                  <span
                    class="mini-status-dot"
                    :class="(pod.phase || '').toLowerCase() === 'running' ? 'running' : 'stopped'"
                  ></span>
                </div>
                <div class="tree-item-name">
                  <span class="truncate">{{ pod.name }}</span>
                </div>
              </div>
            </div>
          </div>

          <div v-if="k8sNamespaceGroups.length === 0" class="empty-search-msg">
            <p class="text-mute">No pods found</p>
          </div>
        </template>
      </div>
    </aside>

    <!-- Mobile Overlay -->
    <div
      v-if="!isSidebarHidden"
      class="mobile-sidebar-overlay show-mobile"
      @click="isSidebarHidden = true"
    ></div>

    <!-- FIXED TRIGGER (Always accessible to reopen sidebar) -->
    <button
      v-if="isSidebarHidden"
      class="fixed-open-trigger glass shadow-xl"
      @click="isSidebarHidden = false"
    >
      <svg
        viewBox="0 0 24 24"
        width="16"
        height="16"
        fill="none"
        stroke="currentColor"
        stroke-width="3"
      >
        <polyline points="9 18 15 12 9 6"></polyline>
      </svg>
    </button>

    <!-- MAIN VIEWPORT -->
    <main class="logs-main-content">
      <div
        v-if="displayStreams.length > 0"
        class="logs-grid"
        :class="gridClass"
      >
        <LogViewer
          v-for="stream in displayStreams"
          :key="stream.key"
          :container="stream.type === 'container' ? stream.data : undefined"
          :pod="stream.type === 'pod' ? stream.data : undefined"
          :service="stream.type === 'service' ? stream.data : undefined"
          showClose
          @close="removeStream(stream)"
          @stats="handleViewerStats"
        />
      </div>

      <!-- PREMIUM HERO EMPTY STATE -->
      <div v-else class="empty-state-wrapper animate-fade-in">
        <div class="observability-hero">
          <div class="hero-visual">
            <div class="radar-scan">
              <div class="circle c1"></div>
              <div class="circle c2"></div>
              <div class="circle c3"></div>
              <svg
                viewBox="0 0 24 24"
                class="hero-icon"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
                ></path>
                <path d="M8 9h8M8 13h5" stroke-linecap="round"></path>
              </svg>
            </div>
          </div>
          <div class="hero-text">
            <h2 class="display-title">Ready for Insight?</h2>
            <p class="subtitle">
              Select a resource from the sidebar to launch a real-time log
              stream. Toggle split-view to monitor up to two resources
              simultaneously.
            </p>
            <div class="hero-actions" v-if="isSidebarHidden">
              <button
                @click="isSidebarHidden = false"
                class="premium-btn primary"
              >
                <svg
                  viewBox="0 0 24 24"
                  width="18"
                  height="18"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                >
                  <path d="M9 18l6-6-6-6"></path>
                </svg>
                Open Resources
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Interactive Container Hover Card -->
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
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { sharedState, dockerEnabled, kubernetesEnabled, swarmEnabled, showToast } from "../utils/sharedState";
import { parsePodKey } from "../utils/logRoutes";
import { usePods } from "../composables/usePods";
import { useSwarm } from "../composables/useSwarm";
import { secureStorage } from "../utils/storage";
import { apiFetch } from "../utils/apiFetch";
import { createAuthenticatedWebSocket } from "../utils/wsAuth";
import ContainerHoverCard from "../components/ContainerHoverCard.vue";
import LogViewer from "../components/LogViewer.vue";

const route = useRoute();
const router = useRouter();

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return "0B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + sizes[i];
};

const containers = ref([]);
const { pods } = usePods({ autoPoll: false });
const { swarmServices, fetchSwarmData } = useSwarm();

const podKey = (pod) => `${pod.namespace}/${pod.name}`;

// Live Stats on Hover Logic
const activeLiveId = ref(null);
const hoveredContainer = ref(null);
const hoverPosition = ref({ top: 0, left: 0 });
const isHoverCardActive = ref(false);
let hoverLeaveTimeout = null;
const liveStats = ref({ cpu: 0, memory: 0 });
let liveInterval = null;

const handleViewerStats = (data) => {
  if (data.id === activeLiveId.value) {
    liveStats.value = { cpu: data.cpu, memory: data.memory };
  }
};

const handleContainerMouseEnter = (c, event) => {
  if (hoverLeaveTimeout) {
    clearTimeout(hoverLeaveTimeout);
    hoverLeaveTimeout = null;
  }
  hoveredContainer.value = c;
  // Initialize with container's cached values so it doesn't flash 0
  liveStats.value = {
    cpu: c.cpu || 0,
    memory: c.memory || 0,
  };
  if (event && event.currentTarget) {
    const rect = event.currentTarget.getBoundingClientRect();
    const top = Math.max(10, Math.min(window.innerHeight - 350, rect.top));
    const left = rect.right + 12;
    hoverPosition.value = { top, left };
  }
  isHoverCardActive.value = true;
  startLiveStats(c.id);
};

const handleContainerMouseLeave = () => {
  hoverLeaveTimeout = setTimeout(() => {
    isHoverCardActive.value = false;
    hoveredContainer.value = null;
    stopLiveStats();
  }, 220);
};

const handleCardMouseEnter = () => {
  if (hoverLeaveTimeout) {
    clearTimeout(hoverLeaveTimeout);
    hoverLeaveTimeout = null;
  }
};

const handleCardMouseLeave = () => {
  isHoverCardActive.value = false;
  hoveredContainer.value = null;
  stopLiveStats();
};

let liveStatsWs = null;

const startLiveStats = (id) => {
  activeLiveId.value = id;
  stopLiveStats();
  activeLiveId.value = id;
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
};

const stopLiveStats = () => {
  activeLiveId.value = null;
  if (liveStatsWs) {
    liveStatsWs.onclose = null;
    liveStatsWs.close();
    liveStatsWs = null;
  }
};

const containerScope = ref('all'); // 'all', 'docker', 'swarm'

const hasSwarmContainers = computed(() => {
  return containers.value.some((c) => c.is_swarm) || swarmServices.value.length > 0;
});

const filteredContainers = computed(() => {
  let list = containers.value;
  if (containerScope.value === 'docker') {
    list = list.filter((c) => !c.is_swarm);
  } else if (containerScope.value === 'swarm') {
    list = list.filter((c) => c.is_swarm);
  }

  if (!sharedState.searchQuery) return list;
  const q = sharedState.searchQuery.toLowerCase();
  return list.filter(
    (c) =>
      c.name.toLowerCase().includes(q) ||
      c.id.toLowerCase().includes(q) ||
      (c.swarm_service && c.swarm_service.toLowerCase().includes(q)),
  );
});

const swarmCollapsed = ref({});
const allContainersCollapsed = ref(false);

const toggleSwarmCollapse = (groupName) => {
  swarmCollapsed.value[groupName] = !swarmCollapsed.value[groupName];
};

const standaloneContainers = computed(() => {
  return containers.value.filter((c) => !c.is_swarm);
});

const filteredStandaloneContainers = computed(() => {
  let list = standaloneContainers.value;
  if (!sharedState.searchQuery) return list;
  const q = sharedState.searchQuery.toLowerCase();
  return list.filter(
    (c) =>
      c.name.toLowerCase().includes(q) ||
      c.id.toLowerCase().includes(q) ||
      c.image.toLowerCase().includes(q)
  );
});

// Group swarm containers and services under their stack / service family
const swarmGroups = computed(() => {
  const groupsMap = {};
  const q = (sharedState.searchQuery || "").toLowerCase();

  // Swarm containers
  containers.value.filter((c) => c.is_swarm).forEach((c) => {
    // Determine group/stack name: e.g. "adminer-swarm_adminer.1..." -> "adminer-swarm"
    let groupName = c.swarm_service || "swarm";
    if (groupName.includes("_")) {
      groupName = groupName.split("_")[0];
    }
    if (!groupsMap[groupName]) {
      groupsMap[groupName] = {
        name: groupName,
        services: [],
        containers: [],
      };
    }
    if (!q || c.name.toLowerCase().includes(q) || c.id.toLowerCase().includes(q) || (c.swarm_service && c.swarm_service.toLowerCase().includes(q))) {
      groupsMap[groupName].containers.push(c);
    }
  });

  // Also include Swarm services if present
  swarmServices.value.forEach((s) => {
    let groupName = s.name || "swarm";
    if (groupName.includes("_")) {
      groupName = groupName.split("_")[0];
    }
    if (!groupsMap[groupName]) {
      groupsMap[groupName] = {
        name: groupName,
        services: [],
        containers: [],
      };
    }
    if (!q || s.name.toLowerCase().includes(q) || s.id.toLowerCase().includes(q) || s.image.toLowerCase().includes(q)) {
      groupsMap[groupName].services.push(s);
    }
  });

  return Object.values(groupsMap).filter((g) => g.containers.length > 0 || g.services.length > 0);
});

const filteredSwarmServices = computed(() => {
  if (!sharedState.searchQuery) return swarmServices.value;
  const q = sharedState.searchQuery.toLowerCase();
  return swarmServices.value.filter(
    (s) =>
      s.name.toLowerCase().includes(q) ||
      s.id.toLowerCase().includes(q) ||
      s.image.toLowerCase().includes(q)
  );
});

const filteredPods = computed(() => {
  if (!sharedState.searchQuery) return pods.value;
  const q = sharedState.searchQuery.toLowerCase();
  return pods.value.filter(
    (pod) =>
      pod.name.toLowerCase().includes(q) ||
      pod.namespace.toLowerCase().includes(q) ||
      (pod.images || []).some((image) => image.toLowerCase().includes(q)),
  );
});

const k8sCollapsed = ref({});
const toggleK8sCollapse = (namespace) => {
  k8sCollapsed.value[namespace] = !k8sCollapsed.value[namespace];
};

const k8sNamespaceGroups = computed(() => {
  const groupsMap = {};
  filteredPods.value.forEach((p) => {
    const ns = p.namespace || "default";
    if (!groupsMap[ns]) {
      groupsMap[ns] = {
        namespace: ns,
        pods: [],
      };
    }
    groupsMap[ns].pods.push(p);
  });
  return Object.values(groupsMap);
});

const selectedContainerIds = ref([]);
const selectedPodKeys = ref([]);
const selectedServiceIds = ref([]);
const isSidebarHidden = ref(window.innerWidth < 1024);
const splitView = ref(route.query.split === "true");
const activeRuntime = ref(dockerEnabled() ? "docker" : "kubernetes");

const hasSwarmSupport = computed(() => {
  return swarmEnabled() || sharedState.swarmActive || hasSwarmContainers.value;
});

const showRuntimeToggle = computed(() => {
  let count = 0;
  if (dockerEnabled()) count++;
  if (hasSwarmSupport.value) count++;
  if (kubernetesEnabled()) count++;
  return count > 1;
});

const syncStateFromUrl = () => {
  const containerParam = route.query.c;
  let podParam = route.query.p;
  // Handle fallback params if navigated from older links: k8s_ns & k8s_pod
  if (!podParam && route.query.k8s_ns && route.query.k8s_pod) {
    podParam = `${route.query.k8s_ns}/${route.query.k8s_pod}`;
  }
  const serviceParam = route.query.s || route.query.service;

  selectedContainerIds.value = containerParam
    ? String(containerParam).split(",").filter(Boolean)
    : [];
  selectedPodKeys.value = podParam
    ? String(podParam).split(",").filter(Boolean)
    : [];
  selectedServiceIds.value = serviceParam
    ? String(serviceParam).split(",").filter(Boolean)
    : [];

  splitView.value = route.query.split === "true";
  if (showRuntimeToggle.value) {
    const rt = String(route.query.rt || "").toLowerCase();
    if (rt === "kubernetes" || rt === "docker" || rt === "swarm") {
      activeRuntime.value = rt;
    } else if (selectedPodKeys.value.length > 0) {
      activeRuntime.value = "kubernetes";
    } else if (selectedServiceIds.value.length > 0) {
      activeRuntime.value = "swarm";
    } else if (selectedContainerIds.value.length > 0) {
      const match = containers.value.find(
        (c) =>
          selectedContainerIds.value.includes(c.id) ||
          selectedContainerIds.value.some((id) => c.id.startsWith(id) || id.startsWith(c.id)),
      );
      if (match?.is_swarm) {
        activeRuntime.value = "swarm";
      } else if (match) {
        activeRuntime.value = "docker";
      }
    } else {
      activeRuntime.value = dockerEnabled() ? "docker" : (hasSwarmSupport.value ? "swarm" : "kubernetes");
    }
  } else {
    activeRuntime.value = dockerEnabled() ? "docker" : (hasSwarmSupport.value ? "swarm" : "kubernetes");
  }
};

// Helper matching functions for container identification & recreation tracking
const cleanContainerName = (name) => {
  if (!name) return "";
  return name.startsWith("/") ? name.slice(1).trim().toLowerCase() : name.trim().toLowerCase();
};

const cleanImageName = (image) => {
  if (!image) return "";
  const trimmed = image.trim().toLowerCase();
  // Strip tag or sha if present for base comparison (e.g., adminer:latest -> adminer)
  const atIdx = trimmed.indexOf("@");
  const withoutSha = atIdx !== -1 ? trimmed.slice(0, atIdx) : trimmed;
  const colonIdx = withoutSha.lastIndexOf(":");
  return colonIdx !== -1 ? withoutSha.slice(0, colonIdx) : withoutSha;
};

const isImageMatch = (img1, img2) => {
  if (!img1 || !img2) return true; // If one doesn't specify image, fall back to name match
  const t1 = img1.trim().toLowerCase();
  const t2 = img2.trim().toLowerCase();
  return t1 === t2 || cleanImageName(t1) === cleanImageName(t2);
};

const isContainerRecreationMatch = (candidate, target) => {
  if (!candidate || !target) return false;
  const nameMatch = cleanContainerName(candidate.name) === cleanContainerName(target.name);
  if (!nameMatch) return false;
  // Must match image as well
  return isImageMatch(candidate.image, target.image);
};

const displayContainers = computed(() => {
  if (containers.value.length === 0 || selectedContainerIds.value.length === 0) {
    return [];
  }

  const ordered = selectedContainerIds.value
    .map((id) => {
      let match = containers.value.find((c) => c.id === id);
      if (!match) {
        match = containers.value.find(
          (c) => c.id.startsWith(id) || id.startsWith(c.id),
        );
      }
      // If container was recreated (ID changed, but same name and same image)
      if (!match && previousContainersMap.size > 0) {
        const prev = previousContainersMap.get(id) ||
          Array.from(previousContainersMap.values()).find(
            (c) => c.id === id || c.id.startsWith(id) || id.startsWith(c.id)
          );
        if (prev && prev.name) {
          match = containers.value.find((c) => isContainerRecreationMatch(c, prev));
          // Seamlessly update selectedContainerIds to new ID
          if (match) {
            const idx = selectedContainerIds.value.indexOf(id);
            if (idx !== -1) {
              selectedContainerIds.value[idx] = match.id;
              updateUrl();
              showToast(
                "Container Reconnected",
                `Container "${match.name}" restarted with new ID (${match.id.substring(0, 12)}). Stream reconnected.`,
                "info"
              );
            }
          }
        }
      }
      return match;
    })
    .filter(Boolean);

  return splitView.value ? ordered.slice(-2) : [ordered[ordered.length - 1]].filter(Boolean);
});

const displayPods = computed(() => {
  if (selectedPodKeys.value.length === 0) {
    return [];
  }

  const ordered = selectedPodKeys.value
    .map((key) => {
      const match = pods.value.find((pod) => podKey(pod) === key);
      if (match) return match;
      const parsed = parsePodKey(key);
      if (!parsed) return null;
      return { ...parsed, phase: "", ready: "", images: [] };
    })
    .filter(Boolean);

  return splitView.value ? ordered.slice(-2) : [ordered[ordered.length - 1]].filter(Boolean);
});

const displayServices = computed(() => {
  if (selectedServiceIds.value.length === 0) {
    return [];
  }

  const ordered = selectedServiceIds.value
    .map((id) => {
      let match = swarmServices.value.find((s) => s.id === id);
      if (!match) {
        match = swarmServices.value.find((s) => s.id.startsWith(id) || s.name === id);
      }
      if (match) return match;
      return { id, name: id, running: 1, replicas: 1 };
    })
    .filter(Boolean);

  return splitView.value ? ordered.slice(-2) : [ordered[ordered.length - 1]].filter(Boolean);
});

const displayStreams = computed(() => {
  const streams = [
    ...displayContainers.value.map((container) => ({
      type: "container",
      key: `container:${container.id}`,
      data: container,
    })),
    ...displayPods.value.map((pod) => ({
      type: "pod",
      key: `pod:${podKey(pod)}`,
      data: pod,
    })),
    ...displayServices.value.map((service) => ({
      type: "service",
      key: `service:${service.id}`,
      data: service,
    })),
  ];
  return splitView.value ? streams.slice(-2) : streams.slice(-1);
});

watch(
  () => [containers.value, pods.value, swarmServices.value],
  () => {
    if (selectedContainerIds.value.length > 0 || selectedPodKeys.value.length > 0 || selectedServiceIds.value.length > 0) {
      syncStateFromUrl();
    }
  },
  { immediate: true },
);

const isContainerVisible = (id) =>
  displayContainers.value.some((c) => c.id === id);
const isPodVisible = (pod) =>
  displayPods.value.some((item) => podKey(item) === podKey(pod));
const isServiceVisible = (serviceId) =>
  displayServices.value.some((s) => s.id === serviceId);

const gridClass = computed(() =>
  displayStreams.value.length > 1 ? "grid-dual" : "grid-single",
);

let previousContainersMap = new Map();
let followTransitionTimers = new Map();

// Update known containers cache so stopped containers retain their metadata (name, image)
const updateKnownContainersCache = (newContainersList) => {
  if (!Array.isArray(newContainersList)) return;
  newContainersList.forEach((c) => {
    if (c && c.id) {
      previousContainersMap.set(c.id, c);
    }
  });
  // Cap map size to avoid memory leak over time
  if (previousContainersMap.size > 200) {
    const keys = Array.from(previousContainersMap.keys());
    for (let i = 0; i < keys.length - 100; i++) {
      // Don't delete currently selected IDs
      if (!selectedContainerIds.value.includes(keys[i])) {
        previousContainersMap.delete(keys[i]);
      }
    }
  }
};

const handleContainerRecreationTracking = (newContainersList) => {
  if (!Array.isArray(newContainersList)) return;
  const followMode = sharedState.generalSettings?.followRecreated || "delayed";
  if (followMode === "never") {
    updateKnownContainersCache(newContainersList);
    return;
  }

  // Check each currently selected container
  if (selectedContainerIds.value.length > 0 && previousContainersMap.size > 0) {
    selectedContainerIds.value.forEach((currentId, index) => {
      // Check if currentId exists in new list
      const stillExists = newContainersList.some(
        (c) => c.id === currentId || c.id.startsWith(currentId) || currentId.startsWith(c.id)
      );

      if (!stillExists) {
        // Container was stopped or recreated. Find previous container details
        const prevContainer = previousContainersMap.get(currentId) ||
          Array.from(previousContainersMap.values()).find(
            (c) => c.id === currentId || c.id.startsWith(currentId) || currentId.startsWith(c.id)
          );

        if (prevContainer && prevContainer.name) {
          // Look for new container with matching name and image
          const replacement = newContainersList.find(
            (c) => isContainerRecreationMatch(c, prevContainer) && c.id !== currentId
          );

          if (replacement) {
            const doSwitch = () => {
              if (selectedContainerIds.value[index] === currentId) {
                selectedContainerIds.value[index] = replacement.id;
                updateUrl();
                showToast(
                  "Container Reconnected",
                  `Container "${replacement.name}" restarted with new ID (${replacement.id.substring(0, 12)}). Stream reconnected.`,
                  "info"
                );
              }
            };

            if (followMode === "immediate") {
              doSwitch();
            } else if (followMode === "delayed") {
              // Delayed (Smooth): wait 1500ms before seamlessly switching
              if (!followTransitionTimers.has(currentId)) {
                showToast(
                  "Recreation Detected",
                  `Switching to new ${replacement.name} container in 1.5s...`,
                  "info"
                );
                const timer = setTimeout(() => {
                  followTransitionTimers.delete(currentId);
                  doSwitch();
                }, 1500);
                followTransitionTimers.set(currentId, timer);
              }
            }
          }
        }
      }
    });
  }

  updateKnownContainersCache(newContainersList);
};

const fetchContainers = async () => {
  try {
    const token = secureStorage.getItem("token");
    const res = await apiFetch("/api/containers", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      const data = await res.json();
      handleContainerRecreationTracking(data);
      containers.value = data;
    }
  } catch (err) {
    console.error(err);
  }
};

const fetchPodsForLogs = async () => {
  if (!kubernetesEnabled() || !sharedState.k8sAvailable) return;
  try {
    const token = secureStorage.getItem("token");
    const nsRes = await apiFetch("/api/namespaces", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!nsRes.ok) return;
    const namespaces = await nsRes.json();
    if (!Array.isArray(namespaces)) return;
    const results = await Promise.all(
      namespaces.map(async ({ name }) => {
        const res = await apiFetch(
          `/api/pods?namespace=${encodeURIComponent(name)}`,
          { headers: { Authorization: `Bearer ${token}` } },
        );
        if (!res.ok) return [];
        return res.json();
      }),
    );
    pods.value = results.flat();
  } catch (err) {
    console.error(err);
  }
};

const fetchResources = async () => {
  const tasks = [];
  if (dockerEnabled()) tasks.push(fetchContainers());
  if (kubernetesEnabled() && sharedState.k8sAvailable) tasks.push(fetchPodsForLogs());
  if (swarmEnabled() || sharedState.swarmActive) tasks.push(fetchSwarmData());
  await Promise.all(tasks);
  syncStateFromUrl();
  ensureUrlReflectsSelection();
};

const ensureUrlReflectsSelection = () => {
  if (!showRuntimeToggle.value) return;

  let expectedRt = activeRuntime.value;
  if (selectedPodKeys.value.length > 0) {
    expectedRt = "kubernetes";
  } else if (selectedServiceIds.value.length > 0) {
    expectedRt = "swarm";
  } else if (selectedContainerIds.value.length > 0) {
    const match = containers.value.find(
      (c) =>
        selectedContainerIds.value.includes(c.id) ||
        selectedContainerIds.value.some((id) => c.id.startsWith(id) || id.startsWith(c.id)),
    );
    if (match?.is_swarm) {
      expectedRt = "swarm";
    } else if (match) {
      expectedRt = "docker";
    }
  }

  const expectedP =
    selectedPodKeys.value.length > 0
      ? selectedPodKeys.value.join(",")
      : "";
  const expectedC =
    selectedContainerIds.value.length > 0
      ? selectedContainerIds.value.join(",")
      : "";
  const expectedS =
    selectedServiceIds.value.length > 0
      ? selectedServiceIds.value.join(",")
      : "";

  const currentRt = String(route.query.rt || "");
  const currentP = String(route.query.p || "");
  const currentC = String(route.query.c || "");
  const currentS = String(route.query.s || route.query.service || "");

  if (
    currentRt !== expectedRt ||
    (expectedP && currentP !== expectedP) ||
    (expectedC && currentC !== expectedC) ||
    (expectedS && currentS !== expectedS) ||
    (!expectedP && currentP) ||
    (!expectedC && currentC) ||
    (!expectedS && currentS)
  ) {
    updateUrl();
  }
};

const updateUrl = () => {
  const query = { ...route.query };

  if (selectedContainerIds.value.length > 0) {
    query.c = selectedContainerIds.value.join(",");
  } else {
    delete query.c;
  }

  if (selectedPodKeys.value.length > 0) {
    query.p = selectedPodKeys.value.join(",");
  } else {
    delete query.p;
  }

  if (selectedServiceIds.value.length > 0) {
    query.s = selectedServiceIds.value.join(",");
    delete query.service;
  } else {
    delete query.s;
    delete query.service;
  }

  if (splitView.value) {
    query.split = "true";
  } else {
    delete query.split;
  }

  if (showRuntimeToggle.value) {
    query.rt = activeRuntime.value;
  } else {
    delete query.rt;
  }

  router.replace({ query });
};

const setRuntime = (runtime) => {
  if (!showRuntimeToggle.value) return;
  if (runtime !== "docker" && runtime !== "kubernetes" && runtime !== "swarm") return;
  if (activeRuntime.value === runtime) return;
  activeRuntime.value = runtime;
  selectedContainerIds.value = [];
  selectedPodKeys.value = [];
  selectedServiceIds.value = [];
  updateUrl();
};

const toggleSplitView = () => {
  splitView.value = !splitView.value;
  updateUrl();
};

const toggleContainerStream = (id) => {
  if (splitView.value) {
    if (selectedContainerIds.value.includes(id)) {
      selectedContainerIds.value = selectedContainerIds.value.filter((sid) => sid !== id);
    } else {
      if (selectedContainerIds.value.length + selectedServiceIds.value.length >= 2) {
        if (selectedServiceIds.value.length > 0) {
          selectedServiceIds.value.shift();
        } else {
          selectedContainerIds.value.shift();
        }
      }
      selectedContainerIds.value.push(id);
    }
  } else {
    selectedPodKeys.value = [];
    selectedServiceIds.value = [];
    selectedContainerIds.value = selectedContainerIds.value.includes(id) ? [] : [id];
  }
  if (window.innerWidth < 900) {
    isSidebarHidden.value = true;
  }
  updateUrl();
};

const toggleServiceStream = (serviceId) => {
  if (splitView.value) {
    if (selectedServiceIds.value.includes(serviceId)) {
      selectedServiceIds.value = selectedServiceIds.value.filter((sid) => sid !== serviceId);
    } else {
      if (selectedContainerIds.value.length + selectedServiceIds.value.length >= 2) {
        if (selectedContainerIds.value.length > 0) {
          selectedContainerIds.value.shift();
        } else {
          selectedServiceIds.value.shift();
        }
      }
      selectedServiceIds.value.push(serviceId);
    }
  } else {
    selectedPodKeys.value = [];
    selectedContainerIds.value = [];
    selectedServiceIds.value = selectedServiceIds.value.includes(serviceId) ? [] : [serviceId];
  }
  if (window.innerWidth < 900) {
    isSidebarHidden.value = true;
  }
  updateUrl();
};

const togglePodStream = (pod) => {
  const key = podKey(pod);
  if (splitView.value) {
    if (selectedPodKeys.value.includes(key)) {
      selectedPodKeys.value = selectedPodKeys.value.filter((sid) => sid !== key);
    } else {
      if (selectedPodKeys.value.length >= 2) {
        selectedPodKeys.value.shift();
      }
      selectedPodKeys.value.push(key);
    }
  } else {
    selectedContainerIds.value = [];
    selectedServiceIds.value = [];
    selectedPodKeys.value = selectedPodKeys.value.includes(key) ? [] : [key];
  }
  if (window.innerWidth < 900) {
    isSidebarHidden.value = true;
  }
  updateUrl();
};

const removeStream = (stream) => {
  if (stream.type === "container") {
    selectedContainerIds.value = selectedContainerIds.value.filter(
      (sid) => sid !== stream.data.id,
    );
  } else if (stream.type === "service") {
    selectedServiceIds.value = selectedServiceIds.value.filter(
      (sid) => sid !== stream.data.id,
    );
  } else {
    selectedPodKeys.value = selectedPodKeys.value.filter(
      (sid) => sid !== podKey(stream.data),
    );
  }
  updateUrl();
};

let logsEventsWs = null;
let statusInterval = null;
let logsReconnectTimer = null;
let logsDebounceTimer = null;

const scheduleLogsResourceRefresh = (delay = 400) => {
  if (logsDebounceTimer) clearTimeout(logsDebounceTimer);
  logsDebounceTimer = setTimeout(() => {
    fetchResources();
  }, delay);
};

const connectLogsEventsWebSocket = () => {
  if (logsEventsWs) return;
  try {
    logsEventsWs = createAuthenticatedWebSocket('/ws/events');
    logsEventsWs.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.Type === 'container' || msg.status || msg.Action) {
          scheduleLogsResourceRefresh(300);
        }
      } catch (e) {
        scheduleLogsResourceRefresh(500);
      }
    };
    logsEventsWs.onclose = () => {
      logsEventsWs = null;
      if (!logsReconnectTimer) {
        logsReconnectTimer = setTimeout(connectLogsEventsWebSocket, 5000);
      }
    };
    logsEventsWs.onerror = () => {
      if (logsEventsWs) logsEventsWs.close();
    };
  } catch (e) {
    logsEventsWs = null;
  }
};

onMounted(() => {
  fetchResources();
  connectLogsEventsWebSocket();
});

onUnmounted(() => {
  if (logsEventsWs) {
    logsEventsWs.onclose = null;
    logsEventsWs.close();
    logsEventsWs = null;
  }
  if (logsReconnectTimer) {
    clearTimeout(logsReconnectTimer);
    logsReconnectTimer = null;
  }
  if (logsDebounceTimer) {
    clearTimeout(logsDebounceTimer);
    logsDebounceTimer = null;
  }
  if (statusInterval) clearInterval(statusInterval);
  stopLiveStats();
  followTransitionTimers.forEach((timer) => clearTimeout(timer));
  followTransitionTimers.clear();
});

watch(() => route.query, syncStateFromUrl, { immediate: true });
</script>

<style scoped>
.logs-container-layout {
  display: flex;
  height: calc(100vh - 80px);
  position: relative;
  overflow: hidden;
  background: var(--bg-main);
}

.resource-section-label {
  margin: 0.75rem 0 0.35rem;
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-mute);
}

.card-status-dot.pending {
  background: var(--warning);
}

.card-status-dot.failed,
.card-status-dot.unknown {
  background: var(--error);
}

/* SIDEBAR UI FIXES */
.resources-sidebar {
  width: 300px;
  height: 100%;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  padding: 1.5rem 1.25rem;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
}

.sidebar-top-nav {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  width: 280px;
}

.logo-img-sidebar {
  width: 36px;
  height: 36px;
  object-fit: contain;
  border-radius: var(--radius-sm);
  background: transparent;
  border: none;
}

.logo-text {
  font-size: 1.2rem;
  font-weight: 950;
  color: var(--text-main);
  letter-spacing: -0.03em;
}

.back-nav-link,
.back-nav-link-mobile {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.75rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-dim);
  font-size: 0.75rem;
  font-weight: 700;
  text-decoration: none;
  transition: all 0.2s;
}

.back-nav-link:hover,
.back-nav-link-mobile:hover {
  color: var(--accent);
  border-color: rgba(var(--accent-rgb), 0.35);
  background: var(--accent-soft);
  transform: none;
}

.sidebar-header {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}

.sidebar-header-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.label-caps {
  text-transform: uppercase;
  font-size: 0.75rem;
  font-weight: 900;
  color: var(--text-mute);
  letter-spacing: 0.05em;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  overflow-y: auto;
  padding: 0 0.5rem;
}

.resource-card {
  padding: 0.85rem 1rem;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  gap: 0.85rem;
  border: 1px solid var(--border);
  background: var(--bg-card);
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s, transform 0.2s;
  position: relative;
  overflow: hidden;
}

.resource-card:hover {
  border-color: var(--accent);
  transform: translateX(4px);
  background: var(--card-hover);
}

/* Stats Peek Styling */
.stats-peek-inline {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%) translateX(10px);
  display: flex;
  gap: 0.75rem;
  background: var(--bg-glass);
  backdrop-filter: blur(8px);
  padding: 0.4rem 0.6rem;
  border-radius: 8px;
  border: 1px solid var(--border-light);
  opacity: 0;
  pointer-events: none;
  transition: all 0.2s ease;
  z-index: 10;
}

.resource-card:hover .stats-peek-inline {
  opacity: 1;
  transform: translateY(-50%) translateX(0);
}

.peek-stat {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.p-label {
  font-size: 0.55rem;
  font-weight: 900;
  color: var(--text-mute);
  text-transform: uppercase;
}

.p-value {
  font-size: 0.7rem;
  font-weight: 800;
  color: var(--accent);
  font-family: var(--font-mono);
  transition: color 0.3s;
}

.text-live {
  color: var(--success) !important;
  text-shadow: 0 0 8px rgba(var(--success-rgb), 0.4);
}

.resource-card.active {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.05);
}

.card-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #4b5563;
  flex-shrink: 0;
}

.card-status-dot.running {
  background: var(--success);
  box-shadow: 0 0 10px var(--success);
}

.card-status-dot.partial {
  background: var(--warning);
  box-shadow: 0 0 10px var(--warning);
}

.service-card {
  border-left: 3px solid var(--accent);
}

.swarm-header-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.count-pill {
  font-size: 0.65rem;
  background: var(--bg-input);
  padding: 0.1rem 0.4rem;
  border-radius: 999px;
  color: var(--text-mute);
  border: 1px solid var(--border);
}

.service-tag {
  font-size: 0.6rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  background: rgba(56, 189, 248, 0.18);
  color: #38bdf8;
  flex-shrink: 0;
}

.card-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
  flex: 1;
}

.card-name-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  min-width: 0;
}

.card-name {
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--text-main);
  white-space: nowrap !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  display: block;
  min-width: 0;
  flex: 1;
}

.swarm-tag {
  font-size: 0.6rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  background: rgba(var(--accent-rgb), 0.18);
  color: var(--accent);
  flex-shrink: 0;
}

.scope-toggle-pills {
  display: flex;
  gap: 0.25rem;
  padding: 0.25rem;
  margin-bottom: 0.65rem;
  border-radius: 8px;
  background: var(--bg-input);
  border: 1px solid var(--border);
}

.scope-pill {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--text-mute);
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.3rem 0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: center;
}

.scope-pill:hover {
  color: var(--text-main);
}

.scope-pill.active {
  background: var(--accent);
  color: #fff;
}

.card-image-tag {
  font-size: 0.7rem;
  color: var(--text-mute);
  font-family: monospace;
  white-space: nowrap !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  max-width: 100%;
  display: block;
}

.sidebar-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.6rem;
  margin-top: 0.5rem;
}

.mini-icon-btn {
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-mute);
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.active-toggle {
  background: var(--accent) !important;
  color: white !important;
  border-color: var(--accent) !important;
  box-shadow: 0 4px 14px rgba(var(--accent-rgb), 0.35);
}

.sidebar-controls {
  display: flex;
  gap: 0.5rem;
}

.sidebar-collapsed .resources-sidebar {
  width: 0;
  margin-right: -1.5rem;
  opacity: 0;
  pointer-events: none;
}

.fixed-open-trigger {
  position: fixed;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 60px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-left: none;
  border-radius: 0 12px 12px 0;
  z-index: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* GRID */
.logs-main-content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  padding: 1.5rem;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.logs-grid {
  display: grid;
  gap: 1.5rem;
  flex: 1;
  min-height: 0;
}
.logs-grid > :deep(.log-viewer) {
  min-height: 0;
  height: 100%;
}
.grid-single {
  grid-template-columns: 1fr;
}
.grid-dual {
  grid-template-columns: 1fr 1fr;
}

/* HERO EMPTY STATE */
.empty-state-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.observability-hero {
  text-align: center;
  max-width: 460px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2.5rem;
}

.hero-visual {
  position: relative;
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.radar-scan {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hero-icon {
  width: 56px;
  height: 56px;
  color: var(--accent);
  z-index: 2;
  filter: drop-shadow(0 0 15px rgba(var(--accent-rgb), 0.4));
}

.circle {
  position: absolute;
  border: 1px solid var(--accent);
  border-radius: 50%;
  opacity: 0;
  animation: radar 4s infinite linear;
}

.c1 {
  width: 60px;
  height: 60px;
  animation-delay: 0s;
}
.c2 {
  width: 60px;
  height: 60px;
  animation-delay: 1.3s;
}
.c3 {
  width: 60px;
  height: 60px;
  animation-delay: 2.6s;
}

@keyframes radar {
  0% {
    transform: scale(1);
    opacity: 0.6;
  }
  100% {
    transform: scale(3);
    opacity: 0;
  }
}

.display-title {
  font-size: 2.2rem;
  font-weight: 950;
  color: var(--text-main);
  letter-spacing: -0.04em;
  margin-bottom: 1rem;
}

.subtitle {
  font-size: 1rem;
  line-height: 1.6;
  color: var(--text-mute);
  font-weight: 500;
  margin-bottom: 2rem;
}

.runtime-toggle {
  display: flex;
  width: 100%;
  gap: 0.25rem;
  padding: 0.25rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-input);
}

.runtime-btn {
  flex: 1;
  text-align: center;
  border: none;
  background: transparent;
  color: var(--text-mute);
  font-size: 0.72rem;
  font-weight: 800;
  padding: 0.35rem 0.4rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.runtime-btn.active {
  background: var(--accent);
  color: #fff;
  box-shadow: 0 2px 8px rgba(var(--accent-rgb), 0.25);
}

@media (max-width: 1024px) {
  .logs-container-layout {
    flex-direction: column;
  }

  .resources-sidebar {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    width: 280px;
    z-index: 10000 !important;
    box-shadow: 20px 0 60px rgba(0, 0, 0, 0.8);
    transform: translateX(0);
    opacity: 1;
    padding: 1.25rem 1rem;
    backdrop-filter: none !important;
    transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
    visibility: visible;
  }

  .sidebar-top-nav {
    margin-bottom: 1.5rem !important;
    display: flex !important;
    flex-direction: column !important;
    gap: 0.75rem !important;
  }

  .mobile-header-row {
    display: flex !important;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--border);
  }

  .sidebar-logo-mobile {
    display: flex;
    align-items: center;
    gap: 0.85rem;
    text-decoration: none;
  }

  .logo-icon-premium {
    width: 36px;
    height: 36px;
    background: var(--accent);
    color: #fff;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 0 20px rgba(var(--accent-rgb), 0.35);
  }

  .logo-text-premium {
    font-size: 1.25rem;
    font-weight: 900;
    letter-spacing: -0.03em;
    color: #fff;
  }

  .minimal-close-btn-glass {
    width: 40px;
    height: 40px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 10px;
    color: var(--text-mute);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  }

  .minimal-close-btn-glass:hover {
    background: rgba(239, 68, 68, 0.15);
    color: var(--error);
    border-color: rgba(239, 68, 68, 0.3);
    transform: rotate(90deg);
  }

  .minimal-close-btn-glass:active {
    transform: scale(0.9);
  }

  .back-nav-link-mobile {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--text-main);
    text-decoration: none;
    font-weight: 850;
    font-size: 0.9rem;
    padding: 0.5rem 0.75rem;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    transition: all 0.2s;
  }

  .back-nav-link-mobile:hover {
    background: var(--accent-soft);
    color: var(--accent);
    border-color: rgba(var(--accent-rgb), 0.3);
  }

  .sidebar-header {
    margin-bottom: 1rem !important;
  }

  .sidebar-collapsed .resources-sidebar {
    transform: translateX(-100%);
    opacity: 0;
    visibility: hidden;
    pointer-events: none;
  }

  .logs-main-content {
    padding: 1rem;
    height: 100%;
    min-height: 0;
    z-index: 1;
    position: relative;
  }

  .grid-dual {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr 1fr;
  }

  .display-title {
    font-size: 1.75rem;
  }

  .hide-mobile {
    display: none !important;
  }

  .show-mobile {
    display: flex !important;
  }

  .mobile-close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    padding: 1rem;
    background: rgba(239, 68, 68, 0.08);
    color: #ef4444;
    border: 1px solid rgba(239, 68, 68, 0.15);
    border-radius: 18px;
    font-size: 0.9rem;
    font-weight: 900;
    cursor: pointer;
    margin-top: 0.5rem;
    width: 100%;
    transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  }

  .mobile-close-btn:hover {
    background: #ef4444;
    color: #fff;
    box-shadow: 0 10px 25px rgba(239, 68, 68, 0.3);
    transform: translateY(-2px);
  }

  .mobile-sidebar-overlay {
    position: fixed;
    inset: 0;
    background: rgba(2, 6, 23, 0.8);
    z-index: 4000;
    animation: fade-in 0.4s ease;
  }
}

.hero-actions {
  justify-items: center;
}

@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.show-mobile {
  display: none;
}

@media (max-width: 600px) {
  .resources-sidebar {
    width: 100%;
    max-width: none;
  }
}

/* TREE VIEW SIDEBAR STYLING MATCHING REFERENCE */
.tree-resource-list {
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.tree-group {
  display: flex;
  flex-direction: column;
}

.tree-group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.45rem 0.6rem;
  border-radius: 6px;
  cursor: pointer;
  user-select: none;
  transition: background 0.15s ease;
}

.tree-group-header:hover {
  background: var(--bg-subtle);
}

.header-main-side {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  color: var(--text-main);
}

.chevron-arrow {
  color: var(--text-mute);
  transition: transform 0.2s ease;
}

.chevron-arrow.collapsed {
  transform: rotate(-90deg);
}

.group-type-icon {
  color: var(--text-mute);
  flex-shrink: 0;
}

.group-title {
  font-size: 0.88rem;
  font-weight: 800;
  color: var(--text-main);
  letter-spacing: -0.01em;
}

.group-count {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-mute);
}

.tree-group-children {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding-left: 0.75rem;
  margin-top: 0.15rem;
}

.tree-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.45rem 0.65rem;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  color: var(--text-dim);
}

.tree-item:hover {
  background: var(--bg-subtle);
  color: var(--text-main);
}

.tree-item.active {
  background: rgba(16, 185, 129, 0.12);
  color: var(--success);
  font-weight: 700;
}

.tree-item-icon-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.item-icon {
  color: #0284c7;
}

[data-theme="dark"] .item-icon {
  color: #38bdf8;
}

.svc-icon {
  color: #7c3aed;
}

[data-theme="dark"] .svc-icon {
  color: #a78bfa;
}

.mini-status-dot {
  position: absolute;
  top: -1px;
  right: -2px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  border: 1.5px solid var(--bg-sidebar);
  background: var(--text-mute);
}

.mini-status-dot.running {
  background: #10b981;
}

.mini-status-dot.partial {
  background: #f59e0b;
}

.mini-status-dot.exited,
.mini-status-dot.stopped {
  background: #64748b;
}

.tree-item-name {
  font-size: 0.82rem;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--text-main);
}

.mini-tag {
  font-size: 0.58rem;
  font-weight: 800;
  text-transform: uppercase;
  padding: 0.05rem 0.25rem;
  border-radius: 3px;
  background: rgba(167, 139, 250, 0.2);
  color: #7c3aed;
}

[data-theme="dark"] .mini-tag {
  color: #c084fc;
}

</style>

<template>
  <div class="k8s-view animate-fade-in">
    <div v-if="!sharedState.k8sAvailable" class="k8s-unavailable-banner">
      <strong>Kubernetes is not connected.</strong>
      <p>
        {{
          sharedState.k8sError ||
          "Mount a kubeconfig or deploy DockLog in-cluster with a ServiceAccount."
        }}
      </p>
    </div>
    <div v-else-if="fetchError" class="k8s-unavailable-banner">
      <strong>Could not load Kubernetes data.</strong>
      <p>{{ fetchError }}</p>
    </div>

    <section class="k8s-hero">
      <div class="hero-copy">
        <span class="hero-eyebrow">Cluster control plane</span>
        <h1>Kubernetes</h1>
        <p class="hero-sub">
          Pods, deployments, HPAs, services, and events across your configured
          namespaces.
        </p>
      </div>
      <div class="hero-stats" v-if="overview">
        <div class="hero-stat">
          <span class="hero-stat-val">{{ overview.pods }}</span>
          <span class="hero-stat-lbl">Pods</span>
        </div>
        <div class="hero-stat success">
          <span class="hero-stat-val">{{ overview.running_pods }}</span>
          <span class="hero-stat-lbl">Running</span>
        </div>
        <div class="hero-stat">
          <span class="hero-stat-val">{{ overview.deployments }}</span>
          <span class="hero-stat-lbl">Deployments</span>
        </div>
        <div class="hero-stat">
          <span class="hero-stat-val">{{ overview.hpas }}</span>
          <span class="hero-stat-lbl">HPAs</span>
        </div>
        <div class="hero-stat warning" v-if="overview.warning_events > 0">
          <span class="hero-stat-val">{{ overview.warning_events }}</span>
          <span class="hero-stat-lbl">Warnings</span>
        </div>
      </div>
      <div class="hero-mesh" aria-hidden="true"></div>
    </section>

    <section class="k8s-panel">
      <div class="panel-toolbar">
        <div class="toolbar-left">
          <div class="namespace-select">
            <label for="namespace">Namespace</label>
            <select
              id="namespace"
              v-model="selectedNamespace"
              :disabled="namespacesLoading || namespaces.length === 0"
            >
              <option v-if="namespacesLoading" value="">Loading...</option>
              <option v-else-if="namespaces.length === 0" value="">
                No namespaces
              </option>
              <option v-for="ns in namespaces" :key="ns" :value="ns">
                {{ ns }}
              </option>
            </select>
          </div>
          <div class="tab-pills">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              class="filter-pill"
              :class="{ active: activeTab === tab.id }"
              @click="setTab(tab.id)"
            >
              {{ tab.label }}
              <span class="pill-count">{{ tab.count }}</span>
            </button>
          </div>
        </div>
        <div class="toolbar-right">
          <div class="search-box">
            <svg
              viewBox="0 0 24 24"
              width="18"
              height="18"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <input
              type="text"
              v-model="sharedState.searchQuery"
              placeholder="Filter resources..."
            />
          </div>
          <button class="refresh-btn" @click="refresh" :disabled="loading">
            Refresh
          </button>
        </div>
      </div>

      <div v-if="activeTab === 'overview' && overview" class="overview-body">
        <div class="overview-intro">
          <div class="overview-intro-copy">
            <span class="overview-eyebrow">
              <span class="live-dot" aria-hidden="true"></span>
              Namespace {{ selectedNamespace }}
            </span>
            <p class="overview-desc">Live snapshot for this namespace. Click a card to open that view.</p>
          </div>
          <div class="overview-chip-row">
            <span class="overview-chip">{{ overview.pods }} pods</span>
            <span class="overview-chip success">{{ overview.running_pods }} running</span>
            <span v-if="overview.warning_events > 0" class="overview-chip warning">
              {{ overview.warning_events }} warnings
            </span>
          </div>
        </div>

        <div class="metrics-bento">
          <button type="button" class="metric-card variant-pods" @click="setTab('pods')">
            <div class="metric-glow" aria-hidden="true"></div>
            <div class="metric-top">
              <div class="metric-icon success">
                <AppIcon name="containers" />
              </div>
              <span class="metric-badge success">Pods</span>
            </div>
            <div class="metric-body">
              <div class="metric-main">
                <span class="metric-value">{{ overview.pods }}</span>
                <span class="metric-label">Total pods</span>
              </div>
              <div class="metric-visual">
                <div class="donut" :style="{ '--pct': podRunningRatio }">
                  <svg viewBox="0 0 36 36">
                    <circle class="donut-track" cx="18" cy="18" r="15.5" />
                    <circle class="donut-fill success" cx="18" cy="18" r="15.5" />
                  </svg>
                  <span class="donut-label">{{ podRunningRatio }}%</span>
                </div>
              </div>
            </div>
            <div class="metric-footer">
              <div class="metric-bar">
                <div class="metric-bar-fill success" :style="{ width: `${podRunningRatio}%` }"></div>
              </div>
              <span>{{ runningCount }} running · {{ pendingCount }} pending · {{ failedCount }} failed</span>
            </div>
          </button>

          <button type="button" class="metric-card variant-deploy" @click="setTab('deployments')">
            <div class="metric-glow" aria-hidden="true"></div>
            <div class="metric-top">
              <div class="metric-icon">
                <AppIcon name="box" />
              </div>
              <span class="metric-badge">Workloads</span>
            </div>
            <div class="metric-body">
              <div class="metric-main">
                <span class="metric-value">{{ overview.deployments }}</span>
                <span class="metric-label">Deployments</span>
              </div>
              <div class="metric-visual">
                <div class="pulse-stack">
                  <span
                    v-for="n in 4"
                    :key="n"
                    class="pulse-bar"
                    :style="{ height: deployPulseHeights[n - 1] + '%', animationDelay: n * 0.12 + 's' }"
                  ></span>
                </div>
              </div>
            </div>
            <div class="metric-footer">{{ deploymentReadySummary }}</div>
          </button>

          <button type="button" class="metric-card variant-services" @click="setTab('services')">
            <div class="metric-glow" aria-hidden="true"></div>
            <div class="metric-top">
              <div class="metric-icon">
                <AppIcon name="server" />
              </div>
              <span class="metric-badge">Network</span>
            </div>
            <div class="metric-body">
              <div class="metric-main">
                <span class="metric-value">{{ overview.services }}</span>
                <span class="metric-label">Services</span>
              </div>
              <div class="metric-visual metric-stat-pill">{{ overview.hpas }} HPAs</div>
            </div>
            <div class="metric-footer">{{ serviceFooterLabel }}</div>
          </button>

          <button
            type="button"
            class="metric-card variant-events"
            :class="{ 'has-warnings': overview.warning_events > 0 }"
            @click="setTab('events')"
          >
            <div class="metric-glow" aria-hidden="true"></div>
            <div class="metric-top">
              <div class="metric-icon" :class="overview.warning_events > 0 ? 'warning' : 'muted'">
                <AppIcon name="bell" />
              </div>
              <span class="metric-badge" :class="overview.warning_events > 0 ? 'warning' : 'dim'">Events</span>
            </div>
            <div class="metric-body">
              <div class="metric-main">
                <span class="metric-value">{{ events.length }}</span>
                <span class="metric-label">Recent events</span>
              </div>
              <div class="metric-visual">
                <span class="warning-count" :class="{ active: overview.warning_events > 0 }">
                  {{ overview.warning_events }}
                </span>
              </div>
            </div>
            <div class="metric-footer" :class="{ warn: overview.warning_events > 0 }">
              {{ overview.warning_events > 0 ? `${overview.warning_events} warnings need review` : 'No warnings in recent events' }}
            </div>
          </button>
        </div>

        <div class="overview-lower">
          <section class="overview-panel warnings-panel">
            <div class="panel-head">
              <div>
                <h3>Recent warnings</h3>
                <p>Latest Warning events in {{ selectedNamespace }}</p>
              </div>
              <button type="button" class="panel-link" @click="setTab('events')">View all</button>
            </div>
            <div v-if="recentWarnings.length" class="warning-list">
              <button
                v-for="event in recentWarnings"
                :key="`${event.name}-${event.last_timestamp}`"
                type="button"
                class="warning-item"
                @click="setTab('events')"
              >
                <span class="warning-reason">{{ event.reason }}</span>
                <span class="warning-target mono">{{ event.involved_kind }}/{{ event.involved_name }}</span>
                <span class="warning-message">{{ event.message }}</span>
                <span class="warning-time">{{ formatDate(event.last_timestamp) }}</span>
              </button>
            </div>
            <p v-else class="panel-empty">No warning events in this namespace.</p>
          </section>

          <section class="overview-panel quick-panel">
            <div class="panel-head">
              <div>
                <h3>Quick access</h3>
                <p>Jump to common views</p>
              </div>
            </div>
            <div class="quick-links">
              <button type="button" class="quick-link" @click="setTab('pods')">
                <AppIcon name="containers" :size="16" />
                <span>Pods</span>
                <strong>{{ overview.pods }}</strong>
              </button>
              <button type="button" class="quick-link" @click="setTab('deployments')">
                <AppIcon name="box" :size="16" />
                <span>Deployments</span>
                <strong>{{ overview.deployments }}</strong>
              </button>
              <button type="button" class="quick-link" @click="setTab('hpas')">
                <AppIcon name="activity" :size="16" />
                <span>HPAs</span>
                <strong>{{ overview.hpas }}</strong>
              </button>
              <button type="button" class="quick-link" @click="setTab('services')">
                <AppIcon name="server" :size="16" />
                <span>Services</span>
                <strong>{{ overview.services }}</strong>
              </button>
            </div>
          </section>
        </div>
      </div>

      <div v-else-if="activeTab === 'pods'" class="filter-pills pod-filters">
        <button
          v-for="f in podFilters"
          :key="f.value"
          class="filter-pill"
          :class="{ active: phaseFilter === f.value }"
          @click="phaseFilter = f.value"
        >
          {{ f.label }} <span class="pill-count">{{ f.count }}</span>
        </button>
      </div>

      <div v-if="activeTab !== 'overview'" class="premium-table-container embedded">
        <table v-if="activeTab === 'pods'" class="premium-table containers-table">
          <thead>
            <tr>
              <th>Pod</th>
              <th>Images</th>
              <th>Created</th>
              <th>Status</th>
              <th>State</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr>
              <td colspan="6">
                <div class="table-loading"><div class="shimmer"></div></div>
              </td>
            </tr>
          </tbody>
          <tbody v-else-if="filteredPods.length">
            <tr
              v-for="pod in filteredPods"
              :key="pod.uid || podKey(pod)"
              class="container-row"
              :class="{ 'is-running': isPodRunning(pod) }"
            >
              <td data-label="Pod">
                <div class="name-cell clickable" @click="goToPodDetail(pod)">
                  <div
                    class="container-avatar"
                    :class="isPodRunning(pod) ? 'running' : 'stopped'"
                  >
                    <svg
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <rect x="3" y="3" width="18" height="18" rx="3"></rect>
                      <path d="M9 9h6v6H9z"></path>
                    </svg>
                  </div>
                  <div class="name-main">
                    <span class="container-title">{{ pod.name }}</span>
                    <div class="pod-sub-row">
                      <span class="container-id">{{ pod.namespace }}</span>
                      <span v-if="pod.node" class="pod-node-badge">{{ pod.node }}</span>
                    </div>
                  </div>
                </div>
              </td>
              <td data-label="Images">
                <div class="image-cell" v-if="pod.images && pod.images.length">
                  <div
                    v-for="img in pod.images"
                    :key="img"
                    class="image-name-row"
                  >
                    <span class="image-name">{{ img.split(':')[0] }}</span>
                    <span class="image-tag">{{ img.split(':')[1] || 'latest' }}</span>
                  </div>
                </div>
                <span v-else class="text-mute">—</span>
              </td>
              <td data-label="Created">
                <span class="date-label">{{ formatDate(pod.created) }}</span>
              </td>
              <td data-label="Status">
                <span
                  :class="[
                    'uptime-label',
                    isPodRunning(pod) ? 'is-running' : 'is-stopped',
                  ]"
                >
                  {{ pod.ready }} Ready<span v-if="pod.restarts"> · {{ pod.restarts }} restarts</span>
                </span>
              </td>
              <td data-label="State">
                <div
                  :class="[
                    'status-pill',
                    phaseClass(pod.phase),
                  ]"
                >
                  <span class="pulse-dot"></span>
                  {{ (pod.phase || 'UNKNOWN').toUpperCase() }}
                </div>
              </td>
              <td class="text-right action-cell" data-label="Actions">
                <div class="action-group justify-end" @click.stop>
                  <div class="action-cluster primary-actions">
                    <button
                      v-if="
                        userCanStart(sharedState.currentUser) &&
                        !isPodRunning(pod)
                      "
                      class="icon-btn start"
                      @click="runPodAction(pod, 'start')"
                      data-tooltip="Start"
                    >
                      <AppIcon name="play" :size="16" :stroke-width="2.75" />
                    </button>
                    <button
                      v-if="
                        userCanStop(sharedState.currentUser) &&
                        isPodRunning(pod)
                      "
                      class="icon-btn stop"
                      @click="runPodAction(pod, 'stop')"
                      data-tooltip="Stop"
                    >
                      <AppIcon
                        name="stopOutline"
                        :size="16"
                        :stroke-width="2.75"
                      />
                    </button>
                    <button
                      v-if="userCanRestart(sharedState.currentUser)"
                      class="icon-btn restart"
                      @click="runPodAction(pod, 'restart')"
                      data-tooltip="Restart"
                    >
                      <AppIcon name="refresh" :size="16" :stroke-width="2.75" />
                    </button>
                  </div>
                  <div class="action-cluster secondary-actions">
                    <button
                      v-if="userCanShell(sharedState.currentUser)"
                      class="icon-btn shell"
                      @click="goToShell(pod)"
                      data-tooltip="Shell"
                    >
                      <AppIcon name="terminal" :size="16" />
                    </button>
                    <button
                      class="icon-btn logs"
                      @click="goToLogs(pod)"
                      data-tooltip="Logs"
                    >
                      <AppIcon name="logsBubble" :size="16" />
                    </button>
                    <button
                      v-if="userCanDelete(sharedState.currentUser)"
                      class="icon-btn delete"
                      @click="runPodAction(pod, 'remove')"
                      data-tooltip="Delete"
                    >
                      <AppIcon name="trash" :size="16" />
                    </button>
                  </div>
                </div>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr>
              <td colspan="6" class="empty-state">No pods found.</td>
            </tr>
          </tbody>
        </table>

        <table v-else-if="activeTab === 'deployments'" class="premium-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Replicas</th>
              <th>Strategy</th>
              <th>Images</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr>
              <td colspan="5">
                <div class="table-loading"><div class="shimmer"></div></div>
              </td>
            </tr>
          </tbody>
          <tbody v-else-if="filteredDeployments.length">
            <tr v-for="dep in filteredDeployments" :key="dep.uid || dep.name">
              <td>
                <strong>{{ dep.name }}</strong>
              </td>
              <td>
                {{ dep.ready }}/{{ dep.replicas }} ready ·
                {{ dep.available }} available
              </td>
              <td>{{ dep.strategy }}</td>
              <td>
                <span v-for="img in dep.images" :key="img" class="image-chip">{{
                  img
                }}</span>
              </td>
              <td>
                <span class="status-pill">{{ dep.status }}</span>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr>
              <td colspan="5" class="empty-state">No deployments found.</td>
            </tr>
          </tbody>
        </table>

        <table v-else-if="activeTab === 'hpas'" class="premium-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Target</th>
              <th>Replicas</th>
              <th>Metrics</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr>
              <td colspan="5">
                <div class="table-loading"><div class="shimmer"></div></div>
              </td>
            </tr>
          </tbody>
          <tbody v-else-if="filteredHPAs.length">
            <tr v-for="hpa in filteredHPAs" :key="hpa.uid || hpa.name">
              <td>
                <strong>{{ hpa.name }}</strong>
              </td>
              <td>{{ hpa.target_kind }}/{{ hpa.target_name }}</td>
              <td>
                {{ hpa.current_replicas }} → {{ hpa.desired_replicas }}
                <span class="sub-text"
                  >({{ hpa.min_replicas }}-{{ hpa.max_replicas }})</span
                >
              </td>
              <td>{{ hpa.metrics }}</td>
              <td>
                <span
                  :class="[
                    'status-pill',
                    hpa.status === 'ScalingLimited'
                      ? 'is-pending'
                      : 'is-running',
                  ]"
                  >{{ hpa.status }}</span
                >
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr>
              <td colspan="5" class="empty-state">
                No HPAs found. Install metrics-server and create a
                HorizontalPodAutoscaler to see it here.
              </td>
            </tr>
          </tbody>
        </table>

        <table v-else-if="activeTab === 'services'" class="premium-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Type</th>
              <th>Cluster IP</th>
              <th>Ports</th>
              <th>Selector</th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr>
              <td colspan="5">
                <div class="table-loading"><div class="shimmer"></div></div>
              </td>
            </tr>
          </tbody>
          <tbody v-else-if="filteredServices.length">
            <tr v-for="svc in filteredServices" :key="svc.uid || svc.name">
              <td>
                <strong>{{ svc.name }}</strong>
              </td>
              <td>{{ svc.type }}</td>
              <td>{{ svc.cluster_ip }}</td>
              <td>{{ svc.ports }}</td>
              <td class="sub-text">{{ svc.selector || "—" }}</td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr>
              <td colspan="5" class="empty-state">No services found.</td>
            </tr>
          </tbody>
        </table>

        <table v-else-if="activeTab === 'events'" class="premium-table">
          <thead>
            <tr>
              <th>Type</th>
              <th>Reason</th>
              <th>Object</th>
              <th>Message</th>
              <th>Count</th>
              <th>Last seen</th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr>
              <td colspan="6">
                <div class="table-loading"><div class="shimmer"></div></div>
              </td>
            </tr>
          </tbody>
          <tbody v-else-if="filteredEvents.length">
            <tr v-for="event in filteredEvents" :key="event.name">
              <td>
                <span
                  :class="[
                    'status-pill',
                    eventTypeClass(event.type),
                  ]"
                  >{{ event.type }}</span
                >
              </td>
              <td>{{ event.reason }}</td>
              <td>{{ event.involved_kind }}/{{ event.involved_name }}</td>
              <td class="message-cell">{{ event.message }}</td>
              <td>{{ event.count }}</td>
              <td>{{ formatDate(event.last_timestamp) }}</td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr>
              <td colspan="6" class="empty-state">No recent events.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  useKubernetes,
  pods,
  deployments,
  hpas,
  services,
  events,
} from "../composables/useKubernetes";
import AppIcon from "../components/AppIcon.vue";
import {
  sharedState,
  showToast,
  userCanStart,
  userCanStop,
  userCanRestart,
  userCanDelete,
  userCanShell,
} from "../utils/sharedState";
import { secureStorage } from "../utils/storage";
import { apiFetch } from "../utils/apiFetch";
import { logsRouteForPod } from "../utils/logRoutes";

const route = useRoute();
const router = useRouter();

const {
  overview,
  namespaces,
  loading,
  namespacesLoading,
  selectedNamespace,
  fetchError,
  activeTab,
  phaseFilter,
  filteredPods,
  filteredDeployments,
  filteredHPAs,
  filteredServices,
  filteredEvents,
  runningCount,
  pendingCount,
  failedCount,
  refresh,
  formatDate,
  podKey,
} = useKubernetes();

const tabs = computed(() => [
  { id: "overview", label: "Overview", count: "·" },
  { id: "pods", label: "Pods", count: pods.value.length },
  { id: "deployments", label: "Deployments", count: deployments.value.length },
  { id: "hpas", label: "HPAs", count: hpas.value.length },
  { id: "services", label: "Services", count: services.value.length },
  { id: "events", label: "Events", count: events.value.length },
]);

const podFilters = computed(() => [
  { label: "All", value: "all", count: pods.value.length },
  { label: "Running", value: "running", count: runningCount.value },
  { label: "Pending", value: "pending", count: pendingCount.value },
  { label: "Failed", value: "failed", count: failedCount.value },
]);

const podRunningRatio = computed(() => {
  if (!overview.value?.pods) return 0;
  return Math.round((overview.value.running_pods / overview.value.pods) * 100);
});

const deploymentReadySummary = computed(() => {
  if (!deployments.value.length) return "No deployments in namespace";
  const ready = deployments.value.reduce((sum, dep) => sum + (dep.ready || 0), 0);
  const total = deployments.value.reduce((sum, dep) => sum + (dep.replicas || 0), 0);
  return `${ready}/${total} replicas ready`;
});

const deployPulseHeights = computed(() => {
  if (!deployments.value.length) return [28, 36, 32, 24];
  const ready = deployments.value.reduce((sum, dep) => sum + (dep.ready || 0), 0);
  const total = deployments.value.reduce((sum, dep) => sum + (dep.replicas || 0), 0);
  const ratio = total ? Math.round((ready / total) * 100) : 0;
  const base = Math.max(30, Math.min(95, ratio));
  return [base * 0.55, base * 0.85, base, base * 0.7];
});

const serviceFooterLabel = computed(() => {
  if (!services.value.length) return "No services exposed";
  const types = [...new Set(services.value.map((svc) => svc.type).filter(Boolean))];
  return types.length ? types.join(" · ") : "Cluster networking active";
});

const recentWarnings = computed(() =>
  events.value.filter((event) => event.type === "Warning").slice(0, 5),
);

const setTab = (tab) => {
  activeTab.value = tab;
  router.replace({ query: { ...route.query, tab } });
};

watch(
  () => route.query.tab,
  (tab) => {
    if (typeof tab === "string" && tab) activeTab.value = tab;
  },
  { immediate: true },
);

watch(
  () => route.query.namespace,
  (namespace) => {
    if (typeof namespace !== "string" || !namespace) {
      selectedNamespace.value = "";
      return;
    }
    if (selectedNamespace.value !== namespace) {
      selectedNamespace.value = namespace;
    }
  },
  { immediate: true },
);

watch(selectedNamespace, (namespace) => {
  if (!namespace || route.query.namespace === namespace) return;
  router.replace({ query: { ...route.query, namespace } });
});

const phaseClass = (phase) => {
  const normalized = (phase || "").toLowerCase();
  if (normalized === "running") return "is-running";
  if (normalized === "pending") return "is-pending";
  if (normalized === "failed" || normalized === "unknown") return "is-failed";
  return "is-stopped";
};

const eventTypeClass = (type) => {
  const normalized = String(type || "").toLowerCase();
  if (normalized === "warning") return "is-warning";
  if (normalized === "normal") return "is-running";
  return "is-neutral";
};

const goToLogs = (pod) => {
  router.push(logsRouteForPod(pod));
};

const goToPodDetail = (pod) => {
  router.push({
    path: `/kubernetes/pods/${encodeURIComponent(pod.namespace)}/${encodeURIComponent(pod.name)}`,
    query: { namespace: pod.namespace, tab: "pods" },
  });
};

const goToShell = (pod) => {
  router.push({ path: "/shell", query: { p: `${pod.namespace}/${pod.name}` } });
};

const goToHistory = (pod) => {
  router.push(logsRouteForPod(pod));
};

const isPodRunning = (pod) =>
  String(pod.phase || "").toLowerCase() === "running";

const runPodAction = async (pod, action) => {
  const confirmed = window.confirm(
    `Are you sure you want to ${action} pod ${pod.name}?`,
  );
  if (!confirmed) return;
  try {
    const token = secureStorage.getItem("token");
    const formData = new URLSearchParams();
    formData.set("action", action);
    const res = await apiFetch(
      `/api/namespaces/${encodeURIComponent(pod.namespace)}/pods/${encodeURIComponent(pod.name)}/action`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: formData.toString(),
      },
    );
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || `Failed to ${action} pod`);
    }
    showToast("Success", `Pod ${action} requested.`, "success");
    await refresh();
  } catch (err) {
    showToast(
      "Action failed",
      err.message || `Failed to ${action} pod`,
      "error",
    );
  }
};
</script>

<style scoped>
.k8s-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-bottom: 2rem;
}
.k8s-unavailable-banner {
  padding: 1rem 1.25rem;
  border-radius: var(--radius-lg);
  border: 1px solid rgba(var(--warning-rgb), 0.35);
  background: rgba(var(--warning-rgb), 0.08);
}
.k8s-unavailable-banner strong {
  display: block;
  margin-bottom: 0.35rem;
  color: var(--warning);
}
.k8s-unavailable-banner p {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-dim);
}
.k8s-hero,
.k8s-panel {
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border);
  background: var(--bg-card);
}
.k8s-hero {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1.5rem;
  flex-wrap: wrap;
  padding: 1.5rem 1.75rem;
  overflow: hidden;
}
.hero-mesh {
  position: absolute;
  inset: 0;
  background: radial-gradient(
    ellipse 70% 80% at 100% 0%,
    rgba(var(--accent-rgb), 0.14),
    transparent 55%
  );
  pointer-events: none;
}
.hero-copy {
  position: relative;
  z-index: 1;
}
.hero-eyebrow {
  display: block;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--accent);
  margin-bottom: 0.4rem;
}
.k8s-hero h1 {
  margin: 0 0 0.35rem;
  font-size: clamp(1.35rem, 2.5vw, 1.75rem);
  font-weight: 800;
}
.hero-sub {
  margin: 0;
  color: var(--text-dim);
  max-width: 560px;
}
.hero-stats {
  position: relative;
  z-index: 1;
  display: flex;
  gap: 0.65rem;
  flex-wrap: wrap;
}
.hero-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 72px;
  padding: 0.65rem 1rem;
  border-radius: var(--radius-md);
  background: var(--bg-input);
  border: 1px solid var(--border);
}
.hero-stat-val {
  font-size: 1.35rem;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}
.hero-stat-lbl {
  margin-top: 0.2rem;
  font-size: 0.62rem;
  font-weight: 800;
  text-transform: uppercase;
  color: var(--text-mute);
}
.hero-stat.success .hero-stat-val {
  color: var(--success);
}
.hero-stat.warning .hero-stat-val {
  color: var(--warning);
}
.k8s-panel {
  padding: 1.15rem 1.15rem 0.35rem;
}
.panel-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}
.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: flex-end;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.namespace-select {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex-shrink: 0;
}
.namespace-select label {
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-mute);
  line-height: 1;
}
.namespace-select select {
  min-width: 180px;
  padding: 0.6rem 0.75rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-main);
  font-size: 0.85rem;
  font-weight: 600;
  min-height: calc(0.6rem * 2 + 1.35rem + 2px);
  box-sizing: border-box;
}
.tab-pills,
.filter-pills {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem;
  border-radius: var(--radius-md);
  background: var(--bg-input);
  border: 1px solid var(--border);
  flex-wrap: wrap;
  min-height: calc(0.6rem * 2 + 1.35rem + 2px);
  box-sizing: border-box;
}
.pod-filters {
  margin-bottom: 1rem;
}
.filter-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.8rem;
  border-radius: calc(var(--radius-md) - 2px);
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-dim);
}
.filter-pill.active {
  background: var(--accent);
  color: #fff;
}
.pill-count {
  font-size: 0.65rem;
  padding: 0.1rem 0.35rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.15);
}
.filter-pill:not(.active) .pill-count {
  background: var(--bg-subtle);
  color: var(--text-mute);
}
.search-box {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 1rem;
  border-radius: var(--radius-md);
  background: var(--bg-input);
  border: 1px solid var(--border);
  min-width: 260px;
  min-height: calc(0.6rem * 2 + 1.35rem + 2px);
  box-sizing: border-box;
}
.search-box input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  width: 100%;
}
.refresh-btn {
  padding: 0.6rem 1rem;
  border-radius: var(--radius-md);
  background: var(--accent);
  color: #fff;
  font-weight: 700;
  font-size: 0.85rem;
  border: none;
  cursor: pointer;
  min-height: calc(0.6rem * 2 + 1.35rem + 2px);
  box-sizing: border-box;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.overview-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-bottom: 1rem;
}

.overview-intro {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1rem;
  flex-wrap: wrap;
  padding: 0.15rem 0.15rem 0;
}

.overview-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--accent);
  margin-bottom: 0.35rem;
}

.live-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--success);
  box-shadow: 0 0 0 3px rgba(var(--success-rgb), 0.25);
  animation: livePulse 2s ease infinite;
}

@keyframes livePulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.7; transform: scale(0.92); }
}

.overview-desc {
  margin: 0;
  font-size: 0.86rem;
  color: var(--text-dim);
}

.overview-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
}

.overview-chip {
  padding: 0.3rem 0.65rem;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--text-dim);
}

.overview-chip.success {
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.3);
  background: rgba(var(--success-rgb), 0.08);
}

.overview-chip.warning {
  color: var(--warning);
  border-color: rgba(var(--warning-rgb), 0.35);
  background: rgba(var(--warning-rgb), 0.08);
}

.metrics-bento {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.metric-card {
  position: relative;
  padding: 0.9rem 1rem 0.8rem;
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  min-height: 138px;
  text-align: left;
  cursor: pointer;
  overflow: hidden;
  transition: transform 0.22s ease, border-color 0.22s ease, box-shadow 0.22s ease;
}

.metric-card:hover {
  transform: translateY(-2px);
  border-color: rgba(var(--accent-rgb), 0.35);
  box-shadow: 0 14px 28px -16px var(--shadow);
}

.metric-card.has-warnings {
  border-color: rgba(var(--warning-rgb), 0.35);
}

.metric-glow {
  position: absolute;
  width: 88px;
  height: 88px;
  border-radius: 50%;
  top: -32px;
  right: -24px;
  pointer-events: none;
  opacity: 0.45;
  filter: blur(22px);
}

.variant-pods .metric-glow { background: rgba(var(--success-rgb), 0.22); }
.variant-deploy .metric-glow { background: rgba(var(--accent-rgb), 0.2); }
.variant-services .metric-glow { background: rgba(var(--accent-rgb), 0.16); }
.variant-events .metric-glow { background: rgba(var(--warning-rgb), 0.2); }

.metric-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.55rem;
  position: relative;
  z-index: 1;
}

.metric-icon {
  width: 32px;
  height: 32px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--accent-soft);
  color: var(--accent);
  border: 1px solid rgba(var(--accent-rgb), 0.12);
}

.metric-icon.success {
  background: rgba(var(--success-rgb), 0.1);
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.15);
}

.metric-icon.warning {
  background: rgba(var(--warning-rgb), 0.1);
  color: var(--warning);
  border-color: rgba(var(--warning-rgb), 0.15);
}

.metric-icon.muted {
  background: var(--bg-subtle);
  color: var(--text-mute);
  border-color: var(--border-subtle);
}

.metric-badge {
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.07em;
  text-transform: uppercase;
  padding: 0.2rem 0.5rem;
  border-radius: 999px;
  background: var(--accent-soft);
  color: var(--accent);
  border: 1px solid rgba(var(--accent-rgb), 0.12);
}

.metric-badge.success {
  background: rgba(var(--success-rgb), 0.1);
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.15);
}

.metric-badge.warning {
  background: rgba(var(--warning-rgb), 0.1);
  color: var(--warning);
  border-color: rgba(var(--warning-rgb), 0.15);
}

.metric-badge.dim {
  background: var(--bg-subtle);
  color: var(--text-mute);
  border-color: var(--border-subtle);
}

.metric-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  flex: 1;
  position: relative;
  z-index: 1;
}

.metric-main {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.metric-value {
  font-size: 1.65rem;
  font-weight: 800;
  letter-spacing: -0.04em;
  color: var(--text-main);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.metric-label {
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--text-dim);
}

.metric-visual {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.metric-stat-pill {
  padding: 0.35rem 0.55rem;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-dim);
}

.warning-count {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  font-weight: 800;
  color: var(--text-mute);
  background: var(--bg-input);
  border: 1px solid var(--border);
}

.warning-count.active {
  color: var(--warning);
  border-color: rgba(var(--warning-rgb), 0.35);
  background: rgba(var(--warning-rgb), 0.1);
}

.metric-footer {
  margin-top: auto;
  padding-top: 0.5rem;
  font-size: 0.68rem;
  font-weight: 600;
  color: var(--text-mute);
  position: relative;
  z-index: 1;
}

.metric-footer.warn {
  color: var(--warning);
}

.metric-bar {
  height: 3px;
  border-radius: 999px;
  overflow: hidden;
  background: var(--bg-subtle);
  margin-bottom: 0.35rem;
}

.metric-bar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.6s cubic-bezier(0.23, 1, 0.32, 1);
}

.metric-bar-fill.success {
  background: linear-gradient(90deg, var(--success), rgba(var(--success-rgb), 0.55));
}

.donut {
  --pct: 0;
  position: relative;
  width: 44px;
  height: 44px;
}

.donut svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.donut-track {
  fill: none;
  stroke: var(--bg-subtle);
  stroke-width: 3;
}

.donut-fill {
  fill: none;
  stroke-width: 3;
  stroke-linecap: round;
  stroke-dasharray: calc(var(--pct) * 0.974) 97.4;
  transition: stroke-dasharray 0.6s cubic-bezier(0.23, 1, 0.32, 1);
}

.donut-fill.success { stroke: var(--success); }

.donut-label {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.6rem;
  font-weight: 800;
  color: var(--text-dim);
  font-variant-numeric: tabular-nums;
}

.pulse-stack {
  display: flex;
  align-items: flex-end;
  gap: 3px;
  height: 36px;
  width: 36px;
}

.pulse-bar {
  flex: 1;
  border-radius: 4px 4px 2px 2px;
  background: linear-gradient(180deg, var(--accent), rgba(var(--accent-rgb), 0.45));
  animation: pulseBar 1.6s ease-in-out infinite;
  min-height: 18%;
}

@keyframes pulseBar {
  0%, 100% { transform: scaleY(1); opacity: 0.85; }
  50% { transform: scaleY(0.72); opacity: 1; }
}

.overview-lower {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(260px, 0.8fr);
  gap: 0.75rem;
}

.overview-panel {
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  padding: 1rem;
}

.overview-panel .panel-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.75rem;
  margin-bottom: 0.85rem;
}

.overview-panel .panel-head h3 {
  margin: 0 0 0.2rem;
  font-size: 0.92rem;
  font-weight: 800;
}

.overview-panel .panel-head p {
  margin: 0;
  font-size: 0.78rem;
  color: var(--text-mute);
}

.panel-link {
  border: none;
  background: transparent;
  color: var(--accent);
  font-size: 0.78rem;
  font-weight: 700;
  cursor: pointer;
}

.panel-link:hover {
  text-decoration: underline;
}

.panel-empty {
  margin: 0;
  font-size: 0.84rem;
  color: var(--text-mute);
}

.warning-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 260px;
  overflow: auto;
}

.warning-item {
  display: grid;
  grid-template-columns: minmax(100px, auto) 1fr;
  gap: 0.25rem 0.75rem;
  padding: 0.65rem 0.75rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  background: var(--bg-input);
  text-align: left;
  cursor: pointer;
  transition: border-color 0.2s ease, background 0.2s ease;
}

.warning-item:hover {
  border-color: rgba(var(--warning-rgb), 0.35);
  background: rgba(var(--warning-rgb), 0.06);
}

.warning-reason {
  font-size: 0.78rem;
  font-weight: 800;
  color: var(--warning);
}

.warning-target {
  font-size: 0.72rem;
  color: var(--text-dim);
  justify-self: end;
}

.warning-message {
  grid-column: 1 / -1;
  font-size: 0.8rem;
  color: var(--text-main);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.warning-time {
  grid-column: 1 / -1;
  font-size: 0.68rem;
  color: var(--text-mute);
}

.quick-links {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.55rem;
}

.quick-link {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.7rem 0.75rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  background: var(--bg-input);
  color: var(--text-dim);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.quick-link strong {
  margin-left: auto;
  color: var(--text-main);
  font-size: 0.9rem;
}

.quick-link:hover {
  border-color: rgba(var(--accent-rgb), 0.35);
  color: var(--accent);
  transform: translateY(-1px);
}
.sub-text {
  font-size: 0.75rem;
  color: var(--text-mute);
}
.image-chip {
  display: inline-block;
  max-width: 240px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 0.25rem;
  padding: 0.15rem 0.45rem;
  border-radius: 999px;
  background: var(--bg-subtle);
  font-size: 0.75rem;
}
.status-pill {
  display: inline-flex;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  font-size: 0.68rem;
  font-weight: 800;
  background: var(--bg-input);
  border: 1px solid var(--border);
}
.status-pill.is-running {
  color: var(--success);
}
.status-pill.is-pending {
  color: var(--warning);
}
.status-pill.is-warning {
  color: var(--warning);
  border-color: rgba(var(--warning-rgb), 0.35);
  background: rgba(var(--warning-rgb), 0.08);
}
.status-pill.is-failed {
  color: var(--error);
}
.status-pill.is-neutral {
  color: var(--text-mute);
}
.action-cell {
  min-width: 330px;
}
.action-group {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.45rem;
}
.action-cluster {
  display: flex;
  gap: 0.35rem;
  padding: 0.25rem;
  border-radius: var(--radius-md);
  background: var(--bg-subtle);
  border: 1px solid var(--border-subtle);
}
.secondary-actions {
  background: transparent;
  border-color: transparent;
  padding: 0.25rem 0;
}
.icon-btn {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  border: 1px solid transparent;
  background: var(--bg-input);
  color: var(--text-dim);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.icon-btn:hover {
  transform: translateY(-1px);
  color: var(--text-main);
  border-color: var(--border);
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.08);
}
.icon-btn.start:hover {
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.35);
  background: rgba(var(--success-rgb), 0.08);
}
.icon-btn.stop:hover {
  color: var(--warning);
  border-color: rgba(var(--warning-rgb), 0.35);
  background: rgba(var(--warning-rgb), 0.08);
}
.icon-btn.restart:hover,
.icon-btn.logs:hover,
.icon-btn.history:hover,
.icon-btn.details:hover {
  color: var(--accent);
  border-color: rgba(var(--accent-rgb), 0.35);
  background: var(--accent-soft);
}
.icon-btn.shell:hover,
.icon-btn.shell:focus-visible {
  color: #8b5cf6;
  border-color: rgba(139, 92, 246, 0.4);
  background: rgba(139, 92, 246, 0.1);
  box-shadow: 0 4px 14px rgba(139, 92, 246, 0.15);
}
.icon-btn.delete:hover {
  color: var(--error);
  border-color: rgba(var(--error-rgb), 0.35);
  background: rgba(var(--error-rgb), 0.08);
}
.pod-link {
  padding: 0;
  background: none;
  border: none;
  color: var(--text-main);
  font-weight: 800;
  cursor: pointer;
  text-align: left;
}
.pod-link:hover {
  color: var(--accent);
  text-decoration: underline;
}

/* Matching ContainerTable UI */
.containers-table tr.container-row {
  transition: background 0.2s ease;
}

.containers-table tr.container-row td:first-child {
  position: relative;
}

.containers-table tr.container-row:hover td {
  background: var(--card-hover);
}

.containers-table tr.container-row.is-running:hover td:first-child::before {
  content: "";
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 0 4px 4px 0;
  background: var(--success);
}

.containers-table tr.container-row:not(.is-running):hover td:first-child::before {
  content: "";
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 0 4px 4px 0;
  background: var(--text-mute);
  opacity: 0.5;
}

.name-cell {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  min-width: 0;
}

.name-cell.clickable {
  cursor: pointer;
}

.container-avatar {
  width: 38px;
  height: 38px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 1px solid var(--border);
}

.container-avatar svg {
  width: 18px;
  height: 18px;
}

.container-avatar.running {
  background: rgba(var(--success-rgb), 0.1);
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.2);
}

.container-avatar.stopped {
  background: var(--bg-subtle);
  color: var(--text-mute);
}

.name-main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.container-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 220px;
}

.name-cell.clickable:hover .container-title {
  color: var(--accent);
}

.container-id {
  font-family: var(--font-mono);
  font-size: 0.72rem;
  color: var(--text-mute);
}

.pod-sub-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-top: 0.1rem;
}

.pod-node-badge {
  font-size: 0.65rem;
  padding: 0.05rem 0.35rem;
  border-radius: 4px;
  background: var(--bg-subtle);
  border: 1px solid var(--border-subtle);
  color: var(--text-dim);
}

.image-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.image-name-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.image-name {
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 160px;
}

.image-tag {
  font-family: var(--font-mono);
  font-size: 0.68rem;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-dim);
}

.date-label {
  font-size: 0.8rem;
  color: var(--text-dim);
  white-space: nowrap;
}

.uptime-label {
  display: inline-block;
  font-size: 0.76rem;
  font-weight: 600;
  white-space: nowrap;
}

.uptime-label.is-running {
  color: var(--success);
}

.uptime-label.is-stopped {
  color: var(--text-dim) !important;
}

.pulse-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  padding: 0.35rem 0.7rem;
  border-radius: 999px;
  width: fit-content;
  border: 1px solid transparent;
}

.status-pill.is-running {
  color: var(--success);
  background: rgba(var(--success-rgb), 0.1);
  border-color: rgba(var(--success-rgb), 0.2);
}

.status-pill.is-running .pulse-dot {
  background: var(--success);
  box-shadow: 0 0 6px rgba(var(--success-rgb), 0.6);
  animation: pulse-mini 2s infinite;
}

.status-pill.is-stopped {
  color: var(--text-mute);
  background: var(--bg-subtle);
  border-color: var(--border-subtle);
}

.status-pill.is-stopped .pulse-dot {
  background: var(--text-mute);
}

.status-pill.is-pending .pulse-dot {
  background: var(--warning);
}

.status-pill.is-failed .pulse-dot {
  background: var(--error);
}

@keyframes pulse-mini {
  0%, 100% { transform: scale(0.95); opacity: 1; }
  50% { transform: scale(1.08); opacity: 0.75; }
}
.empty-state {
  text-align: center;
  padding: 2rem 1rem;
  color: var(--text-mute);
}
.message-cell {
  max-width: 360px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
@media (max-width: 1100px) {
  .metrics-bento {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .overview-lower {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .panel-toolbar {
    flex-direction: column;
    align-items: stretch;
  }
  .toolbar-left,
  .toolbar-right {
    align-items: stretch;
  }
  .toolbar-left,
  .toolbar-right,
  .search-box,
  .namespace-select select {
    width: 100%;
    min-width: 0;
  }

  .metrics-bento,
  .quick-links {
    grid-template-columns: 1fr;
  }
}
</style>

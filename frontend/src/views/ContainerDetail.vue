<template>
  <div class="inspect-view animate-fade-in" :class="{ 'is-fullscreen': isFullscreen }">
    <div v-if="loading && !container" class="inspect-loading">
      <div class="shimmer-block"></div>
      <div class="shimmer-block wide"></div>
    </div>

    <template v-else-if="container">
      <!-- Top Bar: Name, Image, Running Badge, Fullscreen & ESC -->
      <header class="inspect-header">
        <div class="header-left">
          <div class="title-row">
            <h1 class="container-name">{{ container.name }}</h1>
          </div>
          <div class="image-sub">{{ container.image }}</div>
          <div class="status-row">
            <span
              class="state-badge"
              :class="container.state === 'running' ? 'state-running' : 'state-stopped'"
            >
              {{ (container.state || 'stopped').toUpperCase() }}
            </span>
          </div>
        </div>

        <div class="header-actions">
          <!-- Action Buttons with permissions -->
          <button
            type="button"
            class="action-btn logs-btn"
            @click="handleContainerLogs"
            title="View Logs"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
              <line x1="8" y1="8" x2="16" y2="8"></line>
              <line x1="8" y1="12" x2="13" y2="12"></line>
            </svg>
            <span>Logs</span>
          </button>

          <button
            v-if="userCanShell(sharedState.currentUser) && container.state === 'running'"
            type="button"
            class="action-btn shell-btn"
            @click="goToShell(container.id)"
            title="Terminal Shell"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <polyline points="4 17 10 11 4 5"></polyline>
              <line x1="12" y1="19" x2="20" y2="19"></line>
            </svg>
            <span>Shell</span>
          </button>

          <button
            v-if="userCanStart(sharedState.currentUser) && container.state !== 'running'"
            type="button"
            class="action-btn start-btn"
            @click="triggerConfirm(container.id, 'start')"
            title="Start Container"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <polygon points="5 3 19 12 5 21 5 3"></polygon>
            </svg>
            <span>Start</span>
          </button>

          <button
            v-if="userCanStop(sharedState.currentUser) && container.state === 'running'"
            type="button"
            class="action-btn stop-btn"
            @click="triggerConfirm(container.id, 'stop')"
            title="Stop Container"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <rect x="5" y="5" width="14" height="14" rx="2" ry="2"></rect>
            </svg>
            <span>Stop</span>
          </button>

          <button
            v-if="userCanRestart(sharedState.currentUser)"
            type="button"
            class="action-btn restart-btn"
            @click="triggerConfirm(container.id, 'restart')"
            title="Restart Container"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <polyline points="23 4 23 10 17 10"></polyline>
              <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
            </svg>
            <span>Restart</span>
          </button>

          <button
            v-if="userCanDelete(sharedState.currentUser)"
            type="button"
            class="action-btn delete-btn"
            @click="triggerConfirm(container.id, 'remove')"
            title="Delete Container"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <polyline points="3 6 5 6 21 6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
            </svg>
            <span>Delete</span>
          </button>

          <button
            type="button"
            class="back-pill-btn"
            @click="goBack"
            title="Back to containers"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.5">
              <path d="M19 12H5M12 19l-7-7 7-7"></path>
            </svg>
            Back
          </button>
        </div>
      </header>

      <!-- SECTION: OVERVIEW -->
      <section class="inspect-section">
        <h2 class="section-heading">OVERVIEW</h2>
        <div class="card overview-card">
          <div class="overview-grid">
            <div class="overview-cell">
              <span class="cell-label">Started</span>
              <span class="cell-value">{{ overviewData.startedAgo }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">Created</span>
              <span class="cell-value">{{ overviewData.createdAgo }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">ID</span>
              <span
                class="cell-value mono clickable-id"
                @click="copyText(container.id, 'Container ID')"
                title="Click to copy ID"
              >
                {{ container.id.substring(0, 12) }}
              </span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">Restart policy</span>
              <span class="cell-value mono">{{ overviewData.restartPolicy }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">Memory limit</span>
              <span class="cell-value mono">{{ overviewData.memoryLimit }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">CPU limit</span>
              <span class="cell-value mono">{{ overviewData.cpuLimit }}</span>
            </div>
          </div>

          <div class="command-block">
            <span class="command-label">Command</span>
            <div class="command-box mono">
              <code>{{ overviewData.command }}</code>
            </div>
          </div>
        </div>
      </section>

      <!-- SECTION: PORTS -->
      <section class="inspect-section">
        <h2 class="section-heading">PORTS</h2>
        <div class="card ports-card">
          <div v-if="portsList.length > 0" class="ports-list mono">
            <div v-for="port in portsList" :key="port" class="port-item">
              {{ port }}
            </div>
          </div>
          <div v-else class="empty-field-text mono">
            No published ports
          </div>
        </div>
      </section>

      <!-- SECTION: NETWORKS -->
      <section class="inspect-section" v-if="networksList.length > 0">
        <h2 class="section-heading">NETWORKS</h2>
        <div class="card networks-card">
          <div
            v-for="net in networksList"
            :key="net.name"
            class="network-row-group"
          >
            <!-- Accordion Header -->
            <div class="network-header" @click="net.expanded = !net.expanded">
              <div class="net-title-side">
                <svg
                  class="chevron-icon"
                  :class="{ open: net.expanded }"
                  viewBox="0 0 24 24"
                  width="14"
                  height="14"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                >
                  <polyline points="6 9 12 15 18 9"></polyline>
                </svg>
                <span class="net-name bold">{{ net.name }}</span>
                <span class="net-dot">·</span>
                <span class="net-driver">{{ net.driver }}</span>
              </div>
              <div class="net-ip-side mono">
                {{ net.ip || "—" }}
              </div>
            </div>

            <!-- Accordion Body -->
            <div v-if="net.expanded" class="network-body">
              <div class="net-grid">
                <div class="net-cell">
                  <span class="cell-label">Subnet</span>
                  <span class="cell-value mono">{{ net.subnet || "—" }}</span>
                </div>
                <div class="net-cell">
                  <span class="cell-label">Gateway</span>
                  <span class="cell-value mono">{{ net.gateway || "—" }}</span>
                </div>
                <div class="net-cell" v-if="net.ipCount">
                  <span class="cell-label">IPs in use</span>
                  <span class="cell-value mono">{{ net.ipCount }}</span>
                </div>
              </div>

              <!-- Connected containers list -->
              <div class="connected-containers">
                <span class="cell-label">Connected containers</span>
                <div class="connected-row">
                  <div class="conn-name-side">
                    <span class="green-status-dot"></span>
                    <span class="conn-name bold">{{ container.name }}</span>
                    <span class="conn-tag">· this container</span>
                  </div>
                  <div class="conn-ip mono">
                    {{ net.ip || "—" }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- SECTION: MOUNTS -->
      <section class="inspect-section">
        <h2 class="section-heading">MOUNTS</h2>
        <div class="card mounts-card" :class="{ 'is-empty': mountsList.length === 0 }">
          <div v-if="mountsList.length > 0" class="mounts-table">
            <div
              v-for="(m, idx) in mountsList"
              :key="'mount-' + idx"
              class="mount-row"
            >
              <div class="mount-main">
                <div class="mount-path-row">
                  <span class="mount-dest mono bold">{{ m.destination }}</span>
                  <span v-if="m.type" class="mount-type-tag mono">{{ m.type }}</span>
                  <span v-if="m.mode" class="mount-mode-tag mono">{{ m.mode }}</span>
                </div>
                <div class="mount-source-row mono" v-if="m.source">
                  <span class="source-label">Source:</span>
                  <span class="source-val">{{ m.source }}</span>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="empty-field-text mono">
            No volume or bind mounts configured.
          </div>
        </div>
      </section>

      <!-- SECTION: LABELS -->
      <section class="inspect-section">
        <h2 class="section-heading">LABELS</h2>
        <div class="card labels-card" :class="{ 'is-empty': labelItems.length === 0 }">
          <div v-if="labelItems.length > 0" class="labels-table">
            <div
              v-for="item in labelItems"
              :key="item.key"
              class="label-row"
            >
              <div class="label-key mono">{{ item.key }}</div>
              <div class="label-val mono">{{ item.value }}</div>
            </div>
          </div>
          <div v-else class="empty-field-text mono">
            No labels configured.
          </div>
        </div>
      </section>

      <!-- SECTION: ENVIRONMENT -->
      <section class="inspect-section">
        <h2 class="section-heading">ENVIRONMENT</h2>
        <div class="card env-card" :class="{ 'is-empty': envItems.length === 0 }">
          <div v-if="envItems.length > 0" class="env-table">
            <div
              v-for="item in envItems"
              :key="item.key"
              class="env-row"
            >
              <div class="env-key mono">{{ item.key }}</div>
              <div class="env-val mono" :class="{ 'is-masked': item.value === '••••••••••••' }">
                {{ item.value }}
              </div>
            </div>
          </div>
          <div v-else class="empty-field-text mono">
            No environment variables configured.
          </div>
        </div>
        <p v-if="hasMaskedEnv" class="env-disclaimer">
          Environment variables are hidden. Reading them requires shell access.
        </p>
      </section>
    </template>

    <div v-else class="inspect-empty">
      <h2>Container not found</h2>
      <p>This container may have been removed or you do not have permission to inspect it.</p>
      <router-link to="/containers" class="back-btn">Back to containers</router-link>
    </div>

    <!-- Confirm Action Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="showConfirm"
          class="modal-overlay"
          @click.self="closeConfirm"
        >
          <div
            class="modal-content shadow-2xl"
            :class="{ 'modal-content-working': actionPending }"
            :aria-busy="actionPending"
          >
            <div :class="['modal-icon', actionClass]">
              <svg v-if="pendingAction === 'start'" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.5">
                <polygon points="5 3 19 12 5 21 5 3"></polygon>
              </svg>
              <svg v-else-if="pendingAction === 'stop'" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.5">
                <rect x="5" y="5" width="14" height="14" rx="2" ry="2"></rect>
              </svg>
              <svg v-else-if="pendingAction === 'restart'" viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.5">
                <polyline points="23 4 23 10 17 10"></polyline>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
              </svg>
              <svg v-else viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.5">
                <polyline points="3 6 5 6 21 6"></polyline>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              </svg>
            </div>
            <div class="modal-text-center">
              <h3>Confirm {{ pendingAction }}</h3>
              <p>Are you sure you want to <strong>{{ pendingAction }}</strong> this container?</p>
            </div>
            <div class="modal-actions">
              <button
                type="button"
                class="modal-btn cancel"
                :disabled="actionPending"
                @click="closeConfirm"
              >
                Cancel
              </button>
              <button
                type="button"
                :class="['modal-btn confirm', actionClass, { 'is-working': actionPending }]"
                :disabled="actionPending"
                @click="confirmAction"
              >
                {{ actionPending ? `Working...` : `Confirm ${pendingAction}` }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useContainers } from "../composables/useContainers";
import { logsRouteForContainer } from "../utils/logRoutes";
import { apiFetch } from "../utils/apiFetch";
import { secureStorage } from "../utils/storage";
import {
  sharedState,
  showToast,
  userCanStart,
  userCanStop,
  userCanRestart,
  userCanDelete,
  userCanShell,
} from "../utils/sharedState";

const route = useRoute();
const router = useRouter();
const containerId = computed(() => route.params.id);

const {
  containers,
  loading,
  fetchContainers,
  goToLogs,
  goToShell,
  showConfirm,
  pendingAction,
  actionPending,
  actionClass,
  triggerConfirm,
  closeConfirm,
  executeAction,
} = useContainers({ autoPoll: false });

const inspectData = ref(null);
const isFullscreen = ref(false);
const rawNetworks = ref([]);

const container = computed(() => {
  const id = containerId.value;
  return containers.value.find(
    (c) => c.id === id || c.id.startsWith(id) || id.startsWith(c.id)
  );
});

const handleContainerLogs = () => {
  if (!container.value) return;
  router.push(logsRouteForContainer(container.value));
};

const resolvedInspect = computed(() => {
  const data = inspectData.value;
  if (!data) return {};
  if (data.Container && typeof data.Container === "object") {
    return { ...data, ...data.Container };
  }
  return data;
});

const hasMaskedEnv = computed(() => {
  return envItems.value.some((item) => item.value === "••••••••••••");
});

const canShell = computed(() => {
  return userCanShell(sharedState.currentUser);
});

function formatTimeAgo(dateString) {
  if (!dateString || dateString === "0001-01-01T00:00:00Z") return "Never";
  const date = new Date(dateString);
  if (Number.isNaN(date.getTime())) return dateString;

  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
  if (seconds < 0) return "Just now";
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes} minute${minutes === 1 ? "" : "s"} ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} hour${hours === 1 ? "" : "s"} ago`;
  const days = Math.floor(hours / 24);
  return `${days} day${days === 1 ? "" : "s"} ago`;
}

const overviewData = computed(() => {
  const cfg = resolvedInspect.value?.Config || {};
  const state = resolvedInspect.value?.State || {};
  const host = resolvedInspect.value?.HostConfig || {};

  let cmd = "-";
  if (Array.isArray(cfg.Cmd) && cfg.Cmd.length > 0) {
    cmd = cfg.Cmd.join(" ");
  } else if (typeof cfg.Cmd === "string") {
    cmd = cfg.Cmd;
  } else if (Array.isArray(cfg.Entrypoint) && cfg.Entrypoint.length > 0) {
    cmd = cfg.Entrypoint.join(" ");
  }

  // Memory Limit
  let memLimitStr = "Unlimited";
  const mem = host.Memory || 0;
  if (mem > 0) {
    const mb = mem / (1024 * 1024);
    memLimitStr = mb >= 1024 ? `${(mb / 1024).toFixed(1)} GB` : `${mb.toFixed(0)} MB`;
  }

  // CPU Limit
  let cpuLimitStr = "Unlimited";
  const nano = host.NanoCpus || 0;
  if (nano > 0) {
    const cpus = nano / 1e9;
    cpuLimitStr = `${cpus} CPU${cpus === 1 ? "" : "s"}`;
  }

  return {
    startedAgo: formatTimeAgo(state.StartedAt),
    createdAgo: formatTimeAgo(resolvedInspect.value?.Created || container.value?.created),
    restartPolicy: host.RestartPolicy?.Name || "no",
    memoryLimit: memLimitStr,
    cpuLimit: cpuLimitStr,
    command: cmd,
  };
});

const portsList = computed(() => {
  const bindings = resolvedInspect.value?.NetworkSettings?.Ports || {};
  const list = [];
  for (const [containerPort, hostBindings] of Object.entries(bindings)) {
    if (!hostBindings || hostBindings.length === 0) {
      list.push(containerPort);
      continue;
    }
    for (const binding of hostBindings) {
      const hostPort = binding.HostPort;
      list.push(`${hostPort} → ${containerPort}`);
    }
  }
  return list.sort();
});

const networksList = computed(() => {
  const netSettings = resolvedInspect.value?.NetworkSettings?.Networks || {};
  const result = [];

  for (const [name, net] of Object.entries(netSettings)) {
    const ip = net.IPAddress;
    const gateway = net.Gateway;
    const prefix = net.IPPrefixLen;
    let subnet = "";
    if (ip && prefix) {
      const parts = ip.split(".");
      if (prefix === 16 && parts.length === 4) {
        subnet = `${parts[0]}.${parts[1]}.0.0/16`;
      } else if (prefix === 24 && parts.length === 4) {
        subnet = `${parts[0]}.${parts[1]}.${parts[2]}.0/24`;
      } else {
        subnet = `${ip}/${prefix}`;
      }
    }

    let driver = net.Driver;
    if (!driver) {
      if (name === "bridge" || name === "host" || name === "none" || name === "macvlan" || name === "ipvlan") {
        driver = name;
      } else if (name.includes("overlay") || name === "ingress" || name.includes("swarm")) {
        driver = "overlay";
      } else {
        driver = "bridge";
      }
    }

    let ipCount = "";
    if (prefix === 16) {
      ipCount = "4 / 65,536";
    } else if (prefix === 24) {
      ipCount = "4 / 256";
    }

    result.push({
      name,
      driver,
      ip: ip || "",
      gateway: gateway || "",
      subnet: subnet || "",
      ipCount,
      expanded: true,
    });
  }

  return result;
});

const mountsList = computed(() => {
  const inspect = resolvedInspect.value || {};
  const mounts = inspect.Mounts || [];

  if (mounts.length > 0) {
    return mounts.map((m) => ({
      type: m.Type || (m.Name ? "volume" : "bind"),
      destination: m.Destination || m.Target || "",
      source: m.Source || m.Name || "",
      mode: m.Mode || (m.RW ? "rw" : "ro"),
      driver: m.Driver || "",
    }));
  }

  // Fallback to HostConfig.Binds if Mounts array is empty
  const binds = inspect.HostConfig?.Binds || [];
  if (Array.isArray(binds) && binds.length > 0) {
    return binds.map((b) => {
      const parts = String(b).split(":");
      return {
        type: parts[0]?.startsWith("/") ? "bind" : "volume",
        destination: parts[1] || parts[0],
        source: parts[0] || "",
        mode: parts[2] || "rw",
        driver: "",
      };
    });
  }

  return [];
});

const labelItems = computed(() => {
  const labels = resolvedInspect.value?.Config?.Labels || {};
  return Object.entries(labels).map(([key, value]) => ({ key, value }));
});

const envItems = computed(() => {
  const env = resolvedInspect.value?.Config?.Env || [];
  return env.map((line) => {
    const idx = line.indexOf("=");
    if (idx === -1) {
      return { key: line, value: "" };
    }
    const key = line.slice(0, idx);
    const value = line.slice(idx + 1);
    return { key, value };
  });
});

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value;
}

function goBack() {
  router.push("/containers");
}

function handleKeydown(e) {
  if (e.key === "Escape") {
    if (isFullscreen.value) {
      isFullscreen.value = false;
    } else {
      goBack();
    }
  }
}

function copyText(text, label) {
  navigator.clipboard.writeText(text).then(() => {
    showToast("Copied", "Copied to clipboard", "info");
  });
}

async function fetchInspect() {
  try {
    const token = secureStorage.getItem("token");
    const res = await apiFetch(`/api/containers/${containerId.value}/inspect`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      inspectData.value = await res.json();
    }
  } catch (err) {
    console.error("Inspect fetch failed", err);
  }
}

const confirmAction = async () => {
  const action = pendingAction.value;
  await executeAction();
  if (action === "remove") {
    router.push("/containers");
  } else {
    await fetchContainers();
    await fetchInspect();
  }
};

onMounted(async () => {
  window.addEventListener("keydown", handleKeydown);
  await fetchContainers();
  await fetchInspect();
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
});
</script>

<style scoped>
.inspect-view {
  max-width: 960px;
  margin: 0 auto;
  padding: 1.5rem 1rem 3rem;
  color: var(--text-main);
  font-family: var(--font-main);
  transition: all 0.2s ease;
}

.inspect-view.is-fullscreen {
  max-width: 100%;
  padding: 1.5rem 2rem 4rem;
}

/* Header */
.inspect-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
}

.container-name {
  font-size: 1.75rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text-main);
  margin: 0 0 0.2rem 0;
}

.image-sub {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--text-dim);
  margin-bottom: 0.75rem;
}

.status-row {
  margin-top: 0.35rem;
}

.state-badge {
  display: inline-block;
  padding: 0.2rem 0.6rem;
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  border-radius: 4px;
}

.state-running {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.state-stopped {
  background: var(--bg-subtle);
  color: var(--text-mute);
  border: 1px solid var(--border);
}

.header-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.85rem;
  font-size: 0.82rem;
  font-weight: 600;
  border-radius: 8px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn:hover {
  background: var(--card-hover);
  border-color: var(--border-active);
  transform: translateY(-1px);
}

.action-btn.logs-btn:hover {
  color: var(--accent);
  border-color: var(--accent);
}

.action-btn.shell-btn:hover {
  color: #8b5cf6;
  border-color: #8b5cf6;
}

.action-btn.start-btn:hover {
  color: var(--success);
  border-color: var(--success);
}

.action-btn.stop-btn:hover {
  color: var(--error);
  border-color: var(--error);
}

.action-btn.restart-btn:hover {
  color: var(--warning);
  border-color: var(--warning);
}

.action-btn.delete-btn:hover {
  color: var(--error);
  border-color: var(--error);
}

.back-pill-btn {
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-dim);
  font-size: 0.82rem;
  font-weight: 600;
  padding: 0.45rem 0.85rem;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  transition: all 0.15s ease;
}

.back-pill-btn:hover {
  color: var(--text-main);
  background: var(--bg-subtle);
  border-color: var(--border-active);
}

/* Modal styles */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(2, 6, 23, 0.65);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  width: min(400px, 100%);
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 20px 25px -5px var(--shadow);
}

.modal-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1rem;
  background: var(--bg-input);
}

.modal-icon.success {
  color: var(--success);
}

.modal-icon.warning {
  color: var(--warning);
}

.modal-icon.error {
  color: var(--error);
}

.modal-text-center {
  text-align: center;
  margin-bottom: 1.25rem;
}

.modal-text-center h3 {
  font-size: 1.1rem;
  font-weight: 700;
  margin: 0 0 0.35rem;
  color: var(--text-main);
  text-transform: capitalize;
}

.modal-text-center p {
  font-size: 0.85rem;
  color: var(--text-dim);
  margin: 0;
}

.modal-actions {
  display: flex;
  gap: 0.65rem;
}

.modal-btn {
  flex: 1;
  padding: 0.65rem 1rem;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 700;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.modal-btn.cancel {
  background: var(--bg-input);
  color: var(--text-main);
  border-color: var(--border);
}

.modal-btn.cancel:hover {
  background: var(--bg-subtle);
}

.modal-btn.confirm {
  color: #fff;
}

.modal-btn.confirm.success {
  background: var(--success);
}

.modal-btn.confirm.warning {
  background: var(--warning);
}

.modal-btn.confirm.error {
  background: var(--error);
}

/* Section Headings */
.inspect-section {
  margin-bottom: 1.75rem;
}

.section-heading {
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  color: var(--text-mute);
  text-transform: uppercase;
  margin: 0 0 0.65rem 0;
}

/* Theme Cards */
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

/* OVERVIEW CARD */
.overview-card {
  padding: 1.25rem 1.25rem 1rem;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  row-gap: 1.25rem;
  column-gap: 1.5rem;
  margin-bottom: 1.25rem;
}

@media (max-width: 768px) {
  .overview-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.overview-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.cell-label {
  font-size: 0.72rem;
  color: var(--text-mute);
  font-weight: 500;
}

.cell-value {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-main);
}

.mono {
  font-family: var(--font-mono);
}

.clickable-id {
  cursor: pointer;
  color: var(--text-main);
}
.clickable-id:hover {
  color: var(--accent);
  text-decoration: underline;
}

.command-block {
  border-top: 1px solid var(--border);
  padding-top: 1rem;
}

.command-label {
  display: block;
  font-size: 0.72rem;
  color: var(--text-mute);
  margin-bottom: 0.4rem;
}

.command-box {
  background: var(--bg-input);
  border-radius: 6px;
  padding: 0.65rem 0.85rem;
  color: var(--text-main);
  font-size: 0.8rem;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  border: 1px solid var(--border-subtle);
}

/* PORTS CARD */
.ports-card {
  padding: 0.85rem 1.25rem;
}

.port-item {
  font-size: 0.85rem;
  color: var(--text-main);
  padding: 0.2rem 0;
}

/* NETWORKS CARD */
.networks-card {
  padding: 0;
}

.network-row-group {
  border-bottom: 1px solid var(--border);
}
.network-row-group:last-child {
  border-bottom: none;
}

.network-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.85rem 1.25rem;
  cursor: pointer;
  user-select: none;
  transition: background 0.15s ease;
}
.network-header:hover {
  background: var(--bg-subtle);
}

.net-title-side {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.88rem;
}

.chevron-icon {
  color: var(--text-mute);
  transition: transform 0.2s ease;
}
.chevron-icon.open {
  transform: rotate(0deg);
}
.chevron-icon:not(.open) {
  transform: rotate(-90deg);
}

.net-name {
  color: var(--text-main);
  font-weight: 700;
}

.net-dot {
  color: var(--text-mute);
}

.net-driver {
  color: var(--text-dim);
  font-size: 0.82rem;
}

.net-ip-side {
  color: var(--text-dim);
  font-size: 0.85rem;
}

.network-body {
  padding: 0 1.25rem 1.25rem 2.4rem;
  border-top: 1px solid var(--border);
  background: var(--bg-input);
}

.net-grid {
  display: flex;
  gap: 3.5rem;
  padding: 1rem 0;
  border-bottom: 1px solid var(--border);
}

.net-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.connected-containers {
  padding-top: 1rem;
}

.connected-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.5rem;
  font-size: 0.85rem;
}

.conn-name-side {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.green-status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--success);
}

.conn-name {
  color: var(--text-main);
  font-weight: 700;
}

.conn-tag {
  color: var(--text-mute);
  font-size: 0.8rem;
}

.conn-ip {
  color: var(--text-dim);
}

/* MOUNTS CARD */
.mounts-card {
  padding: 0;
}

.mounts-table {
  display: flex;
  flex-direction: column;
}

.mount-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.85rem 1.25rem;
  border-bottom: 1px solid var(--border);
  font-size: 0.84rem;
}
.mount-row:last-child {
  border-bottom: none;
}

.mount-main {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  width: 100%;
}

.mount-path-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.mount-dest {
  color: var(--text-main);
  font-weight: 700;
  font-size: 0.88rem;
}

.mount-type-tag {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  padding: 0.12rem 0.45rem;
  border-radius: 4px;
  background: var(--bg-subtle);
  color: var(--accent);
  border: 1px solid var(--border);
}

.mount-mode-tag {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  padding: 0.12rem 0.45rem;
  border-radius: 4px;
  background: var(--bg-subtle);
  color: var(--text-mute);
  border: 1px solid var(--border);
}

.mount-source-row {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.78rem;
  color: var(--text-dim);
  word-break: break-all;
}

.source-label {
  color: var(--text-mute);
  font-weight: 600;
}

.source-val {
  color: var(--text-dim);
}

/* LABELS CARD */
.labels-card {
  padding: 0;
}

.labels-table {
  display: flex;
  flex-direction: column;
}

.label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid var(--border);
  font-size: 0.82rem;
}
.label-row:last-child {
  border-bottom: none;
}

.label-key {
  color: var(--text-main);
  font-weight: 700;
  letter-spacing: -0.01em;
  word-break: break-all;
}

.label-val {
  color: var(--text-dim);
  font-weight: 600;
  text-align: right;
  max-width: 65%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ENVIRONMENT CARD */
.env-card {
  padding: 0;
}

.env-table {
  display: flex;
  flex-direction: column;
}

.env-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid var(--border);
  font-size: 0.82rem;
}
.env-row:last-child {
  border-bottom: none;
}

.env-key {
  color: var(--text-main);
  font-weight: 700;
  letter-spacing: -0.01em;
}

.env-val {
  color: var(--text-dim);
  font-weight: 600;
  text-align: right;
  max-width: 65%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.env-val.is-masked {
  letter-spacing: 0.15em;
  color: var(--text-mute);
}

.env-disclaimer {
  margin-top: 0.65rem;
  font-size: 0.75rem;
  color: var(--text-mute);
}

.empty-field-text {
  padding: 0.95rem 1.25rem;
  font-size: 0.84rem;
  color: var(--text-mute);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.card.is-empty {
  padding: 0;
  display: flex;
}

/* Loading & Empty States */
.inspect-loading {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.shimmer-block {
  height: 120px;
  border-radius: 8px;
  background: var(--bg-card);
  animation: shimmer 1.5s infinite;
}

.shimmer-block.wide {
  height: 260px;
}

.inspect-empty {
  text-align: center;
  padding: 4rem 1rem;
}

.back-btn {
  display: inline-block;
  margin-top: 1rem;
  padding: 0.5rem 1.2rem;
  border-radius: 6px;
  background: var(--accent);
  color: #fff;
  font-weight: 700;
  text-decoration: none;
}
</style>

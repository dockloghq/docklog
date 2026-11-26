<template>
  <div class="inspect-view animate-fade-in" :class="{ 'is-fullscreen': isFullscreen }">
    <div v-if="loading && !pod" class="inspect-loading">
      <div class="shimmer-block"></div>
      <div class="shimmer-block wide"></div>
    </div>

    <template v-else-if="pod">
      <!-- Top Bar: Name, Image, Phase Badge, Actions -->
      <header class="inspect-header">
        <div class="header-left">
          <div class="title-row">
            <h1 class="container-name">{{ pod.name }}</h1>
          </div>
          <div class="image-sub">
            <span class="mono">{{ pod.namespace }}</span> · {{ primaryImage || 'kubernetes/pod' }}
          </div>
          <div class="status-row">
            <span
              class="state-badge"
              :class="isRunning ? 'state-running' : 'state-stopped'"
            >
              {{ (pod.phase || 'UNKNOWN').toUpperCase() }}
            </span>
          </div>
        </div>

        <div class="header-actions">
          <button
            type="button"
            class="action-btn logs-btn"
            @click="goToLogs"
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
            v-if="userCanShell(sharedState.currentUser) && isRunning"
            type="button"
            class="action-btn shell-btn"
            @click="goToShell"
            title="Terminal Shell"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <polyline points="4 17 10 11 4 5"></polyline>
              <line x1="12" y1="19" x2="20" y2="19"></line>
            </svg>
            <span>Shell</span>
          </button>

          <button
            v-if="userCanStart(sharedState.currentUser) && !isRunning"
            type="button"
            class="action-btn start-btn"
            @click="triggerConfirm('start')"
            title="Start Pod"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <polygon points="5 3 19 12 5 21 5 3"></polygon>
            </svg>
            <span>Start</span>
          </button>

          <button
            v-if="userCanStop(sharedState.currentUser) && isRunning"
            type="button"
            class="action-btn stop-btn"
            @click="triggerConfirm('stop')"
            title="Stop Pod"
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
            @click="triggerConfirm('restart')"
            title="Restart Pod"
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
            @click="triggerConfirm('remove')"
            title="Delete Pod"
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
            title="Back to Kubernetes"
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
              <span class="cell-label">Namespace</span>
              <span class="cell-value mono">{{ pod.namespace }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">Created</span>
              <span class="cell-value">{{ overviewData.createdAgo }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">Pod UID</span>
              <span
                class="cell-value mono clickable-id"
                @click="copyText(pod.uid, 'Pod UID')"
                title="Click to copy UID"
              >
                {{ shortUid }}
              </span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">Node</span>
              <span class="cell-value mono">{{ pod.node || '—' }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">Pod IP</span>
              <span class="cell-value mono">{{ pod.ip || '—' }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">Host IP</span>
              <span class="cell-value mono">{{ pod.host_ip || '—' }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">Memory usage / limit</span>
              <span class="cell-value mono">{{ memoryDisplay }} / {{ memLimit ? formatBytes(memLimit) : 'Unlimited' }}</span>
            </div>
            <div class="overview-cell">
              <span class="cell-label">CPU usage / limit</span>
              <span class="cell-value mono">{{ cpuDisplay }} / {{ cpuLimitLabel }}</span>
            </div>
          </div>

          <div class="command-block" v-if="pod.service_account || pod.qos_class">
            <span class="command-label">Attributes &amp; Service Account</span>
            <div class="command-box mono">
              <span>ServiceAccount: <strong>{{ pod.service_account || 'default' }}</strong></span>
              <span v-if="pod.qos_class" class="ml-4">· QoS: <strong>{{ pod.qos_class }}</strong></span>
              <span v-if="pod.restarts != null" class="ml-4">· Restarts: <strong>{{ pod.restarts }}</strong></span>
            </div>
          </div>
        </div>
      </section>

      <!-- SECTION: CONTAINERS -->
      <section class="inspect-section" v-if="podContainers.length > 0">
        <h2 class="section-heading">CONTAINERS ({{ podContainers.length }})</h2>
        <div class="card containers-card">
          <div
            v-for="c in podContainers"
            :key="c.name"
            class="pod-container-row"
          >
            <div class="p-container-header">
              <div class="p-c-name-side">
                <span
                  class="green-status-dot"
                  :style="{ background: c.state === 'running' ? 'var(--success)' : 'var(--text-mute)' }"
                ></span>
                <span class="conn-name bold">{{ c.name }}</span>
                <span v-if="c.isInit" class="mount-type-tag mono">INIT</span>
                <span class="conn-tag mono">· {{ c.image }}</span>
              </div>
              <div class="p-c-meta-side mono">
                <span>Restarts: {{ c.restart_count ?? 0 }}</span>
              </div>
            </div>

            <div class="p-container-body mono">
              <div class="p-c-grid">
                <div>
                  <span class="source-label">CPU: </span>
                  <span class="source-val">{{ c.cpu_request || '-' }} / {{ c.cpu_limit || '-' }}</span>
                </div>
                <div>
                  <span class="source-label">Memory: </span>
                  <span class="source-val">{{ c.memory_request || '-' }} / {{ c.memory_limit || '-' }}</span>
                </div>
                <div v-if="c.ports && c.ports !== '-'">
                  <span class="source-label">Ports: </span>
                  <span class="source-val">{{ c.ports }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- SECTION: PORTS -->
      <section class="inspect-section">
        <h2 class="section-heading">PORTS</h2>
        <div class="card ports-card">
          <div v-if="allPorts.length > 0" class="ports-list mono">
            <div v-for="port in allPorts" :key="port" class="port-item">
              {{ port }}
            </div>
          </div>
          <div v-else class="empty-field-text mono">
            No published container ports
          </div>
        </div>
      </section>

      <!-- SECTION: MOUNTS -->
      <section class="inspect-section">
        <h2 class="section-heading">MOUNTS</h2>
        <div class="card mounts-card">
          <div v-if="allMountsList.length > 0" class="mounts-table">
            <div
              v-for="(m, idx) in allMountsList"
              :key="'mount-' + idx"
              class="mount-row"
            >
              <div class="mount-main">
                <div class="mount-path-row">
                  <span class="mount-dest mono bold">{{ m.mountPath }}</span>
                  <span class="mount-type-tag mono">{{ m.container }}</span>
                  <span v-if="m.readOnly" class="mount-mode-tag mono">RO</span>
                  <span v-else class="mount-mode-tag mono">RW</span>
                </div>
                <div class="mount-source-row mono" v-if="m.volumeName">
                  <span class="source-label">Volume:</span>
                  <span class="source-val">{{ m.volumeName }}</span>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="empty-field-text mono">
            No volume mounts configured.
          </div>
        </div>
      </section>

      <!-- SECTION: LABELS -->
      <section class="inspect-section">
        <h2 class="section-heading">LABELS</h2>
        <div class="card labels-card">
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
              :key="item.raw"
              class="env-row"
            >
              <div class="env-key mono">
                <span class="container-prefix">{{ item.container }}</span>
                <span class="key-name">{{ item.key }}</span>
              </div>
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
      <h2>Pod not found</h2>
      <p>This pod may have been terminated or you do not have permission to inspect it.</p>
      <button type="button" @click="goBack" class="back-btn">Back to Kubernetes</button>
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
              <p>Are you sure you want to <strong>{{ pendingAction }}</strong> pod <strong>{{ podName }}</strong>?</p>
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
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { apiFetch } from '../utils/apiFetch';
import { secureStorage } from '../utils/storage';
import { logsRouteForPod } from '../utils/logRoutes';
import {
  sharedState,
  showToast,
  formatBytes,
  userCanStart,
  userCanStop,
  userCanRestart,
  userCanDelete,
  userCanShell,
} from '../utils/sharedState';

const route = useRoute();
const router = useRouter();

const loading = ref(true);
const pod = ref(null);
const isFullscreen = ref(false);
const actionPending = ref(false);
const showConfirm = ref(false);
const pendingAction = ref('');
const liveStats = ref({
  cpu: 0,
  memory: 0,
  memory_limit: 0,
  cpu_limit_millicores: 0,
  metrics_available: false,
});
let statsTimer = null;

const namespace = computed(() => String(route.params.namespace || ''));
const podName = computed(() => String(route.params.pod || ''));

const isRunning = computed(() => String(pod.value?.phase || '').toLowerCase() === 'running');

const canShell = computed(() => {
  return userCanShell(sharedState.currentUser);
});

const hasMaskedEnv = computed(() => {
  return envItems.value.some((item) => item.value === "••••••••••••");
});

const shortUid = computed(() => {
  const uid = String(pod.value?.uid || '');
  if (!uid) return '';
  return uid.length > 12 ? uid.slice(0, 12) : uid;
});

const primaryImage = computed(() => pod.value?.containers?.[0]?.image || '');

function formatTimeAgo(timestamp) {
  if (!timestamp) return "Never";
  let date;
  if (typeof timestamp === "number") {
    date = new Date(timestamp * 1000);
  } else {
    date = new Date(timestamp);
  }
  if (Number.isNaN(date.getTime())) return String(timestamp);

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
  return {
    createdAgo: formatTimeAgo(pod.value?.created),
  };
});

const podContainers = computed(() => {
  const containers = (pod.value?.containers || []).map((c) => ({ ...c, isInit: false }));
  const initContainers = (pod.value?.init_containers || []).map((c) => ({ ...c, isInit: true }));
  return [...containers, ...initContainers];
});

const allPorts = computed(() => {
  const ports = new Set();
  for (const c of pod.value?.containers || []) {
    if (c.ports && c.ports !== '-') {
      c.ports.split(',').forEach((p) => {
        const trimmed = p.trim();
        if (trimmed) ports.add(trimmed);
      });
    }
  }
  return [...ports].sort();
});

const allMountsList = computed(() => {
  const list = [];
  const containers = pod.value?.containers || [];
  const initContainers = pod.value?.init_containers || [];
  for (const c of [...containers, ...initContainers]) {
    for (const mount of c.volume_mounts || []) {
      // mount format can be "path" or "name:path" or "name -> path"
      const parts = mount.split(':');
      if (parts.length >= 2) {
        list.push({
          container: c.name,
          volumeName: parts[0].trim(),
          mountPath: parts[1].trim(),
          readOnly: false,
        });
      } else {
        list.push({
          container: c.name,
          volumeName: '',
          mountPath: mount,
          readOnly: false,
        });
      }
    }
  }
  return list;
});

const labelItems = computed(() => {
  const labels = pod.value?.labels || {};
  return Object.entries(labels).map(([key, value]) => ({ key, value }));
});

const envItems = computed(() => {
  const lines = [];
  const containers = pod.value?.containers || [];
  const initContainers = pod.value?.init_containers || [];
  for (const c of [...containers, ...initContainers]) {
    for (const env of c.env || []) {
      const idx = env.indexOf('=');
      const key = idx === -1 ? env : env.slice(0, idx);
      const value = idx === -1 ? '' : env.slice(idx + 1);
      lines.push({
        container: c.name,
        key,
        value,
        raw: `${c.name}:${env}`,
      });
    }
  }
  return lines;
});

const memLimit = computed(() => liveStats.value.memory_limit || 0);

const cpuLimitLabel = computed(() => {
  const milli = liveStats.value.cpu_limit_millicores;
  if (!milli) return 'Unlimited';
  if (milli % 1000 === 0) {
    const cpus = milli / 1000;
    return `${cpus} CPU${cpus === 1 ? '' : 's'}`;
  }
  return `${milli}m`;
});

const cpuDisplay = computed(() => {
  if (!liveStats.value.metrics_available) return '—';
  return `${liveStats.value.cpu.toFixed(1)}%`;
});

const memoryDisplay = computed(() => {
  if (!liveStats.value.metrics_available) return '—';
  return formatBytes(liveStats.value.memory);
});

const actionClass = computed(() => {
  if (pendingAction.value === 'start') return 'success';
  if (pendingAction.value === 'restart') return 'warning';
  return 'error';
});

const goToLogs = () => {
  router.push(logsRouteForPod(`${namespace.value}/${podName.value}`));
};

const goToShell = () => {
  router.push({ path: '/shell', query: { p: `${namespace.value}/${podName.value}` } });
};

function goBack() {
  router.push({
    path: '/kubernetes',
    query: { tab: 'pods', namespace: namespace.value },
  });
}

function copyText(text, label) {
  navigator.clipboard.writeText(text).then(() => {
    showToast("Copied", "Copied to clipboard", "info");
  });
}

function triggerConfirm(action) {
  pendingAction.value = action;
  showConfirm.value = true;
}

function closeConfirm() {
  if (actionPending.value) return;
  showConfirm.value = false;
  pendingAction.value = '';
}

async function confirmAction() {
  await runPodAction(pendingAction.value);
}

const runPodAction = async (action) => {
  if (actionPending.value) return;

  actionPending.value = true;
  try {
    const token = secureStorage.getItem('token');
    const formData = new URLSearchParams();
    formData.set('action', action);
    const res = await apiFetch(
      `/api/namespaces/${encodeURIComponent(namespace.value)}/pods/${encodeURIComponent(podName.value)}/action`,
      {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData.toString(),
      },
    );
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || `Failed to ${action} pod`);
    }
    showToast('Success', `Pod ${action} requested.`, 'success');
    showConfirm.value = false;
    pendingAction.value = '';
    if (action === 'remove' || action === 'stop' || action === 'restart' || action === 'start') {
      goBack();
      return;
    }
    await fetchPod();
  } catch (err) {
    showToast('Action failed', err.message || `Failed to ${action} pod`, 'error');
  } finally {
    actionPending.value = false;
  }
};

async function fetchPod() {
  loading.value = true;
  try {
    const token = secureStorage.getItem('token');
    const res = await apiFetch(
      `/api/namespaces/${encodeURIComponent(namespace.value)}/pods/${encodeURIComponent(podName.value)}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    if (!res.ok) {
      pod.value = null;
      return;
    }
    pod.value = await res.json();
  } catch (err) {
    console.error(err);
    pod.value = null;
  } finally {
    loading.value = false;
  }
}

async function fetchLiveStats() {
  if (!isRunning.value) return;
  try {
    const token = secureStorage.getItem('token');
    const res = await apiFetch(
      `/api/namespaces/${encodeURIComponent(namespace.value)}/pods/${encodeURIComponent(podName.value)}/stats-now`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    if (!res.ok) return;
    const data = await res.json();
    liveStats.value = {
      cpu: Number(data.cpu) || 0,
      memory: Number(data.memory) || 0,
      memory_limit: Number(data.memory_limit) || 0,
      cpu_limit_millicores: Number(data.cpu_limit_millicores) || 0,
      metrics_available: Boolean(data.metrics_available),
    };
  } catch (err) {
    console.debug('Failed to poll pod live stats:', err);
  }
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

onMounted(async () => {
  window.addEventListener("keydown", handleKeydown);
  await fetchPod();
  if (isRunning.value) {
    await fetchLiveStats();
    statsTimer = setInterval(() => {
      if (typeof document !== "undefined" && document.hidden) return;
      fetchLiveStats();
    }, 3000);
  }
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
  if (statsTimer) clearInterval(statsTimer);
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
  gap: 1.5rem;
  flex-wrap: wrap;
}

.container-name {
  font-size: 1.75rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text-main);
  margin: 0 0 0.2rem 0;
  word-break: break-all;
}

.image-sub {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--text-dim);
  margin-bottom: 0.75rem;
  word-break: break-all;
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

/* POD CONTAINERS CARD */
.containers-card {
  padding: 0;
}

.pod-container-row {
  border-bottom: 1px solid var(--border);
  padding: 0.85rem 1.25rem;
}
.pod-container-row:last-child {
  border-bottom: none;
}

.p-container-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.p-c-name-side {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.88rem;
}

.green-status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.conn-name {
  color: var(--text-main);
  font-weight: 700;
}

.conn-tag {
  color: var(--text-mute);
  font-size: 0.8rem;
}

.p-c-meta-side {
  font-size: 0.78rem;
  color: var(--text-mute);
}

.p-container-body {
  font-size: 0.8rem;
  color: var(--text-dim);
}

.p-c-grid {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
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
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.container-prefix {
  font-size: 0.7rem;
  font-weight: 700;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  color: var(--text-mute);
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
  font-size: 0.82rem;
  color: var(--text-mute);
  padding: 0.85rem 1.25rem;
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
  cursor: pointer;
  border: none;
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
</style>

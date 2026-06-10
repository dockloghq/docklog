<template>
  <div class="detail-view animate-fade-in">
    <div v-if="loading && !container" class="detail-loading">
      <div class="shimmer-block"></div>
      <div class="shimmer-block wide"></div>
    </div>

    <template v-else-if="container">
      <section class="detail-hero">
        <div class="hero-top">
          <router-link to="/containers" class="back-link">
            <AppIcon name="chevronLeft" :size="16" :stroke-width="3" />
            <span>Containers</span>
          </router-link>
          <div
            class="state-pill"
            :class="container.state === 'running' ? 'is-running' : 'is-stopped'"
          >
            <span class="pulse-dot"></span>
            {{ container.state }}
          </div>
        </div>

        <div class="hero-main">
          <div
            class="hero-icon"
            :class="container.state === 'running' ? 'running' : 'stopped'"
          >
            <AppIcon name="containers" :size="28" />
          </div>
          <div class="hero-copy">
            <h1>{{ container.name }}</h1>
            <div class="hero-meta">
              <button
                type="button"
                class="id-chip"
                @click="copyText(container.id, 'Container ID')"
              >
                <span>{{ container.id }}</span>
                <AppIcon name="copy" :size="14" />
              </button>
              <span class="meta-dot">·</span>
              <span class="meta-image">{{ container.image }}</span>
            </div>
            <p class="hero-status">{{ container.status }}</p>
          </div>
        </div>

        <div class="action-rail">
          <button
            type="button"
            class="action-chip logs"
            @click="goToLogs(container.id)"
          >
            <AppIcon name="logsBubble" :size="16" />
            <span>Logs</span>
          </button>
          <button
            v-if="
              userCanShell(sharedState.currentUser) &&
              container.state === 'running'
            "
            type="button"
            class="action-chip shell"
            @click="goToShell(container.id)"
          >
            <AppIcon name="terminal" :size="16" />
            <span>Shell</span>
          </button>
          <button
            v-if="
              userCanStart(sharedState.currentUser) &&
              container.state !== 'running'
            "
            type="button"
            class="action-chip start"
            @click="triggerConfirm(container.id, 'start')"
          >
            <AppIcon name="play" :size="16" :stroke-width="3" />
            <span>Start</span>
          </button>
          <button
            v-if="
              userCanStop(sharedState.currentUser) &&
              container.state === 'running'
            "
            type="button"
            class="action-chip stop"
            @click="triggerConfirm(container.id, 'stop')"
          >
            <AppIcon name="stopOutline" :size="16" :stroke-width="3" />
            <span>Stop</span>
          </button>
          <button
            v-if="userCanRestart(sharedState.currentUser)"
            type="button"
            class="action-chip restart"
            @click="triggerConfirm(container.id, 'restart')"
          >
            <AppIcon name="refresh" :size="16" :stroke-width="3" />
            <span>Restart</span>
          </button>
          <button
            v-if="userCanDelete(sharedState.currentUser)"
            type="button"
            class="action-chip delete"
            @click="triggerConfirm(container.id, 'remove')"
          >
            <AppIcon name="trash" :size="16" />
            <span>Delete</span>
          </button>
        </div>
      </section>

      <section v-if="container.state === 'running'" class="stats-grid">
        <article class="stat-card">
          <span class="stat-label">CPU</span>
          <span class="stat-value">{{ liveStats.cpu.toFixed(1) }}%</span>
          <div class="stat-bar">
            <div
              class="stat-bar-fill accent"
              :style="{ width: `${Math.min(liveStats.cpu, 100)}%` }"
            ></div>
          </div>
        </article>
        <article class="stat-card">
          <span class="stat-label">Memory</span>
          <span class="stat-value">{{ formatBytes(liveStats.memory) }}</span>
          <span v-if="memLimit" class="stat-sub"
            >of {{ formatBytes(memLimit) }}</span
          >
        </article>
        <article class="stat-card">
          <span class="stat-label">CPU limit</span>
          <span class="stat-value">{{ cpuLimitLabel }}</span>
        </article>
        <article class="stat-card">
          <span class="stat-label">Created</span>
          <span class="stat-value sm">{{ formatDate(container.created) }}</span>
        </article>
      </section>

      <section class="detail-panels">
        <article class="panel">
          <div class="panel-head">
            <h2>Overview</h2>
          </div>
          <div class="kv-grid">
            <div class="kv-item">
              <span class="kv-label">Image</span>
              <span class="kv-value mono">{{
                overview.image || container.image
              }}</span>
            </div>
            <div class="kv-item">
              <span class="kv-label">Command</span>
              <span class="kv-value mono">{{ overview.command || "—" }}</span>
            </div>
            <div class="kv-item">
              <span class="kv-label">Restart policy</span>
              <span class="kv-value">{{ overview.restartPolicy || "—" }}</span>
            </div>
            <div class="kv-item">
              <span class="kv-label">Platform</span>
              <span class="kv-value">{{ overview.platform || "—" }}</span>
            </div>
            <div class="kv-item">
              <span class="kv-label">Started</span>
              <span class="kv-value">{{ overview.startedAt || "—" }}</span>
            </div>
            <div class="kv-item">
              <span class="kv-label">Finished</span>
              <span class="kv-value">{{ overview.finishedAt || "—" }}</span>
            </div>
          </div>
        </article>

        <article class="panel">
          <div class="panel-head">
            <h2>Network ports</h2>
          </div>
          <div v-if="ports.length" class="port-list">
            <div v-for="port in ports" :key="port" class="port-row mono">
              {{ port }}
            </div>
          </div>
          <p v-else class="panel-empty">No published ports.</p>
        </article>

        <article class="panel panel-wide">
          <div class="panel-head">
            <h2>Environment</h2>
            <div class="env-search">
              <AppIcon name="search" :size="14" />
              <input
                v-model="envQuery"
                type="text"
                placeholder="Filter variables..."
              />
            </div>
          </div>
          <div v-if="filteredEnv.length" class="env-list">
            <div v-for="item in filteredEnv" :key="item.key" class="env-row">
              <span class="env-key mono">{{ item.key }}</span>
              <span class="env-val mono">{{ item.value }}</span>
            </div>
          </div>
          <p v-else class="panel-empty">
            No environment variables match your filter.
          </p>
        </article>
      </section>
    </template>

    <div v-else class="detail-empty">
      <h2>Container not found</h2>
      <p>This container may have been removed or you may not have access.</p>
      <router-link to="/containers" class="back-btn"
        >Back to containers</router-link
      >
    </div>

    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="showConfirm"
          class="modal-overlay"
          @click.self="showConfirm = false"
        >
          <div class="modal-content shadow-2xl">
            <div :class="['modal-icon', actionClass]">
              <AppIcon
                v-if="pendingAction === 'start'"
                name="play"
                :size="28"
                :stroke-width="2.5"
              />
              <AppIcon
                v-else-if="pendingAction === 'restart'"
                name="refresh"
                :size="28"
                :stroke-width="2.5"
              />
              <AppIcon v-else name="alert" :size="28" />
            </div>
            <div class="modal-text-center">
              <h3>Confirm operation</h3>
              <p>
                Are you sure you want to <strong>{{ pendingAction }}</strong>
                <strong>{{ container?.name }}</strong
                >?
              </p>
            </div>
            <div class="modal-actions">
              <button
                type="button"
                class="modal-btn cancel"
                @click="showConfirm = false"
              >
                Cancel
              </button>
              <button
                type="button"
                :class="['modal-btn confirm', actionClass]"
                @click="confirmAction"
              >
                Confirm
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import AppIcon from "../components/AppIcon.vue";
import { useContainers } from "../composables/useContainers";
import { apiFetch } from "../utils/apiFetch";
import { secureStorage } from "../utils/storage";
import {
  sharedState,
  showToast,
  formatBytes,
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
  actionClass,
  triggerConfirm,
  executeAction,
  formatDate,
} = useContainers({ autoPoll: true });

const inspectData = ref(null);
const inspectLoading = ref(true);
const liveStats = ref({ cpu: 0, memory: 0 });
const envQuery = ref("");
let statsTimer = null;

const container = computed(() => {
  const id = containerId.value;
  return containers.value.find(
    (c) => c.id === id || c.id.startsWith(id) || id.startsWith(c.id),
  );
});

const resolvedInspect = computed(() => {
  const data = inspectData.value;
  if (!data) return {};
  if (data.Container && typeof data.Container === "object") {
    return { ...data, ...data.Container };
  }
  return data;
});

const memLimit = computed(
  () =>
    container.value?.memLimit || resolvedInspect.value?.HostConfig?.Memory || 0,
);

const cpuLimitLabel = computed(() => {
  const limit =
    container.value?.cpuLimit || resolvedInspect.value?.HostConfig?.NanoCpus;
  if (!limit) return "No Limit";
  const cpus = typeof limit === "number" && limit > 100 ? limit / 1e9 : limit;
  return `${cpus} CPU${cpus === 1 ? "" : "s"}`;
});

const overview = computed(() => {
  const cfg = resolvedInspect.value?.Config || {};
  const state = resolvedInspect.value?.State || {};
  const host = resolvedInspect.value?.HostConfig || {};
  const cmd = Array.isArray(cfg.Cmd) ? cfg.Cmd.join(" ") : cfg.Cmd;
  return {
    image: cfg.Image,
    command: cmd || "—",
    restartPolicy: host.RestartPolicy?.Name,
    platform: resolvedInspect.value?.Platform,
    startedAt: formatInspectTime(state.StartedAt),
    finishedAt: formatInspectTime(state.FinishedAt),
  };
});

const envVars = computed(() => {
  const env = resolvedInspect.value?.Config?.Env || [];
  return env.map((line) => {
    const idx = line.indexOf("=");
    if (idx === -1) return { key: line, value: "" };
    return { key: line.slice(0, idx), value: line.slice(idx + 1) };
  });
});

const filteredEnv = computed(() => {
  const q = envQuery.value.trim().toLowerCase();
  if (!q) return envVars.value;
  return envVars.value.filter(
    (item) =>
      item.key.toLowerCase().includes(q) ||
      item.value.toLowerCase().includes(q),
  );
});

const ports = computed(() => {
  const bindings = resolvedInspect.value?.NetworkSettings?.Ports || {};
  const rows = [];
  for (const [containerPort, hostBindings] of Object.entries(bindings)) {
    if (!hostBindings?.length) {
      rows.push(containerPort);
      continue;
    }
    for (const binding of hostBindings) {
      rows.push(
        `${binding.HostIp || "0.0.0.0"}:${binding.HostPort} → ${containerPort}`,
      );
    }
  }
  return rows.sort();
});

function formatInspectTime(value) {
  if (!value || value === "0001-01-01T00:00:00Z") return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

async function fetchInspect() {
  inspectLoading.value = true;
  try {
    const token = secureStorage.getItem("token");
    const res = await apiFetch(`/api/containers/${containerId.value}/inspect`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      inspectData.value = await res.json();
    }
  } catch (err) {
    console.error(err);
  } finally {
    inspectLoading.value = false;
  }
}

async function fetchLiveStats() {
  if (!container.value || container.value.state !== "running") return;
  try {
    const token = secureStorage.getItem("token");
    const res = await apiFetch(
      `/api/containers/${containerId.value}/stats-now`,
      {
        headers: { Authorization: `Bearer ${token}` },
      },
    );
    if (res.ok) {
      const data = await res.json();
      liveStats.value = {
        cpu: Number(data.cpu) || 0,
        memory: Number(data.memory) || 0,
      };
    }
  } catch (err) {
    console.error(err);
  }
}

function startStatsPolling() {
  stopStatsPolling();
  fetchLiveStats();
  statsTimer = setInterval(fetchLiveStats, 3000);
}

function stopStatsPolling() {
  if (statsTimer) {
    clearInterval(statsTimer);
    statsTimer = null;
  }
}

async function confirmAction() {
  const action = pendingAction.value;
  await executeAction();
  if (action === "remove") {
    router.push("/containers");
    return;
  }
  await fetchContainers();
  await fetchInspect();
}

function copyText(text, label) {
  navigator.clipboard.writeText(text).then(() => {
    showToast("Copied", `${label} copied to clipboard.`, "success");
  });
}

watch(container, (value) => {
  if (value?.state === "running") startStatsPolling();
  else stopStatsPolling();
});

onMounted(async () => {
  await fetchContainers();
  await fetchInspect();
  if (container.value?.state === "running") startStatsPolling();
});

onUnmounted(() => {
  stopStatsPolling();
});
</script>

<style scoped>
.detail-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-bottom: 2rem;
}

.detail-loading {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.shimmer-block {
  height: 120px;
  border-radius: var(--radius-2xl);
  background: linear-gradient(
    90deg,
    var(--bg-input) 25%,
    var(--bg-subtle) 50%,
    var(--bg-input) 75%
  );
  background-size: 200% 100%;
  animation: shimmer 1.2s infinite;
}

.shimmer-block.wide {
  height: 280px;
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.detail-hero {
  padding: 1.35rem 1.5rem;
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border);
  background: linear-gradient(
    135deg,
    var(--bg-card) 0%,
    rgba(var(--accent-rgb), 0.04) 100%
  );
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}

.hero-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--text-dim);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 700;
}

.back-link:hover {
  color: var(--accent);
}

.state-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.7rem;
  border-radius: 999px;
  font-size: 0.72rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border: 1px solid var(--border);
}

.state-pill.is-running {
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.35);
  background: rgba(var(--success-rgb), 0.08);
}

.state-pill.is-stopped {
  color: var(--text-mute);
}

.pulse-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}

.state-pill.is-running .pulse-dot {
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.35;
  }
}

.hero-main {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.hero-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border);
  flex-shrink: 0;
}

.hero-icon.running {
  color: var(--success);
  background: rgba(var(--success-rgb), 0.08);
}

.hero-icon.stopped {
  color: var(--text-mute);
  background: var(--bg-input);
}

.hero-copy h1 {
  margin: 0 0 0.35rem;
  font-size: clamp(1.35rem, 2.5vw, 1.85rem);
  font-weight: 800;
  letter-spacing: -0.03em;
  color: var(--text-main);
}

.hero-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.45rem;
  margin-bottom: 0.35rem;
}

.id-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem 0.55rem;
  border-radius: 999px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-dim);
  font-family: var(--font-mono);
  font-size: 0.75rem;
  cursor: pointer;
}

.id-chip:hover {
  border-color: var(--border-active);
  color: var(--accent);
}

.meta-dot,
.meta-image {
  color: var(--text-mute);
  font-size: 0.85rem;
}

.hero-status {
  margin: 0;
  color: var(--text-dim);
  font-size: 0.9rem;
}

.action-rail {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
  margin-top: 1.25rem;
  padding-top: 1.15rem;
  border-top: 1px solid var(--border-subtle);
}

.action-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.55rem 0.9rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-main);
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-chip:hover {
  transform: translateY(-1px);
}

.action-chip.logs:hover {
  color: var(--accent);
  border-color: rgba(var(--accent-rgb), 0.35);
  background: var(--accent-soft);
}

.action-chip.shell:hover {
  color: #8b5cf6;
  border-color: rgba(139, 92, 246, 0.35);
  background: rgba(139, 92, 246, 0.1);
}

.action-chip.start:hover {
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.35);
  background: rgba(var(--success-rgb), 0.08);
}

.action-chip.stop:hover,
.action-chip.delete:hover {
  color: var(--error);
  border-color: rgba(var(--error-rgb), 0.35);
  background: rgba(var(--error-rgb), 0.08);
}

.action-chip.restart:hover {
  color: var(--accent);
  border-color: rgba(var(--accent-rgb), 0.35);
  background: var(--accent-soft);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.85rem;
}

.stat-card {
  padding: 1rem 1.1rem;
  border-radius: var(--radius-xl);
  border: 1px solid var(--border);
  background: var(--bg-card);
}

.stat-label {
  display: block;
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-mute);
  margin-bottom: 0.35rem;
}

.stat-value {
  display: block;
  font-size: 1.35rem;
  font-weight: 800;
  color: var(--text-main);
  font-variant-numeric: tabular-nums;
}

.stat-value.sm {
  font-size: 0.95rem;
  font-weight: 700;
}

.stat-sub {
  display: block;
  margin-top: 0.2rem;
  font-size: 0.78rem;
  color: var(--text-dim);
}

.stat-bar {
  margin-top: 0.65rem;
  height: 5px;
  border-radius: 999px;
  background: var(--bg-input);
  overflow: hidden;
}

.stat-bar-fill {
  height: 100%;
  border-radius: inherit;
}

.stat-bar-fill.accent {
  background: linear-gradient(90deg, var(--accent), #22d3ee);
}

.detail-panels {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.85rem;
}

.panel {
  border-radius: var(--radius-xl);
  border: 1px solid var(--border);
  background: var(--bg-card);
  padding: 1rem 1.1rem;
}

.panel-wide {
  grid-column: 1 / -1;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.85rem;
}

.panel-head h2 {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 800;
  color: var(--text-main);
}

.env-search {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.7rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-input);
  min-width: 220px;
}

.env-search input {
  border: none;
  outline: none;
  background: transparent;
  color: var(--text-main);
  font-size: 0.82rem;
  width: 100%;
}

.kv-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem 1rem;
}

.kv-label {
  display: block;
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--text-mute);
  margin-bottom: 0.2rem;
}

.kv-value {
  color: var(--text-main);
  font-size: 0.88rem;
  word-break: break-word;
}

.mono {
  font-family: var(--font-mono);
}

.port-list,
.env-list {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  max-height: 280px;
  overflow: auto;
}

.port-row,
.env-row {
  padding: 0.55rem 0.7rem;
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  border: 1px solid var(--border-subtle);
  font-size: 0.82rem;
}

.env-row {
  display: grid;
  grid-template-columns: minmax(120px, 34%) 1fr;
  gap: 0.75rem;
}

.env-key {
  color: var(--accent);
  font-weight: 600;
}

.env-val {
  color: var(--text-dim);
  word-break: break-all;
}

.panel-empty {
  margin: 0;
  color: var(--text-mute);
  font-size: 0.88rem;
}

.detail-empty {
  padding: 3rem 1rem;
  text-align: center;
}

.detail-empty h2 {
  margin: 0 0 0.5rem;
}

.detail-empty p {
  color: var(--text-dim);
  margin: 0 0 1rem;
}

.back-btn {
  display: inline-flex;
  padding: 0.65rem 1rem;
  border-radius: var(--radius-md);
  background: var(--accent);
  color: #fff;
  text-decoration: none;
  font-weight: 700;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(2, 6, 23, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  width: min(420px, 100%);
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  padding: 1.5rem;
}

.modal-icon {
  width: 52px;
  height: 52px;
  border-radius: 14px;
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
  margin-bottom: 1rem;
}

.modal-text-center h3 {
  margin: 0 0 0.5rem;
}

.modal-text-center p {
  margin: 0;
  color: var(--text-dim);
}

.modal-actions {
  display: flex;
  gap: 0.65rem;
}

.modal-btn {
  flex: 1;
  padding: 0.7rem 1rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  font-weight: 700;
  cursor: pointer;
}

.modal-btn.cancel {
  background: var(--bg-input);
  color: var(--text-main);
}

.modal-btn.confirm {
  color: #fff;
  border: none;
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

@media (max-width: 960px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .detail-panels,
  .kv-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .hero-main {
    flex-direction: column;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .env-row {
    grid-template-columns: 1fr;
  }

  .panel-head {
    flex-direction: column;
    align-items: stretch;
  }

  .env-search {
    min-width: 0;
    width: 100%;
  }
}
</style>

<template>
  <div class="inspect-view animate-fade-in" :class="{ 'is-fullscreen': isFullscreen }">
    <div v-if="loading && !service" class="inspect-loading">
      <div class="shimmer-block"></div>
      <div class="shimmer-block wide"></div>
    </div>

    <template v-else-if="service">
      <!-- Top Bar: Service Name, Image, Mode Badge, Actions -->
      <header class="inspect-header">
        <div class="header-left">
          <div class="title-row">
            <h1 class="container-name">{{ serviceName }}</h1>
          </div>
          <div class="image-sub">
            <span class="mono">{{ serviceImage }}</span> · Swarm Service
          </div>
          <div class="status-row">
            <span
              class="state-badge"
              :class="isHealthy ? 'state-running' : 'state-stopped'"
            >
              {{ serviceMode.toUpperCase() }} · {{ runningTasksCount }} / {{ desiredReplicas }}
            </span>
          </div>
        </div>

        <div class="header-actions">
          <router-link
            :to="{ path: '/logs', query: { rt: 'swarm', s: serviceName } }"
            class="action-btn logs-btn"
            title="Service Logs"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
              <line x1="8" y1="8" x2="16" y2="8"></line>
              <line x1="8" y1="12" x2="13" y2="12"></line>
            </svg>
            <span>Logs</span>
          </router-link>

          <button
            v-if="serviceMode === 'replicated' && canRestart"
            type="button"
            class="action-btn scale-btn"
            @click="isScaleModalOpen = true"
            title="Scale Service Replicas"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
              <polyline points="23 4 23 10 17 10"></polyline>
              <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
            </svg>
            <span>Scale</span>
          </button>

          <button
            v-if="canDelete"
            type="button"
            class="action-btn delete-btn"
            @click="triggerDelete"
            title="Delete Service"
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
            @click="router.push('/swarm')"
            title="Back to Swarm"
          >
            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.5">
              <path d="M19 12H5M12 19l-7-7 7-7"></path>
            </svg>
            Back
          </button>
        </div>
      </header>

      <!-- Section Tabs Navigation -->
      <nav class="inspect-nav">
        <button
          type="button"
          :class="['nav-tab', { active: activeSection === 'overview' }]"
          @click="activeSection = 'overview'"
        >
          Overview
        </button>
        <button
          type="button"
          :class="['nav-tab', { active: activeSection === 'tasks' }]"
          @click="activeSection = 'tasks'"
        >
          Tasks ({{ tasks.length }})
        </button>
        <button
          type="button"
          :class="['nav-tab', { active: activeSection === 'env' }]"
          @click="activeSection = 'env'"
        >
          Environment ({{ envVars.length }})
        </button>
        <button
          type="button"
          :class="['nav-tab', { active: activeSection === 'network' }]"
          @click="activeSection = 'network'"
        >
          Network & Ports
        </button>
        <button
          type="button"
          :class="['nav-tab', { active: activeSection === 'raw' }]"
          @click="activeSection = 'raw'"
        >
          Raw Inspect
        </button>
      </nav>

      <!-- TAB: Overview -->
      <div v-if="activeSection === 'overview'" class="tab-content">
        <!-- Overview Card -->
        <section class="inspect-section">
          <h2 class="section-heading">OVERVIEW</h2>
          <div class="card overview-card">
            <div class="overview-grid">
              <div class="overview-cell">
                <span class="cell-label">Replicas</span>
                <span class="cell-value mono highlight">{{ runningTasksCount }} / {{ desiredReplicas }}</span>
              </div>
              <div class="overview-cell">
                <span class="cell-label">Mode</span>
                <span class="cell-value capitalize">{{ serviceMode }}</span>
              </div>
              <div class="overview-cell">
                <span class="cell-label">Created</span>
                <span class="cell-value">{{ formattedCreated }}</span>
              </div>
              <div class="overview-cell">
                <span class="cell-label">Spec Version</span>
                <span class="cell-value mono">{{ service.Version?.Index || 1 }}</span>
              </div>
              <div class="overview-cell">
                <span class="cell-label">Service ID</span>
                <span
                  class="cell-value mono clickable-id"
                  @click="copyText(service.ID, 'Service ID')"
                  title="Click to copy ID"
                >
                  {{ service.ID?.substring(0, 12) }}
                </span>
              </div>
              <div class="overview-cell">
                <span class="cell-label">Image</span>
                <span class="cell-value mono">{{ serviceImage }}</span>
              </div>
              <div class="overview-cell">
                <span class="cell-label">Restart Condition</span>
                <span class="cell-value mono capitalize">{{ service.Spec?.TaskTemplate?.RestartPolicy?.Condition || 'any' }}</span>
              </div>
              <div class="overview-cell" v-if="service.Spec?.UpdateConfig">
                <span class="cell-label">Update Strategy</span>
                <span class="cell-value mono">Parallel: {{ service.Spec.UpdateConfig.Parallelism || 1 }}</span>
              </div>
            </div>

            <div class="command-block" v-if="service.Spec?.TaskTemplate?.ContainerSpec?.Command || service.Spec?.TaskTemplate?.ContainerSpec?.Args">
              <span class="command-label">Command & Args</span>
              <div class="command-box mono">
                <code>{{ (service.Spec.TaskTemplate.ContainerSpec.Command || []).concat(service.Spec.TaskTemplate.ContainerSpec.Args || []).join(' ') }}</code>
              </div>
            </div>
          </div>
        </section>

        <!-- Service Labels Card -->
        <section class="inspect-section" v-if="labelsList.length > 0">
          <h2 class="section-heading">LABELS ({{ labelsList.length }})</h2>
          <div class="card labels-card">
            <div class="labels-table">
              <div v-for="item in labelsList" :key="item.k" class="label-row">
                <div class="label-key mono">{{ item.k }}</div>
                <div class="label-val mono">{{ item.v }}</div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- TAB: Tasks / Replicas -->
      <div v-if="activeSection === 'tasks'" class="tab-content">
        <section class="inspect-section">
          <h2 class="section-heading">REPLICA TASKS ({{ tasks.length }})</h2>
          <div class="card containers-card">
            <div v-if="tasks.length === 0" class="empty-field-text mono">
              No tasks found for this service.
            </div>
            <div
              v-else
              v-for="t in tasks"
              :key="t.id"
              class="pod-container-row"
            >
              <div class="p-container-header">
                <div class="p-c-name-side">
                  <span
                    class="green-status-dot"
                    :style="{ background: t.current_state === 'running' ? 'var(--success)' : 'var(--text-mute)' }"
                  ></span>
                  <span class="conn-name mono bold">#{{ t.slot || 1 }}</span>
                  <span class="mount-type-tag mono">{{ t.id.slice(0, 12) }}</span>
                  <span class="node-badge mono">
                    <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
                      <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
                      <line x1="6" y1="6" x2="6.01" y2="6"></line>
                      <line x1="6" y1="18" x2="6.01" y2="18"></line>
                    </svg>
                    {{ t.node_hostname || t.node_id?.slice(0, 10) || 'Node' }}
                  </span>
                  <span
                    class="state-pill"
                    :class="t.current_state === 'running' ? 'pill-success' : 'pill-neutral'"
                  >
                    {{ t.current_state }}
                  </span>
                </div>
                <div class="p-c-meta-side mono">
                  <router-link
                    :to="{ path: '/logs', query: { rt: 'swarm', s: serviceName } }"
                    class="action-btn-micro"
                    title="View Logs"
                  >
                    <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2">
                      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                    </svg>
                    <span>Logs</span>
                  </router-link>
                </div>
              </div>

              <div class="p-container-body mono">
                <div class="p-c-grid">
                  <div>
                    <span class="source-label">Desired State: </span>
                    <span class="source-val capitalize">{{ t.desired_state }}</span>
                  </div>
                  <div>
                    <span class="source-label">Updated: </span>
                    <span class="source-val">{{ formatAge(t.updated_at) }}</span>
                  </div>
                  <div v-if="t.error || t.message">
                    <span class="source-label">Message: </span>
                    <span class="source-val text-error">{{ t.error || t.message }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- TAB: Environment -->
      <div v-if="activeSection === 'env'" class="tab-content">
        <section class="inspect-section">
          <h2 class="section-heading">ENVIRONMENT ({{ envVars.length }})</h2>
          <div class="card env-card" :class="{ 'is-empty': envVars.length === 0 }">
            <div v-if="envVars.length > 0" class="env-table">
              <div
                v-for="item in envVars"
                :key="item.key"
                class="env-row"
              >
                <div class="env-key mono">
                  <span class="key-name">{{ item.key }}</span>
                </div>
                <div class="env-val mono" :class="{ 'is-masked': item.value === '••••••••••••' }">
                  {{ item.value }}
                </div>
              </div>
            </div>
            <div v-else class="empty-field-text mono">
              No environment variables defined for this service.
            </div>
          </div>
          <p v-if="isMasked" class="env-disclaimer">
            Environment variables are hidden. Reading them requires shell access.
          </p>
        </section>
      </div>

      <!-- TAB: Network & Ports -->
      <div v-if="activeSection === 'network'" class="tab-content">
        <section class="inspect-section">
          <h2 class="section-heading">PUBLISHED PORTS</h2>
          <div class="card ports-card">
            <div v-if="portsList.length > 0" class="mounts-table">
              <div
                v-for="(p, idx) in portsList"
                :key="idx"
                class="mount-row"
              >
                <div class="mount-main">
                  <div class="mount-path-row">
                    <span class="mount-dest mono bold">{{ p.published }}:{{ p.target }}</span>
                    <span class="mount-type-tag mono">{{ p.protocol }}</span>
                    <span class="mount-mode-tag mono">{{ p.publishMode }}</span>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="empty-field-text mono">
              No published ports configured.
            </div>
          </div>
        </section>

        <section class="inspect-section" v-if="virtualIPs.length > 0">
          <h2 class="section-heading">VIRTUAL IPS</h2>
          <div class="card mounts-card">
            <div class="mounts-table">
              <div
                v-for="vip in virtualIPs"
                :key="vip.networkID"
                class="mount-row"
              >
                <div class="mount-main">
                  <div class="mount-path-row">
                    <span class="mount-dest mono bold text-accent">{{ vip.addr }}</span>
                    <span class="source-label mono">Network ID:</span>
                    <span class="mono">{{ vip.networkID.slice(0, 12) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- TAB: Raw Inspect -->
      <div v-if="activeSection === 'raw'" class="tab-content">
        <section class="inspect-section">
          <div class="raw-header-row">
            <h2 class="section-heading">RAW DOCKER INSPECT JSON</h2>
            <button class="action-btn-micro" @click="copyRawJson">
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
              </svg>
              <span>{{ isCopied ? 'Copied!' : 'Copy JSON' }}</span>
            </button>
          </div>
          <div class="card raw-card">
            <pre class="raw-json mono"><code>{{ formattedRawJson }}</code></pre>
          </div>
        </section>
      </div>

      <!-- Scale Modal -->
      <Teleport to="body">
        <div v-if="isScaleModalOpen" class="modal-backdrop" @click="isScaleModalOpen = false">
          <div class="modal-card" @click.stop>
            <h3>Scale Service Replicas</h3>
            <p class="modal-sub">Set desired number of replicas for <strong>{{ serviceName }}</strong></p>
            <div class="scale-input-row">
              <button class="stepper-btn" @click="targetReplicas = Math.max(0, targetReplicas - 1)">-</button>
              <input type="number" min="0" max="100" v-model.number="targetReplicas" class="scale-number-input" />
              <button class="stepper-btn" @click="targetReplicas = targetReplicas + 1">+</button>
            </div>
            <div class="modal-actions">
              <button class="action-btn back-btn" @click="isScaleModalOpen = false">Cancel</button>
              <button class="action-btn scale-btn-apply" :disabled="scalingLoading" @click="submitScale">
                {{ scalingLoading ? 'Scaling...' : 'Apply Scale' }}
              </button>
            </div>
          </div>
        </div>
      </Teleport>
    </template>

    <div v-else class="inspect-empty">
      <h2>Service Not Found</h2>
      <p>The requested Swarm service could not be located on the cluster.</p>
      <button type="button" @click="router.push('/swarm')" class="back-pill-btn">Back to Swarm</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { apiFetch } from '../utils/apiFetch';
import { sharedState, userCanRestart, userCanDelete, showToast } from '../utils/sharedState';

const route = useRoute();
const router = useRouter();

const service = ref(null);
const tasks = ref([]);
const loading = ref(true);
const activeSection = ref('overview');
const isFullscreen = ref(false);
const isScaleModalOpen = ref(false);
const targetReplicas = ref(1);
const scalingLoading = ref(false);
const isCopied = ref(false);

const canRestart = computed(() => userCanRestart(sharedState.currentUser));
const canDelete = computed(() => userCanDelete(sharedState.currentUser));

const serviceName = computed(() => service.value?.Spec?.Name || route.params.id);
const serviceImage = computed(() => service.value?.Spec?.TaskTemplate?.ContainerSpec?.Image || '—');
const serviceMode = computed(() => {
  if (service.value?.Spec?.Mode?.Global) return 'global';
  return 'replicated';
});
const desiredReplicas = computed(() => {
  if (serviceMode.value === 'global') return 'global';
  return service.value?.Spec?.Mode?.Replicated?.Replicas ?? 1;
});
const runningTasksCount = computed(() => {
  return tasks.value.filter((t) => t.current_state === 'running').length;
});
const isHealthy = computed(() => {
  if (serviceMode.value === 'global') return runningTasksCount.value > 0;
  return runningTasksCount.value >= desiredReplicas.value && desiredReplicas.value > 0;
});

const formattedCreated = computed(() => {
  if (!service.value?.CreatedAt) return '—';
  return new Date(service.value.CreatedAt).toLocaleString();
});

const labelsList = computed(() => {
  const lbls = service.value?.Spec?.Labels || {};
  return Object.entries(lbls).map(([k, v]) => ({ k, v }));
});

const envVars = computed(() => {
  const envList = service.value?.Spec?.TaskTemplate?.ContainerSpec?.Env || [];
  return envList.map((e) => {
    const idx = e.indexOf('=');
    if (idx === -1) return { key: e, value: '' };
    return { key: e.slice(0, idx), value: e.slice(idx + 1) };
  });
});

const isMasked = computed(() => {
  return envVars.value.some((e) => e.value === '••••••••••••');
});

const portsList = computed(() => {
  const rawPorts = service.value?.Endpoint?.Ports || [];
  return rawPorts.map((p) => ({
    published: p.PublishedPort,
    target: p.TargetPort,
    protocol: p.Protocol,
    publishMode: p.PublishMode || 'ingress',
  }));
});

const virtualIPs = computed(() => {
  const rawVIPs = service.value?.Endpoint?.VirtualIPs || [];
  return rawVIPs.map((v) => ({
    networkID: v.NetworkID,
    addr: v.Addr,
  }));
});

const formattedRawJson = computed(() => {
  return JSON.stringify(service.value, null, 2);
});

function copyText(val, label = 'Copied') {
  if (!val) return;
  navigator.clipboard.writeText(String(val));
  showToast(label, 'Copied to clipboard', 'info');
}

function formatAge(unixSeconds) {
  if (!unixSeconds) return '—';
  const now = Math.floor(Date.now() / 1000);
  const diff = Math.max(0, now - unixSeconds);
  const days = Math.floor(diff / 86400);
  const hours = Math.floor((diff % 86400) / 3600);
  const minutes = Math.floor((diff % 3600) / 60);

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m ago`;
}

async function loadService() {
  loading.value = true;
  try {
    const id = route.params.id;
    const [svcRes, tasksRes] = await Promise.all([
      apiFetch(`/api/swarm/services/${id}`),
      apiFetch(`/api/swarm/tasks?service=${id}`).catch(() => null),
    ]);
    if (!svcRes.ok) {
      throw new Error(`Failed to load service (${svcRes.status})`);
    }
    const svcData = await svcRes.json();
    const tasksData = tasksRes && tasksRes.ok ? await tasksRes.json() : [];
    service.value = svcData;
    tasks.value = Array.isArray(tasksData) ? tasksData : [];
    if (svcData?.Spec?.Mode?.Replicated?.Replicas != null) {
      targetReplicas.value = svcData.Spec.Mode.Replicated.Replicas;
    }
  } catch (err) {
    console.error('Failed to load Swarm service details:', err);
    service.value = null;
  } finally {
    loading.value = false;
  }
}

async function submitScale() {
  scalingLoading.value = true;
  try {
    const id = route.params.id;
    const res = await apiFetch(`/api/swarm/services/${id}/scale`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ replicas: targetReplicas.value }),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Scale failed');
    }
    showToast('Service Scaled', `Scaled to ${targetReplicas.value} replicas`, 'success');
    isScaleModalOpen.value = false;
    await loadService();
  } catch (err) {
    showToast('Scaling Failed', err.message, 'error');
  } finally {
    scalingLoading.value = false;
  }
}

async function triggerDelete() {
  if (!confirm(`Are you sure you want to delete swarm service "${serviceName.value}"?`)) {
    return;
  }
  try {
    const id = route.params.id;
    const res = await apiFetch(`/api/swarm/services/${id}`, { method: 'DELETE' });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Delete failed');
    }
    showToast('Service Deleted', `Service ${serviceName.value} removed.`, 'success');
    router.push('/swarm');
  } catch (err) {
    showToast('Delete Failed', err.message, 'error');
  }
}

function copyRawJson() {
  navigator.clipboard.writeText(formattedRawJson.value);
  isCopied.value = true;
  setTimeout(() => (isCopied.value = false), 2000);
}

onMounted(() => {
  loadService();
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
  text-decoration: none;
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

.action-btn.scale-btn:hover {
  color: #06b6d4;
  border-color: #06b6d4;
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

/* Nav Tabs */
.inspect-nav {
  display: flex;
  gap: 0.35rem;
  border-bottom: 1px solid var(--border);
  margin-bottom: 2rem;
  overflow-x: auto;
}

.nav-tab {
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  padding: 0.65rem 1rem;
  color: var(--text-mute);
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.nav-tab:hover {
  color: var(--text-main);
}

.nav-tab.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
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

/* Cards */
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

/* Overview Card */
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

.cell-value.highlight {
  color: var(--accent);
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

.capitalize {
  text-transform: capitalize;
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

/* Replica Tasks Cards */
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
  gap: 0.55rem;
  font-size: 0.88rem;
  flex-wrap: wrap;
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

.node-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  color: var(--text-dim);
  background: var(--bg-input);
  padding: 0.12rem 0.45rem;
  border-radius: 4px;
  border: 1px solid var(--border-subtle);
}

.state-pill {
  display: inline-block;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.pill-success {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.pill-neutral {
  background: var(--bg-subtle);
  color: var(--text-mute);
  border: 1px solid var(--border);
}

.p-c-meta-side {
  font-size: 0.78rem;
  color: var(--text-mute);
}

.action-btn-micro {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem 0.55rem;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 6px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-main);
  cursor: pointer;
  text-decoration: none;
  transition: all 0.15s ease;
}

.action-btn-micro:hover {
  background: var(--card-hover);
  border-color: var(--border-active);
  color: var(--accent);
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

.source-label {
  color: var(--text-mute);
  font-weight: 600;
}

.source-val {
  color: var(--text-dim);
}

.text-error {
  color: #ef4444;
}

/* Labels Card */
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

/* Environment Card */
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

/* Mounts / Ports table styling */
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

.text-accent {
  color: var(--accent);
}

/* Raw Inspect Card */
.raw-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.65rem;
}

.raw-header-row .section-heading {
  margin: 0;
}

.raw-card {
  padding: 1.25rem;
}

.raw-json {
  margin: 0;
  font-size: 0.82rem;
  line-height: 1.5;
  color: var(--text-main);
  white-space: pre-wrap;
  word-break: break-all;
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

.inspect-empty {
  text-align: center;
  padding: 4rem 1rem;
  color: var(--text-mute);
}

.inspect-empty h2 {
  font-size: 1.5rem;
  color: var(--text-main);
  margin-bottom: 0.5rem;
}

.inspect-empty p {
  margin-bottom: 1.5rem;
  font-size: 0.95rem;
}

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.modal-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.75rem 2rem;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
}

.modal-card h3 {
  margin: 0 0 0.5rem 0;
  color: var(--text-main);
  font-size: 1.2rem;
  font-weight: 700;
}

.modal-sub {
  color: var(--text-dim);
  font-size: 0.85rem;
  margin: 0 0 1.5rem 0;
}

.scale-input-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin: 1.5rem 0;
}

.stepper-btn {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  color: var(--text-main);
  font-size: 1.25rem;
  font-weight: 700;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stepper-btn:hover {
  background: var(--card-hover);
  border-color: var(--border-active);
}

.scale-number-input {
  width: 90px;
  text-align: center;
  font-size: 1.4rem;
  font-weight: 800;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text-main);
  padding: 0.35rem 0.5rem;
  font-family: var(--font-mono);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.scale-btn-apply {
  background: var(--accent);
  color: #000;
  border-color: var(--accent);
  font-weight: 700;
}

.scale-btn-apply:hover {
  filter: brightness(1.1);
}
</style>

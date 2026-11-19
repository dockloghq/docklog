<template>
  <div class="swarm-view animate-fade-in">
    <div v-if="!isSwarmActive" class="swarm-unavailable-banner">
      <div class="banner-icon">
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
      </div>
      <div>
        <strong>Docker Swarm is not active on this host.</strong>
        <p>
          Run <code>docker swarm init</code> to initialize a Swarm cluster or <code>docker swarm join</code> to join an existing cluster.
        </p>
      </div>
    </div>
    <div v-else-if="fetchError" class="swarm-unavailable-banner error-banner">
      <div class="banner-icon">
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="15" y1="9" x2="9" y2="15"></line>
          <line x1="9" y1="9" x2="15" y2="15"></line>
        </svg>
      </div>
      <div>
        <strong>Could not load Docker Swarm data.</strong>
        <p>{{ fetchError }}</p>
      </div>
    </div>

    <!-- Hero Section matching Containers.vue design system -->
    <section class="swarm-hero">
      <div class="hero-copy">
        <span class="hero-eyebrow">Cluster Control</span>
        <h1>Docker Swarm</h1>
        <p class="hero-sub">
          Manage services, scale replicas, inspect manager/worker nodes, and track cluster tasks.
        </p>
      </div>
      <div class="hero-stats" v-if="swarmOverview && swarmOverview.active">
        <div class="hero-stat">
          <span class="hero-stat-val">{{ swarmOverview.services_count }}</span>
          <span class="hero-stat-lbl">Services</span>
        </div>
        <div class="hero-stat success">
          <span class="hero-stat-val">{{ swarmOverview.running_tasks_count }}</span>
          <span class="hero-stat-lbl">Running Tasks</span>
        </div>
        <div class="hero-stat">
          <span class="hero-stat-val">{{ swarmOverview.nodes_count }}</span>
          <span class="hero-stat-lbl">Nodes</span>
        </div>
        <div class="hero-stat muted">
          <span class="hero-stat-val">{{ swarmOverview.managers_count }}</span>
          <span class="hero-stat-lbl">Managers</span>
        </div>
        <div class="hero-stat muted">
          <span class="hero-stat-val">{{ swarmOverview.workers_count }}</span>
          <span class="hero-stat-lbl">Workers</span>
        </div>
      </div>
      <div class="hero-mesh" aria-hidden="true"></div>
    </section>

    <!-- Main Panel -->
    <section class="swarm-panel">
      <div class="panel-toolbar">
        <div class="toolbar-left">
          <div class="filter-pills">
            <button
              class="filter-pill"
              :class="{ active: activeTab === 'services' }"
              @click="switchTab('services')"
            >
              Services
              <span class="pill-count">{{ swarmServices.length }}</span>
            </button>
            <button
              class="filter-pill"
              :class="{ active: activeTab === 'nodes' }"
              @click="switchTab('nodes')"
            >
              Nodes
              <span class="pill-count">{{ swarmNodes.length }}</span>
            </button>
            <button
              class="filter-pill"
              :class="{ active: activeTab === 'tasks' }"
              @click="switchTab('tasks')"
            >
              Tasks
              <span class="pill-count">{{ swarmTasks.length }}</span>
            </button>
          </div>
        </div>

        <div class="toolbar-right">
          <div class="search-box">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <input
              type="text"
              v-model="searchFilter"
              placeholder="Filter by name, image, or node..."
            />
            <button v-if="searchFilter" class="clear-search-btn" @click="searchFilter = ''">×</button>
          </div>
          <button class="refresh-btn" @click="fetchSwarmData" :disabled="loading" title="Refresh">
            <svg
              viewBox="0 0 24 24"
              width="15"
              height="15"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              :class="{ spinning: loading }"
            >
              <path d="M23 4v6h-6"></path>
              <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
            </svg>
          </button>
        </div>
      </div>

      <!-- Services Tab -->
      <div v-if="activeTab === 'services'" class="premium-table-container embedded">
        <div v-if="filteredServices.length === 0" class="empty-state">
          <p>No Swarm services found.</p>
        </div>
        <table v-else class="premium-table">
          <thead>
            <tr>
              <th style="min-width: 180px;">Service</th>
              <th>Image</th>
              <th>Mode</th>
              <th>Replicas</th>
              <th>Ports</th>
              <th>Created</th>
              <th class="text-right" style="min-width: 160px;">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="svc in filteredServices" :key="svc.id" class="clickable-row">
              <td class="name-cell">
                <router-link :to="`/swarm/services/${svc.id}`" class="service-link">
                  <div class="container-title" :title="svc.name">{{ svc.name }}</div>
                  <div class="container-id">{{ svc.id.substring(0, 12) }}</div>
                </router-link>
              </td>
              <td class="image-cell">
                <div class="image-pill-box" :title="svc.image">
                  <span class="image-name">{{ formatImageName(svc.image) }}</span>
                  <span v-if="formatImageTag(svc.image)" class="image-tag">{{ formatImageTag(svc.image) }}</span>
                </div>
              </td>
              <td>
                <span class="badge" :class="svc.mode === 'global' ? 'badge-info' : 'badge-primary'">
                  {{ svc.mode }}
                </span>
              </td>
              <td>
                <div class="replica-status">
                  <span
                    class="status-indicator"
                    :class="svc.running === svc.replicas && svc.replicas > 0 ? 'status-up' : (svc.running > 0 ? 'status-partial' : 'status-down')"
                  ></span>
                  <strong class="replica-count">{{ svc.running }} / {{ svc.mode === 'global' ? 'global' : svc.replicas }}</strong>
                </div>
              </td>
              <td>
                <span v-if="svc.ports && svc.ports.length > 0" class="ports-badge">
                  {{ svc.ports.join(', ') }}
                </span>
                <span v-else class="text-muted">—</span>
              </td>
              <td class="date-cell">
                <span class="date-label">{{ formatSwarmDate(svc.created_at) }}</span>
              </td>
              <td class="text-right action-cell">
                <div class="action-group justify-end" @click.stop>
                  <div class="action-cluster">
                    <button
                      v-if="svc.mode === 'replicated' && canRestart"
                      class="icon-btn scale-btn"
                      @click="openScaleModal(svc)"
                      data-tooltip="Scale replicas"
                      type="button"
                    >
                      <AppIcon name="refresh" :size="15" :stroke-width="2.5" />
                    </button>
                    <button
                      class="icon-btn tasks-btn"
                      @click="viewServiceTasks(svc.name)"
                      data-tooltip="View tasks"
                      type="button"
                    >
                      <AppIcon name="containers" :size="15" />
                    </button>
                    <button
                      class="icon-btn logs-btn"
                      @click="viewServiceLogs(svc.id, svc.name)"
                      data-tooltip="Service logs"
                      type="button"
                    >
                      <AppIcon name="logsBubble" :size="15" />
                    </button>
                    <button
                      v-if="canDelete"
                      class="icon-btn delete-btn"
                      @click="confirmDeleteService(svc)"
                      data-tooltip="Delete service"
                      type="button"
                    >
                      <AppIcon name="trash" :size="15" />
                    </button>
                  </div>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Nodes Tab -->
      <div v-if="activeTab === 'nodes'" class="premium-table-container embedded">
        <div v-if="filteredNodes.length === 0" class="empty-state">
          <p>No Swarm nodes found.</p>
        </div>
        <table v-else class="premium-table">
          <thead>
            <tr>
              <th>Hostname</th>
              <th>Role</th>
              <th>Status</th>
              <th>Availability</th>
              <th>Engine</th>
              <th>CPUs / Memory</th>
              <th class="text-right">Manage</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="node in filteredNodes" :key="node.id">
              <td class="name-cell">
                <div class="container-title" :title="node.hostname">
                  {{ node.hostname }}
                  <span v-if="node.leader" class="leader-badge">Leader</span>
                </div>
                <div class="container-id">{{ node.id.substring(0, 12) }}</div>
              </td>
              <td>
                <span class="badge" :class="node.manager ? 'badge-manager' : 'badge-worker'">
                  {{ node.role }}
                </span>
              </td>
              <td>
                <span class="status-pill" :class="node.state === 'ready' ? 'status-ready' : 'status-down'">
                  {{ node.state }}
                </span>
              </td>
              <td>
                <span class="avail-pill" :class="'avail-' + node.availability">
                  {{ node.availability }}
                </span>
              </td>
              <td class="text-muted mono-text">
                {{ node.engine_version || '—' }}
              </td>
              <td class="text-muted">
                {{ node.cpus }} CPUs / {{ formatMemory(node.memory) }}
              </td>
              <td class="text-right">
                <div class="availability-actions" v-if="isAdmin">
                  <select
                    :value="node.availability"
                    @change="handleAvailabilityChange(node, $event.target.value)"
                    class="avail-select"
                  >
                    <option value="active">Active</option>
                    <option value="pause">Pause</option>
                    <option value="drain">Drain</option>
                  </select>
                </div>
                <span v-else class="text-muted">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Tasks Tab -->
      <div v-if="activeTab === 'tasks'" class="premium-table-container embedded">
        <div v-if="filteredTasks.length === 0" class="empty-state">
          <p>No tasks matching current filter.</p>
        </div>
        <table v-else class="premium-table">
          <thead>
            <tr>
              <th>Service / Slot</th>
              <th>Task ID</th>
              <th>Node</th>
              <th>Desired State</th>
              <th>Current State</th>
              <th>Updated</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in filteredTasks" :key="task.id">
              <td class="name-cell">
                <div class="container-title">{{ task.service_name }}</div>
                <div class="sub-text">Slot {{ task.slot || 1 }}</div>
              </td>
              <td class="mono-text text-muted">
                {{ task.id.substring(0, 12) }}
              </td>
              <td>
                {{ task.node_hostname || (task.node_id ? task.node_id.substring(0, 8) : '—') }}
              </td>
              <td>
                <span class="state-pill">{{ task.desired_state }}</span>
              </td>
              <td>
                <span
                  class="state-pill"
                  :class="{
                    'state-running': task.current_state === 'running',
                    'state-failed': task.current_state === 'failed' || task.current_state === 'rejected',
                    'state-shutdown': task.current_state === 'shutdown' || task.current_state === 'complete'
                  }"
                >
                  {{ task.current_state }}
                </span>
                <span v-if="task.error" class="task-err" :title="task.error">⚠️ {{ task.error }}</span>
              </td>
              <td class="date-cell">
                <span class="date-label">{{ formatSwarmDate(task.updated_at) }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Scale Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="scalingService" class="modal-overlay" @click.self="scalingService = null">
          <div class="modal-content shadow-2xl scale-modal-content">
            <div class="scale-modal-header">
              <div class="scale-header-icon">
                <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2">
                  <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
                  <polyline points="2 17 12 22 22 17"></polyline>
                  <polyline points="2 12 12 17 22 12"></polyline>
                </svg>
              </div>
              <div class="scale-header-title">
                <h3>Scale Service</h3>
                <span class="scale-service-chip">{{ scalingService.name }}</span>
              </div>
              <button class="scale-close-btn" @click="scalingService = null">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>

            <p class="scale-modal-sub">
              Set the desired number of running replicas for this Swarm service. Tasks will automatically be rebalanced across available cluster nodes.
            </p>

            <div class="scale-stepper-wrap">
              <label for="replicas-input" class="scale-input-label">Target Replicas</label>
              <div class="scale-counter-box">
                <button
                  type="button"
                  class="stepper-btn"
                  @click="targetReplicas = Math.max(0, targetReplicas - 1)"
                  :disabled="targetReplicas <= 0 || scaleLoading"
                >
                  −
                </button>
                <input
                  id="replicas-input"
                  type="number"
                  min="0"
                  max="1000"
                  v-model.number="targetReplicas"
                  class="scale-input-number"
                  :disabled="scaleLoading"
                />
                <button
                  type="button"
                  class="stepper-btn"
                  @click="targetReplicas = targetReplicas + 1"
                  :disabled="scaleLoading"
                >
                  +
                </button>
              </div>
            </div>

            <div class="scale-actions">
              <button class="scale-btn-cancel" @click="scalingService = null" :disabled="scaleLoading">
                Cancel
              </button>
              <button class="scale-btn-submit" @click="applyScale" :disabled="scaleLoading">
                {{ scaleLoading ? 'Scaling Service...' : 'Apply Scale' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Live Logs Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="activeLogService" class="modal-overlay" @click.self="closeServiceLogs">
          <div class="modal-content shadow-2xl logs-modal-box">
            <div class="logs-modal-header">
              <div>
                <h3>Live Logs: {{ activeLogService.name }}</h3>
                <span class="container-id">{{ activeLogService.id }}</span>
              </div>
              <button class="scale-close-btn" @click="closeServiceLogs">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>
            <div class="logs-stream-container">
              <div v-if="serviceLogLines.length === 0" class="logs-loading">
                Connecting and streaming service logs...
              </div>
              <div v-for="(line, idx) in serviceLogLines" :key="idx" class="log-row" :class="line.stream">
                <span class="log-time">{{ line.timestamp }}</span>
                <span class="log-text">{{ line.message }}</span>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import AppIcon from '../components/AppIcon.vue';
import {
  useSwarm,
  startSwarmPolling,
  stopSwarmPolling,
  formatSwarmDate,
  formatMemory,
} from '../composables/useSwarm';
import {
  sharedState,
  showToast,
  userCanRestart,
  userCanDelete,
} from '../utils/sharedState';
import { secureStorage } from '../utils/storage';

const {
  swarmOverview,
  swarmServices,
  swarmNodes,
  swarmTasks,
  loading,
  fetchError,
  fetchSwarmData,
  scaleService,
  deleteService,
  updateNodeAvailability,
} = useSwarm();

const router = useRouter();

const activeTab = ref('services');
const searchFilter = ref('');

function switchTab(tab) {
  activeTab.value = tab;
  searchFilter.value = '';
}

const scalingService = ref(null);
const targetReplicas = ref(1);
const scaleLoading = ref(false);

const activeLogService = ref(null);
const serviceLogLines = ref([]);
let logSocket = null;

const isSwarmActive = computed(() => {
  return swarmOverview.value ? swarmOverview.value.active : sharedState.swarmActive;
});

const isAdmin = computed(() => {
  return sharedState.currentUser?.is_admin === true || sharedState.isAuthDisabled;
});

const canRestart = computed(() => {
  return userCanRestart(sharedState.currentUser);
});

const canDelete = computed(() => {
  return userCanDelete(sharedState.currentUser);
});

function formatImageName(fullImage) {
  if (!fullImage) return 'N/A';
  // Strip sha256 digest if present
  let img = fullImage.split('@')[0];
  return img.split(':')[0] || img;
}

function formatImageTag(fullImage) {
  if (!fullImage) return '';
  let img = fullImage.split('@')[0];
  const parts = img.split(':');
  if (parts.length > 1) {
    return parts[1];
  }
  return 'latest';
}

const filteredServices = computed(() => {
  const q = searchFilter.value.trim().toLowerCase();
  if (!q) return swarmServices.value;
  return swarmServices.value.filter(
    (s) =>
      s.name.toLowerCase().includes(q) ||
      s.image.toLowerCase().includes(q) ||
      s.id.toLowerCase().includes(q)
  );
});

const filteredNodes = computed(() => {
  const q = searchFilter.value.trim().toLowerCase();
  if (!q) return swarmNodes.value;
  return swarmNodes.value.filter(
    (n) =>
      n.hostname.toLowerCase().includes(q) ||
      n.role.toLowerCase().includes(q) ||
      n.id.toLowerCase().includes(q) ||
      n.availability.toLowerCase().includes(q)
  );
});

const filteredTasks = computed(() => {
  const q = searchFilter.value.trim().toLowerCase();
  if (!q) return swarmTasks.value;
  return swarmTasks.value.filter(
    (t) =>
      t.service_name.toLowerCase().includes(q) ||
      (t.node_hostname && t.node_hostname.toLowerCase().includes(q)) ||
      t.current_state.toLowerCase().includes(q) ||
      t.id.toLowerCase().includes(q)
  );
});

function openScaleModal(svc) {
  scalingService.value = svc;
  targetReplicas.value = svc.replicas;
}

async function applyScale() {
  if (!scalingService.value) return;
  scaleLoading.value = true;
  try {
    await scaleService(scalingService.value.id, targetReplicas.value);
    showToast('Success', `Scaled ${scalingService.value.name} to ${targetReplicas.value} replicas`, 'success');
    scalingService.value = null;
  } catch (err) {
    showToast('Scale Failed', err.message, 'error');
  } finally {
    scaleLoading.value = false;
  }
}

async function confirmDeleteService(svc) {
  if (!confirm(`Are you sure you want to delete Swarm service "${svc.name}"?`)) {
    return;
  }
  try {
    await deleteService(svc.id);
    showToast('Service Deleted', `Deleted service ${svc.name}`, 'success');
  } catch (err) {
    showToast('Delete Failed', err.message, 'error');
  }
}

async function handleAvailabilityChange(node, newAvail) {
  try {
    await updateNodeAvailability(node.id, newAvail);
    showToast('Node Updated', `${node.hostname} availability set to ${newAvail}`, 'success');
  } catch (err) {
    showToast('Update Failed', err.message, 'error');
  }
}

function viewServiceTasks(serviceName) {
  activeTab.value = 'tasks';
  searchFilter.value = serviceName;
}

function viewServiceLogs(serviceId) {
  router.push({ path: '/logs', query: { s: serviceId } });
}

function closeServiceLogs() {
  if (logSocket) {
    logSocket.close();
    logSocket = null;
  }
  activeLogService.value = null;
  serviceLogLines.value = [];
}

onMounted(() => {
  startSwarmPolling();
});

onUnmounted(() => {
  stopSwarmPolling();
  closeServiceLogs();
});
</script>

<style scoped>
.swarm-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-bottom: 2rem;
}

.swarm-unavailable-banner {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1rem 1.25rem;
  border-radius: var(--radius-lg);
  background: rgba(var(--warning-rgb, 234, 179, 8), 0.1);
  border: 1px solid rgba(var(--warning-rgb, 234, 179, 8), 0.3);
  color: var(--text-main);
}

.swarm-unavailable-banner.error-banner {
  background: rgba(var(--error-rgb, 239, 68, 68), 0.1);
  border-color: rgba(var(--error-rgb, 239, 68, 68), 0.3);
}

.banner-icon {
  color: var(--warning, #eab308);
  flex-shrink: 0;
  margin-top: 2px;
}

.error-banner .banner-icon {
  color: var(--error, #ef4444);
}

.swarm-hero {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1.5rem;
  flex-wrap: wrap;
  padding: 1.5rem 1.75rem;
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border);
  background: linear-gradient(135deg, var(--bg-card) 0%, var(--bg-card) 55%, rgba(var(--accent-rgb), 0.04) 100%);
  overflow: hidden;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}

.hero-mesh {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 70% 80% at 100% 0%, rgba(var(--accent-rgb), 0.14), transparent 55%),
    radial-gradient(ellipse 40% 50% at 0% 100%, rgba(var(--success-rgb), 0.06), transparent 50%);
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

.swarm-hero h1 {
  font-size: clamp(1.35rem, 2.5vw, 1.75rem);
  font-weight: 800;
  letter-spacing: -0.03em;
  margin: 0 0 0.35rem;
  color: var(--text-main);
}

.hero-sub {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-dim);
  max-width: 500px;
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
  letter-spacing: -0.03em;
  color: var(--text-main);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.hero-stat-lbl {
  margin-top: 0.2rem;
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-mute);
}

.hero-stat.success .hero-stat-val { color: var(--success); }
.hero-stat.muted .hero-stat-val { color: var(--text-dim); }

.swarm-panel {
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border);
  background: var(--bg-card);
  padding: 1.15rem 1.15rem 0.35rem;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}

.panel-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex-wrap: wrap;
}

.filter-pills {
  display: flex;
  gap: 0.35rem;
  padding: 0.25rem;
  border-radius: var(--radius-md);
  background: var(--bg-input);
  border: 1px solid var(--border);
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
  cursor: pointer;
  background: transparent;
  border: none;
  transition: all 0.2s ease;
}

.filter-pill:hover {
  color: var(--text-main);
  background: var(--bg-subtle);
}

.filter-pill.active {
  background: var(--accent);
  color: #fff;
  box-shadow: 0 4px 12px rgba(var(--accent-rgb), 0.3);
}

.pill-count {
  font-size: 0.65rem;
  padding: 0.1rem 0.35rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.15);
  font-variant-numeric: tabular-nums;
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
  transition: border-color 0.2s, box-shadow 0.2s;
  position: relative;
}

.search-box:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(var(--accent-rgb), 0.1);
}

.search-box input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 0.85rem;
  font-weight: 600;
  width: 100%;
}

.clear-search-btn {
  background: transparent;
  border: none;
  color: var(--text-mute);
  cursor: pointer;
  font-size: 1.1rem;
  padding: 0;
  line-height: 1;
}

.refresh-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.6rem 0.85rem;
  border-radius: var(--radius-md);
  background: var(--accent);
  color: #fff;
  border: none;
  cursor: pointer;
  transition: background 0.2s, transform 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--accent-hover);
  transform: translateY(-1px);
}

.refresh-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.spinning {
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Table cells */
.name-cell {
  min-width: 160px;
}

.image-cell {
  max-width: 240px;
}

.image-pill-box {
  display: inline-flex;
  flex-direction: column;
  max-width: 220px;
  overflow: hidden;
}

.image-name {
  font-weight: 700;
  font-size: 0.85rem;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.replica-status {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-family: var(--font-mono);
  font-size: 0.85rem;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-up { background: var(--success); }
.status-partial { background: var(--warning, #f59e0b); }
.status-down { background: var(--error); }

.ports-badge {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm, 4px);
  color: var(--text-main);
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.6rem;
  border-radius: 99px;
  font-size: 0.68rem;
  font-weight: 850;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.badge-primary { background: rgba(var(--accent-rgb), 0.15); color: var(--accent); }
.badge-info { background: rgba(99, 102, 241, 0.15); color: #818cf8; }
.badge-manager { background: rgba(var(--success-rgb), 0.15); color: var(--success); }
.badge-worker { background: rgba(148, 163, 184, 0.15); color: var(--text-mute); }

.leader-badge {
  font-size: 0.65rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  background: #f59e0b;
  color: #000;
  font-weight: 800;
  margin-left: 0.35rem;
  display: inline-block;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  padding: 0.2rem 0.55rem;
  border-radius: 99px;
  font-size: 0.72rem;
  font-weight: 800;
}

.status-ready { background: rgba(var(--success-rgb), 0.15); color: var(--success); }
.status-down { background: rgba(var(--error-rgb), 0.15); color: var(--error); }

.avail-pill {
  display: inline-flex;
  align-items: center;
  padding: 0.2rem 0.55rem;
  border-radius: 99px;
  font-size: 0.72rem;
  font-weight: 800;
  text-transform: capitalize;
}

.avail-active { background: rgba(var(--success-rgb), 0.15); color: var(--success); }
.avail-pause { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }
.avail-drain { background: rgba(var(--error-rgb), 0.15); color: var(--error); }

.avail-select {
  padding: 0.3rem 0.6rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-main);
  border-radius: var(--radius-sm, 4px);
  font-size: 0.8rem;
  font-weight: 600;
}

.state-pill {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-input);
  border: 1px solid var(--border);
}

.state-running { background: rgba(var(--success-rgb), 0.15); border-color: rgba(var(--success-rgb), 0.3); color: var(--success); }
.state-failed { background: rgba(var(--error-rgb), 0.15); border-color: rgba(var(--error-rgb), 0.3); color: var(--error); }
.state-shutdown { color: var(--text-mute); }

.task-err {
  display: block;
  font-size: 0.72rem;
  color: var(--error);
  margin-top: 0.2rem;
}

.mono-text {
  font-family: var(--font-mono);
  font-size: 0.8rem;
}

.sub-text {
  font-size: 0.72rem;
  color: var(--text-mute);
  font-weight: 600;
}

/* Action group styling matching Kubernetes & ContainerTable */
.action-cell {
  white-space: nowrap;
}

.action-group {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.action-cluster {
  display: flex;
  gap: 0.35rem;
  padding: 0.25rem;
  border-radius: var(--radius-md);
  background: var(--bg-subtle);
  border: 1px solid var(--border-subtle, var(--border));
}

.icon-btn {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 1px solid transparent;
  background: var(--bg-input);
  color: var(--text-dim);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.icon-btn:hover {
  transform: translateY(-1px);
  color: var(--text-main);
  border-color: var(--border);
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.08);
}

.icon-btn.delete-btn:hover {
  color: var(--error);
  background: rgba(var(--error-rgb), 0.12);
  border-color: rgba(var(--error-rgb), 0.3);
}

.icon-btn.scale-btn:hover {
  color: var(--accent);
  border-color: rgba(var(--accent-rgb), 0.3);
}

.icon-btn.logs-btn:hover {
  color: var(--accent);
  border-color: rgba(var(--accent-rgb), 0.3);
}

.text-right { text-align: right; }
.justify-end { justify-content: flex-end; }
.text-muted { color: var(--text-mute); }

.empty-state {
  padding: 3rem 1.5rem;
  text-align: center;
  color: var(--text-mute);
}

/* Scale & Logs Modern Modals matching DockLog design system */
.scale-modal-content {
  max-width: 460px;
  border-radius: 20px;
  padding: 1.75rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  box-shadow: 0 25px 60px -15px var(--shadow);
}

.scale-modal-header {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  position: relative;
  margin-bottom: 0.75rem;
}

.scale-header-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  background: rgba(var(--accent-rgb), 0.12);
  color: var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.scale-header-title {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.scale-header-title h3 {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 800;
  color: var(--text-main);
  letter-spacing: -0.01em;
}

.scale-service-chip {
  font-family: var(--font-mono, monospace);
  font-size: 0.75rem;
  color: var(--accent);
  background: var(--bg-input);
  padding: 0.15rem 0.5rem;
  border-radius: 6px;
  border: 1px solid var(--border);
  display: inline-block;
  max-width: 260px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.scale-close-btn {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-mute);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.scale-close-btn:hover {
  background: var(--bg-subtle);
  color: var(--text-main);
}

.scale-modal-sub {
  color: var(--text-mute);
  font-size: 0.82rem;
  line-height: 1.5;
  margin: 0 0 1.5rem;
}

.scale-stepper-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.75rem;
}

.scale-input-label {
  font-size: 0.72rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-mute);
}

.scale-counter-box {
  display: flex;
  align-items: center;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 0.3rem;
  gap: 0.35rem;
}

.stepper-btn {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  color: var(--text-main);
  font-size: 1.25rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}

.stepper-btn:hover:not(:disabled) {
  border-color: var(--accent);
  color: var(--accent);
  background: rgba(var(--accent-rgb), 0.08);
}

.stepper-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.scale-input-number {
  flex: 1;
  background: transparent;
  border: none;
  text-align: center;
  font-family: var(--font-mono, monospace);
  font-size: 1.4rem;
  font-weight: 800;
  color: var(--text-main);
  outline: none;
  -moz-appearance: textfield;
}

.scale-input-number::-webkit-outer-spin-button,
.scale-input-number::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.scale-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.scale-btn-cancel {
  padding: 0.65rem 1.15rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-dim);
  border-radius: 10px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 700;
  transition: all 0.15s ease;
}

.scale-btn-cancel:hover {
  background: var(--bg-subtle);
  color: var(--text-main);
}

.scale-btn-submit {
  padding: 0.65rem 1.35rem;
  background: var(--accent);
  border: 1px solid var(--accent);
  color: #fff;
  border-radius: 10px;
  font-size: 0.85rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: 0 4px 14px rgba(var(--accent-rgb), 0.35);
}

.scale-btn-submit:hover:not(:disabled) {
  background: var(--accent-hover);
  transform: translateY(-1px);
}

.scale-btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Logs Drawer Modal */
.logs-modal-box {
  background: #090d16;
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  width: 100%;
  max-width: 900px;
  height: 80vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
}

.logs-modal-header {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #0f172a;
}

.logs-modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #f8fafc;
}

.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 1.5rem;
  cursor: pointer;
  line-height: 1;
}

.close-btn:hover {
  color: #fff;
}

.logs-stream-container {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  font-family: var(--font-mono, monospace);
  font-size: 0.82rem;
  background: #090d16;
}

.service-link {
  text-decoration: none;
  color: inherit;
  display: flex;
  flex-direction: column;
  cursor: pointer;
}

.service-link:hover .container-title {
  color: var(--accent);
}

.logs-loading {
  color: #64748b;
  text-align: center;
  padding: 2rem;
}

.log-row {
  display: flex;
  gap: 0.75rem;
  line-height: 1.5;
  padding: 0.15rem 0;
}

.log-time {
  color: #64748b;
  flex-shrink: 0;
}

.log-text {
  color: #e2e8f0;
  word-break: break-all;
  white-space: pre-wrap;
}

.log-row.stderr .log-text {
  color: #f87171;
}

@media (max-width: 900px) {
  .swarm-hero {
    flex-direction: column;
    align-items: stretch;
  }

  .hero-stats {
    width: 100%;
  }

  .hero-stat {
    flex: 1;
  }

  .panel-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-left,
  .toolbar-right {
    width: 100%;
  }

  .filter-pills {
    width: 100%;
    justify-content: space-between;
  }

  .filter-pill {
    flex: 1;
    justify-content: center;
  }

  .search-box {
    width: 100%;
    min-width: 0;
  }

  .refresh-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>

<template>
  <div
    v-if="visible && container"
    class="container-hover-card"
    :class="{ 'metrics-only': metricsOnly }"
    :style="cardPositionStyle"
    @mouseenter="$emit('card-mouseenter')"
    @mouseleave="$emit('card-mouseleave')"
  >
    <!-- Header: Status dot + Name & Image -->
    <div class="hover-card-header">
      <div class="header-status-line">
        <span
          class="status-indicator-dot"
          :class="container.state === 'running' ? 'is-running' : 'is-stopped'"
        ></span>
        <span class="container-title-name">{{ container.name }}</span>
      </div>
      <div class="container-image-sub" v-if="container.image">{{ container.image }}</div>
    </div>

    <!-- Metrics: CPU & MEM -->
    <div class="hover-card-metrics">
      <!-- CPU Row -->
      <div class="metric-row">
        <span class="metric-label">CPU</span>
        <div class="metric-bar-track">
          <!-- Discrete dot/bar cells -->
          <div
            v-for="i in 28"
            :key="'cpu-' + i"
            class="spark-bar"
            :class="{ active: i <= activeCpuBars }"
            :style="{ height: getBarHeight(i, 'cpu') }"
          ></div>
        </div>
        <span class="metric-value-text">{{ (currentCpu || 0).toFixed(2) }}%</span>
      </div>

      <!-- MEM Row -->
      <div class="metric-row">
        <span class="metric-label">MEM</span>
        <div class="metric-bar-track">
          <div
            v-for="i in 28"
            :key="'mem-' + i"
            class="spark-bar"
            :class="{ active: i <= activeMemBars }"
            :style="{ height: getBarHeight(i, 'mem') }"
          ></div>
        </div>
        <span class="metric-value-text">{{ formatMemoryValue(currentMem) }}</span>
      </div>
    </div>

    <!-- Metadata Attributes (Hidden when metricsOnly is true) -->
    <div v-if="!metricsOnly" class="hover-card-details">
      <div class="detail-row">
        <span class="detail-key">STATUS</span>
        <span class="detail-val">{{ container.state || 'running' }}</span>
      </div>
      <div class="detail-row" v-if="portsDisplay">
        <span class="detail-key">PORTS</span>
        <span class="detail-val">{{ portsDisplay }}</span>
      </div>
      <div class="detail-row">
        <span class="detail-key">CREATED</span>
        <span class="detail-val">{{ createdAgo }}</span>
      </div>
      <div class="detail-row">
        <span class="detail-key">UPTIME</span>
        <span class="detail-val">{{ uptimeDisplay }}</span>
      </div>
    </div>

    <!-- Bottom Actions Toolbar (Hidden when metricsOnly is true) -->
    <div v-if="!metricsOnly" class="hover-card-actions">
      <!-- Logs View Button (Logs stream) -->
      <button
        type="button"
        class="card-action-btn"
        @click.stop="goToLogs"
        title="View live logs"
        aria-label="View live logs"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
          <line x1="8" y1="9" x2="16" y2="9"></line>
          <line x1="8" y1="13" x2="13" y2="13"></line>
        </svg>
      </button>

      <!-- Terminal / Shell Button -->
      <button
        type="button"
        class="card-action-btn"
        @click.stop="goToShell"
        title="Attach terminal shell"
        aria-label="Attach terminal shell"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="4 17 10 11 4 5"></polyline>
          <line x1="12" y1="19" x2="20" y2="19"></line>
        </svg>
      </button>

      <!-- Inspect / Detail Info Button -->
      <button
        type="button"
        class="card-action-btn"
        @click.stop="goToInspect"
        title="Inspect container details"
        aria-label="Inspect container details"
      >
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="16" x2="12" y2="12"></line>
          <line x1="12" y1="8" x2="12.01" y2="8"></line>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { formatBytes } from '../utils/sharedState';
import { logsRouteForContainer } from '../utils/logRoutes';

const props = defineProps({
  visible: Boolean,
  container: Object,
  stats: Object,
  position: {
    type: Object,
    default: () => ({ top: 0, left: 0 }),
  },
  metricsOnly: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['card-mouseenter', 'card-mouseleave']);
const router = useRouter();

const currentCpu = computed(() => {
  if (props.stats && props.stats.cpu !== undefined && props.stats.cpu !== null) {
    return props.stats.cpu;
  }
  return props.container?.cpu ?? 0;
});

const currentMem = computed(() => {
  if (props.stats && props.stats.memory > 0) {
    return props.stats.memory;
  }
  return props.container?.memory ?? (props.stats?.memory || 0);
});

const activeCpuBars = computed(() => {
  const percent = Math.min(100, Math.max(0, currentCpu.value));
  return Math.ceil((percent / 100) * 28);
});

const activeMemBars = computed(() => {
  const limit = props.container?.mem_limit > 0 ? props.container.mem_limit : (4 * 1024 * 1024 * 1024);
  const percent = Math.min(100, Math.max(0, (currentMem.value / limit) * 100));
  return Math.max(4, Math.ceil((percent / 100) * 28));
});

function getBarHeight(index, type) {
  if (type === 'cpu') {
    // Generate organic sparkline heights for active bars
    if (index > activeCpuBars.value) return '3px';
    const pattern = [6, 12, 8, 14, 10, 16, 9, 13, 7, 15, 11, 14, 18, 10];
    return `${pattern[index % pattern.length]}px`;
  }
  // Memory uniform level block
  if (index > activeMemBars.value) return '3px';
  return '14px';
}

function formatMemoryValue(bytes) {
  if (!bytes) return '0 MB';
  const mb = bytes / (1024 * 1024);
  if (mb >= 1024) {
    return `${(mb / 1024).toFixed(2)} GB`;
  }
  return `${mb.toFixed(2)} MB`;
}

const portsDisplay = computed(() => {
  if (!props.container?.ports || props.container.ports.length === 0) {
    return null;
  }
  return props.container.ports.join(', ');
});

const createdAgo = computed(() => {
  const unix = props.container?.created;
  if (!unix) return 'Just now';
  const diffSec = Math.floor(Date.now() / 1000 - unix);
  if (diffSec < 60) return `${diffSec} seconds ago`;
  const mins = Math.floor(diffSec / 60);
  if (mins < 60) return `${mins} minute${mins === 1 ? '' : 's'} ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs} hour${hrs === 1 ? '' : 's'} ago`;
  const days = Math.floor(hrs / 24);
  return `${days} day${days === 1 ? '' : 's'} ago`;
});

const uptimeDisplay = computed(() => {
  if (props.container?.status) {
    return props.container.status;
  }
  return props.container?.state === 'running' ? 'Up 3 minutes' : 'Stopped';
});

const cardPositionStyle = computed(() => {
  const top = props.position?.top ?? 100;
  const left = props.position?.left ?? 260;
  return {
    top: `${top}px`,
    left: `${left}px`,
  };
});

function goToLogs() {
  if (props.container?.id) {
    router.push(logsRouteForContainer(props.container));
  }
}

function goToShell() {
  if (props.container?.id) {
    router.push({ path: '/shell', query: { c: props.container.id } });
  }
}

function goToInspect() {
  if (props.container?.id) {
    if (props.container.is_swarm && props.container.swarm_service) {
      router.push({ path: `/swarm/services/${props.container.swarm_service}` });
    } else {
      router.push({ path: `/containers/${props.container.id}` });
    }
  }
}
</script>

<style scoped>
.container-hover-card {
  position: fixed;
  z-index: 9999;
  width: 320px;
  background: #11151c;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  padding: 1.15rem;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.65), 0 0 0 1px rgba(255, 255, 255, 0.05);
  font-family: var(--font-main, -apple-system, BlinkMacSystemFont, sans-serif);
  color: #f1f5f9;
  pointer-events: auto;
  transition: opacity 0.15s ease, transform 0.15s ease;
  animation: fadeInPop 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

[data-theme="light"] .container-hover-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.16), 0 2px 8px rgba(0, 0, 0, 0.08);
  color: #0f172a;
}

@keyframes fadeInPop {
  from {
    opacity: 0;
    transform: scale(0.96) translateX(-4px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateX(0);
  }
}

.hover-card-header {
  margin-bottom: 1.1rem;
}

.header-status-line {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.status-indicator-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-indicator-dot.is-running {
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.6);
}

.status-indicator-dot.is-stopped {
  background: #64748b;
}

.container-title-name {
  font-size: 1.05rem;
  font-weight: 800;
  color: var(--text-main, #ffffff);
  letter-spacing: -0.015em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

[data-theme="light"] .container-title-name {
  color: #0f172a;
}

.container-image-sub {
  font-family: var(--font-mono, monospace);
  font-size: 0.76rem;
  color: #64748b;
  margin-top: 0.2rem;
  margin-left: 1rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

[data-theme="light"] .container-image-sub {
  color: #64748b;
}

/* Metrics rows */
.hover-card-metrics {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  margin-bottom: 1.25rem;
}

.container-hover-card.metrics-only {
  width: 290px;
  padding: 1rem 1.15rem;
}

.container-hover-card.metrics-only .hover-card-header {
  margin-bottom: 0.85rem;
}

.container-hover-card.metrics-only .hover-card-metrics {
  margin-bottom: 0;
}

.metric-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-family: var(--font-mono, monospace);
  font-size: 0.74rem;
}

.metric-label {
  color: #64748b;
  font-weight: 700;
  width: 32px;
  flex-shrink: 0;
}

.metric-bar-track {
  flex: 1;
  display: flex;
  align-items: flex-end;
  gap: 2px;
  height: 16px;
  padding-bottom: 1px;
}

.spark-bar {
  flex: 1;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 1px;
  transition: all 0.15s ease;
}

[data-theme="light"] .spark-bar {
  background: #e2e8f0;
}

.spark-bar.active {
  background: #10b981;
}

.metric-value-text {
  font-weight: 700;
  min-width: 60px;
  text-align: right;
  color: #f1f5f9;
  font-size: 0.76rem;
}

[data-theme="light"] .metric-value-text {
  color: #0f172a;
}

/* Details list */
.hover-card-details {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  padding: 0.85rem 0;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  font-size: 0.78rem;
  margin-bottom: 0.85rem;
}

[data-theme="light"] .hover-card-details {
  border-top-color: #f1f5f9;
  border-bottom-color: #f1f5f9;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.detail-key {
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  color: #64748b;
  text-transform: uppercase;
}

.detail-val {
  font-weight: 600;
  color: #f1f5f9;
  font-family: var(--font-mono, monospace);
  font-size: 0.78rem;
  text-align: right;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 190px;
}

[data-theme="light"] .detail-val {
  color: #1e293b;
}

/* Actions Toolbar */
.hover-card-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding-top: 0.25rem;
}

.card-action-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 0.35rem;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.card-action-btn:hover {
  color: #38bdf8;
  background: rgba(56, 189, 248, 0.12);
}
</style>


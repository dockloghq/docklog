<template>
  <div class="container-table" :class="{ embedded }">
    <div class="premium-table-container" :class="{ embedded }">
      <table class="premium-table containers-table">
        <thead>
          <tr>
            <th>Container</th>
            <th>Image</th>
            <th>Created</th>
            <th>Status</th>
            <th>State</th>
            <th class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody v-if="loading">
          <tr>
            <td colspan="6">
              <div class="table-loading">
                <div class="shimmer"></div>
              </div>
            </td>
          </tr>
        </tbody>
        <tbody v-else-if="displayContainers.length > 0">
          <tr
            v-for="c in displayContainers"
            :key="c.id"
            class="container-row"
            :class="{ 'is-running': c.state === 'running', 'is-hovered': showInlineStats && activeLiveId === c.id }"
            @mouseenter="showInlineStats && c.state === 'running' ? startLiveStats(c.id) : null"
            @mouseleave="showInlineStats ? stopLiveStats() : null"
          >
            <td data-label="Container">
              <div
                class="name-cell clickable"
                @click="goToLogs(c.id)"
              >
                <div class="container-avatar" :class="c.state === 'running' ? 'running' : 'stopped'">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
                  </svg>
                </div>
                <div class="name-main">
                  <span class="container-title">{{ c.name }}</span>
                  <span class="container-id">{{ c.id.substring(0, 12) }}</span>
                  <div
                    v-if="showInlineStats && c.state === 'running' && activeLiveId === c.id"
                    class="inline-stats"
                  >
                    <span class="stat-chip live">
                      CPU {{ liveStats.cpu?.toFixed(1) ?? c.cpu?.toFixed(1) ?? "0.0" }}%
                    </span>
                    <span class="stat-chip live">
                      {{ formatBytes(liveStats.memory ?? c.memory) }}
                    </span>
                  </div>
                </div>
              </div>
            </td>
            <td data-label="Image">
              <div class="image-cell">
                <span class="image-name">{{ c.image.split(":")[0] }}</span>
                <span class="image-tag">{{ c.image.split(":")[1] || "latest" }}</span>
              </div>
            </td>
            <td data-label="Created">
              <span class="date-label">{{ formatDate(c.created) }}</span>
            </td>
            <td data-label="Status">
              <span
                :class="[
                  'uptime-label',
                  c.state === 'running' ? 'is-running' : 'is-stopped',
                ]"
              >
                {{ c.status }}
              </span>
            </td>
            <td data-label="State">
              <div
                :class="[
                  'status-pill',
                  c.state === 'running' ? 'is-running' : 'is-stopped',
                ]"
              >
                <span class="pulse-dot"></span>
                {{ c.state.toUpperCase() }}
              </div>
            </td>
            <td class="text-right" data-label="Actions">
              <div class="action-group justify-end" @click.stop>
                <div class="action-cluster primary-actions">
                  <button
                    v-if="userCanStart(sharedState.currentUser) && c.state !== 'running'"
                    @click="triggerConfirm(c.id, 'start')"
                    class="icon-btn start"
                    data-tooltip="Start"
                  >
                  <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <polygon points="5 3 19 12 5 21 5 3"></polygon>
                  </svg>
                </button>
                <button
                  v-if="userCanStop(sharedState.currentUser) && c.state === 'running'"
                  @click="triggerConfirm(c.id, 'stop')"
                  class="icon-btn stop"
                  data-tooltip="Stop"
                >
                  <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    fill="currentColor"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <rect x="6" y="6" width="12" height="12"></rect>
                  </svg>
                </button>
                <button
                  v-if="userCanRestart(sharedState.currentUser)"
                  @click="triggerConfirm(c.id, 'restart')"
                  class="icon-btn restart"
                  data-tooltip="Restart"
                >
                  <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <path d="M23 4v6h-6"></path>
                    <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
                  </svg>
                </button>
                </div>
                <div class="action-cluster secondary-actions">
                <button
                  @click="goToLogs(c.id)"
                  class="icon-btn logs"
                  data-tooltip="View logs"
                >
                  <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <path
                      d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
                    ></path>
                    <path d="M8 9h8M8 13h5"></path>
                  </svg>
                </button>
                <button
                  v-if="userCanDelete(sharedState.currentUser)"
                  @click="triggerConfirm(c.id, 'remove')"
                  class="icon-btn delete"
                  data-tooltip="Delete"
                >
                  <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <polyline points="3 6 5 6 21 6"></polyline>
                    <path
                      d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                    ></path>
                  </svg>
                </button>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
        <tbody v-else>
          <tr>
            <td colspan="6">
              <div class="empty-state-wrapper">
                <div class="empty-state-content">
                  <div class="empty-icon-box">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                      <rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect>
                      <line x1="8" y1="21" x2="16" y2="21"></line>
                      <line x1="12" y1="17" x2="12" y2="21"></line>
                    </svg>
                  </div>
                  <h4 class="empty-title">No Containers Found</h4>
                  <p class="empty-text">
                    No containers match your search or you may not have access to any yet.
                  </p>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!loading && displayContainers.length" class="table-footer">
        Showing {{ displayContainers.length }} container{{ displayContainers.length === 1 ? "" : "s" }}
      </div>
    </div>

    <!-- Unified Action Confirmation Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showConfirm" class="modal-overlay">
          <div class="modal-content shadow-2xl">
            <div :class="['modal-icon', actionClass]">
              <svg v-if="pendingAction === 'start'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
              <svg v-else-if="pendingAction === 'stop'" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="2.5"><rect x="6" y="6" width="12" height="12"></rect></svg>
              <svg v-else-if="pendingAction === 'restart'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M23 4v6h-6"></path><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path></svg>
              <svg v-else-if="pendingAction === 'remove'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <polyline points="3 6 5 6 21 6"></polyline>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
              </svg>
            </div>
            <div class="modal-text-center">
              <h3>Confirm Operation</h3>
              <p>Are you sure you want to <strong>{{ pendingAction }}</strong> this container? This may affect active services.</p>
            </div>
            <div class="modal-divider"></div>
            <div class="modal-actions">
              <button @click="showConfirm = false" class="modal-btn cancel">Cancel</button>
              <button @click="executeAction" :class="['modal-btn confirm', actionClass]">Confirm {{ pendingAction }}</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useContainers } from '../composables/useContainers';
import { sharedState, userCanStart, userCanStop, userCanRestart, userCanDelete } from '../utils/sharedState';

const props = defineProps({
  stateFilter: {
    type: String,
    default: 'all',
  },
  showInlineStats: {
    type: Boolean,
    default: false,
  },
  embedded: {
    type: Boolean,
    default: false,
  },
});

const {
  loading, filteredContainers, activeLiveId, liveStats,
  showConfirm, pendingAction, actionClass,
  startLiveStats, stopLiveStats, goToLogs, triggerConfirm, executeAction,
  formatBytes, formatDate,
} = useContainers();

const displayContainers = computed(() => {
  if (props.stateFilter === 'running') {
    return filteredContainers.value.filter((c) => c.state === 'running');
  }
  if (props.stateFilter === 'stopped') {
    return filteredContainers.value.filter((c) => c.state !== 'running');
  }
  return filteredContainers.value;
});
</script>

<style scoped>
.premium-table-container.embedded {
  background: transparent;
  border: none;
  border-radius: 0;
  box-shadow: none;
}

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

.icon-btn.restart:hover {
  color: var(--accent);
  border-color: rgba(var(--accent-rgb), 0.35);
  background: var(--accent-soft);
}

.icon-btn.logs:hover {
  color: var(--accent);
  border-color: rgba(var(--accent-rgb), 0.35);
  background: var(--accent-soft);
}

.icon-btn.delete:hover {
  color: var(--error);
  border-color: rgba(var(--error-rgb), 0.35);
  background: rgba(var(--error-rgb), 0.08);
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

.inline-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.35rem;
  animation: stats-in 0.15s ease-out;
}

@keyframes stats-in {
  from {
    opacity: 0;
    transform: translateY(-2px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.stat-chip {
  font-family: var(--font-mono);
  font-size: 0.62rem;
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  border-radius: 6px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-dim);
  transition: color 0.2s, border-color 0.2s;
}

.stat-chip.live {
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.35);
  background: rgba(var(--success-rgb), 0.08);
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

.name-main {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.image-cell {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.image-name {
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
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

.status-pill.is-stopped {
  color: var(--text-mute);
  background: var(--bg-subtle);
  border-color: var(--border-subtle);
}

.pulse-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.is-running .pulse-dot {
  background: var(--success);
  box-shadow: 0 0 6px rgba(var(--success-rgb), 0.6);
  animation: pulse-mini 2s infinite;
}

.is-stopped .pulse-dot {
  background: var(--text-mute);
}

.uptime-label.is-stopped {
  color: var(--text-dim) !important;
  border-color: var(--border) !important;
  background: var(--bg-subtle) !important;
}

@keyframes pulse-mini {
  0%, 100% { transform: scale(0.95); opacity: 1; }
  50% { transform: scale(1.08); opacity: 0.75; }
}

.table-footer {
  padding: 0.75rem 1.25rem;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--text-mute);
  border-top: 1px solid var(--border);
  background: var(--bg-subtle);
}

@media (max-width: 850px) {
  .action-cluster {
    background: transparent;
    border: none;
    padding: 0;
  }

  .action-group {
    flex-wrap: wrap;
    justify-content: flex-start;
  }

  .inline-stats {
    display: none;
  }

  .premium-table thead {
    display: none;
  }

  .premium-table tbody tr {
    display: block;
    padding: 1.15rem;
    margin-bottom: 1rem;
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: var(--radius-xl);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.04);
  }

  .premium-table tbody tr td {
    display: flex;
    flex-direction: column;
    padding: 0.5rem 0;
    border: none;
    text-align: left !important;
    gap: 0.3rem;
  }

  .premium-table tbody tr td::before {
    content: attr(data-label);
    display: block;
    font-size: 0.62rem;
    font-weight: 800;
    color: var(--text-mute);
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .premium-table tbody tr td:first-child::before {
    display: none;
  }
}

.table-loading {
  padding: 2rem 1rem;
}

.table-loading .shimmer {
  min-height: 180px;
  border-radius: var(--radius-lg);
}

.empty-state-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 260px;
}

.empty-state-content {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.empty-icon-box {
  width: 68px;
  height: 68px;
  background: var(--accent-soft);
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  margin-bottom: 1rem;
  border: 1px solid rgba(var(--accent-rgb), 0.15);
}

.empty-icon-box svg {
  width: 32px;
  height: 32px;
}

.empty-title {
  font-size: 1.05rem;
  font-weight: 800;
  color: var(--text-main);
  margin-bottom: 0.4rem;
}

.empty-text {
  font-size: 0.85rem;
  color: var(--text-mute);
  max-width: 300px;
  line-height: 1.55;
}
</style>

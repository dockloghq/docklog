<template>
  <div class="swarm-service-table" :class="{ embedded }">
    <div class="premium-table-container" :class="{ embedded }">
      <table class="premium-table">
        <thead>
          <tr>
            <th style="min-width: 180px;">Service</th>
            <th>Image</th>
            <th>Mode</th>
            <th>Replicas</th>
            <th>Ports</th>
            <th>Created</th>
            <th class="text-right" style="min-width: 140px;">Actions</th>
          </tr>
        </thead>
        <tbody v-if="loading">
          <tr>
            <td colspan="7">
              <div class="table-loading"><div class="shimmer"></div></div>
            </td>
          </tr>
        </tbody>
        <tbody v-else-if="visibleServices.length">
          <tr v-for="svc in visibleServices" :key="svc.id">
            <td class="name-cell">
              <router-link :to="`/swarm/services/${svc.id}`" class="service-identity" style="text-decoration: none; color: inherit;">
                <span class="service-dot" :class="svcDotClass(svc)" aria-hidden="true"></span>
                <div>
                  <div class="service-name" :title="svc.name">{{ svc.name }}</div>
                  <div class="service-id">{{ svc.id ? svc.id.substring(0, 12) : '' }}</div>
                </div>
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
                  <router-link
                    :to="{ path: '/logs', query: { rt: 'swarm', s: svc.name } }"
                    class="icon-btn logs-btn"
                    data-tooltip="Service logs"
                  >
                    <AppIcon name="logsBubble" :size="15" />
                  </router-link>
                  <router-link
                    :to="{ path: '/swarm', query: { tab: 'services', search: svc.name } }"
                    class="icon-btn view-btn"
                    data-tooltip="Manage in Swarm"
                  >
                    <AppIcon name="server" :size="15" />
                  </router-link>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
        <tbody v-else>
          <tr>
            <td colspan="7">
              <div class="table-empty">No Swarm services found</div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="embedded && services.length > maxRows" class="embedded-footer">
      <router-link :to="viewAllTo" class="view-all-link">
        View all {{ services.length }} services
        <AppIcon name="chevronRight" :size="14" />
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import AppIcon from './AppIcon.vue';
import { formatSwarmDate } from '../composables/useSwarm';

const props = defineProps({
  services: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  embedded: { type: Boolean, default: false },
  maxRows: { type: Number, default: 5 },
  viewAllTo: { type: [String, Object], default: '/swarm' },
});

const visibleServices = computed(() => {
  if (!props.embedded || !props.maxRows) return props.services;
  return props.services.slice(0, props.maxRows);
});

function svcDotClass(svc) {
  if (svc.replicas > 0 && svc.running === svc.replicas) return 'online';
  if (svc.running > 0) return 'partial';
  return 'offline';
}

function formatImageName(fullImage) {
  if (!fullImage) return 'N/A';
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
</script>

<style scoped>
.swarm-service-table {
  width: 100%;
}

.premium-table-container {
  width: 100%;
  overflow-x: auto;
}

.premium-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.82rem;
}

.premium-table th {
  text-align: left;
  padding: 0.75rem 1rem;
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-mute);
  border-bottom: 1px solid var(--border);
}

.premium-table td {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--border-subtle, var(--border));
  color: var(--text-main);
  vertical-align: middle;
}

.premium-table tbody tr:hover td {
  background: var(--bg-hover, rgba(255, 255, 255, 0.02));
}

.service-identity {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.service-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.service-dot.online {
  background: var(--success);
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.4);
}

.service-dot.partial {
  background: var(--warning);
  box-shadow: 0 0 8px rgba(245, 158, 11, 0.4);
}

.service-dot.offline {
  background: var(--danger, #ef4444);
}

.service-name {
  font-weight: 700;
  color: var(--text-main);
  font-size: 0.85rem;
}

.service-id {
  font-size: 0.72rem;
  color: var(--text-mute);
  font-family: var(--font-mono, monospace);
}

.image-pill-box {
  display: inline-flex;
  align-items: center;
  background: var(--bg-subtle, rgba(0, 0, 0, 0.03));
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 0.2rem 0.5rem;
  font-size: 0.75rem;
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-name {
  font-weight: 600;
  color: var(--text-main);
}

.image-tag {
  margin-left: 0.35rem;
  color: var(--accent);
  font-weight: 600;
}

.badge {
  display: inline-block;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  font-size: 0.68rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.badge-primary {
  background: rgba(var(--accent-rgb), 0.12);
  color: var(--accent);
  border: 1px solid rgba(var(--accent-rgb), 0.2);
}

.badge-info {
  background: rgba(59, 130, 246, 0.12);
  color: #3b82f6;
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.replica-status {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}

.status-indicator {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.status-up {
  background: var(--success);
}

.status-partial {
  background: var(--warning);
}

.status-down {
  background: var(--danger, #ef4444);
}

.replica-count {
  font-size: 0.82rem;
  font-weight: 700;
}

.ports-badge {
  display: inline-block;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  font-family: var(--font-mono, monospace);
  font-size: 0.72rem;
  color: var(--text-dim);
}

.text-muted {
  color: var(--text-mute);
}

.date-label {
  font-size: 0.75rem;
  color: var(--text-mute);
  white-space: nowrap;
}

.action-cluster {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0.2rem 0.3rem;
}

.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: transparent;
  border: none;
  color: var(--text-mute);
  cursor: pointer;
  transition: all 0.15s ease;
  text-decoration: none;
}

.icon-btn:hover {
  background: var(--bg-hover);
  color: var(--text-main);
}

.icon-btn.logs-btn:hover {
  color: var(--accent);
}

.table-loading {
  padding: 2rem;
  display: flex;
  justify-content: center;
}

.shimmer {
  width: 100%;
  height: 20px;
  background: linear-gradient(90deg, var(--bg-subtle) 25%, var(--bg-hover) 50%, var(--bg-subtle) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
}

@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

.table-empty {
  padding: 2.5rem;
  text-align: center;
  color: var(--text-mute);
  font-size: 0.85rem;
}

.embedded-footer {
  display: flex;
  justify-content: flex-end;
  padding: 0.85rem 1rem;
  border-top: 1px solid var(--border);
  background: var(--bg-subtle, rgba(0, 0, 0, 0.01));
}

.view-all-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--accent);
  text-decoration: none;
  transition: opacity 0.15s;
}

.view-all-link:hover {
  opacity: 0.85;
}

.text-right {
  text-align: right;
}

.justify-end {
  justify-content: flex-end;
}
</style>

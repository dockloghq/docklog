<template>
  <div class="containers-view animate-fade-in">
    <div class="view-header">
      <div class="header-info">
        <h1>Container Management</h1>
        <p class="text-mute">
          Monitor and control your containerized ecosystem
        </p>
      </div>
      <div class="header-actions">
        <div class="search-box glass">
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
            placeholder="Filter by name or image..."
          />
        </div>
        <button
          class="premium-btn primary refresh-trigger"
          @click="fetchContainers"
          :disabled="loading"
        >
          <svg
            viewBox="0 0 24 24"
            width="16"
            height="16"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            :class="{ rotating: loading }"
          >
            <polyline points="23 4 23 10 17 10"></polyline>
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
          </svg>
          Refresh
        </button>
      </div>
    </div>

    <!-- MAIN TABLE -->
    <div class="premium-table-container glass mt-8">
      <table class="premium-table">
        <thead>
          <tr>
            <th>Container Name</th>
            <th>Image & Tag</th>
            <th>Created At</th>
            <th>Uptime</th>
            <th>State</th>
            <th class="text-right">Control</th>
          </tr>
        </thead>
        <tbody v-if="filteredContainers.length > 0">
          <tr v-for="c in filteredContainers" :key="c.id" class="container-row">
            <td data-label="Container Name">
              <div 
                class="name-cell clickable group" 
                @click="goToLogs(c.id)"
                @mouseenter="startLiveStats(c.id)"
                @mouseleave="stopLiveStats"
              >
                <div class="name-main">
                  <span class="container-title">{{ c.name }}</span>
                  <span class="container-id">{{ c.id.substring(0, 12) }}</span>
                </div>
                
                <!-- Stats Peek Hover (Live when hovering) -->
                <div v-if="c.state === 'running'" class="row-stats-peek">
                   <div class="r-stat">
                     <span class="r-val" :class="{ 'text-live': activeLiveId === c.id }">
                       {{ (activeLiveId === c.id ? liveStats.cpu : c.cpu)?.toFixed(1) || 0 }}%
                     </span>
                   </div>
                   <div class="r-stat">
                     <span class="r-val" :class="{ 'text-live': activeLiveId === c.id }">
                       {{ formatBytes(activeLiveId === c.id ? liveStats.memory : c.memory) }}
                     </span>
                   </div>
                </div>
              </div>
            </td>
            <td data-label="Image & Tag">
              <div class="image-cell">
                <span class="image-name">{{ c.image.split(":")[0] }}</span>
                <span class="image-tag">{{
                  c.image.split(":")[1] || "latest"
                }}</span>
              </div>
            </td>
            <td data-label="Created At">
              <span class="date-label">{{ formatDate(c.created) }}</span>
            </td>
            <td data-label="Uptime">
              <span class="uptime-label">{{ c.status }}</span>
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
              <div class="action-group justify-end">
                <button
                  v-if="c.state !== 'running'"
                  @click="triggerConfirm(c.id, 'start')"
                  class="icon-btn start"
                  data-tooltip="Start Container"
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
                  v-if="c.state === 'running'"
                  @click="triggerConfirm(c.id, 'stop')"
                  class="icon-btn stop"
                  data-tooltip="Stop Container"
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
                  @click="triggerConfirm(c.id, 'restart')"
                  class="icon-btn restart"
                  data-tooltip="Restart Container"
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
                <button
                  @click="goToLogs(c.id)"
                  class="icon-btn logs"
                  data-tooltip="View Live Logs"
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
                  @click="triggerConfirm(c.id, 'remove')"
                  class="icon-btn stop"
                  data-tooltip="Delete Container"
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
            </td>
          </tr>
        </tbody>
        <!-- CENTERED EMPTY STATE -->
        <tbody v-else>
          <tr>
            <td colspan="5">
              <div class="empty-state-wrapper">
                <div class="empty-state-content">
                  <div class="empty-icon-box">
                    <svg
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                    >
                      <rect
                        x="2"
                        y="3"
                        width="20"
                        height="14"
                        rx="2"
                        ry="2"
                      ></rect>
                      <line x1="8" y1="21" x2="16" y2="21"></line>
                      <line x1="12" y1="17" x2="12" y2="21"></line>
                    </svg>
                  </div>
                  <h4 class="empty-title">No Containers Found</h4>
                  <p class="empty-text">
                    We couldn't find any containers matching your search
                    criteria or access permissions.
                  </p>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- ACTION CONFIRMATION MODAL -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showConfirm" class="modal-overlay">
          <div class="modal-content shadow-2xl">
            <div :class="['modal-icon', actionClass]">
              <svg
                v-if="pendingAction === 'start'"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <polygon points="5 3 19 12 5 21 5 3"></polygon>
              </svg>
              <svg
                v-else-if="pendingAction === 'stop'"
                viewBox="0 0 24 24"
                fill="currentColor"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <rect x="6" y="6" width="12" height="12"></rect>
              </svg>
              <svg
                v-else-if="pendingAction === 'restart'"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <path d="M23 4v6h-6"></path>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
              </svg>
              <svg
                v-else-if="pendingAction === 'remove'"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <polyline points="3 6 5 6 21 6"></polyline>
                <path
                  d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                ></path>
              </svg>
            </div>
            <div class="modal-text-center">
              <h3>Confirm Operation</h3>
              <p>
                Are you sure you want to
                <strong>{{ pendingAction }}</strong> this container? This may
                affect active services.
              </p>
            </div>
            <div class="modal-divider"></div>
            <div class="modal-actions">
              <button @click="showConfirm = false" class="modal-btn cancel">
                Cancel
              </button>
              <button
                @click="executeAction"
                :class="['modal-btn confirm', actionClass]"
              >
                Confirm {{ pendingAction }}
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
import { useRouter } from "vue-router";
import { secureStorage } from "../utils/storage";
import { sharedState, showToast } from "../utils/sharedState";

const formatDate = (unix) => {
  if (!unix) return "N/A";
  return new Date(unix * 1000).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return "0B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + sizes[i];
};

const router = useRouter();
const containers = ref([]);
const loading = ref(false);

// Live Stats on Hover Logic
const activeLiveId = ref(null);
const liveStats = ref({ cpu: 0, memory: 0 });
let liveInterval = null;

const startLiveStats = (id) => {
  activeLiveId.value = id;
  fetchStatsNow(id);
  if (liveInterval) clearInterval(liveInterval);
  liveInterval = setInterval(() => fetchStatsNow(id), 1000);
};

const stopLiveStats = () => {
  activeLiveId.value = null;
  if (liveInterval) clearInterval(liveInterval);
  liveInterval = null;
};

const fetchStatsNow = async (id) => {
  try {
    const token = secureStorage.getItem("token");
    const res = await fetch(`/api/containers/${id}/stats-now`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      const data = await res.json();
      liveStats.value = { cpu: data.cpu, memory: data.memory };
    }
  } catch (err) {
    console.error("Live stats fetch failed", err);
  }
};

const showConfirm = ref(false);
const pendingId = ref(null);
const pendingAction = ref("");
const actionClass = computed(() => {
  if (pendingAction.value === "start") return "success";
  if (pendingAction.value === "restart") return "warning";
  if (pendingAction.value === "stop" || pendingAction.value === "remove")
    return "error";
  return "";
});
let refreshInterval = null;

const filteredContainers = computed(() => {
  return containers.value.filter(
    (c) =>
      c.name.toLowerCase().includes(sharedState.searchQuery.toLowerCase()) ||
      c.image.toLowerCase().includes(sharedState.searchQuery.toLowerCase()),
  );
});

const fetchContainers = async () => {
  loading.value = true;
  try {
    const token = secureStorage.getItem("token");
    const res = await fetch("/api/containers", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) containers.value = await res.json();
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const triggerConfirm = (id, action) => {
  pendingId.value = id;
  pendingAction.value = action;
  showConfirm.value = true;
};

const executeAction = async () => {
  if (!pendingId.value) return;
  try {
    const token = secureStorage.getItem("token");
    const formData = new FormData();
    formData.append("action", pendingAction.value);
    const res = await fetch(`/api/containers/${pendingId.value}/action`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
    });
    if (res.ok) {
      showToast(
        "Success",
        `Operation ${pendingAction.value}ed successfully`,
        "success",
      );
      fetchContainers();
    }
    showConfirm.value = false;
  } catch (err) {
    console.error(err);
  }
};

const goToLogs = (id) => {
  router.push(`/logs?c=${id}`);
};

onMounted(() => {
  fetchContainers();
  refreshInterval = setInterval(fetchContainers, 3000);
});

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval);
});
</script>

<style scoped>
/* Header Styling */
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 2rem;
}

.view-header h1 {
  font-size: 1.75rem;
  font-weight: 950;
  letter-spacing: -0.05em;
  color: var(--text-main);
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 1.5rem;
  align-items: center;
}

/* Table Styling */
.premium-table-container {
  min-height: 500px;
  display: flex;
  flex-direction: column;
}

.container-row:hover {
  background: rgba(var(--accent-rgb), 0.02);
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  position: relative;
}
.name-cell.clickable {
  cursor: pointer;
}
.name-main {
  display: flex;
  flex-direction: column;
}

/* Row Stats Peek */
.row-stats-peek {
  display: flex;
  gap: 1rem;
  opacity: 0;
  transform: translateX(-10px);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: var(--bg-glass);
  padding: 0.4rem 0.8rem;
  border-radius: 10px;
  border: 1px solid var(--border-light);
  pointer-events: none;
}

.name-cell:hover .row-stats-peek {
  opacity: 1;
  transform: translateX(0);
}

.r-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.r-val {
  font-size: 0.75rem;
  font-weight: 900;
  color: var(--accent);
  font-family: var(--font-mono);
  transition: color 0.3s;
}

.text-live {
  color: var(--success) !important;
  text-shadow: 0 0 8px rgba(var(--success-rgb), 0.4);
}

.r-lab {
  font-size: 0.55rem;
  font-weight: 800;
  color: var(--text-mute);
  text-transform: uppercase;
}

.container-title {
  font-weight: 800;
  color: var(--text-main);
  font-size: 0.95rem;
}
.container-id {
  font-family: "JetBrains Mono", monospace;
  font-size: 0.7rem;
  color: var(--text-mute);
  font-weight: 700;
}

.image-cell {
  display: flex;
  flex-direction: column;
}
.image-name {
  font-weight: 800;
  color: var(--text-main);
  font-size: 0.85rem;
}
.image-tag {
  font-size: 0.7rem;
  color: var(--text-mute);
  font-weight: 700;
  background: var(--bg-input);
  padding: 1px 6px;
  border-radius: 4px;
  width: fit-content;
  margin-top: 2px;
}

.date-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-mute);
}

/* Status Pills */
.status-pill {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.7rem;
  font-weight: 900;
  padding: 0.4rem 0.8rem;
  border-radius: 10px;
  width: fit-content;
  border: 1px solid transparent;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.is-running {
  background: rgba(var(--success-rgb), 0.1);
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.2);
}
.is-running .pulse-dot {
  background: var(--success);
  box-shadow: 0 0 10px var(--success);
  animation: pulse 2s infinite;
}

.is-stopped {
  background: rgba(var(--stop-rgb), 0.1);
  color: var(--stop);
  border-color: rgba(var(--stop-rgb), 0.2);
}
.is-stopped .pulse-dot {
  background: var(--stop);
}

@keyframes pulse {
  0% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(var(--success-rgb), 0.7);
  }
  70% {
    transform: scale(1);
    box-shadow: 0 0 0 6px rgba(var(--success-rgb), 0);
  }
  100% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(var(--success-rgb), 0);
  }
}

/* Control Buttons */
.action-group {
  display: flex;
  gap: 0.5rem;
}

.icon-btn {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-mute);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.icon-btn:hover {
  background: var(--bg-subtle);
  transform: translateY(-2px);
  color: var(--text-main);
}
.icon-btn.start:hover {
  color: var(--success);
  border-color: var(--success);
  box-shadow: 0 4px 12px rgba(var(--success-rgb), 0.2);
}
.icon-btn.stop:hover {
  color: var(--stop);
  border-color: var(--stop);
  box-shadow: 0 4px 12px rgba(var(--stop-rgb), 0.2);
}
.icon-btn.restart:hover {
  color: var(--warning);
  border-color: var(--warning);
  box-shadow: 0 4px 12px rgba(var(--warning-rgb), 0.2);
}
.icon-btn.logs:hover {
  color: var(--accent);
  border-color: var(--accent);
  box-shadow: 0 4px 12px rgba(var(--accent-rgb), 0.2);
}

/* Modal & Empty State (Consistent with your other views) */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-card {
  width: 100%;
  max-width: 400px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 24px;
  overflow: hidden;
}
.modal-card-body {
  padding: 2rem;
}
.modal-card-footer {
  padding: 1.5rem 2rem;
  display: flex;
  gap: 1rem;
}

.action-icon-wrapper {
  width: 64px;
  height: 64px;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
}

.action-icon-wrapper.start {
  background: rgba(var(--success-rgb), 0.1);
  color: var(--success);
}
.action-icon-wrapper.stop {
  background: rgba(var(--stop-rgb), 0.1);
  color: var(--stop);
}
.action-icon-wrapper.restart {
  background: rgba(var(--warning-rgb), 0.1);
  color: var(--warning);
}

.modal-title {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--text-main);
}
.modal-subtitle {
  font-size: 0.9rem;
  color: var(--text-mute);
  margin-top: 0.5rem;
  line-height: 1.5;
}

.btn-primary.start {
  background: var(--success);
}
.btn-primary.stop {
  background: var(--stop);
}
.btn-primary.restart {
  background: var(--warning);
}

/* Empty State */
.empty-state-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 350px;
}
.empty-state-content {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.empty-icon-box {
  width: 80px;
  height: 80px;
  background: var(--bg-input);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-mute);
  opacity: 0.5;
  margin-bottom: 1.5rem;
  border: 1px dashed var(--border);
}
.empty-title {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--text-main);
  margin-bottom: 0.5rem;
}
.empty-text {
  font-size: 0.9rem;
  color: var(--text-mute);
  max-width: 300px;
  line-height: 1.6;
}

/* Utilities */
.rotating {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

button.premium-btn.primary.refresh-trigger {
  max-width: 120px;
}

.modal-bounce-enter-active {
  animation: bounce 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}
@keyframes bounce {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* Responsive Overrides */
@media (max-width: 1024px) {
  .view-header {
    flex-direction: column;
    align-items: stretch;
    gap: 1.25rem;
  }
  .header-info h1 {
    font-size: 1.75rem;
    line-height: 1.2;
  }
  .header-actions {
    width: 100%;
    align-items: stretch;
    gap: 1rem;
  }
  .search-box {
    min-width: 0 !important;
    width: 100%;
  }
  .premium-btn {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 850px) {
  .premium-table thead {
    display: none;
  }
  .premium-table,
  .premium-table tbody,
  .premium-table tr,
  .premium-table td {
    display: block;
    width: 100%;
  }
  .premium-table-container {
    background: transparent;
    border: none;
    box-shadow: none;
    min-height: auto;
    margin-top: 1rem;
  }
  .container-row {
    margin-bottom: 1.25rem;
    padding: 1.25rem;
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 20px;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
  }
  .container-row td {
    padding: 0.6rem 0;
    border: none;
    text-align: left !important;
  }
  .container-row td::before {
    content: attr(data-label);
    display: block;
    font-size: 0.65rem;
    font-weight: 800;
    color: var(--text-mute);
    text-transform: uppercase;
    margin-bottom: 0.35rem;
    letter-spacing: 0.05em;
  }
  .action-group {
    justify-content: flex-start !important;
    margin-top: 0.5rem;
    gap: 0.75rem;
  }
}
</style>

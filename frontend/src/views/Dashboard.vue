<template>
  <div class="dashboard-container">
    <!-- Provider Context Banner -->
    <div v-if="sharedState.activeProvider === 'kubernetes'" class="k8s-banner glass animate-slide-up">
      <div class="banner-content">
        <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.5">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
          <line x1="12" y1="9" x2="12" y2="13"></line>
          <line x1="12" y1="17" x2="12.01" y2="17"></line>
        </svg>
        <div class="banner-text">
          <h4>Kubernetes Integration Pending</h4>
          <p>K8s monitoring is currently in beta. Connection to your cluster is being established.</p>
        </div>
      </div>
    </div>

    <!-- Summary Stats -->
    <div class="summary-grid animate-slide-up">
      <div class="premium-stat-card">
        <div class="stat-header">
          <div class="stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
              <polyline points="2 17 12 22 22 17"></polyline>
              <polyline points="2 12 12 17 22 12"></polyline>
            </svg>
          </div>
          <span class="badge badge-dim">Total</span>
        </div>
        <div class="stat-content">
          <span class="stat-label">TOTAL CONTAINERS</span>
          <span class="stat-value">{{ containers.length }}</span>
        </div>
      </div>

      <div class="premium-stat-card">
        <div class="stat-header">
          <div class="stat-icon success">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
            </svg>
          </div>
          <span class="badge badge-success">Active</span>
        </div>
        <div class="stat-content">
          <span class="stat-label">RUNNING</span>
          <span class="stat-value">{{ runningCount }}</span>
        </div>
      </div>

      <div class="premium-stat-card">
        <div class="stat-header">
          <div class="stat-icon error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <path d="M18.36 6.64a9 9 0 1 1-12.73 0"></path>
              <line x1="12" y1="2" x2="12" y2="12"></line>
            </svg>
          </div>
          <span class="badge badge-error">Stopped</span>
        </div>
        <div class="stat-content">
          <span class="stat-label">STOPPED</span>
          <span class="stat-value">{{ stoppedCount }}</span>
        </div>
      </div>

    </div>

    <!-- Overview Grid -->
    <div class="dashboard-grid full-width">
      <!-- Container Overview Table -->
      <div class="grid-section">
        <div class="section-header">
          <h3>Container Overview</h3>
          <div class="header-actions">
            <div class="search-box glass">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
              </svg>
              <input type="text" v-model="sharedState.searchQuery" placeholder="Search..." />
            </div>
          </div>
        </div>
        
        <div class="premium-table-container">
          <table class="premium-table">
            <thead>
              <tr>
                <th>Container Name</th>
                <th>Image & Tag</th>
                <th>Created</th>
                <th>Uptime</th>
                <th>Status</th>
                <th class="text-right">Control</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="c in filteredContainers"
                :key="c.id"
                class="container-row"
              >
                <td data-label="Container Name">
                  <div class="name-cell">
                    <span class="container-title">{{ c.name }}</span>
                    <span class="container-id">{{ c.id.substring(0, 12) }}</span>
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
                <td data-label="Created">
                  <span class="date-label">{{ formatDate(c.created) }}</span>
                </td>
                <td data-label="Uptime">
                  <span class="uptime-label">{{ c.status }}</span>
                </td>
                <td data-label="Status">
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
          </table>
        </div>
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
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { secureStorage } from "../utils/storage";
import { sharedState, showToast } from "../utils/sharedState";

const formatDate = (unix) => {
  if (!unix) return "N/A";
  return new Date(unix * 1000).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};
const router = useRouter();
const route = useRoute();

const containers = ref([]);
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

const runningCount = computed(
  () => containers.value.filter((c) => c.state === "running").length,
);
const stoppedCount = computed(
  () => containers.value.filter((c) => c.state !== "running").length,
);
const imagesCount = computed(
  () => new Set(containers.value.map((c) => c.image)).size,
);

const filteredContainers = computed(() => {
  return containers.value.filter(
    (c) =>
      c.name.toLowerCase().includes(sharedState.searchQuery.toLowerCase()) ||
      c.image.toLowerCase().includes(sharedState.searchQuery.toLowerCase()),
  );
});

const getContainerById = (id) => containers.value.find((c) => c.id === id);

const goToLogs = (id) => {
  router.push({ path: "/logs", query: { c: id } });
};

const fetchContainers = async () => {
  try {
    const token = secureStorage.getItem("token");
    const res = await fetch("/api/containers", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      containers.value = await res.json();
    }
  } catch (err) {
    console.error(err);
  }
};

const containerAction = async (id, action) => {
  try {
    const token = secureStorage.getItem("token");
    const formData = new FormData();
    formData.append("action", action);
    const res = await fetch(`/api/containers/${id}/action`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
    });
    if (res.ok) {
      showToast("Success", `Action ${action} executed.`, "success");
      fetchContainers();
    } else {
      showToast("Error", "Action failed.", "error");
    }
  } catch (err) {
    console.error(err);
  }
};

const triggerConfirm = (id, action) => {
  pendingId.value = id;
  pendingAction.value = action;
  showConfirm.value = true;
};

const executeAction = () => {
  if (pendingId.value && pendingAction.value) {
    containerAction(pendingId.value, pendingAction.value);
    showConfirm.value = false;
  }
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
.dashboard-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-bottom: 1.5rem;
}

.k8s-banner {
  padding: 1.5rem 2rem;
  border-radius: 20px;
  background: rgba(245, 158, 11, 0.05);
  border: 1px solid rgba(245, 158, 11, 0.2);
  color: #f59e0b;
}

.banner-content {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.banner-text h4 {
  margin: 0;
  font-size: 1rem;
  font-weight: 900;
}

.banner-text p {
  margin: 0.2rem 0 0;
  font-size: 0.85rem;
  font-weight: 500;
  opacity: 0.8;
}

.dashboard-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  align-items: start;
  transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
}

.dashboard-grid.full-width {
  grid-template-columns: 1fr;
}

.grid-section {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  min-width: 0;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-header h3 {
  font-size: 1.1rem;
  font-weight: 950;
  color: var(--text-main);
  letter-spacing: -0.02em;
  margin: 0;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.5rem 0.85rem;
  border-radius: 10px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  width: 180px;
}

.search-box input {
  background: transparent;
  border: none;
  color: var(--text-main);
  font-size: 0.75rem;
  font-weight: 700;
  width: 100%;
  outline: none;
}

.search-box svg {
  color: var(--text-mute);
}

.premium-table tr {
  cursor: pointer;
  transition: all 0.2s;
}

.premium-table tr.active td {
  background: rgba(99, 102, 241, 0.05);
  color: var(--accent);
}

.action-group {
  display: flex;
  gap: 0.5rem;
}

.icon-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-dim);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: var(--bg-card);
  color: var(--text-main);
  border-color: var(--accent);
}

.icon-btn.start:hover { color: var(--success); border-color: var(--success); }
.icon-btn.stop:hover { color: var(--error); border-color: var(--error); }
.icon-btn.restart:hover { color: var(--warning); border-color: var(--warning); }
.icon-btn.logs:hover {
  color: var(--accent);
  border-color: var(--accent);
  box-shadow: 0 4px 12px rgba(var(--accent-rgb), 0.2);
}



.name-cell {
  display: flex;
  flex-direction: column;
  text-align: left;
}
.container-title {
  font-weight: 850;
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
  text-align: left;
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

.status-pill {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.7rem;
  font-weight: 950;
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
  box-shadow: 0 0 8px var(--success);
  animation: pulse-mini 2s infinite;
}

.is-stopped {
  background: rgba(var(--error-rgb), 0.1);
  color: var(--error);
  border-color: rgba(var(--error-rgb), 0.2);
}
.is-stopped .pulse-dot {
  background: var(--error);
}

@keyframes pulse-mini {
  0% { transform: scale(0.95); opacity: 1; }
  50% { transform: scale(1.1); opacity: 0.7; }
  100% { transform: scale(0.95); opacity: 1; }
}

@media (max-width: 850px) {
  .summary-grid {
    grid-template-columns: repeat(2, 1fr) !important;
    gap: 1rem !important;
  }
  .dashboard-grid {
    grid-template-columns: 1fr !important;
    gap: 1.5rem !important;
  }
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1.25rem;
  }
  .header-actions {
    width: 100%;
  }
  .search-box {
    width: 100% !important;
    min-width: 0 !important;
  }
  .premium-table thead {
    display: none;
  }
  .premium-table tbody tr {
    display: block;
    padding: 1.25rem;
    margin-bottom: 1.25rem;
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 20px;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
  }
  .premium-table tbody tr td {
    display: flex;
    flex-direction: column;
    padding: 0.6rem 0;
    border: none;
    text-align: left !important;
    gap: 0.35rem;
  }
  .premium-table tbody tr td::before {
    content: attr(data-label);
    display: block;
    font-size: 0.65rem;
    font-weight: 800;
    color: var(--text-mute);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
}

@media (max-width: 480px) {
  .summary-grid {
    grid-template-columns: 1fr !important;
  }
  .k8s-banner {
    padding: 1rem;
  }
  .banner-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }
  .stat-value {
    font-size: 1.5rem;
  }
  .stat-label {
    font-size: 0.65rem;
  }
  .premium-stat-card {
    padding: 0.85rem 1rem;
    gap: 0.75rem;
  }
  .section-header h3 {
    font-size: 1rem;
  }
}
</style>

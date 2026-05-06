<template>
  <div class="dashboard-page">
    <main class="main-content">
      <button
        v-if="!isCompact && !sharedState.dashboardSidebarOpen"
        class="sidebar-expand-btn"
        @click="sharedState.dashboardSidebarOpen = true"
        title="Expand Sidebar"
      >
        <svg
          viewBox="0 0 24 24"
          width="16"
          height="16"
          fill="none"
          stroke="currentColor"
          stroke-width="3"
        >
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </button>

      <aside
        :class="[
          'sidebar',
          'glass',
          { collapsed: !sharedState.dashboardSidebarOpen },
        ]"
      >
        <div class="sidebar-header">
          <div class="header-top">
            <h2>RESOURCES</h2>
            <div style="display: flex; gap: 0.4rem">
              <button
                v-if="!isCompact"
                @click="sharedState.dashboardSidebarOpen = false"
                class="mode-toggle"
                title="Collapse Sidebar"
              >
                <svg
                  viewBox="0 0 24 24"
                  width="14"
                  height="14"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2.5"
                >
                  <polyline points="15 18 9 12 15 6"></polyline>
                </svg>
              </button>
              <button
                @click="isSplitMode = !isSplitMode"
                :class="['mode-toggle', { active: isSplitMode }]"
                title="Split View"
              >
                <svg
                  viewBox="0 0 24 24"
                  width="14"
                  height="14"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                  <line x1="12" y1="3" x2="12" y2="21"></line>
                </svg>
              </button>
            </div>
          </div>
          
          <div class="segmented-control glass">
            <button 
              :class="['segment-btn', { active: resourceMode === 'docker' }]"
              @click="resourceMode = 'docker'"
            >
              Docker
            </button>
            <button 
              :class="['segment-btn', { active: resourceMode === 'kubernetes' }]"
              @click="resourceMode = 'kubernetes'"
            >
              Kubernetes
            </button>
          </div>
        </div>

        <div class="container-list">
          <div
            v-for="(c, idx) in filteredContainers"
            :key="c.id"
            :class="[
              'container-card',
              { selected: isSelected(c.id), focused: focusedIndex === idx },
            ]"
            @click="selectContainer(c)"
          >
            <div class="card-status" :class="c.state"></div>
            <div class="card-info">
              <div class="c-name-row">
                <span class="c-name">{{ c.name }}</span>
                <span v-if="c.namespace" class="c-namespace">@{{ c.namespace }}</span>
              </div>
              <span class="c-image">{{ c.image }}</span>
            </div>
            <div
              v-if="
                currentUser?.is_admin ||
                currentUser?.can_start ||
                currentUser?.can_stop ||
                currentUser?.can_restart ||
                currentUser?.can_delete
              "
              class="card-actions"
              @click.stop
            >
              <button
                v-if="
                  c.state !== 'running' &&
                  (currentUser?.is_admin || currentUser?.can_start)
                "
                @click="triggerConfirm(c.id, 'start')"
                class="action-btn start"
                title="Start Container"
              >
                <svg
                  viewBox="0 0 24 24"
                  width="14"
                  height="14"
                  stroke="currentColor"
                  stroke-width="3"
                  fill="none"
                >
                  <polygon points="5 3 19 12 5 21 5 3"></polygon>
                </svg>
              </button>
              <button
                v-if="
                  c.state === 'running' &&
                  (currentUser?.is_admin || currentUser?.can_stop)
                "
                @click="triggerConfirm(c.id, 'stop')"
                class="action-btn stop"
                title="Stop Container"
              >
                <svg
                  viewBox="0 0 24 24"
                  width="14"
                  height="14"
                  stroke="currentColor"
                  stroke-width="3"
                  fill="none"
                >
                  <rect x="6" y="6" width="12" height="12"></rect>
                </svg>
              </button>
              <button
                v-if="
                  c.state === 'running' &&
                  (currentUser?.is_admin || currentUser?.can_restart)
                "
                @click="triggerConfirm(c.id, 'restart')"
                class="action-btn restart"
                title="Restart Container"
              >
                <svg
                  viewBox="0 0 24 24"
                  width="14"
                  height="14"
                  stroke="currentColor"
                  stroke-width="3"
                  fill="none"
                >
                  <path d="M23 4v6h-6"></path>
                  <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
                </svg>
              </button>
              <button
                v-if="currentUser?.is_admin || currentUser?.can_delete"
                @click="triggerConfirm(c.id, 'remove')"
                class="action-btn remove"
                title="Remove Container"
              >
                <svg
                  viewBox="0 0 24 24"
                  width="14"
                  height="14"
                  stroke="currentColor"
                  stroke-width="3"
                  fill="none"
                >
                  <polyline points="3 6 5 6 21 6"></polyline>
                  <path
                    d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                  ></path>
                </svg>
              </button>
            </div>
          </div>
        </div>
      </aside>

      <div
        :class="[
          'sidebar-backdrop',
          { active: isCompact && sharedState.dashboardSidebarOpen },
        ]"
        @click="closeSidebar"
      ></div>

      <section class="viewer-area">
        <div v-if="selectedContainers.length === 0" class="standby-view">
          <div class="standby-content">
            <div class="standby-icon">
              <svg
                viewBox="0 0 24 24"
                width="48"
                height="48"
                stroke="currentColor"
                stroke-width="1.5"
                fill="none"
              >
                <rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect>
                <rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect>
                <line x1="6" y1="6" x2="6.01" y2="6"></line>
                <line x1="6" y1="18" x2="6.01" y2="18"></line>
              </svg>
            </div>
            <h2>System Standby</h2>
            <p>
              Select a container from the sidebar to begin real-time monitoring.
            </p>

            <!-- Standby Stats Card -->
            <div v-if="systemStats" class="standby-stats-card glass">
              <div class="s-card-item">
                <span class="s-card-label">SYSTEM CPU</span>
                <div class="s-card-main">
                  <span class="s-card-value highlight">{{ Number(systemStats.cpu || 0).toFixed(1) }}%</span>
                  <div class="s-card-bar"><div class="s-bar-fill" :style="{ width: (systemStats.cpu || 0) + '%', background: getStatColor(systemStats.cpu) }"></div></div>
                </div>
              </div>
              <div class="s-card-item">
                <span class="s-card-label">SYSTEM MEMORY</span>
                <div class="s-card-main">
                  <span class="s-card-value">
                    <span class="highlight">{{ (systemStats.usedMemGB || 0).toFixed(1) }}</span>
                    <span class="unit"> / {{ (systemStats.totalMemGB || 0).toFixed(1) }} GB</span>
                  </span>
                  <div class="s-card-bar">
                    <div
                      class="s-bar-fill accent"
                      :style="{
                        width: ((systemStats.totalMemGB || 0) > 0
                          ? ((systemStats.usedMemGB || 0) / systemStats.totalMemGB) * 100
                          : 0) + '%'
                      }"
                    ></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-else :class="['viewer-grid', { 'split-view': isSplitMode }]">
          <LogViewer
            v-for="c in activeViewers"
            :key="c.id"
            :container="c"
            @close="removeContainer(c)"
            :show-close="isSplitMode"
            :compact-header="isSplitMode"
          />
        </div>
      </section>
    </main>

    <!-- Unified Action Confirmation Modal -->
    <Transition name="fade">
      <div v-if="showConfirm" class="modal-overlay glass">
        <div class="modal-content glass shadow-2xl">
          <div :class="['modal-icon', pendingAction]">
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
              fill="none"
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
              v-else
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <path d="M3 6h18"></path>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"></path>
              <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
            </svg>
          </div>
          <h3>
            Confirm
            {{ pendingAction.charAt(0).toUpperCase() + pendingAction.slice(1) }}
          </h3>
          <p>
            Are you sure you want to
            <strong>{{ pendingAction }}</strong> container
            <strong>{{ pendingId }}</strong
            >?
            <span v-if="pendingAction === 'remove'"
              >This action is permanent and cannot be undone.</span
            >
            <span v-else>This will affect the container's availability.</span>
          </p>
          <div class="modal-actions">
            <button @click="showConfirm = false" class="modal-btn cancel">
              Cancel
            </button>
            <button
              @click="executeAction"
              :class="['modal-btn', `confirm-${pendingAction}`]"
            >
              {{
                pendingAction.charAt(0).toUpperCase() + pendingAction.slice(1)
              }}
              Container
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import LogViewer from "../components/LogViewer.vue";
import { secureStorage } from "../utils/storage";
import { sharedState, showToast } from "../utils/sharedState";

const router = useRouter();
const route = useRoute();

const containers = ref([]);
const selectedContainers = ref([]);
const isSplitMode = ref(false);
const isCompact = ref(window.innerWidth <= 1024);
const resourceMode = ref('docker'); // 'docker' or 'kubernetes'

const handleResize = () => {
  const compact = window.innerWidth <= 1024;
  isCompact.value = compact;
  sharedState.dashboardSidebarOpen = !compact;
};

const closeSidebar = () => {
  sharedState.dashboardSidebarOpen = false;
};

onMounted(() => {
  window.addEventListener("resize", handleResize);
  handleResize();
});

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval);
});

const currentUser = computed(() => sharedState.currentUser);
const systemStats = computed(() => sharedState.systemStats);

const showConfirm = ref(false);
const pendingId = ref(null);
const pendingAction = ref("");
const focusedIndex = ref(-1);

let refreshInterval = null;

const filteredContainers = computed(() => {
  return containers.value.filter(
    (c) =>
      c.name.toLowerCase().includes(sharedState.searchQuery.toLowerCase()) ||
      c.image.toLowerCase().includes(sharedState.searchQuery.toLowerCase()),
  );
});

const activeViewers = computed(() => {
  if (!isSplitMode.value) return selectedContainers.value.slice(-1);
  return selectedContainers.value.slice(-2);
});

const isSelected = (id) => selectedContainers.value.some((c) => c.id === id);

const navigateDown = () => {
  if (focusedIndex.value < filteredContainers.value.length - 1)
    focusedIndex.value++;
};

const navigateUp = () => {
  if (focusedIndex.value > 0) focusedIndex.value--;
};

const selectFocused = () => {
  if (
    focusedIndex.value >= 0 &&
    focusedIndex.value < filteredContainers.value.length
  ) {
    selectContainer(filteredContainers.value[focusedIndex.value]);
  }
};

const selectContainer = (c) => {
  const max = isSplitMode.value ? 2 : 1;
  if (isSelected(c.id)) {
    selectedContainers.value = selectedContainers.value.filter(
      (sc) => sc.id !== c.id,
    );
  } else {
    selectedContainers.value.push(c);
    if (selectedContainers.value.length > max) {
      selectedContainers.value.shift();
    }
  }
  updateUrl();
};

watch(isSplitMode, (newVal) => {
  const max = newVal ? 2 : 1;
  if (selectedContainers.value.length > max) {
    selectedContainers.value = selectedContainers.value.slice(-max);
  }
  updateUrl();
});

watch(() => sharedState.searchQuery, () => {
  focusedIndex.value = -1;
});

watch(resourceMode, () => {
  selectedContainers.value = [];
  containers.value = [];
  fetchContainers();
  updateUrl();
});

const getStatColor = (val) => {
  if (val > 80) return "var(--error)";
  if (val > 50) return "var(--warning)";
  return "var(--accent)";
};

const formatNumber = (val, decimals = 1) => {
  const num = parseFloat(val);
  return isNaN(num) ? (0).toFixed(decimals) : num.toFixed(decimals);
};

// Global stats are now handled by sharedState in MainLayout

const formatBytes = (bytes) => {
  if (bytes === 0) return "0B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + sizes[i];
};

const removeContainer = (c) => {
  selectedContainers.value = selectedContainers.value.filter(
    (sc) => sc.id !== c.id,
  );
  updateUrl();
};

const updateUrl = () => {
  const ids = selectedContainers.value.map((c) => c.id).join(",");
  const query = { ...route.query, c: ids };
  if (isSplitMode.value) query.split = "true";
  else delete query.split;
  router.replace({ query });
};

const fetchContainers = async () => {
  try {
    const token = secureStorage.getItem("token");
    const endpoint = resourceMode.value === 'docker' ? '/api/containers' : '/api/pods';
    const res = await fetch(endpoint, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      containers.value = await res.json();
      if (route.query.split === "true") isSplitMode.value = true;
      const urlIds = route.query.c?.split(",") || [];
      selectedContainers.value = containers.value.filter((c) =>
        urlIds.includes(c.id),
      );
    } else if (res.status === 401) {
      logout();
    }
  } catch (err) {
    console.error(err);
  }
};

// Toast notifications are handled by sharedState globally

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
    if (!res.ok) {
      const data = await res.json();
      showToast(
        "Access Denied",
        data.error || "You do not have permission to perform this action.",
        "error",
      );
    } else {
      showToast(
        "Action Successful",
        `Container successfully sent ${action} command.`,
        "success",
      );
    }
    fetchContainers();
  } catch (err) {
    console.error(err);
    showToast(
      "System Error",
      "An unexpected error occurred while communicating with the engine.",
      "error",
    );
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
    pendingId.value = null;
    pendingAction.value = "";
  }
};

const logout = () => {
  secureStorage.removeItem("token");
  secureStorage.removeItem("user");
  router.push("/login");
};

// Password management is now handled globally by MainLayout

onMounted(async () => {
  fetchContainers();
  refreshInterval = setInterval(fetchContainers, 5000);
});

onUnmounted(() => {
});

watch(
  () => route.query.c,
  () => {
    if (containers.value.length > 0) {
      const urlIds = route.query.c?.split(",") || [];
      selectedContainers.value = containers.value.filter((c) =>
        urlIds.includes(c.id),
      );
    }
  },
);
</script>

<style scoped>
.dashboard-page {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 0;
}

.dashboard-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
  background: var(--bg-main);
  min-height: 0;
}

.sidebar {
  width: 340px;
  min-width: 340px;
  height: calc(100vh - 72px);
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border);
  background: var(--bg-sidebar);
  position: relative;
  z-index: 10;
  overflow: hidden;
  transition:
    width 0.3s cubic-bezier(0.23, 1, 0.32, 1),
    min-width 0.3s cubic-bezier(0.23, 1, 0.32, 1);
}
.sidebar.collapsed {
  width: 0;
  min-width: 0;
  border-right: none;
}

.sidebar-header {
  padding: 1.5rem 1.5rem 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.sidebar-header h2 {
  font-size: 0.7rem;
  font-weight: 900;
  color: var(--text-mute);
  letter-spacing: 0.2em;
  margin: 0;
}

.mode-toggle {
  width: 32px;
  height: 32px;
  border-radius: 9px;
  border: 1px solid var(--border);
  background: rgba(255, 255, 255, 0.03);
  color: var(--text-mute);
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}
.mode-toggle:hover {
  border-color: var(--text-mute);
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.06);
}
.mode-toggle.active {
  background: var(--accent);
  color: #fff;
  border-color: var(--accent);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.container-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem 1.5rem 2rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.container-card {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  padding: 1.35rem 1rem;
  border-radius: 20px;
  cursor: pointer;
  border: 1px solid var(--border);
  background: var(--bg-card);
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  box-shadow: 0 4px 12px var(--shadow);
  position: relative;
  overflow: hidden;
}
.container-card:hover {
  background: var(--card-hover);
  border-color: var(--border);
  transform: translateY(-2px);
  box-shadow: 0 12px 24px -10px var(--shadow);
}
.container-card:hover .c-name {
  color: var(--accent);
}
.container-card.selected {
  background: var(--bg-main);
  border-color: var(--accent);
  box-shadow: 0 10px 30px -10px rgba(99, 102, 241, 0.3);
}
.container-card.selected::before {
  content: "";
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 5px;
  background: var(--accent);
  box-shadow: 2px 0 10px rgba(99, 102, 241, 0.5);
}

.card-status {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 2px solid var(--bg-sidebar);
  box-shadow: 0 0 8px rgba(0, 0, 0, 0.1);
}
.card-status.running {
  background: var(--success);
  box-shadow: 0 0 10px var(--success);
}
.card-status.exited {
  background: var(--text-mute);
}

.card-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding-right: 0.75rem;
  transition: all 0.3s;
}
.c-name {
  font-size: 0.88rem;
  font-weight: 900;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: -0.03em;
  transition: all 0.2s;
}
.c-image {
  font-size: 0.65rem;
  color: var(--text-dim);
  font-family: var(--font-mono);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-top: 0.15rem;
  opacity: 0.4;
  font-weight: 500;
  transition: all 0.2s;
}

.container-card:hover .c-name,
.container-card:hover .c-image {
  opacity: 1;
}

.card-actions {
  display: flex;
  gap: 0.5rem;
  opacity: 0;
  width: 0;
  overflow: hidden;
  transform: translateX(10px);
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  padding: 5px 0px;
}
.container-card:hover .card-actions {
  opacity: 1;
  width: auto;
  transform: translateX(0);
}
.action-btn {
  width: 24px;
  height: 24px;
  border-radius: 7px;
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-sidebar);
  color: var(--text-mute);
  cursor: pointer;
  transition: all 0.2s;
}
.action-btn svg {
  width: 12px;
  height: 12px;
}
.action-btn:hover {
  background: var(--bg-card);
  color: var(--text-main);
  border-color: var(--text-dim);
  transform: translateY(-2px);
}
.action-btn.start:hover {
  color: var(--success);
  border-color: var(--success);
  background: rgba(16, 185, 129, 0.1);
}
.action-btn.stop:hover,
.action-btn.remove:hover {
  color: var(--error);
  border-color: var(--error);
  background: rgba(239, 68, 68, 0.1);
}
.action-btn.restart:hover {
  color: var(--accent);
  border-color: var(--accent);
  background: rgba(99, 102, 241, 0.1);
}

.viewer-area {
  flex: 1;
  background: var(--bg-main);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  position: relative;
  min-height: 0;
}

.sidebar-open-btn {
  display: none;
}

.viewer-grid {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}
.viewer-grid.split-view {
  flex-direction: row;
}
.viewer-grid.split-view > * {
  flex: 1;
  min-width: 0;
  border-right: 1px solid var(--border);
}
.viewer-grid.split-view > *:last-child {
  border-right: none;
}

.standby-view {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  z-index: 1;
  text-align: center;
}
.standby-content {
  max-width: 760px;
  animation: fadeIn 0.8s ease-out;
}
.standby-icon {
  color: var(--accent);
  margin-bottom: 1.75rem;
  opacity: 0.18;
  transform: scale(1.15);
}
.standby-view h2 {
  font-size: 1.65rem;
  font-weight: 900;
  color: var(--text-main);
  margin-bottom: 0.6rem;
  letter-spacing: -0.04em;
}
.standby-view p {
  color: var(--text-dim);
  font-size: 0.9rem;
  font-weight: 500;
  line-height: 1.6;
}

.standby-stats-card {
  margin: 3rem auto 0;
  display: grid;
  grid-template-columns: 1fr 1fr;
  padding: 2.25rem;
  border-radius: 28px;
  border: 1px solid var(--border);
  gap: 2rem;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  box-shadow: 0 30px 80px -25px var(--shadow);
  width: 100%;
  max-width: 760px;
}

.s-card-item {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  text-align: left;
}

.s-card-label {
  font-size: 0.62rem;
  font-weight: 900;
  color: var(--text-mute);
  letter-spacing: 0.18em;
}

.s-card-main {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.s-card-value {
  font-size: 1rem;
  font-weight: 800;
  color: var(--text-dim);
  font-family: 'JetBrains Mono', monospace;
}

.s-card-value.highlight {
  color: var(--text-main);
  font-size: 2rem;
}

.highlight {
  color: var(--text-main);
  font-size: 2rem;
}

.unit {
  font-size: 0.82rem;
  color: var(--text-mute);
}

.s-card-bar {
  width: 100%;
  height: 6px;
  background: var(--bg-input);
  border-radius: 10px;
  overflow: hidden;
}

.s-bar-fill {
  height: 100%;
  transition: width 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.s-bar-fill.accent {
  background: var(--accent);
}

@media (max-width: 768px) {
  .standby-stats-card {
    flex-direction: column;
    padding: 1.5rem;
    gap: 1.25rem;
    width: 100%;
    margin-top: 1.5rem;
  }
  .s-card-divider {
    width: 100% !important;
    height: 1px !important;
  }
  .standby-view h2 {
    font-size: 1.45rem;
  }
}

.s-card-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.s-card-label {
  font-size: 0.62rem;
  font-weight: 900;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.15em;
}
.s-card-value {
  font-size: 1rem;
  font-weight: 800;
  color: var(--text-dim);
  font-family: var(--font-mono);
  letter-spacing: -0.05em;
}
.s-card-divider {
  width: 1px;
  height: 60px;
  background: var(--border);
}

/* Modal Perfection */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-main);
  background-image:
    radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.1) 0px, transparent 50%),
    radial-gradient(at 100% 100%, rgba(16, 185, 129, 0.1) 0px, transparent 50%);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 5000;
  overflow-y: auto;
  padding: 2rem 1rem;
}

[data-theme="dark"] .modal-overlay {
  background: rgba(0, 0, 0, 0.85);
}

.modal-content {
  width: 520px;
  padding: 3rem;
  border-radius: 32px;
  border: 1px solid var(--border);
  background: var(--bg-sidebar);
  text-align: center;
  box-shadow: 0 40px 100px -20px var(--shadow);
  position: relative;
  overflow: hidden;
}
.modal-content::after {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(
    to right,
    transparent,
    var(--accent),
    transparent
  );
}

.modal-icon {
  width: 72px;
  height: 72px;
  border-radius: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.5rem;
}
.modal-icon svg {
  width: 36px;
  height: 36px;
}

.modal-content h3 {
  font-size: 1.5rem;
  font-weight: 950;
  margin-bottom: 0.75rem;
  color: var(--text-main);
  letter-spacing: -0.03em;
}
.modal-content p {
  color: var(--text-dim);
  font-size: 0.95rem;
  line-height: 1.5;
  margin-bottom: 2rem;
  font-weight: 500;
}

.modal-footer {
  display: flex;
  gap: 1.5rem;
}
.modal-btn {
  flex: 1;
  padding: 0.9rem;
  border-radius: 14px;
  font-weight: 800;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  border: none;
}
.modal-btn.cancel {
  background: var(--bg-input);
  color: var(--text-main);
  border: 1px solid var(--border);
}
.modal-btn.cancel:hover {
  background: var(--bg-card);
  transform: translateY(-2px);
}

.confirm-start {
  background: var(--success);
  color: #fff;
  box-shadow: 0 15px 30px -5px rgba(16, 185, 129, 0.4);
}
.confirm-stop,
.confirm-remove {
  background: var(--error);
  color: #fff;
  box-shadow: 0 15px 30px -5px rgba(239, 68, 68, 0.4);
}
.confirm-restart {
  background: var(--accent);
  color: #fff;
  box-shadow: 0 15px 30px -5px rgba(99, 102, 241, 0.4);
}

.modal-btn:hover:not(.cancel) {
  transform: translateY(-4px);
  filter: brightness(1.1);
}
.modal-btn.confirm {
  background: var(--accent);
  color: #fff;
  box-shadow: 0 15px 30px -5px rgba(99, 102, 241, 0.4);
}

.modal-actions {
  display: flex;
  gap: 1rem;
  margin-top: 0.5rem;
}

.glass-input {
  width: 100%;
  padding: 1rem 1.25rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 14px;
  color: var(--text-main);
  font-size: 1rem;
  font-weight: 600;
  transition: all 0.2s;
  outline: none;
}
.glass-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.1);
  background: var(--bg-main);
}
.glass-input::placeholder {
  color: var(--text-mute);
  font-weight: 500;
}

.input-group {
  margin-bottom: 1rem;
}
.input-error {
  color: var(--error);
  font-size: 0.8rem;
  font-weight: 700;
  margin-top: 0.5rem;
}
.force-text-new {
  color: var(--text-dim);
  font-size: 0.95rem;
  font-weight: 500;
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(15px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@media (max-width: 1200px) {
  .sidebar {
    width: 280px;
  }
  .main-header {
    padding: 0 1.5rem;
  }
}

@media (max-width: 1024px) {
  .app-container {
    height: 100%;
    min-height: 0;
    overflow: hidden;
  }

  .main-content {
    height: 100%;
    overflow: hidden;
  }

  .sidebar {
    position: fixed;
    top: 72px;
    left: 0;
    bottom: 0;
    width: min(320px, 88vw) !important;
    z-index: 2000;
    transform: translateX(-100%);
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .sidebar.collapsed {
    transform: translateX(-100%);
    width: min(320px, 88vw) !important;
  }

  .sidebar:not(.collapsed) {
    transform: translateX(0);
    box-shadow: 0 30px 80px -24px rgba(15, 23, 42, 0.35);
  }

  .card-actions {
    opacity: 1 !important;
    width: auto !important;
    transform: none !important;
    gap: 0.75rem;
  }

  .action-btn {
    width: 32px;
    height: 32px;
  }

  .action-btn svg {
    width: 16px;
    height: 16px;
  }

  .viewer-area {
    height: 100%;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .sidebar-open-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0.75rem 0.75rem 0;
    align-self: flex-start;
    padding: 0.7rem 1rem;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: var(--bg-sidebar);
    color: var(--text-main);
    font-size: 0.8rem;
    font-weight: 850;
    box-shadow: 0 10px 30px var(--shadow);
    z-index: 5;
  }

  .viewer-grid {
    flex: 1;
    height: auto !important;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .viewer-grid.split-view > * {
    height: calc((100vh - 72px - 2rem) / 2) !important;
    min-height: 0;
  }

  .viewer-grid:not(.split-view) > * {
    height: 100% !important;
    min-height: 0;
    flex: 1;
  }

  .header-right {
    gap: 0.5rem;
  }

  .header-stats-group {
    display: none;
  }

  .search-wrapper {
    max-width: 180px;
  }

  .user-menu {
    right: -1rem;
    width: calc(100vw - 2rem);
    max-width: 280px;
  }
}

@media (max-width: 600px) {
  .search-wrapper {
    display: none;
  }
  .header-left {
    gap: 1rem;
  }

  .sidebar {
    top: 72px;
  }
}

@media (max-width: 480px) {
  .header-right .nav-icon-btn:not(:last-child) {
    display: none;
  }
}
</style>

<template>
  <div
    class="logs-container-layout"
    :class="{ 'sidebar-collapsed': isSidebarHidden }"
  >
    <!-- RESOURCES SIDEBAR -->
    <aside class="resources-sidebar glass">
      <div class="sidebar-top-nav">
        <!-- Mobile Header Row -->
        <div class="mobile-header-row show-mobile">
          <router-link to="/dashboard" class="back-nav-link-mobile">
            <svg
              viewBox="0 0 24 24"
              width="16"
              height="16"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
            >
              <polyline points="15 18 9 12 15 6"></polyline>
            </svg>
            <span>Dashboard</span>
          </router-link>

          <button
            class="minimal-close-btn-glass"
            @click="isSidebarHidden = true"
          >
            <svg
              viewBox="0 0 24 24"
              width="20"
              height="20"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      </div>

      <div class="sidebar-header">
        <span class="label-caps">Log Resources</span>
        <div class="sidebar-controls">
          <!-- Theme Toggle -->
          <button
            class="mini-icon-btn"
            @click="toggleTheme"
            title="Toggle Theme"
          >
            <svg
              v-if="sharedState.theme === 'dark'"
              viewBox="0 0 24 24"
              width="12"
              height="12"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
            >
              <circle cx="12" cy="12" r="5"></circle>
              <line x1="12" y1="1" x2="12" y2="3"></line>
              <line x1="12" y1="21" x2="12" y2="23"></line>
              <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
              <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
              <line x1="1" y1="12" x2="3" y2="12"></line>
              <line x1="21" y1="12" x2="23" y2="12"></line>
              <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
              <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
            </svg>
            <svg
              v-else
              viewBox="0 0 24 24"
              width="12"
              height="12"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
            >
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
            </svg>
          </button>

          <!-- Collapse Toggle -->
          <button
            class="mini-icon-btn"
            @click="isSidebarHidden = true"
            title="Collapse Sidebar"
          >
            <svg
              viewBox="0 0 24 24"
              width="12"
              height="12"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
            >
              <polyline points="15 18 9 12 15 6"></polyline>
            </svg>
          </button>

          <!-- Split View Toggle (Hide on mobile) -->
          <button
            class="mini-icon-btn hide-mobile"
            :class="{ 'active-toggle': splitView }"
            @click="toggleSplitView"
            title="Toggle Split View"
          >
            <svg
              viewBox="0 0 24 24"
              width="12"
              height="12"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
            >
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
              <line x1="12" y1="3" x2="12" y2="21"></line>
            </svg>
          </button>
        </div>
      </div>

      <div class="resource-list">
        <div
          v-for="c in filteredContainers"
          :key="c.id"
          class="resource-card group"
          :class="{ active: isVisible(c.id) }"
          @click="toggleStream(c.id)"
          @mouseenter="startLiveStats(c.id)"
          @mouseleave="stopLiveStats"
        >
          <!-- Status dot indicator -->
          <div class="card-status-dot" :class="c.state"></div>
          <div class="card-info">
            <span class="card-name">{{ c.name }}</span>
            <span class="card-image-tag">{{ c.image }}</span>
          </div>

          <!-- Stats Peek (Only for running) -->
          <div v-if="c.state === 'running'" class="stats-peek-inline">
            <div class="peek-stat">
              <span class="p-value" :class="{ 'text-live': activeLiveId === c.id }">
                {{ (activeLiveId === c.id ? liveStats.cpu : c.cpu)?.toFixed(1) || 0 }}%
              </span>
            </div>
            <div class="peek-stat">
              <span class="p-value" :class="{ 'text-live': activeLiveId === c.id }">
                {{ formatBytes(activeLiveId === c.id ? liveStats.memory : c.memory) }}
              </span>
            </div>
          </div>
        </div>
        <div v-if="filteredContainers.length === 0" class="empty-search-msg">
          <p class="text-mute">No containers found</p>
        </div>
      </div>
    </aside>

    <!-- Mobile Overlay -->
    <div
      v-if="!isSidebarHidden"
      class="mobile-sidebar-overlay show-mobile"
      @click="isSidebarHidden = true"
    ></div>

    <!-- FIXED TRIGGER (Always accessible to reopen sidebar) -->
    <button
      v-if="isSidebarHidden"
      class="fixed-open-trigger glass shadow-xl"
      @click="isSidebarHidden = false"
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

    <!-- MAIN VIEWPORT -->
    <main class="logs-main-content">
      <div
        v-if="displayContainers.length > 0"
        class="logs-grid"
        :class="gridClass"
      >
        <LogViewer
          v-for="c in displayContainers"
          :key="c.id"
          :container="c"
          showClose
          @close="removeStream(c.id)"
        />
      </div>

      <!-- PREMIUM HERO EMPTY STATE -->
      <div v-else class="empty-state-wrapper animate-fade-in">
        <div class="observability-hero">
          <div class="hero-visual">
            <div class="radar-scan">
              <div class="circle c1"></div>
              <div class="circle c2"></div>
              <div class="circle c3"></div>
              <svg
                viewBox="0 0 24 24"
                class="hero-icon"
                fill="none"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
                ></path>
                <path d="M8 9h8M8 13h5" stroke-linecap="round"></path>
              </svg>
            </div>
          </div>
          <div class="hero-text">
            <h2 class="display-title">Ready for Insight?</h2>
            <p class="subtitle">
              Select a resource from the sidebar to launch a real-time log
              stream. Toggle split-view to monitor up to two containers
              simultaneously.
            </p>
            <div class="hero-actions" v-if="isSidebarHidden">
              <button
                @click="isSidebarHidden = false"
                class="premium-btn primary"
              >
                <svg
                  viewBox="0 0 24 24"
                  width="18"
                  height="18"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="3"
                >
                  <path d="M9 18l6-6-6-6"></path>
                </svg>
                Open Resources
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { sharedState, showToast, fetchCurrentUser } from "../utils/sharedState";
import { secureStorage } from "../utils/storage";
import LogViewer from "../components/LogViewer.vue";

const route = useRoute();
const router = useRouter();

const toggleTheme = () => {
  sharedState.theme = sharedState.theme === "dark" ? "light" : "dark";
  secureStorage.setItem("theme", sharedState.theme);
  document.documentElement.setAttribute("data-theme", sharedState.theme);
};

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return "0B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + sizes[i];
};

const containers = ref([]);

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

const filteredContainers = computed(() => {
  if (!sharedState.searchQuery) return containers.value;
  const q = sharedState.searchQuery.toLowerCase();
  return containers.value.filter(
    (c) => c.name.toLowerCase().includes(q) || c.id.toLowerCase().includes(q),
  );
});
const selectedIds = ref([]);
const isSidebarHidden = ref(window.innerWidth < 1024);
const splitView = ref(route.query.split === "true");

const syncStateFromUrl = () => {
  const urlParam = route.query.c;
  if (!urlParam) {
    selectedIds.value = [];
    return;
  }
  const urlIds = urlParam.split(",").filter(Boolean);
  selectedIds.value = urlIds;
  splitView.value = route.query.split === "true";
  console.log("[Logs] Synced IDs from URL:", selectedIds.value);
};

// Ensure we match containers even if short IDs are provided in the URL
const displayContainers = computed(() => {
  if (containers.value.length === 0) return [];
  
  const ordered = selectedIds.value
    .map((id) => {
      // Try exact match first
      let match = containers.value.find((c) => c.id === id);
      // Fallback: match by prefix (handle short IDs)
      if (!match) {
        match = containers.value.find((c) => c.id.startsWith(id) || id.startsWith(c.id));
      }
      return match;
    })
    .filter(Boolean);

  return splitView.value ? ordered.slice(-2) : [ordered[ordered.length - 1]].filter(Boolean);
});

watch(() => containers.value, () => {
  if (selectedIds.value.length > 0) {
    console.log("[Logs] Containers loaded, applying selection from URL");
  }
}, { immediate: true });

const isVisible = (id) => displayContainers.value.some((c) => c.id === id);
const gridClass = computed(() =>
  displayContainers.value.length > 1 ? "grid-dual" : "grid-single",
);

const fetchContainers = async () => {
  try {
    const token = secureStorage.getItem("token");
    const res = await fetch("/api/containers", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      containers.value = await res.json();
      syncStateFromUrl();
    }
  } catch (err) {
    console.error(err);
  }
};

const updateUrl = () => {
  const query = { ...route.query };

  if (selectedIds.value.length > 0) {
    query.c = selectedIds.value.join(",");
  } else {
    delete query.c;
  }

  if (splitView.value) {
    query.split = "true";
  } else {
    delete query.split;
  }

  router.replace({ query });
};

const toggleSplitView = () => {
  splitView.value = !splitView.value;
  updateUrl();
};

const toggleStream = (id) => {
  if (splitView.value) {
    // SPLIT VIEW: FIFO Logic (Max 2)
    if (selectedIds.value.includes(id)) {
      selectedIds.value = selectedIds.value.filter((sid) => sid !== id);
    } else {
      if (selectedIds.value.length >= 2) {
        selectedIds.value.shift();
      }
      selectedIds.value.push(id);
    }
  } else {
    // SINGLE VIEW: Replace selection entirely
    selectedIds.value = selectedIds.value.includes(id) ? [] : [id];
  }
  if (window.innerWidth < 900) {
    isSidebarHidden.value = true;
  }
  updateUrl();
};

const removeStream = (id) => {
  selectedIds.value = selectedIds.value.filter((sid) => sid !== id);
  updateUrl();
};

let statusInterval = null;

onMounted(() => {
  fetchContainers();
  // Real-time status heartbeat
  statusInterval = setInterval(fetchContainers, 3000);
});

onUnmounted(() => {
  if (statusInterval) clearInterval(statusInterval);
});

watch(() => route.query, syncStateFromUrl);
</script>

<style scoped>
.logs-container-layout {
  display: flex;
  height: calc(100vh - 80px);
  position: relative;
  overflow: hidden;
  background: var(--bg-main);
}

/* SIDEBAR UI FIXES */
.resources-sidebar {
  width: 300px;
  height: 100%;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  padding: 1.5rem 1.25rem;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
}

.sidebar-top-nav {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  width: 280px;
}

.logo-img-sidebar {
  width: 28px;
  height: 28px;
  object-fit: contain;
  border: 1px solid #2b4a64;
  border-radius: 20%;
}

.logo-text {
  font-size: 1.2rem;
  font-weight: 950;
  color: var(--text-main);
  letter-spacing: -0.03em;
}

.back-nav-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text-mute);
  font-size: 0.8rem;
  font-weight: 700;
  text-decoration: none;
  transition: all 0.2s;
}

.back-nav-link:hover {
  color: var(--accent);
  transform: translateX(-3px);
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.label-caps {
  text-transform: uppercase;
  font-size: 0.75rem;
  font-weight: 900;
  color: var(--text-mute);
  letter-spacing: 0.05em;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  overflow-y: auto;
  padding: 0 0.5rem;
}

.resource-card {
  padding: 1rem;
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 1rem;
  cursor: pointer;
  border: 1px solid var(--border);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.resource-card:hover {
  border-color: var(--accent);
  transform: translateX(4px);
  background: var(--bg-card-hover);
}

/* Stats Peek Styling */
.stats-peek-inline {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%) translateX(10px);
  display: flex;
  gap: 0.75rem;
  background: var(--bg-glass);
  backdrop-filter: blur(8px);
  padding: 0.4rem 0.6rem;
  border-radius: 8px;
  border: 1px solid var(--border-light);
  opacity: 0;
  pointer-events: none;
  transition: all 0.2s ease;
  z-index: 10;
}

.resource-card:hover .stats-peek-inline {
  opacity: 1;
  transform: translateY(-50%) translateX(0);
}

.peek-stat {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.p-label {
  font-size: 0.55rem;
  font-weight: 900;
  color: var(--text-mute);
  text-transform: uppercase;
}

.p-value {
  font-size: 0.7rem;
  font-weight: 800;
  color: var(--accent);
  font-family: var(--font-mono);
  transition: color 0.3s;
}

.text-live {
  color: var(--success) !important;
  text-shadow: 0 0 8px rgba(var(--success-rgb), 0.4);
}

.resource-card.active {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.05);
}

.card-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #4b5563;
  flex-shrink: 0;
}

.card-status-dot.running {
  background: var(--success);
  box-shadow: 0 0 10px var(--success);
}

.card-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.card-name {
  font-weight: 800;
  font-size: 0.9rem;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-image-tag {
  font-size: 0.7rem;
  color: var(--text-mute);
  font-family: monospace;
}

/* SIDEBAR CONTROLS */
.mini-icon-btn {
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-mute);
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.active-toggle {
  background: #7c3aed !important;
  color: white !important;
  border-color: #7c3aed !important;
  box-shadow: 0 4px 14px rgba(124, 58, 237, 0.4);
}

.sidebar-collapsed .resources-sidebar {
  width: 0;
  margin-right: -1.5rem;
  opacity: 0;
  pointer-events: none;
}

.fixed-open-trigger {
  position: fixed;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 60px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-left: none;
  border-radius: 0 12px 12px 0;
  z-index: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* GRID */
.logs-main-content {
  flex: 1;
  min-width: 0;
  padding: 2.5rem;
  overflow-y: auto;
}
.logs-grid {
  display: grid;
  gap: 1.5rem;
  height: 100%;
}
.grid-single {
  grid-template-columns: 1fr;
}
.grid-dual {
  grid-template-columns: 1fr 1fr;
}

/* HERO EMPTY STATE */
.empty-state-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.observability-hero {
  text-align: center;
  max-width: 460px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2.5rem;
}

.hero-visual {
  position: relative;
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.radar-scan {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hero-icon {
  width: 56px;
  height: 56px;
  color: var(--accent);
  z-index: 2;
  filter: drop-shadow(0 0 15px rgba(var(--accent-rgb), 0.4));
}

.circle {
  position: absolute;
  border: 1px solid var(--accent);
  border-radius: 50%;
  opacity: 0;
  animation: radar 4s infinite linear;
}

.c1 {
  width: 60px;
  height: 60px;
  animation-delay: 0s;
}
.c2 {
  width: 60px;
  height: 60px;
  animation-delay: 1.3s;
}
.c3 {
  width: 60px;
  height: 60px;
  animation-delay: 2.6s;
}

@keyframes radar {
  0% {
    transform: scale(1);
    opacity: 0.6;
  }
  100% {
    transform: scale(3);
    opacity: 0;
  }
}

.display-title {
  font-size: 2.2rem;
  font-weight: 950;
  color: var(--text-main);
  letter-spacing: -0.04em;
  margin-bottom: 1rem;
}

.subtitle {
  font-size: 1rem;
  line-height: 1.6;
  color: var(--text-mute);
  font-weight: 500;
  margin-bottom: 2rem;
}

.sidebar-controls {
  display: flex;
  gap: 0.5rem;
}

@media (max-width: 1024px) {
  .logs-container-layout {
    flex-direction: column;
  }

  .resources-sidebar {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    width: 280px;
    z-index: 10000 !important;
    box-shadow: 20px 0 60px rgba(0, 0, 0, 0.8);
    transform: translateX(0);
    opacity: 1;
    padding: 1.25rem 1rem;
    backdrop-filter: none !important;
    transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
    visibility: visible;
  }

  .sidebar-top-nav {
    margin-bottom: 1.5rem !important;
    display: flex !important;
    flex-direction: column !important;
    gap: 0.75rem !important;
  }

  .mobile-header-row {
    display: flex !important;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--border);
  }

  .sidebar-logo-mobile {
    display: flex;
    align-items: center;
    gap: 0.85rem;
    text-decoration: none;
  }

  .logo-icon-premium {
    width: 36px;
    height: 36px;
    background: var(--accent);
    color: #fff;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 0 20px rgba(99, 102, 241, 0.4);
  }

  .logo-text-premium {
    font-size: 1.25rem;
    font-weight: 900;
    letter-spacing: -0.03em;
    color: #fff;
  }

  .minimal-close-btn-glass {
    width: 40px;
    height: 40px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 10px;
    color: var(--text-mute);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  }

  .minimal-close-btn-glass:hover {
    background: rgba(239, 68, 68, 0.15);
    color: var(--error);
    border-color: rgba(239, 68, 68, 0.3);
    transform: rotate(90deg);
  }

  .minimal-close-btn-glass:active {
    transform: scale(0.9);
  }

  .back-nav-link-mobile {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--text-main);
    text-decoration: none;
    font-weight: 850;
    font-size: 0.9rem;
    padding: 0.5rem 0.75rem;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    transition: all 0.2s;
  }

  .back-nav-link-mobile:hover {
    background: rgba(99, 102, 241, 0.1);
    color: var(--accent);
    border-color: rgba(99, 102, 241, 0.3);
  }

  .sidebar-header {
    margin-bottom: 1rem !important;
  }

  .sidebar-collapsed .resources-sidebar {
    transform: translateX(-100%);
    opacity: 0;
    visibility: hidden;
    pointer-events: none;
  }

  .logs-main-content {
    padding: 1rem;
    height: 100vh;
    z-index: 1;
    position: relative;
  }

  .grid-dual {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr 1fr;
  }

  .display-title {
    font-size: 1.75rem;
  }

  .hide-mobile {
    display: none !important;
  }

  .show-mobile {
    display: flex !important;
  }

  .mobile-close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    padding: 1rem;
    background: rgba(239, 68, 68, 0.08);
    color: #ef4444;
    border: 1px solid rgba(239, 68, 68, 0.15);
    border-radius: 18px;
    font-size: 0.9rem;
    font-weight: 900;
    cursor: pointer;
    margin-top: 0.5rem;
    width: 100%;
    transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  }

  .mobile-close-btn:hover {
    background: #ef4444;
    color: #fff;
    box-shadow: 0 10px 25px rgba(239, 68, 68, 0.3);
    transform: translateY(-2px);
  }

  .mobile-sidebar-overlay {
    position: fixed;
    inset: 0;
    background: rgba(2, 6, 23, 0.8);
    z-index: 4000;
    animation: fade-in 0.4s ease;
  }
}

.hero-actions {
  justify-items: center;
}

@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.show-mobile {
  display: none;
}

@media (max-width: 600px) {
  .resources-sidebar {
    width: 100%;
    max-width: none;
  }
}
</style>

<template>
  <div
    :class="[
      'log-viewer',
      'glass',
      'animate-fade-in',
      { fullscreen: isFullScreen },
    ]"
  >
    <div :class="['viewer-header', { 'compact-header': compactHeader }]">
      <div class="header-left">
        <div :class="['status-pulse', `status-${container.state}`]"></div>
        <div class="name-group">
          <span class="c-name">{{ container.name }}</span>
          <span class="c-id">{{ container.id.substring(0, 12) }}</span>
        </div>
      </div>

      <div class="header-right">
        <!-- Live Stats in Header -->
        <div v-if="container.state === 'running'" class="header-stats-live">
          <div class="h-stat">
            <span class="h-label">CPU</span>
            <span
              class="h-value"
              :style="{ color: getStatColor(stats.cpu || 0) }"
              >{{ stats.cpu }}%
              <small v-if="stats.assignedCores"
                >/ {{ stats.assignedCores }} Core{{
                  parseFloat(stats.assignedCores) > 1 ? "s" : ""
                }}</small
              ></span
            >
          </div>
          <div class="h-stat">
            <span class="h-label">MEM</span>
            <span class="h-value">{{ stats.memory || "0B / 0B" }}</span>
          </div>
        </div>

        <div class="log-search glass">
          <svg
            viewBox="0 0 24 24"
            width="12"
            height="12"
            stroke="currentColor"
            stroke-width="3"
            fill="none"
          >
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <input
            type="text"
            v-model="logSearchQuery"
            placeholder="Search..."
            class="search-input"
          />
        </div>

        <div class="action-buttons">
          <button
            @click="isFullScreen = !isFullScreen"
            class="icon-btn"
            :title="isFullScreen ? 'Exit Full Screen' : 'Full Screen'"
          >
            <svg
              v-if="!isFullScreen"
              viewBox="0 0 24 24"
              width="14"
              height="14"
              stroke="currentColor"
              stroke-width="2.5"
              fill="none"
            >
              <path d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7"></path>
            </svg>
            <svg
              v-else
              viewBox="0 0 24 24"
              width="14"
              height="14"
              stroke="currentColor"
              stroke-width="2.5"
              fill="none"
            >
              <path d="M4 14h6v6M20 10h-6V4M14 10l7-7M10 14l-7 7"></path>
            </svg>
          </button>
          <button
            @click="showDownloadModal = true"
            class="icon-btn"
            title="Download"
          >
            <svg
              viewBox="0 0 24 24"
              width="14"
              height="14"
              stroke="currentColor"
              stroke-width="2.5"
              fill="none"
            >
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
              <polyline points="7 10 12 15 17 10"></polyline>
              <line x1="12" y1="15" x2="12" y2="3"></line>
            </svg>
          </button>
          <button @click="clearLogs" class="icon-btn" title="Clear">
            <svg
              viewBox="0 0 24 24"
              width="14"
              height="14"
              stroke="currentColor"
              stroke-width="2.5"
              fill="none"
            >
              <polyline points="3 6 5 6 21 6"></polyline>
              <path
                d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
              ></path>
            </svg>
          </button>
          <button
            v-if="showClose"
            @click="$emit('close')"
            class="icon-btn stop"
            title="Close"
          >
            <svg
              viewBox="0 0 24 24"
              width="14"
              height="14"
              stroke="currentColor"
              stroke-width="3"
              fill="none"
            >
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div class="viewer-body">
      <!-- Download Modal -->
      <Teleport to="body">
        <Transition name="fade">
          <div v-if="showDownloadModal" class="modal-overlay">
            <div class="modal-content shadow-2xl">
              <h3>Download Logs</h3>
              <p class="text-mute">
                Export buffer for <strong>{{ container.name }}</strong>
              </p>
              <div class="format-grid mt-6">
                <button
                  @click="downloadLogs('txt')"
                  class="modal-btn secondary"
                >
                  TXT
                </button>
                <button
                  @click="downloadLogs('json')"
                  class="modal-btn secondary"
                >
                  JSON
                </button>
                <button
                  @click="downloadFullLogs"
                  class="modal-btn confirm full-width mt-2"
                >
                  Full History (.log)
                </button>
              </div>
              <button
                @click="showDownloadModal = false"
                class="modal-btn cancel mt-4"
              >
                Close
              </button>
            </div>
          </div>
        </Transition>
      </Teleport>

      <div ref="logContainer" class="log-content" @scroll="handleScroll">
        <!-- Sentinel for IntersectionObserver -->
        <div
          ref="scrollSentinel"
          class="scroll-sentinel"
          style="height: 50px; width: 100%"
        ></div>

        <!-- Manual Load Trigger -->
        <div v-if="hasMoreHistory" class="history-trigger-container">
          <button
            v-if="!isLoadingHistory"
            @click="fetchHistoricalLogs"
            class="history-btn-manual"
          >
            <svg
              viewBox="0 0 24 24"
              width="16"
              height="16"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <polyline points="18 15 12 9 6 15"></polyline>
            </svg>
            <span>Load more history ({{ logs.length }} / {{ totalLogs }})</span>
          </button>
          <div v-else class="history-loading-indicator">
            <div class="mini-spinner"></div>
            <span>Fetching history...</span>
          </div>
        </div>
        <div v-else-if="logs.length > 0" class="history-end-msg">
          Beginning of history reached ({{ totalLogs }} logs)
        </div>

        <div v-for="(log, i) in displayLogs" :key="i" class="log-line">
          <span class="line-num">{{ i + 1 }}</span>
          <span class="line-text" v-html="formatLog(log)"></span>
        </div>
        <div v-if="displayLogs.length === 0" class="log-empty">
          <p class="text-mute">Waiting for stream...</p>
        </div>
      </div>

      <button
        v-if="!autoScroll"
        @click="scrollToBottom"
        class="resume-scroll-btn glass"
      >
        <svg
          viewBox="0 0 24 24"
          width="14"
          height="14"
          stroke="currentColor"
          stroke-width="3"
          fill="none"
        >
          <polyline points="7 13 12 18 17 13"></polyline>
          <polyline points="7 6 12 11 17 6"></polyline>
        </svg>
        Resume Scroll
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from "vue";
import { secureStorage } from "../utils/storage";

const props = defineProps({
  container: Object,
  showClose: Boolean,
  compactHeader: Boolean,
});

const emit = defineEmits(["close", "stats"]);

const logs = ref([]);
const logSearchQuery = ref("");
const logContainer = ref(null);
const scrollSentinel = ref(null);
const autoScroll = ref(true);
const showDownloadModal = ref(false);
const isFullScreen = ref(false);
const stats = ref({ cpu: "0.00", memory: "0B / 0B", memPercent: 0 });
const isLoadingHistory = ref(false);
const hasMoreHistory = ref(true);
const totalLogs = ref(0);
let lastFetchedUntil = null;
let socket = null;
let statsController = null;
let observer = null;

const formatLog = (text) => {
  // Strip Docker timestamp if present (it's always at the start)
  let cleanText = text.replace(
    /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z\s?/,
    "",
  );

  let formatted = cleanText
    .replace(/\033\[[0-9;]*m/g, "")
    .replace(/(ERROR|ERR|Fail|Failed)/gi, '<span class="text-error">$1</span>')
    .replace(/(WARN|Warning)/gi, '<span class="text-warning">$1</span>')
    .replace(/(INFO|OK|Success)/gi, '<span class="text-success">$1</span>');

  if (logSearchQuery.value && logSearchQuery.value.length >= 2) {
    const regex = new RegExp(
      `(${logSearchQuery.value.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, "\\$&")})`,
      "gi",
    );
    formatted = formatted.replace(
      regex,
      '<mark class="log-highlight">$1</mark>',
    );
  }
  return formatted;
};

const displayLogs = computed(() => {
  if (!logSearchQuery.value || logSearchQuery.value.length < 2)
    return logs.value;
  return logs.value.filter((l) =>
    l.toLowerCase().includes(logSearchQuery.value.toLowerCase()),
  );
});

const scrollToBottom = () => {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight;
    autoScroll.value = true;
  }
};

const fetchHistoricalLogs = async () => {
  if (isLoadingHistory.value || !hasMoreHistory.value) return;

  const earliestLog = logs.value[0];
  if (!earliestLog) return;

  // Extract timestamp from the first log line (flexible for varying precision)
  const tsMatch = earliestLog.match(
    /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z/,
  );
  if (!tsMatch) {
    console.warn(
      "No timestamp found in earliest log:",
      earliestLog.substring(0, 50),
    );
    return;
  }

  const until = tsMatch[0];

  if (until === lastFetchedUntil) return;

  console.log("[LogViewer] Fetching history until:", until);
  lastFetchedUntil = until;
  isLoadingHistory.value = true;

  try {
    const token = secureStorage.getItem("token");
    const res = await fetch(
      `/api/containers/${props.container.id}/logs?tail=100&until=${encodeURIComponent(until)}`,
      {
        headers: { Authorization: `Bearer ${token}` },
      },
    );
    if (res.ok) {
      const newLogs = await res.json();
      const logsCount = Array.isArray(newLogs) ? newLogs.length : 0;
      console.log(`[LogViewer] Received ${logsCount} lines from backend`);

      if (!Array.isArray(newLogs) || logsCount === 0) {
        console.log(
          "[LogViewer] No valid historical logs found, assuming start reached",
        );
        hasMoreHistory.value = false;
        return;
      }

      // Filter out duplicates (Until is inclusive)
      const existingLogs = new Set(logs.value.map((l) => l.trim()));
      const filtered = newLogs.filter(
        (nl) => nl && !existingLogs.has(nl.trim()),
      );

      console.log(
        `[LogViewer] ${filtered.length} new lines after filtering duplicates`,
      );

      if (filtered.length === 0) {
        console.log(
          "[LogViewer] No unique historical logs found, assuming start reached",
        );
        hasMoreHistory.value = false;
      } else {
        const container = logContainer.value;
        const oldScrollHeight = container.scrollHeight;

        logs.value = [...filtered, ...logs.value];

        if (newLogs.length < 100) {
          console.log(
            "[LogViewer] Received < 100 lines, end of history reached",
          );
          hasMoreHistory.value = false;
        }

        nextTick(() => {
          if (container) {
            container.scrollTop = container.scrollHeight - oldScrollHeight;
          }
        });
      }
    }
  } catch (err) {
    console.error("Failed to fetch historical logs:", err);
  } finally {
    isLoadingHistory.value = false;
  }
};

const handleScroll = () => {
  if (!logContainer.value) return;
  const { scrollTop, scrollHeight, clientHeight } = logContainer.value;

  // Update auto-scroll state based on distance from bottom
  autoScroll.value = scrollHeight - scrollTop - clientHeight < 100;
};

const setupObserver = () => {
  if (observer) observer.disconnect();

  observer = new IntersectionObserver(
    (entries) => {
      if (entries[0].isIntersecting) {
        if (
          !isLoadingHistory.value &&
          hasMoreHistory.value &&
          logs.value.length > 0
        ) {
          console.log("[LogViewer] Sentinel triggered history fetch");
          fetchHistoricalLogs();
        }
      }
    },
    {
      root: logContainer.value,
      rootMargin: "200px 0px 0px 0px", // Start loading 200px before reaching top
      threshold: 0,
    },
  );

  if (scrollSentinel.value) {
    observer.observe(scrollSentinel.value);
  }
};

const clearLogs = () => (logs.value = []);

const getStatColor = (val) => {
  const n = parseFloat(val);
  if (n > 80) return "var(--error)";
  if (n > 50) return "var(--warning)";
  return "var(--success)";
};

const formatBytes = (bytes) => {
  if (bytes === 0) return "0B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + sizes[i];
};

const downloadLogs = (format) => {
  const rawLogs = logs.value.map((l) => l.replace(/\033\[[0-9;]*m/g, ""));
  let content =
    format === "json" ? JSON.stringify(rawLogs, null, 2) : rawLogs.join("\n");
  const blob = new Blob([content], {
    type: format === "json" ? "application/json" : "text/plain",
  });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${props.container.name}_logs.${format}`;
  a.click();
  showDownloadModal.value = false;
};

const downloadFullLogs = async () => {
  try {
    const token = secureStorage.getItem("token");
    const res = await fetch(
      `/api/containers/${props.container.id}/logs/download`,
      {
        headers: { Authorization: `Bearer ${token}` },
      },
    );
    if (res.ok) {
      const blob = await res.blob();
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = `${props.container.name}_full.log`;
      a.click();
      showDownloadModal.value = false;
    }
  } catch (err) {
    console.error(err);
  }
};

const fetchStats = async () => {
  if (statsController) statsController.abort();
  statsController = new AbortController();
  try {
    const token = secureStorage.getItem("token");
    const response = await fetch(
      `/api/containers/${props.container.id}/stats`,
      {
        headers: { Authorization: `Bearer ${token}` },
        signal: statsController.signal,
      },
    );
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      const chunk = decoder.decode(value);
      const lines = chunk.split("\n").filter((l) => l.trim());
      for (const line of lines) {
        try {
          const data = JSON.parse(line);
          const cpuDelta =
            data.cpu_stats.cpu_usage.total_usage -
            (data.precpu_stats?.cpu_usage?.total_usage || 0);
          const systemDelta =
            data.cpu_stats.system_cpu_usage -
            (data.precpu_stats?.system_cpu_usage || 0);
          
          const onlineCPUs = data.cpu_stats.online_cpus || 1;
          let cpuPercent = 0;
          if (systemDelta > 0 && cpuDelta > 0) {
            cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0;
          }

          const used =
            data.memory_stats.usage - (data.memory_stats.stats?.cache || 0);
          const quota = data.cpu_stats.cpu_quota || 0;
          const period = data.cpu_stats.cpu_period || 100000;
          
          // Priority: 1. HostConfig limit (from props), 2. CFS Quota (from stats), 3. Total Cores
          let assignedCores = props.container.cpu_limit || (quota > 0 ? (quota / period) : onlineCPUs);
          if (typeof assignedCores === 'number') assignedCores = assignedCores.toFixed(1);

          stats.value = {
            cpu: cpuPercent.toFixed(2),
            cores: onlineCPUs,
            assignedCores: assignedCores,
            memory: `${formatBytes(used)} / ${formatBytes(data.memory_stats.limit)}`,
          };
          emit("stats", { id: props.container.id, ...stats.value });
        } catch (e) {}
      }
    }
  } catch (e) {
    if (e.name !== "AbortError") setTimeout(fetchStats, 5000);
  }
};

const fetchLogCount = async () => {
  try {
    const token = secureStorage.getItem("token");
    const res = await fetch(
      `/api/containers/${props.container.id}/logs/count`,
      {
        headers: { Authorization: `Bearer ${token}` },
      },
    );
    if (res.ok) {
      const data = await res.json();
      totalLogs.value = data.total;
    }
  } catch (err) {
    console.error("Failed to fetch log count:", err);
  }
};

const connect = () => {
  if (socket) socket.close();
  const protocol = location.protocol === "https:" ? "wss:" : "ws:";
  const token = secureStorage.getItem("token");
  socket = new WebSocket(
    `${protocol}//${location.host}/ws/logs/${props.container.id}?token=${token}`,
  );
  socket.onmessage = (e) => {
    logs.value.push(e.data);
    // Limit buffer to 5000 lines, only prune if auto-scrolling to preserve history exploration
    if (autoScroll.value && logs.value.length > 5000) {
      logs.value.shift();
    }
    if (autoScroll.value) nextTick(scrollToBottom);
  };
};

onMounted(() => {
  connect();
  fetchStats();
  fetchLogCount();
  nextTick(setupObserver);
});
onUnmounted(() => {
  if (socket) socket.close();
  if (statsController) statsController.abort();
  if (observer) observer.disconnect();
});
watch(
  () => props.container.id,
  () => {
    logs.value = [];
    totalLogs.value = 0;
    connect();
    fetchStats();
    fetchLogCount();
  },
);
</script>

<style scoped>
.log-viewer {
  display: flex;
  flex-direction: column;
  background: var(--log-bg);
  border-radius: 20px;
  overflow: hidden;
  height: calc(100vh - 140px);
}

.viewer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.25rem;
  background: var(--glass-bg);
  border-bottom: 1px solid var(--border);
  backdrop-filter: blur(20px);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-pulse {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.status-running {
  background: var(--success);
  box-shadow: 0 0 8px var(--success);
}
.status-exited {
  background: var(--text-mute);
}

.c-name {
  font-size: 0.85rem;
  font-weight: 900;
  color: var(--text-main);
}

.c-id {
  font-size: 0.65rem;
  color: var(--text-mute);
  font-family: var(--font-mono);
  margin-left: 0.5rem;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header-stats-live {
  display: flex;
  gap: 1rem;
  padding-right: 1rem;
  border-right: 1px solid var(--border);
}

.h-stat {
  display: flex;
  flex-direction: column;
}

.h-label {
  font-size: 0.6rem;
  font-weight: 900;
  color: var(--text-mute);
}

.h-value {
  font-size: 0.75rem;
  font-weight: 800;
  font-family: var(--font-mono);
}

.log-search {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.75rem;
  border-radius: 8px;
  background: var(--bg-input);
}

.search-input {
  background: transparent;
  border: none;
  color: var(--text-main);
  font-size: 0.75rem;
  font-weight: 600;
  width: 80px;
  outline: none;
}

.action-buttons {
  display: flex;
  gap: 0.6rem;
}

.icon-btn {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-mute);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.23, 1, 0.32, 1);
}

.icon-btn:hover {
  background: var(--bg-card);
  color: var(--text-main);
  border-color: var(--accent);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.icon-btn.stop:hover {
  color: var(--error);
  border-color: var(--error);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.2);
}

.viewer-body {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.log-content {
  height: 100%;
  overflow-y: auto;
  padding: 1.5rem;
  font-family: "JetBrains Mono", monospace;
  font-size: 0.75rem;
  line-height: 1.6;
  color: var(--text-main);
  position: relative;
}

.scroll-sentinel {
  height: 1px;
  width: 100%;
  position: absolute;
  top: 0;
  pointer-events: none;
}

.log-line {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 0.2rem;
}

.line-num {
  color: var(--text-mute);
  width: 40px;
  text-align: right;
  flex-shrink: 0;
  user-select: none;
  font-size: 0.7rem;
  opacity: 0.5;
}

.line-text {
  flex: 1;
  min-width: 0;
  word-break: break-word;
  white-space: pre-wrap;
}

.history-trigger-container {
  display: flex;
  justify-content: center;
  padding: 1rem 0;
  margin-bottom: 1rem;
  border-bottom: 1px solid var(--border-light);
}

.history-btn-manual {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 1.2rem;
  background: var(--primary-light);
  border: 1px solid var(--primary);
  border-radius: 8px;
  color: var(--text-main);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.history-btn-manual:hover {
  background: var(--primary);
  color: white;
  transform: translateY(-2px);
}

.history-loading-indicator {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--text-mute);
  font-size: 0.8rem;
}

.history-end-msg {
  text-align: center;
  padding: 1rem;
  color: var(--text-mute);
  font-size: 0.8rem;
  font-style: italic;
  opacity: 0.7;
}

.mini-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.resume-scroll-btn {
  position: absolute;
  bottom: 1.5rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 0.6rem 1rem;
  border-radius: 10px;
  font-size: 0.75rem;
  font-weight: 800;
  color: #fff;
  background: var(--accent);
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.4);
}

.log-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}
.log-viewer.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 9999;
  border-radius: 0;
  background: var(--log-bg);
}

.log-viewer.fullscreen .viewer-body {
  height: calc(100vh - 70px);
}

@media (max-width: 600px) {
  .viewer-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 1rem;
    z-index: 10 !important;
    background: var(--bg-main) !important;
    backdrop-filter: none !important;
  }
  .header-right {
    width: 100%;
    gap: 0.5rem;
    justify-content: space-between;
  }
  .header-stats-live {
    display: none;
  }
  .log-search {
    flex: 1;
    min-width: 0;
  }
  .search-input {
    width: 100%;
  }
  .log-content {
    padding: 1rem;
  }
  .log-line {
    gap: 0.75rem;
  }
  .line-num {
    width: 28px;
    font-size: 0.6rem;
  }
  .line-text {
    font-size: 0.7rem;
  }
  .action-buttons .icon-btn {
    width: 28px;
    height: 28px;
  }
}

@media (max-width: 400px) {
  .log-content {
    padding: 0.75rem;
  }
  .line-num {
    width: 24px;
    font-size: 0.55rem;
  }
  .line-text {
    font-size: 0.65rem;
  }
  .log-line {
    gap: 0.5rem;
  }
}
</style>

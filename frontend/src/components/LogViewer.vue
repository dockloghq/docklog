<template>
  <div class="log-viewer glass" @mouseenter="$emit('stats', stats)">
    <div :class="['viewer-header', { 'compact-header': compactHeader }]">
      <div class="header-left">
        <div :class="['status-pulse', `status-${container.state}`]"></div>
        <div class="name-group">
          <span class="c-name">{{ container.name }}</span>
          <span class="c-id">{{ container.id }}</span>
        </div>
      </div>

      <div class="header-right">
        <!-- Log Search -->
        <div class="log-search">
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
            placeholder="Search logs..."
            class="search-input"
          />
        </div>

        <!-- Live Stats in Header -->
        <div v-if="container.state === 'running'" class="header-stats-live">
          <div class="h-stat">
            <span class="h-label">CPU</span>
            <span class="h-value" :style="{ color: getStatColor(stats.cpu) }"
              >{{ stats.cpu }}%</span
            >
          </div>
          <div class="h-stat">
            <span class="h-label">MEM</span>
            <span class="h-value">{{ stats.memory.split("/")[0] }}</span>
          </div>
        </div>

        <button
          @click="showDownloadModal = true"
          class="header-btn"
          title="Download Logs"
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
        <button @click="clearLogs" class="header-btn" title="Clear Buffer">
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
          class="header-btn close"
          title="Close Pane"
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

    <div class="viewer-body">
      <!-- Download Format Modal -->
      <div
        v-if="showDownloadModal"
        class="download-modal glass"
        @click.self="showDownloadModal = false"
      >
        <div class="modal-content glass-card">
          <h3>Download Logs</h3>
          <p>
            Select your preferred export format for
            <strong>{{ container.name }}</strong>
          </p>

          <div class="format-grid">
            <div class="download-section">
              <label class="section-label">Recent Buffer (Fast)</label>
              <div class="section-btns">
                <button @click="downloadLogs('txt')" class="format-btn mini">
                  .txt
                </button>
                <button @click="downloadLogs('json')" class="format-btn mini">
                  .json
                </button>
                <button @click="downloadLogs('csv')" class="format-btn mini">
                  .csv
                </button>
              </div>
            </div>

            <div class="download-divider"></div>

            <div class="download-section">
              <label class="section-label">Server Stream (Full History)</label>
              <button @click="downloadFullLogs" class="format-btn full-history">
                <svg
                  viewBox="0 0 24 24"
                  width="16"
                  height="16"
                  stroke="currentColor"
                  stroke-width="2.5"
                  fill="none"
                >
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                  <polyline points="7 10 12 15 17 10"></polyline>
                  <line x1="12" y1="15" x2="12" y2="3"></line>
                </svg>
                Download full_history.log
              </button>
              <p class="section-hint">
                Retrieves all historical logs directly from the Docker engine.
              </p>
            </div>
          </div>

          <button @click="showDownloadModal = false" class="modal-close-btn">
            Close
          </button>
        </div>
      </div>

      <div ref="logContainer" class="log-content" @scroll="handleScroll">
        <div v-for="(log, i) in displayLogs" :key="i" class="log-line">
          <span class="line-num">{{ i + 1 }}</span>
          <span class="line-text" v-html="formatLog(log)"></span>
        </div>
      </div>

      <button
        v-if="!autoScroll"
        @click="scrollToBottom"
        class="scroll-bottom-btn glass"
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
const autoScroll = ref(true);
const showDownloadModal = ref(false);
const stats = ref({ cpu: "0.00", memory: "0B / 0B", memPercent: 0 });
let socket = null;
let statsController = null;

const formatLog = (text) => {
  let formatted = text
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

const handleScroll = () => {
  if (!logContainer.value) return;
  const { scrollTop, scrollHeight, clientHeight } = logContainer.value;
  autoScroll.value = scrollHeight - scrollTop - clientHeight < 50;
};

const clearLogs = () => {
  logs.value = [];
};

const getStatColor = (val) => {
  const n = parseFloat(val);
  if (n > 80) return "var(--error)";
  if (n > 50) return "#f59e0b";
  return "var(--success)";
};

const formatBytes = (bytes) => {
  if (bytes === 0) return "0B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + sizes[i];
};

const downloadLogs = (format) => {
  const rawLogs = logs.value.map((l) => l.replace(/\033\[[0-9;]*m/g, ""));
  let content = "";
  let mimeType = "text/plain";
  const filename = `${props.container.name}_buffer_${new Date().getTime()}.${format}`;

  if (format === "txt") {
    content = rawLogs.join("\n");
  } else if (format === "json") {
    content = JSON.stringify(
      rawLogs.map((l) => ({ timestamp: new Date().toISOString(), message: l })),
      null,
      2,
    );
    mimeType = "application/json";
  } else if (format === "csv") {
    content =
      "Timestamp,Message\n" +
      rawLogs
        .map((l) => `"${new Date().toISOString()}","${l.replace(/"/g, '""')}"`)
        .join("\n");
    mimeType = "text/csv";
  }

  const blob = new Blob([content], { type: mimeType });
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  window.URL.revokeObjectURL(url);
  showDownloadModal.value = false;
};

const downloadFullLogs = async () => {
  try {
    const token = secureStorage.getItem("token");
    const response = await fetch(
      `/api/containers/${props.container.id}/logs/download`,
      {
        headers: { Authorization: `Bearer ${token}` },
      },
    );

    if (response.ok) {
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${props.container.name}_full.log`;
      a.click();
      window.URL.revokeObjectURL(url);
      showDownloadModal.value = false;
    } else {
      const err = await response.json();
      alert(err.error || "Failed to download full logs");
    }
  } catch (err) {
    console.error(err);
    alert("Connection error while downloading logs");
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
            data.precpu_stats.cpu_usage.total_usage;
          const systemDelta =
            data.cpu_stats.system_cpu_usage -
            data.precpu_stats.system_cpu_usage;
          let cpuPercent = 0.0;
          if (systemDelta > 0 && cpuDelta > 0) {
            cpuPercent =
              (cpuDelta / systemDelta) *
              (data.cpu_stats.online_cpus || 1) *
              100.0;
          }

          const used =
            data.memory_stats.usage - (data.memory_stats.stats.cache || 0);
          const limit = data.memory_stats.limit;
          const cpu = cpuPercent.toFixed(1);
          const memory = formatBytes(used);
          const limitStr = formatBytes(limit);

          stats.value = {
            cpu: cpu,
            memory: `${memory} / ${limitStr}`,
            memPercent: ((used / limit) * 100).toFixed(2),
          };

          emit("stats", {
            id: props.container.id,
            cpu: stats.value.cpu,
            memory: stats.value.memory,
            rawUsed: used,
            rawLimit: limit,
          });
        } catch (e) {}
      }
    }
  } catch (e) {
    if (e.name !== "AbortError") {
      console.error("Stats error:", e);
      // Attempt to resume after 5 seconds
      setTimeout(fetchStats, 5000);
    }
  }
};

const connect = () => {
  if (socket) socket.close();
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const type = props.container.type || 'docker';
  socket = new WebSocket(
    `${protocol}//${window.location.host}/ws/logs/${props.container.id}?token=${token}&type=${type}`,
  );
  socket.onmessage = (event) => {
    logs.value.push(event.data);
    if (logs.value.length > 1000) logs.value.shift();
    if (autoScroll.value) nextTick(scrollToBottom);
  };
  socket.onclose = () => {
    // Attempt to reconnect after 3 seconds if component is still mounted
    setTimeout(() => {
      if (socket && socket.readyState === WebSocket.CLOSED) {
        connect();
      }
    }, 3000);
  };
  socket.onerror = (err) => {
    console.error("WebSocket error:", err);
    socket.close();
  };
};

onMounted(() => {
  connect();
  if (props.container.type !== 'pod') {
    fetchStats();
  }
});

onUnmounted(() => {
  if (socket) socket.close();
  if (statsController) statsController.abort();
});

watch(
  () => props.container.id,
  () => {
    connect();
    if (props.container.type !== 'pod') {
      fetchStats();
    }
  },
);
</script>

<style scoped>
.log-viewer {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-sidebar);
  border-radius: 0;
  overflow: hidden;
  border-left: 1px solid var(--border);
}

.viewer-header {
  height: 72px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 2.5rem;
  background: var(--glass-bg);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  min-width: 0;
  flex: 1;
}
.name-group {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
}
.status-pulse {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 2px solid var(--bg-sidebar);
  flex-shrink: 0;
}
.status-pulse.status-running {
  background: var(--success);
  box-shadow: 0 0 12px var(--success);
  animation: pulse 2s infinite;
}
.status-pulse.status-exited {
  background: var(--text-mute);
}

.c-name {
  font-size: 0.95rem;
  font-weight: 950;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: -0.03em;
  transition: all 0.2s;
}
.c-name:hover {
  white-space: normal;
  word-break: break-all;
  overflow: visible;
  text-overflow: clip;
}
.c-id {
  font-size: 0.62rem;
  color: var(--text-mute);
  font-family: var(--font-mono);
  opacity: 0.5;
  font-weight: 600;
  letter-spacing: 0.05em;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0;
}

.viewer-header.compact-header {
  padding: 0 1.25rem;
}

.viewer-header.compact-header .header-stats-live {
  display: none;
}

.viewer-header.compact-header .log-search {
  width: 140px;
}

.viewer-header.compact-header .log-search:focus-within {
  width: 170px;
}

.viewer-header.compact-header .search-input {
  width: 70px;
}

.viewer-header.compact-header .header-btn {
  padding: 0.4rem 0.7rem;
}

.viewer-header.compact-header .header-left {
  gap: 0.9rem;
}

.viewer-header.compact-header .header-right {
  gap: 0.5rem;
}

.log-search {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  padding: 0.5rem 1rem;
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
}
.log-search:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.08);
  width: 200px;
}
.log-search svg {
  color: var(--text-mute);
  width: 14px;
  height: 14px;
}
.search-input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 0.8rem;
  font-weight: 600;
  width: 100px;
  transition: width 0.3s;
}
.search-input::placeholder {
  color: var(--text-mute);
  font-weight: 500;
}

.header-stats-live {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-right: 0.25rem;
  padding-right: 1rem;
  border-right: 1px solid var(--border);
}
.h-stat {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  min-width: 60px;
}
.h-label {
  font-size: 0.65rem;
  font-weight: 900;
  color: var(--text-mute);
  letter-spacing: 0.1em;
  text-transform: uppercase;
  margin-bottom: 2px;
}
.h-value {
  font-size: 0.85rem;
  font-weight: 850;
  font-family: var(--font-mono);
  color: var(--text-main);
  letter-spacing: -0.02em;
}

.header-btn {
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-dim);
  padding: 0.45rem 0.85rem;
  border-radius: 10px;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.72rem;
  font-weight: 850;
  cursor: pointer;
  transition: all 0.2s;
}
.header-btn:hover {
  border-color: var(--text-mute);
  color: var(--text-main);
  background: var(--bg-card);
  transform: translateY(-1px);
}
.header-btn.active {
  background: var(--accent);
  color: #fff;
  border-color: var(--accent);
  box-shadow: 0 8px 16px rgba(99, 102, 241, 0.3);
}
.header-btn.close:hover {
  background: var(--error);
  color: #fff;
  border-color: var(--error);
  box-shadow: 0 8px 16px rgba(239, 68, 68, 0.2);
}

.viewer-body {
  flex: 1;
  position: relative;
  overflow: hidden;
  background: var(--log-bg);
  transition: background 0.4s ease;
}

.history-overlay {
  position: absolute;
  inset: 0;
  background: var(--glass-bg);
  backdrop-filter: blur(30px);
  z-index: 20;
  padding: 2.5rem;
  display: flex;
  flex-direction: column;
  animation: fadeIn 0.4s cubic-bezier(0.23, 1, 0.32, 1);
}
.history-header {
  margin-bottom: 2.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.h-title-group h3 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 950;
  color: var(--text-main);
  letter-spacing: -0.03em;
}
.h-subtitle {
  font-size: 0.75rem;
  color: var(--text-mute);
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin-top: 4px;
}

.f-pill {
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-dim);
  padding: 0.5rem 1.25rem;
  border-radius: 100px;
  font-size: 0.8rem;
  font-weight: 900;
  cursor: pointer;
  transition: all 0.2s;
}
.f-pill:hover {
  background: var(--bg-card);
  color: var(--text-main);
  border-color: var(--text-mute);
}
.f-pill.active {
  background: var(--accent);
  color: #fff;
  border-color: var(--accent);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.history-chart-container {
  margin-bottom: 2.5rem;
  background: var(--bg-input);
  border-radius: 24px;
  padding: 1.5rem;
  position: relative;
  border: 1px solid var(--border);
}

.h-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-right: 0.5rem;
}
.h-row {
  display: flex;
  align-items: center;
  padding: 0.75rem 1.25rem;
  background: var(--bg-input);
  border-radius: 14px;
  border: 1px solid var(--border);
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
}
.h-row:hover {
  background: var(--bg-card);
  border-color: var(--text-mute);
  transform: translateX(8px);
}

.h-stats-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  border-left: 1px solid var(--border);
  padding-left: 1.25rem;
  margin-left: 1.25rem;
}
.h-mini-val {
  font-size: 0.85rem;
  font-weight: 900;
  font-family: var(--font-mono);
  color: var(--text-main);
  letter-spacing: -0.05em;
}

.log-content {
  height: 100%;
  overflow-y: auto;
  padding: 2.5rem 3rem;
  font-family: var(--font-mono);
  font-size: 0.78rem;
  line-height: 1.7;
  color: var(--text-main);
}

.log-line {
  display: flex;
  gap: 2rem;
  white-space: pre-wrap;
  word-break: break-all;
  margin-bottom: 0.4rem;
  transition: background 0.2s;
  padding: 0.1rem 0.5rem;
  border-radius: 4px;
}
.log-line:hover {
  background: var(--card-hover);
}
.line-num {
  color: var(--text-mute);
  min-width: 3rem;
  text-align: right;
  user-select: none;
  opacity: 0.4;
  font-size: 0.8rem;
  font-weight: 500;
}

.text-error {
  color: var(--error);
  font-weight: 800;
}
.text-warning {
  color: var(--warning);
  font-weight: 800;
}
.text-success {
  color: var(--success);
  font-weight: 800;
}

.log-highlight {
  background: var(--accent);
  color: #fff;
  padding: 0 6px;
  border-radius: 6px;
  font-weight: 900;
  box-shadow: 0 0 15px rgba(99, 102, 241, 0.5);
}

.scroll-bottom-btn {
  position: absolute;
  bottom: 2.5rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 1rem 2rem;
  border-radius: 100px;
  background: var(--accent);
  color: #fff;
  font-size: 0.95rem;
  font-weight: 900;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  z-index: 10;
  box-shadow: 0 15px 35px rgba(99, 102, 241, 0.5);
  border: none;
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
}
.scroll-bottom-btn:hover {
  transform: translateX(-50%) translateY(-5px);
  box-shadow: 0 20px 45px rgba(99, 102, 241, 0.6);
}

@keyframes pulse {
  0% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.2);
    opacity: 0.7;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

.download-modal {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(20px);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 3rem;
}
[data-theme="dark"] .download-modal {
  background: rgba(0, 0, 0, 0.85);
}
.glass-card {
  background: var(--bg-sidebar);
  border: 1px solid var(--border);
  padding: 4rem;
  border-radius: 32px;
  width: 100%;
  max-width: 500px;
  text-align: center;
  box-shadow: 0 50px 120px -20px var(--shadow);
}
.glass-card h3 {
  font-size: 1.75rem;
  font-weight: 950;
  margin: 0 0 0.85rem 0;
  color: var(--text-main);
  letter-spacing: -0.03em;
}
.glass-card p {
  font-size: 1.1rem;
  color: var(--text-dim);
  margin-bottom: 2.5rem;
  line-height: 1.5;
}

.format-grid {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  margin-bottom: 2.5rem;
  text-align: left;
}
.download-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.section-label {
  font-size: 0.75rem;
  font-weight: 950;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}
.section-btns {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
}
.download-divider {
  height: 1px;
  background: var(--border);
  opacity: 0.5;
}

.format-btn {
  background: var(--bg-input);
  border: 1px solid var(--border);
  padding: 1.25rem;
  border-radius: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.85rem;
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  font-weight: 850;
  color: var(--text-main);
}
.format-btn.mini {
  padding: 0.75rem;
  font-size: 0.9rem;
  font-family: var(--font-mono);
}
.format-btn.full-history {
  background: var(--accent);
  color: #fff;
  border: none;
  padding: 1.4rem;
  width: 100%;
  box-shadow: 0 15px 30px -5px rgba(99, 102, 241, 0.4);
}
.format-btn:hover {
  border-color: var(--accent);
  background: var(--bg-card);
  transform: translateY(-2px);
}
.format-btn.full-history:hover {
  transform: translateY(-4px);
  box-shadow: 0 20px 45px -5px rgba(99, 102, 241, 0.5);
  filter: brightness(1.1);
}

.section-hint {
  font-size: 0.85rem;
  color: var(--text-mute);
  margin: 0.5rem 0 0 0;
  line-height: 1.5;
  font-weight: 500;
}
.modal-close-btn {
  background: transparent;
  border: none;
  color: var(--text-mute);
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  transition: color 0.2s;
}
.modal-close-btn:hover {
  color: var(--text-main);
}
@media (max-width: 1024px) {
  .viewer-header {
    padding: 0 1.5rem;
  }
  .header-stats-live {
    display: none; /* Hide in individual viewer header on tablets */
  }
}

@media (max-width: 768px) {
  .viewer-header {
    height: auto;
    padding: 0.75rem 1rem;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }
  .header-right {
    width: 100%;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .log-search {
    flex: 1;
  }
  .search-input {
    width: 100%;
  }
  .log-content {
    padding: 1.5rem 1rem;
  }
  .log-line {
    gap: 1rem;
  }
  .line-num {
    min-width: 2rem;
  }
  .scroll-bottom-btn {
    bottom: 1.5rem;
    padding: 0.75rem 1.25rem;
    font-size: 0.85rem;
  }
}
</style>

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
        <div :class="['status-pulse', `status-${resourceState}`]"></div>
        <div class="name-group">
          <span class="c-name">{{ resourceName }}</span>
          <span class="c-id">{{ resourceSubtitle }}</span>
        </div>
      </div>

      <div class="header-right">
        <!-- Reconnection Badge if disconnected -->
        <div v-if="isReconnecting" class="reconnect-badge">
          <span class="pulse-dot"></span>
          <span>Reconnecting...</span>
          <button type="button" class="reconnect-now-btn" @click="connect(true)">Retry</button>
        </div>

        <!-- Live Stats in Header -->
        <div v-if="!isPod && !isService && container && container.state === 'running'" class="header-stats-live">
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
            ref="logSearchInputRef"
            type="text"
            v-model="logSearchQuery"
            placeholder="Search logs (⌘F)..."
            class="search-input"
          />
        </div>

        <div class="action-buttons">
          <button
            @click="isFullScreen = !isFullScreen"
            class="icon-btn"
            :data-tooltip="isFullScreen ? 'Exit Full Screen' : 'Full Screen'"
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
            data-tooltip="Download Logs"
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
          <button @click="clearLogs" class="icon-btn" data-tooltip="Clear View">
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
            data-tooltip="Close Viewer"
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
              <div class="modal-header">
                <h3>Download Logs</h3>
                <p class="text-mute">
                  Export logs for <strong>{{ resourceName }}</strong>
                </p>
              </div>
              
              <div class="modal-body">
                <div class="form-group mb-4">
                  <label class="form-label">Time Range</label>
                  <div class="select-container">
                    <select v-model="downloadRangeType" class="premium-input compact select-field">
                      <option value="all">All Logs</option>
                      <option value="1h">Last 1 hour</option>
                      <option value="3h">Last 3 hours</option>
                      <option value="12h">Last 12 hours</option>
                      <option value="24h">Last 24 hours</option>
                      <option value="custom">Custom Date Range</option>
                    </select>
                    <div class="select-arrow">
                      <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2.5" fill="none">
                        <polyline points="6 9 12 15 18 9"></polyline>
                      </svg>
                    </div>
                  </div>
                </div>

                <div v-if="downloadRangeType === 'custom'" class="date-range-fields animate-fade-in mt-4">
                  <div class="grid grid-cols-2 gap-4">
                    <div class="form-group">
                      <label class="form-label-sub">From</label>
                      <div class="datetime-input-container">
                        <input
                          ref="sinceInput"
                          v-model="downloadSince"
                          type="datetime-local"
                          class="premium-input compact datetime-field"
                        />
                        <button
                          type="button"
                          class="calendar-trigger-btn"
                          @click="openSincePicker"
                        >
                          <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2.5" fill="none">
                            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                            <line x1="16" y1="2" x2="16" y2="6"></line>
                            <line x1="8" y1="2" x2="8" y2="6"></line>
                            <line x1="3" y1="10" x2="21" y2="10"></line>
                          </svg>
                        </button>
                      </div>
                    </div>
                    <div class="form-group">
                      <label class="form-label-sub">To</label>
                      <div class="datetime-input-container">
                        <input
                          ref="untilInput"
                          v-model="downloadUntil"
                          type="datetime-local"
                          class="premium-input compact datetime-field"
                        />
                        <button
                          type="button"
                          class="calendar-trigger-btn"
                          @click="openUntilPicker"
                        >
                          <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2.5" fill="none">
                            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                            <line x1="16" y1="2" x2="16" y2="6"></line>
                            <line x1="8" y1="2" x2="8" y2="6"></line>
                            <line x1="3" y1="10" x2="21" y2="10"></line>
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                  <p class="text-mute range-hint mt-1">
                    Dates are parsed in your local time zone.
                  </p>
                </div>

                <div class="form-group mt-6">
                  <label class="form-label">Export Format</label>
                  <div class="format-selector">
                    <button
                      type="button"
                      :class="['format-tab', { active: downloadFormat === 'log' }]"
                      @click="downloadFormat = 'log'"
                    >
                      Raw Log (.log)
                    </button>
                    <button
                      type="button"
                      :class="['format-tab', { active: downloadFormat === 'txt' }]"
                      @click="downloadFormat = 'txt'"
                    >
                      Plain Text (.txt)
                    </button>
                    <button
                      type="button"
                      :class="['format-tab', { active: downloadFormat === 'json' }]"
                      @click="downloadFormat = 'json'"
                    >
                      JSON (.json)
                    </button>
                  </div>
                </div>
              </div>

              <div class="modal-divider"></div>

              <div class="modal-actions">
                <button
                  @click="showDownloadModal = false"
                  class="modal-btn cancel"
                >
                  Cancel
                </button>
                <button
                  @click="startLogDownload"
                  class="modal-btn confirm"
                >
                  Download
                </button>
              </div>
            </div>
          </div>
        </Transition>
      </Teleport>

      <div
        ref="logContainer"
        class="log-content"
        :class="[
          `font-${sharedState.logSettings?.fontSize || 'small'}`,
          {
            'compact-mode': sharedState.logSettings?.compactMode,
            'no-soft-wrap': sharedState.logSettings?.softWrap === false,
          }
        ]"
        @scroll="handleScroll"
      >
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

        <div
          v-for="(log, i) in displayLogs"
          :key="logLineKey(log, i)"
          class="log-line"
          :class="{
            'is-error-line': sharedState.logSettings?.highlightErrors && isErrorLog(log)
          }"
        >
          <span v-if="sharedState.logSettings?.showLineNumbers" class="line-number">
            {{ i + 1 }}
          </span>
          <span v-if="getLogTask(log)" class="task-badge" :style="getTaskBadgeStyle(getLogTask(log))">
            {{ getLogTask(log) }}
          </span>
          <span
            v-if="sharedState.logSettings?.showStreamLabels"
            class="stream-tag"
            :class="getLogStreamLabel(log).toLowerCase()"
          >
            {{ getLogStreamLabel(log) }}
          </span>
          <span v-if="sharedState.logSettings?.showTimestamps !== false" class="line-time">
            {{ formatLogLineTime(log) }}
          </span>
          <span class="line-dot" :class="getLogLevelDot(log)"></span>
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
import { sharedState } from "../utils/sharedState";
import { secureStorage } from "../utils/storage";
import { createAuthenticatedWebSocket } from "../utils/wsAuth";
import { apiFetch } from "../utils/apiFetch";

const props = defineProps({
  container: Object,
  pod: Object,
  service: Object,
  showClose: Boolean,
  compactHeader: Boolean,
});

const emit = defineEmits(["close", "stats"]);

const isPod = computed(() => !!props.pod);
const isService = computed(() => !!props.service);
const resourceName = computed(() => {
  if (isPod.value) return props.pod.name;
  if (isService.value) return props.service.name;
  return props.container?.name;
});
const resourceSubtitle = computed(() => {
  if (isPod.value) return `${props.pod.namespace} · pod`;
  if (isService.value) {
    const running = props.service.running ?? props.service.replicas ?? 0;
    const total = props.service.replicas ?? 1;
    return `${props.service.id?.substring(0, 12) || ''} · swarm service (${running}/${total} replicas)`;
  }
  return props.container?.id?.substring(0, 12) || "";
});
const resourceState = computed(() => {
  if (isPod.value) return (props.pod.phase || "unknown").toLowerCase();
  if (isService.value) {
    const running = props.service.running ?? 0;
    const total = props.service.replicas ?? 0;
    if (running > 0 && running >= total) return "running";
    if (running > 0) return "running";
    return "stopped";
  }
  return props.container?.state || "stopped";
});
const resourceKey = computed(() => {
  if (isPod.value) return `${props.pod.namespace}/${props.pod.name}`;
  if (isService.value) return `swarm:${props.service.id}`;
  return props.container?.id || "";
});
const logsBasePath = computed(() => {
  if (isPod.value) {
    return `/api/namespaces/${encodeURIComponent(props.pod.namespace)}/pods/${encodeURIComponent(props.pod.name)}/logs`;
  }
  if (isService.value) {
    return `/api/swarm/services/${props.service.id}/logs`;
  }
  return `/api/containers/${props.container.id}/logs`;
});
const logsDownloadPath = computed(() => {
  if (isPod.value) {
    return `/api/namespaces/${encodeURIComponent(props.pod.namespace)}/pods/${encodeURIComponent(props.pod.name)}/logs/download`;
  }
  if (isService.value) {
    return `/api/swarm/services/${props.service.id}/logs/download`;
  }
  return `/api/containers/${props.container.id}/logs/download`;
});
const logsCountPath = computed(() => {
  if (isPod.value) {
    return `/api/namespaces/${encodeURIComponent(props.pod.namespace)}/pods/${encodeURIComponent(props.pod.name)}/logs/count`;
  }
  if (isService.value) {
    return `/api/swarm/services/${props.service.id}/logs/count`;
  }
  return `/api/containers/${props.container.id}/logs/count`;
});
const wsPath = computed(() => {
  if (isPod.value) {
    return `/ws/pod-logs/${encodeURIComponent(props.pod.namespace)}/${encodeURIComponent(props.pod.name)}`;
  }
  if (isService.value) {
    return `/ws/swarm/services/${props.service.id}/logs`;
  }
  return `/ws/logs/${props.container.id}`;
});
const downloadFilename = computed(() => {
  if (isPod.value) return `${props.pod.name}_${props.pod.namespace}`;
  if (isService.value) return `${props.service.name || 'service'}_swarm`;
  return props.container?.name || "container";
});

const logs = ref([]);
const logSearchQuery = ref("");
const logSearchInputRef = ref(null);
const logContainer = ref(null);
const scrollSentinel = ref(null);
const autoScroll = ref(true);
const showDownloadModal = ref(false);
const downloadRangeType = ref("all");
const downloadFormat = ref("log");
const downloadSince = ref("");
const downloadUntil = ref("");
const sinceInput = ref(null);
const untilInput = ref(null);

const openSincePicker = () => {
  if (sinceInput.value && typeof sinceInput.value.showPicker === "function") {
    sinceInput.value.showPicker();
  }
};

const openUntilPicker = () => {
  if (untilInput.value && typeof untilInput.value.showPicker === "function") {
    untilInput.value.showPicker();
  }
};
const isFullScreen = ref(false);
const stats = ref({ cpu: "0.00", memory: "0B / 0B", memPercent: 0 });
const isLoadingHistory = ref(false);
const hasMoreHistory = ref(true);
const totalLogs = ref(0);
const isReconnecting = ref(false);
let lastFetchedUntil = null;
let socket = null;
let statsController = null;
let observer = null;
let reconnectTimer = null;
let reconnectAttempts = 0;
let batchBuffer = [];
let rafPending = false;
let historyFetchId = 0;
let viewerMounted = false;

const escapeHtml = (str) =>
  str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");

const LOG_TIMESTAMP_RE =
  /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z)\s?(.*)$/s;

const LOG_MONTHS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

const SWARM_TASK_COLORS = [
  "#38bdf8", // Sky blue
  "#4ade80", // Emerald green
  "#fb923c", // Warm orange
  "#a78bfa", // Purple
  "#f472b6", // Pink
  "#facc15", // Amber yellow
  "#2dd4bf", // Teal
  "#e879f9", // Fuchsia
];

const getTaskBadgeStyle = (taskLabel) => {
  if (!taskLabel) return {};
  let hash = 0;
  for (let i = 0; i < taskLabel.length; i++) {
    hash = (hash << 5) - hash + taskLabel.charCodeAt(i);
    hash |= 0;
  }
  const color = SWARM_TASK_COLORS[Math.abs(hash) % SWARM_TASK_COLORS.length];
  return {
    color: '#0f172a',
    backgroundColor: color,
    borderColor: color,
  };
};

const parseLogLine = (text) => {
  let str = String(text);
  let task = null;

  // 1. Docker Swarm service log format:
  // "adminer.1 | 2026-11-18T... log line"
  // or "service_name.slot.task_id@node_name | 2026-11-18T... log line"
  const swarmPrefixMatch = str.match(/^([a-zA-Z0-9_\-\.]+?)(?:\.[0-9a-z]{10,})?(?:@[a-zA-Z0-9_\-\.]+)?\s*\|\s*(.*)$/s);
  if (swarmPrefixMatch) {
    let rawTask = swarmPrefixMatch[1];
    const parts = rawTask.split('.');
    if (parts.length >= 2) {
      task = `${parts[0].split('_').pop()}.${parts[1]}`;
    } else {
      task = rawTask.split('_').pop();
    }
    str = swarmPrefixMatch[2];
  }

  // 2. Also handle if swarm details appear without pipe:
  // "2026-11-18T... com.docker.swarm.node.id=...,com.docker.swarm.task.id=... message"
  const detailsMatch = str.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z)\s+com\.docker\.swarm\.[^\s]+\s+(.*)$/s);
  if (detailsMatch) {
    const timestamp = new Date(detailsMatch[1]);
    const message = detailsMatch[2];
    if (!task) {
      task = props.service?.name ? `${props.service.name.split('_').pop()}.1` : "service.1";
    }
    return {
      timestamp: Number.isNaN(timestamp.getTime()) ? null : timestamp,
      message,
      task,
    };
  }

  const match = str.match(LOG_TIMESTAMP_RE);
  if (!match) {
    return { timestamp: null, message: str, task };
  }

  const timestamp = new Date(match[1]);
  if (Number.isNaN(timestamp.getTime())) {
    return { timestamp: null, message: str, task };
  }

  return { timestamp, message: match[2], task };
};

const formatLogTimestamp = (date) => {
  const pad = (value) => String(value).padStart(2, "0");
  const timeFormat = sharedState.logSettings?.timeFormat || 'auto';
  const dateFormat = sharedState.logSettings?.dateFormat || 'YYYY-MM-DD';

  const yyyy = date.getFullYear();
  const mm = pad(date.getMonth() + 1);
  const dd = pad(date.getDate());

  let datePart = `${yyyy}-${mm}-${dd}`;
  if (dateFormat === 'YYYY/MM/DD') {
    datePart = `${yyyy}/${mm}/${dd}`;
  } else if (dateFormat === 'YYYY/DD/MM') {
    datePart = `${yyyy}/${dd}/${mm}`;
  } else if (dateFormat === 'DD/MM/YYYY') {
    datePart = `${dd}/${mm}/${yyyy}`;
  } else if (dateFormat === 'DD-MM-YYYY') {
    datePart = `${dd}-${mm}-${yyyy}`;
  } else if (dateFormat === 'MM/DD/YYYY') {
    datePart = `${mm}/${dd}/${yyyy}`;
  }

  let timePart = '';
  if (timeFormat === '12-hour') {
    let hours = date.getHours();
    const ampm = hours >= 12 ? 'PM' : 'AM';
    hours = hours % 12;
    hours = hours ? hours : 12; // 0 becomes 12
    timePart = `${pad(hours)}:${pad(date.getMinutes())}:${pad(date.getSeconds())} ${ampm}`;
  } else {
    // 24-hour (default / auto)
    timePart = `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
  }

  return `${datePart} ${timePart}`;
};

const formatLogLineTime = (text) => {
  const { timestamp } = parseLogLine(text);
  return timestamp ? formatLogTimestamp(timestamp) : "";
};

const getLogStreamLabel = (text) => {
  if (/(stderr|error|err|failed|fail)/i.test(String(text))) {
    return 'STDERR';
  }
  return 'STDOUT';
};

const getLogTask = (text) => {
  const { task } = parseLogLine(text);
  return task;
};

const getLogLevelDot = (text) => {
  const clean = String(text);
  if (/(ERROR|ERR|Fail|Failed|Exception|Fatal)/i.test(clean)) return 'dot-error';
  if (/(WARN|Warning)/i.test(clean)) return 'dot-warn';
  return 'dot-normal';
};

const isErrorLog = (text) => {
  return /(ERROR|ERR|Fail|Failed|Exception|Fatal)/i.test(text);
};

const formatLog = (text) => {
  const { message } = parseLogLine(text);
  let cleanText = message;

  let formatted = escapeHtml(cleanText.replace(/\033\[[0-9;]*m/g, ""))
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

const MAX_RENDERED_LOGS = 2000;

const displayLogs = computed(() => {
  let source = logs.value;
  if (logSearchQuery.value && logSearchQuery.value.length >= 2) {
    source = logs.value.filter((l) =>
      l.toLowerCase().includes(logSearchQuery.value.toLowerCase()),
    );
  }
  if (source.length > MAX_RENDERED_LOGS) {
    return source.slice(-MAX_RENDERED_LOGS);
  }
  return source;
});

const logLineKey = (log, index) => {
  const { timestamp } = parseLogLine(log);
  if (timestamp) {
    return `${timestamp.toISOString()}-${log.length}-${log.slice(-12)}`;
  }
  return `${index}-${log.length}-${log.slice(0, 24)}-${log.slice(-12)}`;
};

const scrollToBottom = () => {
  if (logContainer.value) {
    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight;
        autoScroll.value = true;
      }
    });
  }
};

const fetchHistoricalLogs = async () => {
  if (isLoadingHistory.value || !hasMoreHistory.value) return;

  const earliestLog = logs.value[0];
  if (!earliestLog) return;

  const { timestamp: parsedTs } = parseLogLine(earliestLog);
  let until = "";
  if (parsedTs && !Number.isNaN(parsedTs.getTime())) {
    until = parsedTs.toISOString();
  } else {
    const tsMatch = earliestLog.match(
      /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z/,
    );
    if (tsMatch) {
      until = tsMatch[0];
    }
  }

  if (!until) {
    console.warn(
      "No timestamp found in earliest log:",
      earliestLog.substring(0, 50),
    );
    return;
  }

  if (until === lastFetchedUntil) return;

  const fetchId = ++historyFetchId;
  console.log("[LogViewer] Fetching history until:", until);
  lastFetchedUntil = until;
  isLoadingHistory.value = true;

  try {
    const token = secureStorage.getItem("token");
    const res = await apiFetch(
      `${logsBasePath.value}?tail=100&until=${encodeURIComponent(until)}`,
      {
        headers: { Authorization: `Bearer ${token}` },
      },
    );
    if (res.ok) {
      if (fetchId !== historyFetchId) return;

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
    if (fetchId === historyFetchId) {
      isLoadingHistory.value = false;
    }
  }
};

const handleScroll = () => {
  if (!logContainer.value) return;
  const { scrollTop, scrollHeight, clientHeight } = logContainer.value;

  const mode = sharedState.logSettings?.autoScrollMode || 'always';
  if (mode === 'always') {
    autoScroll.value = true;
    return;
  }
  if (mode === 'manual') {
    // Only user clicks on scroll to bottom can set autoScroll = true
    return;
  }

  // Smart mode (default): automatically re-engage auto-scroll when user scrolls near the bottom (within 150px)
  autoScroll.value = scrollHeight - scrollTop - clientHeight < 150;
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

const startLogDownload = async () => {
  try {
    const token = secureStorage.getItem("token");
    const params = new URLSearchParams();
    
    if (downloadRangeType.value === "custom") {
      const sinceDate = downloadSince.value ? new Date(downloadSince.value) : null;
      const untilDate = downloadUntil.value ? new Date(downloadUntil.value) : null;
      if (sinceDate && !Number.isNaN(sinceDate.getTime())) {
        params.set("since", sinceDate.toISOString());
      }
      if (untilDate && !Number.isNaN(untilDate.getTime())) {
        params.set("until", untilDate.toISOString());
      }
    } else if (downloadRangeType.value !== "all") {
      const hours = parseInt(downloadRangeType.value);
      if (!Number.isNaN(hours)) {
        const sinceDate = new Date(Date.now() - hours * 60 * 60 * 1000);
        params.set("since", sinceDate.toISOString());
      }
    }
    
    const query = params.toString() ? `?${params.toString()}` : "";
    const res = await apiFetch(
      `${logsDownloadPath.value}${query}`,
      {
        headers: { Authorization: `Bearer ${token}` },
      },
    );
    if (res.ok) {
      const text = await res.text();
      let blob;
      let filename = `${downloadFilename.value}_logs.log`;
      
      if (downloadFormat.value === "json") {
        const lines = text.split("\n").filter((l) => l.trim() !== "");
        blob = new Blob([JSON.stringify(lines, null, 2)], {
          type: "application/json",
        });
        filename = `${downloadFilename.value}_logs.json`;
      } else if (downloadFormat.value === "txt") {
        blob = new Blob([text], { type: "text/plain" });
        filename = `${downloadFilename.value}_logs.txt`;
      } else {
        blob = new Blob([text], { type: "text/plain" });
        filename = `${downloadFilename.value}_logs.log`;
      }
      
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      a.click();
      URL.revokeObjectURL(url);
      
      showDownloadModal.value = false;
      downloadSince.value = "";
      downloadUntil.value = "";
    }
  } catch (err) {
    console.error(err);
  }
};

const fetchStats = async () => {
  if (isPod.value || isService.value || !props.container?.id) return;
  if (statsController) statsController.abort();
  statsController = new AbortController();
  try {
    const token = secureStorage.getItem("token");
    const response = await apiFetch(
      `/api/containers/${props.container.id}/stats`,
      {
        headers: { Authorization: `Bearer ${token}` },
        signal: statsController.signal,
      },
    );
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() || "";
      for (const line of lines) {
        if (!line.trim()) continue;
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
          emit("stats", { 
            id: props.container.id, 
            cpu: parseFloat(cpuPercent.toFixed(2)), 
            memory: used 
          });
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
    const res = await apiFetch(logsCountPath.value, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      const data = await res.json();
      totalLogs.value = data.total;
    }
  } catch (err) {
    console.error("Failed to fetch log count:", err);
  }
};

const connect = (immediate = false) => {
  if (immediate) reconnectAttempts = 0;
  if (socket) {
    socket.onclose = null;
    socket.close();
    socket = null;
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (!wsPath.value) return;

  socket = createAuthenticatedWebSocket(wsPath.value);

  socket.onopen = () => {
    isReconnecting.value = false;
    reconnectAttempts = 0;
  };

  socket.onmessage = (e) => {
    batchBuffer.push(e.data);
    if (!rafPending) {
      rafPending = true;
      requestAnimationFrame(() => {
        if (batchBuffer.length > 0) {
          if (logs.value.length === 0) {
            logs.value.push(...batchBuffer);
          } else {
            const existingSet = new Set(logs.value);
            for (const item of batchBuffer) {
              if (!existingSet.has(item)) {
                logs.value.push(item);
                existingSet.add(item);
              }
            }
          }
          batchBuffer = [];
          if (autoScroll.value && logs.value.length > 5000) {
            logs.value.splice(0, logs.value.length - 5000);
          }
          if (autoScroll.value) {
            scrollToBottom();
          }
        }
        rafPending = false;
      });
    }
  };
  socket.onclose = () => {
    socket = null;
    if (!viewerMounted) return;
    isReconnecting.value = true;
    reconnectAttempts++;
    // Exponential backoff with jitter: 1.5s -> 3s -> 6s -> max 10s
    const delay = Math.min(10000, 1500 * Math.pow(1.5, Math.min(reconnectAttempts, 5)));
    reconnectTimer = setTimeout(() => {
      if (viewerMounted && resourceKey.value) connect();
    }, delay);
  };
};

const handleSearchShortcut = (e) => {
  if ((e.metaKey || e.ctrlKey) && (e.key === "f" || e.key === "F")) {
    if (sharedState.generalSettings?.searchIntercept !== false) {
      e.preventDefault();
      if (logSearchInputRef.value) {
        logSearchInputRef.value.focus();
        logSearchInputRef.value.select();
      }
    }
  }
};

onMounted(() => {
  viewerMounted = true;
  window.addEventListener("keydown", handleSearchShortcut);
  connect();
  fetchStats();
  fetchLogCount();
  nextTick(setupObserver);
});
onUnmounted(() => {
  viewerMounted = false;
  window.removeEventListener("keydown", handleSearchShortcut);
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (socket) {
    socket.onclose = null;
    socket.close();
  }
  if (statsController) statsController.abort();
  if (observer) observer.disconnect();
});
watch(
  () => resourceKey.value,
  () => {
    historyFetchId++;
    lastFetchedUntil = null;
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
  height: 100%;
  min-height: 0;
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
  min-height: 0;
  position: relative;
  overflow: hidden;
}

.log-content {
  height: 100%;
  overflow-y: auto;
  padding: 1.5rem 1.5rem 1.5rem 1rem;
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

.log-content.font-small {
  font-size: 0.72rem;
}

.log-content.font-medium {
  font-size: 0.85rem;
}

.log-content.font-large {
  font-size: 1rem;
}

.log-content.compact-mode .log-line {
  margin-bottom: 0.08rem;
  line-height: 1.35;
}

.log-content.no-soft-wrap .line-text {
  white-space: pre !important;
  word-break: normal !important;
}

.log-content.no-soft-wrap {
  overflow-x: auto;
}

.log-line {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 0.25rem;
  padding: 0.15rem 0.5rem;
  border-left: 3px solid transparent;
  border-radius: 2px;
  content-visibility: auto;
  contain-intrinsic-size: 0 24px;
}

.log-line.is-error-line {
  background: rgba(239, 68, 68, 0.12);
  border-left-color: #ef4444;
}

.line-number {
  font-family: var(--font-mono);
  font-size: 0.68rem;
  color: #64748b;
  min-width: 2.2rem;
  text-align: right;
  user-select: none;
  flex-shrink: 0;
  opacity: 0.7;
}

.task-badge {
  font-size: 0.62rem;
  font-weight: 800;
  font-family: var(--font-mono);
  padding: 0.12rem 0.5rem;
  border-radius: 999px;
  flex-shrink: 0;
  display: inline-block;
  line-height: 1.3;
  letter-spacing: 0.02em;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.reconnect-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.2rem 0.6rem;
  border-radius: 6px;
  background: rgba(245, 158, 11, 0.15);
  border: 1px solid rgba(245, 158, 11, 0.35);
  color: #f59e0b;
  font-size: 0.72rem;
  font-weight: 700;
}

.reconnect-now-btn {
  background: rgba(245, 158, 11, 0.25);
  border: none;
  border-radius: 4px;
  color: #fbbf24;
  padding: 0.1rem 0.35rem;
  font-size: 0.68rem;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.15s ease;
}

.reconnect-now-btn:hover {
  background: rgba(245, 158, 11, 0.4);
  color: #ffffff;
}

.pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #f59e0b;
  animation: pulse 1.5s infinite;
}

.stream-tag {
  font-family: var(--font-mono);
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  padding: 0.08rem 0.35rem;
  border-radius: 4px;
  flex-shrink: 0;
  user-select: none;
  line-height: 1.25;
}

.stream-tag.stdout {
  color: #38bdf8;
  background: rgba(56, 189, 248, 0.12);
  border: 1px solid rgba(56, 189, 248, 0.25);
}

.stream-tag.stderr {
  color: #f87171;
  background: rgba(239, 68, 68, 0.14);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.log-content.thin-scrollbars::-webkit-scrollbar {
  width: 5px;
  height: 5px;
}

.log-content.thin-scrollbars::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.3);
  border-radius: 4px;
}

.line-time {
  color: #64748b;
  width: auto;
  min-width: 9.5rem;
  flex-shrink: 0;
  user-select: none;
  font-size: 0.72rem;
  white-space: nowrap;
  font-family: var(--font-mono);
}

.line-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 0.42rem;
}

.line-dot.dot-normal {
  background: #10b981;
}

.line-dot.dot-warn {
  background: #f59e0b;
}

.line-dot.dot-error {
  background: #ef4444;
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
  background: var(--accent-soft);
  border: 1px solid rgba(var(--accent-rgb), 0.35);
  border-radius: 8px;
  color: var(--text-main);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.history-btn-manual:hover {
  background: var(--accent);
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
  border-top-color: var(--accent);
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
  .line-time {
    width: 9.5rem;
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
  .line-time {
    width: 8.5rem;
    font-size: 0.55rem;
  }
  .line-text {
    font-size: 0.65rem;
  }
  .log-line {
    gap: 0.5rem;
  }
}

/* Premium modal form elements styling */
.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  text-align: left;
}
.form-label {
  font-size: 0.72rem;
  font-weight: 850;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-mute);
}
.form-label-sub {
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--text-dim);
}
.range-selector, .format-selector {
  display: flex;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 0.25rem;
  gap: 0.25rem;
}
.range-tab, .format-tab {
  flex: 1;
  padding: 0.55rem;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: var(--text-mute);
  font-size: 0.78rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s ease;
}
.range-tab:hover, .format-tab:hover {
  color: var(--text-main);
}
.range-tab.active, .format-tab.active {
  background: var(--bg-card);
  color: var(--text-main);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  border: 1px solid var(--border);
}
.premium-input.compact {
  padding: 0.55rem 0.85rem;
  font-size: 0.8rem;
  border-radius: 10px;
}
.grid {
  display: grid;
}
.grid-cols-2 {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.gap-4 {
  gap: 1rem;
}
.mb-4 {
  margin-bottom: 1rem;
}
.range-hint {
  font-size: 0.68rem;
  opacity: 0.8;
}

/* Custom dropdown and calendar-only click controls styling */
.select-container {
  position: relative;
  width: 100%;
}
.select-field {
  appearance: none;
  -webkit-appearance: none;
  width: 100%;
  padding-right: 2.5rem !important;
  cursor: pointer;
}
.select-arrow {
  position: absolute;
  right: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-mute);
  pointer-events: none;
  display: flex;
  align-items: center;
}
.datetime-input-container {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}
.datetime-field {
  padding-right: 2.25rem !important;
}
/* Disable opening picker by clicking text area, styling only indicator */
.datetime-field::-webkit-calendar-picker-indicator {
  position: absolute;
  top: 0;
  right: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: default;
  pointer-events: none;
}
.calendar-trigger-btn {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: var(--text-mute);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  border-radius: 6px;
  transition: all 0.2s ease;
}
.calendar-trigger-btn:hover {
  color: var(--accent);
  background: var(--bg-subtle);
}
</style>

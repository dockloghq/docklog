<template>
  <div class="health-view-wrap animate-fade-in">
    <!-- Top Header Banner Card -->
    <header class="health-top-card">
      <div class="header-main-info">
        <h1 class="header-title">System Health</h1>
        <p class="header-subtitle">
          Historical CPU and memory utilization
          <span v-if="isPartialData" class="coverage-gold">
            · {{ formatDuration(availableHours) }} of data
          </span>
        </p>
      </div>

      <div class="header-time-filters">
        <div class="filter-pill-group">
          <button
            v-for="f in filters"
            :key="f.label"
            type="button"
            class="filter-pill-item"
            :class="{
              active: activeFilter === f.value,
              'is-partial': availableHours < f.value,
            }"
            @click="activeFilter = f.value"
            :data-tooltip="availableHours < f.value ? `${f.note} (partial data)` : f.note"
          >
            {{ f.short }}
          </button>
          <button
            type="button"
            class="filter-pill-item"
            :class="{ active: activeFilter === 'custom' }"
            @click="showCustomModal = true"
            data-tooltip="Select specific dates"
          >
            Custom
          </button>
        </div>
      </div>
    </header>

    <!-- 4 Metric Cards Row -->
    <section class="health-metrics-row">
      <!-- 1. Avg CPU Load -->
      <div class="health-stat-card">
        <div class="card-left-icon icon-chip">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="4" y="4" width="16" height="16" rx="2"></rect>
            <rect x="9" y="9" width="6" height="6"></rect>
            <line x1="9" y1="1" x2="9" y2="4"></line>
            <line x1="15" y1="1" x2="15" y2="4"></line>
            <line x1="9" y1="20" x2="9" y2="23"></line>
            <line x1="15" y1="20" x2="15" y2="23"></line>
            <line x1="20" y1="9" x2="23" y2="9"></line>
            <line x1="20" y1="14" x2="23" y2="14"></line>
            <line x1="1" y1="9" x2="4" y2="9"></line>
            <line x1="1" y1="14" x2="4" y2="14"></line>
          </svg>
        </div>
        <div class="card-middle-content">
          <span class="card-label">Avg CPU Load</span>
          <span class="card-number">{{ avgCpu }}%</span>
        </div>
        <div class="card-right-tag tag-avg">AVG</div>
      </div>

      <!-- 2. Peak CPU Spike -->
      <div class="health-stat-card">
        <div class="card-left-icon icon-bell">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
            <path d="M12 2C10.34 2 9 3.34 9 5c0 .35.07.69.18 1C6.28 6.95 4 9.77 4 13v4l-2 2v1h20v-1l-2-2v-4c0-3.23-2.28-6.05-5.18-7 .11-.31.18-.65.18-1 0-1.66-1.34-3-3-3zm0 20a3 3 0 0 0 3-3H9a3 3 0 0 0 3 3z"/>
          </svg>
        </div>
        <div class="card-middle-content">
          <span class="card-label">Peak CPU Spike</span>
          <span class="card-number">{{ maxCpu }}%</span>
          <span v-if="peakCpuTime" class="card-sub-info">at {{ peakCpuTime }}</span>
        </div>
        <div class="card-right-tag tag-peak">PEAK</div>
      </div>

      <!-- 3. Avg Memory -->
      <div class="health-stat-card">
        <div class="card-left-icon icon-ram">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5">
            <rect x="3" y="4" width="18" height="7" rx="2"></rect>
            <rect x="3" y="13" width="18" height="7" rx="2"></rect>
          </svg>
        </div>
        <div class="card-middle-content">
          <span class="card-label">Avg Memory</span>
          <span class="card-number">{{ avgMem }} GB</span>
        </div>
        <div class="card-right-tag tag-mem">MEM</div>
      </div>

      <!-- 4. Peak Memory -->
      <div class="health-stat-card">
        <div class="card-left-icon icon-layers">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
            <polyline points="2 17 12 22 22 17"></polyline>
            <polyline points="2 12 12 17 22 12"></polyline>
          </svg>
        </div>
        <div class="card-middle-content">
          <span class="card-label">Peak Memory</span>
          <span class="card-number">{{ maxMem }} GB</span>
          <span v-if="peakMemTime" class="card-sub-info">at {{ peakMemTime }}</span>
        </div>
        <div class="card-right-tag tag-max">MAX</div>
      </div>
    </section>

    <!-- Side-by-Side Charts -->
    <section class="health-charts-grid">
      <!-- CPU Chart Card -->
      <div class="chart-card">
        <div class="chart-card-top">
          <h2 class="chart-title">CPU Utilization</h2>
          <span class="chart-unit-pill">% Load</span>
        </div>
        <div class="chart-inner-wrap">
          <Line
            v-if="chartData.cpu.labels.length"
            :data="chartData.cpu"
            :options="cpuChartOptions"
          />
          <div v-else class="chart-empty-pulse"></div>
        </div>
      </div>

      <!-- Memory Chart Card -->
      <div class="chart-card">
        <div class="chart-card-top">
          <h2 class="chart-title">Memory Consumption</h2>
          <span class="chart-unit-pill">Gigabytes</span>
        </div>
        <div class="chart-inner-wrap">
          <Line
            v-if="chartData.mem.labels.length"
            :data="chartData.mem"
            :options="memChartOptions"
          />
          <div v-else class="chart-empty-pulse"></div>
        </div>
      </div>
    </section>

    <!-- Custom Range Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showCustomModal" class="modal-overlay">
          <div class="modal-content shadow-2xl">
            <div class="modal-header">
              <h3>Custom Range</h3>
            </div>
            <div class="modal-body">
              <div class="input-group">
                <label>Start Date</label>
                <input type="date" v-model="tempStart" :max="today" class="premium-input" />
              </div>
              <div class="input-group">
                <label>End Date</label>
                <input type="date" v-model="tempEnd" :max="today" class="premium-input" />
              </div>
              <p v-if="modalError" class="error-text">{{ modalError }}</p>
            </div>
            <div class="modal-actions">
              <button @click="showCustomModal = false" class="modal-btn cancel">Cancel</button>
              <button @click="applyCustomRange" class="modal-btn confirm">Apply Range</button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import AppIcon from "../components/AppIcon.vue";
import { ref, onMounted, onUnmounted, watch, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Line } from "vue-chartjs";
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  CategoryScale,
  Filler,
} from "chart.js";
import { secureStorage } from "../utils/storage";
import { sharedState } from "../utils/sharedState";
import { apiFetch } from "../utils/apiFetch";

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  CategoryScale,
  Filler,
);

const isDark = computed(() => sharedState.theme === 'dark');
const route = useRoute();
const router = useRouter();

const filters = [
  { label: "1H", short: "1H", note: "Last hour", value: 1 },
  { label: "6H", short: "6H", note: "Last 6 hours", value: 6 },
  { label: "12H", short: "12H", note: "Last 12 hours", value: 12 },
  { label: "24H", short: "24H", note: "Last 24 hours", value: 24 },
];

const activeFilter = ref(24);
const customStart = ref("");
const customEnd = ref("");
const tempStart = ref("");
const tempEnd = ref("");
const showCustomModal = ref(false);
const modalError = ref("");
const today = new Date().toISOString().split("T")[0];

const history = ref([]);

const parseTime = (ts) => {
  if (!ts) return new Date();
  if (typeof ts === "string" && !ts.endsWith("Z") && !ts.includes("+")) {
    return new Date(ts.replace(" ", "T") + "Z");
  }
  return new Date(ts);
};

const formatTimeLabel = (date) => {
  const now = new Date();
  const isToday = date.toDateString() === now.toDateString();
  return isToday
    ? date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    : `${date.toLocaleDateString([], { month: "short", day: "numeric" })} ${date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;
};

const availableHours = computed(() => {
  if (history.value.length === 0) return 0;
  const timestamps = history.value.map((h) => parseTime(h.timestamp).getTime());
  const oldest = Math.min(...timestamps);
  const now = new Date().getTime();
  return Math.max(0, (now - oldest) / (1000 * 60 * 60));
});

const isPartialData = computed(() => {
  if (activeFilter.value === 'custom') return false;
  // If we have less than 95% of the requested range, call it partial
  return availableHours.value < (activeFilter.value * 0.95);
});

const formatDuration = (hours) => {
  const totalMinutes = Math.round(hours * 60);
  if (totalMinutes < 60) return `${totalMinutes}m`;
  
  const h = Math.floor(totalMinutes / 60);
  const m = totalMinutes % 60;
  
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
};
const chartData = ref({
  cpu: { labels: [], datasets: [] },
  mem: { labels: [], datasets: [] },
});

const activeHistory = computed(() => {
  let start, end;
  if (activeFilter.value === "custom") {
    if (!customStart.value || !customEnd.value) return history.value;
    start = new Date(customStart.value);
    end = new Date(customEnd.value);
    end.setHours(23, 59, 59, 999);
  } else {
    end = new Date();
    start = new Date(end.getTime() - activeFilter.value * 60 * 60 * 1000);
  }
  return history.value.filter((h) => {
    const t = parseTime(h.timestamp);
    return t >= start && t <= end;
  });
});

const activeRange = computed(() => {
  if (activeFilter.value === "custom" && customStart.value && customEnd.value) {
    const start = new Date(customStart.value);
    const end = new Date(customEnd.value);
    end.setHours(23, 59, 59, 999);
    return { start, end };
  }
  const end = new Date();
  const start = new Date(end.getTime() - activeFilter.value * 60 * 60 * 1000);
  return { start, end };
});

const avgCpu = computed(() => {
  if (!activeHistory.value.length) return 0;
  return (activeHistory.value.reduce((acc, h) => acc + h.cpu, 0) / activeHistory.value.length).toFixed(1);
});

const peakCpuObj = computed(() => {
  if (!activeHistory.value.length) return null;
  return activeHistory.value.reduce((max, h) => (h.cpu > (max ? max.cpu : -1) ? h : max), null);
});

const maxCpu = computed(() => {
  return peakCpuObj.value ? peakCpuObj.value.cpu.toFixed(1) : "0.0";
});

const peakCpuTime = computed(() => {
  return peakCpuObj.value ? formatTimeLabel(parseTime(peakCpuObj.value.timestamp)) : "";
});

const avgMem = computed(() => {
  if (!activeHistory.value.length) return 0;
  return (activeHistory.value.reduce((acc, h) => acc + h.memory, 0) / (activeHistory.value.length * 1024 * 1024 * 1024)).toFixed(2);
});

const peakMemObj = computed(() => {
  if (!activeHistory.value.length) return null;
  return activeHistory.value.reduce((max, h) => (h.memory > (max ? max.memory : -1) ? h : max), null);
});

const maxMem = computed(() => {
  return peakMemObj.value ? (peakMemObj.value.memory / (1024 * 1024 * 1024)).toFixed(2) : "0.00";
});

const peakMemTime = computed(() => {
  return peakMemObj.value ? formatTimeLabel(parseTime(peakMemObj.value.timestamp)) : "";
});

const applyCustomRange = () => {
  if (!tempStart.value || !tempEnd.value) {
    modalError.value = "Select both dates";
    return;
  }
  customStart.value = tempStart.value;
  customEnd.value = tempEnd.value;
  activeFilter.value = "custom";
  showCustomModal.value = false;
  updateUrl();
};

const syncStateFromUrl = () => {
  const { range, start, end } = route.query;
  if (range) {
    activeFilter.value = range === "custom" ? "custom" : parseInt(range);
  }
  if (start) customStart.value = start;
  if (end) customEnd.value = end;
};

const updateUrl = () => {
  const query = { ...route.query };
  query.range = activeFilter.value;
  if (activeFilter.value === "custom") {
    query.start = customStart.value;
    query.end = customEnd.value;
  } else {
    delete query.start;
    delete query.end;
  }
  router.replace({ query });
};

const makeChartOptions = (unit) => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: { duration: 1000, easing: "easeOutQuart" },
  interaction: { mode: "index", intersect: false },
  scales: {
    y: {
      beginAtZero: true,
      grid: { color: isDark.value ? "rgba(255, 255, 255, 0.03)" : "rgba(0, 0, 0, 0.03)" },
      ticks: {
        color: isDark.value ? "rgba(255, 255, 255, 0.4)" : "rgba(0, 0, 0, 0.4)",
        font: { size: 10, family: "JetBrains Mono" },
      },
    },
    x: {
      grid: { display: false },
      ticks: {
        color: isDark.value ? "rgba(255, 255, 255, 0.3)" : "rgba(0, 0, 0, 0.3)",
        font: { size: 9, family: "JetBrains Mono" },
        maxRotation: 0,
        maxTicksLimit: 8,
      },
    },
  },
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: isDark.value ? "#0f172a" : "#ffffff",
      titleColor: isDark.value ? "rgba(255,255,255,0.5)" : "rgba(0,0,0,0.5)",
      bodyColor: isDark.value ? "#fff" : "#000",
      borderColor: "var(--border)",
      borderWidth: 1,
      padding: 12,
      cornerRadius: 12,
      displayColors: false,
      callbacks: {
        label: (item) => `${item.formattedValue} ${unit}`,
      },
    },
  },
  elements: {
    point: { radius: 0, hoverRadius: 6 },
    line: { tension: 0.4 },
  },
});

const cpuChartOptions = computed(() => {
  const opts = makeChartOptions("%");
  const values = chartData.value.cpu.datasets[0]?.data?.filter((v) => v != null) ?? [];
  const dataMax = values.length ? Math.max(...values) : 0;
  if (dataMax > 0) {
    opts.scales.y.suggestedMax = Math.min(100, Math.ceil(dataMax * 1.15));
  }
  return opts;
});
const memChartOptions = computed(() => {
  const opts = makeChartOptions("GB");
  const values = chartData.value.mem.datasets[0]?.data?.filter((v) => v != null) ?? [];
  const dataMax = values.length ? Math.max(...values) : 0;
  if (dataMax > 0) {
    opts.scales.y.suggestedMax = Math.ceil(dataMax * 1.15 * 100) / 100;
  }
  return opts;
});

watch(
  () => sharedState.theme,
  () => {
    updateCharts();
  },
);

const fetchData = async () => {
  try {
    const token = secureStorage.getItem("token");
    const res = await apiFetch("/api/system/history", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      const data = await res.json();
      // Ensure history is sorted ascending (oldest first) for easier processing
      history.value = data.sort((a, b) => parseTime(a.timestamp) - parseTime(b.timestamp));

      // If URL does not have explicit range parameter and not in custom mode,
      // auto-select the best fitting filter:
      if (!route.query.range && activeFilter.value !== 'custom') {
        const hours = availableHours.value;
        if (hours >= 20) {
          activeFilter.value = 24;
        } else if (hours >= 10) {
          activeFilter.value = 12;
        } else if (hours >= 4) {
          activeFilter.value = 6;
        } else {
          // If available data is less than ~4h (e.g. only 2m or 30m), auto-select 1H
          activeFilter.value = 1;
        }
      }

      updateCharts();
    }
  } catch (err) {
    console.error(err);
  }
};

const updateCharts = () => {
  const { start: startTime, end: endTime } = activeRange.value;
  const rangeHours = Math.max((endTime.getTime() - startTime.getTime()) / (1000 * 60 * 60), 1 / 60);
  const now = new Date();

  // 1. Generate fixed timeline bins (e.g., 60 bins for any range)
  const binCount = 60;
  const binSizeMs = (rangeHours * 60 * 60 * 1000) / binCount;
  const timeline = [];

  for (let i = 0; i <= binCount; i++) {
    const t = new Date(startTime.getTime() + i * binSizeMs);
    const isToday = t.toDateString() === now.toDateString();
    const label = isToday
      ? t.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
      : `${t.toLocaleDateString([], { month: "short", day: "numeric" })} ${t.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;

    timeline.push({
      time: t,
      label,
      cpu: null,
      mem: null,
    });
  }

  // 2. Map filtered history to bins (peak CPU + peak memory per bin, matching summary cards)
  activeHistory.value.forEach((h) => {
    const hTime = parseTime(h.timestamp);
    const binIndex = Math.floor((hTime.getTime() - startTime.getTime()) / binSizeMs);
    if (binIndex >= 0 && binIndex <= binCount) {
      const bin = timeline[binIndex];
      bin.cpu = bin.cpu === null ? h.cpu : Math.max(bin.cpu, h.cpu);
      bin.mem = bin.mem === null ? h.memory : Math.max(bin.mem, h.memory);
    }
  });

  const labels = timeline.map((t) => t.label);

  chartData.value.cpu = {
    labels,
    datasets: [
      {
        label: "CPU Load",
        data: timeline.map((t) => t.cpu),
        borderColor: "#38bdf8",
        backgroundColor: "rgba(56, 189, 248, 0.08)",
        fill: true,
        borderWidth: 2.5,
        spanGaps: true,
      },
    ],
  };

  chartData.value.mem = {
    labels,
    datasets: [
      {
        label: "Memory Usage",
        data: timeline.map((t) => (t.mem != null ? t.mem / (1024 * 1024 * 1024) : null)),
        borderColor: "#22c55e",
        backgroundColor: "rgba(34, 197, 94, 0.12)",
        fill: true,
        borderWidth: 2.5,
        spanGaps: true,
      },
    ],
  };
};

watch([activeFilter, customStart, customEnd], () => {
  updateCharts();
  updateUrl();
});

onMounted(() => {
  syncStateFromUrl();
  fetchData();
});
</script>

<style scoped>
.health-view-wrap {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.25rem 1.5rem 3rem;
  max-width: 1440px;
  margin: 0 auto;
  color: #f1f5f9;
}

/* 1. Header Card */
.health-top-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  flex-wrap: wrap;
  padding: 1.35rem 1.75rem;
  background: #0d121c;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 14px;
}

[data-theme="light"] .health-top-card {
  background: #ffffff;
  border-color: #e2e8f0;
}

.header-main-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.header-title {
  margin: 0;
  font-size: 1.45rem;
  font-weight: 800;
  color: #f8fafc;
  letter-spacing: -0.02em;
}

[data-theme="light"] .header-title {
  color: #0f172a;
}

.header-subtitle {
  margin: 0;
  font-size: 0.86rem;
  color: #94a3b8;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.coverage-gold {
  color: #f59e0b;
  font-weight: 700;
}

/* Time range filter buttons */
.filter-pill-group {
  display: inline-flex;
  align-items: center;
  background: #080c14;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 3px;
  gap: 2px;
}

[data-theme="light"] .filter-pill-group {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.filter-pill-item {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 0.8rem;
  font-weight: 700;
  padding: 0.4rem 0.9rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.filter-pill-item:hover {
  color: #f8fafc;
}

[data-theme="light"] .filter-pill-item:hover {
  color: #0f172a;
}

.filter-pill-item.active {
  background: #0284c7;
  color: #ffffff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}

.filter-pill-item.is-partial {
  opacity: 0.85;
}

/* 2. 4 Health Stat Cards */
.health-metrics-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

.health-stat-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.15rem 1.25rem;
  background: #0d121c;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 14px;
  gap: 0.85rem;
  transition: transform 0.15s ease, border-color 0.15s ease;
}

[data-theme="light"] .health-stat-card {
  background: #ffffff;
  border-color: #e2e8f0;
}

.health-stat-card:hover {
  border-color: rgba(255, 255, 255, 0.16);
  transform: translateY(-1px);
}

.card-left-icon {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon-chip {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
}

.icon-bell {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.icon-ram {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.icon-layers {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.card-middle-content {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  flex: 1;
  min-width: 0;
}

.card-label {
  font-size: 0.72rem;
  font-weight: 700;
  color: #94a3b8;
  letter-spacing: 0.01em;
  white-space: nowrap;
}

.card-sub-info {
  font-size: 0.68rem;
  font-weight: 600;
  color: var(--accent, #38bdf8);
  font-family: var(--font-mono, monospace);
  letter-spacing: -0.01em;
  margin-top: 1px;
}

.card-number {
  font-size: 1.55rem;
  font-weight: 800;
  color: #f8fafc;
  letter-spacing: -0.02em;
  line-height: 1.1;
  font-family: var(--font-mono, monospace);
}

[data-theme="light"] .card-number {
  color: #0f172a;
}

.card-right-tag {
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  padding: 0.2rem 0.55rem;
  border-radius: 6px;
  align-self: flex-start;
}

.tag-avg {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
}

.tag-peak {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.tag-mem {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.tag-max {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

/* 3. Charts Grid */
.health-charts-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
}

.chart-card {
  padding: 1.25rem 1.5rem;
  background: #0d121c;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 14px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

[data-theme="light"] .chart-card {
  background: #ffffff;
  border-color: #e2e8f0;
}

.chart-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.chart-title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 800;
  color: #f8fafc;
}

[data-theme="light"] .chart-title {
  color: #0f172a;
}

.chart-unit-pill {
  font-size: 0.68rem;
  font-weight: 800;
  color: #94a3b8;
  background: #080c14;
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 0.2rem 0.65rem;
  border-radius: 6px;
  letter-spacing: 0.03em;
}

[data-theme="light"] .chart-unit-pill {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.chart-inner-wrap {
  height: 330px;
  position: relative;
}

.chart-empty-pulse {
  height: 100%;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.02);
  animation: pulse 1.5s infinite;
}

/* Responsive */
@media (max-width: 1200px) {
  .health-metrics-row {
    grid-template-columns: repeat(2, 1fr);
  }
  .health-charts-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .health-metrics-row {
    grid-template-columns: 1fr;
  }
  .chart-inner-wrap {
    height: 250px;
  }
  .health-top-card {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

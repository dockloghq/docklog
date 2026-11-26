import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { secureStorage } from '../utils/storage';
import { sharedState, showToast, formatBytes } from '../utils/sharedState';
import { apiFetch } from '../utils/apiFetch';
import { logsRouteForContainer } from '../utils/logRoutes';
import { createAuthenticatedWebSocket } from '../utils/wsAuth';

const containers = ref([]);
const loading = ref(true);
let eventsWs = null;
let reconnectTimer = null;
let refreshDebounceTimer = null;
let fallbackInterval = null;
let subscriberCount = 0;

const activeLiveId = ref(null);
const liveStats = ref({ cpu: 0, memory: 0 });
let liveStatsWs = null;

export async function fetchContainers() {
  try {
    const token = secureStorage.getItem('token');
    const res = await apiFetch('/api/containers', {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      const data = await res.json();
      containers.value = Array.isArray(data) ? data : [];
    }
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
}

function scheduleRefresh(delay = 300) {
  if (refreshDebounceTimer) clearTimeout(refreshDebounceTimer);
  refreshDebounceTimer = setTimeout(() => {
    fetchContainers();
  }, delay);
}

function connectEventsWebSocket() {
  if (eventsWs) return;
  try {
    eventsWs = createAuthenticatedWebSocket('/ws/events');
    eventsWs.onopen = () => {
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
    };
    eventsWs.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        // If event is container lifecycle (start, die, stop, destroy, pause, unpause, restart)
        if (msg.Type === 'container' || msg.status || msg.Action) {
          scheduleRefresh(300);
        }
      } catch (e) {
        scheduleRefresh(500);
      }
    };
    eventsWs.onclose = () => {
      eventsWs = null;
      if (subscriberCount > 0 && !reconnectTimer) {
        reconnectTimer = setTimeout(connectEventsWebSocket, 5000);
      }
    };
    eventsWs.onerror = () => {
      if (eventsWs) {
        eventsWs.close();
      }
    };
  } catch (err) {
    eventsWs = null;
  }
}

function startRealtimeSync() {
  fetchContainers();
  connectEventsWebSocket();
}

function stopRealtimeSync() {
  if (eventsWs) {
    eventsWs.onclose = null;
    eventsWs.close();
    eventsWs = null;
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (refreshDebounceTimer) {
    clearTimeout(refreshDebounceTimer);
    refreshDebounceTimer = null;
  }
  if (fallbackInterval) {
    clearInterval(fallbackInterval);
    fallbackInterval = null;
  }
}

export function formatContainerDate(unix) {
  if (!unix) return 'N/A';
  return new Date(unix * 1000).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function useContainers(options = {}) {
  const { autoPoll = true } = options;
  const router = useRouter();

  const showConfirm = ref(false);
  const pendingId = ref(null);
  const pendingAction = ref('');
  const actionPending = ref(false);

  const actionClass = computed(() => {
    if (pendingAction.value === 'start') return 'success';
    if (pendingAction.value === 'restart') return 'warning';
    if (pendingAction.value === 'stop' || pendingAction.value === 'remove') return 'error';
    return '';
  });

  const standaloneContainers = computed(() =>
    containers.value.filter((c) => !c.is_swarm),
  );

  const standaloneCount = computed(() => standaloneContainers.value.length);

  const filteredContainers = computed(() =>
    standaloneContainers.value.filter(
      (c) =>
        c.name.toLowerCase().includes(sharedState.searchQuery.toLowerCase()) ||
        c.image.toLowerCase().includes(sharedState.searchQuery.toLowerCase()),
    ),
  );

  const runningCount = computed(
    () => standaloneContainers.value.filter((c) => c.state === 'running').length,
  );

  const stoppedCount = computed(
    () => standaloneContainers.value.filter((c) => c.state !== 'running').length,
  );

  const startLiveStats = (id) => {
    activeLiveId.value = id;
    stopLiveStats();
    activeLiveId.value = id;
    try {
      liveStatsWs = createAuthenticatedWebSocket(`/ws/containers/${id}/stats`);
      liveStatsWs.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          liveStats.value = { cpu: data.cpu, memory: data.memory };
        } catch (e) {}
      };
      liveStatsWs.onerror = () => {
        stopLiveStats();
      };
    } catch (e) {
      liveStatsWs = null;
    }
  };

  const stopLiveStats = () => {
    activeLiveId.value = null;
    if (liveStatsWs) {
      liveStatsWs.onclose = null;
      liveStatsWs.close();
      liveStatsWs = null;
    }
  };

  const goToLogs = (id) => {
    router.push(logsRouteForContainer(id));
  };

  const goToShell = (id) => {
    router.push({ path: '/shell', query: { c: id } });
  };

  const goToDetail = (id) => {
    router.push({ path: `/containers/${id}` });
  };

  const triggerConfirm = (id, action) => {
    if (actionPending.value) return;
    pendingId.value = id;
    pendingAction.value = action;
    showConfirm.value = true;
  };

  const closeConfirm = () => {
    if (actionPending.value) return;
    showConfirm.value = false;
  };

  const executeAction = async () => {
    if (!pendingId.value || !pendingAction.value || actionPending.value) return;

    const id = pendingId.value;
    const action = pendingAction.value;
    actionPending.value = true;

    try {
      const token = secureStorage.getItem('token');
      const formData = new FormData();
      formData.append('action', action);
      const res = await apiFetch(`/api/containers/${id}/action`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: formData,
      });
      if (res.ok) {
        showToast('Success', `${action} completed for container.`, 'success');
        await fetchContainers();
      } else {
        let message = 'Action failed.';
        try {
          const body = await res.json();
          if (body?.error) message = body.error;
        } catch {
          // ignore parse errors
        }
        showToast('Error', message, 'error');
      }
    } catch (err) {
      console.error(err);
      showToast('Error', 'Action failed. Check your connection and try again.', 'error');
    } finally {
      actionPending.value = false;
      showConfirm.value = false;
      pendingId.value = null;
      pendingAction.value = '';
    }
  };

  onMounted(() => {
    if (!autoPoll) return;
    subscriberCount += 1;
    if (subscriberCount === 1) startRealtimeSync();
  });

  onUnmounted(() => {
    if (!autoPoll) return;
    subscriberCount -= 1;
    if (subscriberCount === 0) stopRealtimeSync();
    stopLiveStats();
  });

  return {
    containers,
    standaloneContainers,
    standaloneCount,
    loading,
    filteredContainers,
    runningCount,
    stoppedCount,
    activeLiveId,
    liveStats,
    showConfirm,
    pendingAction,
    actionPending,
    actionClass,
    fetchContainers,
    startLiveStats,
    stopLiveStats,
    goToLogs,
    goToShell,
    goToDetail,
    triggerConfirm,
    closeConfirm,
    executeAction,
    formatBytes,
    formatDate: formatContainerDate,
  };
}

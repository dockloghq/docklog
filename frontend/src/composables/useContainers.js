import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { secureStorage } from '../utils/storage';
import { sharedState, showToast, formatBytes } from '../utils/sharedState';
import { apiFetch } from '../utils/apiFetch';
import { logsRouteForContainer } from '../utils/logRoutes';

const containers = ref([]);
const loading = ref(true);
let pollInterval = null;
let pollSubscribers = 0;

const activeLiveId = ref(null);
const liveStats = ref({ cpu: 0, memory: 0 });
let liveInterval = null;

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

function startPolling() {
  if (pollInterval) return;
  fetchContainers();
  pollInterval = setInterval(() => {
    // Only fetch if tab is active and visible
    if (typeof document !== 'undefined' && document.hidden) return;
    fetchContainers();
  }, 3000);
}

function stopPolling() {
  if (pollInterval) {
    clearInterval(pollInterval);
    pollInterval = null;
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

  const fetchStatsNow = async (id) => {
    try {
      const token = secureStorage.getItem('token');
      const res = await apiFetch(`/api/containers/${id}/stats-now`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const data = await res.json();
        liveStats.value = { cpu: data.cpu, memory: data.memory };
      }
    } catch (err) {
      console.error('Live stats fetch failed', err);
    }
  };

  const startLiveStats = (id) => {
    activeLiveId.value = id;
    fetchStatsNow(id);
    if (liveInterval) clearInterval(liveInterval);
    liveInterval = setInterval(() => fetchStatsNow(id), 2000);
  };

  const stopLiveStats = () => {
    activeLiveId.value = null;
    if (liveInterval) {
      clearInterval(liveInterval);
      liveInterval = null;
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
    pollSubscribers += 1;
    if (pollSubscribers === 1) startPolling();
  });

  onUnmounted(() => {
    if (!autoPoll) return;
    pollSubscribers -= 1;
    if (pollSubscribers === 0) stopPolling();
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

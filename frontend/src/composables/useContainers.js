import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { secureStorage } from '../utils/storage';
import { sharedState, showToast, formatBytes } from '../utils/sharedState';
import { apiFetch } from '../utils/apiFetch';

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
      containers.value = await res.json();
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
  pollInterval = setInterval(fetchContainers, 3000);
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

  const actionClass = computed(() => {
    if (pendingAction.value === 'start') return 'success';
    if (pendingAction.value === 'restart') return 'warning';
    if (pendingAction.value === 'stop' || pendingAction.value === 'remove') return 'error';
    return '';
  });

  const filteredContainers = computed(() =>
    containers.value.filter(
      (c) =>
        c.name.toLowerCase().includes(sharedState.searchQuery.toLowerCase()) ||
        c.image.toLowerCase().includes(sharedState.searchQuery.toLowerCase()),
    ),
  );

  const runningCount = computed(
    () => containers.value.filter((c) => c.state === 'running').length,
  );

  const stoppedCount = computed(
    () => containers.value.filter((c) => c.state !== 'running').length,
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
    router.push({ path: '/logs', query: { c: id } });
  };

  const goToShell = (id) => {
    router.push({ path: '/shell', query: { c: id } });
  };

  const goToDetail = (id) => {
    router.push({ path: `/containers/${id}` });
  };

  const triggerConfirm = (id, action) => {
    pendingId.value = id;
    pendingAction.value = action;
    showConfirm.value = true;
  };

  const executeAction = async () => {
    if (!pendingId.value || !pendingAction.value) return;
    try {
      const token = secureStorage.getItem('token');
      const formData = new FormData();
      formData.append('action', pendingAction.value);
      const res = await apiFetch(`/api/containers/${pendingId.value}/action`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: formData,
      });
      if (res.ok) {
        showToast('Success', `Action ${pendingAction.value} executed.`, 'success');
        await fetchContainers();
      } else {
        showToast('Error', 'Action failed.', 'error');
      }
    } catch (err) {
      console.error(err);
      showToast('Error', 'Action failed.', 'error');
    } finally {
      showConfirm.value = false;
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
    loading,
    filteredContainers,
    runningCount,
    stoppedCount,
    activeLiveId,
    liveStats,
    showConfirm,
    pendingAction,
    actionClass,
    fetchContainers,
    startLiveStats,
    stopLiveStats,
    goToLogs,
    goToShell,
    goToDetail,
    triggerConfirm,
    executeAction,
    formatBytes,
    formatDate: formatContainerDate,
  };
}

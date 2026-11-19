import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { secureStorage } from '../utils/storage';
import { sharedState } from '../utils/sharedState';
import { apiFetch } from '../utils/apiFetch';

const pods = ref([]);
const namespaces = ref([]);
const loading = ref(true);
const namespacesLoading = ref(true);
const selectedNamespace = ref('');
const fetchError = ref('');
let pollInterval = null;
let pollSubscribers = 0;

export function formatPodDate(unix) {
  if (!unix) return 'N/A';
  return new Date(unix * 1000).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export async function fetchNamespaces() {
  namespacesLoading.value = true;
  try {
    const token = secureStorage.getItem('token');
    const res = await apiFetch('/api/namespaces', {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      const data = await res.json();
      namespaces.value = Array.isArray(data) ? data.map((item) => item.name) : [];
      fetchError.value = '';
      if (!selectedNamespace.value) {
        if (namespaces.value.includes(sharedState.k8sDefaultNs)) {
          selectedNamespace.value = sharedState.k8sDefaultNs;
        } else if (namespaces.value.length > 0) {
          selectedNamespace.value = namespaces.value[0];
        }
      }
    } else {
      const data = await res.json().catch(() => ({}));
      fetchError.value = data.error || `Failed to load namespaces (${res.status})`;
      namespaces.value = [];
    }
  } catch (err) {
    console.error(err);
    fetchError.value = 'Failed to load namespaces. Check your connection and try again.';
  } finally {
    namespacesLoading.value = false;
  }
}

export async function fetchPods() {
  if (!selectedNamespace.value) {
    pods.value = [];
    loading.value = false;
    return;
  }

  try {
    const token = secureStorage.getItem('token');
    const res = await apiFetch(
      `/api/pods?namespace=${encodeURIComponent(selectedNamespace.value)}`,
      {
        headers: { Authorization: `Bearer ${token}` },
      },
    );
    if (res.ok) {
      pods.value = await res.json();
      if (!fetchError.value) fetchError.value = '';
    } else {
      const data = await res.json().catch(() => ({}));
      fetchError.value = data.error || `Failed to load pods (${res.status})`;
      pods.value = [];
    }
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
}

function startPolling() {
  if (pollInterval) return;
  fetchNamespaces().then(fetchPods);
  pollInterval = setInterval(fetchPods, 5000);
}

function stopPolling() {
  if (pollInterval) {
    clearInterval(pollInterval);
    pollInterval = null;
  }
}

export function usePods(options = {}) {
  const { autoPoll = true } = options;

  const phaseFilter = ref('all');

  const filteredPods = computed(() => {
    const query = sharedState.searchQuery.toLowerCase();
    return pods.value.filter((pod) => {
      const matchesQuery =
        pod.name.toLowerCase().includes(query) ||
        (pod.images || []).some((image) => image.toLowerCase().includes(query)) ||
        pod.namespace.toLowerCase().includes(query);

      if (!matchesQuery) return false;
      if (phaseFilter.value === 'all') return true;
      const phase = (pod.phase || '').toLowerCase();
      if (phaseFilter.value === 'failed') {
        return phase === 'failed' || phase === 'unknown';
      }
      return phase === phaseFilter.value;
    });
  });

  const runningCount = computed(
    () => pods.value.filter((pod) => pod.phase === 'Running').length,
  );

  const pendingCount = computed(
    () => pods.value.filter((pod) => pod.phase === 'Pending').length,
  );

  const failedCount = computed(
    () =>
      pods.value.filter((pod) =>
        ['Failed', 'Unknown'].includes(pod.phase),
      ).length,
  );

  const refresh = async () => {
    loading.value = true;
    await fetchNamespaces();
    await fetchPods();
  };

  watch(selectedNamespace, () => {
    loading.value = true;
    fetchPods();
  });

  onMounted(() => {
    if (!autoPoll) return;
    pollSubscribers += 1;
    if (pollSubscribers === 1) startPolling();
  });

  onUnmounted(() => {
    if (!autoPoll) return;
    pollSubscribers -= 1;
    if (pollSubscribers === 0) stopPolling();
  });

  return {
    pods,
    namespaces,
    loading,
    namespacesLoading,
    selectedNamespace,
    phaseFilter,
    filteredPods,
    runningCount,
    pendingCount,
    failedCount,
    fetchPods,
    fetchNamespaces,
    refresh,
    fetchError,
    formatDate: formatPodDate,
  };
}

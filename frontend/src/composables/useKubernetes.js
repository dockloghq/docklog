import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { secureStorage } from '../utils/storage';
import { sharedState, kubernetesEnabled } from '../utils/sharedState';
import { apiFetch } from '../utils/apiFetch';

export const pods = ref([]);
export const deployments = ref([]);
export const hpas = ref([]);
export const services = ref([]);
export const events = ref([]);
export const overview = ref(null);
export const namespaces = ref([]);
export const loading = ref(true);
export const namespacesLoading = ref(true);
export const selectedNamespace = ref('');
export const fetchError = ref('');

let pollInterval = null;
let pollSubscribers = 0;

export function podKey(pod) {
  return `${pod.namespace}/${pod.name}`;
}

export function formatK8sDate(unix) {
  if (!unix) return 'N/A';
  return new Date(unix * 1000).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

async function apiGet(path) {
  const token = secureStorage.getItem('token');
  const res = await apiFetch(path, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `Request failed (${res.status})`);
  }
  return res.json();
}

export async function fetchNamespaces() {
  if (!kubernetesEnabled() || !sharedState.k8sAvailable) {
    namespaces.value = [];
    namespacesLoading.value = false;
    return;
  }
  namespacesLoading.value = true;
  try {
    const data = await apiGet('/api/namespaces');
    namespaces.value = Array.isArray(data) ? data.map((item) => item.name) : [];
    fetchError.value = '';
    if (!selectedNamespace.value) {
      if (namespaces.value.includes(sharedState.k8sDefaultNs)) {
        selectedNamespace.value = sharedState.k8sDefaultNs;
      } else if (namespaces.value.length > 0) {
        selectedNamespace.value = namespaces.value[0];
      }
    }
  } catch (err) {
    console.error(err);
    fetchError.value = err.message || 'Failed to load namespaces';
    namespaces.value = [];
  } finally {
    namespacesLoading.value = false;
  }
}

export async function fetchKubernetesData() {
  if (!kubernetesEnabled() || !sharedState.k8sAvailable || !selectedNamespace.value) {
    pods.value = [];
    deployments.value = [];
    hpas.value = [];
    services.value = [];
    events.value = [];
    overview.value = null;
    loading.value = false;
    return;
  }

  const ns = encodeURIComponent(selectedNamespace.value);
  try {
    const [podsData, depData, hpaData, svcData, eventData, overviewData] = await Promise.all([
      apiGet(`/api/pods?namespace=${ns}`),
      apiGet(`/api/deployments?namespace=${ns}`),
      apiGet(`/api/hpas?namespace=${ns}`),
      apiGet(`/api/services?namespace=${ns}`),
      apiGet(`/api/events?namespace=${ns}`),
      apiGet(`/api/k8s/overview?namespace=${ns}`),
    ]);
    pods.value = podsData;
    deployments.value = depData;
    hpas.value = hpaData;
    services.value = svcData;
    events.value = eventData;
    overview.value = overviewData;
    fetchError.value = '';
  } catch (err) {
    console.error(err);
    fetchError.value = err.message || 'Failed to load Kubernetes resources';
  } finally {
    loading.value = false;
  }
}

function startPolling() {
  if (pollInterval) return;
  fetchNamespaces().then(() => {
    loading.value = true;
    fetchKubernetesData();
  });
  pollInterval = setInterval(() => {
    if (typeof document !== 'undefined' && document.hidden) return;
    fetchKubernetesData();
  }, 8000);
}

function stopPolling() {
  if (pollInterval) {
    clearInterval(pollInterval);
    pollInterval = null;
  }
}

function filterByQuery(items, fields) {
  const query = sharedState.searchQuery.toLowerCase();
  if (!query) return items;
  return items.filter((item) =>
    fields.some((field) => String(item[field] || '').toLowerCase().includes(query)),
  );
}

export function useKubernetes(options = {}) {
  const { autoPoll = true } = options;
  const activeTab = ref('overview');
  const phaseFilter = ref('all');

  const filteredPods = computed(() => {
    const query = sharedState.searchQuery.toLowerCase();
    return pods.value.filter((pod) => {
      const matchesQuery =
        pod.name.toLowerCase().includes(query) ||
        pod.namespace.toLowerCase().includes(query) ||
        (pod.images || []).some((image) => image.toLowerCase().includes(query));
      if (!matchesQuery) return false;
      if (phaseFilter.value === 'all') return true;
      const phase = (pod.phase || '').toLowerCase();
      if (phaseFilter.value === 'failed') {
        return phase === 'failed' || phase === 'unknown';
      }
      return phase === phaseFilter.value;
    });
  });

  const filteredDeployments = computed(() => filterByQuery(deployments.value, ['name', 'status']));
  const filteredHPAs = computed(() => filterByQuery(hpas.value, ['name', 'target_name', 'metrics', 'status']));
  const filteredServices = computed(() => filterByQuery(services.value, ['name', 'type', 'ports', 'selector']));
  const filteredEvents = computed(() => filterByQuery(events.value, ['reason', 'message', 'involved_name', 'type']));

  const runningCount = computed(() => pods.value.filter((p) => p.phase === 'Running').length);
  const pendingCount = computed(() => pods.value.filter((p) => p.phase === 'Pending').length);
  const failedCount = computed(() => pods.value.filter((p) => ['Failed', 'Unknown'].includes(p.phase)).length);

  const refresh = async () => {
    loading.value = true;
    await fetchNamespaces();
    await fetchKubernetesData();
  };

  watch(selectedNamespace, () => {
    loading.value = true;
    fetchKubernetesData();
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
    deployments,
    hpas,
    services,
    events,
    overview,
    namespaces,
    loading,
    namespacesLoading,
    selectedNamespace,
    fetchError,
    activeTab,
    phaseFilter,
    filteredPods,
    filteredDeployments,
    filteredHPAs,
    filteredServices,
    filteredEvents,
    runningCount,
    pendingCount,
    failedCount,
    refresh,
    formatDate: formatK8sDate,
    podKey,
  };
}

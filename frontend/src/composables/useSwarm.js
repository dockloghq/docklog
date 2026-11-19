import { ref, computed } from 'vue';
import { secureStorage } from '../utils/storage';
import { sharedState, swarmEnabled } from '../utils/sharedState';
import { apiFetch } from '../utils/apiFetch';

export const swarmOverview = ref(null);
export const swarmServices = ref([]);
export const swarmNodes = ref([]);
export const swarmTasks = ref([]);
export const loading = ref(true);
export const fetchError = ref('');

let pollInterval = null;
let pollSubscribers = 0;

export function formatSwarmDate(unix) {
  if (!unix) return 'N/A';
  return new Date(unix * 1000).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatMemory(bytes) {
  if (!bytes) return 'N/A';
  const gb = bytes / (1024 * 1024 * 1024);
  return `${gb.toFixed(1)} GB`;
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

export async function fetchSwarmData() {
  fetchError.value = '';
  try {
    const [overviewData, servicesData, nodesData, tasksData] = await Promise.all([
      apiGet('/api/swarm/overview').catch(() => null),
      apiGet('/api/swarm/services').catch(() => []),
      apiGet('/api/swarm/nodes').catch(() => []),
      apiGet('/api/swarm/tasks').catch(() => []),
    ]);

    if (overviewData) {
      swarmOverview.value = overviewData;
      sharedState.swarmActive = overviewData.active === true;
    }
    swarmServices.value = Array.isArray(servicesData) ? servicesData : [];
    swarmNodes.value = Array.isArray(nodesData) ? nodesData : [];
    swarmTasks.value = Array.isArray(tasksData) ? tasksData : [];
  } catch (err) {
    fetchError.value = err.message || 'Failed to load Docker Swarm data';
  } finally {
    loading.value = false;
  }
}

export async function scaleService(serviceId, replicas) {
  const token = secureStorage.getItem('token');
  const res = await apiFetch(`/api/swarm/services/${serviceId}/scale`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ replicas: Number(replicas) }),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || 'Failed to scale service');
  }
  await fetchSwarmData();
  return res.json();
}

export async function deleteService(serviceId) {
  const token = secureStorage.getItem('token');
  const res = await apiFetch(`/api/swarm/services/${serviceId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || 'Failed to delete service');
  }
  await fetchSwarmData();
  return res.json();
}

export async function updateNodeAvailability(nodeId, availability) {
  const token = secureStorage.getItem('token');
  const res = await apiFetch(`/api/swarm/nodes/${nodeId}/availability`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ availability }),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || 'Failed to update node availability');
  }
  await fetchSwarmData();
  return res.json();
}

export function startSwarmPolling(intervalMs = 8000) {
  pollSubscribers++;
  if (pollSubscribers === 1) {
    fetchSwarmData();
    pollInterval = setInterval(() => {
      if (typeof document !== 'undefined' && document.hidden) return;
      fetchSwarmData();
    }, intervalMs);
  }
}

export function stopSwarmPolling() {
  pollSubscribers = Math.max(0, pollSubscribers - 1);
  if (pollSubscribers === 0 && pollInterval) {
    clearInterval(pollInterval);
    pollInterval = null;
  }
}

export function useSwarm() {
  return {
    swarmOverview,
    swarmServices,
    swarmNodes,
    swarmTasks,
    loading,
    fetchError,
    fetchSwarmData,
    scaleService,
    deleteService,
    updateNodeAvailability,
    startSwarmPolling,
    stopSwarmPolling,
    formatSwarmDate,
    formatMemory,
  };
}


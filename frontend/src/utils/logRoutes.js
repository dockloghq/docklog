import { dockerEnabled, kubernetesEnabled, swarmEnabled } from './sharedState';

export function logsRouteForPod(podOrKey) {
  const p =
    typeof podOrKey === 'string'
      ? podOrKey
      : `${podOrKey.namespace}/${podOrKey.name}`;
  const query = { p };
  if (dockerEnabled() && kubernetesEnabled()) {
    query.rt = 'kubernetes';
  }
  return { path: '/logs', query };
}

export function logsRouteForContainer(containerOrId, isSwarm = false) {
  let containerId = containerOrId;
  let isSwarmContainer = isSwarm;

  if (containerOrId && typeof containerOrId === 'object') {
    containerId = containerOrId.id;
    if (containerOrId.is_swarm !== undefined) {
      isSwarmContainer = Boolean(containerOrId.is_swarm);
    }
  }

  const query = { c: containerId };
  if (isSwarmContainer) {
    query.rt = 'swarm';
  } else if (dockerEnabled() && (kubernetesEnabled() || swarmEnabled())) {
    query.rt = 'docker';
  }
  return { path: '/logs', query };
}

export function parsePodKey(key) {
  const slash = String(key || '').indexOf('/');
  if (slash <= 0 || slash >= key.length - 1) return null;
  return {
    namespace: key.slice(0, slash),
    name: key.slice(slash + 1),
  };
}

<template>
  <div class="dashboard">
    <section class="dash-hero animate-slide-up">
      <div class="hero-body">
        <div class="hero-copy">
          <span class="hero-eyebrow">
            <span class="live-dot" aria-hidden="true"></span>
            Unified overview
          </span>
          <h1 class="hero-title">{{ greeting }}, {{ username }}</h1>
          <p class="hero-sub">
            {{ standaloneCount }} container{{ standaloneCount === 1 ? '' : 's' }}
            <span class="hero-accent"> · {{ podTotal }} pod{{ podTotal === 1 ? '' : 's' }}</span>
            <span v-if="isSwarmActive && swarmServicesCount" class="hero-accent"> · {{ swarmServicesCount }} swarm service{{ swarmServicesCount === 1 ? '' : 's' }}</span>
            across Docker and Kubernetes
          </p>
        </div>
        <div class="hero-actions">
          <router-link to="/health" class="dash-action-btn">
            <AppIcon name="activity" />
            Diagnostics
          </router-link>
          <router-link to="/containers" class="dash-action-btn">
            <AppIcon name="containers" />
            Containers
          </router-link>
          <router-link v-if="isSwarmActive" to="/swarm" class="dash-action-btn">
            <AppIcon name="server" />
            Swarm
          </router-link>
          <router-link to="/kubernetes" class="dash-action-btn">
            <AppIcon name="box" />
            Kubernetes
          </router-link>
          <button class="dash-action-btn primary" @click="refreshAll" :disabled="loading || k8sLoading || swarmLoading">
            <AppIcon name="refresh" :class="{ spinning: loading || k8sLoading || swarmLoading }" />
            Refresh
          </button>
        </div>
      </div>
      <div class="hero-mesh hero-mesh-hybrid" aria-hidden="true"></div>
    </section>

    <div class="runtime-section-label">Docker</div>
    <section class="metrics-bento animate-slide-up" style="animation-delay: 0.05s">
      <article class="metric-card variant-fleet">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon"><AppIcon name="containers" /></div>
          <span class="metric-badge">Fleet</span>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ standaloneCount }}</span>
            <span class="metric-label">Containers</span>
          </div>
          <div class="metric-visual">
            <div class="donut" :style="{ '--pct': dockerRunningRatio }">
              <svg viewBox="0 0 36 36">
                <circle class="donut-track" cx="18" cy="18" r="15.5" />
                <circle class="donut-fill success" cx="18" cy="18" r="15.5" />
              </svg>
              <span class="donut-label">{{ dockerRunningRatio }}%</span>
            </div>
          </div>
        </div>
        <div class="metric-footer">
          <div class="metric-bar">
            <div class="metric-bar-fill success" :style="{ width: dockerRunningRatio + '%' }"></div>
          </div>
          <span>{{ runningCount }} running · {{ stoppedCount }} stopped</span>
        </div>
      </article>

      <article class="metric-card variant-live">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon success"><AppIcon name="activity" /></div>
          <span class="metric-badge success">Live</span>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ runningCount }}</span>
            <span class="metric-label">Running containers</span>
          </div>
        </div>
        <div class="metric-footer success-text">{{ dockerRunningRatio }}% healthy</div>
      </article>

      <article class="metric-card variant-idle">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon muted"><AppIcon name="stopOutline" /></div>
          <span class="metric-badge dim">Idle</span>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ stoppedCount }}</span>
            <span class="metric-label">Stopped</span>
          </div>
        </div>
        <div class="metric-footer" :class="{ warn: stoppedCount > 0 }">
          {{ stoppedCount ? `${stoppedCount} offline` : 'All containers up' }}
        </div>
      </article>

      <DashboardHostCard />
    </section>

    <!-- Swarm Cluster Section (when Swarm is active) -->
    <template v-if="isSwarmActive">
      <div class="runtime-section-label">Docker Swarm</div>
      <section class="metrics-bento animate-slide-up" style="animation-delay: 0.07s">
        <article class="metric-card variant-fleet">
          <div class="metric-glow" aria-hidden="true"></div>
          <div class="metric-top">
            <div class="metric-icon swarm-accent"><AppIcon name="layers" /></div>
            <span class="metric-badge swarm-badge">Cluster</span>
          </div>
          <div class="metric-body">
            <div class="metric-main">
              <span class="metric-value">{{ swarmServicesCount }}</span>
              <span class="metric-label">Services</span>
            </div>
            <div class="metric-visual">
              <div class="donut" :style="{ '--pct': swarmHealthyRatio }">
                <svg viewBox="0 0 36 36">
                  <circle class="donut-track" cx="18" cy="18" r="15.5" />
                  <circle class="donut-fill success" cx="18" cy="18" r="15.5" />
                </svg>
                <span class="donut-label">{{ swarmHealthyRatio }}%</span>
              </div>
            </div>
          </div>
          <div class="metric-footer">
            <div class="metric-bar">
              <div class="metric-bar-fill success" :style="{ width: swarmHealthyRatio + '%' }"></div>
            </div>
            <span>{{ swarmRunningTasks }} running tasks</span>
          </div>
        </article>

        <article class="metric-card variant-live">
          <div class="metric-glow" aria-hidden="true"></div>
          <div class="metric-top">
            <div class="metric-icon success"><AppIcon name="activity" /></div>
            <span class="metric-badge success">Live</span>
          </div>
          <div class="metric-body">
            <div class="metric-main">
              <span class="metric-value">{{ swarmRunningTasks }}</span>
              <span class="metric-label">Running Tasks</span>
            </div>
          </div>
          <div class="metric-footer success-text">Across {{ swarmNodesCount }} cluster node{{ swarmNodesCount === 1 ? '' : 's' }}</div>
        </article>

        <article class="metric-card variant-idle">
          <div class="metric-glow" aria-hidden="true"></div>
          <div class="metric-top">
            <div class="metric-icon"><AppIcon name="network" /></div>
            <span class="metric-badge">Nodes</span>
          </div>
          <div class="metric-body">
            <div class="metric-main">
              <span class="metric-value">{{ swarmNodesCount }}</span>
              <span class="metric-label">Cluster Nodes</span>
            </div>
          </div>
          <div class="metric-footer">
            {{ swarmManagersCount }} manager{{ swarmManagersCount === 1 ? '' : 's' }} · {{ swarmWorkersCount }} worker{{ swarmWorkersCount === 1 ? '' : 's' }}
          </div>
        </article>

        <article class="metric-card variant-host">
          <div class="metric-glow" aria-hidden="true"></div>
          <div class="metric-top">
            <div class="metric-icon success"><AppIcon name="check" /></div>
            <span class="metric-badge success">Active</span>
          </div>
          <div class="metric-body host-body">
            <div class="swarm-status-box">
              <span class="swarm-status-title">Swarm Mode Active</span>
              <p class="swarm-status-desc">Orchestrating multi-service workloads</p>
            </div>
          </div>
          <div class="metric-footer">
            <router-link to="/swarm" class="swarm-view-link">
              Open Swarm Cluster Control <AppIcon name="chevronRight" :size="12" />
            </router-link>
          </div>
        </article>
      </section>
    </template>

    <div class="runtime-section-label">Kubernetes</div>
    <section class="metrics-bento animate-slide-up" style="animation-delay: 0.08s">
      <article class="metric-card variant-fleet">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon success"><AppIcon name="containers" /></div>
          <span class="metric-badge success">Pods</span>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ podTotal }}</span>
            <span class="metric-label">In {{ selectedNamespace || 'namespace' }}</span>
          </div>
          <div class="metric-visual">
            <div class="donut" :style="{ '--pct': podRunningRatio }">
              <svg viewBox="0 0 36 36">
                <circle class="donut-track" cx="18" cy="18" r="15.5" />
                <circle class="donut-fill success" cx="18" cy="18" r="15.5" />
              </svg>
              <span class="donut-label">{{ podRunningRatio }}%</span>
            </div>
          </div>
        </div>
        <div class="metric-footer">
          {{ k8sRunningCount }} running · {{ pendingCount }} pending · {{ failedCount }} failed
        </div>
      </article>

      <article class="metric-card variant-live">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon"><AppIcon name="box" /></div>
          <span class="metric-badge">Workloads</span>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ overview?.deployments ?? 0 }}</span>
            <span class="metric-label">Deployments</span>
          </div>
          <div class="metric-visual metric-stat-pill">{{ overview?.hpas ?? 0 }} HPAs</div>
        </div>
        <div class="metric-footer">{{ overview?.services ?? 0 }} services exposed</div>
      </article>

      <article class="metric-card variant-idle">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon" :class="warningCount > 0 ? 'warning' : 'muted'">
            <AppIcon name="bell" />
          </div>
          <span class="metric-badge" :class="warningCount > 0 ? 'warning-badge' : 'dim'">Events</span>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ warningCount }}</span>
            <span class="metric-label">Warnings</span>
          </div>
        </div>
        <div class="metric-footer" :class="{ warn: warningCount > 0 }">
          {{ warningCount > 0 ? 'Review cluster warnings' : 'No recent warnings' }}
        </div>
      </article>

      <article class="metric-card variant-host">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon"><AppIcon name="server" /></div>
          <span class="metric-badge">Namespace</span>
        </div>
        <div class="metric-body host-body">
          <select v-model="selectedNamespace" class="namespace-select-inline" :disabled="namespacesLoading">
            <option v-if="namespacesLoading" value="">Loading...</option>
            <option v-for="ns in namespaces" :key="ns" :value="ns">{{ ns }}</option>
          </select>
        </div>
        <div class="metric-footer">Switch namespace for pod metrics below</div>
      </article>
    </section>

    <section class="table-panel animate-slide-up" style="animation-delay: 0.1s">
      <div class="panel-toolbar">
        <div class="toolbar-left">
          <h2>Container overview</h2>
          <p class="toolbar-sub">Docker workloads</p>
        </div>
        <div class="toolbar-right">
          <div class="filter-pills">
            <button
              v-for="f in dockerFilters"
              :key="f.value"
              class="filter-pill"
              :class="{ active: dockerStateFilter === f.value }"
              @click="dockerStateFilter = f.value"
            >
              {{ f.label }}
              <span class="pill-count">{{ f.count }}</span>
            </button>
          </div>
        </div>
      </div>
      <ContainerTable
        :state-filter="dockerStateFilter"
        embedded
        :max-rows="5"
        view-all-to="/containers"
      />
    </section>

    <!-- Swarm Services Table Panel (when Swarm is active) -->
    <section v-if="isSwarmActive" class="table-panel animate-slide-up" style="animation-delay: 0.11s">
      <div class="panel-toolbar">
        <div class="toolbar-left">
          <h2>Swarm services overview</h2>
          <p class="toolbar-sub">Cluster services and active replicas</p>
        </div>
        <div class="toolbar-right">
          <div class="search-box glass">
            <AppIcon name="search" :size="16" />
            <input type="text" v-model="swarmSearchQuery" placeholder="Filter services..." />
          </div>
        </div>
      </div>
      <SwarmServiceTable
        :services="filteredSwarmServices"
        :loading="swarmLoading"
        embedded
        :max-rows="5"
        view-all-to="/swarm"
      />
    </section>

    <section class="table-panel animate-slide-up" style="animation-delay: 0.12s">
      <div class="panel-toolbar">
        <div class="toolbar-left">
          <h2>Pod overview</h2>
          <p class="toolbar-sub">Kubernetes workloads in {{ selectedNamespace }}</p>
        </div>
        <div class="toolbar-right">
          <div class="filter-pills">
            <button
              v-for="f in podFilters"
              :key="f.value"
              class="filter-pill"
              :class="{ active: phaseFilter === f.value }"
              @click="phaseFilter = f.value"
            >
              {{ f.label }}
              <span class="pill-count">{{ f.count }}</span>
            </button>
          </div>
          <div class="search-box glass">
            <AppIcon name="search" :size="16" />
            <input type="text" v-model="sharedState.searchQuery" placeholder="Search pods..." />
          </div>
        </div>
      </div>
      <PodTable
        :pods="filteredPods"
        :loading="k8sLoading"
        embedded
        :max-rows="5"
        :view-all-to="kubernetesPodsLink"
        @refresh="refreshK8s"
      />
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import AppIcon from '../../components/AppIcon.vue';
import ContainerTable from '../../components/ContainerTable.vue';
import PodTable from '../../components/PodTable.vue';
import SwarmServiceTable from '../../components/SwarmServiceTable.vue';
import DashboardHostCard from '../../components/dashboard/DashboardHostCard.vue';
import { useContainers } from '../../composables/useContainers';
import { useKubernetes } from '../../composables/useKubernetes';
import { useSwarm } from '../../composables/useSwarm';
import { useDashboardShared } from '../../composables/useDashboardShared';
import { swarmEnabled } from '../../utils/sharedState';

const dockerStateFilter = ref('all');
const swarmSearchQuery = ref('');
const { containers, standaloneCount, loading, runningCount, stoppedCount, fetchContainers } = useContainers();
const {
  swarmOverview,
  swarmServices,
  swarmNodes,
  loading: swarmLoading,
  fetchSwarmData,
} = useSwarm();

const isSwarmActive = computed(() => {
  return swarmEnabled() || sharedState.swarmActive || swarmOverview.value?.active || swarmServices.value.length > 0;
});

const swarmServicesCount = computed(() => {
  if (swarmOverview.value && swarmOverview.value.services_count !== undefined) {
    return swarmOverview.value.services_count;
  }
  return swarmServices.value.length;
});

const swarmRunningTasks = computed(() => {
  if (swarmOverview.value && swarmOverview.value.running_tasks_count !== undefined) {
    return swarmOverview.value.running_tasks_count;
  }
  return swarmServices.value.reduce((acc, s) => acc + (s.running || 0), 0);
});

const swarmNodesCount = computed(() => {
  if (swarmOverview.value && swarmOverview.value.nodes_count !== undefined) {
    return swarmOverview.value.nodes_count;
  }
  return swarmNodes.value.length;
});

const swarmManagersCount = computed(() => swarmOverview.value?.managers_count ?? 1);
const swarmWorkersCount = computed(() => swarmOverview.value?.workers_count ?? 0);

const swarmHealthyRatio = computed(() => {
  if (!swarmServices.value.length) return 100;
  const totalReplicas = swarmServices.value.reduce((acc, s) => acc + (s.replicas || 1), 0);
  const runningReplicas = swarmServices.value.reduce((acc, s) => acc + (s.running || 0), 0);
  if (!totalReplicas) return 100;
  return Math.min(100, Math.round((runningReplicas / totalReplicas) * 100));
});

const filteredSwarmServices = computed(() => {
  if (!swarmSearchQuery.value) return swarmServices.value;
  const q = swarmSearchQuery.value.toLowerCase().trim();
  return swarmServices.value.filter(
    (s) => (s.name && s.name.toLowerCase().includes(q)) || (s.image && s.image.toLowerCase().includes(q))
  );
});
const {
  pods,
  overview,
  namespaces,
  loading: k8sLoading,
  namespacesLoading,
  selectedNamespace,
  phaseFilter,
  filteredPods,
  runningCount: k8sRunningCount,
  pendingCount,
  failedCount,
  refresh: refreshK8s,
} = useKubernetes();

const { username, greeting, sharedState } = useDashboardShared();

const podTotal = computed(() => pods.value.length);
const warningCount = computed(() => overview.value?.warning_events ?? 0);

const dockerRunningRatio = computed(() => {
  if (!standaloneCount.value) return 0;
  return Math.round((runningCount.value / standaloneCount.value) * 100);
});

const podRunningRatio = computed(() => {
  if (!podTotal.value) return 0;
  return Math.round((k8sRunningCount.value / podTotal.value) * 100);
});

const dockerFilters = computed(() => [
  { label: 'All', value: 'all', count: standaloneCount.value },
  { label: 'Running', value: 'running', count: runningCount.value },
  { label: 'Stopped', value: 'stopped', count: stoppedCount.value },
]);

const podFilters = computed(() => [
  { label: 'All', value: 'all', count: pods.value.length },
  { label: 'Running', value: 'running', count: k8sRunningCount.value },
  { label: 'Pending', value: 'pending', count: pendingCount.value },
  { label: 'Failed', value: 'failed', count: failedCount.value },
]);

const refreshAll = async () => {
  await Promise.all([fetchContainers(), refreshK8s(), fetchSwarmData()]);
};

onMounted(() => {
  fetchSwarmData();
});

const kubernetesPodsLink = computed(() => ({
  path: '/kubernetes',
  query: { tab: 'pods', namespace: selectedNamespace.value },
}));
</script>

<style scoped>
@import './dashboard.css';

.hero-mesh-hybrid {
  background:
    radial-gradient(ellipse 70% 55% at 100% 0%, rgba(var(--accent-rgb), 0.16), transparent 55%),
    radial-gradient(ellipse 55% 45% at 0% 100%, rgba(59, 130, 246, 0.12), transparent 50%);
}

.runtime-section-label {
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-mute);
  margin: 0.15rem 0 -0.35rem;
}

.metric-icon.warning {
  background: rgba(var(--warning-rgb), 0.1);
  color: var(--warning);
  border-color: rgba(var(--warning-rgb), 0.15);
}

.metric-badge.warning-badge {
  background: rgba(var(--warning-rgb), 0.1);
  color: var(--warning);
  border-color: rgba(var(--warning-rgb), 0.15);
}

.metric-icon.swarm-accent {
  background: rgba(var(--accent-rgb), 0.12);
  color: var(--accent);
}

.metric-badge.swarm-badge {
  background: rgba(var(--accent-rgb), 0.12);
  color: var(--accent);
}

.swarm-status-box {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.swarm-status-title {
  font-size: 0.88rem;
  font-weight: 700;
  color: var(--text-main);
}

.swarm-status-desc {
  font-size: 0.75rem;
  color: var(--text-mute);
  margin: 0;
}

.swarm-view-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--accent);
  text-decoration: none;
}

.swarm-view-link:hover {
  text-decoration: underline;
}

.metric-stat-pill {
  font-size: 0.68rem;
  font-weight: 800;
  padding: 0.35rem 0.55rem;
  border-radius: 999px;
  background: var(--bg-subtle);
  color: var(--text-dim);
  border: 1px solid var(--border);
}

.namespace-select-inline {
  width: 100%;
  padding: 0.55rem 0.7rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-main);
  font-size: 0.82rem;
  font-weight: 700;
}
</style>

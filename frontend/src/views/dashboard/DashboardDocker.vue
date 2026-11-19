<template>
  <div class="dashboard">
    <!-- Hero -->
    <section class="dash-hero animate-slide-up">
      <div class="hero-body">
        <div class="hero-copy">
          <span class="hero-eyebrow">
            <span class="live-dot" aria-hidden="true"></span>
            Live overview
          </span>
          <h1 class="hero-title">{{ greeting }}, {{ username }}</h1>
          <p class="hero-sub">
            {{ standaloneCount }} container{{ standaloneCount === 1 ? "" : "s" }} across your fleet
            <span v-if="runningCount" class="hero-accent"> · {{ runningCount }} running</span>
            <span v-if="isSwarmActive && swarmServicesCount" class="hero-accent"> · {{ swarmServicesCount }} swarm service{{ swarmServicesCount === 1 ? '' : 's' }}</span>
          </p>
        </div>
        <div class="hero-actions">
          <router-link to="/health" class="dash-action-btn">
            <AppIcon name="activity" />
            Diagnostics
          </router-link>
          <router-link to="/containers" class="dash-action-btn">
            <AppIcon name="containers" />
            All containers
          </router-link>
          <router-link v-if="isSwarmActive" to="/swarm" class="dash-action-btn">
            <AppIcon name="server" />
            Swarm
          </router-link>
          <button class="dash-action-btn primary" @click="refresh" :disabled="loading || swarmLoading">
            <AppIcon name="refresh" :class="{ spinning: loading || swarmLoading }" />
            Refresh
          </button>
        </div>
      </div>
      <div class="hero-mesh" aria-hidden="true"></div>
    </section>

    <!-- Metrics -->
    <section class="metrics-bento animate-slide-up" style="animation-delay: 0.05s">
      <article class="metric-card variant-fleet">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon">
            <AppIcon name="containers" />
          </div>
          <span class="metric-badge">Fleet</span>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ standaloneCount }}</span>
            <span class="metric-label">Total containers</span>
          </div>
          <div class="metric-visual">
            <div class="donut" :style="{ '--pct': runningRatio }">
              <svg viewBox="0 0 36 36">
                <circle class="donut-track" cx="18" cy="18" r="15.5" />
                <circle class="donut-fill success" cx="18" cy="18" r="15.5" />
              </svg>
              <span class="donut-label">{{ runningRatio }}%</span>
            </div>
          </div>
        </div>
        <div class="metric-footer">
          <div class="metric-bar">
            <div class="metric-bar-fill success" :style="{ width: runningRatio + '%' }"></div>
          </div>
          <span>{{ runningCount }} running · {{ stoppedCount }} stopped</span>
        </div>
      </article>

      <article class="metric-card variant-live">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon success">
            <AppIcon name="activity" />
          </div>
          <span class="metric-badge success">Live</span>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ runningCount }}</span>
            <span class="metric-label">Running now</span>
          </div>
          <div class="metric-visual">
            <div class="pulse-stack">
              <span v-for="n in 4" :key="n" class="pulse-bar" :style="{ height: pulseHeights[n - 1] + '%', animationDelay: n * 0.12 + 's' }"></span>
            </div>
          </div>
        </div>
        <div class="metric-footer success-text">
          {{ runningRatio }}% of fleet is healthy
        </div>
      </article>

      <article class="metric-card variant-idle">
        <div class="metric-glow" aria-hidden="true"></div>
        <div class="metric-top">
          <div class="metric-icon muted">
            <AppIcon name="stopOutline" />
          </div>
          <span class="metric-badge dim">Idle</span>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ stoppedCount }}</span>
            <span class="metric-label">Stopped</span>
          </div>
          <div class="metric-visual">
            <div class="donut muted" :style="{ '--pct': stoppedRatio }">
              <svg viewBox="0 0 36 36">
                <circle class="donut-track" cx="18" cy="18" r="15.5" />
                <circle class="donut-fill muted" cx="18" cy="18" r="15.5" />
              </svg>
              <span class="donut-label">{{ stoppedRatio }}%</span>
            </div>
          </div>
        </div>
        <div class="metric-footer" :class="{ warn: stoppedCount > 0 }">
          {{ stoppedCount ? `${stoppedCount} container${stoppedCount === 1 ? '' : 's'} offline` : "All services operational" }}
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

    <!-- Container table -->
    <section class="table-panel animate-slide-up" style="animation-delay: 0.1s">
      <div class="panel-toolbar">
        <div class="toolbar-left">
          <h2>Container overview</h2>
          <p class="toolbar-sub">Manage workloads and jump into logs</p>
        </div>
        <div class="toolbar-right">
          <div class="filter-pills">
            <button
              v-for="f in filters"
              :key="f.value"
              class="filter-pill"
              :class="{ active: stateFilter === f.value }"
              @click="stateFilter = f.value"
            >
              {{ f.label }}
              <span class="pill-count">{{ f.count }}</span>
            </button>
          </div>
          <div class="search-box glass">
            <AppIcon name="search" :size="16" />
            <input type="text" v-model="sharedState.searchQuery" placeholder="Search..." />
          </div>
        </div>
      </div>
      <ContainerTable
        :state-filter="stateFilter"
        embedded
        :max-rows="5"
        view-all-to="/containers"
      />
    </section>

    <!-- Swarm Services Table Panel (when Swarm is active) -->
    <section v-if="isSwarmActive" class="table-panel animate-slide-up" style="animation-delay: 0.12s">
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import AppIcon from "../../components/AppIcon.vue";
import ContainerTable from "../../components/ContainerTable.vue";
import SwarmServiceTable from "../../components/SwarmServiceTable.vue";
import DashboardHostCard from "../../components/dashboard/DashboardHostCard.vue";
import { useContainers } from "../../composables/useContainers";
import { useSwarm } from "../../composables/useSwarm";
import { useDashboardShared } from "../../composables/useDashboardShared";
import { sharedState, swarmEnabled } from "../../utils/sharedState";

const stateFilter = ref("all");
const swarmSearchQuery = ref("");
const { containers, standaloneCount, loading, runningCount, stoppedCount, fetchContainers } = useContainers();
const {
  swarmOverview,
  swarmServices,
  swarmNodes,
  loading: swarmLoading,
  fetchSwarmData,
} = useSwarm();

const { username, greeting } = useDashboardShared();

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

const runningRatio = computed(() => {
  if (!standaloneCount.value) return 0;
  return Math.round((runningCount.value / standaloneCount.value) * 100);
});

const stoppedRatio = computed(() => {
  if (!standaloneCount.value) return 0;
  return Math.round((stoppedCount.value / standaloneCount.value) * 100);
});

const pulseHeights = computed(() => {
  const base = Math.max(35, Math.min(95, runningRatio.value));
  return [base * 0.55, base * 0.85, base, base * 0.7];
});

const filters = computed(() => [
  { label: "All", value: "all", count: standaloneCount.value },
  { label: "Running", value: "running", count: runningCount.value },
  { label: "Stopped", value: "stopped", count: stoppedCount.value },
]);

const refresh = async () => {
  await Promise.all([fetchContainers(), fetchSwarmData()]);
};

onMounted(() => {
  fetchSwarmData();
});
</script>

<style scoped>
@import './dashboard.css';

.runtime-section-label {
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-mute);
  margin: 0.15rem 0 -0.35rem;
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
</style>

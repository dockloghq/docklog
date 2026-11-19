<template>
  <div class="containers-view animate-fade-in">
    <section class="containers-hero">
      <div class="hero-copy">
        <span class="hero-eyebrow">Fleet control</span>
        <h1>Container management</h1>
        <p class="hero-sub">
          Start, stop, restart, and inspect workloads across your Docker host.
        </p>
      </div>
      <div class="hero-stats">
        <div class="hero-stat">
          <span class="hero-stat-val">{{ standaloneCount }}</span>
          <span class="hero-stat-lbl">Total</span>
        </div>
        <div class="hero-stat success">
          <span class="hero-stat-val">{{ runningCount }}</span>
          <span class="hero-stat-lbl">Running</span>
        </div>
        <div class="hero-stat muted">
          <span class="hero-stat-val">{{ stoppedCount }}</span>
          <span class="hero-stat-lbl">Stopped</span>
        </div>
      </div>
      <div class="hero-mesh" aria-hidden="true"></div>
    </section>

    <section class="containers-panel">
      <div class="panel-toolbar">
        <div class="toolbar-left">
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
        </div>
        <div class="toolbar-right">
          <div class="search-box">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <input
              type="text"
              v-model="sharedState.searchQuery"
              placeholder="Filter by name or image..."
            />
          </div>
          <button class="refresh-btn" @click="fetchContainers" :disabled="loading">
            <svg
              viewBox="0 0 24 24"
              width="16"
              height="16"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              :class="{ spinning: loading }"
            >
              <polyline points="23 4 23 10 17 10"></polyline>
              <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
            </svg>
            Refresh
          </button>
        </div>
      </div>

      <ContainerTable
        :state-filter="stateFilter"
        show-inline-stats
        embedded
      />
    </section>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import ContainerTable from "../components/ContainerTable.vue";
import { useContainers } from "../composables/useContainers";
import { sharedState } from "../utils/sharedState";

const stateFilter = ref("all");

const { containers, loading, runningCount, stoppedCount, fetchContainers } =
  useContainers();

const standaloneCount = computed(
  () => containers.value.filter((c) => !c.is_swarm).length,
);

const filters = computed(() => [
  { label: "All", value: "all", count: standaloneCount.value },
  { label: "Running", value: "running", count: runningCount.value },
  { label: "Stopped", value: "stopped", count: stoppedCount.value },
]);
</script>

<style scoped>
.containers-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-bottom: 2rem;
}

.containers-hero {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1.5rem;
  flex-wrap: wrap;
  padding: 1.5rem 1.75rem;
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border);
  background: linear-gradient(135deg, var(--bg-card) 0%, var(--bg-card) 55%, rgba(var(--accent-rgb), 0.04) 100%);
  overflow: hidden;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}

.hero-mesh {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 70% 80% at 100% 0%, rgba(var(--accent-rgb), 0.14), transparent 55%),
    radial-gradient(ellipse 40% 50% at 0% 100%, rgba(var(--success-rgb), 0.06), transparent 50%);
  pointer-events: none;
}

.hero-copy {
  position: relative;
  z-index: 1;
}

.hero-eyebrow {
  display: block;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--accent);
  margin-bottom: 0.4rem;
}

.containers-hero h1 {
  font-size: clamp(1.35rem, 2.5vw, 1.75rem);
  font-weight: 800;
  letter-spacing: -0.03em;
  margin: 0 0 0.35rem;
  color: var(--text-main);
}

.hero-sub {
  margin: 0;
  font-size: 0.9rem;
  color: var(--text-dim);
  max-width: 480px;
}

.hero-stats {
  position: relative;
  z-index: 1;
  display: flex;
  gap: 0.65rem;
  flex-wrap: wrap;
}

.hero-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 72px;
  padding: 0.65rem 1rem;
  border-radius: var(--radius-md);
  background: var(--bg-input);
  border: 1px solid var(--border);
}

.hero-stat-val {
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: var(--text-main);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.hero-stat-lbl {
  margin-top: 0.2rem;
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-mute);
}

.hero-stat.success .hero-stat-val { color: var(--success); }
.hero-stat.muted .hero-stat-val { color: var(--text-dim); }

.containers-panel {
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border);
  background: var(--bg-card);
  padding: 1.15rem 1.15rem 0.35rem;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}

.panel-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex-wrap: wrap;
}

.filter-pills {
  display: flex;
  gap: 0.35rem;
  padding: 0.25rem;
  border-radius: var(--radius-md);
  background: var(--bg-input);
  border: 1px solid var(--border);
}

.filter-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.8rem;
  border-radius: calc(var(--radius-md) - 2px);
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-dim);
  transition: all 0.2s ease;
}

.filter-pill:hover {
  color: var(--text-main);
  background: var(--bg-subtle);
}

.filter-pill.active {
  background: var(--accent);
  color: #fff;
  box-shadow: 0 4px 12px rgba(var(--accent-rgb), 0.3);
}

.pill-count {
  font-size: 0.65rem;
  padding: 0.1rem 0.35rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.15);
  font-variant-numeric: tabular-nums;
}

.filter-pill:not(.active) .pill-count {
  background: var(--bg-subtle);
  color: var(--text-mute);
}

.search-box {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.6rem 1rem;
  border-radius: var(--radius-md);
  background: var(--bg-input);
  border: 1px solid var(--border);
  min-width: 260px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.search-box:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(var(--accent-rgb), 0.1);
}

.search-box input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 0.85rem;
  font-weight: 600;
  width: 100%;
}

.search-box svg {
  color: var(--text-mute);
  flex-shrink: 0;
}

.refresh-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.6rem 1rem;
  border-radius: var(--radius-md);
  background: var(--accent);
  color: #fff;
  font-size: 0.8rem;
  font-weight: 700;
  transition: background 0.2s, transform 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--accent-hover);
  transform: translateY(-1px);
}

.refresh-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.spinning {
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 900px) {
  .containers-hero {
    flex-direction: column;
    align-items: stretch;
  }

  .hero-stats {
    width: 100%;
  }

  .hero-stat {
    flex: 1;
  }

  .panel-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-left,
  .toolbar-right {
    width: 100%;
  }

  .filter-pills {
    width: 100%;
    justify-content: space-between;
  }

  .filter-pill {
    flex: 1;
    justify-content: center;
  }

  .search-box {
    width: 100%;
    min-width: 0;
  }

  .refresh-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>

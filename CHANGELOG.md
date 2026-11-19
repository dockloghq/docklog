# Changelog

All notable changes to **DockLog** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.0] - 2026-11-19

### 🐝 Complete Docker Swarm Orchestration (New Feature)
- **Full Docker Swarm Engine Support**: Added comprehensive Docker Swarm cluster detection, monitoring, and management across the backend and frontend.
- **Swarm Services & Task Management**: Added dedicated Swarm views (`frontend/src/views/Swarm.vue`, `frontend/src/views/SwarmServiceDetail.vue`) for listing services, replica distributions, placement constraints, and task/container allocations.
- **Fleet & Swarm Overview**: Introduced `frontend/src/views/dashboard/FleetOverview.vue` and `frontend/src/components/SwarmServiceTable.vue` for real-time node topology, service scaling, and replica health.
- **Swarm WebSocket Streaming**: Added live Swarm service log aggregation and container event streaming (`server/routes_swarm.go`, `models/swarm.go`).
- **Interactive Container & Service Hover Cards**: Added `frontend/src/components/ContainerHoverCard.vue` and `frontend/src/composables/useSwarm.js` providing instant service metadata, image digests, and health without navigating away.

### 🚀 Performance & Storage Optimization
- **Hybrid In-Memory Metric Architecture**: Implemented a 24-hour circular ring buffer (`stats/ringbuffer.go`) storing 30-second live metrics entirely in RAM (~2 MB footprint), eliminating millions of write queries to `docklog.db`.
- **Hourly Metric Rollups**: 1-hour downsampled aggregate summaries (min/avg/max) are written to SQLite, cutting database growth by **99.2%** (from ~2,880 rows/day to 24 rows/day per container).
- **SQLite Space Reclamation**: Added `PRAGMA auto_vacuum = INCREMENTAL`, `PRAGMA journal_mode = WAL`, and automated daily incremental vacuum routines that delete data older than 7 days and immediately return freed pages back to the host filesystem.
- **Docker Socket Inspection Caching**: Added an in-memory TTL cache (`containers/specs_cache.go`) for container specs, eliminating repetitive `ContainerInspect` round-trips during 3-second UI polling loops (dropped API response time from ~250ms to ~3ms).
- **HTTP Compression & Caching**: Added Echo Gzip compression middleware (skipping WebSockets) and configured 1-year immutable caching for frontend hashed chunks with `no-cache` for index HTML.
- **Client Polling Throttling**: Added `document.hidden` guards across containers, swarm, and Kubernetes polling loops to freeze background requests when browser tabs are inactive or minimized.

### 📊 System Health & Stats Monitoring
- **Timezone-Standardized Metric History**: Upgraded `/api/system/history` to emit standard ISO 8601 / RFC 3339 UTC timestamps, fixing a bug where local browser timezone offsets filtered out data points across `1H`, `6H`, `12H`, and `Custom` range filters.
- **Peak Spike Visibility**: Added time-tagged indicators to "Peak CPU Spike" and "Peak Memory" cards (e.g. `90.0% at 18:45`), ensuring peak cards precisely match historical spikes on the charts.
- **Database & Storage Health Dashboard**: Added real-time database file size and metric count monitoring to `Settings -> About`, along with a manual **Vacuum Now** action (`POST /api/system/vacuum`).

### 🛠️ Dynamic Versioning & CI/CD Pipeline
- **Dynamic Application Versioning**: Replaced static version strings with dynamic resolution across Go backend, CLI flags, Docker runtime environments, and Vue frontend (`Settings.vue` and `Login.vue`).
- **Automated CI/CD Workflow**: Updated GitHub Actions (`.github/workflows/docker-publish.yml`) to support Git release tags (`v*.*.*`) and automatically inject `DOCKLOG_VERSION` build arguments into multi-arch Docker Buildx builds.
- **In-App Version Update Checker**: Added GitHub Releases API verification with 15-minute caching to detect new releases and prompt users with update banners in the UI.

### 🎨 UI & Streaming Improvements
- **Virtualized Log Viewer**: Optimized live log streaming using CSS `content-visibility: auto; contain-intrinsic-size: 0 24px;` and `requestAnimationFrame` batching for fluid 60–120 FPS scrolling with tens of thousands of lines.
- **Resilient WebSocket Connection**: Implemented exponential backoff reconnection (1.5s $\to$ 10s max) with interactive visual status badges and manual retry buttons in the log viewer.

---

## [1.0.0] - 2026-11-18

### Initial Release
- Docker Standalone and Kubernetes log viewer and metrics monitoring platform.
- Real-time WebSockets streaming for container stdout/stderr.
- Interactive terminal shell execution (`xterm.js`).
- Role-based user access controls (RBAC) and JWT authentication.
- Alerting & outbound webhooks (Slack, Discord, Teams, Custom HTTP).

# Changelog

All notable changes to **DockLog** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.0] - 2026-11-19

### 🐝 New Features
- **Docker Swarm Orchestration**: Added cluster detection and management for Swarm mode with dedicated views (`/swarm` and `/swarm/services/:id`).
- **Swarm Services & Task Management**: Added service inspection, replica count scaling with RBAC checks, node assignment tracking, and task status monitoring.
- **Aggregated Swarm Service Logs**: Implemented real-time WebSocket log streaming aggregated across all running task replicas of a Swarm service.
- **Fleet Overview Dashboard**: Added cluster-wide node status overview with availability controls (`active`, `pause`, `drain`) and replica health breakdown.
- **Service & Container Hover Cards**: Added instant metadata preview cards displaying container image digests, ports, status, and node placement without page navigation.
- **In-Memory Metric Ring Buffer**: Implemented a 24-hour circular ring buffer holding 30-second live metrics entirely in memory (~2 MB RAM), avoiding continuous database writes.
- **Hourly Metric Downsampling**: Added 1-hour aggregate rollups (min, avg, max) written to SQLite, reducing database row growth by 99.2% (24 rows/day per container).
- **Automated Database Maintenance**: Configured SQLite `incremental_vacuum` and automatic 7-day retention cleanup jobs to release unused disk pages back to the operating system.
- **Database Health & Compaction Controls**: Added live SQLite file size inspection (`GET /api/system/db-stats`) and a manual compaction trigger (`POST /api/system/vacuum`) in Settings.
- **Dynamic Application Versioning**: Implemented dynamic version injection across backend Go code, CLI flags, Docker runtime environments, and frontend interfaces.
- **In-App Update Checker**: Added GitHub Releases API verification with 15-minute caching to detect new releases and prompt administrators in the UI.
- **Time-Tagged Peak Indicators**: Added exact timestamp indicators to Peak CPU Spike and Peak Memory summary cards in the System Health view.
- **Container Specs Caching**: Added in-memory TTL caching for container inspect calls to reduce Docker daemon socket load during polling.
- **Off-Screen Log Virtualization**: Added CSS `content-visibility: auto` to offload rendering of non-visible log lines during high-throughput log streaming.

### 🛠️ Bug Fixes
- **Health Timeline Timezone Offset**: Formatted `/api/system/history` timestamps as RFC 3339 UTC strings, fixing an issue where local browser timezone offsets filtered out data points in 1H, 6H, 12H, and Custom range filters.
- **Settings Brand Badge Wrapping**: Fixed title layout in Settings so the version tag stays aligned horizontally beside the brand name on the same line.
- **Active Navigation Indicator**: Replaced inset box-shadow styling on active sidebar links with a vertical accent indicator to prevent distorted corner borders.
- **Background Resource Usage**: Added `document.hidden` guards across containers, swarm, and Kubernetes polling loops to stop background HTTP requests when the browser tab is hidden or minimized.
- **WebSocket Reconnection Backoff**: Added exponential backoff and connection state banners to prevent reconnection loops during network drops.

---

## [1.0.1] - 2026-11-19

### ☸️ New Features
- **Kubernetes Cluster Support**: Added namespace discovery and pod listings for Kubernetes and Hybrid (Docker + K8s) modes.
- **Pod Details View**: Added status inspection, nested container breakdowns, cluster event feeds, live pod logs, and integrated shell access.
- **Runtime Mode Switcher**: Added navigation controls to toggle between Docker, Kubernetes, or Combined Fleet views.
- **Time-Range Log Export**: Added filtering and file export for container log streams over selected date and time intervals.

### 🛠️ Bug Fixes
- **Date Range Validation**: Enforced start and end date requirements before allowing custom log export requests.
- **ANSI Color Highlighting**: Corrected log severity highlight offsets on lines containing ANSI color sequences.
- **Pod Table Layout**: Fixed column alignment and header positioning on smaller screen sizes.

---

## [1.0.0] - 2026-11-18

### 🚀 New Features
- **Docker Log Streaming**: Real-time WebSocket log streaming for standalone Docker containers (stdout/stderr).
- **Interactive Shell**: In-browser interactive terminal execution using `xterm.js`.
- **Role-Based Access Control**: RBAC user accounts, restricted container access permissions, and session management.
- **Audit Logging**: Structured audit trail recording administrative and authentication events.
- **Alert Notifications**: Outbound webhook integrations for Slack, Discord, Microsoft Teams, and custom HTTP endpoints.

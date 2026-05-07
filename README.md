# DockLog 🐳

<p align="center">
  <img src="frontend/public/logo-horizontal.png?v=2" alt="DockLog Logo" width="420">
</p>

<p align="center">
  <strong>High-performance, real-time Docker log viewer built for teams.</strong>
</p>

<p align="center">
  Lightweight. Secure. Modern. Built for real-world Docker environments.
</p>

<p align="center">
  DockLog provides real-time log streaming, RBAC, audit logging, system monitoring, and container management in a clean modern interface.
</p>

<p align="center">
  Think of it as a more powerful and team-friendly alternative to Dozzle.
</p>

<p align="center">
  <img src="https://img.shields.io/docker/pulls/aimldev/docklog" alt="Docker Pulls">
  <img src="https://img.shields.io/github/license/Team-AI-ML/docklog" alt="License">
  <img src="https://img.shields.io/badge/Backend-Go--1.24-00add8" alt="Backend">
  <img src="https://img.shields.io/badge/Frontend-Vue--3-42b883" alt="Frontend">
  <img src="https://img.shields.io/badge/UI-Compact--Airy-6366f1" alt="UI">
  <img src="https://img.shields.io/github/stars/Team-AI-ML/docklog?style=social" alt="GitHub Stars">
</p>

---

> ⚡ Average setup time: under 2 minutes.

DockLog focuses on fast deployment, low resource usage, and team-safe Docker visibility without requiring heavyweight observability tooling.

Built for developers who want real-time Docker visibility with modern operational workflows.

---

# 📸 Preview

## 📊 Dashboard

![Dashboard](assets/dashboard.png)

Real-time Docker monitoring with lightweight system metrics and container controls.

---

## 🐳 Container Management

![Containers](assets/containers.png)

Monitor, control, and manage containers with fast operational actions.

---

## 🔐 RBAC & Staff Management

![RBAC](assets/rbac.png)

Granular container-level permissions with wildcard and regex-based access control.

---

## 🕵️ Security Audit Logs

![Audit Logs](assets/audit-logs.png)

Track administrative actions and security events with a complete audit trail.

---

# 🚀 Why DockLog?

Most Docker log viewers are built for a single administrator.

DockLog is built for the entire team.

- **Team-First Security**  
  Give developers access only to the containers they need using wildcard and regex-based RBAC.

- **Audit Everything**  
  Track exactly who restarted, stopped, or modified containers with full audit logging.

- **Zero-Config Deployment**  
  No external database. No large configuration files. Just run the container and start monitoring.

- **Performance Without Compromise**  
  Written in Go for extremely low memory and CPU usage.

- **Modern Developer Experience**  
  Fast, responsive UI designed for operational workflows.

---

# ✨ Features

## 📜 Real-Time Log Streaming

- WebSocket-based live streaming
- Infinite scrolling through millions of log lines
- Smart auto-scroll behavior
- Manual history loading with progress tracking
- RFC3339Nano precision timestamp filtering
- Automatic reconnection handling
- Search and filtering support

---

## 🔐 Advanced RBAC

- Wildcard container permissions (`backend-*`)
- Full regex support (`^prod-.*$`)
- Per-user action permissions
- Restricted vs. global visibility
- Staff account management
- Container-level access isolation

---

## 🕵️ Audit & Security

- Full audit trail for administrative actions
- First-login forced password reset
- JWT authentication
- Encrypted local storage
- Connectivity monitoring
- Backend availability detection

---

## 📊 System Monitoring

- Global host CPU and memory monitoring
- Historical usage analytics
- Assigned core detection
- Container memory profiling
- Normalized CPU utilization
- Real-time system metrics

---

## ⚡ High Performance

- Lightweight Go backend
- Low memory footprint (~30MB to 50MB typical)
- Handles thousands of log lines per second
- Embedded frontend for single-binary deployment
- Optimized WebSocket streaming
- Fast container filtering

---

# 📈 Performance Benchmarks

| Metric          | Value            |
| --------------- | ---------------- |
| RAM Usage       | ~30MB to 50MB    |
| Backend         | Go 1.24          |
| Frontend        | Vue 3            |
| Log Throughput  | 10k+ lines/sec   |
| Deployment Type | Single Container |

---

# 🆚 DockLog vs Alternatives

| Feature                     | DockLog | Dozzle  | Loki/Grafana   |
| --------------------------- | ------- | ------- | -------------- |
| Real-time logs              | ✅      | ✅      | ✅             |
| Multi-user RBAC             | ✅      | ❌      | Advanced setup |
| Wildcard permissions        | ✅      | ❌      | Complex        |
| Audit logging               | ✅      | ❌      | Partial        |
| Lightweight VPS usage       | ✅      | ✅      | ❌             |
| Single-container deployment | ✅      | ✅      | ❌             |
| Historical observability    | Limited | Limited | ✅             |
| Kubernetes observability    | Planned | ❌      | ✅             |

---

# 🛠 Tech Stack

| Layer            | Technology                |
| ---------------- | ------------------------- |
| Backend          | Go 1.24 + Echo            |
| Frontend         | Vue 3 + Vite              |
| Streaming        | WebSockets                |
| Database         | SQLite                    |
| Container Engine | Docker SDK                |
| Styling          | Vanilla CSS Design System |

---

# ⚙️ Configuration

## Environment Variables

| Variable      | Description          | Default                       |
| ------------- | -------------------- | ----------------------------- |
| `PORT`        | Application port     | `8000`                        |
| `SECRET_KEY`  | JWT signing secret   | `secret-key-change-this`      |
| `DB_PATH`     | SQLite database path | `docklog.db`                  |
| `DOCKER_HOST` | Docker daemon socket | `unix:///var/run/docker.sock` |

---

## Generate Secure Secret Key

```bash
openssl rand -base64 32
```

---

# 👥 User Roles

## 👑 Administrator

- Full container visibility
- User management
- Audit log access
- Full container control

---

## 🛠 Staff Member

Container visibility is controlled using patterns such as:

```text
redis
backend-*
prod-*, *-app
^prod-.*$
```

Users only see containers matching their assigned rules.

---

# 🚀 Getting Started

## 🔑 Default Login

| Username | Password   |
| -------- | ---------- |
| `admin`  | `admin123` |

> [!WARNING]
> Change the default administrator password immediately after first login.

---

# 🐳 Docker Compose

```yaml
version: "3.8"

services:
  docklog:
    image: aimldev/docklog:latest
    container_name: docklog

    ports:
      - "8888:8000"

    environment:
      - SECRET_KEY=your-secure-key-here
      - DB_PATH=/app/data/docklog.db

    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./data:/app/data

    restart: unless-stopped
```

Start DockLog:

```bash
docker compose up -d
```

Open:

```text
http://localhost:8888
```

---

# 🐳 Direct Docker Run

```bash
docker run -d \
  --name docklog \
  -p 8888:8000 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/data:/app/data \
  -e SECRET_KEY=your-secure-key-here \
  -e DB_PATH=/app/data/docklog.db \
  --restart unless-stopped \
  aimldev/docklog:latest
```

---

# 🔒 Docker Socket Security

DockLog requires access to:

```text
/var/run/docker.sock
```

This effectively grants Docker API access to the application.

Recommended production setup:

- place behind Nginx or Traefik
- enable HTTPS
- restrict public access
- expose only trusted networks
- use strong administrator credentials
- rotate `SECRET_KEY` periodically

---

# 📂 Project Structure

| Path                 | Description           |
| -------------------- | --------------------- |
| `main.go`            | Main Go server        |
| `frontend/`          | Vue 3 frontend        |
| `db/`                | SQLite initialization |
| `.github/workflows/` | CI/CD pipelines       |
| `docs/`              | Project documentation |

---

# 📚 Documentation

- [Architecture Overview](docs/ARCHITECTURE.md)
- [Security & RBAC](docs/SECURITY.md)

---

# 🛣️ Roadmap

Planned features:

- log retention controls
- notifications
- API tokens
- multi-host support
- Kubernetes support
- external authentication providers
- organization/team management

---

# 🤝 Contributing

Contributions, issues, and feature requests are welcome.

1. Fork the repository
2. Create a feature branch
3. Submit a pull request

---

# 📦 Docker Hub

https://hub.docker.com/r/aimldev/docklog

---

# 🔓 Open Source License

DockLog is open-source software licensed under the MIT License.

You are free to:

- use
- modify
- distribute
- self-host
- commercialize

See the `LICENSE` file for full details.

---

# ❤️ Support The Project

If DockLog helps your workflow:

- star the repository
- report bugs
- contribute improvements
- share feedback
- suggest new features

---

<p align="center">
  Built for developers who want real-time Docker visibility without deploying an observability cathedral.
</p>

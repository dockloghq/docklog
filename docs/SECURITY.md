# Security & RBAC

DockLog is built with security as a first-class citizen. It implements a multi-layer security model to ensure that container management is both accessible and restricted.

## 🔑 Authentication

DockLog uses **JWT (JSON Web Tokens)** for stateless authentication.
- **Token Issue**: Tokens are issued upon successful login at `/api/token`.
- **Encryption**: Tokens are signed using the `SECRET_KEY` environment variable.
- **Expiry**: Tokens are valid for 24 hours by default.

### Generating a Secure Secret Key
For production deployments, you **must** generate a unique secret key. You can use the following command:
```bash
openssl rand -base64 32
```
Pass this value to the `SECRET_KEY` environment variable.

## 🔐 Role-Based Access Control (RBAC)

Permissions are divided into two categories: **Visibility** and **Actions**.

### 1. Visibility (Regex Filtering)
Administrators can restrict which containers a user can see using **Allowed Containers** patterns.
- **Full Access**: `.*` (Regex for "everything")
- **Specific Match**: `redis` (Only matches the exact name)
- **Wildcard**: `backend*` (Matches `backend-api`, `backend-db`, etc.)
- **Multiple**: `api-*, db-*` (Comma-separated patterns)

The backend translates glob wildcards into anchored regex patterns to prevent accidental exposure.

### 2. Action Rights
Even if a container is visible, a user must have explicit rights to perform actions:
- `can_start`: Ability to start stopped containers.
- `can_stop`: Ability to stop running containers.
- `can_restart`: Ability to restart containers.
- `can_delete`: Ability to permanently remove containers.

## 🕵️ Audit Logging

DockLog maintains a permanent record of all sensitive actions in the `audit_logs` table.
Each entry includes:
- **Timestamp**
- **User ID & Username**
- **Action Performed** (e.g., `START`, `STOP`, `RESET_PASSWORD`)
- **Target Resource** (Container ID or Name)
- **Status** (Success/Failure/Forbidden)

Administrators can view these logs directly in the **Admin Panel**.

## 🛡️ Best Practices

1.  **Reverse Proxy**: Always run DockLog behind a reverse proxy (Nginx, Traefik, Caddy) to handle SSL/TLS termination.
2.  **Docker Socket**: Be careful with mounting the Docker socket. Only expose DockLog to trusted networks or use a VPN.
3.  **Password Policy**: Minimum 8 characters. First login requires a password change. In production, the default `admin` account gets a random password logged once at startup (or set `ADMIN_PASSWORD`).

## 🌐 Client Access Control

DockLog rejects direct `/api` and `/ws` calls from arbitrary browser origins.

| Client | Requirements |
| --- | --- |
| Vue web UI (browser) | `X-DockLog-Client: web` + `Origin` or `Referer` matching the server or `ALLOWED_ORIGINS` |
| Native mobile app (Flutter, Android/iOS) | No browser Origin; JWT authentication after login |
| WebSocket (Vue UI in browser) | Valid Origin (browsers cannot set custom WS headers) |
| WebSocket (native mobile app) | No Origin + JWT subprotocol |

The Flutter app (`com.docklog.app`) is **native mobile only** — it is not a web client. Do not use Flutter Web; use the bundled Vue dashboard in the browser.

Environment variables:

- `CLIENT_ACCESS=strict` — default; set `off` only for local debugging
- `ALLOWED_ORIGINS` — comma-separated extra web origins
- `TRUST_PROXY=true` — honor `X-Forwarded-Host` / `X-Forwarded-Proto` (only when behind a trusted reverse proxy)
- `ADMIN_PASSWORD` — optional initial admin password (min 8 chars); random in production if unset
- `ENV=production` — disables localhost origin bypass

DockLog also sends standard security headers (CSP, `X-Frame-Options`, etc.) on all responses.

Native mobile apps connect with JWT after login (no browser Origin). Mobile setup docs belong in your private Flutter repo.

<template>
  <div class="shell-page">
    <header class="shell-toolbar glass">
      <div class="shell-toolbar-left">
        <router-link to="/containers" class="back-link">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="3">
            <polyline points="15 18 9 12 15 6"></polyline>
          </svg>
          <span>Containers</span>
        </router-link>
        <div class="shell-title">
          <h1>{{ containerName || "Container Shell" }}</h1>
          <span class="shell-subtitle">{{ shortId }}</span>
        </div>
      </div>

      <div class="shell-toolbar-right">
        <label class="shell-select-label" for="shell-binary">Shell</label>
        <select
          id="shell-binary"
          v-model="selectedShell"
          class="shell-select"
          :disabled="isConnecting"
          @change="reconnect"
        >
          <option value="/bin/bash">bash</option>
          <option value="/bin/sh">sh</option>
          <option value="/bin/ash">ash</option>
        </select>

        <span :class="['status-pill', statusClass]">{{ statusLabel }}</span>

        <button class="toolbar-btn" type="button" :disabled="isConnecting" @click="reconnect">
          Reconnect
        </button>
      </div>
    </header>

    <div v-if="errorMessage" class="shell-error">{{ errorMessage }}</div>

    <div ref="terminalHost" class="terminal-host" tabindex="0" />
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { createAuthenticatedWebSocket } from "../utils/wsAuth";
import { sharedState, userCanShell } from "../utils/sharedState";
import { useContainers } from "../composables/useContainers";

const route = useRoute();
const router = useRouter();
const { containers, fetchContainers } = useContainers({ autoPoll: false });

const terminalHost = ref(null);
const selectedShell = ref("/bin/bash");
const isConnecting = ref(false);
const isConnected = ref(false);
const sessionEnded = ref(false);
const errorMessage = ref("");

let terminal = null;
let fitAddon = null;
let socket = null;
let resizeObserver = null;
let reconnectTimer = null;
let reconnectFailures = 0;

const containerId = computed(() => String(route.query.c || ""));
const shortId = computed(() => (containerId.value ? containerId.value.slice(0, 12) : "—"));

const container = computed(() =>
  containers.value.find((c) => c.id === containerId.value || c.id.startsWith(containerId.value)),
);

const containerName = computed(() => container.value?.name || "");

const statusClass = computed(() => {
  if (isConnected.value) return "is-connected";
  if (isConnecting.value) return "is-connecting";
  return "is-disconnected";
});

const statusLabel = computed(() => {
  if (isConnected.value) return "Connected";
  if (isConnecting.value) return "Connecting";
  if (sessionEnded.value) return "Session ended";
  return "Disconnected";
});

function disposeTerminal() {
  terminal?.dispose();
  terminal = null;
  fitAddon = null;
}

function initTerminal() {
  disposeTerminal();
  terminal = new Terminal({
    cursorBlink: true,
    fontFamily: "JetBrains Mono, monospace",
    fontSize: 13,
    lineHeight: 1.35,
    theme: {
      background: "#070c12",
      foreground: "#34d399",
      cursor: "#34d399",
      selectionBackground: "rgba(16, 185, 129, 0.35)",
    },
    allowProposedApi: true,
  });
  fitAddon = new FitAddon();
  terminal.loadAddon(fitAddon);
  terminal.open(terminalHost.value);
  fitTerminal();
  terminal.onData((data) => {
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(data);
    }
  });
  terminal.writeln("\x1b[1;36m[DockLog]\x1b[0m Connecting to container shell...");
}

function fitTerminal() {
  if (!terminal || !fitAddon || !terminalHost.value) return;
  fitAddon.fit();
}

function closeSocket() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (socket) {
    socket.onopen = null;
    socket.onmessage = null;
    socket.onerror = null;
    socket.onclose = null;
    if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
      socket.close();
    }
    socket = null;
  }
}

function scheduleReconnect() {
  if (sessionEnded.value || reconnectFailures >= 8) return;
  const delay = Math.min(1000 * 2 ** reconnectFailures, 15000);
  reconnectFailures += 1;
  reconnectTimer = setTimeout(() => {
    connectWebSocket();
  }, delay);
}

function connectWebSocket() {
  if (!containerId.value) {
    errorMessage.value = "No container selected.";
    return;
  }

  if (!userCanShell(sharedState.currentUser)) {
    errorMessage.value = "Shell access is not enabled for your account on this server.";
    return;
  }

  if (container.value && container.value.state !== "running") {
    errorMessage.value = "Container must be running to open a shell.";
    return;
  }

  closeSocket();
  errorMessage.value = "";
  isConnecting.value = true;
  isConnected.value = false;
  sessionEnded.value = false;

  if (!terminal) {
    initTerminal();
  } else {
    terminal.clear();
    terminal.writeln("\x1b[1;36m[DockLog]\x1b[0m Connecting to container shell...");
  }

  const shell = encodeURIComponent(selectedShell.value);
  const path = `/ws/shell/${containerId.value}?shell=${shell}`;

  try {
    socket = createAuthenticatedWebSocket(path);
  } catch (err) {
    isConnecting.value = false;
    errorMessage.value = err?.message || "Failed to open shell connection.";
    return;
  }

  socket.onopen = () => {
    isConnecting.value = false;
    isConnected.value = true;
    reconnectFailures = 0;
    terminal?.writeln("\x1b[1;32m[DockLog]\x1b[0m Shell connected.");
    terminal?.focus();
    fitTerminal();
  };

  socket.onmessage = (event) => {
    if (typeof event.data === "string") {
      terminal?.write(event.data);
      return;
    }
    if (event.data instanceof Blob) {
      event.data.text().then((text) => terminal?.write(text));
      return;
    }
    if (event.data instanceof ArrayBuffer) {
      terminal?.write(new TextDecoder().decode(event.data));
    }
  };

  socket.onerror = () => {
    isConnecting.value = false;
    isConnected.value = false;
    terminal?.writeln("\r\n\x1b[1;31m[DockLog]\x1b[0m Connection error.");
    scheduleReconnect();
  };

  socket.onclose = () => {
    isConnecting.value = false;
    isConnected.value = false;
    sessionEnded.value = true;
    terminal?.writeln("\r\n\x1b[1;33m[DockLog]\x1b[0m Shell session closed.");
  };
}

function reconnect() {
  reconnectFailures = 0;
  sessionEnded.value = false;
  initTerminal();
  connectWebSocket();
}

watch(
  () => sharedState.theme,
  (theme) => {
    if (!terminal) return;
    const isLight = theme === "light";
    terminal.options.theme = {
      background: isLight ? "#f8fafc" : "#070c12",
      foreground: isLight ? "#065f46" : "#34d399",
      cursor: isLight ? "#047857" : "#34d399",
      selectionBackground: isLight ? "rgba(5, 150, 105, 0.25)" : "rgba(16, 185, 129, 0.35)",
    };
  },
);

onMounted(async () => {
  if (!containerId.value) {
    router.replace("/containers");
    return;
  }

  await fetchContainers();

  if (!userCanShell(sharedState.currentUser)) {
    errorMessage.value = "Shell access is not enabled for your account on this server.";
    return;
  }

  initTerminal();
  connectWebSocket();

  resizeObserver = new ResizeObserver(() => fitTerminal());
  if (terminalHost.value) {
    resizeObserver.observe(terminalHost.value);
  }
});

onUnmounted(() => {
  sessionEnded.value = true;
  closeSocket();
  resizeObserver?.disconnect();
  disposeTerminal();
});
</script>

<style scoped>
.shell-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 64px);
  min-height: 480px;
  background: var(--log-bg);
}

.shell-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--border);
  flex-wrap: wrap;
}

.shell-toolbar-left,
.shell-toolbar-right {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 0;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--text-dim);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 600;
}

.back-link:hover {
  color: var(--accent);
}

.shell-title h1 {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0;
}

.shell-subtitle {
  display: block;
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--text-mute);
}

.shell-select-label {
  font-size: 0.75rem;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.shell-select,
.toolbar-btn {
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-main);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
}

.shell-select {
  padding: 0.45rem 0.65rem;
}

.toolbar-btn {
  padding: 0.45rem 0.8rem;
  font-weight: 600;
  cursor: pointer;
}

.toolbar-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.toolbar-btn:not(:disabled):hover {
  border-color: var(--border-active);
  color: var(--accent);
}

.status-pill {
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.35rem 0.55rem;
  border-radius: 999px;
  border: 1px solid var(--border);
}

.status-pill.is-connected {
  color: var(--success);
  border-color: rgba(var(--success-rgb), 0.35);
  background: rgba(var(--success-rgb), 0.08);
}

.status-pill.is-connecting {
  color: var(--warning);
  border-color: rgba(var(--warning-rgb), 0.35);
  background: rgba(var(--warning-rgb), 0.08);
}

.status-pill.is-disconnected {
  color: var(--text-mute);
}

.shell-error {
  margin: 0.75rem 1rem 0;
  padding: 0.75rem 1rem;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(var(--error-rgb), 0.35);
  background: rgba(var(--error-rgb), 0.08);
  color: var(--error);
  font-size: 0.9rem;
}

.terminal-host {
  flex: 1;
  min-height: 0;
  padding: 0.75rem;
}

.terminal-host :deep(.xterm) {
  height: 100%;
}

.terminal-host :deep(.xterm-viewport) {
  border-radius: var(--radius-sm);
}

@media (max-width: 768px) {
  .shell-page {
    height: calc(100vh - 56px);
  }

  .shell-toolbar {
    align-items: flex-start;
  }

  .shell-toolbar-left,
  .shell-toolbar-right {
    width: 100%;
    flex-wrap: wrap;
  }
}
</style>

<template>
  <div class="app-layout">
    <div
      v-if="isMobileMenuOpen"
      class="mobile-overlay"
      @click="isMobileMenuOpen = false"
    ></div>

    <!-- Redesigned Sidebar -->
    <aside
      v-if="!isFullBleedPage || isMobileMenuOpen"
      :class="[
        'main-sidebar',
        { 'mobile-open': isMobileMenuOpen, collapsed: isSidebarCollapsed },
      ]"
    >
      <div class="sidebar-header">
        <router-link to="/dashboard" class="sidebar-logo">
          <img :src="logoSrc" alt="DockLog" class="logo-img-sidebar" />
          <span class="logo-text">DockLog</span>
        </router-link>

        <button
          class="sidebar-toggle-btn desktop-only"
          @click="isSidebarCollapsed = !isSidebarCollapsed"
          aria-label="Toggle sidebar"
        >
          <svg
            viewBox="0 0 24 24"
            width="18"
            height="18"
            fill="none"
            stroke="currentColor"
            stroke-width="3"
          >
            <line
              v-if="!isSidebarCollapsed"
              x1="3"
              y1="12"
              x2="21"
              y2="12"
            ></line>
            <line
              v-if="!isSidebarCollapsed"
              x1="3"
              y1="6"
              x2="21"
              y2="6"
            ></line>
            <line
              v-if="!isSidebarCollapsed"
              x1="3"
              y1="18"
              x2="21"
              y2="18"
            ></line>
            <path v-else d="M9 18l6-6-6-6"></path>
          </svg>
        </button>

        <!-- Mobile Close Button -->
        <button
          class="sidebar-close-btn mobile-only"
          @click="isMobileMenuOpen = false"
          aria-label="Close menu"
        >
          <svg
            viewBox="0 0 24 24"
            width="20"
            height="20"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>


      <nav class="menu-groups">
        <router-link
          to="/dashboard"
          class="nav-link"
          :class="{ active: route.path === '/dashboard' }"
          :data-tooltip="isSidebarCollapsed ? 'Dashboard' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <rect x="3" y="3" width="7" height="7"></rect>
            <rect x="14" y="3" width="7" height="7"></rect>
            <rect x="14" y="14" width="7" height="7"></rect>
            <rect x="3" y="14" width="7" height="7"></rect>
          </svg>
          Dashboard
        </router-link>
        <router-link
          to="/containers"
          class="nav-link"
          :class="{ active: route.path === '/containers' }"
          :data-tooltip="isSidebarCollapsed ? 'Containers' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path
              d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"
            ></path>
          </svg>
          Containers
        </router-link>

        <router-link
          to="/logs"
          class="nav-link"
          :class="{ active: route.path === '/logs' }"
          :data-tooltip="isSidebarCollapsed ? 'Logs' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <line x1="8" y1="6" x2="21" y2="6"></line>
            <line x1="8" y1="12" x2="21" y2="12"></line>
            <line x1="8" y1="18" x2="21" y2="18"></line>
            <line x1="3" y1="6" x2="3.01" y2="6"></line>
            <line x1="3" y1="12" x2="3.01" y2="12"></line>
            <line x1="3" y1="18" x2="3.01" y2="18"></line>
          </svg>
          Logs
        </router-link>
        <router-link
          to="/health"
          class="nav-link"
          :class="{ active: route.path === '/health' }"
          :data-tooltip="isSidebarCollapsed ? 'System Health' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
          </svg>
          System Health
        </router-link>

        <div class="menu-divider" v-if="!sharedState.isAuthDisabled"></div>

        <router-link
          v-if="!sharedState.isAuthDisabled && sharedState.currentUser?.is_admin"
          to="/admin"
          class="nav-link"
          :class="{ active: route.path === '/admin' }"
          :data-tooltip="isSidebarCollapsed ? 'User Management' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <polyline points="16 3.13 16 3.13 16 3.13"></polyline>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
          </svg>
          Users
        </router-link>
        <router-link
          v-if="!sharedState.isAuthDisabled && sharedState.currentUser?.is_admin"
          to="/audit"
          class="nav-link"
          :class="{ active: route.path === '/audit' }"
          :data-tooltip="isSidebarCollapsed ? 'Audit Logs' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
          </svg>
          Audit Logs
        </router-link>

        <div
          class="menu-divider"
          v-if="!sharedState.isAuthDisabled && sharedState.currentUser?.is_admin"
        ></div>

        <router-link
          v-if="!sharedState.isAuthDisabled"
          to="/settings"
          class="nav-link"
          :class="{ active: route.path === '/settings' }"
          :data-tooltip="isSidebarCollapsed ? 'Settings' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <circle cx="12" cy="12" r="3"></circle>
            <path
              d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
            ></path>
          </svg>
          Settings
        </router-link>
      </nav>

      <div class="sidebar-profile" v-if="!sharedState.isAuthDisabled">
        <div class="profile-card">
          <div class="p-avatar-circle">{{ userInitial }}</div>
          <div class="p-info">
            <span class="p-name">{{ sharedState.currentUser?.username }}</span>
            <span class="p-role">{{
              sharedState.currentUser?.is_admin
                ? "ADMINISTRATOR"
                : "STAFF MEMBER"
            }}</span>
          </div>
        </div>
        <div class="logout-button">
          <button class="logout-link" @click="logout">
            <svg
              viewBox="0 0 24 24"
              width="16"
              height="16"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
            >
              <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"></path>
              <polyline points="10 17 15 12 10 7"></polyline>
              <line x1="15" y1="12" x2="3" y2="12"></line>
            </svg>
            Sign Out
          </button>
        </div>
      </div>
    </aside>

    <div class="layout-main-content">
      <header class="main-header glass">
        <div class="header-left">
          <!-- Mobile Menu Trigger -->
          <button
            class="nav-icon-btn mobile-only"
            v-if="!isFullBleedPage"
            @click="isMobileMenuOpen = true"
            aria-label="Open menu"
          >
            <svg
              viewBox="0 0 24 24"
              width="24"
              height="24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <line x1="3" y1="12" x2="21" y2="12"></line>
              <line x1="3" y1="6" x2="21" y2="6"></line>
              <line x1="3" y1="18" x2="21" y2="18"></line>
            </svg>
          </button>

          <router-link
            to="/dashboard"
            class="sidebar-logo logs-route hide-mobile"
            v-if="isFullBleedPage"
          >
            <img :src="logoSrc" alt="DockLog" class="logo-img-sidebar" />
            <span class="logo-text">DockLog</span>
          </router-link>

          <!-- Mobile Logo Brand -->
          <div class="mobile-logo-brand mobile-only">
            <img :src="logoSrc" alt="DockLog" class="m-logo-img" />
            <span class="m-logo-text">DockLog</span>
          </div>

          <div class="title-group desktop-only">
            <h2>{{ route.name || "DockLog" }}</h2>
          </div>
        </div>

        <div class="header-right">
          <div class="system-stats-global desktop-only">
            <div class="h-stat-global" v-if="sharedState.systemStats">
              <span class="h-label">SYS CPU</span>
              <span
                class="h-value"
                :style="{
                  color: getStatColor(sharedState.systemStats.cpu || 0),
                }"
              >
                {{ parseFloat(sharedState.systemStats.cpu || 0).toFixed(2) }}%
                <small v-if="sharedState.systemStats.cores">
                  / {{ sharedState.systemStats.cores }} Core{{
                    sharedState.systemStats.cores > 1 ? "s" : ""
                  }}
                </small>
              </span>
            </div>
            <div class="h-stat-global" v-if="sharedState.systemStats">
              <span class="h-label">SYS MEM</span>
              <span class="h-value">
                {{ formatBytes(sharedState.systemStats.memory || 0) }} / 
                {{ formatBytes(sharedState.systemStats.total_memory || 0) }}
              </span>
            </div>
          </div>

          <button
            class="nav-icon-btn glass"
            @click="toggleTheme"
            :data-tooltip="themeToggleLabel"
            :aria-label="themeToggleLabel"
          >
            <svg
              v-if="sharedState.theme === 'dark'"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <circle cx="12" cy="12" r="5"></circle>
              <line x1="12" y1="1" x2="12" y2="3"></line>
              <line x1="12" y1="21" x2="12" y2="23"></line>
              <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
              <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
              <line x1="1" y1="12" x2="3" y2="12"></line>
              <line x1="21" y1="12" x2="23" y2="12"></line>
              <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
              <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
            </svg>
            <svg
              v-else
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
            </svg>
          </button>

          <button
            class="nav-icon-btn glass mobile-only"
            @click="isMobileSearchOpen = !isMobileSearchOpen"
            aria-label="Search containers"
            data-tooltip="Search containers"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
          </button>

          <!-- Global Search -->
          <div class="search-wrapper glass desktop-only">
            <svg
              class="search-icon"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <input
              type="text"
              v-model="sharedState.searchQuery"
              placeholder="Search containers..."
              class="search-input"
            />
          </div>
        </div>
      </header>

      <div
        v-if="isMobileSearchOpen"
        class="mobile-search-bar mobile-only"
      >
        <div class="search-wrapper glass mobile-search-input">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <input
            type="text"
            v-model="sharedState.searchQuery"
            placeholder="Search containers..."
            class="search-input"
            aria-label="Search containers"
          />
        </div>
      </div>

      <div :class="['layout-body', { 'no-padding': isFullBleedPage }]">
        <slot v-if="!sharedState.forcePasswordChange" />
      </div>
    </div>

    <!-- Global Toast Notification -->
    <Transition name="slide-up">
      <div
        v-if="sharedState.toast.visible"
        :class="['toast-notification', sharedState.toast.type]"
      >
        <div class="toast-icon">
          <svg
            v-if="sharedState.toast.type === 'success'"
            viewBox="0 0 24 24"
            width="22"
            height="22"
            stroke="currentColor"
            stroke-width="2.5"
            fill="none"
          >
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
            <polyline points="22 4 12 14.01 9 11.01"></polyline>
          </svg>
          <svg
            v-else
            viewBox="0 0 24 24"
            width="22"
            height="22"
            stroke="currentColor"
            stroke-width="2.5"
            fill="none"
          >
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="15" y1="9" x2="9" y2="15"></line>
            <line x1="9" y1="9" x2="15" y2="15"></line>
          </svg>
        </div>
        <div class="toast-content">
          <h4>{{ sharedState.toast.title }}</h4>
          <p>{{ sharedState.toast.message }}</p>
        </div>
        <button class="toast-close" @click="sharedState.toast.visible = false">
          <svg
            viewBox="0 0 24 24"
            width="18"
            height="18"
            stroke="currentColor"
            stroke-width="2.5"
            fill="none"
          >
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>
    </Transition>

    <!-- Password Change Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="sharedState.showPasswordModal" class="modal-overlay" role="presentation">
          <div
            :class="[
              'modal-content',
              'shadow-2xl',
              { 'security-modal': sharedState.forcePasswordChange },
            ]"
            role="dialog"
            aria-modal="true"
            aria-labelledby="password-modal-title"
          >
            <div class="modal-icon">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
              </svg>
            </div>
            <h3 id="password-modal-title">
              {{
                sharedState.forcePasswordChange
                  ? "Welcome to DockLog"
                  : "Update Security"
              }}
            </h3>
            <p v-if="sharedState.forcePasswordChange" class="force-text-new">
              For security, please set a new password for your account to
              continue.
            </p>
            <div v-if="!sharedState.forcePasswordChange" class="input-group">
              <input
                type="password"
                v-model="currentPassword"
                placeholder="Current password"
                class="premium-input"
                @input="passwordError = ''"
              />
            </div>
            <div class="input-group">
              <input
                type="password"
                v-model="newPassword"
                placeholder="Enter new password"
                class="premium-input"
                @input="passwordError = ''"
              />
            </div>
            <div class="input-group">
              <input
                type="password"
                v-model="confirmPassword"
                placeholder="Confirm new password"
                class="premium-input"
                @input="passwordError = ''"
              />
              <p v-if="passwordError" class="input-error">
                {{ passwordError }}
              </p>
            </div>
            <div class="modal-actions">
              <button
                v-if="!sharedState.forcePasswordChange"
                @click="sharedState.showPasswordModal = false"
                class="modal-btn cancel"
              >
                Cancel
              </button>
              <button @click="updatePassword" class="modal-btn confirm">
                Update Password
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  sharedState,
  fetchCurrentUser,
  fetchSystemStats,
  showToast,
  formatBytes,
  toggleTheme,
} from "../utils/sharedState";
import { secureStorage } from "../utils/storage";
import { createAuthenticatedWebSocket } from "../utils/wsAuth";
import { apiFetch } from "../utils/apiFetch";

const route = useRoute();
const router = useRouter();
const showUserMenu = ref(false);
const isMobileMenuOpen = ref(false);
const isMobileSearchOpen = ref(false);
const isSidebarCollapsed = ref(false);
const isFullBleedPage = computed(() => route.name === "Logs" || route.name === "Shell");
let statsWs = null;
let userInterval = null;
let statsReconnectTimer = null;
let layoutMounted = false;

const newPassword = ref("");
const confirmPassword = ref("");
const currentPassword = ref("");
const passwordError = ref("");

const userInitial = computed(
  () => sharedState.currentUser?.username?.charAt(0).toUpperCase() || "A",
);

const logoSrc = computed(() =>
  sharedState.theme === "light" ? "/logo-icon-light.png" : "/logo-icon-dark.png",
);

const themeToggleLabel = computed(() =>
  sharedState.theme === "dark" ? "Switch to light mode" : "Switch to dark mode",
);

const getStatColor = (val) => {
  const v = parseFloat(val);
  if (v > 80) return "var(--error)";
  if (v > 50) return "var(--warning)";
  return "var(--accent)";
};

const toggleDashboardSidebar = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value;
};

const logout = () => {
  secureStorage.removeItem("token");
  secureStorage.removeItem("user");
  sharedState.currentUser = null;
  sharedState.showPasswordModal = false;
  sharedState.forcePasswordChange = false;
  router.push("/login");
};

const openPasswordModal = () => {
  newPassword.value = "";
  confirmPassword.value = "";
  currentPassword.value = "";
  passwordError.value = "";
  sharedState.showPasswordModal = true;
  showUserMenu.value = false;
};

const updatePassword = async () => {
  if (newPassword.value.length < 8) {
    passwordError.value = "Password must be at least 8 characters";
    return;
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = "Passwords do not match";
    return;
  }
  if (!sharedState.forcePasswordChange && !currentPassword.value) {
    passwordError.value = "Current password is required";
    return;
  }

  try {
    const token = secureStorage.getItem("token");
    const formData = new FormData();
    formData.append("password", newPassword.value);
    if (!sharedState.forcePasswordChange) {
      formData.append("current_password", currentPassword.value);
    }
    const res = await apiFetch("/api/user/change-password", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
    });
    if (res.ok) {
      sharedState.showPasswordModal = false;
      showToast("Success", "Password updated successfully", "success");
      // If forced, clear the flag
      if (sharedState.forcePasswordChange) {
        sharedState.forcePasswordChange = false;
      }
    } else {
      const data = await res.json();
      passwordError.value = data.error || "Failed to update password";
    }
  } catch (err) {
    passwordError.value = "Connection error";
  }
};

const handleGlobalClick = () => {
  showUserMenu.value = false;
};

watch(
  () => route.path,
  () => {
    isMobileMenuOpen.value = false;
    isMobileSearchOpen.value = false;
  },
);

const initializeLayoutData = async () => {
  if (sharedState.forcePasswordChange) return;

  await fetchSystemStats();

  const closeStatsWs = () => {
    if (statsReconnectTimer) {
      clearTimeout(statsReconnectTimer);
      statsReconnectTimer = null;
    }
    if (statsWs) {
      statsWs.onclose = null;
      statsWs.close();
      statsWs = null;
    }
  };

  const connectSysStats = () => {
    closeStatsWs();
    statsWs = createAuthenticatedWebSocket("/ws/system-stats");

    statsWs.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        sharedState.systemStats = data;
      } catch (e) {}
    };

    statsWs.onclose = () => {
      statsWs = null;
      if (!layoutMounted) return;
      statsReconnectTimer = setTimeout(connectSysStats, 3000);
    };
  };

  connectSysStats();

  if (userInterval) clearInterval(userInterval);
  userInterval = setInterval(async () => {
    const current = await fetchCurrentUser();
    if (current.status === "forbidden" && !sharedState.isAuthDisabled) {
      clearInterval(userInterval);
      router.replace("/login");
    }
  }, 2000);
};

onMounted(async () => {
  layoutMounted = true;
  window.addEventListener("click", handleGlobalClick);
  const session = await fetchCurrentUser();
  if (session.status === "forbidden" && !sharedState.isAuthDisabled) {
    router.replace("/login");
    return;
  }
  sharedState.forcePasswordChange =
    !sharedState.isAuthDisabled && sharedState.currentUser?.password_changed === false;
  sharedState.showPasswordModal =
    !sharedState.isAuthDisabled && sharedState.currentUser?.password_changed === false;

  if (!sharedState.forcePasswordChange) {
    initializeLayoutData();
  }

  window.addEventListener("online", handleOnline);
  window.addEventListener("offline", handleOffline);
});

const handleOnline = () => {
  showToast(
    "Back Online",
    "Your internet connection has been restored",
    "success",
  );
};

const handleOffline = () => {
  showToast("Offline", "You are currently disconnected", "error");
};

watch(
  () => sharedState.currentUser?.password_changed,
  (passwordChanged) => {
    if (passwordChanged === false) {
      sharedState.forcePasswordChange = true;
      sharedState.showPasswordModal = true;
    }
  },
);

watch(
  () => sharedState.forcePasswordChange,
  (forced) => {
    if (!forced) {
      initializeLayoutData();
    } else {
      if (statsWs) {
        statsWs.onclose = null;
        statsWs.close();
        statsWs = null;
      }
      if (userInterval) clearInterval(userInterval);
    }
  }
);

onUnmounted(() => {
  layoutMounted = false;
  window.removeEventListener("click", handleGlobalClick);
  window.removeEventListener("online", handleOnline);
  window.removeEventListener("offline", handleOffline);
  if (statsReconnectTimer) {
    clearTimeout(statsReconnectTimer);
    statsReconnectTimer = null;
  }
  if (statsWs) {
    statsWs.onclose = null;
    statsWs.close();
    statsWs = null;
  }
  if (userInterval) clearInterval(userInterval);
});
</script>

<style scoped>
.modal-content h3 {
  margin-bottom: 0.5rem;
}

.force-text-new {
  text-align: center;
  color: var(--text-mute);
  font-size: 0.9rem;
  margin-bottom: 2rem;
  line-height: 1.5;
}

.input-group {
  margin-bottom: 1.25rem;
}

.security-modal .modal-actions {
  justify-content: center;
  margin-top: 1.5rem;
}
.app-layout {
  height: 100vh;
  display: flex;
  overflow: hidden;
  background: var(--bg-main);
}

.layout-main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.main-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.75rem;
  background: var(--glass-bg);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-bottom: 1px solid var(--border);
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.title-group h2 {
  font-size: 1.05rem;
  font-weight: 800;
  color: var(--text-main);
  letter-spacing: -0.02em;
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.search-wrapper {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.55rem 1rem;
  border-radius: var(--radius-md);
  width: 280px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  transition: width 0.25s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.search-wrapper:focus-within {
  width: 320px;
  border-color: var(--accent);
  background: var(--bg-card);
  box-shadow: 0 0 0 3px rgba(var(--accent-rgb), 0.12);
}

.search-icon {
  width: 18px;
  height: 18px;
  color: var(--text-mute);
}

.search-input {
  background: transparent;
  border: none;
  color: var(--text-main);
  font-size: 0.9rem;
  font-weight: 600;
  width: 100%;
  outline: none;
}

.mobile-search-bar {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border);
  background: var(--glass-bg);
}

.mobile-search-input {
  width: 100%;
}

.layout-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem 1.5rem;
  scrollbar-width: thin;
  background:
    radial-gradient(ellipse 80% 50% at 50% -20%, rgba(var(--accent-rgb), 0.06), transparent 70%),
    var(--bg-main);
}

.layout-body.no-padding {
  padding: 0;
}

/* Redesigned Sidebar Profile Section */
.sidebar-profile {
  margin-top: auto;
  position: relative;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.profile-card {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.75rem 0.85rem;
  width: 100%;
  min-width: 0;
}

.sidebar-profile:hover {
  border-color: rgba(var(--accent-rgb), 0.25);
}

.p-avatar {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  background: linear-gradient(135deg, var(--accent), var(--accent-hover));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  font-weight: 900;
  color: #fff;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(var(--accent-rgb), 0.3);
}

.p-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.p-name {
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  letter-spacing: -0.01em;
}

.p-role {
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--accent);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  opacity: 0.85;
}

.logout-btn {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  color: #ef4444;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.logout-btn:hover {
  background: #ef4444;
  color: #fff;
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.chevron {
  width: 16px;
  height: 16px;
  color: var(--text-mute);
  transition: transform 0.3s cubic-bezier(0.23, 1, 0.32, 1);
}

.chevron.open {
  transform: rotate(180deg);
  color: var(--accent);
}

/* User Menu Overrides for Sidebar Context */
.user-menu {
  position: absolute;
  bottom: calc(100% + 1rem);
  left: 1rem;
  right: 1rem;
  padding: 0.75rem;
  border-radius: 20px;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  border: 1px solid var(--border);
  box-shadow: 0 20px 50px var(--shadow);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 12px;
  color: var(--text-mute);
  font-size: 0.85rem;
  font-weight: 700;
  text-decoration: none;
  transition: all 0.2s;
  background: transparent;
  border: none;
  width: 100%;
  cursor: pointer;
}

.menu-item:hover {
  background: var(--bg-subtle);
  color: var(--accent);
}

.menu-item.logout:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.menu-divider {
  height: 1px;
  background: var(--border);
  margin: 0.1rem 0;
}

/* Transitions */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
}
.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(10px) scale(0.95);
}
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.95);
}

@media (max-width: 1024px) {
  .main-header {
    padding: 0 1.5rem;
  }

  .layout-body {
    padding: 1.5rem;
  }
}

@media (max-width: 480px) {
  .main-header {
    height: 64px;
    padding: 0 1rem;
  }
  .mobile-logo-brand {
    gap: 0.5rem;
  }
  .m-logo-text {
    font-size: 0.95rem;
  }
  .layout-body {
    padding: 1rem;
  }
  .toast-notification {
    top: 1rem;
    right: 1rem;
    left: 1rem;
    min-width: 0;
  }
  .p-name {
    font-size: 0.85rem;
  }
  .p-role {
    font-size: 0.6rem;
  }
}

@media (max-width: 1024px) {
  .main-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    z-index: 2000;
    transform: translateX(-100%);
    opacity: 0;
    visibility: hidden;
    transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
  }

  .main-sidebar.mobile-open {
    transform: translateX(0);
    opacity: 1;
    visibility: visible;
  }

  .header-center {
    display: none;
  }

  .mobile-logo-brand {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    color: var(--text-main);
  }

  .m-logo-icon {
    width: 32px;
    height: 32px;
    background: var(--accent);
    color: #fff;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .m-logo-text {
    font-size: 1.1rem;
    font-weight: 900;
    letter-spacing: -0.02em;
  }

  .nav-icon-btn.mobile-only {
    height: 40px;
    border-radius: 20%;
    align-items: center;
    justify-content: center;
  }

  .sidebar-close-btn {
    width: 38px;
    height: 38px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border);
    color: var(--text-mute);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  }

  .sidebar-close-btn:hover {
    background: rgba(239, 68, 68, 0.1);
    color: var(--error);
    transform: rotate(90deg);
  }

  .sidebar-close-btn:active {
    transform: scale(0.9);
  }
}

img.logo-img-sidebar {
  width: 36px;
  height: 36px;
  object-fit: contain;
  border-radius: var(--radius-sm);
  background: transparent;
  border: none;
  padding: 0;
}

img.m-logo-img {
  width: 32px;
  height: 32px;
  object-fit: contain;
  border-radius: var(--radius-sm);
  background: transparent;
  border: none;
  padding: 0;
}

/* Premium Toast Notifications */
.toast-notification {
  position: fixed;
  top: 1.25rem;
  right: 1.25rem;
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  padding: 0.9rem 1.1rem;
  min-width: 300px;
  max-width: 400px;
  background: var(--glass-bg);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: 0 16px 40px var(--shadow);
  z-index: 9999;
}

.toast-notification.success {
  border-left: 4px solid var(--success);
}

.toast-notification.error {
  border-left: 4px solid var(--error);
}

.toast-icon {
  margin-top: 0.1rem;
  flex-shrink: 0;
}

.toast-notification.success .toast-icon {
  color: var(--success);
}
.toast-notification.error .toast-icon {
  color: var(--error);
}

.toast-content {
  flex: 1;
  min-width: 0;
}

.toast-content h4 {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-main);
}

.toast-content p {
  margin: 0.25rem 0 0;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-mute);
  line-height: 1.5;
}

.toast-close {
  padding: 0.25rem;
  background: transparent;
  border: none;
  color: var(--text-mute);
  cursor: pointer;
  transition: all 0.2s;
  opacity: 0.6;
}

.toast-close:hover {
  opacity: 1;
  color: var(--text-main);
  transform: scale(1.05);
}

/* Toast Transitions */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
}
.slide-up-enter-from {
  opacity: 0;
  transform: translateY(-20px) translateX(20px) scale(0.9);
}
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(-20px) scale(0.9);
}
.system-stats-global {
  display: flex;
  gap: 1.5rem;
  padding-right: 1.5rem;
}

.h-stat-global {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.h-stat-global .h-label {
  font-size: 0.65rem;
  font-weight: 900;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.h-stat-global .h-value {
  font-size: 0.85rem;
  font-weight: 850;
  color: var(--text-main);
  font-family: "JetBrains Mono", monospace;
}

@media (max-width: 1100px) {
  .system-stats-global {
    display: none;
  }
}
</style>

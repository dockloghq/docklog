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
          v-if="dockerEnabled()"
          to="/containers"
          class="nav-link"
          :class="{ active: route.path === '/containers' || route.path.startsWith('/containers/') }"
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
          v-if="swarmEnabled()"
          to="/swarm"
          class="nav-link"
          :class="{ active: route.path === '/swarm' || route.path.startsWith('/swarm/') }"
          :data-tooltip="isSidebarCollapsed ? 'Swarm' : null"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
            <polyline points="2 17 12 22 22 17"></polyline>
            <polyline points="2 12 12 17 22 12"></polyline>
          </svg>
          Swarm
        </router-link>

        <router-link
          v-if="kubernetesEnabled()"
          to="/kubernetes"
          class="nav-link"
          :class="{ active: route.path === '/kubernetes' || route.path.startsWith('/kubernetes/') || route.path === '/pods' }"
          :data-tooltip="isSidebarCollapsed ? 'Kubernetes' : null"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="9"></circle>
            <circle cx="12" cy="12" r="3"></circle>
            <line x1="12" y1="3" x2="12" y2="9"></line>
            <line x1="12" y1="15" x2="12" y2="21"></line>
            <line x1="4.2" y1="7.5" x2="9.4" y2="10.5"></line>
            <line x1="14.6" y1="13.5" x2="19.8" y2="16.5"></line>
            <line x1="4.2" y1="16.5" x2="9.4" y2="13.5"></line>
            <line x1="14.6" y1="10.5" x2="19.8" y2="7.5"></line>
          </svg>
          Kubernetes
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

        <div class="menu-divider" v-if="!sharedState.isAuthDisabled && sharedState.currentUser?.is_admin"></div>

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
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
            <polyline points="14 2 14 8 20 8"></polyline>
            <line x1="16" y1="13" x2="8" y2="13"></line>
            <line x1="16" y1="17" x2="8" y2="17"></line>
            <polyline points="10 9 9 9 8 9"></polyline>
          </svg>
          Audit Logs
        </router-link>

        <router-link
          v-if="!sharedState.isAuthDisabled && sharedState.currentUser?.is_admin"
          to="/notifications"
          class="nav-link"
          :class="{ active: route.path === '/notifications' }"
          :data-tooltip="isSidebarCollapsed ? 'Notifications' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
            <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
          </svg>
          Notifications
        </router-link>

        <router-link
          v-if="!sharedState.isAuthDisabled && sharedState.currentUser?.is_admin"
          to="/alerts"
          class="nav-link"
          :class="{ active: route.path === '/alerts' }"
          :data-tooltip="isSidebarCollapsed ? 'Alerts' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
            <line x1="12" y1="9" x2="12" y2="13"></line>
            <line x1="12" y1="17" x2="12.01" y2="17"></line>
          </svg>
          Alerts
        </router-link>

        <a
          href="https://www.docklog.dev/docs"
          target="_blank"
          rel="noopener noreferrer"
          class="nav-link"
          :data-tooltip="isSidebarCollapsed ? 'Documentation' : null"
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
            <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5V4.5A2.5 2.5 0 0 1 6.5 2z"></path>
          </svg>
          Documentation
        </a>

        <div
          class="menu-divider"
          v-if="!sharedState.isAuthDisabled"
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
              d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
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

          <!-- Global Search (Trigger Button / Search Bar) -->
          <button
            class="search-trigger-btn glass desktop-only"
            @click="openSpotlight"
            aria-label="Open spotlight search"
          >
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
            <span class="search-trigger-placeholder">Search...</span>
            <span class="search-keybinding-badge"><kbd>⌘</kbd><kbd>K</kbd></span>
          </button>
        </div>
      </header>

      <div
        v-if="isMobileSearchOpen"
        class="mobile-search-bar mobile-only"
      >
        <button
          class="search-trigger-btn glass mobile-search-btn"
          @click="openSpotlight"
          aria-label="Search"
        >
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <span class="search-trigger-placeholder">Search...</span>
        </button>
      </div>

      <div :class="['layout-body', { 'no-padding': isFullBleedPage }]">
        <slot v-if="!sharedState.forcePasswordChange" />
      </div>
    </div>

    <!-- Global Toast Notification -->
    <Transition name="toast-slide">
      <div
        v-if="sharedState.toast.visible"
        :class="['toast-notification', sharedState.toast.type || 'info']"
        role="alert"
        aria-live="polite"
      >
        <div
          class="toast-indicator-bar"
          :style="{ animationDuration: `${sharedState.toast.duration || 3500}ms` }"
        ></div>
        <div class="toast-body-row">
          <div class="toast-icon">
            <!-- Info / Copied / Default -->
            <svg
              v-if="sharedState.toast.type === 'info' || !sharedState.toast.type || sharedState.toast.title === 'Copied'"
              viewBox="0 0 24 24"
              width="18"
              height="18"
              stroke="currentColor"
              stroke-width="2.2"
              fill="none"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="16" x2="12" y2="12"></line>
              <line x1="12" y1="8" x2="12.01" y2="8"></line>
            </svg>
            <!-- Success -->
            <svg
              v-else-if="sharedState.toast.type === 'success'"
              viewBox="0 0 24 24"
              width="18"
              height="18"
              stroke="currentColor"
              stroke-width="2.2"
              fill="none"
            >
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
              <polyline points="22 4 12 14.01 9 11.01"></polyline>
            </svg>
            <!-- Warning -->
            <svg
              v-else-if="sharedState.toast.type === 'warning'"
              viewBox="0 0 24 24"
              width="18"
              height="18"
              stroke="currentColor"
              stroke-width="2.2"
              fill="none"
            >
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
              <line x1="12" y1="9" x2="12" y2="13"></line>
              <line x1="12" y1="17" x2="12.01" y2="17"></line>
            </svg>
            <!-- Error -->
            <svg
              v-else
              viewBox="0 0 24 24"
              width="18"
              height="18"
              stroke="currentColor"
              stroke-width="2.2"
              fill="none"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="15" y1="9" x2="9" y2="15"></line>
              <line x1="9" y1="9" x2="15" y2="15"></line>
            </svg>
          </div>
          <div class="toast-content">
            <h4>{{ sharedState.toast.title }}</h4>
            <p v-if="sharedState.toast.message">{{ sharedState.toast.message }}</p>
          </div>
          <button class="toast-close" @click="sharedState.toast.visible = false" aria-label="Close notification">
            <svg
              viewBox="0 0 24 24"
              width="14"
              height="14"
              stroke="currentColor"
              stroke-width="2.4"
              fill="none"
            >
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      </div>
    </Transition>

    <!-- Spotlight Search Modal (Mac Style) -->
    <Teleport to="body">
      <Transition name="spotlight-fade">
        <div
          v-if="isSpotlightOpen"
          class="spotlight-overlay"
          @click.self="closeSpotlight"
          role="presentation"
        >
          <div
            class="spotlight-modal"
            role="dialog"
            aria-modal="true"
            aria-label="Spotlight search"
          >
            <!-- Spotlight Search Input Bar -->
            <div class="spotlight-search-header">
              <svg
                class="spotlight-search-icon"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
              </svg>
              <input
                ref="spotlightInputRef"
                type="text"
                v-model="spotlightQuery"
                @keydown.down.prevent="navigateResults(1)"
                @keydown.up.prevent="navigateResults(-1)"
                @keydown.enter.prevent="selectHighlightedResult"
                @keydown.esc.prevent="closeSpotlight"
                placeholder="Search containers, pods, services, nodes..."
                class="spotlight-input"
              />
              <button
                v-if="spotlightQuery"
                class="spotlight-clear-btn"
                @click="spotlightQuery = ''"
                aria-label="Clear input"
              >
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
              <kbd class="spotlight-esc-badge" @click="closeSpotlight">ESC</kbd>
            </div>

            <!-- Spotlight Content List -->
            <div class="spotlight-body" ref="spotlightBodyRef">
              <!-- Empty state when no query -->
              <div v-if="!spotlightQuery.trim()" class="spotlight-quick-links">
                <div class="spotlight-group-title">QUICK NAVIGATION</div>
                <div
                  v-for="(link, idx) in quickNavigationLinks"
                  :key="link.path"
                  class="spotlight-item"
                  :class="{ 'is-selected': highlightedIndex === idx }"
                  @mouseenter="highlightedIndex = idx"
                  @click="handleQuickNav(link)"
                >
                  <div class="spotlight-item-icon">
                    <!-- Dashboard -->
                    <svg
                      v-if="link.iconType === 'dashboard'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <rect x="3" y="3" width="7" height="7"></rect>
                      <rect x="14" y="3" width="7" height="7"></rect>
                      <rect x="14" y="14" width="7" height="7"></rect>
                      <rect x="3" y="14" width="7" height="7"></rect>
                    </svg>
                    <!-- Containers -->
                    <svg
                      v-else-if="link.iconType === 'containers'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
                    </svg>
                    <!-- Swarm -->
                    <svg
                      v-else-if="link.iconType === 'swarm'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
                      <polyline points="2 17 12 22 22 17"></polyline>
                      <polyline points="2 12 12 17 22 12"></polyline>
                    </svg>
                    <!-- Kubernetes -->
                    <svg
                      v-else-if="link.iconType === 'kubernetes'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <circle cx="12" cy="12" r="9"></circle>
                      <circle cx="12" cy="12" r="3"></circle>
                      <line x1="12" y1="3" x2="12" y2="9"></line>
                      <line x1="12" y1="15" x2="12" y2="21"></line>
                      <line x1="4.2" y1="7.5" x2="9.4" y2="10.5"></line>
                      <line x1="14.6" y1="13.5" x2="19.8" y2="16.5"></line>
                      <line x1="4.2" y1="16.5" x2="9.4" y2="13.5"></line>
                      <line x1="14.6" y1="10.5" x2="19.8" y2="7.5"></line>
                    </svg>
                    <!-- Live Logs -->
                    <svg
                      v-else-if="link.iconType === 'logs'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <line x1="8" y1="6" x2="21" y2="6"></line>
                      <line x1="8" y1="12" x2="21" y2="12"></line>
                      <line x1="8" y1="18" x2="21" y2="18"></line>
                      <line x1="3" y1="6" x2="3.01" y2="6"></line>
                      <line x1="3" y1="12" x2="3.01" y2="12"></line>
                      <line x1="3" y1="18" x2="3.01" y2="18"></line>
                    </svg>
                    <!-- System Health -->
                    <svg
                      v-else-if="link.iconType === 'health'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
                    </svg>
                    <!-- Users -->
                    <svg
                      v-else-if="link.iconType === 'users'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                      <circle cx="9" cy="7" r="4"></circle>
                      <polyline points="16 3.13 16 3.13 16 3.13"></polyline>
                      <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                      <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                    </svg>
                    <!-- Audit Logs -->
                    <svg
                      v-else-if="link.iconType === 'audit'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                      <polyline points="14 2 14 8 20 8"></polyline>
                      <line x1="16" y1="13" x2="8" y2="13"></line>
                      <line x1="16" y1="17" x2="8" y2="17"></line>
                      <polyline points="10 9 9 9 8 9"></polyline>
                    </svg>
                    <!-- Notifications -->
                    <svg
                      v-else-if="link.iconType === 'notifications'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
                      <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
                    </svg>
                    <!-- Alerts -->
                    <svg
                      v-else-if="link.iconType === 'alerts'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                      <line x1="12" y1="9" x2="12" y2="13"></line>
                      <line x1="12" y1="17" x2="12.01" y2="17"></line>
                    </svg>
                    <!-- Documentation -->
                    <svg
                      v-else-if="link.iconType === 'docs'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
                      <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5V4.5A2.5 2.5 0 0 1 6.5 2z"></path>
                    </svg>
                    <!-- Settings -->
                    <svg
                      v-else-if="link.iconType === 'settings'"
                      viewBox="0 0 24 24"
                      width="16"
                      height="16"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                    >
                      <circle cx="12" cy="12" r="3"></circle>
                      <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
                    </svg>
                  </div>
                  <div class="spotlight-item-info">
                    <span class="spotlight-item-name">{{ link.name }}</span>
                    <span class="spotlight-item-meta">{{ link.desc }}</span>
                  </div>
                  <span class="spotlight-item-badge">Page</span>
                </div>
              </div>

              <!-- Matches found (Pages or Workloads) -->
              <div v-else-if="allSpotlightResults.length > 0" class="spotlight-results">
                <!-- Matched Navigation Pages -->
                <template v-if="matchedNavigationPages.length > 0">
                  <div class="spotlight-group-title">
                    PAGES ({{ matchedNavigationPages.length }})
                  </div>
                  <div
                    v-for="(link, pIdx) in matchedNavigationPages"
                    :key="'page-' + link.path"
                    class="spotlight-item"
                    :class="{ 'is-selected': highlightedIndex === pIdx }"
                    @mouseenter="highlightedIndex = pIdx"
                    @click="handleQuickNav(link)"
                  >
                    <div class="spotlight-item-icon">
                      <!-- Dashboard -->
                      <svg
                        v-if="link.iconType === 'dashboard'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <rect x="3" y="3" width="7" height="7"></rect>
                        <rect x="14" y="3" width="7" height="7"></rect>
                        <rect x="14" y="14" width="7" height="7"></rect>
                        <rect x="3" y="14" width="7" height="7"></rect>
                      </svg>
                      <!-- Containers -->
                      <svg
                        v-else-if="link.iconType === 'containers'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
                      </svg>
                      <!-- Swarm -->
                      <svg
                        v-else-if="link.iconType === 'swarm'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
                        <polyline points="2 17 12 22 22 17"></polyline>
                        <polyline points="2 12 12 17 22 12"></polyline>
                      </svg>
                      <!-- Kubernetes -->
                      <svg
                        v-else-if="link.iconType === 'kubernetes'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <circle cx="12" cy="12" r="9"></circle>
                        <circle cx="12" cy="12" r="3"></circle>
                        <line x1="12" y1="3" x2="12" y2="9"></line>
                        <line x1="12" y1="15" x2="12" y2="21"></line>
                        <line x1="4.2" y1="7.5" x2="9.4" y2="10.5"></line>
                        <line x1="14.6" y1="13.5" x2="19.8" y2="16.5"></line>
                        <line x1="4.2" y1="16.5" x2="9.4" y2="13.5"></line>
                        <line x1="14.6" y1="10.5" x2="19.8" y2="7.5"></line>
                      </svg>
                      <!-- Live Logs -->
                      <svg
                        v-else-if="link.iconType === 'logs'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <line x1="8" y1="6" x2="21" y2="6"></line>
                        <line x1="8" y1="12" x2="21" y2="12"></line>
                        <line x1="8" y1="18" x2="21" y2="18"></line>
                        <line x1="3" y1="6" x2="3.01" y2="6"></line>
                        <line x1="3" y1="12" x2="3.01" y2="12"></line>
                        <line x1="3" y1="18" x2="3.01" y2="18"></line>
                      </svg>
                      <!-- System Health -->
                      <svg
                        v-else-if="link.iconType === 'health'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
                      </svg>
                      <!-- Users -->
                      <svg
                        v-else-if="link.iconType === 'users'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                        <circle cx="9" cy="7" r="4"></circle>
                        <polyline points="16 3.13 16 3.13 16 3.13"></polyline>
                        <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                        <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                      </svg>
                      <!-- Audit Logs -->
                      <svg
                        v-else-if="link.iconType === 'audit'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                        <polyline points="14 2 14 8 20 8"></polyline>
                        <line x1="16" y1="13" x2="8" y2="13"></line>
                        <line x1="16" y1="17" x2="8" y2="17"></line>
                        <polyline points="10 9 9 9 8 9"></polyline>
                      </svg>
                      <!-- Notifications -->
                      <svg
                        v-else-if="link.iconType === 'notifications'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
                        <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
                      </svg>
                      <!-- Alerts -->
                      <svg
                        v-else-if="link.iconType === 'alerts'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
                        <line x1="12" y1="9" x2="12" y2="13"></line>
                        <line x1="12" y1="17" x2="12.01" y2="17"></line>
                      </svg>
                      <!-- Documentation -->
                      <svg
                        v-else-if="link.iconType === 'docs'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
                        <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5V4.5A2.5 2.5 0 0 1 6.5 2z"></path>
                      </svg>
                      <!-- Settings -->
                      <svg
                        v-else-if="link.iconType === 'settings'"
                        viewBox="0 0 24 24"
                        width="16"
                        height="16"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2.2"
                      >
                        <circle cx="12" cy="12" r="3"></circle>
                        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
                      </svg>
                    </div>
                    <div class="spotlight-item-info">
                      <span class="spotlight-item-name">{{ link.name }}</span>
                      <span class="spotlight-item-meta">{{ link.desc }}</span>
                    </div>
                    <span class="spotlight-item-badge">Page</span>
                  </div>
                </template>

                <!-- Matched Workloads -->
                <template v-if="spotlightMatches.length > 0">
                  <div class="spotlight-group-title">
                    WORKLOADS ({{ spotlightMatches.length }})
                  </div>
                  <div
                    v-for="(item, wIdx) in spotlightMatches"
                    :key="item.key"
                    class="spotlight-item"
                    :class="{ 'is-selected': highlightedIndex === (matchedNavigationPages.length + wIdx) }"
                    @mouseenter="highlightedIndex = (matchedNavigationPages.length + wIdx)"
                    @click="handleSpotlightSelect(item)"
                  >
                    <!-- Container Icon -->
                    <div
                      v-if="item.type === 'Container'"
                      class="spotlight-item-icon"
                      :class="item.status === 'running' ? 'running' : 'stopped'"
                    >
                      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.2">
                        <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
                      </svg>
                    </div>
                    <!-- Service Icon -->
                    <div
                      v-else-if="item.type === 'Service' || item.type === 'Deployment'"
                      class="spotlight-item-icon"
                      :class="item.status === 'running' ? 'running' : 'stopped'"
                    >
                      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.2">
                        <polygon points="12 2 2 7 12 12 22 7 12 2" />
                        <polyline points="2 17 12 22 22 17" />
                        <polyline points="2 12 12 17 22 12" />
                      </svg>
                    </div>
                    <!-- Node Icon -->
                    <div
                      v-else-if="item.type === 'Node'"
                      class="spotlight-item-icon"
                      :class="item.status === 'running' ? 'running' : 'stopped'"
                    >
                      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.2">
                        <rect x="2" y="3" width="20" height="7" rx="2" ry="2" />
                        <rect x="2" y="14" width="20" height="7" rx="2" ry="2" />
                        <line x1="6" y1="6.5" x2="6.01" y2="6.5" />
                        <line x1="6" y1="17.5" x2="6.01" y2="17.5" />
                      </svg>
                    </div>
                    <!-- Pod Icon -->
                    <div
                      v-else
                      class="spotlight-item-icon"
                      :class="item.status === 'running' ? 'running' : 'stopped'"
                    >
                      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.2">
                        <path d="M12 2l8.5 4.9v9.8L12 21.6 3.5 16.7V6.9L12 2zm0 2.3L5.5 8v7.9l6.5 3.8 6.5-3.8V8L12 4.3z" />
                      </svg>
                    </div>

                    <div class="spotlight-item-info">
                      <div class="spotlight-item-name-row">
                        <span class="spotlight-item-name">{{ item.name }}</span>
                        <span class="spotlight-type-tag">{{ item.type }}</span>
                        <span class="spotlight-platform-tag">{{ item.platform }}</span>
                      </div>
                      <span class="spotlight-item-meta">{{ item.meta }}</span>
                    </div>

                    <span class="spotlight-item-badge" :class="item.status === 'running' ? 'running' : 'stopped'">
                      {{ item.status }}
                    </span>

                    <div class="spotlight-actions" v-if="item.logsRoute">
                      <button
                        class="spotlight-action-btn"
                        title="View Logs (⌘L)"
                        @click.stop="handleSpotlightLogs(item)"
                      >
                        <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2">
                          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                        </svg>
                        <span>Logs</span>
                      </button>
                    </div>
                  </div>
                </template>
              </div>

              <!-- No matches found -->
              <div v-else class="spotlight-no-results">
                <svg viewBox="0 0 24 24" width="36" height="36" fill="none" stroke="currentColor" stroke-width="1.8">
                  <circle cx="11" cy="11" r="8"></circle>
                  <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
                  <line x1="8" y1="11" x2="14" y2="11"></line>
                </svg>
                <p>No results found matching "<strong>{{ spotlightQuery }}</strong>"</p>
                <span>Try searching by page name (Users, Swarm, Health), container, pod, service, or node.</span>
              </div>
            </div>

            <!-- Spotlight Footer with Shortcuts -->
            <div class="spotlight-footer">
              <div class="spotlight-hints">
                <span class="hint-item"><kbd>↑</kbd><kbd>↓</kbd> to navigate</span>
                <span class="hint-item"><kbd>↵</kbd> to open</span>
                <span class="hint-item"><kbd>esc</kbd> to close</span>
              </div>
              <button
                v-if="spotlightQuery.trim()"
                class="spotlight-view-all-btn"
                @click="submitSpotlightToFleet"
              >
                View in Fleet Overview &rarr;
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

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
import BrandIcon from "./BrandIcon.vue";
import {
  sharedState,
  fetchCurrentUser,
  fetchSystemStats,
  fetchLogcConfig,
  fetchGeneralConfig,
  showToast,
  formatBytes,
  toggleTheme,
  dockerEnabled,
  swarmEnabled,
  kubernetesEnabled,
} from "../utils/sharedState";
import { secureStorage } from "../utils/storage";
import { createAuthenticatedWebSocket } from "../utils/wsAuth";
import { apiFetch } from "../utils/apiFetch";
import { useContainers, fetchContainers } from "../composables/useContainers";
import {
  swarmServices,
  swarmNodes,
  fetchSwarmData,
} from "../composables/useSwarm";
import {
  pods,
  deployments,
  services as k8sServicesList,
  fetchKubernetesData,
  fetchNamespaces,
} from "../composables/useKubernetes";

const route = useRoute();
const router = useRouter();
const showUserMenu = ref(false);
const globalSearchInputRef = ref(null);
const spotlightInputRef = ref(null);
const spotlightBodyRef = ref(null);
const isMobileMenuOpen = ref(false);
const isMobileSearchOpen = ref(false);
const isSidebarCollapsed = ref(false);
const isFullBleedPage = computed(() => route.name === "Logs" || route.name === "Shell");

// Spotlight state
const isSpotlightOpen = ref(false);
const spotlightQuery = ref("");
const highlightedIndex = ref(0);

const { containers } = useContainers();

const allNavigationPages = computed(() => {
  const pages = [
    { name: "Dashboard", desc: "Fleet overview & system metrics", path: "/dashboard", iconType: "dashboard", keywords: "home overview fleet stats" },
    { name: "Containers", desc: "Manage Docker containers", path: "/containers", iconType: "containers", keywords: "docker containers list manage" },
  ];

  if (swarmEnabled()) {
    pages.push({
      name: "Docker Swarm",
      desc: "Manage services, tasks and cluster nodes",
      path: "/swarm",
      iconType: "swarm",
      keywords: "swarm services tasks nodes cluster replicas scale",
    });
  }

  if (kubernetesEnabled()) {
    pages.push({
      name: "Kubernetes",
      desc: "Pods, deployments, services and namespaces",
      path: "/kubernetes",
      iconType: "kubernetes",
      keywords: "k8s kubernetes pods deployments services namespaces",
    });
  }

  pages.push(
    { name: "Live Logs", desc: "Real-time streaming container logs", path: "/logs", iconType: "logs", keywords: "logs stream console stdout stderr" },
    { name: "System Health", desc: "Host CPU, memory and resource metrics", path: "/health", iconType: "health", keywords: "health cpu ram memory system metrics load" }
  );

  if (!sharedState.isAuthDisabled && sharedState.currentUser?.is_admin) {
    pages.push(
      { name: "Users", desc: "User accounts, roles & permissions", path: "/admin", iconType: "users", keywords: "users admin accounts rbac permissions staff password" },
      { name: "Audit Logs", desc: "System activity and security audit trail", path: "/audit", iconType: "audit", keywords: "audit logs security trail activity events history" },
      { name: "Notifications", desc: "System alerts and notification channels", path: "/notifications", iconType: "notifications", keywords: "notifications alerts webhooks channels slack discord telegram email" },
      { name: "Alerts", desc: "System thresholds and alert rules", path: "/alerts", iconType: "alerts", keywords: "alerts rules triggers threshold warnings monitoring" }
    );
  }

  pages.push({
    name: "Documentation",
    desc: "Guides and reference documentation",
    path: "https://www.docklog.dev/docs",
    isExternal: true,
    iconType: "docs",
    keywords: "docs documentation help guide api manual",
  });

  if (!sharedState.isAuthDisabled) {
    pages.push({
      name: "Settings",
      desc: "DockLog system & display preferences",
      path: "/settings",
      iconType: "settings",
      keywords: "settings preferences config password display theme appearance",
    });
  }

  return pages;
});

const quickNavigationLinks = computed(() => {
  return allNavigationPages.value.slice(0, 7);
});

// Matched pages when searching
const matchedNavigationPages = computed(() => {
  const query = (spotlightQuery.value || "").trim().toLowerCase();
  if (!query) return [];
  return allNavigationPages.value.filter((p) => {
    return (
      p.name.toLowerCase().includes(query) ||
      p.desc.toLowerCase().includes(query) ||
      p.path.toLowerCase().includes(query) ||
      (p.keywords && p.keywords.toLowerCase().includes(query))
    );
  });
});

// Unified search matches across Containers, Swarm Services/Nodes, K8s Pods/Deployments/Services
const spotlightMatches = computed(() => {
  const query = (spotlightQuery.value || "").trim().toLowerCase();
  if (!query) return [];

  const results = [];

  // 1. Docker Containers
  (containers.value || []).forEach((c) => {
    const name = (c.name || "").replace(/^\//, "").toLowerCase();
    const image = (c.image || "").toLowerCase();
    const id = (c.id || "").toLowerCase();
    if (name.includes(query) || image.includes(query) || id.startsWith(query)) {
      const isRunning = c.state === "running";
      results.push({
        id: c.id,
        key: `container-${c.id}`,
        name: (c.name || "").replace(/^\//, ""),
        meta: c.image || c.id.slice(0, 12),
        type: "Container",
        platform: c.is_swarm ? "Swarm" : "Docker",
        status: isRunning ? "running" : (c.state || "stopped"),
        targetRoute: `/containers/${c.id}`,
        logsRoute: c.is_swarm ? `/logs?c=${c.id}&rt=swarm` : `/logs?c=${c.id}&rt=docker`,
      });
    }
  });

  // 2. Swarm Services & Nodes (if Swarm enabled)
  if (swarmEnabled()) {
    (swarmServices.value || []).forEach((s) => {
      const name = (s.name || "").toLowerCase();
      const image = (s.image || "").toLowerCase();
      const id = (s.id || "").toLowerCase();
      if (name.includes(query) || image.includes(query) || id.startsWith(query)) {
        const running = s.running ?? 0;
        const replicas = s.replicas ?? 1;
        results.push({
          id: s.id,
          key: `swarm-svc-${s.id}`,
          name: s.name,
          meta: s.image || `Replicas: ${running}/${replicas}`,
          type: "Service",
          platform: "Swarm",
          status: running > 0 ? "running" : "stopped",
          targetRoute: `/swarm/services/${s.id}`,
          logsRoute: `/logs?rt=swarm&s=${s.name}`,
        });
      }
    });

    (swarmNodes.value || []).forEach((n) => {
      const name = (n.hostname || n.id || "").toLowerCase();
      const id = (n.id || "").toLowerCase();
      if (name.includes(query) || id.startsWith(query)) {
        const stateStr = (n.state || n.status || "").toLowerCase();
        const availStr = (n.availability || "").toLowerCase();
        const isReady = stateStr === "ready" || (availStr === "active" && stateStr !== "down" && stateStr !== "unknown");
        results.push({
          id: n.id,
          key: `swarm-node-${n.id}`,
          name: n.hostname || n.id.slice(0, 10),
          meta: n.role === "manager" ? "Manager Node" : "Worker Node",
          type: "Node",
          platform: "Swarm",
          status: isReady ? "running" : "stopped",
          targetRoute: "/swarm?tab=nodes",
          logsRoute: "/swarm?tab=nodes",
        });
      }
    });
  }

  // 3. Kubernetes Pods & Deployments (if K8s enabled)
  if (kubernetesEnabled() && sharedState.k8sAvailable) {
    (pods.value || []).forEach((p) => {
      const name = (p.name || "").toLowerCase();
      const ns = (p.namespace || "").toLowerCase();
      if (name.includes(query) || ns.includes(query)) {
        const isRunning = p.phase === "Running";
        results.push({
          id: p.name,
          key: `k8s-pod-${p.namespace}-${p.name}`,
          name: p.name,
          meta: `Namespace: ${p.namespace}`,
          type: "Pod",
          platform: "Kubernetes",
          status: isRunning ? "running" : (p.phase?.toLowerCase() || "pending"),
          targetRoute: `/kubernetes/pods/${p.namespace}/${p.name}`,
          logsRoute: `/logs?p=${encodeURIComponent(p.namespace + '/' + p.name)}&rt=kubernetes`,
        });
      }
    });

    (deployments.value || []).forEach((d) => {
      const name = (d.name || "").toLowerCase();
      const ns = (d.namespace || "").toLowerCase();
      if (name.includes(query) || ns.includes(query)) {
        const ready = d.ready_replicas ?? d.replicas ?? 1;
        results.push({
          id: d.name,
          key: `k8s-deploy-${d.namespace}-${d.name}`,
          name: d.name,
          meta: `Deployment • ${d.namespace}`,
          type: "Deployment",
          platform: "Kubernetes",
          status: ready > 0 ? "running" : "stopped",
          targetRoute: `/kubernetes?tab=deployments&namespace=${d.namespace}`,
          logsRoute: `/kubernetes?tab=deployments&namespace=${d.namespace}`,
        });
      }
    });
  }

  return results.slice(0, 12);
});

// Combined spotlight results: matched pages followed by matched workloads
const allSpotlightResults = computed(() => {
  return [...matchedNavigationPages.value, ...spotlightMatches.value];
});

// Reset highlighted index on query change
watch(
  () => spotlightQuery.value,
  () => {
    highlightedIndex.value = 0;
  }
);

const openSpotlight = () => {
  isSpotlightOpen.value = true;
  spotlightQuery.value = sharedState.searchQuery || "";
  highlightedIndex.value = 0;
  setTimeout(() => {
    spotlightInputRef.value?.focus();
    spotlightInputRef.value?.select();
  }, 50);
};

const closeSpotlight = () => {
  isSpotlightOpen.value = false;
};

const navigateResults = (direction) => {
  const max = spotlightQuery.value.trim()
    ? allSpotlightResults.value.length
    : quickNavigationLinks.value.length;
  if (max === 0) return;
  highlightedIndex.value = (highlightedIndex.value + direction + max) % max;
};

const selectHighlightedResult = () => {
  if (!spotlightQuery.value.trim()) {
    const link = quickNavigationLinks.value[highlightedIndex.value];
    if (link) handleQuickNav(link);
    return;
  }
  const item = allSpotlightResults.value[highlightedIndex.value];
  if (item) {
    if (item.path) {
      handleQuickNav(item);
    } else {
      handleSpotlightSelect(item);
    }
  } else {
    submitSpotlightToFleet();
  }
};

const handleSpotlightSelect = (item) => {
  closeSpotlight();
  if (item.targetRoute) {
    router.push(item.targetRoute);
  }
};

const handleSpotlightLogs = (item) => {
  closeSpotlight();
  if (item.logsRoute) {
    router.push(item.logsRoute);
  }
};

const handleQuickNav = (link) => {
  closeSpotlight();
  if (typeof link === 'string') {
    if (link.startsWith('http')) {
      window.open(link, '_blank', 'noopener,noreferrer');
    } else {
      router.push(link);
    }
    return;
  }
  if (link.isExternal || link.path?.startsWith('http')) {
    window.open(link.path, '_blank', 'noopener,noreferrer');
  } else if (link.path) {
    router.push(link.path);
  }
};

const submitSpotlightToFleet = () => {
  sharedState.searchQuery = spotlightQuery.value;
  closeSpotlight();
  if (route.path !== "/containers" && route.path !== "/dashboard") {
    router.push("/dashboard");
  }
};
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

  fetchLogcConfig();
  fetchGeneralConfig();
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

const handleGlobalSearchKeydown = (e) => {
  // ⌘K or Ctrl+K opens Spotlight search modal
  if ((e.metaKey || e.ctrlKey) && (e.key === "k" || e.key === "K")) {
    e.preventDefault();
    if (isSpotlightOpen.value) {
      closeSpotlight();
    } else {
      openSpotlight();
    }
    return;
  }

  // ⌘F or Ctrl+F opens Spotlight if on any non-log screen
  if ((e.metaKey || e.ctrlKey) && (e.key === "f" || e.key === "F")) {
    if (sharedState.generalSettings?.searchIntercept !== false) {
      // If we are on Logs page and log viewer search exists, let LogViewer handle it
      const logSearch = document.querySelector(".log-search .search-input");
      if (logSearch) {
        return;
      }
      e.preventDefault();
      openSpotlight();
    }
  }

  if (e.key === "Escape" && isSpotlightOpen.value) {
    closeSpotlight();
  }
};

onMounted(async () => {
  layoutMounted = true;
  window.addEventListener("click", handleGlobalClick);
  window.addEventListener("keydown", handleGlobalSearchKeydown);
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
    fetchContainers();
    fetchSwarmData();
    fetchNamespaces();
    fetchKubernetesData();
  }

  window.addEventListener("online", handleOnline);
  window.addEventListener("offline", handleOffline);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleGlobalSearchKeydown);
  window.removeEventListener("click", handleGlobalClick);
  window.removeEventListener("online", handleOnline);
  window.removeEventListener("offline", handleOffline);
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

.search-trigger-btn {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.45rem 0.9rem;
  border-radius: var(--radius-md);
  width: 240px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-mute);
  cursor: pointer;
  text-align: left;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  overflow: hidden;
}

.search-trigger-btn .search-icon {
  width: 18px;
  min-width: 18px;
  max-width: 18px;
  height: 18px;
  min-height: 18px;
  max-height: 18px;
  color: var(--text-mute);
  flex-shrink: 0;
  display: block;
}

.search-trigger-btn:hover {
  border-color: var(--accent);
  background: var(--bg-card);
  color: var(--text-main);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.search-trigger-placeholder {
  flex: 1;
  font-size: 0.86rem;
  font-weight: 500;
  color: var(--text-mute);
}

.search-keybinding-badge {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.search-keybinding-badge kbd {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  font-size: 0.65rem;
  font-family: inherit;
  font-weight: 700;
  color: var(--text-mute);
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid var(--border);
  border-radius: 4px;
}

.mobile-search-btn {
  width: 100%;
}

/* ==========================================================
   MAC SPOTLIGHT MODAL STYLING
   ========================================================== */
.spotlight-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  z-index: 9999;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 12vh;
}

.spotlight-modal {
  width: 100%;
  max-width: 640px;
  background: var(--bg-card);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 18px;
  box-shadow: 0 25px 60px -10px rgba(0, 0, 0, 0.6), 0 0 0 1px rgba(255, 255, 255, 0.08);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: spotlightPop 0.18s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes spotlightPop {
  from {
    opacity: 0;
    transform: scale(0.97) translateY(-8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.spotlight-fade-enter-active,
.spotlight-fade-leave-active {
  transition: opacity 0.15s ease;
}

.spotlight-fade-enter-from,
.spotlight-fade-leave-to {
  opacity: 0;
}

.spotlight-search-header {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border);
  background: var(--bg-subtle);
}

.spotlight-search-icon {
  width: 22px;
  height: 22px;
  color: var(--accent);
  flex-shrink: 0;
}

.spotlight-input {
  flex: 1;
  background: transparent;
  border: none;
  font-size: 1.15rem;
  font-weight: 500;
  color: var(--text-main);
  outline: none;
}

.spotlight-input::placeholder {
  color: var(--text-mute);
  opacity: 0.7;
}

.spotlight-clear-btn {
  background: transparent;
  border: none;
  color: var(--text-mute);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px;
  border-radius: 6px;
  transition: all 0.15s ease;
}

.spotlight-clear-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-main);
}

.spotlight-esc-badge {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 2px 6px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid var(--border);
  border-radius: 5px;
  color: var(--text-mute);
  cursor: pointer;
}

.spotlight-esc-badge:hover {
  color: var(--text-main);
  border-color: var(--accent);
}

.spotlight-body {
  max-height: 420px;
  overflow-y: auto;
  padding: 0.5rem;
  scrollbar-width: thin;
}

.spotlight-group-title {
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  color: var(--text-mute);
  padding: 0.6rem 0.85rem 0.35rem 0.85rem;
  text-transform: uppercase;
}

.spotlight-item {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.65rem 0.85rem;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.12s ease;
  user-select: none;
}

.spotlight-item.is-selected {
  background: rgba(var(--accent-rgb), 0.15);
}

.spotlight-item:hover {
  background: rgba(var(--accent-rgb), 0.12);
}

.spotlight-item-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-mute);
  flex-shrink: 0;
}

.spotlight-item-icon.running {
  background: rgba(34, 197, 94, 0.16);
  color: #22c55e;
}

.spotlight-item-icon.stopped {
  background: rgba(239, 68, 68, 0.14);
  color: #ef4444;
}

.spotlight-item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.spotlight-item-name-row {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.spotlight-item-name {
  font-size: 0.92rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.spotlight-type-tag {
  font-size: 0.62rem;
  font-weight: 700;
  text-transform: uppercase;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-mute);
  letter-spacing: 0.03em;
}

.spotlight-platform-tag {
  font-size: 0.62rem;
  font-weight: 600;
  color: var(--accent);
}

.spotlight-item-meta {
  font-size: 0.74rem;
  color: var(--text-mute);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: monospace;
}

.spotlight-item-badge {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-mute);
}

.spotlight-item-badge.running {
  background: rgba(34, 197, 94, 0.16);
  color: #22c55e;
}

.spotlight-item-badge.stopped {
  background: rgba(239, 68, 68, 0.14);
  color: #ef4444;
}

.spotlight-actions {
  display: flex;
  align-items: center;
  margin-left: 0.4rem;
}

.spotlight-action-btn {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.3rem 0.65rem;
  border-radius: 7px;
  border: 1px solid var(--border);
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-mute);
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.spotlight-action-btn:hover {
  background: var(--accent);
  color: #fff;
  border-color: var(--accent);
}

.spotlight-no-results {
  padding: 3rem 1.5rem;
  text-align: center;
  color: var(--text-mute);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.spotlight-no-results p {
  margin: 0;
  font-size: 0.95rem;
  color: var(--text-main);
}

.spotlight-no-results span {
  font-size: 0.8rem;
  color: var(--text-mute);
}

.spotlight-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.65rem 1rem;
  border-top: 1px solid var(--border);
  background: var(--bg-subtle);
  font-size: 0.74rem;
}

.spotlight-hints {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  color: var(--text-mute);
}

.hint-item {
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.hint-item kbd {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 1px 5px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid var(--border);
  border-radius: 4px;
  color: var(--text-main);
}

.spotlight-view-all-btn {
  background: transparent;
  border: none;
  color: var(--accent);
  font-weight: 700;
  font-size: 0.76rem;
  cursor: pointer;
  transition: opacity 0.15s ease;
}

.spotlight-view-all-btn:hover {
  text-decoration: underline;
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
    bottom: 1.25rem;
    right: 1.25rem;
    left: 1.25rem;
    min-width: 0;
    max-width: none;
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

/* Minimalist Modern Toast Notifications (bottom-right card with accent line) */
.toast-notification {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  display: flex;
  flex-direction: column;
  min-width: 280px;
  max-width: 380px;
  background: #111418;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.55), 0 0 1px rgba(255, 255, 255, 0.1);
  overflow: hidden;
  z-index: 9999;
}

[data-theme="light"] .toast-notification {
  background: #ffffff;
  border: 1px solid rgba(0, 0, 0, 0.08);
  box-shadow: 0 16px 36px rgba(15, 23, 42, 0.16), 0 0 1px rgba(0, 0, 0, 0.08);
}

.toast-indicator-bar {
  height: 3px;
  width: 100%;
  background: #06b6d4;
  border-radius: 3px 3px 0 0;
  margin: 0;
  transform-origin: left;
  animation: toastProgress linear forwards;
}

@keyframes toastProgress {
  from {
    transform: scaleX(1);
  }
  to {
    transform: scaleX(0);
  }
}

.toast-notification:hover .toast-indicator-bar {
  animation-play-state: paused;
}

.toast-notification.success .toast-indicator-bar {
  background: #10b981;
}

.toast-notification.info .toast-indicator-bar {
  background: #06b6d4;
}

.toast-notification.warning .toast-indicator-bar {
  background: #f59e0b;
}

.toast-notification.error .toast-indicator-bar {
  background: #ef4444;
}

.toast-body-row {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.85rem 1rem 0.95rem;
}

.toast-icon {
  margin-top: 0.15rem;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.toast-notification.info .toast-icon,
.toast-notification .toast-icon {
  color: #06b6d4;
}

.toast-notification.success .toast-icon {
  color: #10b981;
}

.toast-notification.warning .toast-icon {
  color: #f59e0b;
}

.toast-notification.error .toast-icon {
  color: #ef4444;
}

.toast-content {
  flex: 1;
  min-width: 0;
}

.toast-content h4 {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 700;
  color: #f8fafc;
  letter-spacing: -0.01em;
  line-height: 1.35;
}

[data-theme="light"] .toast-content h4 {
  color: #0f172a;
}

.toast-content p {
  margin: 0.2rem 0 0;
  font-size: 0.8rem;
  font-weight: 500;
  color: #94a3b8;
  line-height: 1.4;
}

[data-theme="light"] .toast-content p {
  color: #64748b;
}

.toast-close {
  padding: 0.2rem;
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.15s ease;
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.toast-close:hover {
  color: #f8fafc;
  background: rgba(255, 255, 255, 0.08);
}

[data-theme="light"] .toast-close:hover {
  color: #0f172a;
  background: rgba(0, 0, 0, 0.05);
}

/* Toast Transitions: Bottom Slide + Fade */
.toast-slide-enter-active,
.toast-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.toast-slide-enter-from {
  opacity: 0;
  transform: translateY(16px) scale(0.96);
}

.toast-slide-leave-to {
  opacity: 0;
  transform: translateY(12px) scale(0.96);
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

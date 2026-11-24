<template>
  <div class="settings-page animate-fade-in">
    <!-- Breadcrumb & Top Bar -->
    <header class="settings-top-bar">
      <div class="settings-title-group">
        <div class="breadcrumb">
          <span>Platform</span>
          <span class="breadcrumb-separator">/</span>
          <span class="active-crumb">Preferences &amp; System</span>
        </div>
        <h1 class="settings-title">Settings</h1>
        <p class="settings-subtitle">{{ headerSubtitle }}</p>
      </div>

      <div class="top-bar-actions">
        <button
          type="button"
          class="reset-btn"
          @click="resetToDefaults"
          title="Reset personal preferences to default"
        >
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5">
            <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"></path>
            <path d="M3 3v5h5"></path>
          </svg>
          <span>Reset to defaults</span>
        </button>
      </div>
    </header>

    <!-- Main Container -->
    <div class="settings-layout">
      <!-- Sidebar Navigation -->
      <aside class="settings-sidebar">
        <div class="sidebar-category-label">PERSONAL PREFERENCES</div>
        <button
          type="button"
          class="nav-tab-btn"
          :class="{ active: currentTab === 'general' }"
          @click="selectTab('general')"
        >
          <span class="tab-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3"></circle>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
            </svg>
          </span>
          <div class="tab-label-group">
            <span class="tab-title">General</span>
            <span class="tab-desc">Theme, language &amp; shortcuts</span>
          </div>
        </button>

        <button
          type="button"
          class="nav-tab-btn"
          :class="{ active: currentTab === 'logs' }"
          @click="selectTab('logs')"
        >
          <span class="tab-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="4" y1="6" x2="20" y2="6"></line>
              <line x1="4" y1="12" x2="14" y2="12"></line>
              <line x1="4" y1="18" x2="18" y2="18"></line>
            </svg>
          </span>
          <div class="tab-label-group">
            <span class="tab-title">Logs Stream</span>
            <span class="tab-desc">Font, wrap, time format</span>
          </div>
        </button>

        <div class="sidebar-category-label mt-5">SYSTEM &amp; INSTANCE</div>
        <button
          type="button"
          class="nav-tab-btn"
          :class="{ active: currentTab === 'security' }"
          @click="selectTab('security')"
        >
          <span class="tab-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
            </svg>
          </span>
          <div class="tab-label-group">
            <span class="tab-title">Security &amp; Auth</span>
            <span class="tab-desc">Permissions, shell, password</span>
          </div>
        </button>

        <div class="sidebar-category-label mt-5">SYSTEM INFO</div>
        <button
          type="button"
          class="nav-tab-btn"
          :class="{ active: currentTab === 'about' }"
          @click="selectTab('about')"
        >
          <span class="tab-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="16" x2="12" y2="12"></line>
              <line x1="12" y1="8" x2="12.01" y2="8"></line>
            </svg>
          </span>
          <div class="tab-label-group">
            <span class="tab-title">About DockLog</span>
            <span class="tab-desc">Version, Swarm &amp; K8s status</span>
          </div>
        </button>
      </aside>

      <!-- Main Content Area -->
      <main class="settings-content">
        <!-- ================= TAB: GENERAL ================= -->
        <section v-if="currentTab === 'general'" class="settings-panel animate-fade-in">
          <div class="panel-header">
            <div>
              <h2 class="panel-title">General Preferences</h2>
              <p class="panel-subtitle">Personalize your visual theme, language, and workspace interaction habits.</p>
            </div>
          </div>

          <!-- Section: Appearance -->
          <div class="config-block">
            <div class="config-block-header">
              <span class="config-kicker">APPEARANCE</span>
            </div>

            <!-- Visual Theme Cards -->
            <div class="config-row-stacked">
              <div class="setting-text-col">
                <span class="setting-name">Interface Theme</span>
                <span class="setting-caption">Select light, dark, or follow your operating system appearance.</span>
              </div>
              <div class="theme-card-picker">
                <button
                  type="button"
                  class="theme-card"
                  :class="{ selected: sharedState.themePreference === 'light' }"
                  @click="applyTheme('light')"
                >
                  <div class="theme-preview-box preview-light">
                    <div class="tp-header"></div>
                    <div class="tp-lines">
                      <div class="tp-line" style="width: 70%; background: #0284c7;"></div>
                      <div class="tp-line" style="width: 45%; background: #94a3b8;"></div>
                    </div>
                  </div>
                  <span class="theme-card-label">Light</span>
                </button>

                <button
                  type="button"
                  class="theme-card"
                  :class="{ selected: sharedState.themePreference === 'dark' }"
                  @click="applyTheme('dark')"
                >
                  <div class="theme-preview-box preview-dark">
                    <div class="tp-header"></div>
                    <div class="tp-lines">
                      <div class="tp-line" style="width: 70%; background: #38bdf8;"></div>
                      <div class="tp-line" style="width: 45%; background: #64748b;"></div>
                    </div>
                  </div>
                  <span class="theme-card-label">Dark</span>
                </button>

                <button
                  type="button"
                  class="theme-card"
                  :class="{ selected: sharedState.themePreference === 'auto' }"
                  @click="applyTheme('auto')"
                >
                  <div class="theme-preview-box preview-auto">
                    <div class="tp-split-left">
                      <div class="tp-line" style="width: 80%; background: #0284c7;"></div>
                    </div>
                    <div class="tp-split-right">
                      <div class="tp-line" style="width: 80%; background: #38bdf8;"></div>
                    </div>
                  </div>
                  <span class="theme-card-label">Auto (System)</span>
                </button>
              </div>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Language</span>
                <span class="setting-caption">Interface language. More locales available in upcoming releases.</span>
              </div>
              <div class="custom-select-wrap">
                <select v-model="generalPrefs.language" @change="saveGeneralPrefs">
                  <option value="auto">Auto (Browser default)</option>
                  <option value="en">English (US)</option>
                  <option value="es">Español</option>
                  <option value="de">Deutsch</option>
                  <option value="fr">Français</option>
                </select>
                <div class="select-chevron">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5">
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <!-- Section: Behavior -->
          <div class="config-block">
            <div class="config-block-header">
              <span class="config-kicker">BEHAVIOR &amp; WORKFLOW</span>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Follow recreated containers</span>
                <span class="setting-caption">When a container restarts or is redeployed with a new ID, seamlessly transition log view.</span>
              </div>
              <div class="custom-select-wrap">
                <select v-model="generalPrefs.followRecreated" @change="saveGeneralPrefs">
                  <option value="delayed">Delayed (Smooth)</option>
                  <option value="immediate">Immediate</option>
                  <option value="never">Do not follow</option>
                </select>
                <div class="select-chevron">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5">
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </div>
              </div>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <div class="inline-title-kbd">
                  <span class="setting-name">Search logs with</span>
                  <span class="kbd-badge"><span class="cmd-symbol">&#8984;</span> / Ctrl + F</span>
                </div>
                <span class="setting-caption">Intercept the default browser search dialog with DockLog's built-in log filter.</span>
              </div>
              <label class="dock-switch">
                <input
                  type="checkbox"
                  v-model="generalPrefs.searchIntercept"
                  @change="saveGeneralPrefs"
                />
                <span class="dock-switch-slider"></span>
              </label>
            </div>
          </div>
        </section>

        <!-- ================= TAB: LOGS ================= -->
        <section v-else-if="currentTab === 'logs'" class="settings-panel animate-fade-in">
          <div class="panel-header">
            <div>
              <h2 class="panel-title">Log Stream Customization</h2>
              <p class="panel-subtitle">Configure real-time terminal font, row spacing, line wraps, and error highlighting.</p>
            </div>
          </div>

          <!-- Interactive Live Preview Box -->
          <div class="log-preview-envelope">
            <div class="preview-toolbar">
              <span class="preview-pill">Interactive Live Preview</span>
              <span class="preview-hint">Settings are applied reactively</span>
            </div>
            <div
              class="live-log-console"
              :class="[
                `font-${logPrefs.fontSize}`,
                {
                  'compact-mode': logPrefs.compactMode,
                  'no-soft-wrap': !logPrefs.softWrap,
                }
              ]"
            >
              <div class="preview-line">
                <span v-if="logPrefs.showLineNumbers" class="line-number-preview">1</span>
                <span v-if="logPrefs.showStreamLabels" class="stream-tag stdout">STDOUT</span>
                <span v-if="logPrefs.showTimestamps" class="preview-time">{{ formatSampleTime(sampleDate1) }}</span>
                <span class="preview-dot normal"></span>
                <span class="preview-text">DockLog daemon connection established (client: interactive-web)</span>
              </div>
              <div class="preview-line">
                <span v-if="logPrefs.showLineNumbers" class="line-number-preview">2</span>
                <span v-if="logPrefs.showStreamLabels" class="stream-tag stdout">STDOUT</span>
                <span v-if="logPrefs.showTimestamps" class="preview-time">{{ formatSampleTime(sampleDate2) }}</span>
                <span class="preview-dot warn"></span>
                <span class="preview-text"><span class="text-warning">[WARN] Container buffer memory utilization elevated (74%)</span></span>
              </div>
              <div
                class="preview-line multiline-error"
                :class="{ 'is-highlighted-error': logPrefs.highlightErrors }"
              >
                <span v-if="logPrefs.showLineNumbers" class="line-number-preview">3</span>
                <span v-if="logPrefs.showStreamLabels" class="stream-tag stderr">STDERR</span>
                <span v-if="logPrefs.showTimestamps" class="preview-time">{{ formatSampleTime(sampleDate3) }}</span>
                <span class="preview-dot error"></span>
                <span class="preview-text">
                  <span class="text-error">[ERROR] Handshake timeout: remote peer aborted connection</span><br />
                  &nbsp;&nbsp;at socket.stream.connect (network/socket.go:142)<br />
                  &nbsp;&nbsp;with retries exhausted (3 attempts).
                </span>
              </div>
              <div class="preview-line">
                <span v-if="logPrefs.showLineNumbers" class="line-number-preview">4</span>
                <span v-if="logPrefs.showStreamLabels" class="stream-tag stdout">STDOUT</span>
                <span v-if="logPrefs.showTimestamps" class="preview-time">{{ formatSampleTime(sampleDate4) }}</span>
                <span class="preview-dot normal"></span>
                <span class="preview-text">{"level":"info","event":"heartbeat","latency_ms":1.42,"replica":"web.1"}</span>
              </div>
              <div class="preview-line">
                <span v-if="logPrefs.showLineNumbers" class="line-number-preview">5</span>
                <span v-if="logPrefs.showStreamLabels" class="stream-tag stdout">STDOUT</span>
                <span v-if="logPrefs.showTimestamps" class="preview-time">{{ formatSampleTime(sampleDate5) }}</span>
                <span class="preview-dot purple"></span>
                <span class="preview-text">
                  [SYSTEM] Soft wrap preview message with long payload string. When soft wrap is active, long outputs wrap naturally without requiring tedious horizontal navigation.
                </span>
              </div>
            </div>
          </div>

          <!-- Controls List -->
          <div class="config-block">
            <div class="config-block-header">
              <span class="config-kicker">TYPOGRAPHY &amp; SPACING</span>
            </div>

            <!-- Font Size -->
            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Font size</span>
                <span class="setting-caption">Scale log terminal font for optimal readability.</span>
              </div>
              <div class="dock-segmented">
                <button
                  type="button"
                  class="dock-seg-btn"
                  :class="{ active: logPrefs.fontSize === 'small' }"
                  @click="updateLogPref('fontSize', 'small')"
                >
                  Small
                </button>
                <button
                  type="button"
                  class="dock-seg-btn"
                  :class="{ active: logPrefs.fontSize === 'medium' }"
                  @click="updateLogPref('fontSize', 'medium')"
                >
                  Medium
                </button>
                <button
                  type="button"
                  class="dock-seg-btn"
                  :class="{ active: logPrefs.fontSize === 'large' }"
                  @click="updateLogPref('fontSize', 'large')"
                >
                  Large
                </button>
              </div>
            </div>

            <!-- Compact Mode -->
            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Compact mode</span>
                <span class="setting-caption">Tighter vertical line height so more log lines fit into view.</span>
              </div>
              <label class="dock-switch">
                <input
                  type="checkbox"
                  v-model="logPrefs.compactMode"
                  @change="updateLogPref('compactMode', logPrefs.compactMode)"
                />
                <span class="dock-switch-slider"></span>
              </label>
            </div>

            <!-- Show Timestamps -->
            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Show timestamps</span>
                <span class="setting-caption">Prefix each log line with the RFC timestamp recorded by the daemon.</span>
              </div>
              <label class="dock-switch">
                <input
                  type="checkbox"
                  v-model="logPrefs.showTimestamps"
                  @change="updateLogPref('showTimestamps', logPrefs.showTimestamps)"
                />
                <span class="dock-switch-slider"></span>
              </label>
            </div>

            <!-- Soft Wrap -->
            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Soft wrap</span>
                <span class="setting-caption">Wrap wide terminal messages onto new lines rather than overflowing horizontally.</span>
              </div>
              <label class="dock-switch">
                <input
                  type="checkbox"
                  v-model="logPrefs.softWrap"
                  @change="updateLogPref('softWrap', logPrefs.softWrap)"
                />
                <span class="dock-switch-slider"></span>
              </label>
            </div>

            <!-- Highlight Error Rows -->
            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Highlight error rows</span>
                <span class="setting-caption">Apply distinct colored background and indicator bars to fatal/error rows.</span>
              </div>
              <label class="dock-switch">
                <input
                  type="checkbox"
                  v-model="logPrefs.highlightErrors"
                  @change="updateLogPref('highlightErrors', logPrefs.highlightErrors)"
                />
                <span class="dock-switch-slider"></span>
              </label>
            </div>

            <!-- Time Format -->
            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Time format</span>
                <span class="setting-caption">Format timestamps using 12-hour AM/PM notation or standard 24-hour.</span>
              </div>
              <div class="custom-select-wrap">
                <select v-model="logPrefs.timeFormat" @change="updateLogPref('timeFormat', logPrefs.timeFormat)">
                  <option value="auto">Auto (24-hour)</option>
                  <option value="12-hour">12-hour (hh:mm:ss AM/PM)</option>
                  <option value="24-hour">24-hour (HH:mm:ss)</option>
                </select>
                <div class="select-chevron">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5">
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </div>
              </div>
            </div>

            <!-- Date Format -->
            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Date format</span>
                <span class="setting-caption">Customize year, month, and day ordering for log timestamps.</span>
              </div>
              <div class="custom-select-wrap">
                <select v-model="logPrefs.dateFormat" @change="updateLogPref('dateFormat', logPrefs.dateFormat)">
                  <option value="YYYY-MM-DD">YYYY-MM-DD (ISO / Default)</option>
                  <option value="YYYY/MM/DD">YYYY/MM/DD</option>
                  <option value="YYYY/DD/MM">YYYY/DD/MM</option>
                  <option value="DD/MM/YYYY">DD/MM/YYYY (UK / EU / IN)</option>
                  <option value="DD-MM-YYYY">DD-MM-YYYY</option>
                  <option value="MM/DD/YYYY">MM/DD/YYYY (US)</option>
                </select>
                <div class="select-chevron">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5">
                    <polyline points="6 9 12 15 18 9"></polyline>
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <!-- Collapsible: Advanced Display in Logs Stream -->
          <div class="config-block">
            <button
              type="button"
              class="accordion-toggle"
              @click="showAdvancedLogs = !showAdvancedLogs"
            >
              <div class="flex items-center gap-2">
                <span class="config-kicker">ADVANCED STREAM CONTROLS</span>
                <span class="badge badge-dim text-xs">3 options</span>
              </div>
              <svg
                viewBox="0 0 24 24"
                width="16"
                height="16"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
                class="transition-transform duration-200"
                :style="{ transform: showAdvancedLogs ? 'rotate(180deg)' : 'rotate(0)' }"
              >
                <polyline points="6 9 12 15 18 9"></polyline>
              </svg>
            </button>

            <div v-if="showAdvancedLogs" class="accordion-body">
              <div class="config-row">
                <div class="setting-text-col">
                  <span class="setting-name">Show stdout and stderr labels</span>
                  <span class="setting-caption">Explicitly tag each log entry with its stream origin.</span>
                </div>
                <label class="dock-switch">
                  <input
                    type="checkbox"
                    v-model="logPrefs.showStreamLabels"
                    @change="updateLogPref('showStreamLabels', logPrefs.showStreamLabels)"
                  />
                  <span class="dock-switch-slider"></span>
                </label>
              </div>

              <div class="config-row">
                <div class="setting-text-col">
                  <span class="setting-name">Show line numbers</span>
                  <span class="setting-caption">Display line index gutter for referencing specific log rows.</span>
                </div>
                <label class="dock-switch">
                  <input
                    type="checkbox"
                    v-model="logPrefs.showLineNumbers"
                    @change="updateLogPref('showLineNumbers', logPrefs.showLineNumbers)"
                  />
                  <span class="dock-switch-slider"></span>
                </label>
              </div>

              <div class="config-row">
                <div class="setting-text-col">
                  <span class="setting-name">Auto-scroll behavior</span>
                  <span class="setting-caption">How new incoming live logs interact with your terminal scroll position.</span>
                </div>
                <div class="custom-select-wrap">
                  <select v-model="logPrefs.autoScrollMode" @change="updateLogPref('autoScrollMode', logPrefs.autoScrollMode)">
                    <option value="smart">Smart Resume (pauses when inspecting, resumes at bottom)</option>
                    <option value="always">Always Stick to Bottom</option>
                    <option value="manual">Manual Only</option>
                  </select>
                  <div class="select-chevron">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5">
                      <polyline points="6 9 12 15 18 9"></polyline>
                    </svg>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- ================= TAB: SECURITY ================= -->
        <section v-else-if="currentTab === 'security'" class="settings-panel animate-fade-in">
          <div class="panel-header">
            <div>
              <h2 class="panel-title">Security &amp; Access Controls</h2>
              <p class="panel-subtitle">Instance authentication status, container execution permissions, and account password updates.</p>
            </div>
          </div>

          <!-- Auth Status Banner -->
          <div class="status-banner" :class="sharedState.isAuthDisabled ? 'banner-warn' : 'banner-success'">
            <div class="banner-icon-col">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="16" x2="12" y2="12"></line>
                <line x1="12" y1="8" x2="12.01" y2="8"></line>
              </svg>
            </div>
            <div class="banner-body">
              <span class="banner-title">
                {{ sharedState.isAuthDisabled ? "Authentication Disabled (Open Mode)" : "Role-Based Authentication Active" }}
              </span>
              <span class="banner-desc">
                {{ sharedState.isAuthDisabled
                  ? "Without authentication enabled, any network client with access to this port can view container logs."
                  : "Signed JWT sessions with granular container permissions are strictly enforced." }}
              </span>
            </div>
          </div>

          <!-- Authentication Overview -->
          <div class="config-block">
            <div class="config-block-header">
              <span class="config-kicker">AUTHENTICATION MODE</span>
            </div>

            <div class="config-row">
              <div class="flex items-center gap-3">
                <div class="shield-badge" :class="sharedState.isAuthDisabled ? 'open' : 'locked'">
                  <svg v-if="sharedState.isAuthDisabled" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                    <path d="M7 11V7a5 5 0 0 1 9.9-1"></path>
                  </svg>
                  <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                    <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                  </svg>
                </div>
                <div class="setting-text-col">
                  <span class="setting-name">{{ sharedState.isAuthDisabled ? "No Authentication" : "Password / JWT Session" }}</span>
                  <span class="setting-caption">Provider: {{ sharedState.isAuthDisabled ? "none" : "built-in bcrypt / HMAC-SHA256" }}</span>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <span class="status-pill" :class="sharedState.isAuthDisabled ? 'pill-warn' : 'pill-success'">
                  {{ sharedState.isAuthDisabled ? "No login" : "Enforced" }}
                </span>
                <button
                  v-if="!sharedState.isAuthDisabled"
                  type="button"
                  class="action-sub-btn"
                  @click="openChangePasswordModal"
                >
                  <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                    <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                  </svg>
                  <span>Update password</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Actions & Shell Controls -->
          <div class="config-block">
            <div class="config-block-header">
              <span class="config-kicker">CONTAINER EXECUTION CAPABILITIES</span>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Start, stop and restart</span>
                <span class="setting-caption">Allow privileged users to fix crashes and cycle containers from the UI.</span>
              </div>
              <span class="status-pill" :class="(sharedState.envStartPermission || sharedState.envRestartPermission) ? 'pill-success' : 'pill-neutral'">
                {{ (sharedState.envStartPermission || sharedState.envRestartPermission) ? "Enabled" : "Disabled" }}
              </span>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Interactive Shell Access</span>
                <span class="setting-caption">Allow terminal exec into running containers via secure WebSocket.</span>
              </div>
              <span class="status-pill" :class="sharedState.envShellPermission ? 'pill-success' : 'pill-neutral'">
                {{ sharedState.envShellPermission ? "Enabled" : "Disabled" }}
              </span>
            </div>

            <div class="notice-callout">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="4 17 10 11 4 5"></polyline>
                <line x1="12" y1="19" x2="20" y2="19"></line>
              </svg>
              <span>Container shell access grants direct command execution. Disable in unauthenticated setups.</span>
            </div>
          </div>
        </section>

        <!-- ================= TAB: USERS ================= -->
        <section v-else-if="currentTab === 'users'" class="settings-panel animate-fade-in">
          <div class="panel-header flex justify-between items-center">
            <div>
              <h2 class="panel-title">User Management</h2>
              <p class="panel-subtitle">Manage staff accounts, privileges, shell permissions, and container access scopes.</p>
            </div>
            <button @click="userManagerRef?.openCreateModal()" class="dock-btn-primary">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="3">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
              <span>Add User</span>
            </button>
          </div>

          <div class="embedded-manager-card">
            <UserManager ref="userManagerRef" :token="token" embedded />
          </div>
        </section>

        <!-- ================= TAB: AUDIT ================= -->
        <section v-else-if="currentTab === 'audit'" class="settings-panel animate-fade-in">
          <div class="panel-header">
            <div>
              <h2 class="panel-title">Security &amp; Audit Logs</h2>
              <p class="panel-subtitle">Chronological record of administrative operations, container lifecycle actions, and security events.</p>
            </div>
          </div>

          <div class="embedded-manager-card">
            <AuditManager :token="token" embedded />
          </div>
        </section>

        <!-- ================= TAB: NOTIFICATIONS ================= -->
        <section v-else-if="currentTab === 'notifications'" class="settings-panel animate-fade-in">
          <div class="embedded-manager-card">
            <NotificationsView />
          </div>
        </section>

        <!-- ================= TAB: ALERTS ================= -->
        <section v-else-if="currentTab === 'alerts'" class="settings-panel animate-fade-in">
          <div class="embedded-manager-card">
            <AlertsView />
          </div>
        </section>

        <!-- ================= TAB: ABOUT ================= -->
        <section v-else-if="currentTab === 'about'" class="settings-panel animate-fade-in">
          <div class="about-hero-card">
            <div class="about-meta-row">
              <div class="about-brand-group">
                <img :src="logoSrc" alt="DockLog" class="brand-logo-img" />
                <div class="about-brand-text">
                  <div class="about-title-row">
                    <h2 class="about-brand-title">DockLog</h2>
                    <span class="version-tag">v{{ displayVersion }}</span>
                  </div>
                  <p class="about-brand-desc">Enterprise Container &amp; Cluster Log Aggregation Platform</p>
                </div>
              </div>

              <div class="about-link-row">
                <button
                  type="button"
                  class="link-btn"
                  :disabled="checkingUpdate"
                  @click="checkVersionUpdate"
                >
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" :class="{ 'animate-spin': checkingUpdate }">
                    <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/>
                  </svg>
                  <span>{{ versionStatusText }}</span>
                </button>
                <a href="https://github.com/dockloghq/docklog" target="_blank" rel="noopener noreferrer" class="link-btn">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                  </svg>
                  <span>Repository</span>
                </a>
              </div>
            </div>
          </div>

          <div class="config-block mt-6">
            <div class="config-block-header">
              <span class="config-kicker">INFRASTRUCTURE ENVIRONMENT</span>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Runtime Mode</span>
                <span class="setting-caption">Detected container runtime engine</span>
              </div>
              <span class="status-pill pill-neutral font-mono">{{ sharedState.runtimeMode.toUpperCase() }}</span>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Docker Swarm Clustering</span>
                <span class="setting-caption">Multi-node Swarm orchestration status</span>
              </div>
              <span class="status-pill" :class="sharedState.swarmActive ? 'pill-success' : 'pill-neutral'">
                {{ sharedState.swarmActive ? "ACTIVE CLUSTER" : "STANDALONE" }}
              </span>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Kubernetes Connectivity</span>
                <span class="setting-caption">In-cluster or kubeconfig bridge connection</span>
              </div>
              <span class="status-pill" :class="sharedState.k8sAvailable ? 'pill-success' : 'pill-neutral'">
                {{ sharedState.k8sAvailable ? "CONNECTED" : "NOT CONFIGURED" }}
              </span>
            </div>
          </div>

          <!-- Section: Database Storage Health -->
          <div class="config-block mt-6">
            <div class="config-block-header">
              <span class="config-kicker">DATABASE &amp; STORAGE HEALTH</span>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Database File (SQLite)</span>
                <span class="setting-caption">Storage footprint for accounts, configs, and metric rollups</span>
              </div>
              <span class="status-pill pill-neutral font-mono">{{ dbStats.db_size_human || '...' }}</span>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Historical Stats Records</span>
                <span class="setting-caption">Aggregated hourly metric checkpoints across containers</span>
              </div>
              <span class="status-pill pill-neutral font-mono">{{ dbStats.stats_count !== undefined ? `${dbStats.stats_count} rows` : '...' }}</span>
            </div>

            <div class="config-row">
              <div class="setting-text-col">
                <span class="setting-name">Compact &amp; Vacuum Database</span>
                <span class="setting-caption">Release deleted SQLite pages and return unused disk blocks to OS</span>
              </div>
              <button
                type="button"
                class="action-sub-btn"
                :disabled="vacuuming"
                @click="triggerVacuum"
              >
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" :class="{ 'animate-spin': vacuuming }">
                  <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/>
                  <path d="M3 3v5h5"/>
                  <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/>
                  <path d="M16 21h5v-5"/>
                </svg>
                <span>{{ vacuuming ? 'Compacting...' : 'Vacuum Now' }}</span>
              </button>
            </div>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  sharedState,
  saveLogSettings,
  fetchLogcConfig,
  fetchGeneralConfig,
  saveGeneralConfig,
  applyTheme,
  showToast,
} from "../utils/sharedState";
import { secureStorage } from "../utils/storage";
import { apiFetch } from "../utils/apiFetch";
import UserManager from "../components/UserManager.vue";
import AuditManager from "../components/AuditManager.vue";
import NotificationsView from "./Notifications.vue";
import AlertsView from "./Alerts.vue";

const route = useRoute();
const router = useRouter();

const token = secureStorage.getItem("token");
const userManagerRef = ref(null);

const logoSrc = computed(() =>
  sharedState.theme === "light" ? "/logo-icon-light.png" : "/logo-icon-dark.png"
);

const currentTab = ref(route.query.tab || "general");
const showAdvanced = ref(true);
const showAdvancedLogs = ref(true);

const displayVersion = computed(() => {
  const v = sharedState.appVersion || "1.1.0";
  return v.startsWith("v") ? v.slice(1) : v;
});

const checkingUpdate = ref(false);
const versionStatusText = ref("Check for Updates");
const dbStats = ref({ db_size_human: "", stats_count: 0, system_stats_count: 0 });
const vacuuming = ref(false);

async function checkVersionUpdate() {
  checkingUpdate.value = true;
  versionStatusText.value = "Checking...";
  try {
    const res = await apiFetch("/api/system/version");
    if (res.ok) {
      const data = await res.json();
      if (data.update_available) {
        versionStatusText.value = `Update Available (${data.latest})`;
        showToast("Update Available", `New version ${data.latest} is available on Docker Hub!`, "info", 5000);
      } else {
        versionStatusText.value = "Up to date";
        showToast("Up to date", "DockLog is running the latest available build.", "success", 3000);
      }
    }
  } catch (err) {
    versionStatusText.value = "Check for Updates";
    showToast("Error", "Failed to check Docker Hub for updates.", "error", 4000);
  } finally {
    checkingUpdate.value = false;
  }
}

async function fetchDbStats() {
  try {
    const res = await apiFetch("/api/system/db-stats");
    if (res.ok) {
      dbStats.value = await res.json();
    }
  } catch (err) {
    console.error("Failed to load db stats:", err);
  }
}

async function triggerVacuum() {
  vacuuming.value = true;
  try {
    const res = await apiFetch("/api/system/vacuum", { method: "POST" });
    if (res.ok) {
      const data = await res.json();
      dbStats.value.db_size_human = data.db_size_human;
      showToast("Storage Compacted", "Database successfully vacuumed and disk space reclaimed.", "success", 3500);
    } else {
      showToast("Vacuum Failed", "Could not complete database vacuum.", "error", 3500);
    }
  } catch (err) {
    showToast("Error", "Failed to request database vacuum.", "error", 3500);
  } finally {
    vacuuming.value = false;
  }
}

function selectTab(tab) {
  currentTab.value = tab;
  router.replace({ query: { ...route.query, tab } });
}

watch(
  () => route.query.tab,
  (newTab) => {
    if (newTab && newTab !== currentTab.value) {
      currentTab.value = newTab;
    }
  }
);

const headerSubtitle = computed(() => {
  if (currentTab.value === "general") {
    return "Browser preferences and user interface defaults. Saved locally.";
  }
  if (currentTab.value === "logs") {
    return "Tailored formatting and rendering rules for live log streams.";
  }
  if (currentTab.value === "security") {
    return "Platform permissions, session controls, and credentials.";
  }
  if (currentTab.value === "users") {
    return "Staff accounts, role assignments, and container access scopes.";
  }
  if (currentTab.value === "audit") {
    return "Security logs, administrative trails, and login events.";
  }
  if (currentTab.value === "notifications") {
    return "Outbound alert webhooks (Slack, Teams, Discord, custom HTTP).";
  }
  if (currentTab.value === "alerts") {
    return "Rule-based cluster triggers, pattern monitors, and threshold alerts.";
  }
  if (currentTab.value === "about") {
    return "DockLog build distribution and orchestration engine status.";
  }
  return "Manage your display preferences and platform configuration.";
});

// General preferences (stored in secureStorage)
const generalPrefs = reactive({
  language: secureStorage.getItem("docklog_language") || "auto",
  followRecreated: secureStorage.getItem("docklog_follow_recreated") || "delayed",
  searchIntercept: secureStorage.getItem("docklog_search_intercept") !== "false", // default true
});

function saveGeneralPrefs() {
  saveGeneralConfig(generalPrefs);
}

// Log stream reactive preferences
const logPrefs = reactive({
  fontSize: sharedState.logSettings?.fontSize || "small",
  compactMode: sharedState.logSettings?.compactMode ?? false,
  showTimestamps: sharedState.logSettings?.showTimestamps ?? true,
  softWrap: sharedState.logSettings?.softWrap ?? true,
  highlightErrors: sharedState.logSettings?.highlightErrors ?? true,
  timeFormat: sharedState.logSettings?.timeFormat || "auto",
  dateFormat: sharedState.logSettings?.dateFormat || "YYYY-MM-DD",
  showStreamLabels: sharedState.logSettings?.showStreamLabels ?? true,
  showLineNumbers: sharedState.logSettings?.showLineNumbers ?? false,
  autoScrollMode: sharedState.logSettings?.autoScrollMode || "always",
});

watch(
  () => sharedState.logSettings,
  (newSettings) => {
    if (newSettings) {
      logPrefs.fontSize = newSettings.fontSize;
      logPrefs.compactMode = newSettings.compactMode;
      logPrefs.showTimestamps = newSettings.showTimestamps;
      logPrefs.softWrap = newSettings.softWrap;
      logPrefs.highlightErrors = newSettings.highlightErrors;
      logPrefs.timeFormat = newSettings.timeFormat;
      if (newSettings.dateFormat !== undefined) logPrefs.dateFormat = newSettings.dateFormat;
      if (newSettings.showStreamLabels !== undefined) logPrefs.showStreamLabels = newSettings.showStreamLabels;
      if (newSettings.showLineNumbers !== undefined) logPrefs.showLineNumbers = newSettings.showLineNumbers;
      if (newSettings.autoScrollMode !== undefined) logPrefs.autoScrollMode = newSettings.autoScrollMode;
    }
  },
  { deep: true }
);

watch(
  () => sharedState.generalSettings,
  (newGeneral) => {
    if (newGeneral) {
      if (newGeneral.language) generalPrefs.language = newGeneral.language;
      if (newGeneral.followRecreated) generalPrefs.followRecreated = newGeneral.followRecreated;
      if (newGeneral.searchIntercept !== undefined) generalPrefs.searchIntercept = newGeneral.searchIntercept;
    }
  },
  { deep: true }
);

function updateLogPref(key, value) {
  logPrefs[key] = value;
  saveLogSettings({ [key]: value });
}

onMounted(async () => {
  await Promise.all([fetchLogcConfig(), fetchGeneralConfig()]);
  fetchDbStats();
  const lang = secureStorage.getItem("docklog_language");
  if (lang) generalPrefs.language = lang;
  const follow = secureStorage.getItem("docklog_follow_recreated");
  if (follow) generalPrefs.followRecreated = follow;
  const intercept = secureStorage.getItem("docklog_search_intercept");
  if (intercept !== null) generalPrefs.searchIntercept = intercept === true || intercept === "true";
  const dateFormat = secureStorage.getItem("docklog_log_date_format");
  if (dateFormat) logPrefs.dateFormat = dateFormat;
  const streamLabels = secureStorage.getItem("docklog_stream_labels");
  if (streamLabels !== null) logPrefs.showStreamLabels = streamLabels === true || streamLabels === "true";
  const lineNumbers = secureStorage.getItem("docklog_line_numbers");
  if (lineNumbers !== null) logPrefs.showLineNumbers = lineNumbers === true || lineNumbers === "true";
  const autoScroll = secureStorage.getItem("docklog_autoscroll_mode");
  if (autoScroll) logPrefs.autoScrollMode = autoScroll;
});

// Reset to defaults
function resetToDefaults() {
  applyTheme("auto");
  updateLogPref("fontSize", "small");
  updateLogPref("compactMode", false);
  updateLogPref("showTimestamps", true);
  updateLogPref("softWrap", true);
  updateLogPref("highlightErrors", true);
  updateLogPref("timeFormat", "auto");
  updateLogPref("dateFormat", "YYYY-MM-DD");
  updateLogPref("showStreamLabels", true);
  updateLogPref("showLineNumbers", false);
  updateLogPref("autoScrollMode", "always");
  generalPrefs.language = "auto";
  generalPrefs.followRecreated = "delayed";
  generalPrefs.searchIntercept = true;
  saveGeneralPrefs();
  showToast("Reset", "Preferences reset to defaults", "info");
}

// Sample dates for preview
const now = new Date();
const sampleDate1 = new Date(now.getTime() - 4 * 60 * 60 * 1000);
const sampleDate2 = new Date(now.getTime() - 3 * 60 * 60 * 1000);
const sampleDate3 = new Date(now.getTime() - 2 * 60 * 60 * 1000);
const sampleDate4 = new Date(now.getTime() - 1 * 60 * 60 * 1000);
const sampleDate5 = new Date(now.getTime());

const pad = (n) => String(n).padStart(2, "0");

function formatSampleTime(date) {
  const yyyy = date.getFullYear();
  const mm = pad(date.getMonth() + 1);
  const dd = pad(date.getDate());

  let datePart = `${yyyy}-${mm}-${dd}`;
  const fmt = logPrefs.dateFormat || "YYYY-MM-DD";
  if (fmt === "YYYY/MM/DD") {
    datePart = `${yyyy}/${mm}/${dd}`;
  } else if (fmt === "YYYY/DD/MM") {
    datePart = `${yyyy}/${dd}/${mm}`;
  } else if (fmt === "DD/MM/YYYY") {
    datePart = `${dd}/${mm}/${yyyy}`;
  } else if (fmt === "DD-MM-YYYY") {
    datePart = `${dd}-${mm}-${yyyy}`;
  } else if (fmt === "MM/DD/YYYY") {
    datePart = `${mm}/${dd}/${yyyy}`;
  }

  if (logPrefs.timeFormat === "12-hour") {
    let hh = date.getHours();
    const ampm = hh >= 12 ? "PM" : "AM";
    hh = hh % 12;
    hh = hh ? hh : 12;
    return `${datePart} ${pad(hh)}:${pad(date.getMinutes())}:${pad(date.getSeconds())} ${ampm}`;
  }
  return `${datePart} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

// Password update form
const newPassword = ref("");
const confirmPassword = ref("");
const currentPassword = ref("");
const loading = ref(false);
const pwdError = ref("");

const handlePasswordUpdate = async () => {
  if (newPassword.value !== confirmPassword.value) {
    pwdError.value = "Passwords do not match";
    return;
  }
  if (newPassword.value.length < 8) {
    pwdError.value = "Password must be at least 8 characters";
    return;
  }

  const mustProvideCurrent =
    sharedState.currentUser?.password_changed !== false &&
    !sharedState.forcePasswordChange;

  if (mustProvideCurrent && !currentPassword.value) {
    pwdError.value = "Current password is required";
    return;
  }

  loading.value = true;
  pwdError.value = "";

  try {
    const formData = new FormData();
    formData.append("password", newPassword.value);
    if (mustProvideCurrent) {
      formData.append("current_password", currentPassword.value);
    }

    const res = await apiFetch("/api/user/change-password", {
      method: "POST",
      body: formData,
    });

    if (res.ok) {
      showToast("Success", "Password updated successfully", "success");
      newPassword.value = "";
      confirmPassword.value = "";
      currentPassword.value = "";
      sharedState.forcePasswordChange = false;
      sharedState.showPasswordModal = false;
      if (sharedState.currentUser) {
        sharedState.currentUser.password_changed = true;
      }
    } else {
      const data = await res.json().catch(() => ({}));
      pwdError.value = data.error || "Failed to update password";
    }
  } catch (err) {
    pwdError.value = "System connection failed";
  } finally {
    loading.value = false;
  }
};

function openChangePasswordModal() {
  newPassword.value = "";
  confirmPassword.value = "";
  currentPassword.value = "";
  pwdError.value = "";
  sharedState.showPasswordModal = true;
}
</script>

<style scoped>
.settings-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1.5rem 2rem 3.5rem;
  font-family: var(--font-main);
  color: var(--text-main);
}

/* Top bar */
.settings-top-bar {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 2rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid var(--border);
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-mute);
  margin-bottom: 0.35rem;
}

.breadcrumb-separator {
  opacity: 0.4;
}

.active-crumb {
  color: var(--accent);
}

.settings-title {
  font-size: 1.85rem;
  font-weight: 800;
  letter-spacing: -0.025em;
  color: var(--text-main);
  margin: 0 0 0.25rem;
}

.settings-subtitle {
  font-size: 0.88rem;
  color: var(--text-mute);
  margin: 0;
}

.reset-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.85rem;
  border-radius: 8px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  color: var(--text-dim);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.reset-btn:hover {
  color: var(--text-main);
  border-color: var(--border-active);
  background: var(--bg-subtle);
}

/* Main Layout */
.settings-layout {
  display: flex;
  gap: 2.25rem;
  align-items: flex-start;
}

@media (max-width: 900px) {
  .settings-layout {
    flex-direction: column;
  }
}

/* Sidebar Navigation */
.settings-sidebar {
  width: 250px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  position: sticky;
  top: 1rem;
  max-height: calc(100vh - 120px);
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: thin;
}

@media (max-width: 900px) {
  .settings-sidebar {
    position: static;
    max-height: none;
    width: 100%;
    flex-direction: row;
    overflow-x: auto;
    padding-bottom: 0.75rem;
  }
}

.sidebar-category-label {
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  color: var(--text-mute);
  padding: 0.4rem 0.65rem;
}

.nav-tab-btn {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.85rem;
  border-radius: 10px;
  background: transparent;
  border: 1px solid transparent;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
}

.nav-tab-btn:hover {
  background: var(--bg-card);
  border-color: var(--border);
}

.nav-tab-btn.active {
  background: var(--bg-card);
  border-color: var(--accent);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.tab-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-input);
  color: var(--text-dim);
  flex-shrink: 0;
  transition: all 0.15s ease;
}

.nav-tab-btn.active .tab-icon {
  background: var(--accent-soft);
  color: var(--accent);
}

.tab-label-group {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
}

.tab-title {
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--text-main);
}

.tab-desc {
  font-size: 0.72rem;
  color: var(--text-mute);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Content Area */
.settings-content {
  flex: 1;
  min-width: 0;
}

.panel-header {
  margin-bottom: 1.5rem;
}

.panel-title {
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: -0.015em;
  color: var(--text-main);
  margin: 0 0 0.25rem;
}

.panel-subtitle {
  font-size: 0.85rem;
  color: var(--text-mute);
  margin: 0;
}

/* Config Blocks */
.config-block {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  margin-bottom: 1.5rem;
  overflow: hidden;
}

.config-block-header {
  padding: 0.85rem 1.25rem;
  background: var(--bg-subtle);
  border-bottom: 1px solid var(--border);
}

.config-kicker {
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  color: var(--text-mute);
  text-transform: uppercase;
}

.config-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.05rem 1.25rem;
  border-bottom: 1px solid var(--border);
  gap: 1.5rem;
}

.config-row:last-child {
  border-bottom: none;
}

.config-row-stacked {
  padding: 1.15rem 1.25rem;
  border-bottom: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 0.95rem;
}

.setting-text-col {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.setting-name {
  font-size: 0.88rem;
  font-weight: 700;
  color: var(--text-main);
}

.setting-caption {
  font-size: 0.78rem;
  color: var(--text-mute);
  line-height: 1.4;
}

.inline-title-kbd {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.kbd-badge {
  font-size: 0.7rem;
  font-family: var(--font-mono);
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  border-radius: 5px;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  color: var(--text-dim);
}

/* Theme Card Picker */
.theme-card-picker {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.85rem;
}

@media (max-width: 600px) {
  .theme-card-picker {
    grid-template-columns: 1fr;
  }
}

.theme-card {
  padding: 0.75rem;
  border-radius: 10px;
  background: var(--bg-input);
  border: 2px solid var(--border);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.65rem;
  transition: all 0.15s ease;
}

.theme-card:hover {
  border-color: var(--border-active);
}

.theme-card.selected {
  border-color: var(--accent);
  background: var(--bg-card);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.theme-preview-box {
  width: 100%;
  height: 52px;
  border-radius: 6px;
  border: 1px solid var(--border);
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 0 0.65rem;
  gap: 0.35rem;
}

.preview-light {
  background: #f8fafc;
}

.preview-dark {
  background: #090d16;
}

.preview-auto {
  display: flex;
  flex-direction: row;
  padding: 0;
}

.tp-split-left {
  flex: 1;
  background: #f8fafc;
  display: flex;
  align-items: center;
  padding-left: 0.4rem;
}

.tp-split-right {
  flex: 1;
  background: #090d16;
  display: flex;
  align-items: center;
  padding-left: 0.4rem;
}

.tp-line {
  height: 4px;
  border-radius: 2px;
}

.theme-card-label {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-main);
}

/* Custom Select Wrap */
.custom-select-wrap {
  position: relative;
  min-width: 170px;
}

.custom-select-wrap select {
  width: 100%;
  appearance: none;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-main);
  padding: 0.45rem 2rem 0.45rem 0.85rem;
  border-radius: 8px;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  outline: none;
}

.custom-select-wrap select:focus {
  border-color: var(--accent);
}

.select-chevron {
  position: absolute;
  right: 0.65rem;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  color: var(--text-mute);
}

/* Switch */
.dock-switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
  flex-shrink: 0;
}

.dock-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.dock-switch-slider {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background-color: var(--bg-input);
  border: 1px solid var(--border);
  transition: 0.2s;
  border-radius: 24px;
}

.dock-switch-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 2px;
  bottom: 2px;
  background-color: var(--text-mute);
  transition: 0.2s;
  border-radius: 50%;
}

.dock-switch input:checked + .dock-switch-slider {
  background-color: var(--accent);
  border-color: var(--accent);
}

.dock-switch input:checked + .dock-switch-slider:before {
  transform: translateX(20px);
  background-color: #ffffff;
}

/* Accordion */
.accordion-toggle {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.85rem 1.25rem;
  background: var(--bg-subtle);
  border: none;
  cursor: pointer;
  color: var(--text-main);
  text-align: left;
}

.accordion-body {
  border-top: 1px solid var(--border);
}

/* Segmented Buttons */
.dock-segmented {
  display: inline-flex;
  padding: 3px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.dock-seg-btn {
  padding: 0.35rem 0.8rem;
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-mute);
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.dock-seg-btn:hover {
  color: var(--text-main);
}

.dock-seg-btn.active {
  background: var(--bg-card);
  color: var(--text-main);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

/* Log Envelope */
.log-preview-envelope {
  background: #090d16;
  border: 1px solid #1e293b;
  border-radius: 12px;
  margin-bottom: 1.5rem;
  overflow: hidden;
  font-family: var(--font-mono);
}

[data-theme="light"] .log-preview-envelope {
  background: #0f172a;
}

.preview-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.6rem 1rem;
  background: #030712;
  border-bottom: 1px solid #1e293b;
}

.preview-pill {
  font-size: 0.68rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #38bdf8;
}

.preview-hint {
  font-size: 0.72rem;
  color: #64748b;
}

.live-log-console {
  padding: 1.15rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  line-height: 1.5;
  color: #f1f5f9;
}

.live-log-console.font-small { font-size: 0.72rem; }
.live-log-console.font-medium { font-size: 0.84rem; }
.live-log-console.font-large { font-size: 0.96rem; }

.live-log-console.compact-mode {
  gap: 0.15rem;
  line-height: 1.3;
}

.live-log-console.no-soft-wrap {
  overflow-x: auto;
  white-space: pre !important;
}

.preview-line {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
  padding: 0.15rem 0.5rem;
  border-left: 3px solid transparent;
  border-radius: 2px;
}

.line-number-preview {
  font-family: var(--font-mono);
  font-size: 0.68rem;
  color: #64748b;
  min-width: 1.4rem;
  text-align: right;
  user-select: none;
  opacity: 0.7;
}

.stream-tag {
  font-family: var(--font-mono);
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  padding: 0.08rem 0.35rem;
  border-radius: 4px;
  flex-shrink: 0;
  user-select: none;
  line-height: 1.25;
}

.stream-tag.stdout {
  color: #38bdf8;
  background: rgba(56, 189, 248, 0.12);
  border: 1px solid rgba(56, 189, 248, 0.25);
}

.stream-tag.stderr {
  color: #f87171;
  background: rgba(239, 68, 68, 0.14);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.preview-time {
  color: #64748b;
  font-size: 0.7rem;
  flex-shrink: 0;
  user-select: none;
}

.preview-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}
.preview-dot.normal { background: #10b981; }
.preview-dot.warn { background: #f59e0b; }
.preview-dot.purple { background: #a855f7; }
.preview-dot.error { background: #ef4444; }

.multiline-error.is-highlighted-error {
  background: rgba(239, 68, 68, 0.12);
  border-left-color: #ef4444;
}

.preview-text {
  flex: 1;
  min-width: 0;
}

.text-error { color: #f87171; }
.text-warning { color: #fbbf24; }

/* Status Banners */
.status-banner {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  padding: 1rem 1.25rem;
  border-radius: 10px;
  margin-bottom: 1.5rem;
  border: 1px solid var(--border);
}

.banner-success {
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.25);
  color: #10b981;
}

.banner-warn {
  background: rgba(245, 158, 11, 0.08);
  border-color: rgba(245, 158, 11, 0.25);
  color: #f59e0b;
}

.banner-body {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.banner-title {
  font-size: 0.85rem;
  font-weight: 800;
  color: var(--text-main);
}

.banner-desc {
  font-size: 0.78rem;
  color: var(--text-dim);
  line-height: 1.4;
}

.shield-badge {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  color: var(--text-dim);
}

.shield-badge.locked {
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
  border-color: rgba(16, 185, 129, 0.2);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  font-size: 0.72rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.25rem 0.65rem;
  border-radius: 9999px;
  line-height: 1.2;
}

.pill-success {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.pill-warn {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.pill-neutral {
  background: rgba(148, 163, 184, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

[data-theme="light"] .pill-neutral {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #cbd5e1;
}

.action-sub-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.32rem 0.75rem;
  border-radius: 8px;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  color: var(--text-main);
  font-size: 0.78rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-sub-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
  color: #ffffff;
  transform: translateY(-1px);
}

[data-theme="light"] .action-sub-btn:hover {
  background: rgba(0, 0, 0, 0.06);
  border-color: rgba(0, 0, 0, 0.2);
  color: #000000;
}

.notice-callout {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.85rem 1.25rem;
  background: rgba(245, 158, 11, 0.06);
  border-top: 1px solid rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  font-size: 0.78rem;
  font-weight: 600;
}

/* Forms */
.dock-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 540px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-field label {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-mute);
}

.dock-input {
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-main);
  padding: 0.55rem 0.85rem;
  border-radius: 8px;
  font-size: 0.84rem;
  outline: none;
  transition: all 0.15s ease;
}

.dock-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px var(--accent-soft);
}

.dock-btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 1.25rem;
  background: var(--accent);
  color: #ffffff;
  border: none;
  border-radius: 8px;
  font-size: 0.82rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
}

.dock-btn-primary:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.dock-btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.form-footer {
  display: flex;
  align-items: center;
  gap: 1rem;
}

/* About tab */
.about-hero-card {
  padding: 1.5rem 1.75rem;
  border-radius: 12px;
  background: var(--bg-card);
  border: 1px solid var(--border);
}

.about-meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.about-brand-group {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.about-brand-text {
  display: flex;
  flex-direction: column;
}

.about-title-row {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  line-height: 1;
}

.brand-logo-img {
  width: 48px;
  height: 48px;
  object-fit: contain;
  border-radius: 10px;
}

.about-brand-title {
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text-main);
  margin: 0;
}

.about-brand-desc {
  font-size: 0.78rem;
  color: var(--text-mute);
  margin: 0.15rem 0 0;
}

.version-tag {
  font-size: 0.72rem;
  font-family: var(--font-mono);
  font-weight: 800;
  padding: 0.15rem 0.5rem;
  border-radius: 6px;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  color: var(--text-dim);
}

.about-link-row {
  display: flex;
  gap: 0.5rem;
}

.link-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.85rem;
  border-radius: 8px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  color: var(--text-dim);
  font-size: 0.8rem;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.15s ease;
}

.link-btn:hover {
  color: var(--text-main);
  border-color: var(--border-active);
}

.mono-badge {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.2rem 0.55rem;
  border-radius: 6px;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  color: var(--text-main);
}

.embedded-manager-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
  padding: 1.25rem;
}

.embedded-manager-card :deep(.page-hero),
.embedded-manager-card :deep(.page-metrics) {
  display: none !important;
}

.embedded-manager-card :deep(.page-view) {
  padding: 0 !important;
  max-width: none !important;
}
</style>

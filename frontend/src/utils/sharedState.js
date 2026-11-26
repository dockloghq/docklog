import { reactive } from 'vue';
import { secureStorage } from './storage';
import { apiFetch } from './apiFetch';
import { readApiError } from './authSession';

export function getSystemTheme() {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function normalizeThemePreference(value) {
  if (value === 'system') return 'auto';
  if (value === 'light' || value === 'dark' || value === 'auto') return value;
  return 'auto';
}

function isAutoTheme(preference) {
  return preference === 'auto' || preference === 'system';
}

function loadThemePreference() {
  const stored = secureStorage.getItem('theme');
  if (!stored) return 'auto';
  return normalizeThemePreference(stored);
}

export function resolveTheme(preference = loadThemePreference()) {
  if (isAutoTheme(preference)) return getSystemTheme();
  return preference;
}

const initialThemePreference = loadThemePreference();
const initialResolvedTheme = resolveTheme(initialThemePreference);

if (typeof document !== 'undefined') {
  document.documentElement.setAttribute('data-theme', initialResolvedTheme);
}

function parseStorageBool(val, defaultVal = false) {
  if (val === null || val === undefined) return defaultVal;
  if (typeof val === 'boolean') return val;
  if (val === 'true') return true;
  if (val === 'false') return false;
  return Boolean(val);
}

export const sharedState = reactive({
  currentUser: null,
  systemStats: { cpu: 0, usedMemGB: 0, memory: '0 / 0' },
  searchQuery: '',
  themePreference: initialThemePreference,
  theme: initialResolvedTheme,
  showPasswordModal: false,
  forcePasswordChange: false,
  dashboardSidebarOpen: typeof window !== 'undefined' ? window.innerWidth > 1024 : true,
  adminSidebarOpen: typeof window !== 'undefined' ? window.innerWidth > 1024 : true,
  toast: {
    visible: false,
    title: '',
    message: '',
    type: 'success',
    duration: 4000,
  },
  appVersion: '1.1.0',
  isAuthDisabled: false,
  configLoaded: false,
  envStartPermission: false,
  envStopPermission: false,
  envRestartPermission: false,
  envDeletePermission: false,
  envShellPermission: false,
  isBackendDisconnected: false,
  runtimeMode: 'docker',
  swarmActive: false,
  k8sNamespaces: [],
  k8sDefaultNs: 'default',
  k8sAvailable: false,
  k8sError: '',

  // Log display preferences (synced to settings/log.yaml and localStorage)
  logSettings: {
    fontSize: secureStorage.getItem('docklog_log_font_size') || 'small', // 'small' | 'medium' | 'large'
    compactMode: parseStorageBool(secureStorage.getItem('docklog_log_compact'), false),
    showTimestamps: parseStorageBool(secureStorage.getItem('docklog_log_timestamps'), true),
    softWrap: parseStorageBool(secureStorage.getItem('docklog_log_soft_wrap'), true),
    highlightErrors: parseStorageBool(secureStorage.getItem('docklog_log_highlight_errors'), true),
    timeFormat: secureStorage.getItem('docklog_log_time_format') || 'auto', // 'auto' | '12-hour' | '24-hour'
    dateFormat: secureStorage.getItem('docklog_log_date_format') || 'YYYY-MM-DD', // 'YYYY-MM-DD' | 'YYYY/MM/DD' | 'YYYY/DD/MM' | 'DD/MM/YYYY' | 'MM/DD/YYYY'
    showStreamLabels: parseStorageBool(secureStorage.getItem('docklog_stream_labels'), true),
    showLineNumbers: parseStorageBool(secureStorage.getItem('docklog_line_numbers'), false),
    autoScrollMode: secureStorage.getItem('docklog_autoscroll_mode') || 'always', // 'always' | 'smart' | 'manual'
    defaultTail: secureStorage.getItem('docklog_default_tail') || '100', // '50' | '100' | '250' | '500'
  },

  // General display preferences (synced to settings/general.yaml and localStorage)
  generalSettings: {
    language: secureStorage.getItem('docklog_language') || 'auto',
    followRecreated: secureStorage.getItem('docklog_follow_recreated') || 'delayed',
    searchIntercept: parseStorageBool(secureStorage.getItem('docklog_search_intercept'), true),
  },
});

export async function fetchLogcConfig() {
  try {
    const res = await apiFetch('/api/system/logc');
    if (res.ok) {
      const data = await res.json();
      if (data.font_size) sharedState.logSettings.fontSize = data.font_size;
      if (data.compact_mode !== undefined) sharedState.logSettings.compactMode = data.compact_mode;
      if (data.show_timestamps !== undefined) sharedState.logSettings.showTimestamps = data.show_timestamps;
      if (data.soft_wrap !== undefined) sharedState.logSettings.softWrap = data.soft_wrap;
      if (data.highlight_errors !== undefined) sharedState.logSettings.highlightErrors = data.highlight_errors;
      if (data.time_format) sharedState.logSettings.timeFormat = data.time_format;
      if (data.date_format) {
        sharedState.logSettings.dateFormat = data.date_format;
        secureStorage.setItem('docklog_log_date_format', data.date_format);
      }
      if (data.show_stream_labels !== undefined) {
        sharedState.logSettings.showStreamLabels = data.show_stream_labels;
        secureStorage.setItem('docklog_stream_labels', String(data.show_stream_labels));
      }
      if (data.show_line_numbers !== undefined) {
        sharedState.logSettings.showLineNumbers = data.show_line_numbers;
        secureStorage.setItem('docklog_line_numbers', String(data.show_line_numbers));
      }
      if (data.auto_scroll_mode) {
        sharedState.logSettings.autoScrollMode = data.auto_scroll_mode;
        secureStorage.setItem('docklog_autoscroll_mode', data.auto_scroll_mode);
      }
      if (data.default_tail) {
        sharedState.logSettings.defaultTail = data.default_tail;
        secureStorage.setItem('docklog_default_tail', data.default_tail);
      }

      // Sync to local storage
      secureStorage.setItem('docklog_log_font_size', sharedState.logSettings.fontSize);
      secureStorage.setItem('docklog_log_compact', String(sharedState.logSettings.compactMode));
      secureStorage.setItem('docklog_log_timestamps', String(sharedState.logSettings.showTimestamps));
      secureStorage.setItem('docklog_log_soft_wrap', String(sharedState.logSettings.softWrap));
      secureStorage.setItem('docklog_log_highlight_errors', String(sharedState.logSettings.highlightErrors));
      secureStorage.setItem('docklog_log_time_format', sharedState.logSettings.timeFormat);
      return data;
    }
  } catch (err) {
    console.error('Failed to load logc.yaml config:', err);
  }
  return null;
}

export async function fetchGeneralConfig() {
  try {
    const res = await apiFetch('/api/system/general');
    if (res.ok) {
      const data = await res.json();
      if (data.theme) {
        applyTheme(data.theme, false);
      }
      if (data.language) {
        sharedState.generalSettings.language = data.language;
        secureStorage.setItem('docklog_language', data.language);
      }
      if (data.follow_recreated) {
        sharedState.generalSettings.followRecreated = data.follow_recreated;
        secureStorage.setItem('docklog_follow_recreated', data.follow_recreated);
      }
      if (data.search_intercept !== undefined) {
        sharedState.generalSettings.searchIntercept = data.search_intercept;
        secureStorage.setItem('docklog_search_intercept', String(data.search_intercept));
      }
      return data;
    }
  } catch (err) {
    console.error('Failed to load general.yaml config:', err);
  }
  return null;
}

export async function saveGeneralConfig(prefs) {
  if (prefs.language) {
    sharedState.generalSettings.language = prefs.language;
    secureStorage.setItem('docklog_language', prefs.language);
  }
  if (prefs.followRecreated) {
    sharedState.generalSettings.followRecreated = prefs.followRecreated;
    secureStorage.setItem('docklog_follow_recreated', prefs.followRecreated);
  }
  if (prefs.searchIntercept !== undefined) {
    sharedState.generalSettings.searchIntercept = prefs.searchIntercept;
    secureStorage.setItem('docklog_search_intercept', String(prefs.searchIntercept));
  }

  try {
    const payload = {
      theme: sharedState.themePreference,
      language: prefs.language || secureStorage.getItem('docklog_language') || 'en-US',
      follow_recreated: prefs.followRecreated || secureStorage.getItem('docklog_follow_recreated') || 'delayed',
      search_intercept: prefs.searchIntercept !== undefined ? prefs.searchIntercept : secureStorage.getItem('docklog_search_intercept') !== 'false',
    };
    await apiFetch('/api/system/general', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
  } catch (e) {
    console.error('Failed to persist general.yaml to server:', e);
  }
}

export async function saveLogSettings(newSettings) {
  Object.assign(sharedState.logSettings, newSettings);
  secureStorage.setItem('docklog_log_font_size', sharedState.logSettings.fontSize);
  secureStorage.setItem('docklog_log_compact', String(sharedState.logSettings.compactMode));
  secureStorage.setItem('docklog_log_timestamps', String(sharedState.logSettings.showTimestamps));
  secureStorage.setItem('docklog_log_soft_wrap', String(sharedState.logSettings.softWrap));
  secureStorage.setItem('docklog_log_highlight_errors', String(sharedState.logSettings.highlightErrors));
  secureStorage.setItem('docklog_log_time_format', sharedState.logSettings.timeFormat);
  if (newSettings.dateFormat !== undefined) {
    secureStorage.setItem('docklog_log_date_format', newSettings.dateFormat);
  }
  if (newSettings.showStreamLabels !== undefined) {
    secureStorage.setItem('docklog_stream_labels', String(newSettings.showStreamLabels));
  }
  if (newSettings.showLineNumbers !== undefined) {
    secureStorage.setItem('docklog_line_numbers', String(newSettings.showLineNumbers));
  }
  if (newSettings.autoScrollMode !== undefined) {
    secureStorage.setItem('docklog_autoscroll_mode', newSettings.autoScrollMode);
  }
  if (newSettings.defaultTail !== undefined) {
    secureStorage.setItem('docklog_default_tail', newSettings.defaultTail);
  }

  // Persist to backend settings/logc.yaml
  try {
    const payload = {
      font_size: sharedState.logSettings.fontSize,
      compact_mode: sharedState.logSettings.compactMode,
      show_timestamps: sharedState.logSettings.showTimestamps,
      soft_wrap: sharedState.logSettings.softWrap,
      highlight_errors: sharedState.logSettings.highlightErrors,
      time_format: sharedState.logSettings.timeFormat,
      date_format: sharedState.logSettings.dateFormat,
      show_stream_labels: sharedState.logSettings.showStreamLabels,
      show_line_numbers: sharedState.logSettings.showLineNumbers,
      auto_scroll_mode: sharedState.logSettings.autoScrollMode,
      default_tail: sharedState.logSettings.defaultTail,
    };
    await apiFetch('/api/system/logc', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
  } catch (e) {
    console.error('Failed to persist logc.yaml to server:', e);
  }
}

export function dockerEnabled() {
  return sharedState.runtimeMode === 'docker' || sharedState.runtimeMode === 'both';
}

export function swarmEnabled() {
  return dockerEnabled() && sharedState.swarmActive === true;
}

export function kubernetesEnabled() {
  return sharedState.runtimeMode === 'kubernetes' || sharedState.runtimeMode === 'both';
}

export function applyTheme(preference, persistToServer = true) {
  const normalized = normalizeThemePreference(preference);
  sharedState.themePreference = normalized;
  sharedState.theme = resolveTheme(normalized);
  secureStorage.setItem('theme', normalized);
  document.documentElement.setAttribute('data-theme', sharedState.theme);

  if (persistToServer) {
    saveGeneralConfig({ theme: normalized });
  }
}

export function toggleTheme() {
  applyTheme(sharedState.theme === 'dark' ? 'light' : 'dark');
}

export function initThemeListener() {
  if (typeof window === 'undefined') return () => {};

  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handleChange = () => {
    if (!isAutoTheme(sharedState.themePreference)) return;
    sharedState.theme = getSystemTheme();
    document.documentElement.setAttribute('data-theme', sharedState.theme);
  };

  mediaQuery.addEventListener('change', handleChange);
  return () => mediaQuery.removeEventListener('change', handleChange);
}

let toastTimer = null;

export const showToast = (title, message, type = 'success', duration = 3500) => {
  if (toastTimer) {
    clearTimeout(toastTimer);
    toastTimer = null;
  }
  sharedState.toast.title = title;
  sharedState.toast.message = message;
  sharedState.toast.type = type;
  sharedState.toast.duration = duration;
  sharedState.toast.visible = true;
  toastTimer = setTimeout(() => {
    sharedState.toast.visible = false;
    toastTimer = null;
  }, duration);
};

export const fetchCurrentUser = async () => {
  if (sharedState.isAuthDisabled) {
    sharedState.currentUser = {
      id: 1,
      username: "admin",
      is_admin: true,
      can_start: sharedState.envStartPermission,
      can_stop: sharedState.envStopPermission,
      can_restart: sharedState.envRestartPermission,
      can_delete: sharedState.envDeletePermission,
      can_shell: sharedState.envShellPermission,
      is_active: true
    };
    return { status: 'ok', user: sharedState.currentUser };
  }
  const token = secureStorage.getItem('token');
  if (!token) return { status: 'missing', user: null };
  try {
    const res = await apiFetch('/api/user/me', {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (res.ok) {
      sharedState.currentUser = await res.json();
      return { status: 'ok', user: sharedState.currentUser };
    }
    if (res.status === 403) {
      const err = await readApiError(res, 'Access denied');
      if (err.code === 'ACCOUNT_DEACTIVATED') {
        sharedState.currentUser = null;
        secureStorage.removeItem('token');
        secureStorage.removeItem('user');
        return { status: 'forbidden', user: null };
      }
      return { status: 'error', user: sharedState.currentUser, error: err.message, code: err.code };
    }
  } catch (e) {
    console.error('Failed to fetch user:', e);
  }
  return { status: 'error', user: null };
};

export const fetchSystemStats = async () => {
  const token = secureStorage.getItem('token');
  if (!token) return;
  try {
    const res = await apiFetch('/api/system/stats', {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (res.ok) {
      if (sharedState.isBackendDisconnected) {
        sharedState.isBackendDisconnected = false;
        showToast("Backend Reconnected", "Connection to the server has been restored", "success");
      }
      const data = await res.json();
      sharedState.systemStats = {
        cpu: data.cpu,
        cores: data.cores || 1,
        memory: data.memory,
        total_memory: data.total_memory,
        disk_percent: data.disk_percent || 0,
      };
    } else {
      handleBackendError();
    }
  } catch (e) {
    handleBackendError();
    console.error('Failed to fetch system stats:', e);
  }
};

const handleBackendError = () => {
  if (!sharedState.isBackendDisconnected) {
    sharedState.isBackendDisconnected = true;
    showToast("Server Unreachable", "Cannot connect to the backend server. Please check if it is running.", "error");
  }
};

export function userCanStart(user) {
  return sharedState.envStartPermission && user?.can_start === true;
}

export function userCanStop(user) {
  return sharedState.envStopPermission && user?.can_stop === true;
}

export function userCanRestart(user) {
  return sharedState.envRestartPermission && user?.can_restart === true;
}

export function userCanDelete(user) {
  return sharedState.envDeletePermission && user?.can_delete === true;
}

export function userCanShell(user) {
  return sharedState.envShellPermission && user?.can_shell === true;
}

export function formatBytes(bytes) {
  if (bytes === 0) return '0B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + sizes[i];
}

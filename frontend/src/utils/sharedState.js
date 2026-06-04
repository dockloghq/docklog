import { reactive, ref } from 'vue';
import { secureStorage } from './storage';

export const sharedState = reactive({
  currentUser: null,
  systemStats: { cpu: 0, usedMemGB: 0, memory: '0 / 0' },
  searchQuery: '',
  theme: secureStorage.getItem('theme') || 'dark',
  showPasswordModal: false,
  forcePasswordChange: false,
  dashboardSidebarOpen: typeof window !== 'undefined' ? window.innerWidth > 1024 : true,
  adminSidebarOpen: typeof window !== 'undefined' ? window.innerWidth > 1024 : true,
  toast: {
    visible: false,
    title: '',
    message: '',
    type: 'success'
  },
  isAuthDisabled: false,
  configLoaded: false,
  envStartPermission: true,
  envStopPermission: true,
  envRestartPermission: true,
  envDeletePermission: true,
});

export const showToast = (title, message, type = 'success') => {
  sharedState.toast.title = title;
  sharedState.toast.message = message;
  sharedState.toast.type = type;
  sharedState.toast.visible = true;
  setTimeout(() => {
    sharedState.toast.visible = false;
  }, 4000);
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
      is_active: true
    };
    return { status: 'ok', user: sharedState.currentUser };
  }
  const token = secureStorage.getItem('token');
  if (!token) return { status: 'missing', user: null };
  try {
    const res = await fetch('/api/user/me', {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (res.ok) {
      sharedState.currentUser = await res.json();
      return { status: 'ok', user: sharedState.currentUser };
    }
    if (res.status === 403) {
      sharedState.currentUser = null;
      secureStorage.removeItem('token');
      secureStorage.removeItem('user');
      return { status: 'forbidden', user: null };
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
    const res = await fetch('/api/system/stats', {
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
        total_memory: data.total_memory
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

export function formatBytes(bytes) {
  if (bytes === 0) return '0B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + sizes[i];
}

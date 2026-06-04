import { createApp } from 'vue';
import './style.css';
import App from './App.vue';
import router from './router';
import { sharedState } from './utils/sharedState';
import { secureStorage } from './utils/storage';

const forceLogout = () => {
  secureStorage.removeItem('token');
  secureStorage.removeItem('user');
  sharedState.currentUser = null;
  sharedState.showPasswordModal = false;
  sharedState.forcePasswordChange = false;

  if (router.currentRoute.value.path !== '/login') {
    router.replace('/login').catch(() => {});
  }
};

if (!window.__docklogFetchPatched) {
  const originalFetch = window.fetch.bind(window);

  window.fetch = async (...args) => {
    const response = await originalFetch(...args);

    if (response?.status === 401) {
      const requestUrl = typeof args[0] === 'string' ? args[0] : (args[0] && (args[0].url || args[0].href || String(args[0])));
      const isLoginRequest = typeof requestUrl === 'string' && requestUrl.includes('/api/token');
      if (!isLoginRequest) {
        forceLogout();
      }
    } else if (response?.status === 403) {
      try {
        const contentType = response.headers.get('content-type') || '';
        if (contentType.includes('application/json')) {
          const payload = await response.clone().json();
          if (payload?.code === 'ACCOUNT_DEACTIVATED') {
            forceLogout();
          }
        }
      } catch {
        // Ignore malformed responses and leave normal error handling intact.
      }
    }

    return response;
  };

  window.__docklogFetchPatched = true;
}

const app = createApp(App);
app.use(router);
app.mount('#app');

import { createRouter, createWebHistory } from 'vue-router';
import { secureStorage, parseJwt } from '../utils/storage';
import { sharedState } from '../utils/sharedState';
import { apiFetch } from '../utils/apiFetch';

const routes = [
  { path: '/', redirect: '/dashboard' },
  { 
    path: '/dashboard', 
    name: 'Dashboard', 
    component: () => import('../views/Dashboard.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Dashboard' }
  },
  { 
    path: '/containers', 
    name: 'Containers', 
    component: () => import('../views/Containers.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Container Management' }
  },
  {
    path: '/containers/:id',
    name: 'ContainerDetail',
    component: () => import('../views/ContainerDetail.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Container Details' }
  },
  { 
    path: '/logs', 
    name: 'Logs', 
    component: () => import('../views/Logs.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Live Log Stream' }
  },
  {
    path: '/shell',
    name: 'Shell',
    component: () => import('../views/Shell.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Container Shell' }
  },
  { 
    path: '/health', 
    name: 'Health', 
    component: () => import('../views/Health.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'System Health' }
  },
  { 
    path: '/admin', 
    name: 'Admin', 
    component: () => import('../views/Admin.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, layout: 'main', title: 'Admin Control Center' }
  },
  { 
    path: '/audit', 
    name: 'Audit', 
    component: () => import('../views/Audit.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, layout: 'main', title: 'Security Audits' }
  },
  { 
    path: '/settings', 
    name: 'Settings', 
    component: () => import('../views/Settings.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Account Settings' }
  },
  { 
    path: '/login', 
    name: 'Login', 
    component: () => import('../views/Login.vue'),
    meta: { title: 'Sign In' }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('../views/NotFound.vue'),
    meta: { title: 'Page Not Found' }
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach(async (to, from, next) => {
  if (!sharedState.configLoaded) {
    try {
      const res = await apiFetch('/api/config');
      if (res.ok) {
        const data = await res.json();
        sharedState.isAuthDisabled = data.auth_disabled === true;
        sharedState.envStartPermission = data.allow_start !== false;
        sharedState.envStopPermission = data.allow_stop !== false;
        sharedState.envRestartPermission = data.allow_restart !== false;
        sharedState.envDeletePermission = data.allow_delete !== false;
        sharedState.envShellPermission = data.allow_shell === true;
      }
    } catch (e) {
      console.error('Failed to load auth config:', e);
    }
    sharedState.configLoaded = true;
  }

  // Update Page Title
  const baseTitle = 'DockLog';
  document.title = to.meta.title ? `${to.meta.title} | ${baseTitle}` : baseTitle;

  if (sharedState.isAuthDisabled) {
    if (to.path === '/login') {
      next('/dashboard');
    } else {
      next();
    }
    return;
  }

  const token = secureStorage.getItem('token');
  const claims = parseJwt(token);
  const isAdmin = claims?.is_admin === true;
  const isExpired = claims?.exp ? (claims.exp * 1000 < Date.now()) : false;

  if (to.meta.requiresAuth && (!token || isExpired)) {
    if (isExpired) {
      secureStorage.removeItem('token');
      secureStorage.removeItem('user');
    }
    next('/login');
  } else if (to.meta.requiresAdmin && !isAdmin) {
    next('/dashboard');
  } else {
    next();
  }
});

export default router;

import { createRouter, createWebHistory } from 'vue-router';
import { secureStorage, parseJwt } from '../utils/storage';
import { sharedState, dockerEnabled, kubernetesEnabled, swarmEnabled } from '../utils/sharedState';
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
    meta: { requiresAuth: true, layout: 'main', title: 'Container Management', requiresDocker: true }
  },
  {
    path: '/swarm',
    name: 'Swarm',
    component: () => import('../views/Swarm.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Docker Swarm', requiresDocker: true }
  },
  {
    path: '/swarm/services/:id',
    name: 'SwarmServiceDetail',
    component: () => import('../views/SwarmServiceDetail.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Swarm Service Details', requiresDocker: true }
  },
  {
    path: '/kubernetes',
    name: 'Kubernetes',
    component: () => import('../views/Kubernetes.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Kubernetes', requiresKubernetes: true }
  },
  {
    path: '/pods',
    redirect: (to) => ({ path: '/kubernetes', query: { ...to.query, tab: to.query.tab || 'pods' } }),
  },
  {
    path: '/containers/:id',
    name: 'ContainerDetail',
    component: () => import('../views/ContainerDetail.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Container Details', requiresDocker: true }
  },
  {
    path: '/kubernetes/pods/:namespace/:pod',
    name: 'PodDetail',
    component: () => import('../views/PodDetail.vue'),
    meta: { requiresAuth: true, layout: 'main', title: 'Pod Details', requiresKubernetes: true }
  },
  {
    path: '/pod-detail/:namespace/:pod',
    redirect: (to) => `/kubernetes/pods/${to.params.namespace}/${to.params.pod}`,
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
    path: '/users',
    redirect: '/admin'
  },
  {
    path: '/audit',
    name: 'Audit',
    component: () => import('../views/Audit.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, layout: 'main', title: 'Security Audits' }
  },
  {
    path: '/notifications',
    name: 'Notifications',
    component: () => import('../views/Notifications.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, layout: 'main', title: 'Notifications' }
  },
  {
    path: '/alerts',
    name: 'Alerts',
    component: () => import('../views/Alerts.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, layout: 'main', title: 'Alerts' }
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
        if (data.version) {
          sharedState.appVersion = data.version;
        }
        sharedState.isAuthDisabled = data.auth_disabled === true;
        sharedState.envStartPermission = data.allow_start !== false;
        sharedState.envStopPermission = data.allow_stop !== false;
        sharedState.envRestartPermission = data.allow_restart !== false;
        sharedState.envDeletePermission = data.allow_delete !== false;
        sharedState.envShellPermission = data.allow_shell === true;
        sharedState.runtimeMode = data.runtime_mode || 'docker';
        sharedState.swarmActive = data.swarm_active === true;
        sharedState.k8sNamespaces = Array.isArray(data.k8s_namespaces) ? data.k8s_namespaces : [];
        sharedState.k8sDefaultNs = data.k8s_default_ns || 'default';
        sharedState.k8sAvailable = data.k8s_available === true;
        sharedState.k8sError = data.k8s_error || '';
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
  } else if (to.meta.requiresDocker && !dockerEnabled()) {
    next(kubernetesEnabled() ? '/kubernetes' : '/dashboard');
  } else if (to.meta.requiresKubernetes && !kubernetesEnabled()) {
    next(dockerEnabled() ? '/containers' : '/dashboard');
  } else if (
    to.meta.requiresAdmin &&
    sharedState.currentUser?.password_changed === false
  ) {
    sharedState.forcePasswordChange = true;
    sharedState.showPasswordModal = true;
    next('/dashboard');
  } else {
    next();
  }
});

export default router;

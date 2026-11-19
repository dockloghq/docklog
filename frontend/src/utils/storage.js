// Simple obfuscation for client-side storage
// Note: This is not "military-grade" encryption, but prevents casual tampering via dev tools.

const KEY_PREFIX = 'dl_';

// In-memory fallback if localStorage is blocked (e.g. Firefox Strict Privacy / Private Browsing)
const memoryStore = new Map();

function safeGetStorage(key) {
  try {
    if (typeof localStorage !== 'undefined') {
      return localStorage.getItem(key);
    }
  } catch {
    // LocalStorage blocked or restricted
  }
  return memoryStore.get(key) || null;
}

function safeSetStorage(key, value) {
  try {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(key, value);
    }
  } catch {
    // LocalStorage blocked or restricted
  }
  memoryStore.set(key, value);
}

function safeRemoveStorage(key) {
  try {
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(key);
    }
  } catch {
    // LocalStorage blocked or restricted
  }
  memoryStore.delete(key);
}

export const secureStorage = {
  setItem: (key, value) => {
    try {
      const stringValue = typeof value === 'object' ? JSON.stringify(value) : String(value);
      const obfuscatedValue = btoa(encodeURIComponent(stringValue));
      safeSetStorage(KEY_PREFIX + key, obfuscatedValue);
    } catch (e) {
      console.warn('Failed to set item in secureStorage:', e);
    }
  },
  
  getItem: (key) => {
    try {
      const value = safeGetStorage(KEY_PREFIX + key);
      if (!value) return null;
      const decodedValue = decodeURIComponent(atob(value));
      if (decodedValue === 'true') return true;
      if (decodedValue === 'false') return false;
      try {
        return JSON.parse(decodedValue);
      } catch {
        return decodedValue;
      }
    } catch {
      return null;
    }
  },
  
  removeItem: (key) => {
    try {
      safeRemoveStorage(KEY_PREFIX + key);
    } catch (e) {
      console.warn('Failed to remove item in secureStorage:', e);
    }
  },
  
  clear: () => {
    try {
      if (typeof localStorage !== 'undefined') {
        Object.keys(localStorage).forEach(k => {
          if (k.startsWith(KEY_PREFIX)) localStorage.removeItem(k);
        });
      }
    } catch {
      // Ignore
    }
    memoryStore.clear();
  }
};

export const parseJwt = (token) => {
  if (!token) return null;
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
    return JSON.parse(jsonPayload);
  } catch (e) {
    return null;
  }
};

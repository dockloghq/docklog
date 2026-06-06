<template>
  <div class="login-overlay">
    <div class="login-card glass animate-slide-up">
      <div class="login-header">
        <div class="logo-box">
          <img :src="logoSrc" alt="DockLog" class="login-logo-img" />
        </div>
        <h1>DockLog</h1>
        <p class="text-mute">Enterprise container observability</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="input-group">
          <label>Username</label>
          <div class="premium-input-wrapper">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <input v-model="username" type="text" placeholder="e.g. admin" required />
          </div>
        </div>

        <div class="input-group">
          <label>Password</label>
          <div class="premium-input-wrapper">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.5">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
              <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
            </svg>
            <input v-model="password" type="password" placeholder="••••••••" required />
          </div>
        </div>

        <button type="submit" :disabled="loading" class="premium-btn primary full-width login-btn">
          {{ loading ? "Authenticating..." : "Access Dashboard" }}
        </button>

        <Transition name="fade">
          <p v-if="error" class="error-msg">{{ error }}</p>
        </Transition>
      </form>

      <div class="login-footer">
        <span>v{{ appVersion }}</span>
        <div class="dot-sep"></div>
        <span>Secure Protocol</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import { secureStorage } from "../utils/storage";
import { apiFetch } from "../utils/apiFetch";
import { sharedState } from "../utils/sharedState";
import pkg from "../../package.json";

const appVersion = pkg.version;

const router = useRouter();
const username = ref("");
const password = ref("");
const loading = ref(false);
const error = ref("");

const logoSrc = computed(() =>
  sharedState.theme === "light" ? "/logo-icon-light.png" : "/logo-icon-dark.png",
);

const handleLogin = async () => {
  loading.value = true;
  error.value = "";

  try {
    const formData = new FormData();
    formData.append("username", username.value.trim());
    formData.append("password", password.value);

    const res = await apiFetch("/api/token", {
      method: "POST",
      body: formData,
    });

    if (res.ok) {
      const data = await res.json();
      secureStorage.setItem("token", data.access_token);
      sharedState.forcePasswordChange = data.password_changed === false;
      sharedState.showPasswordModal = data.password_changed === false;
      router.push("/dashboard");
    } else {
      let message = "Invalid credentials";
      try {
        const data = await res.json();
        if (data?.error) {
          message = data.error;
        }
      } catch {
        // keep default message
      }
      if (res.status === 429) {
        message = "Too many login attempts. Try again in 15 minutes.";
      }
      error.value = message;
    }
  } catch (err) {
    error.value = "System connection failed";
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.login-overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-main);
  background-image:
    radial-gradient(ellipse 60% 50% at 20% 0%, rgba(var(--accent-rgb), 0.18), transparent 55%),
    radial-gradient(ellipse 50% 40% at 100% 100%, rgba(var(--success-rgb), 0.08), transparent 50%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.login-overlay::before {
  content: "";
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(var(--border) 1px, transparent 1px),
    linear-gradient(90deg, var(--border) 1px, transparent 1px);
  background-size: 48px 48px;
  opacity: 0.12;
  mask-image: radial-gradient(circle at 50% 40%, black, transparent 75%);
  pointer-events: none;
}

.login-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 440px;
  padding: 2.75rem 2.5rem;
  border-radius: var(--radius-2xl);
  border: 1px solid var(--border);
  background: var(--glass-bg);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  box-shadow: 0 24px 64px -16px var(--shadow);
}

.login-header {
  text-align: center;
  margin-bottom: 2.5rem;
}

.logo-box {
  width: 88px;
  height: 88px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.25rem;
}

.login-logo-img {
  width: 88px;
  height: 88px;
  object-fit: contain;
  border-radius: var(--radius-lg);
  filter: drop-shadow(0 12px 28px rgba(var(--accent-rgb), 0.25));
}

.login-header h1 {
  font-size: 2rem;
  font-weight: 800;
  letter-spacing: -0.04em;
  color: var(--text-main);
  margin: 0;
}

.login-header p {
  font-size: 0.9rem;
  font-weight: 500;
  margin-top: 0.5rem;
  color: var(--text-mute);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.input-group label {
  display: block;
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin-bottom: 0.6rem;
}

.premium-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.premium-input-wrapper svg {
  position: absolute;
  left: 1rem;
  color: var(--text-mute);
  transition: color 0.2s;
  pointer-events: none;
}

.premium-input-wrapper input {
  width: 100%;
  padding: 0.95rem 1rem 0.95rem 3rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  color: var(--text-main);
  font-size: 0.95rem;
  font-weight: 500;
  transition: border-color 0.2s, box-shadow 0.2s, background 0.2s;
}

.premium-input-wrapper input:focus {
  outline: none;
  background: var(--bg-subtle);
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(var(--accent-rgb), 0.12);
}

.premium-input-wrapper:focus-within svg {
  color: var(--accent);
}

.login-btn {
  height: 52px;
  font-size: 0.95rem;
  margin-top: 0.5rem;
}

.error-msg {
  text-align: center;
}

.login-footer {
  margin-top: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.dot-sep {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--border);
}

@media (max-width: 480px) {
  .login-card {
    padding: 2rem 1.5rem;
    border-radius: var(--radius-xl);
  }
  .login-header h1 {
    font-size: 1.65rem;
  }
  .login-header {
    margin-bottom: 2rem;
  }
  .logo-box {
    width: 72px;
    height: 72px;
  }
  .login-logo-img {
    width: 72px;
    height: 72px;
  }
}
</style>

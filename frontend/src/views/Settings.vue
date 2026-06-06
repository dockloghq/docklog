<template>
  <div class="settings-view animate-fade-in">
    <div class="view-header">
      <div class="header-info">
        <h1>Account Settings</h1>
        <p class="text-mute">Manage your profile, security, and preferences</p>
      </div>
    </div>

    <div class="settings-grid mt-8">
      <!-- Profile Section -->
      <div class="settings-card glass">
        <div class="card-header">
          <svg
            viewBox="0 0 24 24"
            width="20"
            height="20"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
            <circle cx="12" cy="7" r="4"></circle>
          </svg>
          <h3>Profile Information</h3>
        </div>
        <div class="card-body">
          <div class="info-row">
            <label>Username</label>
            <div class="value-box">{{ sharedState.currentUser?.username }}</div>
          </div>
          <div class="info-row">
            <label>Role</label>
            <div class="value-box">
              <span
                :class="[
                  'badge',
                  sharedState.currentUser?.is_admin
                    ? 'badge-warning'
                    : 'badge-dim',
                ]"
              >
                {{
                  sharedState.currentUser?.is_admin ? "Administrator" : "Staff"
                }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Security Section -->
      <div class="settings-card glass">
        <div class="card-header">
          <svg
            viewBox="0 0 24 24"
            width="20"
            height="20"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
            <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
          </svg>
          <h3>Security & Password</h3>
        </div>
        <div class="card-body">
          <form @submit.prevent="handlePasswordUpdate" class="settings-form">
            <div class="input-group">
              <label>Current Password</label>
              <input
                type="password"
                v-model="currentPassword"
                placeholder="Enter current password"
                class="premium-input"
                required
              />
            </div>
            <div class="input-group">
              <label>New Password</label>
              <input
                type="password"
                v-model="newPassword"
                placeholder="Enter new password"
                class="premium-input"
                required
              />
            </div>
            <div class="input-group">
              <label>Confirm Password</label>
              <input
                type="password"
                v-model="confirmPassword"
                placeholder="Confirm new password"
                class="premium-input"
                required
              />
            </div>
            <button
              type="submit"
              :disabled="loading"
              class="premium-btn primary full-width mt-4"
            >
              {{ loading ? "Updating..." : "Update Password" }}
            </button>
            <p v-if="error" class="error-msg mt-4">{{ error }}</p>
          </form>
        </div>
      </div>

      <!-- System Preferences -->
      <div class="settings-card glass">
        <div class="card-header">
          <svg
            viewBox="0 0 24 24"
            width="20"
            height="20"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
          >
            <circle cx="12" cy="12" r="3"></circle>
            <path
              d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
            ></path>
          </svg>
          <h3>System Preferences</h3>
        </div>
        <div class="card-body">
          <div class="pref-row theme-pref-row">
            <div class="pref-info">
              <h4>Appearance</h4>
              <p class="text-mute">{{ themeDescription }}</p>
            </div>
            <div class="theme-options">
              <button
                v-for="option in themeOptions"
                :key="option.value"
                type="button"
                :class="[
                  'theme-option',
                  { active: sharedState.themePreference === option.value },
                ]"
                @click="applyTheme(option.value)"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { sharedState, showToast, applyTheme } from "../utils/sharedState";
import { secureStorage } from "../utils/storage";
import { apiFetch } from "../utils/apiFetch";

const themeOptions = [
  { value: "auto", label: "Auto" },
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
];

const themeDescription = computed(() => {
  if (sharedState.themePreference === "auto") {
    return `Auto mode — currently using ${sharedState.theme} (matches your system)`;
  }
  return `Using ${sharedState.themePreference} mode`;
});

const newPassword = ref("");
const confirmPassword = ref("");
const currentPassword = ref("");
const loading = ref(false);
const error = ref("");

const handlePasswordUpdate = async () => {
  if (newPassword.value !== confirmPassword.value) {
    error.value = "Passwords do not match";
    return;
  }
  if (newPassword.value.length < 8) {
    error.value = "Password must be at least 8 characters";
    return;
  }
  if (!currentPassword.value) {
    error.value = "Current password is required";
    return;
  }

  loading.value = true;
  error.value = "";

  try {
    const token = secureStorage.getItem("token");
    const formData = new FormData();
    formData.append("password", newPassword.value);
    formData.append("current_password", currentPassword.value);

    const res = await apiFetch("/api/user/change-password", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: formData,
    });

    if (res.ok) {
      showToast("Success", "Password updated successfully", "success");
      newPassword.value = "";
      confirmPassword.value = "";
      currentPassword.value = "";
    } else {
      const data = await res.json();
      error.value = data.error || "Failed to update password";
    }
  } catch (err) {
    error.value = "System connection failed";
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.settings-view {
  height: 100%;
}

.view-header h1 {
  font-size: 1.75rem;
  font-weight: 950;
  letter-spacing: -0.05em;
  color: var(--text-main);
  margin: 0;
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 2rem;
}

.settings-card {
  padding: 1.75rem;
  border-radius: var(--radius-xl);
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  border: 1px solid var(--border);
  background: var(--bg-card);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  color: var(--accent);
}

.card-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 900;
  color: var(--text-main);
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.info-row {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-row label {
  font-size: 0.75rem;
  font-weight: 800;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.value-box {
  padding: 1rem 1.25rem;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  border-radius: 14px;
  color: var(--text-main);
  font-weight: 600;
}

.pref-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem;
  background: var(--bg-subtle);
  border-radius: 20px;
  border: 1px solid var(--border);
}

.pref-info h4 {
  margin: 0;
  font-size: 1rem;
  font-weight: 800;
  color: var(--text-main);
}

.pref-info p {
  margin: 0.2rem 0 0;
  font-size: 0.8rem;
}

/* Premium Inputs & Buttons */
.premium-input {
  width: 100%;
  padding: 1rem 1.25rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 14px;
  color: var(--text-main);
  font-size: 0.9rem;
  font-weight: 600;
  outline: none;
  transition: all 0.2s;
}

.premium-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(var(--accent-rgb), 0.12);
}

.premium-btn {
  padding: 0.85rem 1.5rem;
  border-radius: 14px;
  border: none;
  font-weight: 850;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.premium-btn.primary {
  background: var(--accent);
  color: #fff;
  box-shadow: 0 8px 16px rgba(var(--accent-rgb), 0.2);
}

.premium-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  filter: brightness(1.1);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.2);
}

.premium-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.full-width {
  width: 100%;
}
.mt-4 {
  margin-top: 1rem;
}

.pref-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem;
  background: var(--bg-input);
  border-radius: 16px;
  border: 1px solid var(--border);
}

.theme-pref-row {
  flex-direction: column;
  align-items: stretch;
  gap: 1.25rem;
}

.theme-options {
  display: flex;
  gap: 0.5rem;
  padding: 0.35rem;
  background: var(--bg-subtle);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
}

.theme-option {
  flex: 1;
  padding: 0.65rem 0.85rem;
  border-radius: calc(var(--radius-md) - 2px);
  border: none;
  background: transparent;
  color: var(--text-mute);
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  transition: background 0.2s ease, color 0.2s ease, box-shadow 0.2s ease;
}

.theme-option:hover {
  color: var(--text-main);
  background: var(--bg-subtle);
}

.theme-option.active {
  background: var(--accent);
  color: #fff;
  box-shadow: 0 4px 12px rgba(var(--accent-rgb), 0.28);
}

@media (max-width: 768px) {
  .view-header h1 {
    font-size: 1.25rem;
  }
  .settings-grid {
    grid-template-columns: 1fr;
    gap: 1.25rem;
  }
  .settings-card {
    padding: 1.5rem;
    border-radius: 24px;
  }
  .pref-row {
    flex-direction: row;
    align-items: center;
    gap: 1rem;
  }
  .theme-pref-row {
    flex-direction: column;
    align-items: stretch;
  }
  .pref-info h4 {
    font-size: 0.9rem;
  }
}

@media (max-width: 480px) {
  .settings-card {
    padding: 1.25rem;
    gap: 1.5rem;
    border-radius: 20px;
  }
  .card-header h3 {
    font-size: 1.1rem;
  }
  .premium-input {
    padding: 0.85rem 1rem;
  }
  .view-header h1 {
    font-size: 1rem;
  }
  .info-row label {
    font-size: 0.65rem;
  }
}
</style>

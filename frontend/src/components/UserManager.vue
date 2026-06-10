<template>
  <div class="user-manager">
    <div class="premium-table-container" :class="{ embedded }">
      <table class="premium-table admin-table">
        <thead>
          <tr>
            <th>User</th>
            <th>Role</th>
            <th>Permissions</th>
            <th>Status</th>
            <th class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody v-if="nonAdminUsers.length > 0">
          <tr v-for="u in nonAdminUsers" :key="u.id">
            <td data-label="User">
              <div class="user-cell">
                <div class="mini-avatar">{{ u.username[0]?.toUpperCase() }}</div>
                <div class="user-info">
                  <span class="user-name">{{ u.username }}</span>
                </div>
              </div>
            </td>
            <td data-label="Role">
              <span
                :class="['badge', u.is_admin ? 'badge-warning' : 'badge-dim']"
              >
                {{ u.is_admin ? "ADMIN" : "STAFF" }}
              </span>
            </td>
            <td data-label="Permissions">
              <div class="perm-tags">
                <span v-if="u.is_admin" class="badge badge-success"
                  >ALL ACCESS</span
                >
                <template v-else>
                  <span v-if="u.can_start" class="badge badge-dim mini"
                    >START</span
                  >
                  <span v-if="u.can_stop" class="badge badge-dim mini"
                    >STOP</span
                  >
                  <span v-if="u.can_restart" class="badge badge-dim mini"
                    >RESTART</span
                  >
                  <span
                    v-if="envShellPermission && u.can_shell"
                    class="badge badge-dim mini"
                    >SHELL</span
                  >
                  <span
                    v-if="!u.can_start && !u.can_stop && !u.can_restart && !(envShellPermission && u.can_shell)"
                    class="badge badge-dim mini"
                    >READ-ONLY</span
                  >
                </template>
              </div>
            </td>
            <td data-label="Status">
              <div
                :class="[
                  'premium-toggle',
                  { active: u.is_active, disabled: u.is_admin },
                ]"
                @click="!u.is_admin && toggleUserStatus(u)"
              >
                <div class="toggle-rail">
                  <div class="toggle-handle"></div>
                </div>
                <span class="status-label">{{
                  u.is_active ? "Active" : "Disabled"
                }}</span>
              </div>
            </td>
            <td class="text-right" data-label="Actions">
              <div class="action-group justify-end" v-if="!u.is_admin">
                <button
                  @click="openResetPassword(u)"
                  class="icon-btn"
                  data-tooltip="Reset Password"
                >
                  <svg
                    viewBox="0 0 24 24"
                    width="14"
                    height="14"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <rect
                      x="3"
                      y="11"
                      width="18"
                      height="11"
                      rx="2"
                      ry="2"
                    ></rect>
                    <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                  </svg>
                </button>
                <button
                  @click="openPermissions(u)"
                  class="icon-btn"
                  data-tooltip="Manage Permissions"
                >
                  <svg
                    viewBox="0 0 24 24"
                    width="14"
                    height="14"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <path
                      d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"
                    ></path>
                  </svg>
                </button>
                <button
                  @click="openDeleteConfirm(u)"
                  class="icon-btn stop"
                  data-tooltip="Delete User"
                >
                  <svg
                    viewBox="0 0 24 24"
                    width="14"
                    height="14"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                  >
                    <polyline points="3 6 5 6 21 6"></polyline>
                    <path
                      d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                    ></path>
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
        <tbody v-else>
          <tr>
            <td colspan="5">
              <div class="empty-state-wrapper">
                <div class="empty-state-content">
                  <div class="empty-icon-box">
                    <svg
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="1.5"
                    >
                      <path
                        d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"
                      ></path>
                      <circle cx="9" cy="7" r="4"></circle>
                      <line x1="23" y1="11" x2="17" y2="11"></line>
                    </svg>
                  </div>
                  <h4 class="empty-title">No Staff Members</h4>
                  <p class="empty-text">
                    Click the 'Add User' button to create your first staff
                    account.
                  </p>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- IMPROVED CREATE/EDIT MODAL -->
    <Teleport to="body">
      <Transition name="modal-bounce">
        <div
          v-if="showCreateModal || showPermissionsModal"
          class="modal-overlay"
        >
          <div class="modal-card wide-modal glass shadow-2xl">
            <div class="modal-card-header">
              <div class="header-content">
                <div class="header-icon">
                  <svg
                    v-if="showCreateModal"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                    <circle cx="9" cy="7" r="4"></circle>
                    <line x1="19" y1="8" x2="19" y2="14"></line>
                    <line x1="16" y1="11" x2="22" y2="11"></line>
                  </svg>
                  <svg
                    v-else
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <path
                      d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"
                    ></path>
                  </svg>
                </div>
                <div>
                  <h3 class="modal-title">
                    {{
                      showCreateModal ? "New Staff Member" : "Edit Permissions"
                    }}
                  </h3>
                  <p class="modal-subtitle">
                    {{
                      showCreateModal
                        ? "Configure credentials and access rights"
                        : `Updating access for ${editingUser?.username}`
                    }}
                  </p>
                </div>
              </div>
              <button class="close-btn" @click="closeAllModals" aria-label="Close">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>

            <div class="modal-card-body">
              <div v-if="showCreateModal" class="form-grid dual">
                <div class="input-group">
                  <label class="label-caps">Username</label>
                  <input
                    type="text"
                    v-model="newUser.username"
                    class="premium-input"
                    placeholder="e.g. john_doe"
                  />
                </div>
                <div class="input-group">
                  <label class="label-caps">Initial Password</label>
                  <input
                    type="password"
                    v-model="newUser.password"
                    class="premium-input"
                    placeholder="••••••••"
                  />
                </div>
              </div>

              <div class="perm-section">
                <label class="label-caps">Container Visibility</label>
                <div class="access-toggle-container">
                  <button
                    :class="['access-choice', { active: isRestricted }]"
                    @click="setRestricted(true)"
                  >
                    <span class="dot"></span> Restricted Access
                  </button>
                  <button
                    :class="['access-choice', { active: !isRestricted }]"
                    @click="setRestricted(false)"
                  >
                    <span class="dot"></span> Full Visibility
                  </button>
                </div>

                <Transition name="slide-down">
                  <div v-if="isRestricted" class="pattern-box">
                    <label class="label-caps">Allowed Patterns</label>
                    <div class="pattern-input-wrap">
                      <input
                        type="text"
                        v-model="activeUser.allowed_containers"
                        class="premium-input"
                        placeholder="e.g. api-*, prod-web, ^front.*"
                        autocomplete="off"
                        @focus="onPatternFocus"
                        @blur="hidePatternSuggestions"
                        @input="patternSuggestionsOpen = true"
                      />
                      <Transition name="fade">
                        <ul
                          v-if="patternSuggestionsOpen && filteredPatternSuggestions.length"
                          class="pattern-suggestions"
                          role="listbox"
                        >
                          <li
                            v-for="name in filteredPatternSuggestions"
                            :key="name"
                            role="option"
                          >
                            <button
                              type="button"
                              class="pattern-suggestion-btn"
                              @mousedown.prevent="applyPatternSuggestion(name)"
                            >
                              <span class="suggestion-dot"></span>
                              {{ name }}
                            </button>
                          </li>
                        </ul>
                      </Transition>
                    </div>

                    <div v-if="runningContainerNames.length" class="pattern-fleet mt-3">
                      <div class="pattern-fleet-head">
                        <span class="label-caps label-caps-inline">Running containers</span>
                        <span class="pattern-fleet-count">{{ availableFleetNames.length }} available</span>
                        <button
                          v-if="runningContainerNames.length > fleetInlineLimit"
                          type="button"
                          class="pattern-fleet-toggle"
                          @click="fleetExpanded = !fleetExpanded"
                        >
                          {{ fleetExpanded ? "Hide list" : "Browse all" }}
                        </button>
                      </div>
                      <p
                        v-if="runningContainerNames.length > fleetInlineLimit && !fleetExpanded"
                        class="pattern-fleet-hint"
                      >
                        Type in the field for quick picks, or browse the full running list.
                      </p>
                      <div
                        v-show="showFleetList"
                        class="pattern-fleet-chips"
                        :class="{ scrollable: runningContainerNames.length > fleetInlineLimit }"
                      >
                        <button
                          v-for="name in availableFleetNames"
                          :key="name"
                          type="button"
                          class="pattern-fleet-chip"
                          @click="applyPatternSuggestion(name)"
                        >
                          {{ name }}
                        </button>
                        <span
                          v-if="!availableFleetNames.length"
                          class="pattern-fleet-all-added"
                        >
                          All running containers are already in the pattern.
                        </span>
                      </div>
                    </div>
                    <p v-else-if="containersLoaded" class="pattern-fleet-empty mt-3">
                      No running containers on this host right now.
                    </p>

                    <div class="pattern-examples mt-3">
                      <div class="example-item">
                        <code class="tag">api-*</code>
                        <span class="desc">Wildcard (matches api-v1, api-db, etc)</span>
                      </div>
                      <div class="example-item">
                        <code class="tag">^prod-.*</code>
                        <span class="desc">Regex (advanced matching)</span>
                      </div>
                      <div class="example-item">
                        <code class="tag">mysql, redis</code>
                        <span class="desc">Multiple (comma separated)</span>
                      </div>
                    </div>
                  </div>
                </Transition>
              </div>

              <div class="perm-section perm-section-compact">
                <label class="label-caps">Action Rights</label>
                <div class="modern-rights-grid">
                  <label
                    v-for="action in actionRights"
                    :key="action.key"
                    class="right-card"
                    :class="{
                      checked:
                        activeUser[action.field] ||
                        activeUser[action.createField],
                    }"
                  >
                    <input
                      type="checkbox"
                      v-model="
                        activeUser[
                          showCreateModal ? action.createField : action.field
                        ]
                      "
                    />
                    <div class="right-card-content">
                      <span class="right-label">{{ action.label }}</span>
                      <div class="custom-check"></div>
                    </div>
                  </label>
                </div>
              </div>
            </div>

            <div class="modal-card-footer">
              <button @click="closeAllModals" class="btn-secondary">
                Cancel
              </button>
              <button
                @click="showCreateModal ? createUser() : updatePermissions()"
                class="btn-primary"
              >
                {{ showCreateModal ? "Create Account" : "Save Changes" }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- DELETE CONFIRMATION MODAL -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showDeleteModal" class="modal-overlay">
          <div class="modal-content shadow-2xl">
            <div class="modal-icon error">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <path
                  d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                ></path>
              </svg>
            </div>
            <div class="modal-text-center">
              <h3>Delete Account?</h3>
              <p>Permanently remove <strong>{{ userToDelete?.username }}</strong>?</p>
            </div>
            <div class="modal-divider"></div>
            <div class="modal-actions">
              <button
                @click="closeAllModals"
                class="modal-btn cancel flex-1"
              >
                Keep User
              </button>
              <button @click="confirmDelete" class="modal-btn confirm error flex-1">
                Yes, Delete
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- RESET PASSWORD MODAL -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showResetModal" class="modal-overlay">
          <div class="modal-content shadow-2xl">
            <div class="modal-icon warning">
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
            <div class="modal-text-center">
              <h3>Reset Password</h3>
              <p>Update credentials for <strong>{{ resetTargetUser?.username }}</strong></p>
            </div>
            
            <div class="modal-body">
              <div class="input-group">
                <label class="label-caps">New Secure Password</label>
                <input
                  type="password"
                  v-model="resetPassword"
                  class="premium-input"
                  placeholder="••••••••"
                  @keyup.enter="confirmResetPassword"
                />
                <p class="hint-text mt-2 text-center">
                  User will be forced to change this upon login.
                </p>
              </div>
            </div>

            <div class="modal-divider"></div>

            <div class="modal-actions">
              <button @click="closeAllModals" class="modal-btn cancel">
                Cancel
              </button>
              <button @click="confirmResetPassword" class="modal-btn confirm warning">
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
import { ref, computed, onMounted } from "vue";
import { showToast } from "../utils/sharedState";
import { sharedState } from "../utils/sharedState";
import { apiFetch } from "../utils/apiFetch";

const props = defineProps({
  token: String,
  embedded: { type: Boolean, default: false },
});
const emit = defineEmits(["update-count"]);

const envShellPermission = computed(() => sharedState.envShellPermission === true);

const baseActionRights = [
  { key: "start", label: "Start", field: "can_start", createField: "canStart" },
  { key: "stop", label: "Stop", field: "can_stop", createField: "canStop" },
  { key: "restart", label: "Restart", field: "can_restart", createField: "canRestart" },
  { key: "delete", label: "Delete", field: "can_delete", createField: "canDelete" },
];

const actionRights = computed(() => {
  if (!envShellPermission.value) {
    return baseActionRights;
  }
  return [
    ...baseActionRights,
    { key: "shell", label: "Shell", field: "can_shell", createField: "canShell" },
  ];
});

const staffUsers = ref([]);
const showCreateModal = ref(false);
const showPermissionsModal = ref(false);
const showDeleteModal = ref(false);
const showResetModal = ref(false);
const resetPassword = ref("");

const newUser = ref({
  username: "",
  password: "",
  is_restricted: false,
  allowed_containers: ".*",
  canStart: true,
  canStop: true,
  canRestart: true,
  canDelete: false,
  canShell: false,
});
const editingUser = ref({});
const resetTargetUser = ref(null);
const userToDelete = ref(null);

const activeUser = computed(() =>
  showCreateModal.value ? newUser.value : editingUser.value,
);
const nonAdminUsers = computed(() =>
  staffUsers.value.filter((u) => !u.is_admin),
);
const isRestricted = computed(() =>
  showCreateModal.value
    ? newUser.value.is_restricted
    : editingUser.value.is_restricted_access,
);

const setRestricted = (val) => {
  if (showCreateModal.value) newUser.value.is_restricted = val;
  else editingUser.value.is_restricted_access = val;
  if (val) fetchRunningContainers();
};

const fleetContainers = ref([]);
const containersLoaded = ref(false);
const patternSuggestionsOpen = ref(false);
const fleetExpanded = ref(false);
const fleetInlineLimit = 6;
let patternBlurTimer = null;

const runningContainerNames = computed(() =>
  fleetContainers.value
    .filter((c) => c.state === "running")
    .map((c) => c.name)
    .sort((a, b) => a.localeCompare(b)),
);

const selectedPatternNames = computed(() => {
  const val = activeUser.value?.allowed_containers || "";
  if (!val.trim() || val.trim() === ".*") return new Set();
  return new Set(
    val
      .split(",")
      .map((p) => p.trim())
      .filter(Boolean),
  );
});

const availableFleetNames = computed(() =>
  runningContainerNames.value.filter((n) => !selectedPatternNames.value.has(n)),
);

const showFleetList = computed(
  () =>
    runningContainerNames.value.length <= fleetInlineLimit || fleetExpanded.value,
);

const patternQuery = computed(() => {
  const val = activeUser.value?.allowed_containers || "";
  const segment = val.split(",").pop()?.trim() || "";
  return segment === ".*" ? "" : segment;
});

const filteredPatternSuggestions = computed(() => {
  const q = patternQuery.value.toLowerCase();
  const selected = new Set(
    (activeUser.value?.allowed_containers || "")
      .split(",")
      .map((p) => p.trim())
      .filter(Boolean),
  );
  let names = runningContainerNames.value.filter((n) => !selected.has(n));
  if (q) {
    names = names.filter((n) => n.toLowerCase().includes(q));
  }
  return names.slice(0, 8);
});

const fetchRunningContainers = async () => {
  try {
    const res = await apiFetch("/api/containers", {
      headers: { Authorization: `Bearer ${props.token}` },
    });
    if (res.ok) {
      fleetContainers.value = await res.json();
    }
  } catch (err) {
    console.error(err);
  } finally {
    containersLoaded.value = true;
  }
};

const onPatternFocus = () => {
  clearTimeout(patternBlurTimer);
  patternSuggestionsOpen.value = true;
  fetchRunningContainers();
};

const hidePatternSuggestions = () => {
  patternBlurTimer = setTimeout(() => {
    patternSuggestionsOpen.value = false;
  }, 150);
};

const applyPatternSuggestion = (name) => {
  const target = activeUser.value;
  const val = (target.allowed_containers || "").trim();
  const existing = val && val !== ".*"
    ? val.split(",").map((p) => p.trim()).filter(Boolean)
    : [];

  if (existing.includes(name)) return;

  const query = patternQuery.value;
  if (!existing.length) {
    target.allowed_containers = name;
    return;
  }

  const lastIdx = existing.length - 1;
  if (
    query &&
    existing[lastIdx].toLowerCase().includes(query.toLowerCase()) &&
    existing[lastIdx] !== name
  ) {
    existing[lastIdx] = name;
    target.allowed_containers = existing.join(", ");
    return;
  }

  target.allowed_containers = [...existing, name].join(", ");
};

const closeAllModals = () => {
  showCreateModal.value = false;
  showPermissionsModal.value = false;
  showDeleteModal.value = false;
  showResetModal.value = false;
  resetPassword.value = "";
  resetTargetUser.value = null;
  patternSuggestionsOpen.value = false;
  fleetExpanded.value = false;
  containersLoaded.value = false;
  fleetContainers.value = [];
};

const fetchStaff = async () => {
  try {
    const res = await apiFetch("/api/admin/users", {
      headers: { Authorization: `Bearer ${props.token}` },
    });
    if (res.ok) {
      const data = await res.json();
      staffUsers.value = data.users || [];
      emit("update-count", nonAdminUsers.value.length);
    }
  } catch (err) {
    console.error(err);
  }
};

const createUser = async () => {
  if (!newUser.value.username || !newUser.value.password) return;
  try {
    const formData = new FormData();
    formData.append("username", newUser.value.username);
    formData.append("password", newUser.value.password);
    formData.append("can_start", newUser.value.canStart ? "true" : "false");
    formData.append("can_stop", newUser.value.canStop ? "true" : "false");
    formData.append("can_restart", newUser.value.canRestart ? "true" : "false");
    formData.append("can_delete", newUser.value.canDelete ? "true" : "false");
    formData.append("can_shell", newUser.value.canShell ? "true" : "false");
    formData.append(
      "is_restricted_access",
      newUser.value.is_restricted ? "true" : "false",
    );
    formData.append(
      "allowed_containers",
      newUser.value.is_restricted ? newUser.value.allowed_containers : ".*",
    );

    const res = await apiFetch("/api/admin/users", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${props.token}`,
      },
      body: formData,
    });

    if (res.ok) {
      showToast("Success", "User created", "success");
      closeAllModals();
      newUser.value = {
        username: "",
        password: "",
        canStart: true,
        canStop: true,
        canRestart: true,
        canDelete: false,
        canShell: false,
        is_restricted: false,
        allowed_containers: ".*",
      };
      fetchStaff();
    } else {
      const errorData = await res.json().catch(() => ({}));
      showToast(
        "Creation Failed",
        errorData.error || "Could not create user account",
        "error",
      );
    }
  } catch (err) {
    console.error(err);
    showToast("Error", "A network error occurred", "error");
  }
};

const toggleUserStatus = async (user) => {
  try {
    const formData = new FormData();
    formData.append("is_active", !user.is_active ? "true" : "false");

    const res = await apiFetch(`/api/admin/users/${user.id}/active`, {
      method: "PUT",
      headers: { Authorization: `Bearer ${props.token}` },
      body: formData,
    });
    if (res.ok) {
      user.is_active = !user.is_active;
      showToast(
        "Updated",
        `User ${user.is_active ? "enabled" : "disabled"}`,
        "success",
      );
    }
  } catch (err) {
    console.error(err);
  }
};

const openPermissions = (user) => {
  editingUser.value = JSON.parse(JSON.stringify(user));
  showPermissionsModal.value = true;
  containersLoaded.value = false;
  if (editingUser.value.is_restricted_access) fetchRunningContainers();
};

const openResetPassword = (user) => {
  resetTargetUser.value = user;
  resetPassword.value = "";
  showResetModal.value = true;
};

const confirmResetPassword = async () => {
  if (!resetPassword.value) {
    showToast("Warning", "Please enter a password", "warning");
    return;
  }
  if (!resetTargetUser.value) return;

  try {
    const formData = new FormData();
    formData.append("password", resetPassword.value);

    const res = await apiFetch(
      `/api/admin/users/${resetTargetUser.value.id}/password`,
      {
        method: "PUT",
        headers: { Authorization: `Bearer ${props.token}` },
        body: formData,
      },
    );

    if (res.ok) {
      showToast("Success", "Password reset successfully", "success");
      closeAllModals();
    } else {
      const errorData = await res.json().catch(() => ({}));
      showToast(
        "Error",
        errorData.error || "Failed to reset password",
        "error",
      );
    }
  } catch (err) {
    console.error(err);
    showToast("Error", "Network error", "error");
  }
};

const updatePermissions = async () => {
  try {
    const formData = new FormData();
    formData.append("can_start", editingUser.value.can_start ? "true" : "false");
    formData.append("can_stop", editingUser.value.can_stop ? "true" : "false");
    formData.append("can_restart", editingUser.value.can_restart ? "true" : "false");
    formData.append("can_delete", editingUser.value.can_delete ? "true" : "false");
    formData.append("can_shell", editingUser.value.can_shell ? "true" : "false");
    formData.append(
      "is_restricted_access",
      editingUser.value.is_restricted_access ? "true" : "false",
    );
    formData.append("allowed_containers", editingUser.value.allowed_containers);

    const res = await apiFetch(
      `/api/admin/users/${editingUser.value.id}/permissions`,
      {
        method: "PUT",
        headers: { Authorization: `Bearer ${props.token}` },
        body: formData,
      },
    );
    if (res.ok) {
      showToast("Success", "Permissions updated", "success");
      closeAllModals();
      fetchStaff();
    }
  } catch (err) {
    console.error(err);
  }
};

const openDeleteConfirm = (user) => {
  userToDelete.value = user;
  showDeleteModal.value = true;
};

const confirmDelete = async () => {
  if (!userToDelete.value) return;
  try {
    const res = await apiFetch(`/api/admin/users/${userToDelete.value.id}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${props.token}` },
    });
    if (res.ok) {
      showToast("Deleted", "User account removed", "success");
      fetchStaff();
      closeAllModals();
    }
  } catch (err) {
    console.error(err);
  }
};

const openCreateModal = () => {
  showCreateModal.value = true;
  containersLoaded.value = false;
};
defineExpose({ openCreateModal });
onMounted(fetchStaff);
</script>

<style scoped>
/* Table Container & Centered Empty State */
.premium-table-container {
  display: flex;
  flex-direction: column;
}

.premium-table {
  width: 100%;
  flex: 1;
}

.empty-state-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 5rem 0;
  min-height: 350px;
}

.empty-state-content {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.empty-icon-box {
  width: 80px;
  height: 80px;
  background: var(--bg-input);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-mute);
  opacity: 0.6;
  margin-bottom: 1.5rem;
  border: 1px dashed var(--border);
}

.empty-icon-box svg {
  width: 40px;
  height: 40px;
}
.empty-title {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--text-main);
  margin-bottom: 0.5rem;
}
.empty-text {
  font-size: 0.9rem;
  color: var(--text-mute);
  max-width: 280px;
  line-height: 1.6;
}

/* Premium Toggle Switch */
.premium-toggle {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s;
}

.premium-toggle.disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.toggle-rail {
  width: 36px;
  height: 20px;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 20px;
  position: relative;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.premium-toggle.active .toggle-rail {
  background: var(--success);
  border-color: var(--success);
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.2);
}

.toggle-handle {
  width: 14px;
  height: 14px;
  background: #fff;
  border-radius: 50%;
  position: absolute;
  top: 2px;
  left: 2px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.premium-toggle.active .toggle-handle {
  transform: translateX(16px);
}

.status-label {
  font-size: 0.8rem;
  font-weight: 800;
  color: var(--text-mute);
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.premium-toggle.active .status-label {
  color: var(--success);
}

.hint-text {
  font-size: 0.75rem;
  color: var(--text-mute);
  font-weight: 500;
}

/* Modal General */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-card {
  width: 100%;
  max-width: 650px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 24px;
  overflow: hidden;
}

.mini-modal {
  max-width: 400px;
}

/* Header */
.modal-card-header {
  padding: 1.5rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
  border-bottom: 1px solid var(--border);
}

.close-btn {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-dim);
  cursor: pointer;
  transition: background 0.2s ease, color 0.2s ease, border-color 0.2s ease;
}

.close-btn svg {
  width: 18px;
  height: 18px;
}

.close-btn:hover {
  background: var(--bg-subtle);
  color: var(--text-main);
  border-color: var(--border-active);
}

.close-btn:active {
  transform: scale(0.96);
}

.header-content {
  display: flex;
  gap: 1.25rem;
  align-items: center;
}
.header-icon {
  width: 48px;
  height: 48px;
  background: var(--bg-subtle);
  color: var(--accent);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-title {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--text-main);
  margin: 0;
}
.modal-subtitle {
  font-size: 0.85rem;
  color: var(--text-mute);
  margin-top: 0.25rem;
}

/* Body */
.modal-card-body {
  padding: 2rem;
}
.form-grid.dual {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
  margin-bottom: 2rem;
}
.label-caps {
  display: block;
  font-family: var(--font-main);
  text-transform: uppercase;
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  color: var(--text-mute);
  margin-bottom: 0.5rem;
}

.label-caps-inline {
  display: inline-block;
  margin-bottom: 0;
}

.perm-section {
  margin-bottom: 1.35rem;
}

.perm-section:last-child {
  margin-bottom: 0;
}

.perm-section-compact {
  margin-top: 0.25rem;
}

.control-text {
  font-family: var(--font-main);
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-dim);
}

.premium-input {
  width: 100%;
  background: var(--bg-input);
  border: 2px solid var(--border);
  border-radius: 12px;
  padding: 0.75rem 1rem;
  font-family: var(--font-main);
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
  transition: all 0.2s;
}
.premium-input:focus {
  outline: none;
  border-color: var(--accent);
}

/* Visibility Selection */
.access-toggle-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  background: var(--bg-input);
  padding: 0.4rem;
  border-radius: 16px;
  gap: 0.4rem;
}

.access-choice {
  border: none;
  background: transparent;
  padding: 0.65rem 0.75rem;
  border-radius: 12px;
  font-family: var(--font-main);
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-dim);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  transition: 0.2s;
}

.access-choice.active {
  background: var(--bg-card);
  color: var(--accent);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
.access-choice .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.pattern-box {
  margin-top: 1.25rem;
  padding: 1.25rem;
  background: var(--bg-subtle);
  border-radius: 16px;
  border: 1px dashed var(--border);
}

.pattern-input-wrap {
  position: relative;
}

.pattern-suggestions {
  position: absolute;
  top: calc(100% + 0.35rem);
  left: 0;
  right: 0;
  z-index: 20;
  margin: 0;
  padding: 0.35rem;
  list-style: none;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  box-shadow: 0 12px 28px -8px var(--shadow);
  max-height: 220px;
  overflow-y: auto;
}

.pattern-suggestion-btn {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.55rem 0.65rem;
  border: none;
  border-radius: 8px;
  background: transparent;
  font-family: var(--font-main);
  color: var(--text-main);
  font-size: 0.78rem;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s ease;
}

.pattern-suggestion-btn:hover {
  background: var(--bg-subtle);
}

.suggestion-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--success);
  box-shadow: 0 0 6px rgba(var(--success-rgb), 0.45);
  flex-shrink: 0;
}

.pattern-fleet-head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 0.45rem;
}

.pattern-fleet-count {
  font-family: var(--font-main);
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--success);
  padding: 0.12rem 0.45rem;
  border-radius: 999px;
  background: rgba(var(--success-rgb), 0.1);
  border: 1px solid rgba(var(--success-rgb), 0.18);
}

.pattern-fleet-toggle {
  margin-left: auto;
  border: none;
  background: transparent;
  font-family: var(--font-main);
  color: var(--accent);
  font-size: 0.68rem;
  font-weight: 700;
  cursor: pointer;
  padding: 0.15rem 0.25rem;
}

.pattern-fleet-toggle:hover {
  text-decoration: underline;
}

.pattern-fleet-hint,
.pattern-fleet-empty,
.pattern-fleet-all-added {
  font-family: var(--font-main);
  font-size: 0.72rem;
  font-weight: 500;
  color: var(--text-mute);
  line-height: 1.4;
}

.pattern-fleet-hint {
  margin: 0 0 0.35rem;
}

.pattern-fleet-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.pattern-fleet-chips.scrollable {
  max-height: 5.5rem;
  overflow-y: auto;
  padding: 0.35rem;
  border-radius: 10px;
  background: var(--bg-input);
  border: 1px solid var(--border);
}

.pattern-fleet-all-added {
  font-style: italic;
}

.pattern-fleet-chip {
  border: 1px solid rgba(var(--success-rgb), 0.22);
  background: rgba(var(--success-rgb), 0.08);
  color: var(--success);
  font-family: var(--font-main);
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.22rem 0.55rem;
  border-radius: 999px;
  cursor: pointer;
  transition: background 0.15s ease, border-color 0.15s ease;
}

.pattern-fleet-chip:hover {
  background: rgba(var(--success-rgb), 0.14);
  border-color: rgba(var(--success-rgb), 0.35);
}

.pattern-fleet-empty {
  margin: 0;
}

/* Rights Grid */
.modern-rights-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.45rem;
}

.right-card {
  cursor: pointer;
}

.right-card input {
  display: none;
}

.right-card-content {
  padding: 0.45rem 0.6rem;
  background: var(--bg-input);
  border: 1px solid var(--border);
  border-radius: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.45rem;
  min-height: 0;
}

.right-label {
  font-family: var(--font-main);
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-dim);
}

.right-card.checked .right-label {
  color: var(--accent);
}

.right-card.checked .right-card-content {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.05);
}

.custom-check {
  width: 14px;
  height: 14px;
  border-radius: 4px;
  border: 1.5px solid var(--border);
  flex-shrink: 0;
}

.right-card.checked .custom-check {
  background: var(--accent);
  border-color: var(--accent);
}

/* Footer */
.modal-card-footer {
  padding: 1.5rem 2rem;
  border-top: 1px solid var(--border);
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}
.btn-primary {
  background: var(--accent);
  color: white;
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 12px;
  font-weight: 700;
  cursor: pointer;
}
.btn-secondary {
  background: var(--bg-subtle);
  color: var(--text-main);
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 12px;
  font-weight: 700;
  cursor: pointer;
}
.btn-danger {
  background: var(--stop);
  color: white;
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 12px;
  font-weight: 700;
  cursor: pointer;
}

/* Warning Icon */
.warning-icon-wrapper {
  width: 64px;
  height: 64px;
  background: rgba(239, 68, 68, 0.1);
  color: var(--stop);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.mini-avatar {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background: var(--accent-soft);
  color: var(--accent);
  border: 1px solid rgba(var(--accent-rgb), 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 800;
  flex-shrink: 0;
}

.user-name {
  font-weight: 800;
  color: var(--text-main);
}

.action-group {
  display: flex;
  gap: 0.5rem;
}

.perm-tags {
  display: flex;
  gap: 0.5rem;
}
.perm-tags span {
  padding: 0.2rem 0.6rem;
  border-radius: 6px;
  font-size: 0.7rem;
}

.perm-tags .tag-start {
  background: var(--success);
  color: white;
}
.perm-tags .tag-stop {
  background: var(--stop);
  color: white;
}
.perm-tags .tag-restart {
  background: var(--accent);
  color: white;
}
.perm-tags .tag-delete {
  background: var(--stop);
  color: white;
}

/* Transitions */
.modal-bounce-enter-active {
  animation: bounce 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}
@keyframes bounce {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

@media (max-width: 850px) {
  .premium-table thead {
    display: none;
  }
  .premium-table, .premium-table tbody, .premium-table tr, .premium-table td {
    display: block;
    width: 100%;
  }
  .premium-table-container {
    background: transparent;
    border: none;
    box-shadow: none;
    min-height: auto;
  }
  .premium-table tbody tr {
    margin-bottom: 1.25rem;
    padding: 1.25rem;
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 20px;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
  }
  .premium-table tbody tr td {
    padding: 0.6rem 0;
    border: none;
    text-align: left !important;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }
  .premium-table tbody tr td::before {
    content: attr(data-label);
    display: block;
    font-size: 0.65rem;
    font-weight: 800;
    color: var(--text-mute);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .action-group {
    justify-content: flex-start !important;
    margin-top: 0.5rem;
    gap: 0.75rem;
    flex-direction: row !important;
  }
  .perm-tags {
    flex-wrap: wrap;
  }
}

@media (max-width: 480px) {
  .premium-table tbody tr {
    padding: 1rem;
    margin-bottom: 1rem;
    border-radius: 16px;
  }
  .modal-card {
    padding: 1.25rem;
    border-radius: 20px;
  }
  .modal-title {
    font-size: 1.25rem;
  }
  .form-grid.dual {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  .modal-card-header, .modal-card-body, .modal-card-footer {
    padding: 1rem 1.25rem;
  }
  .modern-rights-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
.pattern-examples {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.example-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.tag {
  font-family: var(--font-mono);
  font-size: 0.68rem;
  font-weight: 700;
  background: var(--accent-soft);
  color: var(--accent);
  padding: 0.15rem 0.4rem;
  border-radius: 6px;
  border: 1px solid rgba(var(--accent-rgb), 0.2);
  min-width: 80px;
  text-align: center;
}

.desc {
  font-family: var(--font-main);
  font-size: 0.72rem;
  color: var(--text-mute);
  font-weight: 500;
}
</style>

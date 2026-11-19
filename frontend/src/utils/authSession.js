const SESSION_EXEMPT_PATHS = ["/api/token", "/api/user/change-password"];

export function getRequestUrl(input) {
  if (typeof input === "string") return input;
  if (input?.url) return String(input.url);
  if (input?.href) return String(input.href);
  return String(input);
}

export function isSessionExemptUrl(url) {
  return SESSION_EXEMPT_PATHS.some((path) => url.includes(path));
}

export async function shouldForceLogout(response, requestInput) {
  if (!response) return false;

  const url = getRequestUrl(requestInput);
  if (isSessionExemptUrl(url)) return false;

  if (response.status === 401) {
    try {
      const contentType = response.headers.get("content-type") || "";
      if (contentType.includes("application/json")) {
        const payload = await response.clone().json();
        // FORCE_PASSWORD_CHANGE should show modal, not wipe token and force logout
        if (payload?.code === "FORCE_PASSWORD_CHANGE") {
          return false;
        }
      }
    } catch {
      // Ignore json parsing issues
    }
    return true;
  }

  if (response.status === 403) {
    try {
      const contentType = response.headers.get("content-type") || "";
      if (!contentType.includes("application/json")) return false;
      const payload = await response.clone().json();
      return payload?.code === "ACCOUNT_DEACTIVATED";
    } catch {
      return false;
    }
  }

  return false;
}

export async function readApiError(response, fallbackMessage) {
  const data = await response.json().catch(() => ({}));
  const err = new Error(data.error || fallbackMessage);
  err.code = data.code;
  err.status = response.status;
  return err;
}

export function apiErrorMessage(error, fallbackMessage = "Request failed") {
  if (error?.code === "FORCE_PASSWORD_CHANGE") {
    return "Change your password before using this feature.";
  }
  return error?.message || fallbackMessage;
}

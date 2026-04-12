import type { Router } from "vue-router";

export const DEFAULT_LOGIN_REDIRECT = "/conversations";

export function resolveLoginRedirect(router: Router, redirect: unknown) {
  if (typeof redirect !== "string" || !redirect.startsWith("/") || redirect.startsWith("//")) {
    return DEFAULT_LOGIN_REDIRECT;
  }

  const target = router.resolve(redirect);

  if (!target.matched.length || target.path === "/login") {
    return DEFAULT_LOGIN_REDIRECT;
  }

  return target.fullPath;
}

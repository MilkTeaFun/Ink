import { createPinia, setActivePinia } from "pinia";
import { afterEach, vi } from "vitest";
import { createMemoryHistory } from "vue-router";

import { createAppRouter, navigationItems, routes } from "@/router";
import { useWorkspaceStore } from "@/stores/workspace";

afterEach(() => {
  vi.restoreAllMocks();
});

function createAuthenticatedRouter() {
  const pinia = createPinia();
  setActivePinia(pinia);
  const store = useWorkspaceStore();
  store.authUser = {
    id: "user-1",
    email: "name@example.com",
    name: "Ink User",
  };
  store.authSession = {
    accessToken: "access-token",
    refreshToken: "refresh-token",
    accessTokenExpiresAt: new Date(Date.now() + 60_000).toISOString(),
  };

  return createAppRouter(createMemoryHistory(), pinia);
}

describe("router configuration", () => {
  it("keeps navigation items in sync with workspace routes", () => {
    const workspaceRoute = routes.find((route) => route.path === "/");
    const shellChildren = workspaceRoute?.children ?? [];

    expect(navigationItems).toHaveLength(shellChildren.length);
    expect(navigationItems.map((item) => item.path)).toEqual([
      "/status",
      "/conversations",
      "/prints",
      "/settings",
    ]);
  });

  it("redirects anonymous visitors from protected routes to login", async () => {
    const pinia = createPinia();
    const router = createAppRouter(createMemoryHistory(), pinia);

    router.push("/");
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe("/login?redirect=/status");
  });

  it("allows authenticated visitors to reach protected routes", async () => {
    const router = createAuthenticatedRouter();

    router.push("/");
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe("/status");
  });

  it("redirects the retired connections route to /prints for authenticated visitors", async () => {
    const router = createAuthenticatedRouter();

    router.push("/connections");
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe("/prints");
  });

  it("restores a persisted session before redirecting protected routes", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          user: {
            id: "user-1",
            email: "name@example.com",
            name: "Ink User",
          },
        }),
        { status: 200 },
      ),
    );

    const pinia = createPinia();
    setActivePinia(pinia);
    const store = useWorkspaceStore();
    store.authSession = {
      accessToken: "access-token",
      refreshToken: "refresh-token",
      accessTokenExpiresAt: new Date(Date.now() + 60_000).toISOString(),
    };

    const router = createAppRouter(createMemoryHistory(), pinia);
    router.push("/status");
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe("/status");
    expect(store.authUser?.email).toBe("name@example.com");
  });

  it("updates the document title from route metadata", async () => {
    const router = createAuthenticatedRouter();

    router.push("/settings");
    await router.isReady();

    expect(document.title).toBe("Ink · 偏好设置");
  });
});

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
    role: "member",
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

  it("redirects anonymous visitors from the root route to the public status page", async () => {
    const pinia = createPinia();
    const router = createAppRouter(createMemoryHistory(), pinia);

    router.push("/");
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe("/status");
  });

  it.each(["/status", "/conversations", "/prints"])(
    "allows anonymous visitors to reach %s",
    async (path) => {
      const pinia = createPinia();
      const router = createAppRouter(createMemoryHistory(), pinia);

      router.push(path);
      await router.isReady();

      expect(router.currentRoute.value.fullPath).toBe(path);
    },
  );

  it("redirects anonymous visitors from settings to login", async () => {
    const pinia = createPinia();
    const router = createAppRouter(createMemoryHistory(), pinia);

    router.push("/settings");
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe("/login?redirect=/settings");
  });

  it("redirects the retired connections route to /prints for authenticated visitors", async () => {
    const router = createAuthenticatedRouter();

    router.push("/connections");
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe("/prints");
  });

  it("restores a persisted session before entering settings", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            user: {
              id: "user-1",
              email: "name@example.com",
              name: "Ink User",
              role: "member",
            },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            devices: [],
            conversations: [],
            activeConversationId: "",
            printJobs: [],
            schedules: [],
            sources: [],
            preferences: {
              loginProtectionEnabled: false,
              sendConfirmationEnabled: true,
              theme: "light",
              defaultDeviceId: "",
            },
            serviceBinding: {
              providerName: null,
              modelName: "Ink AI",
              bound: false,
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
    router.push("/settings");
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe("/settings");
    expect(store.authUser?.email).toBe("name@example.com");
  });

  it.each([
    ["/login", "/status"],
    ["/login?redirect=/settings", "/settings"],
    ["/login?redirect=/missing", "/status"],
    ["/login?redirect=/login", "/status"],
  ])("redirects authenticated visitors from %s to %s", async (source, destination) => {
    const router = createAuthenticatedRouter();

    router.push(source);
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe(destination);
  });

  it("updates the document title from route metadata", async () => {
    const router = createAuthenticatedRouter();

    router.push("/settings");
    await router.isReady();

    expect(document.title).toBe("Ink · 偏好设置");
  });
});

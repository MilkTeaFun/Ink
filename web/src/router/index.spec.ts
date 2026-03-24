import { createMemoryHistory } from "vue-router";

import { createAppRouter, navigationItems, routes } from "@/router";

describe("router configuration", () => {
  it("keeps navigation items in sync with workspace routes", () => {
    const workspaceRoute = routes.find((route) => route.path === "/");
    const shellChildren = workspaceRoute?.children ?? [];

    expect(navigationItems).toHaveLength(shellChildren.length);
    expect(navigationItems.map((item) => item.path)).toEqual([
      "/status",
      "/conversations",
      "/connections",
      "/settings",
    ]);

    shellChildren.forEach((route, index) => {
      expect(route.name).toBeTruthy();
      expect(route.meta?.label).toBeTruthy();
      expect(route.meta?.title).toBeTruthy();
      expect(route.meta?.description).toBeTruthy();
      expect(route.meta?.navHint).toBeTruthy();
      expect(navigationItems[index]).toMatchObject({
        name: route.name,
        label: route.meta?.label,
        navHint: route.meta?.navHint,
      });
    });
  });

  it("redirects the root route to /status", async () => {
    const router = createAppRouter(createMemoryHistory());

    router.push("/");
    await router.isReady();

    expect(router.currentRoute.value.fullPath).toBe("/status");
  });

  it("keeps the login route outside of workspace navigation", () => {
    expect(navigationItems.some((item) => item.path === "/login")).toBe(false);

    const loginRoute = routes.find((route) => route.path === "/login");

    expect(loginRoute?.name).toBe("login");
    expect(loginRoute?.meta?.title).toBe("欢迎使用 Ink");
  });
});

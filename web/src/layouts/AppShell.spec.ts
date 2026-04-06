import { flushPromises, mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";

import AppShell from "@/layouts/AppShell.vue";
import { createTestRouter, navigationItems } from "@/router";
import { useWorkspaceStore } from "@/stores/workspace";

async function mountShellAt(path: string) {
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

  const router = createTestRouter(pinia);
  router.push(path);
  await router.isReady();

  const wrapper = mount(AppShell, {
    global: {
      plugins: [pinia, router],
    },
  });

  return { wrapper, router, store };
}

describe("AppShell", () => {
  it("renders desktop and mobile navigation from router metadata", async () => {
    const { wrapper } = await mountShellAt("/status");

    const desktopNavLinks = wrapper.findAll("header nav a");
    const mobileNavLinks = wrapper.findAll("nav.fixed a");

    expect(desktopNavLinks).toHaveLength(navigationItems.length);
    expect(mobileNavLinks).toHaveLength(navigationItems.length);
    expect(desktopNavLinks.map((link) => link.text().replace(/\d+/g, ""))).toEqual(
      navigationItems.map((item) => item.label),
    );
    expect(mobileNavLinks.map((link) => link.text().replace(/\s*·\s*\d+/g, ""))).toEqual(
      navigationItems.map((item) => item.label),
    );
  });

  it("shows the pending print badge and authenticated account controls", async () => {
    const { wrapper } = await mountShellAt("/status");

    expect(wrapper.text()).toContain("打印1");
    expect(wrapper.text()).toContain("name@example.com");
    expect(wrapper.text()).toContain("退出");
  });

  it("logs out and returns to login when the header logout action is used", async () => {
    const { wrapper, router, store } = await mountShellAt("/prints");
    const logoutButton = wrapper.findAll("button").find((button) => button.text() === "退出");

    expect(logoutButton?.exists()).toBe(true);

    await logoutButton?.trigger("click");
    await flushPromises();

    expect(store.isAuthenticated).toBe(false);
    expect(router.currentRoute.value.fullPath).toBe("/login");
  });
});

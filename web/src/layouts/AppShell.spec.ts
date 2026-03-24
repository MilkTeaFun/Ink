import { mount } from "@vue/test-utils";

import AppShell from "@/layouts/AppShell.vue";
import { createTestRouter, navigationItems } from "@/router";

async function mountShellAt(path: string) {
  const router = createTestRouter();

  router.push(path);
  await router.isReady();

  const wrapper = mount(AppShell, {
    global: {
      plugins: [router],
    },
  });

  return { wrapper, router };
}

describe("AppShell", () => {
  it("renders matching desktop and mobile navigation from router metadata", async () => {
    const { wrapper } = await mountShellAt("/status");

    const desktopNavLinks = wrapper.findAll("header nav a");
    const mobileNavLinks = wrapper.findAll("nav.fixed a");

    expect(desktopNavLinks).toHaveLength(navigationItems.length);
    expect(mobileNavLinks).toHaveLength(navigationItems.length);
    expect(desktopNavLinks.map((link) => link.text())).toEqual(
      navigationItems.map((item) => item.label),
    );
    expect(mobileNavLinks.map((link) => link.text())).toEqual(
      navigationItems.map((item) => item.label),
    );
  });

  it.each([
    ["/status", "状态", "连接"],
    ["/settings", "设置", "状态"],
  ])(
    "applies active state classes for %s in both desktop and mobile navigation",
    async (path, activeLabel, inactiveLabel) => {
      const { wrapper } = await mountShellAt(path);

      const desktopActiveLink = wrapper
        .findAll("header nav a")
        .find((link) => link.text() === activeLabel);
      const desktopInactiveLink = wrapper
        .findAll("header nav a")
        .find((link) => link.text() === inactiveLabel);
      const mobileActiveLink = wrapper
        .findAll("nav.fixed a")
        .find((link) => link.text() === activeLabel);
      const mobileInactiveLink = wrapper
        .findAll("nav.fixed a")
        .find((link) => link.text() === inactiveLabel);

      expect(desktopActiveLink?.classes()).toContain("bg-stone-100");
      expect(desktopActiveLink?.classes()).toContain("text-stone-900");
      expect(desktopInactiveLink?.classes()).not.toContain("bg-stone-100");
      expect(desktopInactiveLink?.classes()).toContain("text-stone-600");

      expect(mobileActiveLink?.classes()).toContain("text-stone-900");
      expect(mobileInactiveLink?.classes()).toContain("text-stone-500");
    },
  );

  it("keeps login entry points available in both header variants", async () => {
    const { wrapper } = await mountShellAt("/connections");

    expect(wrapper.find("a[href='/login']").exists()).toBe(true);
    expect(wrapper.findAll("a[href='/login']").map((link) => link.text())).toEqual([
      "账号",
      "登录",
    ]);
  });
});

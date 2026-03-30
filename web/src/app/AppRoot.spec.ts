import { mount, type VueWrapper } from "@vue/test-utils";
import { createPinia } from "pinia";

import AppRoot from "@/app/AppRoot.vue";
import { createTestRouter, navigationItems } from "@/router";

async function mountAt(path: string) {
  const router = createTestRouter();
  const pinia = createPinia();

  router.push(path);
  await router.isReady();

  const wrapper = mount(AppRoot, {
    global: {
      plugins: [pinia, router],
    },
  });

  return { wrapper, router };
}

function expectActiveLabelInNav(wrapper: VueWrapper, selector: string, label: string) {
  const links = wrapper.findAll(selector);
  const activeLink = links.find((link) => link.text() === label);

  expect(activeLink?.classes()).toContain("text-stone-900");

  links
    .filter((link) => link.text() !== label)
    .forEach((link) => expect(link.classes()).not.toContain("text-stone-900"));
}

describe("AppRoot", () => {
  it.each([
    ["/status", "状态"],
    ["/conversations", "对话"],
    ["/prints", "打印"],
    ["/settings", "设置"],
  ])("renders the workspace shell for %s", async (path, heading) => {
    const { wrapper } = await mountAt(path);

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
    expectActiveLabelInNav(wrapper, "header nav a", heading);
    expectActiveLabelInNav(wrapper, "nav.fixed a", heading);
    expect(wrapper.text()).toContain(heading);
  });

  it("renders the login view outside the workspace shell", async () => {
    const { wrapper } = await mountAt("/login");

    expect(wrapper.text()).toContain("登录账号");
    expect(wrapper.find("header nav").exists()).toBe(false);
    expect(wrapper.find("nav.fixed").exists()).toBe(false);
    expect(wrapper.text()).not.toContain("状态");
    expect(wrapper.text()).not.toContain("对话");
    expect(wrapper.text()).not.toContain("设置");
  });
});

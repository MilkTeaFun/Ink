import { mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { nextTick } from "vue";

import AppRoot from "@/app/AppRoot.vue";
import { createTestRouter } from "@/router";
import { useWorkspaceStore } from "@/stores/workspace";

async function mountAt(path: string, authenticated = true) {
  const pinia = createPinia();
  setActivePinia(pinia);
  const store = useWorkspaceStore();

  if (authenticated) {
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
  }

  const router = createTestRouter(pinia);

  router.push(path);
  await router.isReady();

  const wrapper = mount(AppRoot, {
    global: {
      plugins: [pinia, router],
    },
  });

  return { wrapper, router, store };
}

describe("AppRoot", () => {
  it.each([
    ["/status", "状态"],
    ["/conversations", "对话"],
    ["/prints", "打印"],
    ["/settings", "设置"],
  ])("renders the workspace shell for %s when authenticated", async (path, heading) => {
    const { wrapper } = await mountAt(path, true);

    expect(wrapper.find("header").exists()).toBe(true);
    expect(wrapper.find("nav.fixed").exists()).toBe(true);
    expect(wrapper.text()).toContain(heading);
  });

  it("renders the workspace shell for anonymous visitors on the public status page", async () => {
    const { wrapper, router } = await mountAt("/status", false);

    expect(router.currentRoute.value.fullPath).toBe("/status");
    expect(wrapper.find("header nav").exists()).toBe(true);
    expect(wrapper.find("nav.fixed").exists()).toBe(true);
    expect(wrapper.text()).not.toContain("登录账号");
    expect(wrapper.text()).not.toContain("退出");
  });

  it("renders the login view outside the workspace shell for anonymous settings access", async () => {
    const { wrapper, router } = await mountAt("/settings", false);

    expect(router.currentRoute.value.fullPath).toBe("/login?redirect=/settings");
    expect(wrapper.text()).toContain("登录账号");
    expect(wrapper.find("header nav").exists()).toBe(false);
    expect(wrapper.find("nav.fixed").exists()).toBe(false);
  });

  it("syncs the selected theme to the document and theme-color meta", async () => {
    let themeMeta = document.querySelector('meta[name="theme-color"]');
    if (!themeMeta) {
      themeMeta = document.createElement("meta");
      themeMeta.setAttribute("name", "theme-color");
      document.head.appendChild(themeMeta);
    }
    themeMeta.setAttribute("content", "#000000");

    const { store } = await mountAt("/status");
    store.selectedTheme = "soft";
    await nextTick();

    expect(document.documentElement.dataset.theme).toBe("soft");
    expect(themeMeta.getAttribute("content")).toBe("#f7f1ea");
  });
});

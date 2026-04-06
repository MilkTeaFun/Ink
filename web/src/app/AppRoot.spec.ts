import { mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";

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

  return { wrapper, router };
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

  it("renders the login view outside the workspace shell for anonymous visitors", async () => {
    const { wrapper, router } = await mountAt("/status", false);

    expect(router.currentRoute.value.fullPath).toBe("/login?redirect=/status");
    expect(wrapper.text()).toContain("登录账号");
    expect(wrapper.find("header nav").exists()).toBe(false);
    expect(wrapper.find("nav.fixed").exists()).toBe(false);
  });
});

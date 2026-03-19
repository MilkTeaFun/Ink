import { mount } from "@vue/test-utils";
import { createPinia } from "pinia";

import AppRoot from "@/app/AppRoot.vue";
import router from "@/router";

describe("AppRoot", () => {
  it("renders the shell navigation for workspace routes", async () => {
    const pinia = createPinia();

    router.push("/status");
    await router.isReady();

    const wrapper = mount(AppRoot, {
      global: {
        plugins: [pinia, router],
      },
    });

    expect(wrapper.text()).toContain("状态");
    expect(wrapper.text()).toContain("对话");
    expect(wrapper.text()).toContain("连接");
    expect(wrapper.text()).toContain("设置");
    expect(wrapper.text()).toContain("已绑定设备");
  });
});

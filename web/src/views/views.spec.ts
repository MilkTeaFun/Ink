import { mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";

import { createTestRouter } from "@/router";
import { useWorkspaceStore } from "@/stores/workspace";
import type { PrintJob } from "@/types/workspace";
import ConversationsView from "@/views/ConversationsView.vue";
import LoginView from "@/views/LoginView.vue";
import PrintsView from "@/views/PrintsView.vue";
import SettingsView from "@/views/SettingsView.vue";
import StatusView from "@/views/StatusView.vue";

async function createWorkspaceContext(path = "/status", authenticated = true) {
  const pinia = createPinia();
  setActivePinia(pinia);
  const store = useWorkspaceStore();

  if (authenticated) {
    store.authUser = {
      id: "user-1",
      email: "name@example.com",
      name: "Ink User",
    };
  }

  const router = createTestRouter(pinia);
  router.push(path);
  await router.isReady();

  return { pinia, router, store };
}

describe("workspace views", () => {
  it("renders the status overview from shared store state", async () => {
    const { pinia, router } = await createWorkspaceContext("/status");
    const wrapper = mount(StatusView, {
      global: {
        plugins: [pinia, router],
      },
    });

    expect(wrapper.text()).toContain("已绑定设备");
    expect(wrapper.text()).toContain("自动打印");
    expect(wrapper.findAll(".ui-list-card article")).not.toHaveLength(0);
  });

  it("allows sending a message and selecting a reply for printing", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/conversations");
    const wrapper = mount(ConversationsView, {
      global: {
        plugins: [pinia, router],
      },
    });

    await wrapper.find("textarea").setValue("请帮我整理成一句适合贴在电脑旁边的提醒");
    await wrapper.find("button.ui-btn-primary").trigger("click");
    await new Promise((resolve) => window.setTimeout(resolve, 120));

    expect(store.activeConversation?.messages.at(-1)?.role).toBe("assistant");
    expect(wrapper.text()).toContain("打印选中回答");

    await wrapper
      .findAll("button.ui-btn-secondary")
      .find((button) => button.text() === "打印选中回答")
      ?.trigger("click");

    expect(store.pendingPrintJobs.at(0)?.source).toBe("对话选中回答");
  });

  it("validates and submits the mock login form", async () => {
    const { pinia, router } = await createWorkspaceContext("/login", false);
    const wrapper = mount(LoginView, {
      global: {
        plugins: [pinia, router],
      },
    });

    await wrapper.find("input[type='email']").setValue("bad-email");
    await wrapper.find("input[type='password']").setValue("demo");
    await wrapper.find("form").trigger("submit");

    expect(wrapper.text()).toContain("请输入有效邮箱和密码。");

    await wrapper.find("input[type='email']").setValue("name@example.com");
    await wrapper.find("form").trigger("submit");
    await new Promise((resolve) => window.setTimeout(resolve, 120));

    expect(router.currentRoute.value.fullPath).toBe("/status");
  });

  it("confirms pending prints and reflects shared defaults", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/prints");
    const wrapper = mount(PrintsView, {
      global: {
        plugins: [pinia, router],
      },
    });

    const pendingButton = wrapper.findAll("button").find((button) => button.text() === "确认打印");

    await pendingButton?.trigger("click");

    expect(store.pendingPrintJobs.some((job: PrintJob) => job.status === "queued")).toBe(true);
    expect(wrapper.text()).toContain("默认打印设置");
    expect(wrapper.text()).toContain("书桌咕咕机");
  });

  it("updates shared settings state through the settings panel", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/settings");
    const wrapper = mount(SettingsView, {
      global: {
        plugins: [pinia, router],
      },
    });

    const selects = wrapper.findAll("select");
    await selects[0].setValue("device-bedroom");
    await selects[1].setValue("gentle");
    await selects[2].setValue("warm-encouraging");

    expect(store.activeDeviceLabel).toBe("卧室咕咕机");
    expect(store.activeNoteStyle).toBe("gentle");
    expect(store.activeAnswerStyle).toBe("warm-encouraging");
    expect(wrapper.text()).toContain("AI 服务");
  });
});

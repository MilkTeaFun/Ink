import { mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { beforeEach, vi } from "vitest";

import { createTestRouter } from "@/router";
import { useWorkspaceStore } from "@/stores/workspace";
import type { PrintJob } from "@/types/workspace";
import ConversationsView from "@/views/ConversationsView.vue";
import LoginView from "@/views/LoginView.vue";
import PrintsView from "@/views/PrintsView.vue";
import SettingsView from "@/views/SettingsView.vue";
import StatusView from "@/views/StatusView.vue";

vi.mock("@/services/auth", () => ({
  changePasswordWithApi: vi.fn(async () => undefined),
  fetchCurrentUser: vi.fn(),
  loginWithApi: vi.fn(async ({ email }: { email: string }) => ({
    user: {
      id: "user-1",
      email,
      name: "Ink User",
    },
    session: {
      accessToken: "access-token",
      refreshToken: "refresh-token",
      accessTokenExpiresAt: new Date(Date.now() + 900_000).toISOString(),
    },
  })),
  logoutWithApi: vi.fn(async () => undefined),
  refreshAuthSession: vi.fn(),
  AuthApiError: class AuthApiError extends Error {},
}));

beforeEach(() => {
  vi.clearAllMocks();
  vi.stubGlobal(
    "confirm",
    vi.fn(() => true),
  );
});

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
    store.authSession = {
      accessToken: "access-token",
      refreshToken: "refresh-token",
      accessTokenExpiresAt: new Date(Date.now() + 60_000).toISOString(),
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

  it("manages devices from the status view", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/status");
    const wrapper = mount(StatusView, {
      global: {
        plugins: [pinia, router],
      },
    });

    const previousLength = store.devices.length;
    await wrapper
      .findAll("button")
      .find((button) => button.text() === "添加设备")
      ?.trigger("click");
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    await wrapper.find("input[placeholder='例如：客厅咕咕机']").setValue("客厅咕咕机");
    await wrapper.find("input[placeholder='例如：窗边打印机']").setValue("窗边打印机");
    await wrapper.find("input[type='checkbox']").setValue(true);
    await wrapper.find("form").trigger("submit");
    expect(store.devices).toHaveLength(previousLength + 1);
    expect(store.defaultDeviceId).toBe(store.devices.at(-1)?.id);

    const bedroomArticle = wrapper
      .findAll("article")
      .find((article) => article.text().includes("卧室咕咕机"));
    await bedroomArticle
      ?.findAll("button")
      .find((button) => button.text() === "设为默认")
      ?.trigger("click");
    expect(store.defaultDeviceId).toBe("device-bedroom");

    await wrapper
      .findAll("button")
      .find((button) => button.text() === "解绑")
      ?.trigger("click");
    expect(store.devices.some((device) => device.id === "device-desk")).toBe(false);
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
    expect(wrapper.text()).toContain("打印选中问答");
    expect(store.selectedConversationMessageIds).toEqual([]);

    const assistantSelectionButton = wrapper
      .findAll("button")
      .filter((button) => button.attributes("aria-label") === "选择这条消息")
      .at(-1);
    await assistantSelectionButton?.trigger("click");
    expect(store.selectedConversationMessageIds).toContain(
      store.activeConversation?.messages.at(-1)?.id,
    );

    await wrapper
      .findAll("button.ui-btn-secondary")
      .find((button) => button.text() === "打印选中问答")
      ?.trigger("click");

    expect(store.pendingPrintJobs.at(0)?.source).toBe("对话选中问答");
  });

  it("allows selecting and printing a user message", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/conversations");
    const wrapper = mount(ConversationsView, {
      global: {
        plugins: [pinia, router],
      },
    });

    const userMessageButton = wrapper
      .findAll("button")
      .find((button) => button.text().includes("帮我整理一张温柔一点的今日提醒"));

    await userMessageButton?.trigger("click");
    expect(store.selectedConversationMessages.at(0)?.role).toBe("user");

    await wrapper
      .findAll("button.ui-btn-secondary")
      .find((button) => button.text() === "打印选中问答")
      ?.trigger("click");

    expect(store.pendingPrintJobs.at(0)?.content).toContain("帮我整理一张温柔一点的今日提醒");
  });

  it("deletes the current conversation from the conversations view", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/conversations");
    const wrapper = mount(ConversationsView, {
      global: {
        plugins: [pinia, router],
      },
    });

    const previousLength = store.conversations.length;
    await wrapper
      .findAll("button")
      .find((button) => button.text() === "删除对话")
      ?.trigger("click");

    expect(store.conversations).toHaveLength(previousLength - 1);
    expect(store.activeConversation).not.toBeNull();
  });

  it("validates and submits the login form", async () => {
    const { pinia, router } = await createWorkspaceContext("/login", false);
    const wrapper = mount(LoginView, {
      global: {
        plugins: [pinia, router],
      },
    });

    await wrapper.find("input[type='text']").setValue("");
    await wrapper.find("input[type='password']").setValue("demo");
    await wrapper.find("form").trigger("submit");

    expect(wrapper.text()).toContain("请输入账号和密码。");

    await wrapper.find("input[type='text']").setValue("admin");
    await wrapper.find("input[type='password']").setValue("demo-password");
    await wrapper.find("form").trigger("submit");
    await new Promise((resolve) => window.setTimeout(resolve, 120));

    expect(router.currentRoute.value.fullPath).toBe("/status");
  });

  it("shows the password-updated notice and supports password visibility toggle on login", async () => {
    const { pinia, router } = await createWorkspaceContext("/login?notice=password-updated", false);
    const wrapper = mount(LoginView, {
      global: {
        plugins: [pinia, router],
      },
    });

    expect(wrapper.text()).toContain("密码已更新，请使用新密码重新登录。");
    const passwordInput = wrapper.find("input[type='password']");
    expect(passwordInput.exists()).toBe(true);

    const toggleButton = wrapper.findAll("button").find((button) => button.text() === "显示");
    await toggleButton?.trigger("click");

    expect(wrapper.find("input[type='text']").exists()).toBe(true);
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

  it("cancels a pending print job from the prints view", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/prints");
    const wrapper = mount(PrintsView, {
      global: {
        plugins: [pinia, router],
      },
    });

    const pendingArticle = wrapper
      .findAll("article")
      .find((article) => article.text().includes("待确认"));
    await pendingArticle
      ?.findAll("button")
      .find((button) => button.text() === "取消打印")
      ?.trigger("click");

    expect(store.pendingPrintJobs.some((job: PrintJob) => job.status === "pending")).toBe(false);
    expect(store.printJobs.some((job: PrintJob) => job.status === "cancelled")).toBe(true);
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

    expect(store.activeDeviceLabel).toBe("卧室咕咕机");
    expect(selects).toHaveLength(1);
    expect(wrapper.text()).toContain("AI 服务");
  });

  it("submits the password change form and returns to login", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/settings");
    const wrapper = mount(SettingsView, {
      global: {
        plugins: [pinia, router],
      },
    });

    const passwordInputs = wrapper.findAll("input[type='password']");
    await passwordInputs[0].setValue("demo-password");
    await passwordInputs[1].setValue("next-password");
    await passwordInputs[2].setValue("next-password");
    const showButtons = wrapper.findAll("button").filter((button) => button.text() === "显示");
    await showButtons[0]?.trigger("click");
    expect(passwordInputs[0].attributes("type")).toBe("text");

    await wrapper.find("form").trigger("submit");
    await new Promise((resolve) => window.setTimeout(resolve, 10));

    expect(store.isAuthenticated).toBe(false);
    expect(router.currentRoute.value.fullPath).toBe("/login?notice=password-updated");
  });
});

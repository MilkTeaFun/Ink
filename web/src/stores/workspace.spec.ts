import { createPinia, setActivePinia } from "pinia";
import { vi } from "vitest";

import { useWorkspaceStore } from "@/stores/workspace";
import type { PrintJob } from "@/types/workspace";

vi.mock("@/services/auth", () => ({
  changePasswordWithApi: vi.fn(async () => undefined),
  fetchCurrentUser: vi.fn(),
  loginWithApi: vi.fn(),
  logoutWithApi: vi.fn(async () => undefined),
  refreshAuthSession: vi.fn(),
  AuthApiError: class AuthApiError extends Error {},
}));

describe("workspace store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.localStorage.clear();
    window.sessionStorage.clear();
  });

  it("exposes stable defaults and derived summaries", () => {
    const store = useWorkspaceStore();

    expect(store.activeDeviceLabel).toBe("书桌咕咕机");
    expect(store.activeModelLabel).toBe("Ink AI");
    expect(store.todayPrintCount).toBe(2);
    expect(store.welcomeLabel).toBe("整理内容，准备打印");
    expect(store.isConfigured).toBe(true);
    expect(store.loginProtectionEnabled).toBe(false);
    expect(store.pendingConfirmationCount).toBe(1);
    expect(store.enabledSchedulesCount).toBe(2);
  });

  it("updates configuration from settings actions instead of local hard-coded values", () => {
    const store = useWorkspaceStore();

    store.setDefaultDevice("device-bedroom");

    expect(store.activeDeviceLabel).toBe("卧室咕咕机");
    expect(store.isConfigured).toBe(true);
    expect(store.activeModelLabel).toBe("Ink AI");
    expect(store.welcomeLabel).toBe("整理内容，准备打印");
  });

  it("adds devices and prevents removing the default device", () => {
    const store = useWorkspaceStore();
    const previousLength = store.devices.length;

    const newDevice = store.addDevice({
      name: "客厅咕咕机",
      note: "窗边打印机",
      setAsDefault: true,
    });

    expect(store.devices).toHaveLength(previousLength + 1);
    expect(store.devices.at(-1)?.id).toBe(newDevice.id);
    expect(store.defaultDeviceId).toBe(newDevice.id);
    expect(store.devices.at(-1)?.name).toBe("客厅咕咕机");
    expect(store.devices.at(-1)?.note).toBe("窗边打印机");
    expect(store.removeDevice(store.defaultDeviceId)).toBe(true);
    expect(store.devices.some((device) => device.id === newDevice.id)).toBe(false);
  });

  it("removes non-default devices and reassigns related items to the default device", () => {
    const store = useWorkspaceStore();

    expect(store.removeDevice("device-bedroom")).toBe(true);
    expect(store.devices.some((device) => device.id === "device-bedroom")).toBe(false);
    expect(store.printJobs.some((job) => job.deviceId === "device-bedroom")).toBe(false);
    expect(store.schedules.some((schedule) => schedule.deviceId === "device-bedroom")).toBe(false);
    expect(store.printJobs.some((job) => job.deviceId === store.defaultDeviceId)).toBe(true);
    expect(store.schedules.some((schedule) => schedule.deviceId === store.defaultDeviceId)).toBe(
      true,
    );
  });

  it("allows removing all devices and clears the default device", () => {
    const store = useWorkspaceStore();

    expect(store.removeDevice("device-bedroom")).toBe(true);
    expect(store.removeDevice("device-desk")).toBe(true);
    expect(store.devices).toHaveLength(0);
    expect(store.defaultDeviceId).toBe("");
    expect(
      store.printJobs.every(
        (job) => job.deviceId === "" || job.status === "cancelled" || job.status === "completed",
      ),
    ).toBe(true);
  });

  it("can generate a reply and create a pending print from the active conversation", async () => {
    const store = useWorkspaceStore();
    const previousCount = store.pendingPrintJobs.length;

    store.updateCurrentDraft("请帮我整理一句适合明早看的提醒");
    await store.sendCurrentDraft();
    store.toggleConversationMessageSelection(store.activeConversation!.messages.at(-1)!.id);
    const printJob = await store.createPrintFromSelectedMessages();

    expect(store.activeConversation?.messages.at(-1)?.role).toBe("assistant");
    expect(store.selectedConversationMessages.at(-1)?.text).toContain(
      "请帮我整理一句适合明早看的提醒",
    );
    expect(store.pendingPrintJobs).toHaveLength(previousCount + 1);
    expect(printJob?.status).toBe("pending");
    expect(printJob?.source).toBe("对话选中问答");
  });

  it("supports multi-select printing in conversation order", async () => {
    const store = useWorkspaceStore();
    const first = store.activeConversation!.messages[0];
    const second = store.activeConversation!.messages[1];

    store.toggleConversationMessageSelection(second.id);
    store.toggleConversationMessageSelection(first.id);

    const printJob = await store.createPrintFromSelectedMessages();

    expect(store.selectedConversationMessages.map((message) => message.id)).toEqual([
      first.id,
      second.id,
    ]);
    expect(printJob?.content).toContain(`我：${first.text}`);
    expect(printJob?.content).toContain(`Ink：${second.text}`);
  });

  it("cancels pending and queued print jobs without removing history", async () => {
    const store = useWorkspaceStore();
    const pendingJob = store.printJobs.find((job) => job.status === "pending");

    expect(pendingJob).toBeTruthy();
    expect(store.cancelPrint(pendingJob!.id)).toBe(true);
    expect(store.printJobs.find((job) => job.id === pendingJob!.id)?.status).toBe("cancelled");

    store.updateCurrentDraft("请整理一条准备直接排队的纸条");
    store.setSendConfirmation(false);
    await store.sendCurrentDraft();
    store.toggleConversationMessageSelection(store.activeConversation!.messages.at(-1)!.id);
    const queuedJob = await store.createPrintFromSelectedMessages();

    expect(queuedJob?.status).toBe("queued");
    expect(store.cancelPrint(queuedJob!.id)).toBe(true);
    expect(store.printJobs.find((job) => job.id === queuedJob!.id)?.status).toBe("cancelled");
  });

  it("can delete the active conversation and keep the workspace usable", () => {
    const store = useWorkspaceStore();
    const currentId = store.activeConversationId;
    const previousLength = store.conversations.length;

    expect(store.deleteConversation(currentId)).toBe(true);
    expect(store.conversations).toHaveLength(previousLength - 1);
    expect(store.activeConversation).not.toBeNull();
    expect(store.activeConversationId).not.toBe(currentId);
  });

  it("supports print queue updates, source state cycling, and service binding", async () => {
    const store = useWorkspaceStore();
    const pendingJob = store.pendingPrintJobs.find((job: PrintJob) => job.status === "pending");
    const schedule = store.schedules[0];
    const source = store.sources[0];

    expect(pendingJob).toBeTruthy();

    await store.confirmPrint(pendingJob!.id);
    store.updateScheduleDevice(schedule.id, "device-bedroom");
    store.toggleSourceConnection(source.id);
    store.setTheme("soft");
    store.setLoginProtection(false);
    store.bindService();
    await store.logout();

    expect(store.printJobs.find((job) => job.id === pendingJob!.id)?.status).toBe("queued");
    expect(store.schedules.find((item) => item.id === schedule.id)?.deviceId).toBe(
      "device-bedroom",
    );
    expect(store.sources.find((item) => item.id === source.id)?.status).toBe("disconnected");
    expect(store.selectedTheme).toBe("soft");
    expect(store.loginProtectionEnabled).toBe(false);
    expect(store.serviceBinding.bound).toBe(true);
    expect(store.isAuthenticated).toBe(false);
  });

  it("clears the local auth state after a successful password change", async () => {
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

    await expect(store.changePassword("demo-password", "next-password")).resolves.toBe(true);
    expect(store.isAuthenticated).toBe(false);
    expect(store.flashMessage).toBe("密码已更新，请重新登录。");
  });

  it("keeps auth tokens out of the workspace snapshot and stores them only for the tab session", async () => {
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

    await Promise.resolve();

    expect(window.sessionStorage.getItem("ink.auth.session.v1")).toContain("refresh-token");
    expect(window.localStorage.getItem("ink.workspace.v1")).not.toContain("refresh-token");
    expect(window.localStorage.getItem("ink.workspace.v1")).not.toContain("access-token");
  });
});

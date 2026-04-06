import { createPinia, setActivePinia } from "pinia";

import { useWorkspaceStore } from "@/stores/workspace";
import type { PrintJob } from "@/types/workspace";

describe("workspace store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("exposes stable defaults and derived summaries", () => {
    const store = useWorkspaceStore();

    expect(store.activeDeviceLabel).toBe("书桌咕咕机");
    expect(store.activeModelLabel).toBe("清楚温柔");
    expect(store.todayPrintCount).toBe(2);
    expect(store.welcomeLabel).toBe("简单一点，也可以很舒服");
    expect(store.isConfigured).toBe(true);
    expect(store.pendingConfirmationCount).toBe(1);
    expect(store.enabledSchedulesCount).toBe(2);
  });

  it("updates configuration from settings actions instead of local hard-coded values", () => {
    const store = useWorkspaceStore();

    store.setDefaultDevice("device-bedroom");

    expect(store.activeDeviceLabel).toBe("卧室咕咕机");
    expect(store.isConfigured).toBe(true);

    store.setAnswerStyle("warm-encouraging");
    expect(store.activeModelLabel).toBe("温柔鼓励");
    expect(store.welcomeLabel).toBe("慢一点，也能把想说的话说好");
  });

  it("can generate a reply and create a pending print from the active conversation", async () => {
    const store = useWorkspaceStore();
    const previousCount = store.pendingPrintJobs.length;

    store.updateCurrentDraft("请帮我整理一句适合明早看的提醒");
    await store.sendCurrentDraft();
    const printJob = await store.createPrintFromSelectedReply();

    expect(store.activeConversation?.messages.at(-1)?.role).toBe("assistant");
    expect(store.selectedAssistantMessage?.text).toContain("请帮我整理一句适合明早看的提醒");
    expect(store.pendingPrintJobs).toHaveLength(previousCount + 1);
    expect(printJob?.status).toBe("pending");
  });

  it("supports print queue updates, source state cycling, and service binding", async () => {
    const store = useWorkspaceStore();
    const pendingJob = store.pendingPrintJobs.find((job: PrintJob) => job.status === "pending");
    const schedule = store.schedules[0];
    const source = store.sources[0];

    expect(pendingJob).toBeTruthy();

    await store.confirmPrint(pendingJob!.id);
    store.updateScheduleDevice(schedule.id, "device-bedroom");
    store.cycleSourceStatus(source.id);
    store.setTheme("soft");
    store.setLoginProtection(false);
    store.bindService();
    store.logout();

    expect(store.printJobs.find((job) => job.id === pendingJob!.id)?.status).toBe("queued");
    expect(store.schedules.find((item) => item.id === schedule.id)?.deviceId).toBe(
      "device-bedroom",
    );
    expect(store.sources.find((item) => item.id === source.id)?.status).toBe("error");
    expect(store.selectedTheme).toBe("soft");
    expect(store.loginProtectionEnabled).toBe(false);
    expect(store.serviceBinding.bound).toBe(true);
    expect(store.isAuthenticated).toBe(false);
  });
});

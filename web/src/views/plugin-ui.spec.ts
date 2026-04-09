import { mount } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { createTestRouter } from "@/router";
import { useWorkspaceStore } from "@/stores/workspace";
import PrintsView from "@/views/PrintsView.vue";
import SettingsView from "@/views/SettingsView.vue";

function createPluginDetails() {
  return {
    installation: {
      id: "plugin-installation-1",
      pluginKey: "demo-source",
      sourceType: "upload" as const,
      displayName: "Demo Source",
      version: "1.0.0",
      runtimeType: "node" as const,
      status: "ready" as const,
    },
    manifest: {
      schemaVersion: 1,
      kind: "source" as const,
      pluginKey: "demo-source",
      name: "Demo Source",
      version: "1.0.0",
      description: "A demo source plugin.",
      runtime: {
        type: "node" as const,
      },
      entrypoints: {
        validate: {
          command: ["pnpm", "validate"],
        },
        fetch: {
          command: ["pnpm", "fetch"],
        },
      },
      workspaceConfigSchema: [
        {
          key: "feedUrl",
          label: "Feed URL",
          type: "url" as const,
          required: true,
          description: "https://example.com/feed",
        },
        {
          key: "apiKey",
          label: "API Key",
          type: "secret" as const,
          required: false,
        },
      ],
      scheduleConfigSchema: [
        {
          key: "mode",
          label: "模式",
          type: "select" as const,
          required: true,
          defaultValue: "brief",
          options: [
            {
              label: "简报",
              value: "brief",
            },
            {
              label: "全文",
              value: "full",
            },
          ],
        },
        {
          key: "keyword",
          label: "关键词",
          type: "text" as const,
          required: false,
          description: "morning digest",
        },
      ],
    },
    binding: {
      id: "binding-1",
      enabled: true,
      status: "connected" as const,
      config: {
        feedUrl: "https://example.com/feed",
      },
    },
  };
}

function createSchedule() {
  return {
    id: "schedule-1",
    title: "Morning Digest",
    source: "Demo Source",
    timeLabel: "每天 08:30",
    deviceId: "device-1",
    enabled: true,
    pluginInstallationId: "plugin-installation-1",
    pluginBindingId: "binding-1",
    frequencyType: "daily" as const,
    timezone: "Asia/Shanghai",
    hour: 8,
    minute: 30,
    weekdays: [],
    scheduleConfig: {
      mode: "brief",
    },
    pluginDisplayName: "Demo Source",
    nextRunAt: new Date("2026-04-11T00:30:00.000Z").toISOString(),
    lastRunAt: new Date("2026-04-10T00:30:00.000Z").toISOString(),
    lastError: "",
    sourceLabel: "Demo Source",
  };
}

async function createWorkspaceContext(path: string, role: "admin" | "member" = "member") {
  const pinia = createPinia();
  setActivePinia(pinia);
  const store = useWorkspaceStore();

  store.authUser = {
    id: "user-1",
    email: role === "admin" ? "admin" : "member",
    name: role === "admin" ? "Administrator" : "Ink User",
    role,
  };
  store.authSession = {
    accessToken: "access-token",
    refreshToken: "refresh-token",
    accessTokenExpiresAt: new Date(Date.now() + 60_000).toISOString(),
  };

  const router = createTestRouter(pinia);
  router.push(path);
  await router.isReady();

  return { pinia, router, store };
}

describe("plugin ui flows", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
    window.sessionStorage.clear();
    vi.stubGlobal(
      "confirm",
      vi.fn(() => true),
    );
  });

  it("renders admin plugin controls and submits plugin upload/test/save actions", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/settings", "admin");
    const plugin = createPluginDetails();
    store.availablePlugins = [plugin];
    store.adminPlugins = [plugin];

    const uploadSpy = vi.spyOn(store, "uploadPlugin").mockResolvedValue(plugin);
    const testSpy = vi
      .spyOn(store, "testPluginConfiguration")
      .mockResolvedValue({ valid: true, errors: [] });
    const saveSpy = vi.spyOn(store, "savePluginConfiguration").mockResolvedValue(plugin);

    const wrapper = mount(SettingsView, {
      global: {
        plugins: [pinia, router],
      },
    });

    expect(wrapper.text()).toContain("本地上传插件 ZIP");
    expect(wrapper.text()).toContain("已安装插件");
    expect(wrapper.text()).toContain("Demo Source");

    const fileInput = wrapper.find("input[type='file']");
    Object.defineProperty(fileInput.element, "files", {
      value: [new File(["zip-content"], "demo-source.zip", { type: "application/zip" })],
    });
    await fileInput.trigger("change");

    const feedInput = wrapper.find("input[placeholder='https://example.com/feed']");
    await feedInput.setValue("https://example.com/updated");

    await wrapper
      .findAll("button")
      .find((button) => button.text() === "测试插件")
      ?.trigger("click");
    await wrapper.findAll("form").at(-1)?.trigger("submit");

    expect(uploadSpy).toHaveBeenCalledTimes(1);
    expect(testSpy).toHaveBeenCalledWith(
      "plugin-installation-1",
      {
        feedUrl: "https://example.com/updated",
      },
      {
        apiKey: "",
      },
      true,
    );
    expect(saveSpy).toHaveBeenCalledWith(
      "plugin-installation-1",
      {
        feedUrl: "https://example.com/updated",
      },
      {
        apiKey: "",
      },
      true,
    );
  });

  it("creates and deletes authenticated plugin schedules from the prints view", async () => {
    const { pinia, router, store } = await createWorkspaceContext("/prints");
    const plugin = createPluginDetails();
    const expectedTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone || "Asia/Shanghai";
    store.availablePlugins = [plugin];
    store.remoteSchedules = [createSchedule()];
    store.devices = [
      {
        id: "device-1",
        name: "Desk Printer",
        status: "connected",
        note: "Primary device",
      },
    ];
    store.defaultDeviceId = "device-1";

    const createSpy = vi.spyOn(store, "createSchedule").mockResolvedValue(createSchedule());
    const deleteSpy = vi.spyOn(store, "deleteSchedule").mockResolvedValue(true);

    const wrapper = mount(PrintsView, {
      global: {
        plugins: [pinia, router],
      },
    });

    expect(wrapper.text()).toContain("Demo Source");

    await wrapper
      .findAll("button")
      .find((button) => button.text() === "新建定时任务")
      ?.trigger("click");

    await wrapper.find("input[placeholder='例如：晨间提醒']").setValue("Weekly Digest");
    const frequencySelect = wrapper
      .findAll("select")
      .find((select) => select.text().includes("每天") && select.text().includes("每周"));
    await frequencySelect?.setValue("weekly");
    await wrapper
      .findAll("button")
      .find((button) => button.text() === "周一")
      ?.trigger("click");
    await wrapper.find("input[placeholder='morning digest']").setValue("paper notes");
    await wrapper.findAll("form").at(-1)?.trigger("submit");

    expect(createSpy).toHaveBeenCalledWith({
      title: "Weekly Digest",
      deviceId: "device-1",
      pluginInstallationId: "plugin-installation-1",
      frequencyType: "weekly",
      timezone: expectedTimezone,
      hour: 19,
      minute: 30,
      weekdays: [1],
      scheduleConfig: {
        mode: "brief",
        keyword: "paper notes",
      },
    });

    await wrapper
      .findAll("button")
      .find((button) => button.text() === "删除")
      ?.trigger("click");

    expect(deleteSpy).toHaveBeenCalledWith("schedule-1");
  });
});

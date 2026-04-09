import { afterEach, beforeEach, expect, it, vi } from "vitest";

import { AuthApiError } from "@/services/auth";
import {
  createPrintSchedule,
  deletePrintSchedule,
  disablePlugin,
  fetchAdminPlugins,
  fetchPlugin,
  fetchPlugins,
  fetchPrintSchedules,
  savePluginBinding,
  testPluginBinding,
  togglePrintSchedule,
  updatePrintSchedule,
  uploadPluginZip,
} from "@/services/plugins";

const fetchMock = vi.fn<typeof fetch>();

function createPluginResponse() {
  return {
    plugin: {
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
        workspaceConfigSchema: [],
        scheduleConfigSchema: [],
      },
      binding: {
        id: "binding-1",
        enabled: true,
        status: "connected" as const,
        config: {
          feedUrl: "https://example.com/feed",
        },
      },
    },
  };
}

function createScheduleResponse() {
  return {
    schedule: {
      id: "schedule-1",
      title: "Morning Digest",
      pluginInstallationId: "plugin-installation-1",
      pluginBindingId: "binding-1",
      pluginDisplayName: "Demo Source",
      frequencyType: "daily" as const,
      timezone: "Asia/Shanghai",
      hour: 8,
      minute: 30,
      weekdays: [],
      scheduleConfig: {
        mode: "brief",
      },
      deviceId: "device-1",
      enabled: true,
      nextRunAt: new Date("2026-04-11T00:30:00.000Z").toISOString(),
      timeLabel: "每天 08:30",
      sourceLabel: "Demo Source",
    },
  };
}

describe("plugins service", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    fetchMock.mockReset();
  });

  it("handles plugin and schedule endpoints through authenticated requests", async () => {
    const pluginResponse = createPluginResponse();
    const scheduleResponse = createScheduleResponse();

    fetchMock
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ plugins: [pluginResponse.plugin] }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(pluginResponse), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(pluginResponse), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ plugins: [pluginResponse.plugin] }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(pluginResponse), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(pluginResponse), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            result: {
              valid: true,
              errors: [],
            },
          }),
          { status: 200 },
        ),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ schedules: [scheduleResponse.schedule] }), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(scheduleResponse), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(scheduleResponse), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(scheduleResponse), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    await expect(fetchAdminPlugins("access-token")).resolves.toEqual({
      plugins: [pluginResponse.plugin],
    });
    await expect(
      uploadPluginZip(
        "access-token",
        new File(["zip-content"], "demo-source.zip", { type: "application/zip" }),
      ),
    ).resolves.toEqual(pluginResponse.plugin);
    await expect(disablePlugin("access-token", "plugin-installation-1")).resolves.toEqual(
      pluginResponse.plugin,
    );
    await expect(fetchPlugins("access-token")).resolves.toEqual({
      plugins: [pluginResponse.plugin],
    });
    await expect(fetchPlugin("access-token", "plugin-installation-1")).resolves.toEqual(
      pluginResponse.plugin,
    );
    await expect(
      savePluginBinding("access-token", "plugin-installation-1", {
        enabled: true,
        config: {
          feedUrl: "https://example.com/feed",
        },
        secrets: {
          apiKey: "secret",
        },
      }),
    ).resolves.toEqual(pluginResponse.plugin);
    await expect(
      testPluginBinding("access-token", "plugin-installation-1", {
        enabled: true,
        config: {
          feedUrl: "https://example.com/feed",
        },
        secrets: {
          apiKey: "secret",
        },
      }),
    ).resolves.toEqual({
      valid: true,
      errors: [],
    });
    await expect(fetchPrintSchedules("access-token")).resolves.toEqual({
      schedules: [scheduleResponse.schedule],
    });
    await expect(
      createPrintSchedule("access-token", {
        title: "Morning Digest",
        pluginInstallationId: "plugin-installation-1",
        frequencyType: "daily",
        timezone: "Asia/Shanghai",
        hour: 8,
        minute: 30,
        weekdays: [],
        scheduleConfig: {
          mode: "brief",
        },
        deviceId: "device-1",
        enabled: true,
      }),
    ).resolves.toEqual(scheduleResponse.schedule);
    await expect(
      updatePrintSchedule("access-token", "schedule-1", {
        title: "Morning Digest",
        pluginInstallationId: "plugin-installation-1",
        frequencyType: "daily",
        timezone: "Asia/Shanghai",
        hour: 8,
        minute: 30,
        weekdays: [],
        scheduleConfig: {
          mode: "brief",
        },
        deviceId: "device-1",
        enabled: true,
      }),
    ).resolves.toEqual(scheduleResponse.schedule);
    await expect(togglePrintSchedule("access-token", "schedule-1")).resolves.toEqual(
      scheduleResponse.schedule,
    );
    await expect(deletePrintSchedule("access-token", "schedule-1")).resolves.toBeUndefined();

    expect(fetchMock).toHaveBeenCalledTimes(12);
    expect(fetchMock.mock.calls[0]?.[0]).toBe("/api/v1/admin/plugins");
    expect(fetchMock.mock.calls[1]?.[1]?.body).toBeInstanceOf(FormData);
    expect(fetchMock.mock.calls[7]?.[0]).toBe("/api/v1/print-schedules");
    expect(fetchMock.mock.calls[11]?.[0]).toBe("/api/v1/print-schedules/schedule-1");
  });

  it("maps api and network failures into AuthApiError", async () => {
    fetchMock
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            code: "forbidden",
            message: "不允许访问该插件。",
            requestId: "req-plugin-1",
          }),
          { status: 403 },
        ),
      )
      .mockRejectedValueOnce(new Error("network down"))
      .mockResolvedValueOnce(new Response("internal error", { status: 500 }));

    await expect(fetchPlugins("access-token")).rejects.toEqual(
      new AuthApiError(403, "forbidden", "不允许访问该插件。", "req-plugin-1"),
    );
    await expect(fetchPrintSchedules("access-token")).rejects.toMatchObject({
      status: 0,
      code: "network_error",
    });
    await expect(togglePrintSchedule("access-token", "schedule-1")).rejects.toEqual(
      new AuthApiError(500, "request_failed", "请求失败，请稍后重试。"),
    );
  });
});

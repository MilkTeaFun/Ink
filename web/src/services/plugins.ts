import { AuthApiError } from "@/services/auth";
import type { PluginDetails, PluginValidationResult, PrintScheduleView } from "@/types/plugins";

interface ApiErrorResponse {
  code?: string;
  message?: string;
  requestId?: string;
}

export interface PluginBindingPayload {
  enabled: boolean;
  config: Record<string, unknown>;
  secrets: Record<string, string>;
}

export interface PrintSchedulePayload {
  title: string;
  pluginInstallationId: string;
  frequencyType: "daily" | "weekly";
  timezone: string;
  hour: number;
  minute: number;
  weekdays: number[];
  scheduleConfig: Record<string, unknown>;
  deviceId: string;
  enabled: boolean;
}

async function request<T>(input: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (!headers.has("Content-Type") && !(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  let response: Response;
  try {
    response = await fetch(input, {
      ...init,
      headers,
    });
  } catch (error) {
    throw new AuthApiError(
      0,
      "network_error",
      error instanceof Error
        ? `网络异常，请检查连接后重试。${error.message ? ` (${error.message})` : ""}`
        : "网络异常，请检查连接后重试。",
    );
  }

  if (!response.ok) {
    let errorPayload: ApiErrorResponse | null = null;

    try {
      errorPayload = (await response.json()) as ApiErrorResponse;
    } catch {
      errorPayload = null;
    }

    throw new AuthApiError(
      response.status,
      errorPayload?.code ?? "request_failed",
      errorPayload?.message ?? "请求失败，请稍后重试。",
      errorPayload?.requestId,
    );
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export async function fetchAdminPlugins(accessToken: string) {
  return request<{ plugins: PluginDetails[] }>("/api/v1/admin/plugins", {
    method: "GET",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

export async function uploadPluginZip(accessToken: string, file: File) {
  const formData = new FormData();
  formData.set("file", file);

  const response = await request<{ plugin: PluginDetails }>("/api/v1/admin/plugins/upload", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: formData,
  });

  return response.plugin;
}

export async function disablePlugin(accessToken: string, installationId: string) {
  const response = await request<{ plugin: PluginDetails }>(
    `/api/v1/admin/plugins/${installationId}/disable`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({}),
    },
  );

  return response.plugin;
}

export async function fetchPlugins(accessToken: string) {
  return request<{ plugins: PluginDetails[] }>("/api/v1/plugins", {
    method: "GET",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

export async function fetchPlugin(accessToken: string, installationId: string) {
  const response = await request<{ plugin: PluginDetails }>(`/api/v1/plugins/${installationId}`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });

  return response.plugin;
}

export async function savePluginBinding(
  accessToken: string,
  installationId: string,
  payload: PluginBindingPayload,
) {
  const response = await request<{ plugin: PluginDetails }>(
    `/api/v1/plugins/${installationId}/binding`,
    {
      method: "PUT",
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(payload),
    },
  );

  return response.plugin;
}

export async function testPluginBinding(
  accessToken: string,
  installationId: string,
  payload: PluginBindingPayload,
) {
  const response = await request<{ result: PluginValidationResult }>(
    `/api/v1/plugins/${installationId}/test`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(payload),
    },
  );

  return response.result;
}

export async function fetchPrintSchedules(accessToken: string) {
  return request<{ schedules: PrintScheduleView[] }>("/api/v1/print-schedules", {
    method: "GET",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

export async function createPrintSchedule(accessToken: string, payload: PrintSchedulePayload) {
  const response = await request<{ schedule: PrintScheduleView }>("/api/v1/print-schedules", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });

  return response.schedule;
}

export async function updatePrintSchedule(
  accessToken: string,
  scheduleId: string,
  payload: PrintSchedulePayload,
) {
  const response = await request<{ schedule: PrintScheduleView }>(
    `/api/v1/print-schedules/${scheduleId}`,
    {
      method: "PUT",
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify(payload),
    },
  );

  return response.schedule;
}

export async function togglePrintSchedule(accessToken: string, scheduleId: string) {
  const response = await request<{ schedule: PrintScheduleView }>(
    `/api/v1/print-schedules/${scheduleId}/toggle`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({}),
    },
  );

  return response.schedule;
}

export async function deletePrintSchedule(accessToken: string, scheduleId: string) {
  return request<void>(`/api/v1/print-schedules/${scheduleId}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

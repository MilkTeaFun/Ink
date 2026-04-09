import { AuthApiError } from "@/services/auth";

export interface AIConfigSummary {
  bound: boolean;
  providerName: string;
  providerType: string;
  baseUrl: string;
  model: string;
  keyConfigured: boolean;
  updatedAt?: string;
}

export interface SaveAIConfigPayload {
  providerName: string;
  providerType: string;
  baseUrl: string;
  model: string;
  apiKey: string;
}

export interface AIReplyMessage {
  role: "system" | "user" | "assistant";
  content: string;
}

export interface AIReplyResult {
  content: string;
  model: string;
  providerName: string;
}

interface ApiErrorResponse {
  code?: string;
  message?: string;
  requestId?: string;
}

async function request<T>(input: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");

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

  return (await response.json()) as T;
}

export async function fetchAIConfigSummary(accessToken: string) {
  return request<AIConfigSummary>("/api/v1/ai/config", {
    method: "GET",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

export async function saveAIConfig(accessToken: string, payload: SaveAIConfigPayload) {
  return request<AIConfigSummary>("/api/v1/admin/ai/config", {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
}

export async function generateAIReply(
  accessToken: string,
  payload: { messages: AIReplyMessage[] },
) {
  return request<AIReplyResult>("/api/v1/ai/reply", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
}

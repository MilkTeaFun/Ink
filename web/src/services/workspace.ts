import { AuthApiError } from "@/services/auth";
import type { User, WorkspaceState } from "@/types/workspace";

export interface CreateUserPayload {
  email: string;
  name: string;
  password: string;
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

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export async function fetchWorkspaceStateWithApi(accessToken: string) {
  return request<WorkspaceState>("/api/v1/workspace", {
    method: "GET",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

export async function saveWorkspaceStateWithApi(accessToken: string, state: WorkspaceState) {
  return request<WorkspaceState>("/api/v1/workspace", {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(state),
  });
}

export async function createUserWithApi(accessToken: string, payload: CreateUserPayload) {
  const response = await request<{ user: User }>("/api/v1/admin/users", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });

  return response.user;
}

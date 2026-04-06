import type { AuthSession, User } from "@/types/workspace";

export interface LoginPayload {
  email: string;
  password: string;
}

export interface LogoutPayload {
  accessToken: string;
  refreshToken: string;
}

export interface ChangePasswordPayload {
  accessToken: string;
  currentPassword: string;
  newPassword: string;
}

interface AuthResponse {
  user: User;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

interface MeResponse {
  user: User;
}

interface ApiErrorResponse {
  code?: string;
  message?: string;
  requestId?: string;
}

export class AuthApiError extends Error {
  status: number;
  code: string;
  requestId?: string;

  constructor(status: number, code: string, message: string, requestId?: string) {
    super(message);
    this.name = "AuthApiError";
    this.status = status;
    this.code = code;
    this.requestId = requestId;
  }
}

function buildSession(response: AuthResponse): AuthSession {
  return {
    accessToken: response.accessToken,
    refreshToken: response.refreshToken,
    accessTokenExpiresAt: new Date(Date.now() + response.expiresIn * 1000).toISOString(),
  };
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

export async function loginWithApi(payload: LoginPayload) {
  const response = await request<AuthResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(payload),
  });

  return {
    user: response.user,
    session: buildSession(response),
  };
}

export async function refreshAuthSession(refreshToken: string) {
  const response = await request<AuthResponse>("/api/v1/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refreshToken }),
  });

  return {
    user: response.user,
    session: buildSession(response),
  };
}

export async function fetchCurrentUser(accessToken: string) {
  const response = await request<MeResponse>("/api/v1/auth/me", {
    method: "GET",
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });

  return response.user;
}

export async function logoutWithApi(payload: LogoutPayload) {
  await request<void>("/api/v1/auth/logout", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${payload.accessToken}`,
    },
    body: JSON.stringify({
      refreshToken: payload.refreshToken,
    }),
  });
}

export async function changePasswordWithApi(payload: ChangePasswordPayload) {
  await request<void>("/api/v1/auth/change-password", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${payload.accessToken}`,
    },
    body: JSON.stringify({
      currentPassword: payload.currentPassword,
      newPassword: payload.newPassword,
    }),
  });
}

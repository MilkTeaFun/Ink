export type DeviceStatus = "connected" | "pending" | "offline";
export type PrintStatus = "pending" | "queued" | "completed" | "failed" | "cancelled";
export type ConversationMessageRole = "user" | "assistant";
export type ThemeMode = "soft" | "light" | "system";
export type SourceConnectionStatus = "connected" | "disconnected" | "error";

export interface User {
  id: string;
  email: string;
  name: string;
}

export interface AuthSession {
  accessToken: string;
  refreshToken: string;
  accessTokenExpiresAt: string;
}

export interface Device {
  id: string;
  name: string;
  status: DeviceStatus;
  note: string;
}

export interface ConversationMessage {
  id: string;
  role: ConversationMessageRole;
  text: string;
  createdAt: string;
}

export interface Conversation {
  id: string;
  title: string;
  preview: string;
  updatedAt: string;
  draft: string;
  messages: ConversationMessage[];
}

export interface PrintJob {
  id: string;
  title: string;
  source: string;
  deviceId: string;
  status: PrintStatus;
  createdAt: string;
  updatedAt: string;
  content: string;
}

export interface Schedule {
  id: string;
  title: string;
  source: string;
  timeLabel: string;
  deviceId: string;
  enabled: boolean;
}

export interface SourceConnection {
  id: string;
  name: string;
  type: string;
  note: string;
  status: SourceConnectionStatus;
}

export interface Preferences {
  loginProtectionEnabled: boolean;
  sendConfirmationEnabled: boolean;
  theme: ThemeMode;
  defaultDeviceId: string;
}

export interface ServiceBinding {
  providerName: string | null;
  modelName: string;
  bound: boolean;
}

export interface PersistedWorkspaceState {
  authUser: User | null;
  authSession: AuthSession | null;
  devices: Device[];
  conversations: Conversation[];
  activeConversationId: string;
  printJobs: PrintJob[];
  schedules: Schedule[];
  sources: SourceConnection[];
  preferences: Preferences;
  serviceBinding: ServiceBinding;
}

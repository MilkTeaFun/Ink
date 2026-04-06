export type DeviceStatus = "connected" | "pending" | "offline";
export type PrintStatus = "pending" | "queued" | "completed" | "failed";
export type ConversationMessageRole = "user" | "assistant";
export type ThemeMode = "soft" | "light" | "system";
export type SourceConnectionStatus = "connected" | "disconnected" | "error";
export type AnswerStyle = "clear-gentle" | "warm-encouraging" | "concise-direct";
export type NoteStyle = "clean" | "gentle" | "list";
export type ResponseLength = "short" | "medium" | "long";

export interface User {
  id: string;
  email: string;
  name: string;
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
  answerStyle: AnswerStyle;
  noteStyle: NoteStyle;
  responseLength: ResponseLength;
  defaultDeviceId: string;
}

export interface ServiceBinding {
  providerName: string | null;
  modelName: string;
  bound: boolean;
}

export interface PersistedWorkspaceState {
  authUser: User | null;
  devices: Device[];
  conversations: Conversation[];
  activeConversationId: string;
  printJobs: PrintJob[];
  schedules: Schedule[];
  sources: SourceConnection[];
  preferences: Preferences;
  serviceBinding: ServiceBinding;
}

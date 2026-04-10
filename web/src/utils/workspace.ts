import type {
  DeviceStatus,
  PrintStatus,
  SourceConnectionStatus,
  ThemeMode,
  UserRole,
} from "@/types/workspace";
import type { PluginDetails, PluginInstallationStatus } from "@/types/plugins";

export type ResolvedThemeMode = Exclude<ThemeMode, "system">;
type BadgeTone = "success" | "warning" | "danger" | "neutral";
type SummaryTone = "green" | "amber" | "stone" | "neutral";

function getInsetBadgeClass(tone: BadgeTone) {
  switch (tone) {
    case "success":
      return "bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20 ring-inset";
    case "warning":
      return "bg-amber-50 text-amber-700 ring-1 ring-amber-600/20 ring-inset";
    case "danger":
      return "bg-rose-50 text-rose-700 ring-1 ring-rose-600/20 ring-inset";
    default:
      return "bg-stone-100 text-stone-700 ring-1 ring-stone-500/10 ring-inset";
  }
}

function getRoleBadgeClass(tone: "admin" | "member") {
  if (tone === "admin") {
    return "bg-amber-100 text-amber-800";
  }

  return "bg-stone-100 text-stone-800";
}

export function createId(prefix: string) {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`;
}

export function formatRelativeTimestamp(iso: string) {
  const target = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - target.getTime();
  const diffMinutes = Math.max(0, Math.floor(diffMs / 60000));

  if (diffMinutes < 1) {
    return "刚刚";
  }

  if (diffMinutes < 60) {
    return `${diffMinutes} 分钟前`;
  }

  if (isSameDay(target, now)) {
    return `今天 ${target.toLocaleTimeString("zh-CN", {
      hour: "2-digit",
      minute: "2-digit",
    })}`;
  }

  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);

  if (isSameDay(target, yesterday)) {
    return `昨天 ${target.toLocaleTimeString("zh-CN", {
      hour: "2-digit",
      minute: "2-digit",
    })}`;
  }

  return `${target.getFullYear()}-${String(target.getMonth() + 1).padStart(2, "0")}-${String(
    target.getDate(),
  ).padStart(2, "0")} ${target.toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  })}`;
}

export function isSameDay(left: Date, right: Date) {
  return (
    left.getFullYear() === right.getFullYear() &&
    left.getMonth() === right.getMonth() &&
    left.getDate() === right.getDate()
  );
}

export function getDeviceStatusLabel(status: DeviceStatus) {
  if (status === "connected") {
    return "已连接";
  }

  if (status === "pending") {
    return "待绑定";
  }

  return "已离线";
}

export function getDeviceStatusBadgeClass(status: DeviceStatus) {
  if (status === "connected") {
    return getInsetBadgeClass("success");
  }

  if (status === "pending") {
    return getInsetBadgeClass("warning");
  }

  return getInsetBadgeClass("neutral");
}

export function getPrintStatusLabel(status: PrintStatus) {
  if (status === "pending") {
    return "待确认";
  }

  if (status === "queued") {
    return "排队中";
  }

  if (status === "completed") {
    return "已完成";
  }

  if (status === "cancelled") {
    return "已取消";
  }

  return "失败";
}

export function getPrintStatusBadgeClass(status: PrintStatus) {
  if (status === "completed") {
    return getInsetBadgeClass("success");
  }

  if (status === "queued") {
    return getInsetBadgeClass("warning");
  }

  if (status === "failed") {
    return getInsetBadgeClass("danger");
  }

  return getInsetBadgeClass("neutral");
}

export function getSourceStatusLabel(status: SourceConnectionStatus) {
  if (status === "connected") {
    return "已连接";
  }

  if (status === "error") {
    return "异常";
  }

  return "未连接";
}

export function getSourceStatusBadgeClass(status: SourceConnectionStatus) {
  if (status === "connected") {
    return getInsetBadgeClass("success");
  }

  if (status === "error") {
    return getInsetBadgeClass("danger");
  }

  return getInsetBadgeClass("neutral");
}

export function getPluginInstallationStatusLabel(status: PluginInstallationStatus) {
  switch (status) {
    case "installing":
      return "安装中";
    case "ready":
      return "可用";
    case "failed":
      return "异常";
    case "disabled":
      return "已停用";
    default:
      return status;
  }
}

export function getPluginInstallationStatusBadgeClass(status: PluginInstallationStatus) {
  switch (status) {
    case "ready":
      return getInsetBadgeClass("success");
    case "failed":
      return getInsetBadgeClass("danger");
    case "disabled":
      return getInsetBadgeClass("neutral");
    default:
      return getInsetBadgeClass("warning");
  }
}

export function getPluginBindingStatusLabel(plugin: PluginDetails) {
  if (plugin.installation.status === "disabled") {
    return "已停用";
  }

  if (!plugin.binding?.enabled) {
    return "未连接";
  }

  return getSourceStatusLabel(plugin.binding.status);
}

export function getPluginBindingStatusBadgeClass(plugin: PluginDetails) {
  if (plugin.installation.status === "disabled" || !plugin.binding?.enabled) {
    return getInsetBadgeClass("neutral");
  }

  return plugin.binding.status === "error"
    ? getInsetBadgeClass("danger")
    : getInsetBadgeClass("success");
}

export function getUserRoleLabel(role: UserRole) {
  return role === "admin" ? "管理员" : "成员";
}

export function getUserRoleBadgeClass(role: UserRole) {
  return getRoleBadgeClass(role);
}

export function getServiceBindingStatusBadgeClass(bound: boolean) {
  return getSourceStatusBadgeClass(bound ? "connected" : "disconnected");
}

export function getSummaryProgressClass(tone: SummaryTone | string) {
  switch (tone) {
    case "green":
      return "bg-emerald-500";
    case "amber":
      return "bg-amber-500";
    case "stone":
      return "bg-stone-400";
    default:
      return "bg-stone-800";
  }
}

export function normalizeThemeMode(theme: unknown): ThemeMode {
  if (theme === "dark" || theme === "light" || theme === "system") {
    return theme;
  }

  if (theme === "soft") {
    return "light";
  }

  return "light";
}

export function resolveThemeMode(
  theme: ThemeMode,
  prefersDark: boolean,
): ResolvedThemeMode {
  if (theme === "system") {
    return prefersDark ? "dark" : "light";
  }

  return theme;
}

export function getThemeDescription(theme: ThemeMode) {
  if (theme === "dark") {
    return "深色";
  }

  if (theme === "system") {
    return "跟随系统";
  }

  return "浅色";
}

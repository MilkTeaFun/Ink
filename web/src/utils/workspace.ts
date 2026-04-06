import type {
  DeviceStatus,
  PrintStatus,
  SourceConnectionStatus,
  ThemeMode,
} from "@/types/workspace";

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

  return "失败";
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

export function getThemeDescription(theme: ThemeMode) {
  if (theme === "soft") {
    return "柔光";
  }

  if (theme === "system") {
    return "系统跟随";
  }

  return "浅色";
}

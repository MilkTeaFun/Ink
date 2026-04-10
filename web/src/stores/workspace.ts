import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";

import type { AIConfigSummary } from "@/services/ai";
import { fetchAIConfigSummary, generateAIReply, saveAIConfig } from "@/services/ai";
import {
  AuthApiError,
  changePasswordWithApi,
  fetchCurrentUser,
  loginWithApi,
  logoutWithApi,
  refreshAuthSession,
} from "@/services/auth";
import { generateReplyWithMockService } from "@/services/mockInk";
import {
  createPrintSchedule,
  deletePrintSchedule,
  disablePlugin,
  fetchAdminPlugins,
  fetchPlugins,
  fetchPrintSchedules,
  savePluginBinding,
  testPluginBinding,
  togglePrintSchedule,
  updatePrintSchedule,
  uploadPluginZip,
} from "@/services/plugins";
import {
  bindPrinter,
  cancelPrintJob,
  createPrintJob,
  deletePrinter,
  fetchPrintJobs,
  fetchPrinters,
  submitPrintJob,
  updatePrintJobDevice as updatePrintJobDeviceWithApi,
} from "@/services/printers";
import {
  createUserWithApi,
  fetchWorkspaceStateWithApi,
  saveWorkspaceStateWithApi,
} from "@/services/workspace";
import type { PluginDetails, PrintScheduleView } from "@/types/plugins";
import type {
  AuthSession,
  Conversation,
  ConversationMessage,
  Device,
  PersistedWorkspaceState,
  Preferences,
  PrintJob,
  Schedule,
  ServiceBinding,
  SourceConnection,
  ThemeMode,
  User,
  WorkspaceState,
} from "@/types/workspace";
import {
  createId,
  formatRelativeTimestamp,
  getDeviceStatusLabel,
  getPrintStatusLabel,
  getSourceStatusLabel,
  normalizeThemeMode,
} from "@/utils/workspace";

const STORAGE_KEY = "ink.workspace.v1";
const AUTH_SESSION_STORAGE_KEY = "ink.auth.session.v1";
const REMOTE_SAVE_DEBOUNCE_MS = 180;
const REMOTE_PRINT_STATUS_POLL_MS = 5000;
const REMOTE_PRINT_STATUS_INITIAL_POLL_MS = 1500;

function getNow() {
  return new Date().toISOString();
}

function createEmptyConversation(): Conversation {
  return {
    id: createId("conversation"),
    title: "新对话",
    preview: "从这里开始整理新的内容",
    updatedAt: getNow(),
    draft: "",
    messages: [],
  };
}

function normalizeConversations(conversations: Conversation[]): Conversation[] {
  return conversations.length > 0 ? conversations : [createEmptyConversation()];
}

function resolveActiveConversationId(
  activeConversationId: string | undefined,
  conversations: Conversation[],
) {
  if (
    activeConversationId &&
    conversations.some((conversation) => conversation.id === activeConversationId)
  ) {
    return activeConversationId;
  }

  return conversations[0]?.id ?? "";
}

function createInitialMessages(): ConversationMessage[] {
  const createdAt = new Date(Date.now() - 1000 * 60 * 10).toISOString();

  return [
    {
      id: createId("message"),
      role: "user",
      text: "帮我整理一张温柔一点的今日提醒，适合打印在小纸条上。",
      createdAt,
    },
    {
      id: createId("message"),
      role: "assistant",
      text: "当然可以。你可以写成：今天也别太赶，先把最重要的一件事做好，晚一点记得给自己买杯热饮。",
      createdAt: new Date(Date.now() - 1000 * 60 * 9).toISOString(),
    },
  ];
}

function createSeedState(): PersistedWorkspaceState {
  const now = Date.now();
  const primaryMessages = createInitialMessages();
  const conversationList: Conversation[] = [
    {
      id: "conv-today",
      title: "今日待办",
      preview: "下班前要记得买牛奶和胶带",
      updatedAt: new Date(now - 1000 * 60 * 2).toISOString(),
      draft: "",
      messages: primaryMessages,
    },
    {
      id: "conv-birthday",
      title: "生日祝福",
      preview: "想写一句温柔一点的话",
      updatedAt: new Date(now - 1000 * 60 * 10).toISOString(),
      draft: "",
      messages: [
        {
          id: createId("message"),
          role: "user",
          text: "想给朋友写一句生日祝福，语气轻一点。",
          createdAt: new Date(now - 1000 * 60 * 12).toISOString(),
        },
        {
          id: createId("message"),
          role: "assistant",
          text: "生日快乐，愿你这一岁也有被认真照顾、被温柔对待的日子。",
          createdAt: new Date(now - 1000 * 60 * 11).toISOString(),
        },
      ],
    },
    {
      id: "conv-shopping",
      title: "购物清单",
      preview: "鸡蛋、吐司、番茄、酸奶",
      updatedAt: new Date(now - 1000 * 60 * 60 * 18).toISOString(),
      draft: "记得补充家里常备的食物。",
      messages: [
        {
          id: createId("message"),
          role: "user",
          text: "帮我整理一个简洁一点的购物清单。",
          createdAt: new Date(now - 1000 * 60 * 60 * 18).toISOString(),
        },
        {
          id: createId("message"),
          role: "assistant",
          text: "鸡蛋、吐司、番茄、酸奶，先买这四样就够了。",
          createdAt: new Date(now - 1000 * 60 * 60 * 17).toISOString(),
        },
      ],
    },
  ];
  const devices: Device[] = [
    {
      id: "device-desk",
      name: "书桌咕咕机",
      status: "connected",
      note: "默认设备",
    },
    {
      id: "device-bedroom",
      name: "卧室咕咕机",
      status: "pending",
      note: "睡前提醒",
    },
  ];
  const schedules: Schedule[] = [
    {
      id: "schedule-morning",
      title: "早报摘要",
      source: "晨间订阅",
      timeLabel: "每天 08:00",
      deviceId: "device-desk",
      enabled: true,
    },
    {
      id: "schedule-night",
      title: "晚安提醒",
      source: "睡前便签",
      timeLabel: "每天 22:00",
      deviceId: "device-bedroom",
      enabled: true,
    },
    {
      id: "schedule-weekend",
      title: "周末清单",
      source: "家庭计划",
      timeLabel: "周六 09:30",
      deviceId: "device-desk",
      enabled: false,
    },
  ];
  const printJobs: PrintJob[] = [
    {
      id: "print-pending-message",
      title: "晚安留言",
      source: "对话草稿",
      deviceId: "device-bedroom",
      status: "pending",
      createdAt: new Date(now - 1000 * 60 * 30).toISOString(),
      updatedAt: new Date(now - 1000 * 60 * 30).toISOString(),
      content: "早点休息，今天已经做得很好了。",
    },
    {
      id: "print-queued-report",
      title: "明日早报",
      source: "晨间订阅",
      deviceId: "device-desk",
      status: "queued",
      createdAt: new Date(now - 1000 * 60 * 25).toISOString(),
      updatedAt: new Date(now - 1000 * 60 * 25).toISOString(),
      content: "明天上午天气晴，记得带水出门。",
    },
    {
      id: "print-done-todo",
      title: "今日待办",
      source: "手动打印",
      deviceId: "device-desk",
      status: "completed",
      createdAt: new Date(now - 1000 * 60 * 70).toISOString(),
      updatedAt: new Date(now - 1000 * 60 * 68).toISOString(),
      content: "先完成最重要的一件事。",
    },
    {
      id: "print-done-shopping",
      title: "购物清单",
      source: "手动打印",
      deviceId: "device-desk",
      status: "completed",
      createdAt: new Date(now - 1000 * 60 * 95).toISOString(),
      updatedAt: new Date(now - 1000 * 60 * 93).toISOString(),
      content: "鸡蛋、吐司、番茄、酸奶。",
    },
  ];
  const sources: SourceConnection[] = [
    {
      id: "source-worth",
      name: "今天值得看",
      type: "RSS",
      note: "每日文章摘要",
      status: "connected",
    },
    {
      id: "source-weather",
      name: "天气提醒",
      type: "在线服务",
      note: "晨间天气简报",
      status: "connected",
    },
    {
      id: "source-calendar",
      name: "家庭日历",
      type: "日历",
      note: "最近同步失败，请重新授权",
      status: "error",
    },
  ];
  const preferences: Preferences = {
    loginProtectionEnabled: false,
    sendConfirmationEnabled: true,
    theme: "light",
    defaultDeviceId: "device-desk",
  };
  const serviceBinding: ServiceBinding = {
    providerName: null,
    modelName: "Ink AI",
    bound: false,
  };

  return {
    authUser: null,
    devices,
    conversations: conversationList,
    activeConversationId: "conv-today",
    printJobs,
    schedules,
    sources,
    preferences,
    serviceBinding,
  };
}

function normalizeWorkspaceState(state: Partial<WorkspaceState>): WorkspaceState {
  const seed = createSeedState();
  const conversations =
    state.conversations === undefined
      ? seed.conversations
      : normalizeConversations(state.conversations);

  return {
    devices: state.devices ?? seed.devices,
    conversations,
    activeConversationId: resolveActiveConversationId(state.activeConversationId, conversations),
    printJobs: state.printJobs ?? seed.printJobs,
    schedules: state.schedules ?? seed.schedules,
    sources: state.sources ?? seed.sources,
    preferences: {
      loginProtectionEnabled:
        state.preferences?.loginProtectionEnabled ?? seed.preferences.loginProtectionEnabled,
      sendConfirmationEnabled:
        state.preferences?.sendConfirmationEnabled ?? seed.preferences.sendConfirmationEnabled,
      theme: normalizeThemeMode(state.preferences?.theme ?? seed.preferences.theme),
      defaultDeviceId: state.preferences?.defaultDeviceId ?? seed.preferences.defaultDeviceId,
    },
    serviceBinding: {
      providerName: state.serviceBinding?.providerName ?? seed.serviceBinding.providerName,
      modelName: state.serviceBinding?.modelName ?? seed.serviceBinding.modelName,
      bound: state.serviceBinding?.bound ?? seed.serviceBinding.bound,
    },
  };
}

function readPersistedWorkspaceState() {
  if (typeof window === "undefined") {
    return createSeedState();
  }

  const raw = window.localStorage.getItem(STORAGE_KEY);

  if (!raw) {
    return createSeedState();
  }

  try {
    const parsed = JSON.parse(raw) as Partial<PersistedWorkspaceState>;

    return {
      authUser: parsed.authUser ?? null,
      ...normalizeWorkspaceState(parsed),
    };
  } catch {
    return createSeedState();
  }
}

function isPersistedAuthSession(value: unknown): value is AuthSession {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Record<string, unknown>;
  return (
    typeof candidate.accessToken === "string" &&
    typeof candidate.refreshToken === "string" &&
    typeof candidate.accessTokenExpiresAt === "string"
  );
}

function readPersistedAuthSession() {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.sessionStorage.getItem(AUTH_SESSION_STORAGE_KEY);
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw);
    return isPersistedAuthSession(parsed) ? parsed : null;
  } catch {
    return null;
  }
}

function writePersistedAuthSession(session: AuthSession | null, persist: boolean) {
  if (typeof window === "undefined") {
    return;
  }

  if (!persist || !session) {
    window.sessionStorage.removeItem(AUTH_SESSION_STORAGE_KEY);
    return;
  }

  window.sessionStorage.setItem(AUTH_SESSION_STORAGE_KEY, JSON.stringify(session));
}

function sortPrintJobsByUpdatedAt(printJobs: PrintJob[]) {
  return printJobs.reduce<PrintJob[]>((sorted, job) => {
    const insertIndex = sorted.findIndex(
      (candidate) => new Date(candidate.updatedAt).getTime() < new Date(job.updatedAt).getTime(),
    );

    if (insertIndex === -1) {
      return [...sorted, job];
    }

    return [...sorted.slice(0, insertIndex), job, ...sorted.slice(insertIndex)];
  }, []);
}

function buildAIReplyMessages(messages: ConversationMessage[]) {
  return messages.map((message) => ({
    role: message.role,
    content: message.text,
  })) as { role: "user" | "assistant"; content: string }[];
}

function cloneRecord(input?: Record<string, unknown>) {
  return { ...input };
}

function mapPluginToSource(plugin: PluginDetails): SourceConnection {
  const note =
    plugin.binding?.lastError ||
    plugin.installation.lastError ||
    plugin.installation.description ||
    "可作为定时打印内容来源";

  return {
    id: plugin.installation.id,
    name: plugin.installation.displayName,
    type: plugin.installation.runtimeType === "node" ? "Node 插件" : "Python 插件",
    note,
    status:
      plugin.installation.status === "disabled" || !plugin.binding?.enabled
        ? "disconnected"
        : plugin.binding.status === "error"
          ? "error"
          : "connected",
  };
}

export const useWorkspaceStore = defineStore("workspace", () => {
  const persisted = readPersistedWorkspaceState();
  const persistedAuthSession = readPersistedAuthSession();

  const authUser = ref<User | null>(persistedAuthSession ? persisted.authUser : null);
  const authSession = ref<AuthSession | null>(persistedAuthSession);
  const authLoading = ref(false);
  const authError = ref("");
  const authBootstrapping = ref(false);
  const passwordChangeLoading = ref(false);
  const workspaceLoading = ref(false);
  const workspaceSyncing = ref(false);
  const workspaceSyncError = ref("");
  const accountCreationLoading = ref(false);
  const accountCreationError = ref("");
  const aiConfigLoading = ref(false);
  const aiConfigSaving = ref(false);
  const aiConfigError = ref("");
  const printerSyncError = ref("");
  const pluginLoading = ref(false);
  const pluginSaving = ref(false);
  const pluginUploadLoading = ref(false);
  const pluginError = ref("");
  const pluginActionError = ref("");
  const pluginTestingId = ref("");
  const pluginSavingId = ref("");
  const pluginUploadingName = ref("");
  const workspaceOwnerId = ref<string | null>(null);
  const workspaceHydrating = ref(false);

  const devices = ref<Device[]>(persisted.devices);
  const conversations = ref<Conversation[]>(persisted.conversations);
  const activeConversationId = ref(persisted.activeConversationId);
  const printJobs = ref<PrintJob[]>(persisted.printJobs);
  const schedules = ref<Schedule[]>(persisted.schedules);
  const sources = ref<SourceConnection[]>(persisted.sources);
  const availablePlugins = ref<PluginDetails[]>([]);
  const adminPlugins = ref<PluginDetails[]>([]);
  const remoteSchedules = ref<PrintScheduleView[]>([]);

  const loginProtectionEnabled = ref(persisted.preferences.loginProtectionEnabled);
  const sendConfirmationEnabled = ref(persisted.preferences.sendConfirmationEnabled);
  const selectedTheme = ref<ThemeMode>(persisted.preferences.theme);
  const defaultDeviceId = ref(persisted.preferences.defaultDeviceId);

  const serviceBinding = ref<ServiceBinding>(persisted.serviceBinding);
  const aiConfigSummary = ref<AIConfigSummary>({
    bound: persisted.serviceBinding.bound,
    providerName: persisted.serviceBinding.providerName ?? "OpenAI Compatible",
    providerType: "openai-compatible",
    baseUrl: "",
    model: persisted.serviceBinding.modelName,
    keyConfigured: false,
  });

  const isGenerating = ref(false);
  const generationError = ref("");
  const selectedConversationMessageIds = ref<string[]>([]);
  const isCreatingPrint = ref(false);
  const flashMessage = ref("");
  const flashTone = ref<"success" | "error" | "info">("info");
  let flashTimer = 0;
  let remoteSaveTimer = 0;
  let remoteSavePromise: Promise<boolean> | null = null;
  let remotePrintStatusTimer = 0;
  let remotePrintStatusPromise: Promise<void> | null = null;

  const deviceMap = computed(() =>
    devices.value.reduce<Record<string, Device>>((accumulator, device) => {
      accumulator[device.id] = device;
      return accumulator;
    }, {}),
  );

  const activeConversation = computed(
    () =>
      conversations.value.find((conversation) => conversation.id === activeConversationId.value) ??
      conversations.value[0] ??
      null,
  );
  const conversationMessages = computed(() => activeConversation.value?.messages ?? []);
  const selectedConversationMessages = computed(() =>
    conversationMessages.value.filter((message) =>
      selectedConversationMessageIds.value.includes(message.id),
    ),
  );
  const defaultDevice = computed(
    () => devices.value.find((device) => device.id === defaultDeviceId.value) ?? null,
  );
  const sortedPrintJobs = computed(() => sortPrintJobsByUpdatedAt(printJobs.value));
  const pendingPrintJobs = computed(() =>
    sortedPrintJobs.value.filter((job) => job.status === "pending" || job.status === "queued"),
  );
  const recentPrintJobs = computed(() => sortedPrintJobs.value.slice(0, 5));
  const activeSchedules = computed(() =>
    isAuthenticated.value
      ? remoteSchedules.value.map((schedule) => ({
          id: schedule.id,
          title: schedule.title,
          source: schedule.sourceLabel,
          timeLabel: schedule.timeLabel,
          deviceId: schedule.deviceId,
          enabled: schedule.enabled,
          pluginInstallationId: schedule.pluginInstallationId,
          frequencyType: schedule.frequencyType,
          timezone: schedule.timezone,
          hour: schedule.hour,
          minute: schedule.minute,
          weekdays: schedule.weekdays,
          scheduleConfig: cloneRecord(schedule.scheduleConfig),
          pluginDisplayName: schedule.pluginDisplayName,
          nextRunAt: schedule.nextRunAt,
          lastRunAt: schedule.lastRunAt,
          lastError: schedule.lastError || "",
        }))
      : schedules.value.map((schedule) => ({
          ...schedule,
          pluginInstallationId: "",
          frequencyType: "daily",
          timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || "Asia/Shanghai",
          hour: 19,
          minute: 30,
          weekdays: [] as number[],
          scheduleConfig: {},
          pluginDisplayName: schedule.source,
          nextRunAt: undefined,
          lastRunAt: undefined,
          lastError: "",
        })),
  );
  const activeSources = computed(() =>
    isAuthenticated.value ? availablePlugins.value.map(mapPluginToSource) : sources.value,
  );
  const connectedDevicesCount = computed(
    () => devices.value.filter((device) => device.status === "connected").length,
  );
  const enabledSchedulesCount = computed(
    () => activeSchedules.value.filter((schedule) => schedule.enabled).length,
  );
  const pendingConfirmationCount = computed(
    () => printJobs.value.filter((job) => job.status === "pending").length,
  );
  const todayCompletedCount = computed(() => {
    const today = new Date();

    return printJobs.value.filter(
      (job) =>
        job.status === "completed" &&
        new Date(job.updatedAt).toDateString() === today.toDateString(),
    ).length;
  });

  const activeDeviceLabel = computed(() => defaultDevice.value?.name ?? "");
  const activeModelLabel = computed(
    () => aiConfigSummary.value.model || serviceBinding.value.modelName,
  );
  const todayPrintCount = computed(() => todayCompletedCount.value);
  const welcomeLabel = computed(() => "整理内容，准备打印");
  const isConfigured = computed(() => activeDeviceLabel.value !== "");
  const isAuthenticated = computed(() => authUser.value !== null && authSession.value !== null);
  const isAdmin = computed(() => authUser.value?.role === "admin");
  const hasQueuedRemotePrintJobs = computed(
    () => isAuthenticated.value && printJobs.value.some((job) => job.status === "queued"),
  );

  const summaryCards = computed(() => [
    {
      label: "已绑定设备",
      value: `${devices.value.length} 台`,
      tone: "neutral",
      progress: devices.value.length > 0 ? Math.min(100, connectedDevicesCount.value * 50) : 10,
    },
    {
      label: "已启用任务",
      value: `${enabledSchedulesCount.value} 条`,
      tone: "amber",
      progress: activeSchedules.value.length
        ? Math.round((enabledSchedulesCount.value / activeSchedules.value.length) * 100)
        : 0,
    },
    {
      label: "待确认打印",
      value: `${pendingConfirmationCount.value} 条`,
      tone: "stone",
      progress: pendingConfirmationCount.value
        ? Math.min(100, pendingConfirmationCount.value * 25)
        : 8,
    },
    {
      label: "今日完成",
      value: `${todayCompletedCount.value} 条`,
      tone: "green",
      progress: todayCompletedCount.value ? Math.min(100, todayCompletedCount.value * 20) : 10,
    },
  ]);

  const workspaceState = computed<WorkspaceState>(() => ({
    devices: devices.value,
    conversations: conversations.value,
    activeConversationId: activeConversationId.value,
    printJobs: printJobs.value,
    schedules: schedules.value,
    sources: sources.value,
    preferences: {
      loginProtectionEnabled: loginProtectionEnabled.value,
      sendConfirmationEnabled: sendConfirmationEnabled.value,
      theme: selectedTheme.value,
      defaultDeviceId: defaultDeviceId.value,
    },
    serviceBinding: serviceBinding.value,
  }));

  const persistableState = computed<PersistedWorkspaceState>(() => ({
    authUser: loginProtectionEnabled.value || !authSession.value ? null : authUser.value,
    ...workspaceState.value,
  }));

  watch(
    persistableState,
    (value) => {
      if (typeof window === "undefined" || authSession.value) {
        return;
      }

      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
    },
    { deep: true, immediate: true },
  );

  watch(
    workspaceState,
    () => {
      if (
        !authSession.value ||
        !authUser.value ||
        workspaceOwnerId.value !== authUser.value.id ||
        workspaceHydrating.value
      ) {
        return;
      }

      if (remoteSaveTimer) {
        window.clearTimeout(remoteSaveTimer);
      }

      remoteSaveTimer = window.setTimeout(() => {
        remoteSaveTimer = 0;
        void persistRemoteWorkspace();
      }, REMOTE_SAVE_DEBOUNCE_MS);
    },
    { deep: true },
  );

  watch(
    [authSession, loginProtectionEnabled],
    ([session, loginProtection]) => {
      writePersistedAuthSession(session, !loginProtection);
    },
    { deep: true, immediate: true },
  );

  watch(
    hasQueuedRemotePrintJobs,
    (hasQueuedJobs) => {
      if (!hasQueuedJobs) {
        clearRemotePrintStatusSync();
        return;
      }

      scheduleRemotePrintStatusSync(true);
    },
    { immediate: true },
  );

  function applyWorkspaceState(nextState: WorkspaceState) {
    const normalized = normalizeWorkspaceState(nextState);
    devices.value = normalized.devices;
    conversations.value = normalized.conversations;
    activeConversationId.value = normalized.activeConversationId;
    printJobs.value = normalized.printJobs;
    schedules.value = normalized.schedules;
    sources.value = normalized.sources;
    loginProtectionEnabled.value = normalized.preferences.loginProtectionEnabled;
    sendConfirmationEnabled.value = normalized.preferences.sendConfirmationEnabled;
    selectedTheme.value = normalized.preferences.theme;
    defaultDeviceId.value = normalized.preferences.defaultDeviceId;
    serviceBinding.value = normalized.serviceBinding;
    selectedConversationMessageIds.value = [];
    generationError.value = "";
  }

  function applyAIConfig(nextConfig: AIConfigSummary) {
    aiConfigSummary.value = nextConfig;
    serviceBinding.value = {
      providerName: nextConfig.bound ? nextConfig.providerName : null,
      modelName: nextConfig.model || "Ink AI",
      bound: nextConfig.bound,
    };
    aiConfigError.value = "";
  }

  function upsertPrintJob(nextJob: PrintJob) {
    printJobs.value = [nextJob, ...printJobs.value.filter((job) => job.id !== nextJob.id)];
  }

  function upsertDevice(nextDevice: Device) {
    devices.value = [nextDevice, ...devices.value.filter((device) => device.id !== nextDevice.id)];
  }

  function clearRemotePrintStatusSync() {
    if (typeof window === "undefined") {
      return;
    }

    if (remotePrintStatusTimer) {
      window.clearTimeout(remotePrintStatusTimer);
      remotePrintStatusTimer = 0;
    }
  }

  function scheduleRemotePrintStatusSync(immediate = false) {
    if (typeof window === "undefined") {
      return;
    }

    clearRemotePrintStatusSync();

    if (!hasQueuedRemotePrintJobs.value) {
      return;
    }

    remotePrintStatusTimer = window.setTimeout(() => {
      remotePrintStatusTimer = 0;
      void syncRemotePrintStatus();
    }, immediate ? REMOTE_PRINT_STATUS_INITIAL_POLL_MS : REMOTE_PRINT_STATUS_POLL_MS);
  }

  async function syncRemotePrintStatus() {
    if (remotePrintStatusPromise) {
      return remotePrintStatusPromise;
    }

    const currentSession = authSession.value;
    if (!currentSession || !hasQueuedRemotePrintJobs.value) {
      clearRemotePrintStatusSync();
      return;
    }

    remotePrintStatusPromise = (async () => {
      try {
        const accessToken = currentSession.accessToken;
        const { printJobs: latestPrintJobs } = await fetchPrintJobs(accessToken);

        if (authSession.value?.accessToken !== accessToken) {
          return;
        }

        printJobs.value = latestPrintJobs;
        printerSyncError.value = "";
      } catch (error) {
        printerSyncError.value =
          error instanceof Error ? error.message : "同步打印状态失败，请稍后重试。";
      } finally {
        remotePrintStatusPromise = null;

        if (hasQueuedRemotePrintJobs.value) {
          scheduleRemotePrintStatusSync();
        }
      }
    })();

    return remotePrintStatusPromise;
  }

  function restoreAnonymousWorkspace() {
    workspaceOwnerId.value = null;
    clearRemotePrintStatusSync();
    workspaceHydrating.value = true;
    applyWorkspaceState(readPersistedWorkspaceState());
    availablePlugins.value = [];
    adminPlugins.value = [];
    remoteSchedules.value = [];
    pluginError.value = "";
    pluginActionError.value = "";
    applyAIConfig({
      bound: false,
      providerName: "OpenAI Compatible",
      providerType: "openai-compatible",
      baseUrl: "",
      model: "Ink AI",
      keyConfigured: false,
    });
    workspaceHydrating.value = false;
    workspaceSyncError.value = "";
    printerSyncError.value = "";
  }

  async function persistRemoteWorkspace() {
    if (remoteSavePromise) {
      return remoteSavePromise;
    }

    const currentSession = authSession.value;
    const currentUser = authUser.value;

    if (!currentSession || !currentUser || workspaceOwnerId.value !== currentUser.id) {
      return true;
    }

    workspaceSyncing.value = true;
    workspaceSyncError.value = "";
    remoteSavePromise = (async () => {
      try {
        await saveWorkspaceStateWithApi(currentSession.accessToken, workspaceState.value);
        return true;
      } catch (error) {
        workspaceSyncError.value =
          error instanceof Error ? error.message : "同步数据失败，请稍后重试。";
        return false;
      } finally {
        workspaceSyncing.value = false;
        remoteSavePromise = null;
      }
    })();

    return remoteSavePromise;
  }

  async function flushRemoteWorkspaceSave() {
    if (remoteSaveTimer) {
      window.clearTimeout(remoteSaveTimer);
      remoteSaveTimer = 0;
      return persistRemoteWorkspace();
    }

    return remoteSavePromise ?? true;
  }

  async function loadRemoteWorkspace() {
    const currentSession = authSession.value;
    const currentUser = authUser.value;

    if (!currentSession || !currentUser) {
      return false;
    }

    workspaceLoading.value = true;
    workspaceSyncError.value = "";
    workspaceHydrating.value = true;

    try {
      const state = await fetchWorkspaceStateWithApi(currentSession.accessToken);
      applyWorkspaceState(state);
      await loadLiveIntegrations(currentSession.accessToken);
      workspaceOwnerId.value = currentUser.id;
      return true;
    } catch (error) {
      workspaceSyncError.value =
        error instanceof Error ? error.message : "加载账号数据失败，请稍后重试。";
      return false;
    } finally {
      workspaceHydrating.value = false;
      workspaceLoading.value = false;
    }
  }

  async function loadLiveIntegrations(accessToken: string) {
    aiConfigLoading.value = true;
    pluginLoading.value = true;
    printerSyncError.value = "";
    pluginError.value = "";

    try {
      const [aiSummary, printerResponse, printJobResponse, pluginResponse, scheduleResponse] =
        await Promise.all([
          fetchAIConfigSummary(accessToken),
          fetchPrinters(accessToken),
          fetchPrintJobs(accessToken),
          fetchPlugins(accessToken),
          fetchPrintSchedules(accessToken),
        ]);

      applyAIConfig(aiSummary);
      devices.value = printerResponse.devices.filter((device) => device.status !== "offline");
      printJobs.value = printJobResponse.printJobs;
      availablePlugins.value = pluginResponse.plugins;
      remoteSchedules.value = scheduleResponse.schedules;

      if (authUser.value?.role === "admin") {
        const adminResponse = await fetchAdminPlugins(accessToken);
        adminPlugins.value = adminResponse.plugins;
      } else {
        adminPlugins.value = [];
      }
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "加载插件、设备与 AI 配置失败，请稍后重试。";
      aiConfigError.value = message;
      printerSyncError.value = message;
      pluginError.value = message;
    } finally {
      aiConfigLoading.value = false;
      pluginLoading.value = false;
    }
  }

  function showFlash(message: string, tone: "success" | "error" | "info" = "info") {
    flashMessage.value = message;
    flashTone.value = tone;

    if (flashTimer) {
      window.clearTimeout(flashTimer);
    }

    if (typeof window !== "undefined") {
      flashTimer = window.setTimeout(() => {
        flashMessage.value = "";
      }, 2600);
    }
  }

  function updateConversation(
    conversationId: string,
    updater: (conversation: Conversation) => Conversation,
  ) {
    conversations.value = conversations.value.map((conversation) =>
      conversation.id === conversationId ? updater(conversation) : conversation,
    );
  }

  function touchConversation(conversationId: string) {
    const conversation = conversations.value.find((item) => item.id === conversationId);

    if (!conversation) {
      return;
    }

    conversations.value = [
      {
        ...conversation,
        updatedAt: getNow(),
      },
      ...conversations.value.filter((item) => item.id !== conversationId),
    ];
  }

  function selectConversation(conversationId: string) {
    activeConversationId.value = conversationId;
    generationError.value = "";
    selectedConversationMessageIds.value = [];
  }

  function ensureActiveConversation() {
    if (activeConversation.value) {
      return activeConversation.value;
    }

    const existingConversation = conversations.value[0];
    if (existingConversation) {
      selectConversation(existingConversation.id);
      return existingConversation;
    }

    const conversation = createEmptyConversation();
    conversations.value = [conversation];
    selectConversation(conversation.id);
    return conversation;
  }

  function createConversation() {
    const conversation = createEmptyConversation();

    conversations.value = [conversation, ...conversations.value];
    selectConversation(conversation.id);
    showFlash("已创建新的对话草稿。");
  }

  function deleteConversation(conversationId: string) {
    const remaining = conversations.value.filter(
      (conversation) => conversation.id !== conversationId,
    );

    if (remaining.length === conversations.value.length) {
      return false;
    }

    if (remaining.length === 0) {
      const replacement = createEmptyConversation();

      conversations.value = [replacement];
      selectConversation(replacement.id);
      showFlash("对话已删除。", "success");
      return true;
    }

    conversations.value = remaining;

    if (activeConversationId.value === conversationId) {
      activeConversationId.value = remaining[0].id;
    }

    selectedConversationMessageIds.value = [];
    generationError.value = "";
    showFlash("对话已删除。", "success");
    return true;
  }

  function updateCurrentDraft(value: string) {
    const currentConversation = ensureActiveConversation();

    updateConversation(currentConversation.id, (conversation) => ({
      ...conversation,
      draft: value,
      updatedAt: getNow(),
    }));
  }

  function saveCurrentDraft() {
    ensureActiveConversation();

    showFlash("草稿已保存。", "success");
  }

  async function sendCurrentDraft() {
    const conversation = ensureActiveConversation();

    if (isGenerating.value) {
      return false;
    }

    const prompt = conversation.draft.trim();

    if (!prompt) {
      generationError.value = "请先输入要整理的内容。";
      return false;
    }

    const userMessage: ConversationMessage = {
      id: createId("message"),
      role: "user",
      text: prompt,
      createdAt: getNow(),
    };

    generationError.value = "";
    isGenerating.value = true;

    updateConversation(conversation.id, (current) => ({
      ...current,
      title: current.messages.length === 0 ? prompt.slice(0, 8) || "新对话" : current.title,
      preview: prompt.slice(0, 22),
      updatedAt: getNow(),
      draft: "",
      messages: [...current.messages, userMessage],
    }));

    touchConversation(conversation.id);

    try {
      const reply =
        isAuthenticated.value && authSession.value
          ? (
              await generateAIReply(authSession.value.accessToken, {
                messages: buildAIReplyMessages([...conversation.messages, userMessage]),
              })
            ).content
          : await generateReplyWithMockService({ prompt });
      const assistantMessage: ConversationMessage = {
        id: createId("message"),
        role: "assistant",
        text: reply,
        createdAt: getNow(),
      };

      updateConversation(conversation.id, (current) => ({
        ...current,
        preview: reply.slice(0, 22),
        updatedAt: getNow(),
        messages: [...current.messages, assistantMessage],
      }));
      touchConversation(conversation.id);
      showFlash("新回复已生成。", "success");
      return true;
    } catch (error) {
      generationError.value =
        error instanceof Error ? error.message : "暂时没能生成回复，请稍后重试。";
      showFlash(generationError.value, "error");
      return false;
    } finally {
      isGenerating.value = false;
    }
  }

  async function regenerateLatestReply() {
    const conversation = activeConversation.value;
    let latestUserMessage: ConversationMessage | undefined;

    for (let index = (conversation?.messages.length ?? 0) - 1; index >= 0; index -= 1) {
      const candidate = conversation?.messages[index];

      if (candidate?.role === "user") {
        latestUserMessage = candidate;
        break;
      }
    }

    if (!conversation || !latestUserMessage || isGenerating.value) {
      generationError.value = "当前没有可重新生成的内容。";
      return false;
    }

    isGenerating.value = true;
    generationError.value = "";

    updateConversation(conversation.id, (current) => ({
      ...current,
      messages:
        current.messages.at(-1)?.role === "assistant"
          ? current.messages.slice(0, -1)
          : current.messages,
      updatedAt: getNow(),
    }));

    try {
      const refreshedConversation =
        conversations.value.find((item) => item.id === conversation.id) ?? conversation;
      const reply =
        isAuthenticated.value && authSession.value
          ? (
              await generateAIReply(authSession.value.accessToken, {
                messages: buildAIReplyMessages(refreshedConversation.messages),
              })
            ).content
          : await generateReplyWithMockService({ prompt: latestUserMessage.text });
      const assistantMessage: ConversationMessage = {
        id: createId("message"),
        role: "assistant",
        text: reply,
        createdAt: getNow(),
      };

      updateConversation(conversation.id, (current) => ({
        ...current,
        preview: reply.slice(0, 22),
        updatedAt: getNow(),
        messages: [...current.messages, assistantMessage],
      }));
      touchConversation(conversation.id);
      showFlash("已经重新生成一版回复。", "success");
      return true;
    } catch (error) {
      generationError.value = error instanceof Error ? error.message : "重新生成失败，请稍后再试。";
      showFlash(generationError.value, "error");
      return false;
    } finally {
      isGenerating.value = false;
    }
  }

  function buildPrintJob(title: string, content: string, source: string): PrintJob {
    const now = getNow();

    return {
      id: createId("print"),
      title,
      source,
      deviceId: defaultDeviceId.value,
      status: sendConfirmationEnabled.value ? "pending" : "queued",
      createdAt: now,
      updatedAt: now,
      content,
    };
  }

  async function maybeCompleteQueuedJob(jobId: string) {
    window.setTimeout(() => {
      const target = printJobs.value.find((job) => job.id === jobId);

      if (!target || target.status !== "queued") {
        return;
      }

      printJobs.value = printJobs.value.map((job) =>
        job.id === jobId
          ? {
              ...job,
              status: "completed",
              updatedAt: getNow(),
            }
          : job,
      );
    }, 500);
  }

  async function addPrintJob(title: string, content: string, source: string) {
    if (isCreatingPrint.value) {
      return null;
    }

    isCreatingPrint.value = true;
    try {
      if (isAuthenticated.value && authSession.value) {
        if (!defaultDeviceId.value) {
          showFlash("请先绑定咕咕机并设为默认设备。", "error");
          return null;
        }

        const job = await createPrintJob(authSession.value.accessToken, {
          title,
          source,
          content,
          printerBindingId: defaultDeviceId.value,
          submitImmediately: !sendConfirmationEnabled.value,
        });
        upsertPrintJob(job);
        showFlash(
          sendConfirmationEnabled.value ? "已加入待确认打印。" : "内容已直接加入打印队列。",
          "success",
        );
        return job;
      }

      const job = buildPrintJob(title, content, source);
      printJobs.value = [job, ...printJobs.value];

      if (job.status === "queued") {
        showFlash("内容已直接加入打印队列。", "success");
        await maybeCompleteQueuedJob(job.id);
      } else {
        showFlash("已加入待确认打印。", "success");
      }

      return job;
    } catch (error) {
      const message = error instanceof Error ? error.message : "创建打印任务失败，请稍后重试。";
      showFlash(message, "error");
      return null;
    } finally {
      isCreatingPrint.value = false;
    }
  }

  async function createPrintFromSelectedMessages() {
    if (selectedConversationMessages.value.length === 0) {
      showFlash("请先选中至少一条消息。", "error");
      return null;
    }

    const content = selectedConversationMessages.value
      .map((message) => `${message.role === "user" ? "我" : "Ink"}：${message.text}`)
      .join("\n\n");

    return addPrintJob(activeConversation.value?.title ?? "选中问答", content, "对话选中问答");
  }

  async function createPrintFromConversation() {
    const conversation = activeConversation.value;

    if (!conversation || conversation.messages.length === 0) {
      showFlash("当前对话还没有内容可打印。", "error");
      return null;
    }

    const content = conversation.messages
      .map((message) => `${message.role === "user" ? "我" : "Ink"}：${message.text}`)
      .join("\n\n");
    return addPrintJob(conversation.title, content, "当前对话");
  }

  async function createManualPrint(options?: { title?: string; content?: string }) {
    return addPrintJob(
      options?.title?.trim() || "手动新建纸条",
      options?.content?.trim() || "新的打印内容已创建，你可以稍后继续编辑。",
      "手动打印",
    );
  }

  async function confirmPrint(jobId: string) {
    const target = printJobs.value.find((job) => job.id === jobId);

    if (!target || target.status !== "pending") {
      return false;
    }

    if (isAuthenticated.value && authSession.value) {
      try {
        const submitted = await submitPrintJob(authSession.value.accessToken, jobId);
        upsertPrintJob(submitted);
        showFlash("已加入打印队列。", "success");
        return true;
      } catch (error) {
        showFlash(error instanceof Error ? error.message : "提交打印失败，请稍后重试。", "error");
        return false;
      }
    }

    printJobs.value = printJobs.value.map((job) =>
      job.id === jobId
        ? {
            ...job,
            status: "queued",
            updatedAt: getNow(),
          }
        : job,
    );
    showFlash("已加入打印队列。", "success");
    await maybeCompleteQueuedJob(jobId);
    return true;
  }

  async function cancelPrint(jobId: string) {
    const target = printJobs.value.find((job) => job.id === jobId);

    if (!target || (target.status !== "pending" && target.status !== "queued")) {
      return false;
    }

    if (isAuthenticated.value && authSession.value) {
      try {
        const cancelled = await cancelPrintJob(authSession.value.accessToken, jobId);
        upsertPrintJob(cancelled);
        showFlash("已取消打印。", "success");
        return true;
      } catch (error) {
        showFlash(error instanceof Error ? error.message : "取消打印失败，请稍后重试。", "error");
        return false;
      }
    }

    printJobs.value = printJobs.value.map((job) =>
      job.id === jobId
        ? {
            ...job,
            status: "cancelled",
            updatedAt: getNow(),
          }
        : job,
    );
    showFlash("已取消打印。", "success");
    return true;
  }

  async function updatePrintDevice(jobId: string, deviceId: string) {
    if (isAuthenticated.value && authSession.value) {
      try {
        const updated = await updatePrintJobDeviceWithApi(authSession.value.accessToken, jobId, {
          printerBindingId: deviceId,
        });
        upsertPrintJob(updated);
        showFlash("打印目标设备已更新。", "success");
        return;
      } catch (error) {
        showFlash(
          error instanceof Error ? error.message : "更新打印设备失败，请稍后重试。",
          "error",
        );
        return;
      }
    }

    printJobs.value = printJobs.value.map((job) =>
      job.id === jobId
        ? {
            ...job,
            deviceId,
            updatedAt: getNow(),
          }
        : job,
    );
    showFlash("打印目标设备已更新。", "success");
  }

  async function addDevice(options?: {
    name?: string;
    note?: string;
    deviceId?: string;
    setAsDefault?: boolean;
  }) {
    if (isAuthenticated.value && authSession.value) {
      try {
        const deviceIdentifier = options?.deviceId?.trim() ?? "";
        if (!deviceIdentifier) {
          showFlash("请填写咕咕机设备编号。", "error");
          return null;
        }

        const device = await bindPrinter(authSession.value.accessToken, {
          name: options?.name?.trim() || `咕咕机 ${devices.value.length + 1}`,
          note: options?.note?.trim() || "",
          deviceId: deviceIdentifier,
        });
        upsertDevice(device);
        if (options?.setAsDefault || !defaultDeviceId.value) {
          defaultDeviceId.value = device.id;
        }
        showFlash("设备已绑定。", "success");
        return device;
      } catch (error) {
        showFlash(error instanceof Error ? error.message : "绑定设备失败，请稍后重试。", "error");
        return null;
      }
    }

    const nextIndex = devices.value.length + 1;
    const device: Device = {
      id: createId("device"),
      name: options?.name?.trim() || `咕咕机 ${nextIndex}`,
      status: "pending",
      note: options?.note?.trim() || "等待绑定",
    };

    devices.value = [...devices.value, device];
    if (options?.setAsDefault) {
      defaultDeviceId.value = device.id;
    }
    showFlash("已添加新设备。", "success");
    return device;
  }

  async function removeDevice(deviceId: string) {
    const target = devices.value.find((device) => device.id === deviceId);

    if (!target) {
      return false;
    }

    if (isAuthenticated.value && authSession.value) {
      try {
        await deletePrinter(authSession.value.accessToken, deviceId);
      } catch (error) {
        showFlash(error instanceof Error ? error.message : "删除设备失败，请稍后重试。", "error");
        return false;
      }

      const remainingDevices = devices.value.filter((device) => device.id !== deviceId);
      const fallbackDeviceId =
        defaultDeviceId.value === deviceId
          ? (remainingDevices.find((device) => device.status !== "offline")?.id ?? "")
          : defaultDeviceId.value;

      devices.value = remainingDevices;
      printJobs.value = printJobs.value.filter((job) => job.deviceId !== deviceId);
      remoteSchedules.value = remoteSchedules.value.filter(
        (schedule) => schedule.deviceId !== deviceId,
      );
      defaultDeviceId.value = fallbackDeviceId;
      showFlash("设备已删除。", "success");
      return true;
    }

    const remainingDevices = devices.value.filter((device) => device.id !== deviceId);
    const fallbackDeviceId =
      defaultDeviceId.value === deviceId ? (remainingDevices[0]?.id ?? "") : defaultDeviceId.value;

    devices.value = remainingDevices;
    defaultDeviceId.value = fallbackDeviceId;
    printJobs.value = printJobs.value.map((job) =>
      job.deviceId === deviceId
        ? {
            ...job,
            deviceId: fallbackDeviceId,
            updatedAt: getNow(),
          }
        : job,
    );
    schedules.value = schedules.value.map((schedule) =>
      schedule.deviceId === deviceId
        ? {
            ...schedule,
            deviceId: fallbackDeviceId,
          }
        : schedule,
    );
    showFlash(target.status === "pending" ? "已移除设备。" : "已删除设备。", "success");
    return true;
  }

  function upsertRemoteSchedule(nextSchedule: PrintScheduleView) {
    remoteSchedules.value = [
      nextSchedule,
      ...remoteSchedules.value.filter((schedule) => schedule.id !== nextSchedule.id),
    ];
  }

  function upsertPlugin(nextPlugin: PluginDetails) {
    availablePlugins.value = [
      nextPlugin,
      ...availablePlugins.value.filter(
        (plugin) => plugin.installation.id !== nextPlugin.installation.id,
      ),
    ];
    if (isAdmin.value) {
      adminPlugins.value = [
        nextPlugin,
        ...adminPlugins.value.filter(
          (plugin) => plugin.installation.id !== nextPlugin.installation.id,
        ),
      ];
    }
  }

  async function toggleSchedule(scheduleId: string) {
    if (isAuthenticated.value && authSession.value) {
      try {
        const updated = await togglePrintSchedule(authSession.value.accessToken, scheduleId);
        upsertRemoteSchedule(updated);
        showFlash("定时任务状态已更新。", "success");
        return true;
      } catch (error) {
        showFlash(
          error instanceof Error ? error.message : "更新定时任务失败，请稍后重试。",
          "error",
        );
        return false;
      }
    }

    schedules.value = schedules.value.map((schedule) =>
      schedule.id === scheduleId ? { ...schedule, enabled: !schedule.enabled } : schedule,
    );
    showFlash("定时任务状态已更新。", "success");
    return true;
  }

  async function createSchedule(options?: {
    title?: string;
    source?: string;
    deviceId?: string;
    pluginInstallationId?: string;
    frequencyType?: "daily" | "weekly";
    timezone?: string;
    hour?: number;
    minute?: number;
    weekdays?: number[];
    scheduleConfig?: Record<string, unknown>;
  }) {
    const nextDeviceId = options?.deviceId || defaultDeviceId.value;
    const targetDevice = devices.value.find((device) => device.id === nextDeviceId);

    if (!targetDevice || targetDevice.status === "offline") {
      showFlash("请先选择可用设备。", "error");
      return null;
    }

    if (isAuthenticated.value && authSession.value) {
      const pluginInstallationId = options?.pluginInstallationId?.trim() ?? "";
      if (!pluginInstallationId) {
        showFlash("请选择插件来源。", "error");
        return null;
      }

      try {
        const schedule = await createPrintSchedule(authSession.value.accessToken, {
          title: options?.title?.trim() || "新的定时任务",
          pluginInstallationId,
          frequencyType: options?.frequencyType || "daily",
          timezone: options?.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone,
          hour: options?.hour ?? 19,
          minute: options?.minute ?? 30,
          weekdays: options?.weekdays ?? [],
          scheduleConfig: cloneRecord(options?.scheduleConfig),
          deviceId: nextDeviceId,
          enabled: true,
        });
        upsertRemoteSchedule(schedule);
        showFlash("已创建新的定时任务。", "success");
        return schedule;
      } catch (error) {
        showFlash(
          error instanceof Error ? error.message : "创建定时任务失败，请稍后重试。",
          "error",
        );
        return null;
      }
    }

    const schedule: Schedule = {
      id: createId("schedule"),
      title: options?.title?.trim() || "新的定时任务",
      source: options?.source?.trim() || "手动创建",
      timeLabel: "每天 19:30",
      deviceId: nextDeviceId,
      enabled: true,
    };

    schedules.value = [schedule, ...schedules.value];
    showFlash("已创建新的定时任务。", "success");
    return schedule;
  }

  async function updateScheduleDevice(scheduleId: string, deviceId: string) {
    const targetDevice = devices.value.find((device) => device.id === deviceId);

    if (!targetDevice || targetDevice.status === "offline") {
      showFlash("请先选择可用设备。", "error");
      return;
    }

    if (isAuthenticated.value && authSession.value) {
      const current = remoteSchedules.value.find((schedule) => schedule.id === scheduleId);
      if (!current) {
        return;
      }

      try {
        const updated = await updatePrintSchedule(authSession.value.accessToken, scheduleId, {
          title: current.title,
          pluginInstallationId: current.pluginInstallationId,
          frequencyType: current.frequencyType,
          timezone: current.timezone,
          hour: current.hour,
          minute: current.minute,
          weekdays: current.weekdays,
          scheduleConfig: cloneRecord(current.scheduleConfig),
          deviceId,
          enabled: current.enabled,
        });
        upsertRemoteSchedule(updated);
        showFlash("定时任务设备已更新。", "success");
        return;
      } catch (error) {
        showFlash(
          error instanceof Error ? error.message : "更新定时任务失败，请稍后重试。",
          "error",
        );
        return;
      }
    }

    schedules.value = schedules.value.map((schedule) =>
      schedule.id === scheduleId ? { ...schedule, deviceId } : schedule,
    );
    showFlash("定时任务设备已更新。", "success");
  }

  async function deleteSchedule(scheduleId: string) {
    if (isAuthenticated.value && authSession.value) {
      try {
        await deletePrintSchedule(authSession.value.accessToken, scheduleId);
        remoteSchedules.value = remoteSchedules.value.filter(
          (schedule) => schedule.id !== scheduleId,
        );
        showFlash("定时任务已删除。", "success");
        return true;
      } catch (error) {
        showFlash(
          error instanceof Error ? error.message : "删除定时任务失败，请稍后重试。",
          "error",
        );
        return false;
      }
    }

    schedules.value = schedules.value.filter((schedule) => schedule.id !== scheduleId);
    showFlash("定时任务已删除。", "success");
    return true;
  }

  function toggleSourceConnection(sourceId: string) {
    const target = sources.value.find((source) => source.id === sourceId);

    if (!target) {
      return false;
    }

    const nextStatus = target.status === "connected" ? "disconnected" : "connected";
    sources.value = sources.value.map((source) =>
      source.id === sourceId
        ? {
            ...source,
            status: nextStatus,
            note: nextStatus === "connected" ? "连接正常" : "尚未连接到此来源",
          }
        : source,
    );
    showFlash(nextStatus === "connected" ? "插件已连接。" : "插件已解绑。", "success");
    return true;
  }

  async function uploadPlugin(file: File) {
    const current = authSession.value;

    if (!current) {
      pluginActionError.value = "登录状态已失效，请重新登录。";
      return null;
    }

    pluginUploadLoading.value = true;
    pluginActionError.value = "";
    pluginUploadingName.value = file.name;

    try {
      const uploaded = await uploadPluginZip(current.accessToken, file);
      upsertPlugin(uploaded);
      if (isAdmin.value) {
        const adminResponse = await fetchAdminPlugins(current.accessToken);
        adminPlugins.value = adminResponse.plugins;
      }
      showFlash("插件已上传并完成安装。", "success");
      return uploaded;
    } catch (error) {
      pluginActionError.value =
        error instanceof Error ? error.message : "上传插件失败，请稍后重试。";
      showFlash(pluginActionError.value, "error");
      return null;
    } finally {
      pluginUploadLoading.value = false;
      pluginUploadingName.value = "";
    }
  }

  async function disablePluginInstallation(installationId: string) {
    const current = authSession.value;

    if (!current) {
      pluginActionError.value = "登录状态已失效，请重新登录。";
      return null;
    }

    pluginSaving.value = true;
    pluginSavingId.value = installationId;
    pluginActionError.value = "";

    try {
      const updated = await disablePlugin(current.accessToken, installationId);
      upsertPlugin(updated);
      adminPlugins.value = adminPlugins.value.map((plugin) =>
        plugin.installation.id === installationId ? updated : plugin,
      );
      availablePlugins.value = availablePlugins.value.map((plugin) =>
        plugin.installation.id === installationId
          ? {
              ...plugin,
              installation: updated.installation,
            }
          : plugin,
      );
      showFlash("插件已停用。", "success");
      return updated;
    } catch (error) {
      pluginActionError.value =
        error instanceof Error ? error.message : "停用插件失败，请稍后重试。";
      showFlash(pluginActionError.value, "error");
      return null;
    } finally {
      pluginSaving.value = false;
      pluginSavingId.value = "";
    }
  }

  async function testPluginConfiguration(
    installationId: string,
    config: Record<string, unknown>,
    secrets: Record<string, string>,
    enabled = true,
  ) {
    const current = authSession.value;

    if (!current) {
      pluginActionError.value = "登录状态已失效，请重新登录。";
      return null;
    }

    pluginTestingId.value = installationId;
    pluginActionError.value = "";

    try {
      const result = await testPluginBinding(current.accessToken, installationId, {
        enabled,
        config,
        secrets,
      });
      showFlash(
        result.valid ? "插件连接测试通过。" : "插件连接测试未通过。",
        result.valid ? "success" : "error",
      );
      return result;
    } catch (error) {
      pluginActionError.value =
        error instanceof Error ? error.message : "测试插件连接失败，请稍后重试。";
      showFlash(pluginActionError.value, "error");
      return null;
    } finally {
      pluginTestingId.value = "";
    }
  }

  async function savePluginConfiguration(
    installationId: string,
    config: Record<string, unknown>,
    secrets: Record<string, string>,
    enabled = true,
  ) {
    const current = authSession.value;

    if (!current) {
      pluginActionError.value = "登录状态已失效，请重新登录。";
      return null;
    }

    pluginSaving.value = true;
    pluginSavingId.value = installationId;
    pluginActionError.value = "";

    try {
      const updated = await savePluginBinding(current.accessToken, installationId, {
        enabled,
        config,
        secrets,
      });
      upsertPlugin(updated);
      showFlash(enabled ? "插件配置已保存并启用。" : "插件配置已保存。", "success");
      return updated;
    } catch (error) {
      pluginActionError.value =
        error instanceof Error ? error.message : "保存插件配置失败，请稍后重试。";
      showFlash(pluginActionError.value, "error");
      return null;
    } finally {
      pluginSaving.value = false;
      pluginSavingId.value = "";
    }
  }

  function setDefaultDevice(deviceId: string) {
    const targetDevice = devices.value.find((device) => device.id === deviceId);

    if (!targetDevice || targetDevice.status === "offline") {
      showFlash("已解绑设备不能设为默认。", "error");
      return;
    }

    defaultDeviceId.value = deviceId;
    showFlash("默认设备已更新。", "success");
  }

  function setTheme(theme: ThemeMode) {
    selectedTheme.value = theme;
    showFlash("主题设置已更新。");
  }

  function setSendConfirmation(enabled: boolean) {
    sendConfirmationEnabled.value = enabled;
    showFlash(enabled ? "已开启发送前确认。" : "新内容会直接进入打印队列。");
  }

  function setLoginProtection(enabled: boolean) {
    loginProtectionEnabled.value = enabled;
    showFlash(enabled ? "刷新后会要求重新登录。" : "刷新后将保留登录状态。");
  }

  async function saveAIServiceConfig(config: {
    providerName: string;
    providerType: string;
    baseUrl: string;
    model: string;
    apiKey: string;
  }) {
    const current = authSession.value;

    if (!current) {
      aiConfigError.value = "登录状态已失效，请重新登录。";
      return false;
    }

    aiConfigSaving.value = true;
    aiConfigError.value = "";

    try {
      const summary = await saveAIConfig(current.accessToken, config);
      applyAIConfig(summary);
      showFlash("AI 服务配置已保存。", "success");
      return true;
    } catch (error) {
      aiConfigError.value =
        error instanceof Error ? error.message : "保存 AI 配置失败，请稍后重试。";
      showFlash(aiConfigError.value, "error");
      return false;
    } finally {
      aiConfigSaving.value = false;
    }
  }

  function setAuthState(user: User, session: AuthSession) {
    authUser.value = user;
    authSession.value = session;
    authError.value = "";
    accountCreationError.value = "";
    aiConfigError.value = "";
  }

  function clearAuthState() {
    if (remoteSaveTimer) {
      window.clearTimeout(remoteSaveTimer);
      remoteSaveTimer = 0;
    }

    authUser.value = null;
    authSession.value = null;
    authError.value = "";
    accountCreationError.value = "";
    aiConfigError.value = "";
    pluginError.value = "";
    pluginActionError.value = "";
    workspaceOwnerId.value = null;
  }

  async function refreshSessionIfNeeded() {
    const current = authSession.value;

    if (!current) {
      return false;
    }

    const expiresAt = new Date(current.accessTokenExpiresAt).getTime();

    if (Number.isNaN(expiresAt)) {
      clearAuthState();
      return false;
    }

    if (expiresAt > Date.now() + 30_000) {
      return true;
    }

    try {
      const refreshed = await refreshAuthSession(current.refreshToken);
      setAuthState(refreshed.user, refreshed.session);
      return true;
    } catch (error) {
      clearAuthState();

      if (error instanceof AuthApiError) {
        authError.value = error.message;
      }

      return false;
    }
  }

  async function initializeAuth() {
    if (!authSession.value) {
      clearAuthState();
      return false;
    }

    authBootstrapping.value = true;

    try {
      const sessionReady = await refreshSessionIfNeeded();

      if (!sessionReady || !authSession.value) {
        return false;
      }

      if (!authUser.value) {
        const user = await fetchCurrentUser(authSession.value.accessToken);
        authUser.value = user;
      }

      const loaded = await loadRemoteWorkspace();
      if (!loaded) {
        clearAuthState();
        authError.value = workspaceSyncError.value || "加载账号数据失败，请重新登录。";
        restoreAnonymousWorkspace();
        return false;
      }

      authError.value = "";
      return true;
    } catch (error) {
      clearAuthState();
      restoreAnonymousWorkspace();
      authError.value = error instanceof Error ? error.message : "登录状态已失效，请重新登录。";
      return false;
    } finally {
      authBootstrapping.value = false;
    }
  }

  async function login(email: string, password: string) {
    authLoading.value = true;
    authError.value = "";

    try {
      const result = await loginWithApi({ email, password });
      setAuthState(result.user, result.session);
      const loaded = await loadRemoteWorkspace();

      if (!loaded) {
        const workspaceError = workspaceSyncError.value || "加载账号数据失败，请稍后重试。";
        clearAuthState();
        restoreAnonymousWorkspace();
        authError.value = workspaceError;
        showFlash(workspaceError, "error");
        return false;
      }

      showFlash("登录成功。", "success");
      return true;
    } catch (error) {
      authError.value = error instanceof Error ? error.message : "登录失败，请稍后重试。";
      showFlash(authError.value, "error");
      return false;
    } finally {
      authLoading.value = false;
    }
  }

  async function changePassword(currentPassword: string, newPassword: string) {
    const current = authSession.value;

    if (!current) {
      authError.value = "登录状态已失效，请重新登录。";
      return false;
    }

    passwordChangeLoading.value = true;
    authError.value = "";

    try {
      await changePasswordWithApi({
        accessToken: current.accessToken,
        currentPassword,
        newPassword,
      });
      clearAuthState();
      restoreAnonymousWorkspace();
      showFlash("密码已更新，请重新登录。", "success");
      return true;
    } catch (error) {
      authError.value = error instanceof Error ? error.message : "修改密码失败，请稍后重试。";
      showFlash(authError.value, "error");
      return false;
    } finally {
      passwordChangeLoading.value = false;
    }
  }

  async function logout() {
    const current = authSession.value;

    await flushRemoteWorkspaceSave();

    if (current) {
      try {
        await logoutWithApi({
          accessToken: current.accessToken,
          refreshToken: current.refreshToken,
        });
      } catch {
        // Local sign-out should still succeed if the backend session already expired.
      }
    }

    clearAuthState();
    restoreAnonymousWorkspace();
    showFlash("已退出当前账号。");
  }

  async function createAccount(email: string, name: string, password: string) {
    const current = authSession.value;

    if (!current) {
      accountCreationError.value = "登录状态已失效，请重新登录。";
      return false;
    }

    accountCreationLoading.value = true;
    accountCreationError.value = "";

    try {
      await createUserWithApi(current.accessToken, {
        email,
        name,
        password,
      });
      showFlash("新账号已创建。", "success");
      return true;
    } catch (error) {
      accountCreationError.value =
        error instanceof Error ? error.message : "创建账号失败，请稍后重试。";
      showFlash(accountCreationError.value, "error");
      return false;
    } finally {
      accountCreationLoading.value = false;
    }
  }

  function formatPrintTime(iso: string) {
    return formatRelativeTimestamp(iso);
  }

  function getDeviceName(deviceId: string) {
    return deviceMap.value[deviceId]?.name ?? "未设置设备";
  }

  function toggleConversationMessageSelection(messageId: string) {
    selectedConversationMessageIds.value = selectedConversationMessageIds.value.includes(messageId)
      ? selectedConversationMessageIds.value.filter((current) => current !== messageId)
      : [...selectedConversationMessageIds.value, messageId];
  }

  return {
    authUser,
    authSession,
    authLoading,
    authError,
    authBootstrapping,
    passwordChangeLoading,
    workspaceLoading,
    workspaceSyncing,
    workspaceSyncError,
    accountCreationLoading,
    accountCreationError,
    aiConfigSummary,
    aiConfigLoading,
    aiConfigSaving,
    aiConfigError,
    printerSyncError,
    pluginLoading,
    pluginSaving,
    pluginUploadLoading,
    pluginError,
    pluginActionError,
    pluginTestingId,
    pluginSavingId,
    pluginUploadingName,
    devices,
    conversations,
    activeConversationId,
    printJobs,
    schedules,
    sources,
    availablePlugins,
    adminPlugins,
    remoteSchedules,
    loginProtectionEnabled,
    sendConfirmationEnabled,
    selectedTheme,
    defaultDeviceId,
    serviceBinding,
    isGenerating,
    generationError,
    selectedConversationMessageIds,
    isCreatingPrint,
    flashMessage,
    flashTone,
    activeConversation,
    conversationMessages,
    selectedConversationMessages,
    defaultDevice,
    pendingPrintJobs,
    recentPrintJobs,
    activeSchedules,
    activeSources,
    connectedDevicesCount,
    enabledSchedulesCount,
    pendingConfirmationCount,
    summaryCards,
    activeDeviceLabel,
    activeModelLabel,
    todayPrintCount,
    welcomeLabel,
    isConfigured,
    isAuthenticated,
    isAdmin,
    selectConversation,
    createConversation,
    deleteConversation,
    updateCurrentDraft,
    saveCurrentDraft,
    sendCurrentDraft,
    regenerateLatestReply,
    createPrintFromSelectedMessages,
    createPrintFromConversation,
    createManualPrint,
    confirmPrint,
    cancelPrint,
    updatePrintDevice,
    addDevice,
    removeDevice,
    toggleSchedule,
    createSchedule,
    updateScheduleDevice,
    deleteSchedule,
    toggleSourceConnection,
    uploadPlugin,
    disablePluginInstallation,
    testPluginConfiguration,
    savePluginConfiguration,
    setDefaultDevice,
    setTheme,
    setSendConfirmation,
    setLoginProtection,
    saveAIServiceConfig,
    initializeAuth,
    refreshSessionIfNeeded,
    changePassword,
    createAccount,
    login,
    logout,
    formatPrintTime,
    getDeviceName,
    getDeviceStatusLabel,
    getPrintStatusLabel,
    getSourceStatusLabel,
    toggleConversationMessageSelection,
  };
});

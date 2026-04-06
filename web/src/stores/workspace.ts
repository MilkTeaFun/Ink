import { defineStore } from "pinia";
import { computed, ref, watch } from "vue";

import { generateReplyWithMockService, loginWithMockService } from "@/services/mockInk";
import type {
  AnswerStyle,
  Conversation,
  ConversationMessage,
  Device,
  NoteStyle,
  PersistedWorkspaceState,
  Preferences,
  PrintJob,
  ResponseLength,
  Schedule,
  ServiceBinding,
  SourceConnection,
  ThemeMode,
  User,
} from "@/types/workspace";
import {
  createId,
  formatRelativeTimestamp,
  getDeviceStatusLabel,
  getPrintStatusLabel,
  getSourceStatusLabel,
} from "@/utils/workspace";

const STORAGE_KEY = "ink.workspace.v1";

function getNow() {
  return new Date().toISOString();
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
    loginProtectionEnabled: true,
    sendConfirmationEnabled: true,
    theme: "light",
    answerStyle: "clear-gentle",
    noteStyle: "clean",
    responseLength: "medium",
    defaultDeviceId: "device-desk",
  };
  const serviceBinding: ServiceBinding = {
    providerName: null,
    modelName: "清楚温柔",
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

function readPersistedWorkspaceState() {
  if (typeof window === "undefined") {
    return createSeedState();
  }

  const raw = window.localStorage.getItem(STORAGE_KEY);

  if (!raw) {
    return createSeedState();
  }

  try {
    return {
      ...createSeedState(),
      ...JSON.parse(raw),
    } as PersistedWorkspaceState;
  } catch {
    return createSeedState();
  }
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

export const useWorkspaceStore = defineStore("workspace", () => {
  const persisted = readPersistedWorkspaceState();

  const authUser = ref<User | null>(persisted.authUser);
  const authLoading = ref(false);
  const authError = ref("");

  const devices = ref<Device[]>(persisted.devices);
  const conversations = ref<Conversation[]>(persisted.conversations);
  const activeConversationId = ref(persisted.activeConversationId);
  const printJobs = ref<PrintJob[]>(persisted.printJobs);
  const schedules = ref<Schedule[]>(persisted.schedules);
  const sources = ref<SourceConnection[]>(persisted.sources);

  const loginProtectionEnabled = ref(persisted.preferences.loginProtectionEnabled);
  const sendConfirmationEnabled = ref(persisted.preferences.sendConfirmationEnabled);
  const selectedTheme = ref<ThemeMode>(persisted.preferences.theme);
  const activeAnswerStyle = ref<AnswerStyle>(persisted.preferences.answerStyle);
  const activeNoteStyle = ref<NoteStyle>(persisted.preferences.noteStyle);
  const responseLength = ref<ResponseLength>(persisted.preferences.responseLength);
  const defaultDeviceId = ref(persisted.preferences.defaultDeviceId);

  const serviceBinding = ref<ServiceBinding>(persisted.serviceBinding);

  const isGenerating = ref(false);
  const generationError = ref("");
  const selectedAssistantMessageId = ref("");
  const isCreatingPrint = ref(false);
  const flashMessage = ref("");
  const flashTone = ref<"success" | "error" | "info">("info");
  let flashTimer = 0;

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
  const assistantMessages = computed(
    () =>
      activeConversation.value?.messages.filter((message) => message.role === "assistant") ?? [],
  );
  const selectedAssistantMessage = computed(() => {
    if (selectedAssistantMessageId.value) {
      return (
        assistantMessages.value.find(
          (message) => message.id === selectedAssistantMessageId.value,
        ) ?? null
      );
    }

    return assistantMessages.value.at(-1) ?? null;
  });
  const defaultDevice = computed(
    () => devices.value.find((device) => device.id === defaultDeviceId.value) ?? null,
  );
  const sortedPrintJobs = computed(() => sortPrintJobsByUpdatedAt(printJobs.value));
  const pendingPrintJobs = computed(() =>
    sortedPrintJobs.value.filter((job) => job.status === "pending" || job.status === "queued"),
  );
  const recentPrintJobs = computed(() => sortedPrintJobs.value.slice(0, 5));
  const connectedDevicesCount = computed(
    () => devices.value.filter((device) => device.status === "connected").length,
  );
  const enabledSchedulesCount = computed(
    () => schedules.value.filter((schedule) => schedule.enabled).length,
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
  const activeModelLabel = computed(() => serviceBinding.value.modelName);
  const todayPrintCount = computed(() => todayCompletedCount.value);
  const welcomeLabel = computed(() => {
    if (activeAnswerStyle.value === "warm-encouraging") {
      return "慢一点，也能把想说的话说好";
    }

    if (activeAnswerStyle.value === "concise-direct") {
      return "说重点，剩下的交给打印";
    }

    return "简单一点，也可以很舒服";
  });
  const isConfigured = computed(() => activeDeviceLabel.value !== "");
  const isAuthenticated = computed(() => authUser.value !== null);

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
      progress: schedules.value.length
        ? Math.round((enabledSchedulesCount.value / schedules.value.length) * 100)
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

  const persistableState = computed<PersistedWorkspaceState>(() => ({
    authUser: loginProtectionEnabled.value ? null : authUser.value,
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
      answerStyle: activeAnswerStyle.value,
      noteStyle: activeNoteStyle.value,
      responseLength: responseLength.value,
      defaultDeviceId: defaultDeviceId.value,
    },
    serviceBinding: serviceBinding.value,
  }));

  watch(
    persistableState,
    (value) => {
      if (typeof window === "undefined") {
        return;
      }

      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
    },
    { deep: true, immediate: true },
  );

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
    selectedAssistantMessageId.value = "";
  }

  function createConversation() {
    const conversation: Conversation = {
      id: createId("conversation"),
      title: "新对话",
      preview: "从这里开始整理新的内容",
      updatedAt: getNow(),
      draft: "",
      messages: [],
    };

    conversations.value = [conversation, ...conversations.value];
    selectConversation(conversation.id);
    showFlash("已创建新的对话草稿。");
  }

  function updateCurrentDraft(value: string) {
    if (!activeConversation.value) {
      return;
    }

    updateConversation(activeConversation.value.id, (conversation) => ({
      ...conversation,
      draft: value,
      updatedAt: getNow(),
    }));
  }

  function saveCurrentDraft() {
    if (!activeConversation.value) {
      return;
    }

    showFlash("草稿已保存在本地。", "success");
  }

  function ensureSelectedAssistantMessage() {
    const latestAssistant = assistantMessages.value.at(-1);
    selectedAssistantMessageId.value = latestAssistant?.id ?? "";
  }

  async function sendCurrentDraft() {
    const conversation = activeConversation.value;

    if (!conversation || isGenerating.value) {
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
      const reply = await generateReplyWithMockService({
        prompt,
        answerStyle: activeAnswerStyle.value,
        noteStyle: activeNoteStyle.value,
        responseLength: responseLength.value,
      });
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
      selectedAssistantMessageId.value = assistantMessage.id;
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
      const reply = await generateReplyWithMockService({
        prompt: latestUserMessage.text,
        answerStyle: activeAnswerStyle.value,
        noteStyle: activeNoteStyle.value,
        responseLength: responseLength.value,
      });
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
      selectedAssistantMessageId.value = assistantMessage.id;
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
    const job = buildPrintJob(title, content, source);
    printJobs.value = [job, ...printJobs.value];

    if (job.status === "queued") {
      showFlash("内容已直接加入打印队列。", "success");
      await maybeCompleteQueuedJob(job.id);
    } else {
      showFlash("已加入待确认打印。", "success");
    }

    isCreatingPrint.value = false;
    return job;
  }

  async function createPrintFromLatestReply() {
    const message = assistantMessages.value.at(-1);

    if (!message) {
      showFlash("当前还没有可打印的回答。", "error");
      return null;
    }

    return addPrintJob(activeConversation.value?.title ?? "最新回答", message.text, "对话最新回答");
  }

  async function createPrintFromSelectedReply() {
    if (!selectedAssistantMessage.value) {
      showFlash("请先选中一条回答。", "error");
      return null;
    }

    return addPrintJob(
      activeConversation.value?.title ?? "选中回答",
      selectedAssistantMessage.value.text,
      "对话选中回答",
    );
  }

  async function createPrintFromConversation() {
    const conversation = activeConversation.value;

    if (!conversation || conversation.messages.length === 0) {
      showFlash("当前对话还没有内容可打印。", "error");
      return null;
    }

    const content = conversation.messages.map((message) => message.text).join("\n");
    return addPrintJob(conversation.title, content, "整段对话");
  }

  async function createManualPrint() {
    return addPrintJob("手动新建纸条", "新的打印内容已创建，你可以稍后继续编辑。", "手动打印");
  }

  async function confirmPrint(jobId: string) {
    const target = printJobs.value.find((job) => job.id === jobId);

    if (!target || target.status !== "pending") {
      return false;
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

  function updatePrintDevice(jobId: string, deviceId: string) {
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

  function toggleSchedule(scheduleId: string) {
    schedules.value = schedules.value.map((schedule) =>
      schedule.id === scheduleId ? { ...schedule, enabled: !schedule.enabled } : schedule,
    );
    showFlash("定时任务状态已更新。", "success");
  }

  function createSchedule() {
    const schedule: Schedule = {
      id: createId("schedule"),
      title: "新的定时任务",
      source: "手动创建",
      timeLabel: "每天 19:30",
      deviceId: defaultDeviceId.value,
      enabled: true,
    };

    schedules.value = [schedule, ...schedules.value];
    showFlash("已创建新的定时任务。", "success");
  }

  function updateScheduleDevice(scheduleId: string, deviceId: string) {
    schedules.value = schedules.value.map((schedule) =>
      schedule.id === scheduleId ? { ...schedule, deviceId } : schedule,
    );
    showFlash("定时任务设备已更新。", "success");
  }

  function cycleSourceStatus(sourceId: string) {
    const nextStatus = {
      connected: "error",
      error: "disconnected",
      disconnected: "connected",
    } as const;

    sources.value = sources.value.map((source) =>
      source.id === sourceId
        ? {
            ...source,
            status: nextStatus[source.status],
            note:
              nextStatus[source.status] === "connected"
                ? "连接正常"
                : nextStatus[source.status] === "error"
                  ? "最近同步失败，请重新授权"
                  : "尚未连接到此来源",
          }
        : source,
    );
    showFlash("来源状态已更新。", "success");
  }

  function setDefaultDevice(deviceId: string) {
    defaultDeviceId.value = deviceId;
    showFlash("默认设备已更新。", "success");
  }

  function setAnswerStyle(style: AnswerStyle) {
    activeAnswerStyle.value = style;
    serviceBinding.value = {
      ...serviceBinding.value,
      modelName:
        style === "warm-encouraging"
          ? "温柔鼓励"
          : style === "concise-direct"
            ? "直接简洁"
            : "清楚温柔",
    };
    showFlash("回答风格已更新。", "success");
  }

  function setNoteStyle(style: NoteStyle) {
    activeNoteStyle.value = style;
    showFlash("纸条风格已更新。", "success");
  }

  function setResponseLength(length: ResponseLength) {
    responseLength.value = length;
    showFlash("回复长度偏好已更新。");
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
    showFlash(enabled ? "刷新后会要求重新登录。" : "已允许保留本地登录状态。");
  }

  function bindService() {
    serviceBinding.value = {
      providerName: "Ink Mock AI",
      modelName: activeModelLabel.value,
      bound: true,
    };
    showFlash("已绑定前端 mock AI 服务。", "success");
  }

  async function login(email: string, password: string) {
    authLoading.value = true;
    authError.value = "";

    try {
      authUser.value = await loginWithMockService({ email, password });
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

  function logout() {
    authUser.value = null;
    authError.value = "";
    showFlash("已退出当前账号。");
  }

  function formatPrintTime(iso: string) {
    return formatRelativeTimestamp(iso);
  }

  function getDeviceName(deviceId: string) {
    return deviceMap.value[deviceId]?.name ?? "未设置设备";
  }

  function selectAssistantMessage(messageId: string) {
    selectedAssistantMessageId.value = messageId;
  }

  ensureSelectedAssistantMessage();

  return {
    authUser,
    authLoading,
    authError,
    devices,
    conversations,
    activeConversationId,
    printJobs,
    schedules,
    sources,
    loginProtectionEnabled,
    sendConfirmationEnabled,
    selectedTheme,
    activeAnswerStyle,
    activeNoteStyle,
    responseLength,
    defaultDeviceId,
    serviceBinding,
    isGenerating,
    generationError,
    selectedAssistantMessageId,
    isCreatingPrint,
    flashMessage,
    flashTone,
    activeConversation,
    assistantMessages,
    selectedAssistantMessage,
    defaultDevice,
    pendingPrintJobs,
    recentPrintJobs,
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
    selectConversation,
    createConversation,
    updateCurrentDraft,
    saveCurrentDraft,
    sendCurrentDraft,
    regenerateLatestReply,
    createPrintFromLatestReply,
    createPrintFromSelectedReply,
    createPrintFromConversation,
    createManualPrint,
    confirmPrint,
    updatePrintDevice,
    toggleSchedule,
    createSchedule,
    updateScheduleDevice,
    cycleSourceStatus,
    setDefaultDevice,
    setAnswerStyle,
    setNoteStyle,
    setResponseLength,
    setTheme,
    setSendConfirmation,
    setLoginProtection,
    bindService,
    login,
    logout,
    formatPrintTime,
    getDeviceName,
    getDeviceStatusLabel,
    getPrintStatusLabel,
    getSourceStatusLabel,
    selectAssistantMessage,
  };
});

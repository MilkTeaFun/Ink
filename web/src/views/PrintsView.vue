<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink } from "vue-router";

import AppDialog from "@/components/AppDialog.vue";
import { useWorkspaceStore } from "@/stores/workspace";
import type { PluginFieldSpec } from "@/types/plugins";
import { buildPluginFieldDefaults, getPluginFieldDefaultValue } from "@/utils/plugins";
import { getPrintStatusBadgeClass, getSourceStatusBadgeClass } from "@/utils/workspace";

const workspaceStore = useWorkspaceStore();

const weekdayOptions = [
  { label: "周日", value: 0 },
  { label: "周一", value: 1 },
  { label: "周二", value: 2 },
  { label: "周三", value: 3 },
  { label: "周四", value: 4 },
  { label: "周五", value: 5 },
  { label: "周六", value: 6 },
] as const;

const defaultSettings = computed(() => [
  { label: "默认设备", value: workspaceStore.activeDeviceLabel || "暂未设置" },
  { label: "发送前确认", value: workspaceStore.sendConfirmationEnabled ? "已开启" : "已关闭" },
]);

const printDialogOpen = ref(false);
const scheduleDialogOpen = ref(false);
const printTitle = ref("");
const printContent = ref("");
const printError = ref("");
const scheduleTitle = ref("");
const scheduleSource = ref("");
const scheduleTime = ref("每天 19:30");
const schedulePluginInstallationId = ref("");
const scheduleFrequencyType = ref<"daily" | "weekly">("daily");
const scheduleTimezone = ref(Intl.DateTimeFormat().resolvedOptions().timeZone || "Asia/Shanghai");
const scheduleHour = ref(19);
const scheduleMinute = ref(30);
const scheduleWeekdays = ref<number[]>([]);
const scheduleConfigDraft = ref<Record<string, unknown>>({});
const scheduleDeviceId = ref("");
const scheduleError = ref("");

const connectedPlugins = computed(() =>
  workspaceStore.availablePlugins.filter(
    (plugin) =>
      plugin.installation.status === "ready" &&
      plugin.binding?.enabled &&
      plugin.binding.status === "connected",
  ),
);

const selectedSchedulePlugin = computed(
  () =>
    connectedPlugins.value.find(
      (plugin) => plugin.installation.id === schedulePluginInstallationId.value,
    ) ?? null,
);

watch(
  () => selectedSchedulePlugin.value?.installation.id ?? "",
  (pluginID, previousPluginID) => {
    if (!pluginID) {
      scheduleConfigDraft.value = {};
      return;
    }

    if (pluginID === previousPluginID) {
      return;
    }

    scheduleConfigDraft.value = buildPluginFieldDefaults(
      selectedSchedulePlugin.value?.manifest.scheduleConfigSchema ?? [],
    );
  },
  { immediate: true },
);

function handlePrintDeviceChange(jobId: string, event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.updatePrintDevice(jobId, target?.value ?? workspaceStore.defaultDeviceId);
}

function handleScheduleDeviceChange(scheduleId: string, event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.updateScheduleDevice(scheduleId, target?.value ?? workspaceStore.defaultDeviceId);
}

async function handleScheduleDelete(scheduleId: string) {
  if (typeof window !== "undefined" && !window.confirm("确认删除这条定时任务吗？")) {
    return;
  }

  await workspaceStore.deleteSchedule(scheduleId);
}

function scheduleFieldValue(field: PluginFieldSpec) {
  return scheduleConfigDraft.value[field.key] ?? getPluginFieldDefaultValue(field);
}

function updateScheduleField(field: PluginFieldSpec, value: unknown) {
  scheduleConfigDraft.value = {
    ...scheduleConfigDraft.value,
    [field.key]:
      field.type === "number" && value !== ""
        ? Number(value)
        : field.type === "checkbox"
          ? Boolean(value)
          : value,
  };
}

function handleScheduleFieldInput(field: PluginFieldSpec, event: Event) {
  const target = event.target as HTMLInputElement | HTMLTextAreaElement | null;
  updateScheduleField(field, target?.value ?? "");
}

function handleScheduleFieldSelect(field: PluginFieldSpec, event: Event) {
  const target = event.target as HTMLSelectElement | null;
  updateScheduleField(field, target?.value ?? "");
}

function handleScheduleFieldCheckbox(field: PluginFieldSpec, event: Event) {
  const target = event.target as HTMLInputElement | null;
  updateScheduleField(field, target?.checked ?? false);
}

function toggleWeekday(weekday: number) {
  scheduleWeekdays.value = scheduleWeekdays.value.includes(weekday)
    ? scheduleWeekdays.value.filter((value) => value !== weekday)
    : weekdayOptions
        .map((option) => option.value)
        .filter((value) => [...scheduleWeekdays.value, weekday].includes(value));
}

function openPrintDialog() {
  printDialogOpen.value = true;
  printTitle.value = "";
  printContent.value = "";
  printError.value = "";
}

function closePrintDialog() {
  printDialogOpen.value = false;
}

async function submitPrintDialog() {
  printError.value = "";

  if (!printTitle.value.trim()) {
    printError.value = "请输入打印标题。";
    return;
  }

  if (!printContent.value.trim()) {
    printError.value = "请输入打印内容。";
    return;
  }

  const created = await workspaceStore.createManualPrint({
    title: printTitle.value,
    content: printContent.value,
  });

  if (!created) {
    printError.value =
      workspaceStore.flashTone === "error" ? workspaceStore.flashMessage : "创建打印失败。";
    return;
  }

  closePrintDialog();
}

function openScheduleDialog() {
  scheduleDialogOpen.value = true;
  scheduleTitle.value = "";
  scheduleSource.value = "";
  scheduleTime.value = "每天 19:30";
  schedulePluginInstallationId.value = connectedPlugins.value[0]?.installation.id ?? "";
  scheduleFrequencyType.value = "daily";
  scheduleTimezone.value = Intl.DateTimeFormat().resolvedOptions().timeZone || "Asia/Shanghai";
  scheduleHour.value = 19;
  scheduleMinute.value = 30;
  scheduleWeekdays.value = [];
  scheduleConfigDraft.value = buildPluginFieldDefaults(
    connectedPlugins.value[0]?.manifest.scheduleConfigSchema ?? [],
  );
  scheduleDeviceId.value = workspaceStore.defaultDeviceId;
  scheduleError.value = "";
}

function closeScheduleDialog() {
  scheduleDialogOpen.value = false;
}

async function submitScheduleDialog() {
  scheduleError.value = "";

  if (!scheduleTitle.value.trim()) {
    scheduleError.value = "请输入任务名称。";
    return;
  }

  if (!(scheduleDeviceId.value || workspaceStore.defaultDeviceId)) {
    scheduleError.value = "请先绑定设备，再创建定时任务。";
    return;
  }

  if (workspaceStore.isAuthenticated) {
    if (!schedulePluginInstallationId.value) {
      scheduleError.value = "请先选择一个已启用的插件来源。";
      return;
    }

    if (scheduleFrequencyType.value === "weekly" && scheduleWeekdays.value.length === 0) {
      scheduleError.value = "每周任务至少选择一天。";
      return;
    }

    const created = await workspaceStore.createSchedule({
      title: scheduleTitle.value,
      deviceId: scheduleDeviceId.value || workspaceStore.defaultDeviceId,
      pluginInstallationId: schedulePluginInstallationId.value,
      frequencyType: scheduleFrequencyType.value,
      timezone: scheduleTimezone.value,
      hour: scheduleHour.value,
      minute: scheduleMinute.value,
      weekdays: scheduleWeekdays.value,
      scheduleConfig: scheduleConfigDraft.value,
    });

    if (!created) {
      scheduleError.value =
        workspaceStore.flashTone === "error" ? workspaceStore.flashMessage : "创建定时任务失败。";
      return;
    }

    closeScheduleDialog();
    return;
  }

  if (!scheduleTime.value.trim()) {
    scheduleError.value = "请输入执行时间。";
    return;
  }

  await workspaceStore.createSchedule({
    title: scheduleTitle.value,
    source: scheduleSource.value || "手动创建",
    deviceId: scheduleDeviceId.value || workspaceStore.defaultDeviceId,
  });
  closeScheduleDialog();
}
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-6 pt-4 sm:space-y-8">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h2 class="text-2xl font-semibold tracking-tight text-stone-900">打印</h2>
      </div>
      <div class="flex flex-wrap gap-2">
        <RouterLink to="/tutorial" class="ui-btn-secondary px-3 py-1.5 text-sm">
          绑定教程
        </RouterLink>
        <button class="ui-btn-primary px-3 py-1.5 text-sm" @click="openPrintDialog">
          新建打印
        </button>
        <button class="ui-btn-secondary px-3 py-1.5 text-sm" @click="openScheduleDialog">
          新建定时任务
        </button>
      </div>
    </div>

    <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div class="space-y-8">
        <section
          v-if="workspaceStore.printerSyncError"
          class="rounded-2xl border border-amber-200 bg-amber-50 px-5 py-4"
        >
          <p class="text-sm font-medium text-amber-900">设备状态同步异常</p>
          <p class="mt-1 text-sm text-amber-700">{{ workspaceStore.printerSyncError }}</p>
        </section>

        <section>
          <div class="mb-4">
            <h3 class="text-base leading-6 font-semibold text-stone-900">待处理打印</h3>
          </div>

          <div
            v-if="workspaceStore.pendingPrintJobs.length === 0"
            class="rounded-2xl border border-dashed border-stone-200 bg-stone-50 px-6 py-10 text-center"
          >
            <h4 class="text-base font-semibold text-stone-900">当前没有待处理打印</h4>
            <p class="mt-2 text-sm text-stone-500">
              {{
                workspaceStore.isAuthenticated
                  ? "绑定设备后可以先在对话页生成内容，再回到这里确认是否出纸。"
                  : "当前未登录时显示的是演示数据流，登录后会切到各账号自己的真实打印记录。"
              }}
            </p>
          </div>

          <div v-else class="ui-list-card">
            <article
              v-for="item in workspaceStore.pendingPrintJobs"
              :key="item.id"
              class="ui-list-row flex flex-col gap-4"
            >
              <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <p class="text-sm font-medium text-stone-900">{{ item.title }}</p>
                    <span
                      class="ui-status-badge"
                      :class="getPrintStatusBadgeClass(item.status)"
                    >
                      {{ workspaceStore.getPrintStatusLabel(item.status) }}
                    </span>
                  </div>
                  <p class="mt-1 text-sm text-stone-500">
                    {{ item.source }} · {{ workspaceStore.getDeviceName(item.deviceId) }} ·
                    {{ workspaceStore.formatPrintTime(item.updatedAt) }}
                  </p>
                  <p
                    class="mt-2 rounded-lg bg-stone-50 px-3 py-2 text-sm leading-relaxed text-stone-600"
                  >
                    {{ item.content }}
                  </p>
                </div>

                <div
                  class="flex w-full flex-wrap items-start gap-2 self-stretch sm:w-auto sm:self-start"
                >
                  <button
                    v-if="item.status === 'pending'"
                    class="ui-btn-primary px-3 py-1.5 text-sm whitespace-nowrap"
                    @click="workspaceStore.confirmPrint(item.id)"
                  >
                    确认打印
                  </button>
                  <button
                    v-if="!workspaceStore.isAuthenticated || item.status === 'pending'"
                    class="ui-btn-secondary px-3 py-1.5 text-sm whitespace-nowrap"
                    @click="workspaceStore.cancelPrint(item.id)"
                  >
                    取消打印
                  </button>
                </div>
              </div>

              <div class="flex flex-col gap-2 md:flex-row md:items-center">
                <label class="text-sm font-medium text-stone-700">目标设备</label>
                <select
                  :value="item.deviceId"
                  :disabled="workspaceStore.isAuthenticated && item.status === 'queued'"
                  class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 md:w-auto"
                  @change="handlePrintDeviceChange(item.id, $event)"
                >
                  <option
                    v-for="device in workspaceStore.devices.filter(
                      (device) => device.status !== 'offline',
                    )"
                    :key="device.id"
                    :value="device.id"
                  >
                    {{ device.name }}
                  </option>
                </select>
                <p
                  v-if="workspaceStore.isAuthenticated && item.status === 'queued'"
                  class="text-sm text-stone-500"
                >
                  已提交到咕咕机后不能再取消或改绑设备。
                </p>
              </div>
            </article>
          </div>
        </section>

        <section>
          <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <h3 class="text-base leading-6 font-semibold text-stone-900">定时任务</h3>
          </div>

          <div
            v-if="workspaceStore.activeSchedules.length === 0"
            class="rounded-2xl border border-dashed border-stone-200 bg-stone-50 px-6 py-10 text-center"
          >
            <h4 class="text-base font-semibold text-stone-900">还没有定时任务</h4>
          </div>

          <div v-else class="ui-list-card">
            <article
              v-for="task in workspaceStore.activeSchedules"
              :key="task.id"
              class="ui-list-row flex flex-col gap-4"
            >
              <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                <div class="min-w-0">
                  <p class="text-sm font-medium text-stone-900">{{ task.title }}</p>
                  <p class="mt-1 text-sm text-stone-500">
                    {{ task.source }} · {{ task.timeLabel }} · 发往
                    {{ workspaceStore.getDeviceName(task.deviceId) }}
                  </p>
                  <p v-if="task.nextRunAt" class="mt-1 text-xs text-stone-500">
                    下次执行 {{ workspaceStore.formatPrintTime(task.nextRunAt) }}
                  </p>
                  <p
                    v-if="task.lastError"
                    class="mt-3 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700"
                  >
                    {{ task.lastError }}
                  </p>
                </div>

                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    class="ui-btn-secondary px-3 py-1.5 text-sm whitespace-nowrap"
                    @click="handleScheduleDelete(task.id)"
                  >
                    删除
                  </button>
                  <button
                    type="button"
                    class="ui-toggle"
                    :class="{ 'is-on': task.enabled }"
                    :aria-label="`${task.enabled ? '关闭' : '开启'}${task.title}`"
                    :aria-pressed="task.enabled"
                    @click="workspaceStore.toggleSchedule(task.id)"
                  >
                    <span class="ui-toggle-thumb" />
                  </button>
                </div>
              </div>

              <div class="flex flex-col gap-2 md:flex-row md:items-center">
                <label class="text-sm font-medium text-stone-700">发送设备</label>
                <select
                  :value="task.deviceId"
                  class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 md:w-auto"
                  @change="handleScheduleDeviceChange(task.id, $event)"
                >
                  <option
                    v-for="device in workspaceStore.devices.filter(
                      (device) => device.status !== 'offline',
                    )"
                    :key="device.id"
                    :value="device.id"
                  >
                    {{ device.name }}
                  </option>
                </select>
              </div>
            </article>
          </div>
        </section>

        <section>
          <div class="mb-4">
            <h3 class="text-base leading-6 font-semibold text-stone-900">最近打印</h3>
          </div>

          <div class="ui-list-card p-4">
            <div class="ui-timeline">
              <article
                v-for="item in workspaceStore.recentPrintJobs"
                :key="item.id"
                class="ui-timeline-item"
              >
                <div class="ui-timeline-row">
                  <div class="ui-timeline-copy">
                    <p class="truncate text-sm font-medium text-stone-900">{{ item.title }}</p>
                    <p class="mt-0.5 text-sm text-stone-500">
                      {{ workspaceStore.getDeviceName(item.deviceId) }} ·
                      {{ workspaceStore.formatPrintTime(item.updatedAt) }}
                    </p>
                  </div>
                  <span
                    class="ui-status-badge sm:self-center"
                    :class="getPrintStatusBadgeClass(item.status)"
                  >
                    {{ workspaceStore.getPrintStatusLabel(item.status) }}
                  </span>
                </div>
              </article>
            </div>
          </div>
        </section>
      </div>

      <aside class="space-y-8">
        <section>
          <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">默认打印设置</h3>
            </div>
            <RouterLink to="/settings" class="ui-btn-secondary px-3 py-1.5 text-sm">
              调整
            </RouterLink>
          </div>

          <div class="ui-list-card">
            <div v-for="item in defaultSettings" :key="item.label" class="ui-list-row">
              <p class="text-sm font-medium text-stone-900">{{ item.label }}</p>
              <p class="mt-1 text-sm text-stone-500">{{ item.value }}</p>
            </div>
          </div>
          <p class="mt-3 text-sm text-stone-500">
            如果你还没绑定咕咕机，先去教程页拿到设备编号，再回到状态页完成绑定。
          </p>
        </section>

        <section>
          <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <h3 class="text-base leading-6 font-semibold text-stone-900">已连接插件</h3>
            <RouterLink to="/settings" class="ui-btn-secondary px-3 py-1.5 text-sm"
              >更多设置</RouterLink
            >
          </div>

          <div class="ui-list-card">
            <article
              v-for="source in workspaceStore.activeSources"
              :key="source.id"
              class="ui-list-row"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-stone-900">{{ source.name }}</p>
                  <p class="mt-0.5 text-sm text-stone-500">{{ source.type }} · {{ source.note }}</p>
                </div>
                <span
                  class="ui-status-badge"
                  :class="getSourceStatusBadgeClass(source.status)"
                >
                  {{ workspaceStore.getSourceStatusLabel(source.status) }}
                </span>
              </div>
            </article>
          </div>
        </section>
      </aside>
    </div>

    <AppDialog
      :open="printDialogOpen"
      title="新建打印"
      description="创建一条新的打印内容。"
      @close="closePrintDialog"
    >
      <form class="space-y-4" @submit.prevent="submitPrintDialog">
        <label class="block">
          <span class="mb-2 block text-sm font-medium text-stone-900">打印标题</span>
          <input
            v-model="printTitle"
            type="text"
            placeholder="例如：晚安留言"
            class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
          />
        </label>

        <label class="block">
          <span class="mb-2 block text-sm font-medium text-stone-900">打印内容</span>
          <textarea
            v-model="printContent"
            rows="4"
            placeholder="输入要打印的内容"
            class="w-full rounded-xl border border-stone-200 bg-white px-4 py-3 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
          />
        </label>

        <p v-if="printError" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {{ printError }}
        </p>

        <div class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <button
            type="button"
            class="ui-btn-secondary px-4 py-2 text-sm"
            @click="closePrintDialog"
          >
            取消
          </button>
          <button type="submit" class="ui-btn-primary px-4 py-2 text-sm">创建打印</button>
        </div>
      </form>
    </AppDialog>

    <AppDialog
      :open="scheduleDialogOpen"
      title="新建定时任务"
      :description="
        workspaceStore.isAuthenticated
          ? '选择已连接插件作为来源，并配置执行时间。'
          : '创建一条自动打印计划。'
      "
      @close="closeScheduleDialog"
    >
      <form class="space-y-4" @submit.prevent="submitScheduleDialog">
        <label class="block">
          <span class="mb-2 block text-sm font-medium text-stone-900">任务名称</span>
          <input
            v-model="scheduleTitle"
            type="text"
            placeholder="例如：晨间提醒"
            class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
          />
        </label>

        <template v-if="workspaceStore.isAuthenticated">
          <label class="block">
            <span class="mb-2 block text-sm font-medium text-stone-900">来源插件</span>
            <select
              v-model="schedulePluginInstallationId"
              class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
            >
              <option value="">请选择一个插件</option>
              <option
                v-for="plugin in connectedPlugins"
                :key="plugin.installation.id"
                :value="plugin.installation.id"
              >
                {{ plugin.installation.displayName }}
              </option>
            </select>
          </label>

          <div
            v-if="connectedPlugins.length === 0"
            class="rounded-lg bg-stone-50 px-4 py-3 text-sm text-stone-500"
          >
            当前没有可用插件，请先去设置页完成插件安装和工作区配置。
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <label class="block">
              <span class="mb-2 block text-sm font-medium text-stone-900">频率</span>
              <select
                v-model="scheduleFrequencyType"
                class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
              >
                <option value="daily">每天</option>
                <option value="weekly">每周</option>
              </select>
            </label>

            <label class="block">
              <span class="mb-2 block text-sm font-medium text-stone-900">时区</span>
              <input
                v-model="scheduleTimezone"
                type="text"
                placeholder="例如：Asia/Shanghai"
                class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
              />
            </label>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <label class="block">
              <span class="mb-2 block text-sm font-medium text-stone-900">小时</span>
              <input
                v-model.number="scheduleHour"
                type="number"
                min="0"
                max="23"
                class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
              />
            </label>

            <label class="block">
              <span class="mb-2 block text-sm font-medium text-stone-900">分钟</span>
              <input
                v-model.number="scheduleMinute"
                type="number"
                min="0"
                max="59"
                class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
              />
            </label>
          </div>

          <div v-if="scheduleFrequencyType === 'weekly'" class="block">
            <span class="mb-2 block text-sm font-medium text-stone-900">执行日期</span>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="weekday in weekdayOptions"
                :key="weekday.value"
                type="button"
                class="rounded-full px-3 py-1.5 text-sm ring-1 transition ring-inset"
                :class="
                  scheduleWeekdays.includes(weekday.value)
                    ? 'bg-stone-900 text-white ring-stone-900'
                    : 'bg-white text-stone-700 ring-stone-200'
                "
                @click="toggleWeekday(weekday.value)"
              >
                {{ weekday.label }}
              </button>
            </div>
          </div>

          <div
            v-if="selectedSchedulePlugin"
            class="rounded-xl border border-stone-200 bg-stone-50 px-4 py-4"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-sm font-medium text-stone-900">
                  {{ selectedSchedulePlugin.installation.displayName }}
                </p>
                <p class="mt-1 text-sm text-stone-500">
                  {{ selectedSchedulePlugin.manifest.description }}
                </p>
              </div>
              <span class="text-xs text-stone-500">
                {{ selectedSchedulePlugin.installation.runtimeType === "node" ? "Node" : "Python" }}
              </span>
            </div>

            <div
              v-if="selectedSchedulePlugin.manifest.scheduleConfigSchema.length === 0"
              class="mt-4 rounded-lg border border-dashed border-stone-200 bg-white px-4 py-3 text-sm text-stone-500"
            >
              这个插件没有本次任务的额外参数。
            </div>

            <div v-else class="mt-4 grid gap-4 md:grid-cols-2">
              <div
                v-for="field in selectedSchedulePlugin.manifest.scheduleConfigSchema"
                :key="field.key"
                class="block"
                :class="{ 'md:col-span-2': field.type === 'textarea' }"
              >
                <span class="mb-2 block text-sm font-medium text-stone-900">
                  {{ field.label }}
                  <span v-if="field.required" class="text-rose-500">*</span>
                </span>

                <textarea
                  v-if="field.type === 'textarea'"
                  :value="String(scheduleFieldValue(field) ?? '')"
                  rows="4"
                  :placeholder="field.description || ''"
                  class="w-full rounded-xl border border-stone-200 bg-white px-4 py-3 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                  @input="handleScheduleFieldInput(field, $event)"
                />

                <select
                  v-else-if="field.type === 'select'"
                  :value="String(scheduleFieldValue(field) ?? '')"
                  class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                  @change="handleScheduleFieldSelect(field, $event)"
                >
                  <option
                    v-for="option in field.options ?? []"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </option>
                </select>

                <label
                  v-else-if="field.type === 'checkbox'"
                  class="flex items-center gap-3 rounded-xl border border-stone-200 bg-white px-4 py-3"
                >
                  <input
                    :checked="Boolean(scheduleFieldValue(field))"
                    type="checkbox"
                    class="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-stone-900"
                    @change="handleScheduleFieldCheckbox(field, $event)"
                  />
                  <span class="text-sm text-stone-700">
                    {{ field.description || "启用此选项" }}
                  </span>
                </label>

                <input
                  v-else
                  :value="scheduleFieldValue(field)"
                  :type="
                    field.type === 'secret'
                      ? 'password'
                      : field.type === 'number'
                        ? 'number'
                        : field.type === 'url'
                          ? 'url'
                          : 'text'
                  "
                  :placeholder="field.description || ''"
                  autocomplete="off"
                  class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                  @input="handleScheduleFieldInput(field, $event)"
                />

                <span v-if="field.description" class="mt-2 block text-xs text-stone-500">
                  {{ field.description }}
                </span>
              </div>
            </div>
          </div>
        </template>

        <template v-else>
          <label class="block">
            <span class="mb-2 block text-sm font-medium text-stone-900">内容来源</span>
            <input
              v-model="scheduleSource"
              type="text"
              placeholder="例如：手动创建"
              class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
            />
          </label>

          <label class="block">
            <span class="mb-2 block text-sm font-medium text-stone-900">执行时间</span>
            <input
              v-model="scheduleTime"
              type="text"
              placeholder="例如：每天 19:30"
              class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
            />
          </label>
        </template>

        <label class="block">
          <span class="mb-2 block text-sm font-medium text-stone-900">发送设备</span>
          <select
            v-model="scheduleDeviceId"
            class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
          >
            <option
              v-for="device in workspaceStore.devices.filter(
                (device) => device.status !== 'offline',
              )"
              :key="device.id"
              :value="device.id"
            >
              {{ device.name }}
            </option>
          </select>
        </label>

        <p v-if="scheduleError" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {{ scheduleError }}
        </p>

        <div class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <button
            type="button"
            class="ui-btn-secondary px-4 py-2 text-sm"
            @click="closeScheduleDialog"
          >
            取消
          </button>
          <button
            type="submit"
            class="ui-btn-primary px-4 py-2 text-sm"
            :disabled="workspaceStore.isAuthenticated && connectedPlugins.length === 0"
          >
            创建任务
          </button>
        </div>
      </form>
    </AppDialog>
  </section>
</template>

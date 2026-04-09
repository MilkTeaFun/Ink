<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink } from "vue-router";

import AppDialog from "@/components/AppDialog.vue";
import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();

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
const scheduleDeviceId = ref("");
const scheduleError = ref("");

function handlePrintDeviceChange(jobId: string, event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.updatePrintDevice(jobId, target?.value ?? workspaceStore.defaultDeviceId);
}

function handleScheduleDeviceChange(scheduleId: string, event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.updateScheduleDevice(scheduleId, target?.value ?? workspaceStore.defaultDeviceId);
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
  scheduleDeviceId.value = workspaceStore.defaultDeviceId;
  scheduleError.value = "";
}

function closeScheduleDialog() {
  scheduleDialogOpen.value = false;
}

function submitScheduleDialog() {
  scheduleError.value = "";

  if (!scheduleTitle.value.trim()) {
    scheduleError.value = "请输入任务名称。";
    return;
  }

  if (!scheduleTime.value.trim()) {
    scheduleError.value = "请输入执行时间。";
    return;
  }

  if (!(scheduleDeviceId.value || workspaceStore.defaultDeviceId)) {
    scheduleError.value = "请先绑定设备，再创建定时任务。";
    return;
  }

  workspaceStore.createSchedule({
    title: scheduleTitle.value,
    source: scheduleSource.value || "手动创建",
    timeLabel: scheduleTime.value,
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
                      class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium"
                      :class="
                        item.status === 'pending'
                          ? 'bg-stone-100 text-stone-700 ring-1 ring-stone-500/10 ring-inset'
                          : 'bg-amber-50 text-amber-700 ring-1 ring-amber-600/20 ring-inset'
                      "
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
            v-if="workspaceStore.schedules.length === 0"
            class="rounded-2xl border border-dashed border-stone-200 bg-stone-50 px-6 py-10 text-center"
          >
            <h4 class="text-base font-semibold text-stone-900">还没有定时任务</h4>
          </div>

          <div v-else class="ui-list-card">
            <article
              v-for="task in workspaceStore.schedules"
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
                </div>

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
                <div
                  class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between sm:gap-3"
                >
                  <div class="min-w-0">
                    <p class="truncate text-sm font-medium text-stone-900">{{ item.title }}</p>
                    <p class="mt-0.5 text-sm text-stone-500">
                      {{ workspaceStore.getDeviceName(item.deviceId) }} ·
                      {{ workspaceStore.formatPrintTime(item.updatedAt) }}
                    </p>
                  </div>
                  <span
                    class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium"
                    :class="
                      item.status === 'completed'
                        ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20 ring-inset'
                        : item.status === 'queued'
                          ? 'bg-amber-50 text-amber-700 ring-1 ring-amber-600/20 ring-inset'
                          : item.status === 'cancelled'
                            ? 'bg-stone-100 text-stone-700 ring-1 ring-stone-500/10 ring-inset'
                            : item.status === 'failed'
                              ? 'bg-rose-50 text-rose-700 ring-1 ring-rose-600/20 ring-inset'
                              : 'bg-stone-100 text-stone-700 ring-1 ring-stone-500/10 ring-inset'
                    "
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
            <article v-for="source in workspaceStore.sources" :key="source.id" class="ui-list-row">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-stone-900">{{ source.name }}</p>
                  <p class="mt-0.5 text-sm text-stone-500">{{ source.type }} · {{ source.note }}</p>
                </div>
                <span
                  class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset"
                  :class="
                    source.status === 'connected'
                      ? 'bg-emerald-50 text-emerald-700 ring-emerald-600/20'
                      : source.status === 'error'
                        ? 'bg-rose-50 text-rose-700 ring-rose-600/20'
                        : 'bg-stone-100 text-stone-700 ring-stone-500/10'
                  "
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
      description="创建一条自动打印计划。"
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
          <button type="submit" class="ui-btn-primary px-4 py-2 text-sm">创建任务</button>
        </div>
      </form>
    </AppDialog>
  </section>
</template>

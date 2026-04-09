<script setup lang="ts">
import { ref } from "vue";
import { RouterLink } from "vue-router";

import AppDialog from "@/components/AppDialog.vue";
import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();
const addDeviceOpen = ref(false);
const deviceName = ref("");
const deviceNote = ref("");
const deviceIdentifier = ref("");
const setAsDefault = ref(false);
const addDeviceError = ref("");

function openAddDeviceDialog() {
  addDeviceOpen.value = true;
  addDeviceError.value = "";
  deviceName.value = "";
  deviceNote.value = "";
  deviceIdentifier.value = "";
  setAsDefault.value = false;
}

function closeAddDeviceDialog() {
  addDeviceOpen.value = false;
}

async function submitAddDevice() {
  addDeviceError.value = "";

  if (!deviceName.value.trim()) {
    addDeviceError.value = "请输入设备名称。";
    return;
  }

  if (workspaceStore.isAuthenticated && !deviceIdentifier.value.trim()) {
    addDeviceError.value = "请输入咕咕机设备编号。";
    return;
  }

  const created = await workspaceStore.addDevice({
    name: deviceName.value,
    note: deviceNote.value,
    deviceId: deviceIdentifier.value,
    setAsDefault: setAsDefault.value,
  });
  if (!created) {
    addDeviceError.value =
      workspaceStore.flashTone === "error" ? workspaceStore.flashMessage : "绑定设备失败。";
    return;
  }
  closeAddDeviceDialog();
}
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-6 pt-4 sm:space-y-8">
    <div>
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">状态</h2>
    </div>

    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <article
        v-for="item in workspaceStore.summaryCards"
        :key="item.label"
        class="rounded-xl border border-stone-200 bg-white p-5 shadow-sm"
      >
        <p class="text-xs font-medium text-stone-500">{{ item.label }}</p>
        <div class="mt-3 flex items-end justify-between gap-3">
          <p class="text-2xl font-semibold text-stone-900">
            {{ item.value }}
          </p>
          <div class="mb-1.5 h-1.5 w-16 overflow-hidden rounded-full bg-stone-100">
            <div
              class="h-full rounded-full transition-all duration-500"
              :class="
                item.tone === 'green'
                  ? 'bg-emerald-500'
                  : item.tone === 'amber'
                    ? 'bg-amber-500'
                    : item.tone === 'stone'
                      ? 'bg-stone-400'
                      : 'bg-stone-800'
              "
              :style="{ width: `${item.progress}%` }"
            />
          </div>
        </div>
      </article>
    </div>

    <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div class="space-y-8">
        <section>
          <div
            v-if="workspaceStore.printerSyncError"
            class="mb-4 rounded-2xl border border-amber-200 bg-amber-50 px-5 py-4"
          >
            <p class="text-sm font-medium text-amber-900">设备同步异常</p>
            <p class="mt-1 text-sm text-amber-700">{{ workspaceStore.printerSyncError }}</p>
          </div>

          <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex flex-wrap items-center gap-3">
              <h3 class="text-base leading-6 font-semibold text-stone-900">已绑定设备</h3>
              <RouterLink to="/tutorial" class="text-sm text-stone-500 hover:text-stone-900">
                查看教程
              </RouterLink>
            </div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="ui-btn-secondary px-3 py-1.5 text-sm"
                @click="openAddDeviceDialog"
              >
                添加设备
              </button>
            </div>
          </div>

          <div
            v-if="workspaceStore.devices.length === 0"
            class="rounded-2xl border border-dashed border-stone-200 bg-stone-50 px-6 py-10 text-center"
          >
            <h4 class="text-base font-semibold text-stone-900">还没有绑定任何咕咕机</h4>
            <p class="mt-2 text-sm text-stone-500">
              {{
                workspaceStore.isAuthenticated
                  ? "先去教程页看一眼，再回来填写设备编号完成真实绑定。"
                  : "当前未登录时展示的是演示工作区；登录后，这里会切到你账号自己的真实设备列表。"
              }}
            </p>
            <RouterLink
              to="/tutorial"
              class="ui-btn-secondary mt-4 inline-flex px-3 py-1.5 text-sm"
            >
              打开教程
            </RouterLink>
          </div>

          <div v-else class="ui-list-card">
            <article
              v-for="device in workspaceStore.devices"
              :key="device.id"
              class="ui-list-row flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div class="flex min-w-0 items-center gap-3">
                <div
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-stone-100 text-sm font-semibold text-stone-700"
                >
                  {{ device.name.slice(0, 1) }}
                </div>
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium text-stone-900">{{ device.name }}</p>
                  <p class="mt-0.5 text-sm text-stone-500">
                    {{ device.id === workspaceStore.defaultDeviceId ? "默认设备 · " : "" }}
                    {{
                      device.status === "offline" && !device.note
                        ? "已解绑，仅保留历史记录"
                        : device.note
                    }}
                  </p>
                </div>
              </div>

              <div class="flex w-full flex-wrap items-center gap-2 sm:w-auto sm:justify-end">
                <span
                  class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium"
                  :class="
                    device.status === 'connected'
                      ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20 ring-inset'
                      : device.status === 'pending'
                        ? 'bg-amber-50 text-amber-700 ring-1 ring-amber-600/20 ring-inset'
                        : 'bg-stone-100 text-stone-700 ring-1 ring-stone-500/10 ring-inset'
                  "
                >
                  {{ workspaceStore.getDeviceStatusLabel(device.status) }}
                </span>
                <button
                  v-if="device.id !== workspaceStore.defaultDeviceId && device.status !== 'offline'"
                  type="button"
                  class="ui-btn-secondary px-3 py-1.5 text-sm"
                  @click="workspaceStore.setDefaultDevice(device.id)"
                >
                  设为默认
                </button>
                <button
                  v-if="device.status !== 'offline'"
                  type="button"
                  class="ui-btn-secondary px-3 py-1.5 text-sm"
                  @click="workspaceStore.removeDevice(device.id)"
                >
                  {{ device.status === "pending" ? "移除" : "解绑" }}
                </button>
              </div>
            </article>
          </div>
        </section>

        <section>
          <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">自动打印</h3>
            </div>
            <RouterLink to="/prints" class="ui-btn-secondary px-3 py-1.5 text-sm">
              前往打印
            </RouterLink>
          </div>

          <div
            v-if="workspaceStore.schedules.length === 0"
            class="rounded-2xl border border-dashed border-stone-200 bg-stone-50 px-6 py-10 text-center"
          >
            <h4 class="text-base font-semibold text-stone-900">还没有自动打印计划</h4>
          </div>

          <div v-else class="ui-list-card">
            <article
              v-for="task in workspaceStore.schedules"
              :key="task.id"
              class="ui-list-row flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div class="min-w-0">
                <p class="text-sm font-medium text-stone-900">{{ task.title }}</p>
                <p class="mt-0.5 text-sm text-stone-500">
                  {{ task.source }} · {{ task.timeLabel }} ·
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
            </article>
          </div>
        </section>
      </div>

      <aside>
        <div class="mb-4">
          <div>
            <h3 class="text-base leading-6 font-semibold text-stone-900">最近状态</h3>
          </div>
        </div>

        <div class="ui-list-card p-4">
          <div class="ui-timeline">
            <article
              v-for="item in workspaceStore.recentPrintJobs"
              :key="item.id"
              class="ui-timeline-item"
            >
              <div class="flex items-start justify-between gap-3">
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
      </aside>
    </div>

    <AppDialog
      :open="addDeviceOpen"
      title="添加设备"
      :description="
        workspaceStore.isAuthenticated
          ? '登录后会把设备真实绑定到当前账号下，并可继续设为默认或解绑。'
          : '当前为演示模式，添加后只会保存在本地示例数据里。'
      "
      @close="closeAddDeviceDialog"
    >
      <form class="space-y-4" @submit.prevent="submitAddDevice">
        <label class="block">
          <span class="mb-2 block text-sm font-medium text-stone-900">设备名称</span>
          <input
            v-model="deviceName"
            type="text"
            placeholder="例如：客厅咕咕机"
            class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
          />
        </label>

        <label class="block">
          <span class="mb-2 block text-sm font-medium text-stone-900">设备备注</span>
          <input
            v-model="deviceNote"
            type="text"
            placeholder="例如：窗边打印机"
            class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
          />
        </label>

        <label v-if="workspaceStore.isAuthenticated" class="block">
          <span class="mb-2 block text-sm font-medium text-stone-900">咕咕机设备编号</span>
          <input
            v-model="deviceIdentifier"
            type="text"
            placeholder="例如：xxxxxx"
            class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
          />
        </label>

        <label
          class="flex items-center gap-3 rounded-2xl border border-stone-200 bg-stone-50 px-4 py-3"
        >
          <input
            v-model="setAsDefault"
            type="checkbox"
            class="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-stone-900"
          />
          <span class="text-sm text-stone-900">设为默认设备</span>
        </label>

        <p v-if="addDeviceError" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {{ addDeviceError }}
        </p>

        <div class="flex justify-end gap-3">
          <button
            type="button"
            class="ui-btn-secondary px-4 py-2 text-sm"
            @click="closeAddDeviceDialog"
          >
            取消
          </button>
          <button type="submit" class="ui-btn-primary px-4 py-2 text-sm">添加设备</button>
        </div>
      </form>
    </AppDialog>
  </section>
</template>

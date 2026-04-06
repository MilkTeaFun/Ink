<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";

import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();

const defaultSettings = computed(() => [
  { label: "默认设备", value: workspaceStore.activeDeviceLabel || "暂未设置" },
  { label: "发送前确认", value: workspaceStore.sendConfirmationEnabled ? "已开启" : "已关闭" },
  { label: "纸条风格", value: workspaceStore.activeNoteStyle },
]);

function handlePrintDeviceChange(jobId: string, event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.updatePrintDevice(jobId, target?.value ?? workspaceStore.defaultDeviceId);
}

function handleScheduleDeviceChange(scheduleId: string, event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.updateScheduleDevice(scheduleId, target?.value ?? workspaceStore.defaultDeviceId);
}
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 px-4 pt-4 pb-24 sm:px-0 lg:pb-12">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h2 class="text-2xl font-semibold tracking-tight text-stone-900">打印</h2>
        <p class="mt-1 text-sm text-stone-500">待确认内容、定时任务和打印记录都集中放在这里。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <button
          class="ui-btn-primary px-3 py-1.5 text-sm"
          @click="workspaceStore.createManualPrint"
        >
          新建打印
        </button>
        <button class="ui-btn-secondary px-3 py-1.5 text-sm" @click="workspaceStore.createSchedule">
          新建定时任务
        </button>
      </div>
    </div>

    <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div class="space-y-8">
        <section>
          <div class="mb-4">
            <h3 class="text-base leading-6 font-semibold text-stone-900">待处理打印</h3>
            <p class="mt-1 text-sm text-stone-500">这里会收纳待确认和已进队列的打印任务。</p>
          </div>

          <div
            v-if="workspaceStore.pendingPrintJobs.length === 0"
            class="rounded-2xl border border-dashed border-stone-200 bg-stone-50 px-6 py-10 text-center"
          >
            <h4 class="text-base font-semibold text-stone-900">当前没有待处理打印</h4>
            <p class="mt-2 text-sm text-stone-500">
              去对话页生成一条回答，或者直接手动新建一张纸条。
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

                <div class="flex flex-wrap gap-2">
                  <button
                    v-if="item.status === 'pending'"
                    class="ui-btn-primary px-3 py-1.5 text-sm"
                    @click="workspaceStore.confirmPrint(item.id)"
                  >
                    确认打印
                  </button>
                  <button v-else class="ui-btn-secondary px-3 py-1.5 text-sm" disabled>
                    等待完成
                  </button>
                </div>
              </div>

              <div class="flex flex-col gap-2 md:flex-row md:items-center">
                <label class="text-sm font-medium text-stone-700">目标设备</label>
                <select
                  :value="item.deviceId"
                  class="rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900"
                  @change="handlePrintDeviceChange(item.id, $event)"
                >
                  <option
                    v-for="device in workspaceStore.devices"
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
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">定时任务</h3>
              <p class="mt-1 text-sm text-stone-500">这里可以启停自动打印计划，并修改目标设备。</p>
            </div>
            <div class="flex items-center gap-3">
              <span class="text-xs text-stone-500">模板库筹备中</span>
              <button class="ui-btn-secondary px-3 py-1.5 text-sm" disabled>管理模板</button>
            </div>
          </div>

          <div
            v-if="workspaceStore.schedules.length === 0"
            class="rounded-2xl border border-dashed border-stone-200 bg-stone-50 px-6 py-10 text-center"
          >
            <h4 class="text-base font-semibold text-stone-900">还没有定时任务</h4>
            <p class="mt-2 text-sm text-stone-500">
              先创建一个自动打印计划，再回来调整时间和设备。
            </p>
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
                  class="rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900"
                  @change="handleScheduleDeviceChange(task.id, $event)"
                >
                  <option
                    v-for="device in workspaceStore.devices"
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
            <p class="mt-1 text-sm text-stone-500">
              查看最近完成的任务，确认有没有卡住或失败的记录。
            </p>
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
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">默认打印设置</h3>
              <p class="mt-1 text-sm text-stone-500">新建打印和对话页会优先套用这些配置。</p>
            </div>
            <RouterLink to="/settings" class="ui-btn-secondary px-3 py-1.5 text-sm"
              >调整</RouterLink
            >
          </div>

          <div class="ui-list-card">
            <div v-for="item in defaultSettings" :key="item.label" class="ui-list-row">
              <p class="text-sm font-medium text-stone-900">{{ item.label }}</p>
              <p class="mt-1 text-sm text-stone-500">{{ item.value }}</p>
            </div>
          </div>
        </section>

        <section>
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">内容来源</h3>
              <p class="mt-1 text-sm text-stone-500">
                直接在这里切换来源状态，模拟授权和异常恢复。
              </p>
            </div>
            <RouterLink to="/settings" class="ui-btn-secondary px-3 py-1.5 text-sm"
              >更多设置</RouterLink
            >
          </div>

          <div class="ui-list-card">
            <article
              v-for="source in workspaceStore.sources"
              :key="source.id"
              class="ui-list-row grid gap-3"
            >
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
              <button
                class="ui-btn-secondary px-3 py-1.5 text-sm"
                @click="workspaceStore.cycleSourceStatus(source.id)"
              >
                切换状态
              </button>
            </article>
          </div>
        </section>
      </aside>
    </div>
  </section>
</template>

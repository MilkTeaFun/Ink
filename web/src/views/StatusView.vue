<script setup lang="ts">
import { RouterLink } from "vue-router";

import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 px-4 pt-4 pb-24 sm:px-0 lg:pb-12">
    <div>
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">状态</h2>
      <p class="mt-1 text-sm text-stone-500">
        设备、任务和最近打印记录会从同一份共享状态里实时更新。
      </p>
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
          <div class="mb-4">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">已绑定设备</h3>
              <p class="mt-1 text-sm text-stone-500">
                默认设备、离线设备和待绑定设备都显示在这里。
              </p>
            </div>
          </div>

          <div class="ui-list-card">
            <article
              v-for="device in workspaceStore.devices"
              :key="device.id"
              class="ui-list-row flex items-center justify-between gap-3"
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
                    {{ device.id === workspaceStore.defaultDeviceId ? "默认设备 · " : ""
                    }}{{ device.note }}
                  </p>
                </div>
              </div>

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
            </article>
          </div>
        </section>

        <section>
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">自动打印</h3>
              <p class="mt-1 text-sm text-stone-500">
                状态页保留启停，创建和详细调整集中到打印页。
              </p>
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
            <p class="mt-2 text-sm text-stone-500">去打印页创建第一条定时任务。</p>
          </div>

          <div v-else class="ui-list-card">
            <article
              v-for="task in workspaceStore.schedules"
              :key="task.id"
              class="ui-list-row flex items-center justify-between gap-4"
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
            <p class="mt-1 text-sm text-stone-500">
              打印队列、完成记录和失败状态都从同一处读出来。
            </p>
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
  </section>
</template>

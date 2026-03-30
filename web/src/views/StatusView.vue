<script setup lang="ts">
import { ref } from "vue";

const summary = [
  { label: "已绑定设备", value: "2 台", tone: "neutral", progress: 100 },
  { label: "已启用任务", value: "2 条", tone: "amber", progress: 64 },
  { label: "待确认打印", value: "2 条", tone: "stone", progress: 42 },
  { label: "今日完成", value: "18 条", tone: "green", progress: 92 },
];

const devices = [
  {
    name: "书桌咕咕机",
    status: "已连接",
    note: "默认设备",
  },
  {
    name: "卧室咕咕机",
    status: "待绑定",
    note: "睡前提醒",
  },
];

const schedules = ref([
  {
    title: "早报摘要",
    source: "晨间订阅",
    time: "每天 08:00",
    enabled: true,
  },
  {
    title: "晚安提醒",
    source: "睡前便签",
    time: "每天 22:00",
    enabled: true,
  },
  {
    title: "周末清单",
    source: "家庭计划",
    time: "周六 09:30",
    enabled: false,
  },
]);

const prints = [
  {
    title: "今日待办",
    device: "书桌咕咕机",
    time: "14:12",
    status: "已完成",
  },
  {
    title: "购物清单",
    device: "书桌咕咕机",
    time: "13:47",
    status: "打印中",
  },
  {
    title: "晚安留言",
    device: "卧室咕咕机",
    time: "昨天",
    status: "待确认",
  },
];
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 px-4 pt-4 pb-24 sm:px-0 lg:pb-12">
    <div>
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">状态</h2>
      <p class="mt-1 text-sm text-stone-500">这里只看设备是否在线、任务是否启用。</p>
    </div>

    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <article
        v-for="item in summary"
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
              <p class="mt-1 text-sm text-stone-500">先确认哪台设备在待命。</p>
            </div>
          </div>

          <div class="ui-list-card">
            <article
              v-for="device in devices"
              :key="device.name"
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
                  <p class="mt-0.5 text-sm text-stone-500">{{ device.note }}</p>
                </div>
              </div>

              <span
                class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium"
                :class="
                  device.status === '已连接'
                    ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20 ring-inset'
                    : 'bg-amber-50 text-amber-700 ring-1 ring-amber-600/20 ring-inset'
                "
              >
                {{ device.status }}
              </span>
            </article>
          </div>
        </section>

        <section>
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">自动打印</h3>
              <p class="mt-1 text-sm text-stone-500">这里只保留启停，创建和编辑请去打印页。</p>
            </div>
            <RouterLink to="/prints" class="ui-btn-secondary px-3 py-1.5 text-sm">
              前往打印
            </RouterLink>
          </div>

          <div class="ui-list-card">
            <article
              v-for="task in schedules"
              :key="`${task.title}-${task.time}`"
              class="ui-list-row flex items-center justify-between gap-4"
            >
              <div class="min-w-0">
                <p class="text-sm font-medium text-stone-900">{{ task.title }}</p>
                <p class="mt-0.5 text-sm text-stone-500">{{ task.source }} · {{ task.time }}</p>
              </div>

              <button
                type="button"
                class="ui-toggle"
                :class="{ 'is-on': task.enabled }"
                :aria-label="`${task.enabled ? '关闭' : '开启'}${task.title}`"
                :aria-pressed="task.enabled"
                @click="task.enabled = !task.enabled"
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
            <p class="mt-1 text-sm text-stone-500">看一下最近有没有卡住或待确认的任务。</p>
          </div>
        </div>

        <div class="ui-list-card p-4">
          <div class="ui-timeline">
            <article
              v-for="item in prints"
              :key="`${item.title}-${item.time}`"
              class="ui-timeline-item"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium text-stone-900">{{ item.title }}</p>
                  <p class="mt-0.5 text-sm text-stone-500">{{ item.device }} · {{ item.time }}</p>
                </div>
                <span
                  class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium"
                  :class="
                    item.status === '已完成'
                      ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20 ring-inset'
                      : item.status === '打印中'
                        ? 'bg-amber-50 text-amber-700 ring-1 ring-amber-600/20 ring-inset'
                        : 'bg-stone-100 text-stone-700 ring-1 ring-stone-500/10 ring-inset'
                  "
                >
                  {{ item.status }}
                </span>
              </div>
            </article>
          </div>
        </div>
      </aside>
    </div>
  </section>
</template>

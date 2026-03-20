<script setup lang="ts">
const summary = [
  { label: "已绑定设备", value: "2 台", tone: "neutral", progress: 100 },
  { label: "执行中任务", value: "3 条", tone: "amber", progress: 64 },
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


const schedules = [
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
];


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
  <section class="space-y-6 pb-24 lg:pb-0">
    <div class="space-y-2">
      <p class="text-[0.62rem] uppercase tracking-[0.24em] text-stone-500">状态</p>
      <h2 class="text-[1.95rem] font-semibold tracking-[-0.05em] text-stone-950">
        设备、任务和打印记录都在这里。
      </h2>
    </div>

    <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
      <article
        v-for="item in summary"
        :key="item.label"
        class="bg-white/72 rounded-[1.35rem] border border-white/60 px-4 py-4 backdrop-blur"
      >
        <p class="text-[0.62rem] uppercase tracking-[0.22em] text-stone-500">{{ item.label }}</p>
        <div class="mt-2 flex items-end justify-between gap-3">
          <p class="text-[1.9rem] font-semibold leading-none tracking-[-0.05em] text-stone-950">
            {{ item.value }}
          </p>
          <div class="mb-1 h-1.5 w-16 overflow-hidden rounded-full bg-stone-200/70">
            <div
              class="h-full rounded-full"
              :class="
                item.tone === 'green'
                  ? 'bg-emerald-500'
                  : item.tone === 'amber'
                    ? 'bg-amber-500'
                    : item.tone === 'stone'
                      ? 'bg-stone-500'
                      : 'bg-stone-800'
              "
              :style="{ width: `${item.progress}%` }"
            />
          </div>
        </div>
      </article>
    </div>

    <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div class="space-y-4">
        <section class="ui-panel">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="text-[0.62rem] uppercase tracking-[0.22em] text-stone-500">已绑定设备</p>
              <h3 class="mt-2 text-[1.65rem] font-semibold tracking-[-0.05em] text-stone-950">
                先确认哪台设备在待命。
              </h3>
            </div>
            <button class="ui-btn-secondary px-3 py-2 text-xs">管理设备</button>
          </div>

          <div class="ui-list-card mt-4">
            <article
              v-for="device in devices"
              :key="device.name"
              class="ui-list-row flex items-center justify-between gap-3"
            >
              <div class="flex min-w-0 items-center gap-3">
                <div
                  class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-[#f6ede3] text-xs font-semibold text-[#6d4d31]"
                >
                  {{ device.name.slice(0, 1) }}
                </div>
                <div class="min-w-0">
                  <p class="truncate text-sm font-semibold text-stone-950">{{ device.name }}</p>
                  <p class="mt-1 text-xs text-stone-500">{{ device.note }}</p>
                </div>
              </div>

              <span
                class="inline-flex rounded-full px-3 py-1 text-[0.7rem] font-semibold"
                :class="
                  device.status === '已连接'
                    ? 'bg-emerald-100 text-emerald-700'
                    : 'bg-amber-100 text-amber-700'
                "
              >
                {{ device.status }}
              </span>
            </article>
          </div>
        </section>

        <section class="ui-panel">
          <div>
            <p class="text-[0.62rem] uppercase tracking-[0.22em] text-stone-500">定时任务</p>
            <h3 class="mt-2 text-[1.65rem] font-semibold tracking-[-0.05em] text-stone-950">
              哪些内容会按计划自动准备。
            </h3>
          </div>

          <div class="mt-4 space-y-4">
            <article
              v-for="task in schedules"
              :key="`${task.title}-${task.time}`"
              class="border-stone-900/8 flex items-center justify-between gap-4 border-b pb-4 last:border-b-0 last:pb-0"
            >
              <div class="min-w-0">
                <p class="text-sm font-semibold text-stone-950">{{ task.title }}</p>
                <p class="mt-1 text-xs text-stone-500">{{ task.source }} · {{ task.time }}</p>
              </div>

              <button
                class="relative h-6 w-11 shrink-0 rounded-full transition"
                :class="task.enabled ? 'bg-stone-700' : 'bg-stone-300'"
                :aria-pressed="task.enabled"
                type="button"
              >
                <span
                  class="absolute top-1 h-4 w-4 rounded-full bg-white transition"
                  :class="task.enabled ? 'left-6' : 'left-1'"
                />
              </button>
            </article>
          </div>
        </section>
      </div>

      <aside class="ui-panel">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-[0.62rem] uppercase tracking-[0.22em] text-stone-500">打印记录</p>
            <h3 class="mt-2 text-[1.65rem] font-semibold tracking-[-0.05em] text-stone-950">
              最近状态
            </h3>
          </div>
          <button class="ui-btn-secondary px-3 py-2 text-xs">筛选</button>
        </div>

        <div class="ui-list-card mt-4">
          <div class="ui-timeline">
            <article
              v-for="item in prints"
              :key="`${item.title}-${item.time}`"
              class="ui-timeline-item"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate text-sm font-semibold text-stone-950">{{ item.title }}</p>
                  <p class="mt-1 text-xs text-stone-500">{{ item.device }} · {{ item.time }}</p>
                </div>
                <span
                  class="inline-flex rounded-full px-2.5 py-1 text-[0.68rem] font-semibold"
                  :class="
                    item.status === '已完成'
                      ? 'bg-lime-100 text-lime-700'
                      : item.status === '打印中'
                        ? 'bg-amber-100 text-amber-700'
                        : 'bg-stone-200 text-stone-700'
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

<script setup lang="ts">
const summary = [
  { label: "已绑定设备", value: "2 台" },
  { label: "执行中任务", value: "3 条" },
  { label: "待确认打印", value: "2 条" },
  { label: "今日完成", value: "18 条" },
];


const devices = [
  {
    name: "书桌咕咕机",
    status: "已连接",
    note: "默认设备",
  },
  {
    name: "卧室咕咕机",
    status: "等待绑定",
    note: "睡前提醒",
  },
];


const schedules = [
  {
    title: "早报摘要",
    source: "晨间订阅",
    time: "每天 08:00",
    status: "已开启",
  },
  {
    title: "晚安提醒",
    source: "睡前便签",
    time: "每天 22:00",
    status: "已开启",
  },
  {
    title: "周末清单",
    source: "家庭计划",
    time: "周六 09:30",
    status: "待确认",
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
  <section class="space-y-8 pb-24 lg:pb-0">
    <div>
      <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">状态</p>
      <h2 class="mt-2 text-3xl text-stone-950">设备、任务和打印记录都在这里。</h2>
    </div>

    <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
      <article v-for="item in summary" :key="item.label" class="min-w-0">
        <div class="bg-white/72 rounded-[1.4rem] border border-white/55 px-4 py-4 backdrop-blur">
          <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">{{ item.label }}</p>
          <p class="mt-3 text-3xl font-semibold text-stone-950">{{ item.value }}</p>
        </div>
      </article>
    </div>

    <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_340px]">
      <div class="space-y-8">
        <section class="ui-panel">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">已绑定设备</p>
              <h3 class="mt-2 text-2xl font-semibold text-stone-950">先确认哪台设备在待命。</h3>
            </div>
            <button class="ui-btn-secondary px-3 py-2">管理设备</button>
          </div>

          <div class="ui-list-card mt-5">
            <article
              v-for="device in devices"
              :key="device.name"
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-base font-semibold text-stone-950">{{ device.name }}</p>
                <p class="mt-1 text-sm text-stone-500">{{ device.note }}</p>
              </div>
              <span
                class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
                :class="
                  device.status === '已连接'
                    ? 'bg-emerald-100 text-emerald-800'
                    : 'bg-amber-100 text-amber-800'
                "
              >
                {{ device.status }}
              </span>
            </article>
          </div>
        </section>

        <section class="ui-panel">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">定时任务</p>
              <h3 class="mt-2 text-2xl font-semibold text-stone-950">哪些内容会按计划自动准备。</h3>
            </div>
            <button class="ui-btn-secondary px-3 py-2">查看全部</button>
          </div>

          <div class="ui-list-card mt-5">
            <article
              v-for="task in schedules"
              :key="`${task.title}-${task.time}`"
              class="ui-list-row grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-center"
            >
              <div>
                <p class="text-base font-semibold text-stone-950">{{ task.title }}</p>
                <p class="mt-1 text-sm text-stone-500">{{ task.source }} · {{ task.time }}</p>
              </div>
              <span
                class="inline-flex h-fit rounded-full px-3 py-1 text-xs font-semibold"
                :class="
                  task.status === '已开启'
                    ? 'bg-emerald-100 text-emerald-800'
                    : 'bg-stone-200 text-stone-700'
                "
              >
                {{ task.status }}
              </span>
            </article>
          </div>
        </section>
      </div>

      <aside class="ui-panel">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">打印记录</p>
            <h3 class="mt-2 text-2xl font-semibold text-stone-950">最近状态</h3>
          </div>
          <button class="ui-btn-secondary px-3 py-2">筛选</button>
        </div>

        <div class="ui-list-card mt-5">
          <div class="ui-timeline">
            <article
              v-for="item in prints"
              :key="`${item.title}-${item.time}`"
              class="ui-timeline-item"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-base font-semibold text-stone-950">{{ item.title }}</p>
                  <p class="mt-1 text-sm text-stone-500">{{ item.device }} · {{ item.time }}</p>
                </div>
                <span
                  class="inline-flex rounded-full px-3 py-1 text-xs font-semibold"
                  :class="
                    item.status === '已完成'
                      ? 'bg-lime-100 text-lime-800'
                      : item.status === '打印中'
                        ? 'bg-amber-100 text-amber-800'
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

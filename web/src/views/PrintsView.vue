<script setup lang="ts">
import { ref } from "vue";

const pendingPrints = [
  {
    title: "晚安留言",
    source: "对话草稿",
    device: "卧室咕咕机",
    time: "今晚 22:00 前",
    status: "待确认",
  },
  {
    title: "明日早报",
    source: "晨间订阅",
    device: "书桌咕咕机",
    time: "明早 08:00",
    status: "排队中",
  },
];

const scheduledPrints = ref([
  {
    title: "早报摘要",
    source: "晨间订阅",
    time: "每天 08:00",
    device: "书桌咕咕机",
    enabled: true,
  },
  {
    title: "晚安提醒",
    source: "睡前便签",
    time: "每天 22:00",
    device: "卧室咕咕机",
    enabled: true,
  },
  {
    title: "周末清单",
    source: "家庭计划",
    time: "周六 09:30",
    device: "书桌咕咕机",
    enabled: false,
  },
]);

const printHistory = [
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

const defaultSettings = [
  { label: "默认设备", value: "书桌咕咕机" },
  { label: "发送前确认", value: "已开启" },
  { label: "纸条风格", value: "简洁清晰" },
];

const sources = [
  { name: "今天值得看", type: "RSS", note: "每日文章摘要" },
  { name: "天气提醒", type: "在线服务", note: "晨间天气简报" },
];
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 px-4 pt-4 pb-24 sm:px-0 lg:pb-12">
    <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h2 class="text-2xl font-semibold tracking-tight text-stone-900">打印</h2>
        <p class="mt-1 text-sm text-stone-500">待确认内容、定时任务和打印记录都集中放在这里。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <button class="ui-btn-primary px-3 py-1.5 text-sm">新建打印</button>
        <button class="ui-btn-secondary px-3 py-1.5 text-sm">新建定时任务</button>
      </div>
    </div>

    <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div class="space-y-8">
        <section>
          <div class="mb-4">
            <h3 class="text-base leading-6 font-semibold text-stone-900">待处理打印</h3>
            <p class="mt-1 text-sm text-stone-500">先处理还没发出的内容，再回看队列状态。</p>
          </div>

          <div class="ui-list-card">
            <article
              v-for="item in pendingPrints"
              :key="`${item.title}-${item.time}`"
              class="ui-list-row flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
            >
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="text-sm font-medium text-stone-900">{{ item.title }}</p>
                  <span
                    class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium"
                    :class="
                      item.status === '待确认'
                        ? 'bg-stone-100 text-stone-700 ring-1 ring-stone-500/10 ring-inset'
                        : 'bg-amber-50 text-amber-700 ring-1 ring-amber-600/20 ring-inset'
                    "
                  >
                    {{ item.status }}
                  </span>
                </div>
                <p class="mt-1 text-sm text-stone-500">
                  {{ item.source }} · {{ item.device }} · {{ item.time }}
                </p>
              </div>

              <div class="flex flex-wrap gap-2">
                <button
                  class="px-3 py-1.5 text-sm"
                  :class="item.status === '待确认' ? 'ui-btn-primary' : 'ui-btn-secondary'"
                >
                  {{ item.status === "待确认" ? "确认打印" : "查看队列" }}
                </button>
                <button class="ui-btn-secondary px-3 py-1.5 text-sm">编辑</button>
              </div>
            </article>
          </div>
        </section>

        <section>
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">定时任务</h3>
              <p class="mt-1 text-sm text-stone-500">在这里创建、编辑和启停自动打印计划。</p>
            </div>
            <button class="ui-btn-secondary px-3 py-1.5 text-sm">管理模板</button>
          </div>

          <div class="ui-list-card">
            <article
              v-for="task in scheduledPrints"
              :key="`${task.title}-${task.time}`"
              class="ui-list-row flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
            >
              <div class="min-w-0">
                <p class="text-sm font-medium text-stone-900">{{ task.title }}</p>
                <p class="mt-1 text-sm text-stone-500">
                  {{ task.source }} · {{ task.time }} · 发往 {{ task.device }}
                </p>
              </div>

              <div class="flex flex-wrap items-center gap-2">
                <button class="ui-btn-secondary px-3 py-1.5 text-sm">编辑</button>
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
              </div>
            </article>
          </div>
        </section>

        <section>
          <div class="mb-4">
            <h3 class="text-base leading-6 font-semibold text-stone-900">最近打印</h3>
            <p class="mt-1 text-sm text-stone-500">回看发出记录，确认有没有卡住的任务。</p>
          </div>

          <div class="ui-list-card p-4">
            <div class="ui-timeline">
              <article
                v-for="item in printHistory"
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
        </section>
      </div>

      <aside class="space-y-8">
        <section>
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">默认打印设置</h3>
              <p class="mt-1 text-sm text-stone-500">打印前会套用这些默认项。</p>
            </div>
            <button class="ui-btn-secondary px-3 py-1.5 text-sm">调整</button>
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
              <p class="mt-1 text-sm text-stone-500">确认自动打印依赖的来源是否正常。</p>
            </div>
            <button class="ui-btn-secondary px-3 py-1.5 text-sm">管理来源</button>
          </div>

          <div class="ui-list-card">
            <article
              v-for="source in sources"
              :key="source.name"
              class="ui-list-row grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-center"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">{{ source.name }}</p>
                <p class="mt-0.5 text-sm text-stone-500">{{ source.type }} · {{ source.note }}</p>
              </div>
              <span
                class="inline-flex items-center rounded-full bg-emerald-50 px-2.5 py-0.5 text-xs font-medium text-emerald-700 ring-1 ring-emerald-600/20 ring-inset"
              >
                已连接
              </span>
            </article>
          </div>
        </section>
      </aside>
    </div>
  </section>
</template>

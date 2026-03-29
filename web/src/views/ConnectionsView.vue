<script setup lang="ts">
const connectedSources = [
  { name: "今天值得看", type: "RSS", status: "已连接", note: "每日文章摘要" },
  { name: "天气提醒", type: "在线服务", status: "待配置", note: "晨间天气简报" },
];


const plannedSources = [
  { name: "RSS 订阅", detail: "订阅博客、新闻、公告，自动整理成适合打印的小纸条。" },
  { name: "日历事件", detail: "把日程、会议和提醒汇总成每天的打印摘要。" },
  { name: "在线表单 / API", detail: "把来自表单或接口的数据做成通知、标签或提醒。" },
];


const accessMethods = [
  { name: "RSS / Atom", detail: "面向持续更新的文章流与订阅内容。" },
  { name: "在线服务授权", detail: "适合天气、日历、待办等第三方平台。" },
  { name: "自定义接口", detail: "把内部系统或脚本结果直接送进打印流程。" },
];
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 px-4 pt-4 pb-24 sm:px-0 lg:pb-12">
    <div>
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">连接</h2>
      <p class="mt-1 text-sm text-stone-500">未来扩展会从这里开始。</p>
    </div>

    <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div class="space-y-8">
        <section>
          <div class="mb-4 flex items-center justify-between">
            <div>
              <h3 class="text-base leading-6 font-semibold text-stone-900">已连接来源</h3>
              <p class="mt-1 text-sm text-stone-500">哪些外部内容已经接进来。</p>
            </div>
            <button class="ui-btn-primary px-3 py-1.5 text-sm">新增连接</button>
          </div>

          <div class="ui-list-card">
            <article
              v-for="source in connectedSources"
              :key="source.name"
              class="ui-list-row grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-center"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">{{ source.name }}</p>
                <p class="mt-0.5 text-sm text-stone-500">{{ source.type }} · {{ source.note }}</p>
              </div>
              <span
                class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
                :class="
                  source.status === '已连接'
                    ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20 ring-inset'
                    : 'bg-stone-100 text-stone-700 ring-1 ring-stone-500/10 ring-inset'
                "
              >
                {{ source.status }}
              </span>
            </article>
          </div>
        </section>

        <section>
          <div class="mb-4">
            <h3 class="text-base leading-6 font-semibold text-stone-900">规划中的连接</h3>
            <p class="mt-1 text-sm text-stone-500">即将支持的更多数据来源。</p>
          </div>

          <div class="ui-list-card">
            <article
              v-for="(item, index) in plannedSources"
              :key="item.name"
              class="ui-list-row flex gap-4"
            >
              <span
                class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-stone-100 text-sm font-semibold text-stone-500"
              >
                {{ String(index + 1).padStart(2, "0") }}
              </span>
              <div class="min-w-0">
                <p class="text-sm font-medium text-stone-900">{{ item.name }}</p>
                <p class="mt-0.5 text-sm text-stone-500">{{ item.detail }}</p>
              </div>
            </article>
          </div>
        </section>
      </div>

      <aside class="space-y-8">
        <section>
          <h3 class="mb-4 text-base leading-6 font-semibold text-stone-900">接入方式</h3>
          <div class="ui-list-card">
            <div v-for="method in accessMethods" :key="method.name" class="ui-list-row">
              <p class="text-sm font-medium text-stone-900">{{ method.name }}</p>
              <p class="mt-0.5 text-sm text-stone-500">{{ method.detail }}</p>
            </div>
          </div>
        </section>
      </aside>
    </div>
  </section>
</template>

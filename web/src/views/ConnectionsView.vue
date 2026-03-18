<script setup lang="ts">
const connectedSources = [
  { name: '今天值得看', type: 'RSS', status: '已连接', note: '每日文章摘要' },
  { name: '天气提醒', type: '在线服务', status: '待配置', note: '晨间天气简报' },
]

const plannedSources = [
  { name: 'RSS 订阅', detail: '订阅博客、新闻、公告，自动整理成适合打印的小纸条。'},
  { name: '日历事件', detail: '把日程、会议和提醒汇总成每天的打印摘要。'},
  { name: '在线表单 / API', detail: '把来自表单或接口的数据做成通知、标签或提醒。'},
]

const accessMethods = [
  { name: 'RSS / Atom', detail: '面向持续更新的文章流与订阅内容。' },
  { name: '在线服务授权', detail: '适合天气、日历、待办等第三方平台。' },
  { name: '自定义接口', detail: '把内部系统或脚本结果直接送进打印流程。' },
]
</script>

<template>
  <section class="space-y-8 pb-24 lg:pb-0">
    <div>
      <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">连接</p>
      <h2 class="mt-2 text-3xl text-stone-950">未来扩展会从这里开始。</h2>
    </div>

    <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div class="space-y-8">
        <section class="ui-panel">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">已连接来源</p>
              <h3 class="mt-2 text-2xl font-semibold text-stone-950">哪些外部内容已经接进来。</h3>
            </div>
            <button class="ui-btn-primary">新增连接</button>
          </div>

          <div class="ui-list-card mt-5">
            <article
              v-for="source in connectedSources"
              :key="source.name"
              class="ui-list-row grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-center"
            >
              <div>
                <p class="text-base font-semibold text-stone-950">{{ source.name }}</p>
                <p class="mt-1 text-sm text-stone-500">{{ source.type }} · {{ source.note }}</p>
              </div>
              <span
                class="inline-flex h-fit rounded-full px-3 py-1 text-xs font-semibold"
                :class="
                  source.status === '已连接'
                    ? 'bg-emerald-100 text-emerald-800'
                    : 'bg-stone-200 text-stone-700'
                "
            >
                {{ source.status }}
              </span>
            </article>
          </div>
        </section>

        <section class="ui-panel">
          <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">规划中的连接</p>
          <div class="ui-list-card mt-5">
            <article
              v-for="(item, index) in plannedSources"
              :key="item.name"
              class="ui-list-row flex gap-4"
            >
              <span class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-stone-900/10 bg-stone-950 text-xs font-semibold tracking-[0.18em] text-stone-50">
                {{ String(index + 1).padStart(2, '0') }}
              </span>
              <div class="min-w-0">
                <p class="text-base font-semibold text-stone-950">{{ item.name }}</p>
                <p class="mt-1 text-sm leading-6 text-stone-500">{{ item.detail }}</p>
              </div>
            </article>
          </div>
        </section>
      </div>

      <aside class="ui-panel">
        <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">接入方式</p>
        <div class="ui-list-card mt-5">
          <div
            v-for="method in accessMethods"
            :key="method.name"
            class="ui-list-row"
          >
            <p class="text-sm font-semibold text-stone-950">{{ method.name }}</p>
            <p class="mt-1 text-sm leading-6 text-stone-500">{{ method.detail }}</p>
          </div>
        </div>
      </aside>
    </div>
  </section>
</template>

<script setup lang="ts">
const chats = [
  { title: '今日待办', preview: '下班前要记得买牛奶和胶带', time: '刚刚', active: true },
  { title: '生日祝福', preview: '想写一句温柔一点的话', time: '10 分钟前', active: false },
  { title: '购物清单', preview: '鸡蛋、吐司、番茄、酸奶', time: '昨天', active: false },
]

const messages = [
  { role: 'user', text: '帮我整理一张温柔一点的今日提醒，适合打印在小纸条上。'},
  { role: 'assistant', text: '当然可以。你可以写成：今天也别太赶，先把最重要的一件事做好，晚一点记得给自己买杯热饮。'},
]

const previewSummary = [
  { label: '内容语气', value: '清楚温柔', note: '适合提醒、祝福和日常短句。' },
  { label: '推荐长度', value: '两段以内', note: '控制在热敏纸上更容易一眼读完。' },
  { label: '当前输出', value: '最新回答', note: '更适合直接发往默认设备。' },
]

const defaultTargets = [
  { label: '默认设备', value: '书桌咕咕机', note: '工作台上的常用打印位。' },
  { label: '默认打印方式', value: '最新回答', note: '不额外选择时会直接取最新一段。' },
]
</script>

<template>
  <section class="space-y-6 pb-24 lg:pb-0">
    <div>
      <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">对话</p>
      <h2 class="mt-2 text-3xl text-stone-950">通过对话整理内容，再决定打印哪一段。</h2>
    </div>

    <section class="space-y-3 lg:hidden">
      <div class="flex items-center justify-between">
        <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">最近对话</p>
        <button class="ui-btn-secondary px-3 py-2">新建</button>
      </div>

      <div class="flex gap-3 overflow-x-auto pb-2">
        <article
          v-for="chat in chats"
          :key="chat.title"
          class="min-w-[220px] rounded-[1.2rem] border px-4 py-4 transition"
          :class="
            chat.active
              ? 'border-stone-900/80 bg-stone-950 text-stone-50'
              : 'border-white/60 bg-white/74 text-stone-900'
          "
        >
          <div class="flex items-start justify-between gap-2">
            <p class="text-sm font-semibold" :class="chat.active ? 'text-stone-50' : 'text-stone-950'">{{ chat.title }}</p>
            <span class="text-xs" :class="chat.active ? 'text-stone-300' : 'text-stone-500'">{{ chat.time }}</span>
          </div>
          <p class="mt-2 text-sm leading-6" :class="chat.active ? 'text-stone-300' : 'text-stone-600'">
            {{ chat.preview }}
          </p>
        </article>
      </div>
    </section>

    <div class="grid gap-8 lg:grid-cols-[280px_minmax(0,1fr)_280px]">
      <aside class="ui-panel hidden min-w-0 lg:block">
        <div class="flex items-center justify-between">
          <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">最近对话</p>
          <button class="ui-btn-secondary px-3 py-2">新建</button>
        </div>

        <div class="ui-list-card mt-5">
          <article
            v-for="chat in chats"
            :key="chat.title"
            class="ui-list-row"
            :class="
              chat.active
                ? 'is-active text-stone-50'
                : 'text-stone-900'
            "
          >
            <div class="flex items-start justify-between gap-2">
              <p class="text-sm font-semibold" :class="chat.active ? 'text-stone-50' : 'text-stone-950'">{{ chat.title }}</p>
              <span class="text-xs" :class="chat.active ? 'text-stone-300' : 'text-stone-500'">{{ chat.time }}</span>
            </div>
            <p class="mt-2 text-sm leading-6" :class="chat.active ? 'text-stone-300' : 'text-stone-600'">
              {{ chat.preview }}
            </p>
          </article>
        </div>
      </aside>

      <div class="ui-panel min-w-0">
        <div class="flex flex-col gap-3 border-b border-stone-900/10 pb-5 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">当前对话</p>
            <h3 class="mt-2 text-2xl font-semibold text-stone-950">先把内容整理顺，再决定打印哪一句。</h3>
          </div>
          <div class="text-sm text-stone-600">
            当前助手：清楚温柔
          </div>
        </div>

        <div class="mt-6">
          <div class="space-y-4">
            <article
              v-for="message in messages"
              :key="message.text"
              class="max-w-[88%] rounded-[1.45rem] px-4 py-3 text-sm leading-7"
              :class="
                message.role === 'user'
                  ? 'ml-auto bg-stone-950 text-stone-50'
                  : 'bg-white text-stone-700 shadow-sm'
              "
            >
              {{ message.text }}
            </article>
          </div>

          <div class="mt-6 flex flex-col gap-2 border-t border-stone-900/10 pt-4 sm:flex-row">
            <button class="ui-btn-primary w-full sm:w-auto">打印整段对话</button>
            <button class="ui-btn-secondary w-full sm:w-auto">打印最新回答</button>
            <button class="ui-btn-secondary w-full sm:w-auto">选择指定回答</button>
          </div>

          <div class="mt-5 border-t border-stone-900/10 pt-4">
            <textarea
              rows="5"
              placeholder="比如：帮我写一张今天的购物清单，语气轻松一点。"
              class="w-full resize-none border-0 bg-transparent text-sm leading-6 text-stone-700 outline-none"
            />
            <div class="mt-3 flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
              <div class="flex flex-wrap gap-2">
                <button class="ui-btn-secondary px-4 py-2">重新生成</button>
                <button class="ui-btn-secondary px-4 py-2">换个语气</button>
              </div>
              <button class="ui-btn-primary w-full sm:w-auto">发送</button>
            </div>
          </div>
        </div>
      </div>

      <aside class="ui-panel space-y-5">
        <article class="border-b border-stone-900/10 pb-5">
          <p class="text-sm font-semibold text-stone-950">预览摘要</p>
          <div class="ui-list-card mt-4">
            <div
              v-for="item in previewSummary"
              :key="item.label"
              class="ui-list-row"
            >
              <p class="text-sm text-stone-500">{{ item.label }}</p>
              <p class="mt-1 text-base font-semibold text-stone-950">{{ item.value }}</p>
              <p class="mt-2 text-sm leading-6 text-stone-500">{{ item.note }}</p>
            </div>
          </div>
        </article>

        <article>
          <p class="text-sm font-semibold text-stone-950">默认目标</p>
          <div class="ui-list-card mt-3">
            <div
              v-for="item in defaultTargets"
              :key="item.label"
              class="ui-list-row"
            >
              <p class="text-sm text-stone-500">{{ item.label }}</p>
              <p class="mt-1 text-base font-semibold text-stone-950">{{ item.value }}</p>
              <p class="mt-2 text-sm leading-6 text-stone-500">{{ item.note }}</p>
            </div>
            <div class="ui-list-row">
              <button class="ui-btn-secondary w-full">先保存草稿</button>
            </div>
          </div>
        </article>
      </aside>
    </div>
  </section>
</template>

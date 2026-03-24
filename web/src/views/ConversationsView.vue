<script setup lang="ts">
const chats = [
  { title: "今日待办", preview: "下班前要记得买牛奶和胶带", time: "刚刚", active: true },
  { title: "生日祝福", preview: "想写一句温柔一点的话", time: "10 分钟前", active: false },
  { title: "购物清单", preview: "鸡蛋、吐司、番茄、酸奶", time: "昨天", active: false },
];


const messages = [
  { role: "user", text: "帮我整理一张温柔一点的今日提醒，适合打印在小纸条上。" },
  {
    role: "assistant",
    text: "当然可以。你可以写成：今天也别太赶，先把最重要的一件事做好，晚一点记得给自己买杯热饮。",
  },
];
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 px-4 pb-24 pt-4 sm:px-0 lg:pb-12">
    <div>
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">对话</h2>
      <p class="mt-1 text-sm text-stone-500">通过对话整理内容，再决定打印哪一段。</p>
    </div>

    <section class="space-y-4 lg:hidden">
      <div class="flex items-center justify-between">
        <h3 class="text-base font-semibold leading-6 text-stone-900">最近对话</h3>
        <button class="ui-btn-secondary px-3 py-1.5 text-sm">新建</button>
      </div>

      <div class="flex snap-x gap-4 overflow-x-auto pb-4">
        <article
          v-for="chat in chats"
          :key="chat.title"
          class="min-w-[240px] snap-center rounded-xl border p-5 transition-colors"
          :class="
            chat.active
              ? 'border-stone-900 bg-stone-900 text-white'
              : 'border-stone-200 bg-white text-stone-900 hover:border-stone-300'
          "
        >
          <div class="flex items-start justify-between gap-2">
            <p class="text-sm font-medium" :class="chat.active ? 'text-white' : 'text-stone-900'">
              {{ chat.title }}
            </p>
            <span class="text-xs" :class="chat.active ? 'text-stone-300' : 'text-stone-500'">{{
              chat.time
            }}</span>
          </div>
          <p
            class="mt-2 text-sm leading-relaxed"
            :class="chat.active ? 'text-stone-300' : 'text-stone-500'"
          >
            {{ chat.preview }}
          </p>
        </article>
      </div>
    </section>

    <div class="grid gap-8 lg:grid-cols-[280px_minmax(0,1fr)]">
      <aside class="hidden min-w-0 space-y-4 lg:block">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-semibold leading-6 text-stone-900">最近对话</h3>
          <button class="ui-btn-secondary px-3 py-1.5 text-sm">新建</button>
        </div>

        <div class="ui-list-card">
          <article
            v-for="chat in chats"
            :key="chat.title"
            class="ui-list-row cursor-pointer"
            :class="chat.active ? 'bg-stone-50' : ''"
          >
            <div class="flex items-start justify-between gap-2">
              <p
                class="text-sm font-medium"
                :class="chat.active ? 'text-stone-900' : 'text-stone-700'"
              >
                {{ chat.title }}
              </p>
              <span class="text-xs" :class="chat.active ? 'text-stone-500' : 'text-stone-400'">{{
                chat.time
              }}</span>
            </div>
            <p class="mt-1 line-clamp-2 text-sm text-stone-500">
              {{ chat.preview }}
            </p>
          </article>
        </div>
      </aside>

      <div class="flex h-[calc(100vh-16rem)] min-h-[500px] min-w-0 flex-col">
        <div
          class="mb-4 flex shrink-0 flex-col gap-3 border-b border-stone-200 pb-4 sm:flex-row sm:items-center sm:justify-between"
        >
          <div>
            <h3 class="text-base font-semibold leading-6 text-stone-900">当前对话</h3>
            <p class="mt-1 text-sm text-stone-500">默认发往：书桌咕咕机</p>
          </div>
          <div class="flex gap-2">
            <button
              class="inline-flex items-center rounded-full bg-stone-100 px-2.5 py-1 text-xs font-medium text-stone-800 transition-colors hover:bg-stone-200"
              title="切换回答语气"
            >
              语气：清楚温柔
            </button>
            <button
              class="inline-flex items-center rounded-full bg-stone-100 px-2.5 py-1 text-xs font-medium text-stone-800 transition-colors hover:bg-stone-200"
              title="设置目标长度"
            >
              长度：两段以内
            </button>
          </div>
        </div>

        <div class="flex-1 space-y-6 overflow-y-auto pr-4">
          <article
            v-for="message in messages"
            :key="message.text"
            class="max-w-[85%] rounded-2xl px-5 py-3.5 text-[15px] leading-relaxed"
            :class="
              message.role === 'user'
                ? 'ml-auto rounded-br-sm bg-stone-900 text-white'
                : 'rounded-bl-sm border border-stone-200 bg-white text-stone-900 shadow-sm'
            "
          >
            {{ message.text }}
          </article>
        </div>

        <div class="mt-6 shrink-0 space-y-4 border-t border-stone-200 pt-4">
          <div class="flex items-center justify-between">
            <div class="flex snap-x gap-2 overflow-x-auto pb-2">
              <button class="ui-btn-secondary snap-start whitespace-nowrap">打印最新回答</button>
              <button class="ui-btn-secondary snap-start whitespace-nowrap">选择指定回答</button>
              <button class="ui-btn-secondary snap-start whitespace-nowrap">打印整段对话</button>
            </div>
            <button
              class="ui-btn-secondary mb-2 whitespace-nowrap text-stone-500 hover:text-stone-900"
            >
              保存草稿
            </button>
          </div>

          <div
            class="relative rounded-xl border border-stone-200 bg-white shadow-sm transition-all focus-within:ring-2 focus-within:ring-stone-900 focus-within:ring-offset-2"
          >
            <textarea
              rows="3"
              placeholder="发送消息..."
              class="w-full resize-none border-0 bg-transparent p-4 text-[15px] leading-relaxed text-stone-900 placeholder:text-stone-400 focus:outline-none focus:ring-0"
            />
            <div
              class="flex items-center justify-between rounded-b-xl border-t border-stone-100 bg-stone-50/50 px-4 py-2"
            >
              <div class="flex gap-2">
                <button
                  class="rounded-md p-1.5 text-stone-500 transition-colors hover:bg-stone-200/50 hover:text-stone-900"
                  title="重新生成"
                >
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                    />
                  </svg>
                </button>
              </div>
              <button class="ui-btn-primary px-4 py-1.5">发送</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

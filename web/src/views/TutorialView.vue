<script setup lang="ts">
import { ref } from "vue";
import { RouterLink } from "vue-router";

import AppDialog from "@/components/AppDialog.vue";
import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();

const steps = [
  {
    number: "01",
    title: "在对话页通过聊天整理后打印",
    body: "想到什么就直接输入，Ink 会把内容整理成更适合打印的小纸条。你可以边聊边改，满意后直接发去打印。",
    note: "适合提醒、留言、鼓励语和临时清单。",
  },
  {
    number: "02",
    title: "在打印页直接新建并打印",
    body: "如果你已经想好内容，可以直接去打印页手动创建纸条，不需要先走对话流程。",
    note: "适合快速打印一句固定文案或临时通知。",
  },
  {
    number: "03",
    title: "用定时打印安排固定内容",
    body: "你可以在打印页设置自动任务，让某些内容按固定时间发送到设备，比如晨间提醒、晚安便签或周期性通知。",
    note: "适合重复出现、想稳定出纸的内容。",
  },
] as const;

const faqs = [
  {
    question: "我应该先用对话页还是打印页？",
    answer:
      "如果你还在想内容、想让 Ink 帮你润色，就先去对话页；如果内容已经确定，只想立刻出纸，就直接去打印页。",
  },
  {
    question: "定时打印适合拿来做什么？",
    answer: "适合每天、每周都要重复发送的内容，比如固定提醒、待办摘要或周期性问候。",
  },
  {
    question: "对话里的内容会不会太长，不适合打印？",
    answer: "可以直接继续追问，让 Ink 帮你压缩成更短、更适合出纸的小段内容，再发送打印。",
  },
  {
    question: "为什么纸条没有马上打印出来？",
    answer: "先去打印页看看任务状态；如果是定时任务，就等到设定时间；如果是手动任务，通常会很快进入打印队列。",
  },
] as const;

const feedbackOpen = ref(false);
const feedbackDraft = ref("");
const feedbackFormError = ref("");

function openFeedbackDialog() {
  feedbackOpen.value = true;
  feedbackDraft.value = "";
  feedbackFormError.value = "";
}

function closeFeedbackDialog() {
  feedbackOpen.value = false;
  feedbackFormError.value = "";
}

async function handleFeedbackSubmit() {
  feedbackFormError.value = "";

  if (!feedbackDraft.value.trim()) {
    feedbackFormError.value = "请先输入反馈内容。";
    return;
  }

  const success = await workspaceStore.submitFeedback(feedbackDraft.value);
  if (!success) {
    feedbackFormError.value = workspaceStore.feedbackError;
    return;
  }

  feedbackDraft.value = "";
  closeFeedbackDialog();
}
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-6 pt-4 sm:space-y-8 lg:space-y-10">
    <AppDialog
      :open="feedbackOpen"
      title="反馈"
      description="这里适合提交问题、建议或吐槽；提交后会提醒作者。"
      @close="closeFeedbackDialog"
    >
      <form class="space-y-4" @submit.prevent="handleFeedbackSubmit">
        <label class="block">
          <span class="mb-2 block text-sm font-medium text-stone-900">反馈内容</span>
          <textarea
            v-model="feedbackDraft"
            rows="6"
            placeholder="反馈（功能 / 建议 / 吐槽）"
            class="w-full rounded-xl border border-stone-200 bg-white px-4 py-3 text-sm leading-7 text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
          />
        </label>

        <p v-if="feedbackFormError" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {{ feedbackFormError }}
        </p>

        <div class="flex flex-col gap-3 sm:flex-row sm:justify-end">
          <button
            type="button"
            class="ui-btn-secondary px-4 py-2.5 text-sm"
            @click="closeFeedbackDialog"
          >
            取消
          </button>
          <button
            class="ui-btn-primary px-4 py-2.5 text-sm"
            :disabled="workspaceStore.feedbackSubmitting"
          >
            {{ workspaceStore.feedbackSubmitting ? "发送中..." : "提交反馈" }}
          </button>
        </div>
      </form>
    </AppDialog>

    <div
      class="relative overflow-hidden rounded-[2rem] border border-stone-200 bg-[linear-gradient(135deg,rgba(250,250,249,0.95),rgba(245,245,244,0.92))] px-5 py-6 shadow-sm sm:px-7 sm:py-8 lg:px-10"
    >
      <div
        aria-hidden="true"
        class="absolute top-0 right-0 h-40 w-40 rounded-full bg-amber-100/70 blur-3xl"
      />
      <div
        aria-hidden="true"
        class="absolute bottom-0 left-0 h-32 w-32 rounded-full bg-stone-200/70 blur-3xl"
      />

      <div class="relative space-y-6">
        <div class="space-y-4">
          <p class="text-sm font-medium tracking-[0.2em] text-stone-500 uppercase">使用教程</p>
          <div class="space-y-3">
            <h2
              class="max-w-4xl text-[clamp(2rem,4.8vw,3.5rem)] font-semibold tracking-tight text-stone-900"
            >
              Ink 里最常用的三种打印方式
              <span class="block text-stone-600">对话打印、直接打印、定时打印</span>
            </h2>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-3">
          <RouterLink
            to="/conversations"
            class="rounded-2xl border border-white/70 bg-white/85 px-4 py-4 shadow-sm backdrop-blur transition-colors hover:border-stone-300 hover:bg-white"
          >
            <p class="text-sm font-semibold text-stone-900">对话页</p>
            <p class="mt-2 text-sm leading-6 text-stone-600">边聊边整理内容，再直接发去打印。</p>
          </RouterLink>
          <RouterLink
            to="/prints"
            class="rounded-2xl border border-white/70 bg-white/85 px-4 py-4 shadow-sm backdrop-blur transition-colors hover:border-stone-300 hover:bg-white"
          >
            <p class="text-sm font-semibold text-stone-900">打印页</p>
            <p class="mt-2 text-sm leading-6 text-stone-600">手动新建纸条，适合快速直接出纸。</p>
          </RouterLink>
          <RouterLink
            to="/prints"
            class="rounded-2xl border border-white/70 bg-white/85 px-4 py-4 shadow-sm backdrop-blur transition-colors hover:border-stone-300 hover:bg-white"
          >
            <p class="text-sm font-semibold text-stone-900">定时打印</p>
            <p class="mt-2 text-sm leading-6 text-stone-600">设置固定时间，按计划自动发送内容。</p>
          </RouterLink>
        </div>

        <div class="flex flex-col gap-3 sm:flex-row">
          <RouterLink to="/conversations" class="ui-btn-primary px-4 py-2.5 text-center text-sm">
            去对话页
          </RouterLink>
          <RouterLink to="/prints" class="ui-btn-secondary px-4 py-2.5 text-center text-sm">
            去打印页
          </RouterLink>
        </div>
      </div>
    </div>

    <div class="grid gap-6 lg:grid-cols-[minmax(0,1.15fr)_0.85fr]">
      <article class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
        <div class="max-w-2xl">
          <p class="text-sm font-medium tracking-[0.18em] text-stone-500 uppercase">操作步骤</p>
          <h3 class="mt-3 text-2xl font-semibold tracking-tight text-stone-900">
            三种常用使用方式
          </h3>
        </div>

        <div class="mt-6 space-y-4">
          <article
            v-for="(step, index) in steps"
            :key="step.number"
            class="rounded-[1.5rem] border px-5 py-5"
            :class="
              index === 0 ? 'border-amber-200 bg-amber-50/80' : 'border-stone-200 bg-stone-50/70'
            "
          >
            <div class="grid gap-4 md:grid-cols-[auto_minmax(0,1fr)] md:items-start">
              <div
                class="inline-flex h-12 w-12 items-center justify-center rounded-2xl text-sm font-semibold"
                :class="index === 0 ? 'bg-amber-100 text-amber-900' : 'bg-white text-stone-900'"
              >
                {{ step.number }}
              </div>
              <div>
                <h4 class="text-lg font-semibold text-stone-900">{{ step.title }}</h4>
                <p class="mt-2 text-sm leading-7 text-stone-600">{{ step.body }}</p>
                <p
                  class="mt-4 rounded-2xl px-4 py-3 text-sm leading-6"
                  :class="index === 0 ? 'bg-white text-amber-900' : 'bg-white text-stone-600'"
                >
                  {{ step.note }}
                </p>
              </div>
            </div>
          </article>
        </div>
      </article>

      <div class="space-y-6">
        <article class="rounded-[1.8rem] border border-stone-200 bg-stone-50 p-6 shadow-sm sm:p-7">
          <p class="text-sm font-medium tracking-[0.18em] text-stone-500 uppercase">
            作为手机应用使用
          </p>
          <h3 class="mt-3 text-2xl font-semibold tracking-tight text-stone-900">
            把 Ink 添加到 iPhone 主屏幕
          </h3>
          <p class="mt-3 text-sm leading-7 text-stone-600">
            用 Safari 打开 Ink 后，点击底部分享按钮，选择“添加到主屏幕”。之后你可以像普通 App
            一样从桌面进入 Ink，顶部和底部导航都会按手机安全区适配。
          </p>
        </article>

        <article class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
          <h3 class="text-2xl font-semibold tracking-tight text-stone-900">常见问题</h3>
          <div class="mt-5 space-y-4">
            <article
              v-for="item in faqs"
              :key="item.question"
              class="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4"
            >
              <p class="text-sm font-semibold text-stone-900">{{ item.question }}</p>
              <p class="mt-2 text-sm leading-7 text-stone-600">{{ item.answer }}</p>
            </article>
          </div>

          <div class="mt-6 rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4">
            <p class="text-sm font-semibold text-stone-900">没找到问题？</p>
            <p class="mt-2 text-sm leading-7 text-stone-600">
              可以直接把你的问题、建议或吐槽发给作者。
            </p>
            <button
              type="button"
              class="ui-btn-secondary mt-4 w-full px-4 py-2.5 text-sm sm:w-auto"
              @click="openFeedbackDialog"
            >
              反馈给作者
            </button>
          </div>
        </article>

        <aside class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
          <h3 class="text-xl font-semibold text-stone-900">开始使用</h3>
          <div class="mt-5 space-y-3">
            <RouterLink
              to="/conversations"
              class="ui-btn-primary block w-full px-4 py-2.5 text-center text-sm"
            >
              去对话页
            </RouterLink>
            <RouterLink
              to="/prints"
              class="ui-btn-secondary block w-full px-4 py-2.5 text-center text-sm"
            >
              去打印页
            </RouterLink>
          </div>
        </aside>
      </div>
    </div>
  </section>
</template>

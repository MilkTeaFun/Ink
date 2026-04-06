<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";

import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();

const answerStyleOptions = [
  { label: "清楚温柔", value: "clear-gentle" },
  { label: "温柔鼓励", value: "warm-encouraging" },
  { label: "直接简洁", value: "concise-direct" },
] as const;

const responseLengthOptions = [
  { label: "一句到两句", value: "short" },
  { label: "两段以内", value: "medium" },
  { label: "稍微详细", value: "long" },
] as const;

const hasMessages = computed(() => (workspaceStore.activeConversation?.messages.length ?? 0) > 0);

function handleDraftInput(event: Event) {
  const target = event.target as HTMLTextAreaElement | null;
  workspaceStore.updateCurrentDraft(target?.value ?? "");
}
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 px-4 pt-4 pb-24 sm:px-0 lg:pb-12">
    <div>
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">对话</h2>
      <p class="mt-1 text-sm text-stone-500">通过对话整理内容，再决定打印哪一段。</p>
    </div>

    <section class="space-y-4 lg:hidden">
      <div class="flex items-center justify-between gap-3">
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">最近对话</h3>
          <p class="mt-1 text-sm text-stone-500">草稿和消息会保存在本地。</p>
        </div>
        <button
          class="ui-btn-secondary px-3 py-1.5 text-sm"
          @click="workspaceStore.createConversation"
        >
          新建
        </button>
      </div>

      <div class="flex snap-x gap-4 overflow-x-auto pb-4">
        <button
          v-for="chat in workspaceStore.conversations"
          :key="chat.id"
          type="button"
          class="min-w-[240px] snap-center rounded-xl border p-5 text-left transition-colors"
          :class="
            workspaceStore.activeConversationId === chat.id
              ? 'border-stone-900 bg-stone-900 text-white'
              : 'border-stone-200 bg-white text-stone-900 hover:border-stone-300'
          "
          @click="workspaceStore.selectConversation(chat.id)"
        >
          <div class="flex items-start justify-between gap-2">
            <p
              class="text-sm font-medium"
              :class="
                workspaceStore.activeConversationId === chat.id ? 'text-white' : 'text-stone-900'
              "
            >
              {{ chat.title }}
            </p>
            <span
              class="text-xs"
              :class="
                workspaceStore.activeConversationId === chat.id
                  ? 'text-stone-300'
                  : 'text-stone-500'
              "
            >
              {{ workspaceStore.formatPrintTime(chat.updatedAt) }}
            </span>
          </div>
          <p
            class="mt-2 text-sm leading-relaxed"
            :class="
              workspaceStore.activeConversationId === chat.id ? 'text-stone-300' : 'text-stone-500'
            "
          >
            {{ chat.preview }}
          </p>
        </button>
      </div>
    </section>

    <div class="grid gap-8 lg:grid-cols-[280px_minmax(0,1fr)]">
      <aside class="hidden min-w-0 space-y-4 lg:block">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-base leading-6 font-semibold text-stone-900">最近对话</h3>
            <p class="mt-1 text-sm text-stone-500">会话、草稿和选择结果实时同步。</p>
          </div>
          <button
            class="ui-btn-secondary px-3 py-1.5 text-sm"
            @click="workspaceStore.createConversation"
          >
            新建
          </button>
        </div>

        <div v-if="workspaceStore.conversations.length" class="ui-list-card">
          <button
            v-for="chat in workspaceStore.conversations"
            :key="chat.id"
            type="button"
            class="ui-list-row block w-full cursor-pointer text-left"
            :class="{ 'is-active': workspaceStore.activeConversationId === chat.id }"
            @click="workspaceStore.selectConversation(chat.id)"
          >
            <div class="flex items-start justify-between gap-2">
              <p
                class="text-sm font-medium"
                :class="
                  workspaceStore.activeConversationId === chat.id
                    ? 'text-stone-900'
                    : 'text-stone-700'
                "
              >
                {{ chat.title }}
              </p>
              <span
                class="text-xs"
                :class="
                  workspaceStore.activeConversationId === chat.id
                    ? 'text-stone-500'
                    : 'text-stone-400'
                "
              >
                {{ workspaceStore.formatPrintTime(chat.updatedAt) }}
              </span>
            </div>
            <p class="mt-1 line-clamp-2 text-sm text-stone-500">
              {{ chat.preview }}
            </p>
          </button>
        </div>
      </aside>

      <div class="flex h-[calc(100vh-16rem)] min-h-[500px] min-w-0 flex-col">
        <div
          class="mb-4 flex shrink-0 flex-col gap-3 border-b border-stone-200 pb-4 sm:flex-row sm:items-center sm:justify-between"
        >
          <div>
            <h3 class="text-base leading-6 font-semibold text-stone-900">
              {{ workspaceStore.activeConversation?.title ?? "当前对话" }}
            </h3>
            <p class="mt-1 text-sm text-stone-500">
              默认发往：{{ workspaceStore.activeDeviceLabel || "尚未设置" }}
            </p>
          </div>
          <div class="flex flex-col gap-2 sm:items-end">
            <div class="flex flex-wrap gap-2">
              <button
                v-for="option in answerStyleOptions"
                :key="option.value"
                type="button"
                class="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium transition-colors"
                :class="
                  workspaceStore.activeAnswerStyle === option.value
                    ? 'bg-stone-900 text-white'
                    : 'bg-stone-100 text-stone-800 hover:bg-stone-200'
                "
                @click="workspaceStore.setAnswerStyle(option.value)"
              >
                {{ option.label }}
              </button>
            </div>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="option in responseLengthOptions"
                :key="option.value"
                type="button"
                class="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium transition-colors"
                :class="
                  workspaceStore.responseLength === option.value
                    ? 'bg-stone-900 text-white'
                    : 'bg-stone-100 text-stone-800 hover:bg-stone-200'
                "
                @click="workspaceStore.setResponseLength(option.value)"
              >
                {{ option.label }}
              </button>
            </div>
          </div>
        </div>

        <div
          v-if="!hasMessages"
          class="flex flex-1 items-center justify-center rounded-2xl border border-dashed border-stone-200 bg-stone-50 px-6 text-center"
        >
          <div class="max-w-sm">
            <h4 class="text-base font-semibold text-stone-900">这里还没有消息</h4>
            <p class="mt-2 text-sm leading-relaxed text-stone-500">
              输入一段想整理的内容，Ink 会先帮你生成一版适合打印的小纸条文案。
            </p>
          </div>
        </div>

        <div v-else class="flex-1 space-y-4 overflow-y-auto pr-2">
          <article
            v-for="message in workspaceStore.activeConversation?.messages"
            :key="message.id"
            class="space-y-2"
          >
            <button
              type="button"
              class="block max-w-[88%] rounded-2xl px-5 py-3.5 text-left text-[15px] leading-relaxed"
              :class="
                message.role === 'user'
                  ? 'ml-auto cursor-default rounded-br-sm bg-stone-900 text-white'
                  : workspaceStore.selectedAssistantMessage?.id === message.id
                    ? 'rounded-bl-sm border border-stone-900 bg-stone-50 text-stone-900 shadow-sm'
                    : 'rounded-bl-sm border border-stone-200 bg-white text-stone-900 shadow-sm'
              "
              :disabled="message.role === 'user'"
              @click="
                message.role === 'assistant' && workspaceStore.selectAssistantMessage(message.id)
              "
            >
              {{ message.text }}
            </button>
            <div
              v-if="message.role === 'assistant'"
              class="px-1 text-xs text-stone-500"
              :class="workspaceStore.selectedAssistantMessage?.id === message.id ? 'ml-1' : 'ml-1'"
            >
              {{
                workspaceStore.selectedAssistantMessage?.id === message.id
                  ? "已选中用于打印"
                  : "点击可选中这条回答"
              }}
            </div>
          </article>

          <article
            v-if="workspaceStore.isGenerating"
            class="max-w-[85%] rounded-2xl rounded-bl-sm border border-stone-200 bg-white px-5 py-3.5 text-[15px] leading-relaxed text-stone-500 shadow-sm"
          >
            正在整理新的回复...
          </article>
        </div>

        <div class="mt-6 shrink-0 space-y-4 border-t border-stone-200 pt-4">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div class="flex flex-wrap gap-2">
              <button
                class="ui-btn-secondary whitespace-nowrap"
                @click="workspaceStore.createPrintFromLatestReply"
              >
                打印最新回答
              </button>
              <button
                class="ui-btn-secondary whitespace-nowrap"
                @click="workspaceStore.createPrintFromSelectedReply"
              >
                打印选中回答
              </button>
              <button
                class="ui-btn-secondary whitespace-nowrap"
                @click="workspaceStore.createPrintFromConversation"
              >
                打印整段对话
              </button>
            </div>
            <div class="flex items-center gap-2">
              <button
                class="ui-btn-secondary whitespace-nowrap"
                @click="workspaceStore.saveCurrentDraft"
              >
                保存草稿
              </button>
              <RouterLink to="/prints" class="ui-btn-secondary whitespace-nowrap"
                >查看打印队列</RouterLink
              >
            </div>
          </div>

          <p
            v-if="workspaceStore.generationError"
            class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700"
          >
            {{ workspaceStore.generationError }}
          </p>

          <div
            class="relative rounded-xl border border-stone-200 bg-white shadow-sm transition-all focus-within:ring-2 focus-within:ring-stone-900 focus-within:ring-offset-2"
          >
            <textarea
              :value="workspaceStore.activeConversation?.draft ?? ''"
              rows="4"
              placeholder="发送消息..."
              class="w-full resize-none border-0 bg-transparent p-4 text-[15px] leading-relaxed text-stone-900 placeholder:text-stone-400 focus:ring-0 focus:outline-none"
              @input="handleDraftInput"
            />
            <div
              class="flex flex-col gap-3 rounded-b-xl border-t border-stone-100 bg-stone-50/50 px-4 py-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div class="flex flex-wrap items-center gap-2 text-xs text-stone-500">
                <span>纸条风格：{{ workspaceStore.activeNoteStyle }}</span>
                <span>回复长度：{{ workspaceStore.responseLength }}</span>
                <span
                  >发送前确认：{{
                    workspaceStore.sendConfirmationEnabled ? "已开启" : "已关闭"
                  }}</span
                >
              </div>
              <div class="flex gap-2">
                <button
                  class="rounded-md p-1.5 text-stone-500 transition-colors hover:bg-stone-200/50 hover:text-stone-900"
                  title="重新生成"
                  :disabled="workspaceStore.isGenerating"
                  @click="workspaceStore.regenerateLatestReply"
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
                <button
                  class="ui-btn-primary px-4 py-1.5"
                  :disabled="workspaceStore.isGenerating"
                  @click="workspaceStore.sendCurrentDraft"
                >
                  {{ workspaceStore.isGenerating ? "生成中..." : "发送" }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";

import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();

const hasMessages = computed(() => (workspaceStore.activeConversation?.messages.length ?? 0) > 0);

function handleDraftInput(event: Event) {
  const target = event.target as HTMLTextAreaElement | null;
  workspaceStore.updateCurrentDraft(target?.value ?? "");
}

function handleDeleteCurrentConversation() {
  const current = workspaceStore.activeConversation;

  if (!current) {
    return;
  }

  const hasContent = current.messages.length > 0 || current.draft.trim().length > 0;
  if (hasContent && typeof window !== "undefined") {
    const confirmed = window.confirm(`删除“${current.title}”？`);
    if (!confirmed) {
      return;
    }
  }

  workspaceStore.deleteConversation(current.id);
}
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-6 pt-4 sm:space-y-8">
    <div>
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">对话</h2>
    </div>

    <section class="space-y-4 lg:hidden">
      <div class="flex items-center justify-between gap-3">
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">最近对话</h3>
        </div>
        <button
          class="ui-btn-secondary px-3 py-1.5 text-sm"
          @click="workspaceStore.createConversation"
        >
          新建
        </button>
      </div>

      <div class="flex snap-x gap-4 overflow-x-auto pb-2">
        <button
          v-for="chat in workspaceStore.conversations"
          :key="chat.id"
          type="button"
          class="max-w-[18rem] min-w-[85%] snap-center rounded-xl border p-5 text-left transition-colors"
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

    <div class="grid gap-6 lg:grid-cols-[280px_minmax(0,1fr)] lg:gap-8">
      <aside class="hidden min-w-0 space-y-4 lg:block">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-base leading-6 font-semibold text-stone-900">最近对话</h3>
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

      <div
        class="flex min-h-[24rem] min-w-0 flex-col rounded-[1.5rem] border border-stone-200 bg-white/90 p-4 shadow-sm sm:min-h-[28rem] lg:h-[calc(100dvh-16rem)] lg:min-h-[500px] lg:rounded-none lg:border-0 lg:bg-transparent lg:p-0 lg:shadow-none"
      >
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
          <button
            type="button"
            class="ui-btn-secondary w-full px-3 py-1.5 text-sm sm:w-auto"
            @click="handleDeleteCurrentConversation"
          >
            删除对话
          </button>
        </div>

        <div
          v-if="!hasMessages"
          class="flex flex-1 items-center justify-center rounded-2xl border border-dashed border-stone-200 bg-stone-50 px-6 text-center"
        >
          <h4 class="text-base font-semibold text-stone-900">这里还没有消息</h4>
        </div>

        <div v-else class="space-y-4 lg:flex-1 lg:overflow-y-auto lg:pr-2">
          <article
            v-for="message in workspaceStore.activeConversation?.messages"
            :key="message.id"
            class="space-y-2"
          >
            <div
              class="flex items-center gap-3"
              :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
            >
              <button
                type="button"
                class="block max-w-[88%] rounded-2xl border px-5 py-3.5 text-left text-[15px] leading-relaxed shadow-sm transition-colors"
                :class="
                  message.role === 'user'
                    ? workspaceStore.selectedConversationMessageIds.includes(message.id)
                      ? 'rounded-br-sm border-stone-900 bg-stone-800 text-white ring-1 ring-stone-400'
                      : 'rounded-br-sm border-stone-900 bg-stone-900 text-white'
                    : workspaceStore.selectedConversationMessageIds.includes(message.id)
                      ? 'rounded-bl-sm border-amber-500 bg-amber-50 text-stone-900 ring-1 ring-amber-200'
                      : 'rounded-bl-sm border-stone-200 bg-white text-stone-900'
                "
                @click="workspaceStore.toggleConversationMessageSelection(message.id)"
              >
                {{ message.text }}
              </button>
              <button
                type="button"
                class="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full border transition-colors"
                :class="
                  workspaceStore.selectedConversationMessageIds.includes(message.id)
                    ? 'border-amber-500 bg-amber-500'
                    : 'border-stone-300 bg-white'
                "
                :aria-label="
                  workspaceStore.selectedConversationMessageIds.includes(message.id)
                    ? '取消选择这条消息'
                    : '选择这条消息'
                "
                :aria-pressed="workspaceStore.selectedConversationMessageIds.includes(message.id)"
                @click="workspaceStore.toggleConversationMessageSelection(message.id)"
              >
                <span
                  class="h-2 w-2 rounded-full"
                  :class="
                    workspaceStore.selectedConversationMessageIds.includes(message.id)
                      ? 'bg-white'
                      : 'bg-transparent'
                  "
                />
              </button>
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
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex flex-wrap gap-2">
              <button
                class="ui-btn-secondary whitespace-nowrap"
                @click="workspaceStore.createPrintFromSelectedMessages"
              >
                打印选中问答
              </button>
              <button
                class="ui-btn-secondary whitespace-nowrap"
                @click="workspaceStore.createPrintFromConversation"
              >
                打印当前对话
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
                <span>已选中 {{ workspaceStore.selectedConversationMessageIds.length }} 条</span>
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

<script setup lang="ts">
import { useRouter } from "vue-router";

import { useWorkspaceStore } from "@/stores/workspace";
import { getThemeDescription } from "@/utils/workspace";

const router = useRouter();
const workspaceStore = useWorkspaceStore();

const themes = [
  { label: "柔光", value: "soft" },
  { label: "浅色", value: "light" },
  { label: "系统跟随", value: "system" },
] as const;

const answerStyleOptions = [
  { label: "清楚温柔", value: "clear-gentle" },
  { label: "温柔鼓励", value: "warm-encouraging" },
  { label: "直接简洁", value: "concise-direct" },
] as const;

const noteStyleOptions = [
  { label: "简洁清晰", value: "clean" },
  { label: "柔和留白", value: "gentle" },
  { label: "清单格式", value: "list" },
] as const;

function handleDefaultDeviceChange(event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.setDefaultDevice(target?.value ?? workspaceStore.defaultDeviceId);
}

function handleNoteStyleChange(event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.setNoteStyle((target?.value ?? workspaceStore.activeNoteStyle) as never);
}

function handleAnswerStyleChange(event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.setAnswerStyle((target?.value ?? workspaceStore.activeAnswerStyle) as never);
}

async function handleLogout() {
  workspaceStore.logout();
  await router.replace("/login");
}
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 pt-4 pb-24 lg:pb-12">
    <div class="max-w-2xl px-4 sm:px-0">
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">设置</h2>
      <p class="mt-1 text-sm text-stone-500">管理账号、打印、主题和授权选项。</p>
    </div>

    <div class="space-y-12">
      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">账号管理</h3>
          <p class="mt-1 text-sm text-stone-500">登录状态、退出和本地会话保护都在这里。</p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">当前账号</p>
                <p class="mt-0.5 text-sm text-stone-500">
                  {{ workspaceStore.authUser?.email ?? "当前未登录" }}
                </p>
              </div>
              <button
                type="button"
                class="ui-btn-secondary px-3 py-1.5 text-sm"
                @click="handleLogout"
              >
                退出
              </button>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">登录保护</p>
                <p class="mt-0.5 text-sm text-stone-500">
                  {{
                    workspaceStore.loginProtectionEnabled
                      ? "刷新后需要重新登录，不保留本地会话。"
                      : "刷新后继续保留当前本地会话。"
                  }}
                </p>
              </div>
              <button
                type="button"
                class="ui-toggle"
                :class="{ 'is-on': workspaceStore.loginProtectionEnabled }"
                :aria-label="`${workspaceStore.loginProtectionEnabled ? '关闭' : '开启'}登录保护`"
                :aria-pressed="workspaceStore.loginProtectionEnabled"
                @click="workspaceStore.setLoginProtection(!workspaceStore.loginProtectionEnabled)"
              >
                <span class="ui-toggle-thumb" />
              </button>
            </div>
          </div>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">打印设置</h3>
          <p class="mt-1 text-sm text-stone-500">
            默认设备、发送确认和纸条风格会立即同步到其他页面。
          </p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">默认设备</p>
                <p class="mt-0.5 text-sm text-stone-500">对话和新建打印都会优先发往这里。</p>
              </div>
              <select
                :value="workspaceStore.defaultDeviceId"
                class="rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900"
                @change="handleDefaultDeviceChange"
              >
                <option
                  v-for="device in workspaceStore.devices"
                  :key="device.id"
                  :value="device.id"
                >
                  {{ device.name }}
                </option>
              </select>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">发送前确认</p>
                <p class="mt-0.5 text-sm text-stone-500">
                  {{
                    workspaceStore.sendConfirmationEnabled
                      ? "生成的新内容会先进入待确认列表。"
                      : "生成的新内容会直接进入打印队列。"
                  }}
                </p>
              </div>
              <button
                type="button"
                class="ui-toggle"
                :class="{ 'is-on': workspaceStore.sendConfirmationEnabled }"
                :aria-label="`${workspaceStore.sendConfirmationEnabled ? '关闭' : '开启'}发送前确认`"
                :aria-pressed="workspaceStore.sendConfirmationEnabled"
                @click="workspaceStore.setSendConfirmation(!workspaceStore.sendConfirmationEnabled)"
              >
                <span class="ui-toggle-thumb" />
              </button>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">纸条风格</p>
                <p class="mt-0.5 text-sm text-stone-500">会影响对话页生成内容的组织方式。</p>
              </div>
              <select
                :value="workspaceStore.activeNoteStyle"
                class="rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900"
                @change="handleNoteStyleChange"
              >
                <option
                  v-for="option in noteStyleOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </div>
          </div>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">页面主题</h3>
          <p class="mt-1 text-sm text-stone-500">切换应用视觉模式，设置会保存在本地。</p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-choice-grid">
            <button
              v-for="theme in themes"
              :key="theme.value"
              type="button"
              class="ui-btn-secondary justify-center py-2 text-sm"
              :class="
                theme.value === workspaceStore.selectedTheme
                  ? 'border-stone-300 bg-white text-stone-900 ring-1 ring-stone-200/70'
                  : 'border-transparent bg-transparent text-stone-600 shadow-none hover:border-stone-200 hover:bg-white'
              "
              :aria-pressed="theme.value === workspaceStore.selectedTheme"
              @click="workspaceStore.setTheme(theme.value)"
            >
              {{ theme.label }}
            </button>
          </div>
          <p class="mt-2 text-sm text-stone-500">
            当前主题：{{ getThemeDescription(workspaceStore.selectedTheme) }}
          </p>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">AI 服务</h3>
          <p class="mt-1 text-sm text-stone-500">
            当前阶段使用前端 mock service，也保留服务商绑定状态。
          </p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">默认服务商</p>
                <p class="mt-0.5 text-sm text-stone-500">
                  {{
                    workspaceStore.serviceBinding.bound
                      ? workspaceStore.serviceBinding.providerName
                      : "暂未绑定"
                  }}
                </p>
              </div>
              <button
                type="button"
                class="ui-btn-secondary px-3 py-1.5 text-sm"
                @click="workspaceStore.bindService"
              >
                {{ workspaceStore.serviceBinding.bound ? "重新绑定" : "去绑定" }}
              </button>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">回答风格</p>
                <p class="mt-0.5 text-sm text-stone-500">会直接影响对话页新的回复生成方式。</p>
              </div>
              <select
                :value="workspaceStore.activeAnswerStyle"
                class="rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900"
                @change="handleAnswerStyleChange"
              >
                <option
                  v-for="option in answerStyleOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">当前模型标签</p>
                <p class="mt-0.5 text-sm text-stone-500">{{ workspaceStore.activeModelLabel }}</p>
              </div>
              <span
                class="inline-flex items-center rounded-full bg-stone-100 px-2.5 py-0.5 text-xs font-medium text-stone-800"
              >
                {{ workspaceStore.serviceBinding.bound ? "mock 已绑定" : "mock 未绑定" }}
              </span>
            </div>
          </div>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">授权与隐私</h3>
          <p class="mt-1 text-sm text-stone-500">无后端阶段用来源状态来模拟授权、异常和重连。</p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div v-for="source in workspaceStore.sources" :key="source.id" class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">{{ source.name }}</p>
                <p class="mt-0.5 text-sm text-stone-500">
                  {{ source.type }} · {{ source.note }} ·
                  {{ workspaceStore.getSourceStatusLabel(source.status) }}
                </p>
              </div>
              <button
                type="button"
                class="ui-btn-secondary px-3 py-1.5 text-sm"
                @click="workspaceStore.cycleSourceStatus(source.id)"
              >
                切换状态
              </button>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

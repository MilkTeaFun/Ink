<script setup lang="ts">
import { ref } from "vue";
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

const currentPassword = ref("");
const newPassword = ref("");
const confirmPassword = ref("");
const passwordFormError = ref("");
const currentPasswordVisible = ref(false);
const newPasswordVisible = ref(false);
const confirmPasswordVisible = ref(false);

function handleDefaultDeviceChange(event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.setDefaultDevice(target?.value ?? workspaceStore.defaultDeviceId);
}

async function handleLogout() {
  await workspaceStore.logout();
  await router.replace("/login");
}

async function handlePasswordSubmit() {
  passwordFormError.value = "";
  const trimmedNewPassword = newPassword.value.trim();
  const trimmedConfirmPassword = confirmPassword.value.trim();

  if (!currentPassword.value.trim()) {
    passwordFormError.value = "请输入当前密码。";
    return;
  }

  if (trimmedNewPassword.length < 8) {
    passwordFormError.value = "新密码至少需要 8 位。";
    return;
  }

  if (trimmedNewPassword !== trimmedConfirmPassword) {
    passwordFormError.value = "两次输入的新密码不一致。";
    return;
  }

  const success = await workspaceStore.changePassword(currentPassword.value, trimmedNewPassword);
  if (!success) {
    passwordFormError.value = workspaceStore.authError;
    return;
  }

  currentPassword.value = "";
  newPassword.value = "";
  confirmPassword.value = "";
  await router.replace({
    path: "/login",
    query: {
      notice: "password-updated",
    },
  });
}
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 pt-4 pb-24 lg:pb-12">
    <div class="max-w-2xl px-4 sm:px-0">
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">设置</h2>
    </div>

    <div class="space-y-12">
      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">账号管理</h3>
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
                      ? "刷新后需要重新登录"
                      : "刷新后保留登录状态"
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
            <div class="rounded-xl border border-stone-200 bg-stone-50 p-4">
              <p class="text-sm font-medium text-stone-900">修改密码</p>

              <form class="mt-4 space-y-4" @submit.prevent="handlePasswordSubmit">
                <div class="grid gap-4 md:grid-cols-3">
                  <label class="block">
                    <span class="mb-2 block text-sm font-medium text-stone-900">当前密码</span>
                    <div
                      class="flex items-center gap-2 rounded-lg border border-stone-200 bg-white px-3"
                    >
                      <input
                        v-model="currentPassword"
                        :type="currentPasswordVisible ? 'text' : 'password'"
                        autocomplete="current-password"
                        class="min-w-0 flex-1 bg-transparent py-2 text-sm text-stone-900 focus:outline-none"
                      />
                      <button
                        type="button"
                        class="shrink-0 text-xs font-medium text-stone-500 hover:text-stone-900"
                        @click="currentPasswordVisible = !currentPasswordVisible"
                      >
                        {{ currentPasswordVisible ? "隐藏" : "显示" }}
                      </button>
                    </div>
                  </label>
                  <label class="block">
                    <span class="mb-2 block text-sm font-medium text-stone-900">新密码</span>
                    <div
                      class="flex items-center gap-2 rounded-lg border border-stone-200 bg-white px-3"
                    >
                      <input
                        v-model="newPassword"
                        :type="newPasswordVisible ? 'text' : 'password'"
                        autocomplete="new-password"
                        class="min-w-0 flex-1 bg-transparent py-2 text-sm text-stone-900 focus:outline-none"
                      />
                      <button
                        type="button"
                        class="shrink-0 text-xs font-medium text-stone-500 hover:text-stone-900"
                        @click="newPasswordVisible = !newPasswordVisible"
                      >
                        {{ newPasswordVisible ? "隐藏" : "显示" }}
                      </button>
                    </div>
                  </label>
                  <label class="block">
                    <span class="mb-2 block text-sm font-medium text-stone-900">确认新密码</span>
                    <div
                      class="flex items-center gap-2 rounded-lg border border-stone-200 bg-white px-3"
                    >
                      <input
                        v-model="confirmPassword"
                        :type="confirmPasswordVisible ? 'text' : 'password'"
                        autocomplete="new-password"
                        class="min-w-0 flex-1 bg-transparent py-2 text-sm text-stone-900 focus:outline-none"
                      />
                      <button
                        type="button"
                        class="shrink-0 text-xs font-medium text-stone-500 hover:text-stone-900"
                        @click="confirmPasswordVisible = !confirmPasswordVisible"
                      >
                        {{ confirmPasswordVisible ? "隐藏" : "显示" }}
                      </button>
                    </div>
                  </label>
                </div>
                <p
                  v-if="passwordFormError"
                  class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700"
                >
                  {{ passwordFormError }}
                </p>
                <button
                  type="submit"
                  class="ui-btn-primary px-4 py-2 text-sm"
                  :disabled="workspaceStore.passwordChangeLoading"
                >
                  {{ workspaceStore.passwordChangeLoading ? "提交中..." : "更新密码" }}
                </button>
              </form>
            </div>
          </div>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">打印设置</h3>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">默认设备</p>
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
          </div>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">页面主题</h3>
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
                      : "未连接"
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
                <p class="text-sm font-medium text-stone-900">当前模型标签</p>
                <p class="mt-0.5 text-sm text-stone-500">{{ workspaceStore.activeModelLabel }}</p>
              </div>
              <span
                class="inline-flex items-center rounded-full bg-stone-100 px-2.5 py-0.5 text-xs font-medium text-stone-800"
              >
                {{ workspaceStore.serviceBinding.bound ? "已连接" : "未连接" }}
              </span>
            </div>
          </div>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">插件</h3>
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
                @click="workspaceStore.toggleSourceConnection(source.id)"
              >
                {{ source.status === "connected" ? "解绑" : "连接" }}
              </button>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

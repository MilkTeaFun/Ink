<script setup lang="ts">
import { computed, ref, watch } from "vue";
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
const newAccountEmail = ref("");
const newAccountName = ref("");
const newAccountPassword = ref("");
const newAccountFormError = ref("");
const newAccountPasswordVisible = ref(false);
const aiProviderName = ref("OpenAI Compatible");
const aiBaseUrl = ref("");
const aiModel = ref("gpt-4.1-mini");
const aiApiKey = ref("");
const aiFormError = ref("");

const AI_PROVIDER_TYPE = "openai-compatible";
const AI_PROVIDER_NAME_FALLBACK = "OpenAI Compatible";
const AI_MODEL_FALLBACK = "gpt-4.1-mini";

const aiConfigErrorMessage = computed(() => aiFormError.value || workspaceStore.aiConfigError);

watch(
  () => workspaceStore.aiConfigSummary,
  (summary) => {
    aiProviderName.value = summary.providerName || AI_PROVIDER_NAME_FALLBACK;
    aiBaseUrl.value = summary.baseUrl;
    aiModel.value = summary.model || AI_MODEL_FALLBACK;
  },
  { deep: true, immediate: true },
);

function handleDefaultDeviceChange(event: Event) {
  const target = event.target as HTMLSelectElement | null;
  workspaceStore.setDefaultDevice(target?.value ?? workspaceStore.defaultDeviceId);
}

async function handleLogout() {
  await workspaceStore.logout();
  await router.replace("/status");
}

async function handlePasswordSubmit() {
  passwordFormError.value = "";
  const nextPassword = newPassword.value;
  const confirmNextPassword = confirmPassword.value;

  if (!currentPassword.value.trim()) {
    passwordFormError.value = "请输入当前密码。";
    return;
  }

  if (nextPassword.length < 8) {
    passwordFormError.value = "新密码至少需要 8 位。";
    return;
  }

  if (nextPassword !== confirmNextPassword) {
    passwordFormError.value = "两次输入的新密码不一致。";
    return;
  }

  const success = await workspaceStore.changePassword(currentPassword.value, nextPassword);
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

async function handleCreateAccountSubmit() {
  newAccountFormError.value = "";

  if (!newAccountEmail.value.trim()) {
    newAccountFormError.value = "请输入新账号。";
    return;
  }

  if (newAccountPassword.value.trim().length < 8) {
    newAccountFormError.value = "新账号密码至少需要 8 位。";
    return;
  }

  const success = await workspaceStore.createAccount(
    newAccountEmail.value.trim(),
    newAccountName.value.trim(),
    newAccountPassword.value,
  );

  if (!success) {
    newAccountFormError.value = workspaceStore.accountCreationError;
    return;
  }

  newAccountEmail.value = "";
  newAccountName.value = "";
  newAccountPassword.value = "";
}

async function handleAIConfigSubmit() {
  aiFormError.value = "";

  if (!aiBaseUrl.value.trim()) {
    aiFormError.value = "请输入兼容接口的 API URL。";
    return;
  }

  if (!aiModel.value.trim()) {
    aiFormError.value = "请输入默认模型名称。";
    return;
  }

  if (!aiApiKey.value.trim() && !workspaceStore.aiConfigSummary.keyConfigured) {
    aiFormError.value = "请先输入 API Key。";
    return;
  }

  const success = await workspaceStore.saveAIServiceConfig({
    providerName: aiProviderName.value.trim() || AI_PROVIDER_NAME_FALLBACK,
    providerType: AI_PROVIDER_TYPE,
    baseUrl: aiBaseUrl.value.trim(),
    model: aiModel.value.trim() || AI_MODEL_FALLBACK,
    apiKey: aiApiKey.value.trim(),
  });

  if (!success) {
    aiFormError.value = workspaceStore.aiConfigError;
    return;
  }

  aiApiKey.value = "";
}
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 pt-4">
    <div class="max-w-2xl">
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">设置</h2>
    </div>

    <div class="space-y-12">
      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 md:grid-cols-[minmax(0,13rem)_1fr]"
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
              <div class="flex flex-wrap items-center gap-3">
                <span
                  class="inline-flex shrink-0 items-center rounded-full bg-stone-100 px-2.5 py-0.5 text-xs font-medium whitespace-nowrap text-stone-800"
                >
                  {{ workspaceStore.isAdmin ? "管理员" : "成员" }}
                </span>
                <button
                  type="button"
                  class="ui-btn-secondary px-3 py-1.5 text-sm"
                  @click="handleLogout"
                >
                  退出
                </button>
              </div>
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
            <div
              v-if="workspaceStore.isAdmin"
              class="rounded-xl border border-stone-200 bg-stone-50 p-4"
            >
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p class="text-sm font-medium text-stone-900">创建新账号</p>
                  <p class="mt-1 text-sm text-stone-500">
                    为成员创建独立账号，登录后会加载各自的工作区。
                  </p>
                </div>
                <span
                  class="inline-flex shrink-0 self-start rounded-full bg-amber-100 px-2.5 py-0.5 text-xs font-medium whitespace-nowrap text-amber-800"
                >
                  管理员
                </span>
              </div>

              <form class="mt-4 space-y-4" @submit.prevent="handleCreateAccountSubmit">
                <div class="grid gap-4 md:grid-cols-3">
                  <label class="block">
                    <span class="mb-2 block text-sm font-medium text-stone-900">账号</span>
                    <input
                      v-model="newAccountEmail"
                      type="text"
                      autocomplete="username"
                      placeholder="例如：alice"
                      class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                    />
                  </label>
                  <label class="block">
                    <span class="mb-2 block text-sm font-medium text-stone-900">显示名称</span>
                    <input
                      v-model="newAccountName"
                      type="text"
                      placeholder="例如：Alice"
                      class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                    />
                  </label>
                  <label class="block">
                    <span class="mb-2 block text-sm font-medium text-stone-900">初始密码</span>
                    <div
                      class="flex items-center gap-2 rounded-lg border border-stone-200 bg-white px-3"
                    >
                      <input
                        v-model="newAccountPassword"
                        :type="newAccountPasswordVisible ? 'text' : 'password'"
                        autocomplete="new-password"
                        class="min-w-0 flex-1 bg-transparent py-2 text-sm text-stone-900 focus:outline-none"
                      />
                      <button
                        type="button"
                        class="shrink-0 text-xs font-medium text-stone-500 hover:text-stone-900"
                        @click="newAccountPasswordVisible = !newAccountPasswordVisible"
                      >
                        {{ newAccountPasswordVisible ? "隐藏" : "显示" }}
                      </button>
                    </div>
                  </label>
                </div>
                <p
                  v-if="newAccountFormError"
                  class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700"
                >
                  {{ newAccountFormError }}
                </p>
                <button
                  type="submit"
                  class="ui-btn-primary px-4 py-2 text-sm"
                  :disabled="workspaceStore.accountCreationLoading"
                >
                  {{ workspaceStore.accountCreationLoading ? "创建中..." : "创建账号" }}
                </button>
              </form>
            </div>
          </div>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">打印设置</h3>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div v-if="workspaceStore.workspaceSyncError" class="rounded-xl bg-amber-50 p-4">
              <p class="text-sm font-medium text-amber-900">账号数据同步异常</p>
              <p class="mt-1 text-sm text-amber-700">
                {{ workspaceStore.workspaceSyncError }}
              </p>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">默认设备</p>
              </div>
              <select
                :value="workspaceStore.defaultDeviceId"
                :disabled="workspaceStore.devices.every((device) => device.status === 'offline')"
                class="rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900"
                @change="handleDefaultDeviceChange"
              >
                <option
                  v-if="workspaceStore.devices.every((device) => device.status === 'offline')"
                  value=""
                >
                  暂未设置设备
                </option>
                <option
                  v-for="device in workspaceStore.devices.filter(
                    (device) => device.status !== 'offline',
                  )"
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
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 md:grid-cols-[minmax(0,13rem)_1fr]"
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
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">AI 服务</h3>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div
              v-if="workspaceStore.aiConfigLoading"
              class="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3"
            >
              <p class="text-sm text-stone-600">正在加载当前 AI 配置…</p>
            </div>

            <div class="rounded-xl border border-stone-200 bg-stone-50 p-4">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-stone-900">当前接入状态</p>
                  <p class="mt-1 text-sm text-stone-500">
                    {{
                      workspaceStore.aiConfigSummary.bound
                        ? "问答会通过服务端代理转发到你配置的 OpenAI 兼容模型。"
                        : "当前还没有可用于登录用户的真实 AI 服务。"
                    }}
                  </p>
                </div>
                <span
                  class="inline-flex items-center rounded-full bg-stone-100 px-2.5 py-0.5 text-xs font-medium text-stone-800"
                >
                  {{ workspaceStore.aiConfigSummary.bound ? "已连接" : "未连接" }}
                </span>
              </div>

              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <div>
                  <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                    服务商
                  </p>
                  <p class="mt-1 text-sm text-stone-900">
                    {{ workspaceStore.aiConfigSummary.providerName || AI_PROVIDER_NAME_FALLBACK }}
                  </p>
                </div>
                <div>
                  <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">模型</p>
                  <p class="mt-1 text-sm text-stone-900">
                    {{ workspaceStore.aiConfigSummary.model || AI_MODEL_FALLBACK }}
                  </p>
                </div>
                <div>
                  <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                    API URL
                  </p>
                  <p class="mt-1 text-sm break-all text-stone-900">
                    {{ workspaceStore.aiConfigSummary.baseUrl || "尚未设置" }}
                  </p>
                </div>
                <div>
                  <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                    API Key
                  </p>
                  <p class="mt-1 text-sm text-stone-900">
                    {{
                      workspaceStore.aiConfigSummary.keyConfigured
                        ? "已在服务端加密保存"
                        : "尚未配置"
                    }}
                  </p>
                </div>
              </div>
            </div>

            <form
              v-if="workspaceStore.isAdmin"
              class="rounded-xl border border-stone-200 bg-white p-4"
              @submit.prevent="handleAIConfigSubmit"
            >
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p class="text-sm font-medium text-stone-900">管理员配置</p>
                  <p class="mt-1 text-sm text-stone-500">
                    保存供应商、API URL、模型和服务端密钥。已保存的 Key 不会回显。
                  </p>
                </div>
                <span
                  class="inline-flex shrink-0 self-start rounded-full bg-amber-100 px-2.5 py-0.5 text-xs font-medium whitespace-nowrap text-amber-800"
                >
                  管理员
                </span>
              </div>

              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <label class="block">
                  <span class="mb-2 block text-sm font-medium text-stone-900">供应商名称</span>
                  <input
                    v-model="aiProviderName"
                    type="text"
                    placeholder="例如：OpenAI Compatible"
                    class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                  />
                </label>
                <label class="block">
                  <span class="mb-2 block text-sm font-medium text-stone-900">API URL</span>
                  <input
                    v-model="aiBaseUrl"
                    type="url"
                    placeholder="例如：https://api.openai.com/v1"
                    class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                  />
                </label>
                <label class="block">
                  <span class="mb-2 block text-sm font-medium text-stone-900">默认模型</span>
                  <input
                    v-model="aiModel"
                    type="text"
                    placeholder="例如：gpt-4.1-mini"
                    class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                  />
                </label>
                <label class="block">
                  <span class="mb-2 block text-sm font-medium text-stone-900">API Key</span>
                  <input
                    v-model="aiApiKey"
                    :placeholder="
                      workspaceStore.aiConfigSummary.keyConfigured
                        ? '留空则沿用当前服务端密钥'
                        : '输入新的服务端密钥'
                    "
                    type="password"
                    autocomplete="off"
                    class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                  />
                </label>
              </div>

              <p class="mt-4 text-sm text-stone-500">
                API Key 只会通过后端保存到服务端加密存储，前端只支持新增或替换，不支持读取回显。
              </p>

              <p
                v-if="aiConfigErrorMessage"
                class="mt-4 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700"
              >
                {{ aiConfigErrorMessage }}
              </p>

              <div class="mt-4 flex justify-end">
                <button
                  type="submit"
                  class="ui-btn-primary px-4 py-2 text-sm"
                  :disabled="workspaceStore.aiConfigSaving || workspaceStore.aiConfigLoading"
                >
                  {{ workspaceStore.aiConfigSaving ? "保存中..." : "保存 AI 配置" }}
                </button>
              </div>
            </form>

            <div v-else class="rounded-xl border border-stone-200 bg-stone-50 p-4">
              <p class="text-sm font-medium text-stone-900">仅管理员可修改</p>
              <p class="mt-1 text-sm text-stone-500">
                当前账号只能查看 AI 接入摘要，不能读取或编辑服务端保存的 API Key。
              </p>
            </div>
          </div>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 md:grid-cols-[minmax(0,13rem)_1fr]"
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

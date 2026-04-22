<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRouter } from "vue-router";

import AppDialog from "@/components/AppDialog.vue";
import { DEFAULT_LOGIN_REDIRECT } from "@/router/authRedirect";
import { storeToRefs } from "@/stores/pinia";
import { useWorkspaceStore } from "@/stores/workspace";
import type { PluginDetails, PluginFieldSpec } from "@/types/plugins";
import { getPluginFieldDefaultValue } from "@/utils/plugins";
import {
  getPluginBindingStatusBadgeClass,
  getPluginBindingStatusLabel,
  getPluginInstallationStatusBadgeClass,
  getPluginInstallationStatusLabel,
  getServiceBindingStatusBadgeClass,
  getThemeDescription,
  getUserRoleBadgeClass,
  getUserRoleLabel,
} from "@/utils/workspace";

const router = useRouter();
const workspaceStore = useWorkspaceStore();
const { adminPlugins, aiConfigSummary, availablePlugins } = storeToRefs(workspaceStore);

const themes = [
  { label: "浅色", value: "light" },
  { label: "深色", value: "dark" },
  { label: "跟随系统", value: "system" },
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
const passwordDialogOpen = ref(false);
const accountCreationDialogOpen = ref(false);
const aiConfigDialogOpen = ref(false);
const pluginUploadError = ref("");
const pluginGitInstallError = ref("");
const pluginGitRepoUrl = ref("");
const pluginGitRepoRef = ref("");
const pluginGitRepoSubdir = ref("");
const pluginGitInstallDialogOpen = ref(false);
const activePluginConfigId = ref<string | null>(null);
const pluginEnabledDrafts = ref<Record<string, boolean>>({});
const pluginDrafts = ref<Record<string, Record<string, unknown>>>({});
const pluginSecretDrafts = ref<Record<string, Record<string, string>>>({});
const pluginTestMessages = ref<Record<string, string>>({});

const AI_PROVIDER_TYPE = "openai-compatible";
const AI_PROVIDER_NAME_FALLBACK = "OpenAI Compatible";
const AI_MODEL_FALLBACK = "gpt-4.1-mini";

const aiConfigErrorMessage = computed(() => aiFormError.value || workspaceStore.aiConfigError);
const pluginInstallations = computed(() =>
  workspaceStore.isAdmin ? adminPlugins.value : availablePlugins.value,
);
const activePluginConfig = computed(() => {
  const installationId = activePluginConfigId.value;
  const plugins = availablePlugins.value;
  for (const plugin of plugins) {
    if (plugin.installation.id === installationId) {
      return plugin;
    }
  }
  return null;
});

void AppDialog;
void pluginInstallations.value;
void activePluginConfig.value;

watch(
  aiConfigSummary,
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

function createPluginDraftState(plugin: PluginDetails) {
  const config: Record<string, unknown> = {};
  const secrets: Record<string, string> = {};

  for (const field of plugin.manifest.workspaceConfigSchema) {
    if (field.type === "secret") {
      secrets[field.key] = "";
      continue;
    }

    config[field.key] = plugin.binding?.config?.[field.key] ?? getPluginFieldDefaultValue(field);
  }

  return {
    enabled: plugin.binding?.enabled ?? false,
    config,
    secrets,
  };
}

function resetPluginDraft(plugin: PluginDetails) {
  const nextState = createPluginDraftState(plugin);
  pluginEnabledDrafts.value[plugin.installation.id] = nextState.enabled;
  pluginDrafts.value[plugin.installation.id] = nextState.config;
  pluginSecretDrafts.value[plugin.installation.id] = nextState.secrets;
}

function ensurePluginDraft(plugin: PluginDetails) {
  pluginEnabledDrafts.value[plugin.installation.id] =
    pluginEnabledDrafts.value[plugin.installation.id] ?? plugin.binding?.enabled ?? false;

  const nextDraft = { ...pluginDrafts.value[plugin.installation.id] };
  for (const field of plugin.manifest.workspaceConfigSchema) {
    if (field.type === "secret" || field.key in nextDraft) {
      continue;
    }
    nextDraft[field.key] = plugin.binding?.config?.[field.key] ?? getPluginFieldDefaultValue(field);
  }
  pluginDrafts.value[plugin.installation.id] = nextDraft;

  const nextSecrets = { ...pluginSecretDrafts.value[plugin.installation.id] };
  for (const field of plugin.manifest.workspaceConfigSchema) {
    if (field.type !== "secret" || field.key in nextSecrets) {
      continue;
    }
    nextSecrets[field.key] = "";
  }
  pluginSecretDrafts.value[plugin.installation.id] = nextSecrets;
}

watch(
  () => workspaceStore.availablePlugins,
  (plugins) => {
    const activeIDs = new Set(plugins.map((plugin) => plugin.installation.id));
    for (const drafts of [
      pluginEnabledDrafts.value,
      pluginDrafts.value,
      pluginSecretDrafts.value,
    ]) {
      for (const installationID of Object.keys(drafts)) {
        if (!activeIDs.has(installationID)) {
          delete drafts[installationID];
        }
      }
    }
    for (const plugin of plugins) {
      resetPluginDraft(plugin);
    }
  },
  { deep: true, immediate: true },
);

function pluginDraftValue(plugin: PluginDetails, field: PluginFieldSpec) {
  return field.type === "secret"
    ? (pluginSecretDrafts.value[plugin.installation.id]?.[field.key] ?? "")
    : (pluginDrafts.value[plugin.installation.id]?.[field.key] ??
        getPluginFieldDefaultValue(field));
}

function updatePluginDraft(plugin: PluginDetails, field: PluginFieldSpec, value: unknown) {
  ensurePluginDraft(plugin);

  if (field.type === "secret") {
    pluginSecretDrafts.value[plugin.installation.id] = {
      ...pluginSecretDrafts.value[plugin.installation.id],
      [field.key]: String(value ?? ""),
    };
    return;
  }

  pluginDrafts.value[plugin.installation.id] = {
    ...pluginDrafts.value[plugin.installation.id],
    [field.key]:
      field.type === "number" && value !== ""
        ? Number(value)
        : field.type === "checkbox"
          ? Boolean(value)
          : value,
  };
}

function setPluginEnabled(plugin: PluginDetails, enabled: boolean) {
  ensurePluginDraft(plugin);
  pluginEnabledDrafts.value[plugin.installation.id] = enabled;
}

function handlePluginEnabledChange(plugin: PluginDetails, event: Event) {
  const target = event.target as HTMLInputElement | null;
  setPluginEnabled(plugin, target?.checked ?? false);
}

function handlePluginInput(plugin: PluginDetails, field: PluginFieldSpec, event: Event) {
  const target = event.target as HTMLInputElement | HTMLTextAreaElement | null;
  updatePluginDraft(plugin, field, target?.value ?? "");
}

function handlePluginSelect(plugin: PluginDetails, field: PluginFieldSpec, event: Event) {
  const target = event.target as HTMLSelectElement | null;
  updatePluginDraft(plugin, field, target?.value ?? "");
}

function handlePluginCheckbox(plugin: PluginDetails, field: PluginFieldSpec, event: Event) {
  const target = event.target as HTMLInputElement | null;
  updatePluginDraft(plugin, field, target?.checked ?? false);
}

async function handlePluginUpload(event: Event) {
  pluginUploadError.value = "";
  const target = event.target as HTMLInputElement | null;
  const file = target?.files?.[0];

  if (!file) {
    return;
  }

  const uploaded = await workspaceStore.uploadPlugin(file);
  if (!uploaded) {
    pluginUploadError.value = workspaceStore.pluginActionError;
  }

  if (target) {
    target.value = "";
  }
}

async function handlePluginInstallFromGit() {
  pluginGitInstallError.value = "";

  const repoUrl = pluginGitRepoUrl.value.trim();
  const repoRef = pluginGitRepoRef.value.trim();
  const repoSubdir = pluginGitRepoSubdir.value.trim();

  if (!repoUrl) {
    pluginGitInstallError.value = "请输入 Git 仓库地址。";
    return;
  }

  const installed = await workspaceStore.installPluginRepository({
    repoUrl,
    repoRef,
    repoSubdir,
  });
  if (!installed) {
    pluginGitInstallError.value = workspaceStore.pluginActionError;
    return;
  }

  pluginGitRepoUrl.value = "";
  pluginGitRepoRef.value = "";
  pluginGitRepoSubdir.value = "";
  pluginGitInstallDialogOpen.value = false;
}

async function handlePluginDisable(installationId: string) {
  await workspaceStore.disablePluginInstallation(installationId);
}

function openPluginGitInstallDialog() {
  pluginGitInstallError.value = "";
  pluginGitInstallDialogOpen.value = true;
}

function closePluginGitInstallDialog() {
  pluginGitInstallDialogOpen.value = false;
}

function openPluginConfigDialog(plugin: PluginDetails) {
  ensurePluginDraft(plugin);
  activePluginConfigId.value = plugin.installation.id;
}

function closePluginConfigDialog() {
  activePluginConfigId.value = null;
}

async function handlePluginTest(plugin: PluginDetails) {
  ensurePluginDraft(plugin);
  const result = await workspaceStore.testPluginConfiguration(
    plugin.installation.id,
    pluginDrafts.value[plugin.installation.id] ?? {},
    pluginSecretDrafts.value[plugin.installation.id] ?? {},
    pluginEnabledDrafts.value[plugin.installation.id] ?? false,
  );

  pluginTestMessages.value[plugin.installation.id] = result?.valid
    ? "连接测试通过。"
    : result?.errors?.map((error) => error.message).join("；") || workspaceStore.pluginActionError;
}

async function handlePluginSave(plugin: PluginDetails) {
  ensurePluginDraft(plugin);
  const saved = await workspaceStore.savePluginConfiguration(
    plugin.installation.id,
    pluginDrafts.value[plugin.installation.id] ?? {},
    pluginSecretDrafts.value[plugin.installation.id] ?? {},
    pluginEnabledDrafts.value[plugin.installation.id] ?? false,
  );

  if (saved) {
    pluginSecretDrafts.value[plugin.installation.id] = createPluginDraftState(plugin).secrets;
    closePluginConfigDialog();
  }
}

async function handleLogout() {
  await workspaceStore.logout();
  await router.replace(DEFAULT_LOGIN_REDIRECT);
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
  passwordDialogOpen.value = false;
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
  accountCreationDialogOpen.value = false;
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
  aiConfigDialogOpen.value = false;
}

function openAIConfigDialog() {
  aiFormError.value = "";
  aiConfigDialogOpen.value = true;
}

function closeAIConfigDialog() {
  aiConfigDialogOpen.value = false;
}

function resetPasswordForm() {
  currentPassword.value = "";
  newPassword.value = "";
  confirmPassword.value = "";
  passwordFormError.value = "";
  currentPasswordVisible.value = false;
  newPasswordVisible.value = false;
  confirmPasswordVisible.value = false;
}

function openPasswordDialog() {
  passwordFormError.value = "";
  passwordDialogOpen.value = true;
}

function closePasswordDialog() {
  passwordDialogOpen.value = false;
  resetPasswordForm();
}

function resetAccountCreationForm() {
  newAccountEmail.value = "";
  newAccountName.value = "";
  newAccountPassword.value = "";
  newAccountFormError.value = "";
  newAccountPasswordVisible.value = false;
}

function openAccountCreationDialog() {
  newAccountFormError.value = "";
  accountCreationDialogOpen.value = true;
}

function closeAccountCreationDialog() {
  accountCreationDialogOpen.value = false;
  resetAccountCreationForm();
}

void openPluginGitInstallDialog;
void closePluginGitInstallDialog;
void openPluginConfigDialog;
void openAIConfigDialog;
void closeAIConfigDialog;
void openPasswordDialog;
void closePasswordDialog;
void openAccountCreationDialog;
void closeAccountCreationDialog;
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
                  class="ui-status-badge"
                  :class="getUserRoleBadgeClass(workspaceStore.isAdmin ? 'admin' : 'member')"
                >
                  {{ getUserRoleLabel(workspaceStore.isAdmin ? "admin" : "member") }}
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
                      ? "关闭浏览器后需要重新登录"
                      : "关闭浏览器后保留登录状态"
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
            <div class="rounded-xl border border-stone-200 bg-white p-4">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p class="text-sm font-medium text-stone-900">修改密码</p>
                  <p class="mt-1 text-sm text-stone-500">
                    密码编辑收进独立窗口，设置页只保留安全状态摘要。
                  </p>
                </div>
                <button
                  type="button"
                  class="ui-btn-primary px-4 py-2 text-sm"
                  @click="openPasswordDialog"
                >
                  修改密码
                </button>
              </div>
              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <div class="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
                  <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                    安全规则
                  </p>
                  <p class="mt-1 text-sm text-stone-900">新密码至少 8 位</p>
                </div>
                <div class="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
                  <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                    提交结果
                  </p>
                  <p class="mt-1 text-sm text-stone-900">更新后会跳回登录页重新认证</p>
                </div>
              </div>
            </div>
            <div
              v-if="workspaceStore.isAdmin"
              class="rounded-xl border border-stone-200 bg-white p-4"
            >
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p class="text-sm font-medium text-stone-900">创建新账号</p>
                  <p class="mt-1 text-sm text-stone-500">
                    为成员创建独立账号，登录后会加载各自的工作区。
                  </p>
                </div>
                <div class="flex items-center gap-3">
                  <span class="ui-status-badge self-start" :class="getUserRoleBadgeClass('admin')">
                    {{ getUserRoleLabel("admin") }}
                  </span>
                  <button
                    type="button"
                    class="ui-btn-primary px-4 py-2 text-sm"
                    @click="openAccountCreationDialog"
                  >
                    创建账号
                  </button>
                </div>
              </div>
              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <div class="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
                  <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                    账号形式
                  </p>
                  <p class="mt-1 text-sm text-stone-900">推荐使用简短用户名，例如 alice</p>
                </div>
                <div class="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
                  <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                    初始权限
                  </p>
                  <p class="mt-1 text-sm text-stone-900">新账号默认创建为成员角色</p>
                </div>
              </div>
            </div>

            <AppDialog
              :open="passwordDialogOpen"
              title="修改密码"
              description="输入当前密码并设置新的登录密码。提交成功后会回到登录页重新认证。"
              @close="closePasswordDialog"
            >
              <form class="space-y-4" @submit.prevent="handlePasswordSubmit">
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
                <div class="flex justify-end gap-3">
                  <button
                    type="button"
                    class="ui-btn-secondary px-4 py-2 text-sm"
                    @click="closePasswordDialog"
                  >
                    取消
                  </button>
                  <button
                    type="submit"
                    class="ui-btn-primary px-4 py-2 text-sm"
                    :disabled="workspaceStore.passwordChangeLoading"
                  >
                    {{ workspaceStore.passwordChangeLoading ? "提交中..." : "更新密码" }}
                  </button>
                </div>
              </form>
            </AppDialog>

            <AppDialog
              :open="accountCreationDialogOpen"
              title="创建新账号"
              description="为成员创建独立登录账号，提交后会立即同步到当前工作区。"
              @close="closeAccountCreationDialog"
            >
              <form class="space-y-4" @submit.prevent="handleCreateAccountSubmit">
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
                <div class="flex justify-end gap-3">
                  <button
                    type="button"
                    class="ui-btn-secondary px-4 py-2 text-sm"
                    @click="closeAccountCreationDialog"
                  >
                    取消
                  </button>
                  <button
                    type="submit"
                    class="ui-btn-primary px-4 py-2 text-sm"
                    :disabled="workspaceStore.accountCreationLoading"
                  >
                    {{ workspaceStore.accountCreationLoading ? "创建中..." : "创建账号" }}
                  </button>
                </div>
              </form>
            </AppDialog>
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
                <p class="text-sm font-medium text-stone-900">教程标签</p>
                <p class="mt-0.5 text-sm text-stone-500">
                  {{
                    workspaceStore.tutorialTabEnabled
                      ? "顶部和底部导航会显示“教程”标签。"
                      : "顶部和底部导航会隐藏“教程”标签。"
                  }}
                </p>
              </div>
              <button
                type="button"
                class="ui-toggle"
                :class="{ 'is-on': workspaceStore.tutorialTabEnabled }"
                :aria-label="`${workspaceStore.tutorialTabEnabled ? '关闭' : '开启'}教程标签`"
                :aria-pressed="workspaceStore.tutorialTabEnabled"
                @click="workspaceStore.setTutorialTabEnabled(!workspaceStore.tutorialTabEnabled)"
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
          <p class="mt-1 text-sm text-stone-500">
            选择“跟随系统”后，会自动根据设备当前的深浅色设置切换。
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
                  class="ui-status-badge"
                  :class="getServiceBindingStatusBadgeClass(workspaceStore.aiConfigSummary.bound)"
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

            <div
              v-if="workspaceStore.isAdmin"
              class="rounded-xl border border-stone-200 bg-white p-4"
            >
              <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <p class="text-sm font-medium text-stone-900">管理员配置</p>
                  <p class="mt-1 text-sm text-stone-500">
                    编辑动作收进独立窗口，设置页只保留当前服务摘要。
                  </p>
                </div>
                <div class="flex items-center gap-3">
                  <span class="ui-status-badge self-start" :class="getUserRoleBadgeClass('admin')">
                    {{ getUserRoleLabel("admin") }}
                  </span>
                  <button
                    type="button"
                    class="ui-btn-primary px-4 py-2 text-sm"
                    @click="openAIConfigDialog"
                  >
                    编辑 AI 配置
                  </button>
                </div>
              </div>
            </div>

            <div v-else class="rounded-xl border border-stone-200 bg-stone-50 p-4">
              <p class="text-sm font-medium text-stone-900">仅管理员可修改</p>
              <p class="mt-1 text-sm text-stone-500">
                当前账号只能查看 AI 接入摘要，不能读取或编辑服务端保存的 API Key。
              </p>
            </div>

            <AppDialog
              :open="aiConfigDialogOpen"
              title="编辑 AI 配置"
              description="保存供应商、API URL、模型和服务端密钥。已保存的 Key 不会回显。"
              @close="closeAIConfigDialog"
            >
              <form class="space-y-4" @submit.prevent="handleAIConfigSubmit">
                <div class="grid gap-4 md:grid-cols-2">
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

                <p class="text-sm text-stone-500">
                  API Key 只会通过后端保存到服务端加密存储，前端只支持新增或替换，不支持读取回显。
                </p>

                <p
                  v-if="aiConfigErrorMessage"
                  class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700"
                >
                  {{ aiConfigErrorMessage }}
                </p>

                <div class="flex justify-end gap-3">
                  <button
                    type="button"
                    class="ui-btn-secondary px-4 py-2 text-sm"
                    @click="closeAIConfigDialog"
                  >
                    取消
                  </button>
                  <button
                    type="submit"
                    class="ui-btn-primary px-4 py-2 text-sm"
                    :disabled="workspaceStore.aiConfigSaving || workspaceStore.aiConfigLoading"
                  >
                    {{ workspaceStore.aiConfigSaving ? "保存中..." : "保存 AI 配置" }}
                  </button>
                </div>
              </form>
            </AppDialog>
          </div>
        </div>
      </article>

      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">插件</h3>
        </div>
        <div class="min-w-0 space-y-4">
          <template v-if="workspaceStore.isAuthenticated">
            <section
              v-if="workspaceStore.isAdmin"
              class="rounded-2xl border border-stone-200 bg-stone-50 px-5 py-4"
            >
              <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p class="text-sm font-medium text-stone-900">本地上传插件 ZIP</p>
                  <p class="mt-1 text-sm text-stone-500">
                    上传后会自动校验 manifest、安装依赖，并把成功版本切换为当前可用版本。
                  </p>
                </div>
                <label class="ui-btn-secondary inline-flex cursor-pointer px-4 py-2 text-sm">
                  <input
                    type="file"
                    accept=".zip,application/zip"
                    class="hidden"
                    :disabled="workspaceStore.pluginUploadLoading"
                    @change="handlePluginUpload"
                  />
                  {{ workspaceStore.pluginUploadLoading ? "上传中..." : "选择本地 ZIP" }}
                </label>
              </div>

              <p
                v-if="workspaceStore.pluginUploadLoading && workspaceStore.pluginUploadingName"
                class="mt-3 text-sm text-stone-500"
              >
                正在处理 {{ workspaceStore.pluginUploadingName }}
              </p>

              <p
                v-if="pluginUploadError"
                class="mt-3 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700"
              >
                {{ pluginUploadError }}
              </p>
            </section>

            <section
              v-if="workspaceStore.isAdmin"
              class="rounded-2xl border border-stone-200 bg-stone-50 px-5 py-4"
            >
              <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p class="text-sm font-medium text-stone-900">从 Git 仓库安装插件</p>
                  <p class="mt-1 text-sm text-stone-500">
                    改为在弹窗里填写仓库、分支和子目录，避免设置页被长表单撑开。
                  </p>
                </div>
                <button
                  type="button"
                  class="ui-btn-secondary px-4 py-2 text-sm"
                  @click="openPluginGitInstallDialog"
                >
                  打开安装窗口
                </button>
              </div>
            </section>

            <section class="ui-settings-group">
              <div class="ui-settings-row !items-start">
                <div class="ui-settings-copy">
                  <p class="text-sm font-medium text-stone-900">插件工作台</p>
                  <p class="mt-0.5 text-sm text-stone-500">
                    页面只保留摘要和入口，具体安装与配置在独立窗口中完成，避免插件多时界面臃肿。
                  </p>
                </div>
              </div>

              <div v-if="pluginInstallations.length === 0" class="ui-settings-row !items-start">
                <div class="ui-settings-copy">
                  <p class="text-sm font-medium text-stone-900">还没有已安装插件</p>
                  <p class="mt-0.5 text-sm text-stone-500">
                    {{
                      workspaceStore.isAdmin
                        ? "先上传 ZIP 或从 Git 安装一个插件。"
                        : "管理员安装插件后，这里会出现插件卡片。"
                    }}
                  </p>
                </div>
              </div>

              <div
                v-for="plugin in pluginInstallations"
                :key="plugin.installation.id"
                class="rounded-2xl border border-stone-200 bg-white px-5 py-5 shadow-xs"
              >
                <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                  <div class="min-w-0 flex-1">
                    <div class="flex flex-wrap items-center gap-2">
                      <p class="text-sm font-medium text-stone-900">
                        {{ plugin.installation.displayName }}
                      </p>
                      <span
                        class="ui-status-badge"
                        :class="getPluginInstallationStatusBadgeClass(plugin.installation.status)"
                      >
                        {{ getPluginInstallationStatusLabel(plugin.installation.status) }}
                      </span>
                      <span
                        v-if="plugin.binding"
                        class="ui-status-badge"
                        :class="getPluginBindingStatusBadgeClass(plugin)"
                      >
                        {{ getPluginBindingStatusLabel(plugin) }}
                      </span>
                    </div>
                    <p class="mt-1 text-sm text-stone-500">
                      {{ plugin.manifest.description || plugin.installation.description }}
                    </p>
                    <div class="mt-4 grid gap-3 md:grid-cols-3">
                      <div class="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
                        <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                          来源
                        </p>
                        <p class="mt-1 text-sm text-stone-900">
                          {{ plugin.installation.sourceType === "git" ? "Git 仓库" : "本地上传" }}
                        </p>
                        <p class="mt-1 text-xs text-stone-500">
                          {{ plugin.installation.runtimeType === "node" ? "Node.js" : "Python" }} ·
                          v{{ plugin.installation.version }}
                        </p>
                      </div>
                      <div class="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
                        <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                          抓取节奏
                        </p>
                        <p class="mt-1 text-sm text-stone-900">
                          每 {{ plugin.manifest.fetchPolicy.minutes }} 分钟抓取一次
                        </p>
                        <p class="mt-1 text-xs text-stone-500">
                          {{
                            plugin.binding?.nextFetchAt
                              ? `下次抓取 ${new Date(plugin.binding.nextFetchAt).toLocaleString()}`
                              : "尚未启用工作区绑定"
                          }}
                        </p>
                      </div>
                      <div class="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
                        <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                          标识
                        </p>
                        <p class="mt-1 text-sm text-stone-900">
                          {{ plugin.installation.pluginKey }}
                        </p>
                        <p class="mt-1 text-xs break-all text-stone-500">
                          {{ plugin.installation.repoUrl || "无 Git 地址" }}
                        </p>
                      </div>
                    </div>
                    <p
                      v-if="plugin.binding?.lastError || plugin.installation.lastError"
                      class="mt-4 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700"
                    >
                      {{ plugin.binding?.lastError || plugin.installation.lastError }}
                    </p>
                  </div>

                  <div class="flex shrink-0 flex-wrap items-center gap-3">
                    <button
                      v-if="
                        workspaceStore.availablePlugins.some(
                          (item) => item.installation.id === plugin.installation.id,
                        )
                      "
                      type="button"
                      class="ui-btn-primary px-4 py-2 text-sm"
                      @click="openPluginConfigDialog(plugin)"
                    >
                      配置插件
                    </button>
                    <button
                      v-if="plugin.installation.status !== 'disabled' && workspaceStore.isAdmin"
                      type="button"
                      class="ui-btn-secondary px-3 py-1.5 text-sm"
                      :disabled="
                        workspaceStore.pluginSaving &&
                        workspaceStore.pluginSavingId === plugin.installation.id
                      "
                      @click="handlePluginDisable(plugin.installation.id)"
                    >
                      {{
                        workspaceStore.pluginSaving &&
                        workspaceStore.pluginSavingId === plugin.installation.id
                          ? "停用中..."
                          : "停用"
                      }}
                    </button>
                  </div>
                </div>
              </div>
            </section>

            <AppDialog
              :open="pluginGitInstallDialogOpen"
              title="从 Git 安装插件"
              description="推荐直接安装单仓库单插件。只有仓库里放了多个插件时，才需要填写插件子目录。"
              @close="closePluginGitInstallDialog"
            >
              <form class="space-y-4" @submit.prevent="handlePluginInstallFromGit">
                <div class="grid gap-4">
                  <label class="block">
                    <span class="mb-2 block text-sm font-medium text-stone-900">仓库地址</span>
                    <input
                      v-model="pluginGitRepoUrl"
                      type="url"
                      placeholder="例如：https://github.com/MilkTeaFun/Ink-plugin.git"
                      class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                    />
                  </label>

                  <label class="block">
                    <span class="mb-2 block text-sm font-medium text-stone-900">
                      分支 / Tag / Commit
                    </span>
                    <input
                      v-model="pluginGitRepoRef"
                      type="text"
                      placeholder="默认 main"
                      class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                    />
                  </label>

                  <label class="block">
                    <span class="mb-2 block text-sm font-medium text-stone-900">
                      插件子目录（高级可选）
                    </span>
                    <input
                      v-model="pluginGitRepoSubdir"
                      type="text"
                      placeholder="例如：plugins/acme-source"
                      class="w-full rounded-lg border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                    />
                  </label>
                </div>

                <p class="text-sm text-stone-500">
                  推荐留空。只有仓库根目录不是插件目录时，才填写“插件子目录”。
                </p>

                <p
                  v-if="pluginGitInstallError"
                  class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700"
                >
                  {{ pluginGitInstallError }}
                </p>

                <div class="flex justify-end gap-3">
                  <button
                    type="button"
                    class="ui-btn-secondary px-4 py-2 text-sm"
                    @click="closePluginGitInstallDialog"
                  >
                    取消
                  </button>
                  <button
                    type="submit"
                    class="ui-btn-primary px-4 py-2 text-sm"
                    :disabled="workspaceStore.pluginGitInstallLoading"
                  >
                    {{ workspaceStore.pluginGitInstallLoading ? "安装中..." : "从 Git 安装" }}
                  </button>
                </div>
              </form>
            </AppDialog>

            <AppDialog
              :open="activePluginConfig !== null"
              :title="activePluginConfig?.installation.displayName ?? '插件配置'"
              :description="activePluginConfig?.manifest.description"
              @close="closePluginConfigDialog"
            >
              <template v-if="activePluginConfig">
                <div class="grid gap-3 md:grid-cols-2">
                  <div class="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
                    <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                      安装状态
                    </p>
                    <div class="mt-2 flex flex-wrap items-center gap-2">
                      <span
                        class="ui-status-badge"
                        :class="
                          getPluginInstallationStatusBadgeClass(
                            activePluginConfig.installation.status,
                          )
                        "
                      >
                        {{
                          getPluginInstallationStatusLabel(activePluginConfig.installation.status)
                        }}
                      </span>
                      <span
                        class="ui-status-badge"
                        :class="getPluginBindingStatusBadgeClass(activePluginConfig)"
                      >
                        {{ getPluginBindingStatusLabel(activePluginConfig) }}
                      </span>
                    </div>
                  </div>
                  <div class="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
                    <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
                      抓取策略
                    </p>
                    <p class="mt-2 text-sm text-stone-900">
                      每 {{ activePluginConfig.manifest.fetchPolicy.minutes }} 分钟抓取一次
                    </p>
                    <p class="mt-1 text-xs text-stone-500">
                      {{
                        activePluginConfig.binding?.nextFetchAt
                          ? `下次抓取 ${new Date(activePluginConfig.binding.nextFetchAt).toLocaleString()}`
                          : "当前未安排自动抓取"
                      }}
                    </p>
                  </div>
                </div>

                <form class="mt-5 space-y-4" @submit.prevent="handlePluginSave(activePluginConfig)">
                  <label
                    class="flex items-center justify-between gap-4 rounded-xl border border-stone-200 bg-stone-50 px-4 py-3"
                  >
                    <div>
                      <span class="block text-sm font-medium text-stone-900">
                        启用当前工作区绑定
                      </span>
                      <span class="mt-1 block text-sm text-stone-500">
                        启用后，这个工作区就可以测试该插件并创建相关定时任务。
                      </span>
                    </div>
                    <input
                      :checked="
                        pluginEnabledDrafts[activePluginConfig.installation.id] ??
                        activePluginConfig.binding?.enabled ??
                        false
                      "
                      type="checkbox"
                      class="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-stone-900"
                      @change="handlePluginEnabledChange(activePluginConfig, $event)"
                    />
                  </label>

                  <div v-if="activePluginConfig.manifest.workspaceConfigSchema.length === 0">
                    <p
                      class="rounded-xl border border-dashed border-stone-200 px-4 py-3 text-sm text-stone-500"
                    >
                      这个插件没有工作区级配置项，直接测试或保存即可。
                    </p>
                  </div>

                  <div v-else class="grid gap-4 md:grid-cols-2">
                    <div
                      v-for="field in activePluginConfig.manifest.workspaceConfigSchema"
                      :key="field.key"
                      class="block"
                      :class="{ 'md:col-span-2': field.type === 'textarea' }"
                    >
                      <span class="mb-2 block text-sm font-medium text-stone-900">
                        {{ field.label }}
                        <span v-if="field.required" class="text-rose-500">*</span>
                      </span>

                      <textarea
                        v-if="field.type === 'textarea'"
                        :value="String(pluginDraftValue(activePluginConfig, field) ?? '')"
                        rows="4"
                        :placeholder="field.description || ''"
                        class="w-full rounded-xl border border-stone-200 bg-white px-4 py-3 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                        @input="handlePluginInput(activePluginConfig, field, $event)"
                      />

                      <select
                        v-else-if="field.type === 'select'"
                        :value="String(pluginDraftValue(activePluginConfig, field) ?? '')"
                        class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                        @change="handlePluginSelect(activePluginConfig, field, $event)"
                      >
                        <option
                          v-for="option in field.options ?? []"
                          :key="option.value"
                          :value="option.value"
                        >
                          {{ option.label }}
                        </option>
                      </select>

                      <label
                        v-else-if="field.type === 'checkbox'"
                        class="flex items-center gap-3 rounded-xl border border-stone-200 bg-white px-4 py-3"
                      >
                        <input
                          :checked="Boolean(pluginDraftValue(activePluginConfig, field))"
                          type="checkbox"
                          class="h-4 w-4 rounded border-stone-300 text-stone-900 focus:ring-stone-900"
                          @change="handlePluginCheckbox(activePluginConfig, field, $event)"
                        />
                        <span class="text-sm text-stone-700">
                          {{ field.description || "启用此选项" }}
                        </span>
                      </label>

                      <input
                        v-else
                        :value="pluginDraftValue(activePluginConfig, field)"
                        :type="
                          field.type === 'secret'
                            ? 'password'
                            : field.type === 'number'
                              ? 'number'
                              : field.type === 'url'
                                ? 'url'
                                : 'text'
                        "
                        :placeholder="
                          field.type === 'secret' ? '留空则保持当前密钥' : field.description || ''
                        "
                        autocomplete="off"
                        class="w-full rounded-xl border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
                        @input="handlePluginInput(activePluginConfig, field, $event)"
                      />

                      <span v-if="field.description" class="mt-2 block text-xs text-stone-500">
                        {{ field.description }}
                      </span>
                    </div>
                  </div>

                  <p
                    v-if="pluginTestMessages[activePluginConfig.installation.id]"
                    class="rounded-lg bg-stone-100 px-3 py-2 text-sm text-stone-700"
                  >
                    {{ pluginTestMessages[activePluginConfig.installation.id] }}
                  </p>

                  <div class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
                    <button
                      type="button"
                      class="ui-btn-secondary px-4 py-2 text-sm"
                      :disabled="
                        workspaceStore.pluginTestingId === activePluginConfig.installation.id
                      "
                      @click="handlePluginTest(activePluginConfig)"
                    >
                      {{
                        workspaceStore.pluginTestingId === activePluginConfig.installation.id
                          ? "测试中..."
                          : "测试插件"
                      }}
                    </button>
                    <button
                      type="submit"
                      class="ui-btn-primary px-4 py-2 text-sm"
                      :disabled="
                        workspaceStore.pluginSavingId === activePluginConfig.installation.id
                      "
                    >
                      {{
                        workspaceStore.pluginSavingId === activePluginConfig.installation.id
                          ? "保存中..."
                          : "保存配置"
                      }}
                    </button>
                  </div>
                </form>
              </template>
            </AppDialog>
          </template>

          <div v-else class="ui-settings-group">
            <div
              v-for="source in workspaceStore.activeSources"
              :key="source.id"
              class="ui-settings-row"
            >
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

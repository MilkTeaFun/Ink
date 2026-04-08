<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { resolveLoginRedirect } from "@/router/authRedirect";
import { useWorkspaceStore } from "@/stores/workspace";

const router = useRouter();
const route = useRoute();
const workspaceStore = useWorkspaceStore();

const email = ref("admin");
const password = ref("");
const formError = ref("");
const passwordVisible = ref(false);

const isFormValid = computed(
  () => email.value.trim().length > 0 && password.value.trim().length > 0,
);
const noticeMessage = computed(() =>
  route.query.notice === "password-updated" ? "密码已更新，请使用新密码重新登录。" : "",
);

async function handleSubmit() {
  formError.value = "";

  if (!isFormValid.value) {
    formError.value = "请输入账号和密码。";
    return;
  }

  const success = await workspaceStore.login(email.value.trim(), password.value.trim());

  if (!success) {
    formError.value = workspaceStore.authError;
    return;
  }

  await router.replace(resolveLoginRedirect(router, route.query.redirect));
}
</script>

<template>
  <div class="min-h-screen bg-white px-4 py-6 text-stone-900">
    <div
      class="mx-auto grid min-h-[calc(100vh-3rem)] max-w-6xl items-center gap-12 lg:grid-cols-[1.1fr_0.9fr]"
    >
      <section class="p-8 lg:p-10">
        <h1
          class="text-[clamp(2rem,5vw,3.5rem)] leading-[1.1] font-semibold tracking-tight text-stone-900"
        >
          <span class="block">打开 Ink</span>
          <span class="mt-2 block">继续你的纸条灵感</span>
        </h1>
      </section>

      <section class="rounded-2xl border border-stone-200 bg-white p-8 shadow-sm lg:p-10">
        <h2 class="text-xl font-semibold text-stone-900">登录账号</h2>

        <p
          v-if="noticeMessage"
          class="mt-4 rounded-lg border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm text-emerald-700"
        >
          {{ noticeMessage }}
        </p>

        <form class="mt-8 space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label for="email" class="mb-2 block text-sm font-medium text-stone-900">账号</label>
            <input
              id="email"
              v-model="email"
              type="text"
              placeholder="admin"
              class="w-full rounded-lg border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 transition-colors placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
            />
          </div>
          <div>
            <label for="password" class="mb-2 block text-sm font-medium text-stone-900">密码</label>
            <div class="flex items-center gap-2 rounded-lg border border-stone-200 bg-white px-3">
              <input
                id="password"
                v-model="password"
                :type="passwordVisible ? 'text' : 'password'"
                placeholder="请输入密码"
                class="min-w-0 flex-1 bg-transparent px-1 py-2.5 text-sm text-stone-900 transition-colors placeholder:text-stone-400 focus:outline-none"
              />
              <button
                type="button"
                class="shrink-0 text-xs font-medium text-stone-500 hover:text-stone-900"
                @click="passwordVisible = !passwordVisible"
              >
                {{ passwordVisible ? "隐藏" : "显示" }}
              </button>
            </div>
          </div>
          <p v-if="formError" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
            {{ formError }}
          </p>

          <div class="flex flex-col gap-3 sm:flex-row">
            <button
              class="ui-btn-primary w-full px-6 py-2.5 sm:w-auto"
              :disabled="workspaceStore.authLoading"
            >
              {{ workspaceStore.authLoading ? "登录中..." : "登录" }}
            </button>
          </div>
        </form>
      </section>
    </div>
  </div>
</template>

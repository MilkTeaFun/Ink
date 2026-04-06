<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";

import { useWorkspaceStore } from "@/stores/workspace";

const router = useRouter();
const route = useRoute();
const workspaceStore = useWorkspaceStore();

const email = ref("name@example.com");
const password = ref("demo-password");
const formError = ref("");

const isFormValid = computed(
  () => /\S+@\S+\.\S+/.test(email.value) && password.value.trim().length > 0,
);

async function handleSubmit() {
  formError.value = "";

  if (!isFormValid.value) {
    formError.value = "请输入有效邮箱和密码。";
    return;
  }

  const success = await workspaceStore.login(email.value.trim(), password.value.trim());

  if (!success) {
    formError.value = workspaceStore.authError;
    return;
  }

  const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/status";
  await router.replace(redirect === "/login" ? "/status" : redirect);
}
</script>

<template>
  <div class="min-h-screen bg-white px-4 py-6 text-stone-900">
    <div
      class="mx-auto grid min-h-[calc(100vh-3rem)] max-w-6xl items-center gap-12 lg:grid-cols-[1.1fr_0.9fr]"
    >
      <section class="p-8 lg:p-10">
        <p class="text-[0.7rem] font-medium tracking-[0.28em] text-stone-500 uppercase">欢迎回来</p>
        <h1
          class="mt-5 text-[clamp(2rem,5vw,3.5rem)] leading-[1.1] font-semibold tracking-tight text-stone-900"
        >
          打开 Ink，继续你的纸条灵感。
        </h1>
        <p class="mt-6 max-w-xl text-base leading-relaxed text-stone-500">
          管理设备、整理内容、选好要打印的小纸条。整个过程会尽量保持简单，不需要你记很多技术细节。
        </p>
      </section>

      <section class="rounded-2xl border border-stone-200 bg-white p-8 shadow-sm lg:p-10">
        <h2 class="text-xl font-semibold text-stone-900">登录账号</h2>
        <p class="mt-1 text-sm text-stone-500">继续管理你的设备和打印内容。</p>

        <form class="mt-8 space-y-5" @submit.prevent="handleSubmit">
          <div>
            <label for="email" class="mb-2 block text-sm font-medium text-stone-900">邮箱</label>
            <input
              id="email"
              v-model="email"
              type="email"
              placeholder="name@example.com"
              class="w-full rounded-lg border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 transition-colors placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
            />
          </div>
          <div>
            <label for="password" class="mb-2 block text-sm font-medium text-stone-900">密码</label>
            <input
              id="password"
              v-model="password"
              type="password"
              placeholder="请输入密码"
              class="w-full rounded-lg border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 transition-colors placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
            />
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
            <RouterLink to="/" class="ui-btn-secondary w-full px-6 py-2.5 sm:w-auto"
              >先看看首页</RouterLink
            >
          </div>
        </form>

        <div
          class="mt-4 rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 text-sm text-stone-600"
        >
          这是无后端阶段的前端 mock 登录。输入任意有效邮箱即可登录，密码填 `wrong` 或邮箱包含 `fail`
          会触发错误提示。
        </div>
      </section>
    </div>
  </div>
</template>

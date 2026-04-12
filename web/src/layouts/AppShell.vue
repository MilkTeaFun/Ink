<script setup lang="ts">
import { computed } from "vue";
import { RouterLink, RouterView, useRoute, useRouter } from "vue-router";

import AppDialog from "@/components/AppDialog.vue";
import { DEFAULT_LOGIN_REDIRECT } from "@/router/authRedirect";
import { navigationItems } from "@/router";
import { useWorkspaceStore } from "@/stores/workspace";

const route = useRoute();
const router = useRouter();
const workspaceStore = useWorkspaceStore();
const postLoginTutorialSteps = [
  {
    title: "双击开机键，先打印状态纸条",
    detail: "只看 Device ID 那一行，别把 WiFi Name 或 MAC Address 填进设备编号。",
  },
  {
    title: "回到 Ink，把 Device ID 完整填进添加设备",
    detail: "设备名称和备注按习惯填写，真正决定能否绑定成功的是那串 Device ID。",
  },
  {
    title: "绑定后设为默认，再去试打一张",
    detail: "之后你在对话页生成的新内容，就会按当前默认设置直接进入打印队列。",
  },
] as const;

const pendingBadge = computed(() =>
  workspaceStore.pendingConfirmationCount > 0 ? workspaceStore.pendingConfirmationCount : "",
);
const anonymousDemoRouteNames = new Set(["status", "conversations", "prints"]);
const loginTarget = computed(() => ({
  path: "/login",
  query: route.fullPath === DEFAULT_LOGIN_REDIRECT ? undefined : { redirect: route.fullPath },
}));
const showAnonymousDemoBanner = computed(
  () =>
    !workspaceStore.isAuthenticated &&
    anonymousDemoRouteNames.has(String(route.name ?? "")),
);
const visibleNavigationItems = computed(() =>
  navigationItems.filter((item) => item.name !== "tutorial" || workspaceStore.tutorialTabEnabled),
);

function closePostLoginTutorial() {
  workspaceStore.closePostLoginTutorial();
}

async function handlePostLoginTutorialNavigate(path: string) {
  closePostLoginTutorial();
  await router.push(path);
}

async function handleLogout() {
  await workspaceStore.logout();
  await router.replace(DEFAULT_LOGIN_REDIRECT);
}
</script>

<template>
  <div class="flex min-h-[100dvh] flex-col bg-white text-stone-900">
    <AppDialog
      :open="workspaceStore.postLoginTutorialOpen"
      title="登录成功后先绑定设备"
      description="绑定教程现在统一放在这里。先按下面三步完成咕咕机绑定，之后你在对话页生成的内容就能直接发往默认设备。"
      @close="closePostLoginTutorial"
    >
      <div class="space-y-4">
        <article
          v-for="(step, index) in postLoginTutorialSteps"
          :key="step.title"
          class="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-3"
        >
          <p class="text-xs font-medium tracking-[0.12em] text-stone-500 uppercase">
            步骤 {{ index + 1 }}
          </p>
          <p class="mt-1 text-sm font-medium text-stone-900">{{ step.title }}</p>
          <p class="mt-1 text-sm leading-6 text-stone-600">{{ step.detail }}</p>
        </article>

        <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <button
            type="button"
            class="ui-btn-secondary px-4 py-2 text-sm"
            @click="closePostLoginTutorial"
          >
            稍后再看
          </button>
          <button
            type="button"
            class="ui-btn-primary px-4 py-2 text-sm"
            @click="handlePostLoginTutorialNavigate('/tutorial')"
          >
            去看教程
          </button>
        </div>
      </div>
    </AppDialog>

    <header
      class="sticky top-0 z-40 border-b border-stone-200 bg-white/90 px-4 pt-[calc(env(safe-area-inset-top)+0.75rem)] pb-3 backdrop-blur sm:px-5 lg:bg-white lg:px-8 lg:py-3"
    >
      <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 lg:hidden">
        <div class="flex items-center gap-3">
          <img src="/icon.jpg" alt="Ink Icon" class="h-8 w-8 rounded-lg object-contain" />
          <div class="flex items-center gap-2">
            <p class="text-sm font-semibold text-stone-950">Ink</p>
            <a
              href="https://github.com/ruhuang2001"
              target="_blank"
              rel="noreferrer"
              class="text-xs text-stone-400 transition-colors hover:text-stone-700"
            >
              Powered by ruhuang2001
            </a>
          </div>
          <div class="hidden sm:block">
            <p class="text-xs text-stone-500">
              {{ route.meta.navHint ?? route.meta.title ?? "纸条工作区" }}
            </p>
          </div>
        </div>

        <button
          v-if="workspaceStore.isAuthenticated"
          type="button"
          class="text-sm font-medium text-stone-600 hover:text-stone-900"
          @click="handleLogout"
        >
          退出
        </button>
        <RouterLink
          v-else
          :to="loginTarget"
          class="text-sm font-medium text-stone-600 hover:text-stone-900"
        >
          登录
        </RouterLink>
      </div>

      <div class="mx-auto hidden max-w-7xl items-center justify-between lg:flex">
        <div class="flex items-center gap-8">
          <div class="flex items-center gap-3">
            <img src="/icon.jpg" alt="Ink Icon" class="h-8 w-8 rounded-lg object-contain" />
            <p class="text-sm font-semibold text-stone-950">Ink</p>
            <a
              href="https://github.com/ruhuang2001"
              target="_blank"
              rel="noreferrer"
              class="text-xs text-stone-400 transition-colors hover:text-stone-700"
            >
              Powered by ruhuang2001
            </a>
          </div>

          <nav class="flex items-center gap-1">
            <RouterLink
              v-for="item in visibleNavigationItems"
              :key="item.name"
              :to="item.path"
              class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
              :class="
                route.name === item.name
                  ? 'bg-stone-100 text-stone-900'
                  : 'text-stone-600 hover:bg-stone-50 hover:text-stone-900'
              "
            >
              <span>{{ item.label }}</span>
              <span
                v-if="item.name === 'prints' && pendingBadge"
                class="ml-2 inline-flex min-w-5 shrink-0 items-center justify-center whitespace-nowrap rounded-full bg-stone-900 px-1.5 py-0.5 text-[11px] text-white"
              >
                {{ pendingBadge }}
              </span>
            </RouterLink>
          </nav>
        </div>

        <div class="flex items-center gap-4">
          <template v-if="workspaceStore.isAuthenticated">
            <p class="text-sm text-stone-500">{{ workspaceStore.authUser?.email }}</p>
            <button
              type="button"
              class="text-sm font-medium text-stone-600 hover:text-stone-900"
              @click="handleLogout"
            >
              退出
            </button>
          </template>
          <RouterLink
            v-else
            :to="loginTarget"
            class="text-sm font-medium text-stone-600 hover:text-stone-900"
          >
            登录
          </RouterLink>
        </div>
      </div>
    </header>

    <div
      v-if="workspaceStore.flashMessage"
      class="border-b px-4 py-3 text-sm lg:px-8"
      :class="
        workspaceStore.flashTone === 'success'
          ? 'border-emerald-100 bg-emerald-50 text-emerald-700'
          : workspaceStore.flashTone === 'error'
            ? 'border-rose-100 bg-rose-50 text-rose-700'
            : 'border-stone-200 bg-stone-50 text-stone-700'
      "
    >
      <div class="mx-auto max-w-7xl">
        {{ workspaceStore.flashMessage }}
      </div>
    </div>

    <div
      v-if="showAnonymousDemoBanner"
      class="border-b border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 lg:px-8"
    >
      <div class="mx-auto flex max-w-7xl flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p class="leading-6">
          当前设备、对话、打印页均为演示内容，具体使用请登录后继续。
        </p>
        <RouterLink
          :to="loginTarget"
          class="inline-flex items-center justify-center rounded-lg border border-amber-300 bg-white px-3 py-1.5 text-sm font-medium text-amber-900 transition-colors hover:border-amber-400 hover:bg-amber-100"
        >
          去登录
        </RouterLink>
      </div>
    </div>

    <main
      class="mx-auto w-full max-w-7xl flex-1 px-4 pt-4 pb-[calc(5.75rem+env(safe-area-inset-bottom))] sm:px-5 sm:pt-5 lg:px-8 lg:py-8"
    >
      <RouterView v-slot="{ Component, route: currentRoute }">
        <Transition name="page-swap" mode="out-in">
          <component :is="Component" :key="currentRoute.fullPath" />
        </Transition>
      </RouterView>
    </main>

    <nav
      class="fixed inset-x-0 bottom-0 z-30 border-t border-stone-200 bg-white/95 px-3 pt-2.5 pb-[calc(env(safe-area-inset-bottom)+0.5rem)] backdrop-blur lg:hidden"
    >
      <div
        class="mx-auto grid max-w-lg gap-1"
        :style="{ gridTemplateColumns: `repeat(${visibleNavigationItems.length}, minmax(0, 1fr))` }"
      >
        <RouterLink
          v-for="item in visibleNavigationItems"
          :key="item.name"
          :to="item.path"
          class="flex min-h-12 flex-col items-center justify-center rounded-xl px-2 py-2.5 text-center transition-colors"
          :class="
            route.name === item.name
              ? 'bg-stone-900 text-white shadow-sm'
              : 'text-stone-500 hover:bg-stone-50 hover:text-stone-900'
          "
        >
          <span class="block text-xs leading-tight font-medium">
            {{ item.label }}
            <span v-if="item.name === 'prints' && pendingBadge">· {{ pendingBadge }}</span>
          </span>
        </RouterLink>
      </div>
    </nav>
  </div>
</template>

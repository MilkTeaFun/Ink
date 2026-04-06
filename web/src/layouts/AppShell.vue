<script setup lang="ts">
import { computed } from "vue";
import { RouterLink, RouterView, useRoute, useRouter } from "vue-router";

import { navigationItems } from "@/router";
import { useWorkspaceStore } from "@/stores/workspace";

const route = useRoute();
const router = useRouter();
const workspaceStore = useWorkspaceStore();

const pendingBadge = computed(() =>
  workspaceStore.pendingConfirmationCount > 0 ? workspaceStore.pendingConfirmationCount : "",
);

async function handleLogout() {
  await workspaceStore.logout();
  await router.replace("/login");
}
</script>

<template>
  <div class="flex min-h-screen flex-col bg-white text-stone-900">
    <header class="sticky top-0 z-40 border-b border-stone-200 bg-white px-4 py-3 lg:px-8">
      <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 lg:hidden">
        <div class="flex items-center gap-3">
          <img src="/icon.jpg" alt="Ink Icon" class="h-8 w-8 rounded-lg object-contain" />
          <p class="text-sm font-semibold text-stone-950">Ink</p>
        </div>

        <button
          type="button"
          class="text-sm font-medium text-stone-600 hover:text-stone-900"
          @click="handleLogout"
        >
          退出
        </button>
      </div>

      <div class="mx-auto hidden max-w-7xl items-center justify-between lg:flex">
        <div class="flex items-center gap-8">
          <div class="flex items-center gap-3">
            <img src="/icon.jpg" alt="Ink Icon" class="h-8 w-8 rounded-lg object-contain" />
            <p class="text-sm font-semibold text-stone-950">Ink</p>
          </div>

          <nav class="flex items-center gap-1">
            <RouterLink
              v-for="item in navigationItems"
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
                class="ml-2 inline-flex min-w-5 items-center justify-center rounded-full bg-stone-900 px-1.5 py-0.5 text-[11px] text-white"
              >
                {{ pendingBadge }}
              </span>
            </RouterLink>
          </nav>
        </div>

        <div class="flex items-center gap-4">
          <p class="text-sm text-stone-500">{{ workspaceStore.authUser?.email }}</p>
          <button
            type="button"
            class="text-sm font-medium text-stone-600 hover:text-stone-900"
            @click="handleLogout"
          >
            退出
          </button>
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

    <main class="mx-auto w-full max-w-7xl flex-1 px-4 py-8 lg:px-8">
      <RouterView v-slot="{ Component, route: currentRoute }">
        <Transition name="page-swap" mode="out-in">
          <component :is="Component" :key="currentRoute.fullPath" />
        </Transition>
      </RouterView>
    </main>

    <nav
      class="pb-safe fixed inset-x-0 bottom-0 z-30 border-t border-stone-200 bg-white/80 p-2 backdrop-blur lg:hidden"
    >
      <div class="mx-auto grid max-w-md grid-cols-4 gap-1">
        <RouterLink
          v-for="item in navigationItems"
          :key="item.name"
          :to="item.path"
          class="flex flex-col items-center justify-center rounded-lg px-2 py-2 transition-colors"
          :class="
            route.name === item.name
              ? 'text-stone-900'
              : 'text-stone-500 hover:bg-stone-50 hover:text-stone-900'
          "
        >
          <span class="mt-1 block text-xs font-medium">
            {{ item.label }}
            <span v-if="item.name === 'prints' && pendingBadge">· {{ pendingBadge }}</span>
          </span>
        </RouterLink>
      </div>
    </nav>
  </div>
</template>

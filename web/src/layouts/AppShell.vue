<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from "vue-router";

import { navigationItems } from "@/router";

const route = useRoute();
</script>

<template>
  <div class="flex min-h-screen flex-col bg-white text-stone-900">
    <header class="sticky top-0 z-40 border-b border-stone-200 bg-white px-4 py-3 lg:px-8">
      <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 lg:hidden">
        <div class="flex items-center gap-3">
          <img src="/icon.jpg" alt="Ink Icon" class="h-8 w-8 rounded-lg object-contain" />
          <div>
            <p class="text-sm font-semibold text-stone-950">Ink</p>
          </div>
        </div>

        <RouterLink to="/login" class="text-sm font-medium text-stone-600 hover:text-stone-900">
          账号
        </RouterLink>
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
              {{ item.label }}
            </RouterLink>
          </nav>
        </div>

        <div class="flex items-center gap-4">
          <RouterLink to="/login" class="text-sm font-medium text-stone-600 hover:text-stone-900">
            登录
          </RouterLink>
        </div>
      </div>
    </header>

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
          <span class="mt-1 block text-xs font-medium">{{ item.label }}</span>
        </RouterLink>
      </div>
    </nav>
  </div>
</template>

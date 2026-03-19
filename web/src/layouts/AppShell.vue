<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from "vue-router";

import { navigationItems } from "@/router";

const route = useRoute();
</script>

<template>
  <div
    class="min-h-screen bg-[radial-gradient(circle_at_top,rgba(255,225,198,0.72),transparent_34%),linear-gradient(180deg,#fcfaf7_0%,#f5f1eb_100%)] text-stone-900"
  >
    <div class="mx-auto min-h-screen max-w-[1480px] px-4 py-4 lg:px-6 lg:py-6">
      <div
        class="rounded-[2rem] border border-stone-900/10 bg-white/55 shadow-[0_30px_80px_-48px_rgba(74,53,34,0.25)] backdrop-blur-2xl"
      >
        <header class="border-b border-stone-900/10 px-4 py-4 lg:px-6">
          <div class="flex items-center justify-between gap-4 lg:hidden">
            <div class="flex items-center gap-3">
              <span
                class="bg-white/78 inline-flex h-10 w-10 items-center justify-center rounded-2xl border border-stone-900/10 text-sm font-semibold text-stone-900"
              >
                Ink
              </span>
              <div>
                <p class="text-sm font-semibold text-stone-950">Ink 咕咕机助手</p>
              </div>
            </div>

            <RouterLink to="/login" class="ui-btn-secondary px-4 py-2"> 账号 </RouterLink>
          </div>

          <div class="hidden items-center gap-4 lg:grid lg:grid-cols-[220px_minmax(0,1fr)_220px]">
            <div class="flex items-center gap-3">
              <span
                class="bg-white/82 inline-flex h-11 w-11 items-center justify-center rounded-2xl border border-stone-900/10 text-sm font-semibold text-stone-900"
              >
                Ink
              </span>
              <p class="text-base font-semibold text-stone-950">Ink 咕咕机助手</p>
            </div>

            <nav class="overflow-x-auto">
              <div class="mx-auto max-w-[560px]">
                <div
                  class="bg-white/88 grid grid-cols-4 gap-2 rounded-[1.75rem] border border-stone-900/10 p-2 backdrop-blur"
                >
                  <RouterLink
                    v-for="item in navigationItems"
                    :key="item.name"
                    :to="item.path"
                    class="rounded-[1.2rem] px-2 py-3 text-center transition duration-200"
                    :class="
                      route.name === item.name
                        ? 'border border-[#e4cfb8] bg-[linear-gradient(180deg,rgba(255,248,241,0.96),rgba(245,233,220,0.96))] text-stone-950'
                        : 'bg-transparent text-stone-600 hover:bg-stone-100'
                    "
                  >
                    <span class="block text-sm font-semibold">{{ item.label }}</span>
                  </RouterLink>
                </div>
              </div>
            </nav>

            <div class="flex items-center justify-end gap-2">
              <RouterLink to="/login" class="ui-btn-secondary px-4 py-2"> 登录 </RouterLink>
              <button class="ui-btn-secondary px-4 py-2">退出</button>
            </div>
          </div>
        </header>

        <main class="px-5 py-5 lg:px-8 lg:py-6">
          <RouterView v-slot="{ Component, route: currentRoute }">
            <Transition name="page-swap" mode="out-in">
              <component :is="Component" :key="currentRoute.fullPath" />
            </Transition>
          </RouterView>
        </main>
      </div>

      <nav
        class="fixed inset-x-3 bottom-3 z-30 rounded-[1.75rem] border border-stone-900/10 bg-white/90 p-2 shadow-[0_24px_50px_-30px_rgba(35,24,19,0.6)] backdrop-blur lg:hidden"
      >
        <div class="grid grid-cols-4 gap-2">
          <RouterLink
            v-for="item in navigationItems"
            :key="item.name"
            :to="item.path"
            class="rounded-[1.2rem] px-2 py-3 text-center transition"
            :class="
              route.name === item.name
                ? 'border border-[#e4cfb8] bg-[linear-gradient(180deg,rgba(255,248,241,0.96),rgba(245,233,220,0.96))] text-stone-950'
                : 'bg-transparent text-stone-600 hover:bg-stone-100'
            "
          >
            <span class="block text-sm font-semibold">{{ item.label }}</span>
          </RouterLink>
        </div>
      </nav>
    </div>
  </div>
</template>

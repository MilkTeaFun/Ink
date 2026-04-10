<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watchEffect } from "vue";

import { useWorkspaceStore } from "@/stores/workspace";
import { resolveThemeMode } from "@/utils/workspace";

const workspaceStore = useWorkspaceStore();
const themeColors = {
  light: "#fafaf9",
  dark: "#14110f",
} as const;
const prefersDark = ref(false);
let colorSchemeMediaQuery: MediaQueryList | null = null;

function syncSystemColorScheme() {
  prefersDark.value = colorSchemeMediaQuery?.matches ?? false;
}

function handleSystemColorSchemeChange(event: MediaQueryListEvent) {
  prefersDark.value = event.matches;
}

onMounted(() => {
  if (workspaceStore.authSession && !workspaceStore.authUser && !workspaceStore.authBootstrapping) {
    void workspaceStore.initializeAuth();
  }

  if (typeof window === "undefined" || typeof window.matchMedia !== "function") {
    return;
  }

  colorSchemeMediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
  syncSystemColorScheme();

  if (typeof colorSchemeMediaQuery.addEventListener === "function") {
    colorSchemeMediaQuery.addEventListener("change", handleSystemColorSchemeChange);
    return;
  }

  colorSchemeMediaQuery.addListener(handleSystemColorSchemeChange);
});

onBeforeUnmount(() => {
  if (!colorSchemeMediaQuery) {
    return;
  }

  if (typeof colorSchemeMediaQuery.removeEventListener === "function") {
    colorSchemeMediaQuery.removeEventListener("change", handleSystemColorSchemeChange);
  } else {
    colorSchemeMediaQuery.removeListener(handleSystemColorSchemeChange);
  }
});

watchEffect(() => {
  if (typeof document === "undefined") {
    return;
  }

  const resolvedTheme = resolveThemeMode(workspaceStore.selectedTheme, prefersDark.value);
  const root = document.documentElement;

  root.dataset.theme = workspaceStore.selectedTheme;
  root.dataset.colorMode = resolvedTheme;
  root.style.colorScheme = resolvedTheme;
  document
    .querySelector('meta[name="theme-color"]')
    ?.setAttribute("content", themeColors[resolvedTheme] ?? themeColors.light);
});
</script>

<template>
  <RouterView />
</template>

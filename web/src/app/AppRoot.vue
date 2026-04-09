<script setup lang="ts">
import { onMounted, watchEffect } from "vue";

import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();
const themeColors = {
  light: "#fafaf9",
  soft: "#f7f1ea",
  system: "#fafaf9",
} as const;

onMounted(() => {
  if (workspaceStore.authSession && !workspaceStore.authUser && !workspaceStore.authBootstrapping) {
    void workspaceStore.initializeAuth();
  }
});

watchEffect(() => {
  if (typeof document === "undefined") {
    return;
  }

  document.documentElement.dataset.theme = workspaceStore.selectedTheme;
  document
    .querySelector('meta[name="theme-color"]')
    ?.setAttribute("content", themeColors[workspaceStore.selectedTheme] ?? themeColors.light);
});
</script>

<template>
  <RouterView />
</template>

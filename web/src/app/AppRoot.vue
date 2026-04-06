<script setup lang="ts">
import { onMounted, watchEffect } from "vue";

import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();

onMounted(() => {
  void workspaceStore.initializeAuth();
});

watchEffect(() => {
  if (typeof document === "undefined") {
    return;
  }

  document.documentElement.dataset.theme = workspaceStore.selectedTheme;
});
</script>

<template>
  <RouterView />
</template>

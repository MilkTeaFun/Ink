<script setup lang="ts">
import { watchEffect } from "vue";

import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();

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

import { defineStore } from "pinia";
import { computed, ref } from "vue";

export const useWorkspaceStore = defineStore("workspace", () => {
  const activeDeviceLabel = ref("书桌咕咕机");
  const activeModelLabel = ref("清楚温柔");
  const todayPrintCount = ref(3);
  const welcomeLabel = ref("简单一点，也可以很舒服");

  const isConfigured = computed(() => activeDeviceLabel.value !== "");

  return {
    activeDeviceLabel,
    activeModelLabel,
    todayPrintCount,
    welcomeLabel,
    isConfigured,
  };
});

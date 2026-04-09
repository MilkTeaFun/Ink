import { createApp } from "vue";

import "@fontsource/noto-sans-sc/chinese-simplified-400.css";
import "@fontsource/noto-sans-sc/chinese-simplified-500.css";
import "@fontsource/noto-sans-sc/chinese-simplified-600.css";
import AppRoot from "@/app/AppRoot.vue";
import { registerServiceWorker } from "@/pwa";
import router from "@/router";
import { pinia } from "@/stores/pinia";

import "@/styles.css";

const app = createApp(AppRoot);

app.use(pinia);
app.use(router);
app.mount("#app");

registerServiceWorker();

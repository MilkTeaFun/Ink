import { createPinia } from "pinia";
import { createApp } from "vue";

import "@fontsource/noto-sans-sc/chinese-simplified-400.css";
import "@fontsource/noto-sans-sc/chinese-simplified-500.css";
import "@fontsource/noto-sans-sc/chinese-simplified-600.css";
import AppRoot from "@/app/AppRoot.vue";
import router from "@/router";

import "@/styles.css";

const app = createApp(AppRoot);

app.use(createPinia());
app.use(router);
app.mount("#app");

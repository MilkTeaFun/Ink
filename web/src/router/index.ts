import {
  createMemoryHistory,
  createRouter,
  createWebHistory,
  type RouteRecordRaw,
} from "vue-router";
import type { RouterHistory } from "vue-router";

import AppShell from "@/layouts/AppShell.vue";
import ConversationsView from "@/views/ConversationsView.vue";
import LoginView from "@/views/LoginView.vue";
import PrintsView from "@/views/PrintsView.vue";
import SettingsView from "@/views/SettingsView.vue";
import StatusView from "@/views/StatusView.vue";

declare module "vue-router" {
  interface RouteMeta {
    label?: string;
    title?: string;
    description?: string;
    navHint?: string;
  }
}

const shellChildren: RouteRecordRaw[] = [
  {
    path: "status",
    name: "status",
    component: StatusView,
    meta: {
      label: "状态",
      title: "状态",
      description: "查看设备绑定情况、定时任务和最近的打印记录。",
      navHint: "设备与任务",
    },
  },
  {
    path: "conversations",
    name: "conversations",
    component: ConversationsView,
    meta: {
      label: "对话",
      title: "内容对话",
      description: "像聊天一样整理内容，确认满意后，再把它发去打印。",
      navHint: "整理内容",
    },
  },
  {
    path: "prints",
    name: "prints",
    component: PrintsView,
    meta: {
      label: "打印",
      title: "打印",
      description: "管理待确认内容、定时任务和打印记录，把打印流程集中在一起。",
      navHint: "打印流程",
    },
  },
  {
    path: "settings",
    name: "settings",
    component: SettingsView,
    meta: {
      label: "设置",
      title: "偏好设置",
      description: "调整默认设备、助手风格和打印习惯，让每次使用都更顺手。",
      navHint: "习惯与偏好",
    },
  },
];

export const navigationItems = shellChildren.map((route) => ({
  name: route.name as string,
  path: `/${route.path}`,
  label: route.meta?.label as string,
  navHint: route.meta?.navHint as string,
}));

export const routes: RouteRecordRaw[] = [
  {
    path: "/",
    component: AppShell,
    redirect: "/status",
    children: shellChildren,
  },
  {
    path: "/login",
    name: "login",
    component: LoginView,
    meta: {
      title: "欢迎使用 Ink",
      description: "登录后就可以继续管理设备、整理对话内容，并把纸条发到你想要的咕咕机。",
    },
  },
  {
    path: "/connections",
    redirect: "/prints",
  },
];

export function createAppRouter(
  history: RouterHistory = createWebHistory(import.meta.env.BASE_URL),
) {
  return createRouter({
    history,
    routes,
  });
}

export function createTestRouter() {
  return createAppRouter(createMemoryHistory());
}

const router = createAppRouter();

export default router;

import { mount } from "@vue/test-utils";

import ConnectionsView from "@/views/ConnectionsView.vue";
import ConversationsView from "@/views/ConversationsView.vue";
import LoginView from "@/views/LoginView.vue";
import SettingsView from "@/views/SettingsView.vue";
import StatusView from "@/views/StatusView.vue";

describe("workspace views", () => {
  it("renders the status overview surface", () => {
    const wrapper = mount(StatusView);

    expect(wrapper.text()).toContain("设备、任务和打印记录都在这里。");
    expect(wrapper.text()).toContain("书桌咕咕机");
    expect(wrapper.text()).toContain("定时任务");
  });

  it("renders the mobile-first conversation surface", () => {
    const wrapper = mount(ConversationsView);

    expect(wrapper.text()).toContain("通过对话整理内容，再决定打印哪一段。");
    expect(wrapper.text()).toContain("今日待办");
    expect(wrapper.text()).toContain("打印整段对话");
  });

  it("renders the login handoff preview", () => {
    const wrapper = mount(LoginView, {
      global: {
        stubs: {
          RouterLink: {
            template: "<a><slot /></a>",
          },
        },
      },
    });

    expect(wrapper.text()).toContain("继续管理你的设备和打印内容。");
    expect(wrapper.text()).toContain("先看看首页");
  });

  it("renders the future connections surface", () => {
    const wrapper = mount(ConnectionsView);

    expect(wrapper.text()).toContain("未来扩展会从这里开始。");
    expect(wrapper.text()).toContain("RSS 订阅");
    expect(wrapper.text()).toContain("新增连接");
  });

  it("renders the settings management surface", () => {
    const wrapper = mount(SettingsView);

    expect(wrapper.text()).toContain("账号、打印、主题和授权都从这里管理。");
    expect(wrapper.text()).toContain("页面主题");
    expect(wrapper.text()).toContain("AI 服务绑定");
  });
});

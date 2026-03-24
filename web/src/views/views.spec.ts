import { mount } from "@vue/test-utils";

import ConnectionsView from "@/views/ConnectionsView.vue";
import ConversationsView from "@/views/ConversationsView.vue";
import LoginView from "@/views/LoginView.vue";
import SettingsView from "@/views/SettingsView.vue";
import StatusView from "@/views/StatusView.vue";

describe("workspace views", () => {
  it("renders the status overview surface", () => {
    const wrapper = mount(StatusView);

    expect(wrapper.findAll(".grid.grid-cols-2 article")).toHaveLength(4);
    expect(wrapper.text()).toContain("已绑定设备");
    expect(wrapper.findAll(".ui-list-card article").at(0)?.text()).toContain("书桌咕咕机");
    expect(wrapper.findAll('button[aria-pressed="true"]')).toHaveLength(2);
    expect(wrapper.findAll('button[aria-pressed="false"]')).toHaveLength(1);
  });

  it("renders the mobile-first conversation surface", () => {
    const wrapper = mount(ConversationsView);

    expect(wrapper.findAll("aside article")).toHaveLength(3);
    expect(wrapper.findAll(".max-w-\\[85\\%\\]")).toHaveLength(2);
    expect(wrapper.find("textarea[placeholder='发送消息...']").exists()).toBe(true);
    expect(wrapper.text()).toContain("打印整段对话");
  });

  it("renders the login handoff preview", () => {
    const wrapper = mount(LoginView, {
      global: {
        stubs: {
          RouterLink: {
            props: ["to"],
            template: "<a :href='to'><slot /></a>",
          },
        },
      },
    });

    expect(wrapper.find("input[type='email']").exists()).toBe(true);
    expect(wrapper.find("input[type='password']").exists()).toBe(true);
    expect(wrapper.find("a[href='/']").text()).toBe("先看看首页");
  });

  it("renders the future connections surface", () => {
    const wrapper = mount(ConnectionsView);

    expect(wrapper.findAll(".ui-list-card").at(0)?.findAll("article")).toHaveLength(2);
    expect(wrapper.findAll(".ui-list-card").at(1)?.findAll("article")).toHaveLength(3);
    expect(wrapper.findAll(".ui-list-card").at(2)?.findAll(".ui-list-row")).toHaveLength(3);
    expect(wrapper.text()).toContain("新增连接");
  });

  it("renders the settings management surface", () => {
    const wrapper = mount(SettingsView);

    expect(wrapper.findAll("article")).toHaveLength(5);
    expect(wrapper.text()).toContain("账号管理");
    expect(wrapper.text()).toContain("页面主题");
    expect(wrapper.text()).toContain("AI 服务");
  });
});

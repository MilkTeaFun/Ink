import { createPinia, setActivePinia } from "pinia";

import { useWorkspaceStore } from "@/stores/workspace";

describe("workspace store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("exposes stable defaults for the workspace shell", () => {
    const store = useWorkspaceStore();

    expect(store.activeDeviceLabel).toBe("书桌咕咕机");
    expect(store.activeModelLabel).toBe("清楚温柔");
    expect(store.todayPrintCount).toBe(3);
    expect(store.welcomeLabel).toBe("简单一点，也可以很舒服");
    expect(store.isConfigured).toBe(true);
  });

  it("derives configuration state from the active device label", () => {
    const store = useWorkspaceStore();

    store.activeDeviceLabel = "";
    expect(store.isConfigured).toBe(false);

    store.activeDeviceLabel = "卧室咕咕机";
    expect(store.isConfigured).toBe(true);
  });
});

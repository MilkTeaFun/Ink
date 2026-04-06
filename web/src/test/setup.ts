import { config } from "@vue/test-utils";
import { beforeEach } from "vitest";

config.global.stubs = {
  transition: false,
};

beforeEach(() => {
  window.localStorage.clear();
  document.title = "Ink";
  delete document.documentElement.dataset.theme;
});

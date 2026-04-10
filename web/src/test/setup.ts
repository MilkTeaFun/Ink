import { config } from "@vue/test-utils";
import { beforeEach, vi } from "vitest";

type MatchMediaEventHandler = (
  type: string,
  listener: EventListenerOrEventListenerObject | null,
) => void;
type LegacyMediaQueryListenerHandler = (
  listener: ((event: MediaQueryListEvent) => void) | null,
) => void;
type MatchMediaDispatchEventHandler = (event: Event) => boolean;
type MatchMediaFactory = (query: string) => MediaQueryList;

config.global.stubs = {
  transition: false,
};

function createMemoryStorage(): Storage {
  const values = new Map<string, string>();

  return {
    get length() {
      return values.size;
    },
    clear() {
      values.clear();
    },
    getItem(key) {
      return values.get(key) ?? null;
    },
    key(index) {
      return [...values.keys()][index] ?? null;
    },
    removeItem(key) {
      values.delete(key);
    },
    setItem(key, value) {
      values.set(key, value);
    },
  };
}

function hasStorageApi(storage: Storage | undefined) {
  return (
    typeof storage?.getItem === "function" &&
    typeof storage.setItem === "function" &&
    typeof storage.removeItem === "function" &&
    typeof storage.clear === "function"
  );
}

if (!hasStorageApi(window.localStorage)) {
  Object.defineProperty(window, "localStorage", {
    configurable: true,
    value: createMemoryStorage(),
  });
}

if (!hasStorageApi(window.sessionStorage)) {
  Object.defineProperty(window, "sessionStorage", {
    configurable: true,
    value: createMemoryStorage(),
  });
}

beforeEach(() => {
  window.localStorage.clear();
  window.sessionStorage.clear();
  document.title = "Ink";
  delete document.documentElement.dataset.theme;
  delete document.documentElement.dataset.colorMode;
  document.documentElement.style.colorScheme = "";
});

if (typeof window.matchMedia !== "function") {
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: vi.fn<MatchMediaFactory>((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: vi.fn<MatchMediaEventHandler>(() => undefined),
      removeEventListener: vi.fn<MatchMediaEventHandler>(() => undefined),
      addListener: vi.fn<LegacyMediaQueryListenerHandler>(() => undefined),
      removeListener: vi.fn<LegacyMediaQueryListenerHandler>(() => undefined),
      dispatchEvent: vi.fn<MatchMediaDispatchEventHandler>(() => true),
    })),
  });
}

import { createI18n } from "vue-i18n";

import type { LocaleCode, LocalePreference } from "@/types/workspace";

import enUS from "./messages/en-US";
import zhCN from "./messages/zh-CN";

export const SUPPORTED_LOCALES = ["zh-CN", "en-US"] as const;
export const DEFAULT_LOCALE_PREFERENCE: LocalePreference = "system";

function readNavigatorLanguage() {
  if (typeof navigator === "undefined") {
    return "zh-CN";
  }

  return navigator.language;
}

export function resolveBrowserLocale(language = readNavigatorLanguage()): LocaleCode {
  return language.toLowerCase().startsWith("zh") ? "zh-CN" : "en-US";
}

export function isLocaleCode(value: unknown): value is LocaleCode {
  return SUPPORTED_LOCALES.includes(value as LocaleCode);
}

export function isLocalePreference(value: unknown): value is LocalePreference {
  return value === DEFAULT_LOCALE_PREFERENCE || isLocaleCode(value);
}

export function normalizeLocalePreference(value: unknown): LocalePreference {
  return isLocalePreference(value) ? value : DEFAULT_LOCALE_PREFERENCE;
}

export function resolveLocalePreference(
  preference: LocalePreference = DEFAULT_LOCALE_PREFERENCE,
  browserLocale = readNavigatorLanguage(),
): LocaleCode {
  if (preference === DEFAULT_LOCALE_PREFERENCE) {
    return resolveBrowserLocale(browserLocale);
  }

  return preference as LocaleCode;
}

export const i18n = createI18n({
  legacy: false,
  locale: resolveLocalePreference(DEFAULT_LOCALE_PREFERENCE) as LocaleCode,
  fallbackLocale: "zh-CN",
  messages: {
    "zh-CN": zhCN,
    "en-US": enUS,
  },
});

export function setI18nLocale(locale: LocaleCode) {
  i18n.global.locale.value = locale;
}

export function getI18nLocale() {
  return i18n.global.locale.value as LocaleCode;
}

export function translate(key: string, values?: Record<string, unknown>) {
  return values ? i18n.global.t(key, values) : i18n.global.t(key);
}

export default i18n;

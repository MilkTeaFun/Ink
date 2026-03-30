<script setup lang="ts">
import { ref } from "vue";

const isLoginProtectionEnabled = ref(true);
const isSendConfirmationEnabled = ref(true);
const themes = [
  { label: "柔光", value: "soft" },
  { label: "浅色", value: "light" },
  { label: "系统跟随", value: "system" },
] as const;
const selectedTheme = ref<(typeof themes)[number]["value"]>("light");
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-8 pt-4 pb-24 lg:pb-12">
    <div class="max-w-2xl px-4 sm:px-0">
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">设置</h2>
      <p class="mt-1 text-sm text-stone-500">管理账号、打印、主题和授权选项。</p>
    </div>

    <div class="space-y-12">
      <!-- 账号管理 -->
      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">账号管理</h3>
          <p class="mt-1 text-sm text-stone-500">登录、退出和基本账号信息。</p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">当前账号</p>
                <p class="mt-0.5 text-sm text-stone-500">name@example.com</p>
              </div>
              <button type="button" class="ui-btn-secondary px-3 py-1.5 text-sm">管理</button>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">登录保护</p>
                <p class="mt-0.5 text-sm text-stone-500">
                  {{
                    isLoginProtectionEnabled
                      ? "长时间未操作后会自动退出"
                      : "保持登录，直到你主动退出"
                  }}
                </p>
              </div>
              <button
                type="button"
                class="ui-toggle"
                :class="{ 'is-on': isLoginProtectionEnabled }"
                :aria-label="`${isLoginProtectionEnabled ? '关闭' : '开启'}登录保护`"
                :aria-pressed="isLoginProtectionEnabled"
                @click="isLoginProtectionEnabled = !isLoginProtectionEnabled"
              >
                <span class="ui-toggle-thumb" />
              </button>
            </div>
          </div>
        </div>
      </article>

      <!-- 打印设置 -->
      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">打印设置</h3>
          <p class="mt-1 text-sm text-stone-500">默认设备和打印方式。</p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">默认设备</p>
                <p class="mt-0.5 text-sm text-stone-500">书桌咕咕机</p>
              </div>
              <span
                class="inline-flex items-center rounded-full bg-stone-100 px-2.5 py-0.5 text-xs font-medium text-stone-800"
                >常用</span
              >
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">发送前确认</p>
                <p class="mt-0.5 text-sm text-stone-500">
                  {{ isSendConfirmationEnabled ? "发送前会再确认一次" : "会直接发送到默认设备" }}
                </p>
              </div>
              <button
                type="button"
                class="ui-toggle"
                :class="{ 'is-on': isSendConfirmationEnabled }"
                :aria-label="`${isSendConfirmationEnabled ? '关闭' : '开启'}发送前确认`"
                :aria-pressed="isSendConfirmationEnabled"
                @click="isSendConfirmationEnabled = !isSendConfirmationEnabled"
              >
                <span class="ui-toggle-thumb" />
              </button>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">纸条风格</p>
                <p class="mt-0.5 text-sm text-stone-500">简洁清晰</p>
              </div>
              <button type="button" class="ui-btn-secondary px-3 py-1.5 text-sm">设置</button>
            </div>
          </div>
        </div>
      </article>

      <!-- 页面主题 -->
      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">页面主题</h3>
          <p class="mt-1 text-sm text-stone-500">切换应用程序的视觉风格。</p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-choice-grid">
            <button
              v-for="theme in themes"
              :key="theme.value"
              type="button"
              class="ui-btn-secondary justify-center py-2 text-sm"
              :class="
                theme.value === selectedTheme
                  ? 'border-stone-300 bg-white text-stone-900 ring-1 ring-stone-200/70'
                  : 'border-transparent bg-transparent text-stone-600 shadow-none hover:border-stone-200 hover:bg-white'
              "
              :aria-pressed="theme.value === selectedTheme"
              @click="selectedTheme = theme.value"
            >
              {{ theme.label }}
            </button>
          </div>
        </div>
      </article>

      <!-- AI 服务绑定 -->
      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">AI 服务</h3>
          <p class="mt-1 text-sm text-stone-500">管理模型与服务商配置。</p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">默认服务商</p>
                <p class="mt-0.5 text-sm text-stone-500">暂未绑定</p>
              </div>
              <button type="button" class="ui-btn-secondary px-3 py-1.5 text-sm">去绑定</button>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">回答风格</p>
                <p class="mt-0.5 text-sm text-stone-500">清楚温柔</p>
              </div>
              <button type="button" class="ui-btn-secondary px-3 py-1.5 text-sm">调整</button>
            </div>
          </div>
        </div>
      </article>

      <!-- 授权 -->
      <article
        class="grid grid-cols-1 items-start gap-x-10 gap-y-5 px-4 sm:px-0 md:grid-cols-[minmax(0,13rem)_1fr]"
      >
        <div>
          <h3 class="text-base leading-6 font-semibold text-stone-900">授权与隐私</h3>
          <p class="mt-1 text-sm text-stone-500">外部来源和服务权限管理。</p>
        </div>
        <div class="min-w-0">
          <div class="ui-settings-group">
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">连接服务授权</p>
                <p class="mt-0.5 text-sm text-stone-500">RSS、天气、日历或自定义接口</p>
              </div>
              <button type="button" class="ui-btn-secondary px-3 py-1.5 text-sm">管理</button>
            </div>
            <div class="ui-settings-row">
              <div class="ui-settings-copy">
                <p class="text-sm font-medium text-stone-900">打印设备授权</p>
                <p class="mt-0.5 text-sm text-stone-500">默认设备、绑定状态和访问范围</p>
              </div>
              <button type="button" class="ui-btn-secondary px-3 py-1.5 text-sm">管理</button>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

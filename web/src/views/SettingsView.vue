<template>
  <section class="mx-auto max-w-5xl space-y-10 pb-24 pt-4 lg:pb-12">
    <div class="px-4 sm:px-0">
      <h2 class="text-2xl font-semibold tracking-tight text-stone-900">设置</h2>
      <p class="mt-1 text-sm text-stone-500">管理账号、打印、主题和授权选项。</p>
    </div>

    <div class="space-y-10 divide-y divide-stone-200">
      <!-- 账号管理 -->
      <article class="grid grid-cols-1 gap-x-8 gap-y-6 px-4 pt-10 sm:px-0 md:grid-cols-3">
        <div>
          <h3 class="text-base font-semibold leading-6 text-stone-900">账号管理</h3>
          <p class="mt-1 text-sm text-stone-500">登录、退出和基本账号信息。</p>
        </div>
        <div class="md:col-span-2">
          <div class="ui-list-card">
            <div
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">当前账号</p>
                <p class="mt-0.5 text-sm text-stone-500">name@example.com</p>
              </div>
              <button class="ui-btn-secondary px-3 py-1.5 text-sm">管理</button>
            </div>
            <div
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">登录保护</p>
                <p class="mt-0.5 text-sm text-stone-500">长时间未操作自动退出</p>
              </div>
              <button class="ui-btn-secondary px-3 py-1.5 text-sm">设置</button>
            </div>
          </div>
        </div>
      </article>

      <!-- 打印设置 -->
      <article class="grid grid-cols-1 gap-x-8 gap-y-6 px-4 pt-10 sm:px-0 md:grid-cols-3">
        <div>
          <h3 class="text-base font-semibold leading-6 text-stone-900">打印设置</h3>
          <p class="mt-1 text-sm text-stone-500">默认设备和打印方式。</p>
        </div>
        <div class="md:col-span-2">
          <div class="ui-list-card">
            <div
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">默认设备</p>
                <p class="mt-0.5 text-sm text-stone-500">书桌咕咕机</p>
              </div>
              <span
                class="inline-flex items-center rounded-full bg-stone-100 px-2.5 py-0.5 text-xs font-medium text-stone-800"
                >常用</span
              >
            </div>
            <div
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">发送前确认</p>
                <p class="mt-0.5 text-sm text-stone-500">已开启</p>
              </div>
              <span
                class="inline-flex items-center rounded-full bg-emerald-50 px-2.5 py-0.5 text-xs font-medium text-emerald-700 ring-1 ring-inset ring-emerald-600/20"
                >开启</span
              >
            </div>
            <div
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">纸条风格</p>
                <p class="mt-0.5 text-sm text-stone-500">简洁清晰</p>
              </div>
              <button class="ui-btn-secondary px-3 py-1.5 text-sm">设置</button>
            </div>
          </div>
        </div>
      </article>

      <!-- 页面主题 -->
      <article class="grid grid-cols-1 gap-x-8 gap-y-6 px-4 pt-10 sm:px-0 md:grid-cols-3">
        <div>
          <h3 class="text-base font-semibold leading-6 text-stone-900">页面主题</h3>
          <p class="mt-1 text-sm text-stone-500">切换应用程序的视觉风格。</p>
        </div>
        <div class="md:col-span-2">
          <div class="grid gap-3 sm:grid-cols-3">
            <button class="ui-btn-secondary justify-center py-2 text-sm">柔光</button>
            <button
              class="ui-btn-secondary justify-center py-2 text-sm shadow-sm ring-1 ring-stone-900"
            >
              浅色
            </button>
            <button class="ui-btn-secondary justify-center py-2 text-sm">系统跟随</button>
          </div>
        </div>
      </article>

      <!-- AI 服务绑定 -->
      <article class="grid grid-cols-1 gap-x-8 gap-y-6 px-4 pt-10 sm:px-0 md:grid-cols-3">
        <div>
          <h3 class="text-base font-semibold leading-6 text-stone-900">AI 服务</h3>
          <p class="mt-1 text-sm text-stone-500">管理模型与服务商配置。</p>
        </div>
        <div class="md:col-span-2">
          <div class="ui-list-card">
            <div
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">默认服务商</p>
                <p class="mt-0.5 text-sm text-stone-500">暂未绑定</p>
              </div>
              <button class="ui-btn-primary px-3 py-1.5 text-sm">去绑定</button>
            </div>
            <div
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">回答风格</p>
                <p class="mt-0.5 text-sm text-stone-500">清楚温柔</p>
              </div>
              <button class="ui-btn-secondary px-3 py-1.5 text-sm">调整</button>
            </div>
          </div>
        </div>
      </article>

      <!-- 授权 -->
      <article class="grid grid-cols-1 gap-x-8 gap-y-6 px-4 pt-10 sm:px-0 md:grid-cols-3">
        <div>
          <h3 class="text-base font-semibold leading-6 text-stone-900">授权与隐私</h3>
          <p class="mt-1 text-sm text-stone-500">外部来源和服务权限管理。</p>
        </div>
        <div class="md:col-span-2">
          <div class="ui-list-card">
            <div
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">连接服务授权</p>
                <p class="mt-0.5 text-sm text-stone-500">RSS、天气、日历或自定义接口</p>
              </div>
              <button class="ui-btn-secondary px-3 py-1.5 text-sm">管理</button>
            </div>
            <div
              class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p class="text-sm font-medium text-stone-900">打印设备授权</p>
                <p class="mt-0.5 text-sm text-stone-500">默认设备、绑定状态和访问范围</p>
              </div>
              <button class="ui-btn-secondary px-3 py-1.5 text-sm">管理</button>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

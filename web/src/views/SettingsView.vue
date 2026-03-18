<template>
  <section class="space-y-8 pb-24 lg:pb-0">
    <div>
      <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">设置</p>
      <h2 class="mt-2 text-3xl text-stone-950">账号、打印、主题和授权都从这里管理。</h2>
    </div>

    <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
      <section class="space-y-6">
        <article class="ui-panel">
          <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">账号管理</p>
          <h3 class="mt-2 text-2xl font-semibold text-stone-950">登录、退出和基本账号信息。</h3>
          <div class="ui-list-card mt-4">
            <div class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-sm font-semibold text-stone-950">当前账号</p>
                <p class="mt-1 text-sm text-stone-500">name@example.com</p>
              </div>
              <button class="ui-btn-secondary px-3 py-2">管理</button>
            </div>
            <div class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-sm font-semibold text-stone-950">登录保护</p>
                <p class="mt-1 text-sm text-stone-500">长时间未操作自动退出</p>
              </div>
              <button class="ui-btn-secondary px-3 py-2">设置</button>
            </div>
          </div>
        </article>

        <article class="ui-panel">
          <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">打印设置</p>
          <h3 class="mt-2 text-2xl font-semibold text-stone-950">默认设备和打印方式。</h3>
          <div class="ui-list-card mt-4">
            <div class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-sm font-semibold text-stone-950">默认设备</p>
                <p class="mt-1 text-sm text-stone-500">书桌咕咕机</p>
              </div>
              <span class="text-xs font-medium uppercase tracking-[0.2em] text-stone-400">常用</span>
            </div>
            <div class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-sm font-semibold text-stone-950">发送前确认</p>
                <p class="mt-1 text-sm text-stone-500">已开启</p>
              </div>
              <span class="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-800">开启</span>
            </div>
            <div class="ui-list-row">
              <p class="text-sm font-semibold text-stone-950">纸条风格</p>
              <p class="mt-1 text-sm text-stone-500">简洁清晰</p>
            </div>
          </div>
        </article>
      </section>

      <section class="space-y-6">
        <article class="ui-panel">
          <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">页面主题</p>
          <h3 class="mt-2 text-2xl font-semibold text-stone-950">未来会在这里切换主题风格。</h3>
          <div class="mt-4 grid gap-3 sm:grid-cols-3">
            <button class="ui-btn-secondary justify-center">柔光</button>
            <button class="ui-btn-secondary justify-center">浅色</button>
            <button class="ui-btn-secondary justify-center">系统跟随</button>
          </div>
        </article>

        <article class="ui-panel">
          <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">AI 服务绑定</p>
          <h3 class="mt-2 text-2xl font-semibold text-stone-950">以后会在这里管理模型与服务商。</h3>
          <div class="ui-list-card mt-4">
            <div class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-sm font-semibold text-stone-950">默认 AI 服务</p>
                <p class="mt-1 text-sm text-stone-500">暂未绑定</p>
              </div>
              <button class="ui-btn-primary">去绑定</button>
            </div>
            <div class="ui-list-row flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-sm font-semibold text-stone-950">回答风格</p>
                <p class="mt-1 text-sm text-stone-500">清楚温柔</p>
              </div>
              <button class="ui-btn-secondary px-3 py-2">调整</button>
            </div>
          </div>
        </article>

        <article class="ui-panel">
          <p class="text-[0.68rem] uppercase tracking-[0.24em] text-stone-500">授权</p>
          <h3 class="mt-2 text-2xl font-semibold text-stone-950">外部来源和服务权限都会放在这里。</h3>
          <div class="ui-list-card mt-4">
            <div class="ui-list-row">
              <p class="text-sm font-semibold text-stone-950">连接服务授权</p>
              <p class="mt-1 text-sm text-stone-500">RSS、天气、日历或自定义接口</p>
            </div>
            <div class="ui-list-row">
              <p class="text-sm font-semibold text-stone-950">打印设备授权</p>
              <p class="mt-1 text-sm text-stone-500">默认设备、绑定状态和访问范围</p>
            </div>
          </div>
        </article>
      </section>
    </div>
  </section>
</template>

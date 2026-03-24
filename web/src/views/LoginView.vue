<template>
  <div class="min-h-screen bg-white px-4 py-6 text-stone-900">
    <div
      class="mx-auto grid min-h-[calc(100vh-3rem)] max-w-6xl items-center gap-12 lg:grid-cols-[1.1fr_0.9fr]"
    >
      <section class="p-8 lg:p-10">
        <p class="text-[0.7rem] font-medium uppercase tracking-[0.28em] text-stone-500">欢迎回来</p>
        <h1
          class="mt-5 text-[clamp(2rem,5vw,3.5rem)] font-semibold leading-[1.1] tracking-tight text-stone-900"
        >
          打开 Ink，继续你的纸条灵感。
        </h1>
        <p class="mt-6 max-w-xl text-base leading-relaxed text-stone-500">
          管理设备、整理内容、选好要打印的小纸条。整个过程会尽量保持简单，不需要你记很多技术细节。
        </p>
      </section>

      <section class="rounded-2xl border border-stone-200 bg-white p-8 shadow-sm lg:p-10">
        <h2 class="text-xl font-semibold text-stone-900">登录账号</h2>
        <p class="mt-1 text-sm text-stone-500">继续管理你的设备和打印内容。</p>

        <div class="mt-8 space-y-5">
          <div>
            <label class="mb-2 block text-sm font-medium text-stone-900">邮箱</label>
            <input
              type="email"
              placeholder="name@example.com"
              class="w-full rounded-lg border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 transition-colors placeholder:text-stone-400 focus:border-stone-900 focus:outline-none focus:ring-1 focus:ring-stone-900"
            />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-stone-900">密码</label>
            <input
              type="password"
              placeholder="请输入密码"
              class="w-full rounded-lg border border-stone-200 bg-white px-4 py-2.5 text-sm text-stone-900 transition-colors placeholder:text-stone-400 focus:border-stone-900 focus:outline-none focus:ring-1 focus:ring-stone-900"
            />
          </div>
        </div>

        <div class="mt-8 flex flex-col gap-3 sm:flex-row">
          <button class="ui-btn-primary w-full px-6 py-2.5 sm:w-auto">登录</button>
          <RouterLink to="/" class="ui-btn-secondary w-full px-6 py-2.5 sm:w-auto"
            >先看看首页</RouterLink
          >
        </div>
      </section>
    </div>
  </div>
</template>

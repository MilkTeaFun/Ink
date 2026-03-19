<template>
  <div
    class="min-h-screen bg-[linear-gradient(180deg,#f7efe2_0%,#efe4d4_100%)] px-4 py-6 text-stone-900"
  >
    <div
      class="mx-auto grid min-h-[calc(100vh-3rem)] max-w-6xl items-center gap-6 lg:grid-cols-[1.1fr_0.9fr]"
    >
      <section
        class="bg-white/58 rounded-[2rem] border border-stone-900/10 p-8 shadow-panel backdrop-blur lg:p-10"
      >
        <p class="text-[0.7rem] uppercase tracking-[0.28em] text-stone-500">欢迎回来</p>
        <h1 class="mt-5 text-[clamp(3rem,7vw,5.8rem)] leading-[0.92] text-stone-950">
          打开 Ink，继续你的纸条灵感。
        </h1>
        <p class="mt-5 max-w-xl text-base leading-7 text-stone-600">
          管理设备、整理内容、选好要打印的小纸条。整个过程会尽量保持简单，不需要你记很多技术细节。
        </p>
      </section>

      <section
        class="bg-white/58 rounded-[2rem] border border-stone-900/10 p-6 shadow-panel backdrop-blur lg:p-8"
      >
        <p class="text-[0.68rem] uppercase tracking-[0.3em] text-stone-500">登录账号</p>
        <h2 class="mt-3 text-3xl text-stone-950">继续管理你的设备和打印内容。</h2>

        <div class="mt-8 rounded-[1.5rem] border border-stone-900/10 bg-stone-50/80 p-5">
          <label class="block text-sm font-medium text-stone-700">
            邮箱
            <input
              type="email"
              placeholder="name@example.com"
              class="mt-2 w-full rounded-2xl border border-white/50 bg-white/75 px-4 py-3 text-sm text-stone-500 outline-none backdrop-blur"
            />
          </label>
          <label class="mt-4 block text-sm font-medium text-stone-700">
            密码
            <input
              type="password"
              placeholder="请输入密码"
              class="mt-2 w-full rounded-2xl border border-white/50 bg-white/75 px-4 py-3 text-sm text-stone-500 outline-none backdrop-blur"
            />
          </label>
        </div>

        <div class="mt-6 flex flex-col gap-3 sm:flex-row">
          <button class="ui-btn-primary px-5 py-3">登录</button>
          <RouterLink to="/devices" class="ui-btn-secondary px-5 py-3"> 先看看首页 </RouterLink>
        </div>
      </section>
    </div>
  </div>
</template>

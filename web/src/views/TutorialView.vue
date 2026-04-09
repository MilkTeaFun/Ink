<script setup lang="ts">
import { RouterLink } from "vue-router";
</script>

<template>
  <section class="mx-auto max-w-5xl space-y-6 px-4 pt-4 pb-24 sm:space-y-8 sm:px-0 lg:pb-12">
    <div class="max-w-3xl space-y-4">
      <p class="text-sm font-medium tracking-[0.2em] text-stone-500 uppercase">教程</p>
      <div class="space-y-3">
        <h2 class="text-3xl font-semibold tracking-tight text-stone-900">
          绑定咕咕机并打印第一张纸条
        </h2>
        <p class="text-base leading-7 text-stone-600">
          Ink 登录后会使用你的真实账号数据。完成一次设备绑定后，你就可以把 AI
          整理出的内容直接发到自己的咕咕机。
        </p>
      </div>
    </div>

    <article class="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
      <p class="text-sm font-medium text-stone-500">作为手机应用使用</p>
      <h3 class="mt-2 text-lg font-semibold text-stone-900">把 Ink 添加到 iPhone 主屏幕</h3>
      <p class="mt-3 text-sm leading-7 text-stone-600">
        用 Safari 打开 Ink 后，点击底部分享按钮，选择“添加到主屏幕”。之后你可以像普通 App
        一样从桌面进入 Ink，顶部和底部导航都会按手机安全区适配。
      </p>
    </article>

    <div class="grid gap-6 lg:grid-cols-3">
      <article class="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
        <p class="text-sm font-medium text-stone-500">步骤 1</p>
        <h3 class="mt-3 text-lg font-semibold text-stone-900">准备设备编号</h3>
        <p class="mt-3 text-sm leading-7 text-stone-600">
          打开咕咕机 App 或设备说明，确认你的 Memobird 设备编号。登录 Ink
          后，去状态页点击“添加设备”，填写设备名称、备注和设备编号。
        </p>
      </article>

      <article class="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
        <p class="text-sm font-medium text-stone-500">步骤 2</p>
        <h3 class="mt-3 text-lg font-semibold text-stone-900">完成绑定并设为默认</h3>
        <p class="mt-3 text-sm leading-7 text-stone-600">
          绑定成功后，Ink
          会把这台设备记到你的账号下。你可以在状态页把常用设备设为默认，后续对话打印和手动打印都会优先发到这里。
        </p>
      </article>

      <article class="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
        <p class="text-sm font-medium text-stone-500">步骤 3</p>
        <h3 class="mt-3 text-lg font-semibold text-stone-900">打印第一张纸条</h3>
        <p class="mt-3 text-sm leading-7 text-stone-600">
          在对话页整理一段内容，点击“打印当前对话”或在打印页手动新建打印。如果你开启了发送前确认，任务会先进入待确认列表，再由你决定是否真正发往咕咕机。
        </p>
      </article>
    </div>

    <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
      <article class="rounded-2xl border border-stone-200 bg-stone-50 p-6">
        <h3 class="text-lg font-semibold text-stone-900">常见问题</h3>
        <div class="mt-5 space-y-5">
          <div>
            <p class="text-sm font-medium text-stone-900">绑定失败</p>
            <p class="mt-1 text-sm leading-7 text-stone-600">
              检查设备编号是否正确，以及服务器是否已经配置了 Memobird
              的访问密钥。若仍失败，稍后重试一次。
            </p>
          </div>
          <div>
            <p class="text-sm font-medium text-stone-900">任务已提交但没有出纸</p>
            <p class="mt-1 text-sm leading-7 text-stone-600">
              先确认设备在线、纸张充足，再回到打印页刷新任务状态。Ink 会继续显示最近一次提交的结果。
            </p>
          </div>
          <div>
            <p class="text-sm font-medium text-stone-900">还没配置 AI</p>
            <p class="mt-1 text-sm leading-7 text-stone-600">
              管理员需要先在设置页配置 OpenAI 兼容模型的供应商、API URL、模型和 API
              Key。密钥只会保存在服务端，不会直接暴露到浏览器。
            </p>
          </div>
        </div>
      </article>

      <aside class="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
        <h3 class="text-lg font-semibold text-stone-900">开始使用</h3>
        <div class="mt-5 space-y-3">
          <RouterLink
            to="/status"
            class="ui-btn-primary block w-full px-4 py-2.5 text-center text-sm"
          >
            去绑定设备
          </RouterLink>
          <RouterLink
            to="/prints"
            class="ui-btn-secondary block w-full px-4 py-2.5 text-center text-sm"
          >
            去打印页
          </RouterLink>
          <RouterLink
            to="/settings"
            class="ui-btn-secondary block w-full px-4 py-2.5 text-center text-sm"
          >
            去设置 AI
          </RouterLink>
        </div>
      </aside>
    </div>
  </section>
</template>

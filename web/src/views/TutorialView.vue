<script setup lang="ts">
import { RouterLink } from "vue-router";

const quickChecks = [
  {
    title: "原咕咕机 App 已可用",
    detail: "先确认这台设备已经在手机原来的咕咕机 App 里绑定成功。",
  },
  {
    title: "设备已经连上 Wi-Fi",
    detail: "状态纸条顶部应显示 WiFi Connected，说明设备当前在线。",
  },
  {
    title: "双击开机键吐出纸条",
    detail: "这张纸条会打印出 Device ID，Ink 绑定就靠这一串编号。",
  },
] as const;

const fieldGuides = [
  {
    label: "要填的内容",
    value: "Device ID",
    detail: "把冒号后的一长串字符完整填进“咕咕机设备编号”。",
    positive: true,
  },
  {
    label: "不要填错",
    value: "MAC Address",
    detail: "这是设备网卡地址，不是 Ink 用来绑定的编号。",
    positive: false,
  },
  {
    label: "也不用扫码",
    value: "底部二维码",
    detail: "Ink 走设备编号绑定，不需要扫描状态纸条上的二维码。",
    positive: false,
  },
] as const;

const steps = [
  {
    number: "01",
    title: "先拿到状态纸条",
    body: "确保设备已经在手机原有的咕咕机 App 里可用，并且已经连上 Wi-Fi。然后双击开机键，设备会打印一张当前状态纸条。",
    note: "你应该看到类似右侧示例里的 WiFi Connected、Device ID 和 MAC Address。",
  },
  {
    number: "02",
    title: "把 Device ID 填进添加设备",
    body: "登录 Ink 后进入状态页，点击“添加设备”。设备名称和备注按你的习惯填写，关键是把 Device ID 后面那一长串字符完整填进“咕咕机设备编号”。",
    note: "真正决定能否绑定成功的是 Device ID，不是 WiFi Name，也不是 MAC Address。",
  },
  {
    number: "03",
    title: "绑定后设为默认并试打一张",
    body: "绑定成功后，可以把常用设备设为默认。接着去对话页或打印页试发一张短纸条，确认这台咕咕机已经能正常接收 Ink 的任务。",
    note: "如果开启了发送前确认，任务会先进入待确认列表，再由你手动提交。",
  },
] as const;

const faqs = [
  {
    question: "纸条上哪一项最重要？",
    answer:
      "只看 Device ID 那一行。Ink 绑定时需要的是 Device ID 后面的长串字符，别把 MAC Address 或 WiFi Name 填进去。",
  },
  {
    question: "为什么双击开机键后没出状态纸条？",
    answer:
      "先检查设备是否开机、有纸、并且已经连接电源或电量足够。确认后再双击一次，通常会重新打印状态纸条。",
  },
  {
    question: "绑定失败怎么办？",
    answer:
      "先重新核对 Device ID 是否完整，再确认服务器已经配置了 Memobird 访问密钥。若仍失败，可以重新打印状态纸条后再试一次。",
  },
  {
    question: "还没配置 AI 会影响绑定吗？",
    answer:
      "不会影响绑定设备本身，但管理员仍需要在设置页配置 OpenAI 兼容服务，后续对话整理和自动生成内容才会真正启用。",
  },
] as const;
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-6 pt-4 sm:space-y-8 lg:space-y-10">
    <div
      class="relative overflow-hidden rounded-[2rem] border border-stone-200 bg-[linear-gradient(135deg,rgba(250,250,249,0.95),rgba(245,245,244,0.92))] px-5 py-6 shadow-sm sm:px-7 sm:py-8 lg:px-10"
    >
      <div
        aria-hidden="true"
        class="absolute top-0 right-0 h-40 w-40 rounded-full bg-amber-100/70 blur-3xl"
      />
      <div
        aria-hidden="true"
        class="absolute bottom-0 left-0 h-32 w-32 rounded-full bg-stone-200/70 blur-3xl"
      />

      <div class="relative grid gap-8 xl:grid-cols-[minmax(0,1.05fr)_420px] xl:items-center">
        <div class="space-y-6">
          <div class="space-y-4">
            <p class="text-sm font-medium tracking-[0.2em] text-stone-500 uppercase">绑定教程</p>
            <div class="space-y-3">
              <h2
                class="max-w-3xl text-[clamp(2rem,4.8vw,3.5rem)] font-semibold tracking-tight text-stone-900"
              >
                把纸条上的 Device ID
                <span class="block text-stone-600">准确填进 Ink 的设备编号</span>
              </h2>
              <p class="max-w-2xl text-base leading-7 text-stone-600 sm:text-[15px]">
                这页只解决一件事：让你第一次绑定咕咕机时，不用猜、不用试错，直接知道该看哪张纸、该抄哪一行、该填到哪里。
              </p>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-3">
            <article
              v-for="item in quickChecks"
              :key="item.title"
              class="rounded-2xl border border-white/70 bg-white/85 px-4 py-4 shadow-sm backdrop-blur"
            >
              <p class="text-sm font-semibold text-stone-900">{{ item.title }}</p>
              <p class="mt-2 text-sm leading-6 text-stone-600">{{ item.detail }}</p>
            </article>
          </div>

          <div class="flex flex-col gap-3 sm:flex-row">
            <RouterLink to="/status" class="ui-btn-primary px-4 py-2.5 text-center text-sm">
              去状态页添加设备
            </RouterLink>
            <RouterLink to="/prints" class="ui-btn-secondary px-4 py-2.5 text-center text-sm">
              绑定完后去试打印
            </RouterLink>
          </div>
        </div>

        <figure
          class="rounded-[1.8rem] border border-stone-200 bg-white p-4 shadow-xl shadow-stone-900/5"
        >
          <div class="flex items-center justify-between gap-3 px-1 pb-3">
            <div>
              <p class="text-sm font-medium text-stone-500">示例纸条</p>
              <p class="mt-1 text-base font-semibold text-stone-900">你会拿到一张这样的状态纸条</p>
            </div>
            <span
              class="inline-flex rounded-full border border-amber-200 bg-amber-50 px-3 py-1 text-xs font-medium text-amber-800"
            >
              重点看 Device ID
            </span>
          </div>

          <div class="overflow-hidden rounded-[1.35rem] border border-stone-200 bg-stone-100">
            <img
              src="/tutorial-device-id.jpg"
              alt="咕咕机连接 Wi-Fi 后打印出的状态纸条示例"
              class="h-full w-full object-cover"
              loading="lazy"
            />
          </div>

          <figcaption class="mt-4 grid gap-3 sm:grid-cols-2">
            <div class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3">
              <p class="text-sm font-semibold text-amber-900">先确认纸条顶部</p>
              <p class="mt-1 text-sm leading-6 text-amber-800">
                看到 <span class="font-medium">WiFi Connected</span>，说明设备当前已经联网。
              </p>
            </div>
            <div class="rounded-2xl border border-stone-200 bg-stone-900 px-4 py-3">
              <p class="text-sm font-semibold text-white">真正要抄的只有一行</p>
              <p class="mt-1 text-sm leading-6 text-stone-300">
                抄
                <span class="font-medium text-white">Device ID：</span>
                后面那一长串字符，不要抄错成 MAC Address。
              </p>
            </div>
          </figcaption>
        </figure>
      </div>
    </div>

    <div class="grid gap-6 xl:grid-cols-[minmax(0,1.2fr)_340px]">
      <article class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
        <div class="max-w-2xl">
          <p class="text-sm font-medium tracking-[0.18em] text-stone-500 uppercase">读取规则</p>
          <h3 class="mt-3 text-2xl font-semibold tracking-tight text-stone-900">
            你在 Ink 里应该填什么，不该填什么
          </h3>
          <p class="mt-3 text-sm leading-7 text-stone-600">
            状态纸条的信息不少，但 Ink 真正需要的只有
            <span class="font-medium text-stone-900"> Device ID </span>
            。下面这三项最容易看错，第一次绑定时建议对着纸条逐项核对。
          </p>
        </div>

        <div class="mt-6 grid gap-4 md:grid-cols-3">
          <article
            v-for="item in fieldGuides"
            :key="item.value"
            class="rounded-[1.4rem] border px-4 py-5"
            :class="
              item.positive ? 'border-emerald-200 bg-emerald-50' : 'border-stone-200 bg-stone-50'
            "
          >
            <p
              class="text-xs font-medium tracking-[0.16em] uppercase"
              :class="item.positive ? 'text-emerald-700' : 'text-stone-500'"
            >
              {{ item.label }}
            </p>
            <p
              class="mt-3 text-lg font-semibold"
              :class="item.positive ? 'text-emerald-950' : 'text-stone-900'"
            >
              {{ item.value }}
            </p>
            <p
              class="mt-2 text-sm leading-6"
              :class="item.positive ? 'text-emerald-800' : 'text-stone-600'"
            >
              {{ item.detail }}
            </p>
          </article>
        </div>
      </article>

      <aside class="rounded-[1.8rem] border border-stone-200 bg-stone-900 p-6 shadow-sm sm:p-7">
        <p class="text-sm font-medium tracking-[0.18em] text-stone-400 uppercase">绑定前检查</p>
        <h3 class="mt-3 text-2xl font-semibold tracking-tight text-white">开始前只核对这 3 件事</h3>
        <div class="mt-6 space-y-4">
          <article
            v-for="(item, index) in quickChecks"
            :key="item.title"
            class="rounded-2xl border border-white/10 bg-white/5 px-4 py-4"
          >
            <div class="flex items-start gap-3">
              <span
                class="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-white/10 text-sm font-semibold text-white"
              >
                {{ index + 1 }}
              </span>
              <div>
                <p class="text-sm font-semibold text-white">{{ item.title }}</p>
                <p class="mt-1 text-sm leading-6 text-stone-300">{{ item.detail }}</p>
              </div>
            </div>
          </article>
        </div>
      </aside>
    </div>

    <div class="grid gap-6 lg:grid-cols-[minmax(0,1.15fr)_0.85fr]">
      <article class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
        <div class="max-w-2xl">
          <p class="text-sm font-medium tracking-[0.18em] text-stone-500 uppercase">操作步骤</p>
          <h3 class="mt-3 text-2xl font-semibold tracking-tight text-stone-900">
            三步完成绑定并打印第一张纸条
          </h3>
        </div>

        <div class="mt-6 space-y-4">
          <article
            v-for="(step, index) in steps"
            :key="step.number"
            class="rounded-[1.5rem] border px-5 py-5"
            :class="
              index === 0 ? 'border-amber-200 bg-amber-50/80' : 'border-stone-200 bg-stone-50/70'
            "
          >
            <div class="grid gap-4 md:grid-cols-[auto_minmax(0,1fr)] md:items-start">
              <div
                class="inline-flex h-12 w-12 items-center justify-center rounded-2xl text-sm font-semibold"
                :class="index === 0 ? 'bg-amber-100 text-amber-900' : 'bg-white text-stone-900'"
              >
                {{ step.number }}
              </div>
              <div>
                <h4 class="text-lg font-semibold text-stone-900">{{ step.title }}</h4>
                <p class="mt-2 text-sm leading-7 text-stone-600">{{ step.body }}</p>
                <p
                  class="mt-4 rounded-2xl px-4 py-3 text-sm leading-6"
                  :class="index === 0 ? 'bg-white text-amber-900' : 'bg-white text-stone-600'"
                >
                  {{ step.note }}
                </p>
              </div>
            </div>
          </article>
        </div>
      </article>

      <div class="space-y-6">
        <article class="rounded-[1.8rem] border border-stone-200 bg-stone-50 p-6 shadow-sm sm:p-7">
          <p class="text-sm font-medium tracking-[0.18em] text-stone-500 uppercase">
            作为手机应用使用
          </p>
          <h3 class="mt-3 text-2xl font-semibold tracking-tight text-stone-900">
            把 Ink 添加到 iPhone 主屏幕
          </h3>
          <p class="mt-3 text-sm leading-7 text-stone-600">
            用 Safari 打开 Ink 后，点击底部分享按钮，选择“添加到主屏幕”。之后你可以像普通 App
            一样从桌面进入 Ink，顶部和底部导航都会按手机安全区适配。
          </p>
        </article>

        <article class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
          <h3 class="text-2xl font-semibold tracking-tight text-stone-900">常见问题</h3>
          <div class="mt-5 space-y-4">
            <article
              v-for="item in faqs"
              :key="item.question"
              class="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4"
            >
              <p class="text-sm font-semibold text-stone-900">{{ item.question }}</p>
              <p class="mt-2 text-sm leading-7 text-stone-600">{{ item.answer }}</p>
            </article>
          </div>
        </article>

        <aside class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
          <h3 class="text-xl font-semibold text-stone-900">开始使用</h3>
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
    </div>
  </section>
</template>

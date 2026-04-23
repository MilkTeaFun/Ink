<script setup lang="ts">
import { ref } from "vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { RouterLink } from "vue-router";

import AppDialog from "@/components/AppDialog.vue";
import { useWorkspaceStore } from "@/stores/workspace";

const workspaceStore = useWorkspaceStore();
const { t } = useI18n();

const steps = computed(
  () =>
    [
      {
        number: "01",
        title: t("tutorial.steps.chat.title"),
        body: t("tutorial.steps.chat.body"),
        note: t("tutorial.steps.chat.note"),
      },
      {
        number: "02",
        title: t("tutorial.steps.print.title"),
        body: t("tutorial.steps.print.body"),
        note: t("tutorial.steps.print.note"),
      },
      {
        number: "03",
        title: t("tutorial.steps.schedule.title"),
        body: t("tutorial.steps.schedule.body"),
        note: t("tutorial.steps.schedule.note"),
      },
    ] as const,
);

const faqs = computed(
  () =>
    [
      {
        question: t("tutorial.faqs.chatOrPrint.question"),
        answer: t("tutorial.faqs.chatOrPrint.answer"),
      },
      {
        question: t("tutorial.faqs.scheduleUse.question"),
        answer: t("tutorial.faqs.scheduleUse.answer"),
      },
      {
        question: t("tutorial.faqs.contentTooLong.question"),
        answer: t("tutorial.faqs.contentTooLong.answer"),
      },
      {
        question: t("tutorial.faqs.notPrinted.question"),
        answer: t("tutorial.faqs.notPrinted.answer"),
      },
    ] as const,
);

const feedbackOpen = ref(false);
const feedbackDraft = ref("");
const feedbackFormError = ref("");

function openFeedbackDialog() {
  feedbackOpen.value = true;
  feedbackDraft.value = "";
  feedbackFormError.value = "";
}

function closeFeedbackDialog() {
  feedbackOpen.value = false;
  feedbackFormError.value = "";
}

async function handleFeedbackSubmit() {
  feedbackFormError.value = "";

  if (!feedbackDraft.value.trim()) {
    feedbackFormError.value = t("feedback.errors.required");
    return;
  }

  const success = await workspaceStore.submitFeedback(feedbackDraft.value);
  if (!success) {
    feedbackFormError.value = workspaceStore.feedbackError;
    return;
  }

  feedbackDraft.value = "";
  closeFeedbackDialog();
}
</script>

<template>
  <section class="mx-auto max-w-6xl space-y-6 pt-4 sm:space-y-8 lg:space-y-10">
    <AppDialog
      :open="feedbackOpen"
      :title="t('feedback.dialog.title')"
      :description="t('feedback.dialog.description')"
      @close="closeFeedbackDialog"
    >
      <form class="space-y-4" @submit.prevent="handleFeedbackSubmit">
        <label class="block">
          <span class="mb-2 block text-sm font-medium text-stone-900">
            {{ t("feedback.dialog.contentLabel") }}
          </span>
          <textarea
            v-model="feedbackDraft"
            rows="6"
            :placeholder="t('feedback.dialog.placeholder')"
            class="w-full rounded-xl border border-stone-200 bg-white px-4 py-3 text-sm leading-7 text-stone-900 placeholder:text-stone-400 focus:border-stone-900 focus:ring-1 focus:ring-stone-900 focus:outline-none"
          />
        </label>

        <p v-if="feedbackFormError" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {{ feedbackFormError }}
        </p>

        <div class="flex flex-col gap-3 sm:flex-row sm:justify-end">
          <button
            type="button"
            class="ui-btn-secondary px-4 py-2.5 text-sm"
            @click="closeFeedbackDialog"
          >
            {{ t("common.actions.cancel") }}
          </button>
          <button
            class="ui-btn-primary px-4 py-2.5 text-sm"
            :disabled="workspaceStore.feedbackSubmitting"
          >
            {{
              workspaceStore.feedbackSubmitting
                ? t("feedback.dialog.submitting")
                : t("feedback.dialog.submit")
            }}
          </button>
        </div>
      </form>
    </AppDialog>

    <div
      class="tutorial-hero relative overflow-hidden rounded-[2rem] border border-stone-200 px-5 py-6 shadow-sm sm:px-7 sm:py-8 lg:px-10"
    >
      <div
        aria-hidden="true"
        class="tutorial-hero-glow tutorial-hero-glow-primary absolute top-0 right-0 h-40 w-40 rounded-full blur-3xl"
      />
      <div
        aria-hidden="true"
        class="tutorial-hero-glow tutorial-hero-glow-secondary absolute bottom-0 left-0 h-32 w-32 rounded-full blur-3xl"
      />

      <div class="relative space-y-6">
        <div class="space-y-4">
          <p class="text-sm font-medium tracking-[0.2em] text-stone-500 uppercase">
            {{ t("tutorial.hero.eyebrow") }}
          </p>
          <div class="space-y-3">
            <h2
              class="max-w-4xl text-[clamp(2rem,4.8vw,3.5rem)] font-semibold tracking-tight text-stone-900"
            >
              {{ t("tutorial.hero.title") }}
              <span class="block text-stone-600">{{ t("tutorial.hero.subtitle") }}</span>
            </h2>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-3">
          <RouterLink
            to="/conversations"
            class="tutorial-feature-card rounded-2xl border px-4 py-4 shadow-sm backdrop-blur transition-colors"
          >
            <p class="text-sm font-semibold text-stone-900">
              {{ t("tutorial.features.chat.title") }}
            </p>
            <p class="mt-2 text-sm leading-6 text-stone-600">
              {{ t("tutorial.features.chat.body") }}
            </p>
          </RouterLink>
          <RouterLink
            to="/prints"
            class="tutorial-feature-card rounded-2xl border px-4 py-4 shadow-sm backdrop-blur transition-colors"
          >
            <p class="text-sm font-semibold text-stone-900">
              {{ t("tutorial.features.print.title") }}
            </p>
            <p class="mt-2 text-sm leading-6 text-stone-600">
              {{ t("tutorial.features.print.body") }}
            </p>
          </RouterLink>
          <RouterLink
            to="/prints"
            class="tutorial-feature-card rounded-2xl border px-4 py-4 shadow-sm backdrop-blur transition-colors"
          >
            <p class="text-sm font-semibold text-stone-900">
              {{ t("tutorial.features.schedule.title") }}
            </p>
            <p class="mt-2 text-sm leading-6 text-stone-600">
              {{ t("tutorial.features.schedule.body") }}
            </p>
          </RouterLink>
        </div>

        <div class="flex flex-col gap-3 sm:flex-row">
          <RouterLink to="/conversations" class="ui-btn-primary px-4 py-2.5 text-center text-sm">
            {{ t("tutorial.actions.goToConversations") }}
          </RouterLink>
          <RouterLink to="/prints" class="ui-btn-secondary px-4 py-2.5 text-center text-sm">
            {{ t("tutorial.actions.goToPrints") }}
          </RouterLink>
        </div>
      </div>
    </div>

    <div class="grid gap-6 lg:grid-cols-[minmax(0,1.15fr)_0.85fr]">
      <article class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
        <div class="max-w-2xl">
          <p class="text-sm font-medium tracking-[0.18em] text-stone-500 uppercase">
            {{ t("tutorial.stepsSection.eyebrow") }}
          </p>
          <h3 class="mt-3 text-2xl font-semibold tracking-tight text-stone-900">
            {{ t("tutorial.stepsSection.title") }}
          </h3>
        </div>

        <div class="mt-6 space-y-4">
          <article
            v-for="(step, index) in steps"
            :key="step.number"
            class="rounded-[1.5rem] border px-5 py-5"
            :class="
              index === 0
                ? 'tutorial-step-highlight border-amber-200 bg-amber-50/80'
                : 'border-stone-200 bg-stone-50/70'
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
            {{ t("tutorial.mobile.eyebrow") }}
          </p>
          <h3 class="mt-3 text-2xl font-semibold tracking-tight text-stone-900">
            {{ t("tutorial.mobile.title") }}
          </h3>
          <p class="mt-3 text-sm leading-7 text-stone-600">
            {{ t("tutorial.mobile.body") }}
          </p>
        </article>

        <article class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
          <h3 class="text-2xl font-semibold tracking-tight text-stone-900">
            {{ t("tutorial.faqSection.title") }}
          </h3>
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

          <div class="mt-6 rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4">
            <p class="text-sm font-semibold text-stone-900">
              {{ t("tutorial.faqSection.missingQuestionTitle") }}
            </p>
            <p class="mt-2 text-sm leading-7 text-stone-600">
              {{ t("tutorial.faqSection.missingQuestionBody") }}
            </p>
            <button
              type="button"
              class="ui-btn-secondary mt-4 w-full px-4 py-2.5 text-sm sm:w-auto"
              @click="openFeedbackDialog"
            >
              {{ t("feedback.card.action") }}
            </button>
          </div>
        </article>

        <aside class="rounded-[1.8rem] border border-stone-200 bg-white p-6 shadow-sm sm:p-7">
          <h3 class="text-xl font-semibold text-stone-900">
            {{ t("tutorial.start.title") }}
          </h3>
          <div class="mt-5 space-y-3">
            <RouterLink
              to="/conversations"
              class="ui-btn-primary block w-full px-4 py-2.5 text-center text-sm"
            >
              {{ t("tutorial.actions.goToConversations") }}
            </RouterLink>
            <RouterLink
              to="/prints"
              class="ui-btn-secondary block w-full px-4 py-2.5 text-center text-sm"
            >
              {{ t("tutorial.actions.goToPrints") }}
            </RouterLink>
          </div>
        </aside>
      </div>
    </div>
  </section>
</template>

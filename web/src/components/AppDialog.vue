<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from "vue";

const props = defineProps<{
  open: boolean;
  title: string;
  description?: string;
}>();

const emit = defineEmits<{
  close: [];
}>();

const dialogRef = ref<HTMLElement | null>(null);
const closeButtonRef = ref<HTMLButtonElement | null>(null);
let previousActiveElement: HTMLElement | null = null;

function getFocusableElements() {
  if (!dialogRef.value) {
    return [];
  }

  return Array.from(
    dialogRef.value.querySelectorAll<HTMLElement>(
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  ).filter((element) => !element.hasAttribute("hidden") && element.tabIndex !== -1);
}

function restoreFocus() {
  previousActiveElement?.focus();
  previousActiveElement = null;
}

function trapFocus(event: KeyboardEvent) {
  const focusableElements = getFocusableElements();
  if (focusableElements.length === 0) {
    event.preventDefault();
    dialogRef.value?.focus();
    return;
  }

  const first = focusableElements[0];
  const last = focusableElements.at(-1);
  if (!last) {
    return;
  }

  const activeElement = document.activeElement;
  if (event.shiftKey && activeElement === first) {
    event.preventDefault();
    last.focus();
    return;
  }

  if (!event.shiftKey && activeElement === last) {
    event.preventDefault();
    first.focus();
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (!props.open) {
    return;
  }

  if (event.key === "Escape") {
    event.preventDefault();
    emit("close");
    return;
  }

  if (event.key === "Tab") {
    trapFocus(event);
  }
}

watch(
  () => props.open,
  async (open, wasOpen) => {
    if (typeof document === "undefined") {
      return;
    }

    if (open) {
      previousActiveElement =
        document.activeElement instanceof HTMLElement ? document.activeElement : null;
      document.addEventListener("keydown", handleKeydown);
      await nextTick();
      closeButtonRef.value?.focus();
      return;
    }

    document.removeEventListener("keydown", handleKeydown);
    if (wasOpen) {
      restoreFocus();
    }
  },
);

onBeforeUnmount(() => {
  if (typeof document !== "undefined") {
    document.removeEventListener("keydown", handleKeydown);
  }

  restoreFocus();
});
</script>

<template>
  <Transition name="dialog-fade">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center bg-stone-950/35 px-4 py-6 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <section
        ref="dialogRef"
        class="w-full max-w-md rounded-3xl border border-stone-200 bg-[var(--app-surface)] p-6 shadow-2xl shadow-stone-900/10"
        role="dialog"
        aria-modal="true"
        tabindex="-1"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <h3 class="text-lg font-semibold text-stone-900">{{ title }}</h3>
            <p v-if="description" class="mt-1 text-sm text-stone-500">{{ description }}</p>
          </div>
          <button
            ref="closeButtonRef"
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded-full border border-stone-200 bg-white text-stone-500 transition-colors hover:border-stone-300 hover:text-stone-900"
            aria-label="关闭窗口"
            @click="emit('close')"
          >
            ×
          </button>
        </div>

        <div class="mt-5">
          <slot />
        </div>
      </section>
    </div>
  </Transition>
</template>

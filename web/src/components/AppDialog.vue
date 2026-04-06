<script setup lang="ts">
defineProps<{
  open: boolean;
  title: string;
  description?: string;
}>();

const emit = defineEmits<{
  close: [];
}>();
</script>

<template>
  <Transition name="dialog-fade">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center bg-stone-950/35 px-4 py-6 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <section
        class="w-full max-w-md rounded-3xl border border-stone-200 bg-[var(--app-surface)] p-6 shadow-2xl shadow-stone-900/10"
        role="dialog"
        aria-modal="true"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <h3 class="text-lg font-semibold text-stone-900">{{ title }}</h3>
            <p v-if="description" class="mt-1 text-sm text-stone-500">{{ description }}</p>
          </div>
          <button
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

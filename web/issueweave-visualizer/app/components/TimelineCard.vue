<template>
  <div :class="['flex w-full', alignRight ? 'justify-end' : 'justify-start']">
    <div
      v-if="variant === 'explicit'"
      class="w-full md:w-2/3 bg-white p-10 border border-zinc-200 shadow-xl relative group hover:-translate-y-2 transition-transform"
    >
      <span
        class="absolute -top-8 -left-4 text-7xl font-black text-zinc-200 z-[-1] opacity-50 group-hover:text-blue-100 transition-colors"
      >
        Explicit
      </span>
      <div class="flex items-center justify-between mb-4">
        <span class="text-2xl font-bold text-zinc-400">
          #{{ issue.number }}
        </span>
        <span
          class="px-4 py-1 text-sm font-bold uppercase tracking-widest bg-zinc-100 text-zinc-600"
        >
          {{ issue.state }}
        </span>
      </div>
      <h3 class="text-4xl font-bold leading-tight mb-6">{{ issue.title }}</h3>
      <a
        :href="issue.url"
        target="_blank"
        class="text-lg font-bold text-blue-500 hover:text-blue-700"
      >
        Explore →
      </a>
    </div>

    <!-- Comment Mention Variant -->
    <div
      v-else
      class="w-full md:w-1/2 bg-zinc-900 text-white p-10 shadow-2xl relative hover:scale-105 transition-transform"
    >
      <span
        class="absolute -bottom-6 -right-4 text-6xl font-black text-zinc-800 z-[-1]"
      >
        Mention
      </span>
      <div class="flex items-center justify-between mb-4">
        <span class="text-xl font-bold text-zinc-500">#{{ issue.number }}</span>
      </div>
      <h3 class="text-3xl font-semibold leading-tight mb-6">
        {{ issue.title }}
      </h3>
      <a
        :href="issue.url"
        target="_blank"
        class="text-lg font-bold text-zinc-400 hover:text-white border-b border-zinc-600 pb-1"
      >
        View Context
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  issue: { number: number; title: string; state: string; url: string };
  variant: "explicit" | "mention";
  alignRight: boolean;
}>();
</script>

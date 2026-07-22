<template>
  <div class="relative flex pt-4 pb-4">
    <div
      class="absolute left-0 mt-1 flex h-8 w-8 items-center justify-center rounded-full bg-[#161b22] border border-[#30363d]"
    >
      <svg
        v-if="variant === 'explicit'"
        class="h-4 w-4 text-[#8b949e]"
        viewBox="0 0 16 16"
        fill="currentColor"
      >
        <path
          fill-rule="evenodd"
          d="M8 0a8 8 0 100 16A8 8 0 008 0zm1.5 4.5a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0zm-3 4c0-.55.45-1 1-1h2c.55 0 1 .45 1 1v4c0 .55-.45 1-1 1h-2c-.55 0-1-.45-1-1v-4z"
        ></path>
      </svg>
      <svg
        v-else
        class="h-4 w-4 text-[#8b949e]"
        viewBox="0 0 16 16"
        fill="currentColor"
      >
        <path
          fill-rule="evenodd"
          d="M1.5 2.75a.25.25 0 01.25-.25h8.5a.25.25 0 01.25.25v5.5a.25.25 0 01-.25.25h-3.5a.75.75 0 00-.53.22L3.5 11.44V9.25a.75.75 0 00-.75-.75h-1a.25.25 0 01-.25-.25v-5.5zM1.75 1A1.75 1.75 0 000 2.75v5.5C0 9.216.784 10 1.75 10H2.75v2.25c0 .604.767.873 1.156.402l2.36-2.652h4.484A1.75 1.75 0 0012.5 8.25v-5.5A1.75 1.75 0 0010.75 1h-9z"
        ></path>
      </svg>
    </div>

    <div class="ml-12 w-full max-w-4xl">
      <div
        class="rounded-md border border-[#30363d] bg-[#0d1117] overflow-hidden"
      >
        <div
          class="flex items-center justify-between border-b border-[#30363d] bg-[#161b22] px-4 py-2"
        >
          <div class="text-sm text-[#8b949e]">
            <span
              v-if="variant === 'explicit'"
              class="font-semibold text-[#c9d1d9]"
            >
              Cross-referenced
            </span>
            <span v-else class="font-semibold text-[#c9d1d9]">
              Mentioned in comment
            </span>
          </div>
          <span
            class="rounded-full px-2.5 py-0.5 text-xs font-medium capitalize"
            :class="
              issue.state === 'open'
                ? 'bg-[#238636]/10 text-[#3fb950]'
                : 'bg-[#8957e5]/10 text-[#a371f7]'
            "
          >
            {{ issue.state }}
          </span>
        </div>

        <div class="p-4">
          <h3 class="text-base font-semibold text-[#c9d1d9] mb-1">
            <a
              :href="issue.url"
              target="_blank"
              class="hover:text-[#58a6ff] hover:underline"
            >
              {{ issue.title }}
            </a>
            <span class="text-[#8b949e] font-normal ml-2">
              #{{ issue.number }}
            </span>
          </h3>
          <div class="text-sm text-[#8b949e] mt-3">
            <a
              :href="issue.url"
              target="_blank"
              class="text-[#58a6ff] hover:underline"
            >
              View on GitHub →
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  issue: {
    number: number;
    title: string;
    state: string;
    url: string;
    date?: string;
  };
  variant: "explicit" | "mention";
}>();
</script>

<template>
  <main class="min-h-screen bg-[#0d1117] text-[#c9d1d9] font-sans">
    <div
      class="bg-[#161b22] border-b border-[#30363d] px-6 py-4 flex items-center gap-2 text-sm"
    >
      <span class="text-[#58a6ff] font-semibold">IssueWeave</span>
      <span class="text-[#8b949e]">/</span>
      <span class="font-semibold text-[#c9d1d9]">Timeline View</span>
    </div>

    <div class="max-w-5xl mx-auto p-6 md:p-12">
      <div v-if="pending" class="text-center py-12 text-[#8b949e]">
        Loading weave data...
      </div>
      <div v-else-if="error" class="text-center py-12 text-[#f85149]">
        Failed to fetch data. Ensure the Go API is running.
      </div>

      <div v-else-if="data">
        <header class="mb-8 pb-6 border-b border-[#30363d]">
          <h1
            class="text-3xl font-normal mb-3 tracking-tight text-white flex flex-wrap items-baseline gap-2"
          >
            {{ data.main_issue.title }}
            <span class="text-[#8b949e] font-light">
              #{{ data.main_issue.number }}
            </span>
          </h1>
          <div class="flex items-center gap-3 text-sm text-[#8b949e]">
            <span
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-full font-medium text-white capitalize"
              :class="
                data.main_issue.state === 'open'
                  ? 'bg-[#238636]'
                  : 'bg-[#8957e5]'
              "
            >
              <svg
                v-if="data.main_issue.state === 'open'"
                class="w-4 h-4"
                viewBox="0 0 16 16"
                fill="currentColor"
              >
                <path d="M8 9.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3z"></path>
                <path
                  fill-rule="evenodd"
                  d="M8 0a8 8 0 100 16A8 8 0 008 0zM1.5 8a6.5 6.5 0 1113 0 6.5 6.5 0 01-13 0z"
                ></path>
              </svg>
              <svg
                v-else
                class="w-4 h-4"
                viewBox="0 0 16 16"
                fill="currentColor"
              >
                <path
                  fill-rule="evenodd"
                  d="M1.5 8a6.5 6.5 0 0110.65-5.09l-2.85 2.84a1.5 1.5 0 00-2.12 2.12l-2.84 2.85A6.5 6.5 0 011.5 8zM8 14.5A6.5 6.5 0 013.35 13.09l2.85-2.84a1.5 1.5 0 002.12-2.12l2.84-2.85A6.5 6.5 0 018 14.5z"
                ></path>
              </svg>
              {{ data.main_issue.state }}
            </span>
            <span>Main Issue</span>
            <span>•</span>
            <a
              :href="data.main_issue.url"
              target="_blank"
              class="text-[#58a6ff] hover:underline"
            >
              View on GitHub
            </a>
          </div>
        </header>

        <div class="relative ml-4">
          <div
            class="absolute top-0 bottom-0 left-[15px] w-0.5 bg-[#21262d] -z-10"
          ></div>

          <TimelineEvent
            v-for="ref in data.direct_references"
            :key="'ref-' + ref.number"
            :issue="ref"
            variant="explicit"
          />

          <TimelineEvent
            v-for="mention in data.comment_references"
            :key="'mention-' + mention.number"
            :issue="mention"
            variant="mention"
          />
        </div>

        <div class="mt-4 ml-4 flex items-center gap-4 text-[#8b949e]">
          <div
            class="flex h-8 w-8 items-center justify-center rounded-full bg-[#161b22] border border-[#30363d]"
          >
            <svg class="h-4 w-4" viewBox="0 0 16 16" fill="currentColor">
              <path
                fill-rule="evenodd"
                d="M10.5 7.75a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0zm1.43.75a4.002 4.002 0 01-7.86 0H.75a.75.75 0 110-1.5h3.32a4.001 4.001 0 017.86 0h3.32a.75.75 0 110 1.5h-3.32z"
              ></path>
            </svg>
          </div>
          <span class="text-sm font-semibold">End of Timeline</span>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
interface IssueData {
  number: number;
  title: string;
  state: string;
  url: string;
  date?: string;
}

interface WeaveResult {
  main_issue: IssueData;
  direct_references: IssueData[];
  comment_references: IssueData[];
}

const { data, pending, error } = await useFetch<WeaveResult>(
  "http://localhost:8080/api/weave",
);
</script>

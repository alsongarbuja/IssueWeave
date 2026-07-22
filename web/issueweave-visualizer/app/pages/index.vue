<template>
  <main class="min-h-screen bg-zinc-50 p-12 overflow-x-hidden">
    <!-- Loading & Error States -->
    <div v-if="pending" class="text-2xl font-bold text-zinc-400">
      Loading weave data...
    </div>
    <div v-else-if="error" class="text-2xl text-red-500">
      Failed to connect to Go server.
    </div>

    <!-- Main Layout -->
    <div v-else-if="data">
      <header class="mb-24 max-w-5xl">
        <h1
          class="text-8xl font-black tracking-tighter text-zinc-900 mb-6 leading-none"
        >
          Issue
          <br />
          #{{ data?.main_issue.number }}
        </h1>
        <p class="text-3xl text-zinc-500 font-medium max-w-3xl">
          {{ data?.main_issue.title }}
        </p>
        <a
          :href="data?.main_issue.url"
          target="_blank"
          class="inline-block mt-8 text-xl font-bold text-blue-600 hover:underline"
        >
          View Original →
        </a>
      </header>

      <h2 class="text-4xl font-bold text-zinc-800 mb-12">The Weave Timeline</h2>

      <div
        class="flex flex-col gap-16 md:gap-32 w-full max-w-6xl mx-auto pb-32"
      >
        <TimelineCard
          v-for="(ref, i) in data?.direct_references"
          :key="'ref-' + ref.number"
          :issue="ref"
          variant="explicit"
          :alignRight="i % 2 !== 0"
        />

        <TimelineCard
          v-for="(mention, i) in data?.comment_references"
          :key="'mention-' + mention.number"
          :issue="mention"
          variant="mention"
          :alignRight="i % 2 === 0"
        />
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
// Define the expected shape of the JSON from your Go API
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

// Call the local Go API. Nuxt handles the async state automatically.
const { data, pending, error } = await useFetch<WeaveResult>(
  "http://localhost:8080/api/weave",
);

console.log(data.value);
</script>

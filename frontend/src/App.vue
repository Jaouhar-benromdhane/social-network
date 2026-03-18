<template>
  <main class="layout">
    <section class="card">
      <p class="eyebrow">Mandatory Project</p>
      <h1>Social Network</h1>
      <p class="subtitle">
        M1 skeleton is running. Backend and migrations are wired.
      </p>

      <div class="status-row">
        <span class="label">API health:</span>
        <strong :class="statusClass">{{ status }}</strong>
      </div>

      <p v-if="error" class="error">{{ error }}</p>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";

const status = ref("checking");
const error = ref("");

const statusClass = computed(() => {
  if (status.value === "ok") return "ok";
  if (status.value === "degraded") return "warn";
  return "pending";
});

onMounted(async () => {
  try {
    const response = await fetch("/api/health");
    const payload = await response.json();
    status.value = payload.status || "unknown";
  } catch (_err) {
    status.value = "unreachable";
    error.value = "Backend is not reachable yet. Start backend server first.";
  }
});
</script>

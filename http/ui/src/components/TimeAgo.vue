<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue';
import TimeAgo from 'javascript-time-ago';
import timeAgoEn from 'javascript-time-ago/locale/en.json';

TimeAgo.addLocale(timeAgoEn);
const timeAgo = new TimeAgo('en-US');

const props = defineProps<{ timestamp: string }>();
const formattedTimestamp = ref<string>();

function formatTimestamp(timestamp: string): string {
  return timeAgo.format(new Date(timestamp), 'round-minute') as string;
}

function updateformattedTimestamp(): void {
  formattedTimestamp.value = formatTimestamp(props.timestamp);
}

watch(() => props.timestamp, updateformattedTimestamp, { immediate: true });

let updateFormattedTimestampInterval: number;

onMounted(() => {
  updateFormattedTimestampInterval = setInterval(updateformattedTimestamp, 1_000);
});

onBeforeUnmount(() => {
  clearInterval(updateFormattedTimestampInterval);
});
</script>

<template>
  <span>{{ formattedTimestamp }}</span>
</template>

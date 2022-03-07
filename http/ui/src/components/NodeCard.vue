<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { EventData } from '../gesundheit'
import EventCard from './EventCard.vue';
import Dot from './Dot.vue';
import Card from './Card.vue';

const props = defineProps<{
  name: string,
  events: Array<EventData>,
  forceOpen: boolean,
}>()

const healthy = computed(() => (
  props.events.every((event) => event.Status === 0)
));

const isOpen = ref(!healthy.value || props.forceOpen);

watch(healthy, (healthy) => {
  if (!healthy) isOpen.value = true;
});

watch(() => props.forceOpen, (forceOpen) => {
  if (forceOpen) isOpen.value = true;
});

const sortedEvents = computed(() => (
  [...props.events].sort((a, b) => {
    if (a.Status < b.Status) return 1;
    if (a.Status > b.Status) return -1;
    return b.Timestamp.localeCompare(a.Timestamp);
  })
))
</script>

<template>
  <Card v-model:is-open="isOpen">
    <template #header>
      <Dot
        :pulse="!healthy"
        :danger="!healthy"
        class="flex-shrink-0 me-3"
      />
      <div class="me-auto">
        {{ name }}
      </div>
    </template>
    <template #body>
      <EventCard
        v-for="event in sortedEvents"
        :key="event.CheckId"
        :event="event"
        class="event mb-2"
      />
    </template>
  </Card>
</template>

<style lang="scss" scoped>
.event:last-child {
  margin-bottom: 0 !important;
}
</style>

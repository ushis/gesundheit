<script setup lang="ts">
import { computed } from 'vue';
import { EventData } from '../gesundheit';
import EventCard from './EventCard.vue';
import Dot from './Dot.vue';
import Card from './Card.vue';

const props = defineProps<{
  name: string,
  events: Array<EventData>,
  isOpen: boolean,
}>();

const isHealthy = computed(() => (
  props.events.every((event) => event.Status === 0)
));

const isOpen = computed(() => (
  !isHealthy.value || props.isOpen
));

const sortedEvents = computed(() => (
  [...props.events].sort((a, b) => {
    if (a.Status < b.Status) return 1;
    if (a.Status > b.Status) return -1;
    return b.Timestamp.localeCompare(a.Timestamp);
  })
));
</script>

<template>
  <Card :is-open="isOpen">
    <template #header>
      <Dot
        :pulse="!isHealthy"
        :danger="!isHealthy"
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

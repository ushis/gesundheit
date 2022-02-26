<script setup lang="ts">
import { computed } from 'vue';
import { Event } from '../gesundheit'
import EventComponent from './Event.vue';
import Dot from './Dot.vue';

const props = defineProps<{
  name: string,
  events: Array<Event>,
}>()

const healthy = computed(() => (
  props.events.every((event) => event.Result === 0)
));

const sortedEvents = computed(() => (
  [...props.events].sort((a, b) => {
    if (a.Result < b.Result) return 1;
    if (a.Result > b.Result) return -1;
    return b.Timestamp.localeCompare(a.Timestamp);
  })
))
</script>

<template>
  <div class="card">
    <div class="card-header d-flex align-items-center">
      <Dot
        :pulse="!healthy"
        :danger="!healthy"
        class="flex-shrink-0 me-3"
      />
      <div>{{ name }}</div>
    </div>

    <div class="card-body">
      <EventComponent
        v-for="event in sortedEvents"
        :key="event.CheckId"
        :event="event"
        class="event mb-2"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.event:last-child {
  margin-bottom: 0 !important;
}
</style>

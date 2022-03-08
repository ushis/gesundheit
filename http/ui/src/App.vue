<script setup lang="ts">
import { ref, computed, onBeforeMount } from 'vue';
import { EventData, EventStream } from './gesundheit';
import { groupBy } from './util';
import NavBar from './components/NavBar.vue';
import NodeCard from './components/NodeCard.vue';

const allEvents = ref([] as Array<EventData>);
const filter = ref('');

const normalFilter = computed(() => (
  filter.value.trim().toLocaleLowerCase()
));

const filteredEvents = computed(() => {
  if (normalFilter.value === '') return allEvents.value;

  return allEvents.value.filter((e) => (
    e.CheckDescription.toLocaleLowerCase().includes(normalFilter.value)
  ));
});

const eventsByNode = computed(() => (
  groupBy(filteredEvents.value, (e) => e.NodeName)
    .sort(([a], [b]) => a.localeCompare(b))
));

const healthy = computed(() => (
  allEvents.value.every((event) => event.Status === 0)
));

const stream = new EventStream((event) => {
  const i = allEvents.value.findIndex((e) => (
    e.NodeName === event.NodeName &&
      e.CheckId === event.CheckId
  ));

  if (i < 0) {
    allEvents.value.push(event);
  } else {
    allEvents.value[i] = event;
  }
});

onBeforeMount(() => stream.connect());
</script>

<template>
  <NavBar
    v-model:filter="filter"
    :is-healthy="healthy"
  />
  <div class="container py-4">
    <NodeCard
      v-for="([nodeName, events]) in eventsByNode"
      :key="nodeName"
      :name="nodeName"
      :events="events"
      :is-open="normalFilter !== '' || eventsByNode.length === 1"
      class="mb-3"
    />
  </div>
</template>


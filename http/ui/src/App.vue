<script setup lang="ts">
import { ref, computed, onBeforeMount } from 'vue';
import { EventData, EventStream } from './gesundheit';
import { groupBy } from './util';
import NavBar from './components/NavBar.vue';
import NodeCard from './components/NodeCard.vue';

const filter = ref('');
const events = ref([] as Array<EventData>);

const eventsByNode = computed(() => (
  groupBy(events.value, (e) => e.NodeName)
    .sort(([a], [b]) => a.localeCompare(b))
));

const healthy = computed(() => (
  events.value.every((event) => event.Status === 0)
));

const stream = new EventStream((event) => {
  const i = events.value.findIndex((e) => (
    e.NodeName === event.NodeName &&
      e.CheckId === event.CheckId
  ));

  if (i < 0) {
    events.value.push(event);
  } else {
    events.value[i] = event;
  }
});

onBeforeMount(() => stream.connect());
</script>

<template>
  <NavBar
    v-model:filter="filter"
    :is-healthy="healthy"
  />
  <div class="container py-3">
    <NodeCard
      v-for="([nodeName, nodeEvents]) in eventsByNode"
      :key="nodeName"
      :name="nodeName"
      :events="nodeEvents"
      :filter="filter"
      :is-open="eventsByNode.length === 1"
      class="mb-3"
    />
  </div>
</template>


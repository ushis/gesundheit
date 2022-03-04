<script setup lang="ts">
import { ref, computed, onBeforeMount } from 'vue'
import { EventData, EventStream } from './gesundheit';
import Dot from './components/Dot.vue';
import Node from './components/Node.vue';

const nodes = ref(new Map() as Map<string, Array<EventData>>);

const sortedNodes = computed(() => (
  Array.from(nodes.value.entries())
    .sort(([a], [b]) => a.localeCompare(b))
));

const healthy = computed(() => (
  sortedNodes.value.every(([, events]) => (
    events.every((event) => event.Status === 0)
  ))
));

const stream = new EventStream((event) => {
  let events = nodes.value.get(event.NodeName);

  if (events === undefined) {
    events = [event];
  } else {
    events = events.filter((e) => e.CheckId !== event.CheckId);
    events.push(event);
  }
  nodes.value.set(event.NodeName, events);
});

onBeforeMount(() => stream.connect());
</script>

<template>
  <nav class="navbar navbar-expand-lg navbar-light bg-light">
    <div class="container">
      <div class="navbar-brand">
        <Dot
          :danger="!healthy"
          :pulse="!healthy"
          class="me-3"
        />
        <span>gesundheit</span>
      </div>
    </div>
  </nav>
  <div class="container py-4">
    <Node
      v-for="([name, events]) in sortedNodes"
      :key="name"
      :name="name"
      :events="events"
      class="mb-4"
    />
  </div>
</template>


<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { EventData } from '../gesundheit'
import Dot from './Dot.vue';
import TimeAgo from './TimeAgo.vue';
import Card from './Card.vue';

const props = defineProps<{ event: EventData }>();
const healthy = computed(() => props.event.Status === 0);
const isOpen = ref(!healthy.value);

watch(healthy, (healthy) => {
  if (!healthy) isOpen.value = true;
});
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
        {{ event.CheckDescription }}
      </div>
      <TimeAgo
        :timestamp="event.Timestamp"
        class="text-truncate d-none d-sm-block"
      />
    </template>
    <template #body>
      <code class="text-dark">{{ event.Message }}</code>
    </template>
  </Card>
</template>

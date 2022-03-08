<script setup lang="ts">
import { computed } from 'vue';
import { EventData } from '../gesundheit';
import Dot from './Dot.vue';
import TimeAgo from './TimeAgo.vue';
import Card from './Card.vue';

const props = defineProps<{ event: EventData }>();
const isHealthy = computed(() => props.event.Status === 0);
</script>

<template>
  <Card :is-open="!isHealthy">
    <template #header>
      <Dot
        :pulse="!isHealthy"
        :danger="!isHealthy"
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

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { Event } from '../gesundheit'
import Dot from './Dot.vue';
import TimeAgo from './TimeAgo.vue';

const props = defineProps<{ event: Event }>();
const healthy = computed(() => props.event.Result === 0);
const isOpen = ref(!healthy.value);

watch(healthy, (healthy) => {
  if (!healthy) isOpen.value = true;
});
</script>

<template>
  <div class="card">
    <div
      class="card-header d-flex align-items-center"
      @click="isOpen = !isOpen"
    >
      <Dot
        :pulse="!healthy"
        :danger="!healthy"
        class="flex-shrink-0 me-3"
      />
      <div class="me-auto">
        {{ event.CheckDescription }}
      </div>
      <div class="text-truncate d-none d-sm-block">
        <TimeAgo :timestamp="event.Timestamp" />
      </div>
    </div>
    <div
      class="card-body"
      :class="{ 'd-none': !isOpen }"
    >
      <code class="text-dark">{{ event.Message }}</code>
    </div>
  </div>
</template>

<style scoped lang="scss">
.card-header {
  cursor: pointer;
}
</style>

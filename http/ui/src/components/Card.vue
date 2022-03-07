<script setup lang="ts">
import ExpandTransition from './ExpandTransition.vue';
import Chevron from './Chevron.vue';

const props = defineProps<{ isOpen: boolean }>()
const emit = defineEmits<{ (event: 'update:isOpen', isOpen: boolean): void }>();
</script>

<template>
  <div class="card">
    <div
      class="card-header d-flex align-items-center"
      @click.prevent="emit('update:isOpen', !props.isOpen)"
    >
      <slot name="header" />
      <Chevron
        class="flex-shrink-0 ms-3"
        :up="props.isOpen"
      />
    </div>

    <ExpandTransition>
      <div v-show="props.isOpen">
        <div class="card-body">
          <slot name="body" />
        </div>
      </div>
    </ExpandTransition>
  </div>
</template>

<style scoped lang="scss">
.card-header {
  cursor: pointer;
}

.card-header:hover {
  background-color: rgba(0, 0, 0, 0.06);
}
</style>

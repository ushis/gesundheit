<script setup lang="ts">
import { ref, watch } from 'vue';
import ExpandTransition from './ExpandTransition.vue';
import Chevron from './Chevron.vue';

const props = defineProps<{ isOpen: boolean }>();
const isReallyOpen = ref(props.isOpen);
const isOpenedByUser = ref(false);

const onHeaderClick = () => {
  isReallyOpen.value = !isReallyOpen.value;
  isOpenedByUser.value = isReallyOpen.value;
};

watch(() => props.isOpen, () => {
  if (props.isOpen) {
    isReallyOpen.value = true;
  } else if (!isOpenedByUser.value) {
    isReallyOpen.value = false;
  }
});
</script>

<template>
  <div class="card">
    <div
      class="card-header d-flex align-items-center"
      @click.prevent="onHeaderClick"
    >
      <slot name="header" />
      <Chevron
        class="flex-shrink-0 ms-3"
        :up="isReallyOpen"
      />
    </div>

    <ExpandTransition>
      <div v-show="isReallyOpen">
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

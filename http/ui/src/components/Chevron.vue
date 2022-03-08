<script setup lang="ts">
const props = defineProps<{ up: boolean }>();
</script>

<template>
  <svg
    width="12"
    height="12"
    viewBox="0 0 12 12"
    class="chevron"
    :class="{ 'chevron-up': props.up }"
  >
    <path
      fill="currentColor"
      d="M 0 3 L 6 9 L 12 3 L 10 3 L 6 7 L 2 3"
    />
  </svg>
</template>

<style scoped lang="scss">
.chevron {
  transition: transform 0.2s;
}

.chevron.chevron-up {
  transform: rotate(-180deg);
}
</style>

<script setup lang="ts">
withDefaults(defineProps<{
  pulse: boolean,
  danger: boolean,
}>(), {
  pulse: false,
  danger: false,
});
</script>

<template>
  <span
    class="dot"
    :class="{ 'dot-pulse': pulse, 'dot-danger': danger }"
  />
</template>

<style scoped lang="scss">
@import 'bootstrap/scss/functions';
@import 'bootstrap/scss/_variables.scss';

@keyframes dot-pulse {
  0%   { box-shadow: 0 0 0 0 rgba($success, 0.5); }
  70%  { box-shadow: 0 0 0 10px rgba($success, 0); }
  100% { box-shadow: 0 0 0 0 rgba($success, 0); }
}

@keyframes dot-pulse-danger {
  0%   { box-shadow: 0 0 0 0 rgba($danger, 0.5); }
  70%  { box-shadow: 0 0 0 10px rgba($danger, 0); }
  100% { box-shadow: 0 0 0 0 rgba($danger, 0); }
}

.dot {
  display: inline-block;
  border-radius: 50%;
  width: 10px;
  height: 10px;
  background-color: $success;
}

.dot.dot-danger {
  background-color: $danger;
}

.dot.dot-pulse {
  animation-name: dot-pulse;
  animation-duration: 2s;
  animation-timing-function: ease-in-out;
  animation-iteration-count: infinite;
  animation-direction: normal;
  animation-fill-mode: none;
}

.dot.dot-pulse.dot-danger {
  animation-name: dot-pulse-danger;
}
</style>

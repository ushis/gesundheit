<script setup lang="ts">
function enter(el: HTMLElement) {
  const { height } = getComputedStyle(el);

  requestAnimationFrame(() => {
    el.style.height = '0';

    requestAnimationFrame(() => {
      el.style.height = height;
    });
  })
}

function leave(el: HTMLElement) {
  const { height } = getComputedStyle(el);

  requestAnimationFrame(() => {
    el.style.height = height;

    requestAnimationFrame(() => {
      el.style.height = '0';
    });
  })
}

function after(el: HTMLElement) {
  el.style.height = '';
}
</script>

<template>
  <Transition
    name="expand"
    @enter="enter"
    @after-enter="after"
    @leave="leave"
    @after-leave="after"
  >
    <slot />
  </Transition>
</template>

<style scoped lang="scss">
.expand-enter-active,
.expand-leave-active {
  transition: height 0.2s;
  overflow: hidden;
}
</style>

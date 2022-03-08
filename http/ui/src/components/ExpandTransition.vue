<script setup lang="ts">
function enter(elem: Element) {
  const el = elem as HTMLElement;
  const { height } = getComputedStyle(el);

  requestAnimationFrame(() => {
    el.style.height = '0';

    requestAnimationFrame(() => {
      el.style.height = height;
    });
  });
}

function leave(elem: Element) {
  const el = elem as HTMLElement;
  const { height } = getComputedStyle(el);

  requestAnimationFrame(() => {
    el.style.height = height;

    requestAnimationFrame(() => {
      el.style.height = '0';
    });
  });
}

function after(elem: Element) {
  const el = elem as HTMLElement;
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

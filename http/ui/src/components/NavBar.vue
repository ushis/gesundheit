<script setup lang="ts">
import { ref, watch } from 'vue';
import Dot from './Dot.vue';

const props = defineProps<{
  isHealthy: boolean,
  filter: string,
}>();

const emit = defineEmits<{
  (event: 'update:filter', value: string): void,
}>();

const onFilterInput = (e: Event) => {
  emit('update:filter', (e.target as HTMLInputElement).value)
};

const menu = ref<HTMLElement | null>(null);
const menuOpen = ref(false);

const openMenu = (el: HTMLElement) => {
  el.classList.add('show');
  const { height } = getComputedStyle(el);

  requestAnimationFrame(() => {
    el.classList.add('collapsing');

    requestAnimationFrame(() => {
      el.style.height = height;
    });
  });

  el.addEventListener('transitionend', () => {
    el.classList.remove('collapsing');
    el.style.height = '';
  }, { once: true });
};

const closeMenu = (el: HTMLElement) => {
  const { height } = getComputedStyle(el);

  requestAnimationFrame(() => {
    el.style.height = height;

    requestAnimationFrame(() => {
      el.classList.add('collapsing');
      el.style.height = '';
    });
  });

  el.addEventListener('transitionend', () => {
    el.classList.remove('collapsing');
    el.classList.remove('show');
  }, { once: true });
};

watch(menuOpen, () => {
  if (menu.value === null) return;

  if (menuOpen.value) {
    openMenu(menu.value)
  } else {
    closeMenu(menu.value);
  }
});
</script>

<template>
  <nav class="navbar navbar-expand-lg navbar-light bg-light">
    <div class="container">
      <div class="navbar-brand">
        <Dot
          :danger="!props.isHealthy"
          :pulse="!props.isHealthy"
          class="me-3"
        />
        <span>gesundheit</span>
      </div>
      <button
        class="navbar-toggler"
        @click.prevent="menuOpen = !menuOpen"
      >
        <span class="navbar-toggler-icon" />
      </button>
      <div
        ref="menu"
        class="collapse navbar-collapse"
      >
        <form
          class="ms-auto pt-2"
          @submit.prevent
        >
          <input
            :value="props.filter"
            class="form-control"
            type="search"
            placeholder="Search..."
            @input="onFilterInput"
          >
        </form>
      </div>
    </div>
  </nav>
</template>

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
  emit('update:filter', (e.target as HTMLInputElement).value);
};

const menu = ref<HTMLElement | null>(null);
const menuOpen = ref(false);

const openMenu = () => {
  if (menu.value === null) return;

  const el = menu.value;
  el.classList.add('show');
  const { height } = getComputedStyle(el);

  requestAnimationFrame(() => {
    el.style.height = '0';
    el.style.overflow = 'hidden';
    el.style.transition = 'height 0.2s';

    requestAnimationFrame(() => {
      el.style.height = height;
    });
  });

  el.addEventListener('transitionend', () => {
    el.style.height = '';
    el.style.overflow = '';
    el.style.transition = '';
  }, { once: true });
};

const closeMenu = () => {
  if (menu.value === null) return;

  const el = menu.value;
  const { height } = getComputedStyle(el);

  requestAnimationFrame(() => {
    el.style.height = height;
    el.style.overflow = 'hidden';
    el.style.transition = 'height 0.2s';

    requestAnimationFrame(() => {
      el.style.height = '0';
    });
  });

  el.addEventListener('transitionend', () => {
    el.style.height = '';
    el.style.overflow = '';
    el.style.transition = '';
    el.classList.remove('show');
  }, { once: true });
};

watch(menuOpen, () => {
  if (menuOpen.value) {
    openMenu();
  } else {
    closeMenu();
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

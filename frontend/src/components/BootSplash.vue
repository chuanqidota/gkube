<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import Logo from './Logo.vue'

const emit = defineEmits<{ done: [] }>()

const reduced = ref(false)
let timer: number | undefined

onMounted(() => {
  const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
  reduced.value = mq.matches
  // Hard ceiling — the splash never strands the user, no matter what.
  timer = window.setTimeout(() => emit('done'), reduced.value ? 900 : 1500)
})

onBeforeUnmount(() => {
  if (timer) window.clearTimeout(timer)
})
</script>

<template>
  <div class="gk-boot" :class="{ 'gk-boot--reduced': reduced }">
    <div class="gk-boot-glow" />
    <div class="gk-boot-mark">
      <Logo :size="72" show-text :text-size="26" tone="light" />
    </div>
  </div>
</template>

<style scoped>
.gk-boot {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #07080f;
  pointer-events: none;
  animation: gk-boot-out 0.4s ease forwards;
  animation-delay: 1s;
}

.gk-boot-glow {
  position: absolute;
  width: 320px;
  height: 320px;
  border-radius: 50%;
  background: radial-gradient(
    circle,
    rgba(99, 102, 241, 0.35) 0%,
    rgba(59, 130, 246, 0.12) 45%,
    rgba(59, 130, 246, 0) 70%
  );
  filter: blur(8px);
  opacity: 0;
  animation: gk-boot-bloom 0.8s ease forwards;
}

.gk-boot-mark {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  opacity: 0;
  transform: scale(0.8) rotate(-50deg);
  animation: gk-boot-in 0.7s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
}

/* Stagger the wordmark so it lands just after the helm locks. */
.gk-boot-mark :deep(.gk-logo-text) {
  opacity: 0;
  transform: translateY(6px);
  animation: gk-boot-text 0.5s ease forwards;
  animation-delay: 0.35s;
}

@keyframes gk-boot-in {
  to {
    opacity: 1;
    transform: scale(1) rotate(0deg);
  }
}

@keyframes gk-boot-bloom {
  to {
    opacity: 1;
    transform: scale(1.4);
  }
}

@keyframes gk-boot-text {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes gk-boot-out {
  to {
    opacity: 0;
    transform: scale(1.03);
  }
}

/* Reduced motion: drop the spin, keep a gentle fade in/out. */
.gk-boot--reduced {
  animation-duration: 0.4s;
  animation-delay: 0.5s;
}
.gk-boot--reduced .gk-boot-mark {
  transform: scale(1);
  animation: gk-boot-fade 0.4s ease forwards;
}
.gk-boot--reduced .gk-boot-glow {
  animation: gk-boot-fade 0.4s ease forwards;
}
.gk-boot--reduced .gk-boot-mark :deep(.gk-logo-text) {
  transform: none;
  animation-delay: 0.1s;
}

@keyframes gk-boot-fade {
  to {
    opacity: 1;
  }
}
</style>

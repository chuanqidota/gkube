<script setup lang="ts">
import { useId, computed } from 'vue'

const props = withDefaults(
  defineProps<{
    /** Hexagon mark edge length in px. */
    size?: number
    /** Show the "GKube" wordmark next to the mark. */
    showText?: boolean
    /** Wordmark font-size in px. */
    textSize?: number
    /** Wordmark tone — light for dark backgrounds, dark for light backgrounds. */
    tone?: 'light' | 'dark'
  }>(),
  {
    size: 32,
    showText: false,
    textSize: 18,
    tone: 'light',
  },
)

// Unique gradient ids so multiple Logo instances never collide.
const uid = useId()
const hexId = computed(() => `gk-logo-hex-${uid}`)
const innerId = computed(() => `gk-logo-inner-${uid}`)
</script>

<template>
  <div class="gk-logo" :style="{ '--gk-logo-size': `${props.size}px` }">
    <svg
      :viewBox="'0 0 48 48'"
      :width="props.size"
      :height="props.size"
      class="gk-logo-mark"
      aria-hidden="true"
      focusable="false"
    >
      <defs>
        <linearGradient :id="hexId" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#6366f1" />
          <stop offset="50%" stop-color="#818cf8" />
          <stop offset="100%" stop-color="#3b82f6" />
        </linearGradient>
        <linearGradient :id="innerId" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#c7d2fe" />
          <stop offset="100%" stop-color="#e0e7ff" />
        </linearGradient>
      </defs>
      <path d="M24 2 L44 14 L44 34 L24 46 L4 34 L4 14 Z" :fill="`url(#${hexId})`" />
      <path d="M24 8 L38 16 L38 32 L24 40 L10 32 L10 16 Z" :fill="`url(#${innerId})`" opacity="0.15" />
      <path
        d="M30 16 C26.5 13.5 21.5 13.5 18 16 C14.5 18.5 14 23 14 24 C14 28 16 31 20 33 C23 34.5 27 34.5 30 33 L30 26 L24 26 L24 23 L30 23 Z"
        fill="white"
        opacity="0.95"
      />
      <circle cx="36" cy="12" r="3" fill="#a78bfa" opacity="0.8" />
    </svg>
    <transition name="gk-logo-fade">
      <span
        v-show="showText"
        class="gk-logo-text"
        :class="`gk-logo-text--${tone}`"
        :style="{ fontSize: `${props.textSize}px` }"
      >GKube</span>
    </transition>
  </div>
</template>

<style scoped>
.gk-logo {
  display: inline-flex;
  align-items: center;
  gap: var(--gk-space-2);
}

.gk-logo-mark {
  display: block;
  flex-shrink: 0;
  filter: drop-shadow(0 2px 8px rgba(59, 130, 246, 0.18));
}

.gk-logo-text {
  font-family: var(--gk-font-sans);
  font-weight: 700;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.gk-logo-text--light {
  color: #f1f5f9;
}

.gk-logo-text--dark {
  color: var(--gk-neutral-900);
}

.gk-logo-fade-enter-active,
.gk-logo-fade-leave-active {
  transition: opacity var(--gk-transition-fast);
}

.gk-logo-fade-enter-from,
.gk-logo-fade-leave-to {
  opacity: 0;
}
</style>

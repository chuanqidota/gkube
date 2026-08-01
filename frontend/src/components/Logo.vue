<script setup lang="ts">
import { useId, computed } from 'vue'

const props = withDefaults(
  defineProps<{
    /** Mark edge length in px. */
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
const ringId = computed(() => `gk-logo-ring-${uid}`)
const gemId = computed(() => `gk-logo-gem-${uid}`)
const spokeId = computed(() => `gk-logo-spoke-${uid}`)
const glowId = computed(() => `gk-logo-glow-${uid}`)
const softId = computed(() => `gk-logo-soft-${uid}`)

// Below this size the spokes and fine detail smear — switch to the bold hub tier.
const detailed = computed(() => props.size >= 24)
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
        <linearGradient :id="ringId" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#818cf8" />
          <stop offset="45%" stop-color="#4f46e5" />
          <stop offset="100%" stop-color="#1e1b4b" />
        </linearGradient>
        <linearGradient :id="gemId" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#e0e7ff" />
          <stop offset="35%" stop-color="#818cf8" />
          <stop offset="75%" stop-color="#4f46e5" />
          <stop offset="100%" stop-color="#312e81" />
        </linearGradient>
        <linearGradient :id="spokeId" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#93c5fd" />
          <stop offset="100%" stop-color="#3b82f6" />
        </linearGradient>
        <radialGradient :id="glowId" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stop-color="#6366f1" stop-opacity="0.55" />
          <stop offset="60%" stop-color="#3b82f6" stop-opacity="0.18" />
          <stop offset="100%" stop-color="#3b82f6" stop-opacity="0" />
        </radialGradient>
        <filter :id="softId" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="0.9" />
        </filter>
      </defs>

      <!-- ambient glow -->
      <circle cx="24" cy="24" r="22" :fill="`url(#${glowId})`" />

      <!-- outer ring -->
      <path
        d="M45 24 A21 21 0 1 1 3 24 A21 21 0 1 1 45 24 Z M42 24 A18 18 0 1 0 6 24 A18 18 0 1 0 42 24 Z"
        :fill="`url(#${ringId})`"
        fill-rule="evenodd"
      />
      <!-- top-left glass highlight -->
      <path
        d="M5.5 17.3 A19.7 19.7 0 0 1 42.5 17.3"
        fill="none"
        stroke="#ffffff"
        stroke-width="1.8"
        stroke-linecap="round"
        opacity="0.4"
        :filter="`url(#${softId})`"
      />
      <!-- outer rim -->
      <circle cx="24" cy="24" r="21" fill="none" stroke="#a78bfa" stroke-width="0.8" opacity="0.75" />
      <!-- inner edge shadow -->
      <circle cx="24" cy="24" r="18" fill="none" stroke="#0b0a1f" stroke-width="0.6" opacity="0.5" />

      <!-- 6 spokes (hex-vertex aligned) — only at full size -->
      <g
        v-if="detailed"
        :stroke="`url(#${spokeId})`"
        stroke-width="1.5"
        stroke-linecap="round"
        opacity="0.92"
      >
        <line x1="24" y1="15" x2="24" y2="6" />
        <line x1="31.79" y1="19.5" x2="39.59" y2="15" />
        <line x1="31.79" y1="28.5" x2="39.59" y2="33" />
        <line x1="24" y1="33" x2="24" y2="42" />
        <line x1="16.21" y1="28.5" x2="8.41" y2="33" />
        <line x1="16.21" y1="19.5" x2="8.41" y2="15" />
      </g>

      <!-- G gem hub -->
      <circle cx="24" cy="24" :r="detailed ? 8.2 : 9.2" :fill="`url(#${gemId})`" />
      <circle
        cx="24"
        cy="24"
        :r="detailed ? 8.2 : 9.2"
        fill="none"
        stroke="#a78bfa"
        stroke-width="0.7"
        opacity="0.6"
      />
      <!-- the G — bold open bowl (right opening) + inward crossbar -->
      <path
        :d="detailed
          ? 'M28.73 20.31 A6 6 0 1 0 28.73 27.69 M29.4 24 L23 24'
          : 'M29.04 20.06 A6.4 6.4 0 1 0 29.04 27.94 M29.7 24 L22.6 24'"
        fill="none"
        stroke="#e0e7ff"
        :stroke-width="detailed ? 3 : 3.2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
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
  filter: drop-shadow(0 4px 14px rgba(59, 130, 246, 0.35));
}

.gk-logo-text {
  font-family: var(--gk-font-sans);
  font-weight: 700;
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.gk-logo-text--light {
  background: linear-gradient(135deg, #c7d2fe 0%, #818cf8 45%, #3b82f6 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  -webkit-text-fill-color: transparent;
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

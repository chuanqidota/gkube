<script setup lang="ts">
import { ref } from 'vue'
import BootSplash from '@/components/BootSplash.vue'

const showSplash = ref(true)
</script>

<template>
  <BootSplash v-if="showSplash" @done="showSplash = false" />
  <router-view />
</template>

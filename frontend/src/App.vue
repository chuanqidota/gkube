<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import BootSplash from '@/components/BootSplash.vue'

const { locale } = useI18n()
const elLocale = computed(() => locale.value === 'zh-CN' ? zhCn : en)
const showSplash = ref(true)
</script>

<template>
  <ElConfigProvider :locale="elLocale">
    <BootSplash v-if="showSplash" @done="showSplash = false" />
    <router-view />
  </ElConfigProvider>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  /** Description text; falls back to the localized "no data" string */
  description?: string
  /** Image size in px */
  imageSize?: number
}>()

const { t } = useI18n()
</script>

<template>
  <el-empty
    :description="description || t('common.noData')"
    :image-size="imageSize || 80"
  >
    <slot />
  </el-empty>
</template>

<script setup lang="ts">
import WorkloadForm from './WorkloadForm.vue'

defineProps<{
  isEdit?: boolean
  initialData?: any
  onSubmit?: (yaml: string) => Promise<void>
}>()

defineEmits<{ success: []; cancel: [] }>()
</script>

<template>
  <WorkloadForm kind="DaemonSet" :is-edit="isEdit" :initial-data="initialData" :on-submit="onSubmit" />
</template>

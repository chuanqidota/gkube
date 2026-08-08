<script setup lang="ts">
import WorkloadCreate from './WorkloadCreate.vue'
import { createStatefulSet } from '@/api/resource'
</script>

<template>
  <WorkloadCreate
    kind="StatefulSet"
    :create-api="createStatefulSet"
    list-route="/workloads/statefulsets"
  />
</template>

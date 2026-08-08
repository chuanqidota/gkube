<script setup lang="ts">
import WorkloadCreate from './WorkloadCreate.vue'
import { createDaemonSet } from '@/api/resource'
</script>

<template>
  <WorkloadCreate
    kind="DaemonSet"
    :create-api="createDaemonSet"
    list-route="/workloads/daemonsets"
  />
</template>

<script setup lang="ts">
import WorkloadCreate from './WorkloadCreate.vue'
import { createDeployment } from '@/api/resource'
</script>

<template>
  <WorkloadCreate
    kind="Deployment"
    :create-api="createDeployment"
    list-route="/workloads/deployments"
  />
</template>

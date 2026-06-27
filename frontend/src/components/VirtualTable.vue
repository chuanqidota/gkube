<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useVirtualizer } from '@tanstack/vue-virtual'

const props = defineProps<{
  data: any[]
  rowHeight?: number
  overscan?: number
}>()

const rowHeight = props.rowHeight || 48
const overscan = props.overscan || 10

const parentRef = ref<HTMLElement | null>(null)

const virtualizer = useVirtualizer({
  count: props.data.length,
  getScrollElement: () => parentRef.value,
  estimateSize: () => rowHeight,
  overscan,
})

const totalSize = computed(() => virtualizer.value.getTotalSize())
const virtualRows = computed(() => virtualizer.value.getVirtualItems())

watch(() => props.data.length, () => {
  virtualizer.value.measure()
})
</script>

<template>
  <div
    ref="parentRef"
    class="virtual-table-container"
    :style="{ overflow: 'auto', maxHeight: 'calc(100vh - 280px)' }"
  >
    <div :style="{ height: `${totalSize}px`, position: 'relative' }">
      <div
        v-for="row in virtualRows"
        :key="row.index"
        :style="{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: `${row.size}px`,
          transform: `translateY(${row.start}px)`,
        }"
        class="virtual-row"
      >
        <slot name="row" :item="data[row.index]" :index="row.index" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.virtual-table-container {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
}
.virtual-row {
  display: flex;
  align-items: center;
  padding: 0 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  transition: background-color 0.2s;
}
.virtual-row:hover {
  background-color: var(--el-fill-color-light);
}
</style>

<script setup lang="ts">
import { useResizable } from '@/composables/useResizable'

interface Props {
  resizable?: boolean
  initialLeftWidth?: number
  initialTopHeight?: number
  minLeftWidth?: number
  minTopHeight?: number
}

const props = withDefaults(defineProps<Props>(), {
  resizable: false,
  initialLeftWidth: 400,
  initialTopHeight: 300,
  minLeftWidth: 300,
  minTopHeight: 150,
})

const { leftWidth, rightTopHeight, resizingH, resizingV, onHResizeStart, onVResizeStart } =
  useResizable({
    initialWidth: props.initialLeftWidth,
    minWidth: props.minLeftWidth,
    minTopHeight: props.minTopHeight,
  })
</script>

<template>
  <div class="detail-page">
    <!-- Default slot: page header content outside main-layout -->
    <slot />
    <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">
      <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
        <slot name="left" />
      </div>
      <div class="resize-handle-h" :class="{ active: resizingH }" @mousedown="onHResizeStart" />
      <div class="right-panel">
        <template v-if="resizable">
          <div class="right-top" :style="{ height: (rightTopHeight ?? initialTopHeight) + 'px' }">
            <slot name="right-top" />
          </div>
          <div class="resize-handle-v" :class="{ active: resizingV }" @mousedown="onVResizeStart" />
          <div class="right-bottom">
            <slot name="right-bottom" />
          </div>
        </template>
        <template v-else>
          <slot name="right" />
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.detail-page {
  padding: var(--gk-space-4) var(--gk-space-5);
  height: calc(100dvh - var(--gk-header-height));
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

.main-layout {
  display: flex;
  gap: 2px;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  position: relative;
}

.left-panel {
  border: 1px solid var(--gk-color-border-light);
  border-radius: var(--gk-radius-md);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  overflow: hidden;
}

.right-top,
.right-bottom {
  border: 1px solid var(--gk-color-border-light);
  border-radius: var(--gk-radius-md);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.right-top {
  flex-shrink: 0;
}

.right-bottom {
  flex: 1;
  min-height: 0;
}

.resize-handle-h {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 8px;
  cursor: col-resize;
  z-index: 10;
}

.resize-handle-h:hover,
.resize-handle-h.active {
  background: var(--gk-color-primary-bg);
}

.resize-handle-v {
  height: 4px;
  cursor: row-resize;
  flex-shrink: 0;
  position: relative;
  z-index: 5;
  margin: -2px 0;
}

.resize-handle-v:hover,
.resize-handle-v.active {
  background: var(--gk-color-primary-bg);
}

.is-resizing {
  user-select: none;
}

.is-resizing * {
  pointer-events: none;
}

@media (max-width: 768px) {
  .main-layout {
    flex-direction: column;
    overflow: auto;
  }
  .left-panel {
    width: 100% !important;
    min-width: 100% !important;
    max-height: 300px;
  }
  .resize-handle-h {
    display: none;
  }
}
</style>

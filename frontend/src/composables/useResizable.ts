import { ref } from 'vue'

export interface ResizableOptions {
  /** 左侧面板初始宽度（px） */
  initialWidth?: number
  /** 左侧面板最小/最大宽度（px） */
  minWidth?: number
  maxWidth?: number
  /** 上下分栏时顶部区域的最小高度（px） */
  minTopHeight?: number
  /** 上下分栏时底部区域保留的最小高度（px） */
  minBottomHeight?: number
  /** 垂直分栏所参考的父面板选择器 */
  panelSelector?: string
}

/**
 * 详情页通用分栏拖拽逻辑：左侧宽度（左右拖拽）+ 右侧上下分栏高度（上下拖拽）。
 *
 * 关键修复：垂直拖拽时先确认 rightPanel 可查到再置 resizingV=true，
 * 否则 early-return 会导致 resizingV 永真、整页 pointer-events:none 卡死。
 */
export function useResizable(options: ResizableOptions = {}) {
  const {
    initialWidth = 320,
    minWidth = 220,
    maxWidth = 500,
    minTopHeight = 120,
    minBottomHeight = 120,
    panelSelector = '.right-panel',
  } = options

  const leftWidth = ref(initialWidth)
  const rightTopHeight = ref<number | null>(null)
  const resizingH = ref(false)
  const resizingV = ref(false)

  function onHResizeStart(e: MouseEvent) {
    e.preventDefault()
    resizingH.value = true
    const startX = e.clientX
    const startW = leftWidth.value
    const onMove = (ev: MouseEvent) => {
      leftWidth.value = Math.min(Math.max(startW + ev.clientX - startX, minWidth), maxWidth)
    }
    const onUp = () => {
      resizingH.value = false
      document.removeEventListener('mousemove', onMove)
      document.removeEventListener('mouseup', onUp)
    }
    document.addEventListener('mousemove', onMove)
    document.addEventListener('mouseup', onUp)
  }

  function onVResizeStart(e: MouseEvent) {
    e.preventDefault()
    const rightPanel = (e.target as HTMLElement).closest(panelSelector)
    // 先校验面板存在再进入 resizing 态，避免 early-return 后 resizingV 永真卡死
    if (!rightPanel) return
    resizingV.value = true
    const startY = e.clientY
    const startH = rightPanel.getBoundingClientRect().height
    const onMove = (ev: MouseEvent) => {
      const delta = ev.clientY - startY
      rightTopHeight.value = Math.min(Math.max(startH * 0.3 + delta, minTopHeight), startH - minBottomHeight)
    }
    const onUp = () => {
      resizingV.value = false
      document.removeEventListener('mousemove', onMove)
      document.removeEventListener('mouseup', onUp)
    }
    document.addEventListener('mousemove', onMove)
    document.addEventListener('mouseup', onUp)
  }

  return { leftWidth, rightTopHeight, resizingH, resizingV, onHResizeStart, onVResizeStart }
}

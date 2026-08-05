import { onBeforeUnmount, onMounted, reactive, type Ref } from 'vue'

/**
 * Ambient "control plane" canvas for the login page.
 *
 * Hand-rolled canvas so we control the premium rendering: glowing hexagon
 * cluster hubs, dot workers, curved gradient edges, and particles travelling
 * along edges (real "live traffic" motion).
 *
 * It is driven by `graphState` (a reactive object the host view mutates):
 *  - activeField: 'username' | 'password' | null  -> lights up a node cluster
 *  - typing: { field, at } last keystroke marker    -> spawns a particle burst
 *  - status: 'idle' | 'loading' | 'error'           -> global pulse / red wash
 *
 * The rAF loop READS graphState but never writes to it; all heavy mutation
 * happens on plain closure arrays (node hi lerp, edge particles, burst pool).
 * Pauses when the tab is hidden. Under prefers-reduced-motion the graph is a
 * static frame (no travelling particles, no breathing) but still reflects
 * highlight + status tint so the left-right link survives.
 *
 * Coordinates are normalized [0..1] so the layout adapts to any resize.
 */

export type GraphStatus = 'idle' | 'loading' | 'error'
export type GraphField = 'username' | 'password' | null

export const graphState = reactive({
  activeField: null as GraphField,
  /** bumped on every keystroke; the loop reads .at to detect new bursts */
  typing: { field: null as GraphField, at: 0 },
  status: 'idle' as GraphStatus,
  /** monotonic counter the view bumps to request a burst (e.g. submit click) */
  burstTick: 0,
})

type Node = {
  id: string
  kind: 'cluster' | 'worker'
  x: number
  y: number
  r: number
  hi: number
  hiTarget: number
  phase: number
  /** logical zone so an active field can light up a region */
  zone: 'username' | 'password' | 'shared'
}

type Edge = {
  from: string
  to: string
  cx: number
  cy: number
  hi: number
  hiTarget: number
  zone: 'username' | 'password' | 'shared'
  particles: { t: number; speed: number }[]
}

type Vec = { x: number; y: number }

// Burst particle: travels outward from a node, fades, dies.
type Burst = { x: number; y: number; vx: number; vy: number; life: number; max: number; zone: Node['zone'] }

const PALETTE = {
  cluster: '#3b82f6',
  clusterGlow: '#60a5fa',
  worker: '#1e293b',
  workerBorder: '#475569',
  workerHi: '#60a5fa',
  edgeDim: 'rgba(148, 163, 184, 0.10)',
  particle: '#93c5fd',
  error: '#f87171',
  errorGlow: '#ef4444',
}

// Layout: cluster c1 lights for username, c3 for password, c2 shared; workers
// are partitioned to the same zones so a field highlights its neighbourhood.
const NODES: Node[] = [
  { id: 'c1', kind: 'cluster', x: 0.24, y: 0.46, r: 11, hi: 0, hiTarget: 0, phase: 0.0, zone: 'username' },
  { id: 'c2', kind: 'cluster', x: 0.52, y: 0.26, r: 11, hi: 0, hiTarget: 0, phase: 1.7, zone: 'shared' },
  { id: 'c3', kind: 'cluster', x: 0.80, y: 0.52, r: 11, hi: 0, hiTarget: 0, phase: 3.1, zone: 'password' },

  { id: 'w1', kind: 'worker', x: 0.10, y: 0.30, r: 4, hi: 0, hiTarget: 0, phase: 0.3, zone: 'username' },
  { id: 'w2', kind: 'worker', x: 0.12, y: 0.66, r: 4, hi: 0, hiTarget: 0, phase: 2.1, zone: 'username' },
  { id: 'w3', kind: 'worker', x: 0.30, y: 0.68, r: 4, hi: 0, hiTarget: 0, phase: 4.0, zone: 'username' },
  { id: 'w4', kind: 'worker', x: 0.34, y: 0.30, r: 4, hi: 0, hiTarget: 0, phase: 5.2, zone: 'shared' },

  { id: 'w5', kind: 'worker', x: 0.44, y: 0.10, r: 4, hi: 0, hiTarget: 0, phase: 1.1, zone: 'shared' },
  { id: 'w6', kind: 'worker', x: 0.62, y: 0.12, r: 4, hi: 0, hiTarget: 0, phase: 3.3, zone: 'shared' },
  { id: 'w7', kind: 'worker', x: 0.66, y: 0.40, r: 4, hi: 0, hiTarget: 0, phase: 0.7, zone: 'password' },

  { id: 'w8', kind: 'worker', x: 0.92, y: 0.34, r: 4, hi: 0, hiTarget: 0, phase: 2.8, zone: 'password' },
  { id: 'w9', kind: 'worker', x: 0.90, y: 0.74, r: 4, hi: 0, hiTarget: 0, phase: 4.6, zone: 'password' },
  { id: 'w10', kind: 'worker', x: 0.70, y: 0.74, r: 4, hi: 0, hiTarget: 0, phase: 1.4, zone: 'password' },
  { id: 'w11', kind: 'worker', x: 0.50, y: 0.78, r: 4, hi: 0, hiTarget: 0, phase: 3.8, zone: 'shared' },
]

function makeEdge(from: string, to: string): Edge {
  const a = NODES.find((n) => n.id === from)!
  const b = NODES.find((n) => n.id === to)!
  const mx = (a.x + b.x) / 2
  const my = (a.y + b.y) / 2
  const dx = b.x - a.x
  const dy = b.y - a.y
  const len = Math.hypot(dx, dy) || 1
  const nx = -dy / len
  const ny = dx / len
  const bend = 0.04
  const zone: Edge['zone'] = a.zone === b.zone ? a.zone : 'shared'
  return {
    from,
    to,
    cx: mx + nx * bend,
    cy: my + ny * bend,
    hi: 0,
    hiTarget: 0,
    zone,
    particles: [
      { t: Math.random(), speed: 0.06 + Math.random() * 0.05 },
      { t: Math.random(), speed: 0.06 + Math.random() * 0.05 },
    ],
  }
}

const EDGES: Edge[] = [
  makeEdge('c1', 'w1'), makeEdge('c1', 'w2'), makeEdge('c1', 'w3'), makeEdge('c1', 'w4'),
  makeEdge('c2', 'w4'), makeEdge('c2', 'w5'), makeEdge('c2', 'w6'), makeEdge('c2', 'w7'),
  makeEdge('c3', 'w7'), makeEdge('c3', 'w8'), makeEdge('c3', 'w9'), makeEdge('c3', 'w10'),
  makeEdge('c2', 'w11'), makeEdge('c3', 'w11'),
]

const nodeById = new Map(NODES.map((n) => [n.id, n]))
const adjacency = new Map<string, Set<string>>()
for (const e of EDGES) {
  if (!adjacency.has(e.from)) adjacency.set(e.from, new Set())
  if (!adjacency.has(e.to)) adjacency.set(e.to, new Set())
  adjacency.get(e.from)!.add(e.to)
  adjacency.get(e.to)!.add(e.from)
}

function bezierPoint(a: Vec, b: Vec, c: Vec, t: number): Vec {
  const u = 1 - t
  return {
    x: u * u * a.x + 2 * u * t * b.x + t * t * c.x,
    y: u * u * a.y + 2 * u * t * b.y + t * t * c.y,
  }
}

function hexagonPath(ctx: CanvasRenderingContext2D, cx: number, cy: number, r: number) {
  ctx.beginPath()
  for (let i = 0; i < 6; i++) {
    const ang = (Math.PI / 3) * i - Math.PI / 6
    const px = cx + r * Math.cos(ang)
    const py = cy + r * Math.sin(ang)
    if (i === 0) ctx.moveTo(px, py)
    else ctx.lineTo(px, py)
  }
  ctx.closePath()
}

export function useClusterGraph(canvasRef: Ref<HTMLCanvasElement | undefined>) {
  let ctx: CanvasRenderingContext2D | null = null
  let rafId = 0
  let width = 0
  let height = 0
  let dpr = 1
  let running = false
  let lastTs = 0
  let lastNearest: string | null = null
  let lastTypingAt = 0
  let lastBurstTick = 0
  let errorWash = 0 // 0..1 decays after an error

  let pointerHandler: ((e: PointerEvent) => void) | null = null
  let leaveHandler: (() => void) | null = null
  let resizeObserver: ResizeObserver | null = null
  let visibilityHandler: (() => void) | null = null

  const bursts: Burst[] = []

  const prefersReduced =
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches

  function toPx(n: Node): Vec {
    return { x: n.x * width, y: n.y * height }
  }

  function resize() {
    const canvas = canvasRef.value
    if (!canvas || !ctx) return
    const rect = canvas.getBoundingClientRect()
    dpr = window.devicePixelRatio || 1
    width = Math.max(1, rect.width)
    height = Math.max(1, rect.height)
    canvas.width = Math.round(width * dpr)
    canvas.height = Math.round(height * dpr)
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  }

  /** Highlight driven by field focus. Returns the zone currently lit, if any. */
  function applyFieldHighlight(): 'username' | 'password' | null {
    const f = graphState.activeField
    if (f !== 'username' && f !== 'password') return null
    for (const n of NODES) {
      if (n.zone === f) n.hiTarget = 1
      else if (n.zone === 'shared') n.hiTarget = 0.55
    }
    for (const e of EDGES) {
      if (e.zone === f) e.hiTarget = 1
      else if (e.zone === 'shared') e.hiTarget = 0.45
    }
    return f
  }

  function setHoverHighlight(nearest: string | null): boolean {
    // Hover highlight layers ON TOP of the field highlight; when hovering, it
    // also dims non-hover nodes so the cursor region reads clearly.
    if (nearest === lastNearest) return false
    lastNearest = nearest

    const fieldZone = applyFieldHighlight()
    const hoverSet = new Set<string>()
    if (nearest) {
      hoverSet.add(nearest)
      adjacency.get(nearest)?.forEach((id) => hoverSet.add(id))
    }

    for (const n of NODES) {
      let target = 0
      if (fieldZone && n.zone === fieldZone) target = 1
      else if (fieldZone && n.zone === 'shared') target = 0.55
      if (hoverSet.has(n.id)) target = 1
      n.hiTarget = target
    }
    for (const e of EDGES) {
      let target = 0
      if (fieldZone && e.zone === fieldZone) target = 1
      else if (fieldZone && e.zone === 'shared') target = 0.45
      if (hoverSet.has(e.from) && hoverSet.has(e.to)) target = 1
      e.hiTarget = target
    }
    return true
  }

  function onPointerMove(e: PointerEvent) {
    const canvas = canvasRef.value
    if (!canvas) return
    const rect = canvas.getBoundingClientRect()
    const px = e.clientX - rect.left
    const py = e.clientY - rect.top

    let nearest: string | null = null
    let nearestDist = 70 * 70
    for (const n of NODES) {
      const p = toPx(n)
      const d = (p.x - px) ** 2 + (p.y - py) ** 2
      if (d < nearestDist) {
        nearestDist = d
        nearest = n.id
      }
    }
    if (setHoverHighlight(nearest) && prefersReduced) renderStatic()
  }

  function onPointerLeave() {
    if (setHoverHighlight(null) && prefersReduced) renderStatic()
  }

  /** Spawn a burst across every node in the zone so the whole region erupts,
   *  not just a single hub. Bigger, faster, longer-lived than a single-origin burst. */
  function spawnBurst(zone: Node['zone'], perNode = 10) {
    const nodes = NODES.filter((n) => n.zone === zone || n.zone === 'shared')
    for (const origin of nodes) {
      const p = toPx(origin)
      for (let i = 0; i < perNode; i++) {
        const ang = Math.random() * Math.PI * 2
        const sp = 70 + Math.random() * 160
        bursts.push({
          x: p.x,
          y: p.y,
          vx: Math.cos(ang) * sp,
          vy: Math.sin(ang) * sp,
          life: 0,
          max: 0.8 + Math.random() * 0.7,
          zone,
        })
      }
    }
    // cap the pool
    if (bursts.length > 400) bursts.splice(0, bursts.length - 400)
  }

  function step(ts: number) {
    if (!ctx) return
    const dt = lastTs ? Math.min(50, ts - lastTs) / 1000 : 0
    lastTs = ts

    // React to view-driven state (read-only on the reactive object).
    if (graphState.typing.at !== lastTypingAt) {
      lastTypingAt = graphState.typing.at
      const z = graphState.typing.field
      if (z === 'username' || z === 'password') spawnBurst(z)
    }
    if (graphState.burstTick !== lastBurstTick) {
      lastBurstTick = graphState.burstTick
      spawnBurst('username')
      spawnBurst('password')
    }
    if (graphState.status === 'error') {
      errorWash = 1
      graphState.status = 'idle' // one-shot
    }
    errorWash = Math.max(0, errorWash - dt * 1.4)

    // Apply field highlight as soon as activeField changes (focus). Mouse hover
    // will layer on top later; we let it override only when actually hovered.
    if (graphState.activeField) applyFieldHighlight()

    // If no field is active and no hover, ensure targets reset.
    if (!graphState.activeField && lastNearest === null) {
      for (const n of NODES) if (n.hiTarget !== 0) n.hiTarget = 0
      for (const e of EDGES) if (e.hiTarget !== 0) e.hiTarget = 0
    }

    ctx.clearRect(0, 0, width, height)

    // error wash overlay
    if (errorWash > 0.01) {
      ctx.fillStyle = `rgba(239, 68, 68, ${(errorWash * 0.12).toFixed(3)})`
      ctx.fillRect(0, 0, width, height)
    }

    const loadingPulse = graphState.status === 'loading'
    const loadingAmp = loadingPulse ? 0.5 + 0.5 * Math.sin(ts / 260) : 0

    for (const n of NODES) n.hi += (n.hiTarget - n.hi) * 0.15
    for (const e of EDGES) e.hi += (e.hiTarget - e.hi) * 0.15

    // edges + travelling particles
    for (const e of EDGES) {
      const a = toPx(nodeById.get(e.from)!)
      const b = toPx(nodeById.get(e.to)!)
      const c = { x: e.cx * width, y: e.cy * height }
      const hi = Math.min(1, e.hi + loadingAmp * 0.4)
      ctx.lineWidth = 1 + hi * 1.2
      ctx.strokeStyle = errorWash > 0.05
        ? `rgba(248, 113, 113, ${(0.18 + hi * 0.6).toFixed(3)})`
        : hi > 0.01
          ? `rgba(96, 165, 250, ${(0.18 + hi * 0.7).toFixed(3)})`
          : PALETTE.edgeDim
      ctx.beginPath()
      ctx.moveTo(a.x, a.y)
      ctx.quadraticCurveTo(c.x, c.y, b.x, b.y)
      ctx.stroke()

      if (!prefersReduced) {
        for (const p of e.particles) {
          p.t += p.speed * dt * (1 + hi * 2.2 + loadingAmp * 1.5)
          if (p.t > 1) p.t -= 1
          const pos = bezierPoint(a, c, b, p.t)
          const pr = 1.4 + hi * 1.4
          ctx.shadowBlur = 8 + hi * 10
          ctx.shadowColor = errorWash > 0.05 ? PALETTE.error : PALETTE.particle
          ctx.fillStyle = errorWash > 0.05
            ? `rgba(252, 165, 165, ${(0.45 + hi * 0.5).toFixed(3)})`
            : `rgba(147, 197, 253, ${(0.45 + hi * 0.5).toFixed(3)})`
          ctx.beginPath()
          ctx.arc(pos.x, pos.y, pr, 0, Math.PI * 2)
          ctx.fill()
        }
        ctx.shadowBlur = 0
      }
    }

    // burst particles
    if (!prefersReduced) {
      for (let i = bursts.length - 1; i >= 0; i--) {
        const b = bursts[i]
        b.life += dt
        if (b.life >= b.max) {
          bursts.splice(i, 1)
          continue
        }
        b.x += b.vx * dt
        b.y += b.vy * dt
        b.vx *= 0.94
        b.vy *= 0.94
        const k = 1 - b.life / b.max
        ctx.shadowBlur = 14
        ctx.shadowColor = errorWash > 0.05 ? PALETTE.error : PALETTE.particle
        ctx.fillStyle = errorWash > 0.05
          ? `rgba(252, 165, 165, ${k.toFixed(3)})`
          : `rgba(147, 197, 253, ${(0.5 + k * 0.5).toFixed(3)})`
        ctx.beginPath()
        ctx.arc(b.x, b.y, 2 + k * 2.4, 0, Math.PI * 2)
        ctx.fill()
      }
      ctx.shadowBlur = 0
    }

    // nodes
    for (const n of NODES) {
      const p = toPx(n)
      const hi = Math.min(1, n.hi + loadingAmp * 0.3)
      const breath = prefersReduced ? 0 : Math.sin(ts / 1400 + n.phase) * 0.5 + 0.5
      const isError = errorWash > 0.05

      if (n.kind === 'cluster') {
        const r = n.r + hi * 3 + breath * 1.2
        ctx.shadowBlur = 18 + hi * 22
        ctx.shadowColor = isError ? PALETTE.errorGlow : PALETTE.clusterGlow
        hexagonPath(ctx, p.x, p.y, r)
        const g = ctx.createLinearGradient(p.x - r, p.y - r, p.x + r, p.y + r)
        if (isError) {
          g.addColorStop(0, '#ef4444')
          g.addColorStop(0.5, '#f87171')
          g.addColorStop(1, '#dc2626')
        } else {
          g.addColorStop(0, '#6366f1')
          g.addColorStop(0.5, '#818cf8')
          g.addColorStop(1, PALETTE.cluster)
        }
        ctx.fillStyle = g
        ctx.fill()
        ctx.shadowBlur = 0
        hexagonPath(ctx, p.x, p.y, r * 0.5)
        ctx.fillStyle = 'rgba(255,255,255,0.92)'
        ctx.fill()
      } else {
        const r = n.r + hi * 1.5
        ctx.shadowBlur = 6 + hi * 14
        ctx.shadowColor = isError ? PALETTE.error : PALETTE.workerHi
        ctx.beginPath()
        ctx.arc(p.x, p.y, r, 0, Math.PI * 2)
        ctx.fillStyle = hi > 0.01
          ? (isError ? PALETTE.error : PALETTE.workerHi)
          : PALETTE.worker
        ctx.fill()
        ctx.shadowBlur = 0
        ctx.lineWidth = 1
        ctx.strokeStyle = hi > 0.01
          ? (isError ? 'rgba(254,202,202,0.9)' : 'rgba(191,219,254,0.9)')
          : PALETTE.workerBorder
        ctx.stroke()
      }
    }
    ctx.shadowBlur = 0

    if (running) rafId = requestAnimationFrame(step)
  }

  function renderStatic() {
    if (!ctx) return
    // Reflect current graphState highlight + error wash in one static frame.
    applyFieldHighlight()
    if (!graphState.activeField && lastNearest === null) {
      for (const n of NODES) n.hiTarget = 0
      for (const e of EDGES) e.hiTarget = 0
    }
    for (const n of NODES) n.hi = n.hiTarget
    for (const e of EDGES) e.hi = e.hiTarget

    ctx.clearRect(0, 0, width, height)
    if (errorWash > 0.01) {
      ctx.fillStyle = `rgba(239, 68, 68, ${(errorWash * 0.12).toFixed(3)})`
      ctx.fillRect(0, 0, width, height)
    }
    for (const e of EDGES) {
      const a = toPx(nodeById.get(e.from)!)
      const b = toPx(nodeById.get(e.to)!)
      const c = { x: e.cx * width, y: e.cy * height }
      const hi = e.hi
      ctx.lineWidth = 1 + hi * 1.2
      ctx.strokeStyle = hi > 0.01
        ? `rgba(96, 165, 250, ${(0.18 + hi * 0.7).toFixed(3)})`
        : PALETTE.edgeDim
      ctx.beginPath()
      ctx.moveTo(a.x, a.y)
      ctx.quadraticCurveTo(c.x, c.y, b.x, b.y)
      ctx.stroke()
    }
    for (const n of NODES) {
      const p = toPx(n)
      const hi = n.hi
      if (n.kind === 'cluster') {
        const r = n.r + hi * 3
        ctx.shadowBlur = 18 + hi * 22
        ctx.shadowColor = PALETTE.clusterGlow
        hexagonPath(ctx, p.x, p.y, r)
        const g = ctx.createLinearGradient(p.x - r, p.y - r, p.x + r, p.y + r)
        g.addColorStop(0, '#6366f1')
        g.addColorStop(0.5, '#818cf8')
        g.addColorStop(1, PALETTE.cluster)
        ctx.fillStyle = g
        ctx.fill()
        ctx.shadowBlur = 0
        hexagonPath(ctx, p.x, p.y, r * 0.5)
        ctx.fillStyle = 'rgba(255,255,255,0.92)'
        ctx.fill()
      } else {
        const r = n.r + hi * 1.5
        ctx.shadowBlur = 6 + hi * 14
        ctx.shadowColor = PALETTE.workerHi
        ctx.beginPath()
        ctx.arc(p.x, p.y, r, 0, Math.PI * 2)
        ctx.fillStyle = hi > 0.01 ? PALETTE.workerHi : PALETTE.worker
        ctx.fill()
        ctx.shadowBlur = 0
        ctx.lineWidth = 1
        ctx.strokeStyle = hi > 0.01 ? 'rgba(191,219,254,0.9)' : PALETTE.workerBorder
        ctx.stroke()
      }
    }
    ctx.shadowBlur = 0
  }

  function start() {
    if (running) return
    running = true
    lastTs = 0
    rafId = requestAnimationFrame(step)
  }

  function stop() {
    running = false
    if (rafId) cancelAnimationFrame(rafId)
    rafId = 0
  }

  function mount() {
    const canvas = canvasRef.value
    if (!canvas) return
    ctx = canvas.getContext('2d')
    if (!ctx) return

    resize()
    resizeObserver = new ResizeObserver(() => {
      resize()
      if (prefersReduced) renderStatic()
    })
    resizeObserver.observe(canvas)

    pointerHandler = onPointerMove
    leaveHandler = onPointerLeave
    canvas.addEventListener('pointermove', pointerHandler, { passive: true })
    canvas.addEventListener('pointerleave', leaveHandler)

    visibilityHandler = () => {
      if (document.hidden) stop()
      else if (!prefersReduced) start()
    }
    document.addEventListener('visibilitychange', visibilityHandler)

    if (prefersReduced) {
      renderStatic()
    } else {
      start()
    }
  }

  function destroy() {
    stop()
    const canvas = canvasRef.value
    if (canvas && pointerHandler) canvas.removeEventListener('pointermove', pointerHandler)
    if (canvas && leaveHandler) canvas.removeEventListener('pointerleave', leaveHandler)
    if (resizeObserver) resizeObserver.disconnect()
    if (visibilityHandler) document.removeEventListener('visibilitychange', visibilityHandler)
    ctx = null
  }

  onMounted(mount)
  onBeforeUnmount(destroy)
}

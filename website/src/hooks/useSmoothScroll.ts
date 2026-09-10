import { useEffect } from 'react'
import {
  readScrollMetrics,
  setSmoothScrollApi,
  type SmoothScrollApi,
} from '../lib/smooth-scroll-api'

/** Interpolation — smooth glide between wheel / touch steps. */
const LERP = 0.072
/** Wheel delta multiplier. */
const WHEEL_SCALE = 0.4

function maxScrollY() {
  return Math.max(0, document.documentElement.scrollHeight - window.innerHeight)
}

function startSmoothScroll(): () => void {
  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  const listeners = new Set<() => void>()

  function notify() {
    listeners.forEach((fn) => fn())
  }

  let target = window.scrollY
  let current = window.scrollY
  let frame = 0
  let ignoreScrollSync = false
  let touchStartY = 0
  let touching = false

  function metrics() {
    const max = maxScrollY()
    return {
      scrollY: current,
      maxScrollY: max,
      progress: max > 0 ? current / max : 0,
    }
  }

  function clampY(y: number) {
    return Math.min(maxScrollY(), Math.max(0, y))
  }

  function syncFromNative() {
    if (ignoreScrollSync) return
    target = window.scrollY
    current = window.scrollY
    notify()
  }

  function onWheel(event: WheelEvent) {
    if (event.ctrlKey) return
    const max = maxScrollY()
    if (max <= 0) return

    event.preventDefault()
    const delta =
      event.deltaMode === 1
        ? event.deltaY * 16
        : event.deltaMode === 2
          ? event.deltaY * window.innerHeight
          : event.deltaY
    target = clampY(target + delta * WHEEL_SCALE)
    notify()
  }

  function shouldSkipSmoothTouch(target: EventTarget | null) {
    if (!(target instanceof Element)) return false
    let node: Element | null = target
    while (node) {
      if (node.matches('input, textarea, select, [contenteditable="true"]')) return true
      const style = getComputedStyle(node)
      if (
        (style.overflowY === 'auto' || style.overflowY === 'scroll') &&
        node.scrollHeight > node.clientHeight + 1
      ) {
        return true
      }
      node = node.parentElement
    }
    return false
  }

  function onTouchStart(event: TouchEvent) {
    if (event.touches.length !== 1) return
    if (shouldSkipSmoothTouch(event.target)) {
      touching = false
      return
    }
    touching = true
    touchStartY = event.touches[0]?.clientY ?? 0
  }

  function onTouchMove(event: TouchEvent) {
    if (!touching || event.touches.length !== 1) return
    if (shouldSkipSmoothTouch(event.target)) return

    const y = event.touches[0]?.clientY ?? touchStartY
    const delta = touchStartY - y
    touchStartY = y
    if (Math.abs(delta) < 0.25) return

    event.preventDefault()
    target = clampY(target + delta)
    notify()
  }

  function onTouchEnd() {
    touching = false
  }

  function tick() {
    const delta = target - current
    if (Math.abs(delta) > 0.5) {
      current += delta * LERP
      ignoreScrollSync = true
      window.scrollTo(0, current)
      notify()
      requestAnimationFrame(() => {
        ignoreScrollSync = false
      })
    } else if (Math.abs(target - current) > 0.01) {
      current = target
      ignoreScrollSync = true
      window.scrollTo(0, current)
      notify()
      requestAnimationFrame(() => {
        ignoreScrollSync = false
      })
    }
    frame = requestAnimationFrame(tick)
  }

  const api: SmoothScrollApi = {
    scrollTo: (y) => {
      const next = clampY(y)
      target = next
      current = next
      ignoreScrollSync = true
      window.scrollTo(0, next)
      notify()
      requestAnimationFrame(() => {
        ignoreScrollSync = false
      })
    },
    setTarget: (y) => {
      target = clampY(y)
      notify()
    },
    getMetrics: metrics,
    subscribe: (listener) => {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
  }

  if (reduced) {
    setSmoothScrollApi({
      scrollTo: (y) => {
        window.scrollTo(0, clampY(y))
        notify()
      },
      setTarget: (y) => {
        window.scrollTo(0, clampY(y))
        notify()
      },
      getMetrics: readScrollMetrics,
      subscribe: (listener) => {
        const onScroll = () => listener()
        window.addEventListener('scroll', onScroll, { passive: true })
        listener()
        return () => window.removeEventListener('scroll', onScroll)
      },
    })
    document.documentElement.classList.remove('smooth-scroll-active')
    return () => setSmoothScrollApi(null)
  }

  setSmoothScrollApi(api)
  document.documentElement.classList.add('smooth-scroll-active')

  document.addEventListener('wheel', onWheel, { passive: false, capture: true })
  document.addEventListener('touchstart', onTouchStart, { passive: true, capture: true })
  document.addEventListener('touchmove', onTouchMove, { passive: false, capture: true })
  document.addEventListener('touchend', onTouchEnd, { passive: true, capture: true })
  document.addEventListener('touchcancel', onTouchEnd, { passive: true, capture: true })
  window.addEventListener('scroll', syncFromNative, { passive: true })
  frame = requestAnimationFrame(tick)

  return () => {
    cancelAnimationFrame(frame)
    document.removeEventListener('wheel', onWheel, { capture: true })
    document.removeEventListener('touchstart', onTouchStart, { capture: true })
    document.removeEventListener('touchmove', onTouchMove, { capture: true })
    document.removeEventListener('touchend', onTouchEnd, { capture: true })
    document.removeEventListener('touchcancel', onTouchEnd, { capture: true })
    window.removeEventListener('scroll', syncFromNative)
    document.documentElement.classList.remove('smooth-scroll-active')
    setSmoothScrollApi(null)
  }
}

let activeCleanups = 0
let cleanupSmoothScroll: (() => void) | null = null

/** Eased wheel + touch scroll site-wide (matches SOCIAL_THINGS). */
export function useSmoothScroll(enabled: boolean) {
  useEffect(() => {
    if (!enabled) return

    activeCleanups += 1
    if (activeCleanups === 1) {
      cleanupSmoothScroll = startSmoothScroll()
    }

    return () => {
      activeCleanups -= 1
      if (activeCleanups === 0 && cleanupSmoothScroll) {
        cleanupSmoothScroll()
        cleanupSmoothScroll = null
      }
    }
  }, [enabled])
}

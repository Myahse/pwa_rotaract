import { useCallback, useEffect, useRef, useState } from 'react'
import {
  getSmoothScrollApi,
  readScrollMetrics,
  type SmoothScrollMetrics,
} from '../lib/smooth-scroll-api'

const INDICATOR_SIZE = 8
const RAIL_TOP_OFFSET = '2.5rem'
const RAIL_BOTTOM_OFFSET = '11rem'

type Props = {
  onDark?: boolean
}

export function ScrollRail({ onDark = false }: Props) {
  const trackRef = useRef<HTMLDivElement>(null)
  const dragRef = useRef(false)
  const [metrics, setMetrics] = useState<SmoothScrollMetrics>(() => readScrollMetrics())

  const sync = useCallback(() => {
    const api = getSmoothScrollApi()
    setMetrics(api ? api.getMetrics() : readScrollMetrics())
  }, [])

  useEffect(() => {
    sync()
    const api = getSmoothScrollApi()
    if (api) return api.subscribe(sync)

    const onScroll = () => sync()
    window.addEventListener('scroll', onScroll, { passive: true })
    window.addEventListener('resize', sync)
    const observer = new ResizeObserver(sync)
    observer.observe(document.documentElement)
    return () => {
      window.removeEventListener('scroll', onScroll)
      window.removeEventListener('resize', sync)
      observer.disconnect()
    }
  }, [sync])

  const scrollToProgress = useCallback((progress: number) => {
    const api = getSmoothScrollApi()
    const { maxScrollY } = readScrollMetrics()
    const y = Math.min(maxScrollY, Math.max(0, progress * maxScrollY))
    if (api) api.setTarget(y)
    else window.scrollTo(0, y)
  }, [])

  const pointerToProgress = useCallback((clientY: number) => {
    const track = trackRef.current
    if (!track) return 0
    const rect = track.getBoundingClientRect()
    const usable = Math.max(1, rect.height - INDICATOR_SIZE)
    const y = clientY - rect.top - INDICATOR_SIZE / 2
    return Math.min(1, Math.max(0, y / usable))
  }, [])

  useEffect(() => {
    function onPointerMove(event: PointerEvent) {
      if (!dragRef.current) return
      scrollToProgress(pointerToProgress(event.clientY))
    }

    function onPointerUp() {
      dragRef.current = false
    }

    window.addEventListener('pointermove', onPointerMove)
    window.addEventListener('pointerup', onPointerUp)
    window.addEventListener('pointercancel', onPointerUp)
    return () => {
      window.removeEventListener('pointermove', onPointerMove)
      window.removeEventListener('pointerup', onPointerUp)
      window.removeEventListener('pointercancel', onPointerUp)
    }
  }, [pointerToProgress, scrollToProgress])

  if (metrics.maxScrollY <= 8) return null

  const indicatorTop =
    metrics.maxScrollY > 0
      ? `calc(${metrics.progress * 100}% - ${metrics.progress * INDICATOR_SIZE}px)`
      : '0px'

  return (
    <div
      className="scroll-rail"
      style={{
        top: `calc(var(--header-h) + ${RAIL_TOP_OFFSET})`,
        height: `calc(100dvh - var(--header-h) - ${RAIL_TOP_OFFSET} - ${RAIL_BOTTOM_OFFSET})`,
      }}
      aria-hidden
    >
      <div
        ref={trackRef}
        role="scrollbar"
        aria-orientation="vertical"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(metrics.progress * 100)}
        aria-label="Page scroll"
        className="scroll-rail-track"
        onPointerDown={(event) => {
          if (event.button !== 0) return
          const target = event.target as HTMLElement
          if (target.dataset.scrollIndicator === 'true') return
          scrollToProgress(pointerToProgress(event.clientY))
        }}
      >
        <div className={`scroll-rail-line ${onDark ? 'scroll-rail-line--dark' : ''}`} aria-hidden />
        <div
          data-scroll-indicator="true"
          className={`scroll-rail-dot ${onDark ? 'scroll-rail-dot--dark' : ''}`}
          style={{ top: indicatorTop }}
          onPointerDown={(event) => {
            event.stopPropagation()
            dragRef.current = true
            ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)
          }}
        />
      </div>
    </div>
  )
}

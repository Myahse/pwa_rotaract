import { useEffect, useRef, useState, type ReactNode } from 'react'

type Props = {
  id?: string
  className?: string
  children: ReactNode
  immediate?: boolean
}

export function RevealSection({ id, className = '', children, immediate = false }: Props) {
  const ref = useRef<HTMLElement>(null)
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    if (immediate) {
      const frame = window.requestAnimationFrame(() => setVisible(true))
      return () => window.cancelAnimationFrame(frame)
    }

    const node = ref.current
    if (!node) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) {
          setVisible(true)
          observer.disconnect()
        }
      },
      { threshold: 0.1, rootMargin: '0px 0px -2% 0px' },
    )

    observer.observe(node)
    return () => observer.disconnect()
  }, [immediate])

  const classes = ['reveal-section', visible ? 'is-visible' : '', className].filter(Boolean).join(' ')

  return (
    <section id={id} ref={ref} className={classes}>
      {children}
    </section>
  )
}

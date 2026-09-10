export function safeWavePayUrl(value: string | undefined) {
  try {
    const parsed = new URL((value ?? '').trim())
    if (parsed.protocol !== 'https:') return ''
    if (parsed.hostname !== 'pay.wave.com') return ''
    return parsed.toString()
  } catch {
    return ''
  }
}

export const WAVE_PAY_URL = safeWavePayUrl(import.meta.env.VITE_WAVE_PAY_URL) ||
  'https://pay.wave.com/m/M_ci_pHlyZFYyH1Su/c/ci/'

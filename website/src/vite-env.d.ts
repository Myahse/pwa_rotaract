/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_CLUB_APP_URL?: string
  readonly VITE_WAVE_PAY_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

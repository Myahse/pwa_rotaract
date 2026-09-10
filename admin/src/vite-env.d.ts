/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_WEBSITE_URL?: string
  readonly VITE_TOMBOLA_ORGANIZER_URL?: string
  readonly VITE_TOMBOLA_CAMPAIGN_URL?: string
  readonly VITE_TOMBOLA_MONITOR_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

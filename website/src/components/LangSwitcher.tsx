import { useI18n, type Lang } from '../i18n'

export function LangSwitcher() {
  const { lang, setLang, t } = useI18n()

  function select(next: Lang) {
    setLang(next)
  }

  return (
    <div className="lang-switch" role="group" aria-label={t.langLabel}>
      <button
        type="button"
        className={lang === 'fr' ? 'active' : ''}
        onClick={() => select('fr')}
        aria-pressed={lang === 'fr'}
      >
        {t.langFr}
      </button>
      <button
        type="button"
        className={lang === 'en' ? 'active' : ''}
        onClick={() => select('en')}
        aria-pressed={lang === 'en'}
      >
        {t.langEn}
      </button>
    </div>
  )
}

import { useEffect, useState, type FormEvent } from 'react'
import type { Donation } from '../api/types'
import { apiUpload } from '../api/client'
import { RevealSection } from '../components/RevealSection'
import { useI18n } from '../i18n'
import { prepareReceiptFile, RECEIPT_ACCEPT } from '../lib/receiptUpload'
import { WAVE_PAY_URL } from '../lib/wave'

export function DonatePage() {
  const { t } = useI18n()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [amount, setAmount] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState('')
  const [fileError, setFileError] = useState('')
  const [error, setError] = useState('')
  const [done, setDone] = useState<Donation | null>(null)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!file) {
      setPreviewUrl('')
      return
    }
    if (file.type.startsWith('image/')) {
      const url = URL.createObjectURL(file)
      setPreviewUrl(url)
      return () => URL.revokeObjectURL(url)
    }
    setPreviewUrl('')
  }, [file])

  async function onPick(next: File | undefined) {
    if (!next) {
      setFile(null)
      setFileError('')
      return
    }
    try {
      await prepareReceiptFile(next)
      setFile(next)
      setFileError('')
    } catch (err) {
      const code = err instanceof Error ? err.message : ''
      setFile(null)
      setFileError(code === 'too_large' ? t.receiptTooLarge : t.receiptInvalidType)
    }
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    const gift = Number(amount)
    if (!name.trim() || !email.trim() || !Number.isFinite(gift) || gift <= 0 || !file) {
      setError(t.donateInvalid)
      return
    }
    setBusy(true)
    setError('')
    try {
      const prepared = await prepareReceiptFile(file)
      const body = new FormData()
      body.append('name', name.trim())
      body.append('email', email.trim())
      body.append('amount_xof', String(Math.round(gift)))
      body.append('file', prepared.blob, prepared.filename)
      const donation = await apiUpload<Donation>('/donations', body)
      setDone(donation)
      setAmount('')
      setFile(null)
    } catch {
      setError(t.donateInvalid)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="screen-scroll">
      <RevealSection className="screen-section screen-section--alt screen-section--donate" immediate>
        <div className="screen-section-inner screen-section-inner--wide screen-section-inner--fill screen-section-inner--stack">
          <p className="screen-eyebrow">{t.donateKicker}</p>
          <h1>{t.donateTitle}</h1>
          <p className="lede">{t.donateLede}</p>
          <div className="cta-row">
            <a className="btn-primary" href={WAVE_PAY_URL} target="_blank" rel="noopener noreferrer">
              {t.donateWaveCta}
            </a>
          </div>
          <h2>{t.donateRefTitle}</h2>
          <p className="lede">{t.donateRefHelp}</p>
          {done ? (
            <p className="form-ok">{done.known_member ? t.donateSentKnown : t.donateSent}</p>
          ) : (
            <form className="donate-form screen-section--fill-grid" onSubmit={onSubmit}>
              <label>
                {t.donateName}
                <input value={name} onChange={(e) => setName(e.target.value)} autoComplete="name" required minLength={2} />
              </label>
              <label>
                {t.donateEmailLabel}
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoComplete="email"
                  required
                />
              </label>
              <label>
                {t.donateAmount}
                <input
                  type="number"
                  inputMode="numeric"
                  min={100}
                  step={100}
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  required
                />
              </label>
              <label className="receipt-picker">
                <span>{t.receiptFileLabel}</span>
                <span className="receipt-picker-box">
                  <input
                    type="file"
                    accept={RECEIPT_ACCEPT}
                    capture="environment"
                    disabled={busy}
                    onChange={(e) => void onPick(e.target.files?.[0])}
                  />
                  {previewUrl ? (
                    <img src={previewUrl} alt="" className="receipt-preview" />
                  ) : file ? (
                    <span className="receipt-file-name">{file.name}</span>
                  ) : (
                    <span className="receipt-picker-hint">{t.receiptPickHint}</span>
                  )}
                </span>
              </label>
              {fileError ? <p className="form-error">{fileError}</p> : null}
              {error ? <p className="form-error">{error}</p> : null}
              <button type="submit" className="btn-primary btn-block" disabled={busy || !file}>
                {busy ? t.donateSending : t.donateSend}
              </button>
            </form>
          )}
        </div>
      </RevealSection>
    </div>
  )
}

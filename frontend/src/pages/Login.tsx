import { FormEvent, useState } from 'react'
import api, { errorMessage } from '../api/api'
import useStore from '../store/useStore'

export default function Login() {
  const [email, setEmail] = useState('')
  const [otp, setOTP] = useState('')
  const [sent, setSent] = useState(false)
  const [delivery, setDelivery] = useState<'email' | 'development-log' | ''>('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setBusy(true); setError('')
    try {
      if (!sent) {
        const response = await api.post('/auth/send-otp', { email })
        setDelivery(response.data.delivery === 'email' ? 'email' : 'development-log')
        setSent(true)
      } else {
        const response = await api.post('/auth/verify-otp', { email, otp })
        useStore.getState().setUser(response.data.user)
      }
    } catch (error) { setError(errorMessage(error)) }
    finally { setBusy(false) }
  }
  return <main className="login-shell">
    <section className="login-intro"><span className="brand">UnAlone<span className="brand-dot">●</span></span>
      <p className="eyebrow">A little closer to your people</p>
      <h1>Find the places.<br/>Feel the company.</h1>
      <p>Discover where people are gathering right now, and find a public meetup nearby.</p>
      <div className="intro-facts"><span>Live local hotspots</span><span>Public meetups</span><span>Ephemeral presence</span></div>
    </section>
    <section className="login-card"><h2>{sent ? 'Enter your code' : 'Start nearby'}</h2>
      <p>{sent ? 'Your six-digit code expires after five minutes.' : 'Sign in with your email. No password to remember.'}</p>
      <form onSubmit={submit}>
        <label>Email<input required type="email" autoComplete="email" value={email} disabled={sent} onChange={e => setEmail(e.target.value)} placeholder="you@example.com"/></label>
        {sent && <label>Verification code<input required autoFocus inputMode="numeric" autoComplete="one-time-code" pattern="[0-9]{6}" maxLength={6} value={otp} onChange={e => setOTP(e.target.value.replace(/\D/g,''))}/></label>}
        {error && <p className="error" role="alert">{error}</p>}
        <button className="primary" disabled={busy}>{busy ? 'One moment…' : sent ? 'Sign in' : 'Get a sign-in code'}</button>
      </form>
      {sent && <button className="text-button" onClick={() => {setSent(false);setOTP('');setError('');setDelivery('')}}>Change email or request another code</button>}
      <p className="local-note">{sent && delivery === 'email'
        ? 'We sent a six-digit code to your inbox. It expires in five minutes.'
        : 'This local server is using development log mode. Your code appears in the backend terminal.'}</p>
    </section>
  </main>
}

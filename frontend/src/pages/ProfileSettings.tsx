import { useState } from 'react'
import api, { errorMessage } from '../api/api'
import useStore from '../store/useStore'

export default function ProfileSettings() {
  const user = useStore(s => s.user)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const logout = async () => {
    setBusy(true)
    try {
      await api.post('/auth/logout')
      localStorage.removeItem('token')
      localStorage.removeItem('email')
      useStore.getState().setUser(null)
    } catch (error) { setError(errorMessage(error)) }
    finally { setBusy(false) }
  }
  return <section className="page-card"><p className="eyebrow">Your account</p><h1>Keep it simple.</h1>
    <p>Signed in as <strong>{user?.email}</strong></p>
    <p>Your live location is shared only while this tab is visible and sharing is enabled. Server location records expire within 30 seconds. Public meetup locations remain saved.</p>
    {error && <p className="error" role="alert">{error}</p>}
    <button className="primary" onClick={logout} disabled={busy}>{busy ? 'Signing out…' : 'Sign out'}</button>
  </section>
}

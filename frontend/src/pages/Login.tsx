import React, { useState } from 'react'
import api from '../api/api'

export default function Login({ onLogin }:{ onLogin:()=>void }){
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [sent, setSent] = useState(false)

  const send = async () => {
    await api.post('/auth/send-otp', { email })
    setSent(true)
  }

  const verify = async () => {
    const res = await api.post('/auth/verify-otp', { email, code })
    const token = res.data.token
    localStorage.setItem('token', token)
    onLogin()
  }

  return (
    <div className="login-page">
      <h2>Login</h2>
      <input value={email} onChange={(e)=>setEmail(e.target.value)} placeholder="you@example.com" />
      <button onClick={send}>Send OTP</button>
      {sent && (
        <div>
          <input value={code} onChange={(e)=>setCode(e.target.value)} placeholder="Enter code" />
          <button onClick={verify}>Verify</button>
        </div>
      )}
    </div>
  )
}

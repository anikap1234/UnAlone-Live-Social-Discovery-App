import React from 'react'

export default function ProfileSettings({ onDone }:{ onDone:()=>void }){
  return (
    <div className="profile-settings">
      <h3>Profile</h3>
      <p>Logged in as: {localStorage.getItem('email') || 'unknown'}</p>
      <button onClick={() => { localStorage.removeItem('token'); onDone() }}>Logout</button>
    </div>
  )
}

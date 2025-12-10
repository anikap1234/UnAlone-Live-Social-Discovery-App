import React, { useState } from 'react'
import api from '../api/api'

export default function MeetupCreator({ onDone }:{ onDone:()=>void }){
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [lat, setLat] = useState('')
  const [lon, setLon] = useState('')

  const submit = async () => {
    await api.post('/meetups', { title, description, lat: parseFloat(lat), lon: parseFloat(lon) })
    onDone()
  }

  return (
    <div className="meetup-creator">
      <h3>Create Meetup</h3>
      <input placeholder="Title" value={title} onChange={(e)=>setTitle(e.target.value)} />
      <textarea placeholder="Description" value={description} onChange={(e)=>setDescription(e.target.value)} />
      <input placeholder="Lat" value={lat} onChange={(e)=>setLat(e.target.value)} />
      <input placeholder="Lon" value={lon} onChange={(e)=>setLon(e.target.value)} />
      <button onClick={submit}>Create</button>
    </div>
  )
}

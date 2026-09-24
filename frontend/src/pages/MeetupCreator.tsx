import { FormEvent, useState } from 'react'
import api, { errorMessage } from '../api/api'
import useStore, { Meetup } from '../store/useStore'

function localDate(date: Date) { return new Date(date.getTime()-date.getTimezoneOffset()*60000).toISOString().slice(0,16) }
export default function MeetupCreator({ onDone }: { onDone: (meetup?: Meetup) => void }) {
  const position = useStore(s => s.position)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [lat, setLat] = useState(position?.lat.toString() || '')
  const [lon, setLon] = useState(position?.lon.toString() || '')
  const [time, setTime] = useState(localDate(new Date(Date.now()+3600000)))
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setBusy(true); setError('')
    try {
      const response = await api.post('/meetups', { title, description, lat: Number(lat), lon: Number(lon), time: Math.floor(new Date(time).getTime()/1000) })
      const meetup: Meetup = response.data.meetup
      useStore.getState().addMeetup(meetup)
      onDone(meetup)
    } catch (error) { setError(errorMessage(error)) }
    finally { setBusy(false) }
  }
  return <section className="page-card"><p className="eyebrow">Make a little room for company</p><h1>Create a meetup</h1>
    <p>A public invitation to do something together. Your meetup's location and details will be visible to other signed-in people.</p>
    <form onSubmit={submit} className="meetup-form">
      <label>What are you planning?<input required minLength={3} maxLength={100} value={title} onChange={e=>setTitle(e.target.value)} placeholder="Coffee and a good book"/></label>
      <label>A few details<textarea maxLength={1000} rows={4} value={description} onChange={e=>setDescription(e.target.value)} placeholder="What should people know?"/></label>
      <label>When<input required type="datetime-local" min={localDate(new Date())} value={time} onChange={e=>setTime(e.target.value)}/></label>
      <div className="form-row"><label>Latitude<input required type="number" step="any" min={-85.05112878} max={85.05112878} value={lat} onChange={e=>setLat(e.target.value)}/></label>
      <label>Longitude<input required type="number" step="any" min={-180} max={180} value={lon} onChange={e=>setLon(e.target.value)}/></label></div>
      <button className="text-button" type="button" disabled={!position} onClick={()=>{if(position){setLat(String(position.lat));setLon(String(position.lon))}}}>Use my current location</button>
      {error && <p className="error" role="alert">{error}</p>}
      <div className="actions"><button className="primary" disabled={busy}>{busy ? 'Saving…' : 'Create public meetup'}</button><button type="button" onClick={()=>onDone()}>Cancel</button></div>
    </form>
  </section>
}

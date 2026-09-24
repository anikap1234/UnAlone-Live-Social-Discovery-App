import { useMemo, useState } from 'react'
import MapboxMap from '../components/MapboxMap'
import MeetupCard from '../components/MeetupCard'
import useStore, { Meetup } from '../store/useStore'
import { distance } from '../utils/distance'

export default function MapView({ onOpenCreate, created }: { onOpenCreate: () => void; created: Meetup | null }) {
  const position = useStore(s=>s.position)
  const allMeetups = useStore(s=>s.meetups)
  const allHotspots = useStore(s=>s.hotspots)
  const [selected,setSelected] = useState<Meetup | null>(created)
  const meetups = useMemo(()=>allMeetups.filter(m=>m.time>Date.now()/1000 && (!position || distance(position,m)<=1000)).sort((a,b)=>a.time-b.time),[allMeetups,position])
  const hotspots = useMemo(()=>allHotspots.filter(h=>!position || distance(position,h)<=1000),[allHotspots,position])
  return <main className="discovery"><aside className="discovery-panel"><p className="eyebrow">A little company, close by</p><h1>What's happening?</h1>
    <p>{position ? 'Live hotspots and upcoming meetups within 1 km of you.' : 'Share your location to discover what is nearby.'}</p>
    <div className="stats"><div><strong>{hotspots.length}</strong><span>live hotspots</span></div><div><strong>{meetups.length}</strong><span>upcoming meetups</span></div></div>
    <button className="primary" onClick={onOpenCreate}>+ Create a meetup</button>
    <div className="list-heading"><h2>Meet someone nearby</h2><span>Public meetups</span></div>
    {meetups.length===0 ? <div className="empty">Nothing here yet. Start a meetup and make the first move.</div> : meetups.map(m=><MeetupCard key={m.meetupId} m={m} onSelect={()=>setSelected(m)}/>)}
  </aside><MapboxMap position={position} meetups={meetups} hotspots={hotspots} selected={selected}/></main>
}

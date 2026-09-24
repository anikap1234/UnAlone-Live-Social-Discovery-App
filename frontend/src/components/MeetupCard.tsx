import { Meetup } from '../store/useStore'
export default function MeetupCard({ m, onSelect }: { m: Meetup; onSelect?: () => void }) {
  return <article className="meetup-card">
    <p className="eyebrow">{new Date(m.time*1000).toLocaleString([], {month:'short',day:'numeric',hour:'2-digit',minute:'2-digit'})}</p>
    <h3>{m.title}</h3><p>{m.description || 'Come along and say hello.'}</p>
    {onSelect && <button className="text-button" onClick={onSelect}>Show on map ↗</button>}
  </article>
}

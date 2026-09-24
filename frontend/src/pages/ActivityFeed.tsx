import useStore from '../store/useStore'

export default function ActivityFeed() {
  const feed = useStore(s => s.feed)
  return <section className="page-card"><p className="eyebrow">Happening live</p><h1>Activity</h1>
    <p>This feed belongs to your current connection. It clears on refresh or reconnect.</p>
    {feed.length === 0 ? <div className="empty">A quiet moment. New hotspots and meetups will appear here as they happen.</div>
      : <ol className="activity-list">{feed.map(event => <li key={event.id}><span className="activity-dot"/><div><span>{event.text}</span><small>{new Date(event.time).toLocaleTimeString([], {hour:'2-digit',minute:'2-digit'})}</small></div></li>)}</ol>}
  </section>
}

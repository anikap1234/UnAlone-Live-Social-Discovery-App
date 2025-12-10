import { useEffect, useRef } from 'react'
import useStore from '../store/useStore'

export default function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const addHotspot = useStore(state => state.addHotspot)
  const addMeetup = useStore(state => state.addMeetup)

  useEffect(() => {
    const apiBase = (import.meta.env.VITE_API_BASE || 'http://localhost:8080').replace(/^http/, 'ws')
    const ws = new WebSocket(`${apiBase.replace(/\/$/, '')}/ws`)
    wsRef.current = ws
    ws.onmessage = (ev) => {
      try {
        const parsed = JSON.parse(ev.data)
        if (parsed.type === 'HOTSPOT_UPDATE') {
          addHotspot({ lat: parsed.lat, lon: parsed.lon, activeUsers: parsed.activeUsers })
        }
        if (parsed.type === 'MEETUP_CREATED') {
          const m = parsed.data
          addMeetup({ id: m.id, title: m.title, description: m.description, lat: m.lat, lon: m.lon })
        }
      } catch (e) {
        // ignore
      }
    }
    return () => {
      ws.close()
    }
  }, [addHotspot, addMeetup])
}

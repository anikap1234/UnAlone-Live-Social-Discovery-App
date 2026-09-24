import { useEffect } from 'react'
import api, { API_BASE } from '../api/api'
import useStore, { LiveEvent } from '../store/useStore'

export default function useWebSocket(enabled: boolean) {
  useEffect(() => {
    if (!enabled) return
    let stopped = false
    let retries = 0
    let timer: number | undefined
    let socket: WebSocket | undefined
    let controller: AbortController | undefined
    const connect = () => {
      if (stopped) return
      if (!navigator.onLine) { useStore.getState().setConnection('offline'); return }
      useStore.getState().setConnection('connecting')
      const base = new URL(API_BASE.replace(/\/$/, '') + '/ws', window.location.origin)
      base.protocol = base.protocol === 'https:' ? 'wss:' : 'ws:'
      const ws = new WebSocket(base)
      socket = ws
      let ready = false
      const queued: LiveEvent[] = []
      ws.onopen = async () => {
        if (stopped) return
        controller = new AbortController()
        useStore.getState().resetLive()
        try {
          const response = await api.get('/hotspots', { signal: controller.signal })
          if (stopped || socket !== ws || ws.readyState !== WebSocket.OPEN) return
          useStore.getState().setHotspots(response.data.hotspots)
          queued.forEach(event => useStore.getState().applyEvent(event))
          ready = true
          retries = 0
          useStore.getState().setConnection('live')
        } catch { ws.close() }
      }
      ws.onmessage = message => {
        try {
          const event: LiveEvent = JSON.parse(message.data)
          if (!['HOTSPOT_FORMED', 'HOTSPOT_DISSOLVED', 'MEETUP_CREATED'].includes(event.type)) return
          if (ready) useStore.getState().applyEvent(event)
          else if (queued.length < 200) queued.push(event)
          else ws.close()
        } catch { /* An invalid frame cannot change application state. */ }
      }
      ws.onerror = () => ws.close()
      ws.onclose = () => {
        if (stopped || socket !== ws) return
        controller?.abort()
        useStore.getState().setConnection('offline')
        useStore.getState().setHotspots([])
        if (navigator.onLine) timer = window.setTimeout(connect, Math.min(1000 * 2 ** retries++, 15000))
      }
    }
    const offline = () => {
      window.clearTimeout(timer)
      controller?.abort()
      socket?.close()
      useStore.getState().setConnection('offline')
      useStore.getState().setHotspots([])
    }
    const online = () => {
      window.clearTimeout(timer)
      socket?.close()
      connect()
    }
    window.addEventListener('offline', offline)
    window.addEventListener('online', online)
    connect()
    return () => {
      stopped = true
      window.clearTimeout(timer)
      controller?.abort()
      socket?.close()
      window.removeEventListener('offline', offline)
      window.removeEventListener('online', online)
    }
  }, [enabled])
}

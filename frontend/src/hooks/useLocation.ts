import { useEffect, useRef } from 'react'
import api from '../api/api'

export default function useLocation(enabled: boolean = true) {
  const watchId = useRef<number | null>(null)

  useEffect(() => {
    if (!enabled) return
    if (!('geolocation' in navigator)) return
    const sendPos = (pos: GeolocationPosition) => {
      api.post('/location/update', { lat: pos.coords.latitude, lon: pos.coords.longitude }).catch(() => {})
    }
    const success = (pos: GeolocationPosition) => { sendPos(pos) }
    const err = () => {}
    watchId.current = navigator.geolocation.watchPosition(success, err, { enableHighAccuracy: true, maximumAge: 5000, timeout: 5000 }) as unknown as number
    // also send every 7s via getCurrentPosition
    const iv = setInterval(() => {
      navigator.geolocation.getCurrentPosition((pos) => sendPos(pos))
    }, 7000)
    return () => {
      if (watchId.current !== null) navigator.geolocation.clearWatch(watchId.current as any)
      clearInterval(iv)
    }
  }, [enabled])
}

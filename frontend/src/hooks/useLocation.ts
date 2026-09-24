import { useEffect, useRef } from 'react'
import api, { errorMessage } from '../api/api'
import useStore from '../store/useStore'

export default function useLocation(enabled: boolean) {
  const lastSent = useRef(0)
  useEffect(() => {
    let active = true
    let generation = 0
    let pending = false
    const state = useStore.getState
    const send = () => {
      if (!active || !enabled || document.hidden || pending || Date.now() - lastSent.current < 5000) return
      if (!navigator.geolocation) { state().setLocationStatus('This browser does not support location'); return }
      pending = true
      const current = generation
      navigator.geolocation.getCurrentPosition(async pos => {
        pending = false
        if (!active || document.hidden || current !== generation) return
        const position = { lat: pos.coords.latitude, lon: pos.coords.longitude }
        state().setPosition(position)
        lastSent.current = Date.now()
        try {
          await api.post('/location/update', position)
          if (active && current === generation) state().setLocationStatus('Sharing while this tab is visible')
        } catch (error) {
          if (active && current === generation) state().setLocationStatus(errorMessage(error))
        }
      }, error => {
        pending = false
        if (active && current === generation) state().setLocationStatus(error.code === 1
          ? 'Location permission is off. Allow it in your browser, then retry sharing.'
          : 'Could not find your location. We will retry.')
      }, { enableHighAccuracy: true, maximumAge: 0, timeout: 10000 })
    }
    const visibility = () => {
      generation++
      pending = false
      if (document.hidden) {
        state().setLocationStatus('Sharing paused while this tab is hidden')
        state().setPosition(null)
      } else send()
    }
    if (enabled) send()
    else {
      state().setPosition(null)
      state().setLocationStatus('Location sharing is paused')
    }
    const timer = enabled ? window.setInterval(send, 7000) : undefined
    document.addEventListener('visibilitychange', visibility)
    return () => {
      active = false
      generation++
      window.clearInterval(timer)
      document.removeEventListener('visibilitychange', visibility)
    }
  }, [enabled])
}

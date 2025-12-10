import React, { useEffect, useRef } from 'react'
import mapboxgl from 'mapbox-gl'
import useStore from '../store/useStore'

mapboxgl.accessToken = import.meta.env.VITE_MAPBOX_TOKEN || ''

export default function MapboxMap(){
  const ref = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<mapboxgl.Map | null>(null)
  const hotspots = useStore(s => s.hotspots)
  const meetups = useStore(s => s.meetups)

  useEffect(() => {
    if (!ref.current) return
    mapRef.current = new mapboxgl.Map({ container: ref.current, style: 'mapbox://styles/mapbox/streets-v11', center: [0,0], zoom: 2 })
    const map = mapRef.current
    map.addControl(new mapboxgl.NavigationControl())
    return () => { map.remove() }
  }, [])

  useEffect(() => {
    const map = mapRef.current
    if (!map) return
    // remove existing markers
    const existing = document.getElementsByClassName('meetup-marker')
    while (existing.length) existing[0].remove()
    for (const m of meetups){
      const el = document.createElement('div')
      el.className = 'meetup-marker'
      el.style.background = '#e74c3c'
      el.style.width = '16px'
      el.style.height = '16px'
      el.style.borderRadius = '8px'
      new mapboxgl.Marker(el).setLngLat([m.lon, m.lat]).addTo(map)
    }
  }, [meetups])

  useEffect(() => {
    const map = mapRef.current
    if (!map) return
    // draw hotspot pulses using circle layers or markers
    // For simplicity add circle via canvas markers
    const pulses = document.getElementsByClassName('hotspot-pulse')
    while (pulses.length) pulses[0].remove()
    for (const h of hotspots){
      const el = document.createElement('div')
      el.className = 'hotspot-pulse'
      el.style.width = '120px'
      el.style.height = '120px'
      el.style.marginLeft = '-60px'
      el.style.marginTop = '-60px'
      el.style.borderRadius = '60px'
      el.style.background = 'rgba(52, 152, 219, 0.25)'
      el.style.boxShadow = '0 0 20px rgba(52,152,219,0.5)'
      new mapboxgl.Marker(el).setLngLat([h.lon, h.lat]).addTo(map)
    }
  }, [hotspots])

  return <div ref={ref} style={{ width: '100%', height: '100vh' }} />
}

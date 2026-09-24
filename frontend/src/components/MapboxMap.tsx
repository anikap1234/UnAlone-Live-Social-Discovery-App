import { useEffect, useRef, useState } from 'react'
import mapboxgl from 'mapbox-gl'
import 'mapbox-gl/dist/mapbox-gl.css'
import { Hotspot, Meetup, Position } from '../store/useStore'

const token = import.meta.env.VITE_MAPBOX_TOKEN
const localStyle: mapboxgl.Style = {
  version: 8,
  sources: { streets: { type: 'raster', tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'], tileSize: 256,
    attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors', maxzoom: 19 } },
  layers: [{ id: 'streets', type: 'raster', source: 'streets' }]
}

export default function MapboxMap({ position, meetups, hotspots, selected }: {position: Position | null; meetups: Meetup[]; hotspots: Hotspot[]; selected: Meetup | null}) {
  const container = useRef<HTMLDivElement>(null)
  const instance = useRef<mapboxgl.Map | null>(null)
  const centered = useRef(false)
  const [error, setError] = useState('')
  useEffect(() => {
    if (!container.current) return
    try {
      if (token) mapboxgl.accessToken = token
      const map = new mapboxgl.Map({ container: container.current,
        style: token ? 'mapbox://styles/mapbox/streets-v12' : localStyle,
        center: [77.5946,12.9716], zoom: 12, attributionControl: false })
      map.addControl(new mapboxgl.NavigationControl(), 'top-right')
      map.addControl(new mapboxgl.AttributionControl({compact:false}), 'bottom-right')
      map.on('error', () => setError('Some map tiles could not load. Live data and meetup details are still available.'))
      instance.current = map
      centered.current = false
      return () => { instance.current = null; map.remove() }
    } catch { setError('This browser could not start the map. Check that WebGL is enabled.') }
  }, [])
  useEffect(() => {
    const map = instance.current
    if (!map || !position) return
    const element = document.createElement('div')
    element.className = 'user-marker'
    element.setAttribute('aria-label','Your current location')
    const marker = new mapboxgl.Marker(element).setLngLat([position.lon,position.lat]).addTo(map)
    if (!centered.current) { map.jumpTo({center:[position.lon,position.lat],zoom:15}); centered.current = true }
    return () => {marker.remove()}
  }, [position])
  useEffect(() => {
    const map = instance.current
    if (!map) return
    const markers: mapboxgl.Marker[] = []
    for (const meetup of meetups) {
      const element = document.createElement('button')
      element.className = 'meetup-marker'
      element.textContent = '●'
      element.setAttribute('aria-label',meetup.title)
      const content = document.createElement('section')
      const title = document.createElement('h3'); title.textContent = meetup.title
      const time = document.createElement('p'); time.textContent = new Date(meetup.time*1000).toLocaleString()
      const description = document.createElement('p'); description.textContent = meetup.description
      content.append(title,time,description)
      const popup = new mapboxgl.Popup({offset:20}).setDOMContent(content)
      markers.push(new mapboxgl.Marker(element).setLngLat([meetup.lon,meetup.lat]).setPopup(popup).addTo(map))
    }
    for (const hotspot of hotspots) {
      const element = document.createElement('button')
      element.className = 'hotspot-pulse'
      element.textContent = String(hotspot.activeUsers)
      element.setAttribute('aria-label',hotspot.activeUsers+' people at this hotspot')
      const popup = new mapboxgl.Popup({offset:24}).setText(hotspot.activeUsers+' people are active here')
      markers.push(new mapboxgl.Marker(element).setLngLat([hotspot.lon,hotspot.lat]).setPopup(popup).addTo(map))
    }
    return () => markers.forEach(marker => marker.remove())
  }, [meetups,hotspots])
  useEffect(() => {
    if (selected) instance.current?.flyTo({center:[selected.lon,selected.lat],zoom:16})
  }, [selected])
  return <div className="map-container"><div ref={container} className="map-canvas"/>
    {error && <p className="map-error" role="status">{error}</p>}
    <div className="map-legend"><span><i className="legend-user"/>You</span><span><i className="legend-hotspot"/>Live hotspot</span><span><i className="legend-meetup"/>Meetup</span></div>
  </div>
}

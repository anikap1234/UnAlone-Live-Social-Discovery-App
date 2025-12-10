import React from 'react'
import MapboxMap from '../components/MapboxMap'
import useWebSocket from '../hooks/useWebSocket'
import useLocation from '../hooks/useLocation'

export default function MapView({ onOpenCreate, onOpenProfile }:{ onOpenCreate:()=>void; onOpenProfile:()=>void }){
  useWebSocket()
  useLocation(true)

  return (
    <div className="mapview-root">
      <div className="topbar">
        <button onClick={onOpenCreate}>Create Meetup</button>
        <button onClick={onOpenProfile}>Profile</button>
      </div>
      <MapboxMap />
    </div>
  )
}

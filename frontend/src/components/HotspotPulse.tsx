import React from 'react'

export default function HotspotPulse({activeUsers}:{activeUsers:number}){
  return <div className="hotspot-pulse-ui">{activeUsers} people</div>
}

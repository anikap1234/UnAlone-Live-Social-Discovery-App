import React from 'react'

export default function MeetupCard({m}:{m:any}){
  return (
    <div className="meetup-card">
      <h4>{m.title}</h4>
      <p>{m.description}</p>
    </div>
  )
}

import React from 'react'
import { useState } from 'react'
import Login from './pages/Login'
import MapView from './pages/MapView'
import MeetupCreator from './pages/MeetupCreator'
import ProfileSettings from './pages/ProfileSettings'
import useStore from './store/useStore'

export default function App(){
  const token = localStorage.getItem('token')
  const [route, setRoute] = useState<string>(token ? 'map' : 'login')
  const { meetups } = useStore()

  return (
    <div className="app-root">
      {route === 'login' && <Login onLogin={() => setRoute('map')} />}
      {route === 'map' && <MapView onOpenCreate={() => setRoute('create')} onOpenProfile={() => setRoute('profile')} />}
      {route === 'create' && <MeetupCreator onDone={() => setRoute('map')} />}
      {route === 'profile' && <ProfileSettings onDone={() => setRoute('map')} />}
    </div>
  )
}

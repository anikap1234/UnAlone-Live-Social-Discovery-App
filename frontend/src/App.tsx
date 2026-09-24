import { useEffect, useState } from 'react'
import api, { errorMessage } from './api/api'
import Login from './pages/Login'
import MapView from './pages/MapView'
import MeetupCreator from './pages/MeetupCreator'
import ProfileSettings from './pages/ProfileSettings'
import ActivityFeed from './pages/ActivityFeed'
import useStore, { Meetup } from './store/useStore'
import useLocation from './hooks/useLocation'
import useWebSocket from './hooks/useWebSocket'

type Route = 'map' | 'create' | 'activity' | 'profile'
export default function App() {
  const user=useStore(s=>s.user)
  const position=useStore(s=>s.position)
  const sharing=useStore(s=>s.sharing)
  const locationStatus=useStore(s=>s.locationStatus)
  const connection=useStore(s=>s.connection)
  const [route,setRoute]=useState<Route>('map')
  const [checking,setChecking]=useState(true)
  const [startupError,setStartupError]=useState('')
  const [dataError,setDataError]=useState('')
  const [retry,setRetry]=useState(0)
  const [created,setCreated]=useState<Meetup|null>(null)
  useLocation(Boolean(user)&&sharing)
  useWebSocket(Boolean(user))
  useEffect(()=>{
    let active=true
    setChecking(true);setStartupError('')
    api.get('/auth/me').then(r=>{if(active)useStore.getState().setUser(r.data.user)}).catch(error=>{
      if(active && error.response?.status!==401)setStartupError(errorMessage(error))
    }).finally(()=>{if(active)setChecking(false)})
    return()=>{active=false}
  },[retry])
  useEffect(()=>{
    const expire=()=>{useStore.getState().setUser(null);setRoute('map')}
    window.addEventListener('session-expired',expire)
    return()=>window.removeEventListener('session-expired',expire)
  },[])
  useEffect(()=>{setRoute('map');setCreated(null)},[user?.userId])
  useEffect(()=>{
    if(!user || !position)return
    const controller=new AbortController()
    const eventSequence=useStore.getState().eventSequence
    Promise.all([
      api.get('/meetups/near',{params:{...position,radius:1000},signal:controller.signal}),
      api.get('/hotspots',{signal:controller.signal})
    ]).then(([meetups,hotspots])=>{
      if(controller.signal.aborted)return
      useStore.getState().setMeetups(meetups.data.meetups)
      if(useStore.getState().eventSequence===eventSequence)useStore.getState().setHotspots(hotspots.data.hotspots)
      setDataError('')
    }).catch(error=>{if(!controller.signal.aborted)setDataError(errorMessage(error))})
    return()=>controller.abort()
  },[user?.userId,position,connection])
  const toggleSharing=async()=>{
    useStore.getState().setSharing(!sharing)
    if(sharing){try{await api.delete('/location')}catch{setDataError('Sharing stopped locally. Your server presence will expire within 30 seconds.')}}
  }
  if(checking)return <main className="loading">Opening UnAlone…</main>
  if(startupError)return <main className="loading"><p role="alert">{startupError}</p><button className="primary" onClick={()=>setRetry(x=>x+1)}>Retry connection</button></main>
  if(!user)return <Login/>
  return <div className="app-root">
    <header className="app-header"><button className="brand" onClick={()=>setRoute('map')}>UnAlone<span className="brand-dot">●</span></button>
      <nav aria-label="Main navigation">{([['map','Discover'],['create','Create meetup'],['activity','Activity'],['profile','Profile']] as [Route,string][]).map(([key,label])=><button key={key} aria-current={route===key?'page':undefined} onClick={()=>setRoute(key)}>{label}</button>)}</nav>
      <span className={'connection '+connection}><i/>{connection==='live'?'Live':connection==='connecting'?'Connecting':'Reconnecting'}</span>
    </header>
    <div className="presence-bar"><span>{locationStatus}</span><button onClick={toggleSharing}>{sharing?'Pause sharing':'Share location'}</button></div>
    {dataError && <div className="data-error" role="alert">{dataError}</div>}
    {route==='map'&&<MapView onOpenCreate={()=>setRoute('create')} created={created}/>}
    {route==='create'&&<MeetupCreator onDone={meetup=>{if(meetup)setCreated(meetup);setRoute('map')}}/>}
    {route==='activity'&&<ActivityFeed/>}
    {route==='profile'&&<ProfileSettings/>}
  </div>
}

import { create } from 'zustand'

export type User = { userId: string; email: string }
export type Position = { lat: number; lon: number }
export type Hotspot = Position & { hotspotId: string; activeUsers: number }
export type Meetup = Position & { meetupId: string; title: string; description: string; time: number; createdBy: string; createdAt: number }
export type LiveEvent =
  | ({ type: 'HOTSPOT_FORMED' | 'HOTSPOT_DISSOLVED' } & Hotspot)
  | { type: 'MEETUP_CREATED'; data: Meetup }
export type Activity = { id: string; text: string; time: number }
type State = {
  user: User | null
  position: Position | null
  sharing: boolean
  locationStatus: string
  connection: 'connecting' | 'live' | 'offline'
  hotspots: Hotspot[]
  meetups: Meetup[]
  feed: Activity[]
  eventSequence: number
  setUser: (user: User | null) => void
  setPosition: (position: Position | null) => void
  setSharing: (sharing: boolean) => void
  setLocationStatus: (status: string) => void
  setConnection: (connection: State['connection']) => void
  setHotspots: (hotspots: Hotspot[]) => void
  setMeetups: (meetups: Meetup[]) => void
  addMeetup: (meetup: Meetup) => void
  applyEvent: (event: LiveEvent) => void
  resetLive: () => void
}
const useStore = create<State>((set) => ({
  user: null, position: null, sharing: true, locationStatus: 'Waiting for your location',
  connection: 'connecting', hotspots: [], meetups: [], feed: [], eventSequence: 0,
  setUser: user => set({ user, position: null, hotspots: [], meetups: [], feed: [], sharing: true }),
  setPosition: position => set({ position }),
  setSharing: sharing => set({ sharing }),
  setLocationStatus: locationStatus => set({ locationStatus }),
  setConnection: connection => set({ connection }),
  setHotspots: hotspots => set({ hotspots }),
  setMeetups: meetups => set(state => ({
    meetups: [...new Map([...meetups, ...state.meetups.filter(m => m.time > Date.now()/1000)].map(m => [m.meetupId, m])).values()]
  })),
  addMeetup: meetup => set(state => ({ meetups: [...state.meetups.filter(m => m.meetupId !== meetup.meetupId), meetup] })),
  resetLive: () => set({ hotspots: [], feed: [] }),
  applyEvent: event => set(state => {
    let text: string
    let hotspots = state.hotspots
    let meetups = state.meetups
    if (event.type === 'MEETUP_CREATED') {
      const meetup = event.data
      if (!meetup?.meetupId || !Number.isFinite(meetup.lat) || !Number.isFinite(meetup.lon)) return state
      meetups = [...meetups.filter(m => m.meetupId !== meetup.meetupId), meetup]
      text = 'New meetup: ' + meetup.title
    } else if (event.type === 'HOTSPOT_FORMED' || event.type === 'HOTSPOT_DISSOLVED') {
      if (!event.hotspotId || !Number.isFinite(event.lat) || !Number.isFinite(event.lon)) return state
      hotspots = hotspots.filter(h => h.hotspotId !== event.hotspotId)
      if (event.type === 'HOTSPOT_FORMED') hotspots = [...hotspots, event]
      text = event.type === 'HOTSPOT_FORMED' ? 'A hotspot formed · ' + event.activeUsers + ' people' : 'A hotspot is no longer active'
    } else return state
    return { hotspots, meetups, eventSequence: state.eventSequence + 1, feed: [{ id: crypto.randomUUID(), text, time: Date.now() }, ...state.feed].slice(0, 100) }
  })
}))
export default useStore

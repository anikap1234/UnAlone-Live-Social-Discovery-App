import create from 'zustand'

type Hotspot = { lat: number; lon: number; activeUsers: number }
type Meetup = { id: string; title: string; description: string; lat: number; lon: number }

type State = {
  hotspots: Hotspot[]
  meetups: Meetup[]
  addHotspot: (h: Hotspot) => void
  addMeetup: (m: Meetup) => void
  setMeetups: (ms: Meetup[]) => void
}

const useStore = create<State>((set) => ({
  hotspots: [],
  meetups: [],
  addHotspot: (h) => set((s) => ({ hotspots: [...s.hotspots.filter(x => !(x.lat===h.lat && x.lon===h.lon)), h] })),
  addMeetup: (m) => set((s) => ({ meetups: [...s.meetups, m] })),
  setMeetups: (ms) => set(() => ({ meetups: ms })),
}))

export type { Hotspot, Meetup }
export default useStore

import { Position } from '../store/useStore'
export function distance(a: Position, b: Position) {
  const rad = Math.PI/180
  const x = Math.sin((b.lat-a.lat)*rad/2)**2 + Math.cos(a.lat*rad)*Math.cos(b.lat*rad)*Math.sin((b.lon-a.lon)*rad/2)**2
  return 6371000*2*Math.atan2(Math.sqrt(x), Math.sqrt(Math.max(0,1-x)))
}

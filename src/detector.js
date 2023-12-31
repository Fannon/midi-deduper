import { ext } from './main.js'
import { log } from './log.js'

export function detectDuplicateNote(msg) {
  const history = ext.history.playedNotes || []
  const lastNote = findLatestNote(msg, history)

  if (lastNote) {
    const timeDiff = Math.round(msg.timestamp - lastNote.timestamp)
    if (msg.rawVelocity < ext.config.velocityThreshold) {
      log.warn(`Duplicate Note detected: Note: ${msg.note.identifier} | Velocity: ${msg.rawVelocity} | Interval: ${timeDiff}ms`)
      console.debug(msg)
      return timeDiff
    }
  }

  return false
}

function findLatestNote(latestMessage, history) {
  console.log(latestMessage)
  for (let i = history.length - 1; i >= 0; i--) {
    console.log(i, history[i])
    if (history[i].noteNumber !== latestMessage.note.number) {
      continue
    }
    const timeDiff = latestMessage.timestamp - history[i].timestamp
    if (timeDiff > ext.config.timeThreshold) {
      return undefined // stop looking if entries are too old anyway
    } 
    return history[i]
  }
  return undefined
}

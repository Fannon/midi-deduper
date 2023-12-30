import { ext } from './main.js'
import { log } from './log.js'

export function detectDuplicateNote(msg) {
  const history = ext.history.playedNotes || []
  const lastNote = history.findLast((el) => el.noteNumber === msg.note.number);

  if (lastNote) {
    const timeDiff = Math.round(msg.timestamp - lastNote.timestamp)

    if (timeDiff < ext.config.timeThreshold) {
      log.warn(`Duplicate Note detected: ${msg.note.identifier} (${msg.rawVelocity}) with interval: ${timeDiff}ms`)
      console.debug(msg)
      return true
    }
  
    console.debug(msg, lastNote, timeDiff)
  }

  return false
}

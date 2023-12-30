import { ext } from './main.js'
import { log } from './log.js'

export function detectDuplicateNote(msg) {
  const history = ext.history.playedNotes || []
  const lastNote = history.findLast((el) => el.noteNumber === msg.note.number);

  if (lastNote) {
    const timeDiff = Math.round(msg.timestamp - lastNote.timestamp)

    if (timeDiff < ext.config.timeThreshold) {
      if (msg.rawVelocity < ext.config.velocityThreshold) {
        log.warn(`Duplicate Note detected: ${msg.note.identifier} (${msg.rawVelocity}) with interval: ${timeDiff}ms`)
        console.debug(msg)
        return true
      }
    }

    // TODO: Support note-off de-duplication
    // Idea: Create map which notes are currently on. 
    //       Remember when the last duplicate note-on was
    //       If note is currently off, ignore the second note-off
    //       If note is currently on, look how quick the note-off came after the last duplicate note
    //          If this is lower than timeThreshold, ignore it
    //          Somehow ensure that note-off will be triggered after some time ?
  }

  return false
}

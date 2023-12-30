import { ext } from './main.js'
import { log } from './log.js'

export function calculateStatistics() {
  const stats = {
    notesPlayed: ext.history.playedNotes.length,
    duplicatedNotes: ext.history.duplicatedNotes.length,
  }

  stats.avgDuplicatedNoteRatio = Math.round((stats.duplicatedNotes / (stats.notesPlayed || 1)) * 100) / 100
  const intervals = ext.history.duplicatedNotes.map((el) => {
    return el.timeDiff
  })
  let total = 0;
  for (let i = 0; i < intervals.length; i++) {
      total += intervals[i];
  }
  stats.avgTimeDiff = total / intervals.length;

  console.debug(`Statistics`, stats)

  let output = `<strong>Statistics:</strong> `
  output += `Notes played: ${stats.notesPlayed} | Duplicate Notes: ${stats.duplicatedNotes} | Ratio: ${Math.round(stats.avgDuplicatedNoteRatio * 100)}% | `
  output += `Avg. time diff: ${stats.avgTimeDiff || 0}ms`
}

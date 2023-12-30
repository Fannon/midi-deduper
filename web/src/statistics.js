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

  // let table = `Statistics:`
  // table += `<table class="table table-sm">`
  // table += `<thead><tr><th scope="col"></th><th scope="col"># Notes</th><th scope="col">Ratio</th><th scope="col">Avg.</th></tr></thead>`
  // table += `<tbody>`

  // table += `<tr><th>Notes Played</th><td>${stats.notesPlayed}</td><td></td></tr>`

  
  // table += `<tr><th class="text-danger">Duplicated Notes</th><td>${stats.duplicatedNotes}</td><td>${Math.round(stats.avgDuplicatedNoteRatio * 100)}%</td></tr>`

  // table += `</tbody>`
  // table += `</table>`

  log.info(output)
}

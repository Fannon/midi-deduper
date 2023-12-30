import { log } from "./log.js";
import { initConfig, resetConfig, saveConfig } from "./config.js";
import { detectDuplicateNote } from "./detector.js";

/**
 * Global namespace, aliased to `window.ext`
 */
export const ext = {
  config: {},
  history: {
    /** Only played note-on messages */
    playedNotes: [],
    duplicatedNotes: [],
  },
  fn: {
    resetConfig,
    init,
    clearLog,
    clearHistory,
  }
}
window.ext = ext

//////////////////////////////////////////
// INIT                                 //
//////////////////////////////////////////

WebMidi.enable().then(init).catch(console.error);

// Function triggered when WEBMIDI.js is ready
async function init() {

  // Load Config
  ext.config = initConfig()

  // Setup MIDI callbacks / event listeners
  await registerUiEvents()
  await registerMidiEvents()

  log.info(`Successfully initialized.`)
}

/**
 * Register UI Events and Listeners
 */
async function registerUiEvents() {
  // UI Buttons Listeners
  document.getElementById("save").addEventListener("click", (event) => {
    saveConfig(ext.config, event)
  });
  document.getElementById("reset-config").addEventListener("click", resetConfig);
  document.getElementById("clear-log").addEventListener("click", clearLog);
  document.getElementById("clear-history").addEventListener("click", clearHistory);

  // Enable tooltips
  const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'))
  tooltipTriggerList.map(function (tooltipTriggerEl) {
    return new bootstrap.Tooltip(tooltipTriggerEl)
  })
}

/**
 * Register listeners and callbacks to MIDI events / messages
 */
async function registerMidiEvents() {

  //////////////////////////////////////////
  // INSTRUMENT INPUT                     //
  //////////////////////////////////////////

  if (ext.config.instrumentInputPort) {
    try {
      ext.input = WebMidi.getInputByName(ext.config.instrumentInputPort)
      if (!ext.input) {
        throw new Error('Could not connect to Instrument MIDI Input')
      }
      ext.input.addListener("noteon", (msg) => {
        const duplicate = detectDuplicateNote(msg, ext.history.playedNotes)

        if (!duplicate) {

          if (ext.forwardPort1) {
            ext.forwardPort1.sendNoteOn(msg.note.number, { 
              channels: msg.message.channel, 
              rawAttack: msg.rawAttack
            })
          }
          if (ext.forwardPort2) {
            ext.forwardPort2.sendNoteOn(msg.note.number, { 
              channels: msg.message.channel, 
              rawAttack: msg.rawAttack
            })
          }

          // TODO: Ensure that history does not grow endless
          if (ext.history.playedNotes.length >= 1500) {
            ext.history.playedNotes = ext.history.playedNotes.splice(500)
            console.log('Spliced', ext.history.playedNotes)
          }

          ext.history.playedNotes.push({
            time: performance.now(),
            timestamp: msg.timestamp,
            noteNumber: msg.note.number,
            rawAttack: msg.rawAttack
          })
  
          // Add it to MIDI input recording
          const jzzMsg = JZZ.MIDI.noteOn(msg.message.channel, msg.note.number, msg.rawVelocity)
          ext.recording.midiInput.track.add(ext.recording.tick, jzzMsg);
        } else {
          ext.history.duplicatedNotes.push({
            time: performance.now(),
            timestamp: msg.timestamp,
            noteNumber: msg.note.number,
            rawAttack: msg.rawAttack
          })
        }

      });
      ext.input.addListener("noteoff", (msg) => {

        // TODO: Somehow detect duplicate noteoff as well? 
        // Not entirely sure how to do this best

        // Add it to MIDI input recording
        const jzzMsg = JZZ.MIDI.noteOff(msg.message.channel, msg.note.number, msg.rawVelocity)
        ext.recording.midiInput.track.add(ext.recording.tick, jzzMsg);
      });

      log.success(`Connected to Instrument MIDI Input: ${ext.config.instrumentInputPort}`)

    } catch (err) {
      log.error(`Could not connect to Instrument MIDI Input: ${ext.config.instrumentInputPort}`)
      console.error(err)
    }
  } else {
    log.error(`No Instrument MIDI Input given.`)
  }

  //////////////////////////////////////////
  // MIDI THRU FORWARDS                   //
  //////////////////////////////////////////

  if (ext.input) {
    // Skipping 'noteon'
    const autoForwardTypes = [
      // https://webmidijs.org/api/classes/Enumerations#MIDI_CHANNEL_MESSAGES
      'noteoff', 'keyaftertouch', 'controlchange', 'programchange', 'channelaftertouch', 'pitchbend',
      // https://webmidijs.org/api/classes/Enumerations#SYSTEM_MESSAGES
      'sysex', 'timecode', 'songposition', 'songselect', 'tunerequest', 'sysexend'
    ]
    if (ext.config.forwardPort1) {
      try {
        ext.forwardPort1 = WebMidi.getOutputByName(ext.config.forwardPort1)
        if (!ext.forwardPort1) {
          throw new Error('Could not connect to Forward MIDI Port 1')
        }
        ext.input.addForwarder(ext.forwardPort1, { types: autoForwardTypes })

        log.success(`Connected MIDI Forward Port 1: ${ext.config.forwardPort1}`)
      } catch (err) {
        log.warn(`Could not connect to optional Forward Port 1: ${ext.config.forwardPort1}`)
      }
    }
    if (ext.config.forwardPort2) {
      try {
        ext.forwardPort2 = WebMidi.getOutputByName(ext.config.forwardPort2)
        if (!ext.forwardPort2) {
          throw new Error('Could not connect to Forward MIDI Port 1')
        }
        ext.input.addForwarder(ext.forwardPort2, { types: autoForwardTypes })
        log.success(`Connected MIDI Forward Port 2: ${ext.config.forwardPort2}`)
      } catch (err) {
        log.warn(`Could not connect to optional Forward Port 2: ${ext.config.forwardPort2}`)
      }
    }
  } else {
    log.warn(`No Instrument input found, cannot forward MIDI from it.`)
  }

  return
}

//////////////////////////////////////////
// HELPER FUNCTIONS                     //
//////////////////////////////////////////

function clearLog() {
  document.getElementById("log").innerHTML = ''
}

function clearHistory() {
  ext.history.playedNotes = []
  ext.history.duplicatedNotes = []
}

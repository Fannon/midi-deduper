# MIDI Deduper

## Description

This [web app](https://fannon.github.io/midi-deduper/) allows you to de-duplicate accidental MIDI double triggers, as caused e.g. by some MIDI pad controllers.
It runs on any modern [Web MIDI enabled browser](https://caniuse.com/midi). 

Why did I create this project? I have a (finger) drum pad controller, which tends to produce accidental double triggers. 
This is actually not uncommon for pad controllers, especially when the sensitivity is increased / high. 

Such accidental double triggers may be annoying when recording, but will also especially cause problems when playing with [Melodics](https://melodics.com/) or similar software that scores you on correct playing.

## Screenshot

![Screenshot](./assets/screenshot.png)

## User Guide

First, you need to have a (virtual) MIDI Loop device, e.g. like [loopMIDI](https://www.tobias-erichsen.de/software/loopmidi.html) for Windows.

In the web app, choose your MIDI Input (the device you play) and then choose the MIDI loop device.
Don't forget to hit the "Save & Apply" button after you made changes.

> 🤙 If you have questions, ideas or issues, feel free to create an [issue](https://github.com/Fannon/midi-deduper/issues) or write in the [KVR forum thread](https://www.kvraudio.com/forum/viewtopic.php?p=8819564).

### How the detection works

* A history of your played notes is kept (until a certain size)
* When a new note is incoming, the history is searched for recent notes
  * with the same note number
  * within the time threshold (configurable option)
  * below the velocity threshold (configurable option)
* If all of the above criteria apply, the note-on event is filtered out
* Note-off events are also filtered out, but by a simple "first-wins" approach (could be improved in the future)

## Developer Guide

* Requires a recent [Node.js](https://nodejs.org/en) runtime
* Git clone this repository
* Run `npm install`
* Run `npm run build` (to copy over dependencies to webapp)
* Run `npm start` 

## Go CLI Implementation

A standalone Go executable is available for running as a background tool or CLI application.

### Build Instructions

> **New to Go?** Check out our detailed [Building Guide](BUILDING.md) for step-by-step instructions.

**Quick Build (Windows):**
Simply run the included `build.bat` script. It will create two executables:
*   `midi-deduper.exe`: Standard console application (for testing/debugging).
*   `midi-deduper-headless.exe`: Hidden background application (for auto-start).

**Manual Build:**
```powershell
go mod tidy
# Standard
go build -o midi-deduper.exe ./cmd/midi-deduper
# Headless (Hidden Window)
go build -ldflags "-H=windowsgui" -o midi-deduper-headless.exe ./cmd/midi-deduper
```

### Usage

```bash
# List devices
midi-deduper.exe -list

# Run with defaults (looks for "Finger Drum Pad" -> "loop1")
midi-deduper.exe

# Run with specific devices and thresholds
midi-deduper.exe -input "My Keyboard" -output "LoopMIDI" -time 60 -velocity 100

# Run with debug logging (writes to ./logs/)
midi-deduper.exe -debug
```

> **Note:** The application includes a smart retry loop. If your MIDI devices are not connected yet (or LoopMIDI hasn't started), it will wait and retry every 5 seconds until they appear.

### Auto-Start (Windows Startup)

To run the deduper automatically when you log in without a popup window:

1.  Build the headless version (`midi-deduper-headless.exe`).
2.  Create a shortcut to `midi-deduper-headless.exe`.
3.  Press `Win+R`, type `shell:startup`, and press Enter.
4.  Move the shortcut into this folder.
5.  (Optional) Right-click the shortcut -> Properties -> Target, and add flags like `-debug` or `-input "..."`.

**Stopping the Headless Version:**
Since there is no window, use the included `stop-midi-deduper.bat` script or run `taskkill /F /IM midi-deduper-headless.exe` in a terminal.

# MIDI Deduper

## Description

This [web app](https://fannon.github.io/midi-deduper/) allows you to de-duplicate accidental MIDI double triggers, as caused e.g. by some MIDI pad controllers.
It runs on any modern [Web MIDI enabled browser](https://caniuse.com/midi). 

## Screenshot

![Synthesia Light Guide](./assets/screenshot.png)

## User Guide

First, you need to have a (virtual) MIDI Loop device, e.g. like [loopMIDI](https://www.tobias-erichsen.de/software/loopmidi.html) for Windows.

In the web app, choose your MIDI Input (the device you play) and then choose the MIDI loop device.

## TODO and Ideas

* Create a GitHub issue in this project if you have an idea or run into a problem.

## Developer Guide

* Install Node.js if not there
* Check out this repository
* Run `npm install`
* Run `npm run build` (to copy over dependencies to webapp)
* Run `npm start` 

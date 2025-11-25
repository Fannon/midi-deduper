package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fannon/midi-deduper/internal/deduper"
	"github.com/Fannon/midi-deduper/internal/logger"
	midiutil "github.com/Fannon/midi-deduper/internal/midi"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregister driver
)

const version = "1.0.0"

var (
	// CLI flags
	inputDevice       = flag.String("input", "", "Input MIDI device name (default: auto-detect from list)")
	outputDevice      = flag.String("output", "", "Output MIDI device name (default: auto-detect from list)")
	timeThreshold     = flag.Int("time", 50, "Time threshold in milliseconds for duplicate detection")
	velocityThreshold = flag.Int("velocity", 127, "Velocity threshold (0-127) for duplicate detection")
	listDevices       = flag.Bool("list", false, "List available MIDI devices and exit")
	debug             = flag.Bool("debug", false, "Enable debug logging (also writes to ./tmp/<ISO-date>.log)")
	showVersion       = flag.Bool("version", false, "Show version and exit")
	enableFlam        = flag.Bool("flam", false, "Enable flam detection (allow louder notes within threshold)")

	// Default device lists
	defaultInputs  = []string{"Finger Drum Pad"}
	defaultOutputs = []string{"loop1", "loopMIDI Port"}
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("MIDI Deduper v%s\n", version)
		os.Exit(0)
	}

	if *listDevices {
		midiutil.ListDevices()
		os.Exit(0)
	}

	// Setup logger
	appLogger, err := logger.New(*debug)
	if err != nil {
		log.Printf("Warning: Could not setup logger: %v\n", err)
	}
	defer appLogger.Close()

	if *debug {
		appLogger.Debug(fmt.Sprintf("MIDI Deduper v%s started at %s", version, time.Now().Format(time.RFC3339)))
	}

	// Initialize MIDI
	defer midi.CloseDriver()

	// Find devices with retry loop
	var inputPort drivers.In
	var inputName string
	var outputPort drivers.Out
	var outputName string

	log.Println("Waiting for MIDI devices...")

	for {
		var inErr, outErr error

		// Find input device
		if *inputDevice != "" {
			inputPort, inErr = midiutil.FindInput(*inputDevice)
			if inErr == nil {
				inputName = *inputDevice
			}
		} else {
			inputPort, inputName, inErr = midiutil.FindInputFromList(defaultInputs)
		}

		// Find output device
		if *outputDevice != "" {
			outputPort, outErr = midiutil.FindOutput(*outputDevice)
			if outErr == nil {
				outputName = *outputDevice
			}
		} else {
			outputPort, outputName, outErr = midiutil.FindOutputFromList(defaultOutputs)
		}

		// If both found, proceed
		if inErr == nil && outErr == nil {
			log.Printf("Devices found: Input=%q, Output=%q\n", inputName, outputName)
			break
		}

		// Log status and wait
		if *debug {
			appLogger.Debug(fmt.Sprintf("Devices not ready. Input err: %v. Output err: %v. Retrying in 5s...", inErr, outErr))
		}
		time.Sleep(5 * time.Second)
	}

	log.Printf("MIDI Deduper v%s started with effective configuration:\n", version)
	log.Printf("  -input=%q\n", inputName)
	log.Printf("  -output=%q\n", outputName)
	log.Printf("  -time=%d\n", *timeThreshold)
	log.Printf("  -velocity=%d\n", *velocityThreshold)
	log.Printf("  -flam=%v\n", *enableFlam)
	log.Printf("  -debug=%v\n", *debug)

	// Create deduper
	deduperConfig := deduper.Config{
		TimeThreshold:     time.Duration(*timeThreshold) * time.Millisecond,
		VelocityThreshold: uint8(*velocityThreshold),
		HistoryMaxSize:    25000,
		FlamDetection:     *enableFlam,
		Debug:             *debug,
		Logger:            appLogger.Debug,
		WarnLogger:        nil,
	}

	// Only enable warn logger if debug is enabled (to avoid spamming console in normal operation)
	if *debug {
		deduperConfig.WarnLogger = appLogger.Warn
	}

	d := deduper.New(deduperConfig)

	// Open output port
	if err := outputPort.Open(); err != nil {
		log.Fatalf("Error opening output port: %v\n", err)
	}
	defer outputPort.Close()

	// Open MIDI ports
	stop, err := midi.ListenTo(inputPort, func(msg midi.Message, timestampms int32) {
		handleMIDIMessage(msg, outputPort, d, appLogger)
	})
	if err != nil {
		log.Fatalf("Error listening to MIDI input: %v\n", err)
	}
	defer stop()

	log.Println("MIDI Deduper running. Press Ctrl+C to exit.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}

func handleMIDIMessage(msg midi.Message, output drivers.Out, d *deduper.Deduper, l *logger.Logger) {
	var channel, note, velocity uint8

	switch {
	case msg.GetNoteOn(&channel, &note, &velocity):
		// Treat Note On with velocity 0 as Note Off
		if velocity == 0 {
			handleNoteOff(output, channel, note, l)
			return
		}

		deduperNote := deduper.Note{
			Timestamp: time.Now(),
			Number:    note,
			Velocity:  velocity,
		}

		// Check if should filter
		if !d.ShouldFilter(deduperNote) {
			// Not a duplicate, forward the note
			send := midi.NoteOn(channel, note, velocity)
			err := output.Send(send)
			if err != nil {
				l.Debug(fmt.Sprintf("Error sending note on: %v", err))
			}
			l.Debug(fmt.Sprintf("Note ON:  ch=%d note=%d vel=%d", channel, note, velocity))
		} else {
			l.Debug(fmt.Sprintf("FILTERED: ch=%d note=%d vel=%d", channel, note, velocity))
		}

	case msg.GetNoteOff(&channel, &note, &velocity):
		handleNoteOff(output, channel, note, l)

	default:
		// Forward all other MIDI messages (CC, pitch bend, etc.)
		err := output.Send(msg)
		if err != nil {
			l.Debug(fmt.Sprintf("Error forwarding message: %v", err))
		}
	}
}

func handleNoteOff(output drivers.Out, channel, note uint8, l *logger.Logger) {
	// Forward note off messages (no deduplication for note off)
	send := midi.NoteOff(channel, note)
	err := output.Send(send)
	if err != nil {
		l.Debug(fmt.Sprintf("Error sending note off: %v", err))
	}
}

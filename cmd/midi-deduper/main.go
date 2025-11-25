package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Fannon/midi-deduper/internal/deduper"
	midiutil "github.com/Fannon/midi-deduper/internal/midi"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv" // autoregister driver
)

const version = "1.0.0"

var (
	// CLI flags
	inputDevice      = flag.String("input", "", "Input MIDI device name (default: auto-detect from list)")
	outputDevice     = flag.String("output", "", "Output MIDI device name (default: auto-detect from list)")
	timeThreshold    = flag.Int("time", 50, "Time threshold in milliseconds for duplicate detection")
	velocityThreshold = flag.Int("velocity", 127, "Velocity threshold (0-127) for duplicate detection")
	listDevices      = flag.Bool("list", false, "List available MIDI devices and exit")
	waitSeconds      = flag.Int("wait", 0, "Startup delay in seconds (useful for Windows Service)")
	debug            = flag.Bool("debug", false, "Enable debug logging (also writes to ./tmp/<ISO-date>.log)")
	showVersion      = flag.Bool("version", false, "Show version and exit")

	// Default device lists
	defaultInputs = []string{"Finger Drum Pad"}
	defaultOutputs = []string{"loop1", "loopMIDI Port"}

	// Debug log file
	logFile *os.File
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

	// Startup delay (e.g., wait for loopMIDI to start)
	if *waitSeconds > 0 {
		log.Printf("Waiting %d seconds before starting...\n", *waitSeconds)
		time.Sleep(time.Duration(*waitSeconds) * time.Second)
	}

	// Setup debug logging
	var logger func(string)
	if *debug {
		setupDebugLogging()
		defer closeDebugLogging()
		logger = debugLog
	}

	// Initialize MIDI
	defer midi.CloseDriver()

	// Find input device
	var inputPort drivers.In
	var inputName string
	var err error

	if *inputDevice != "" {
		inputPort, err = midiutil.FindInput(*inputDevice)
		if err != nil {
			log.Fatalf("Error finding input device '%s': %v\n", *inputDevice, err)
		}
		inputName = *inputDevice
	} else {
		inputPort, inputName, err = midiutil.FindInputFromList(defaultInputs)
		if err != nil {
			log.Fatalf("Error finding input device from defaults: %v\n", err)
		}
	}

	// Find output device
	var outputPort drivers.Out
	var outputName string

	if *outputDevice != "" {
		outputPort, err = midiutil.FindOutput(*outputDevice)
		if err != nil {
			log.Fatalf("Error finding output device '%s': %v\n", *outputDevice, err)
		}
		outputName = *outputDevice
	} else {
		outputPort, outputName, err = midiutil.FindOutputFromList(defaultOutputs)
		if err != nil {
			log.Fatalf("Error finding output device from defaults: %v\n", err)
		}
	}

	log.Printf("Connected to input: %s\n", inputName)
	log.Printf("Connected to output: %s\n", outputName)
	log.Printf("Time threshold: %dms, Velocity threshold: %d\n", *timeThreshold, *velocityThreshold)

	// Create deduper
	deduperConfig := deduper.Config{
		TimeThreshold:     time.Duration(*timeThreshold) * time.Millisecond,
		VelocityThreshold: uint8(*velocityThreshold),
		HistoryMaxSize:    25000,
		Debug:             *debug,
		Logger:            logger,
	}
	d := deduper.New(deduperConfig)

	// Open MIDI ports
	stop, err := midi.ListenTo(inputPort, func(msg midi.Message, timestampms int32) {
		handleMIDIMessage(msg, outputPort, d)
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

func handleMIDIMessage(msg midi.Message, output drivers.Out, d *deduper.Deduper) {
	var channel, note, velocity uint8

	switch {
	case msg.GetNoteOn(&channel, &note, &velocity):
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
			if err != nil && *debug {
				debugLog(fmt.Sprintf("Error sending note on: %v", err))
			}
			if *debug {
				debugLog(fmt.Sprintf("Note ON:  ch=%d note=%d vel=%d", channel, note, velocity))
			}
		} else {
			if *debug {
				debugLog(fmt.Sprintf("FILTERED: ch=%d note=%d vel=%d", channel, note, velocity))
			}
		}

	case msg.GetNoteOff(&channel, &note, &velocity):
		// Forward note off messages (no deduplication for note off)
		send := midi.NoteOff(channel, note)
		err := output.Send(send)
		if err != nil && *debug {
			debugLog(fmt.Sprintf("Error sending note off: %v", err))
		}
		if *debug {
			debugLog(fmt.Sprintf("Note OFF: ch=%d note=%d", channel, note))
		}

	default:
		// Forward all other MIDI messages (CC, pitch bend, etc.)
		err := output.Send(msg)
		if err != nil && *debug {
			debugLog(fmt.Sprintf("Error forwarding message: %v", err))
		}
	}
}

func setupDebugLogging() {
	// Create tmp directory if it doesn't exist
	tmpDir := "./tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		log.Printf("Warning: Could not create tmp directory: %v\n", err)
		return
	}

	// Create log file with ISO date
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(tmpDir, fmt.Sprintf("%s.log", timestamp))

	var err error
	logFile, err = os.Create(logPath)
	if err != nil {
		log.Printf("Warning: Could not create log file: %v\n", err)
		return
	}

	log.Printf("Debug logging to: %s\n", logPath)
	debugLog(fmt.Sprintf("MIDI Deduper v%s started at %s", version, time.Now().Format(time.RFC3339)))
	debugLog(fmt.Sprintf("Config: time=%dms, velocity=%d", *timeThreshold, *velocityThreshold))
}

func closeDebugLogging() {
	if logFile != nil {
		logFile.Close()
	}
}

func debugLog(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// Print to console
	fmt.Print(logLine)

	// Write to file
	if logFile != nil {
		logFile.WriteString(logLine)
	}
}

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

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Supervisor loop for auto-reconnection
	for {
		// Find devices with retry loop
		var inputPort drivers.In
		var inputName string
		var outputPort drivers.Out
		var outputName string

		log.Println("Waiting for MIDI devices...")

		// Device Discovery Loop
		for {
			// Check for exit signal during discovery
			select {
			case <-sigChan:
				log.Println("Shutting down...")
				return
			default:
			}

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

			// Wait before retry, but listen for exit signal
			select {
			case <-sigChan:
				log.Println("Shutting down...")
				return
			case <-time.After(5 * time.Second):
				// Continue loop
			}
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

		// Only enable warn logger if debug is enabled
		if *debug {
			deduperConfig.WarnLogger = appLogger.Warn
		}

		d := deduper.New(deduperConfig)

		// Run the session (blocks until error, device loss, or signal)
		err := runSession(inputPort, outputPort, d, appLogger, sigChan, inputName, outputName)

		// Check if we should exit or reconnect
		if err == nil {
			// Normal shutdown requested via signal
			log.Println("Shutting down...")
			return
		}

		log.Printf("Session ended: %v. Restarting discovery in 2s...\n", err)
		time.Sleep(2 * time.Second)
	}
}

// runSession manages a single connected session. It returns nil on graceful shutdown, or error on failure/disconnect.
func runSession(input drivers.In, output drivers.Out, d *deduper.Deduper, l *logger.Logger, sigChan chan os.Signal, inName, outName string) error {
	// Open output port
	if err := output.Open(); err != nil {
		return fmt.Errorf("error opening output port: %v", err)
	}
	defer output.Close()

	// Open MIDI ports
	stop, err := midi.ListenTo(input, func(msg midi.Message, timestampms int32) {
		handleMIDIMessage(msg, output, d, l)
	})
	if err != nil {
		return fmt.Errorf("error listening to MIDI input: %v", err)
	}
	defer stop()

	log.Println("MIDI Deduper running. Press Ctrl+C to exit.")

	// Watchdog ticker to check device presence
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			return nil // Graceful shutdown

		case <-ticker.C:
			// Check if devices are still present
			if !midiutil.IsDevicePresent(inName, "input") {
				return fmt.Errorf("input device %q lost", inName)
			}
			if !midiutil.IsDevicePresent(outName, "output") {
				return fmt.Errorf("output device %q lost", outName)
			}
		}
	}
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

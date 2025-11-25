package deduper

import (
	"fmt"
	"time"
)

// Config holds the deduplication configuration
type Config struct {
	TimeThreshold     time.Duration // Time threshold for duplicate detection
	VelocityThreshold uint8         // Velocity threshold (0-127)
	HistoryMaxSize    int           // Maximum size of history
	Debug             bool          // Enable debug logging
	Logger            func(string)  // Logger function
}

// Note represents a MIDI note event
type Note struct {
	Timestamp time.Time
	Number    uint8
	Velocity  uint8
}

// Deduper handles MIDI note deduplication
type Deduper struct {
	config  Config
	history []Note
}

// New creates a new Deduper instance
func New(config Config) *Deduper {
	return &Deduper{
		config:  config,
		history: make([]Note, 0, config.HistoryMaxSize),
	}
}

// ShouldFilter determines if a note should be filtered out as a duplicate
// Returns true if the note is a duplicate and should be filtered
func (d *Deduper) ShouldFilter(note Note) bool {
	lastNote := d.findLatestNote(note)

	if lastNote != nil {
		timeDiff := note.Timestamp.Sub(lastNote.Timestamp)
		if note.Velocity < d.config.VelocityThreshold {
			if d.config.Debug && d.config.Logger != nil {
				d.config.Logger(fmt.Sprintf("Duplicate Note detected: Note: %d | Velocity: %d | Interval: %v",
					note.Number, note.Velocity, timeDiff))
			}
			return true
		}
	}

	// Add to history if not a duplicate
	d.addToHistory(note)
	return false
}

// findLatestNote finds the latest matching note in history within the time threshold
func (d *Deduper) findLatestNote(latestNote Note) *Note {
	// Search backwards through history
	for i := len(d.history) - 1; i >= 0; i-- {
		if d.history[i].Number != latestNote.Number {
			continue
		}
		timeDiff := latestNote.Timestamp.Sub(d.history[i].Timestamp)
		if timeDiff > d.config.TimeThreshold {
			return nil // stop looking if entries are too old anyway
		}
		return &d.history[i]
	}
	return nil
}

// addToHistory adds a note to the history, managing size limits
func (d *Deduper) addToHistory(note Note) {
	// Ensure history does not grow endless
	if len(d.history) >= d.config.HistoryMaxSize {
		// Remove oldest 10% of history
		removeCount := d.config.HistoryMaxSize / 10
		d.history = d.history[removeCount:]
	}

	d.history = append(d.history, note)
}

package deduper

import (
	"fmt"
	"sync"
	"time"
)

// Config holds the deduplication configuration
type Config struct {
	TimeThreshold     time.Duration // Time threshold for duplicate detection
	VelocityThreshold uint8         // Velocity threshold (0-127)
	HistoryMaxSize    int           // Maximum size of history
	FlamDetection     bool          // Enable flam detection (allow louder notes)
	Debug             bool          // Enable debug logging
	Logger            func(string)  // Logger function (Debug)
	WarnLogger        func(string)  // Logger function (Warn)
}

// Note represents a MIDI note event
type Note struct {
	Timestamp time.Time
	Number    uint8
	Velocity  uint8
}

// Deduper handles MIDI note deduplication
type Deduper struct {
	config         Config
	history        []Note
	statsTotal     uint64
	statsDuplicate uint64
	mu             sync.Mutex
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
	d.mu.Lock()
	defer d.mu.Unlock()

	d.statsTotal++
	lastNote := d.findLatestNote(note)

	if lastNote != nil {
		timeDiff := note.Timestamp.Sub(lastNote.Timestamp)

		// Improvement: Allow "flams" or accents (quiet -> loud)
		// If the new note is significantly louder than the previous one, it's likely intentional
		if d.config.FlamDetection && note.Velocity > lastNote.Velocity {
			if d.config.Debug && d.config.Logger != nil {
				d.config.Logger(fmt.Sprintf("Flam/Accent detected (Allowed): Note: %d | Vel: %d > %d | Interval: %v",
					note.Number, note.Velocity, lastNote.Velocity, timeDiff))
			}
			// Update history with the new louder note so subsequent bounces are checked against this one
			d.addToHistory(note)
			return false
		}

		if note.Velocity < d.config.VelocityThreshold {
			d.statsDuplicate++
			if d.config.WarnLogger != nil {
				percentage := float64(d.statsDuplicate) / float64(d.statsTotal) * 100
				d.config.WarnLogger(fmt.Sprintf("Duplicate Note detected: Note: %d | Velocity: %d | Interval: %v | Stats: %d/%d (%.2f%%)",
					note.Number, note.Velocity, timeDiff, d.statsDuplicate, d.statsTotal, percentage))
			}
			return true
		}
	}

	// Add to history if not a duplicate
	d.addToHistory(note)
	return false
}

// findLatestNote finds the latest matching note in history within the time threshold
// This method should be called while holding the lock
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
// This method should be called while holding the lock
func (d *Deduper) addToHistory(note Note) {
	// Ensure history does not grow endless
	if len(d.history) >= d.config.HistoryMaxSize {
		// Remove oldest 10% of history
		removeCount := d.config.HistoryMaxSize / 10
		d.history = d.history[removeCount:]
	}

	d.history = append(d.history, note)
}

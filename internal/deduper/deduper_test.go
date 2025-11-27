package deduper

import (
	"testing"
	"time"
)

func TestShouldFilter_NoDuplicate(t *testing.T) {
	config := Config{
		TimeThreshold:     50 * time.Millisecond,
		VelocityThreshold: 127,
		HistoryMaxSize:    1000,
		Debug:             false,
	}
	d := New(config)

	note1 := Note{
		Timestamp: time.Now(),
		Number:    60,
		Velocity:  100,
	}

	// First note should not be filtered
	if d.ShouldFilter(note1) {
		t.Error("First note should not be filtered")
	}

	// Second note with different pitch should not be filtered
	note2 := Note{
		Timestamp: time.Now(),
		Number:    61,
		Velocity:  100,
	}
	if d.ShouldFilter(note2) {
		t.Error("Different note should not be filtered")
	}

	// Wait beyond threshold
	time.Sleep(60 * time.Millisecond)

	// Same note after threshold should not be filtered
	note3 := Note{
		Timestamp: time.Now(),
		Number:    60,
		Velocity:  100,
	}
	if d.ShouldFilter(note3) {
		t.Error("Note after threshold should not be filtered")
	}
}

func TestShouldFilter_Duplicate(t *testing.T) {
	config := Config{
		TimeThreshold:     50 * time.Millisecond,
		VelocityThreshold: 127,
		HistoryMaxSize:    1000,
		Debug:             false,
	}
	d := New(config)

	now := time.Now()
	note1 := Note{
		Timestamp: now,
		Number:    60,
		Velocity:  80,
	}

	// First note should not be filtered
	if d.ShouldFilter(note1) {
		t.Error("First note should not be filtered")
	}

	// Second note within threshold should be filtered
	note2 := Note{
		Timestamp: now.Add(10 * time.Millisecond),
		Number:    60,
		Velocity:  75,
	}
	if !d.ShouldFilter(note2) {
		t.Error("Duplicate note should be filtered")
	}
}

func TestShouldFilter_HighVelocityNotFiltered(t *testing.T) {
	config := Config{
		TimeThreshold:     50 * time.Millisecond,
		VelocityThreshold: 80,
		HistoryMaxSize:    1000,
		Debug:             false,
	}
	d := New(config)

	now := time.Now()
	note1 := Note{
		Timestamp: now,
		Number:    60,
		Velocity:  100,
	}

	d.ShouldFilter(note1)

	// Second note within threshold but high velocity should not be filtered
	note2 := Note{
		Timestamp: now.Add(10 * time.Millisecond),
		Number:    60,
		Velocity:  100,
	}
	if d.ShouldFilter(note2) {
		t.Error("Note with velocity >= threshold should not be filtered")
	}
}

func TestHistorySizeLimit(t *testing.T) {
	config := Config{
		TimeThreshold:     50 * time.Millisecond,
		VelocityThreshold: 127,
		HistoryMaxSize:    100,
		Debug:             false,
	}
	d := New(config)

	// Add more notes than max size
	for i := 0; i < 150; i++ {
		note := Note{
			Timestamp: time.Now(),
			Number:    uint8(i % 128),
			Velocity:  100,
		}
		d.ShouldFilter(note)
	}

	// History should be trimmed
	if len(d.history) > config.HistoryMaxSize {
		t.Errorf("History size %d exceeds max %d", len(d.history), config.HistoryMaxSize)
	}
}

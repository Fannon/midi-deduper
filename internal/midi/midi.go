package midi

import (
	"fmt"
	"strings"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

// FindInput finds a MIDI input device by name (case-insensitive, suffix-ignoring)
// Windows sometimes adds suffixes like " 2", " 3" when USB port changes
func FindInput(name string) (drivers.In, error) {
	inputs := midi.GetInPorts()
	
	// Try exact match first
	for _, in := range inputs {
		if strings.EqualFold(in.String(), name) {
			return in, nil
		}
	}
	
	// Try prefix match (ignore Windows suffixes)
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	for _, in := range inputs {
		deviceName := strings.ToLower(strings.TrimSpace(in.String()))
		// Check if device name starts with the search name
		if strings.HasPrefix(deviceName, normalizedName) {
			return in, nil
		}
	}
	
	return nil, fmt.Errorf("MIDI input device '%s' not found", name)
}

// FindOutput finds a MIDI output device by name (case-insensitive, suffix-ignoring)
// Windows sometimes adds suffixes like " 2", " 3" when USB port changes
func FindOutput(name string) (drivers.Out, error) {
	outputs := midi.GetOutPorts()
	
	// Try exact match first
	for _, out := range outputs {
		if strings.EqualFold(out.String(), name) {
			return out, nil
		}
	}
	
	// Try prefix match (ignore Windows suffixes)
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	for _, out := range outputs {
		deviceName := strings.ToLower(strings.TrimSpace(out.String()))
		// Check if device name starts with the search name
		if strings.HasPrefix(deviceName, normalizedName) {
			return out, nil
		}
	}
	
	return nil, fmt.Errorf("MIDI output device '%s' not found", name)
}

// FindInputFromList tries to find the first available input device from a list of names
func FindInputFromList(names []string) (drivers.In, string, error) {
	for _, name := range names {
		if name == "" {
			continue
		}
		in, err := FindInput(name)
		if err == nil {
			return in, name, nil
		}
	}
	return nil, "", fmt.Errorf("no MIDI input device found from list: %v", names)
}

// FindOutputFromList tries to find the first available output device from a list of names
func FindOutputFromList(names []string) (drivers.Out, string, error) {
	for _, name := range names {
		if name == "" {
			continue
		}
		out, err := FindOutput(name)
		if err == nil {
			return out, name, nil
		}
	}
	return nil, "", fmt.Errorf("no MIDI output device found from list: %v", names)
}

// ListDevices prints all available MIDI devices
func ListDevices() {
	fmt.Println("Available MIDI Input Devices:")
	inputs := midi.GetInPorts()
	if len(inputs) == 0 {
		fmt.Println("  (none)")
	}
	for i, in := range inputs {
		fmt.Printf("  %d: %s\n", i, in.String())
	}
	
	fmt.Println("\nAvailable MIDI Output Devices:")
	outputs := midi.GetOutPorts()
	if len(outputs) == 0 {
		fmt.Println("  (none)")
	}
	for i, out := range outputs {
		fmt.Printf("  %d: %s\n", i, out.String())
	}
}


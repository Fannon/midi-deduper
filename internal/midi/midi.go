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
	return findPort(midi.GetInPorts(), name, "input")
}

// FindOutput finds a MIDI output device by name (case-insensitive, suffix-ignoring)
// Windows sometimes adds suffixes like " 2", " 3" when USB port changes
func FindOutput(name string) (drivers.Out, error) {
	return findPort(midi.GetOutPorts(), name, "output")
}

// findPort is a generic helper to find a MIDI port by name
func findPort[P fmt.Stringer](ports []P, name, portType string) (P, error) {
	var zero P

	// Try exact match first
	for _, port := range ports {
		if strings.EqualFold(port.String(), name) {
			return port, nil
		}
	}

	// Try prefix match (ignore Windows suffixes)
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	for _, port := range ports {
		deviceName := strings.ToLower(strings.TrimSpace(port.String()))
		// Check if device name starts with the search name
		if strings.HasPrefix(deviceName, normalizedName) {
			return port, nil
		}
	}

	return zero, fmt.Errorf("MIDI %s device '%s' not found", portType, name)
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

// IsDevicePresent checks if a device with the given name is currently available.
// It uses the same flexible matching as FindInput/FindOutput (exact match first, then prefix).
func IsDevicePresent(name, portType string) bool {
	var ports []fmt.Stringer
	if portType == "input" {
		inPorts := midi.GetInPorts()
		ports = make([]fmt.Stringer, len(inPorts))
		for i, p := range inPorts {
			ports[i] = p
		}
	} else {
		outPorts := midi.GetOutPorts()
		ports = make([]fmt.Stringer, len(outPorts))
		for i, p := range outPorts {
			ports[i] = p
		}
	}

	// Try exact match first
	for _, port := range ports {
		if strings.EqualFold(port.String(), name) {
			return true
		}
	}

	// Try prefix match (ignore Windows suffixes like " 2", " 3")
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	for _, port := range ports {
		deviceName := strings.ToLower(strings.TrimSpace(port.String()))
		if strings.HasPrefix(deviceName, normalizedName) {
			return true
		}
	}

	return false
}

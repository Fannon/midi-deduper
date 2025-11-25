# Go Migration Report

## Status Summary
The Go executable implementation is code-complete but currently fails to build in the WSL2 environment due to missing Linux audio drivers (ALSA). This is expected behavior for WSL2 without specific configuration. The code is designed to work on Windows using the standard `rtmididrv` driver.

## 1. Created Components

### Core Logic (`internal/deduper/`)
- **Status**: ✅ Complete & Verified
- **Details**: 
  - Ported the exact logic from `web/src/detector.js`.
  - Includes `TimeThreshold` and `VelocityThreshold` logic.
  - Includes history management (circular buffer/slice trimming).
  - **Verification**: Unit tests in `deduper_test.go` pass successfully.

### MIDI Handling (`internal/midi/`)
- **Status**: ✅ Code Complete (Build Pending)
- **Details**:
  - Implemented device discovery.
  - Added smart matching:
    - Case-insensitive matching.
    - Prefix matching to handle Windows suffixes (e.g., "Device Name 2").
  - Helper functions to find input/output from a priority list.

### CLI Application (`cmd/midi-deduper/`)
- **Status**: ✅ Code Complete (Build Pending)
- **Details**:
  - **Flags**:
    - `-input`, `-output`: Specify devices (fuzzy matched).
    - `-time`, `-velocity`: Configuration thresholds.
    - `-wait`: Startup delay (useful for Services).
    - `-debug`: Verbose logging.
    - `-list`: List available devices.
  - **Logging**: 
    - Implemented debug logging to both console and file (`./tmp/<timestamp>.log`).
  - **Loopback**:
    - Implemented the main loop: Read Note -> Dedupe -> Forward.
    - Handles NoteOn and NoteOff correctly.
    - Forwards other messages (CC, etc.) transparently.

## 2. Build Issues & Resolution

### Current Error
The build fails with `fatal error: alsa/asoundlib.h: No such file or directory`.
- **Cause**: The `rtmididrv` (RealTime MIDI Driver) uses CGO to interface with the OS audio subsystem. On Linux (WSL2), it looks for ALSA. On Windows, it will look for the Windows MM API.

### How to Build on Windows
1. **Prerequisites**:
   - **Go**: Install Go for Windows.
   - **C Compiler**: Since `rtmididrv` uses CGO, you need a C compiler.
     - Recommended: **TDM-GCC** (http://tdm-gcc.tdragon.net/) or **MinGW-w64**.
     - Ensure `gcc` is in your system PATH.

2. **Build Command**:
   ```powershell
   go mod tidy
   go build -o midi-deduper.exe ./cmd/midi-deduper
   ```

## 3. Remaining Tasks

### Documentation
The following documentation needs to be added to `README.md` once the build is verified.

#### Usage
```bash
# List devices
midi-deduper.exe -list

# Run with defaults (looks for "Finger Drum Pad" -> "loop1")
midi-deduper.exe

# Run with specific devices and thresholds
midi-deduper.exe -input "My Keyboard" -output "LoopMIDI" -time 60 -velocity 100

# Run with debug logging (writes to ./tmp/)
midi-deduper.exe -debug
```

#### Windows Service Setup (NSSM)
To run as a background service that starts automatically:

1. Download **NSSM** (Non-Sucking Service Manager).
2. Install the service:
   ```powershell
   nssm install MidiDeduper "C:\path\to\midi-deduper.exe"
   ```
3. Configure arguments in NSSM GUI:
   - Arguments: `-wait 10 -input "My Device" -output "loopMIDI"`
   - The `-wait 10` is important to allow the loopMIDI driver to initialize before the deduper starts.
4. Start the service:
   ```powershell
   nssm start MidiDeduper
   ```

## 4. Next Steps
1. Pull this branch (`try-golang-version`) on your Windows machine.
2. Ensure GCC is installed.
3. Build the executable.
4. Verify with real MIDI devices.

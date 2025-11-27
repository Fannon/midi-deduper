# Building and Testing MIDI Deduper (Go)

Welcome! This guide will help you compile and test the Go version of MIDI Deduper. 

Since this project interacts with hardware (MIDI devices), it uses a library that connects to the operating system's audio/MIDI subsystem. This means there are a few extra steps compared to a "pure" Go program.

## Prerequisites

Before you start, you need two main things installed:

1.  **Go Language**: The programming language itself.
2.  **C Compiler (GCC)**: Required because the MIDI driver (`rtmididrv`) uses CGO to talk to Windows/Linux/macOS APIs.

### 1. Install Go
*   Download and install from [go.dev/dl](https://go.dev/dl/).
*   Verify installation by opening a terminal and running:
    ```bash
    go version
    ```

#### macOS
Install Xcode Command Line Tools:
```bash
xcode-select --install
```

---

## How to Build

Once prerequisites are met, building is simple.

### Option 1: Automatic Build

**Bash (Git Bash / WSL):**
```bash
./build.sh
```

**Windows CMD:**
```cmd
build.bat
```

This will create both:
*   `midi-deduper.exe` (Console version)
*   `midi-deduper-headless.exe` (Background version)

### Option 2: Manual Build

1.  Open your terminal in the project directory.
2.  Download dependencies:
    ```bash
    go mod tidy
    ```
3.  Compile the standard code:
    ```bash
    go build -ldflags "-s -w" -o midi-deduper.exe ./cmd/midi-deduper
    ```
4.  Compile the headless code (Windows only):
    ```bash
    go build -ldflags "-H=windowsgui -s -w" -o midi-deduper-headless.exe ./cmd/midi-deduper
    ```

5.  **Success!** You should now see the executables in your folder.

---

## Common Issues

### "build constraints exclude all Go files..."
**Cause**: Go cannot find a C compiler (GCC), so it disabled CGO. The MIDI driver requires CGO.
**Fix**: Install GCC (see Prerequisites above) and ensure `gcc` works in your terminal.

### "fatal error: alsa/asoundlib.h: No such file or directory"
**Cause**: You are on Linux (or WSL) and missing the ALSA development headers.
**Fix**: Run `sudo apt-get install libasound2-dev`.

### "gcc: command not found"
**Cause**: GCC is not installed or not in your system PATH.
**Fix**: Reinstall TDM-GCC (Windows) and restart your terminal.

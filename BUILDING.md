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

### 2. Install GCC (C Compiler)

#### Windows
You need a GCC compiler for Windows. We recommend **TDM-GCC**.
1.  Download **TDM-GCC** from [jmeubank.github.io/tdm-gcc/](https://jmeubank.github.io/tdm-gcc/).
2.  Run the installer. **Important**: Ensure the "Add to Path" option is checked (it usually is by default).
3.  After installation, **restart your terminal** (or VS Code) to pick up the new path.
4.  Verify by running:
    ```bash
    gcc --version
    ```
    *If you see a version number, you are ready!*

#### Linux (Ubuntu/Debian/WSL)
You need the build essentials and ALSA sound headers.
```bash
sudo apt-get update
sudo apt-get install build-essential libasound2-dev
```

#### macOS
Install Xcode Command Line Tools:
```bash

To run all tests:
```bash
go test ./...
```

To run tests with verbose output (shows each test name):
```bash
go test -v ./...
```

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

# Go Language Tutorial: Learning with `midi-deduper`

Welcome to Go! This tutorial is designed to help you learn the basics of the Go programming language (Golang) by exploring the code within this project (`midi-deduper`). Instead of abstract examples, we'll look at how real concepts are applied in this MIDI deduplication tool.

## 1. Why Go?

Go is known for:
*   **Simplicity**: It has a small keyword set and is easy to read.
*   **Performance**: It compiles to fast machine code.
*   **Concurrency**: It handles multiple tasks (like processing MIDI events) efficiently.
*   **Static Typing**: It catches errors at compile time, not runtime.

## 2. Project Structure

Go projects typically follow a standard layout. Let's look at ours:

*   **`go.mod`**: This file defines the module and its dependencies. It's like `package.json` in Node.js.
*   **`cmd/`**: Contains the main applications.
    *   `cmd/midi-deduper/main.go`: The entry point for our program.
*   **`internal/`**: Contains code that is private to this project and shouldn't be imported by other projects.
    *   `internal/deduper/`: The core logic for detecting duplicate notes.
    *   `internal/midi/`: Helper functions for MIDI device discovery.

## 3. Language Basics

Let's dive into the code to see Go concepts in action.

### Packages & Imports

Every Go file starts with a `package` declaration.

*   **Executable Programs**: Files that run as an application must belong to `package main`.
    *   *See `cmd/midi-deduper/main.go` line 1.*
*   **Libraries**: Reusable code belongs to other packages.
    *   *See `internal/deduper/deduper.go` line 1: `package deduper`.*

Imports allow you to use code from other packages.

```go
// cmd/midi-deduper/main.go

import (
    "fmt"   // Standard library for formatting text
    "time"  // Standard library for time handling

    // Third-party library (defined in go.mod)
    "gitlab.com/gomidi/midi/v2"
)
```

### Variables

Go has a strong type system. You can declare variables in a few ways.

**Global Variables (with `var` block):**
In `cmd/midi-deduper/main.go`, we define command-line flags:

```go
var (
    inputDevice   = flag.String("input", "", "Input MIDI device name")
    timeThreshold = flag.Int("time", 50, "Time threshold in ms")
)
```
*Note: These return pointers (e.g., `*int`), so we access their values with `*timeThreshold`.*

**Short Variable Declaration (`:=`):**
Inside functions, you'll often see `:=`. This tells Go to infer the type automatically.

```go
// internal/deduper/deduper.go

// 'd' is automatically inferred as type *Deduper
d := &Deduper{
    config: config,
}
```

### Structs (Data Structures)

Structs are collections of fields. They are similar to Classes in other languages but without inheritance.

Look at `internal/deduper/deduper.go`:

```go
type Note struct {
    Timestamp time.Time
    Number    uint8
    Velocity  uint8
}

type Deduper struct {
    config     Config
    history    []Note      // A slice (dynamic array) of Notes
    mu         sync.Mutex  // A lock for thread safety
}
```

### Functions & Methods

**Functions** are standalone blocks of code.
**Methods** are functions attached to a specific type (struct).

**Function Example:**
In `internal/deduper/deduper.go`, `New` is a "constructor" function (a common Go pattern):

```go
func New(config Config) *Deduper {
    return &Deduper{ ... }
}
```

**Method Example:**
`ShouldFilter` is a method attached to the `*Deduper` struct. It has access to the struct's data via `d`.

```go
//      Receiver
//         |
func (d *Deduper) ShouldFilter(note Note) bool {
    // Logic to check if note is a duplicate
}
```

### Control Flow

**If / Else:**
Standard conditional logic. Note that parentheses `()` are not required around the condition.

```go
// internal/midi/midi.go
if err == nil {
    return in, name, nil
}
```

**Loops (`for`):**
Go only has `for`. It acts as `while`, `foreach`, and standard `for`.

*   **Infinite Loop (While True):** Used in `main.go` for the supervisor loop.
    ```go
    for {
        // Keep trying to connect to devices
    }
    ```
*   **Range Loop (Foreach):** Used in `midi.go` to iterate over devices.
    ```go
    for _, name := range names {
        // check name
    }
    ```

**Switch:**
Used in `main.go` to handle different MIDI message types.

```go
switch {
case msg.GetNoteOn(&channel, &note, &velocity):
    // Handle Note On
case msg.GetNoteOff(&channel, &note, &velocity):
    // Handle Note Off
default:
    // Handle everything else
}
```

### Error Handling

Go doesn't use exceptions (try/catch). Instead, functions return an `error` value. You must check it.

```go
// cmd/midi-deduper/main.go

// output.Open() returns an error if it fails
if err := output.Open(); err != nil {
    // Handle the error
    return fmt.Errorf("error opening output port: %v", err)
}
```

### Defer

`defer` schedules a function call to run immediately before the current function returns. It's perfect for cleanup.

```go
// cmd/midi-deduper/main.go

// Ensure the logger file is closed when main() exits
defer appLogger.Close()
```

### Concurrency (Goroutines & Channels)

Go is famous for concurrency.

*   **Goroutines**: Lightweight threads.
*   **Channels**: Pipes for communicating between goroutines.

In `cmd/midi-deduper/main.go`, we use a channel to listen for the "Stop" signal (Ctrl+C):

```go
// Create a channel that can hold os.Signal data
sigChan := make(chan os.Signal, 1)

// Tell the OS to send Interrupt signals to this channel
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// Block and wait until we receive a signal
<-sigChan
```

## 4. Running the Project

To run the project locally without compiling a binary file:

```bash
go run ./cmd/midi-deduper
```

To compile it into an executable (`.exe` on Windows):

```bash
go build ./cmd/midi-deduper
```

## 5. Try It Yourself!

To practice, try making a small change:

1.  Open `cmd/midi-deduper/main.go`.
2.  Find the `main()` function.
3.  Add a `fmt.Println("Hello from the tutorial!")` at the very beginning of the function.
4.  Run `go run ./cmd/midi-deduper` and see your message.

Happy Coding!

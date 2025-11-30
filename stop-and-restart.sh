#!/bin/bash

echo "Stopping MIDI Deduper..."

# Forcefully terminate the processes using taskkill.
# Redirecting output to null to suppress "not found" errors.
taskkill /F /IM midi-deduper-headless.exe > /dev/null 2>&1
taskkill /F /IM midi-deduper.exe > /dev/null 2>&1

echo "Done."
echo ""
echo "Starting MIDI Deduper..."

# Check for the executable in the current directory first, then in PATH.
if [ -f "midi-deduper.exe" ]; then
    ./midi-deduper.exe
else
    midi-deduper.exe
fi

echo ""
read -p "Press Enter to exit."

#!/bin/bash
echo "Building MIDI Deduper..."

echo "1. Building Standard Version (Console)..."
go build -ldflags "-s -w" -o midi-deduper.exe ./cmd/midi-deduper
if [ $? -ne 0 ]; then exit 1; fi

echo "2. Building Headless Version (No Window)..."
go build -ldflags "-H=windowsgui -s -w" -o midi-deduper-headless.exe ./cmd/midi-deduper
if [ $? -ne 0 ]; then exit 1; fi

echo ""
echo "Build Complete!"
echo "- midi-deduper.exe (Standard)"
echo "- midi-deduper-headless.exe (Background/Startup)"

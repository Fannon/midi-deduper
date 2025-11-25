@echo off
echo Building MIDI Deduper...

echo 1. Building Standard Version (Console)...
go build -o midi-deduper.exe ./cmd/midi-deduper
if %errorlevel% neq 0 exit /b %errorlevel%

echo 2. Building Headless Version (No Window)...
go build -ldflags "-H=windowsgui" -o midi-deduper-headless.exe ./cmd/midi-deduper
if %errorlevel% neq 0 exit /b %errorlevel%

echo.
echo Build Complete!
echo - midi-deduper.exe (Standard)
echo - midi-deduper-headless.exe (Background/Startup)
pause

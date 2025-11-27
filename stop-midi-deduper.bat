@echo off
echo Stopping MIDI Deduper (Headless)...
taskkill /F /IM midi-deduper-headless.exe
taskkill /F /IM midi-deduper.exe
echo.
echo Done.
pause

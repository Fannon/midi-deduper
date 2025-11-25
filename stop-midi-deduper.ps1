Write-Host "Stopping MIDI Deduper..."
Stop-Process -Name "midi-deduper-headless" -ErrorAction SilentlyContinue
Stop-Process -Name "midi-deduper" -ErrorAction SilentlyContinue
Write-Host "Done."

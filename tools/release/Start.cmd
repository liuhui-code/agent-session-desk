@echo off
"%~dp0asd-launch.exe"
if errorlevel 1 (
  echo Could not open Agent Session Desk. Keep the data directory for diagnosis.
  pause
)
